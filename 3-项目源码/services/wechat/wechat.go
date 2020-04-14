package wechat

import "code.gitea.io/gitea/modules/setting"

type Message struct {
	openid  string
	content string
	url     string
}

// SendWechatMsg 将消息放到发送队列中，异步发送微信通知
func SendWechatMsg(message, openid, url string) bool {
	msg := &Message{
		openid:  openid,
		content: message,
		url:     url,
	}

	go func() {
		messageQueue <- msg
	}()

	return true
}

// GetQRCode 生成扫描后重定向到服务号的二维码, 返回二维码的内容
// data 为二维码携带的数据，微信用户扫描二维码后微信会将扫描事件和二维码携带的数据发送给Gitea服务器
// expireSeconds 为二维码有效期
func GetQRCode(data string, expireSeconds int) string {
	token := getCachedToken(setting.Wechat.AppId, setting.Wechat.AppSecret)
	qrLink := getQRCode(token, data, expireSeconds)
	return qrLink
}
