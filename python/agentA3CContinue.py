# import wandb
import time

import tensorflow as tf
from tensorflow.keras.layers import Input, Dense, Lambda, InputLayer, concatenate
from env.env import E, channel, version
import env.env as env

import gym
import argparse
import numpy as np
from threading import Thread
import sys
from rpc.train import close
from multiprocessing import cpu_count

tf.keras.backend.set_floatx('float64')
# wandb.init(name='A3C', project="deep-rl-tf2")

parser = argparse.ArgumentParser()
parser.add_argument('--gamma', type=float, default=0.99)
parser.add_argument('--update_interval', type=int, default=5)
parser.add_argument('--actor_lr', type=float, default=0.0005)
parser.add_argument('--critic_lr', type=float, default=0.001)

args = parser.parse_args()

CUR_EPISODE = 0


def get_input_model():
    inbound_bandwidth_used_model = tf.keras.Sequential([
        InputLayer(E, name='inbound_bandwidth_used'),
        # Activation('relu')
    ])
    outbound_bandwidth_used_model = tf.keras.Sequential([
        InputLayer(E + 1, name='outbound_bandwidth_used'),
        # Activation('relu')
    ])
    computation_resource_usage_model = tf.keras.Sequential([
        InputLayer(E, name='computation_resource_usage'),
        # Activation('relu')
    ])
    viewer_connection_table_model = tf.keras.Sequential([
        InputLayer(E * channel * version, name='viewer_connection_table'),
        # Activation('relu')
    ])
    viewer_request_model = tf.keras.Sequential([
        InputLayer(4, name='user_info'),
        # Activation('relu')
    ])
    qoe_model = tf.keras.Sequential([
        InputLayer(3, name='qoe'),
        # Activation('relu')
    ])
    model = concatenate([inbound_bandwidth_used_model.output,
                         outbound_bandwidth_used_model.output,
                         computation_resource_usage_model.output,
                         viewer_connection_table_model.output,
                         viewer_request_model.output,
                         qoe_model.output], name='model_concatenate')
    model.input = [inbound_bandwidth_used_model.input,
                   outbound_bandwidth_used_model.input,
                   computation_resource_usage_model.input,
                   viewer_connection_table_model.input,
                   viewer_request_model.input,
                   qoe_model.input]
    return model


class Actor:
    def __init__(self, state_dim, action_dim, action_bound, std_bound):
        self.state_dim = state_dim
        self.action_dim = E + 1
        self.action_bound = action_bound
        self.std_bound = std_bound
        self.model = self.create_model()
        self.opt = tf.keras.optimizers.Adam(args.actor_lr)
        self.entropy_beta = 0.01

    def create_model(self):
        state_input = get_input_model()
        dense_1 = Dense(1024, activation='relu')(state_input)
        dense_2 = Dense(512, activation='relu')(dense_1)
        out_mu = Dense(self.action_dim, activation='tanh')(dense_2)
        mu_output = Lambda(lambda x: x * self.action_bound)(out_mu)
        std_output = Dense(self.action_dim, activation='softplus')(dense_2)
        return tf.keras.models.Model(state_input.input, [mu_output, std_output])

    def get_action(self, state):
        # state = np.reshape(state, [1, self.state_dim])
        mu, std = self.model.predict(state)
        mu, std = mu[0], std[0]
        return np.random.normal(mu, std, size=self.action_dim)

    def log_pdf(self, mu, std, action):
        std = tf.clip_by_value(std, self.std_bound[0], self.std_bound[1])
        var = std ** 2
        log_policy_pdf = -0.5 * (action - mu) ** 2 / \
                         var - 0.5 * tf.math.log(var * 2 * np.pi)
        return tf.reduce_sum(log_policy_pdf, 1, keepdims=True)

    def compute_loss(self, mu, std, actions, advantages):
        log_policy_pdf = self.log_pdf(mu, std, actions)
        loss_policy = log_policy_pdf * advantages
        return tf.reduce_sum(-loss_policy)

    def train(self, states, actions, advantages):
        with tf.GradientTape() as tape:
            mus = []
            stds = []
            for state in states:
                mu, std = self.model(state, training=True)
                mus.append(mu)
                stds.append(stds)
            loss = self.compute_loss(tf.convert_to_tensor(mu), tf.convert_to_tensor(std), actions, advantages)
        grads = tape.gradient(loss, self.model.trainable_variables)
        self.opt.apply_gradients(zip(grads, self.model.trainable_variables))
        return loss


