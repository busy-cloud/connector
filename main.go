package main

import (
	"github.com/busy-cloud/boat/api"
	"github.com/busy-cloud/boat/boot"
	"github.com/busy-cloud/boat/log"
	"github.com/busy-cloud/boat/web"
	_ "github.com/busy-cloud/connector/apis"
	_ "github.com/busy-cloud/connector/connect"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs

		//关闭web，出发
		_ = web.Shutdown()
	}()

	//安全退出
	defer boot.Shutdown()

	err := boot.Startup()
	if err != nil {
		log.Fatal(err)
		return
	}

	//注册接口
	api.RegisterRoutes(web.Engine.Group("api"))

	err = web.ServeHTTP()
	if err != nil {
		log.Fatal(err)
	}
}
