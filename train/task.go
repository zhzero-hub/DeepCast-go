package train

import (
	rpc "DeepCast/grpc"
	"container/heap"
	"context"
	"log"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"
)

type ViewerWithWatchChannel struct {
	*Viewer
	*LiveInfo
}

type ViewerHeap []*ViewerWithWatchChannel

func (h ViewerHeap) Len() int {
	return len(h)
}

func (h ViewerHeap) Less(i, j int) bool {
	return h[i].EndTime < h[j].EndTime
}

func (h ViewerHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *ViewerHeap) Push(x interface{}) {
	*h = append(*h, x.(*ViewerWithWatchChannel))
}

func (h *ViewerHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

func (h *ViewerHeap) Top() *ViewerWithWatchChannel {
	return (*h)[0]
}

type Task struct {
	watchInfo *ViewerWithWatchChannel
}

func (t *Task) GetTask() *ViewerWithWatchChannel {
	return t.watchInfo
}

type TaskManager struct {
	ctx        *context.Context
	time       int64
	task       map[int64][]*Task
	taskLock   sync.RWMutex
	viewerList ViewerHeap
	solved     map[*ViewerWithWatchChannel]*DeviceCommon
	solvedLock sync.RWMutex
	maxTime    int64
}

func (t *TaskManager) Lock() {
	t.taskLock.Lock()
	t.solvedLock.Lock()
}

func (t *TaskManager) Unlock() {
	t.taskLock.Unlock()
	t.solvedLock.Unlock()
}

func (t *TaskManager) GetTime() int64 {
	return t.time
}

func (t *TaskManager) GetMaxTime() int64 {
	return t.maxTime
}

func (t *TaskManager) GetAllTasks() []*Task {
	var tasks []*Task
	for _time := int64(0); _time < t.maxTime; _time++ {
		if _, ok := t.task[_time]; ok {
			for _, v := range t.task[_time] {
				tasks = append(tasks, v)
			}
		}
	}
	return tasks
}

func (t *TaskManager) GetAllSolved() map[*ViewerWithWatchChannel]string {
	solved := make(map[*ViewerWithWatchChannel]string)
	for u, v := range t.solved {
		solved[u] = v.Name
	}
	return solved
}

func InitTaskManager(ctx *context.Context) {
	viewerWithWatchChannel := make(ViewerHeap, 0)
	heap.Init(&viewerWithWatchChannel)
	maxEndTime := int64(0)
	viewers := (*ctx).Value("viewer").(*ViewerInfo)
	for _, viewer := range viewers.viewer {
		for _, live := range viewer.LiveInfo {
			if live.EndTime > maxEndTime {
				maxEndTime = live.EndTime
			}
		}
	}
	taskManager := TaskManager{
		ctx:        ctx,
		time:       -1,
		task:       make(map[int64][]*Task),
		taskLock:   sync.RWMutex{},
		viewerList: viewerWithWatchChannel,
		solved:     make(map[*ViewerWithWatchChannel]*DeviceCommon),
		solvedLock: sync.RWMutex{},
		maxTime:    maxEndTime,
	}
	*ctx = context.WithValue(*ctx, "taskManager", &taskManager)
}

func (t *TaskManager) AddTask(viewer *Viewer, liveInfo *LiveInfo) {
	startTime := liveInfo.StartTime
	if _, ok := t.task[startTime]; !ok {
		task := make([]*Task, 0)
		task = append(task, &Task{
			watchInfo: &ViewerWithWatchChannel{
				Viewer:   viewer,
				LiveInfo: liveInfo,
			},
		})
		t.task[startTime] = task
	} else {
		task := t.task[startTime]
		task = append(task, &Task{
			watchInfo: &ViewerWithWatchChannel{
				Viewer:   viewer,
				LiveInfo: liveInfo,
			},
		})
		t.task[startTime] = task
	}
}

func (t *TaskManager) RefreshTasks() {
	count := 0
	for viewerWithWatchInfo, device := range t.solved {
		viewer := viewerWithWatchInfo.Viewer
		liveInfo := viewerWithWatchInfo.LiveInfo
		if t.time >= liveInfo.EndTime {
			device.BandWidthInfo.OutBandWidthUsed -= viewer.DownThroughput
			// todo: 如果这是edge上看的最后一个人，理论上计算资源也需要被释放
			delete(t.solved, viewerWithWatchInfo)
			count++
		}
	}
	log.Printf("solved count: %d\n", count)
	log.Printf("current watch count: %d\n", len(t.solved))
	// time.Sleep(2 * time.Second)
	//for top := t.viewerList.Top(); top.EndTime <= t.time; top = t.viewerList.Top() {
	//	watchInfo := t.viewerList.Pop().(*ViewerWithWatchChannel)
	//	system := (*t.ctx).Value("system").(*System)
	//	t.solved[watchInfo.Viewer] = nil
	//	system.RemoveViewer(watchInfo.Viewer)
	//	if len(t.viewerList) == 0 {
	//		return
	//	}
	//}
}

func (t *TaskManager) TimeGrowth() {
	t.time++
	log.Printf("Time growth: %d", t.time)
	time.Sleep(1 * time.Second)
	t.RefreshTasks()
	for _, task := range t.task[t.time] {
		t.viewerList.Push(task.watchInfo)
	}
}

func (t *TaskManager) GetTask() *Task {
	for ; len(t.viewerList) == 0; t.TimeGrowth() {
		if t.time > t.maxTime {
			return nil
		}
	}
	return &Task{
		watchInfo: t.viewerList.Pop().(*ViewerWithWatchChannel),
	}
}

func (t *TaskManager) TakeAction(ctx *context.Context, req *rpc.TrainStepRequest) float64 {
	// 从req.Base.Extra拿了什么: version, deviceId
	action := req.Action
	system := (*t.ctx).Value("system").(*System)
	viewerId := action.GetViewerId()
	viewerInfo := (*t.ctx).Value("viewer").(*ViewerInfo)
	viewer := viewerInfo.viewer[viewerId]
	var liveInfo *LiveInfo
	for _, live := range viewer.LiveInfo {
		if live.ChannelId == req.Action.ChannelId && t.time == live.StartTime {
			liveInfo = live
		}
	}
	if liveInfo == nil {
		log.Fatalf("不能找到相应的liveInfo\n")
	}
	version, _ := strconv.ParseInt(req.Base.Extra["version"], 10, 64)
	viewer.AssignInfo = AssignInfo{
		DeviceId:  req.Base.Extra["deviceId"],
		Version:   version,
		ChannelId: req.Action.ChannelId,
	}
	if edge, ok := system.Edge["Edge"+strconv.FormatInt(action.GetAction(), 10)]; ok {
		if outBandUsage := edge.BandWidthInfo.OutBandWidthUsed + viewer.DownThroughput; outBandUsage < edge.BandWidthInfo.OutBandWidthLimit {
			edge.BandWidthInfo.OutBandWidthUsed = outBandUsage
			viewer.AssignInfo.DeviceId = edge.Name
			t.solved[&ViewerWithWatchChannel{
				Viewer:   viewer,
				LiveInfo: liveInfo,
			}] = &edge.DeviceCommon
		} else {
			// DONE: 无论如何都不够了，需要redirect 直接转去Cdn
			// DONE: 不能直接去cdn，还是得看edge的情况，有的edge可能都没被用到过
			// DONE: 看edge的情况，edge中选择占用率小于80%中最小的占用率的
			log.Printf("Edge redirect！\n")
			var device *DeviceCommon
			minDeviceUsed := 0.8
			for _, _edge := range system.Edge {
				if deviceUsedPercent := _edge.BandWidthInfo.OutBandWidthUsed / _edge.BandWidthInfo.OutBandWidthLimit; deviceUsedPercent < minDeviceUsed {
					minDeviceUsed = deviceUsedPercent
					device = &_edge.DeviceCommon
				}
			}
			if device == nil {
				for name, cdn := range system.Cdn {
					viewer.AssignInfo.DeviceId = name
					t.solved[&ViewerWithWatchChannel{
						Viewer:   viewer,
						LiveInfo: liveInfo,
					}] = &cdn.DeviceCommon
				}
			} else {
				device.BandWidthInfo.OutBandWidthUsed += viewer.DownThroughput
				viewer.AssignInfo.DeviceId = device.Name
				t.solved[&ViewerWithWatchChannel{
					Viewer:   viewer,
					LiveInfo: liveInfo,
				}] = device
			}
		}
		//else {
		// 尽可能给；ps: 尽可能给没意义，要给也只能给一个人，而且对于OutBandWidthUsed，加的究竟是BitRateMap还是DownThroughput也不统一
		//	var realRate int64
		//	for index, rate := range BitRateMap {
		//		if edge.BandWidthInfo.OutBandWidthUsed+rate < edge.BandWidthInfo.OutBandWidthLimit {
		//			if realRate < index {
		//				realRate = index
		//			}
		//		}
		//	}
		//	edge.BandWidthInfo.OutBandWidthUsed += BitRateMap[realRate]
		//	viewer.AssignInfo.DeviceId = edge.Name
		//	viewer.AssignInfo.Version = realRate
		//	t.solved[&ViewerWithWatchChannel{
		//		Viewer:   viewer,
		//		LiveInfo: liveInfo,
		//	}] = &edge.DeviceCommon
		//}
	} else if cdn, ok := system.Cdn["Cdn"+strconv.FormatInt(action.GetAction(), 10)]; ok {
		viewer.AssignInfo.DeviceId = cdn.Name
		t.solved[&ViewerWithWatchChannel{
			Viewer:   viewer,
			LiveInfo: liveInfo,
		}] = &cdn.DeviceCommon
	} else {
		log.Fatalf("不存在的action: %v\n", action)
	}
	// t.viewerList.Pop()

	reward := GetReward(ctx, viewer, req)
	log.Printf("Channel Id: %s\tDevice name: %s\tVersion: %d\tReward: %f\n", viewer.AssignInfo.ChannelId, viewer.AssignInfo.DeviceId, viewer.AssignInfo.Version, reward)
	return reward
}

func (t *TaskManager) NextState(ctx *context.Context) *rpc.State {
	task := t.GetTask()
	if task == nil {
		return nil
	}
	system := (*t.ctx).Value("system").(*System)
	inboundUsed := make([]float64, 0)
	outboundUsed := make([]float64, 0)
	computeUsed := make([]float64, 0)
	for _, inbound := range system.InboundMap {
		inboundUsed = append(inboundUsed, *inbound)
	}
	for _, outbound := range system.OutboundMap {
		outboundUsed = append(outboundUsed, *outbound)
	}
	for _, compute := range system.ComputationMap {
		computeUsed = append(computeUsed, *compute)
	}
	conn := GetConnMap(t.solved) // edge->channel->version->number
	var viewerConnectionMap rpc.ViewerConnection
	viewerConnectionMap.ViewerConnectionTable = make(map[string]*rpc.H2V, 0)
	for deviceId, c2v := range *conn {
		_v2number := make(map[string]*rpc.V2Number, 0)
		for channelId, v2number := range *c2v {
			_number := make(map[string]int64, 0)
			for version, number := range *v2number {
				_number[strconv.FormatInt(version, 10)] = *number
			}
			_v2number[channelId] = &rpc.V2Number{
				Number: _number,
			}
		}
		viewerConnectionMap.ViewerConnectionTable[deviceId] = &rpc.H2V{
			H2V: _v2number,
		}
	}
	qoePreference := GetQoePreference(task)
	state := rpc.State{
		InboundBandwidthUsage: &rpc.InboundBandwidthUsage{
			InboundBandwidthUsage: inboundUsed,
		},
		OutboundBandwidthUsage: &rpc.OutboundBandwidthUsage{
			OutboundBandwidthUsage: outboundUsed,
		},
		ComputationResourceUsage: &rpc.ComputationResourceUsage{
			ComputationResourceUsage: computeUsed,
		},
		QoePreference: &rpc.QoEPreference{
			Alpha1: qoePreference[0],
			Alpha2: qoePreference[1],
			Alpha3: qoePreference[2],
		},
		UserInfo: &rpc.UserInfo{
			Location: &rpc.Location{
				Latitude:  task.watchInfo.Location.Lat,
				Longitude: task.watchInfo.Location.Long,
			},
			ChannelId: task.watchInfo.ChannelId,
			Version:   GetVersion(task.watchInfo.VersionBit),
			UserId:    task.watchInfo.Viewer.Id,
		},
		ViewerConnection: &viewerConnectionMap,
	}
	return &state
}

func GetVersion(version int64) int64 {
	switch version {
	case 768000:
		return 480
	case 1250000:
		return 720
	case 2250000:
		return 1080
	case 3750000:
		return 1440
	default:
		return 1440
	}
}

func GetReward(ctx *context.Context, viewer *Viewer, req *rpc.TrainStepRequest) float64 {
	action := req.Action
	var rewarda, rewardb float64
	rewarda += float64(action.GetQoePreference().Alpha1) * GetStreamingDelay(ctx, viewer, viewer.AssignInfo.DeviceId)
	rewarda += float64(action.GetQoePreference().Alpha2) * GetChannelSwitchingDelay(ctx, viewer)
	rewarda += float64(action.GetQoePreference().Alpha3) * GetMismatchLevel(ctx, viewer, action, req.Base.Extra["version"])
	rewarda *= Alpha

	rewardb += GetComputationCost(ctx, viewer)
	rewardb += GetBandwidthCost(ctx, viewer)
	rewardb *= Beta
	return -1 * (rewarda + rewardb)
}

func GetStreamingDelay(ctx *context.Context, viewer *Viewer, deviceId string) float64 {
	var streamingDelay float64
	system := (*ctx).Value("system").(*System)
	if edge, ok := system.Edge[deviceId]; ok {
		streamingDelay += edge.ViewerLatencyCal(viewer)
		streamingDelay += edge.TranscodingLatencyCal(viewer)
		streamingDelay += edge.LatencyToUpper
	} else if cdn, ok := system.Cdn[deviceId]; ok {
		streamingDelay += cdn.ViewerLatencyCal(viewer)
	} else {
		log.Fatalf("不存在的edge: %v\n", deviceId)
	}
	return streamingDelay
}

func GetChannelSwitchingDelay(ctx *context.Context, viewer *Viewer) float64 {
	return viewer.Latency
}

func GetMismatchLevel(ctx *context.Context, viewer *Viewer, action *rpc.Action, reqVersion string) float64 {
	var mismatchLevel float64
	version, _ := strconv.ParseInt(reqVersion, 10, 64)
	mismatchLevel += math.Log(float64(version)) - math.Log(float64(viewer.AssignInfo.Version))
	return mismatchLevel
}

func GetComputationCost(ctx *context.Context, viewer *Viewer) float64 {
	if strings.Contains(viewer.AssignInfo.DeviceId, "Edge") {
		return TransCodingCpuMap[viewer.AssignInfo.Version] * EdgeComputationPrice
	} else {
		return 0
	}
}

func GetBandwidthCost(ctx *context.Context, viewer *Viewer) float64 {
	var bandwidthCost float64
	if strings.Contains(viewer.AssignInfo.DeviceId, "Edge") {
		bandwidthCost += BitRateMap[viewer.AssignInfo.Version] * EdgeBandwidthPrice
		bandwidthCost += BitRateMap[1440] * CdnToEdgeBandwidthPrice // cdn->edge
	} else {
		bandwidthCost += BitRateMap[viewer.AssignInfo.Version] * CdnBandwidthPrice
	}
	return bandwidthCost
}

func GetConnMap(solved map[*ViewerWithWatchChannel]*DeviceCommon) *map[string]*map[string]*map[int64]*int64 {
	conn := make(map[string]*map[string]*map[int64]*int64, 0)
	for viewerWithWatchInfo, device := range solved {
		viewer := viewerWithWatchInfo.Viewer
		var c2v *map[string]*map[int64]*int64
		var v2number *map[int64]*int64
		var ok bool
		if c2v, ok = conn[strconv.Itoa(int(device.Id))]; !ok {
			v := make(map[string]*map[int64]*int64, 0)
			conn[strconv.Itoa(int(device.Id))] = &v
			c2v = &v
		}
		if v2number, ok = (*c2v)[viewer.AssignInfo.ChannelId]; !ok {
			number := make(map[int64]*int64, 0)
			(*c2v)[viewer.AssignInfo.ChannelId] = &number
			v2number = &number
		}
		if number, ok := (*v2number)[viewer.AssignInfo.Version]; ok {
			*number++
		} else {
			n := int64(1)
			(*v2number)[viewer.AssignInfo.Version] = &n
		}
	}
	return &conn
}

func GetQoePreference(task *Task) []float32 {
	viewer := task.watchInfo.Viewer
	liveInfo := task.watchInfo.LiveInfo
	n := len(viewer.LiveInfo)
	t := (liveInfo.EndTime - liveInfo.StartTime) * 10
	if n <= 2 && t >= 30 {
		return []float32{2, 1.5, 2}
	} else if n >= 5 && t <= 10 {
		return []float32{0.5, 6, 2}
	} else if n > 4 && t >= 30 {
		return []float32{0.5, 1.5, 8}
	} else {
		return []float32{1, 3, 4}
	}
}
