package internal

import (
	"github.com/busy-cloud/boat/api"
	"github.com/busy-cloud/boat/curd"
	"github.com/gin-gonic/gin"
)

func init() {
	api.Register("GET", "tcp-incoming/list", curd.ApiList[TcpIncoming]())
	api.Register("POST", "tcp-incoming/create", curd.ApiCreate[TcpIncoming]())
	api.Register("POST", "tcp-incoming/search", curd.ApiSearch[TcpIncoming]())
	api.Register("GET", "tcp-incoming/:id", curd.ParseParamStringId, curd.ApiGet[TcpIncoming]())
	api.Register("POST", "tcp-incoming/:id", curd.ParseParamStringId, curd.ApiUpdate[TcpIncoming]("id", "name", "disabled", "protocol"))
	api.Register("GET", "tcp-incoming/:id/delete", curd.ParseParamStringId, curd.ApiDelete[TcpIncoming]())
	api.Register("GET", "tcp-incoming/:id/enable", curd.ParseParamStringId, curd.ApiDisable[TcpIncoming](false))
	api.Register("GET", "tcp-incoming/:id/disable", curd.ParseParamStringId, curd.ApiDisable[TcpIncoming](true))
	api.Register("GET", "tcp-incoming/:id/close", curd.ParseParamStringId, incomingClose)
}

func incomingClose(ctx *gin.Context) {
	c := GetIncoming(ctx.Param("id"))
	if c == nil {
		api.Fail(ctx, "找不到连接")
		return
	}

	err := c.Close()
	if err != nil {
		api.Error(ctx, err)
		return
	}

	api.OK(ctx, nil)
}
