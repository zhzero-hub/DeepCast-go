from gym import spaces, core
import numpy


# core.Env 是 gym 的环境基类,自定义的环境就是根据自己的需要重写其中的方法；
# 必须要重写的方法有:
# __init__()：构造函数
# reset()：初始化环境
# step()：环境动作,即环境对agent的反馈
# render()：如果要进行可视化则实现

E = 100
channel = 1000000
version = 4


class DeepCastEnv(core.Env):
    def __init__(self):
        self.action_space = spaces.Box(low=0, high=1, shape=(E + 1,), dtype=numpy.float64)  # 动作空间
        self.observation_space = spaces.Dict({
            "Inbound_bandwidth_usage": spaces.Box(low=-1.0, high=1.0, shape=(E,)),
            "Outbound_bandwidth_usage": spaces.Box(low=-1.0, high=1.0, shape=(E + 1,)),
            "Computation_resource_usage": spaces.Box(low=-1.0, high=1.0, shape=(E,)),
            "Viewer_connection_table": spaces.Box(low=-1.0, high=1.0, shape=(E, channel, version)),
            "location": spaces.Box(low=-360.0, high=360.0, shape=(1, 1)),
            "channel": spaces.Discrete(channel),
            "version": spaces.Discrete(version),
            "QoE_preference": spaces.Discrete(3)
        })  # 状态空间

    def reset(self):
        obs = self._get_observation(None)
        return obs

    def step(self, action):
        reward = self._get_reward()
        done = self._get_done()
        obs = self._get_observation(action)
        info = {}  # 用于记录训练过程中的环境信息,便于观察训练状态
        return obs, reward, done, info

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
