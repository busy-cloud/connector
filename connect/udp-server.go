package connect

import (
	"encoding/hex"
	"fmt"
	"github.com/busy-cloud/boat/log"
	"github.com/busy-cloud/boat/mqtt"
	"github.com/panjf2000/gnet/v2"
	"time"
)

type UdpServer struct {
	ServerId       string
	IdStart, IdEnd int

	buf [4096]byte
}

func (h *UdpServer) OnBoot(eng gnet.Engine) (action gnet.Action) {
	return gnet.None
}

func (h *UdpServer) OnShutdown(eng gnet.Engine) {
}

func (h *UdpServer) OnOpen(c gnet.Conn) (out []byte, action gnet.Action) {
	log.Println("UdpServer OnOpen")
	return nil, gnet.None
}

func (h *UdpServer) OnClose(c gnet.Conn, err error) (action gnet.Action) {
	log.Println("UdpServer OnClose")
	return gnet.None
}

func (h *UdpServer) OnTraffic(c gnet.Conn) (action gnet.Action) {
	n, e := c.Read(h.buf[:])
	if e != nil {
		return gnet.Close
	}

	//验证合法性
	if n < h.IdEnd {
		_, _ = c.Write([]byte("packet too short"))
		return gnet.Close
	}

	//id := string(h.buf[h.IdStart:h.IdEnd])
	id := hex.EncodeToString(h.buf[h.IdStart:h.IdEnd])

	topic := fmt.Sprintf("tunnel/%s/%s/up", id, h.ServerId)
	mqtt.Client.Publish(topic, 0, false, h.buf[:n])

	//保存连接
	connections.LoadOrStore(id, c)

	return gnet.None
}

func (h *UdpServer) OnTick() (delay time.Duration, action gnet.Action) {
	return 0, gnet.None
}
