package internal

import (
	"github.com/busy-cloud/boat/api"
	"github.com/busy-cloud/boat/curd"
	"github.com/gin-gonic/gin"
	"go.bug.st/serial"
)

func init() {
	api.Register("GET", "linker/serials", linkerSerials)
	api.Register("GET", "linker/list", curd.ApiList[Linker]())
	api.Register("POST", "linker/search", curd.ApiSearch[Linker]())
	api.Register("POST", "linker/create", curd.ApiCreateHook[Linker](nil, FromLinker))
	api.Register("GET", "linker/:id", curd.ApiGet[Linker]())

	api.Register("POST", "linker/:id", curd.ApiUpdateHook[Linker](nil, func(m *Linker) error {
		return FromLinker(m)
	}, "id", "name", "type", "address", "port", "serial_options", "register_options", "disabled", "protocol", "protocol_options"))

	api.Register("GET", "linker/:id/delete", curd.ApiDeleteHook[Linker](nil, func(m *Linker) error {
		return UnloadLink(m.Id)
	}))

	api.Register("GET", "linker/:id/enable", curd.ApiDisableHook[Linker](false, nil, func(id any) error {
		return LoadLink(id.(string))
	}))

	api.Register("GET", "linker/:id/disable", curd.ApiDisableHook[Linker](true, nil, func(id any) error {
		return UnloadLink(id.(string))
	}))

	api.Register("GET", "linker/:id/open", linkerOpen)
	api.Register("GET", "linker/:id/close", linkerClose)
}

func linkerSerials(ctx *gin.Context) {
	ss, err := serial.GetPortsList()
	if err != nil {
		api.Error(ctx, err)
		return
	}

	api.OK(ctx, ss)
}

func linkerClose(ctx *gin.Context) {
	c := GetLink(ctx.Param("id"))
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
	err := LoadLink(ctx.Param("id"))
	if err != nil {
		api.Error(ctx, err)
		return
	}

	api.OK(ctx, nil)
}
