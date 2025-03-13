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
	"strings"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/api/services/tasks/v1"
	"github.com/containerd/containerd/namespaces"
	"k8s.io/apimachinery/pkg/util/sets"

	"Ascend-device-plugin/pkg/common"
	"ascend-common/common-utils/hwlog"
)

// GetUsedChips return chips used by process and containerd
func (tool *AscendTools) GetUsedChips() sets.String {
	procUsedChips := tool.getChipsUsedByProcess()
	containerUsedChips := tool.getChipsUsedByContainerd()
	usedChips := procUsedChips.Union(containerUsedChips)
	hwlog.RunLog.Debugf("get used chips: %v", usedChips)
	return usedChips
}

// getChipsUsedByProcess return chips used by process
func (tool *AscendTools) getChipsUsedByProcess() sets.String {
	if !common.ParamOption.PresetVDevice {
		return sets.String{}
	}
	_, logicIDs, err := tool.dmgr.GetDeviceList()
	if err != nil {
		hwlog.RunLog.Warnf("get device list failed, err: %v", err)
		return sets.String{}
	}
	if len(logicIDs) < 1 {
		hwlog.RunLog.Warn("get device list failed, logicID is empty")
		return sets.String{}
	}
	usedChips := make([]string, 0, len(logicIDs))
	for _, logicID := range logicIDs {
		chipInfo, err := tool.dmgr.GetDevProcessInfo(logicID)
		if err != nil {
			// use vnpu will report an 8255 error
			hwlog.RunLog.Debugf("get device process info failed, err: %v", err)
			continue
		}
		if chipInfo.ProcNum != 0 {
			hwlog.RunLog.Debugf("the card logicID:[%d] is used, chipInfo: %#v", logicID, chipInfo)
			davinCidev, err := tool.getDavinCiDev(logicID)
			if err != nil {
				hwlog.RunLog.Errorf("get davinci dev by logicID:[%d] failed, err: %v", logicID, err)
				continue
			}
			chipName := fmt.Sprintf("%s-%d", tool.name, davinCidev.PhyID)
			usedChips = append(usedChips, chipName)
		}
	}
	hwlog.RunLog.Debugf("process used chips: %v", usedChips)
	return sets.NewString(usedChips...)
}

// getChipsUsedByContainerd return chips used by process
func (tool *AscendTools) getChipsUsedByContainerd() sets.String {
	usedChips := sets.NewString()
	if tool.containerdClient == nil {
		hwlog.RunLog.Debug("containerd client is nil")
		return usedChips
	}
	nss, err := tool.containerdClient.NamespaceService().List(context.Background())
	if err != nil {
		hwlog.RunLog.Warnf("failed to get namespace list: %v", err)
		return usedChips
	}
	hwlog.RunLog.Debugf("containerd namespace list: %v", nss)
	for _, ns := range nss {
		ctx := namespaces.WithNamespace(context.Background(), ns)
		taskList, err := tool.containerdClient.TaskService().List(ctx, &tasks.ListTasksRequest{})
		if err != nil {
			hwlog.RunLog.Warnf("failed to get task list: %v", err)
			continue
		}
		if len(taskList.Tasks) == 0 {
			hwlog.RunLog.Debugf("no tasks found in namespace %s", ns)
			continue
		}
		for _, taskInfo := range taskList.Tasks {
			hwlog.RunLog.Debugf("Task ID: %s, PID: %d", taskInfo.ID, taskInfo.Pid)
			containerObj, err := tool.containerdClient.LoadContainer(ctx, taskInfo.ID)
			if err != nil {
				hwlog.RunLog.Warnf("failed to load container %s, err: %v", taskInfo.ID, err)
				continue
			}
			usedChipsWithAscendRuntime := tool.getDeviceWithAscendRuntime(containerObj, ctx)
			if usedChipsWithAscendRuntime.Len() > 0 {
				usedChips = usedChips.Union(usedChipsWithAscendRuntime)
				continue
			}
			usedChipsWithoutAscendRuntime := tool.getDeviceWithoutAscendRuntime(containerObj, ctx)
			usedChips = usedChips.Union(usedChipsWithoutAscendRuntime)
		}
	}
	hwlog.RunLog.Debugf("containerd used chips: %v", usedChips)
	return usedChips
}

func (tool *AscendTools) getDeviceWithAscendRuntime(containerObj containerd.Container,
	ctx context.Context) sets.String {
	usedChips := sets.NewString()
	containerInfo, err := containerObj.Info(ctx, containerd.WithoutRefreshedMetadata)
	if err != nil {
		hwlog.RunLog.Warnf("failed to get container info: %v", err)
		return usedChips
	}
	spec, err := getContainerValidSpec(containerObj, ctx)
	if err != nil {
		hwlog.RunLog.Warnf("failed to get container valid spec: %v", err)
		return usedChips
	}
	envs := spec.Process.Env
	for i := len(envs) - 1; i >= 0; i-- {
		devInfo := strings.Split(envs[i], "=")
		if len(devInfo) != ascendEnvPart {
			if len(devInfo) > 0 && devInfo[0] == common.AscendVisibleDevicesEnv {
				hwlog.RunLog.Warnf("an invalid %s env(%s)", common.AscendVisibleDevicesEnv, envs[i])
				return usedChips
			}
			hwlog.RunLog.Debugf("an invalid env(%s)", envs[i])
			continue
		}
		if devInfo[0] == common.AscendVisibleDevicesEnv {
			hwlog.RunLog.Debugf("get device info by env (%s) in %s", envs[i], containerInfo.ID)
			devicesIDs := parseDiffEnvFmt(devInfo[1], containerInfo.ID)
			hwlog.RunLog.Debugf("parse diffEnv get devicesIDs %v", devicesIDs)
			for _, deviceID := range devicesIDs {
				chipName := fmt.Sprintf("%s-%d", tool.name, deviceID)
				usedChips.Insert(chipName)
			}
			break
		}
	}
	return usedChips
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
