import sys
import json
import argparse
from pathlib import Path
from configparser import ConfigParser
import importlib
import server_config
from typing import Dict, List
from db_operation import DBOperation
import mod_redirect
import grpc
from concurrent import futures
import service_pb2_grpc
import mod_service
import logging

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
            "object": obj
            }
    return True

def init_database(config: Dict[str, Dict[str, str]]):
    global db
    section = config["gitea"]
    host = server_config.get_option_or_default(section, "host", "localhost")
    port = server_config.get_option_or_default(section, "port", 3306, int)
    user = server_config.get_option_or_default(section, "user")
    password = server_config.get_option_or_default(section, "password")
    database = server_config.get_option_or_default(section, "database")
    db = DBOperation()
    db.db_init(host, port, user, password, database)
    return db

def init_module(config: Dict[str, Dict[str, str]]):
    section = config["gitea"]
    modules = server_config.get_option_or_default(section, "enabled_module",
            [], lambda s: s.split(","))
    for module in modules:
        if module not in config:
            config[module] = {}
        load_module(module, config[module])

def start_server(port):
    server = grpc.server(futures.ThreadPoolExecutor(max_workers = 10))
    service_pb2_grpc.add_NotifyServiceServicer_to_server(mod_service.NotifyService(), server)
    server.add_insecure_port('[::]:{}'.format(port))
    server.start()
    server.wait_for_termination()

def main(rootPath: str, port: int) -> int:
    global ROOT_PATH
    global config
    ROOT_PATH = Path(rootPath)
    ROOT_PATH.mkdir(parents = True, exist_ok = True)
    current_path = Path(__file__).parent.absolute()
    sys.path.insert(0, str(current_path))
    sys.path.insert(0, str(ROOT_PATH))
    config_path = Path(rootPath).joinpath(CONFIG_FILE)
    config = server_config.load_config(config_path)
    if "gitea" not in config:
        config["gitea"] = {}
    init_database(config)
    init_module(config)
    logging.basicConfig()
    start_server(port)
    return 0

if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument('-r', '--root_path', action = 'store',
            required = True, dest = 'root_path',
            help = 'specify the root path')
    parser.add_argument('-p', '--port', action = 'store',
            type = int, required = True, dest = 'port',
            help = 'specify the port to be used')
    args = parser.parse_args()
    sys.exit(main(args.root_path, args.port))
