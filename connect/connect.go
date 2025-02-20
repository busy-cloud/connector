package connect

import (
	"github.com/busy-cloud/boat/db"
	"github.com/busy-cloud/boat/log"
	"github.com/busy-cloud/connector/types"
)

func Startup() error {

	//订阅通知
	subscribe()

	//加载连接器
	var linkers []*types.Linker
	err := db.Engine.Find(&linkers)
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
