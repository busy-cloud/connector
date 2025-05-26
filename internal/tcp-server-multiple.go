package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/busy-cloud/boat/db"
	"github.com/busy-cloud/boat/mqtt"
	"go.uber.org/multierr"
	"net"
	"regexp"
)

type TcpServerMultiple struct {
	*Linker

	buf    [4096]byte
	opened bool

	listener net.Listener
	children map[string]*TcpIncoming

	regex *regexp.Regexp
}

func NewTcpServerMultiple(l *Linker) *TcpServerMultiple {
	server := &TcpServerMultiple{
		Linker:   l,
		children: make(map[string]*TcpIncoming),
	}
	if server.RegisterOptions != nil && server.RegisterOptions.Regex != "" {
		server.regex, _ = regexp.Compile("^" + server.RegisterOptions.Regex + "$")
	}
	if server.regex == nil {
		server.regex = idReg
	}
	return server
}

func (s *TcpServerMultiple) Read(p []byte) (n int, err error) {
	return 0, errors.New("unsupported read")
}

func (s *TcpServerMultiple) Write(p []byte) (n int, err error) {
	return 0, errors.New("unsupported write")
}

func (s *TcpServerMultiple) Opened() bool {
	return s.opened
}

func (s *TcpServerMultiple) Connected() bool {
	return s.listener != nil
}

func (s *TcpServerMultiple) Error() string {
	return s.Linker.Error
}

func (s *TcpServerMultiple) Open() (err error) {
	defer func() {
		if err != nil {
			s.Linker.Error = err.Error()
		} else {
			s.Linker.Error = ""
		}
	}()

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
	s.children = make(map[string]*TcpIncoming)
	if s.listener != nil {
		err = multierr.Append(err, s.listener.Close())
		s.listener = nil
	}
	return err
}

func (s *TcpServerMultiple) receive(id string, reg []byte, conn net.Conn) {
	//从数据库中查询
	var incoming TcpIncoming
	//xorm.ErrNotExist //db.Engine.Exist()
	has, err := db.Engine().ID(id).Get(&incoming)
	if err != nil {
		_, _ = conn.Write([]byte(err.Error()))
		_ = conn.Close()
		return
	}
	//查不到
	if !has {
		incoming.Id = id
		incoming.LinkerId = s.Id
		incoming.Protocol = s.Protocol //继承协议
		incoming.ProtocolOptions = s.ProtocolOptions
		_, err = db.Engine().InsertOne(&incoming)
		if err != nil {
			_, _ = conn.Write([]byte(err.Error()))
			_ = conn.Close()
			return
		}
	} else {
		if incoming.Disabled {
			_, _ = conn.Write([]byte("disabled"))
			_ = conn.Close()
			return
		}
	}

	//赋值连接
	incoming.conn = conn

	s.children[id] = &incoming
	links.Store(id, &incoming)

	//连接
	topicOpen := fmt.Sprintf("link/%s/%s/open", s.Id, id)
	mqtt.Publish(topicOpen, reg)
	if incoming.Protocol != "" {
		topicOpen = fmt.Sprintf("%s/%s/%s/open", incoming.Protocol, s.Id, id)
		mqtt.Publish(topicOpen, incoming.ProtocolOptions)
	}

	topicUp := fmt.Sprintf("link/%s/%s/up", s.Id, id)
	topicUpProtocol := fmt.Sprintf("%s/%s/%s/up", s.Protocol, s.Id, id)

	var n int
	var e error
	buf := make([]byte, 4096)
	for {
		n, e = conn.Read(buf)
		if e != nil {
			_ = conn.Close()
			break
		}

		data := buf[:n]
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

	delete(s.children, id)
	links.Delete(id)
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

		//开始接收数据
		go s.receive(id, data, conn)
	}

	_ = s.listener.Close()
	s.listener = nil

	//下线
	topicClose := fmt.Sprintf("link/%s/close", s.Id)
	mqtt.Publish(topicClose, "")
}
