package connect

import (
	"fmt"
	"github.com/busy-cloud/boat/mqtt"
	"github.com/panjf2000/gnet/v2"
	"regexp"
	"time"
)

var idReg = regexp.MustCompile(`^\w{2,128}`)

type TcpServer struct {
	ServerId string

	buf [4096]byte
}

func (h *TcpServer) OnBoot(eng gnet.Engine) (action gnet.Action) {
	return gnet.None
}

func (h *TcpServer) OnShutdown(eng gnet.Engine) {
}

func (h *TcpServer) OnOpen(c gnet.Conn) (out []byte, action gnet.Action) {
	return nil, gnet.None
}

func (h *TcpServer) OnClose(c gnet.Conn, err error) (action gnet.Action) {
	ctx := c.Context()
	if ctx == nil {
		return gnet.None
	}

	cc, ok := ctx.(map[string]interface{})
	if !ok {
		return gnet.None
	}
	id := cc["id"].(string)

	//下线
	topic := fmt.Sprintf("tunnel/%s/%s/close", id, h.ServerId)
	mqtt.Client.Publish(topic, 0, false, err.Error())

	//从池中清除
	connections.Delete(id)

	return gnet.None
}

func (h *TcpServer) OnTraffic(c gnet.Conn) (action gnet.Action) {
	ctx := c.Context()

	//检查首个包, 作为注册包
	if ctx == nil {
		n, e := c.Read(h.buf[:])
		if e != nil {
			return gnet.Close
		}
		id := string(h.buf[:n])

		//验证合法性
		if !idReg.MatchString(id) {
			_, _ = c.Write([]byte("invalid id"))
			return gnet.Close
		}

		ctx = map[string]interface{}{"id": id}
		c.SetContext(ctx)

		//上线
		topic := fmt.Sprintf("tunnel/%s/%s/open", id, h.ServerId)
		mqtt.Client.Publish(topic, 0, false, c.RemoteAddr().String())

		//保存连接
		connections.Store(id, c)

		return gnet.None
	}

	//取出上下文
	cc, ok := ctx.(map[string]interface{})
	if !ok {
		_, _ = c.Write([]byte("context is not map"))
		return gnet.Close
	}

	n, e := c.Read(h.buf[:])
	if e != nil {
		return gnet.Close
	}
	read := string(h.buf[:n])

	//_, _ = c.Write([]byte("you are " + cc["id"].(string)))
	topic := fmt.Sprintf("tunnel/%s/%s/up", cc["id"], h.ServerId)
	mqtt.Client.Publish(topic, 0, false, read)

	return gnet.None
}

func (h *TcpServer) OnTick() (delay time.Duration, action gnet.Action) {
	return 0, gnet.None
}
