package connect

import (
	"context"
	"github.com/busy-cloud/boat/exception"
	"github.com/busy-cloud/boat/log"
	"github.com/panjf2000/gnet/v2"
	"time"
)

type Server struct {
	*Connect

	engine gnet.Engine

	opened    bool
	connected bool
}

func NewServer(l *Connect) *Server {
	return &Server{Connect: l}
}

func (s *Server) Opened() bool {
	return s.opened
}

func (s *Server) Connected() bool {
	return s.connected
}

func (s *Server) Open() error {
	var h gnet.EventHandler
	switch s.Type {
	case "tcp":
		if s.Singleton {
			h = &TcpServerSingleton{Connect: s.Connect, Server: s}
		} else {
			h = &TcpServer{Connect: s.Connect, Server: s}
		}
	case "udp":
		h = &UdpServer{Connect: s.Connect, Server: s}
	default:
		return exception.New("错误类型")
	}

	protoAddr := s.Type + "://" + s.Addr
	go func() {
		err := gnet.Run(h, protoAddr,
			gnet.WithMulticore(true),
			gnet.WithLockOSThread(true),
			gnet.WithTCPKeepAlive(30*time.Second),
			gnet.WithTCPNoDelay(gnet.TCPDelay),
			//gnet.WithTicker(true), //严重占用CPU
		)
		if err != nil {
			log.Error(err)
		}
	}()

	return nil
}

func (s *Server) Close() error {
	s.connected = false
	s.opened = false
	return s.engine.Stop(context.Background())
}
