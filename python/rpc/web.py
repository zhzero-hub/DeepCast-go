from concurrent import futures
import time
import grpc
import common_pb2
import bridge_pb2
import bridge_pb2_grpc
from util import *

import numpy as np


_HOST = '127.0.0.1'
_PORT = '5050'
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


def service(base, form):
    if base is None:
        base = common_pb2.Base(
            RetCode=0,
            RetMsg="Success",
            Extra={'desc': form['desc']}
        )
    user_id = form['userId']
    channel_id = form['channelId']
    start_time = form['startTime']
    end_time = form['endTime']
    encrypt = form['encrypt']
    lat = form['lat']
    long = form['long']
    resource = form['resource']
    version = form['version']
    # version = int(form['version'][:-1])
    if user_id == '' or channel_id == '' or start_time == '' or end_time == '' or lat == '' or long == '' or resource == '' or version == '':
        return ''
    location = common_pb2.Location(
            latitude=float(lat),
            longitude=float(long)
        )
    user_info = common_pb2.UserInfo(
        location=location,
        channel_id=str(channel_id),
        version=int(form['version'][:-1]),
        user_id=str(user_id)
    )
    service_info = common_pb2.ServiceInfo(
        start_time=start_time,
        end_time=end_time,
        encrypted=encrypt,
        resource=resource
    )
    resp = client.Service(bridge_pb2.ServiceRequest(
        Base=base,
        user_info=user_info,
        service_info=service_info
    ))
    return convert_state(resp)
