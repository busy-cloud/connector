package connect

import (
	"fmt"
	"github.com/busy-cloud/boat/db"
	"github.com/busy-cloud/boat/mqtt"
	"github.com/busy-cloud/connector/types"
	"github.com/panjf2000/gnet/v2"
	"regexp"
	"sync/atomic"
	"time"
)

var idReg = regexp.MustCompile(`^\w{2,128}$`)

type GNetHandler struct {
	*types.Linker
	*GNetServer

	buf   [4096]byte
	count atomic.Int64

	regex *regexp.Regexp
}

func NewGNetHandlerTcp(link *types.Linker, server *GNetServer) *GNetHandler {
	h := &GNetHandler{
		Linker:     link,
		GNetServer: server,
	}
	if link.IdRegex != "" {
		h.regex, _ = regexp.Compile("^" + link.IdRegex + "$")
		if h.regex == nil {
			h.regex = idReg
		}
	}
	return h
}

func (h *GNetHandler) OnBoot(eng gnet.Engine) (action gnet.Action) {
	h.GNetServer.opened = true
	h.GNetServer.engine = eng
	return gnet.None
}

func (h *GNetHandler) OnShutdown(eng gnet.Engine) {
	h.GNetServer.opened = false
}

func (h *GNetHandler) OnOpen(c gnet.Conn) (out []byte, action gnet.Action) {
	h.count.Add(1)
	h.connected = true
	return nil, gnet.None
}

func (h *GNetHandler) OnClose(c gnet.Conn, err error) (action gnet.Action) {
	n := h.count.Add(-1)
	//可以使用h.engine.CountConnections()替代，就不知道效率怎么样
	if n <= 0 {
		h.GNetServer.connected = false
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

func (h *GNetHandler) OnTraffic(c gnet.Conn) (action gnet.Action) {
	ctx := c.Context()

	//检查首个包, 作为注册包
	if ctx == nil {
		n, e := c.Read(h.buf[:])
		if e != nil {
			return gnet.Close
		}
		id := string(h.buf[:n])

		//验证合法性
		if !h.regex.MatchString(id) {
			_, _ = c.Write([]byte("invalid id"))
			return gnet.Close
		}

		//从数据库中查询
		var i types.Incoming
		//xorm.ErrNotExist //db.Engine.Exist()
		has, err := db.Engine.ID(id).Get(&i)
		if err != nil {
			_, _ = c.Write([]byte(err.Error()))
			return gnet.Close
		}
		//查不到
		if !has {
			i.Id = id
			i.ServerId = h.Id
			_, err = db.Engine.InsertOne(&i)
			if err != nil {
				_, _ = c.Write([]byte(err.Error()))
				return gnet.Close
			}
		}
		incoming := Incoming{Incoming: &i, conn: c}

		ctx = map[string]interface{}{"id": id}
		c.SetContext(ctx)

		//上线
		topic := fmt.Sprintf("link/%s/%s/open", id, h.Id)
		mqtt.Client.Publish(topic, 0, false, c.RemoteAddr().String())

		//保存连接
		connections.Store(id, incoming)

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
	read := h.buf[:n] //TODO 可能要复制read

	//_, _ = c.Write([]byte("you are " + cc["id"].(string)))
	topic := fmt.Sprintf("link/%s/%s/up", cc["id"], h.Id)
	mqtt.Client.Publish(topic, 0, false, read)

	return gnet.None
}

func (h *GNetHandler) OnTick() (delay time.Duration, action gnet.Action) {
	return 0, gnet.None
}
