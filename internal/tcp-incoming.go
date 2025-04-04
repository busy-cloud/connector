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
	Id              string         `json:"id,omitempty" xorm:"pk"`
	LinkerId        string         `json:"linker_id,omitempty" xorm:"index"`
	Name            string         `json:"name,omitempty"`
	Disabled        bool           `json:"disabled,omitempty"`                        //禁用
	Protocol        string         `json:"protocol,omitempty"`                        //通讯协议
	ProtocolOptions map[string]any `json:"protocol_options,omitempty" xorm:"json"`    //通讯协议
	Created         time.Time      `json:"created,omitempty,omitzero" xorm:"created"` //创建时间

	Running bool `json:"running,omitempty" xorm:"-"` //实时状态不入库

	conn   net.Conn
	closed bool
}

func (s *TcpIncoming) Read(p []byte) (n int, err error) {
	return s.conn.Read(p)
}

func (s *TcpIncoming) Error() string {
	return ""
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
