package train

import (
	"context"
	"encoding/csv"
	"io"
	"log"
	"math/rand"
	"os"
	"sort"
	"strconv"
)

type UserViewingInfo struct {
	UserId    string
	ChannelId string
	LiverName string
	StartTime int64
	EndTime   int64
}

func LoadUserViewingDataset(ctx context.Context) (*map[string]*Viewer, error) {
	pwd, _ := os.Getwd()
	if csvData, err := ReadFromFile(pwd+"/data/", "user_viewing_dataset.csv", "csv"); err != nil {
		log.Fatalf("User_viewing_dataset 数据加载错误: %v", err)
		return nil, err
	} else {
		topLivers := GetTopLivers(ctx, csvData)
		var topLiverNames []string
		for k := range *topLivers {
			topLiverNames = append(topLiverNames, k)
		}
		topLiverIndex := 0
		viewerInfoMap := make(map[string]*Viewer, 0)
		for _, data := range csvData {
			csvRow := data
			var liverName, channelId string
			if _, ok := (*topLivers)[csvRow[2]]; ok {
				liverName = csvRow[2]
				channelId = csvRow[1]
			} else {
				liverName = topLiverNames[topLiverIndex]
				channelId = (*topLivers)[liverName]
				topLiverIndex++
				if topLiverIndex == len(topLiverNames) {
					topLiverIndex = 0
				}
			}
			startTimeInt, _ := strconv.ParseInt(csvRow[3], 10, 64)
			endTimeInt, _ := strconv.ParseInt(csvRow[4], 10, 64)
			userViewingInfo := UserViewingInfo{
				UserId:    csvRow[0],
				ChannelId: channelId,
				LiverName: liverName,
				StartTime: startTimeInt,
				EndTime:   endTimeInt,
			}
			if viewer, ok := viewerInfoMap[userViewingInfo.UserId]; !ok {
				liveInfo := make([]*LiveInfo, 0)
				viewerInfoMap[userViewingInfo.UserId] = &Viewer{
					Id: userViewingInfo.UserId,
					LiveInfo: append(liveInfo, &LiveInfo{
						ChannelId: userViewingInfo.ChannelId,
						LiverName: userViewingInfo.LiverName,
						StartTime: userViewingInfo.StartTime,
						EndTime:   userViewingInfo.EndTime,
					}),
				}
			} else {
				liveInfo := viewer.LiveInfo
				liveInfo = append(liveInfo, &LiveInfo{
					ChannelId: userViewingInfo.ChannelId,
					LiverName: userViewingInfo.LiverName,
					StartTime: userViewingInfo.StartTime,
					EndTime:   userViewingInfo.EndTime,
				})
			}
		}
		return &viewerInfoMap, nil
	}
}

func LoadUserLocationDataset(ctx context.Context) error {
	pwd, _ := os.Getwd()
	if csvData, err := ReadFromFile(pwd+"/data/", "user_location_dataset.csv", "csv"); err != nil {
		log.Fatalf("User_viewing_dataset 数据加载错误: %v", err)
		return err
	} else {
		userMap := ctx.Value("viewer").(*map[string]*Viewer)
		var locationInfo []struct {
			Lat float64
			Lot float64
		}
		for index := 1; index < len(csvData); index++ {
			longitude, _ := strconv.ParseFloat(csvData[index][3], 64)
			latitude, _ := strconv.ParseFloat(csvData[index][4], 64)
			if longitude > 1 && latitude > 1 {
				locationInfo = append(locationInfo, struct {
					Lat float64
					Lot float64
				}{Lat: latitude, Lot: longitude})
			}
		}
		index := 0
		for userId, viewer := range *userMap {
			viewer.Location = Location{
				Name: userId + "->" + strconv.FormatFloat(locationInfo[index].Lot, 'f', 2, 64) + ", " + strconv.FormatFloat(locationInfo[index].Lat, 'f', 2, 64),
				Lat:  locationInfo[index].Lat,
				Long: locationInfo[index].Lot,
			}
			index++
			if index == len(locationInfo) {
				index = 0
			}
		}
		return nil
	}
}

func LoadUserBandWidthDataset(ctx context.Context) error {
	pwd, _ := os.Getwd()
	if csvData, err := ReadFromFile(pwd+"/data/", "user_bandwidth_dataset.csv", "csv"); err != nil {
		log.Fatalf("User_viewing_dataset 数据加载错误: %v", err)
		return err
	} else {
		userMap := ctx.Value("viewer").(*map[string]*Viewer)
		userBandwidthInfo := make([]int64, 0)
		// userBandwidthMap := make(map[string]*[]int64, 0)
		for index := 0; index < len(csvData); index++ {
			bandWidth, _ := strconv.ParseInt(csvData[index][3], 10, 64)
			if bandWidth > 0 {
				userBandwidthInfo = append(userBandwidthInfo, bandWidth*8) // B/s -> bps
				//if userBandwidth, ok := userBandwidthMap[csvData[index][0]]; ok {
				//	*userBandwidth = append(*userBandwidth, bandWidth*8)
				//} else {
				//	b := make([]int64, 0)
				//	b = append(b, bandWidth*8)
				//	userBandwidthMap[csvData[index][0]] = &b
				//}
			}
		}
		index := 0
		for _, viewer := range *userMap {
			viewer.DownThroughput = userBandwidthInfo[index]
			index++
			if index == len(userBandwidthInfo) {
				index = 0
			}
		}
		return nil
	}
}

