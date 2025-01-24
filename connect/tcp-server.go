package connect

import (
	"fmt"
	"github.com/busy-cloud/boat/mqtt"
	"github.com/panjf2000/gnet/v2"
	"regexp"
	"sync/atomic"
	"time"
)

var idReg = regexp.MustCompile(`^\w{2,128}`)

type TcpServer struct {
	*Connect
	*Server

	buf   [4096]byte
	count atomic.Int64

	regex *regexp.Regexp
}

func (h *TcpServer) OnBoot(eng gnet.Engine) (action gnet.Action) {
	h.Server.opened = true
	h.Server.engine = eng
	return gnet.None
}

func (h *TcpServer) OnShutdown(eng gnet.Engine) {
	h.Server.opened = true
}

func (h *TcpServer) OnOpen(c gnet.Conn) (out []byte, action gnet.Action) {
	h.count.Add(1)
	h.connected = true
	return nil, gnet.None
}

func (h *TcpServer) OnClose(c gnet.Conn, err error) (action gnet.Action) {
	n := h.count.Add(-1)
	if n <= 0 {
		h.Server.connected = false
	}

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
	topic := fmt.Sprintf("link/%s/%s/close", id, h.Id)
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

		if h.regex == nil {
			if h.IdOptions != nil && h.IdOptions.Regex != "" {
				h.regex, e = regexp.Compile(h.IdOptions.Regex) //重复编译
				if e != nil {
					//_, _ = c.Write([]byte("id regex compile error"))
					//return gnet.Close
				} else {
					h.regex = idReg
				}
			} else {
				h.regex = idReg
			}
		}

		//验证合法性
		if !h.regex.MatchString(id) {
			_, _ = c.Write([]byte("invalid id"))
			return gnet.Close
		}

		ctx = map[string]interface{}{"id": id}
		c.SetContext(ctx)

		//上线
		topic := fmt.Sprintf("link/%s/%s/opened", id, h.Id)
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
	topic := fmt.Sprintf("link/%s/%s/up", cc["id"], h.Id)
	mqtt.Client.Publish(topic, 0, false, read)

	return gnet.None
}

func (h *TcpServer) OnTick() (delay time.Duration, action gnet.Action) {
	return 0, gnet.None
}
