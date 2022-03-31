from flask import Flask, json, request
from rpc.web import *
from .util import *
import env.env as env

app = Flask(__name__)


@app.after_request
def cors(environ):
    environ.headers['Access-Control-Allow-Origin'] = '*'
    environ.headers['Access-Control-Allow-Method'] = '*'
    environ.headers['Access-Control-Allow-Headers'] = 'x-requested-with,content-type'
    return environ


@app.route('/getSystemDevices')
def get_system_devices():
    system = get_system_info(None)
    return json.dumps(system.SystemInfo, default=system_info_to_json)


@app.route('/getTaskManager')
def get_task_manager():
    task_manager = get_task_manager_info(None)
    return json.dumps(task_manager.TaskManagerInfo, default=task_manager_to_json)


@app.route('/setTrainStatus', methods=['POST'])
def set_train_status():
    status = request.json.get('status')
    env.STATUS = status
    return ''


@app.route('/getBackground')
def get_background():
    _type = request.args.get('type')
    extra = {}
    if _type is not None and _type == 'location':
        extra['location'] = ''
    background = get_background_info(None, extra=extra)
    return json.dumps(background.BackgroundInfo, default=background_info_to_json)


def start_web_server():
    app.run()
