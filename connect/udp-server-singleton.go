package connect

import (
	"fmt"
	"github.com/busy-cloud/boat/log"
	"github.com/busy-cloud/boat/mqtt"
	"github.com/panjf2000/gnet/v2"
	"time"
)

type UdpServerSingleton struct {
	ServerId string

	buf [4096]byte
}

func (h *UdpServerSingleton) OnBoot(eng gnet.Engine) (action gnet.Action) {
	return gnet.None
}

func (h *UdpServerSingleton) OnShutdown(eng gnet.Engine) {
}

func (h *UdpServerSingleton) OnOpen(c gnet.Conn) (out []byte, action gnet.Action) {
	log.Println("UdpServerSingleton OnOpen")
	return nil, gnet.None
}

func (h *UdpServerSingleton) OnClose(c gnet.Conn, err error) (action gnet.Action) {
	log.Println("UdpServerSingleton OnClose")
	return gnet.None
}

func (h *UdpServerSingleton) OnTraffic(c gnet.Conn) (action gnet.Action) {
	n, e := c.Read(h.buf[:])
	if e != nil {
		return gnet.Close
	}

	topic := fmt.Sprintf("tunnel/%s/up", h.ServerId)
	mqtt.Client.Publish(topic, 0, false, h.buf[:n])

	//保存连接
	connections.LoadOrStore(h.ServerId, c)

	return gnet.None
}

func (h *UdpServerSingleton) OnTick() (delay time.Duration, action gnet.Action) {
	return 0, gnet.None
}
