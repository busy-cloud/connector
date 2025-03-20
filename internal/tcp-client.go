package internal

import (
	"errors"
	"fmt"
	"github.com/busy-cloud/boat/log"
	"github.com/busy-cloud/boat/mqtt"
	"net"
	"time"
)

type TcpClient struct {
	*Linker

	net.Conn
	buf    [4096]byte
	opened bool
}

func NewTcpClient(l *Linker) *TcpClient {
	c := &TcpClient{Linker: l}
	return c
}

func (c *TcpClient) Opened() bool {
	return c.opened
}

func (c *TcpClient) Connected() bool {
	return c.Conn != nil
}

func (c *TcpClient) Error() string {
	return c.Linker.Error
}

func (c *TcpClient) connect() (err error) {
	defer func() {
		if err != nil {
			c.Linker.Error = err.Error()
		} else {
			c.Linker.Error = ""
		}
	}()

	if c.Conn != nil {
		_ = c.Conn.Close()
	}

	//连接
	addr := fmt.Sprintf("%s:%d", c.Address, c.Port)
	c.Conn, err = net.Dial("tcp", addr)
	if err != nil {
		return err
	}

	go c.keep()

	return
}

func (c *TcpClient) Open() (err error) {
	if c.opened {
		return errors.New("already opened")
	}
	c.opened = true

	go c.keep()

	return
}

func (c *TcpClient) Close() (err error) {
	c.opened = false

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
	links.Store(c.Id, c)

	//连接
	topicOpen := fmt.Sprintf("link/%s/open", c.Id)
	mqtt.Publish(topicOpen, c.Conn.RemoteAddr().String())
	if c.Protocol != "" {
		topic := fmt.Sprintf("%s/%s/open", c.Protocol, c.Id)
		mqtt.Publish(topic, c.ProtocolOptions)
	}

	topicUp := fmt.Sprintf("link/%s/up", c.Id)
	topicUpProtocol := fmt.Sprintf("%s/%s/up", c.Protocol, c.Id)
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
		mqtt.Publish(topicUp, data)
		if c.Protocol != "" {
			mqtt.Publish(topicUpProtocol, data)
		}
	}

	//下线
	topicClose := fmt.Sprintf("link/%s/close", c.Id)
	mqtt.Publish(topicClose, e.Error())
	if c.Protocol != "" {
		topic := fmt.Sprintf("%s/%s/close", c.Protocol, c.Id)
		mqtt.Publish(topic, c.SerialOptions.PortName)
	}

	links.Delete(c.Id)
}
