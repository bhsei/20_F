import json
import re
from typing import Dict, List

import requests

from module_definition import ModuleAbstract, SettingType, RedirectUrl


class FeiShuModule(ModuleAbstract):

    @staticmethod
    def feishu_msg(bot_url, title, text):
        headers = {
            'Content-Type': 'application/json',
        }
        data = {"title": title, "text": text}

        response = requests.post(bot_url, headers=headers,
                                 data=json.dumps(data),
                                 timeout=20)
        return response.text

    def __init__(self, db_proxy, config, global_setting: SettingType = None):
        """
        :param db_proxy: 数据库代理对象，有如下方法:
            load(user_id: int) -> Dict[str, str] 返回当前模块的个人设置，若没有则为None
            store(user_id: int, dict: Dict[str, str])->bool 将dict对应设置绑定到user_id对应的用户上，返回操作状态标志
        :param dict config: 模块config.json文件内容解析成的dict
        :param dict global_setting: 模块全局设置
        """
        super(FeiShuModule).__init__(db_proxy, config, global_setting)

    def send(self, title: str, content: str, url: str, uer_id: int) -> int:
        setting = self.db_proxy.load(uer_id)
        if not setting.get('bot_url'):
            return 1

        type(self).feishu_msg(text=content, title='%s通知' % self.global_setting.get('site_name', "Gitea"),
                              bot_url=setting.get('bot_url'))
        return 0

    def get_redirect_urls(self) -> List['RedirectUrl']:
        return []

    def global_setting_check(self, setting: Dict[str, str]) -> bool:
        return True

    def user_setting_check(self, user_id: int, setting: Dict[str, str]) -> bool:
        if not re.match(r"https://open\.feishu\.cn/open-apis/bot/hook/[0-9a-z]*?", setting.get('bot_url')):
            return False

        self.db_proxy.store(user_id, setting)
        return True


def load_module(db_proxy, config, global_setting: SettingType):
    return FeiShuModule(db_proxy, config, global_setting)
