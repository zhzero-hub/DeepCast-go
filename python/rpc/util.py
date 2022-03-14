channel_id_to_channel = {}
version_to_version_id = {240: 0, 360: 1, 480: 2, 720: 3, 1080: 4, 1440: 5}


def get_channel_id(channel_id):
    if channel_id_to_channel.get(channel_id) is None:
        channel_id_to_channel[channel_id] = len(channel_id_to_channel)
    return channel_id_to_channel[channel_id]


def get_version_id(version):
    return version_to_version_id[version]
