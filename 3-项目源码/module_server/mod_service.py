import service_pb2_grpc
import service_pb2
from mod_redirect import ModuleRedirect
from mod_manage import ModuleManage
from module_definition import RedirectUrl
import urllib.parse

def redirect_urls_transform(urls):
    import_resp = service_pb2.ModuleImportResp

    def tr(url):
        url_id = url[1]
        pattern = url[0].url_pattern
        if url[0].url_type == RedirectUrl.URL_GET:
            url_type = import_resp.UrlRedirect.GET
        else:
            url_type = import_resp.UrlRedirect.POST
        return import_resp.UrlRedirect(id=url_id, pattern=pattern, url_type=url_type)
    return list(map(tr, urls))

def get_setting_tmpls(tmpls):
    def F(m, t, o):
        return service_pb2.SettingResp.Setting(module = m, data = t, old_settings = o)
    settings = map(lambda x: F(x[0], x[1], x[2]), tmpls)
    status = service_pb2.Resp.SUCCESS
    resp = service_pb2.Resp(status = status)
    return service_pb2.SettingResp(resp = resp, settings = list(settings))

class NotifyService(service_pb2_grpc.NotifyServiceServicer):

    _module_manager: "ModuleManage" = None
    _redirect_manager: "ModuleRedirect" = None

    def __init__(self, module_manager: "ModuleManage", redirect_manager: "ModuleRedirect"):
        self._module_manager = module_manager
        self._redirect_manager = redirect_manager

    def Init(self, request, context):
        print("Init called")
        urls = self._redirect_manager.get_urls()
        urls = redirect_urls_transform(urls)
        resp = service_pb2.Resp(status=service_pb2.Resp.SUCCESS)
        return service_pb2.ModuleImportResp(resp=resp, redirect=urls)

    def ModuleImport(self, request, context):
        print("ModuleImport called")
        urls = self._module_manager.load_module(request.data)
        ret = service_pb2.Resp.ERROR
        if urls is not None:
            urls = redirect_urls_transform(urls)
            ret = service_pb2.Resp.SUCCESS
            return service_pb2.ModuleImportResp(resp=service_pb2.Resp(status=ret), redirect=urls)
        return service_pb2.ModuleImportResp(resp=service_pb2.Resp(status=ret))

    def GlobalSettingRequest(self, request, context):
        print("GlobalSettingRequest called")
        return get_setting_tmpls(self._module_manager.global_tmpls())

    def UserSettingRequest(self, request, context):
        print("UserSettingRequest called")
        user = request.user
        return get_setting_tmpls(self._module_manager.user_tmpls(user))

    def Redirect(self, request, context):
        print("Redirect called")
        rid = request.id
        form = dict(request.form)
        data = bytes(request.data)
        url = self._redirect_manager.get_url_by_id(rid)
        if url is not None:
            content_type, payload = url.handler(form, data)
            return service_pb2.RedirectResp(
                resp=service_pb2.Resp(status=service_pb2.Resp.SUCCESS),
                content_type=content_type,
                data=payload)
        return service_pb2.RedirectResp(
            resp=service_pb2.Resp(status=service_pb2.Resp.ERROR))

    def SendMessage(self, request, context):
        print("SendMessage called")
        title = request.body.title
        content = request.body.content
        url = request.body.url
        users = request.ids
        self._module_manager.send(title, content, url, users)
        return service_pb2.MsgControlResp()

    def GlobalSettingCommit(self, request, context):
        print("GlobalSettingCommit called")
        module = request.module
        encode_form = request.encode_form
        form = urllib.parse.parse_qs(encode_form)
        for key in form.keys():
            form[key] = form[key][0]
        ok = self._module_manager.add_global_setting(module, form)
        status = service_pb2.Resp.SUCCESS
        if not ok:
            status = service_pb2.Resp.ERROR
        return service_pb2.Resp(status = status)

    def UserSettingCommit(self, request, context):
        print("UserSettingCommit called")
        module = request.req.module
        encode_form = request.req.encode_form
        uid = request.user
        form = urllib.parse.parse_qs(encode_form, keep_blank_values = True)
        for key in form.keys():
            form[key] = form[key][0]
        ok = self._module_manager.add_user_setting(module, uid, form)
        status=service_pb2.Resp.ERROR
        if ok:
            status = service_pb2.Resp.SUCCESS
        return service_pb2.Resp(status=status)
