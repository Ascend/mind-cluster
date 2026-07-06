// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package switchinfo a series of switchinfo function
package switchinfo

import (
	"encoding/json"
	"fmt"
	"strings"

	v1 "k8s.io/api/core/v1"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
)

const (
	// maxCmDataSize is the max data size for a single ConfigMap (~1MB limit, using 800KB for safety margin)
	maxCmDataSize = 800 * 1024
)

// ParseSwitchInfoCM get node info from configmap obj
func ParseSwitchInfoCM(switchCm *v1.ConfigMap) (*constant.SwitchInfo, error) {
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

// GetSafeData splits switchInfos into chunks that fit within K8s ConfigMap size limit (~1MB).
// Each chunk is as close to maxCmDataSize (800KB) as possible.
func GetSafeData(switchInfos map[string]*constant.SwitchInfo) []string {
	return util.SplitMapToSafeChunks(switchInfos, maxCmDataSize,
		func(m map[string]*constant.SwitchInfo) string {
			return util.ObjToString(getReportSwitchInfo(m))
		})
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
