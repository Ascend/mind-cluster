// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package stresstest contain filtering fault handling method for stress test faults
package stresstest

import (
	"sync"
	"time"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/faultdomain"
)

// StressTestProcessor stress test fault processor
var StressTestProcessor *stressTestFaultProcessor

type stressTestFaultProcessor struct {
	devCMInfo      *constant.AdvanceDeviceFaultCm
	jobFilterFault map[string][]string // jobID -> filter nodes
	mutex          sync.Mutex
}

func init() {
	StressTestProcessor = &stressTestFaultProcessor{
		devCMInfo:      &constant.AdvanceDeviceFaultCm{},
		jobFilterFault: make(map[string][]string),
		mutex:          sync.Mutex{},
	}
}

// Process stress test fault process
func (p *stressTestFaultProcessor) Process(info any) any {
	processContent, ok := info.(constant.OneConfigmapContent[*constant.AdvanceDeviceFaultCm])
	if !ok {
		hwlog.RunLog.Errorf("input is not DeviceInfo type, info:%v", info)
		return info
	}

	deviceInfos := processContent.AllConfigmap
	for _, filterNodes := range p.jobFilterFault {
		for _, node := range filterNodes {
			p.processNodeFaults(deviceInfos[node], node)
		}
	}
	processContent.AllConfigmap = deviceInfos
	return processContent
}

func (p *stressTestFaultProcessor) processNodeFaults(devInfo *constant.AdvanceDeviceFaultCm, node string) {
	if devInfo == nil {
		return
	}
	sortData := false
	for _, devFaults := range devInfo.FaultDeviceList {
		for _, fault := range devFaults {
			if faultdomain.IsStressTestFault(fault.FaultCode) {
				hwlog.RunLog.Infof("node:%v filter stress test fault:%v", node, fault)
				devInfo.DelFaultAndFix(fault)
				sortData = true
			}
		}
	}
	if sortData {
		devInfo.UpdateTime = time.Now().Unix()
		faultdomain.SortDataForAdvanceDeviceInfo(devInfo)
	}
}

// SetFilterAicFault set filter aic fault
func (p *stressTestFaultProcessor) SetFilterAicFault(jobID string, nodes []string, filter bool) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if filter {
		p.jobFilterFault[jobID] = nodes
		hwlog.RunLog.Infof("jobId:%v add filter nodes: %v", jobID, nodes)
	} else {
		delete(p.jobFilterFault, jobID)
		hwlog.RunLog.Infof("jobId:%v delete filter nodes", jobID)
	}

}
