package module

import (
	"bytes"
	"code.gitea.io/gitea/modules/base"
	"code.gitea.io/gitea/modules/context"
	"code.gitea.io/gitea/modules/log"
	"code.gitea.io/gitea/modules/setting"
	"code.gitea.io/gitea/modules/upload"
	module_service "code.gitea.io/gitea/services/module"
	"fmt"
	"gitea.com/macaron/csrf"
	gouuid "github.com/satori/go.uuid"
	"github.com/unknwon/com"
	"html/template"
	"strings"
	"sync"
)

const (
	tplModule              base.TplName = "admin/modules"
	tplUserModule          base.TplName = "user/settings/modules"
	attachmentAllowedTypes string       = "application/zip"
)

var (
	attachments sync.Map = sync.Map{}
)

type ModuleSpec struct {
	ModuleName     string
	ModuleTemplate string
	ModuleSetting  map[string]string
	ModuleSubUrl   string
	RedirectSubUrl string
	CsrfTokenHtml  template.HTML
}

func renderAttachmentSetting(ctx *context.Context) {
	ctx.Data["RequireDropzone"] = true
	ctx.Data["IsAttachmentEnabled"] = setting.AttachmentEnabled
	ctx.Data["AttachmentAllowedTypes"] = attachmentAllowedTypes
	ctx.Data["AttachmentMaxSize"] = setting.AttachmentMaxSize
	ctx.Data["AttachmentMaxFiles"] = setting.AttachmentMaxFiles
}

func UploadAttachments(ctx *context.Context) {
	file, _, err := ctx.Req.FormFile("file")
	if err != nil {
		ctx.Error(500, fmt.Sprintf("FormFile: %v", err))
		return
	}
	defer file.Close()
	buf := make([]byte, 0)
	for {
		tmp := make([]byte, 1024)
		n, _ := file.Read(tmp)
		if n > 0 {
			buf = append(buf, tmp[:n]...)
		} else {
			break
		}
	}
	err = upload.VerifyAllowedContentType(buf, strings.Split(attachmentAllowedTypes, ","))
	if err != nil {
		ctx.Error(400, err.Error())
		return
	}
	uuid := gouuid.NewV4().String()
	attachments.Store(uuid, buf)
	ctx.JSON(200, map[string]string{
		"uuid": uuid,
	})
}

func ModuleImport(ctx *context.Context) {
	name := ctx.Req.FormValue("files")
	file, ok := attachments.Load(name)
	if !ok {
		ctx.Flash.Error("No attachment selected")
		ctx.Redirect(setting.AppSubURL + "/admin/modules")
		return
	}
	attachments.Delete(name)
	data, _ := file.([]byte)
	ok = module_service.ModuleImport(data)
	if !ok {
		ctx.Flash.Error("Import module error")
		ctx.Redirect(setting.AppSubURL + "/admin/modules")
		return
	}
	ctx.Flash.Success("Import module succeed")
	ctx.Redirect(setting.AppSubURL + "/admin/modules")
	return
}

//TODO: add self-defined setting data
func SetModules(ctx *context.Context, x csrf.CSRF) {
	settings, ok := module_service.GlobalSettings()
	ctx.Data["Title"] = ctx.Tr("admin.modules")
	ctx.Data["PageIsAdmin"] = true
	ctx.Data["PageIsAdminModules"] = true
	renderAttachmentSetting(ctx)
	moduleSubUrl := setting.AppSubURL + "/module"
	csrfToken := x.GetToken()
	csrfTokenHtml := template.HTML(`<input type="hidden" name="_csrf" value="` + csrfToken + `">`)

	if !ok {
		log.Info("SetModule", settings)
		return
	}
	modules := make([]template.HTML, 0, 10)
	for module, setting := range settings {
		s := ModuleSpec{
			ModuleName:     module,
			ModuleSetting:  setting.OldSetting,
			ModuleSubUrl:   moduleSubUrl,
			RedirectSubUrl: moduleSubUrl + "/redirect/" + module,
			CsrfTokenHtml:  csrfTokenHtml,
		}
		parser, err := template.New(module).Parse(setting.Tmpl)
		if err != nil {
			log.Warn("SetModule parse %s error", module, err)
			continue
		}
		buffer := bytes.Buffer{}
		err = parser.Execute(&buffer, s)
		if err != nil {
			log.Warn("Set Module execute %s error", module, err)
			continue
		}
		modules = append(modules, template.HTML(buffer.String()))
	}
	ctx.Data["Modules"] = modules
	ctx.HTML(200, tplModule)
}

