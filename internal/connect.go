package internal

import (
	"github.com/busy-cloud/boat/db"
	"github.com/busy-cloud/boat/log"
	"github.com/busy-cloud/connector/interfaces"
	"io"
)

func Startup() error {

	//订阅通知
	subscribe()

	//加载连接器
	var linkers []*Linker
	err := db.Engine().Find(&linkers)
	if err != nil {
		return err
	}
	for _, linker := range linkers {
		err := FromLinker(linker)
		if err != nil {
			log.Error(err)
		}
	}

	return nil
}

func Shutdown() error {
	linkers.Range(func(key, value interface{}) bool {
		linker := value.(interfaces.Linker)
		_ = linker.Close()
		return true
	})
	incomingConnections.Range(func(key, value any) bool {
		conn := value.(io.ReadWriteCloser)
		if conn != nil {
			_ = conn.Close()
		}
		return true
	})
	return nil
}
