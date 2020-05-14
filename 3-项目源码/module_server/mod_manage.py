from pathlib import Path
import mod_config
import mod_upload
from module_definition import RedirectUrl
import db_operation
import db_proxy
import sys
import json
import importlib
from typing import List, Union, Tuple, Dict

class ModuleManage(object):

    """manage module directory"""

    GLOBAL_INIT = 1
    NORMAL = 2

    _root_path = None
    _mlist = None
    _db = None
    _check = None
    _load = None

    def __init__(self, root_path: 'Path'):
        """read root_path directory, initialize Config and load all modules
        """
        self._root_path = root_path
        self._mlist = {}
        root_path.mkdir(parents = True, exist_ok = True)
        config_path = root_path.joinpath("config.ini")
        self._config = mod_config.Config(config_path)
        modules = self._config.get("gitea", "enabled_module")
        modules = modules.split(",")
        sys.path.append(str(root_path))

        def is_str_list(data):
            return bool(data) and all(isinstance(item, str) for item in data)

        def global_setting_check(name):
            gs = self._mlist[name]["config"]["globalSetting"]
            obj = self._mlist[name]["object"]
            f = {}
            for g in gs:
                f[g] = self._config.get(name, g, "")
            if all(f.values()) and obj.global_setting_check(f):
                status = self.NORMAL
            else:
                status = self.GLOBAL_INIT
            self._mlist[name]["status"] = status
            self._mlist[name]["global_setting"] = f

        def init_database():
            host = self._config.get("gitea", "host", "localhost")
            port = self._config.get("gitea", "port", "3306")
            port = int(port)
            user = self._config.get("gitea", "user")
            password = self._config.get("gitea", "password")
            database = self._config.get("gitea", "database")
            self._db = db_operation.DBOperation()
            self._db.db_init(host, port, user, password, database)

        def load_module(name: str) -> List['RedirectUrl']:
            module_path = self._root_path.joinpath(name)
            conf = module_path.joinpath("config.json")
            assets_path = module_path.joinpath("assets")
            global_tmpl = assets_path.joinpath("global.tmpl")
            user_tmpl = assets_path.joinpath("user.tmpl")

            global_tmpl = global_tmpl.read_text()
            user_tmpl = user_tmpl.read_text()
            conf = json.loads(conf.read_text())
            
            if not ("name" in conf and
                    conf["name"] == name):
                raise ValueError("{}: name attribute is invalid".format(name))
            if not ("globalSetting" in conf and
                    is_str_list(conf["globalSetting"]) and
                    "userSetting" in conf and
                    is_str_list(conf["userSetting"])):
                raise ValueError("{}: globalSetting/userSetting attribute is invalid".format(name))

            gs = conf["globalSetting"]
            us = conf["userSetting"]

            proxy = db_proxy.DBProxy(name, gs, self._db)
            pkg = importlib.import_module(name + ".entry")
            obj = pkg.load_module(proxy, conf)

            self._mlist[name] = {
                    "global_tmpl": global_tmpl,
                    "user_tmpl": user_tmpl,
                    "config": conf,
                    "object": obj,
                    }
            self._check(name)
            print("load {}".format(name))
            return obj.get_redirect_urls()


        def load(name):
            try:
                load_module(name)
            except RuntimeError:
                pass
        
        self._check = global_setting_check
        self._load = load_module
        init_database()
        list(map(load, modules))

    def get_redirect_urls(self, module = None) -> List['RedirectUrl']:
        if module is None:
            urls = list(map(lambda val: val["object"].get_redirect_urls(), self._mlist.values()))
            return [item for sub in urls for item in sub]
        if module not in self._mlist:
            return []
        return self._mlist[module]["object"].get_redirect_urls()

    def load_module(self, data: bytes) -> Union[List['RedirectUrl'], type(None)]:
        ok, name = mod_upload.extract_module(data, self._root_path)
        if not ok:
            return None
        return self._load(name)

    def global_tmpls(self) -> List[str]:
        return list(map(lambda k: (k, self._mlist[k]["global_tmpl"]), self._mlist.keys()))

    def user_tmpls(self) -> List[Tuple[str, str]]:
        valid = filter(lambda k: self._mlist[k]["status"] == self.NORMAL, self._mlist.keys())
        tmpls = map(lambda k: (k, self._mlist[k]["user_tmpl"]), valid)
        return list(tmpls)

    def add_global_setting(self, module: str, settings: Dict[str, str]) -> bool:
        if module not in self._mlist:
            return False
        gs = self._mlist[module]["config"]["globalSetting"]
        obj = self._mlist[module]["object"]
        f = {}
        for g in gs:
            if g not in settings or not settings[g]:
                return False
            f[g] = settings[g]
        if not obj.global_setting_check(f):
            return False
        self._mlist[module]["status"] = self.NORMAL
        list(map(lambda key: self._config.set(module, key, f[key]), f.keys()))
        return True

    def add_user_setting(self, module: str, user_id: int, settings: Dict[str, str]) -> bool:
        if module not in self._mlist:
            return False
        us = self._mlist[module]["config"]["userSetting"]
        obj = self._mlist[module]["object"]
        f = {}
        for u in us:
            if u not in settings or not settings[u]:
                return False
            f[u] = settings[u]
        if not obj.user_setting_check(user_id, f):
            return False
        return True

    def send(self, title, content, url, users):
        for user in users:
            for d in self._mlist.values():
                if d["status"] == NORMAL:
                    d["object"].send(title, content, url, user)

    def save_state(self):
        enabled = self._mlist.keys()
        self._config.set("gitea", "enabled_module", ",".join(enabled))
        self._config.save()
