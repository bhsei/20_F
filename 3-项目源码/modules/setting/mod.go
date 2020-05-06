package setting

import (
	"path"
)

const (
	mod_server_bin_name = "mod_server.py"
)

var (
	Mod = struct {
		RootPath       string
		ServiceBinPath string
	}{}
)

func newModService() {
	sec := Cfg.Section("mod")
	Mod.RootPath = sec.Key("ROOT_PATH").MustString(path.Join(AppWorkPath, "mod"))
	appDir, _ := path.Split(AppPath)
	Mod.ServiceBinPath = sec.Key("SERVICE_BIN_PATH").MustString(path.Join(appDir, mod_server_bin_name))
}
