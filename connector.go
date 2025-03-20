package connector

import (
	"embed"
	"github.com/busy-cloud/boat/menu"
	"github.com/busy-cloud/boat/page"
	_ "github.com/busy-cloud/connector/boot"
)

//go:embed pages
var pages embed.FS

//go:embed menus
var menus embed.FS

func init() {
	//注册页面
	page.EmbedFS(pages, "pages")

	//注册菜单
	menu.EmbedFS(menus, "menus")
}
