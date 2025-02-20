package apis

import (
	"github.com/busy-cloud/boat/api"
	"github.com/busy-cloud/boat/curd"
	"github.com/busy-cloud/connector/connect"
	"github.com/busy-cloud/connector/types"
	"github.com/gin-gonic/gin"
)

func init() {
	api.Register("GET", "connector/incoming/list", curd.ApiList[types.Incoming]())
	api.Register("POST", "connector/incoming/create", curd.ApiCreate[types.Incoming]())
	api.Register("GET", "connector/incoming/:id", curd.ParseParamStringId, curd.ApiGet[types.Incoming]())
	api.Register("POST", "connector/incoming/:id", curd.ParseParamStringId, curd.ApiUpdate[types.Incoming]("id", "name"))
	api.Register("GET", "connector/incoming/:id/delete", curd.ParseParamStringId, curd.ApiDelete[types.Incoming]())
	api.Register("GET", "connector/incoming/:id/enable", curd.ParseParamStringId, curd.ApiDisable[types.Incoming](false))
	api.Register("GET", "connector/incoming/:id/disable", curd.ParseParamStringId, curd.ApiDisable[types.Incoming](true))
	api.Register("GET", "connector/incoming/:id/close", curd.ParseParamStringId, incomingClose)
}

func incomingClose(ctx *gin.Context) {
	c := connect.GetIncoming(ctx.Param("id"))
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
