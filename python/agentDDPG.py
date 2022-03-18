import tensorflow as tf
from tensorflow.keras.layers import Input, Dense, Lambda, concatenate, InputLayer, Activation
from env.env import E, channel, version

import gym
import argparse
import numpy as np
import random
from collections import deque

tf.keras.backend.set_floatx('float64')

parser = argparse.ArgumentParser()
parser.add_argument('--gamma', type=float, default=0.99)
parser.add_argument('--actor_lr', type=float, default=0.0005)
parser.add_argument('--critic_lr', type=float, default=0.001)
parser.add_argument('--batch_size', type=int, default=64)
parser.add_argument('--tau', type=float, default=0.05)
parser.add_argument('--train_start', type=int, default=200)

args = parser.parse_args()


def get_input_model():
    inbound_bandwidth_used_model = tf.keras.Sequential([
        InputLayer(E, name='inbound_bandwidth_used'),
        Activation('relu')
    ])
    outbound_bandwidth_used_model = tf.keras.Sequential([
        InputLayer(E + 1, name='outbound_bandwidth_used'),
        Activation('relu')
    ])
    computation_resource_usage_model = tf.keras.Sequential([
        InputLayer(E, name='computation_resource_usage'),
        Activation('relu')
    ])
    viewer_connection_table_model = tf.keras.Sequential([
        InputLayer(E * channel * version, name='viewer_connection_table'),
        Activation('relu')
    ])
    viewer_request_model = tf.keras.Sequential([
        InputLayer(4, name='user_info'),
        Activation('relu')
    ])
    qoe_model = tf.keras.Sequential([
        InputLayer(3, name='qoe'),
        Activation('relu')
    ])
    model = concatenate([inbound_bandwidth_used_model.output,
                         outbound_bandwidth_used_model.output,
                         computation_resource_usage_model.output,
                         viewer_connection_table_model.output,
                         viewer_request_model.output,
                         qoe_model.output], name='model_concatenate')
    # model = [inbound_bandwidth_used_model, outbound_bandwidth_used_model, computation_resource_usage_model,
    #          viewer_connection_table_model, viewer_request_model, qoe_model]
    model.input = [inbound_bandwidth_used_model.input,
                   outbound_bandwidth_used_model.input,
                   computation_resource_usage_model.input,
                   viewer_connection_table_model.input,
                   viewer_request_model.input,
                   qoe_model.input]
    return model


class ReplayBuffer:
    def __init__(self, capacity=20000):
        self.buffer = deque(maxlen=capacity)

    def put(self, state, action, reward, next_state, done):
        self.buffer.append([state, action, reward, next_state, done])

    def sample(self):
        sample = random.sample(self.buffer, args.batch_size)
        states, actions, rewards, next_states, done = map(np.asarray, zip(*sample))
        states = np.array(states).reshape(args.batch_size, -1)
        next_states = np.array(next_states).reshape(args.batch_size, -1)
        return states, actions, rewards, next_states, done

    def size(self):
        return len(self.buffer)


class Actor:
    def __init__(self, state_dim, action_dim, action_bound):
        self.state_dim = state_dim
        self.action_dim = action_dim
        self.action_bound = action_bound
        self.model = self.create_model()
        self.opt = tf.keras.optimizers.Adam(args.actor_lr)

    def create_model(self):
        model = get_input_model()
        z = Dense(1024, activation='relu')(model)
        z = Dense(512, activation='relu')(z)
        z = Dense(self.action_dim, activation='tanh', name='output')(z)
        z = Lambda(lambda x: x * self.action_bound)(z)
        z = tf.keras.Model(inputs=model.input, outputs=z)
        return z

    def train(self, states, q_grads):
        with tf.GradientTape() as tape:
            grads = []
            for i in range(args.batch_size):
                state = states[i][0]
                x = [np.array(np.reshape(state['inbound_bandwidth_used'], (1, E))),
                     np.array(np.reshape(state['outbound_bandwidth_used'], (1, E + 1))),
                     np.array(np.reshape(state['computation_resource_usage'], (1, E))),
                     np.array(np.reshape(state['viewer_connection_table'], (1, 4000)), dtype=np.float64),
                     np.array(np.reshape(state['user_info'], (1, 4))),
                     np.array(np.reshape(state['qoe'], (1, 3)))]
                grads.append(self.model(x))
            grads = tf.convert_to_tensor(grads)
            grads = tf.squeeze(grads)
            trainable_variable = self.model.trainable_variables
            grads = tape.gradient(grads, self.model.trainable_variables, -q_grads)
        self.opt.apply_gradients(zip(grads, self.model.trainable_variables))

    def predict(self, state):
        y = [np.array(np.reshape(state['inbound_bandwidth_used'], (1, E))),
             np.array(np.reshape(state['outbound_bandwidth_used'], (1, E + 1))),
             np.array(np.reshape(state['computation_resource_usage'], (1, E))),
             np.array(np.reshape(state['viewer_connection_table'], (1, 4000))),
             np.array(np.reshape(state['user_info'], (1, 4))),
             np.array(np.reshape(state['qoe'], (1, 3)))]
        return self.model.predict(y)

    def get_action(self, state):
        x = [np.array(np.reshape(state['inbound_bandwidth_used'], (1, E)), dtype=np.float64),
             np.array(np.reshape(state['outbound_bandwidth_used'], (1, E + 1)), dtype=np.float64),
             np.array(np.reshape(state['computation_resource_usage'], (1, E)), dtype=np.float64),
             np.array(np.reshape(state['viewer_connection_table'], (1, 4000)), dtype=np.float64),
             np.array(np.reshape(state['user_info'], (1, 4)), dtype=np.float64),
             np.array(np.reshape(state['qoe'], (1, 3)), dtype=np.float64)]
        return self.model.predict(x)[0]


