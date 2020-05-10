import configparser
from pathlib import Path
from typing import Dict

def load_config(config_path: 'Path'):
    parser = configparser.ConfigParser()
    with config_path.open("r") as f:
        parser.readfp(f)
    sections = parser.sections()
    result = {}
    for section in sections:
        temp_dict = {}
        options = parser.options(section)
        for option in options:
            temp_dict[option] = parser.get(section, option)
        result[section] = temp_dict
    return result

def get_option_or_default(section: Dict[str, str], option: str, default = None, convert_func = None) -> str:
    if option not in section:
        if default is None:
            raise KeyError
        return default
    data = section[option]
    if convert_func is not None:
        data = convert_func(data)
    return data
