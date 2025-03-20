package internal

import (
	"github.com/busy-cloud/boat/db"
	"time"
)

func init() {
	db.Register(&Linker{})
}

type Linker struct {
	Id              string           `json:"id,omitempty" xorm:"pk"`
	Name            string           `json:"name,omitempty"`
	Type            string           `json:"type,omitempty"`                            //serial tcp-client tcp-server tcp-server-multiple tcp-incoming
	Address         string           `json:"address,omitempty"`                         //地址，域名或IP
	Port            uint16           `json:"port,omitempty"`                            //端口号
	SerialOptions   *SerialOptions   `json:"serial_options,omitempty" xorm:"json"`      //串口参数
	RegisterOptions *RegisterOptions `json:"register_options,omitempty" xorm:"json"`    //注册包参数
	Disabled        bool             `json:"disabled,omitempty"`                        //禁用
	Protocol        string           `json:"protocol,omitempty"`                        //通讯协议
	ProtocolOptions map[string]any   `json:"protocol_options,omitempty" xorm:"json"`    //通讯协议参数
	Created         time.Duration    `json:"created,omitempty,omitzero" xorm:"created"` //创建时间
}

type SerialOptions struct {
	PortName   string `json:"port_name,omitempty"`   //port, e.g. COM1 "/dev/ttySerial1".
	BaudRate   int    `json:"baud_rate,omitempty"`   //9600 115200
	DataBits   int    `json:"data_bits,omitempty"`   //5 6 7 8
	StopBits   int    `json:"stop_bits,omitempty"`   //1 2
	ParityMode int    `json:"parity_mode,omitempty"` //0 1 2 NONE ODD EVEN
}

type RegisterOptions struct {
	Type   string `json:"type,omitempty"`   //注册类型 string, json
	Regex  string `json:"regex,omitempty"`  //ID正则表达式
	Field  string `json:"field,omitempty"`  //注册包为JSON时，取一个字段作为ID
	Offset uint16 `json:"offset,omitempty"` //偏移，用于处理固定包头
	Length uint16 `json:"length,omitempty"` //取长度
}
