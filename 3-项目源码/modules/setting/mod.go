package setting

var (
	Mod = struct {
		Host string
		Port int
	}{}
)

func newModService() {
	sec := Cfg.Section("mod")
	Mod.Host = sec.Key("HOST").MustString("localhost")
	Mod.Port = sec.Key("PORT").MustInt(3001)
}
