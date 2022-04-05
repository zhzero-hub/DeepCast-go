channel_id_to_channel = {}
version_to_version_id = {240: 0, 360: 1, 480: 2, 720: 3, 1080: 4, 1440: 5}


def get_channel_id(channel_id):
    if channel_id_to_channel.get(channel_id) is None:
        channel_id_to_channel[channel_id] = len(channel_id_to_channel)
    return channel_id_to_channel[channel_id]


def get_version_id(version):
    return version_to_version_id[version]


def convert_state(resp):
    E = int(resp.Base.Extra['E'])
    channels = int(resp.Base.Extra['channel'])
    versions = int(resp.Base.Extra['versions'])

    state = {}
    inbound_bandwidth_used = []
    outbound_bandwidth_used = []
    computation_resource_usage = []
    for inbound in resp.State.inbound_bandwidth_usage.inbound_bandwidth_usage:
        inbound_bandwidth_used.append(inbound)
    state['inbound_bandwidth_used'] = inbound_bandwidth_used
    # state.append(inbound_bandwidth_used)
    for outbound in resp.State.outbound_bandwidth_usage.outbound_bandwidth_usage:
        outbound_bandwidth_used.append(outbound)
    state['outbound_bandwidth_used'] = outbound_bandwidth_used
    # state.append(outbound_bandwidth_used)
    for compute in resp.State.computation_resource_usage.computation_resource_usage:
        computation_resource_usage.append(compute)
    state['computation_resource_usage'] = computation_resource_usage
    # state.append(computation_resource_usage)
    viewer_connection_table = []
    for _ in range(E * channels * versions):
        viewer_connection_table.append(0)
    for edge, h2v in resp.State.viewer_connection.viewer_connection_table.items():
        _edge = int(edge)
        for channel_id, v2number in h2v.h2v.items():
            _channel = get_channel_id(channel_id)
            for version, number in v2number.number.items():
                _version = get_version_id(int(version))
                index = _edge * channels * versions + _channel * versions + _version
                viewer_connection_table[index] = number
    state['viewer_connection_table'] = viewer_connection_table
    # state.append(viewer_connection_table)
    user_info = [resp.State.user_info.location.latitude, resp.State.user_info.location.longitude,
                 get_channel_id(resp.State.user_info.channel_id), get_version_id(resp.State.user_info.version)]
    state['user_info'] = user_info
    # state.append(user_info)
    qoe = [resp.State.qoe_preference.alpha1, resp.State.qoe_preference.alpha2, resp.State.qoe_preference.alpha3]
    state['qoe'] = qoe
    # state.append(qoe)

    state['viewer_id'] = resp.State.user_info.user_id
    state['channel_id'] = resp.State.user_info.channel_id
    state['version'] = resp.State.user_info.version
    return state
