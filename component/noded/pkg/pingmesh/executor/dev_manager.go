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
	"sync"
	"time"

	"ascend-common/common-utils/hwlog"
	"ascend-common/devmanager"
	"ascend-common/devmanager/common"
	common2 "nodeD/pkg/common"
	"nodeD/pkg/pingmesh/types"
)

const (
	notFoundFunctionErrCode = "-99998"
	notSupportErrCode       = "-8255"
	collectPeriodFactor     = 10
	// pingMeshTaskStopped pingmesh task stopped
	pingMeshTaskStopped = 0
)

// DevManager execute action of hccsping mesh
type DevManager struct {
	devManager    devmanager.DeviceInterface
	commandChan   chan *types.HccspingMeshPolicy
	wg            *sync.WaitGroup
	currentPolicy *types.HccspingMeshPolicy
	chips         map[string]*common.ChipBaseInfo
	resultHandler func(result *types.HccspingMeshResult)
	SuperPodId    uint32
	RackId        uint32
	ServerIndex   uint32
}

// New create new device manager
func New() (*DevManager, error) {
	dm, err := devmanager.GetDeviceManager(common2.ParamOption.DeviceResetTimeout)
	if err != nil {
		return nil, err
	}

	chips, err := dm.GetChipBaseInfos()
	if err != nil {
		return nil, err
	}
	var superPodId uint32 = 0
	var serverIndex uint32 = 0
	var rackId uint32 = 0
	for _, chip := range chips {
		_, err = dm.DcGetHccsPingMeshState(chip.CardID, chip.DeviceID, 0, common.InternalPingMeshTaskID)
		if err != nil {
			hwlog.RunLog.Warnf("deviceManager get hccsPingMeshState failed, err: %v", err)
			if strings.Contains(err.Error(), notSupportErrCode) ||
				strings.Contains(err.Error(), notFoundFunctionErrCode) {
				return nil, err
			}
		}
		superPodInfo, err := dm.GetSuperPodInfo(chip.LogicID)
		if err != nil {
			return nil, fmt.Errorf("deviceManager get cgoSuperPodInfo failed, err: %v", err)
		}
		superPodId = superPodInfo.SuperPodId
		serverIndex = superPodInfo.ServerId
		hwlog.RunLog.Infof("new devManager get devType %s", dm.DevType)
		if dm.DevType == common.Ascend910A5 {
			rackId = superPodInfo.RackId
		}
		break
	}

	var physicID2ChipInfo = make(map[string]*common.ChipBaseInfo, len(chips))
	for _, chip := range chips {
		physicID2ChipInfo[strconv.Itoa(int(chip.PhysicID))] = chip
	}

	return &DevManager{
		devManager:  dm,
		chips:       physicID2ChipInfo,
		SuperPodId:  superPodId,
		RackId:      rackId,
		ServerIndex: serverIndex,
		wg:          &sync.WaitGroup{},
		commandChan: make(chan *types.HccspingMeshPolicy, 1),
	}, nil
}

// UpdateConfig update config
func (d *DevManager) UpdateConfig(config *types.HccspingMeshPolicy) {
	if d == nil {
		hwlog.RunLog.Error("deviceManager is nil")
		return
	}
	d.commandChan <- config
}

// SetResultHandler set result handler
func (d *DevManager) SetResultHandler(handler func(result *types.HccspingMeshResult)) {
	if d == nil {
		hwlog.RunLog.Error("deviceManager is nil")
		return
	}
	d.resultHandler = handler
}

// Start executor
func (d *DevManager) Start(stopCh <-chan struct{}) {
	if d == nil {
		hwlog.RunLog.Error("deviceManager is nil")
		return
	}
	var currentStop chan struct{} = nil

	for {
		select {
		case <-stopCh:
			// when main goroutine exit, children goroutine should exit
			if currentStop != nil {
				close(currentStop)
				d.wg.Wait()
			}
			return
		case cmd := <-d.commandChan:
			if cmd == nil || cmd.Config == nil {
				hwlog.RunLog.Warn("received nil hccspingmesh command, ignore")
				continue
			}
			hwlog.RunLog.Infof("executor receive cmd, activate: %s, uid: %s", cmd.Config.Activate, cmd.UID)
			// need stop collect goroutine and wait the goroutine done
			if currentStop != nil {
				close(currentStop)
				d.wg.Wait()
			}
			d.stopHccspingMesh()
			if cmd.Config.Activate == types.ActivateOff {
				currentStop = nil
				continue
			}
			d.currentPolicy = cmd
			d.startPingMesh()
			currentStop = make(chan struct{})
			d.wg.Add(1)
			go d.startCollect(currentStop)
		}
	}
}

