import configparser
from pathlib import Path
from typing import Dict


def load_config(config_path: 'Path'):
    parser = configparser.ConfigParser()
    with config_path.open("r") as f:
        parser.readfp(f)
    return parser


def get_option_or_default(section: Dict[str, str], option: str, default=None, convert_func=None) -> str:
    if option not in section:
        if default is None:
            raise KeyError
        return default
    data = section[option]
    if convert_func is not None:
        data = convert_func(data)
    return data
