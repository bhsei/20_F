import service_pb2_grpc
import service_pb2
import grpc
import mod_redirect
import mod_upload
import mod_server
from typing import List, Tuple

def redirect_urls_transform(urls):
    import_resp = service_pb2.ModuleImportResp
    def tr(url):
        url_id = url[1]
        pattern = url[0].url_pattern
        if url[0].url_type == mod_redirect.RedirectUrl.URL_GET:
            url_type = import_resp.UrlRedirect.GET
        else:
            url_type = import_resp.UrlRedirect.POST
        return import_resp.UrlRedirect(id = url_id, pattern = pattern, url_type = url_type)
    return list(map(tr, urls))

def get_setting_tmpls(setting_name):
    module_list = mod_server.module_list
    data = map(lambda m: (m, module_list[m][setting_name]), module_list.keys())
    data = map(lambda m: service_pb2.SettingResp.Setting(module = m[0], data = m[1]), data)
    resp = service_pb2.Resp(status = service_pb2.Resp.SUCCESS)
    return service_pb2.SettingResp(resp = resp, settings = list(data))

def get_setting_tmpl(request, setting):
    module = request.module
    if module not in mod_server.module_list:
        resp = service_pb2.Resp(status = service_pb2.Resp.ERROR)
    resp = service_pb2.Resp(status = service_pb2.Resp.SUCCESS)
    data = mod_server.module_list[module][setting]
    return service_pb2.SettingResp(resp = resp, module = module, data = data)

def get_user_setting(module_name: str):
    if module_name not in mod_server.module_list:
        return []
    settings = mod_server.module_list[module_name]["user_setting"]

def get_user_module_settings(user_id):
    module_list = mod_server.module_list
    settings = mod_server.db.db_query(user_id)
    def collect(module):
        targets = module_list[module]["user_setting"]
        ret = {}
        for target in targets:
            ret[target] = setting[target]
        return module, ret

    def module_set(module, settings):
        for value in settings.values():
            if len(value) == 0:
                return False
        return True
    modules = map(collect, module_list.keys())
    valids = filter(lambda m: module_set(m[0], m[1]), modules)
    return list(valids)

class NotifyService(service_pb2_grpc.NotifyServiceServicer):

    def Init(self, request, context):
        urls = redirect_urls_transfor(mod_redirect.get_urls())
        resp = service_pb2.Resp(status = service_pb2.Resp.SUCCESS)
        return service_pb2.ModuleImportResp(resp = resp, redirect = urls)

    def ModuleImport(self, request, context):
        ok, name = mod_upload.extract_module(request.data, mod_server.ROOT_PATH)
        ret = service_pb2.Resp.ERROR
        if ok and mod_server.load_module(name, {}):
            urls = mod_server.module_list[name]["object"].get_redirect_urls()
            urls = redirect_urls_transform(urls)
            ret = service_pb2.Resp.SUCCESS
            return service_pb2.ModuleImportResp(resp = service_pb2.Resp(status = ret), redirect = urls)
        return service_pb2.ModuleImportResp(resp = service_pb2.Resp(status = ret))

    def GlobalSettingRequest(self, request, context):
        return get_setting_tmpls("global_setting")

    def UserSettingRequest(self, request, context):
        return get_setting_tmpls("user_setting")

    def Redirect(self, request, context):
        rid = request.id
        form = dict(request.form)
        data = bytes(request.data)
        url = mod_redirect.get_url_by_id(rid)
        if url is not None:
            content_type, payload = url.handler(form, data)
            return service_pb2.RedirectResp(
                    resp = service_pb2.Resp(status = service_pb2.Resp.SUCCESS),
                    content_type = content_type,
                    data = payload)
        return service_pb2.RedirectResp(
                resp = service_pb2.Resp(status = service_pb2.Resp.ERROR))

    def SendMessage(self, request, context):
        title = request.body.title
        content = request.body.content
        url = request.body.url
        users = request.ids
        for user in users:
            modules = get_user_module_settings(user)
            for module, setting in modules:
                mod_server.module_list[module]["object"].send(title, content, url, setting)
        return service_pb2.MsgControlResp()
