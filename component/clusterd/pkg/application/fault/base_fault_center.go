package fault

import (
	"clusterd/pkg/application/job"
	"clusterd/pkg/common/util"
	"clusterd/pkg/interface/kube"
	"sync"
	"time"

	"huawei.com/npu-exporter/v6/common-utils/hwlog"

	"clusterd/pkg/common/constant"
)

// The faultProcessor process the fault information.
type faultProcessor interface {
	process()
}

type baseFaultCenter struct {
	processorList        []faultProcessor
	lastProcessTime      int64
	subscribeChannelList []chan struct{}
	mutex                sync.Mutex
	processPeriod        int64
	jobServerInfoMap     job.JobServerInfoMap
}

func newBaseFaultCenter() baseFaultCenter {
	return baseFaultCenter{
		processorList:        make([]faultProcessor, 0),
		lastProcessTime:      0,
		subscribeChannelList: make([]chan struct{}, 0),
		mutex:                sync.Mutex{},
		processPeriod:        constant.FaultCenterProcessPeriod,
	}
}

func (baseCenter *baseFaultCenter) isProcessLimited(currentTime int64) bool {
	return baseCenter.lastProcessTime+baseCenter.processPeriod > currentTime
}

func (baseCenter *baseFaultCenter) process() {
	currentTime := time.Now().UnixMilli()
	if baseCenter.isProcessLimited(currentTime) {
		hwlog.RunLog.Infof("process limited, last time %d, current time %d",
			baseCenter.lastProcessTime, currentTime)
		return
	}
	baseCenter.lastProcessTime = currentTime
	baseCenter.jobServerInfoMap = kube.JobMgr.GetJobServerInfoMap()
	hwlog.RunLog.Infof("job server info map: %v", util.ObjToString(baseCenter.jobServerInfoMap))
	for _, processor := range baseCenter.processorList {
		processor.process()
	}
	for _, ch := range baseCenter.subscribeChannelList {
		ch <- struct{}{}
	}
}

func (baseCenter *baseFaultCenter) addProcessors(processors []faultProcessor) {
	baseCenter.processorList = append(baseCenter.processorList, processors...)
}

func (baseCenter *baseFaultCenter) register(ch chan struct{}) {
	baseCenter.mutex.Lock()
	defer baseCenter.mutex.Unlock()
	if baseCenter.subscribeChannelList == nil {
		baseCenter.subscribeChannelList = make([]chan struct{}, 0)
	}
	length := len(baseCenter.subscribeChannelList)
	if length > constant.MAX_FAULT_CENTER_SUBSCRIBER {
		hwlog.RunLog.Errorf("The number of registrants is %d, cannot add any more", length)
	}
	baseCenter.subscribeChannelList = append(baseCenter.subscribeChannelList, ch)
	hwlog.RunLog.Infof("The number of registrants is %d", length+1)
}
