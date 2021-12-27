package train

type Edge struct {
	DeviceCommon
}

type CDN struct {
	DeviceCommon
}

type DeviceSystem struct {
	edge Edge
	cdns map[string]CDN
	viewers map[Viewer]CDN
}
