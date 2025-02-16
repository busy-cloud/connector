package connect

import (
	"fmt"
	"github.com/busy-cloud/boat/log"
	"github.com/busy-cloud/boat/mqtt"
	"github.com/busy-cloud/connector/types"
	"net"
	"time"
)

type TcpClient struct {
	*types.Linker

	net.Conn
	buf    [4096]byte
	opened bool
}

func NewTcpClient(l *types.Linker) *TcpClient {
	c := &TcpClient{Linker: l}
	return c
}

func (c *TcpClient) Opened() bool {
	return c.opened
}

func (c *TcpClient) Connected() bool {
	return c.Conn != nil
}

func (c *TcpClient) Open() (err error) {
	if c.opened {
		//重复打开关闭上次连接
		if c.Conn != nil {
			_ = c.Conn.Close()
		}
	}

	//连接
	addr := fmt.Sprintf("%s:%d", c.Address, c.Port)
	c.Conn, err = net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	c.opened = true
	go c.keep()
	go c.receive()
	return
}

func (c *TcpClient) Close() (err error) {
	c.opened = true
	if c.Conn != nil {
		return c.Conn.Close()
	}
	return nil
}

func (c *TcpClient) keep() {
	for c.opened {
		time.Sleep(time.Minute)

		if c.Conn != nil {
			continue
		}

		err := c.Open()
		if err != nil {
			log.Error(err)
		}
	}
}

func (c *TcpClient) receive() {
	topicOpen := fmt.Sprintf("link/%s/open", c.Id)
	topicUp := fmt.Sprintf("link/%s/up", c.Id)
	topicClose := fmt.Sprintf("link/%s/close", c.Id)

	//连接
	mqtt.Client.Publish(topicOpen, 0, false, c.Conn.RemoteAddr().String())

	var n int
	var e error
	for {
		n, e = c.Conn.Read(c.buf[:])
		if e != nil {
			_ = c.Conn.Close()
			c.Conn = nil
			break
		}
		data := c.buf[:n]
		//mqtt.TcpClient.IsConnected()
		//转发
		mqtt.Client.Publish(topicUp, 0, false, data)
	}

	//下线
	mqtt.Client.Publish(topicClose, 0, false, e.Error())
}
