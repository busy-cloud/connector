package apis

import (
	"encoding/json"
	"github.com/busy-cloud/boat/api"
	"github.com/busy-cloud/connector/connect"
	"github.com/gin-gonic/gin"
	"go.uber.org/multierr"
	"os"
	"path/filepath"
)

const BasePath = "app/connector/connect"

func init() {
	api.Register("GET", BasePath+"/list", connects)
	api.Register("POST", BasePath+"/create", connectCreate)
	api.Register("GET", BasePath+"/:id", connectDetail)
	api.Register("POST", BasePath+"/:id", connectUpdate)
	api.Register("GET", BasePath+"/:id/delete", connectDelete)
	api.Register("GET", BasePath+"/:id/enable", connectEnable)
	api.Register("GET", BasePath+"/:id/disable", connectDisable)
	api.Register("GET", BasePath+"/:id/open", connectOpen)
	api.Register("GET", BasePath+"/:id/close", connectClose)
}

func connectClose(ctx *gin.Context) {
	c := connect.GetConnect(ctx.Param("id"))
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

func connectOpen(ctx *gin.Context) {
	c := connect.GetConnect(ctx.Param("id"))
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

func connectEnable(ctx *gin.Context) {
	id := ctx.Param("id")

	buf, err := os.ReadFile("connects/" + id + ".json")
	if err != nil {
		api.Error(ctx, err)
		return
	}
	var c map[string]any //c.Connect
	err = json.Unmarshal(buf, &c)
	if err != nil {
		api.Error(ctx, err)
		return
	}

	c["disabled"] = false
	buf, err = json.Marshal(c)
	err = os.WriteFile("connects/"+id+".json", buf, os.ModePerm)
	if err != nil {
		api.Error(ctx, err)
		return
	}

	err = connect.LoadConnect(id)
	if err != nil {
		api.Error(ctx, err)
		return
	}

	api.OK(ctx, nil)
}

func connectDisable(ctx *gin.Context) {
	id := ctx.Param("id")

	buf, err := os.ReadFile("connects/" + id + ".json")
	if err != nil {
		api.Error(ctx, err)
		return
	}
	var c map[string]any //c.Connect
	err = json.Unmarshal(buf, &c)
	if err != nil {
		api.Error(ctx, err)
		return
	}

	c["disabled"] = true
	buf, err = json.Marshal(c)
	err = os.WriteFile("connects/"+id+".json", buf, os.ModePerm)
	if err != nil {
		api.Error(ctx, err)
		return
	}

	err = connect.UnloadConnect(id)
	if err != nil {
		api.Error(ctx, err)
		return
	}

	api.OK(ctx, nil)
}

func connectDelete(ctx *gin.Context) {
	id := ctx.Param("id")

	//备份
	_ = os.Rename("connects/"+id+".json", "connects/"+id+".bak")

	err := connect.UnloadConnect(id)
	if err != nil {
		api.Error(ctx, err)
		return
	}

	api.OK(ctx, nil)
}

func connectDetail(ctx *gin.Context) {
	id := ctx.Param("id")
	//ctx.File("connects/*.json")

	buf, err := os.ReadFile("connects/" + id + ".json")
	if err != nil {
		api.Error(ctx, err)
		return
	}
	var c map[string]any //c.Connect
	err = json.Unmarshal(buf, &c)
	if err != nil {
		api.Error(ctx, err)
		return
	}

	cc := connect.GetConnect(ctx.Param("id"))
	if cc != nil {
		c["opened"] = cc.Opened()
		c["connected"] = cc.Connected()
		return
	}

	api.OK(ctx, c)
}

func connects(ctx *gin.Context) {
	var ls []map[string]any //[]*connect.Connect
	files, err := filepath.Glob("connects/*.json")
	if err != nil {
		api.Error(ctx, err)
		return
	}
	var e error
	for _, f := range files {
		buf, err := os.ReadFile("connects/" + f)
		if err != nil {
			//log.Error(err)
			e = multierr.Append(e, err)
			continue
		}
		var c map[string]any //c.Connect
		err = json.Unmarshal(buf, &c)
		if err != nil {
			//log.Error(err)
			e = multierr.Append(e, err)
			continue
		}
		ls = append(ls, c)

		//补充状态
		cc := connect.GetConnect(ctx.Param("id"))
		if cc != nil {
			c["opened"] = cc.Opened()
			c["connected"] = cc.Connected()
			return
		}

	}
	if e != nil {
		api.Error(ctx, e)
		return
	}

	api.OK(ctx, ls)
}

func connectCreate(ctx *gin.Context) {
	var c map[string]any //connect.Connect
	err := ctx.BindJSON(&c)
	if err != nil {
		api.Error(ctx, err)
		return
	}

	id := c["id"].(string)
	if id == "" {
		api.Fail(ctx, "ID不能为空")
		return
	}
	file, err := os.OpenFile("connects/"+id+".json", os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		api.Error(ctx, err)
		return
	}
	defer file.Close()

	err = json.NewEncoder(file).Encode(c)
	if err != nil {
		api.Error(ctx, err)
		return
	}

	connect.LoadConnect(id)

	api.OK(ctx, nil)
}

func connectUpdate(ctx *gin.Context) {
	id := ctx.Param("id")

	var c map[string]any // connect.Connect
	err := ctx.BindJSON(&c)
	if err != nil {
		api.Error(ctx, err)
		return
	}

	file, err := os.OpenFile("connects/"+id+".json", os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		api.Error(ctx, err)
		return
	}
	defer file.Close()

	err = json.NewEncoder(file).Encode(c)
	if err != nil {
		api.Error(ctx, err)
		return
	}

	//connect.UnloadConnect(id)
	connect.LoadConnect(id)

	api.OK(ctx, nil)
}
