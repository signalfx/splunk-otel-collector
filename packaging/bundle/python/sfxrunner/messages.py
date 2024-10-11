"""
Logic for communicating with FIFO pipes between Python and the agent.  The
protocol is very simple: messages are sent in frames that are prefixed with a 4
byte integer indicating their length.
"""


import io
import logging
import os
import struct
import sys
import threading
from collections import namedtuple

import ujson

logger = logging.getLogger(__name__)

MSG_TYPE_CONFIGURE = 1
MSG_TYPE_CONFIGURE_RESULT = 2
MSG_TYPE_SHUTDOWN = 3
MSG_TYPE_LOG = 4

Message = namedtuple("Message", "type size payload")


def setup_io_pipes():
    """
    Creates an input_reader and output_writer that are connected to the
    process's stdin/out.

    Make sys.stdout be directed to the original stderr fd to make debugging
    easier and so we don't have to worry about random code corrupting the
    real stdout.  stderr will continue to go to its original location and
    should be used for debugging information.
    """
    real_out_fd = os.dup(sys.stdout.fileno())
    os.dup2(sys.stderr.fileno(), sys.stdout.fileno())

    # This process sends data back to the agent through stdout
    output_writer = PipeMessageWriter(real_out_fd)
    output_writer.open()

    # The agent sends control messages to this process via stdin
    input_reader = PipeMessageReader(sys.stdin.fileno())
    input_reader.open()

    return input_reader, output_writer


class _PipeMessageBase(object):
    def __init__(self, fd):
        """
        `path` is the path to a local fifo pipe
        """
        self.fd = fd
        self.file = None
        self.closed = False

    def open(self):
        """
        Open the pipe for either reading or writing
        """
        raise NotImplementedError()

    def close(self):
        """
        Close the pipe when we are completely done using it
        """
        self.closed = True
        if self.file:
            self.file.close()


class PipeMessageReader(_PipeMessageBase):
    """
    A message-oriented reader from a fifo pipe.
    """

    def open(self):
        self.file = io.open(self.fd, "rb", buffering=0)  # pylint: disable=consider-using-with

    def recv_msg(self):
        """
        Block until we receive a complete message from the pipe
        """
        msg_type = struct.unpack(">i", self.file.read(4))[0]
        size = struct.unpack(">i", self.file.read(4))[0]
        msg = self.file.read(size)

        logger.debug("Received control message with length %d", len(msg))
        return Message(type=msg_type, size=size, payload=ujson.loads(msg))


class PipeMessageWriter(_PipeMessageBase):
    """
    A message-oriented writer to a fifo pipe.  It sends length-prefixed
    messages, which should be efficient enough since messages generally won't
    be that big, so precalculating the length isn't that big of a deal.  The
    send_msg method is thread-safe.
    """

    def __init__(self, *args):
        super().__init__(*args)
        self.lock = threading.Lock()

    def open(self):
        self.file = io.open(self.fd, "wb", buffering=0)  # pylint: disable=consider-using-with

    def send_msg(self, msg_type, msg_obj):
        """
        Sends a message with the with the size prefixed to determine the
        message boundary on the receiving side.
        """
        msg_bytes = ujson.dumps(msg_obj).encode("utf-8")

        with self.lock:
            self.file.write(struct.pack(">i", msg_type))
            self.file.write(struct.pack(">i", len(msg_bytes)))
            self.file.write(msg_bytes)
