from typing import List, Tuple
from module_definition import RedirectUrl


def register_urls(urls: List['RedirectUrl']) -> List[int]:
    pass


def get_urls() -> List[Tuple['RedirectUrl', int]]:
    pass


def remove_urls() -> List[int]:
    pass


def get_url_by_id(id: int) -> 'RedirectUrl':
    pass
