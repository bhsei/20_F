import json
import importlib
import mod_config
from typing import Dict
from db_operation import DBOperation
import mod_redirect

# 配置文件存放在rootPath目录下
CONFIG_FILE = "config.ini"
ROOT_PATH = None

# 模块状态
GLOBAL_INIT = 1
NORMAL = 2
BROKEN = 3

module_list = {}
config = {}
db = None


def is_str_list(data):
    return bool(data) and all(isinstance(item, str) for item in data)


def load_module(module: str, config: Dict[str, str]) -> bool:
    global module_list
    module_path = ROOT_PATH.joinpath(module)
    conf_path = module_path.joinpath("config.json")
    assets_path = module_path.joinpath("assets")
    global_setting = assets_path.joinpath("global.tmpl")
    user_setting = assets_path.joinpath("user.tmpl")
    try:
        global_setting = global_setting.read_text()
        user_setting = user_setting.read_text()
    except Exception:
        return False
    status = NORMAL
    try:
        pkg = importlib.import_module(module + ".entry")
    except Exception:
        return False
    try:
        conf = json.loads(conf_path.read_text())
    except Exception:
        return False
    if not ("name" in conf and
            conf["name"] == module):
        return False
    if not ("globalSetting" in conf and
            is_str_list(conf["globalSetting"]) and
            "userSetting" in conf and
            is_str_list(conf["userSetting"])):
        return False
    g = conf["globalSetting"]
    for s in g:
        if s not in config or config[s] == "":
            status = GLOBAL_INIT
            break
    try:
        obj = pkg.load_module(config)
        urls = obj.get_redirect_urls()
        mod_redirect.register_urls(urls)
    except Exception:
        return False
    module_list[module] = {
        "status": status,
        "global_setting": global_setting,
        "user_setting": user_setting,
        "config": config,
        "module_conf": conf,
        "object": obj,
    }
    print("load module:", module_list[module])
    return True


def init_database(config: Dict[str, Dict[str, str]]):
    global db
    section = config["gitea"]
    host = mod_config.get_option_or_default(section, "host", "localhost")
    port = mod_config.get_option_or_default(section, "port", 3306, int)
    user = mod_config.get_option_or_default(section, "user")
    password = mod_config.get_option_or_default(section, "password")
    database = mod_config.get_option_or_default(section, "database")
    db = DBOperation()
    db.db_init(host, port, user, password, database)
    return db


def init_module(config: Dict[str, Dict[str, str]]):
    section = config["gitea"]
    modules = mod_config.get_option_or_default(section, "enabled_module",
                                               [], lambda s: s.split(","))
    for module in modules:
        if module not in config:
            config[module] = {}
        load_module(module, config[module])
