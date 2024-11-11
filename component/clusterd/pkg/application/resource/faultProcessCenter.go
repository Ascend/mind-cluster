package resource

import (
	"clusterd/pkg/application/resource/fault"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
	"context"
	"huawei.com/npu-exporter/v6/common-utils/hwlog"
	"time"
)

var GlobalFaultProcessCenter *FaultProcessCenter

// The FaultProcessCenter process the faults
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
	hwlog.RunLog.Info("notify fault center process switch fault for add")
	hwlog.RunLog.Debugf("old switch info: %s, new switch info %s",
		util.ObjToString(oldInfo), util.ObjToString(newInfo))
	GlobalFaultProcessCenter.notifyFaultCenterProcess(constant.SWITCH_FAULT)
}

func (center *FaultProcessCenter) informSwitchInfoDel(newInfo *constant.SwitchInfo) {
	center.switchCenter.InformerDelCallback(newInfo)
	hwlog.RunLog.Info("notify fault center process switch fault for delete")
	hwlog.RunLog.Debugf("delete switch info: %s", util.ObjToString(newInfo))
	GlobalFaultProcessCenter.notifyFaultCenterProcess(constant.SWITCH_FAULT)
}

func (center *FaultProcessCenter) informDeviceInfoAdd(oldInfo, newInfo *constant.DeviceInfo) {
	center.deviceCenter.InformerAddCallback(oldInfo, newInfo)
	hwlog.RunLog.Info("notify fault center process device fault for add")
	hwlog.RunLog.Debugf("old device info: %s, new device info %s",
		util.ObjToString(oldInfo), util.ObjToString(newInfo))
	GlobalFaultProcessCenter.notifyFaultCenterProcess(constant.DEVICE_FAULT)
}

func (center *FaultProcessCenter) informDeviceInfoDel(newInfo *constant.DeviceInfo) {
	center.deviceCenter.InformerDelCallback(newInfo)
	hwlog.RunLog.Info("notify fault center process device fault for delete")
	hwlog.RunLog.Debugf("delete device info: %s", util.ObjToString(newInfo))
	GlobalFaultProcessCenter.notifyFaultCenterProcess(constant.DEVICE_FAULT)
}

func (center *FaultProcessCenter) informNodeInfoAdd(oldInfo, newInfo *constant.NodeInfo) {
	center.nodeCenter.InformerAddCallback(oldInfo, newInfo)
	hwlog.RunLog.Info("notify fault center process node fault for add")
	hwlog.RunLog.Debugf("old node info: %s, new node info %s",
		util.ObjToString(oldInfo), util.ObjToString(newInfo))
	GlobalFaultProcessCenter.notifyFaultCenterProcess(constant.NODE_FAULT)
}

func (center *FaultProcessCenter) informNodeInfoDel(newInfo *constant.NodeInfo) {
	center.nodeCenter.InformerDelCallback(newInfo)
	hwlog.RunLog.Info("notify fault center process node fault for delete")
	hwlog.RunLog.Debugf("delete node info: %s", util.ObjToString(newInfo))
	GlobalFaultProcessCenter.notifyFaultCenterProcess(constant.NODE_FAULT)
}

func (center *FaultProcessCenter) notifyFaultCenterProcess(whichToProcess int) {
	center.notifyProcessChan <- whichToProcess
}

func (center *FaultProcessCenter) work(ctx context.Context) {
	hwlog.RunLog.Info("FaultProcessCenter start work.")
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

// CallbackForReportUceInfo cluster grpc should call back for report uce fault situation
type MindIoReportRecoverInfo struct {
	jobId       string
	rankId      string
	recoverTime int64
}

func (center *FaultProcessCenter) CallbackForReportUceInfo(infos []MindIoReportRecoverInfo) error {
	for _, info := range infos {
		center.deviceCenter.CallbackForReportUceInfo(info.jobId, info.rankId, info.recoverTime)
	}
	center.notifyFaultCenterProcess(constant.DEVICE_FAULT)
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
	hwlog.RunLog.Errorf("Wrong number %d, cannot decide which to register.", which)
}

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
