from concurrent import futures
import time
import grpc
import common_pb2
import bridge_pb2
import bridge_pb2_grpc


_HOST = '127.0.0.1'
_PORT = '5001'
channel = grpc.insecure_channel("{0}:{1}".format(_HOST, _PORT))
client = bridge_pb2_grpc.TrainApiStub(channel=channel)


def say_hello():
    response = client.SayHello(bridge_pb2.SayHelloRequest(
        msg="python_"
    ))
    print("received: " + response.msg)


def train_step(base, action):
    req = bridge_pb2.TrainStepRequest(
        Base=base
    )
    resp = client.TrainStep(req)
    print(resp)


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
    state = []
    inboundBandwidthUsed = []
    outboundBandwidthUsed = []
    computationResourceUsage = []
    for inbound in resp.State.inbound_bandwidth_usage:
        inboundBandwidthUsed.append(inbound)
    state.append(inboundBandwidthUsed)
    for outbound in resp.State.outbound_bandwidth_usage:
        outboundBandwidthUsed.append(outbound)
    state.append(outboundBandwidthUsed)
    for compute in resp.State.computation_resource_usage:
        computationResourceUsage.append(compute)
    state.append(computationResourceUsage)

    return state


if __name__ == '__main__':
    reset_env(None)
