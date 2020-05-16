package module

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

type SettingPack struct {
	Tmpl       string
	OldSetting map[string]string
}

func getConnect() (conn *grpc.ClientConn, err error) {
	addr := fmt.Sprintf("%s:%d", setting.Module.Host, setting.Module.Port)
	conn, err = grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		log.Error("gRPC getConnect: %v", err)
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
		log.Error("gRPC SendMessage %v", err)
		return
	}
}

//TODO: remove registered urls if something error happens at middle
func registerUrls(redirects []*ModuleImportResp_UrlRedirect) error {
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
			return err
		}
	}
	return nil
}

// Init: initialize module subsystem. Disable it if fail
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
		log.Error("gRPC Init %v", r.GetResp().GetDetail())
		setting.Module.Enabled = false
		return
	}
	err = registerUrls(r.GetRedirect())
	if err != nil {
		log.Error("gRPC Init %v", err)
		setting.Module.Enabled = false
		return
	}
	return
}

func ModuleImport(data []byte) bool {
	conn, err := getConnect()
	if err != nil {
		return false
	}
	defer conn.Close()
	c := NewNotifyServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeOut)
	defer cancel()
	r, err := c.ModuleImport(ctx, &ModuleImportReq{
		Data: data,
	})
	if err != nil {
		log.Error("gRPC ModuleImport %v", err)
		return false
	}
	if r.GetResp().GetStatus() != Resp_SUCCESS {
		log.Warn("gRPC ModuleImport %v", r.GetResp().GetDetail())
		return false
	}
	err = registerUrls(r.GetRedirect())
	if err != nil {
		log.Warn("gRPC ModuleImport %v", err)
		return false
	}
	return true
}

func GlobalSettings() (gs map[string]SettingPack, ok bool) {
	gs = map[string]SettingPack{}
	conn, err := getConnect()
	ok = false
	if err != nil {
		return
	}
	defer conn.Close()
	c := NewNotifyServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeOut)
	defer cancel()
	r, err := c.GlobalSettingRequest(ctx, &Req{})
	if err != nil {
		log.Error("gRPC GlobalSettings %v", err)
		return
	}
	if r.GetResp().GetStatus() != Resp_SUCCESS {
		log.Warn("gRPC GlobalSettings %s", r.GetResp().GetDetail())
		return
	}
	s := r.GetSettings()
	for _, item := range s {
		gs[item.GetModule()] = SettingPack{
			Tmpl:       item.GetData(),
			OldSetting: item.GetOldSettings(),
		}
	}
	ok = true
	return
}

func UserSettings(user_id int64) (us map[string]SettingPack, ok bool) {
	us = map[string]SettingPack{}
	conn, err := getConnect()
	ok = false
	if err != nil {
		return
	}
	defer conn.Close()
	c := NewNotifyServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeOut)
	defer cancel()
	r, err := c.UserSettingRequest(ctx, &Req{
		User: user_id,
	})
	if err != nil {
		log.Error("gRPC UserSettings %v", err)
		return
	}
	if r.GetResp().GetStatus() != Resp_SUCCESS {
		log.Warn("gRPC UserSettings %s", r.GetResp().GetDetail())
		return
	}
	s := r.GetSettings()
	for _, item := range s {
		us[item.GetModule()] = SettingPack{
			Tmpl:       item.GetData(),
			OldSetting: item.GetOldSettings(),
		}
	}
	ok = true
	return
}

func GlobalSetingCommit(module string, form string) bool {
	conn, err := getConnect()
	if err != nil {
		return false
	}
	defer conn.Close()
	c := NewNotifyServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeOut)
	defer cancel()
	r, err := c.GlobalSettingCommit(ctx, &SettingReq{
		Module:     module,
		EncodeForm: form,
	})
	if err != nil {
		log.Error("gRPC GlobalSettingCommit %s %v", module, err)
		return false
	}
	if r.GetStatus() != Resp_SUCCESS {
		log.Warn("gRPC GlobalSettingCommit %s %s", module, r.GetDetail())
		return false
	}
	return true
}

func UserSettingCommit(uid int64, module string, form string) bool {
	conn, err := getConnect()
	if err != nil {
		return false
	}
	defer conn.Close()
	c := NewNotifyServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeOut)
	defer cancel()
	r, err := c.UserSettingCommit(ctx, &UserSettingReq{
		Req: &SettingReq{
			Module:     module,
			EncodeForm: form,
		},
		User: uid,
	})
	if err != nil {
		log.Error("gRPC UserSettingCommit %s %v", module, err)
		return false
	}
	if r.GetStatus() != Resp_SUCCESS {
		log.Warn("gRPC UserSettingCommit %s %s", module, r.GetDetail())
		return false
	}
	return true
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
		log.Error("gRPC Redirect %v", err)
		return
	}
	if r.GetResp().GetStatus() != Resp_SUCCESS {
		log.Warn("gRPC Redirect %s", r.GetResp().GetDetail())
		return
	}
	content_type = r.GetContentType()
	load = r.GetData()
	ok = true
	return
}
