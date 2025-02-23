package connect

import (
	"context"
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
	if server.IdRegex != "" {
		server.regex, _ = regexp.Compile("^" + server.IdRegex + "$")
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

func (s *GNetServer) OnOpen(c gnet.Conn) (out []byte, action gnet.Action) {
	s.count.Add(1)
	s.connected = true
	return nil, gnet.None
}

func (s *GNetServer) OnClose(c gnet.Conn, err error) (action gnet.Action) {
	n := s.count.Add(-1)
	//可以使用h.engine.CountConnections()替代，就不知道效率怎么样
	if n <= 0 {
		s.connected = false
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
	topic := fmt.Sprintf("link/%s/%s/close", s.Id, id)
	mqtt.Client.Publish(topic, 0, false, err.Error())

	//从池中清除
	incomingConnections.Delete(id)

	return gnet.None
}

func (s *GNetServer) OnTraffic(c gnet.Conn) (action gnet.Action) {
	ctx := c.Context()

	//检查首个包, 作为注册包
	if ctx == nil {
		n, e := c.Read(s.buf[:])
		if e != nil {
			return gnet.Close
		}
		id := string(s.buf[:n])

		//验证合法性
		if !s.regex.MatchString(id) {
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
			i.ServerId = s.Id
			_, err = db.Engine.InsertOne(&i)
			if err != nil {
				_, _ = c.Write([]byte(err.Error()))
				return gnet.Close
			}
		}
		//incoming := Incoming{Incoming: &i, conn: c}

		ctx = map[string]interface{}{"id": id}
		c.SetContext(ctx)

		//上线
		topic := fmt.Sprintf("link/%s/%s/open", s.Id, id)
		mqtt.Client.Publish(topic, 0, false, c.RemoteAddr().String())

		//保存连接
		incomingConnections.Store(id, c)

		return gnet.None
	}

	//取出上下文
	cc, ok := ctx.(map[string]interface{})
	if !ok {
		_, _ = c.Write([]byte("context is not map"))
		return gnet.Close
	}

	n, e := c.Read(s.buf[:])
	if e != nil {
		return gnet.Close
	}
	read := s.buf[:n] //TODO 可能要复制read

	//_, _ = c.Write([]byte("you are " + cc["id"].(string)))
	topic := fmt.Sprintf("link/%s/%s/up", s.Id, cc["id"])
	mqtt.Client.Publish(topic, 0, false, read)

	return gnet.None
}

func (s *GNetServer) OnTick() (delay time.Duration, action gnet.Action) {
	return 0, gnet.None
}
