from abc import ABC, abstractmethod
from typing import Dict, List, Callable, Tuple

SettingType = Dict[str, str]


class RedirectUrl(object):
    URL_GET = 1
    URL_POST = 2

    def __init__(self, url_pattern: str, url_type: int,
                 handler: Callable[[Dict[str, str], bytes], Tuple[str, bytes]]):
        """initialize RedirectUrl object
        handler: return content_type and data
        """
        self.url_pattern = url_pattern
        self.url_type = url_type
        self.handler = handler


"""
在创建模块时，原则上不允许导入任何其他mod_*模块，若有其他不能实现的需求，再讨论
"""
class ModuleAbstract(ABC):

    def __init__(self, db_proxy, config, setting: SettingType = None):
        """
        db_proxy: 数据库代理对象，有如下方法:
            load(user_id: int) -> Dict[str, str] 返回当前模块的个人设置，若没有则为None
            store(user_id: int, dict: Dict[str, str])->bool 将dict对应设置绑定到user_id对应的用户上，返回操作状态标志
        config: 模块config.json文件内容解析成的dict
        """
        self.db_proxy = db_proxy
        self.config = config
        self.global_setting = setting

    @abstractmethod
    def send(self, title: str, content: str, url: str, uer_id: int) -> int:
        return 0

    def get_redirect_urls(self) -> List['RedirectUrl']:
        return []

    def global_setting_check(self, setting: Dict[str, str]) -> bool:
        self.global_setting = setting
        return True

    def user_setting_check(self, user_id: int, setting: Dict[str, str]) -> bool:
        return self.db_proxy.store(user_id, setting)
