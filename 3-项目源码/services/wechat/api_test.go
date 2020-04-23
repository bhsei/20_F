package wechat

/*
测试方式:
APPID="xxx" APPSECRET="xxx" OPENID="xxx" go test -v
*/

import (
	"github.com/stretchr/testify/assert"
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
	assert := assert.New(t)

	t.Logf("Use appID:%s appSecret:%s", appID, appSecret)
	token, _, err := getToken(appID, appSecret)
	assert.Equal(err, nil, "getToken() error")
	t.Logf("Use access token:%.10s...", token)

	qrStr, err := getQRCode(token, "test", 3600)
	assert.Equal(err, nil, "getQRCode() error")

	t.Logf("getQRCode -> %s", qrStr)
}

func TestSendWechatMsg(t *testing.T) {
	assert := assert.New(t)

	t.Logf("Use appID:%s appSecret:%s", appID, appSecret)
	token, _, err := getToken(appID, appSecret)
	assert.Equal(err, nil, "getToken() error")
	t.Logf("Use access token:%.10s...", token)

	err = sendMessage(token, openID, "Hello", "https://baidu.com")
	assert.Equal(err, nil, "sendMessage() error")
	t.Logf("sendMessage to %s", openID)
}

func TestGetCachedToken(t *testing.T) {
	assert := assert.New(t)

	t.Logf("Use appID:%s appSecret:%s", appID, appSecret)
	token, err := getCachedToken(appID, appSecret)
	assert.Equal(err, nil, "getCachedToken() error")

	t.Logf("Sleep some seconds...")
	time.Sleep(8 * time.Second)
	token2, err := getCachedToken(appID, appSecret)
	assert.Equal(err, nil, "getCachedToken() error")
	if token != token2 {
		t.Errorf("两次获取到的缓存不一致")
	}
}

func TestUserInfo(t *testing.T) {
	assert := assert.New(t)

	t.Logf("Use appID:%s appSecret:%s", appID, appSecret)
	token, _, err := getToken(appID, appSecret)
	assert.Equal(err, nil, "getToken() error")
	t.Logf("Use access token:%.10s...", token)

	t.Logf("Use openID:%s ", openID)

	info, err := userInfo(token, openID)
	assert.Equal(err, nil, "userInfo() error")

	t.Logf("User info: %+v", info)
}
