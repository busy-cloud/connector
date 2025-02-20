package apis

import (
	"github.com/busy-cloud/boat/api"
	"github.com/busy-cloud/boat/curd"
	"github.com/busy-cloud/connector/connect"
	"github.com/busy-cloud/connector/types"
	"github.com/gin-gonic/gin"
)

func init() {
	api.Register("GET", "connector/linker/list", curd.ApiList[types.Linker]())
	api.Register("POST", "connector/linker/create", curd.ApiCreate[types.Linker]())
	api.Register("GET", "connector/linker/:id", curd.ParseParamStringId, curd.ApiGet[types.Linker]())
	api.Register("POST", "connector/linker/:id", curd.ParseParamStringId, curd.ApiUpdate[types.Linker]("id"))
	api.Register("GET", "connector/linker/:id/delete", curd.ParseParamStringId, curd.ApiDelete[types.Linker]())
	api.Register("GET", "connector/linker/:id/enable", curd.ParseParamStringId, curd.ApiDisable[types.Linker](false))
	api.Register("GET", "connector/linker/:id/disable", curd.ParseParamStringId, curd.ApiDisable[types.Linker](true))
	api.Register("GET", "connector/linker/:id/open", curd.ParseParamStringId, linkerOpen)
	api.Register("GET", "connector/linker/:id/close", curd.ParseParamStringId, linkerClose)
}

func linkerClose(ctx *gin.Context) {
	c := connect.GetLinker(ctx.Param("id"))
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

func linkerOpen(ctx *gin.Context) {
	c := connect.GetLinker(ctx.Param("id"))
	if c == nil {
		api.Fail(ctx, "找不到连接")
		return
	}

	err := c.Open()
	if err != nil {
		api.Error(ctx, err)
		return
	}

	api.OK(ctx, nil)
}