func (d *DevManager) startPingMesh() {
	if d.devManager.GetDevType() == common.Ascend910A5 {
		d.startUbPingMesh()
	} else {
		d.startHccspingMesh()
	}
}

func (d *DevManager) startHccspingMesh() {
	for physicID, addrs := range d.currentPolicy.DestAddr {
		chip, ok := d.chips[physicID]
		if !ok || chip == nil {
			continue
		}

		for taskID := range addrs {

			hwlog.RunLog.Infof("execute starting hccspingmesh, cardID: %d, deviceID: %d, taskID: %d, "+
				"destination address: %v", chip.CardID, chip.DeviceID, taskID, addrs[taskID])
			if err := d.devManager.DcStartHccsPingMesh(chip.CardID, chip.DeviceID, 0, common.HccspingMeshOperate{
				DstAddr:      addrs[taskID],
				PktSize:      common.DefaultPktSize,
				PktSendNum:   common.DefaultPktSendNum,
				PktInterval:  common.DefaultPktInterval,
				Timeout:      common.DefaultTimeout,
				TaskInterval: d.currentPolicy.Config.TaskInterval,
				TaskId:       int(taskID),
			}); err != nil {
				hwlog.RunLog.Errorf("start hccspingmesh failed, err: %v", err)
			}
		}
	}
}

func (d *DevManager) stopHccspingMesh() {
	if d.currentPolicy == nil {
		d.stopAllTasks()
		return
	}
	d.stopLastTasks()
}

func (d *DevManager) stopAllTasks() {
	for _, chip := range d.chips {
		var taskIDs []uint
		if d.devManager.GetDevType() == common.Ascend910A5 {
			taskIDs = []uint{common.InternalPingMeshTaskID}
		} else {
			taskIDs = []uint{common.InternalPingMeshTaskID, common.ExternalPingMeshTaskID}
		}

		for _, taskID := range taskIDs {
			if err := d.devManager.DcStopHccsPingMesh(chip.CardID, chip.DeviceID, 0, taskID); err != nil {
				hwlog.RunLog.Errorf("stop hccspingmesh failed, err: %v", err)
				continue
			}

			hwlog.RunLog.Infof("stop hccspingmesh success, cardID: %d, deviceID: %d, taskID: %d",
				chip.CardID, chip.DeviceID, taskID)
		}
	}
}

