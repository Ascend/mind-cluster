/* Copyright(C) 2022. Huawei Technologies Co.,Ltd. All rights reserved.
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
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"

	"Ascend-device-plugin/pkg/common"
	"Ascend-device-plugin/pkg/device"
	"Ascend-device-plugin/pkg/device/deviceswitch"
	"Ascend-device-plugin/pkg/device/dpucontrol"
	"Ascend-device-plugin/pkg/kubeclient"
	"Ascend-device-plugin/pkg/next/devicefactory/customname"
	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"ascend-common/devmanager"
	npuCommon "ascend-common/devmanager/common"
)

var resourceVersion = ""

const (
	memoryRadix                  = 1024
	nodeAnnotationUpdateInterval = 60
	serverIndexKey               = "serverIndex"
	serverTypeKey                = "serverType"
	cardTypeKey                  = "cardType"
)

// HwDevManager manages huawei device devices.
type HwDevManager struct {
	SwitchDevManager *deviceswitch.SwitchDevManager
	groupDevice      map[string][]*common.NpuDevice
	ServerMap        map[string]InterfaceServer
	allInfo          common.NpuAllInfo
	manager          device.DevManager
	RunMode          string
	WorkMode         string
	baseNPUInfo      map[string]*common.NpuBaseInfo
	dpuManager       *dpucontrol.DpuFilter
	ManagerLock      sync.Mutex
}

// NewHwDevManager function is used to new a dev manager.
func NewHwDevManager(devM devmanager.DeviceInterface) *HwDevManager {
	var hdm HwDevManager
	hdm.dpuManager = &dpucontrol.DpuFilter{}
	if err := hdm.setAscendManager(devM); err != nil {
		hwlog.RunLog.Errorf("init hw dev manager failed, err: %v", err)
		return nil
	}
	if err := hdm.setAllDeviceAndType(); err != nil {
		hwlog.RunLog.Errorf("set all device and type failed, err: %v", err)
		return nil
	}
	device.InitResetInfoMgr(hdm.manager.GetKubeClient())
	if err := hdm.checkSupportedProductType(); err != nil {
		hwlog.RunLog.Errorf("check supported product type failed, err: %v", err)
		return nil
	}
	hdm.setSuperPodInfo()
	if err := hdm.UpdateNode(); err != nil {
		hwlog.RunLog.Errorf("update node label failed, err: %v", err)
		return nil
	}
	if err := hdm.initPluginServer(); err != nil {
		hwlog.RunLog.Errorf("init plugin server failed, err: %v", err)
		return nil
	}
	return &hdm
}

func (hdm *HwDevManager) setAscendManager(dmgr devmanager.DeviceInterface) error {
	devType := dmgr.GetDevType()
	if !common.ParamOption.PresetVDevice && devType != api.Ascend310P && devType != api.Ascend910B {
		return fmt.Errorf("only 310p and 910b support to set presetVirtualDevice false")
	}
	switch devType {
	case api.Ascend310, api.Ascend310B:
		hdm.RunMode = api.Ascend310
		hdm.manager = device.NewHwAscend310Manager()
	case api.Ascend910A, api.Ascend910B, api.Ascend910A3, api.Ascend910A5:
		hdm.RunMode = api.Ascend910
		hdm.manager = device.NewHwAscend910Manager()
		hdm.WorkMode = dmgr.GetNpuWorkMode()
	case api.Ascend310P:
		hdm.RunMode = api.Ascend310P
		hdm.manager = device.NewHwAscend310PManager()
	default:
		hwlog.RunLog.Error("found an unsupported device type")
		return fmt.Errorf("an unsupported device type")
	}
	common.ParamOption.RealCardType = devType
	hdm.manager.SetDmgr(dmgr)
	productTypes, err := hdm.manager.GetDmgr().GetAllProductType()
	if err != nil {
		return err
	}
	common.ParamOption.ProductTypes = productTypes
	if err = common.CheckCardUsageMode(common.ParamOption.Use310PMixedInsert, productTypes); err != nil {
		return err
	}

	if common.ParamOption.BuildScene != common.EdgeScene {
		aiCoreCount, err := hdm.manager.GetChipAiCoreCount()
		if err != nil {
			hwlog.RunLog.Errorf("get chip aicore count failed, err: %v", err)
			return err
		}
		common.ParamOption.AiCoreCount = aiCoreCount
	}
	return nil
}

// UpdateNode update server type, like Ascend910-32, and label of 910b infer card
// other common label will be updated in the future
func (hdm *HwDevManager) UpdateNode() error {
	if common.ParamOption.BuildScene == common.EdgeScene {
		return nil
	}
	hdm.manager.GetKubeClient().InitPodInformer()
	hwlog.RunLog.Info("init kube client success")

	return hdm.updateNode()
}

func getDevType(cardType string) string {
	if strings.Contains(cardType, common.DevA3) {
		return common.DevA3
	}
	if strings.Contains(cardType, common.DevA5) {
		return common.DevA5
	}
	return ""
}

func (hdm *HwDevManager) updateNode() error {
	oldNode, err := hdm.manager.GetKubeClient().GetNode()
	if err != nil || oldNode == nil {
		hwlog.RunLog.Errorf("failed to get node, err: %v, node is nil: %v", err, oldNode == nil)
		return err
	}
	// rank table needs this info for A5
	hdm.SetNodeInternalIPInK8s(oldNode)
	newLabelMap, err := hdm.getNewNodeLabel(oldNode)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get new node label, err: %v", err)
		return err
	}
	if len(newLabelMap) == 0 {
		return nil
	}
	newNode := oldNode.DeepCopy()
	for key, value := range newLabelMap {
		newNode.Labels[key] = value
	}

	newAnnotationMap, err := hdm.getNewNodeAnnotation(oldNode)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get new node annotation, err: %v", err)
		return err
	}
	for key, value := range newAnnotationMap {
		newNode.Annotations[key] = value
	}

	for i := 0; i < common.RetryUpdateCount; i++ {
		if _, _, err = hdm.manager.GetKubeClient().PatchNodeState(oldNode, newNode); err == nil {
			hwlog.RunLog.Info("update node label success")
			return nil
		}
		hwlog.RunLog.Warnf("failed to patch new label to node, err: %s, retry count: %d", err.Error(), i+1)
		time.Sleep(time.Second)
	}
	return fmt.Errorf("update node label failed")
}

func (hdm *HwDevManager) getNewNodeAnnotation(oldNode *v1.Node) (map[string]string, error) {
	annotationMap := make(map[string]string)
	cardType, err := hdm.getCardType()
	if err != nil {
		hwlog.RunLog.Errorf("failed to get node board info, err: %v", err)
	}
	if cardType != "" {
		annotationMap[cardTypeKey] = cardType
		common.ParamOption.CardType = cardType
	}
	mashaledNpuInfo, err := json.Marshal(hdm.getNpuBaseInfo())
	if err != nil {
		return nil, fmt.Errorf("failed to marshal device ip map: %w", err)
	}
	hdm.baseNPUInfo = hdm.getNpuBaseInfo()
	newMashaledNpuInfo := customname.ReplaceDevicePublicName(hdm.RunMode, string(mashaledNpuInfo))
	annotationMap[api.BaseDevInfoAnno] = newMashaledNpuInfo
	annotationMap[common.SuperPodIDKey] = strconv.Itoa(int(hdm.getSuperPodInfo().SuperPodId))
	annotationMap[serverIndexKey] = strconv.Itoa(int(hdm.getSuperPodInfo().ServerId))
	annotationMap[serverTypeKey] = getDevType(common.ParamOption.RealCardType)
	annotationMap[api.RackIDKey] = strconv.Itoa(int(hdm.getSuperPodInfo().RackId))

	return annotationMap, nil
}

func (hdm *HwDevManager) getNewNodeLabel(node *v1.Node) (map[string]string, error) {
	newLabelMap, err := hdm.updateChipNameToNode()
	if err != nil {
		return nil, err
	}
	if _, ok := node.Labels[common.ServerTypeLabelKey]; !ok {
		cardType := common.ParamOption.RealCardType + common.MiddelLine +
			strconv.Itoa(int(common.ParamOption.AiCoreCount))
		newLabelMap[common.ServerTypeLabelKey] = customname.ReplaceDevicePublicName(hdm.RunMode, cardType)

	}
	if len(hdm.allInfo.AllDevs) <= common.FirstDevice {
		return nil, fmt.Errorf("index(%d) exceeds the range of alldevs", common.FirstDevice)
	}
	boardInfo, err := hdm.manager.GetDmgr().GetBoardInfo(hdm.allInfo.AllDevs[common.FirstDevice].LogicID)
	if err != nil {
		return nil, fmt.Errorf("failed to get node board info, err: %s", err.Error())
	}

	if common.HasOnChipMemory() {
		hwlog.RunLog.Debug("get node on-chip-memory info")
		hbmInfo, err := hdm.manager.GetDmgr().GetDeviceHbmInfo(hdm.allInfo.AllDevs[common.FirstDevice].LogicID)
		if err != nil {
			hwlog.RunLog.Warnf("failed to get node on-chip-memory info, err: %s", err)
		} else {
			newLabelMap[api.NPUChipMemoryLabel] = fmt.Sprintf("%dG", hbmInfo.MemorySize/memoryRadix)
		}
	}

	if common.ParamOption.RealCardType == api.Ascend910B && hdm.manager.GetDeviceUsage() == common.Infer {
		// only auto label 300IA2 with910B card
		if boardInfo.BoardId == common.A300IA2BoardId || boardInfo.BoardId == common.A300IA2GB64BoardId {
			newLabelMap[common.AcceleratorTypeKey] = api.A300IA2Label
		}
	}
	if common.IsContainAll300IDuo() {
		newLabelMap[common.InferCardKey] = api.A300IDuoLabel
	}
	return newLabelMap, nil
}

func (hdm *HwDevManager) getNpuBaseInfo() map[string]*common.NpuBaseInfo {
	if common.ParamOption.RealCardType == api.Ascend910A5 {
		if err := hdm.dpuManager.SaveDpuConfToNode(hdm.manager.GetDmgr()); err != nil {
			hwlog.RunLog.Errorf("%s failed to save dpu info to node, err: %v", api.DpuLogPrefix, err)
		}
	}
	ipMap := make(map[string]*common.NpuBaseInfo, len(hdm.allInfo.AllDevs))
	for index, dev := range hdm.allInfo.AllDevs {
		tmpDev := dev
		levelList := hdm.getLevelList(&tmpDev)
		ipMap[tmpDev.DeviceName] = &common.NpuBaseInfo{
			IP:            tmpDev.IP,
			SuperDeviceID: tmpDev.SuperDeviceID,
			// node baseDeviceInfo levelList -> rank table for A5
			LevelList: levelList,
		}
		hdm.allInfo.AllDevs[index].LevelList = levelList
	}
	return ipMap
}

func (hdm *HwDevManager) updateChipNameToNode() (map[string]string, error) {
	newLabelMap := make(map[string]string, 1)
	chipInfo, err := hdm.manager.GetDmgr().GetValidChipInfo()
	if err != nil {
		return nil, err
	}
	newLabelMap[common.ChipNameLabel] = chipInfo.Name
	return newLabelMap, nil
}

func (hdm *HwDevManager) setAllDeviceAndType() error {
	kubeClient, err := kubeclient.NewClientK8s()
	if err != nil {
		hwlog.RunLog.Errorf("init k8s client failed err: %v", err.Error())
		return err
	}
	hdm.manager.SetKubeClient(kubeClient)

	if hdm.allInfo, err = hdm.manager.GetNPUs(); err != nil {
		return err
	}
	if len(hdm.allInfo.AllDevTypes) == 0 {
		return fmt.Errorf("no devices type found")
	}
	if len(hdm.allInfo.AllDevs) == 0 {
		return fmt.Errorf("no devices found")
	}
	if err = hdm.manager.SetDeviceUsage(hdm.allInfo.AllDevs[0].LogicID); err != nil {
		return err
	}
	hdm.groupDevice = device.ClassifyDevices(hdm.allInfo.AllDevs, hdm.allInfo.AllDevTypes)
	return nil
}

func (hdm *HwDevManager) getSuperPodInfo() common.SuperPodInfo {
	result := common.SuperPodInfo{
		ScaleType:    common.ScaleTypeAbnormal,
		SuperPodId:   common.SuperPodIdAbnormal,
		ServerId:     common.ServerIdAbnormal,
		RackId:       common.RackIdAbnormal,
		SuperPodType: common.SuperPodTypeAbnormal,
		Reserve:      make([]int8, 0),
	}
	for _, npuDevices := range hdm.groupDevice {
		for _, npuDevice := range npuDevices {
			superPodInfo, err := hdm.manager.GetDmgr().GetSuperPodInfo(npuDevice.LogicID)
			if err != nil {
				hwlog.RunLog.Warnf("failed to get super pod info, error: %v", err)
				continue
			}
			if common.ParamOption.RealCardType == api.Ascend910A5 && int(superPodInfo.RackId) < 0 {
				hwlog.RunLog.Warnf("failed to get super pod info, rack id invalid: %v", superPodInfo.RackId)
				continue
			}
			hwlog.RunLog.Infof("get super pod info: %v", superPodInfo)
			npuDevice.SuperDeviceID = superPodInfo.SdId
			if result.ScaleType != common.ScaleTypeAbnormal {
				continue
			}
			result = common.SuperPodInfo{
				ScaleType:    int32(superPodInfo.ScaleType),
				SuperPodId:   int32(superPodInfo.SuperPodId),
				ServerId:     int32(superPodInfo.ServerId),
				RackId:       int32(superPodInfo.RackId),
				SuperPodType: int8(superPodInfo.SuperPodType),
			}
			for i := 0; i < len(superPodInfo.Reserve); i++ {
				result.Reserve = append(result.Reserve, int8(superPodInfo.Reserve[i]))
			}
		}
	}

	return result
}

// setSuperPodInfo get super pod info then cache it
func (hdm *HwDevManager) setSuperPodInfo() {
	superPodInfo := hdm.getSuperPodInfo()
	hwlog.RunLog.Infof("get super pod id: %d, server index: %d", superPodInfo.SuperPodId, superPodInfo.ServerId)
	hdm.manager.SetSuperPodID(superPodInfo.SuperPodId)
	hdm.manager.SetServerIndex(superPodInfo.ServerId)
	if common.ParamOption.RealCardType == api.Ascend910A5 {
		hwlog.RunLog.Infof("get rack id: %d", superPodInfo.RackId)
		hdm.manager.SetRackID(superPodInfo.RackId)
		hwlog.RunLog.Infof("get super pod type: %d", superPodInfo.SuperPodType)
		if _, exist := hcclTopoFilePathMap[superPodInfo.SuperPodType]; !exist {
			hwlog.RunLog.Warnf("device super pod type[%d] invalid", superPodInfo.SuperPodType)
		}
		hdm.manager.SetSuperPodType(superPodInfo.SuperPodType)
		hwlog.RunLog.Infof("get super pod size: %d", superPodInfo.ScaleType)
		hdm.manager.SetSuperPodSize(superPodInfo.ScaleType)
	}
}

func (hdm *HwDevManager) initPluginServer() error {
	hdm.ServerMap = make(map[string]InterfaceServer, len(hdm.allInfo.AllDevTypes))
	defaultDevices, err := common.GetDefaultDevices(common.ParamOption.GetFdFlag)
	if err != nil {
		hwlog.RunLog.Error("get default device error")
		return err
	}
	if !common.ParamOption.PresetVDevice {
		hdm.ServerMap[common.AiCoreResourceName] = NewPluginServer(common.AiCoreResourceName,
			hdm.allInfo.AICoreDevs, defaultDevices, hdm.manager)
		return nil
	}
	for _, deviceType := range hdm.allInfo.AllDevTypes {
		hdm.ServerMap[deviceType] = NewPluginServer(deviceType, hdm.groupDevice[deviceType], defaultDevices,
			hdm.manager)
	}
	return nil
}

func (hdm *HwDevManager) checkSupportedProductType() error {
	if !common.ParamOption.PresetVDevice && common.IsContainAtlas300IDuo() {
		return fmt.Errorf("%s is not supported to dynamic virtual instance", common.Atlas300IDuo)
	}
	return nil
}

// GetNPUs will set device default health, actually, it should be based on the last status if exist
func (hdm *HwDevManager) updateDeviceHealth(curAllDevs []common.NpuDevice) {
	lastAllDevs := make(map[string]int, len(hdm.allInfo.AllDevs))
	for index, dev := range hdm.allInfo.AllDevs {
		lastAllDevs[dev.DeviceName] = index
	}
	for i, dev := range curAllDevs {
		if index, exist := lastAllDevs[dev.DeviceName]; exist && index < len(hdm.allInfo.AllDevs) {
			curAllDevs[i].Health = hdm.allInfo.AllDevs[index].Health
			curAllDevs[i].NetworkHealth = hdm.allInfo.AllDevs[index].NetworkHealth
			curAllDevs[i].DpuHealth = hdm.allInfo.AllDevs[index].DpuHealth
			curAllDevs[i].FaultCodes = hdm.allInfo.AllDevs[index].FaultCodes
			curAllDevs[i].AlarmRaisedTime = hdm.allInfo.AllDevs[index].AlarmRaisedTime
			curAllDevs[i].NetworkFaultCodes = hdm.allInfo.AllDevs[index].NetworkFaultCodes
			curAllDevs[i].NetworkAlarmRaisedTime = hdm.allInfo.AllDevs[index].NetworkAlarmRaisedTime
		}
	}
}

func (hdm *HwDevManager) updateAllInfo() error {
	if common.ParamOption.PresetVDevice {
		return nil
	}
	element, exist := hdm.ServerMap[common.AiCoreResourceName]
	if !exist {
		return fmt.Errorf("not found %s plugin server", common.AiCoreResourceName)
	}
	pluginServer, ok := element.(*PluginServer)
	if !ok {
		return fmt.Errorf("serverMap convert %s failed", common.AiCoreResourceName)
	}
	err := pluginServer.DestroyNotUsedVNPU()
	if err != nil {
		return err
	}
	if err := hdm.manager.CheckDeviceTypeLabel(); err != nil {
		hwlog.RunLog.Warnf("device type label may not correct, %v", err)
	}
	allInfo, err := hdm.manager.GetNPUs()
	if err != nil {
		return err
	}
	hdm.updateDeviceHealth(allInfo.AllDevs)
	hdm.groupDevice = device.ClassifyDevices(allInfo.AllDevs, allInfo.AllDevTypes)
	hdm.allInfo = allInfo
	return nil
}

func (hdm *HwDevManager) loadDeviceInfoCm() {
	hdm.manager.LoadDeviceInfoCm()
}

func (hdm *HwDevManager) handleDeviceInfoUpdate(ctx context.Context, initTime *time.Time) {
	common.LockAllDeviceInfo()
	defer common.UnlockAllDeviceInfo()

	if err := hdm.updateAllInfo(); err != nil {
		hwlog.RunLog.Error(err)
		return
	}

	// complete the fault codes that cannot be reported by the event subscribe interface
	hdm.mendSubscribeFaultEvents()
	if err := hdm.updatePodAnnotation(); err != nil {
		hwlog.RunLog.Error(err)
	}
	hdm.updateDeviceUsedInfo(hdm.groupDevice)
	hdm.notifyToK8s(ctx, initTime)

	// if node annotation has reset fail devices but all devices are healthy, clear node annotation
	hdm.checkNodeResetInfo()
	hdm.useVolcanoNotify()
	hdm.chipHotReset()
	common.DelOnceRecoverFault(hdm.groupDevice)
	common.DelOnceFrequencyFault()
	common.Synchronize = true
}

// ListenDevice ListenDevice coroutine
func (hdm *HwDevManager) ListenDevice(ctx context.Context) {
	hwlog.RunLog.Info("starting the listen device")
	hdm.subscribeFaultEvent()
	if common.ParamOption.RealCardType == api.Ascend910A3 && common.ParamOption.EnableSwitchFault {
		// will set a goroutine to query all switch faults every 5 min
		go hdm.SwitchDevManager.GetSwitchFaultCodeByInterval(ctx, time.Second*common.GetSwitchFaultCodeInterval)
	}
	// when device-plugin is started, the value of ManuallySeparateNPU and upgrade fault reason in device info configmap
	// needs to be written into cache to prevent manually separate npu IDs in cache from been lost
	hdm.loadDeviceInfoCm()
	go hdm.pollFaultCodeCM(ctx)
	go hdm.Serve(ctx)
	if common.ParamOption.CheckCachedPods {
		go hdm.manager.GetKubeClient().PodInformerInspector(ctx)
	}
	go hdm.updateNodeAnnotations(ctx)

	// report device fault to k8s event
	go hdm.manager.WriteFaultToEvent(ctx)
	initTime := time.Now()
	ticker := time.NewTicker(time.Duration(common.ParamOption.ListAndWatchPeriod) * time.Second)
	defer ticker.Stop()
	triggerTicker := time.NewTicker(time.Second)
	defer triggerTicker.Stop()

	for {
		select {
		case _, ok := <-ctx.Done():
			if !ok {
				hwlog.RunLog.Info("catch stop signal channel closed")
			}
			hwlog.RunLog.Info("listen device stop")
			return
		case <-triggerTicker.C:
			hdm.parseTriggers(ctx, initTime)
		case <-ticker.C:
			hwlog.RunLog.Debug("Periodic device info update")
			hdm.handleDeviceInfoUpdate(ctx, &initTime)
		}
	}
}

func (hdm *HwDevManager) updateNodeAnnotations(ctx context.Context) {
	ticker := time.NewTicker(nodeAnnotationUpdateInterval * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			hdm.doUpdateNodeAnnotations()
		}
	}
}

func (hdm *HwDevManager) doUpdateNodeAnnotations() {
	baseInfoChange, newBaseInfo := hdm.compareBaseNPUInfo()
	if !baseInfoChange {
		hwlog.RunLog.Debug("npu IP is not changed")
		return
	}
	hwlog.RunLog.Info("base npu info changed, update node annotation")
	mashaledNpuInfo, err := json.Marshal(newBaseInfo)
	if err != nil {
		hwlog.RunLog.Errorf("failed to marshal device ip map, err: %v", err)
		return
	}

	for i := 0; i < common.RetryUpdateCount; i++ {
		if err = hdm.manager.GetKubeClient().AddAnnotation(api.BaseDevInfoAnno, string(mashaledNpuInfo)); err == nil {
			hwlog.RunLog.Info("update node annotations success")
			hdm.baseNPUInfo = newBaseInfo
			return
		}
		hwlog.RunLog.Warnf("failed to patch new label to node, err: %s, retry count: %d", err.Error(), i+1)
		time.Sleep(time.Second)
	}
}

func (hdm *HwDevManager) compareBaseNPUInfo() (bool, map[string]*common.NpuBaseInfo) {
	baseInfoChange := false
	newInfo := make(map[string]*common.NpuBaseInfo, len(hdm.baseNPUInfo))
	for _, dev := range hdm.allInfo.AllDevs {
		info, ok := hdm.baseNPUInfo[dev.DeviceName]
		if !ok {
			continue
		}
		newItem := &common.NpuBaseInfo{
			IP:            info.IP,
			SuperDeviceID: info.SuperDeviceID,
		}

		newInfo[dev.DeviceName] = newItem

		ip, err := hdm.manager.GetDeviceIP(dev.DevType, int(dev.PhyID))
		if err != nil {
			hwlog.RunLog.Warnf("get %s device ip failed, err: %v", dev.DeviceName, err)
			continue
		}
		if info.IP != ip {
			baseInfoChange = true
			newItem.IP = ip
		}
	}

	return baseInfoChange, newInfo
}

func (hdm *HwDevManager) parseTriggers(ctx context.Context, initTime time.Time) {
	select {
	case <-common.GetUpdateChan():
		hwlog.RunLog.Info("Received update trigger, processing device info update")
		hdm.handleDeviceInfoUpdate(ctx, &initTime)
	default:
		hwlog.RunLog.Debug("No update trigger, skipping device info update")
	}
}

func deepCopyGroupDevice(groupDevice map[string][]*common.NpuDevice) map[string][]*common.NpuDevice {
	newGroupDevice := make(map[string][]*common.NpuDevice, len(groupDevice))
	for deviceType, npuDevices := range groupDevice {
		newNpuDevices := make([]*common.NpuDevice, 0, len(npuDevices))
		for _, npuDevice := range npuDevices {
			newNpuDevice := &common.NpuDevice{
				FaultCodes:             npuDevice.FaultCodes,
				AlarmRaisedTime:        npuDevice.AlarmRaisedTime,
				NetworkFaultCodes:      npuDevice.NetworkFaultCodes,
				NetworkAlarmRaisedTime: npuDevice.NetworkAlarmRaisedTime,
				FaultTimeMap:           npuDevice.FaultTimeMap,
				DevType:                npuDevice.DevType,
				DeviceName:             npuDevice.DeviceName,
				Health:                 npuDevice.Health,
				NetworkHealth:          npuDevice.NetworkHealth,
				DpuHealth:              npuDevice.DpuHealth,
				IP:                     npuDevice.IP,
				LogicID:                npuDevice.LogicID,
				PhyID:                  npuDevice.PhyID,
				CardID:                 npuDevice.CardID,
				Status:                 npuDevice.Status,
				PodUsed:                npuDevice.PodUsed,
			}
			newNpuDevices = append(newNpuDevices, newNpuDevice)
		}
		newGroupDevice[deviceType] = newNpuDevices
	}
	return newGroupDevice
}

func (hdm *HwDevManager) updateDeviceUsedInfo(groupDevice map[string][]*common.NpuDevice) {
	podUsedChips := hdm.manager.GetKubeClient().GetPodsUsedNPUByKlt()
	hwlog.RunLog.Debugf("update deviceUsedInfo podUsedChips: %v", podUsedChips)
	for _, devices := range groupDevice {
		for _, deviceInfo := range devices {
			deviceInfo.PodUsed = podUsedChips.Has(deviceInfo.DeviceName)
		}
	}
}

func (hdm *HwDevManager) pluginNotify(classifyDev []*common.NpuDevice, devType string) {
	serverMap, ok := hdm.ServerMap[devType]
	if !ok {
		hwlog.RunLog.Warnf("server map (%s) not exist", devType)
		return
	}
	pluginServer, ok := serverMap.(*PluginServer)
	if !ok {
		hwlog.RunLog.Warnf("pluginServer (%s) not ok", devType)
		return
	}
	if !pluginServer.Notify(classifyDev) {
		hwlog.RunLog.Warnf("deviceType(%s) notify failed, server may not start, please check", devType)
	}
}

func (hdm *HwDevManager) notifyToK8s(ctx context.Context, initTime *time.Time) {
	hdm.isSupportGraceTolerance()
	oldGroupDevice := deepCopyGroupDevice(hdm.groupDevice)
	hdm.manager.UpdateHealth(hdm.groupDevice, hdm.allInfo.AICoreDevs, hdm.RunMode)
	if hdm.manager.GetDmgr().GetDevType() == api.Ascend910A5 {
		hdm.updateDpuHealthy(hdm.groupDevice)
	}
	// If hot reset is used, the health of the device being reset is set here to healthy
	hdm.graceTolerance(ctx, hdm.groupDevice)
	isDevStateChange := hdm.manager.GetChange(hdm.groupDevice, oldGroupDevice)

	for devType, isChanged := range isDevStateChange {
		server := hdm.ServerMap[devType]
		if server == nil {
			continue
		}
		if !isChanged &&
			(time.Now().Sub(*initTime) < time.Minute || server.LastSendSuccess()) &&
			time.Now().Sub(*initTime) < time.Hour {
			continue
		}
		*initTime = time.Now()
		if !common.ParamOption.PresetVDevice {
			hdm.pluginNotify(hdm.allInfo.AICoreDevs, common.AiCoreResourceName)
			return
		}
		hdm.pluginNotify(hdm.groupDevice[devType], devType)
	}
}

func (hdm *HwDevManager) chipHotReset() {
	// both 910B[A800IA2] and 310 will be used as infer device
	if common.ParamOption.HotReset != common.HotResetInfer {
		hwlog.RunLog.Debugf("infer device hot reset mode error: %d", common.ParamOption.HotReset)
		return
	}
	prClient := NewPodResource()
	for devType, devices := range hdm.groupDevice {
		if common.IsVirtualDev(devType) || len(devices) == 0 {
			continue
		}
		if common.IsContainAtlas300IDuo() {
			hdm.resetDuoCard(devType, devices, prClient)
			continue
		}
		hdm.resetCommonInferCard(devType, devices, prClient)
	}
}

func (hdm *HwDevManager) resetCommonInferCard(devType string, devices []*common.NpuDevice, prClient *PodResource) {
	if hdm == nil || len(hdm.allInfo.AllDevs) == 0 {
		hwlog.RunLog.Error("invalid params")
		return
	}
	if common.ParamOption.RealCardType == api.Ascend910A3 {
		hdm.ResetServerForA3(devType, devices, prClient)
		return
	}

	usage, boardId, err := hdm.getServerUsageAndBoardId()
	if err != nil {
		hwlog.RunLog.Error(err)
		return
	}

	// A800IA2 server, node labeled with server-usage=infer
	if usage == common.Infer {
		// server without hccs is 0x33 or 0x3c, 0x28, 0x29
		if boardId == common.A800IA2NoneHccsBoardId || boardId == common.A800IA2NoneHccsBoardIdOld ||
			boardId == common.A300IA2BoardId || boardId == common.A300IA2GB64BoardId {
			hdm.ResetWithoutHccsServer(devType, devices, prClient)
			return
		}
		hdm.ResetHccsServer(devType, devices, prClient)
		return
	}
	for _, device := range devices {
		if device.Health == v1beta1.Healthy {
			continue
		}
		if !hdm.isPodRemove(devType, device, prClient) {
			continue
		}
		if !hdm.checkNoProc(device.LogicID) {
			continue
		}
		hdm.hotReset(device, []*common.NpuDevice{device})
	}
}

func (hdm *HwDevManager) getServerUsageAndBoardId() (string, uint32, error) {
	boardId, err := hdm.manager.GetServerBoardId(hdm.allInfo.AllDevs[common.FirstDevice].LogicID)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get node board info, err: %s", err.Error())
		return "", common.EmptyBoardId, err
	}

	client := hdm.manager.GetKubeClient()
	if client == nil {
		return "", common.EmptyBoardId, fmt.Errorf("k8s client is nil")
	}
	// try to get server-usage label
	usage, err := client.GetServerUsageLabelCache()
	if err != nil {
		hwlog.RunLog.Errorf("failed to get server usage")
		return "", common.EmptyBoardId, err
	}
	return usage, boardId, nil
}

// ResetWithoutHccsServer reset server without hccs, which can reset one card at one time
func (hdm *HwDevManager) ResetWithoutHccsServer(devType string, devices []*common.NpuDevice, prClient *PodResource) {
	for _, device := range devices {
		inReset := hdm.manager.GetIfCardsInResetting(device.LogicID)
		resetFailedTimes := hdm.manager.GetResetFailedTimes(device.LogicID)
		if device.Health == v1beta1.Healthy {
			hwlog.RunLog.Warnf("Ascend910-%d is health, would not reset", device.LogicID)
			continue
		}
		if inReset {
			hwlog.RunLog.Warnf("Ascend910-%d is inReset, would not reset", device.LogicID)
			continue
		}
		if resetFailedTimes >= common.MaxResetTimes {
			hwlog.RunLog.Warnf("Ascend910-%d exceeds MaxResetTimes, would not reset", device.LogicID)
			continue
		}
		if !hdm.isPodRemove(devType, device, prClient) {
			hwlog.RunLog.Warnf("Ascend910-%d contains pod, would not reset", device.LogicID)
			continue
		}
		if !hdm.checkNoProc(device.LogicID) {
			hwlog.RunLog.Warnf("Ascend910-%d contains proc, would not reset", device.LogicID)
			continue
		}
		// to avoid blocking for minutes
		go hdm.hotReset(device, []*common.NpuDevice{device})
	}
}

func (hdm *HwDevManager) checkNoProc(logicID int32) bool {
	logicIDForCompare := fmt.Sprintf("Ascend910-%d", logicID)
	processInfo, err := hdm.manager.GetDmgr().GetDevProcessInfo(logicID)
	if err != nil || processInfo == nil {
		hwlog.RunLog.Errorf("failed to get device process, logicId: %s, err: %v, devProcessInfo: %v",
			logicIDForCompare, err, processInfo)
		return false
	}
	if processInfo.ProcNum != 0 {
		hwlog.RunLog.Errorf("found busy chip: %v", logicIDForCompare)
		return false
	}
	return true
}

// ResetHccsServer try to reset server with hccs, which need to reset all cards at once
func (hdm *HwDevManager) ResetHccsServer(devType string, devices []*common.NpuDevice, prClient *PodResource) {
	//  if all cards are healthy will do no more action, to log less
	allHealthy := true
	for _, npu := range devices {
		allHealthy = allHealthy && (npu.Health == v1beta1.Healthy)
	}
	if hdm.manager.GetResetFailedTimes(common.FirstDevice) > common.MaxResetTimes {
		hwlog.RunLog.Warnf("reset failed more than %d times without success, hot reset will be disabled "+
			"util device-plugin restarted", common.MaxResetTimes)
		return
	}
	if allHealthy || hdm.manager.GetIfCardsInResetting(common.FirstDevice) {
		return
	}

	freeDeviceNum := 0
	needReset := false
	for _, device := range devices {
		if device.Health != v1beta1.Healthy {
			needReset = true
		}
		if !hdm.isPodRemove(devType, device, prClient) {
			break
		}
		if !hdm.checkNoProc(device.LogicID) {
			break
		}
		freeDeviceNum++
	}

	if needReset && freeDeviceNum == common.Ascend910BRingsNumTrain {
		if common.FirstDevice >= len(devices) {
			hwlog.RunLog.Errorf("index out of range: giving devices index %d, "+
				"real length %d", common.FirstDevice, len(devices))
			return
		}
		hdm.hotReset(devices[common.FirstDevice], devices)
	}
}

// ResetServerForA3 reset server device for a3
func (hdm *HwDevManager) ResetServerForA3(devType string, devices []*common.NpuDevice, prClient *PodResource) {
	coverIdSet := sets.NewInt32()
	for _, npuDevice := range devices {
		if npuDevice.Health == v1beta1.Healthy || coverIdSet.Has(npuDevice.LogicID) {
			continue
		}
		cardID, deviceID, err := hdm.manager.GetDmgr().GetCardIDDeviceID(npuDevice.LogicID)
		if err != nil {
			hwlog.RunLog.Errorf("get card id and device id failed, logic id: %d err: %v",
				npuDevice.LogicID, err)
			continue
		}
		logicIDs, err := hdm.manager.GetAssociatedLogicIDs(npuDevice.LogicID, cardID, deviceID)
		if err != nil || len(logicIDs) == 0 {
			hwlog.RunLog.Errorf("invalid associated logic id list %v, err: %v", logicIDs, err)
			continue
		}
		idSet := sets.NewInt32(logicIDs...)
		deviceList := make([]*common.NpuDevice, 0, len(logicIDs))
		freeDeviceNum := 0
		for _, dev := range devices {
			if !idSet.Has(dev.LogicID) {
				continue
			}
			deviceList = append(deviceList, dev)
			inReset := hdm.manager.GetIfCardsInResetting(dev.LogicID)
			resetFailedTimes := hdm.manager.GetResetFailedTimes(dev.LogicID)
			podRemoved := hdm.isPodRemove(devType, dev, prClient)
			noProc := hdm.checkNoProc(dev.LogicID)
			if inReset || resetFailedTimes >= common.MaxResetTimes || !podRemoved || !noProc {
				hwlog.RunLog.Infof("device %v cant reset, "+
					"inReset: %v, resetFailedTimes: %v, podRemoved: %v, noProc: %v",
					dev.DeviceName, inReset, resetFailedTimes, podRemoved, noProc)
				break
			}
			freeDeviceNum++
		}
		if freeDeviceNum == len(logicIDs) {
			hwlog.RunLog.Infof("start reset device, logic id list %v", logicIDs)
			// to avoid blocking for minutes
			go hdm.hotReset(npuDevice, deviceList)
		}
		coverIdSet.Insert(logicIDs...)
	}
}

func (hdm *HwDevManager) resetDuoCard(devType string, devices []*common.NpuDevice, prClient *PodResource) {
	var cardResetOnce = make(map[int32][]*common.NpuDevice, 1)
	for _, device := range devices {
		cardResetOnce[device.CardID] = append(cardResetOnce[device.CardID], device)
	}
	for _, deviceChip := range cardResetOnce {
		if hdm.isDuoCardChipHealthy(deviceChip) {
			continue
		}
		if !hdm.isDuoRemove(devType, deviceChip, prClient) {
			continue
		}
		if len(deviceChip) == 0 {
			hwlog.RunLog.Error("device chip is empty")
			continue
		}
		hdm.hotReset(deviceChip[0], deviceChip)
	}
}

func (hdm *HwDevManager) isDuoRemove(devType string, deviceChip []*common.NpuDevice, prClient *PodResource) bool {
	for _, dev := range deviceChip {
		if !hdm.isPodRemove(devType, dev, prClient) {
			return false
		}
	}
	return true
}

func (hdm *HwDevManager) isDuoCardChipHealthy(deviceChip []*common.NpuDevice) bool {
	for _, dev := range deviceChip {
		if dev.Health == v1beta1.Unhealthy {
			return false
		}
	}
	return true
}

func (hdm *HwDevManager) useVolcanoNotify() {
	if common.ParamOption.BuildScene == common.EdgeScene {
		return
	}
	if hdm.manager.GetKubeClient() == nil {
		hwlog.RunLog.Error("kube client is nil, can't interacting with k8s")
		return
	}
	common.DpStartReset.Do(func() {
		if err := hdm.manager.GetKubeClient().AnnotationReset(); err != nil {
			hwlog.RunLog.Warn("device plugin first reset annotation and config map error")
		}
	})
	hdm.manager.DoWithVolcanoListAndWatch(hdm.groupDevice)
}

// SignCatch stop system sign catch
func (hdm *HwDevManager) SignCatch(cancel context.CancelFunc) {
	osSignChan := common.NewSignWatcher(syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGKILL)
	if osSignChan == nil {
		hwlog.RunLog.Error("the stop signal is not initialized")
		return
	}
	select {
	case s, signEnd := <-osSignChan:
		if signEnd == false {
			hwlog.RunLog.Info("catch stop signal channel is closed")
			return
		}
		hwlog.RunLog.Infof("Received signal: %s, shutting down.", s.String())
		cancel()
		hdm.stopAllSever()
		hdm.manager.GetDmgr().ShutDown()
		hdm.SwitchDevManager.ShutDownSwitch()
	}
}

// Serve Serve function
func (hdm *HwDevManager) Serve(ctx context.Context) {
	// initiate a global socket path watcher
	hwlog.RunLog.Info("Serve start")
	watcher, err := common.NewFileWatch()
	if err != nil {
		hwlog.RunLog.Error("createSocketWatcher error")
		return
	}
	defer func() {
		if watcher == nil {
			hwlog.RunLog.Error("watcher is nil")
			return
		}
		if err := watcher.FileWatcher.Close(); err != nil {
			hwlog.RunLog.Errorf("close file watcher, err: %v", err)
		}
	}()

	// create restart signal
	restartSignal := common.NewSignWatcher(syscall.SIGHUP)

	for {
		allSuccess := hdm.startAllServer(watcher)
		if hdm.handleEvents(ctx, restartSignal, watcher) {
			break
		}
		if !allSuccess {
			time.Sleep(common.SleepTime * time.Second)
		}
	}
}

func (hdm *HwDevManager) handleEvents(ctx context.Context, restartSignal chan os.Signal,
	watcher *common.FileWatch) bool {

	if restartSignal == nil {
		hwlog.RunLog.Error("the restart signal is not initialized")
		return true
	}

	select {
	case <-ctx.Done():
		hwlog.RunLog.Info("stop signal received, stop device plugin")
		return true
	case sig, ok := <-restartSignal:
		if ok {
			hwlog.RunLog.Infof("restart signal %s received, restart device plugin", sig)
			hdm.setRestartForAll()
		}
	case event := <-watcher.FileWatcher.Events:
		if event.Op&fsnotify.Remove == fsnotify.Remove {
			_, deleteFile := filepath.Split(event.Name)
			hdm.handleDeleteEvent(deleteFile)
		}
		if event.Name == v1beta1.KubeletSocket && event.Op&fsnotify.Create == fsnotify.Create {
			hwlog.RunLog.Info("notify: kubelet.sock file created.")
		}
	default:
		time.Sleep(common.CheckFailurePeriodSecond)
	}
	return false
}

func (hdm *HwDevManager) stopAllSever() {
	for deviceType := range hdm.ServerMap {
		hwlog.RunLog.Infof("stop server type %s", deviceType)
		hdm.ServerMap[deviceType].Stop()
	}
	hwlog.RunLog.Info("stop all server done")
}

func (hdm *HwDevManager) setRestartForAll() {
	for deviceType := range hdm.ServerMap {
		hdm.ServerMap[deviceType].SetRestartFlag(true)
	}
}

func (hdm *HwDevManager) startAllServer(socketWatcher *common.FileWatch) bool {
	success := true
	for deviceType, serverInterface := range hdm.ServerMap {
		if !serverInterface.GetRestartFlag() {
			continue
		}
		if err := serverInterface.Start(socketWatcher); err != nil {
			hwlog.RunLog.Errorf("Could not contact Kubelet for %s, retrying. "+
				"Did you enable the device plugin feature gate?", deviceType)
			success = false
		} else {
			serverInterface.SetRestartFlag(false)
		}
	}
	return success
}

func (hdm *HwDevManager) handleDeleteEvent(deleteFile string) {
	for deviceType := range hdm.ServerMap {
		candidateSocketFilename := fmt.Sprintf("%s.sock", deviceType)
		if candidateSocketFilename == deleteFile {
			hwlog.RunLog.Warnf("notify: sock file %s deleted, please check !", deleteFile)
		}
	}
}

func (hdm *HwDevManager) updatePodAnnotation() error {
	nodeIp, err := hdm.manager.GetKubeClient().GetNodeIpCache()
	if err != nil {
		return fmt.Errorf("get node server id failed: %v", err)
	}
	if !common.ParamOption.PresetVDevice {
		return hdm.updateSpecTypePodAnnotation(common.AiCoreResourceName, nodeIp)
	}
	for _, devType := range hdm.allInfo.AllDevTypes {
		// for 310P vnpu no need update
		if common.IsVirtualDev(devType) && !strings.HasPrefix(devType, api.Ascend910) {
			continue
		}
		if err := hdm.updateSpecTypePodAnnotation(devType, nodeIp); err != nil {
			hwlog.RunLog.Warnf("update pod annotation failed, %v", err)
		}
	}
	return nil
}

// tryToClearResetInfoCM try to clear reset info config map
func (hdm *HwDevManager) tryToClearResetInfoCM(pod v1.Pod) error {
	taskName, ok := pod.Annotations[common.ResetTaskNameKey]
	if !ok {
		taskName, ok = pod.Labels[common.ResetTaskNameKeyInLabel]
		if !ok {
			hwlog.RunLog.Error("failed to get task name by task key in tryToClearResetInfoCM")
			return fmt.Errorf("failed to get task name by task key")
		}
	}
	resetInfo, err := hdm.manager.GetKubeClient().GetConfigMap(
		common.ResetInfoCMNamePrefix+taskName, pod.Namespace)
	if err != nil {
		hwlog.RunLog.Warnf("get reset configMap failed, because: %v", err)
		return err
	}

	data, ok := resetInfo.Data[common.ResetInfoCMDataKey]
	if !ok {
		return fmt.Errorf("%s not exist", common.ResetInfoCMDataKey)
	}
	if len(data) > common.CMDataMaxLength {
		return fmt.Errorf("configmap data size is out of memory")
	}
	var taskResetInfo common.TaskResetInfo
	if err := json.Unmarshal([]byte(data), &taskResetInfo); err != nil {
		return fmt.Errorf("unmarshal configmap data failed, err: %v", err)
	}
	// skip it when the reset info config map is initialized
	if taskResetInfo.UpdateTime == 0 {
		return nil
	}

	if err := hdm.manager.GetKubeClient().ClearResetInfo(taskName, pod.Namespace); err != nil {
		return fmt.Errorf("clear reset configMap failed err is: %v", err)
	}
	return nil
}

// updateSpecTypePodAnnotation will update annotation of pod and
// try to clear reset info config map which may not be initialized after rescheduling
func (hdm *HwDevManager) updateSpecTypePodAnnotation(deviceType, serverID string) error {
	element, exist := hdm.ServerMap[deviceType]
	if !exist {
		return fmt.Errorf("not found %s plugin server", deviceType)
	}
	pluginServer, ok := element.(*PluginServer)
	if !ok {
		return fmt.Errorf("serverMap convert %s failed", deviceType)
	}
	podList := hdm.manager.GetKubeClient().GetActivePodListCache()
	podDeviceInfo, err := pluginServer.GetKltAndRealAllocateDev(podList)
	if err != nil {
		return err
	}
	for _, deviceInfo := range podDeviceInfo {
		hwlog.RunLog.Debugf("pods: %s, %s, %s", deviceInfo.Pod.Name, deviceInfo.Pod.Status.Phase, deviceInfo.Pod.UID)
		_, existRealAlloc := deviceInfo.Pod.Annotations[api.PodAnnotationAscendReal]
		if existRealAlloc {
			hwlog.RunLog.Debug("The field AscendReal exists; device plugin skips writing the annotation")
			continue
		}
		if len(deviceInfo.KltDevice) == 0 || len(deviceInfo.RealDevice) == 0 {
			hwlog.RunLog.Warnf("%s %s klt device or real device is empty", deviceInfo.Pod.Namespace,
				deviceInfo.Pod.Name)
			continue
		}
		hwlog.RunLog.Debugf("%s, %d, %v", deviceInfo.Pod.Name, len(deviceInfo.KltDevice), deviceInfo.RealDevice)
		hwlog.RunLog.Debug("Write annotation via device plugin")
		if err := hdm.manager.AddPodAnnotation(deviceInfo, deviceType, serverID, hdm.allInfo.AllDevs); err != nil {
			hwlog.RunLog.Errorf("update pod %s_%s annotation failed, %v", deviceInfo.Pod.Namespace,
				deviceInfo.Pod.Name, err)
		}

		if common.ParamOption.HotReset != common.HotResetTrainOnLine {
			continue
		}

		// need to clear reset info config map after rescheduling
		if err = hdm.tryToClearResetInfoCM(deviceInfo.Pod); err != nil {
			hwlog.RunLog.Warnf("try to clear configMap failed, err is: %v", err)
		}
	}
	return nil
}

func (hdm *HwDevManager) hotReset(device *common.NpuDevice, devices []*common.NpuDevice) {
	hwlog.RunLog.Infof("will start to reset device %s", device.DeviceName)
	hdm.manager.SetCardsInResetting(device.LogicID, true)
	var isResetExec = false
	successResetDevList := sets.NewInt32()
	if err := wait.PollImmediate(time.Second, time.Minute, func() (bool, error) {
		if err := hdm.execResetChip(device.LogicID, &isResetExec); err != nil {
			hwlog.RunLog.Errorf("get device boot status failed, err: %v", err)
			return false, err
		}
		// check all device state that hot reset together
		for _, dev := range devices {
			if successResetDevList.Has(dev.LogicID) {
				continue
			}
			bootState, err := hdm.manager.GetDmgr().GetDeviceBootStatus(dev.LogicID)
			if err != nil {
				hwlog.RunLog.Errorf("get device %v boot status failed, err: %v", dev.LogicID, err)
				return false, err
			}
			if bootState != common.BootStartFinish {
				hwlog.RunLog.Warnf("device %v bootState(%d), starting...", dev.LogicID, bootState)
				return false, nil
			}
			successResetDevList.Insert(dev.LogicID)
		}
		common.SetDeviceInit(device.LogicID)
		return true, nil
	}); err != nil {
		hwlog.RunLog.Warnf("hot reset failed, timeout or err: %v", err)
		hdm.manager.SetCardsInResetting(device.LogicID, false)
		hdm.manager.SetResetFailedTimes(device.LogicID, hdm.manager.GetResetFailedTimes(device.LogicID)+1)
		return
	}
	hdm.manager.SetResetFailedTimes(device.LogicID, 0)
	hdm.manager.SetCardsInResetting(device.LogicID, false)
	hwlog.RunLog.Info("hot reset success")
}

func (hdm *HwDevManager) isPodRemove(devType string, device *common.NpuDevice, prClient *PodResource) bool {
	podList := hdm.manager.GetKubeClient().GetAllPodListCache()
	element, exist := hdm.ServerMap[devType]
	if !exist {
		hwlog.RunLog.Errorf("not found %s plugin server", devType)
		return false
	}
	pluginServer, ok := element.(*PluginServer)
	if !ok {
		hwlog.RunLog.Errorf("serverMap convert %s failed", devType)
		return false
	}
	if !prClient.IsPodMoveComplete(device.DeviceName, podList, pluginServer) {
		hwlog.RunLog.Warn("service pod has not been migrated or destroyed, wait for scanning again.")
		return false
	}
	return true
}

func (hdm *HwDevManager) execResetChip(logicID int32, isResetExec *bool) error {
	if *isResetExec {
		return nil
	}
	cardID, deviceID, err := hdm.manager.GetDmgr().GetCardIDDeviceID(logicID)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get cardID and deviceID by logicID(%d)", logicID)
		return err
	}
	if common.IsContainAtlas300IDuo() {
		deviceID = 0
	}
	hwlog.RunLog.Infof("start device card(%d) and deviceID(%d) reset...", cardID, deviceID)
	if err := hdm.manager.GetDmgr().SetDeviceReset(cardID, deviceID); err != nil {
		hwlog.RunLog.Errorf("hot reset failed, err: %v", err)
		return err
	}
	*isResetExec = true
	hwlog.RunLog.Infof("card(%d) and deviceID(%d) exec set device reset function success", cardID, deviceID)
	return nil
}

func (hdm *HwDevManager) subscribeFaultEvent() {
	hdm.subscribeNpuFaultEvent()
	hdm.subscribeSwitchFaultEvent()
}

func (hdm *HwDevManager) subscribeSwitchFaultEvent() {
	if common.ParamOption.RealCardType != api.Ascend910A3 || !common.ParamOption.EnableSwitchFault {
		return
	}
	for i := 0; i < common.GeneralSubscribeTime; i++ {
		if err := hdm.SwitchDevManager.SubscribeSwitchFaults(); err != nil {
			time.Sleep(time.Second)
			continue
		}
		return
	}
	common.SwitchSubscribeFailed = true
	hwlog.RunLog.Error("request Subscribe Switch FaultEvent failed, the subscribe way is closed")
}

// subscribeNpuFaultEvent subscribe fault happend on npus
func (hdm *HwDevManager) subscribeNpuFaultEvent() {
	if err := common.LoadFaultCodeFromFile(); err != nil {
		common.SubscribeFailed = true
		hwlog.RunLog.Errorf("load faultCode.json failed, the subscribe way is closed, err: %v", err)
		return
	}
	if hdm.RunMode != api.Ascend910 {
		hwlog.RunLog.Debug("subscribe mode only support 910 now")
		common.SubscribeFailed = true
		return
	}
	if err := hdm.manager.GetDmgr().SetFaultEventCallFunc(common.SaveDevFaultInfo); err != nil {
		common.SubscribeFailed = true
		hwlog.RunLog.Errorf("set fault event call back function failed, the subscribe way is closed, err: %v", err)
		return
	}
	for i := 0; i < common.GeneralSubscribeTime; i++ {
		if err := hdm.manager.GetDmgr().SubscribeDeviceFaultEvent(npuCommon.SubscribeAllDevice); err != nil {
			time.Sleep(time.Second)
			continue
		}
		return
	}
	common.SubscribeFailed = true
	hwlog.RunLog.Errorf("request SubscribeDeviceFaultEvent failed, the subscribe way is closed")
}

// graceTolerance start fault tolerance for training tasks
func (hdm *HwDevManager) graceTolerance(ctx context.Context, groupDevice map[string][]*common.NpuDevice) {
	hdm.manager.GraceTolerance(ctx, groupDevice)
	return
}

func (hdm *HwDevManager) isSupportGraceTolerance() {
	if common.ParamOption.HotReset != common.HotResetTrainOnLine &&
		common.ParamOption.HotReset != common.HotResetTrainOffLine {
		hwlog.RunLog.Debugf("train device hot reset mode error: %d", common.ParamOption.HotReset)
		return
	}

	if hdm.RunMode != api.Ascend910 {
		hwlog.RunLog.Debugf("grace tolerance only support training chip")
		return
	}
	if common.ParamOption.RealCardType == api.Ascend910A && hdm.WorkMode != common.SMPMode {
		hwlog.RunLog.Debug("grace tolerance only support SMP chip mode for 910")
		return
	}
	common.ParamOption.GraceToleranceOn = true
}

func (hdm *HwDevManager) pollFaultCodeCM(ctx context.Context) {
	var interval = common.PollFaultCodeCMInterval
	for {
		select {
		case _, ok := <-ctx.Done():
			if !ok {
				hwlog.RunLog.Info("stop signal chanel closed")
			}
			hwlog.RunLog.Info("poll fault code cm stop")
			return
		default:
			hwlog.RunLog.Debugf("polling '%s' configmap", common.FaultCodeCMName)
			configMap, err := hdm.manager.GetKubeClient().GetConfigMap(common.FaultCodeCMName, api.KubeNS)
			if err != nil {
				hwlog.RunLog.Debugf("cannot find '%s' configmap, reason: %v", common.FaultCodeCMName, err)
				initFaultInfoFromFile()
				time.Sleep(time.Duration(interval) * time.Second)
				continue
			}
			interval = getFaultCodeCMPollInterval(configMap)
			updateFaultConfigFromCm(configMap)
			time.Sleep(time.Duration(interval) * time.Second)
		}
	}
}

func updateFaultConfigFromCm(configMap *v1.ConfigMap) {
	if resourceVersion == configMap.ResourceVersion {
		return
	}
	hwlog.RunLog.Infof("detect '%s' configmap changed", common.FaultCodeCMName)
	resourceVersion = configMap.ResourceVersion
	loadFaultCode(configMap)
	if common.ParamOption.RealCardType == api.Ascend910A3 && common.ParamOption.EnableSwitchFault {
		loadSwitchFaultCode(configMap)
		deviceswitch.UpdateSwitchFaultLevel()
	}
	loadFaultCustomization(configMap)
	hwlog.RunLog.Infof("handling '%s' configmap change complete", common.FaultCodeCMName)
}

func initFaultInfoFromFile() {
	if err := common.LoadFaultCodeFromFile(); err != nil {
		hwlog.RunLog.Errorf("load fault code from file failed, err: %v", err)
	}
	if err := common.LoadFaultCustomizationFromFile(); err != nil {
		hwlog.RunLog.Errorf("load fault customization from file failed, err: %v", err)
	}
	if common.ParamOption.RealCardType == api.Ascend910A3 && common.ParamOption.EnableSwitchFault {
		if err := common.LoadSwitchFaultCodeFromFile(); err != nil {
			hwlog.RunLog.Errorf("load switch fault code from file failed, err: %v", err)
			return
		}
		deviceswitch.UpdateSwitchFaultLevel()
	}
}

func loadFaultCode(configMap *v1.ConfigMap) {
	faultCode, ok := configMap.Data[common.FaultCodeKey]
	if !ok {
		hwlog.RunLog.Errorf("cannot find key '%s' in CM, try to load faultCode.json", common.FaultCodeKey)
		if err := common.LoadFaultCodeFromFile(); err != nil {
			hwlog.RunLog.Errorf("load fault code from faultCode.json failed, err: %v", err)
			return
		}
		hwlog.RunLog.Infof("load fault code from faultCode.json success")
		return
	}
	if err := common.LoadFaultCode([]byte(faultCode)); err != nil {
		hwlog.RunLog.Errorf("load fault code from configmap failed, try to load faultCode.json, err: %v", err)
		if err = common.LoadFaultCodeFromFile(); err != nil {
			hwlog.RunLog.Errorf("load fault code from faultCode.json failed, err: %v", err)
			return
		}
		hwlog.RunLog.Infof("load fault code from faultCode.json success")
		return
	}
	hwlog.RunLog.Infof("load fault code from configmap success")
}

func loadSwitchFaultCode(configMap *v1.ConfigMap) {
	switchFaultCode, ok := configMap.Data[common.SwitchFaultCodeKey]
	if !ok {
		hwlog.RunLog.Errorf("cannot find key '%s' in CM, try to load SwitchFaultCode.json", common.SwitchFaultCodeKey)
		if err := common.LoadSwitchFaultCodeFromFile(); err != nil {
			hwlog.RunLog.Errorf("load switch fault code from SwitchFaultCode.json failed, err: %v", err)
			return
		}
		hwlog.RunLog.Info("load switch fault code from file success")
		return
	}
	if err := common.LoadSwitchFaultCode([]byte(switchFaultCode)); err != nil {
		hwlog.RunLog.Errorf("failed to load switch fault code from configmap, err: %s, "+
			"will try to load from file", err.Error())
		if err := common.LoadSwitchFaultCodeFromFile(); err != nil {
			hwlog.RunLog.Errorf("load switch fault code from SwitchFaultCode.json failed, err: %v", err)
			return
		}
		hwlog.RunLog.Info("load switch fault code from file success")
		return
	}
	hwlog.RunLog.Info("load switch fault code from configmap success")
}

func loadFaultCustomization(configMap *v1.ConfigMap) {
	faultCustomization, ok := configMap.Data[common.FaultCustomizationKey]
	if !ok {
		hwlog.RunLog.Warnf("did not find key(%s) in configmap, "+
			"reset fault customization", common.FaultCustomizationKey)
		common.ResetFaultCustomizationCache()
		if err := common.LoadFaultCustomizationFromFile(); err != nil {
			hwlog.RunLog.Errorf("load fault customization from faultCustomization.json failed, err: %v", err)
			return
		}
		hwlog.RunLog.Infof("load fault customization from faultCustomization.json success")
		return
	}
	if err := common.LoadFaultCustomization([]byte(faultCustomization)); err != nil {
		hwlog.RunLog.Errorf("load fault customization from cm failed, err: %v", err)
		common.ResetFaultCustomizationCache()
		if err = common.LoadFaultCustomizationFromFile(); err != nil {
			hwlog.RunLog.Errorf("load fault customization from faultCustomization.json failed, err: %v", err)
			return
		}
		hwlog.RunLog.Infof("Use default faultCustomization.json")
		return
	}
	hwlog.RunLog.Infof("load fault customization from configmap complete")
}

func getFaultCodeCMPollInterval(configMap *v1.ConfigMap) int {
	intervalStr, ok := configMap.Data[common.PollIntervalKey]
	if !ok {
		hwlog.RunLog.Infof("cannot find 'PollInterval', use default interval: %d", common.PollFaultCodeCMInterval)
		return common.PollFaultCodeCMInterval
	}
	interval, err := strconv.Atoi(intervalStr)
	if err != nil {
		hwlog.RunLog.Errorf("failed to parse 'PollInterval': %s, use default interval: %d", intervalStr,
			common.PollFaultCodeCMInterval)
		return common.PollFaultCodeCMInterval
	}
	if interval < common.PollFaultCodeCMMinInterval || interval > common.PollFaultCodeCMMaxInterval {
		hwlog.RunLog.Errorf("'PollInterval' exceed limit (%d~%d), 'PollInterval': %d, use default interval: %d",
			common.PollFaultCodeCMMinInterval, common.PollFaultCodeCMMaxInterval, interval,
			common.PollFaultCodeCMInterval)
		return common.PollFaultCodeCMInterval
	}
	return interval
}

func (hdm *HwDevManager) mendSubscribeFaultEvents() {
	initLogicIDs := common.GetAndCleanLogicID()
	for _, npuDevices := range hdm.groupDevice {
		for _, npuDevice := range npuDevices {
			if common.SubscribeFailed {
				hdm.manager.LogFaultModeChange(npuDevice, initLogicIDs, common.Polling)
			} else {
				hdm.manager.LogFaultModeChange(npuDevice, initLogicIDs, common.Subscribe)
			}

			hdm.manager.HandleDropCardFaultEvents(npuDevice)
			hdm.manager.HandleLostChipFaultEvents(npuDevice, initLogicIDs)
			hdm.manager.HandleLostNetworkFaultEvents(npuDevice, initLogicIDs)
		}
	}
}

func (hdm *HwDevManager) checkNodeResetInfo() {
	client := hdm.manager.GetKubeClient()
	if client == nil {
		return
	}
	resetInfo := device.ReadResetInfo()
	newResetInfo := device.ResetInfo{}
	newThirdPartyResetDevs, tpChanged := checkDeviceStatus(resetInfo.ThirdPartyResetDevs, hdm.groupDevice)
	newManualResetDevs, manChanged := checkDeviceStatus(resetInfo.ManualResetDevs, hdm.groupDevice)
	if !tpChanged && !manChanged {
		return
	}
	newResetInfo.ThirdPartyResetDevs = newThirdPartyResetDevs
	newResetInfo.ManualResetDevs = newManualResetDevs
	newResetInfo = checkOverRetryDev(newResetInfo)
	device.WriteResetInfo(newResetInfo, device.WMOverwrite, true)
}

func flattenMap(m map[string][]*common.NpuDevice) []*common.NpuDevice {
	var result []*common.NpuDevice
	for _, values := range m {
		for _, value := range values {
			result = append(result, value)
		}
	}
	return result
}

func checkDeviceStatus(failDevs []device.ResetDevice,
	groupDev map[string][]*common.NpuDevice) ([]device.ResetDevice, bool) {
	isChange := false
	var newDevs []device.ResetDevice = nil
	devMap := make(map[int32]*common.NpuDevice)
	for _, dev := range flattenMap(groupDev) {
		if dev.Health != v1beta1.Healthy {
			hwlog.RunLog.Debugf("device not recover, health %v, faultCode num %v", dev.Health,
				len(dev.FaultCodes))
			continue
		}
		devMap[dev.PhyID] = dev
		device.FreeBusyDev(dev.CardID, dev.DeviceID)
		// device recovered, set reset times to 0, then that device could be reset again
		device.SetResetCnt(dev.CardID, dev.DeviceID, 0)
	}
	for _, failDev := range failDevs {
		if _, exist := devMap[failDev.PhyID]; !exist {
			newDevs = append(newDevs, failDev)
			continue
		}

		isChange = true
	}
	return newDevs, isChange
}

func checkOverRetryDev(info device.ResetInfo) device.ResetInfo {
	ret := device.ResetInfo{
		ThirdPartyResetDevs: make([]device.ResetDevice, 0, len(info.ThirdPartyResetDevs)),
		ManualResetDevs:     info.ManualResetDevs,
	}
	for _, dev := range info.ThirdPartyResetDevs {
		if device.GetResetCnt(dev.CardId, dev.DeviceId) <= common.MaxResetTimes {
			ret.ThirdPartyResetDevs = append(ret.ThirdPartyResetDevs, dev)
			continue
		}
		ret.ManualResetDevs = append(ret.ManualResetDevs, dev)
	}
	return ret
}
