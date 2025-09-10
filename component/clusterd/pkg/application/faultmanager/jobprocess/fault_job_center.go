// Copyright (c) Huawei Technologies Co., Ltd. 2024-2025. All rights reserved.

// Package jobprocess contain job fault process
package jobprocess

import (
	"fmt"
	"sync"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/application/faultmanager/cmprocess"
	"clusterd/pkg/application/faultmanager/jobprocess/faultrank"
	"clusterd/pkg/application/faultmanager/jobprocess/relationfault"
	"clusterd/pkg/common/constant"
)

type faultJobProcessCenter struct {
	processorList        []constant.FaultProcessor
	subscribeChannelList []*subscriber
	mutex                sync.Mutex
}

type subscriber struct {
	ch  chan map[string]constant.JobFaultInfo
	src string
}

// FaultJobCenter process fault about job
var FaultJobCenter *faultJobProcessCenter

func init() {
	FaultJobCenter = &faultJobProcessCenter{
		processorList: []constant.FaultProcessor{
			relationfault.RelationProcessor,
			faultrank.JobFaultRankProcessor,
		},
		mutex:                sync.Mutex{},
		subscribeChannelList: make([]*subscriber, 0),
	}
}

func (fJobCenter *faultJobProcessCenter) Process() {
	content := constant.AllConfigmapContent{
		DeviceCm: cmprocess.DeviceCenter.GetProcessedCm(),
		SwitchCm: cmprocess.SwitchCenter.GetProcessedCm(),
		NodeCm:   cmprocess.NodeCenter.GetProcessedCm(),
	}
	for _, processor := range fJobCenter.processorList {
		processor.Process(content)
	}
	fJobCenter.notifySubscriber()
}

// Register notify chan
func (fJobCenter *faultJobProcessCenter) Register(ch chan map[string]constant.JobFaultInfo, src string) error {
	if ch == nil {
		return fmt.Errorf("invalid chanel for send")
	}
	fJobCenter.mutex.Lock()
	defer fJobCenter.mutex.Unlock()
	length := len(fJobCenter.subscribeChannelList)
	if length > constant.MaxFaultCenterSubscriber {
		return fmt.Errorf("the number of registrants is %d, cannot add any more", length)
	}
	fJobCenter.subscribeChannelList = append(fJobCenter.subscribeChannelList, &subscriber{
		ch:  ch,
		src: src,
	})
	return nil
}

func (fJobCenter *faultJobProcessCenter) notifySubscriber() {
	for _, sub := range fJobCenter.subscribeChannelList {
		if sub.ch == nil {
			continue
		}
		faultRankInfos := faultrank.JobFaultRankProcessor.GetJobFaultRankInfosFilterLevel(
			[]string{constant.NotHandleFault, constant.PreSeparateNPU})
		select {
		case sub.ch <- faultRankInfos:
		default:
			hwlog.RunLog.Warnf("notify %s fault rank failed.", sub.src)
		}
	}
}
