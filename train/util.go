package train

import (
	"context"
	"github.com/muesli/clusters"
	"github.com/muesli/kmeans"
	"log"
	"math"
	"math/rand"
	"strconv"
	"strings"
)

func ChooseEdgeLocationWithKMeans(ctx *context.Context) {
	viewer := (*ctx).Value("viewer").(*ViewerInfo)
	system := (*ctx).Value("system").(*System)

	var d clusters.Observations
	for _, viewerInfo := range viewer.viewer {
		d = append(d, clusters.Coordinates{
			viewerInfo.Location.Lat,
			viewerInfo.Location.Long,
		})
	}

	// Partition the data points into 16 clusters
	km := kmeans.New()
	if cluster, err := km.Partition(d, len(system.Cdn)+len(system.Edge)); err != nil {
		log.Fatalf("k-means生成edge location信息失败: %v\n", err)
	} else {
		for i, c := range cluster {
			if i < len(system.Edge) {
				system.Edge["Edge"+strconv.FormatInt(int64(i), 10)].Location = Location{
					Name: "Cluster" + strconv.FormatInt(int64(i), 10),
					Lat:  c.Center[0],
					Long: c.Center[1],
				}
			} else {
				system.Cdn["Cdn"+strconv.FormatInt(int64(i), 10)].Location = Location{
					Name: "Cluster" + strconv.FormatInt(int64(i), 10),
					Lat:  c.Center[0],
					Long: c.Center[1],
				}
			}
			//log.Printf("Centered at x: %.2f y: %.2f\n", c.Center[0], c.Center[1])
			//log.Printf("Matching data points: %+v\n\n", c.Observations)
		}
	}
}

func GenerateRandomUserLocation(ctx *context.Context) {
	viewer := (*ctx).Value("viewer").(*ViewerInfo)
	var d clusters.Observations
	for _, viewerInfo := range viewer.viewer {
		d = append(d, clusters.Coordinates{
			viewerInfo.Location.Lat,
			viewerInfo.Location.Long,
		})
	}
	// Partition the data points into 16 clusters
	km := kmeans.New()
	if cluster, err := km.Partition(d, 50); err != nil {
		log.Fatalf("k-means生成edge location信息失败: %v\n", err)
	} else {
		locations := make([]Location, 0)
		for i, c := range cluster {
			locations = append(locations, Location{
				Name: "UserLocation" + strconv.FormatInt(int64(i), 10),
				Lat:  c.Center[0],
				Long: c.Center[1],
			})
		}
		*ctx = context.WithValue(*ctx, "locations", locations)
	}
}

func (e *Edge) TranscodingLatencyCal(viewer *Viewer) float64 {
	if _, ok := e.rates[viewer.AssignInfo.ChannelId]; !ok {
		// 没有这个channelId, 把这个channel的1440 version拿进来
		// TODO: 没人看了应该要拿出去
		v := make([]VersionInfo, 0)
		v = append(v, VersionInfo{version: 1440, number: 1})
		e.rates[viewer.AssignInfo.ChannelId] = &v
		e.BandWidthInfo.InBandWidthUsed += BitRateMap[1440]
		if e.BandWidthInfo.InBandWidthUsed > e.BandWidthInfo.InBandWidthLimit {
			// 应该不会触发
			log.Fatalf("超过边缘节点的带宽限制, 边缘节点名称: %s, 带宽限制: %f, 已使用: %f\n", e.Name, e.BandWidthInfo.InBandWidthLimit, e.BandWidthInfo.InBandWidthUsed)
		}
	}
	availableVersion := int64(1440)
	for _, info := range *e.rates[viewer.AssignInfo.ChannelId] {
		if info.version == viewer.AssignInfo.Version { // 有这个channelId对应的version
			availableVersion = info.version
			info.number++
			break
			//} else if availableVersion > viewer.AssignInfo.Version && info.version > availableVersion { // 没有，
			//	availableVersion = info.version
		}
	}
	if availableVersion == viewer.AssignInfo.Version {
		return 0
	} else if availableVersion > viewer.AssignInfo.Version {
		e.ComputationUsed += TransCodingCpuMap[viewer.AssignInfo.Version]
		// e.BandWidthInfo.OutBandWidthUsed += BitRateMap[viewer.AssignInfo.Version]
		*e.rates[viewer.AssignInfo.ChannelId] = append(*e.rates[viewer.AssignInfo.ChannelId], VersionInfo{version: availableVersion, number: 1})
		return TransCodingTimeMap[viewer.AssignInfo.Version]
	} else {
		log.Fatalf("默认edge能够拿到所有channel的1440 version\n")
		return 0
	}
}

func (l *Location) DistanceCal(other *Location) float64 {
	return math.Sqrt(math.Pow(l.Lat-other.Lat, 2)+math.Pow(l.Long-other.Long, 2)) * 1000
}

func (device *DeviceCommon) ViewerLatencyCal(v *Viewer) float64 {
	if strings.Contains(device.Name, "Cdn") {
		v.Latency = rand.Float64()*(ViewerToCdnLatencyUpperLimit-ViewerToCdnLatencyLowerLimit) + ViewerToCdnLatencyLowerLimit
	} else if strings.Contains(device.Name, "Edge") {
		if latency := v.Location.DistanceCal(&device.Location) * LatencyInGeo; latency > ViewerToEdgeLatencyUpperLimit {
			v.Latency = ViewerToEdgeLatencyUpperLimit
		} else {
			v.Latency = latency
		}
	}
	return v.Latency
}

func (s *System) RemoveViewer(viewer *Viewer) {
	// 要做的事: device对应的资源要补回去 channel对应的version如果没人看了要删掉
	channelId := viewer.AssignInfo.ChannelId
	version := viewer.AssignInfo.Version
	deviceId := viewer.AssignInfo.DeviceId
	if strings.Contains(deviceId, "Edge") {
		edge := s.Edge[deviceId]
		for index, info := range *edge.rates[channelId] {
			if info.version == version {
				if info.number == int64(1) { // 这是最后一个人
					edge.ComputationUsed -= TransCodingCpuMap[version]
					edge.BandWidthInfo.OutBandWidthUsed -= BitRateMap[version]
					seq := append((*edge.rates[channelId])[:index], (*edge.rates[channelId])[index+1:]...)
					edge.rates[channelId] = &seq
				} else {
					info.number--
				}
				break
			}
		}

	}
}
