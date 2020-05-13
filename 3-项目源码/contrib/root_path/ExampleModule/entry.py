from module_definition import ModuleAbstract, SettingType

class ExampleModule(ModuleAbstract):

    def send(self, title: str, content: str, url: str, user_id: int) -> int:
        return 0

def load_module(db_proxy, config, global_setting: SettingType):
    return ExampleModule(db_proxy, config, global_setting)