class Critic:
    def __init__(self, state_dim):
        self.state_dim = state_dim
        self.model = self.create_model()
        self.opt = tf.keras.optimizers.Adam(args.critic_lr)

    def create_model(self):
        model = get_input_model()
        z = Dense(1024, activation='relu')(model)
        z = Dense(512, activation='relu')(z)
        z = Dense(32, activation='relu')(z)
        z = Dense(1, activation='linear', name='output')(z)
        z = tf.keras.Model(inputs=model.input, outputs=z)
        return z

    def compute_loss(self, v_pred, td_targets):
        mse = tf.keras.losses.MeanSquaredError()
        return mse(td_targets, v_pred)

    def train(self, states, td_targets):
        with tf.GradientTape() as tape:
            v_preds = []
            for state in states:
                v_pred = self.model(state, training=True)
                v_preds.append(v_pred)
            v_preds = tf.convert_to_tensor(v_preds)
            td_targets = tf.reshape(td_targets, shape=v_preds.shape)
            assert v_preds.shape == td_targets.shape
            loss = self.compute_loss(v_preds, tf.stop_gradient(td_targets))
        grads = tape.gradient(loss, self.model.trainable_variables)
        self.opt.apply_gradients(zip(grads, self.model.trainable_variables))
        return loss


class Agent:
    def __init__(self, env_name):
        env = gym.make(env_name)
        self.env_name = env_name
        self.state_dim = env.observation_space.shape
        self.action_dim = env.action_space.shape
        self.action_bound = 1
        self.std_bound = [1e-2, 1.0]

        self.global_actor = Actor(
            self.state_dim, self.action_dim, self.action_bound, self.std_bound)
        self.global_critic = Critic(self.state_dim)
        self.num_workers = 1

    def train(self, max_episodes=1):
        workers = []

        for i in range(self.num_workers):
            env = gym.make(self.env_name)
            workers.append(WorkerAgent(
                env, self.global_actor, self.global_critic, max_episodes))

        for worker in workers:
            worker.start()

        for worker in workers:
            worker.join()


