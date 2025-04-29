// Copyright (c) Huawei Technologies Co., Ltd. 2024-2025. All rights reserved.

// Package faultmanager contain fault process
package faultmanager

import (
	"context"
	"time"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/application/faultmanager/cmprocess"
	"clusterd/pkg/application/faultmanager/jobprocess"
	"clusterd/pkg/application/faultmanager/jobprocess/faultrank"
	"clusterd/pkg/application/publicfault"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/faultdomain/collector"
)

// GlobalFaultProcessCenter is a global instance of faultProcessCenter used for processing faults.
var GlobalFaultProcessCenter *faultProcessCenter

func init() {
	GlobalFaultProcessCenter = &faultProcessCenter{
		notifyProcessChan: make(chan int, constant.MaxNotifyChanLen),
	}
}

// faultProcessCenter processes the faults and coordinates the fault handling among different components.
type faultProcessCenter struct {
	notifyProcessChan chan int
}

func (center *faultProcessCenter) Process() {
	cmprocess.DeviceCenter.Process()
	cmprocess.NodeCenter.Process()
	cmprocess.SwitchCenter.Process()
	jobprocess.FaultJobCenter.Process()
}

func (center *faultProcessCenter) notifyFaultCenterProcess(whichToProcess int) {
	center.notifyProcessChan <- whichToProcess
}

// Work faultProcessCenter work goroutine
func (center *faultProcessCenter) Work(ctx context.Context) {
	go func() {
		hwlog.RunLog.Info("faultProcessCenter start work!")
		centerTicker := time.NewTicker(time.Second)
		for {
			select {
			case <-ctx.Done():
				hwlog.RunLog.Info("faultProcessCenter stop work!")
				return
			case whichToProcess := <-center.notifyProcessChan:
				switch whichToProcess {
				case constant.AllProcessType:
					center.Process()
				case constant.DeviceProcessType:
					cmprocess.DeviceCenter.Process()
				case constant.NodeProcessType:
					cmprocess.NodeCenter.Process()
				case constant.SwitchProcessType:
					cmprocess.SwitchCenter.Process()
				default:
					hwlog.RunLog.Errorf("wrong number %d to process", whichToProcess)
				}
			case <-centerTicker.C:
				center.Process()
			}
		}
	}()
}

// Register to notify fault occurrence
func (center *faultProcessCenter) Register(ch chan int, whichToRegister int) {
	switch whichToRegister {
	case constant.SwitchProcessType:
		cmprocess.SwitchCenter.Register(ch)
	case constant.NodeProcessType:
		cmprocess.NodeCenter.Register(ch)
	case constant.DeviceProcessType:
		cmprocess.DeviceCenter.Register(ch)
	case constant.AllProcessType:
		cmprocess.SwitchCenter.Register(ch)
		cmprocess.NodeCenter.Register(ch)
		cmprocess.DeviceCenter.Register(ch)
	default:
		hwlog.RunLog.Errorf("Wrong number %d, cannot decide which to register", whichToRegister)
	}
}

// CallbackForReportUceInfo callback function to report uce info
func CallbackForReportUceInfo(infos []constant.ReportRecoverInfo) {
	for _, info := range infos {
		collector.ReportInfoCollector.ReportUceInfo(info.JobId, info.Rank, info.RecoverTime)
	}
	GlobalFaultProcessCenter.notifyFaultCenterProcess(constant.DeviceProcessType)
}

// QueryJobsFaultInfo query jobs fault rank info, and filter fault below `faultLevel`
func QueryJobsFaultInfo(faultLevel string) map[string]constant.JobFaultInfo {
	return faultrank.JobFaultRankProcessor.GetJobFaultRankInfosFilterLevel(faultLevel)
}

// QueryDeviceInfoToReport query device info to report
func QueryDeviceInfoToReport() map[string]*constant.AdvanceDeviceFaultCm {
	infos := cmprocess.DeviceCenter.GetProcessedCm()
	for _, info := range infos {
		info.UpdateTime = time.Now().Unix()
	}
	return infos
}

// QuerySwitchInfoToReport query switch info to report
func QuerySwitchInfoToReport() map[string]*constant.SwitchInfo {
	infos := cmprocess.SwitchCenter.GetProcessedCm()
	for _, info := range infos {
		info.UpdateTime = time.Now().Unix()
	}
	return infos
}

// QueryNodeInfoToReport query node info to report
func QueryNodeInfoToReport() map[string]*constant.NodeInfo {
	return cmprocess.NodeCenter.GetProcessedCm()
}

// DeviceInfoCollector collects device info
func DeviceInfoCollector(oldDevInfo, newDevInfo *constant.DeviceInfo, operator string) {
	collector.DeviceInfoCollector(oldDevInfo, newDevInfo, operator)
}

// SwitchInfoCollector collects switchinfo info of 900A3
func SwitchInfoCollector(oldSwitchInfo, newSwitchInfo *constant.SwitchInfo, operator string) {
	collector.SwitchInfoCollector(oldSwitchInfo, newSwitchInfo, operator)
}

// NodeCollector collects node info
func NodeCollector(oldNodeInfo, newNodeInfo *constant.NodeInfo, operator string) {
	collector.NodeCollector(oldNodeInfo, newNodeInfo, operator)
}

// PubFaultCollector collects public fault info
func PubFaultCollector(oldPubFaultInfo, newPubFaultInfo *api.PubFaultInfo, operator string) {
	if operator == constant.DeleteOperator {
		return
	}
	publicfault.PubFaultCollector(newPubFaultInfo)
}

// RegisterForJobFaultRank register for job fault info
func RegisterForJobFaultRank(ch chan map[string]constant.JobFaultInfo, src string) error {
	return jobprocess.FaultJobCenter.Register(ch, src)
}
