package main

import (
	"github.com/busy-cloud/boat/boot"
	"github.com/busy-cloud/boat/log"
	_ "github.com/busy-cloud/connector/connect"
)

func main() {
	err := boot.Startup()
	if err != nil {
		log.Fatal(err)
		return
	}
	select {}
}
