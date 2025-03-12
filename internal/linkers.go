package internal

import (
	"fmt"
	"github.com/busy-cloud/boat/db"
	"github.com/busy-cloud/boat/log"
	"sync"
)

var linkers sync.Map //[string, Instance]

func GetLinker(id string) Instance {
	val, ok := linkers.Load(id)
	if ok {
		return val.(Instance)
	}
	return nil
}

func FromLinker(l *Linker) error {
	var linker Instance

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
		err := val.(Instance).Close()
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
		return val.(Instance).Close()
	}
	return nil
}
