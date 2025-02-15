package connect

import (
	"context"
	"fmt"
	"github.com/busy-cloud/boat/log"
	"github.com/panjf2000/gnet/v2"
	"time"
)

type GNetServer struct {
	*Linker

	engine gnet.Engine //在Handler的OnBoot中复制

	opened    bool
	connected bool
}

func NewGNetServer(l *Linker) *GNetServer {
	return &GNetServer{Linker: l}
}

func (s *GNetServer) Opened() bool {
	return s.opened
}

func (s *GNetServer) Connected() bool {
	return s.connected
}

func (s *GNetServer) Open() error {
	handler := &GNetHandler{Linker: s.Linker, GNetServer: s}
	addr := fmt.Sprintf("tcp://:%d", s.Port)
	go func() {
		//这里全阻塞等待
		err := gnet.Run(handler, addr,
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
