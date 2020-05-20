from Wechat.wechat import Wechat

def load_module(db_proxy, config):
    return Wechat(db_proxy, config)
