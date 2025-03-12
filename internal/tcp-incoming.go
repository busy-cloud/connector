package internal

import (
	"github.com/busy-cloud/boat/db"
	"sync"
	"time"
)

var tcpIncoming sync.Map //[string, io.ReadWriteCloser]

func init() {
	db.Register(&TcpIncoming{})
}

type TcpIncoming struct {
	Id              string        `json:"id,omitempty" xorm:"pk"`
	ServerId        string        `json:"server_id,omitempty" xorm:"index"`
	Name            string        `json:"name,omitempty"`
	Disabled        bool          `json:"disabled,omitempty"`               //禁用
	Protocol        string        `json:"protocol,omitempty"`               //通讯协议
	ProtocolOptions string        `json:"protocol_options,omitempty"`       //通讯协议
	Created         time.Duration `json:"created,omitempty" xorm:"created"` //创建时间
}
