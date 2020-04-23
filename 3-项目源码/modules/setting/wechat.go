package setting

import (
	"code.gitea.io/gitea/modules/log"
)

var (
	Wechat = struct {
		Enable        bool
		AppId         string
		AppSecret     string
		ExpireSeconds int
		QrcodeSize    int
		Token         string
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
	Wechat.Token = sec.Key("TOKEN").String()
}

func newNotifyWechatService() {
	if !Wechat.Enable || !Cfg.Section("service").Key("ENABLE_NOTIFY_WECHAT").MustBool() {
		return
	}
	Service.EnableNotifyWechat = true
	log.Info("Notify Wechat Service Enabled")
}
