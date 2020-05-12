package module

import (
	"bytes"
	"code.gitea.io/gitea/modules/base"
	"code.gitea.io/gitea/modules/context"
	"code.gitea.io/gitea/modules/log"
	"code.gitea.io/gitea/modules/notification"
	"code.gitea.io/gitea/modules/setting"
	"gitea.com/macaron/csrf"
	"html/template"
)

const (
	tplModule base.TplName = "admin/modules"
)

type ModuleSpec struct {
	ModuleName     string
	ModuleTemplate string
	ModuleSetting  map[string]string
	ModuleSubUrl   string
	CsrfTokenHtml  template.HTML
}

//TODO: add self-defined setting data
func SetModules(ctx *context.Context, x csrf.CSRF) {
	settings, ok := notification.GlobalSettings()
	ctx.Data["PageIsAdmin"] = true
	ctx.Data["PageIsAdminModules"] = true
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
			ModuleName:    module,
			ModuleSetting: map[string]string{},
			ModuleSubUrl:  moduleSubUrl,
			CsrfTokenHtml: csrfTokenHtml,
		}
		parser, err := template.New(module).Parse(setting)
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

func ModuleSettingCommit(ctx *context.Context) {
	if !ctx.User.IsAdmin {
		ctx.Flash.Error("Only Admin can set modules")
		return
	}
	form := ctx.Req.Form.Encode()
	module := ctx.Params(":module")
	log.Info("ModuleSettingCommit", form)
	ok := notification.GlobalSetingCommit(module, form)
	if !ok {
		ctx.Flash.Error("Setting error")
	} else {
		ctx.Flash.Success("Setting suceed")
	}
	ctx.Redirect(setting.AppSubURL + "/admin/modules")
	return
}
