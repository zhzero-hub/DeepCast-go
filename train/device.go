package train

const (
	EdgeNumber                    = 10
	EdgeCpuCore                   = 36                // 4 * 8
	EdgeInboundBandwidth          = 200 * 1024 * 1024 // Mbps -> bps
	EdgeOutboundBandwidth         = 400 * 1024 * 1024 // Mbps -> bps
	CdnCpuCore                    = 36
	CdnInboundBandwidth           = 20000 * 1024 * 1024 // Mbps -> bps
	CdnOutboundBandwidth          = 40000 * 1024 * 1024 // Mbps -> bps
	CdnToEdgeLatencyUpperLimit    = 0.1                 // s
	CdnToEdgeLatencyLowerLimit    = 0.02                // s
	ViewerToCdnLatencyUpperLimit  = 0.7
	ViewerToCdnLatencyLowerLimit  = 0.1
	ViewerToEdgeLatencyUpperLimit = 0.1
)

type Edge struct {
	DeviceCommon
}

type CDN struct {
	DeviceCommon
}

type System struct {
	Edge           map[string]*Edge
	Cdn            map[string]*CDN
	InboundMap     []*float64
	OutboundMap    []*float64
	ComputationMap []*float64
}
