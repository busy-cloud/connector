package internal

import (
	"fmt"
	"github.com/busy-cloud/boat/db"
	"github.com/busy-cloud/boat/log"
	"github.com/busy-cloud/connector/interfaces"
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

func FromLinker(l *Linker) error {
	var linker interfaces.Linker

	switch l.Type {
	case "serial":
		linker = NewSerial(l)
	case "tcp":
		linker = NewTcpClient(l)
	case "tcp-server":
		linker = NewTcpServer(l)
	case "tcp-server-multiple":
		linker = NewTcpServerMultiple(l)
	case "gnet-server":
		linker = NewGNetServer(l)
	default:
		return fmt.Errorf("unknown connector type: %s", l.Type)
	}

	//保存
	val, loaded := linkers.LoadOrStore(l.Id, linker)
	if loaded {
		err := val.(interfaces.Linker).Close()
		if err != nil {
			log.Error(err)
		}
	}

	//启动
	err := linker.Open()
	if err != nil {
		return err
	}

	return nil
}

func LoadLinker(id string) error {
	var l Linker
	has, err := db.Engine().ID(id).Get(&l)
	if err != nil {
		return err
	}
	if !has {
		return fmt.Errorf("linker %s not found", id)
	}

	return FromLinker(&l)
}

func UnloadLinker(id string) error {
	val, loaded := linkers.LoadAndDelete(id)
	if loaded {
		return val.(interfaces.Linker).Close()
	}
	return nil
}
