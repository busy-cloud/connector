package connect

import (
	"fmt"
	"github.com/busy-cloud/boat/mqtt"
	"net"
)

type UdpClient struct {
	*Connect

	net.Conn
	buf    [4096]byte
	opened bool
}

func NewUdpClient(l *Connect) *UdpClient {
	c := &UdpClient{Connect: l}
	return c
}

func (c *UdpClient) Opened() bool {
	return c.opened
}

func (c *UdpClient) Connected() bool {
	return c.Conn != nil
}

func (c *UdpClient) Open() (err error) {
	if !c.opened {
		c.opened = true
	}
	addr := fmt.Sprintf("%s:%d", c.Address, c.Port)
	c.Conn, err = net.Dial("udp", addr)
	if err != nil {
		return err
	}
	go c.receive()
	return
}

func (c *UdpClient) Close() (err error) {
	c.opened = true
	if c.Conn != nil {
		return c.Conn.Close()
	}
	return nil
}

func (c *UdpClient) receive() {
	topicOpen := fmt.Sprintf("link/%s/open", c.Id)
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
		//mqtt.UdpClient.IsConnected()
		//转发
		mqtt.Client.Publish(topicUp, 0, false, data)
	}

	//下线
	mqtt.Client.Publish(topicClose, 0, false, e.Error())

	connections.Delete(c.Id)
}
