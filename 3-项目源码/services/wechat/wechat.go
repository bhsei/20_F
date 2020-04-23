package wechat

import "code.gitea.io/gitea/modules/setting"

// 微信用户信息
type UserInfo struct {
	Subscribe     int    `json:"subscribe"`      // 用户是否订阅该公众号标识，值为0时，代表此用户没有关注该公众号，其他字段为空
	Openid        string `json:"openid"`         // 用户的标识，对当前公众号唯一
	Nickname      string `json:"nickname"`       // 用户的昵称
	Sex           int    `json:"sex"`            // 用户的性别，值为1时是男性，值为2时是女性，值为0时是未知
	Language      string `json:"language"`       // 用户的语言，简体中文为zh_CN
	City          string `json:"city"`           // 用户所在城市
	Province      string `json:"province"`       // 用户所在省份
	Country       string `json:"country"`        // 用户所在国家
	HeadImgUrl    string `json:"headimgurl"`     // 用户头像. 用户没有头像时该项为空。若用户更换头像，原有头像URL将失效
	SubscribeTime int64  `json:"subscribe_time"` // 用户关注时间，为时间戳。如果用户曾多次关注，则取最后关注时间
	QrSceneStr    string `json:"qr_scene_str"`   // 关注服务号时扫描的二维码中携带的数据. 即 GetQRCode() 中的 data
}

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
func GetQRCode(data string, expireSeconds int) (qrcodeContent string, err error) {
	token, err := getCachedToken(setting.Wechat.AppId, setting.Wechat.AppSecret)
	if err == nil {
		qrcodeContent, err = getQRCode(token, data, expireSeconds)
	}
	return qrcodeContent, err
}

// UserInfo 获取微信用户信息
func GetUserInfo(openid string) (info UserInfo, err error) {
	token, err := getCachedToken(setting.Wechat.AppId, setting.Wechat.AppSecret)
	if err == nil {
		info, err = userInfo(token, openid)
	}
	return info, err
}
