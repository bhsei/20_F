package wechat

import (
	"code.gitea.io/gitea/models"
	"code.gitea.io/gitea/modules/auth"
	"code.gitea.io/gitea/modules/context"
	"encoding/xml"
	"strconv"
	"strings"
)

const (
	SUBSCRIBE   = "subscribe"
	UNSUBSCRIBE = "unsubscribe"
)

// Event: subscribe/unsubscribe
type WatchEventForm struct {
	Self       string `xml:"ToUserName"`
	OpenId     string `xml:"FromUserName"`
	CreateTime int64
	MsgType    string
	Event      string
	EventKey   string
	Ticket     string
}

func WatchEventPost(ctx *context.Context) {
	data, err := ctx.Req.Body().Bytes()
	if err != nil {
		ctx.ServerError("WatchEventPost", err)
		return
	}
	var form WatchEventForm
	err = xml.Unmarshal(data, &form)
	if err != nil {
		ctx.ServerError("WatchEventPost", err)
		return
	}
	userid_str := strings.TrimPrefix(form.EventKey, "qrscene_")
	userid, err := strconv.ParseInt(userid_str, 10, 64)
	if err != nil {
		ctx.ServerError("WatchEventPost", err)
		return
	}
	user, err := models.GetUserByID(userid)
	if err != nil {
		ctx.ServerError("WatchEventPost", err)
		return
	}
	if form.Event == SUBSCRIBE {
		user.UpdateWechatOpenId(form.OpenId)
	} else if form.Event == UNSUBSCRIBE {
		user.UpdateWechatOpenId("")
	}
	ctx.RawData(200, []byte(""))
}

func UrlValidateGet(ctx *context.Context, form auth.ValidateForm) {
	//TODO: validate signature
	ctx.RawData(200, []byte(form.Echostr))
}
