package internal

import (
	"errors"
	"fmt"
	"github.com/busy-cloud/boat/log"
	"github.com/busy-cloud/boat/mqtt"
	"go.bug.st/serial"
	"time"
)

type Serial struct {
	*Linker

	serial.Port
	buf    [4096]byte
	opened bool
}

func NewSerial(l *Linker) *Serial {
	s := &Serial{Linker: l}
	//go s.keep()
	return s
}

func (c *Serial) Opened() bool {
	return c.opened
}

func (c *Serial) Connected() bool {
	return c.Port != nil
}

func (c *Serial) Error() string {
	return c.Linker.Error
}

func (c *Serial) connect() (err error) {
	defer func() {
		if err != nil {
			c.Linker.Error = err.Error()
		} else {
			c.Linker.Error = ""
		}
	}()

	if c.Port != nil {
		_ = c.Port.Close()
	}

	if c.SerialOptions == nil {
		return errors.New("serial options is blank")
	}

	opts := serial.Mode{
		BaudRate: c.SerialOptions.BaudRate,
		DataBits: c.SerialOptions.DataBits,
		StopBits: serial.StopBits(c.SerialOptions.StopBits),
		Parity:   serial.Parity(c.SerialOptions.ParityMode),
	}

	log.Trace("create serial ", c.Address, opts)
	c.Port, err = serial.Open(c.SerialOptions.PortName, &opts)
	if err != nil {
		return err
	}

	//读超时
	err = c.Port.SetReadTimeout(time.Second * 5)
	if err != nil {
		return err
	}

	go c.receive()

	return
}

func (c *Serial) Open() (err error) {
	if c.opened {
		return errors.New("already open")
	}
	c.opened = true

	//保持连接
	go c.keep()

	return c.connect()
}

func (c *Serial) Close() (err error) {
	c.opened = false

	if c.Port != nil {
		return c.Port.Close()
	}
	return nil
}

func (c *Serial) keep() {
	for c.opened {
		time.Sleep(time.Minute)

		if c.Port == nil {
			err := c.connect()
			if err != nil {
				log.Error(err)
			}
		}
	}
}

func (c *Serial) receive() {
	links.Store(c.Id, c)

	//连接
	topicOpen := fmt.Sprintf("link/%s/open", c.Id)
	mqtt.Publish(topicOpen, c.SerialOptions.PortName)
	if c.Protocol != "" {
		topic := fmt.Sprintf("%s/%s/open", c.Protocol, c.Id)
		mqtt.Publish(topic, c.ProtocolOptions)
	}

	//接收数据
	topicUp := fmt.Sprintf("link/%s/up", c.Id)
	topicUpProtocol := fmt.Sprintf("%s/%s/up", c.Protocol, c.Id)

	var n int
	var e error
	for {
		n, e = c.Port.Read(c.buf[:])
		if e != nil {
			_ = c.Port.Close()
			c.Port = nil
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
		mqtt.Publish(topic, e.Error())
	}

	links.Delete(c.Id)

	//清空连接
	c.Port = nil
}
