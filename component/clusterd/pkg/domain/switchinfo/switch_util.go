// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package switchinfo a series of switchinfo function
package switchinfo

import (
	"encoding/json"
	"fmt"
	"strings"

	"k8s.io/api/core/v1"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
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
	switchInfoCM := constant.SwitchFaultInfoFromCm{}
	data, ok := switchCm.Data[api.SwitchInfoCMDataKey]
	if !ok {
		return &constant.SwitchInfo{},
			fmt.Errorf("configmap %s has no key: %s", switchCm.Name, api.SwitchInfoCMDataKey)
	}

	if unmarshalErr := json.Unmarshal([]byte(data), &switchInfoCM); unmarshalErr != nil {
		return &constant.SwitchInfo{}, fmt.Errorf("unmarshal failed: %v, configmap name: %s", unmarshalErr, switchCm.Name)
	}
	faultInfo, err := parseSimpleSwitchFaultInfo(switchInfoCM.FaultCode, switchCm.Name)
	if err != nil {
		return &constant.SwitchInfo{}, err
	}
	nodeName := strings.TrimPrefix(switchCm.Name, constant.DeviceInfoPrefix)
	node := constant.SwitchInfo{
		SwitchFaultInfo: constant.SwitchFaultInfo{
			FaultInfo:            faultInfo,
			FaultLevel:           switchInfoCM.FaultLevel,
			UpdateTime:           switchInfoCM.UpdateTime,
			NodeStatus:           switchInfoCM.NodeStatus,
			FaultTimeAndLevelMap: switchInfoCM.FaultTimeAndLevelMap,
		},
		CmName: constant.SwitchInfoPrefix + nodeName,
	}
	return &node, nil
}

func parseSimpleSwitchFaultInfo(dataList []string, cm string) ([]constant.SimpleSwitchFaultInfo, error) {
	faultInfos := make([]constant.SimpleSwitchFaultInfo, 0, len(dataList))
	for _, data := range dataList {
		faultInfo := constant.SimpleSwitchFaultInfo{}
		unmarshalErr := json.Unmarshal([]byte(data), &faultInfo)
		if unmarshalErr != nil {
			return faultInfos, fmt.Errorf("unmarshal failed: %v, configmap name: %s", unmarshalErr, cm)
		}
		faultInfos = append(faultInfos, faultInfo)
	}
	return faultInfos, nil
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

// GetSafeData get data every 2000 SwitchInfo
func GetSafeData(switchInfos map[string]*constant.SwitchInfo) []string {
	if len(switchInfos) == 0 {
		return []string{}
	}
	if len(switchInfos) <= safeSwitchSize {
		return []string{util.ObjToString(getReportSwitchInfo(switchInfos))}
	}
	SwitchSlice := make([]string, 0, len(switchInfos)/safeSwitchSize+1)
	childSwitchInfos := make(map[string]*constant.SwitchInfo, safeSwitchSize)
	for cmName, switchInfo := range switchInfos {
		childSwitchInfos[cmName] = switchInfo
		if len(childSwitchInfos)%safeSwitchSize == 0 {
			SwitchSlice = append(SwitchSlice, util.ObjToString(getReportSwitchInfo(childSwitchInfos)))
			childSwitchInfos = make(map[string]*constant.SwitchInfo, safeSwitchSize)
		}
	}
	if len(childSwitchInfos) != 0 {
		SwitchSlice = append(SwitchSlice, util.ObjToString(getReportSwitchInfo(childSwitchInfos)))
	}
	return SwitchSlice
}

func getReportSwitchInfo(switchInfoMap map[string]*constant.SwitchInfo) map[string]*constant.SwitchInfoFromCM {
	reportSwitchInfo := make(map[string]*constant.SwitchInfoFromCM, len(switchInfoMap))
	for k, v := range switchInfoMap {
		reportFaultCodes := make([]string, 0, len(v.FaultInfo))
		for _, faultInfo := range v.FaultInfo {
			faultBytes, err := json.Marshal(faultInfo)
			if err != nil {
				hwlog.RunLog.Warnf("failed to convert fault:%v, err: %v", faultInfo, err)
				continue
			}
			reportFaultCodes = append(reportFaultCodes, string(faultBytes))
		}
		reportSwitchInfo[k] = &constant.SwitchInfoFromCM{
			SwitchFaultInfoFromCm: constant.SwitchFaultInfoFromCm{
				FaultCode:            reportFaultCodes,
				FaultLevel:           v.FaultLevel,
				UpdateTime:           v.UpdateTime,
				NodeStatus:           v.NodeStatus,
				FaultTimeAndLevelMap: v.FaultTimeAndLevelMap,
			},
			CmName: v.CmName,
		}
	}
	return reportSwitchInfo
}
