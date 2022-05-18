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


def say_hello(msg):
    response = client.SayHello(bridge_pb2.SayHelloRequest(
        msg=msg
    ))
    # print("received: " + response.msg)


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
    return convert_state(resp), resp.Base.Extra['mode']


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
        return next_state, reward, accuracy, False, resp.Base.Extra['mode']
    elif resp.Base.RetCode == 1:
        accuracy = resp.Feedback.accuracy
        reward = resp.Feedback.reward
        return None, reward, accuracy, True, resp.Base.Extra['mode']


def close():
    say_hello('Over')


if __name__ == '__main__':
    state = reset_env(None)
    print(state)
