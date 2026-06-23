/* Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.
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

package server

import (
	"context"
	"fmt"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"

	"Ascend-device-plugin/pkg/common"
	"Ascend-device-plugin/pkg/device"
	"Ascend-device-plugin/pkg/kubeclient"
	"Ascend-device-plugin/pkg/plugin"
	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"ascend-common/devmanager"
	npuCommon "ascend-common/devmanager/common"
)

const (
	idleWaitSeconds        = 60
	bootStatusPollInterval = 1 * time.Second
	bootStatusPollTimeout  = 350 * time.Second
)

type UnifiedHotResetManager struct {
	dmgr        devmanager.DeviceInterface
	devManager  device.DevManager
	tokenBucket *TokenBucketMgr
	idleTimeMgr *IdleTimeMgr
	pluginMgr   *plugin.PluginManager
	mu          sync.Mutex
}

func NewUnifiedHotResetManager(dmgr devmanager.DeviceInterface,
	devManager device.DevManager, kubeClient *kubeclient.ClientK8s) *UnifiedHotResetManager {

	return &UnifiedHotResetManager{
		dmgr:        dmgr,
		devManager:  devManager,
		tokenBucket: NewTokenBucketMgr(),
		idleTimeMgr: NewIdleTimeMgr(),
	}
}

func (m *UnifiedHotResetManager) SetPluginManager(pm *plugin.PluginManager) {
	m.pluginMgr = pm
}

func (m *UnifiedHotResetManager) UnifiedHotReset(groupDevice map[string][]*common.NpuDevice) {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.checkDeviceRecovered(groupDevice)
	if common.ParamOption.HotReset == common.HotResetClose {
		hwlog.RunLog.Debug("hot reset is closed, return")
		return
	}
	faultDevs := m.filterFaultDevices(groupDevice)
	if len(faultDevs) == 0 {
		return
	}
	prClient := NewPodResource()
	for _, dev := range faultDevs {
		hwlog.RunLog.Infof("UnifiedHotReset: faultDev: %v", dev.LogicID)
		if m.devManager.GetIfCardsInResetting(dev.LogicID) {
			hwlog.RunLog.Infof("device %d is in resetting, skip", dev.LogicID)
			continue
		}
		ringDevs, faultDev := m.getResetRingDevices(dev, groupDevice)
		if len(ringDevs) == 0 {
			hwlog.RunLog.Infof("getResetRingDevices: no ring devices for device %d", dev.LogicID)
			continue
		}
		if !m.isRingFree(ringDevs, prClient, groupDevice) {
			continue
		}
		m.idleTimeMgr.RecordIdleTime(dev.LogicID)
		if !m.idleTimeMgr.IsIdleTimeExceeded(dev.LogicID, idleWaitSeconds) {
			hwlog.RunLog.Debugf("idle time not exceeded for device %d, skip", dev.LogicID)
			continue
		}
		if !m.tokenBucket.HasToken(dev.LogicID) {
			hwlog.RunLog.Warnf("device %d token exhausted, need external ops", dev.LogicID)
			m.markNeedExternalOps(dev)
			continue
		}
		if !m.confirmAndPrepareReset(ringDevs, faultDev) {
			continue
		}
		go m.executeReset(ringDevs, faultDev)
	}
}

func (m *UnifiedHotResetManager) checkDeviceRecovered(groupDevice map[string][]*common.NpuDevice) {
	var recoveredDevs []*common.NpuDevice
	for _, devices := range groupDevice {
		for _, dev := range devices {
			if dev.Health == v1beta1.Healthy {
				hwlog.RunLog.Debugf("checkDeviceRecovered: device %d recovered to healthy", dev.LogicID)
				m.idleTimeMgr.DeleteIdleTime(dev.LogicID)
				recoveredDevs = append(recoveredDevs, dev)
			}
		}
	}
	if len(recoveredDevs) > 0 {
		m.clearNeedExternalOps(recoveredDevs)
	}
}

func (m *UnifiedHotResetManager) filterFaultDevices(groupDevice map[string][]*common.NpuDevice) []*common.NpuDevice {
	var faultDevs []*common.NpuDevice
	for devType, devices := range groupDevice {
		if common.IsVirtualDev(devType) || len(devices) == 0 {
			continue
		}
		for _, dev := range devices {
			if dev.Health != v1beta1.Healthy && !m.devManager.GetIfCardsInResetting(dev.LogicID) {
				faultType := common.GetFaultType(dev.FaultCodes, dev.LogicID)
				if faultType == common.NotHandleFault || faultType == common.PreSeparateNPU {
					hwlog.RunLog.Debugf("filterFaultDevices: device %d fault type %s, skip", dev.LogicID, faultType)
					continue
				}
				if faultType == common.SeparateNPU || faultType == common.ManuallySeparateNPU {
					hwlog.RunLog.Warnf("device %d is L6 isolated", dev.LogicID)
					m.markNeedExternalOps(dev)
					continue
				}
				if faultType == common.NormalNPU || faultType == common.SubHealthFault {
					hwlog.RunLog.Debugf("filterFaultDevices: device %d fault type %s, continue", dev.LogicID, faultType)
					continue
				}
				faultDevs = append(faultDevs, dev)
			}
		}
	}
	hwlog.RunLog.Debugf("filterFaultDevices: faultDevs: %v", faultDevs)
	return faultDevs
}

func (m *UnifiedHotResetManager) getResetRingDevices(faultDev *common.NpuDevice,
	groupDevice map[string][]*common.NpuDevice) ([]*common.NpuDevice, *common.NpuDevice) {

	boardId, err := m.devManager.GetServerBoardId(common.FirstDevice)
	if err != nil {
		hwlog.RunLog.Warnf("get board id failed: %v, use default 0", err)
		boardId = 0
	}
	devType := m.findFaultDevType(faultDev, groupDevice)
	deviceNum := len(groupDevice[devType])

	if common.ParamOption.RealCardType == api.Ascend910A3 {
		return m.getA3AssociatedDevices(faultDev, groupDevice)
	}

	if common.IsContainAtlas300IDuo() {
		return m.getDuoCardDevices(faultDev, groupDevice)
	}

	ringSize := m.getRingSize(faultDev, boardId, deviceNum)
	if ringSize == 1 {
		return []*common.NpuDevice{faultDev}, faultDev
	}

	return m.getHccsRingDevices(faultDev, ringSize, groupDevice)
}

func (m *UnifiedHotResetManager) getRingSize(dev *common.NpuDevice, boardId uint32, deviceNum int) int {
	switch common.ParamOption.RealCardType {
	case api.Ascend910A:
		return common.Ascend910RingsNum
	case api.Ascend910B:
		if boardId == common.A300IA2BoardId || boardId == common.A300IA2GB64BoardId ||
			boardId == common.A800IA2NoneHccsBoardId || boardId == common.A800IA2NoneHccsBoardIdOld {
			return common.Ascend910BRingsNumInfer
		}
		if deviceNum > common.Ascend910BRingsNumTrain {
			return common.A200TA2RingsNum
		}
		return common.Ascend910BRingsNumTrain
	case api.Ascend910A3:
		return common.Ascend910A3RingsNum
	case api.Ascend910A5:
		return common.Ascend910A5RingsNum
	default:
		return 1
	}
}

func (m *UnifiedHotResetManager) findFaultDevType(faultDev *common.NpuDevice,
	groupDevice map[string][]*common.NpuDevice) string {
	for dt, devices := range groupDevice {
		for _, dev := range devices {
			if dev.LogicID == faultDev.LogicID {
				return dt
			}
		}
	}
	return ""
}

func (m *UnifiedHotResetManager) getA3AssociatedDevices(faultDev *common.NpuDevice,
	groupDevice map[string][]*common.NpuDevice) ([]*common.NpuDevice, *common.NpuDevice) {

	cardID, deviceID, err := m.dmgr.GetCardIDDeviceID(faultDev.LogicID)
	if err != nil {
		hwlog.RunLog.Errorf("get card id and device id failed, logicID: %d, err: %v", faultDev.LogicID, err)
		return nil, nil
	}
	logicIDs, err := m.devManager.GetAssociatedLogicIDs(faultDev.LogicID, cardID, deviceID)
	if err != nil || len(logicIDs) == 0 {
		hwlog.RunLog.Errorf("get associated logic ids failed, logicID: %d, err: %v", faultDev.LogicID, err)
		return nil, nil
	}

	idSet := sets.NewInt32(logicIDs...)
	var ringDevs []*common.NpuDevice
	for _, devices := range groupDevice {
		for _, dev := range devices {
			if idSet.Has(dev.LogicID) {
				ringDevs = append(ringDevs, dev)
			}
		}
	}
	return ringDevs, faultDev
}

func (m *UnifiedHotResetManager) getDuoCardDevices(faultDev *common.NpuDevice,
	groupDevice map[string][]*common.NpuDevice) ([]*common.NpuDevice, *common.NpuDevice) {

	var ringDevs []*common.NpuDevice
	for _, devices := range groupDevice {
		for _, dev := range devices {
			if dev.CardID == faultDev.CardID {
				ringDevs = append(ringDevs, dev)
			}
		}
	}
	return ringDevs, faultDev
}

func (m *UnifiedHotResetManager) getHccsRingDevices(faultDev *common.NpuDevice,
	ringSize int, groupDevice map[string][]*common.NpuDevice) ([]*common.NpuDevice, *common.NpuDevice) {

	ringStart := (int(faultDev.LogicID) / ringSize) * ringSize
	var ringDevs []*common.NpuDevice
	for _, devices := range groupDevice {
		for _, dev := range devices {
			if int(dev.LogicID) >= ringStart && int(dev.LogicID) < ringStart+ringSize {
				ringDevs = append(ringDevs, dev)
			}
		}
	}
	return ringDevs, faultDev
}

func (m *UnifiedHotResetManager) isRingFree(ringDevs []*common.NpuDevice,
	prClient *PodResource, groupDevice map[string][]*common.NpuDevice) bool {

	for _, dev := range ringDevs {
		devType := dev.DevType
		if devType == "" {
			for dt, devices := range groupDevice {
				for _, d := range devices {
					if d.LogicID == dev.LogicID {
						devType = dt
						break
					}
				}
			}
		}
		if dev.PodUsed {
			hwlog.RunLog.Infof("device %d has pod, skip reset", dev.LogicID)
			return false
		}
		processInfo, err := m.dmgr.GetDevProcessInfo(dev.LogicID)
		if err != nil || processInfo == nil {
			hwlog.RunLog.Errorf("device %d has process, skip reset, err: %v", dev.LogicID, err)
			return false
		}
		if processInfo.ProcNum != 0 {
			hwlog.RunLog.Infof("isRingFree: device %d process num is %d, skip reset", dev.LogicID, processInfo.ProcNum)
			return false
		}
		hwlog.RunLog.Infof("isRingFree: device %d process num is 0, continue reset", dev.LogicID)
	}
	hwlog.RunLog.Infof("isRingFree: all devices in ring are free, continue reset, len of ringDevs: %d", len(ringDevs))
	return true
}

func (m *UnifiedHotResetManager) confirmAndPrepareReset(ringDevs []*common.NpuDevice,
	faultDev *common.NpuDevice) bool {
	for _, dev := range ringDevs {
		if m.devManager.GetIfCardsInResetting(dev.LogicID) {
			hwlog.RunLog.Infof("confirmAndPrepareReset: ring device %d is in resetting, abort", dev.LogicID)
			return false
		}
	}
	healthCode, err := m.dmgr.GetDeviceHealth(faultDev.LogicID)
	if err != nil && !common.CheckErrorMessage(err, npuCommon.DeviceNotReadyErrCodeStr) {
		hwlog.RunLog.Warnf("device %d is not ready, err: %v", faultDev.LogicID, err)
		return false
	}
	if err == nil && healthCode == 0 {
		hwlog.RunLog.Infof("confirmAndPrepareReset: device %d is healthy, skip reset", faultDev.LogicID)
		return false
	}
	hwlog.RunLog.Infof("confirmAndPrepareReset: consuming token and setting resetting for device %d", faultDev.LogicID)
	m.tokenBucket.ConsumeToken(faultDev.LogicID)
	for _, dev := range ringDevs {
		m.devManager.SetCardsInResetting(dev.LogicID, true)
	}
	return true
}

func (m *UnifiedHotResetManager) executeReset(ringDevs []*common.NpuDevice, faultDev *common.NpuDevice) {
	resetDevices := m.convertToResetDevices(ringDevs)
	tokensLeft := int32(m.tokenBucket.GetTokens(faultDev.LogicID))
	for i := range resetDevices {
		if resetDevices[i].LogicID == faultDev.LogicID {
			resetDevices[i].IsFaultDev = true
			resetDevices[i].TokensLeft = tokensLeft
			break
		}
	}
	ctx := context.Background()

	hwlog.RunLog.Infof("executing PreReset for device %d", faultDev.LogicID)
	if m.pluginMgr != nil {
		m.pluginMgr.ExecutePreReset(ctx, resetDevices)
	}

	hwlog.RunLog.Infof("executing driver reset for device %d", faultDev.LogicID)
	var resetErr error
	if err := m.execDriverReset(faultDev.LogicID, ringDevs); err != nil {
		resetErr = fmt.Errorf("driver reset failed: %w", err)
		hwlog.RunLog.Warnf("driver reset failed for logicID %d: %v", faultDev.LogicID, err)
	}

	hwlog.RunLog.Infof("executing CustomReset for device %d", faultDev.LogicID)
	if m.pluginMgr != nil {
		resetErr = m.pluginMgr.ExecuteCustomReset(ctx, resetDevices, resetErr)
	}

	if resetErr == nil {
		m.idleTimeMgr.DeleteIdleTime(faultDev.LogicID)
		common.SetDeviceInit(faultDev.LogicID)
		hwlog.RunLog.Infof("hot reset success, logicID: %d", faultDev.LogicID)
	} else {
		hwlog.RunLog.Warnf("hot reset failed, logicID: %d, err: %v", faultDev.LogicID, resetErr)
	}

	hwlog.RunLog.Infof("executing AfterReset for device %d", faultDev.LogicID)
	if m.pluginMgr != nil {
		m.pluginMgr.ExecuteAfterReset(ctx, resetDevices, resetErr)
	}

	hwlog.RunLog.Infof("resetting device %d reset state", faultDev.LogicID)
	for _, dev := range ringDevs {
		m.devManager.SetCardsInResetting(dev.LogicID, false)
	}
}

func (m *UnifiedHotResetManager) execDriverReset(logicID int32, ringDevs []*common.NpuDevice) error {
	var isResetExec bool
	if err := wait.PollImmediate(bootStatusPollInterval, bootStatusPollTimeout, func() (bool, error) {
		if !isResetExec {
			if err := m.dmgr.SetDeviceReset(logicID); err != nil {
				return false, err
			}
			isResetExec = true
		}
		for _, dev := range ringDevs {
			bootState, err := m.dmgr.GetDeviceBootStatus(dev.LogicID)
			if err != nil {
				return false, err
			}
			if bootState != common.BootStartFinish {
				return false, nil
			}
		}
		return true, nil
	}); err != nil {
		return err
	}
	return nil
}

func (m *UnifiedHotResetManager) convertToResetDevices(devs []*common.NpuDevice) []plugin.ResetDevice {
	result := make([]plugin.ResetDevice, 0, len(devs))
	for _, dev := range devs {
		result = append(result, plugin.ResetDevice{
			LogicID:  dev.LogicID,
			CardID:   dev.CardID,
			DeviceID: dev.DeviceID,
			PhyID:    dev.PhyID,
			CardType: common.ParamOption.RealCardType,
		})
	}
	return result
}

func (m *UnifiedHotResetManager) markNeedExternalOps(dev *common.NpuDevice) {
	hwlog.RunLog.Warnf("markNeedExternalOps: device %d, marking node annotation", dev.LogicID)
	resetInfo := device.ResetInfo{
		ManualResetDevs: []device.ResetDevice{
			{LogicID: dev.LogicID, CardId: dev.CardID, DeviceId: dev.DeviceID, PhyID: dev.PhyID},
		},
	}
	device.WriteResetInfo(resetInfo, device.WMAppend, true)
}

func (m *UnifiedHotResetManager) clearNeedExternalOps(devs []*common.NpuDevice) {
	var logicIDs []int32
	for _, dev := range devs {
		logicIDs = append(logicIDs, dev.LogicID)
	}
	resetInfo := device.ReadResetInfo()
	needClear := filterOpsDevices(resetInfo, logicIDs)
	if len(needClear) == 0 {
		return
	}
	hwlog.RunLog.Infof("step0: clearing ops annotation for %d devices: %v", len(needClear), needClear)
	needClearSet := make(map[int32]struct{}, len(needClear))
	for _, id := range needClear {
		needClearSet[id] = struct{}{}
	}
	var delDevs []device.ResetDevice
	for _, dev := range devs {
		if _, ok := needClearSet[dev.LogicID]; !ok {
			continue
		}
		delDevs = append(delDevs, device.ResetDevice{
			LogicID: dev.LogicID, CardId: dev.CardID, DeviceId: dev.DeviceID, PhyID: dev.PhyID,
		})
	}
	if len(delDevs) == 0 {
		return
	}
	delInfo := device.ResetInfo{
		ThirdPartyResetDevs: delDevs,
		ManualResetDevs:     delDevs,
	}
	device.WriteResetInfo(delInfo, device.WMDelete, true)
}

func filterOpsDevices(resetInfo device.ResetInfo, logicIDs []int32) []int32 {
	idSet := make(map[int32]bool)
	for _, d := range resetInfo.ThirdPartyResetDevs {
		idSet[d.LogicID] = true
	}
	for _, d := range resetInfo.ManualResetDevs {
		idSet[d.LogicID] = true
	}
	var needClear []int32
	for _, id := range logicIDs {
		if idSet[id] {
			needClear = append(needClear, id)
		}
	}
	return needClear
}
