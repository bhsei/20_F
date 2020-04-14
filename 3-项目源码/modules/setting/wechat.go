package setting

var (
	Wechat = struct {
		AppId     string
		AppSecret string
	}{}
)

func newWechatService() {
	sec := Cfg.Section("wechat")
	if !sec.Key("ENABLED").MustBool() {
		return
	}

	Wechat.AppId = sec.Key("APP_ID").String()
	Wechat.AppSecret = sec.Key("APP_SECRET").String()
}
