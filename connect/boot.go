package connect

import "github.com/busy-cloud/boat/boot"

func init() {
	boot.Register("connect", &boot.Task{
		Startup:  Startup,
		Shutdown: nil,
		Depends:  []string{"log"},
	})
}
