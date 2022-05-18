import tensorflow as tf
import gym
from gym import spaces
import numpy as np
from tensorflow.python.client import device_lib
import math


def C(m, n):
    return math.factorial(n) / (math.factorial(m) * math.factorial(n - m))


if __name__ == '__main__':
    prob = 0
    for i in range(14, 29):
        prob += C(i, 28) * 0.58 ** i * (1 - 0.58) ** (28 - i)
    print(prob)
    # # 列出所有的本地机器设备
    # local_device_protos = device_lib.list_local_devices()
    # # 打印
    # print(local_device_protos)
    # # 只打印GPU设备
    # [print(x) for x in local_device_protos if x.device_type == 'GPU']

