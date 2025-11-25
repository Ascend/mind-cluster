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

// Package app control container
package app

import (
	"context"
	"fmt"
	"syscall"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/namespaces"

	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
	"container-manager/pkg/common"
)

func (cm *CtrCtl) pauseCtr(onRing bool) {
	ctrNeedPaused := cm.devInfoMap.GetNeedPausedCtr(onRing)
	ctrHasPaused := cm.ctrInfoMap.GetCtrsByStatus(common.StatusPaused)
	needPaused := utils.RemoveEleSli(ctrNeedPaused, ctrHasPaused)
	for _, id := range needPaused {
		if err := cm.doPauseCtr(id); err != nil {
			hwlog.RunLog.Errorf("pause container %s failed, error: %v", id, err)
			continue
		}
	}
}

func (cm *CtrCtl) resumeCtr(onRing bool) {
	ctrHasPaused := cm.ctrInfoMap.GetCtrsByStatus(common.StatusPaused)
	var ctrNeedResume []string
	for _, id := range ctrHasPaused {
		if !onRing {
			if cm.isDevsNeedPause(cm.ctrInfoMap.GetCtrUsedDevs(id)) {
				continue
			}
			ctrNeedResume = append(ctrNeedResume, id)
			continue
		}
		if utils.Contains(ctrNeedResume, id) {
			continue
		}
		ctrsOnRings := cm.ctrInfoMap.GetCtrsOnRing(id)
		// can all containers on the ring be resumed.
		// as long as one of the cards used by the containers on the ring does not meet the condition,
		// the entire container on the ring cannot be resumed
		if cm.isDevsNeedPause(cm.ctrInfoMap.GetCtrRelatedDevs(ctrsOnRings)) {
			continue
		}
		for _, ctrId := range ctrsOnRings {
			if utils.Contains(ctrHasPaused, ctrId) {
				ctrNeedResume = append(ctrNeedResume, ctrId)
			}
		}
	}

	for _, id := range ctrNeedResume {
		if err := cm.doResumeCtr(id); err != nil {
			hwlog.RunLog.Errorf("resume container %s failed, error: %v", id, err)
			continue
		}
	}
}

func (cm *CtrCtl) doPauseCtr(containerID string) error {
	hwlog.RunLog.Infof("start pausing container: %s", containerID)
	cm.ctrInfoMap.SetCtrsStatus(containerID, common.StatusPausing)
	ns := cm.ctrInfoMap.GetCtrNs(containerID)
	if ns == "" {
		return fmt.Errorf("failed to get namespace of container: %s", containerID)
	}
	ctx := namespaces.WithNamespace(context.Background(), ns)
	container, err := cm.client.LoadContainer(ctx, containerID)
	if err != nil {
		return fmt.Errorf("failed to load container: %v", err)
	}
	task, err := container.Task(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to get container %s , error: %v", containerID, err)
	}
	if err = task.Kill(ctx, syscall.SIGKILL); err != nil {
		return fmt.Errorf("failed to kill container %s, error: %v", containerID, err)
	}
	// force kill again to make sure the container is killed
	if err = task.Kill(ctx, syscall.SIGKILL, containerd.WithKillAll); err != nil {
		return fmt.Errorf("failed to kill container %s, error: %v", containerID, err)
	}
	hwlog.RunLog.Infof("successfully pause container: %s", containerID)
	cm.ctrInfoMap.SetDetailedInfo(containerID, container)
	cm.ctrInfoMap.SetCtrsStatus(containerID, common.StatusPaused)
	return nil
}

func (cm *CtrCtl) doResumeCtr(containerID string) error {
	hwlog.RunLog.Infof("start resuming container: %s", containerID)
	cm.ctrInfoMap.SetCtrsStatus(containerID, common.StatusResuming)
	ns := cm.ctrInfoMap.GetCtrNs(containerID)
	if ns == "" {
		return fmt.Errorf("failed to get namespace of container: %s", containerID)
	}
	ctx := namespaces.WithNamespace(context.Background(), ns)
	container := cm.ctrInfoMap.GetDetailedInfo(containerID)
	if container == nil {
		return fmt.Errorf("failed to get detailed info of container: %s", containerID)
	}
	task, err := container.NewTask(ctx, cio.NewCreator(cio.WithStdio))
	if err != nil {
		return fmt.Errorf("failed to new task for container: %v", err)
	}
	if err = task.Start(ctx); err != nil {
		return fmt.Errorf("failed to start task for container: %v", err)
	}
	hwlog.RunLog.Infof("successfully resume container: %s", containerID)
	cm.ctrInfoMap.SetCtrsStatus(containerID, common.StatusRunning)
	return nil
}
