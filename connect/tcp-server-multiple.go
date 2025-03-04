package connect

import (
	"encoding/json"
	"fmt"
	"github.com/busy-cloud/boat/db"
	"github.com/busy-cloud/boat/mqtt"
	"github.com/busy-cloud/connector/types"
	"go.uber.org/multierr"
	"net"
	"regexp"
)

type TcpServerMultiple struct {
	*types.Linker

	buf    [4096]byte
	opened bool

	listener net.Listener
	children map[string]net.Conn

	regex *regexp.Regexp
}

func NewTcpServerMultiple(l *types.Linker) *TcpServerMultiple {
	server := &TcpServerMultiple{
		Linker:   l,
		children: make(map[string]net.Conn),
	}
	if server.RegisterOptions != nil && server.RegisterOptions.Regex != "" {
		server.regex, _ = regexp.Compile("^" + server.RegisterOptions.Regex + "$")
	}
	if server.regex == nil {
		server.regex = idReg
	}
	return server
}

func (s *TcpServerMultiple) Opened() bool {
	return s.opened
}

func (s *TcpServerMultiple) Connected() bool {
	return s.listener != nil
}

func (s *TcpServerMultiple) Open() (err error) {
	if s.opened {
		_ = s.Close()
	}

	//addr := fmt.Sprintf("%s:%d", s.Address, s.Port)
	addr := fmt.Sprintf("%s:%d", "", s.Port)
	s.listener, err = net.Listen("tcp", addr)
	if err != nil {
		return
	}

	s.opened = true

	go s.accept()

	topic := fmt.Sprintf("link/%s/open", s.Id)
	mqtt.Publish(topic, nil)

	return
}

func (s *TcpServerMultiple) Close() error {
	s.opened = false
	var err error
	for _, conn := range s.children {
		err = multierr.Append(err, conn.Close())
	}
	s.children = make(map[string]net.Conn)
	if s.listener != nil {
		err = multierr.Append(err, s.listener.Close())
		s.listener = nil
	}
	return err
}

func (s *TcpServerMultiple) receive(id string, reg []byte, conn net.Conn) {
	//从数据库中查询
	var i types.Incoming
	//xorm.ErrNotExist //db.Engine.Exist()
	has, err := db.Engine.ID(id).Get(&i)
	if err != nil {
		_, _ = conn.Write([]byte(err.Error()))
		_ = conn.Close()
		return
	}
	//查不到
	if !has {
		i.Id = id
		i.ServerId = s.Id
		i.Protocol = s.Protocol //继承协议
		_, err = db.Engine.InsertOne(&i)
		if err != nil {
			_, _ = conn.Write([]byte(err.Error()))
			_ = conn.Close()
			return
		}
	}

	s.children[id] = conn
	incomingConnections.Store(id, conn)

	//连接
	topicOpen := fmt.Sprintf("link/%s/%s/open", s.Id, id)
	mqtt.Publish(topicOpen, reg)
	if i.Protocol != "" {
		topicOpen = fmt.Sprintf("%s/%s/%s/open", i.Protocol, s.Id, id)
		mqtt.Publish(topicOpen, reg)
	}

	topicUp := fmt.Sprintf("link/%s/%s/up", s.Id, id)
	topicUpProtocol := fmt.Sprintf("%s/%s/up", s.Protocol, s.Id)

	var n int
	var e error
	buf := make([]byte, 4096)
	for {
		n, e = conn.Read(buf)
		if e != nil {
			_ = conn.Close()
			conn = nil
			delete(s.children, id)
			break
		}

		data := s.buf[:n]
		//转发
		mqtt.Publish(topicUp, data)
		if s.Protocol != "" {
			mqtt.Publish(topicUpProtocol, data)
		}
	}

	//下线
	topicClose := fmt.Sprintf("link/%s/%s/close", s.Id, id)
	mqtt.Publish(topicClose, e.Error())
	if s.Protocol != "" {
		topic := fmt.Sprintf("%s/%s/%s/close", s.Protocol, s.Id, id)
		mqtt.Publish(topic, e.Error())
	}

	incomingConnections.Delete(id)
}

func (s *TcpServerMultiple) accept() {
	for s.opened {
		conn, err := s.listener.Accept()
		if err != nil {
			break
		}

		//TODO 读超时
		n, e := conn.Read(s.buf[:])
		if e != nil {
			//log.Error(e)
			_ = conn.Close()
			continue
		}
		data := s.buf[:n]

		if s.RegisterOptions != nil {
			//去头
			if s.RegisterOptions.Offset > 0 {
				if int(s.RegisterOptions.Offset) > len(data) {
					_, _ = conn.Write([]byte("id too small"))
					_ = conn.Close()
					continue
				}
				data = data[s.RegisterOptions.Offset:]
			}
			//取定长
			if s.RegisterOptions.Length > 0 {
				if int(s.RegisterOptions.Length) > len(data) {
					_, _ = conn.Write([]byte("id too small"))
					_ = conn.Close()
					continue
				}
				data = data[:s.RegisterOptions.Length]
			}
		}

		id := string(data)

		//处理json包
		if s.RegisterOptions != nil && s.RegisterOptions.Type == "json" {
			var reg map[string]any
			err = json.Unmarshal(data, &reg)
			if err != nil {
				_, _ = conn.Write([]byte(err.Error()))
				_ = conn.Close()
				continue
			}

			var ok bool
			id, ok = reg[s.RegisterOptions.Field].(string)
			if !ok {
				_, _ = conn.Write([]byte("require field " + s.RegisterOptions.Field))
				_ = conn.Close()
				continue
			}
		}

		//验证合法性
		if !s.regex.MatchString(id) {
			_, _ = conn.Write([]byte("invalid id"))
			_ = conn.Close()
			continue
		}

		//接口数据
		go s.receive(id, s.buf[:n], conn)
	}

	_ = s.listener.Close()
	s.listener = nil

	//下线
	topicClose := fmt.Sprintf("link/%s/close", s.Id)
	mqtt.Publish(topicClose, "")
}
