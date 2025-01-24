package connect

import (
	"fmt"
	"github.com/busy-cloud/boat/log"
	"github.com/busy-cloud/boat/mqtt"
	"net"
	"time"
)

type Client struct {
	*Connect

	net.Conn
	buf    [4096]byte
	opened bool
}

func NewClient(l *Connect) *Client {
	c := &Client{Connect: l}
	go c.keep()
	return c
}

func (c *Client) Opened() bool {
	return c.opened
}

func (c *Client) Connected() bool {
	return c.Conn != nil
}

func (c *Client) Open() (err error) {
	if !c.opened {
		c.opened = true
		go c.keep()
	}
	c.Conn, err = net.Dial(c.Type, c.Addr)
	if err != nil {
		return err
	}
	go c.receive()
	return
}

func (c *Client) Close() (err error) {
	c.opened = true
	if c.Conn != nil {
		return c.Conn.Close()
	}
	return nil
}

func (c *Client) keep() {
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

func (c *Client) receive() {
	topicOpen := fmt.Sprintf("link/%s/opened", c.Id)
	topicUp := fmt.Sprintf("link/%s/up", c.Id)
	topicClose := fmt.Sprintf("link/%s/close", c.Id)

	//连接
	mqtt.Client.Publish(topicOpen, 0, false, c.Conn.RemoteAddr().String())

	connections.Store(c.Id, c)

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
		//mqtt.Client.IsConnected()
		//转发
		mqtt.Client.Publish(topicUp, 0, false, data)
	}

	//下线
	mqtt.Client.Publish(topicClose, 0, false, e.Error())

	connections.Delete(c.Id)
}
