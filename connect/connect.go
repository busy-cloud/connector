package connect

import (
	"encoding/json"
	"fmt"
	"github.com/busy-cloud/boat/log"
	"os"
	"sync"
)

var connections sync.Map //[string, io.ReadWriteCloser]

var connectors sync.Map

type Connector interface {
	Open() error
	Close() error
	Opened() bool
	Connected() bool
}

type Connect struct {
	Id            string         `json:"id,omitempty" xorm:"pk"`
	Type          string         `json:"type,omitempty"`           //serial tcp-client tcp-server
	Address       string         `json:"address,omitempty"`        //地址，域名或IP
	Port          uint16         `json:"port,omitempty"`           //端口号
	SerialOptions *SerialOptions `json:"serial_options,omitempty"` //串口参数
	IdOptions     *IdOptions     `json:"id_options,omitempty"`     //ID参数，只在UDP时有用
	Disabled      bool           `json:"disabled,omitempty"`       //禁用
}

type SerialOptions struct {
	PortName   string `json:"port_name,omitempty"`   //port, e.g. COM1 "/dev/ttySerial1".
	BaudRate   int    `json:"baud_rate,omitempty"`   //9600 115200
	DataBits   int    `json:"data_bits,omitempty"`   //5 6 7 8
	StopBits   int    `json:"stop_bits,omitempty"`   //1 2
	ParityMode int    `json:"parity_mode,omitempty"` //0 1 2 NONE ODD EVEN
}

type IdOptions struct {
	Regex string `json:"regex,omitempty"`
	Start int    `json:"start,omitempty"`
	End   int    `json:"end,omitempty"`
}

func GetConnect(id string) Connector {
	val, ok := connectors.Load(id)
	if ok {
		return val.(Connector)
	}
	return nil
}

func LoadConnect(id string) error {
	buf, err := os.ReadFile("connects/" + id + ".json")
	if err != nil {
		return err
	}
	var c Connect
	err = json.Unmarshal(buf, &c)
	if err != nil {
		return err
	}

	var cc Connector

	switch c.Type {
	case "serial":
		cc = NewSerial(&c)
	case "tcp":
		cc = NewTcpClient(&c)
	case "tcp-server":
		cc = NewGNetServer(&c)
	default:
		return fmt.Errorf("unknown connector type: %s", c.Type)
	}

	//保存
	val, loaded := connectors.LoadOrStore(id, cc)
	if loaded {
		err = val.(Connector).Close()
		if err != nil {
			log.Error(err)
		}
	}

	//启动
	err = cc.Open()
	if err != nil {
		return err
	}

	return nil
}

func UnloadConnect(id string) error {
	val, loaded := connectors.LoadAndDelete(id)
	if loaded {
		return val.(Connector).Close()
	}
	return nil
}
