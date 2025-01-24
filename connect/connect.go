package connect

import (
	"github.com/panjf2000/gnet/v2"
	"time"
)

func Startup() error {
	var h handler
	err := gnet.Run(&h, "tcp://:60000",
		gnet.WithMulticore(true),
		gnet.WithLockOSThread(true),
		gnet.WithTCPKeepAlive(30*time.Second),
		gnet.WithTCPNoDelay(gnet.TCPDelay),
		gnet.WithTicker(true),
	)
	return err
}
