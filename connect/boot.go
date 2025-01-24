package connect

import (
	"github.com/busy-cloud/boat/boot"
	"github.com/busy-cloud/boat/log"
	"os"
	"path/filepath"
	"strings"
)

func init() {
	boot.Register("connect", &boot.Task{
		Startup:  Startup,
		Shutdown: nil,
		Depends:  []string{"log"},
	})
}

func Startup() error {
	_ = os.MkdirAll("connects", 0777)
	files, err := filepath.Glob("connects/*.json")
	if err != nil {
		return err
	}
	for _, f := range files {
		//log.Println(filepath.Base(f))
		id, _ := strings.CutSuffix(filepath.Base(f), ".json")
		err = LoadConnect(id)
		if err != nil {
			//e = multierr.Append(e, err)
			log.Error(err)
		}
	}

	subscribe()
	
	return nil
}