func (d *DevManager) stopLastTasks() {
	for physicID, address := range d.currentPolicy.DestAddr {
		chip, ok := d.chips[physicID]
		if !ok || chip == nil {
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

func (d *DevManager) startCollect(stop <-chan struct{}) {
	hwlog.RunLog.Info("start collect hccsping mesh info")
	defer d.wg.Done()
	ticker := time.NewTicker(time.Duration(d.currentPolicy.Config.TaskInterval*collectPeriodFactor) * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-stop:
			hwlog.RunLog.Info("stop collect hccsping mesh info")
			return
		case <-ticker.C:
			d.getHccspingMeshInfo()
		}
	}
}

func (d *DevManager) getHccspingMeshInfo() {
	hwlog.RunLog.Infof("deviceManager get hccspingmesh info, time: %s", time.Now().Format(time.RFC3339))
	res := make(map[string]map[uint]*common.HccspingMeshInfo)
	for physicID, tasks := range d.currentPolicy.DestAddr {
		chip, ok := d.chips[physicID]
		if !ok || chip == nil {
			continue
		}
		infos := make(map[uint]*common.HccspingMeshInfo, len(tasks))
		for taskID := range tasks {
			d.checkPingMeshTaskState(chip.CardID, chip.DeviceID, taskID)
			hwlog.RunLog.Infof("get HccspingMeshInfo info, cardID: %d, deviceID: %d, physicID: %s, taskID: %d",
				chip.CardID, chip.DeviceID, physicID, taskID)
			info, err := d.devManager.DcGetHccsPingMeshInfo(chip.CardID, chip.DeviceID, 0, taskID) // 超时时间是30s
			if err != nil {
				hwlog.RunLog.Errorf("deviceManager get hccspingmesh info failed, err: %v", err)
				continue
			}
			if info == nil {
				hwlog.RunLog.Warn("deviceManager get hccspingmesh info is empty")
				continue
			}
			// when reset chip, pingmesh task will be stopped, so we should restart pingmesh task
			if d.GetDeviceType() == common.Ascend910A5 && len(info.UBPingMeshInfoList) == 0 {
				d.restartStoppedPingMeshTask(chip.CardID, chip.DeviceID, taskID, tasks[taskID])
				continue
			} else if d.GetDeviceType() != common.Ascend910A5 && info.DestNum == 0 {
				d.restartStoppedPingMeshTask(chip.CardID, chip.DeviceID, taskID, tasks[taskID])
				continue
			}
			infos[taskID] = info
			hwlog.RunLog.Infof("the infos len is %d", len(infos))
		}
		res[physicID] = infos
	}

	if d.resultHandler != nil {
		d.resultHandler(&types.HccspingMeshResult{
			Policy:  d.currentPolicy,
			Results: res,
		})
	}
}

func (d *DevManager) restartStoppedPingMeshTask(cardID, deviceID int32, taskID uint, addr string) {
	state, err := d.devManager.DcGetHccsPingMeshState(cardID, deviceID, 0, taskID)
	if err != nil {
		hwlog.RunLog.Errorf("deviceManager get hccspingmesh state failed, cardID: %d, "+
			"deviceID: %d, taskID: %d, err:%v", cardID, deviceID, taskID, err)
		return
	}
	if state != pingMeshTaskStopped {
		return
	}

	if d.devManager.GetDevType() == common.Ascend910A5 {
		d.restartUbPingMesh(cardID, deviceID)
	} else {
		d.restartHccsPingMesh(cardID, deviceID, taskID, addr)
	}
}

func (d *DevManager) restartHccsPingMesh(cardID, deviceID int32, taskID uint, addr string) {
	hwlog.RunLog.Infof("hccspingmesh task stopped, ready to restart, cardID: %d, "+
		"deviceID: %d, taskID: %d", cardID, deviceID, taskID)
	err := d.devManager.DcStartHccsPingMesh(cardID, deviceID, 0, common.HccspingMeshOperate{
		DstAddr:      addr,
		PktSize:      common.DefaultPktSize,
		PktSendNum:   common.DefaultPktSendNum,
		PktInterval:  common.DefaultPktInterval,
		Timeout:      common.DefaultTimeout,
		TaskInterval: d.currentPolicy.Config.TaskInterval,
		TaskId:       int(taskID),
	})
	if err != nil {
		hwlog.RunLog.Errorf("restart hccspingmesh failed, cardID: %d, deviceID: %d, taskID: %d err: %v",
			cardID, deviceID, taskID, err)
		return
	}
	hwlog.RunLog.Infof("restart hccspingmesh success, cardID: %d, deviceID: %d, taskID: %d",
		cardID, deviceID, taskID)
}

func (d *DevManager) checkPingMeshTaskState(cardID, deviceID int32, taskID uint) {
	state, err := d.devManager.DcGetHccsPingMeshState(cardID, deviceID, 0, taskID)
	if err != nil {
		hwlog.RunLog.Errorf("deviceManager get pingmesh state failed, cardID: %d, deviceID: %d, taskID: %d, err:%v",
			cardID, deviceID, taskID, err)
		return
	}
	hwlog.RunLog.Infof("get pingmesh state %d", state)
	if state != pingMeshTaskStopped {
		return
	}
	hwlog.RunLog.Infof("pingmesh task stopped, ready to restart, cardID: %d, deviceID: %d, taskID: %d",
		cardID, deviceID, taskID)
}

// GetDeviceType call devManager devType
func (d *DevManager) GetDeviceType() string {
	if d.devManager != nil {
		return d.devManager.GetDevType()
	}
	return ""
}
