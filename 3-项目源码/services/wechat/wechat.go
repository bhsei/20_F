package wechat

import (
	"code.gitea.io/gitea/modules/setting"
)

func SendWechatMsg(message, openid, url string) bool {
	return true
}

func GetQRCode(data string, expireSeconds int) string {
	return ""
}
