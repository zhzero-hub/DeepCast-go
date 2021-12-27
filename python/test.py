import tensorflow as tf
import gym
from gym import spaces
import numpy as np
from tensorflow.python.client import device_lib


if __name__ == '__main__':
    # 列出所有的本地机器设备
    local_device_protos = device_lib.list_local_devices()
    # 打印
    print(local_device_protos)
    # 只打印GPU设备
    [print(x) for x in local_device_protos if x.device_type == 'GPU']
