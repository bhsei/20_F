from module_definition import ModuleAbstract, SettingType

class ExampleModule(ModuleAbstract):

    def send(self, title: str, content: str, url: str, user_setting: SettingType) -> int:
        return 0

def load_module(global_setting: SettingType):
    return ExampleModule(global_setting)
