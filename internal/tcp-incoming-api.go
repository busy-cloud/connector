package internal

import (
	"errors"
	"github.com/busy-cloud/boat/api"
	"github.com/busy-cloud/boat/curd"
	"github.com/gin-gonic/gin"
)

func init() {
	api.Register("GET", "connector/tcp-incoming/list", curd.ApiListHook[TcpIncoming](getIncomingsInfo))
	api.Register("POST", "connector/tcp-incoming/create", curd.ApiCreate[TcpIncoming]())
	api.Register("POST", "connector/tcp-incoming/search", curd.ApiSearchHook[TcpIncoming](getIncomingsInfo))
	api.Register("GET", "connector/tcp-incoming/:id", curd.ApiGetHook[TcpIncoming](getIncomingInfo))

	api.Register("POST", "connector/tcp-incoming/:id", curd.ApiUpdateHook[TcpIncoming](nil, func(m *TcpIncoming) error {
		_ = unloadIncoming(m.Id)
		return nil
	}, "id", "name", "disabled", "protocol", "protocol_options"))

	api.Register("GET", "connector/tcp-incoming/:id/delete", curd.ApiDeleteHook[TcpIncoming](nil, func(m *TcpIncoming) error {
		_ = unloadIncoming(m.Id)
		return nil
	}))

	api.Register("GET", "connector/tcp-incoming/:id/enable", curd.ApiDisable[TcpIncoming](false))
	api.Register("GET", "connector/tcp-incoming/:id/disable", curd.ApiDisableHook[TcpIncoming](true, nil, func(id any) error {
		_ = unloadIncoming(id.(string))
		return nil
	}))

	api.Register("GET", "connector/tcp-incoming/:id/close", incomingClose)
}

func getIncomingsInfo(ds []*TcpIncoming) error {
	for _, d := range ds {
		_ = getIncomingInfo(d)
	}
	return nil
}

func getIncomingInfo(d *TcpIncoming) error {
	l := GetLink(d.Id)
	if l != nil {
		d.Running = l.Connected()
	}
	return nil
}

func unloadIncoming(id string) error {
	c, ok := links.LoadAndDelete(id)
	if !ok {
		return errors.New("找不到连接")
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
