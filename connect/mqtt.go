package connect

import (
	"github.com/busy-cloud/boat/mqtt"
	paho "github.com/eclipse/paho.mqtt.golang"
	"io"
	"strings"
)

func subscribe() {
	//订阅数据变化
	mqtt.Client.SubscribeMultiple(map[string]byte{
		"link/+/+/down": 0,
		"link/+/down":   0,
	}, func(client paho.Client, message paho.Message) {
		ss := strings.Split(message.Topic(), "/")
		conn, ok := links.Load(ss[1])
		if !ok {
			return
		}
		c, ok := conn.(io.ReadWriteCloser)
		if !ok {
			return
		}
		_, e := c.Write(message.Payload())
		if e != nil {
			links.Delete(ss[1])
		}
	})
}
