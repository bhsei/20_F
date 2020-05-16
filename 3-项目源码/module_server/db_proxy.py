from typing import Dict, List, Union

def remove_prefix(text: str, prefix: str):
    if text.startswith(prefix):
        return text[len(prefix):]
    return text

class DBProxy(object):

    def __init__(self, module: str, setting_list: List[str], db):
        self.module = module
        self.prefix = module + "_"
        self.setting_list = setting_list
        self.plist = list(map(lambda s: self.prefix + s, setting_list))
        self.db = db
        self.db.db_add_setting(self.plist)

    def _load_by_id(self, uid: int) -> Dict[str, str]:
        settings = self.db.db_query(uid)
        if settings is None:
            return None
        result = {}
        for item in self.plist:
            if item not in settings:
                return None
            if settings[item] is None or settings[item] == "":
                return None
            t = remove_prefix(item, self.prefix)
            result[t] = settings[item]
        return result

    def _load_by_dict(self, cols: Dict[str, str]) -> List[int]:
        target = {}
        for key in cols.keys():
            target[self.prefix + key] = cols[key]
        return self.db.db_query_setting(target)

    def load(self, key: Union[Dict[str, str], int]) -> Union[Dict[str, str], List[int]]:
        if type(key) == dict:
            return self._load_by_dict(key)
        if type(key) == int:
            return self._load_by_id(key)
        return None

    def store(self, uid: int, data: Dict[str, str]) -> bool:
        f = {}
        for i in range(len(self.setting_list)):
            item = self.setting_list[i]
            pitem = self.plist[i]
            if item not in data:
                print("store not in {} {}".format(item, data))
                return False
            f[pitem] = data[item]
        if self.db.db_insert_or_update(uid, f) != 1:
            return False
        return True
