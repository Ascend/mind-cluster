package fault

import (
	"context"
	"time"

	"huawei.com/npu-exporter/v6/common-utils/hwlog"

	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
)

var GlobalFaultProcessCenter *FaultProcessCenter

// The FaultProcessCenter process the faults
type FaultProcessCenter struct {
	deviceCenter      *deviceFaultProcessCenter
	nodeCenter        *nodeFaultProcessCenter
	switchCenter      *switchFaultProcessCenter
	notifyProcessChan chan int
}

func (center *FaultProcessCenter) process() {
	center.deviceCenter.process()
	center.nodeCenter.process()
	center.switchCenter.process()
}

func NewFaultProcessCenter(ctx context.Context) {
	GlobalFaultProcessCenter = &FaultProcessCenter{
		deviceCenter:      newDeviceFaultProcessCenter(),
		nodeCenter:        newNodeFaultProcessCenter(),
		switchCenter:      newSwitchFaultProcessCenter(),
		notifyProcessChan: make(chan int, 1000),
	}
	go GlobalFaultProcessCenter.work(ctx)
}

func (center *FaultProcessCenter) informSwitchInfoAdd(newInfo *constant.SwitchInfo) {
	center.switchCenter.updateInfoFromCm(newInfo)
	hwlog.RunLog.Infof("notify fault center process switch fault for add")
	hwlog.RunLog.Infof("new switch info %s", util.ObjToString(newInfo))
	GlobalFaultProcessCenter.notifyFaultCenterProcess(constant.SWITCH_FAULT)
}

func (center *FaultProcessCenter) informSwitchInfoDel(newInfo *constant.SwitchInfo) {
	center.switchCenter.delInfoFromCm(newInfo)
	hwlog.RunLog.Infof("notify fault center process switch fault for delete")
	hwlog.RunLog.Infof("delete switch info: %s", util.ObjToString(newInfo))
	GlobalFaultProcessCenter.notifyFaultCenterProcess(constant.SWITCH_FAULT)
}

func (center *FaultProcessCenter) informDeviceInfoAdd(newInfo *constant.DeviceInfo) {
	center.deviceCenter.updateInfoFromCm(newInfo)
	hwlog.RunLog.Infof("notify fault center process device fault for add")
	hwlog.RunLog.Infof("new device info %s", util.ObjToString(newInfo))
	GlobalFaultProcessCenter.notifyFaultCenterProcess(constant.DEVICE_FAULT)
}

func (center *FaultProcessCenter) informDeviceInfoDel(newInfo *constant.DeviceInfo) {
	center.deviceCenter.delInfoFromCm(newInfo)
	hwlog.RunLog.Infof("notify fault center process device fault for delete")
	hwlog.RunLog.Infof("delete device info: %s", util.ObjToString(newInfo))
	GlobalFaultProcessCenter.notifyFaultCenterProcess(constant.DEVICE_FAULT)
}

func (center *FaultProcessCenter) informNodeInfoAdd(newInfo *constant.NodeInfo) {
	center.nodeCenter.updateInfoFromCm(newInfo)
	hwlog.RunLog.Infof("notify fault center process node fault for add")
	hwlog.RunLog.Infof("new node info %s", util.ObjToString(newInfo))
	GlobalFaultProcessCenter.notifyFaultCenterProcess(constant.NODE_FAULT)
}

func (center *FaultProcessCenter) informNodeInfoDel(newInfo *constant.NodeInfo) {
	center.nodeCenter.delInfoFromCm(newInfo)
	hwlog.RunLog.Infof("notify fault center process node fault for delete")
	hwlog.RunLog.Infof("delete node info: %s", util.ObjToString(newInfo))
	GlobalFaultProcessCenter.notifyFaultCenterProcess(constant.NODE_FAULT)
}

func (center *FaultProcessCenter) notifyFaultCenterProcess(whichToProcess int) {
	center.notifyProcessChan <- whichToProcess
}

func (center *FaultProcessCenter) work(ctx context.Context) {
	hwlog.RunLog.Info("FaultProcessCenter start work")
	centerTicker := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-ctx.Done():
			hwlog.RunLog.Info("FaultProcessCenter stop work")
			return
		case whichToProcess := <-center.notifyProcessChan:
			switch whichToProcess {
			case constant.ALL_FAULT:
				center.process()
			case constant.DEVICE_FAULT:
				center.deviceCenter.process()
			case constant.NODE_FAULT:
				center.nodeCenter.process()
			case constant.SWITCH_FAULT:
				center.switchCenter.process()
			}
		case <-centerTicker.C:
			center.process()
		}
	}
}

func (center *FaultProcessCenter) getJobFaultRankProcessor() (*jobRankFaultInfoProcessor, error) {
	return center.deviceCenter.getJobFaultRankProcessor()
}

// callbackForReportUceInfo cluster grpc should call back for report uce fault situation
type ReportRecoverInfo struct {
	JobId       string
	Rank        string
	RecoverTime int64
}

// CallbackForReportUceInfo callback function to report uce info
func (center *FaultProcessCenter) CallbackForReportUceInfo(infos []ReportRecoverInfo) error {
	for _, info := range infos {
		center.deviceCenter.callbackForReportUceInfo(info.JobId, info.Rank, info.RecoverTime)
	}
	center.notifyFaultCenterProcess(constant.DEVICE_FAULT)
	return nil
}

// Register to notify fault occurrence
func (center *FaultProcessCenter) Register(ch chan struct{}, which int) {
	switch which {
	case constant.SWITCH_FAULT:
		center.switchCenter.register(ch)
	case constant.NODE_FAULT:
		center.nodeCenter.register(ch)
	case constant.DEVICE_FAULT:
		center.deviceCenter.register(ch)
	case constant.ALL_FAULT:
		center.switchCenter.register(ch)
		center.nodeCenter.register(ch)
		center.deviceCenter.register(ch)
	}
	hwlog.RunLog.Errorf("Wrong number %d, cannot decide which to register", which)
}

// QueryJobsFaultInfo query jobs fault rank info
func (center *FaultProcessCenter) QueryJobsFaultInfo() map[string]JobFaultInfo {
	processor, err := center.getJobFaultRankProcessor()
	if err != nil {
		hwlog.RunLog.Error(err)
		return nil
	}
	return processor.getJobFaultRankInfos()
}

// QueryDeviceInfoToReport query device info to report
func (center *FaultProcessCenter) QueryDeviceInfoToReport() map[string]*constant.DeviceInfo {
	return center.deviceCenter.getInfoMap()
}

// QuerySwitchInfoToReport query switch info to report
func (center *FaultProcessCenter) QuerySwitchInfoToReport() map[string]*constant.SwitchInfo {
	return center.switchCenter.getInfoMap()
}

// QueryNodeInfoToReport query node info to report
func (center *FaultProcessCenter) QueryNodeInfoToReport() map[string]*constant.NodeInfo {
	return center.nodeCenter.getInfoMap()
}
