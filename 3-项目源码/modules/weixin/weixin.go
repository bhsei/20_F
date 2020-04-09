package weixin

// weixin 包提供了对微信服务号可进行的操作

var appID string
var appSecret string

// Init 函数使用提供的微信服务号的appid和appsecret初始化weixin包
func Init(appid string, appsecret string) {
	appID = appid
	appSecret = appsecret
}

// SendWeixinMsg 函数向openid的微信用户发送内容为message的消息，并且可选提供消息点击后跳转的url
func SendWeixinMsg(message string, openid string, url string) bool {
	return true
}

// GetQRCode 函数获取带参数(data)的二维码，返回二维码的Url，url在 expireSeconds 秒内有效。
// 微信用户扫描二维码后，可以关注服务号，微信会将带参数(data)的事件推送给Gitea后台
func GetQRCode(data string, expireSeconds int) string {
	return ""
}
