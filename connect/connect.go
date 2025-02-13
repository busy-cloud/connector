package connect

import (
	"encoding/json"
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
	Type          string         `json:"type,omitempty"`           //tcp udp serial
	Addr          string         `json:"addr,omitempty"`           // :6000 192.168.0.100:8088 COM1
	Server        bool           `json:"server,omitempty"`         //服务端
	Singleton     bool           `json:"singleton,omitempty"`      //tcp server时，单例模式
	SerialOptions *SerialOptions `json:"serial_options,omitempty"` //串口参数
	IdOptions     *IdOptions     `json:"id_options,omitempty"`     //ID参数，只在UDP时有用
	Disabled      bool           `json:"disabled,omitempty"`       //禁用
}

type SerialOptions struct {
	//PortName   string `json:"port_name,omitempty"`   //port, e.g. COM1 "/dev/ttySerial1".
	BaudRate   int `json:"baud_rate,omitempty"`   //9600 115200
	DataBits   int `json:"data_bits,omitempty"`   //5 6 7 8
	StopBits   int `json:"stop_bits,omitempty"`   //1 2
	ParityMode int `json:"parity_mode,omitempty"` //0 1 2 NONE ODD EVEN
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

	if c.Type == "serial" {
		cc = NewSerial(&c)
	} else if c.Server {
		cc = NewGNetServer(&c)
	} else {
		cc = NewTcpClient(&c)
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