class Critic:
    def __init__(self, state_dim, action_dim):
        self.state_dim = state_dim
        self.action_dim = action_dim
        self.model = self.create_model()
        self.opt = tf.keras.optimizers.Adam(args.critic_lr)

    def create_model(self):
        model = get_input_model()
        s1 = Dense(64, activation='relu')(model)
        s2 = Dense(32, activation='relu')(s1)
        action_input = Input((self.action_dim,))
        a1 = Dense(32, activation='relu')(action_input)
        c1 = concatenate([s2, a1], axis=-1)
        c2 = Dense(16, activation='relu')(c1)
        output = Dense(1, activation='linear')(c2)
        return tf.keras.Model([model.input, action_input], output)

    def predict(self, state):
        z = [np.array(np.reshape(state['inbound_bandwidth_used'], (1, E))),
             np.array(np.reshape(state['outbound_bandwidth_used'], (1, E + 1))),
             np.array(np.reshape(state['computation_resource_usage'], (1, E))),
             np.array(np.reshape(state['viewer_connection_table'], (1, 4000))),
             np.array(np.reshape(state['user_info'], (1, 4))),
             np.array(np.reshape(state['qoe'], (1, 3))),
             np.array(np.reshape(state['input2'], (1, E + 1)))]
        return self.model.predict(z)

    def q_grads(self, states, actions):
        actions = tf.convert_to_tensor(actions)
        q_values = []
        with tf.GradientTape() as tape:
            tape.watch(actions)
            for i in range(args.batch_size):
                state = states[i][0]
                action = actions[i]
                x = [np.array(np.reshape(state['inbound_bandwidth_used'], (1, E))),
                     np.array(np.reshape(state['outbound_bandwidth_used'], (1, E + 1))),
                     np.array(np.reshape(state['computation_resource_usage'], (1, E))),
                     np.array(np.reshape(state['viewer_connection_table'], (1, 4000)), dtype=np.float64),
                     np.array(np.reshape(state['user_info'], (1, 4))),
                     np.array(np.reshape(state['qoe'], (1, 3))),
                     tf.reshape(action, (1, E + 1))]
                q_values.append(self.model(x))
            q_values = tf.convert_to_tensor(q_values)
            # q_values = tf.squeeze(q_values)
            return tape.gradient(q_values, actions)

    def compute_loss(self, v_pred, td_targets):
        mse = tf.keras.losses.MeanSquaredError()
        return mse(td_targets, v_pred)

    def train(self, states, actions, td_targets):
        with tf.GradientTape() as tape:
            v_preds = []
            for i in range(args.batch_size):
                state = states[i][0]
                action = actions[i]
                x = [np.array(np.reshape(state['inbound_bandwidth_used'], (1, E))),
                     np.array(np.reshape(state['outbound_bandwidth_used'], (1, E + 1))),
                     np.array(np.reshape(state['computation_resource_usage'], (1, E))),
                     np.array(np.reshape(state['viewer_connection_table'], (1, 4000)), dtype=np.float64),
                     np.array(np.reshape(state['user_info'], (1, 4))),
                     np.array(np.reshape(state['qoe'], (1, 3))),
                     np.array(np.reshape(action, (1, E + 1)))]
                v_pred = self.model(x, training=True)
                v_preds.append(v_pred)
            v_preds = tf.convert_to_tensor(v_preds)
            assert v_preds.shape == td_targets.shape
            loss = self.compute_loss(v_preds, tf.stop_gradient(td_targets))
            grads = tape.gradient(loss, self.model.trainable_variables)
            self.opt.apply_gradients(zip(grads, self.model.trainable_variables))
            return loss


