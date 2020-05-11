package notification

import (
	"code.gitea.io/gitea/modules/log"
	"code.gitea.io/gitea/modules/setting"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"time"
)

const (
	DefaultTimeOut = time.Second
)

func getConnect() (conn *grpc.ClientConn, err error) {
	addr := fmt.Sprintf("%s:%d", setting.Mod.Host, setting.Mod.Port)
	conn, err = grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		log.Error("gRPC getConnect", err)
	}
	return
}

func SendMessage(title, content, url string, users []int64) {
	conn, err := getConnect()
	if err != nil {
		return
	}
	defer conn.Close()
	c := NewNotifyServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeOut)
	defer cancel()
	_, err = c.SendMessage(ctx, &MsgControl{
		Ids: users,
		Body: &MsgControl_MsgBody{
			Title:   title,
			Content: content,
			Url:     url,
		},
	})
	if err != nil {
		log.Error("gRPC SendMessage", err)
		return
	}
}

func registerUrls(redirects []*ModuleImportResp_UrlRedirect) {
	for _, redirect := range redirects {
		id := redirect.GetId()
		pattern := redirect.GetPattern()
		var reqtype ReqType
		if redirect.GetUrlType() == ModuleImportResp_UrlRedirect_GET {
			reqtype = GET
		} else if redirect.GetUrlType() == ModuleImportResp_UrlRedirect_POST {
			reqtype = POST
		}
		err := UrlRegister(id, pattern, reqtype)
		if err != nil {
			log.Error("gRPC Init", err)
		}
	}
}

func Init() {
	conn, err := getConnect()
	if err != nil {
		return
	}
	defer conn.Close()
	c := NewNotifyServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeOut)
	defer cancel()
	r, err := c.Init(ctx, &Req{})
	if r.GetResp().GetStatus() != Resp_SUCCESS {
		log.Warn("gRPC Init", r.GetResp().GetDetail())
		return
	}
	registerUrls(r.GetRedirect())
}

func ModuleImport(data []byte) {
	conn, err := getConnect()
	if err != nil {
		return
	}
	defer conn.Close()
	c := NewNotifyServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeOut)
	defer cancel()
	r, err := c.ModuleImport(ctx, &ModuleImportReq{
		Data: data,
	})
	if err != nil {
		log.Error("gRPC ModuleImport", err)
		return
	}
	if r.GetResp().GetStatus() != Resp_SUCCESS {
		log.Warn("gRPC ModuleImport", r.GetResp().GetDetail())
		return
	}
	registerUrls(r.GetRedirect())
}

func GlobalSettings() (gs map[string]string) {
	gs = map[string]string{}
	conn, err := getConnect()
	if err != nil {
		return
	}
	defer conn.Close()
	c := NewNotifyServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeOut)
	defer cancel()
	r, err := c.GlobalSettingRequest(ctx, &Req{})
	if err != nil {
		log.Error("gRPC GlobalSettings", err)
		return
	}
	if r.GetResp().GetStatus() != Resp_SUCCESS {
		log.Warn("gRPC GlobalSettings", r.GetResp().GetDetail())
		return
	}
	s := r.GetSettings()
	for _, item := range s {
		gs[item.GetModule()] = item.GetData()
	}
	return
}

func UserSettings() (us map[string]string) {
	us = map[string]string{}
	conn, err := getConnect()
	if err != nil {
		return
	}
	defer conn.Close()
	c := NewNotifyServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeOut)
	defer cancel()
	r, err := c.UserSettingRequest(ctx, &Req{})
	if err != nil {
		log.Error("gRPC UserSettings", err)
		return
	}
	if r.GetResp().GetStatus() != Resp_SUCCESS {
		log.Warn("gRPC UserSettings", r.GetResp().GetDetail())
		return
	}
	s := r.GetSettings()
	for _, item := range s {
		us[item.GetModule()] = item.GetData()
	}
	return
}

func Redirect(form map[string]string, data []byte, id int64) (content_type string, load []byte, ok bool) {
	conn, err := getConnect()
	ok = false
	if err != nil {
		return
	}
	defer conn.Close()
	c := NewNotifyServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeOut)
	defer cancel()
	r, err := c.Redirect(ctx, &RedirectData{
		Id:   id,
		Form: form,
		Data: data,
	})
	if err != nil {
		log.Error("gRPC Redirect", err)
		return
	}
	if r.GetResp().GetStatus() != Resp_SUCCESS {
		log.Warn("gRPC Redirect", r.GetResp().GetDetail())
		return
	}
	content_type = r.GetContentType()
	load = r.GetData()
	ok = true
	return
}
