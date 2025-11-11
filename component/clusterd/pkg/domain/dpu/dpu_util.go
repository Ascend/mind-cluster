// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package dpu a series of dpu function
package dpu

import (
	"encoding/json"
	"fmt"

	"k8s.io/api/core/v1"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"

	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
)

const safeDpuCMSize = 1000

// ParseDpuInfoCM get dpu info from configmap obj
func ParseDpuInfoCM(obj interface{}) (*constant.DpuInfoCM, error) {
	dpuCM, ok := obj.(*v1.ConfigMap)
	if !ok {
		return nil, fmt.Errorf("%s not a ConfigMap", api.DpuLogPrefix)
	}

	var dpuList constant.DpuCMDataList
	dpuListStr, ok := dpuCM.Data[api.DpuInfoCMDataKey]
	if !ok {
		return nil, fmt.Errorf("%s <%s> configmap has no key <%s>", api.DpuLogPrefix, dpuCM.Name, api.DpuInfoCMDataKey)
	}
	if err := json.Unmarshal([]byte(dpuListStr), &dpuList); err != nil {
		return nil, fmt.Errorf("%s <%s> configmap unmarshal data failed: <%v>", api.DpuLogPrefix, err, dpuCM.Name)
	}
	busType, getBusTypeOk := dpuCM.Data[api.DpuInfoCMBusTypeKey]
	if !getBusTypeOk {
		return nil, fmt.Errorf("%s <%s> configmap has no key <%s>", api.DpuLogPrefix, dpuCM.Name,
			api.DpuInfoCMBusTypeKey)
	}
	var npuToDpusMap map[string][]string
	npuToDpusStr, getMapOk := dpuCM.Data[api.DpuInfoCMNpuToDpusMapKey]
	if !getMapOk {
		return nil, fmt.Errorf("%s <%s> configmap has no key <%s>", api.DpuLogPrefix, dpuCM.Name,
			api.DpuInfoCMNpuToDpusMapKey)
	}
	if err := json.Unmarshal([]byte(npuToDpusStr), &npuToDpusMap); err != nil {
		return nil, fmt.Errorf("%s <%s> configmap unmarshal data failed: <%v>", api.DpuLogPrefix, err, dpuCM.Name)
	}

	result := constant.DpuInfoCM{
		CmName:       dpuCM.Name,
		DPUList:      dpuList,
		BusType:      busType,
		NpuToDpusMap: npuToDpusMap,
	}
	return &result, nil
}

// DeepCopy deep copy dpu info
func DeepCopy(info *constant.DpuInfoCM) *constant.DpuInfoCM {
	if info == nil {
		return nil
	}
	data, err := json.Marshal(info)
	if err != nil {
		hwlog.RunLog.Errorf("%s json marshal failed: %v", api.DpuLogPrefix, err)
		return nil
	}
	newInfo := &constant.DpuInfoCM{}
	if unErr := json.Unmarshal(data, newInfo); unErr != nil {
		hwlog.RunLog.Errorf("%s json unmarshal failed: %v", api.DpuLogPrefix, unErr)
		return nil
	}
	return newInfo
}

// GetSafeData put every 1000 dpu cm info together
func GetSafeData(dpuCMInfos map[string]*constant.DpuInfoCM) []string {
	dpuCMCount := len(dpuCMInfos)
	if dpuCMCount == 0 {
		return []string{}
	}
	if dpuCMCount <= safeDpuCMSize {
		return []string{util.ObjToString(dpuCMInfos)}
	}

	dpuSlice := make([]string, 0, (dpuCMCount+safeDpuCMSize-1)/safeDpuCMSize)
	dpuSliceTemp := make(map[string]*constant.DpuInfoCM, safeDpuCMSize)
	for cmName, dpuInfo := range dpuCMInfos {
		dpuSliceTemp[cmName] = dpuInfo
		if len(dpuSliceTemp)%safeDpuCMSize == 0 {
			dpuSlice = append(dpuSlice, util.ObjToString(dpuSliceTemp))
			dpuSliceTemp = make(map[string]*constant.DpuInfoCM, safeDpuCMSize)
		}
	}
	if len(dpuSliceTemp) != 0 {
		dpuSlice = append(dpuSlice, util.ObjToString(dpuSliceTemp))
	}
	return dpuSlice
}
