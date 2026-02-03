/* Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

// Package server holds the implementation of registration to kubelet, k8s pod resource interface.
package server

import (
	"context"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"

	"Ascend-device-plugin/pkg/common"
	"Ascend-device-plugin/pkg/device/dpucontrol"
	"ascend-common/api"
	"ascend-common/common-utils/ethtool"
	"ascend-common/common-utils/hwlog"
)

const (
	listenDpuInterval = 5 * time.Second
	maxUpdateInterval = 6 * time.Hour
)

var (
	lastData       []common.DpuCMData
	lastUpdateTime time.Time
	uniqueDpuMap   map[string]dpucontrol.BaseDpuInfo
	npuToDpuMap    map[string][]string
)

// ListenDpu periodically query the DPU operstate and write to the configmap
func (hdm *HwDevManager) ListenDpu(ctx context.Context) {
	if common.ParamOption.RealCardType != api.Ascend910A5 || len(hdm.dpuManager.NpuWithDpuInfos) == 0 {
		return
	}

	uniqueDpuMap = make(map[string]dpucontrol.BaseDpuInfo, api.NpuCountPerNode)
	npuToDpuMap = make(map[string][]string, api.NpuCountPerNode)
	for _, npuWithDpu := range hdm.dpuManager.NpuWithDpuInfos {
		strKey := strconv.Itoa(int(npuWithDpu.NpuId))
		npuToDpuMap[strKey] = make([]string, 0)
		for _, dpu := range npuWithDpu.DpuInfo {
			if _, exists := uniqueDpuMap[dpu.DeviceName]; !exists {
				uniqueDpuMap[dpu.DeviceName] = dpu
			}
			npuToDpuMap[strKey] = append(npuToDpuMap[strKey], dpu.DeviceName)
		}
	}

	ticker := time.NewTicker(listenDpuInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			hdm.handleDpu()
		case <-ctx.Done():
			hwlog.RunLog.Warnf("%s ListenDpu stop work due to ctx.Done(): %v", api.DpuLogPrefix, ctx.Err())
			return
		}
	}
}

func (hdm *HwDevManager) handleDpu() {
	var dpuList []common.DpuCMData
	for _, dpu := range uniqueDpuMap {
		state, err := ethtool.GetInterfaceOperState(dpu.DeviceName)
		if err != nil {
			hwlog.RunLog.Errorf("%s GetInterfaceOperState err: %v", api.DpuLogPrefix, err)
			state = api.DpuStatusDown
		}
		dpuList = append(dpuList, common.DpuCMData{
			Name:      dpu.DeviceName,
			Operstate: state,
			DeviceID:  dpu.DeviceId,
			VendorID:  dpu.Vendor,
		})
	}
	sort.Slice(dpuList, func(i, j int) bool {
		return dpuList[i].Name < dpuList[j].Name
	})
	needUpdate := !reflect.DeepEqual(dpuList, lastData) || time.Since(lastUpdateTime) > maxUpdateInterval
	if !needUpdate {
		return
	}
	hdm.manager.SetDpu(hdm.dpuManager.UserConfig.BusType, dpuList, npuToDpuMap)
	lastData = dpuList
	lastUpdateTime = time.Now()
}

func (hdm *HwDevManager) updateDpuHealthy(groupDevice map[string][]*common.NpuDevice) {
	uniqueMap := make(map[string]dpucontrol.BaseDpuInfo, api.NpuCountPerNode)
	npuToDpusMap := make(map[string][]string, api.NpuCountPerNode)
	for _, npuWithDpu := range hdm.dpuManager.NpuWithDpuInfos {
		strKey := strconv.Itoa(int(npuWithDpu.NpuId))
		npuToDpusMap[strKey] = make([]string, 0)
		for _, dpu := range npuWithDpu.DpuInfo {
			if _, exists := uniqueMap[dpu.DeviceName]; !exists {
				uniqueMap[dpu.DeviceName] = dpu
			}
			npuToDpusMap[strKey] = append(npuToDpusMap[strKey], dpu.DeviceName)
		}
	}
	dpuOperstateMap := make(map[string]string, api.NpuCountPerNode)
	for _, dpu := range uniqueMap {
		state, err := ethtool.GetInterfaceOperState(dpu.DeviceName)
		if err != nil {
			hwlog.RunLog.Errorf("%s GetInterfaceOperState err: %v", api.DpuLogPrefix, err)
			state = api.DpuStatusDown
		}
		dpuOperstateMap[dpu.DeviceName] = state
	}
	for _, devices := range groupDevice {
		for _, device := range devices {
			device.DpuHealth = getDpuFaultInfoOfNpu(device.DeviceName, npuToDpusMap, dpuOperstateMap)
		}
	}
}

// getDpuFaultInfoOfNpu get the dpu fault status of one npu
func getDpuFaultInfoOfNpu(npuName string, npuToDpuMap map[string][]string,
	dpuOperstateMap map[string]string) string {
	dpuFaultCount := 0
	for npu, dpus := range npuToDpuMap {
		if !isNpuMatched(npuName, npu) {
			continue
		}
		// PCIe: One NPU has only one DPU, so the for loop only iterates once.
		// UB: One NPU has two DPUs, but as long as one DPU is active, it's fine.
		for _, dpu := range dpus {
			if dpuOperstateMap[dpu] != api.DpuStatusUp {
				dpuFaultCount++
			}
		}
		if dpuFaultCount == len(dpus) {
			hwlog.RunLog.Debugf("%s for npu<%s> dpu<%v> are faulty", api.DpuLogPrefix, npu, dpus)
			return v1beta1.Unhealthy
		} else if dpuFaultCount == common.OneDpuFault {
			hwlog.RunLog.Debugf("%s for npu<%s> one of its dpus<%v> is faulty", api.DpuLogPrefix, npu, dpus)
			return api.DpuSubHealthy
		}
		return v1beta1.Healthy
	}
	return v1beta1.Healthy
}

func isNpuMatched(npuName string, target string) bool {
	parts := strings.Split(npuName, "-")
	if len(parts) <= 1 {
		hwlog.RunLog.Errorf("get id by spliting npu card failed")
		return false
	}
	cardId, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil {
		hwlog.RunLog.Errorf("get card id failed:%s", err)
		return false
	}
	cardIdStr := strconv.Itoa(cardId % api.NpuCountPerNode)
	if cardIdStr == target {
		return true
	}

	return false
}
