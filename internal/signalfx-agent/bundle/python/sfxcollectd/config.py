"""
Logic for converting from the agent config format to the Collectd-python config
object format
"""

import logging

logger = logging.getLogger(__name__)


class Config(object):  # pylint: disable=too-few-public-methods
    """
    Dummy class that we use to put config that conforms to the collectd-python
    Config class

    See https://collectd.org/documentation/manpages/collectd-python.5.shtml#config
    """

    def __init__(self, root=None, key=None, values=None, children=None):
        self.root = root
        self.key = key
        self.values = values
        self.children = children

    # pylint:disable=too-many-branches
    @classmethod
    def from_monitor_config(cls, monitor_plugin_config):
        """
        Converts config as expressed in the monitor to the Collectd Config
        interface.
        """
        assert isinstance(monitor_plugin_config, dict)

        conf = cls(root=None)
        conf.children = []
        for key, val in list(monitor_plugin_config.items()):
            values = None
            children = None
            if val is None:
                logging.debug("dropping configuration %s because its value is None", key)
                continue
            if isinstance(val, (tuple, list)):
                if not val:
                    logging.debug("dropping configuration %s because its value is an empty list or tuple", key)
                    continue
                values = val
            elif isinstance(val, str):  # pylint: disable=undefined-variable
                if not val:
                    logging.debug("dropping configuration %s because its value is an empty string", key)
                    continue
                values = (val,)
            elif isinstance(val, bytes):
                if not val:
                    logging.debug("dropping configuration %s because its value is an empty string", key)
                    continue
                values = (val.decode("utf-8"),)
            elif isinstance(val, (int, float, bool)):
                values = (val,)
            elif isinstance(val, dict):
                if not val:
                    logging.debug("dropping configuration %s because its value is an empty dictionary", key)
                    continue
                if "#flatten" in val and "values" in val:
                    conf.children += [
                        cls(root=conf, key=key, values=item if isinstance(item, (list, tuple)) else [item], children=[])
                        for item in val.get("values") or []
                        if item is not None
                    ]
                    continue
                dict_conf = cls.from_monitor_config(val)
                children = dict_conf.children
                values = dict_conf.values
            else:
                logging.error(
                    "Cannot convert monitor config to collectd config: %s: %s (type: %s)", key, val, type(val)
                )
                continue

            conf.children.append(cls(root=conf, key=key, values=values, children=children))

        return conf
