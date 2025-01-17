// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package faultmanager contain fault process
package faultmanager

import (
	"context"
	"time"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/application/faultmanager/collector"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/job"
)

func (center *FaultProcessCenter) Process() {
	center.jobServerInfoMap = job.GetJobServerInfoMap()
	center.DeviceCenter.Process()
	center.NodeCenter.Process()
	center.SwitchCenter.Process()
	center.faultJobProcessor.Process()
	center.FaultJobCenter.Process()
}

// NewFaultProcessCenter create deviceCenter,nodeCenter,switchCenter and work goroutine
func NewFaultProcessCenter() *FaultProcessCenter {
	GlobalFaultProcessCenter = &FaultProcessCenter{
		DeviceCenter:      NewDeviceFaultProcessCenter(),
		NodeCenter:        NewNodeFaultProcessCenter(),
		SwitchCenter:      NewSwitchFaultProcessCenter(),
		FaultJobCenter:    NewFaultJobProcessCenter(),
		NotifyProcessChan: make(chan int, 1000),
	}

	processor, err := GlobalFaultProcessCenter.DeviceCenter.getJobFaultRankProcessor()
	if err != nil {
		hwlog.RunLog.Errorf("get device fault rank processor failed, err: %v", err)
		return nil
	}

	GlobalFaultProcessCenter.faultJobProcessor = &faultProcessorImpl{
		jobRankFaultInfoProcessor: processor,
	}
	return GlobalFaultProcessCenter
}

func (center *FaultProcessCenter) NotifyFaultCenterProcess(whichToProcess int) {
	center.NotifyProcessChan <- whichToProcess
}

func (center *FaultProcessCenter) Work(ctx context.Context) {
	go func() {
		hwlog.RunLog.Info("FaultProcessCenter start work")
		centerTicker := time.NewTicker(time.Second)
		for {
			select {
			case <-ctx.Done():
				hwlog.RunLog.Info("FaultProcessCenter stop work")
				return
			case whichToProcess := <-center.NotifyProcessChan:
				switch whichToProcess {
				case constant.AllProcessType:
					center.Process()
				case constant.DeviceProcessType:
					center.DeviceCenter.Process()
				case constant.NodeProcessType:
					center.NodeCenter.Process()
				case constant.SwitchProcessType:
					center.SwitchCenter.Process()
				default:
					hwlog.RunLog.Errorf("wrong number %d to process", whichToProcess)
				}
			case <-centerTicker.C:
				center.Process()
			}
		}
	}()
}

func (center *FaultProcessCenter) getJobFaultRankProcessor() (*jobRankFaultInfoProcessor, error) {
	return center.DeviceCenter.getJobFaultRankProcessor()
}

// ReportRecoverInfo cluster grpc should call back for report uce fault
type ReportRecoverInfo struct {
	JobId       string
	Rank        string
	RecoverTime int64
}

// CallbackForReportUceInfo callback function to report uce info
func (center *FaultProcessCenter) CallbackForReportUceInfo(infos []ReportRecoverInfo) error {
	for _, info := range infos {
		collector.ReportInfoCollector.ReportUceInfo(info.JobId, info.Rank, info.RecoverTime)
	}
	center.NotifyFaultCenterProcess(constant.DeviceProcessType)
	return nil
}

// Register to notify fault occurrence
func (center *FaultProcessCenter) Register(ch chan int, whichToRegister int) {
	switch whichToRegister {
	case constant.SwitchProcessType:
		center.SwitchCenter.register(ch)
	case constant.NodeProcessType:
		center.NodeCenter.register(ch)
	case constant.DeviceProcessType:
		center.DeviceCenter.register(ch)
	case constant.AllProcessType:
		center.SwitchCenter.register(ch)
		center.NodeCenter.register(ch)
		center.DeviceCenter.register(ch)
	default:
		hwlog.RunLog.Errorf("Wrong number %d, cannot decide which to register", whichToRegister)
	}
}

// QueryJobsFaultInfo query jobs fault rank info, and filter fault below `faultLevel`
func (center *FaultProcessCenter) QueryJobsFaultInfo(faultLevel string) map[string]JobFaultInfo {
	processor, err := center.getJobFaultRankProcessor()
	if err != nil {
		hwlog.RunLog.Error(err)
		return nil
	}
	return processor.getJobFaultRankInfosFilterLevel(faultLevel)
}

// QueryDeviceInfoToReport query device info to report
func (center *FaultProcessCenter) QueryDeviceInfoToReport() map[string]*constant.DeviceInfo {
	return center.DeviceCenter.getProcessedCm()
}

// QuerySwitchInfoToReport query switch info to report
func (center *FaultProcessCenter) QuerySwitchInfoToReport() map[string]*constant.SwitchInfo {
	return center.SwitchCenter.getProcessedCm()
}

// QueryNodeInfoToReport query node info to report
func (center *FaultProcessCenter) QueryNodeInfoToReport() map[string]*constant.NodeInfo {
	return center.NodeCenter.getProcessedCm()
}
