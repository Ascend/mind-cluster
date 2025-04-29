// Copyright (c) Huawei Technologies Co., Ltd. 2024-2025. All rights reserved.

// Package collector collect information to process
package collector

import (
	"sync"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/faultdomain"
)

var DeviceCmCollectBuffer *ConfigmapCollectBuffer[*constant.AdvanceDeviceFaultCm]
var NodeCmCollectBuffer *ConfigmapCollectBuffer[*constant.NodeInfo]
var SwitchCmCollectBuffer *ConfigmapCollectBuffer[*constant.SwitchInfo]

type ConfigmapCollectBuffer[T constant.ConfigMapInterface] struct {
	mutex    sync.Mutex
	buffer   map[string]*[]constant.InformerCmItem[T]
	lastItem map[string]constant.InformerCmItem[T]
}

func init() {
	DeviceCmCollectBuffer = &ConfigmapCollectBuffer[*constant.AdvanceDeviceFaultCm]{
		mutex:    sync.Mutex{},
		buffer:   make(map[string]*[]constant.InformerCmItem[*constant.AdvanceDeviceFaultCm]),
		lastItem: make(map[string]constant.InformerCmItem[*constant.AdvanceDeviceFaultCm]),
	}
	NodeCmCollectBuffer = &ConfigmapCollectBuffer[*constant.NodeInfo]{
		mutex:    sync.Mutex{},
		buffer:   make(map[string]*[]constant.InformerCmItem[*constant.NodeInfo]),
		lastItem: make(map[string]constant.InformerCmItem[*constant.NodeInfo]),
	}
	SwitchCmCollectBuffer = &ConfigmapCollectBuffer[*constant.SwitchInfo]{
		mutex:    sync.Mutex{},
		buffer:   make(map[string]*[]constant.InformerCmItem[*constant.SwitchInfo]),
		lastItem: make(map[string]constant.InformerCmItem[*constant.SwitchInfo]),
	}
}

func (cmCollector *ConfigmapCollectBuffer[T]) Push(info T, isAdd bool) bool {
	cmCollector.mutex.Lock()
	defer cmCollector.mutex.Unlock()
	queue, found := cmCollector.buffer[info.GetCmName()]
	if !found {
		queue = &[]constant.InformerCmItem[T]{}
		cmCollector.buffer[info.GetCmName()] = queue
	}
	if len(*queue) > constant.MaxCmQueueLen {
		hwlog.RunLog.Warnf("queue of %s is exceeded %d", info.GetCmName(), constant.MaxCmQueueLen)
	}
	newItem := constant.InformerCmItem[T]{
		IsAdd: isAdd,
		Data:  info,
	}
	lastItem, found := cmCollector.lastItem[info.GetCmName()]
	if !found || !informerItemEqual(lastItem, newItem) {
		cmCollector.lastItem[info.GetCmName()] = newItem
		*queue = append(*queue, newItem)
		return true
	}
	return false
}

func (cmCollector *ConfigmapCollectBuffer[T]) Pop() []constant.InformerCmItem[T] {
	cmCollector.mutex.Lock()
	defer cmCollector.mutex.Unlock()
	result := make([]constant.InformerCmItem[T], 0)
	for _, queue := range cmCollector.buffer {
		if len(*queue) == 0 {
			continue
		}
		result = append(result, (*queue)[0])
		*queue = (*queue)[1:]
	}
	return result
}

func informerItemEqual[T constant.ConfigMapInterface](lastItem, newItem constant.InformerCmItem[T]) bool {
	if lastItem.IsAdd == newItem.IsAdd && lastItem.Data.IsSame(newItem.Data) {
		return true
	}
	return false
}

func informInfoUpdate(newInfo any, whichToInformer int, isAdd bool) {
	switch whichToInformer {
	case constant.DeviceProcessType:
		advanceFaultForNode :=
			faultdomain.GetAdvanceFaultForNode(newInfo.(*constant.DeviceInfo)).(*constant.AdvanceDeviceFaultCm)
		faultdomain.SortDataForAdvanceDeviceInfo(advanceFaultForNode)
		DeviceCmCollectBuffer.Push(advanceFaultForNode, isAdd)
	case constant.NodeProcessType:
		NodeCmCollectBuffer.Push(newInfo.(*constant.NodeInfo), isAdd)
	case constant.SwitchProcessType:
		SwitchCmCollectBuffer.Push(newInfo.(*constant.SwitchInfo), isAdd)
	default:
		hwlog.RunLog.Errorf("cannot process %d", whichToInformer)
		return
	}
}

// DeviceInfoCollector collects device info
func DeviceInfoCollector(oldDevInfo, newDevInfo *constant.DeviceInfo, operator string) {
	if operator == constant.AddOperator || operator == constant.UpdateOperator {
		informInfoUpdate(newDevInfo, constant.DeviceProcessType, true)
	} else if operator == constant.DeleteOperator {
		informInfoUpdate(newDevInfo, constant.DeviceProcessType, false)
	}
}

// SwitchInfoCollector collects switchinfo info of 900A3
func SwitchInfoCollector(oldSwitchInfo, newSwitchInfo *constant.SwitchInfo, operator string) {
	if operator == constant.AddOperator || operator == constant.UpdateOperator {
		informInfoUpdate(newSwitchInfo, constant.SwitchProcessType, true)
	} else if operator == constant.DeleteOperator {
		informInfoUpdate(newSwitchInfo, constant.SwitchProcessType, false)
	}
}

// NodeCollector collects node info
func NodeCollector(oldNodeInfo, newNodeInfo *constant.NodeInfo, operator string) {
	if operator == constant.AddOperator || operator == constant.UpdateOperator {
		informInfoUpdate(newNodeInfo, constant.NodeProcessType, true)
	} else if operator == constant.DeleteOperator {
		informInfoUpdate(newNodeInfo, constant.NodeProcessType, false)
	}
}
