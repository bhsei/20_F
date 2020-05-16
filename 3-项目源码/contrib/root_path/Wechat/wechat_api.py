import requests
import json
import time

class WechatAPI(object):

    cached_access_token: str = ""
    expire_time: float = 0

    app_id: str = ""
    app_secret: str = ""

    def __init__(self, app_id, app_secret):
        self.app_id = app_id
        self.app_secret = app_secret
        self._get_cached_token()

    def _get_cached_token(self):
        if self.expire_time < time.time():
            token, e = self._get_token()
            self.cached_access_token = token
            self.expire_time = time.time() + e - 100
        return self.cached_access_token

    def _get_token(self):
        api = "https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid={}&secret={}".format(self.app_id, self.app_secret)
        r = requests.get(api)
        data = r.json()
        return data["access_token"], int(data["expires_in"])

    @staticmethod
    def _err(r):
        if "errcode" in r:
            code = float(r["errcode"])
            msg = r["errmsg"]
            if code != 0:
                print("response error", msg)

    def send(self, open_id, message, url):
        token = self._get_cached_token()
        api = "https://api.weixin.qq.com/cgi-bin/message/template/send?access_token={}".format(token)
        data = {
                "touser": open_id,
                "template_id": "K2lk62bRKlzrhPB57D-hf2aHrBtqVL1-mIDYqHxoBG8",
                "url": url,
                "data": {
                    "data": {
                        "value": message
                        }
                    }
                }
        r = requests.post(api, json = data)
        self._err(r.json())

    def get_qr_code(self, data, expire):
        token = self._get_cached_token()
        api = "https://api.weixin.qq.com/cgi-bin/qrcode/create?access_token={}".format(token)
        data = {
                "expire_seconds": int(expire),
                "action_name": "QR_STR_SCENE",
                "action_info": {
                    "scene": {
                        "scene_str": data
                        }
                    }
                }
        r = requests.post(api, json = data).json()
        self._err(r)
        if "url" in r:
            return r["url"]
        return ""

