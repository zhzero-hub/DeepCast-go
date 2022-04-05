import gym
import numpy as np
from env.env import E, channel, version
import env.env as env
from agentA3CContinue import Actor, Critic


class ServiceAgent:
    def __init__(self, env_name):
        env = gym.make(env_name)
        self.env_name = env_name
        self.state_dim = env.observation_space.shape
        self.action_dim = env.action_space.shape
        self.action_bound = 1
        self.std_bound = [1e-2, 1.0]

        self.actor = Actor(
            self.state_dim, self.action_dim, self.action_bound, self.std_bound)
        self.critic = Critic(self.state_dim)

        self.actor.model.load_weights('./model/actor_model.h5')
        self.critic.model.load_weights('./model/critic_model.h5')

    def service(self, state):
        x = [np.array(np.reshape(state['inbound_bandwidth_used'], (1, E)), dtype=np.float64) / 1024 / 1024,
             np.array(np.reshape(state['outbound_bandwidth_used'], (1, E + 1)), dtype=np.float64) / 1024 / 1024,
             np.array(np.reshape(state['computation_resource_usage'], (1, E)), dtype=np.float64),
             np.array(np.reshape(state['viewer_connection_table'], (1, E * channel * version)),
                      dtype=np.float64),
             np.array(np.reshape(state['user_info'], (1, 4)), dtype=np.float64),
             np.array(np.reshape(state['qoe'], (1, 3)), dtype=np.float64)]
        actor_result = self.actor.get_action(x)
        actor_result = (np.clip(actor_result, -self.actor.action_bound, self.actor.action_bound) + self.actor.action_bound) / 2
        device_id = np.random.choice(np.where(actor_result == np.max(actor_result))[0])
        return actor_result, device_id


service_agent = ServiceAgent('env_deepcast-v0')
