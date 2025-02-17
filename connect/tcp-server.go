package connect

import (
	"fmt"
	"github.com/busy-cloud/boat/mqtt"
	"github.com/busy-cloud/connector/types"
	"go.uber.org/multierr"
	"net"
)

type TcpServer struct {
	*types.Linker

	net.Conn
	buf    [4096]byte
	opened bool

	listener net.Listener
}

func NewTcpServer(l *types.Linker) *TcpServer {
	c := &TcpServer{Linker: l}
	return c
}

func (c *TcpServer) Opened() bool {
	return c.opened
}

func (c *TcpServer) Connected() bool {
	return c.Conn != nil
}

func (c *TcpServer) Open() (err error) {
	if c.opened {
		//重复打开关闭上次连接
		if c.listener != nil {
			_ = c.listener.Close()
		}
		if c.Conn != nil {
			_ = c.Conn.Close()
		}
	}

	//addr := fmt.Sprintf("%s:%d", c.Address, c.Port)
	addr := fmt.Sprintf("%s:%d", "", c.Port)
	c.listener, err = net.Listen("tcp", addr)
	if err != nil {
		return
	}

	c.opened = true
	go c.receive()

	return
}

func (c *TcpServer) Close() error {
	c.opened = false
	var err error
	if c.Conn != nil {
		err = multierr.Append(err, c.Conn.Close())
	}
	if c.listener != nil {
		err = multierr.Append(err, c.listener.Close())
	}
	return err
}

func (c *TcpServer) receive() {
	topicOpen := fmt.Sprintf("link/%s/open", c.Id)
	topicUp := fmt.Sprintf("link/%s/up", c.Id)
	topicClose := fmt.Sprintf("link/%s/close", c.Id)

	var err error
	for c.opened {
		c.Conn, err = c.listener.Accept()
		if err != nil {
			break
		}

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
			//mqtt.TcpServer.IsConnected()
			//转发
			mqtt.Client.Publish(topicUp, 0, false, data)
		}

		//下线
		mqtt.Client.Publish(topicClose, 0, false, e.Error())

		c.Conn = nil
	}

	_ = c.listener.Close()
	c.listener = nil
}
