package boot

import (
	"github.com/busy-cloud/boat/boot"
	"github.com/busy-cloud/connector/internal"
)

func init() {
	boot.Register("connector", &boot.Task{
		Startup:  internal.Startup,
		Shutdown: internal.Shutdown,
		Depends:  []string{"log", "mqtt", "database"},
	})
}
