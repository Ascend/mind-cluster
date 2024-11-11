package fault

import (
	"clusterd/pkg/common/constant"
	"huawei.com/npu-exporter/v6/common-utils/hwlog"
	"time"
)

// The FaultProcessor process the fault information.
type FaultProcessor interface {
	Process()
}

type BaseFaultCenter struct {
	processors        []FaultProcessor
	lastProcessTime   int64
	subscribeChannels []chan struct{}
	processPeriod     int64
}

func (baseCenter *BaseFaultCenter) isProcessLimited(currentTime int64) bool {
	return baseCenter.lastProcessTime+baseCenter.processPeriod < currentTime
}

func (baseCenter *BaseFaultCenter) Process() {
	currentTime := time.Now().UnixMilli()
	if baseCenter.isProcessLimited(currentTime) {
		return
	}
	baseCenter.lastProcessTime = currentTime
	for _, processor := range baseCenter.processors {
		processor.Process()
	}
	for _, ch := range baseCenter.subscribeChannels {
		ch <- struct{}{}
	}
}

func newBaseFaultCenter() BaseFaultCenter {
	return BaseFaultCenter{
		processors:        make([]FaultProcessor, 0),
		lastProcessTime:   0,
		subscribeChannels: make([]chan struct{}, 0),
		processPeriod:     0,
	}
}

func (baseCenter *BaseFaultCenter) addProcessors(processors []FaultProcessor) {
	baseCenter.processors = append(baseCenter.processors, processors...)
}

func (baseCenter *BaseFaultCenter) RegisterSubscriber(ch chan struct{}) {
	if baseCenter.subscribeChannels == nil {
		baseCenter.subscribeChannels = make([]chan struct{}, 0)
	}
	length := len(baseCenter.subscribeChannels)
	if length > constant.MAX_FAULT_CENTER_SUBSCRIBER {
		hwlog.RunLog.Errorf("The number of registrants is %d, cannot add any more.", length)
	}
	baseCenter.subscribeChannels = append(baseCenter.subscribeChannels, ch)
	hwlog.RunLog.Infof("The number of registrants is %d.", length+1)
}
