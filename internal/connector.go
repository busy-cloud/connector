package internal

import (
	"github.com/busy-cloud/boat/db"
	"github.com/busy-cloud/boat/log"
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
	links.Range(func(key, value interface{}) bool {
		linker := value.(Link)
		_ = linker.Close()
		return true
	})
	return nil
}
