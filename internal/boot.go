package internal

import (
	"github.com/busy-cloud/boat/boot"
	"time"
)

func init() {
	boot.Register("connector", &boot.Task{
		Startup:  Startup,
		Shutdown: Shutdown,
		Depends:  []string{"log", "mqtt", "database"},
	})
}

func Startup() error {

	//订阅通知
	subscribe()

	//go LoadLinkers()
	//5秒后再启动，先让其他准备好
	time.AfterFunc(time.Second*5, LoadLinkers)

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
