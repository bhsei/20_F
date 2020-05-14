from typing import List, Tuple
from module_definition import RedirectUrl

class ModuleRedirect(object):

    """manage redirect urls"""

    def __init__(self):
        pass

    def register_urls(self, urls: List['RedirectUrl']) -> List[int]:
        pass

    def get_urls(self) -> List[Tuple['RedirectUrl', int]]:
        return []

    def get_url_by_id(self, id) -> 'RedirectUrl':
        return None

