package internal

import (
	"fmt"
	"github.com/busy-cloud/boat/mqtt"
	"go.uber.org/multierr"
	"net"
)

type TcpServer struct {
	*Linker

	net.Conn
	buf    [4096]byte
	opened bool

	listener net.Listener
}

func NewTcpServer(l *Linker) *TcpServer {
	c := &TcpServer{Linker: l}
	return c
}

func (s *TcpServer) Opened() bool {
	return s.opened
}

func (s *TcpServer) Connected() bool {
	return s.Conn != nil
}

func (s *TcpServer) Open() (err error) {
	if s.opened {
		//重复打开关闭上次连接
		if s.listener != nil {
			_ = s.listener.Close()
		}
		if s.Conn != nil {
			_ = s.Conn.Close()
		}
	}

	//addr := fmt.Sprintf("%s:%d", s.Address, s.Port)
	addr := fmt.Sprintf("%s:%d", "", s.Port)
	s.listener, err = net.Listen("tcp", addr)
	if err != nil {
		return
	}

	s.opened = true
	go s.receive()

	return
}

func (s *TcpServer) Close() error {
	s.opened = false
	var err error
	if s.Conn != nil {
		err = multierr.Append(err, s.Conn.Close())
	}
	if s.listener != nil {
		err = multierr.Append(err, s.listener.Close())
	}
	return err
}

func (s *TcpServer) receive() {
	links.Store(s.Id, s)

	var err error
	for s.opened {
		s.Conn, err = s.listener.Accept()
		if err != nil {
			break
		}

		//连接
		topicOpen := fmt.Sprintf("link/%s/open", s.Id)
		mqtt.Publish(topicOpen, s.Conn.RemoteAddr().String())
		if s.Protocol != "" {
			topic := fmt.Sprintf("%s/%s/open", s.Protocol, s.Id)
			mqtt.Publish(topic, s.Conn.RemoteAddr().String())
		}

		topicUp := fmt.Sprintf("link/%s/up", s.Id)
		topicUpProtocol := fmt.Sprintf("%s/%s/up", s.Protocol, s.Id)

		var n int
		var e error
		for {
			n, e = s.Conn.Read(s.buf[:])
			if e != nil {
				_ = s.Conn.Close()
				s.Conn = nil
				break
			}
			data := s.buf[:n]
			//mqtt.TcpServer.IsConnected()
			//转发
			mqtt.Publish(topicUp, data)
			if s.Protocol != "" {
				mqtt.Publish(topicUpProtocol, data)
			}
		}

		//下线
		topicClose := fmt.Sprintf("link/%s/close", s.Id)
		mqtt.Publish(topicClose, e.Error())
		if s.Protocol != "" {
			topic := fmt.Sprintf("%s/%s/close", s.Protocol, s.Id)
			mqtt.Publish(topic, s.SerialOptions.PortName)
		}

		s.Conn = nil
	}

	_ = s.listener.Close()
	s.listener = nil

	links.Delete(s.Id)
}
