from zipp import zipfile, Path
from typing import Tuple
import io
import json
import pathlib
import shutil

def is_str_list(data):
    return bool(data) and all(isinstance(item, str) for item in data)


def check_module_config(config):
    return ("name" in config and
            config["name"] and
            "globalSetting" in config and
            is_str_list(config["globalSetting"]) and
            "userSetting" in config and
            is_str_list(config["userSetting"]))


def check_module(root_path):
    conf_path = root_path / "config.json"
    assets_path = root_path / "assets"
    global_path = assets_path / "global.tmpl"
    user_path = assets_path / "user.tmpl"
    entry = root_path / "entry.py"
    if not (conf_path.is_file() and
            assets_path.is_dir() and
            global_path.is_file() and
            user_path.is_file() and
            entry.is_file()):
        return False, "invalid file structure"
    try:
        conf = json.loads(conf_path.read_text())
    except json.JSONDecodeError:
        return False, "failed to parse config.json"
    if not check_module_config(conf):
        return False, "lack essential setting items in config.json"
    return True, conf["name"]


def extract_module_zip(module, target_dir):
    ok, name = check_module(Path(module))
    if not ok:
        return False, name
    target = target_dir.joinpath(name)
    if target.exists():
        return False, "found the same module"
    try:
        target.mkdir(parents=True, exist_ok=False)
        module.extractall(str(target))
    except Exception as e:
        shutil.rmtree(target)
        return False, "failed to create module directory: {}".format(e)
    return True, name


def extract_module(module_bytes: bytes, target_dir: 'pathlib.Path') -> Tuple[bool, str]:
    data = io.BytesIO(module_bytes)
    zip_file = zipfile.ZipFile(data, "r")
    return extract_module_zip(zip_file, target_dir)
