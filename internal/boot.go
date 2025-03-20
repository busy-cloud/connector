package internal

import (
	"github.com/busy-cloud/boat/boot"
)

func init() {
	boot.Register("connector", &boot.Task{
		Startup:  Startup,
		Shutdown: Shutdown,
		Depends:  []string{"log", "mqtt", "database"},
	})
}
