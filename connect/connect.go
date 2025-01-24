package connect

import (
	"github.com/busy-cloud/boat/log"
	"github.com/busy-cloud/boat/mqtt"
	paho "github.com/eclipse/paho.mqtt.golang"
	"github.com/panjf2000/gnet/v2"
	"io"
	"strings"
	"sync"
	"time"
)

var connections sync.Map //[string, io.ReadWriteCloser]

func Startup() error {

	//订阅数据变化
	mqtt.Client.SubscribeMultiple(map[string]byte{
		"tunnel/+/+/down": 0,
		"tunnel/+/down":   0,
	}, func(client paho.Client, message paho.Message) {
		ss := strings.Split(message.Topic(), "/")
		conn, ok := connections.Load(ss[1])
		if !ok {
			return
		}
		c, ok := conn.(io.ReadWriteCloser)
		if !ok {
			return
		}
		_, e := c.Write(message.Payload())
		if e != nil {
			connections.Delete(ss[1])
		}
	})

	var h = UdpServer{
		ServerId: "test",
		IdStart:  0,
		IdEnd:    4,
	}
	go func() {
		err := gnet.Run(&h, "udp://:60000",
			gnet.WithMulticore(true),
			gnet.WithLockOSThread(true),
			gnet.WithTCPKeepAlive(30*time.Second),
			gnet.WithTCPNoDelay(gnet.TCPDelay),
			//gnet.WithTicker(true),
		)
		if err != nil {
			log.Fatal(err)
		}
	}()
	return nil
}
