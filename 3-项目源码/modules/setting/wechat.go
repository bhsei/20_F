package setting

var (
	Wechat = struct {
		Enable        bool
		AppId         string
		AppSecret     string
		ExpireSeconds int
		QrcodeSize    int
	}{}
)

func newWechatService() {
	sec := Cfg.Section("wechat")
	if !sec.Key("ENABLED").MustBool() {
		return
	}

	Wechat.Enable = sec.Key("ENABLED").MustBool()
	Wechat.AppId = sec.Key("APP_ID").String()
	Wechat.AppSecret = sec.Key("APP_SECRET").String()
	Wechat.ExpireSeconds = sec.Key("EXPIRE_SECONDS").MustInt(120)
	Wechat.QrcodeSize = sec.Key("QRCODE_SIZE").MustInt(256)
}
