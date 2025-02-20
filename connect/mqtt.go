package connect

import (
	"github.com/busy-cloud/boat/mqtt"
	"io"
	"strings"
)

func subscribe() {

	//订阅数据变化
	mqtt.Subscribe("link/+/down", func(topic string, payload []byte) {
		ss := strings.Split(topic, "/")
		conn, ok := linkers.Load(ss[1])
		if !ok {
			return
		}
		c, ok := conn.(io.ReadWriteCloser)
		if !ok {
			return
		}
		_, _ = c.Write(payload)
	})

	//订阅数据变化，服务类型
	mqtt.Subscribe("link/+/+/down", func(topic string, payload []byte) {
		ss := strings.Split(topic, "/")
		conn, ok := incomingConnections.Load(ss[2])
		if !ok {
			return
		}
		c, ok := conn.(io.ReadWriteCloser)
		if !ok {
			return
		}
		_, _ = c.Write(payload)
	})
}
