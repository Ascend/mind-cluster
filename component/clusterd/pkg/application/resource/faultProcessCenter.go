package resource

import (
	"clusterd/pkg/application/resource/fault"
	"clusterd/pkg/common/constant"
	"context"
	"huawei.com/npu-exporter/v6/common-utils/hwlog"
	"time"
)

var GlobalFaultProcessCenter *FaultProcessCenter

// The FaultProcessCenter maintain the fault information.
type FaultProcessCenter struct {
	deviceCenter      *fault.DeviceFaultProcessCenter
	nodeCenter        *fault.NodeFaultProcessCenter
	switchCenter      *fault.SwitchFaultProcessCenter
	notifyProcessChan chan int
}

func (center *FaultProcessCenter) Process() {
	center.deviceCenter.Process()
	center.nodeCenter.Process()
	center.switchCenter.Process()
}

func NewFaultProcessCenter(ctx context.Context) {
	GlobalFaultProcessCenter = &FaultProcessCenter{
		deviceCenter:      fault.NewDeviceFaultProcessCenter(),
		nodeCenter:        fault.NewNodeFaultProcessCenter(),
		switchCenter:      fault.NewSwitchFaultProcessCenter(),
		notifyProcessChan: make(chan int),
	}
	go GlobalFaultProcessCenter.work(ctx)
}

func (center *FaultProcessCenter) informSwitchInfoAdd(oldInfo, newInfo *constant.SwitchInfo) {
	center.switchCenter.InformerAddCallback(oldInfo, newInfo)
}

func (center *FaultProcessCenter) informSwitchInfoDel(oldInfo, newInfo *constant.SwitchInfo) {
	center.switchCenter.InformerDelCallback(oldInfo, newInfo)
}

func (center *FaultProcessCenter) informDeviceInfoAdd(oldInfo, newInfo *constant.DeviceInfo) {
	changed := center.deviceCenter.InformerAddCallback(oldInfo, newInfo)
	if changed {
		GlobalFaultProcessCenter.NotifyFaultCenterProcess(constant.DEVICE_FAULT)
	}
}

func (center *FaultProcessCenter) informDeviceInfoDel(oldInfo, newInfo *constant.DeviceInfo) {
	changed := center.deviceCenter.InformerDelCallback(oldInfo, newInfo)
	if changed {
		GlobalFaultProcessCenter.NotifyFaultCenterProcess(constant.DEVICE_FAULT)
	}
}

func (center *FaultProcessCenter) informNodeInfoAdd(oldInfo, newInfo *constant.NodeInfo) {
	changed := center.nodeCenter.InformerAddCallback(oldInfo, newInfo)
	if changed {
		GlobalFaultProcessCenter.NotifyFaultCenterProcess(constant.NODE_FAULT)
	}
}

func (center *FaultProcessCenter) informNodeInfoDel(oldInfo, newInfo *constant.NodeInfo) {
	changed := center.nodeCenter.InformerDelCallback(oldInfo, newInfo)
	if changed {
		GlobalFaultProcessCenter.NotifyFaultCenterProcess(constant.NODE_FAULT)
	}
}

func (center *FaultProcessCenter) NotifyFaultCenterProcess(whichToProcess int) {
	center.notifyProcessChan <- whichToProcess
}

func (center *FaultProcessCenter) work(ctx context.Context) {
	centerTicker := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-ctx.Done():
			hwlog.RunLog.Info("FaultProcessCenter stop work.")
			return
		case which := <-center.notifyProcessChan:
			switch which {
			case constant.ALL_FAULT:
				center.Process()
			case constant.DEVICE_FAULT:
				center.deviceCenter.Process()
			case constant.NODE_FAULT:
				center.nodeCenter.Process()
			case constant.SWITCH_FAULT:
				center.switchCenter.Process()
			}
		case <-centerTicker.C:
			center.Process()
		}
	}
}

func (center *FaultProcessCenter) getUceFaultProcessor() (*fault.UceFaultProcessor, error) {
	return center.deviceCenter.GetUceFaultProcessor()
}

func (center *FaultProcessCenter) getUceAccompanyFaultProcessor() (*fault.UceAccompanyFaultProcessor, error) {
	return center.deviceCenter.GetUceAccompanyFaultProcessor()
}

func (center *FaultProcessCenter) getJobFaultRankProcessor() (*fault.JobRankFaultInfoProcessor, error) {
	return center.deviceCenter.GetJobFaultRankProcessor()
}

func (center *FaultProcessCenter) GetDeviceFaultCenter() *fault.DeviceFaultProcessCenter {
	return center.deviceCenter
}

// CallbackForReportUceInfo cluster grpc should call back for report uce fault situation
func (center *FaultProcessCenter) CallbackForReportUceInfo(jobId, rankId string, recoverTime int64) error {
	err := center.deviceCenter.CallbackForReportUceInfo(jobId, rankId, reportTime)
	if err != nil {
		return err
	}
	center.NotifyFaultCenterProcess(constant.DEVICE_FAULT)
	return nil
}

// RegisterSubscriber to notify fault occurrence
func (center *FaultProcessCenter) RegisterSubscriber(ch chan struct{}, which int) {
	switch which {
	case constant.SWITCH_FAULT:
		center.switchCenter.RegisterSubscriber(ch)
	case constant.NODE_FAULT:
		center.nodeCenter.RegisterSubscriber(ch)
	case constant.DEVICE_FAULT:
		center.deviceCenter.RegisterSubscriber(ch)
	case constant.ALL_FAULT:
		center.switchCenter.RegisterSubscriber(ch)
		center.nodeCenter.RegisterSubscriber(ch)
		center.deviceCenter.RegisterSubscriber(ch)
	}
	hwlog.RunLog.Errorf("Wrong number %d, cannot decide which to regiter.", which)
}

// TODO 他们需要什么？该函数有点定制化
func (center *FaultProcessCenter) QueryJobsFaultInfo() map[string]fault.FaultInfo {
	processor, err := center.getJobFaultRankProcessor()
	if err != nil {
		hwlog.RunLog.Error(err)
		return nil
	}
	return processor.GetJobFaultRankInfos()
}

func (center *FaultProcessCenter) QueryDeviceInfoToReport() map[string]*constant.DeviceInfo {
	return center.deviceCenter.GetDeviceInfos()
}

func (center *FaultProcessCenter) QuerySwitchInfoToReport() map[string]*constant.SwitchInfo {
	return center.switchCenter.GetSwitchInfos()
}

func (center *FaultProcessCenter) QueryNodeInfoToReport() map[string]*constant.NodeInfo {
	return center.nodeCenter.GetNodeInfos()
}

// util functions
