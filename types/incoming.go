package types

import (
	"github.com/busy-cloud/boat/db"
	"time"
)

func init() {
	db.Register(&Incoming{})
}

type Incoming struct {
	Id       string        `json:"id,omitempty" xorm:"pk"`
	ServerId string        `json:"server_id,omitempty" xorm:"index"`
	Name     string        `json:"name,omitempty"`
	Disabled bool          `json:"disabled,omitempty"`               //禁用
	Protocol string        `json:"protocol,omitempty"`               //通讯协议
	Created  time.Duration `json:"created,omitempty" xorm:"created"` //创建时间
}
