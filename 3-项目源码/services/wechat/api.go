package wechat

// 微信服务号api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

var httpClient http.Client

func init() {
	httpClient = http.Client{
		Timeout: 8 * time.Second,
	}
}

// 获取微信access token
func getToken(appID, appSecret string) (accessToken string, expire int) {
	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s", appID, appSecret)
	res, err := httpClient.Get(url)
	if err != nil {
		panic(err)
	}
	defer func() { _ = res.Body.Close() }()
	body, _ := ioutil.ReadAll(res.Body)

	type Response struct {
		AccessToken string `json:"access_token"`
		Expires     int    `json:"expires_in"`
	}
	var msg Response
	err = json.Unmarshal(body, &msg)
	if err != nil {
		panic(err)
	}

	return msg.AccessToken, msg.Expires
}

var cachedAccessToken string // 缓存的微信access token
var expireTime int64         // 缓存的access token的失效时间

// 获取缓存的微信access token
func getCachedToken(appID, appSecret string) string {
	if expireTime < time.Now().Unix() {
		accessToken, expire := getToken(appID, appSecret)
		cachedAccessToken = accessToken
		expireTime = time.Now().Unix() + int64(expire) - 100
	}

	return cachedAccessToken
}

func sendMessage(accessToken, openID, message, url string) {
	api := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/message/template/send?access_token=%s", accessToken)
	bodyJson, _ := json.Marshal(map[string]interface{}{
		"touser":      openID,
		"template_id": "K2lk62bRKlzrhPB57D-hf2aHrBtqVL1-mIDYqHxoBG8", // 微信测试模板ID, 模板内容 {{data.DATA}}
		"url":         url,
		"data": map[string]interface{}{
			"data": map[string]interface{}{
				"value": message,
			},
		},
	})

	res, err := httpClient.Post(api, "application/json", strings.NewReader(string(bodyJson)))
	if err != nil {
		panic(err)
	}
	defer func() { _ = res.Body.Close() }()
	body, _ := ioutil.ReadAll(res.Body)

	var v interface{}
	json.Unmarshal(body, &v)
	data := v.(map[string]interface{})

	fmt.Printf("body: %s", body)

	if code, ok := data["errcode"]; ok && code.(float64) != 0 {
		panic(fmt.Sprintf("Error in send wechat message: %s[%.0f]", data["errmsg"], data["errcode"]))
	}
}

/*
生成带参数二维码
expire	该二维码有效时间，以秒为单位。 最大不超过2592000（即30天）
data 参数，长度限制为1到64
*/
func getQRCode(accessToken, data string, expire int) string {
	api := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/qrcode/create?access_token=%s", accessToken)
	bodyJson, _ := json.Marshal(map[string]interface{}{
		"expire_seconds": expire,
		"action_name":    "QR_STR_SCENE",
		"action_info": map[string]interface{}{
			"scene": map[string]interface{}{
				"scene_str": data,
			},
		},
	})

	res, err := httpClient.Post(api, "application/json", strings.NewReader(string(bodyJson)))
	if err != nil {
		panic(err)
	}
	defer func() { _ = res.Body.Close() }()
	body, _ := ioutil.ReadAll(res.Body)

	var v interface{}
	json.Unmarshal(body, &v)
	respData := v.(map[string]interface{})

	url, ok := respData["url"]
	if !ok {
		panic(fmt.Sprintf("Error in get wechat qrcode: %s[%.0f]", respData["errmsg"], respData["errcode"]))
	}

	return url.(string)
}
