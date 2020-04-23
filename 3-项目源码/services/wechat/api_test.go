package wechat

/*
测试方式:
APPID="xxx" APPSECRET="xxx" OPENID="xxx" go test -v
*/

import (
	"os"
	"testing"
	"time"
)

var appID, appSecret, openID string

func init() {
	appID = os.Getenv("APPID")
	appSecret = os.Getenv("APPSECRET")
	openID = os.Getenv("OPENID")
}

func TestGetQRCode(t *testing.T) {
	t.Logf("Use appID:%s appSecret:%s", appID, appSecret)
	token, _ := getToken(appID, appSecret)
	t.Logf("Use access token:%.10s...", token)

	qrStr := getQRCode(token, "test", 3600)
	t.Logf("getQRCode -> %s", qrStr)
}

func TestSendWechatMsg(t *testing.T) {
	t.Logf("Use appID:%s appSecret:%s", appID, appSecret)
	token, _ := getToken(appID, appSecret)
	t.Logf("Use access token:%.10s...", token)

	sendMessage(token, openID, "Hello", "https://baidu.com")
	t.Logf("sendMessage to %s", openID)
}

func TestGetCachedToken(t *testing.T) {
	t.Logf("Use appID:%s appSecret:%s", appID, appSecret)
	token := getCachedToken(appID, appSecret)
	t.Logf("Sleep some seconds...")
	time.Sleep(8 * time.Second)
	if token != getCachedToken(appID, appSecret) {
		t.Errorf("两次获取到的缓存不一致")
	}
}

func TestUserInfo(t *testing.T) {
	token, _ := getToken(appID, appSecret)
	t.Logf("Use access token:%.10s...", token)

	t.Logf("Use openID:%s ", openID)

	info := userInfo(token, openID)

	t.Logf("User info: %+v", info)
}
