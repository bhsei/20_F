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

class ModuleAbstract(ABC):

    def __init__(self, setting: SettingType = None):
        self.globalSetting = setting

    @abstractmethod
    def send(self, title: str, content: str, url: str, user_setting: SettingType) -> int:
        return 0

    def get_redirect_urls(self) -> List['RedirectUrl']:
        return []
