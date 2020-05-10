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

def get_setting(request, setting):
    module = request.module
    if module not in mod_server.module_list:
        resp = service_pb2.Resp(status = service_pb2.Resp.ERROR)
    resp = service_pb2.Resp(status = service_pb2.Resp.SUCCESS)
    data = mod_server.module_list[module][setting]
    return service_pb2.SettingResp(resp = resp, module = module, data = data)

def get_user_setting(module_name: str):
    if module_name not in mod_server.module_list:
        return []
    settings = mod_server.module_list[module_name]["userSetting"]

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
        return get_setting(request, "global_setting")

    def UserSettingRequest(self, request, context):
        return get_setting(request, "user_setting")

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

    def DisableModule(self, request, context):
        return service_pb2.Resp(status = service_pb2.Resp.SUCCESS)

    def SendMessage(self, request, context):
        title = request.body.title
        content = request.body.content
        url = request.body.url
        users = request.meta
        module_list = mod_server.module_list
        for user in users:
            modules = users[user]
            for module in modules:
                if module not in module_list:
                    continue
                settings = module_list[module]["userSetting"]
                settings = mod_server.db.db_query(user, settings)
                if settings is None:
                    continue
                module_list[module]["object"].send(title, content, url, settings)
        return service_pb2.MsgControlResp()
