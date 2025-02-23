package connect

import (
	"fmt"
	"github.com/busy-cloud/boat/log"
	"github.com/busy-cloud/boat/mqtt"
	"github.com/busy-cloud/connector/types"
	"go.bug.st/serial"
	"time"
)

type Serial struct {
	*types.Linker

	serial.Port
	buf    [4096]byte
	opened bool
}

func NewSerial(l *types.Linker) *Serial {
	s := &Serial{Linker: l}
	go s.keep()
	return s
}

func (s *Serial) Opened() bool {
	return s.opened
}

func (s *Serial) Connected() bool {
	return s.Port != nil
}

func (s *Serial) Open() (err error) {
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

	s.opened = true
	go s.keep()
	go s.receive()

	return nil
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

		err := s.Open()
		if err != nil {
			log.Error(err)
		}
	}
}

func (s *Serial) receive() {
	topicOpen := fmt.Sprintf("link/%s/open", s.Id)
	topicUp := fmt.Sprintf("link/%s/up", s.Id)
	topicClose := fmt.Sprintf("link/%s/close", s.Id)

	//连接
	mqtt.Client.Publish(topicOpen, 0, false, s.SerialOptions.PortName)

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
		mqtt.Client.Publish(topicUp, 0, false, data)
	}

	//下线
	mqtt.Client.Publish(topicClose, 0, false, e.Error())
}
