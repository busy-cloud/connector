package internal

import (
	"errors"
	"github.com/busy-cloud/boat/api"
	"github.com/busy-cloud/boat/curd"
	"github.com/gin-gonic/gin"
)

func init() {
	api.Register("GET", "tcp-incoming/list", curd.ApiList[TcpIncoming]())
	api.Register("POST", "tcp-incoming/create", curd.ApiCreate[TcpIncoming]())
	api.Register("POST", "tcp-incoming/search", curd.ApiSearch[TcpIncoming]())
	api.Register("GET", "tcp-incoming/:id", curd.ApiGet[TcpIncoming]())

	api.Register("POST", "tcp-incoming/:id", curd.ApiUpdateHook[TcpIncoming](nil, func(m *TcpIncoming) error {
		return unloadIncoming(m.Id)
	}, "id", "name", "disabled", "protocol", "protocol_options"))

	api.Register("GET", "tcp-incoming/:id/delete", curd.ApiDeleteHook[TcpIncoming](nil, func(m *TcpIncoming) error {
		return unloadIncoming(m.Id)
	}))

	api.Register("GET", "tcp-incoming/:id/enable", curd.ApiDisable[TcpIncoming](false))
	api.Register("GET", "tcp-incoming/:id/disable", curd.ApiDisableHook[TcpIncoming](true, nil, func(id any) error {
		return unloadIncoming(id.(string))
	}))

	api.Register("GET", "tcp-incoming/:id/close", incomingClose)
}

func unloadIncoming(id string) error {
	c, ok := links.LoadAndDelete(id)
	if !ok {
		return errors.New("tcp-incoming not found")
	}
	ti := c.(Link)
	return ti.Close()
}

func incomingClose(ctx *gin.Context) {
	err := unloadIncoming(ctx.Param("id"))
	if err != nil {
		api.Error(ctx, err)
		return
	}

	api.OK(ctx, nil)
}
