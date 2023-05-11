"""
Logging helpers
"""


import logging
import sys
import traceback

from .messages import MSG_TYPE_LOG

logger = logging.getLogger()


class PipeLogHandler(logging.Handler):
    """
    Python log handler that converts log messages to a json object and sends
    them back to the agent through the given writeable pipe
    """

    def __init__(self, writer):
        """
        `pipe` should be a PipeMessageWriter that is already opened
        """
        self.writer = writer

        super().__init__()

    def emit(self, record):
        self.writer.send_msg(
            MSG_TYPE_LOG,
            {
                "message": record.getMessage(),
                "logger": record.name,
                "source_path": record.pathname,
                "lineno": record.lineno,
                "created": record.created,
                "level": record.levelname,
            },
        )


def format_exception():
    """
    Format the current exception as a traceback
    """
    exc_type, exc_value, exc_traceback = sys.exc_info()
    return "\n".join(traceback.format_exception(exc_type, exc_value, exc_traceback))


def log_exc_traceback_as_error():
    """
    Log the current exception at the error level.  Meant to be called when in
    an except block
    """
    logger.error(format_exception())
