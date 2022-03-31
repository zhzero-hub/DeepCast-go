from concurrent import futures
import time
import grpc
import common_pb2
import bridge_pb2
import bridge_pb2_grpc
from util import *

import numpy as np


_HOST = '127.0.0.1'
_PORT = '5002'
channel = grpc.insecure_channel("{0}:{1}".format(_HOST, _PORT))
client = bridge_pb2_grpc.ServiceApiStub(channel=channel)


def get_system_info(base):
    if base is None:
        base = common_pb2.Base(
            RetCode=0,
            RetMsg="Success"
        )
    req = bridge_pb2.SystemInfoRequest(
        Base=base
    )
    resp = client.SystemInfo(req)
    return resp


def get_task_manager_info(base):
    if base is None:
        base = common_pb2.Base(
            RetCode=0,
            RetMsg="Success",
            Extra={'max': '5'}
        )
    req = bridge_pb2.TaskManagerInfoRequest(
        Base=base
    )
    resp = client.TaskManagerInfo(req)
    return resp


def get_background_info(base, extra):
    if base is None:
        base = common_pb2.Base(
            RetCode=0,
            RetMsg="Success",
            Extra=extra
        )
    req = bridge_pb2.BackgroundInfoRequest(
        Base=base
    )
    resp = client.BackgroundInfo(req)
    return resp
