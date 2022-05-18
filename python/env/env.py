from gym import spaces, core
import numpy
from rpc.train import *

# core.Env 是 gym 的环境基类,自定义的环境就是根据自己的需要重写其中的方法；
# 必须要重写的方法有:
# __init__()：构造函数
# reset()：初始化环境
# step()：环境动作,即环境对agent的反馈
# render()：如果要进行可视化则实现

E = 10
channel = 50
version = 6
STATUS = False


def set_status(_STATUS):
    global STATUS
    STATUS = _STATUS


class DeepCastEnv(core.Env):
    def __init__(self):
        self.action_space = spaces.Box(low=0, high=1, shape=(E + 1,), dtype=numpy.float64)  # 动作空间
        self.observation_space = spaces.Dict({
            "Inbound_bandwidth_usage": spaces.Box(low=-1.0, high=1.0, shape=(E,)),
            "Outbound_bandwidth_usage": spaces.Box(low=-1.0, high=1.0, shape=(E + 1,)),
            "Computation_resource_usage": spaces.Box(low=-1.0, high=1.0, shape=(E,)),
            "Viewer_connection_table": spaces.Box(low=-1.0, high=1.0, shape=(E, channel, version)),
            "User_info": spaces.Box(low=0.0, high=200, shape=(4,)),
            "QoE_preference": spaces.Box(low=0, high=10, shape=(3,))
        })  # 状态空间

    def reset(self):
        obs, mode = reset_env(None)
        return obs, int(mode)

    def step(self, action):
        obs, reward, accuracy, done, mode = train_step(None, action['device_id'], action['viewer_id'], action['channel_id'],
                                                 action['version'], action['qoe'])
        return obs, reward, done, int(mode)

    def render(self, mode='human'):
        return 0

    # 根据需要设计相关辅助函数
    def _reset(self):
        return 0

    def _get_observation(self, action):
        return 0

    def _get_reward(self):
        return 0

    def _get_done(self):
        return True