class WorkerAgent(Thread):
    def __init__(self, env, global_actor, global_critic, max_episodes):
        Thread.__init__(self)
        self.env = env
        self.state_dim = self.env.observation_space.shape
        self.action_dim = E + 1
        self.action_bound = 1
        self.std_bound = [1e-2, 1.0]

        self.max_episodes = max_episodes
        self.global_actor = global_actor
        self.global_critic = global_critic
        self.actor = Actor(self.state_dim, self.action_dim,
                           self.action_bound, self.std_bound)
        self.critic = Critic(self.state_dim)

        self.actor.model.set_weights(self.global_actor.model.get_weights())
        self.critic.model.set_weights(self.global_critic.model.get_weights())

    def n_step_td_target(self, rewards, next_v_value, done):
        td_targets = np.zeros_like(rewards)
        cumulative = 0
        if not done:
            cumulative = next_v_value

        for k in reversed(range(0, len(rewards))):
            cumulative = args.gamma * cumulative + rewards[k]
            td_targets[k] = cumulative
        return td_targets

    def advatnage(self, td_targets, baselines):
        return td_targets - baselines

    def list_to_batch(self, list):
        batch = list[0]
        for elem in list[1:]:
            batch = np.append(batch, elem, axis=0)
        return batch

    def train(self):
        global CUR_EPISODE

        while self.max_episodes >= CUR_EPISODE:
            state_batch = []
            action_batch = []
            reward_batch = []
            episode_reward, done = 0, False

            state = self.env.reset()

            while not done:
                while True:
                    if not env.STATUS:
                        time.sleep(1)
                    else:
                        break
                # self.env.render()
                x = [np.array(np.reshape(state['inbound_bandwidth_used'], (1, E)), dtype=np.float64) / 1024 / 1024,
                     np.array(np.reshape(state['outbound_bandwidth_used'], (1, E + 1)), dtype=np.float64) / 1024 / 1024,
                     np.array(np.reshape(state['computation_resource_usage'], (1, E)), dtype=np.float64),
                     np.array(np.reshape(state['viewer_connection_table'], (1, E * channel * version)),
                              dtype=np.float64),
                     np.array(np.reshape(state['user_info'], (1, 4)), dtype=np.float64),
                     np.array(np.reshape(state['qoe'], (1, 3)), dtype=np.float64)]
                device_id = self.actor.get_action(x)
                device_id = np.clip(device_id, -self.action_bound, self.action_bound)
                device_id = np.random.choice(np.where(device_id == np.max(device_id))[0])

                print('Action: {}, Channel id: {}, Version: {}'.format(device_id, state['channel_id'], state['version']))

                action = {'device_id': device_id, 'viewer_id': state['viewer_id'], 'channel_id': state['channel_id'],
                          'qoe': state['qoe'], 'version': state['version']}
                next_state, reward, done, _ = self.env.step(action)

                # state = np.reshape(state, [1, self.state_dim])
                action = np.reshape(device_id, [1, 1])
                # next_state = np.reshape(next_state, [1, self.state_dim])
                reward = np.reshape(reward, [1, 1])

                state_batch.append(x)
                action_batch.append(action)
                reward_batch.append(reward)

                if len(state_batch) >= args.update_interval and next_state is not None:
                    # states = self.list_to_batch(state_batch)
                    states = state_batch
                    actions = self.list_to_batch(action_batch)
                    rewards = self.list_to_batch(reward_batch)

                    y = [np.array(np.reshape(next_state['inbound_bandwidth_used'], (1, E))) / 1024 / 1024,
                         np.array(np.reshape(next_state['outbound_bandwidth_used'], (1, E + 1))) / 1024 / 1024,
                         np.array(np.reshape(next_state['computation_resource_usage'], (1, E))),
                         np.array(np.reshape(next_state['viewer_connection_table'], (1, E * channel * version))),
                         np.array(np.reshape(next_state['user_info'], (1, 4))),
                         np.array(np.reshape(next_state['qoe'], (1, 3)))]
                    next_v_value = self.critic.model.predict(y)
                    td_targets = self.n_step_td_target(
                        (rewards + 8) / 8, next_v_value, done)
                    critic = []
                    for state in states:
                        critic.append(self.critic.model(state)[0][0])
                    critic = tf.convert_to_tensor(critic)
                    advantages = td_targets - critic

                    actor_loss = self.global_actor.train(
                        states, actions, advantages)
                    critic_loss = self.global_critic.train(
                        states, td_targets)

                    self.actor.model.set_weights(
                        self.global_actor.model.get_weights())
                    self.critic.model.set_weights(
                        self.global_critic.model.get_weights())

                    state_batch = []
                    action_batch = []
                    reward_batch = []
                    td_target_batch = []
                    advatnage_batch = []

                episode_reward += reward[0][0]
                state = next_state

                print('EP{} EpisodeReward={}'.format(CUR_EPISODE, episode_reward))
                # wandb.log({'Reward': episode_reward})
                CUR_EPISODE += 1
            self.actor.model.save_weights('model/actor_model.h5')
            self.critic.model.save('model/critic_model.h5')

    def run(self):
        self.train()


def main():
    output = sys.stdout
    outputfile = open('output/stdout', 'w')
    sys.stdout = outputfile

    env_name = 'env_deepcast-v0'
    agent = Agent(env_name)
    agent.train()

    close()
    outputfile.close()
    sys.stdout = output
