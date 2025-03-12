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

func (s *Serial) Opened() bool {
	return s.opened
}

func (s *Serial) Connected() bool {
	return s.Port != nil
}

func (s *Serial) connect() (err error) {
	if s.Port != nil {
		_ = s.Port.Close()
	}

	opts := serial.Mode{
		BaudRate: s.SerialOptions.BaudRate,
		DataBits: s.SerialOptions.DataBits,
		StopBits: serial.StopBits(s.SerialOptions.StopBits),
		Parity:   serial.Parity(s.SerialOptions.ParityMode),
	}

	log.Trace("create serial ", s.Address, opts)
	s.Port, err = serial.Open(s.SerialOptions.PortName, &opts)
	if err != nil {
		return err
	}

	//读超时
	err = s.Port.SetReadTimeout(time.Second * 5)
	if err != nil {
		return err
	}

	go s.receive()

	return
}

func (s *Serial) Open() (err error) {
	if s.opened {
		return errors.New("already open")
	}
	s.opened = true

	//保持连接
	go s.keep()

	return s.connect()
}

func (s *Serial) Close() (err error) {
	s.opened = false

	if s.Port != nil {
		return s.Port.Close()
	}
	return nil
}

func (s *Serial) keep() {
	for s.opened {
		time.Sleep(time.Minute)

		if s.Port != nil {
			continue
		}

		err := s.connect()
		if err != nil {
			log.Error(err)
		}
	}
}

func (s *Serial) receive() {
	links.Store(s.Id, s)

	//连接
	topicOpen := fmt.Sprintf("link/%s/open", s.Id)
	mqtt.Publish(topicOpen, s.SerialOptions.PortName)
	if s.Protocol != "" {
		topic := fmt.Sprintf("%s/%s/open", s.Protocol, s.Id)
		mqtt.Publish(topic, s.SerialOptions.PortName)
	}

	//接收数据
	topicUp := fmt.Sprintf("link/%s/up", s.Id)
	topicUpProtocol := fmt.Sprintf("%s/%s/up", s.Protocol, s.Id)

	var n int
	var e error
	for {
		n, e = s.Port.Read(s.buf[:])
		if e != nil {
			_ = s.Port.Close()
			s.Port = nil
			break
		}
		data := s.buf[:n]
		//mqtt.TcpClient.IsConnected()
		//转发
		mqtt.Publish(topicUp, data)
		if s.Protocol != "" {
			mqtt.Publish(topicUpProtocol, data)
		}
	}

	//下线
	topicClose := fmt.Sprintf("link/%s/close", s.Id)
	mqtt.Publish(topicClose, e.Error())
	if s.Protocol != "" {
		topic := fmt.Sprintf("%s/%s/close", s.Protocol, s.Id)
		mqtt.Publish(topic, s.SerialOptions.PortName)
	}

	links.Delete(s.Id)
}
