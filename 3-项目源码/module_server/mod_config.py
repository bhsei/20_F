import configparser
from pathlib import Path
from typing import Dict, Union

class Config(object):

    """read/write config.ini"""

    def __init__(self, config_path: 'Path'):
        """read configuration from config_path
        caller must guarantee that all parents of config_path has been created
        """
        self._path = config_path
        self._parser = configparser.ConfigParser()
        if config_path.is_file():
            with config_path.open("r") as f:
                self._parser.readfp(f)

    def get(self, module: str, item: str, default = None) -> str:
        """return global settings of module stored in config

        :module: module name
        :item: global setting item
        :default: it module and item not exist, use default
        :returns: setting if exists or default

        """
        if module not in self._parser:
            if default is None:
                raise ValueError("no {} specified".format(item))
            return default
        if self._parser.has_option(module, item) and self._parser[module][item]:
            return self._parser[module][item]
        if default is not None:
            return default
        raise ValueError("no {} specified".format(item))

    def set(self, module, item, value):
        if module not in self._parser:
            self._parser[module] = {}
        self._parser[module][item] = value

    def save(self):
        with self._path.open("w") as f:
            self._parser.write(f)
        print("config saved {}".format(self._parser))

