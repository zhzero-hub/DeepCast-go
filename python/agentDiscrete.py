from tensorflow.keras.layers import concatenate
import tensorflow as tf
from tensorflow.keras.layers import InputLayer, Dense, Activation, Reshape
from env.env import E, channel, version

import gym
import argparse
import numpy as np
from threading import Thread, Lock
from multiprocessing import cpu_count
import env

tf.keras.backend.set_floatx('float64')

parser = argparse.ArgumentParser()
parser.add_argument('--gamma', type=float, default=0.99)
parser.add_argument('--update_interval', type=int, default=50)
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


def main():
    env_name = 'env_deepcast-v0'
    agent = Agent(env_name)
    agent.train()


class Actor:
    def __init__(self, state_dim, action_dim):
        self.state_dim = state_dim
        self.action_dim = E + 1
        self.model = self.create_model()
        self.opt = tf.keras.optimizers.Adam(args.actor_lr)
        self.entropy_beta = 0.01

    def create_model(self):
        model = get_input_model()
        z = Dense(4096, activation='relu')(model)
        z = Dense(2048, activation='relu')(z)
        z = Dense(self.action_dim, activation='softmax', name='output')(z)
        z = tf.keras.Model(inputs=model.input, outputs=z)
        return z

    def compute_loss(self, actions, logits, advantages):
        ce_loss = tf.keras.losses.CategoricalCrossentropy(
            from_logits=True)
        entropy_loss = tf.keras.losses.CategoricalCrossentropy(
            from_logits=True)
        # actions = tf.cast(actions, tf.int32)
        actions = tf.one_hot(actions, depth=E + 1)
        # actions = tf.reshape(actions, shape=(args.update_interval, E + 1))
        logits = tf.reshape(logits, shape=actions.shape)
        advantages = tf.reshape(advantages, shape=(args.update_interval, 1))
        policy_loss = ce_loss(actions, logits, sample_weight=tf.stop_gradient(advantages))
        entropy = entropy_loss(logits, logits)
        return policy_loss - self.entropy_beta * entropy

    def train(self, states, actions, advantages):
        with tf.GradientTape() as tape:
            logits = []
            for state in states:
                predict = self.model(state, training=True)
                logits.append(predict)
            loss = self.compute_loss(actions, logits, advantages)
            grads = tape.gradient(loss, self.model.trainable_variables)
        # print(grads)
        self.opt.apply_gradients(zip(grads, self.model.trainable_variables))
        return loss


class Critic:
    def __init__(self, state_dim):
        self.state_dim = 1
        self.model = self.create_model()
        self.opt = tf.keras.optimizers.Adam(args.critic_lr)

    def create_model(self):
        model = get_input_model()
        z = Dense(512, activation='relu')(model)
        z = Dense(256, activation='relu')(z)
        z = Dense(self.state_dim, activation='linear', name='output')(z)
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
        # self.state_dim = env.observation_space.shape[0].
        self.state_dim = 0
        for name, space in env.observation_space.spaces.items():
            if len(space.shape) > 0:
                tmp_state_dim = 1
                for shape in space.shape:
                    tmp_state_dim *= shape
                self.state_dim += tmp_state_dim
            else:
                self.state_dim += 1
        self.action_dim = env.action_space.shape[0]
        # self.action_dim = 100

        self.global_actor = Actor(self.state_dim, self.action_dim)
        self.global_critic = Critic(self.state_dim)
        # self.num_workers = cpu_count()
        self.num_workers = 1

    def train(self, max_episodes=1000):
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
        self.lock = Lock()
        self.env = env
        self.state_dim = 0
        for name, space in env.observation_space.spaces.items():
            if len(space.shape) > 0:
                tmp_state_dim = 1
                for shape in space.shape:
                    tmp_state_dim *= shape
                self.state_dim += tmp_state_dim
            else:
                self.state_dim += 1
        self.action_dim = env.action_space.shape[0]

        self.max_episodes = max_episodes
        self.global_actor = global_actor
        self.global_critic = global_critic
        self.actor = Actor(self.state_dim, self.action_dim)
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
            batch = np.append(batch, elem)
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
                # self.env.render()
                x = [np.array(np.reshape(state['inbound_bandwidth_used'], (1, E)), dtype=np.float64) / 1024 / 1024,
                     np.array(np.reshape(state['outbound_bandwidth_used'], (1, E + 1)), dtype=np.float64) / 1024 / 1024,
                     np.array(np.reshape(state['computation_resource_usage'], (1, E)), dtype=np.float64),
                     np.array(np.reshape(state['viewer_connection_table'], (1, 4000)), dtype=np.float64),
                     np.array(np.reshape(state['user_info'], (1, 4)), dtype=np.float64),
                     np.array(np.reshape(state['qoe'], (1, 3)), dtype=np.float64)]
                probs = self.actor.model.predict(x)
                print('Current episode: {}\nprobs: {}'.format(CUR_EPISODE, probs))
                device_id = np.random.choice(E + 1, p=probs[0])

                action = {'device_id': device_id, 'viewer_id': state['viewer_id'], 'channel_id': state['channel_id'],
                          'qoe': state['qoe'], 'version': state['version']}
                next_state, reward, done, _ = self.env.step(action)

                # state = np.reshape(state, [1, self.state_dim])
                # device_id = np.reshape(device_id, [1, 1])
                # next_state = np.reshape(next_state, [1, self.state_dim])
                reward = np.reshape(reward, [1, 1])

                state_batch.append(x)
                action_batch.append(device_id)
                reward_batch.append(reward)

                if len(state_batch) >= args.update_interval or done:
                    states = state_batch
                    actions = self.list_to_batch(action_batch)
                    rewards = self.list_to_batch(reward_batch)

                    y = [np.array(np.reshape(next_state['inbound_bandwidth_used'], (1, E))) / 1024 / 1024,
                         np.array(np.reshape(next_state['outbound_bandwidth_used'], (1, E + 1))) / 1024 / 1024,
                         np.array(np.reshape(next_state['computation_resource_usage'], (1, E))),
                         np.array(np.reshape(next_state['viewer_connection_table'], (1, 4000))),
                         np.array(np.reshape(next_state['user_info'], (1, 4))),
                         np.array(np.reshape(next_state['qoe'], (1, 3)))]
                    next_v_value = self.critic.model.predict(y)
                    td_targets = self.n_step_td_target(rewards, next_v_value, done)
                    critic = []
                    for state in states:
                        critic.append(self.critic.model(state)[0][0])
                    critic = tf.convert_to_tensor(critic)
                    advantages = td_targets - critic

                    with self.lock:
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
                CUR_EPISODE += 1

            print('EP{} EpisodeReward={}'.format(CUR_EPISODE, episode_reward))

    def run(self):
        self.train()