func UserSetModule(ctx *context.Context, x csrf.CSRF) {
	settings, ok := module_service.UserSettings(ctx.User.ID)
	ctx.Data["Title"] = ctx.Tr("settings")
	ctx.Data["PageIsSettingsModules"] = true
	moduleSubUrl := setting.AppSubURL + "/module"
	csrfToken := x.GetToken()
	csrfTokenHtml := template.HTML(`<input type="hidden" name="_csrf" value="` + csrfToken + `">`)

	if !ok {
		log.Info("UserSetModule", settings)
		return
	}
	modules := make([]template.HTML, 0, 10)
	for module, setting := range settings {
		s := ModuleSpec{
			ModuleName:     module,
			ModuleSetting:  setting.OldSetting,
			ModuleSubUrl:   moduleSubUrl,
			RedirectSubUrl: moduleSubUrl + "/redirect/" + module,
			CsrfTokenHtml:  csrfTokenHtml,
		}
		parser, err := template.New(module).Parse(setting.Tmpl)
		if err != nil {
			log.Warn("UserSetModule parse %s error", module, err)
			continue
		}
		buffer := bytes.Buffer{}
		err = parser.Execute(&buffer, s)
		if err != nil {
			log.Warn("UserSetModule execute %s error", module, err)
			continue
		}
		modules = append(modules, template.HTML(buffer.String()))
	}
	ctx.Data["Modules"] = modules
	ctx.HTML(200, tplUserModule)
}

func ModuleSettingCommit(ctx *context.Context) {
	if !ctx.User.IsAdmin {
		ctx.Flash.Error("Only Admin can set modules")
		return
	}
	form := ctx.Req.Form.Encode()
	module := ctx.Params(":module")
	log.Info("ModuleSettingCommit", form)
	ok := module_service.GlobalSetingCommit(module, form)
	if !ok {
		ctx.Flash.Error("Setting error")
	} else {
		ctx.Flash.Success("Setting succeed")
	}
	ctx.Redirect(setting.AppSubURL + "/admin/modules")
	return
}

func UserModuleSettingCommit(ctx *context.Context) {
	form := ctx.Req.Form.Encode()
	module := ctx.Params(":module")
	log.Info("UserModuleSettingCommit %d", ctx.User.ID, form)
	ok := module_service.UserSettingCommit(ctx.User.ID, module, form)
	if !ok {
		ctx.Flash.Error("User setting error")
	} else {
		ctx.Flash.Success("User setting succeed")
	}
	ctx.Redirect(setting.AppSubURL + "/user/settings/modules")
	return
}

func ModuleRedirect(ctx *context.Context) {
	form := map[string]string{}
	form["param"] = ctx.Req.Form.Encode()
	if ctx.User != nil {
		form["uid"] = com.ToStr(ctx.User.ID)
		form["admin"] = com.ToStr(ctx.User.IsAdmin)
	}
	body, err := ctx.Req.Body().Bytes()
	if err != nil {
		ctx.NotFound("", nil)
		return
	}
	m := ctx.Req.Method
	var method module_service.ReqType
	if m == "GET" {
		method = module_service.GET
	} else if m == "POST" {
		method = module_service.POST
	} else {
		ctx.NotFound("", nil)
		return
	}
	url := ctx.Req.URL.Path
	url = strings.TrimPrefix(url, "/module/redirect")
	log.Info("Module redirect", url)
	contentType, payload, ok := module_service.UrlRedirectRequest(form, body, url, method)
	if !ok {
		log.Info("%s for %s Not Found", url, method)
		ctx.NotFound("", nil)
		return
	}
	ctx.Resp.Header().Set("Content-Type", contentType)
	ctx.Write(payload)
}
