package connector

import (
	"embed"
	"encoding/json"
	"github.com/busy-cloud/boat/apps"
	"github.com/busy-cloud/boat/log"
	_ "github.com/busy-cloud/connector/internal"
)

//go:embed assets
var assets embed.FS

//go:embed manifest.json
var manifest []byte

func init() {
	//注册为内部插件
	var a apps.App
	err := json.Unmarshal(manifest, &a)
	if err != nil {
		log.Fatal(err)
	}
	apps.Register(&a)

	apps.Assets().EmbedFS(a.Id, assets, "assets")
}
