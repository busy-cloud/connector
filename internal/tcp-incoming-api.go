package internal

import (
	"github.com/busy-cloud/boat/api"
	"github.com/busy-cloud/boat/curd"
	"github.com/gin-gonic/gin"
	"io"
)

func init() {
	api.Register("GET", "tcp-incoming/list", curd.ApiList[TcpIncoming]())
	api.Register("POST", "tcp-incoming/create", curd.ApiCreate[TcpIncoming]())
	api.Register("POST", "tcp-incoming/search", curd.ApiSearch[TcpIncoming]())
	api.Register("GET", "tcp-incoming/:id", curd.ApiGet[TcpIncoming]())
	api.Register("POST", "tcp-incoming/:id", curd.ApiUpdate[TcpIncoming]("id", "name", "disabled", "protocol", "protocol_options"))
	api.Register("GET", "tcp-incoming/:id/delete", curd.ApiDelete[TcpIncoming]())
	api.Register("GET", "tcp-incoming/:id/enable", curd.ApiDisable[TcpIncoming](false))
	api.Register("GET", "tcp-incoming/:id/disable", curd.ApiDisable[TcpIncoming](true))
	api.Register("GET", "tcp-incoming/:id/close", incomingClose)
}

func incomingClose(ctx *gin.Context) {
	c, ok := tcpIncoming.Load(ctx.Param("id"))
	if !ok {
		api.Fail(ctx, "找不到连接")
		return
	}

	ti := c.(io.ReadWriteCloser)

	err := ti.Close()
	if err != nil {
		api.Error(ctx, err)
		return
	}

	api.OK(ctx, nil)
}
