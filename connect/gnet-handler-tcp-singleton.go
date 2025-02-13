package connect

import (
	"fmt"
	"github.com/busy-cloud/boat/mqtt"
	"github.com/panjf2000/gnet/v2"
	"sync/atomic"
	"time"
)

type GNetHandlerTcpSingleton struct {
	*Connect
	*GNetServer

	buf   [4096]byte
	count atomic.Int64
}

func (h *GNetHandlerTcpSingleton) OnBoot(eng gnet.Engine) (action gnet.Action) {
	h.GNetServer.opened = true
	h.GNetServer.engine = eng
	return gnet.None
}

func (h *GNetHandlerTcpSingleton) OnShutdown(eng gnet.Engine) {
	h.GNetServer.opened = true
}

func (h *GNetHandlerTcpSingleton) OnOpen(c gnet.Conn) (out []byte, action gnet.Action) {
	h.count.Add(1)
	h.connected = true
	//上线
	topic := fmt.Sprintf("link/%s/open", h.Id)
	mqtt.Client.Publish(topic, 0, false, c.RemoteAddr().String())

	connections.Store(h.Id, c)

	return nil, gnet.None
}

func (h *GNetHandlerTcpSingleton) OnClose(c gnet.Conn, err error) (action gnet.Action) {
	n := h.count.Add(-1)
	if n <= 0 {
		h.GNetServer.connected = false
	}

	//下线
	topic := fmt.Sprintf("link/%s/close", h.Id)
	mqtt.Client.Publish(topic, 0, false, err.Error())

	connections.Delete(h.Id)

	return gnet.None
}

func (h *GNetHandlerTcpSingleton) OnTraffic(c gnet.Conn) (action gnet.Action) {

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

func (h *GNetHandlerTcpSingleton) OnTick() (delay time.Duration, action gnet.Action) {
	return 0, gnet.None
}