func GetTopLivers(ctx context.Context, data [][]string) *map[string]string {
	liverMap := make(map[string]int64, 0)
	channelMap := make(map[string]string, 0)
	for _, csvData := range data {
		if count, ok := liverMap[csvData[2]]; !ok {
			liverMap[csvData[2]] = 1
		} else {
			liverMap[csvData[2]] = count + 1
		}
		if _, ok := channelMap[csvData[2]]; !ok {
			channelMap[csvData[2]] = csvData[1]
		}
	}
	var topCount []struct {
		Liver string
		Count int64
	}
	for liver, count := range liverMap {
		topCount = append(topCount, struct {
			Liver string
			Count int64
		}{Liver: liver, Count: count})
	}
	sort.Slice(topCount, func(i, j int) bool {
		return topCount[i].Count > topCount[j].Count
	})
	topLiverMap := make(map[string]string, 0)
	for index := 0; index < channels; index++ {
		topLiverMap[topCount[index].Liver] = channelMap[topCount[index].Liver]
	}
	return &topLiverMap
}

func ReadFromFile(filePath string, fileName string, fileType string) ([][]string, error) {
	if file, err := os.Open(filePath + "/" + fileName); err != nil {
		return nil, err
	} else {
		defer func(file *os.File) {
			err := file.Close()
			if err != nil {
				log.Fatalf("%v", err)
				return
			}
		}(file)
		switch fileType {
		case "csv":
			csvReader := csv.NewReader(file)
			var csvData [][]string
			for {
				if row, err := csvReader.Read(); err == io.EOF {
					break
				} else if err != nil {
					return nil, err
				} else {
					csvData = append(csvData, row)
				}
			}
			return csvData, nil
		}
	}
	return nil, nil
}

func InitEdgeSystemInfo(ctx context.Context) *System {
	var system System
	edgeMap := make(map[string]*Edge, 0)
	inboundBandPointer := make([]*float64, 0)
	outboundBandPointer := make([]*float64, 0)
	computationPointer := make([]*float64, 0)
	for index := 0; index < EdgeNumber; index++ {
		edge := Edge{
			DeviceCommon{
				Id:      int32(index),
				Name:    "Edge" + strconv.Itoa(index),
				CpuCore: EdgeCpuCore,
				BandWidthInfo: BandWidthInfo{
					InBandWidthLimit:  EdgeInboundBandwidth,
					OutBandWidthLimit: EdgeOutboundBandwidth,
					InBandWidthUsed:   0,
					OutBandWidthUsed:  0,
				},
				LatencyToUpper:  rand.Float64()*(EdgeToCdnLatencyUpperLimit-EdgeToCdnLatencyLowerLimit) + EdgeToCdnLatencyLowerLimit,
				ComputationUsed: 0,
			},
		}
		edgeMap["Edge"+strconv.Itoa(index)] = &edge
		inboundBandPointer = append(inboundBandPointer, &edge.BandWidthInfo.InBandWidthUsed)
		outboundBandPointer = append(outboundBandPointer, &edge.BandWidthInfo.OutBandWidthUsed)
		computationPointer = append(computationPointer, &edge.ComputationUsed)
	}
	cdn := CDN{
		DeviceCommon{
			Id:      int32(0),
			Name:    "Cdn" + strconv.Itoa(EdgeNumber),
			CpuCore: CdnCpuCore,
			BandWidthInfo: BandWidthInfo{
				InBandWidthLimit:  CdnInboundBandwidth,
				OutBandWidthLimit: CdnOutboundBandwidth,
				InBandWidthUsed:   0,
				OutBandWidthUsed:  0,
			},
			LatencyToUpper:  0,
			ComputationUsed: 0,
		},
	}
	outboundBandPointer = append(outboundBandPointer, &cdn.BandWidthInfo.OutBandWidthUsed)

	system.Cdn = make(map[string]*CDN, 0)
	system.Cdn["Cdn"+strconv.Itoa(EdgeNumber)] = &cdn
	system.Edge = edgeMap
	system.InboundMap = inboundBandPointer
	system.OutboundMap = outboundBandPointer
	system.ComputationMap = computationPointer
	return &system
}
