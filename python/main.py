import sys
from threading import Thread
import time

sys.path.append('./A3C')
sys.path.append('./rpc')
sys.path.append('./env')

# from agent import main
from agentA3CContinue import main
from server.server import start_web_server

if __name__ == '__main__':
    t1 = Thread(target=main)
    t2 = Thread(target=start_web_server)
    t1.start()
    t2.start()
    t1.join()
    t2.join()