class Agent:
    def __init__(self, env):
        self.env = env
        self.state_dim = E
        self.action_dim = E + 1
        self.action_bound = 1

        self.buffer = ReplayBuffer()

        self.actor = Actor(self.state_dim, self.action_dim, self.action_bound)
        self.critic = Critic(self.state_dim, self.action_dim)

        self.target_actor = Actor(self.state_dim, self.action_dim, self.action_bound)
        self.target_critic = Critic(self.state_dim, self.action_dim)

        actor_weights = self.actor.model.get_weights()
        critic_weights = self.critic.model.get_weights()
        self.target_actor.model.set_weights(actor_weights)
        self.target_critic.model.set_weights(critic_weights)

    def target_update(self):
        actor_weights = self.actor.model.get_weights()
        t_actor_weights = self.target_actor.model.get_weights()
        critic_weights = self.critic.model.get_weights()
        t_critic_weights = self.target_critic.model.get_weights()

        for i in range(len(actor_weights)):
            t_actor_weights[i] = args.tau * actor_weights[i] + (1 - args.tau) * t_actor_weights[i]

        for i in range(len(critic_weights)):
            t_critic_weights[i] = args.tau * critic_weights[i] + (1 - args.tau) * t_critic_weights[i]

        self.target_actor.model.set_weights(t_actor_weights)
        self.target_critic.model.set_weights(t_critic_weights)

    def td_target(self, rewards, q_values, dones):
        targets = np.asarray(q_values)
        for i in range(q_values.shape[0]):
            if dones[i]:
                targets[i] = rewards[i]
            else:
                targets[i] = args.gamma * q_values[i]
        return targets

    def list_to_batch(self, list):
        batch = list[0]
        for elem in list[1:]:
            batch = np.append(batch, elem, axis=0)
        return batch

    def ou_noise(self, x, rho=0.15, mu=0, dt=1e-1, sigma=0.2, dim=1):
        return x + rho * (mu - x) * dt + sigma * np.sqrt(dt) * np.random.normal(size=dim)

    def replay(self):
        for _ in range(10):
            states, actions, rewards, next_states, dones = self.buffer.sample()
            target_q_values = []
            for i in range(args.batch_size):
                next_state = next_states[i][0]
                target_actor_predict = self.target_actor.predict(next_state)
                next_state['input2'] = target_actor_predict
                target_critic_predict = self.target_critic.predict(next_state)
                target_q_values.append(target_critic_predict)
            # target_q_values = self.target_critic.predict([next_states, self.target_actor.predict(next_states)])
            target_q_values = np.asarray(target_q_values)
            td_targets = self.td_target(rewards, target_q_values, dones)

            self.critic.train(states, actions, td_targets)

            s_actions = []
            for i in range(args.batch_size):
                s_actions.append(self.actor.predict(states[i][0]))
            s_grads = self.critic.q_grads(states, s_actions)
            grads = np.array(s_grads).reshape((-1, self.action_dim))
            self.actor.train(states, grads)
            self.target_update()

    def train(self, max_episodes=1000):
        for ep in range(max_episodes):
            episode_reward, done = 0, False

            state = self.env.reset()
            bg_noise = np.zeros(self.action_dim)
            while not done:
                # self.env.render()
                action = self.actor.get_action(state)
                noise = self.ou_noise(bg_noise, dim=self.action_dim)
                action = np.clip(action + noise, -self.action_bound, self.action_bound)

                device_id = np.random.choice(np.where(action == np.max(action))[0])
                print('Action: device id {}'.format(device_id))
                step_action = {'device_id': device_id, 'viewer_id': state['viewer_id'],
                               'channel_id': state['channel_id'], 'qoe': state['qoe'], 'version': state['version']}
                next_state, reward, done, _ = self.env.step(step_action)
                self.buffer.put(state, action, (reward + 8) / 8, next_state, done)
                bg_noise = noise
                episode_reward += reward
                state = next_state
                # print('Buffer size{}'.format(self.buffer.size()))
                if self.buffer.size() >= args.batch_size and self.buffer.size() >= args.train_start:
                    self.replay()
            print('EP{} EpisodeReward={}'.format(ep, episode_reward))


def main():
    env_name = 'env_deepcast-v0'
    env = gym.make(env_name)
    agent = Agent(env)
    agent.train()
