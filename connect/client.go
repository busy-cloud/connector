package connect

import (
	"fmt"
	"github.com/busy-cloud/boat/log"
	"github.com/busy-cloud/boat/mqtt"
	"net"
	"strconv"
	"time"
)

type Client struct {
	Id   string
	Type string //tcp, udp
	Ip   string
	Port int

	conn net.Conn
	buf  [4096]byte
}

func NewClient(conn net.Conn) *Client {
	return &Client{}
}

func (c *Client) Connect() (err error) {
	c.conn, err = net.Dial(c.Type, c.Ip+":"+strconv.Itoa(c.Port))
	go c.receive()
	go c.keep()
	return
}

func (c *Client) keep() {
	for {
		time.Sleep(time.Minute)
		err := c.Connect()
		if err != nil {
			log.Error(err)
		}
	}
}

func (c *Client) receive() {
	topicOpen := fmt.Sprintf("tunnel/%s/open", c.Id)
	topicUp := fmt.Sprintf("tunnel/%s/up", c.Id)
	topicClose := fmt.Sprintf("tunnel/%s/close", c.Id)

	//连接
	mqtt.Client.Publish(topicOpen, 0, false, c.conn.RemoteAddr().String())

	var n int
	var e error
	for {
		n, e = c.conn.Read(c.buf[:])
		if e != nil {
			_ = c.conn.Close()
			break
		}
		data := c.buf[:n]
		//mqtt.Client.IsConnected()
		//转发
		mqtt.Client.Publish(topicUp, 0, false, data)
	}

	//下线
	mqtt.Client.Publish(topicClose, 0, false, e.Error())
}
