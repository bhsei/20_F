from module_definition import ModuleAbstract, RedirectUrl
from Wechat.wechat_api import WechatAPI
import qrcode
import xmltodict
import io
import hashlib
import urllib.parse

class Wechat(ModuleAbstract):

    _API = None
    _qr_expire = 60

    def __init__(self, db_proxy, config):
        super(Wechat, self).__init__(db_proxy, config)

    def get_qr_code(self, form, data):
        content_type = "image/png"
        payload = b""
        if self._API is None:
            return content_type, payload
        url = self._API.get_qr_code(str(form["uid"]), self._qr_expire)
        if url:
            data = qrcode.make(url)
            target = io.BytesIO()
            data.save(target, format="PNG")
            payload = target.getvalue()
        return content_type, payload

    def validate(self, form, data):
        token = self.global_setting["token"]
        form = urllib.parse.parse_qs(form["param"])
        try:
            nonce = form["nonce"][0]
            timestamp = form["timestamp"][0]
            signature = form["signature"][0]
            echostr = form["echostr"][0]
        except KeyError:
            return "text/html", b""
        payload = [token, nonce, timestamp]
        payload.sort()
        sha1 = hashlib.sha1(bytes("".join(payload), "utf8"))
        sha1 = sha1.hexdigest()
        if sha1 != signature:
            return "text/html", b""
        return "text/html", bytes(echostr, "utf8")

    def unsubscribe(self, form):
        open_id = form["FromUserName"]
        ids = self.db_proxy.load({"open_id": open_id})
        for user in ids:
            self.db_proxy.store(user, {"open_id": ""})

    def subscribe(self, form):
        open_id = form["FromUserName"]
        uid = form["EventKey"][len("qrscene_"):]
        uid = int(uid)
        self.db_proxy.store(uid, {"open_id": open_id})

    def scan(self, form):
        open_id = form["FromUserName"]
        uid = form["EventKey"]
        uid = int(uid)
        self.db_proxy.store(uid, {"open_id": open_id})

    def watch_event_post(self, form, data):
        try:
            data = xmltodict.parse(data)
            data = data["xml"]
            event = data["Event"]
        except KeyError:
            return "text/html", b""

        if event == "SCAN":
            self.scan(data)
        elif event == "subscribe":
            self.subscribe(data)
        elif event == "unsubscribe":
            self.unsubscribe(data)

        return "text/html", b""

    def send(self, title, content, url, user_id):
        if self._API is None:
            return -1
        settings = self.db_proxy.load(user_id)
        if settings is None:
            return -1
        data = "{}\n{}".format(title, content)
        self._API.send(settings["open_id"], data, url)

    def global_setting_check(self, settings):
        super(Wechat, self).global_setting_check(settings)
        self._API = WechatAPI(settings["app_id"], settings["app_secret"])
        return True

    def get_redirect_urls(self):
        qrcode_handler = lambda f, d: self.get_qr_code(f, d)
        validate_handler = lambda f, d: self.validate(f, d)
        event_handler = lambda f, d: self.watch_event_post(f, d)
        return [RedirectUrl("/qrcode", RedirectUrl.URL_GET, qrcode_handler),
                RedirectUrl("/event", RedirectUrl.URL_GET, validate_handler),
                RedirectUrl("/event", RedirectUrl.URL_POST, event_handler)]
