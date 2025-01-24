package connect

import (
	"fmt"
	"github.com/busy-cloud/boat/mqtt"
	"github.com/panjf2000/gnet/v2"
	"time"
)

type TcpServerSingleton struct {
	ServerId  string
	Singleton bool //单例

	buf [4096]byte
}

func (h *TcpServerSingleton) OnBoot(eng gnet.Engine) (action gnet.Action) {
	return gnet.None
}

func (h *TcpServerSingleton) OnShutdown(eng gnet.Engine) {
}

func (h *TcpServerSingleton) OnOpen(c gnet.Conn) (out []byte, action gnet.Action) {

	//上线
	topic := fmt.Sprintf("tunnel/%s/open", h.ServerId)
	mqtt.Client.Publish(topic, 0, false, c.RemoteAddr().String())

	connections.Store(h.ServerId, c)

	return nil, gnet.None
}

func (h *TcpServerSingleton) OnClose(c gnet.Conn, err error) (action gnet.Action) {

	//下线
	topic := fmt.Sprintf("tunnel/%s/close", h.ServerId)
	mqtt.Client.Publish(topic, 0, false, err.Error())

	connections.Delete(h.ServerId)

	return gnet.None
}

func (h *TcpServerSingleton) OnTraffic(c gnet.Conn) (action gnet.Action) {

	n, e := c.Read(h.buf[:])
	if e != nil {
		return gnet.Close
	}
	read := string(h.buf[:n])

	//_, _ = c.Write([]byte("you are " + cc["id"].(string)))
	topic := fmt.Sprintf("tunnel/%s/up", h.ServerId)
	mqtt.Client.Publish(topic, 0, false, read)

	return gnet.None
}

func (h *TcpServerSingleton) OnTick() (delay time.Duration, action gnet.Action) {
	return 0, gnet.None
}
