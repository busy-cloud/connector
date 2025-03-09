package connect

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/busy-cloud/boat/db"
	"github.com/busy-cloud/boat/log"
	"github.com/busy-cloud/boat/mqtt"
	"github.com/busy-cloud/connector/types"
	"github.com/panjf2000/gnet/v2"
	"regexp"
	"sync/atomic"
	"time"
)

var idReg = regexp.MustCompile(`^\w{2,128}$`)

type GNetServer struct {
	*types.Linker

	engine gnet.Engine //在Handler的OnBoot中复制

	opened    bool
	connected bool

	buf   [4096]byte
	count atomic.Int64

	regex *regexp.Regexp
}

func NewGNetServer(l *types.Linker) *GNetServer {
	server := &GNetServer{Linker: l}
	if server.RegisterOptions != nil && server.RegisterOptions.Regex != "" {
		server.regex, _ = regexp.Compile("^" + server.RegisterOptions.Regex + "$")
	}
	if server.regex == nil {
		server.regex = idReg
	}
	return server
}

func (s *GNetServer) Opened() bool {
	return s.opened
}

func (s *GNetServer) Connected() bool {
	return s.connected
}

func (s *GNetServer) Open() error {
	//handler := &GNetServer{Linker: s.Linker, GNetServer: s}
	addr := fmt.Sprintf("tcp://:%d", s.Port)
	log.Println("GNet Server Opening: ", addr)

	go func() {
		//这里全阻塞等待
		err := gnet.Run(s, addr,
			gnet.WithMulticore(true),
			//gnet.WithLockOSThread(true), //依赖CGO，容易编译出错
			gnet.WithTCPKeepAlive(30*time.Second),
			gnet.WithTCPNoDelay(gnet.TCPDelay),
			//gnet.WithTicker(true), //严重占用CPU
			//gnet.WithLogger()
		)
		if err != nil {
			log.Error(err)
		}
	}()

	return nil
}

func (s *GNetServer) Close() error {
	s.connected = false
	s.opened = false
	return s.engine.Stop(context.Background())
}

func (s *GNetServer) OnBoot(eng gnet.Engine) (action gnet.Action) {
	s.opened = true
	s.engine = eng
	return gnet.None
}

func (s *GNetServer) OnShutdown(eng gnet.Engine) {
	s.opened = false
}

func (s *GNetServer) OnOpen(conn gnet.Conn) (out []byte, action gnet.Action) {
	s.count.Add(1)
	s.connected = true
	return nil, gnet.None
}

func (s *GNetServer) OnClose(conn gnet.Conn, err error) (action gnet.Action) {
	n := s.count.Add(-1)
	//可以使用h.engine.CountConnections()替代，就不知道效率怎么样
	if n <= 0 {
		s.connected = false
	}

	ctx := conn.Context()
	if ctx == nil {
		return gnet.None
	}

	cc, ok := ctx.(map[string]interface{})
	if !ok {
		return gnet.None
	}
	id := cc["id"]

	last, ok := incomingConnections.Load(id)
	if !ok {
		return gnet.None
	}

	//同一连接才算关闭，应对移动网络抖动问题，新连接发起后，旧连接才关闭
	if last != nil && last.(gnet.Conn) == conn {
		//下线
		topic := fmt.Sprintf("link/%s/%s/close", s.Id, id)
		mqtt.Publish(topic, err.Error())
		if p, ok := cc["protocol"]; ok {
			//向协议转发
			topic := fmt.Sprintf("%s/%s/%s/close", p, s.Id, id)
			mqtt.Publish(topic, err.Error())
		}

		//从池中清除
		incomingConnections.Delete(id)
	}

	return gnet.None
}

func (s *GNetServer) OnTraffic(conn gnet.Conn) (action gnet.Action) {
	ctx := conn.Context()

	//检查首个包, 作为注册包
	if ctx == nil {
		n, e := conn.Read(s.buf[:])
		if e != nil {
			return gnet.Close
		}

		data := s.buf[:n]

		if s.RegisterOptions != nil {
			//去头
			if s.RegisterOptions.Offset > 0 {
				if int(s.RegisterOptions.Offset) > len(data) {
					_, _ = conn.Write([]byte("id too small"))
					return gnet.Close
				}
				data = data[s.RegisterOptions.Offset:]
			}
			//取定长
			if s.RegisterOptions.Length > 0 {
				if int(s.RegisterOptions.Length) > len(data) {
					_, _ = conn.Write([]byte("id too small"))
					return gnet.Close
				}
				data = data[:s.RegisterOptions.Length]
			}
		}

		id := string(data)

		//处理json包
		if s.RegisterOptions != nil && s.RegisterOptions.Type == "json" {
			var reg map[string]any
			err := json.Unmarshal(data, &reg)
			if err != nil {
				_, _ = conn.Write([]byte(err.Error()))
				return gnet.Close
			}

			var ok bool
			id, ok = reg[s.RegisterOptions.Field].(string)
			if !ok {
				_, _ = conn.Write([]byte("require field " + s.RegisterOptions.Field))
				return gnet.Close
			}
		}

		//验证合法性
		if !s.regex.MatchString(id) {
			_, _ = conn.Write([]byte("invalid id"))
			return gnet.Close
		}

		//从数据库中查询
		var i types.Incoming
		//xorm.ErrNotExist //db.Engine.Exist()
		has, err := db.Engine().ID(id).Get(&i)
		if err != nil {
			_, _ = conn.Write([]byte(err.Error()))
			return gnet.Close
		}
		//查不到
		if !has {
			i.Id = id
			i.ServerId = s.Id
			i.Protocol = s.Protocol //继承协议
			_, err = db.Engine().InsertOne(&i)
			if err != nil {
				_, _ = conn.Write([]byte(err.Error()))
				return gnet.Close
			}
		}
		//incoming := Incoming{Incoming: &i, conn: conn}

		c := map[string]interface{}{"id": id}
		conn.SetContext(c)

		//上线
		topic := fmt.Sprintf("link/%s/%s/open", s.Id, id)
		mqtt.Publish(topic, s.buf[:n])
		if i.Protocol != "" {
			c["protocol"] = i.Protocol //协议也保存进去
			topic = fmt.Sprintf("%s/%s/%s/open", i.Protocol, s.Id, id)
			mqtt.Publish(topic, s.buf[:n])
		}

		//保存连接
		incomingConnections.Store(id, conn)

		return gnet.None
	}

	//取出上下文
	c, ok := ctx.(map[string]interface{})
	if !ok {
		_, _ = conn.Write([]byte("context is not map"))
		return gnet.Close
	}
	id := c["id"]

	n, e := conn.Read(s.buf[:])
	if e != nil {
		return gnet.Close
	}
	read := s.buf[:n] //TODO 可能要复制read

	topic := fmt.Sprintf("link/%s/%s/up", s.Id, id)
	mqtt.Publish(topic, read)
	if p, ok := c["protocol"]; ok {
		//向协议转发
		topic := fmt.Sprintf("%s/%s/%s/up", p, s.Id, id)
		mqtt.Publish(topic, read)
	}

	return gnet.None
}

func (s *GNetServer) OnTick() (delay time.Duration, action gnet.Action) {
	return 0, gnet.None
}
