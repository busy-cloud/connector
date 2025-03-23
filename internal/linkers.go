package internal

import (
	"fmt"
	"github.com/busy-cloud/boat/db"
	"github.com/busy-cloud/boat/log"
	"sync"
)

var links sync.Map //[string, Link]

func GetLink(id string) Link {
	val, ok := links.Load(id)
	if ok {
		return val.(Link)
	}
	return nil
}

func FromLinker(l *Linker) error {
	var link Link

	switch l.Type {
	case "serial":
		link = NewSerial(l)
	case "tcp-client":
		link = NewTcpClient(l)
	case "tcp-server":
		link = NewTcpServer(l)
	case "tcp-server-multiple":
		link = NewTcpServerMultiple(l)
	case "gnet-server":
		link = NewGNetServer(l)
	default:
		return fmt.Errorf("unknown connector type: %s", l.Type)
	}

	//保存
	val, loaded := links.LoadOrStore(l.Id, link)
	if loaded {
		err := val.(Link).Close()
		if err != nil {
			log.Error(err)
		}
	}

	//启动
	err := link.Open()
	if err != nil {
		return err
	}

	return nil
}

func LoadLink(id string) error {
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

func UnloadLink(id string) error {
	val, loaded := links.LoadAndDelete(id)
	if loaded {
		return val.(Link).Close()
	}
	return nil
}

func LoadLinkers() {
	//加载连接器
	var linkers []*Linker
	err := db.Engine().Find(&linkers)
	if err != nil {
		log.Error(err)
		return
	}
	for _, linker := range linkers {
		if linker.Disabled {
			log.Info("linker %s is disabled", linker.Id)
			continue
		}
		err := FromLinker(linker)
		if err != nil {
			log.Error(err)
		}
	}
}
