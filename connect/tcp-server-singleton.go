package connect

import (
	"fmt"
	"github.com/busy-cloud/boat/mqtt"
	"github.com/panjf2000/gnet/v2"
	"sync/atomic"
	"time"
)

type TcpServerSingleton struct {
	*Connect
	*Server

	buf   [4096]byte
	count atomic.Int64
}

func (h *TcpServerSingleton) OnBoot(eng gnet.Engine) (action gnet.Action) {
	h.Server.opened = true
	h.Server.engine = eng
	return gnet.None
}

func (h *TcpServerSingleton) OnShutdown(eng gnet.Engine) {
	h.Server.opened = true
}

func (h *TcpServerSingleton) OnOpen(c gnet.Conn) (out []byte, action gnet.Action) {
	h.count.Add(1)
	h.connected = true
	//上线
	topic := fmt.Sprintf("link/%s/opened", h.Id)
	mqtt.Client.Publish(topic, 0, false, c.RemoteAddr().String())

	connections.Store(h.Id, c)

	return nil, gnet.None
}

func (h *TcpServerSingleton) OnClose(c gnet.Conn, err error) (action gnet.Action) {
	n := h.count.Add(-1)
	if n <= 0 {
		h.Server.connected = false
	}

	//下线
	topic := fmt.Sprintf("link/%s/close", h.Id)
	mqtt.Client.Publish(topic, 0, false, err.Error())

	connections.Delete(h.Id)

	return gnet.None
}

func (h *TcpServerSingleton) OnTraffic(c gnet.Conn) (action gnet.Action) {

	n, e := c.Read(h.buf[:])
	if e != nil {
		return gnet.Close
	}
	read := string(h.buf[:n])

	//_, _ = c.Write([]byte("you are " + cc["id"].(string)))
	topic := fmt.Sprintf("link/%s/up", h.Id)
	mqtt.Client.Publish(topic, 0, false, read)

	return gnet.None
}

func (h *TcpServerSingleton) OnTick() (delay time.Duration, action gnet.Action) {
	return 0, gnet.None
}
