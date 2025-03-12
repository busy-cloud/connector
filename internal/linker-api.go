package internal

import (
	"github.com/busy-cloud/boat/api"
	"github.com/busy-cloud/boat/curd"
	"github.com/gin-gonic/gin"
)

func init() {
	api.Register("GET", "linker/list", curd.ApiList[Linker]())
	api.Register("POST", "linker/search", curd.ApiSearch[Linker]())
	api.Register("POST", "linker/create", curd.ApiCreate[Linker]())
	api.Register("GET", "linker/:id", curd.ParseParamStringId, curd.ApiGet[Linker]())
	api.Register("POST", "linker/:id", curd.ParseParamStringId, curd.ApiUpdate[Linker]("id", "name", "type", "address", "port", "serial", "id_regex", "disabled", "protocol", "protocol_options"))
	api.Register("GET", "linker/:id/delete", curd.ParseParamStringId, curd.ApiDelete[Linker]())
	api.Register("GET", "linker/:id/enable", curd.ParseParamStringId, curd.ApiDisable[Linker](false))
	api.Register("GET", "linker/:id/disable", curd.ParseParamStringId, curd.ApiDisable[Linker](true))
	api.Register("GET", "linker/:id/open", curd.ParseParamStringId, linkerOpen)
	api.Register("GET", "linker/:id/close", curd.ParseParamStringId, linkerClose)
}

func linkerClose(ctx *gin.Context) {
	c := GetLinker(ctx.Param("id"))
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
	c := GetLinker(ctx.Param("id"))
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
