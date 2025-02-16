package connect

import (
	"encoding/json"
	"fmt"
	"github.com/busy-cloud/boat/log"
	"github.com/busy-cloud/connector/interfaces"
	"github.com/busy-cloud/connector/types"
	"os"
	"sync"
)

var linkers sync.Map //[string, Linker]

func GetLinker(id string) interfaces.Linker {
	val, ok := linkers.Load(id)
	if ok {
		return val.(interfaces.Linker)
	}
	return nil
}

func LoadLinker(id string) error {
	buf, err := os.ReadFile("connects/" + id + ".json")
	if err != nil {
		return err
	}
	var l types.Linker
	err = json.Unmarshal(buf, &l)
	if err != nil {
		return err
	}

	var linker interfaces.Linker

	switch l.Type {
	case "serial":
		linker = NewSerial(&l)
	case "tcp":
		linker = NewTcpClient(&l)
	case "tcp-server":
		linker = NewTcpServer(&l)
	case "tcp-server-multiple":
		linker = NewGNetServer(&l)
	default:
		return fmt.Errorf("unknown connector type: %s", l.Type)
	}

	//保存
	val, loaded := linkers.LoadOrStore(id, linker)
	if loaded {
		err = val.(interfaces.Linker).Close()
		if err != nil {
			log.Error(err)
		}
	}

	//启动
	err = linker.Open()
	if err != nil {
		return err
	}

	return nil
}

func UnloadLinker(id string) error {
	val, loaded := linkers.LoadAndDelete(id)
	if loaded {
		return val.(interfaces.Linker).Close()
	}
	return nil
}
