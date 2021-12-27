package train

const (
	BitRate1440 = 4.3
	BitRate1080 = 2.85
	BitRate720  = 1.85
	BitRate480  = 1.2
	BitRate360  = 0.75
	BitRate240  = 0.3
)

const (
	TransCodingCpu1440 = 0
	TransCodingCpu1080 = 3.3
	TransCodingCpu720  = 1.42
	TransCodingCpu480  = 0.82
	TransCodingCpu360  = 0.51
	TransCodingCpu240  = 0.41
)

const (
	TransCodingTime1440 = 0
	TransCodingTime1080 = 0.27
	TransCodingTime720  = 0.19
	TransCodingTime480  = 0.16
	TransCodingTime360  = 0.13
	TransCodingTime240  = 0.11
)

type Location struct {
	Name string
	Lat  float64
	Long float64
}

type BandWidthInfo struct {
	InBandWidthLimit float64
	OutBandWidthLimit float64
	InBandWidthUsed float64
	OutBandWidthUsed float64
}

type DeviceCommon struct {
	Id        string
	Name      string
	CpuCore   int32
	Location  Location
	BandWidthInfo BandWidthInfo
}

type Viewer struct {
	ID        string
	Name      string
	Location  Location
	Latency  float64
}
