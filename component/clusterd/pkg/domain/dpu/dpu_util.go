// Copyright (c) Huawei Technologies Co., Ltd. 2026-2026. All rights reserved.

// Package dpu a series of dpu function
package dpu

import (
	"encoding/json"
	"fmt"

	v1 "k8s.io/api/core/v1"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
)

const (
	// maxCmDataSize is the max data size for a single ConfigMap (~1MB limit, using 800KB for safety margin)
	maxCmDataSize = 800 * 1024
)

// DpuInfoCMDataKey is the data key in dpu configmap, aligned with dpu-dp
const DpuInfoCMDataKey = "DpuInfoCfg"

// ParseDpuInfoCM get dpu info from configmap obj
func ParseDpuInfoCM(dpuCm *v1.ConfigMap) (*constant.DpuInfo, error) {
	data, ok := dpuCm.Data[DpuInfoCMDataKey]
	if !ok {
		return &constant.DpuInfo{}, fmt.Errorf("configmap %s has no %s", dpuCm.Name, DpuInfoCMDataKey)
	}

	var dpuInfoCfg constant.DpuInfoCfg
	if unmarshalErr := json.Unmarshal([]byte(data), &dpuInfoCfg); unmarshalErr != nil {
		return &constant.DpuInfo{}, fmt.Errorf("unmarshal failed: %v, configmap name: %s", unmarshalErr, dpuCm.Name)
	}

	return &constant.DpuInfo{
		DpuInfoCfg: dpuInfoCfg,
		CmName:     dpuCm.Name,
	}, nil
}

// DeepCopy deep copy DpuInfo
func DeepCopy(info *constant.DpuInfo) *constant.DpuInfo {
	if info == nil {
		return nil
	}
	data, err := json.Marshal(info)
	if err != nil {
		hwlog.RunLog.Errorf("marshal dpu failed , err is %v", err)
		return nil
	}
	newDpuInfo := &constant.DpuInfo{}
	if err := json.Unmarshal(data, newDpuInfo); err != nil {
		hwlog.RunLog.Errorf("unmarshal dpu failed , err is %v", err)
		return nil
	}
	return newDpuInfo
}

// GetSafeData splits dpuInfos into chunks that fit within K8s ConfigMap size limit (~1MB).
// Each chunk is as close to maxCmDataSize (800KB) as possible.
func GetSafeData(dpuInfos map[string]*constant.DpuInfo) []string {
	return util.SplitMapToSafeChunks(dpuInfos, maxCmDataSize,
		func(m map[string]*constant.DpuInfo) string {
			return util.ObjToString(getReportDpuInfo(m))
		})
}

// getReportDpuInfo converts DpuInfo to a map of DpuInfoCfg keyed by cm name for serialization
func getReportDpuInfo(dpuInfoMap map[string]*constant.DpuInfo) map[string]*constant.DpuInfoCfg {
	reportDpuInfo := make(map[string]*constant.DpuInfoCfg, len(dpuInfoMap))
	for k, v := range dpuInfoMap {
		cfg := v.DpuInfoCfg
		reportDpuInfo[k] = &cfg
	}
	return reportDpuInfo
}
