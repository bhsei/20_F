package setting

import (
	"code.gitea.io/gitea/modules/log"
)

var (
	Module = struct {
		Enabled bool
		Host    string
		Port    int
	}{}
)

func newModuleService() {
	sec := Cfg.Section("module")
	Module.Enabled = sec.Key("ENABLED").MustBool(false)
	if !Module.Enabled {
		log.Info("Module disabled")
		return
	}
	Module.Host = sec.Key("HOST").MustString("localhost")
	Module.Port = sec.Key("PORT").MustInt(3001)
}
