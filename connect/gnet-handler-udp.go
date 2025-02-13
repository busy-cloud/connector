package connect

import (
	"fmt"
	"github.com/busy-cloud/boat/log"
	"github.com/busy-cloud/boat/mqtt"
	"github.com/panjf2000/gnet/v2"
	"time"
)

type GNetHandlerUdp struct {
	*Connect
	*GNetServer

	buf [4096]byte
}

func (h *GNetHandlerUdp) OnBoot(eng gnet.Engine) (action gnet.Action) {
	h.GNetServer.opened = true
	h.GNetServer.engine = eng
	return gnet.None
}

func (h *GNetHandlerUdp) OnShutdown(eng gnet.Engine) {
	h.GNetServer.opened = true
}

func (h *GNetHandlerUdp) OnOpen(c gnet.Conn) (out []byte, action gnet.Action) {
	log.Println("GNetHandlerUdp OnOpen")
	h.connected = true
	return nil, gnet.None
}

func (h *GNetHandlerUdp) OnClose(c gnet.Conn, err error) (action gnet.Action) {
	log.Println("GNetHandlerUdp OnClose")
	return gnet.None
}

func (h *GNetHandlerUdp) OnTraffic(c gnet.Conn) (action gnet.Action) {
	h.connected = true

	n, e := c.Read(h.buf[:])
	if e != nil {
		return gnet.Close
	}

	topic := fmt.Sprintf("link/%s/up", h.Id)
	mqtt.Client.Publish(topic, 0, false, h.buf[:n])

	//保存连接
	connections.LoadOrStore(h.Id, c)

	return gnet.None
}

func (h *GNetHandlerUdp) OnTick() (delay time.Duration, action gnet.Action) {
	return 0, gnet.None
}
