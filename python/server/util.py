def device_to_json(device):
    return {
        'id': device.id,
        'name': device.name,
        'cpu_core': device.cpu_core,
        'location': {
            'latitude': device.location.latitude,
            'longitude': device.location.longitude,
        },
        'bandwidth_info': {
            'inbound_bandwidth_usage': device.band_width_info.inbound_bandwidth_usage,
            'outbound_bandwidth_usage': device.band_width_info.outbound_bandwidth_usage,
            'inbound_bandwidth_limit': device.band_width_info.inbound_bandwidth_limit,
            'outbound_bandwidth_limit': device.band_width_info.outbound_bandwidth_limit,
        },
        'latency_to_upper': device.latency_to_upper,
        'computation_usage': device.computation_usage,
    }


def system_info_to_json(system):
    if system is None:
        return ''
    js = {}
    for edge in system.edges:
        js[edge.name] = device_to_json(edge)
    for cdn in system.cdn:
        js[cdn.name] = device_to_json(cdn)
    return js


def user_info_to_json(user):
    return {
        'channel_id': user.channel_id,
        'version': user.version,
        'user_id': user.user_id,
        'location': {
            'latitude': user.location.latitude,
            'longitude': user.location.longitude,
        }
    }


def task_manager_to_json(task_manager):
    if task_manager is None:
        return ''
    js = {'time': task_manager.time}
    user_js = {}
    solve_js = {}
    for user in task_manager.user_info:
        user_js[user.user_id] = user_info_to_json(user)
    for solve in task_manager.solved:
        solve_js[solve.user_info.user_id] = {
            'user_info': user_info_to_json(solve.user_info),
            'device_name': solve.device_name,
        }
    js['user_info'] = user_js
    js['solved'] = solve_js
    return js


def background_info_to_json(background):
    if background is None:
        return ''
    js = {'time': background.time, 'maxTime': background.max_time}
    if background.location is not None:
        js['location'] = {
            'lat': background.location.latitude,
            'long': background.location.longitude
        }
    return js


def service_result_to_json(actor_result, device_id):
    result = []
    for value in actor_result:
        result.append(float(value))
    return {
        'result': result,
        'device_id': int(device_id)
    }
