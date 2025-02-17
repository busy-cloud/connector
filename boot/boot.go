package boot

import (
	"github.com/busy-cloud/boat/boot"
	_ "github.com/busy-cloud/connector/apis"
	"github.com/busy-cloud/connector/connect"
)

func init() {
	boot.Register("connector", &boot.Task{
		Startup:  connect.Startup,
		Shutdown: nil,
		Depends:  []string{"log", "mqtt", "database"},
	})
}
