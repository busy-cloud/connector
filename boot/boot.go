package boot

import (
	"github.com/busy-cloud/boat/boot"
	"github.com/busy-cloud/connector/connect"
)

func init() {
	boot.Register("connector", &boot.Task{
		Startup:  connect.Startup,
		Shutdown: nil,
		Depends:  []string{"log", "mqtt", "database"},
	})
}
