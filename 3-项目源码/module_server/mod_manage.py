from pathlib import Path
import mod_config
import mod_util
from mod_redirect import ModuleRedirect
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

    def __init__(self, root_path: 'Path', redirect_manager: 'ModuleRedirect'):
        """read root_path directory, initialize Config and load all modules
        """
        self._root_path = root_path
        self._redirect_manager = redirect_manager
        self._mlist = {}
        root_path.mkdir(parents = True, exist_ok = True)
        config_path = root_path.joinpath("config.ini")
        self._config = mod_config.Config(config_path)
        modules = self._config.get("gitea", "enabled_module")
        modules = modules.split(",")
        sys.path.append(str(root_path))

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

            if not mod_util.check_module_config(conf):
                raise ValueError("{}: config.json parse error".format(name))

            gs = conf["globalSetting"]
            us = conf["userSetting"]

            proxy = db_proxy.DBProxy(name, us, self._db)
            pkg = importlib.import_module(name + ".entry")
            obj = pkg.load_module(proxy, conf)

            self._mlist[name] = {
                    "global_tmpl": global_tmpl,
                    "user_tmpl": user_tmpl,
                    "config": conf,
                    "object": obj,
                    }
            self._check(name)
            urls = self._redirect_manager.register_urls(name, obj.get_redirect_urls())
            print("load {}".format(name))
            return urls


        def load(name):
            try:
                load_module(name)
            except RuntimeError:
                pass

        self._check = global_setting_check
        self._load = load_module
        init_database()
        list(map(load, modules))

    def load_module(self, data: bytes) -> Union[List['RedirectUrl'], type(None)]:
        ok, name = mod_util.extract_module(data, self._root_path)
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
            print("module {} not in module list".format(module))
            return False
        gs = self._mlist[module]["config"]["globalSetting"]
        obj = self._mlist[module]["object"]
        f = {}
        for g in gs:
            if g not in settings or not settings[g]:
                print("setting {} is not valid".format(g))
                return False
            f[g] = settings[g]
        if not obj.global_setting_check(f):
            print("setting check error")
            return False
        self._mlist[module]["status"] = self.NORMAL
        list(map(lambda key: self._config.set(module, key, f[key]), f.keys()))
        return True

    def add_user_setting(self, module: str, user_id: int, settings: Dict[str, str]) -> bool:
        if module not in self._mlist:
            print("module {} not in module list".format(module))
            return False
        us = self._mlist[module]["config"]["userSetting"]
        obj = self._mlist[module]["object"]
        f = {}
        for u in us:
            if u not in settings or not settings[u]:
                print("{} not in setting".format(u))
                return False
            f[u] = settings[u]
        if not obj.user_setting_check(user_id, f):
            print("setting check error")
            return False
        return True

    def send(self, title, content, url, users):
        for user in users:
            for d in self._mlist.values():
                if d["status"] == self.NORMAL:
                    d["object"].send(title, content, url, user)

    def save_state(self):
        enabled = self._mlist.keys()
        self._config.set("gitea", "enabled_module", ",".join(enabled))
        self._config.save()
