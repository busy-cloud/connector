package connect

import (
	"context"
	"github.com/busy-cloud/boat/exception"
	"github.com/busy-cloud/boat/log"
	"github.com/panjf2000/gnet/v2"
	"time"
)

type GNetServer struct {
	*Connect

	engine gnet.Engine //在Handler的OnBoot中复制

	opened    bool
	connected bool
}

func NewGNetServer(l *Connect) *GNetServer {
	return &GNetServer{Connect: l}
}

func (s *GNetServer) Opened() bool {
	return s.opened
}

func (s *GNetServer) Connected() bool {
	return s.connected
}

func (s *GNetServer) Open() error {
	var h gnet.EventHandler
	switch s.Type {
	case "tcp":
		if s.Singleton {
			h = &GNetHandlerTcpSingleton{Connect: s.Connect, GNetServer: s}
		} else {
			h = &GNetHandlerTcp{Connect: s.Connect, GNetServer: s}
		}
	case "udp":
		h = &GNetHandlerUdp{Connect: s.Connect, GNetServer: s}
	default:
		return exception.New("错误类型")
	}

	protoAddr := s.Type + "://" + s.Addr
	go func() {
		//这里全阻塞等待
		err := gnet.Run(h, protoAddr,
			gnet.WithMulticore(true),
			gnet.WithLockOSThread(true),
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
