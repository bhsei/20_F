from typing import List, Tuple
from module_definition import RedirectUrl

class ModuleRedirect(object):

    """manage redirect urls"""

    _url_list = []

    def __init__(self):
        pass

    @staticmethod
    def _add_prefix(module: str, url: 'RedirectUrl'):
        return RedirectUrl("/{}{}".format(module, url.url_pattern), url.url_type, url.handler)

    def register_urls(self, module: str, urls: List['RedirectUrl']) -> List[Tuple['RedirectUrl', int]]:
        base = len(self._url_list)
        urls = list(map(lambda u: self._add_prefix(module, u), urls))
        self._url_list += urls
        return list(map(lambda a, b: (a, b), urls, range(base, base + len(urls))))

    def get_urls(self) -> List[Tuple['RedirectUrl', int]]:
        return list(map(lambda a, b: (a, b), self._url_list, range(len(self._url_list))))

    def get_url_by_id(self, id) -> 'RedirectUrl':
        return self._url_list[id]

