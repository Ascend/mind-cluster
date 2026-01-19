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

// Package device a series of device function
package device

import (
	"context"
	"fmt"

	"github.com/containerd/containerd"
	"k8s.io/apimachinery/pkg/util/sets"

	"ascend-common/common-utils/hwlog"
)

// GetUsedChips return chips used by process and containerd
func (tool *AscendTools) GetUsedChips() sets.String {
	procUsedChips := tool.getChipsUsedByProcess()
	containerUsedChips := tool.getChipsUsedByContainerd()
	usedChips := procUsedChips.Union(containerUsedChips)
	return usedChips
}

// getChipsUsedByProcess return chips used by process
func (tool *AscendTools) getChipsUsedByProcess() sets.String {
	return sets.String{}
}

// getChipsUsedByContainerd return chips used by process
func (tool *AscendTools) getChipsUsedByContainerd() sets.String {
	return sets.String{}
}

func (tool *AscendTools) getDeviceWithoutAscendRuntime(containerObj containerd.Container,
	ctx context.Context) sets.String {
	usedChips := sets.NewString()
	spec, err := getContainerValidSpec(containerObj, ctx)
	if err != nil {
		hwlog.RunLog.Debugf("failed to get container valid spec: %v", err)
		return usedChips
	}
	deviceIDs, err := filterNPUDevices(spec)
	if err != nil {
		hwlog.RunLog.Debugf("failed to get device ids: %v", err)
		return usedChips
	}
	hwlog.RunLog.Debugf("filter npu devices get deviceIDs: %v", deviceIDs)
	for _, deviceID := range deviceIDs {
		chipName := fmt.Sprintf("%s-%d", tool.name, deviceID)
		usedChips.Insert(chipName)
	}
	return usedChips
}
