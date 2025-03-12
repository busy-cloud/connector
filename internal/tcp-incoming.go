package internal

import (
	"errors"
	"github.com/busy-cloud/boat/db"
	"net"
	"time"
)

func init() {
	db.Register(&TcpIncoming{})
}

type TcpIncoming struct {
	Id              string        `json:"s.Id,omitempty" xorm:"pk"`
	ServerId        string        `json:"server_s.Id,omitempty" xorm:"index"`
	Name            string        `json:"name,omitempty"`
	Disabled        bool          `json:"disabled,omitempty"`               //禁用
	Protocol        string        `json:"protocol,omitempty"`               //通讯协议
	ProtocolOptions string        `json:"protocol_options,omitempty"`       //通讯协议
	Created         time.Duration `json:"created,omitempty" xorm:"created"` //创建时间

	conn   net.Conn
	closed bool
}

func (s *TcpIncoming) Read(p []byte) (n int, err error) {
	return s.conn.Read(p)
}

func (s *TcpIncoming) Write(p []byte) (n int, err error) {
	return s.conn.Write(p)
}

func (s *TcpIncoming) Close() (err error) {
	s.closed = true
	return s.conn.Close()
}

func (s *TcpIncoming) Opened() bool {
	return !s.closed
}

func (s *TcpIncoming) Connected() bool {
	return !s.closed
}

func (s *TcpIncoming) Open() (err error) {
	return errors.New("unsupported open")
}
