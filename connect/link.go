package connect

import (
	"encoding/json"
	"fmt"
	"github.com/busy-cloud/boat/log"
	"os"
	"sync"
)

var links sync.Map //[string, io.ReadWriteCloser]

var linkers sync.Map //[string, Linker]

type Link interface {
	Open() error
	Close() error
	Opened() bool
	Connected() bool
}

type Linker struct {
	Id            string         `json:"id,omitempty" xorm:"pk"`
	Type          string         `json:"type,omitempty"`           //serial tcp-client tcp-server
	Address       string         `json:"address,omitempty"`        //地址，域名或IP
	Port          uint16         `json:"port,omitempty"`           //端口号
	SerialOptions *SerialOptions `json:"serial_options,omitempty"` //串口参数
	IdRegex       string         `json:"id_regex,omitempty"`       //ID正则表达式
	Disabled      bool           `json:"disabled,omitempty"`       //禁用
}

type SerialOptions struct {
	PortName   string `json:"port_name,omitempty"`   //port, e.g. COM1 "/dev/ttySerial1".
	BaudRate   int    `json:"baud_rate,omitempty"`   //9600 115200
	DataBits   int    `json:"data_bits,omitempty"`   //5 6 7 8
	StopBits   int    `json:"stop_bits,omitempty"`   //1 2
	ParityMode int    `json:"parity_mode,omitempty"` //0 1 2 NONE ODD EVEN
}

func GetConnect(id string) Link {
	val, ok := linkers.Load(id)
	if ok {
		return val.(Link)
	}
	return nil
}

func LoadConnect(id string) error {
	buf, err := os.ReadFile("connects/" + id + ".json")
	if err != nil {
		return err
	}
	var l Linker
	err = json.Unmarshal(buf, &l)
	if err != nil {
		return err
	}

	var link Link

	switch l.Type {
	case "serial":
		link = NewSerial(&l)
	case "tcp":
		link = NewTcpClient(&l)
	case "tcp-server":
		link = NewTcpServer(&l)
	case "tcp-server-multiple":
		link = NewGNetServer(&l)
	default:
		return fmt.Errorf("unknown connector type: %s", l.Type)
	}

	//保存
	val, loaded := linkers.LoadOrStore(id, link)
	if loaded {
		err = val.(Link).Close()
		if err != nil {
			log.Error(err)
		}
	}

	//启动
	err = link.Open()
	if err != nil {
		return err
	}

	return nil
}

func UnloadConnect(id string) error {
	val, loaded := linkers.LoadAndDelete(id)
	if loaded {
		return val.(Link).Close()
	}
	return nil
}
