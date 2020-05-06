from abc import ABC, abstractmethod
from typing import Dict

SettingType = Dict[str, str]

class ModuleAbstract(ABC):

    def __init__(self, setting: SettingType = None):
        self.globalSetting = setting

    @abstractmethod
    def send(self, title: str, content: str, url: str, user_setting: SettingType) -> int:
        return 0

    def globalSettingCheck(self, setting: SettingType) -> int:
        self.globalSetting = setting
        return 0
    
    def userSettingCheck(self, setting: SettingType) -> int:
        return 0
