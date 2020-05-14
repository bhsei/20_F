from zipp import zipfile, Path
from typing import Tuple
import io
import json
import pathlib


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
        return False, ""
    try:
        conf = json.loads(conf_path.read_text())
    except json.JSONDecodeError:
        return False, ""
    if not ("name" in conf and
            "globalSetting" in conf and
            "userSetting" in conf):
        return False, ""
    return True, conf["name"]


def extract_module_zip(module, target_dir):
    ok, name = check_module(Path(module))
    if not ok:
        return False, ""
    target = target_dir.joinpath(name)
    try:
        target.mkdir(parents=True, exist_ok=False)
        module.extractall(str(target))
    except Exception:
        # TODO: remove target directory
        return False, ""
    return True, name


def extract_module(module_bytes: bytes, target_dir: 'pathlib.Path') -> Tuple[bool, str]:
    data = io.BytesIO(module_bytes)
    zip_file = zipfile.ZipFile(data, "r")
    return extract_module_zip(zip_file, target_dir)
