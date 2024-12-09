// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package faultmanager contain fault process
package faultmanager

import (
	"context"
	"time"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
)

func (center *FaultProcessCenter) process() {
	center.deviceCenter.process()
	center.nodeCenter.process()
	center.switchCenter.process()
}

// NewFaultProcessCenter create deviceCenter,nodeCenter,switchCenter and work goroutine
func NewFaultProcessCenter(ctx context.Context) {
	GlobalFaultProcessCenter = &FaultProcessCenter{
		deviceCenter:      newDeviceFaultProcessCenter(),
		nodeCenter:        newNodeFaultProcessCenter(),
		switchCenter:      newSwitchFaultProcessCenter(),
		notifyProcessChan: make(chan int, 1000),
	}
	go GlobalFaultProcessCenter.work(ctx)
}

func (center *FaultProcessCenter) notifyFaultCenterProcess(whichToProcess int) {
	center.notifyProcessChan <- whichToProcess
}

func (center *FaultProcessCenter) work(ctx context.Context) {
	hwlog.RunLog.Info("FaultProcessCenter start work")
	centerTicker := time.NewTicker(time.Second)
	for {
		select {
		case <-ctx.Done():
			hwlog.RunLog.Info("FaultProcessCenter stop work")
			return
		case whichToProcess := <-center.notifyProcessChan:
			switch whichToProcess {
			case constant.AllProcessType:
				center.process()
			case constant.DeviceProcessType:
				center.deviceCenter.process()
			case constant.NodeProcessType:
				center.nodeCenter.process()
			case constant.SwitchProcessType:
				center.switchCenter.process()
			default:
				hwlog.RunLog.Errorf("wrong number %d to process", whichToProcess)
			}
		case <-centerTicker.C:
			center.process()
		}
	}
}

func (center *FaultProcessCenter) getJobFaultRankProcessor() (*jobRankFaultInfoProcessor, error) {
	return center.deviceCenter.getJobFaultRankProcessor()
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
		center.deviceCenter.callbackForReportUceInfo(info.JobId, info.Rank, info.RecoverTime)
	}
	center.notifyFaultCenterProcess(constant.DeviceProcessType)
	return nil
}

// Register to notify fault occurrence
func (center *FaultProcessCenter) Register(ch chan int, whichToRegister int) {
	switch whichToRegister {
	case constant.SwitchProcessType:
		center.switchCenter.register(ch)
	case constant.NodeProcessType:
		center.nodeCenter.register(ch)
	case constant.DeviceProcessType:
		center.deviceCenter.register(ch)
	case constant.AllProcessType:
		center.switchCenter.register(ch)
		center.nodeCenter.register(ch)
		center.deviceCenter.register(ch)
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
	return center.deviceCenter.getProcessedCm()
}

// QuerySwitchInfoToReport query switch info to report
func (center *FaultProcessCenter) QuerySwitchInfoToReport() map[string]*constant.SwitchInfo {
	return center.switchCenter.getProcessedCm()
}

// QueryNodeInfoToReport query node info to report
func (center *FaultProcessCenter) QueryNodeInfoToReport() map[string]*constant.NodeInfo {
	return center.nodeCenter.getProcessedCm()
}
