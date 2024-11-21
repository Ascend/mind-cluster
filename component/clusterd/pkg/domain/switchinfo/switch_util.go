// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package switchinfo a series of switchinfo function
package switchinfo

import (
	"encoding/json"
	"fmt"
	"strings"

	"huawei.com/npu-exporter/v6/common-utils/hwlog"
	"k8s.io/api/core/v1"

	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
)

const safeSwitchSize = 2000

// ParseSwitchInfoCM get node info from configmap obj
func ParseSwitchInfoCM(obj interface{}) (*constant.SwitchInfo, error) {
	switchCm, ok := obj.(*v1.ConfigMap)
	if !ok {
		return &constant.SwitchInfo{}, fmt.Errorf("not configmap")
	}
	switchInfoCM := constant.SwitchFaultInfo{}
	data, ok := switchCm.Data[constant.SwitchInfoCmKey]
	if !ok {
		return &constant.SwitchInfo{},
			fmt.Errorf("configmap %s has no key: %s", switchCm.Name, constant.SwitchInfoCmKey)
	}

	if unmarshalErr := json.Unmarshal([]byte(data), &switchInfoCM); unmarshalErr != nil {
		return &constant.SwitchInfo{}, fmt.Errorf("unmarshal failed: %v, configmap name: %s", unmarshalErr, switchCm.Name)
	}

	nodeName := strings.TrimPrefix(switchCm.Name, constant.DeviceInfoPrefix)
	node := constant.SwitchInfo{
		SwitchFaultInfo: constant.SwitchFaultInfo{
			FaultCode:  switchInfoCM.FaultCode,
			FaultLevel: switchInfoCM.FaultLevel,
			UpdateTime: switchInfoCM.UpdateTime,
			NodeStatus: switchInfoCM.NodeStatus,
		},
		CmName: constant.SwitchInfoPrefix + nodeName,
	}
	return &node, nil
}

// DeepCopy deep copy NodeInfo
func DeepCopy(info *constant.SwitchInfo) (*constant.SwitchInfo, error) {
	if info == nil {
		return nil, nil
	}
	data, err := json.Marshal(info)
	if err != nil {
		hwlog.RunLog.Errorf("marshal switchinfo failed , err is %v", err)
		return nil, err
	}
	newSwitchInfo := &constant.SwitchInfo{}
	if err := json.Unmarshal(data, newSwitchInfo); err != nil {
		hwlog.RunLog.Errorf("unmarshal switchinfo failed , err is %v", err)
		return nil, err
	}
	return newSwitchInfo, nil
}

// DeepCopyInfos deep copy NodeInfo
func DeepCopyInfos(infos map[string]*constant.SwitchInfo) map[string]*constant.SwitchInfo {
	res := make(map[string]*constant.SwitchInfo)
	for key, val := range infos {
		res[key], _ = DeepCopy(val)
	}
	return res
}

// GetSafeData get data every 2000 SwitchInfo
func GetSafeData(switchInfos map[string]*constant.SwitchInfo) []string {
	if len(switchInfos) == 0 {
		return []string{}
	}
	if len(switchInfos) <= safeSwitchSize {
		return []string{util.ObjToString(switchInfos)}
	}
	SwitchSlice := make([]string, 0, len(switchInfos)/safeSwitchSize+1)
	childSwitchInfos := make(map[string]*constant.SwitchInfo, safeSwitchSize)
	for cmName, switchInfo := range switchInfos {
		childSwitchInfos[cmName] = switchInfo
		if len(childSwitchInfos)%safeSwitchSize == 0 {
			SwitchSlice = append(SwitchSlice, util.ObjToString(childSwitchInfos))
			childSwitchInfos = make(map[string]*constant.SwitchInfo, safeSwitchSize)
		}
	}
	if len(childSwitchInfos) != 0 {
		SwitchSlice = append(SwitchSlice, util.ObjToString(childSwitchInfos))
	}
	return SwitchSlice
}

// BusinessDataIsNotEqual judge is the faultcode and fault level is the same as known, if is not same returns true
func BusinessDataIsNotEqual(oldSwitch, newSwitch *constant.SwitchInfo) bool {
	if oldSwitch == nil && newSwitch == nil {
		return false
	}
	if (oldSwitch != nil && newSwitch == nil) || (oldSwitch == nil && newSwitch != nil) {
		return true
	}
	if newSwitch.FaultLevel != oldSwitch.FaultLevel || newSwitch.NodeStatus != oldSwitch.NodeStatus ||
		len(newSwitch.FaultCode) != len(oldSwitch.FaultCode) {
		return true
	}
	return false
}
