// Copyright (c) Huawei Technologies Co., Ltd. 2024-2025. All rights reserved.

// Package jobprocess contain job fault process
package jobprocess

import (
	"time"

	"clusterd/pkg/application/faultmanager/cmprocess"
	"clusterd/pkg/application/faultmanager/jobprocess/faultrank"
	"clusterd/pkg/application/faultmanager/jobprocess/relationfault"
	"clusterd/pkg/common/constant"
)

// FaultJobCenter process fault about job
var FaultJobCenter *faultJobProcessCenter

type faultJobProcessCenter struct {
	lastProcessTime int64
	processorList   []constant.FaultProcessor
}

func init() {
	FaultJobCenter = &faultJobProcessCenter{
		lastProcessTime: 0,
		processorList: []constant.FaultProcessor{
			relationfault.RelationProcessor,
			faultrank.JobFaultRankProcessor,
		},
	}
}

func (fJobCenter *faultJobProcessCenter) Process() {
	currentTime := time.Now().UnixMilli()
	if fJobCenter.isProcessLimited(currentTime) {
		return
	}
	fJobCenter.lastProcessTime = currentTime
	content := constant.AllConfigmapContent{
		DeviceCm: cmprocess.DeviceCenter.GetProcessedCm(),
		SwitchCm: cmprocess.SwitchCenter.GetProcessedCm(),
		NodeCm:   cmprocess.NodeCenter.GetProcessedCm(),
	}
	for _, processor := range fJobCenter.processorList {
		processor.Process(content)
	}
}

func (fJobCenter *faultJobProcessCenter) isProcessLimited(currentTime int64) bool {
	return fJobCenter.lastProcessTime+constant.FaultCenterProcessPeriod > currentTime
}
