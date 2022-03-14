from concurrent import futures
import time
import grpc
import common_pb2
import bridge_pb2
import bridge_pb2_grpc
from util import *

import numpy as np


_HOST = '127.0.0.1'
_PORT = '5001'
channel = grpc.insecure_channel("{0}:{1}".format(_HOST, _PORT))
client = bridge_pb2_grpc.TrainApiStub(channel=channel)


def convert_state(resp):
    state = {}
    inbound_bandwidth_used = []
    outbound_bandwidth_used = []
    computation_resource_usage = []
    for inbound in resp.State.inbound_bandwidth_usage.inbound_bandwidth_usage:
        inbound_bandwidth_used.append(inbound)
    state['inbound_bandwidth_used'] = inbound_bandwidth_used
    # state.append(inbound_bandwidth_used)
    for outbound in resp.State.outbound_bandwidth_usage.outbound_bandwidth_usage:
        outbound_bandwidth_used.append(outbound)
    state['outbound_bandwidth_used'] = outbound_bandwidth_used
    # state.append(outbound_bandwidth_used)
    for compute in resp.State.computation_resource_usage.computation_resource_usage:
        computation_resource_usage.append(compute)
    state['computation_resource_usage'] = computation_resource_usage
    # state.append(computation_resource_usage)
    viewer_connection_table = []
    for _ in range(4000):
        viewer_connection_table.append(0)

    state['viewer_connection_table'] = viewer_connection_table
    # state.append(viewer_connection_table)
    user_info = [resp.State.user_info.location.latitude, resp.State.user_info.location.longitude,
                 get_channel_id(resp.State.user_info.channel_id), get_version_id(resp.State.user_info.version)]
    state['user_info'] = user_info
    # state.append(user_info)
    qoe = [resp.State.qoe_preference.alpha1, resp.State.qoe_preference.alpha2, resp.State.qoe_preference.alpha3]
    state['qoe'] = qoe
    # state.append(qoe)

    state['viewer_id'] = resp.State.user_info.user_id
    state['channel_id'] = resp.State.user_info.channel_id
    state['version'] = resp.State.user_info.version
    return state


def say_hello():
    response = client.SayHello(bridge_pb2.SayHelloRequest(
        msg="python_"
    ))
    print("received: " + response.msg)


def reset_env(base):
    if base is None:
        base = common_pb2.Base(
            RetCode=0,
            RetMsg="Success"
        )
    req = bridge_pb2.ResetEnvRequest(
        Base=base
    )
    resp = client.ResetEnv(req)
    return convert_state(resp)


def train_step(base, device_id, viewer_id, channel_id, version, qoe):
    if base is None:
        base = common_pb2.Base(
            RetCode=0,
            RetMsg="Success",
            Extra={"version": str(version), "deviceId": str(device_id)}
        )
    action = common_pb2.Action(
        viewer_id=viewer_id,
        channel_id=channel_id,
        version=version,
        action=device_id,
        qoe_preference=common_pb2.QoEPreference(
            alpha1=qoe[0],
            alpha2=qoe[1],
            alpha3=qoe[2],
        )
    )
    req = bridge_pb2.TrainStepRequest(
        Base=base,
        Action=action,
    )
    resp = client.TrainStep(req)
    if resp.Base.RetCode == 0:
        # print(resp)
        accuracy = resp.Feedback.accuracy
        reward = resp.Feedback.reward
        next_state = convert_state(resp)
        return next_state, reward, accuracy, False
    elif resp.Base.RetCode == 1:
        accuracy = resp.Feedback.accuracy
        reward = resp.Feedback.reward
        return None, reward, accuracy, True


if __name__ == '__main__':
    state = reset_env(None)
    print(state)
