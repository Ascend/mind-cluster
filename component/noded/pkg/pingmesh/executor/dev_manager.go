/*
Copyright(C)2025. Huawei Technologies Co.,Ltd. All rights reserved.

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

/*
Package executor is using for execute hccsping mesh
*/

package executor

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"

	"ascend-common/common-utils/hwlog"
	"ascend-common/devmanager"
	"ascend-common/devmanager/common"
	"nodeD/pkg/pingmesh/types"
)

const (
	notSupportErrCode   = "-99996"
	collectPeriodFactor = 5
)

// Executor execute action of hccsping mesh
type Executor struct {
	devManager    devmanager.DeviceInterface
	stopChan      chan struct{}
	commandChan   chan *types.HccspingMeshPolicy
	currentPolicy *types.HccspingMeshPolicy
	chips         map[string]*common.ChipBaseInfo
	resultHandler func(result *types.HccspingMeshResult)
	SuperPodId    uint32
}

// New create new device manager
func New() (*Executor, error) {
	dm, err := devmanager.GetDeviceManager()
	if err != nil {
		return nil, err
	}

	chips, err := dm.GetChipBaseInfos()
	if err != nil {
		return nil, err
	}

	var superPodId uint32
	for _, chip := range chips {
		_, err = dm.DcGetHccsPingMeshState(chip.CardID, chip.DeviceID, 0, common.InternalPingMeshTaskID)
		if err != nil && strings.Contains(err.Error(), notSupportErrCode) {
			return nil, err
		}
		superPodInfo, err := dm.GetSuperPodInfo(chip.LogicID)
		if err != nil {
			return nil, fmt.Errorf("deviceManager get cgoSuperPodInfo failed, err: %v", err)
		}
		superPodId = superPodInfo.SuperPodId
		break
	}

	var physicID2ChipInfo = make(map[string]*common.ChipBaseInfo, len(chips))
	for _, chip := range chips {
		physicID2ChipInfo[strconv.Itoa(int(chip.PhysicID))] = chip
	}

	return &Executor{
		devManager:  dm,
		chips:       physicID2ChipInfo,
		SuperPodId:  superPodId,
		commandChan: make(chan *types.HccspingMeshPolicy, 1),
	}, nil
}

// UpdateConfig update config
func (d *Executor) UpdateConfig(config *types.HccspingMeshPolicy) {
	d.commandChan <- config
}

// SetResultHandler set result handler
func (d *Executor) SetResultHandler(handler func(result *types.HccspingMeshResult)) {
	d.resultHandler = handler
}

// Start executor
func (d *Executor) Start(stopCh <-chan struct{}) {
	for {
		select {
		case <-stopCh:
			// todo consider to stop hccsping-mesh
			d.stopCollect()
			return
		case cmd := <-d.commandChan:
			hwlog.RunLog.Infof("executor receive cmd, start: %v, uid: %s", cmd.Config.Activate, cmd.UID)
			d.stopHccspingMesh()
			d.stopCollect()
			if cmd.Config.Activate == types.ActivateOff {
				continue
			}
			d.currentPolicy = cmd
			d.startHccspingMesh()
			go d.startCollect()
		}
	}
}

func (d *Executor) startHccspingMesh() {
	for physicID, addrs := range d.currentPolicy.DestAddr {
		chip, ok := d.chips[physicID]
		if !ok {
			continue
		}

		for taskID := range addrs {
			err := d.devManager.DcStopHccsPingMesh(chip.CardID, chip.DeviceID, 0, taskID)
			if err != nil {
				hwlog.RunLog.Errorf("deviceManager stop hccspingmesh failed, err: %v", err)
				continue
			}
			hwlog.RunLog.Debugf("deviceManager start hccspingmesh, cardID: %d, deviceID: %d, taskID: %d",
				chip.CardID, chip.DeviceID, taskID)
			if err := d.devManager.DcStartHccsPingMesh(chip.CardID, chip.DeviceID, 0, common.HccspingMeshOperate{
				DstAddr:      addrs[taskID],
				PktSize:      common.DefaultPktSize,
				PktSendNum:   common.DefaultPktSendNum,
				PktInterval:  common.DefaultPktInterval,
				Timeout:      common.DefaultTimeout,
				TaskInterval: d.currentPolicy.Config.TaskInterval,
				TaskId:       int(taskID),
			}); err != nil {
				hwlog.RunLog.Errorf("deviceManager start hccspingmesh failed, err: %v", err)
			}
		}
	}
}

func (d *Executor) stopHccspingMesh() {
	if d.currentPolicy == nil {
		d.stopAllTasks()
		return
	}
	d.stopLastTasks()
}

func (d *Executor) stopAllTasks() {
	for _, chip := range d.chips {
		for _, taskID := range []uint{common.InternalPingMeshTaskID, common.ExternalPingMeshTaskID} {
			if err := d.devManager.DcStopHccsPingMesh(chip.CardID, chip.DeviceID, 0, taskID); err != nil {
				hwlog.RunLog.Errorf("deviceManager stop hccspingmesh failed, err: %v", err)
				continue
			}

			hwlog.RunLog.Infof("deviceManager stop hccspingmesh success, cardID: %d, deviceID: %d, taskID: %d",
				chip.CardID, chip.DeviceID, taskID)
		}
	}
}

func (d *Executor) stopLastTasks() {
	for physicID, address := range d.currentPolicy.DestAddr {
		chip, ok := d.chips[physicID]
		if !ok {
			continue
		}
		for taskID := range address {
			if err := d.devManager.DcStopHccsPingMesh(chip.CardID, chip.DeviceID, 0, taskID); err != nil {
				hwlog.RunLog.Errorf("deviceManager stop hccspingmesh failed, err: %v", err)
				continue
			}
			hwlog.RunLog.Infof("deviceManager stop hccspingmesh success, cardID: %d, deviceID: %d, taskID: %d",
				chip.CardID, chip.DeviceID, taskID)
		}
	}
}

func (d *Executor) startCollect() {
	if d.stopChan == nil {
		d.stopChan = make(chan struct{}, 1)
	}

	wait.Until(
		d.getHccspingMeshInfo,
		time.Duration(d.currentPolicy.Config.TaskInterval*collectPeriodFactor)*time.Second,
		d.stopChan)
}

func (d *Executor) stopCollect() {
	if d.stopChan != nil {
		d.stopChan <- struct{}{}
	}
}

func (d *Executor) getHccspingMeshInfo() {
	hwlog.RunLog.Infof("deviceManager get hccspingmesh info, time: %s", time.Now().Format(time.RFC3339))
	res := make(map[string]map[uint]*common.HccspingMeshInfo)
	for physicID, tasks := range d.currentPolicy.DestAddr {
		chip, ok := d.chips[physicID]
		if !ok {
			continue
		}
		infos := make(map[uint]*common.HccspingMeshInfo, len(tasks))
		for taskID := range tasks {
			info, err := d.devManager.DcGetHccsPingMeshInfo(chip.CardID, chip.DeviceID, 0, taskID) // 超时时间是30s
			if err != nil {
				hwlog.RunLog.Errorf("deviceManager get hccspingmesh info failed, err: %v", err)
				continue
			}
			infos[taskID] = info
		}
		res[physicID] = infos
	}

	if d.resultHandler != nil {
		d.resultHandler(&types.HccspingMeshResult{
			UID:     d.currentPolicy.UID,
			Address: d.currentPolicy.DestAddr,
			Config:  d.currentPolicy.Config,
			Results: res,
		})
	}
}
