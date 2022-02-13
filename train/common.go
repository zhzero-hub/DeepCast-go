package train

const channels = 50

const LatencyInGeo = 5 * 10e-6 // s 每1000m增加5us时延

var BitRateMap = map[int64]float64{
	1440: 4.3 * 1024 * 1024,
	1080: 2.85 * 1024 * 1024,
	720:  1.85 * 1024 * 1024,
	480:  1.2 * 1024 * 1024,
	360:  0.75 * 1024 * 1024,
	240:  0.3 * 1024 * 1024,
}

var TransCodingCpuMap = map[int64]float64{
	1440: 0,
	1080: 3.3,
	720:  1.42,
	480:  0.82,
	360:  0.51,
	240:  0.41,
}

var TransCodingTimeMap = map[int64]float64{
	1440: 0,
	1080: 0.27,
	720:  0.19,
	480:  0.16,
	360:  0.13,
	240:  0.11,
}

type Location struct {
	Name string
	Lat  float64
	Long float64
}

type BandWidthInfo struct {
	InBandWidthLimit  float64
	OutBandWidthLimit float64
	InBandWidthUsed   float64
	OutBandWidthUsed  float64
}

type DeviceCommon struct {
	Id              int32
	Name            string
	CpuCore         int32
	Location        Location
	BandWidthInfo   BandWidthInfo
	LatencyToUpper  float64 // s
	ComputationUsed float64
}

type AssignInfo struct {
	ChannelId string
	Version   int64
	DeviceId  string
}

type Viewer struct {
	Id             string
	Location       Location
	Latency        float64
	LiveInfo       []*LiveInfo
	DownThroughput float64 // bps
	AssignInfo     AssignInfo
}

type LiveInfo struct {
	ChannelId string
	LiverName string
	StartTime int64
	EndTime   int64
}
