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

// Package device a series of device function
package device

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/containerd/containerd"
	"golang.org/x/time/rate"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	"k8s.io/utils/strings/slices"

	"Ascend-device-plugin/pkg/common"
	"Ascend-device-plugin/pkg/device/deviceswitch"
	"Ascend-device-plugin/pkg/kubeclient"
	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"ascend-common/devmanager"
	npuCommon "ascend-common/devmanager/common"
)

// isFirstFlushFault for device fault init
var (
	isFirstFlushFault      = true
	subscribeToPollingTime = common.DefaultSubscribeToPollingTime * time.Minute.Milliseconds()
	faultMode              = make(map[int32]string, common.GeneralMapSize)
	lastCheckNodeLabel     int64
	useIpv4                = true
	re                     = regexp.MustCompile(`"fault_time":\d+,`)
	allFaultInfo           = make(chan npuCommon.DevFaultInfo, common.WriteEventChanLenLimit)
	limiter                = rate.NewLimiter(
		rate.Every(time.Minute/common.WriteEventRateLimit), common.WriteEventRateLimit)
)

const (
	ipAddrTypeV4       = 0
	ipAddrTypeV6       = 1
	ipv6LinkTypePrefix = "fe80"

	checkNodeLabelPolling     = 60 * 60
	defaultUpdateTimeInterval = 5
)

// AscendTools struct definition
type AscendTools struct {
	client           *kubeclient.ClientK8s
	containerdClient *containerd.Client
	dmgr             devmanager.DeviceInterface
	name             string
	deviceUsage      string
	unHealthyKey     string
	devCount         int32
	healthDevice     sets.String
	boardId          uint32
	superPodID       int32
	serverIndex      int32
	// record map[device_logic_id]inresetting to show
	cardInResetMap  map[int32]bool
	cardInResetLock sync.Mutex
	// record map[device_logic_id]failed times
	resetFailedTimesMap       map[int32]int
	resetFailedTimesLock      sync.Mutex
	lastUpdateTimeStamp       time.Time
	lastManuallySeparateNPU   string
	lastSwitchFaultInfo       common.SwitchFaultInfo
	lastUsedChipsByProcess    sets.String
	lastUsedChipsContainerMap map[string]sets.String
}

// DevManager interface for manager device
type DevManager interface {
	GetNPUs() (common.NpuAllInfo, error)
	DoWithVolcanoListAndWatch(map[string][]*common.NpuDevice)
	GraceTolerance(context.Context, map[string][]*common.NpuDevice)
	SetDmgr(devmanager.DeviceInterface)
	GetDmgr() devmanager.DeviceInterface
	GetChipAICore() int32
	GetName() string
	SetKubeClient(*kubeclient.ClientK8s)
	GetKubeClient() *kubeclient.ClientK8s
	SetContainerdClient(*containerd.Client)
	GetContainerdClient() *containerd.Client
	UpdateHealth(map[string][]*common.NpuDevice, []*common.NpuDevice, string)
	GetChange(map[string][]*common.NpuDevice, map[string][]*common.NpuDevice) map[string]bool
	AddPodAnnotation(*common.PodDeviceInfo, string, string, []common.NpuDevice) error
	AppendVGroupInfo([]string)
	CheckDeviceTypeLabel() error
	CreateVirtualDevice(int32, string) (string, error)
	DestroyVirtualDevice(string) error
	GetChipAiCoreCount() (int32, error)
	SetDeviceUsage(int32) error
	GetDeviceUsage() string
	SetSuperPodID(superPodID int32)
	GetSuperPodID() int32
	SetServerIndex(serverIndex int32)
	GetServerIndex() int32
	GetServerBoardId(devLogicID int32) (uint32, error)
	SetCardsInResetting(int32, bool)
	GetIfCardsInResetting(int32) bool
	GetResetFailedTimes(int32) int
	SetResetFailedTimes(int32, int)
	HandleDropCardFaultEvents(*common.NpuDevice)
	HandleLostChipFaultEvents(*common.NpuDevice, []int32)
	HandleLostNetworkFaultEvents(*common.NpuDevice, []int32)
	LogFaultModeChange(*common.NpuDevice, []int32, string)
	GetUsedChips() sets.String
	GetDeviceIP(deviceType string, phyID int) (string, error)
	WriteFaultToEvent(ctx context.Context)
}

// SetDmgr set devmanager
func (tool *AscendTools) SetDmgr(dmgr devmanager.DeviceInterface) {
	tool.dmgr = dmgr
}

// GetDmgr get devmanager
func (tool *AscendTools) GetDmgr() devmanager.DeviceInterface {
	return tool.dmgr
}

// SetKubeClient set ClientK8s
func (tool *AscendTools) SetKubeClient(client *kubeclient.ClientK8s) {
	tool.client = client
}

// GetKubeClient get ClientK8s
func (tool *AscendTools) GetKubeClient() *kubeclient.ClientK8s {
	return tool.client
}

// SetContainerdClient set containerd Client
func (tool *AscendTools) SetContainerdClient(client *containerd.Client) {
	tool.containerdClient = client
}

// GetContainerdClient get containerd Client
func (tool *AscendTools) GetContainerdClient() *containerd.Client {
	return tool.containerdClient
}

// GetChipAICore get ai core
func (tool *AscendTools) GetChipAICore() int32 {
	return common.ParamOption.AiCoreCount
}

// GetName get chip name
func (tool *AscendTools) GetName() string {
	return tool.name
}

func (tool *AscendTools) convertLogicIDsToDeviceNames(logicIds []int32) string {
	deviceRunMode, err := common.GetDeviceRunMode()
	if err != nil {
		hwlog.RunLog.Warnf("failed to get device run mode, error: %v", err)
		return ""
	}
	deviceNamesSlice := make([]string, 0)
	for _, logicId := range logicIds {
		physicId, err := tool.GetDmgr().GetPhysicIDFromLogicID(logicId)
		if err != nil {
			hwlog.RunLog.Warnf("get physic id failed, err: %v", err)
			continue
		}
		deviceName := fmt.Sprintf("%s-%d", deviceRunMode, physicId)
		deviceNamesSlice = append(deviceNamesSlice, deviceName)
	}

	deviceNames := strings.Join(deviceNamesSlice, ",")

	return deviceNames
}

func (tool *AscendTools) handleManuallySeparateNPUFaultInfo() string {
	deviceRunMode, err := common.GetDeviceRunMode()
	if err != nil {
		hwlog.RunLog.Warnf("failed to get device run mode, error: %v", err)
		return ""
	}
	if manuallyFaultCache := common.QueryManuallyFaultNPULogicIDsByHandleStatus(common.
		ManuallySeparateNpuAll); len(manuallyFaultCache) == 0 {
		hwlog.RunLog.Debug("manually separate npu cache is empty, no need to handle manually separate npu " +
			"fault, the value of ManuallySeparateNPU field in device info configmap will be cleared")
		return ""
	}
	logicIDsHandledFromCache := common.QueryManuallyFaultNPULogicIDsByHandleStatus(common.ManuallySeparateNpuHandled)
	deviceInfoName := tool.client.DeviceInfoName
	physicIDsFromDeviceInfo := tool.client.
		GetManuallySeparateNPUIDFromDeviceInfo(deviceInfoName, api.KubeNS)
	for _, logicId := range logicIDsHandledFromCache {
		physicId, err := tool.GetDmgr().GetPhysicIDFromLogicID(logicId)
		if err != nil {
			hwlog.RunLog.Warnf("get physic id failed, err: %v", err)
			common.DeleteManuallyFaultInfo(logicId)
			continue
		}
		deviceName := fmt.Sprintf("%s-%d", deviceRunMode, physicId)

		if !common.Int32Tool.Contains(physicIDsFromDeviceInfo, physicId) {
			hwlog.RunLog.Infof("%s is not in ManuallySeparateNPU of device info configmap, will be removed in "+
				"cache", deviceName)
			common.DeleteManuallyFaultInfo(logicId)
		}
	}
	logicIDsAllFromCache := common.QueryManuallyFaultNPULogicIDsByHandleStatus(common.ManuallySeparateNpuAll)
	sort.Slice(logicIDsAllFromCache, func(i, j int) bool {
		return logicIDsAllFromCache[i] < logicIDsAllFromCache[j]
	})
	for _, physicId := range physicIDsFromDeviceInfo {
		logicId, err := tool.GetDmgr().GetLogicIDFromPhysicID(physicId)
		if err != nil {
			hwlog.RunLog.Warnf("get logic id failed, err: %v", err)
			continue
		}
		deviceName := fmt.Sprintf("%s-%d", deviceRunMode, physicId)
		if !common.Int32Tool.Contains(logicIDsAllFromCache, logicId) {
			hwlog.RunLog.Infof("cache does not contain %v, %v will be removed in ManuallySeparateNPU field "+
				"of device info configmap", deviceName, deviceName)
		}
	}
	common.SetManuallyFaultNPUHandled()
	manuallySeparateNPU := tool.convertLogicIDsToDeviceNames(logicIDsAllFromCache)
	return manuallySeparateNPU
}

// UpdateNodeDeviceInfo update device info
func (tool *AscendTools) UpdateNodeDeviceInfo(devStatusSet common.DevStatusSet,
	updateDeviceInfoFunc func(map[string]string, map[string]string, common.DevStatusSet) error) error {
	waitErr := wait.PollImmediate(common.Interval*time.Second, common.Timeout*time.Second, func() (bool, error) {
		nodeDeviceInfo := tool.GetKubeClient().GetDeviceInfoCMCache()
		deviceList := nodeDeviceInfo.DeviceInfo.DeviceList
		newDeviceList := common.MapDeepCopy(deviceList)
		if err := updateDeviceInfoFunc(deviceList, newDeviceList, devStatusSet); err != nil {
			hwlog.RunLog.Errorf("update device info failed, err: %v", err)
			return false, nil
		}
		tool.delVirDevInfo(newDeviceList)
		manuallySeparateNPU := tool.handleManuallySeparateNPUFaultInfo()
		// if subscribe failed, will use get interface
		if common.SwitchSubscribeFailed && common.ParamOption.EnableSwitchFault {
			var err error
			newFaults, err := deviceswitch.GetSwitchFaults()
			common.SetSwitchFaultCode(newFaults)
			if err != nil {
				hwlog.RunLog.Error("failed to query all fault codes of switch")
			}
		}
		switchFaultInfo := common.GetSwitchFaultInfo()
		if common.GetSyncMapLen(resetGoroutine) != 0 {
			common.UpdateSwitchFaultInfoAndFaultLevel(&switchFaultInfo)
		}
		dataSame := compareDeviceList(deviceList, newDeviceList) &&
			common.DeepEqualSwitchFaultInfo(switchFaultInfo, tool.lastSwitchFaultInfo) &&
			manuallySeparateNPU == tool.lastManuallySeparateNPU
		timeDiff := time.Now().Sub(tool.lastUpdateTimeStamp)
		hwlog.RunLog.Debugf("dataSame is %v, timeDiff is %v", dataSame, timeDiff)
		if dataSame && timeDiff < defaultUpdateTimeInterval*time.Minute {
			hwlog.RunLog.Debug("device info is not changed and " +
				"timeDiff less than 5 minutes, no need to update")
			return true, nil
		}
		if err := tool.client.WriteDeviceInfoDataIntoCMCache(newDeviceList, manuallySeparateNPU, switchFaultInfo,
			tool.GetSuperPodID(), tool.GetServerIndex()); err != nil {
			hwlog.RunLog.Errorf("write device info failed: %v", err)
			return false, nil
		}
		tool.lastUpdateTimeStamp = time.Now()
		tool.lastManuallySeparateNPU = manuallySeparateNPU
		tool.lastSwitchFaultInfo = switchFaultInfo
		hwlog.RunLog.Infof("write deviceInfo into configMap cache success, time is %s",
			tool.lastUpdateTimeStamp.Format(time.RFC3339))
		return true, nil
	})
	return waitErr
}

// compareDeviceList compare deviceList and newDeviceList are exactly the same without fault_time
func compareDeviceList(deviceList, newDeviceList map[string]string) bool {
	if deviceList == nil && newDeviceList == nil {
		return true
	}
	if deviceList == nil || newDeviceList == nil {
		return false
	}
	if len(deviceList) != len(newDeviceList) {
		return false
	}
	for key, value := range deviceList {
		val, exists := newDeviceList[key]
		if !exists {
			return false
		}
		str1 := re.ReplaceAllString(value, "")
		str2 := re.ReplaceAllString(val, "")
		if str1 != str2 {
			return false
		}
	}
	return true
}

func (tool *AscendTools) delVirDevInfo(newDeviceList map[string]string) {
	for annotationTag := range common.GetAllDeviceInfoTypeList() {
		if _, ok := newDeviceList[annotationTag]; !ok {
			continue
		}
		if common.IsVirtualDev(annotationTag) {
			delete(newDeviceList, annotationTag)
		}
	}
}

func (tool *AscendTools) assembleNpuDeviceStruct(deviType, deviceName string,
	davinCiDev common.DavinCiDev) common.NpuDevice {
	hwlog.RunLog.Debugf("Found Huawei Ascend, deviceType: %s, deviceName: %s", deviType, deviceName)
	return common.NpuDevice{
		DevType:       deviType,
		DeviceName:    deviceName,
		Health:        v1beta1.Healthy,
		NetworkHealth: v1beta1.Healthy,
		LogicID:       davinCiDev.LogicID,
		PhyID:         davinCiDev.PhyID,
		CardID:        davinCiDev.CardID,
		IP:            davinCiDev.IP,
		DeviceID:      davinCiDev.DeviceID,
	}
}

func (tool *AscendTools) assemblePhyDevices(davinCiDev common.DavinCiDev, devices *[]common.NpuDevice,
	deviceTypes *[]string) {
	deviceName := fmt.Sprintf("%s-%d", tool.name, davinCiDev.PhyID)
	device := tool.assembleNpuDeviceStruct(tool.name, deviceName, davinCiDev)
	*deviceTypes = append(*deviceTypes, tool.name)
	*devices = append(*devices, device)
}

func (tool *AscendTools) assembleVirtualDevices(davinCiDev common.DavinCiDev, vDevInfos npuCommon.VirtualDevInfo,
	devices *[]common.NpuDevice, vDeviceTypes *[]string) {
	for _, subVDevInfo := range vDevInfos.VDevInfo {
		vDeviType, deviceName, err := tool.assembleSpecVirtualDevice(davinCiDev.PhyID, subVDevInfo)
		if err != nil {
			hwlog.RunLog.Error(err)
			continue
		}
		device := tool.assembleNpuDeviceStruct(vDeviType, deviceName, davinCiDev)
		*devices = append(*devices, device)
		*vDeviceTypes = append(*vDeviceTypes, vDeviType)
	}
}

func (tool *AscendTools) assembleSpecVirtualDevice(phyID int32, vDevInfo npuCommon.CgoVDevQueryStru) (string,
	string, error) {
	coreNum := int32(vDevInfo.QueryInfo.Computing.Aic)
	if coreNum <= 0 {
		return "", "", fmt.Errorf("invalid vdev info, ai core is 0")
	}
	vDeviType, exist := common.GetTemplateName2DeviceTypeMap()[vDevInfo.QueryInfo.Name]
	if !exist {
		return "", "", fmt.Errorf("check templatename failed, templatename is %s", vDevInfo.QueryInfo.Name)
	}
	vDeviType = fmt.Sprintf("%s-%s", tool.name, vDeviType)
	devID := fmt.Sprintf("%s-%d-%d", vDeviType, vDevInfo.VDevID, phyID)
	return vDeviType, devID, nil
}

func (tool *AscendTools) assemble310PMixedPhyDevices(davinCiDev common.DavinCiDev, devices *[]common.NpuDevice,
	deviceTypes *[]string) error {
	cardID, deviceID, err := tool.dmgr.GetCardIDDeviceID(davinCiDev.LogicID)
	if err != nil {
		return fmt.Errorf("get cardID and deviceID failed: LogicID[%v]", davinCiDev.LogicID)
	}
	productType, err := tool.dmgr.GetProductType(cardID, deviceID)
	if err != nil {
		return fmt.Errorf("get product type failed:cardID[%v] deviceID[%v]", cardID, deviceID)
	}
	ProductTypeMap := common.Get310PProductType()
	if _, ok := ProductTypeMap[productType]; !ok {
		return fmt.Errorf("%v not found", productType)
	}
	deviceName := fmt.Sprintf("%s-%d", ProductTypeMap[productType], davinCiDev.PhyID)
	device := tool.assembleNpuDeviceStruct(ProductTypeMap[productType], deviceName, davinCiDev)
	*deviceTypes = append(*deviceTypes, ProductTypeMap[productType])
	*devices = append(*devices, device)
	return nil
}

func (tool *AscendTools) removeDuplicate(allDeviceTypes *[]string) []string {
	deviceTypesMap := make(map[string]string, len(*allDeviceTypes))
	var rmDupDeviceTypes []string
	for _, deviType := range *allDeviceTypes {
		deviceTypesMap[deviType] = deviType
	}
	for _, deviType := range deviceTypesMap {
		rmDupDeviceTypes = append(rmDupDeviceTypes, deviType)
	}
	return rmDupDeviceTypes
}

func getResetInfoData(resetInfo *v1.ConfigMap) ([]*common.TaskDevInfo, error) {
	data, ok := resetInfo.Data[common.ResetInfoCMDataKey]
	if !ok {
		return nil, fmt.Errorf("%s not exist", common.ResetInfoCMDataKey)
	}
	if len(data) > common.CMDataMaxLength {
		return nil, fmt.Errorf("configmap data size is out of memory")
	}
	var taskResetInfo common.TaskResetInfo
	if err := json.Unmarshal([]byte(data), &taskResetInfo); err != nil {
		return nil, fmt.Errorf("unmarshal configmap data failed, err: %v", err)
	}
	if taskResetInfo.UpdateTime == 0 {
		hwlog.RunLog.Debugf("reset configmap is initializing")
		return nil, nil
	}
	return taskResetInfo.RankList, nil
}

func (tool *AscendTools) getRealUsedDevices() sets.String {
	podList := tool.client.GetActivePodListCache()
	usedDevice := sets.String{}
	for _, pod := range podList {
		realDevice, exist := pod.Annotations[api.ResourceNamePrefix+common.PodRealAlloc]
		if !exist {
			continue
		}
		usedDevice.Insert(strings.Split(realDevice, common.CommaSepDev)...)
	}
	return usedDevice
}

func (tool *AscendTools) getDevStatesDevSet(classifyDevs map[string][]*common.NpuDevice) common.DevStatusSet {
	totalFreeDevices := make(map[string]sets.String, len(classifyDevs))
	totalUHDevices, totalNetUHDevices, podUsedDev, totalRCDevices :=
		sets.String{}, sets.String{}, sets.String{}, sets.String{}
	totalDeviceFaults := make([]common.DeviceFault, 0, common.GeneralMapSize)
	allDevices := sets.String{}
	if !common.ParamOption.PresetVDevice {
		podUsedDev = tool.getRealUsedDevices()
	}
	for devType, classifyDev := range classifyDevs {
		partDevStatusSet := tool.groupDevsByStatus(classifyDev, tool.name)
		allDevices = allDevices.Union(partDevStatusSet.AllDevices)
		kltUsedDevices := tool.getUsedDevices(classifyDev)
		totalFreeDevices[devType] = partDevStatusSet.HealthDevices.Difference(kltUsedDevices)
		if !common.ParamOption.PresetVDevice {
			totalFreeDevices[devType] = totalFreeDevices[devType].Difference(podUsedDev)
		}
		totalUHDevices = totalUHDevices.Union(partDevStatusSet.UnHealthyDevice)
		totalNetUHDevices = totalNetUHDevices.Union(partDevStatusSet.NetUnHealthyDevice)
		totalRCDevices = totalRCDevices.Union(partDevStatusSet.RecoveringDevices)
		totalDeviceFaults = append(totalDeviceFaults, partDevStatusSet.DeviceFault...)
	}
	devStatusSet := common.DevStatusSet{
		FreeHealthyDevice:  totalFreeDevices,
		UnHealthyDevice:    totalUHDevices,
		NetUnHealthyDevice: totalNetUHDevices,
		RecoveringDevices:  totalRCDevices,
		DeviceFault:        totalDeviceFaults,
		AllDevices:         allDevices,
	}
	hwlog.RunLog.Debugf("get device status devStatusSet=%#v", devStatusSet)
	return devStatusSet
}

func (tool *AscendTools) getUsedDevices(subClassDevices []*common.NpuDevice) sets.String {
	usedDevices := make([]string, 0, len(subClassDevices))
	for _, device := range subClassDevices {
		if device.PodUsed {
			usedDevices = append(usedDevices, device.DeviceName)
		}
	}
	return sets.NewString(usedDevices...)
}

func (tool *AscendTools) groupDevsByStatus(subClassDevices []*common.NpuDevice, runMode string) common.DevStatusSet {
	healthDevice, totalUHDevices, totalNetworkUHDevices, totalRCDevices :=
		sets.String{}, sets.String{}, sets.String{}, sets.String{}
	deviceFaults := make([]common.DeviceFault, 0, common.GeneralMapSize)
	allDevices := sets.NewString()
	for _, device := range subClassDevices {
		allDevices.Insert(device.DeviceName)
		deviceFaults = append(deviceFaults, tool.getDeviceFaults(device)...)
		if device.NetworkHealth == v1beta1.Unhealthy {
			totalNetworkUHDevices.Insert(device.DeviceName)
		}
		if device.Status == common.NPUResettingStatus {
			totalRCDevices.Insert(device.DeviceName)
		}
		if device.Health == v1beta1.Healthy {
			healthDevice.Insert(device.DeviceName)
			continue
		}
		if !common.IsVirtualDev(device.DeviceName) {
			totalUHDevices.Insert(device.DeviceName)
			continue
		}
		if dev := fmt.Sprintf("%s-%d", runMode, device.PhyID); !totalUHDevices.Has(dev) {
			totalUHDevices.Insert(dev)
		}
	}
	hwlog.RunLog.Debugf("healthy device %#v", healthDevice)
	hwlog.RunLog.Debugf("total unhealthy devices %#v", totalUHDevices)
	hwlog.RunLog.Debugf("total network unhealthy devices %#v", totalNetworkUHDevices)
	hwlog.RunLog.Debugf("total recovering devices %#v", totalRCDevices)
	hwlog.RunLog.Debugf("device fault list is %#v", deviceFaults)
	return common.DevStatusSet{
		HealthDevices:      healthDevice,
		UnHealthyDevice:    totalUHDevices,
		NetUnHealthyDevice: totalNetworkUHDevices,
		RecoveringDevices:  totalRCDevices,
		DeviceFault:        deviceFaults,
		AllDevices:         allDevices,
	}
}

func (tool *AscendTools) getFaultTimeAndLevelMap(
	device *common.NpuDevice, allFaults map[int64]common.FaultTimeAndLevel,
	isNetworkFault bool) map[string]common.FaultTimeAndLevel {
	result := make(map[string]common.FaultTimeAndLevel)
	var events []int64
	var getFaultLevelFunc func(events []int64, logicId int32) string
	if isNetworkFault {
		events = device.NetworkFaultCodes
		getFaultLevelFunc = common.GetNetworkFaultType
	} else {
		events = device.FaultCodes
		getFaultLevelFunc = common.GetFaultType
	}
	for _, eventId := range events {
		faultLevel := getFaultLevelFunc([]int64{eventId}, device.LogicID)
		faultTime, found := device.FaultTimeMap[eventId]
		if !found {
			hwlog.RunLog.Warnf("fault time map is inconsistance with faults, map: %s, codes: %s",
				common.ObjToString(device.FaultTimeMap), common.ObjToString(events))
		}
		faultTimeAndLevel := common.FaultTimeAndLevel{
			FaultTime:  faultTime,
			FaultLevel: faultLevel,
		}
		hexFaultCode := strings.ToUpper(strconv.FormatInt(eventId, common.Hex))
		result[hexFaultCode] = faultTimeAndLevel
	}
	for eventId, timeAndLevel := range allFaults {
		hexFaultCode := strings.ToUpper(strconv.FormatInt(eventId, common.Hex))
		result[hexFaultCode] = timeAndLevel
	}
	return result
}

func (tool *AscendTools) combineFaultTimeMaps(timeoutFaultLevelAndTime,
	frequencyFaultLevelAndTime map[int64]common.FaultTimeAndLevel) map[int64]common.FaultTimeAndLevel {
	combineMap := make(map[int64]common.FaultTimeAndLevel)
	for key, value := range timeoutFaultLevelAndTime {
		combineMap[key] = value
	}
	for key, value := range frequencyFaultLevelAndTime {
		combineMap[key] = value
	}
	return combineMap
}

// getDeviceFaults get device fault list
func (tool *AscendTools) getDeviceFaults(device *common.NpuDevice) []common.DeviceFault {
	deviceFaults := make([]common.DeviceFault, 0, common.MapSizeTwo)
	if len(device.NetworkFaultCodes) != 0 || device.NetworkHealth == v1beta1.Unhealthy {
		deviceFaults = tool.getDeviceFaultsWithMode(device, device.NetworkFaultCodes, deviceFaults,
			common.NetworkFaultMode, common.CardNetworkUnhealthy)
	}
	if len(device.FaultCodes) != 0 || device.Health == v1beta1.Unhealthy {
		deviceFaults = tool.getDeviceFaultsWithMode(device, device.FaultCodes, deviceFaults,
			common.ChipFaultMode, common.CardUnhealthy)
	}
	return deviceFaults
}

func (tool *AscendTools) getDeviceFaultsWithMode(device *common.NpuDevice, faultCodes []int64,
	deviceFaults []common.DeviceFault, mode string, unhealthyType string) []common.DeviceFault {
	timeoutFaultLevelAndTime := common.GetTimeoutFaultLevelAndCodes(mode, device.LogicID)
	frequencyFaultLevelAndTime := common.GetFrequencyFaultLevelAndCodes(mode, device.LogicID)
	allFaultLevelAndTime := tool.combineFaultTimeMaps(timeoutFaultLevelAndTime, frequencyFaultLevelAndTime)
	allFaultCodes := append(faultCodes, common.Keys(allFaultLevelAndTime)...)
	newCode := tool.removeDuplicateErr(allFaultCodes)
	var faultType = ""
	var isNetworkFault bool
	if mode == common.NetworkFaultMode {
		faultType = common.GetNetworkFaultType(faultCodes, device.LogicID)
		isNetworkFault = true
	}
	if mode == common.ChipFaultMode {
		faultType = common.GetFaultType(faultCodes, device.LogicID)
		isNetworkFault = false
	}
	deviceFaults = append(deviceFaults, common.DeviceFault{
		FaultType:            unhealthyType,
		NPUName:              device.DeviceName,
		LargeModelFaultLevel: faultType,
		FaultLevel:           faultType,
		FaultHandling:        faultType,
		FaultCode:            strings.ToUpper(common.Int64Tool.ToHexString(newCode)),
		FaultTimeAndLevelMap: tool.getFaultTimeAndLevelMap(device, allFaultLevelAndTime, isNetworkFault),
	})
	return deviceFaults
}

func (tool *AscendTools) removeDuplicateErr(faultCodes []int64) []int64 {
	isExist := make(map[int64]bool, common.GeneralMapSize)
	newCode := make([]int64, 0)
	for _, code := range faultCodes {
		if _, exist := isExist[code]; !exist {
			newCode = append(newCode, code)
			isExist[code] = true
		}
	}
	return newCode
}

func (tool *AscendTools) getDavinCiDev(logicID int32) (common.DavinCiDev, error) {
	phyID, err := tool.dmgr.GetPhysicIDFromLogicID(logicID)
	if err != nil {
		return common.DavinCiDev{}, err
	}
	cardID, deviceID, err := tool.dmgr.GetCardIDDeviceID(logicID)
	if err != nil {
		return common.DavinCiDev{}, err
	}
	ip, err := tool.GetDeviceIP("", int(phyID))
	if err != nil {
		hwlog.RunLog.Warnf("get device ip failed, err: %v", err)
		ip = ""
	}
	return common.DavinCiDev{
		LogicID:  logicID,
		PhyID:    phyID,
		CardID:   cardID,
		IP:       ip,
		DeviceID: deviceID,
	}, nil
}

func (tool *AscendTools) getVirtualDevice(logicID int32) (npuCommon.VirtualDevInfo, error) {
	virtualDevInfos, err := tool.dmgr.GetVirtualDeviceInfo(logicID)
	if err != nil {
		return npuCommon.VirtualDevInfo{}, fmt.Errorf("query virtual device info failure: %s", err)
	}
	return virtualDevInfos, nil
}

// GetDeviceIP get device ip
func (tool *AscendTools) GetDeviceIP(deviceType string, phyID int) (string, error) {
	if common.IsVirtualDev(deviceType) {
		return common.DefaultDeviceIP, nil
	}
	logicID, err := tool.dmgr.GetLogicIDFromPhysicID(int32(phyID))
	if err != nil {
		return "", fmt.Errorf("transfor phyID %d to logicID failed, err: %v", phyID, err)
	}
	chip, err := tool.dmgr.GetChipInfo(logicID)
	if err != nil {
		return "", fmt.Errorf("get logicID %d chip info failed, err: %v", logicID, err)
	}
	if strings.Contains(chip.Name, common.VirMark) {
		return common.DefaultDeviceIP, nil
	}
	return tool.getDcmiDeviceIP(logicID)
}

func (tool *AscendTools) getDcmiDeviceIP(logicID int32) (string, error) {
	var deviceIp string
	var err error
	if useIpv4 {
		if deviceIp, err = tool.dmgr.GetDeviceIPAddress(logicID, ipAddrTypeV4); err == nil {
			return deviceIp, nil
		}
		useIpv4 = false
	}

	if !useIpv4 {
		deviceIp, err = tool.dmgr.GetDeviceIPAddress(logicID, ipAddrTypeV6)
		if err != nil {
			useIpv4 = true
			return "", err
		}

		if strings.Index(deviceIp, ipv6LinkTypePrefix) == 0 {
			return "", fmt.Errorf("logicID(%d) device ip %v is a link type ipv6 address", logicID, deviceIp)
		}
	}

	return deviceIp, nil
}

func (tool *AscendTools) getDeviceListIP(devices []string, deviceType string) (map[int]string, error) {
	ascendRuntimeOptions := ""
	if common.IsVirtualDev(deviceType) {
		ascendRuntimeOptions = common.VirtualDev
	}
	_, ascendDevices, err := common.GetDeviceListID(devices, ascendRuntimeOptions)
	if err != nil {
		hwlog.RunLog.Errorf("get device list id err: %v", err)
		return nil, err
	}
	devicesWithIP := make(map[int]string, len(devices))
	for _, id := range ascendDevices {
		if tool.deviceUsage == common.Infer {
			devicesWithIP[id] = ""
			continue
		}
		deviceIP, err := tool.GetDeviceIP(deviceType, id)
		if err != nil {
			hwlog.RunLog.Errorf("get device %d ip err: %v", id, err)
			return nil, err
		}
		devicesWithIP[id] = deviceIP
	}
	return devicesWithIP, nil
}

func (tool *AscendTools) getConfigAnno(podDev *common.PodDeviceInfo, deviceType, serverID string,
	allDevices []common.NpuDevice) (string, error) {
	ascendRuntimeOptions := ""
	if common.IsVirtualDev(deviceType) {
		ascendRuntimeOptions = common.VirtualDev
	}
	phyDevMapVirtualDev, _, err := common.GetDeviceListID(podDev.RealDevice, ascendRuntimeOptions)
	if err != nil {
		hwlog.RunLog.Errorf("get device list id failed, error: %v", err)
		return "", errors.New("get device list id failed")
	}
	ascendVisibleDevices, err := tool.getDeviceListIP(podDev.RealDevice, deviceType)
	if err != nil {
		hwlog.RunLog.Errorf("get device list ip failed, error: %v", err)
		return "", errors.New("get device list ip failed")
	}
	info := common.ServerInfo{
		ServerID:   serverID,
		DeviceType: deviceType,
		SuperPodID: tool.GetSuperPodID(),
	}
	configuration := common.GetPodConfiguration(phyDevMapVirtualDev, ascendVisibleDevices,
		podDev.Pod.Name, info, allDevices)
	return configuration, nil
}

// AddPodAnnotation check and update pod annotations
// correct annotation 'AscendReal', 'Ascend910', 'kltDev', 'ascend-910-configuration' per 5s
func (tool *AscendTools) AddPodAnnotation(podDev *common.PodDeviceInfo, deviceType, serverID string,
	allDevices []common.NpuDevice) error {
	if !common.ParamOption.PresetVDevice {
		tool.AppendVGroupInfo(podDev.RealDevice)
	}
	sort.Strings(podDev.KltDevice)
	sort.Strings(podDev.RealDevice)
	allUsedDev := strings.Join(podDev.RealDevice, common.CommaSepDev)
	annoChecker := []struct {
		annoKey   string
		annoValue string
	}{
		{annoKey: api.ResourceNamePrefix + common.Pod2kl, annoValue: strings.Join(podDev.KltDevice, common.CommaSepDev)},
		{annoKey: api.ResourceNamePrefix + common.PodRealAlloc, annoValue: allUsedDev},
		{annoKey: fmt.Sprintf("%s%s", api.ResourceNamePrefix, deviceType), annoValue: allUsedDev},
	}
	annotation := make(map[string]string)
	if !common.IsVirtualDev(deviceType) {
		for _, checker := range annoChecker {
			// dynamic scene, not correct huawei.com/npu-core
			if deviceType == common.AiCoreResourceName && checker.annoKey ==
				fmt.Sprintf("%s%s", api.ResourceNamePrefix, deviceType) {
				continue
			}
			if podDev.Pod.Annotations[checker.annoKey] != checker.annoValue {
				hwlog.RunLog.Warnf("need correct: annotKey: %s, old value: %s, new value: %s",
					checker.annoKey, podDev.Pod.Annotations[checker.annoKey], checker.annoValue)
				annotation[checker.annoKey] = checker.annoValue
			}
		}
	}
	if tool.name == common.Ascend910 || common.IsContainAll300IDuo() {
		config, err := tool.getConfigAnno(podDev, deviceType, serverID, allDevices)
		if err == nil {
			if podDev.Pod.Annotations[api.Pod910DeviceAnno] != config {
				hwlog.RunLog.Warnf("need correct: annotKey: %s, old value: %s, new value: %s",
					api.Pod910DeviceAnno, podDev.Pod.Annotations[api.Pod910DeviceAnno], config)
				annotation[api.Pod910DeviceAnno] = config
			}
		}
	}
	if len(annotation) == 0 {
		hwlog.RunLog.Debug("annotation do not need to be updated")
		return nil
	}
	if err := tool.client.TryUpdatePodAnnotation(&podDev.Pod, annotation); err != nil {
		hwlog.RunLog.Errorf("correct pod[%s:%s] annotation failed, error: %v", podDev.Pod.Namespace, podDev.Pod.Name, err)
		return err
	}
	hwlog.RunLog.Infof("correct pod[%s:%s] annotation success", podDev.Pod.Namespace, podDev.Pod.Name)
	return nil
}

// UpdateHealth update group device healthy
func (tool *AscendTools) UpdateHealth(groupDevice map[string][]*common.NpuDevice,
	aiCoreDevs []*common.NpuDevice, runMode string) {
	// update health of device
	tool.writeNewFaultCode(groupDevice, runMode)

	setHealthyIfDuoCard(groupDevice)
	setAICoreHealthyIfVNpu(groupDevice, aiCoreDevs)
}

// GetChange check if groupDevice changes
func (tool *AscendTools) GetChange(groupDevice, oldGroupDevice map[string][]*common.NpuDevice) map[string]bool {
	isStateChange := make(map[string]bool, len(groupDevice))
	for devType, devices := range groupDevice {
		isStateChange[devType] = false
		for idx, device := range devices {
			if device.Health != oldGroupDevice[devType][idx].Health {
				isStateChange[devType] = true
			}
		}
	}
	return isStateChange
}

func setAICoreHealthyIfVNpu(groupDevice map[string][]*common.NpuDevice, aiCoreDevs []*common.NpuDevice) {
	if common.ParamOption.PresetVDevice {
		return
	}
	logicDeviceMap := make(map[int32]*common.NpuDevice, common.GeneralMapSize)
	for _, devices := range groupDevice {
		for _, device := range devices {
			logicDeviceMap[device.LogicID] = device
		}
	}
	for _, device := range aiCoreDevs {
		device.Health = logicDeviceMap[device.LogicID].Health
		device.NetworkHealth = logicDeviceMap[device.LogicID].NetworkHealth
		device.FaultCodes = logicDeviceMap[device.LogicID].FaultCodes
		device.AlarmRaisedTime = logicDeviceMap[device.LogicID].AlarmRaisedTime
		device.NetworkFaultCodes = logicDeviceMap[device.LogicID].NetworkFaultCodes
		device.NetworkAlarmRaisedTime = logicDeviceMap[device.LogicID].NetworkAlarmRaisedTime
	}
}

func setHealthyIfDuoCard(groupDevice map[string][]*common.NpuDevice) {
	if !common.IsContainAtlas300IDuo() {
		return
	}
	if common.ParamOption.HotReset != common.HotResetInfer {
		hwlog.RunLog.Debugf("not open infer device hot reset function, it's %d", common.ParamOption.HotReset)
		return
	}
	ascend310PDevices, ok := groupDevice[common.Ascend310P]
	if !ok {
		hwlog.RunLog.Debugf("not found 310P devices")
		return
	}
	unHealthyCards := getUnHealthyCard(ascend310PDevices)
	for _, device := range ascend310PDevices {
		if _, ok := unHealthyCards[device.CardID]; ok {
			device.Health = v1beta1.Unhealthy
		}
	}
}

func getUnHealthyCard(ascend310PDevices []*common.NpuDevice) map[int32]int8 {
	unHealthyCards := make(map[int32]int8, len(ascend310PDevices))
	for _, device := range ascend310PDevices {
		if device.Health == v1beta1.Healthy {
			continue
		}
		unHealthyCards[device.CardID] = 0
	}
	return unHealthyCards
}

// ClassifyDevices classify diff type devices
func ClassifyDevices(allDevs []common.NpuDevice, devTypes []string) map[string][]*common.NpuDevice {
	var classifyMap = make(map[string][]*common.NpuDevice, len(devTypes))
	for _, suffix := range devTypes {
		classifyMap[suffix] = classifyDevByType(allDevs, suffix)
	}
	return classifyMap
}

func classifyDevByType(allDevs []common.NpuDevice, suffix string) []*common.NpuDevice {
	var classifyDev []*common.NpuDevice
	for index, device := range allDevs {
		if device.DevType == suffix {
			classifyDev = append(classifyDev, &allDevs[index])
		}
	}
	return classifyDev
}

func (tool *AscendTools) isHealthy(device *common.NpuDevice) string {
	faultType := common.GetFaultType(device.FaultCodes, device.LogicID)
	if faultType == common.NormalNPU || faultType == common.NotHandleFault || faultType == common.SubHealthFault ||
		(faultType == common.FreeRestartNPU &&
			tool.npuIsUsedNow(device.DeviceName) && common.ParamOption.GraceToleranceOn == true) {
		return v1beta1.Healthy
	}
	if faultType == common.PreSeparateNPU && tool.npuIsUsedNow(device.DeviceName) {
		hwlog.RunLog.Infof("detect %s but device is used, device name: %s", faultType, device.DeviceName)
		return v1beta1.Healthy
	}
	return v1beta1.Unhealthy
}

func (tool *AscendTools) isNetworkHealthy(device *common.NpuDevice) string {
	faultType := common.GetNetworkFaultType(device.NetworkFaultCodes, device.LogicID)
	if faultType == common.NormalNPU || faultType == common.NotHandleFault {
		return v1beta1.Healthy
	}

	return v1beta1.Unhealthy
}

func (tool *AscendTools) npuIsUsedNow(deviceName string) bool {
	podList := tool.client.GetActivePodListCache()
	for _, pod := range podList {
		annotationTag := fmt.Sprintf("%s%s", api.ResourceNamePrefix, common.Ascend910)
		tmpNpu, ok := pod.Annotations[annotationTag]
		if !ok || len(tmpNpu) == 0 || len(tmpNpu) > common.PodAnnotationMaxLength {
			continue
		}
		deviceStrList := strings.Split(tmpNpu, common.CommaSepDev)
		if slices.Index(deviceStrList, deviceName) != -1 {
			return true
		}
	}
	return false
}

func (tool *AscendTools) getVGroupID(device string) (uint32, error) {
	phyID, virID, err := common.GetDeviceID(device, common.VirtualDev)
	if err != nil {
		return 0, err
	}
	logicID, err := tool.dmgr.GetLogicIDFromPhysicID(int32(phyID))
	if err != nil {
		return 0, err
	}
	virtualDevInfos, err := tool.dmgr.GetVirtualDeviceInfo(logicID)
	if err != nil {
		return 0, fmt.Errorf("query virtual device info failure: %s", err)
	}
	for _, vDevInfo := range virtualDevInfos.VDevInfo {
		if vDevInfo.VDevID == uint32(virID) {
			return vDevInfo.QueryInfo.Base.VfgID, nil
		}
	}
	return 0, fmt.Errorf("not found virutal device info, %s", device)
}

// AppendVGroupInfo append virtual group id info after device name
func (tool *AscendTools) AppendVGroupInfo(allocateDevice []string) {
	hwlog.RunLog.Debugf("allocateDevice:%v", allocateDevice)
	for i, device := range allocateDevice {
		if !common.IsVirtualDev(device) {
			continue
		}
		vGroupID, err := tool.getVGroupID(device)
		if err != nil {
			hwlog.RunLog.Warn(err)
			continue
		}
		allocateDevice[i] = fmt.Sprintf("%s%s%d", device, common.UnderLine, vGroupID)
	}
}

// CheckDeviceTypeLabel check device type label
func (tool *AscendTools) CheckDeviceTypeLabel() error {
	if time.Now().Unix()-lastCheckNodeLabel < checkNodeLabelPolling {
		return nil
	}
	curNode, err := tool.client.GetNode()
	if err != nil {
		return err
	}
	deviceType, exist := curNode.Labels[common.ServerTypeLabelKey]
	if !exist {
		return fmt.Errorf("label of %s not exist", common.ServerTypeLabelKey)
	}
	deviceTypeInfos := strings.Split(deviceType, common.MiddelLine)
	if len(deviceTypeInfos) < common.ServerTypeInfoMinLen {
		return fmt.Errorf("length of device type info %d is invalid", len(deviceTypeInfos))
	}
	if !strings.HasPrefix(deviceTypeInfos[0], tool.name) {
		return fmt.Errorf("label chip name %s is not meet real chip name %s", deviceTypeInfos[0], tool.name)
	}
	aiCore, err := strconv.Atoi(deviceTypeInfos[1])
	if err != nil {
		return fmt.Errorf("covert label ai core failed, error is %v", err)
	}
	if aiCore != int(common.ParamOption.AiCoreCount) {
		return fmt.Errorf("label ai core %d not equal real chip ai core %d", aiCore, common.ParamOption.AiCoreCount)
	}
	return nil
}

// CreateVirtualDevice create virtual device
func (tool *AscendTools) CreateVirtualDevice(phyID int32, templateName string) (string, error) {
	createInfo := npuCommon.CgoCreateVDevRes{
		VDevID:       common.DefaultIDForCreateVNPU,
		VfgID:        common.DefaultIDForCreateVNPU,
		TemplateName: templateName,
	}
	logicID, err := tool.dmgr.GetLogicIDFromPhysicID(phyID)
	if err != nil {
		return "", err
	}
	createOut, err := tool.dmgr.CreateVirtualDevice(logicID, createInfo)
	if err != nil {
		hwlog.RunLog.Error(err)
		return "", fmt.Errorf(common.NPUSegmentFailed)
	}
	hwlog.RunLog.Infof("create %s from device %d success", createInfo.TemplateName, phyID)
	vDevType, exist := common.GetTemplateName2DeviceTypeMap()[templateName]
	if !exist {
		return "", fmt.Errorf("check templatename failed, templatename is %s", templateName)
	}
	vDevName := fmt.Sprintf("%s-%s-%d-%d", tool.name, vDevType, createOut.VDevID, phyID)
	return vDevName, nil
}

// DestroyVirtualDevice destroy virtual device
func (tool *AscendTools) DestroyVirtualDevice(deviceName string) error {
	phyID, virID, err := common.GetDeviceID(deviceName, common.VirtualDev)
	if err != nil {
		return fmt.Errorf("get device id failed, %v", err)
	}
	logicID, err := tool.dmgr.GetLogicIDFromPhysicID(int32(phyID))
	if err != nil {
		return err
	}
	for i := 0; i < common.RetryUpdateCount; i++ {
		if err = tool.dmgr.DestroyVirtualDevice(logicID, uint32(virID)); err == nil {
			break
		}
		time.Sleep(time.Second)
	}
	return err
}

// GetChipAiCoreCount get chip aicore count
func (tool *AscendTools) GetChipAiCoreCount() (int32, error) {
	_, logicIDs, err := tool.dmgr.GetDeviceList()
	if err != nil {
		return 0, err
	}
	if len(logicIDs) < 1 {
		return 0, fmt.Errorf("not found logicIDs")
	}
	for _, logicID := range logicIDs {
		cgoVDevInfo, err := tool.dmgr.GetVirtualDeviceInfo(logicID)
		if err != nil && strings.Contains(err.Error(), strconv.Itoa(common.DeviceNotSupport)) {
			return common.DeviceNotSupport, nil
		}
		if err != nil {
			// if not support found aicore number, setting a default value
			hwlog.RunLog.Infof("not found aicore number by dcmi: %v", err)
			return common.DefaultAiCoreNum, nil
		}
		return tool.getAiCoreCount(cgoVDevInfo)
	}
	return 0, fmt.Errorf("not get aicore count")
}

func (tool *AscendTools) getAiCoreCount(cgoVDevInfo npuCommon.VirtualDevInfo) (int32, error) {
	chipAICore := cgoVDevInfo.TotalResource.Computing.Aic
	if chipAICore < common.MinAICoreNum || chipAICore > common.MaxAICoreNum {
		return 0, fmt.Errorf("invalid ai core num %f", chipAICore)
	}
	return int32(chipAICore), nil
}

// writeNewFaultCode writes fault code and health to device
func (tool *AscendTools) writeNewFaultCode(deviceMap map[string][]*common.NpuDevice, runMode string) {
	devFaultInfoMap := common.GetAndCleanFaultInfo()
	for _, devices := range deviceMap {
		for _, device := range devices {
			tool.flushFaultCodesWithInit(device, devFaultInfoMap)
			common.CountFaultDuration(device, devFaultInfoMap)
			device.Health = tool.isHealthy(device)
			if runMode == common.Ascend910 {
				device.NetworkHealth = tool.isNetworkHealthy(device)
			}
		}
	}
	isFirstFlushFault = false
}

func (tool *AscendTools) flushFaultCodesWithInit(device *common.NpuDevice,
	devFaultInfoMap map[int32][]npuCommon.DevFaultInfo) {
	if devFaultInfo, ok := devFaultInfoMap[device.LogicID]; ok {
		for _, faultInfo := range devFaultInfo {
			select {
			case allFaultInfo <- faultInfo:
				hwlog.RunLog.Debugf("fault %#v has been put into queue", faultInfo)
			default:
				hwlog.RunLog.Warnf("there is too many fault already in queue, queue len:%d, "+
					"will not write to k8s event: %#v", len(allFaultInfo), faultInfo)
			}
		}
	}
	common.SetNewFaultAndCacheOnceRecoverFault(device.LogicID, devFaultInfoMap[device.LogicID], device)
	common.SetNetworkNewFaultAndCacheOnceRecoverFault(device.LogicID, devFaultInfoMap[device.LogicID], device)
}

func moreThanFiveMin(device *common.NpuDevice) bool {
	if device.AlarmRaisedTime == 0 {
		return false
	}
	return time.Now().UnixMilli()-device.AlarmRaisedTime > subscribeToPollingTime
}

func networkMoreThanFiveMin(device *common.NpuDevice) bool {
	if device.NetworkAlarmRaisedTime == 0 {
		return false
	}
	return time.Now().UnixMilli()-device.NetworkAlarmRaisedTime > subscribeToPollingTime
}

// LogFaultModeChange print logs when fault mode changed
func (tool *AscendTools) LogFaultModeChange(device *common.NpuDevice, initLogicIDs []int32, newMode string) {
	var oldMode string
	var ok bool
	if oldMode, ok = faultMode[device.LogicID]; !ok {
		faultMode[device.LogicID] = newMode
		return
	}
	if oldMode == newMode {
		return
	}
	faultMode[device.LogicID] = newMode
	if newMode == common.Polling {
		var reason string
		if device.Health == v1beta1.Unhealthy && moreThanFiveMin(device) {
			reason = "fault raised more than five minutes"
		} else if device.NetworkHealth == v1beta1.Unhealthy && networkMoreThanFiveMin(device) {
			reason = "network fault raised more than five minutes"
		} else if common.Int32Tool.Contains(initLogicIDs, device.LogicID) {
			reason = "device reset"
		} else if common.SubscribeFailed {
			reason = "subscribe failed"
		} else if isFirstFlushFault {
			reason = "first flush fault"
		} else {
			reason = "unknown reason"
		}
		hwlog.RunLog.Infof("fault get mode downgrade. logicId: %v, cause by : %v", device.LogicID, reason)
		return
	}
	hwlog.RunLog.Infof("fault get mode upgrade. logicId: %v", device.LogicID)
}

func (tool *AscendTools) getNPUsByShareMode(davinCiDev common.DavinCiDev) []common.NpuDevice {
	shareDevices := make([]common.NpuDevice, 0, common.ParamOption.ShareCount)
	for index := uint(0); index < common.ParamOption.ShareCount; index++ {
		deviceName := fmt.Sprintf("%s-%d-%d", tool.name, davinCiDev.PhyID, index)
		device := tool.assembleNpuDeviceStruct(tool.name, deviceName, davinCiDev)
		shareDevices = append(shareDevices, device)
	}
	return shareDevices
}

func (tool *AscendTools) assembleShareModeDevices(davinCiDev common.DavinCiDev, devices *[]common.NpuDevice,
	deviceTypes *[]string) {
	device := tool.getNPUsByShareMode(davinCiDev)
	*devices = append(*devices, device...)
	*deviceTypes = append(*deviceTypes, tool.name)
}

// SetDeviceUsage set usage of device according to board info
func (tool *AscendTools) SetDeviceUsage(devLogicID int32) error {
	devType := tool.dmgr.GetDevType()
	if strings.HasPrefix(devType, common.Ascend310) {
		tool.deviceUsage = common.Infer
		return nil
	}

	node, err := tool.client.GetNode()
	if err != nil {
		hwlog.RunLog.Errorf("failed to get node %s info, err: %s", tool.client.NodeName, err.Error())
		return fmt.Errorf("failed to get node info")
	}
	// A800IA2 with has to label the node as server-usage:infer to divide with A800T
	if serverUsage, ok := node.Labels[common.ServerUsageLabelKey]; ok && serverUsage == common.Infer {
		tool.deviceUsage = common.Infer
		return nil
	}

	boardId, err := tool.GetServerBoardId(devLogicID)
	if err != nil {
		hwlog.RunLog.Errorf("%v", err)
		return fmt.Errorf("set device usage error")
	}
	// A800IA2 without hccs can be auto set usage as infer
	if devType == common.Ascend910B && (boardId == common.A300IA2BoardId ||
		boardId == common.A800IA2NoneHccsBoardId || boardId == common.A800IA2NoneHccsBoardIdOld) {
		tool.deviceUsage = common.Infer
		return nil
	}

	tool.deviceUsage = common.Train
	return nil
}

// GetDeviceUsage return usage of device, infer or train
func (tool *AscendTools) GetDeviceUsage() string {
	return tool.deviceUsage
}

// GetServerBoardId get server board id
func (tool *AscendTools) GetServerBoardId(devLogicID int32) (uint32, error) {
	if tool.boardId != common.EmptyBoardId {
		return tool.boardId, nil
	}
	boardInfo, err := tool.dmgr.GetBoardInfo(devLogicID)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get boardId, err: %s", err.Error())
		return common.EmptyBoardId, fmt.Errorf("get device usage error")
	}
	tool.boardId = boardInfo.BoardId
	return tool.boardId, nil
}

// GetIfCardsInResetting get whether all cards in resetting process
func (tool *AscendTools) GetIfCardsInResetting(deviceLogicId int32) bool {
	tool.cardInResetLock.Lock()
	defer tool.cardInResetLock.Unlock()
	return tool.cardInResetMap[deviceLogicId]
}

// SetCardsInResetting set the indicator of whether all cards in resetting process
func (tool *AscendTools) SetCardsInResetting(deviceLogicId int32, reset bool) {
	tool.cardInResetLock.Lock()
	defer tool.cardInResetLock.Unlock()
	tool.cardInResetMap[deviceLogicId] = reset
}

// GetResetFailedTimes get how many times has the reset process failed in a row
func (tool *AscendTools) GetResetFailedTimes(deviceLogicId int32) int {
	tool.resetFailedTimesLock.Lock()
	defer tool.resetFailedTimesLock.Unlock()
	counter, exist := tool.resetFailedTimesMap[deviceLogicId]
	if !exist {
		return 0
	}
	return counter
}

// SetResetFailedTimes set the counter of how many times the reset process has failed
func (tool *AscendTools) SetResetFailedTimes(deviceLogicId int32, count int) {
	tool.resetFailedTimesLock.Lock()
	defer tool.resetFailedTimesLock.Unlock()
	tool.resetFailedTimesMap[deviceLogicId] = count
}

// WriteFaultToEvent write fault to event
func (tool *AscendTools) WriteFaultToEvent(ctx context.Context) {
	for {
		select {
		case _, ok := <-ctx.Done():
			if !ok {
				hwlog.RunLog.Info("stop signal chanel closed")
			}
			hwlog.RunLog.Info("write fault to k8s event stop")
			return
		case faultInfo := <-allFaultInfo:
			if !limiter.Allow() {
				hwlog.RunLog.Warn("write k8s event limiter overflowed")
				hwlog.RunLog.Warnf("current event for fault:%#v will be discard", faultInfo)
				continue
			}
			if err := tool.doWriteFaultToEvent(faultInfo); err != nil {
				hwlog.RunLog.Errorf("failed to write device fault to k8s event, %v", err)
				continue
			}
		}
	}
}

// doWriteFaultToEvent writing fault event to cache
func (tool *AscendTools) doWriteFaultToEvent(faultInfo npuCommon.DevFaultInfo) error {
	cardID, deviceID, err := tool.dmgr.GetCardIDDeviceID(faultInfo.LogicID)
	if err != nil {
		return fmt.Errorf("failed to get cardID and deviceID, %w", err)
	}
	nodeName, err := kubeclient.GetNodeNameFromEnv()
	if err != nil {
		return fmt.Errorf("failed to get node name, %w", err)
	}
	podName, err := common.GetPodNameFromEnv()
	if err != nil {
		return fmt.Errorf("failed to get pod name, %w", err)
	}
	assertionName := common.GetFaultAssertionName(faultInfo.Assertion)
	if assertionName == "" {
		return fmt.Errorf("failed to get name of assertion: %d", faultInfo.Assertion)
	}
	var faultLevelName string
	if !common.NetworkFaultCodes.Has(faultInfo.EventID) {
		faultLevelName = common.GetFaultTypeByCode([]int64{faultInfo.EventID})
	} else {
		faultLevelName = common.GetNetworkFaultTypeByCode([]int64{faultInfo.EventID})
	}
	faultInfo.AlarmRaisedTime = time.Now().UnixMilli()
	event := &v1.Event{
		ObjectMeta: metav1.ObjectMeta{Namespace: api.KubeNS,
			Name: fmt.Sprintf("%s.%d%d", podName, faultInfo.AlarmRaisedTime, faultInfo.LogicID),
		},
		Type: v1.EventTypeWarning,
		Message: fmt.Sprintf("device fault, nodeName:%s, assertion:%s, cardID:%d, deviceID:%d, "+
			"faultCodes:%s, faultLevelName:%s, alarmRaisedTime:%s", nodeName, assertionName, cardID, deviceID,
			strings.ToUpper(strconv.FormatInt(faultInfo.EventID, common.Hex)), faultLevelName,
			time.UnixMilli(faultInfo.AlarmRaisedTime).Format(common.TimeFormat)),
		EventTime: metav1.MicroTime{Time: time.UnixMilli(faultInfo.AlarmRaisedTime)},
		Reason:    assertionName, Action: faultLevelName,
		Source: v1.EventSource{Component: common.Component, Host: nodeName},
		InvolvedObject: v1.ObjectReference{
			Kind: common.ResourceKindPod, Namespace: api.KubeNS, Name: podName,
		},
		ReportingController: common.Component, ReportingInstance: podName,
	}
	if faultInfo.Assertion != npuCommon.FaultOccur {
		event.Type = v1.EventTypeNormal
	}
	if _, err := tool.client.CreateEvent(event); err != nil {
		return fmt.Errorf("failed to create event, %v", err)
	}
	hwlog.RunLog.Infof("successfully create event for fault: %#v", faultInfo)
	return nil
}

// SetSuperPodID set super pod id
func (tool *AscendTools) SetSuperPodID(superPodID int32) {
	tool.superPodID = superPodID
}

// GetSuperPodID getting super pod id
func (tool *AscendTools) GetSuperPodID() int32 {
	return tool.superPodID
}

// SetServerIndex setting the index from server
func (tool *AscendTools) SetServerIndex(serverIndex int32) {
	tool.serverIndex = serverIndex
}

// GetServerIndex getting the index from server
func (tool *AscendTools) GetServerIndex() int32 {
	return tool.serverIndex
}

// HandleDropCardFaultEvents handle drop card fault events that may be lost by the fault subscription interface
func (tool *AscendTools) HandleDropCardFaultEvents(npuDevice *common.NpuDevice) {
	if common.SubscribeFailed {
		return
	}
	tool.generateCardDropFaultEvents(npuDevice)
}

func (tool *AscendTools) generateCardDropFaultEvents(npuDevice *common.NpuDevice) {
	if !npuDevice.CardDrop && tool.checkCardDropFault(npuDevice.LogicID) {
		faultInfo := npuCommon.DevFaultInfo{
			EventID:         common.CardDropFaultCode,
			LogicID:         npuDevice.LogicID,
			Assertion:       npuCommon.FaultOccur,
			AlarmRaisedTime: time.Now().UnixMilli(),
		}
		npuDevice.CardDrop = true
		hwlog.RunLog.Info("generate card drop occur fault event")
		common.SaveDevFaultInfo(faultInfo)
	}

	if npuDevice.CardDrop && !tool.checkCardDropFault(npuDevice.LogicID) {
		faultInfo := npuCommon.DevFaultInfo{
			EventID:         common.CardDropFaultCode,
			LogicID:         npuDevice.LogicID,
			Assertion:       npuCommon.FaultRecover,
			AlarmRaisedTime: time.Now().UnixMilli(),
		}
		npuDevice.CardDrop = false
		hwlog.RunLog.Info("generate card drop recover fault event")
		common.SaveDevFaultInfo(faultInfo)
	}
}

func (tool *AscendTools) checkCardDropFault(logicID int32) bool {
	_, err := tool.dmgr.GetDeviceHealth(logicID)
	if common.CheckErrorMessage(err, npuCommon.DeviceNotReadyErrCodeStr) {
		hwlog.RunLog.Errorf("logic id %d, error message contains %s, device does not ready, "+
			"the card may be dropped", logicID, npuCommon.DeviceNotReadyErrCodeStr)
		return true
	}

	return false
}

// HandleLostChipFaultEvents handle chip fault events that may be lost by the fault subscription interface
func (tool *AscendTools) HandleLostChipFaultEvents(device *common.NpuDevice, initLogicIDs []int32) {
	needHandleLostChipFaultCondition := isFirstFlushFault || (common.Int32Tool.Contains(initLogicIDs,
		device.LogicID)) || common.SubscribeFailed || (device.Health == v1beta1.Unhealthy && moreThanFiveMin(device))
	if !needHandleLostChipFaultCondition {
		return
	}
	tool.generateChipFaultEventsBasedOnFaultCacheChange(device)
}

func (tool *AscendTools) generateChipFaultEventsBasedOnFaultCacheChange(device *common.NpuDevice) {
	_, errCodes, err := tool.dmgr.GetDeviceAllErrorCode(device.LogicID)
	if err != nil {
		hwlog.RunLog.Errorf("get device fault failed logic: %d, err: %v", device.LogicID, err)
		return
	}
	chipFaultCodes := make([]int64, 0, npuCommon.MaxErrorCodeCount)
	for _, faultCode := range errCodes {
		if common.NetworkFaultCodes.Has(faultCode) {
			continue
		}
		chipFaultCodes = append(chipFaultCodes, faultCode)
	}

	chipFaultEvents := common.GetChangedDevFaultInfo(device, device.FaultCodes, chipFaultCodes)
	for _, chipFaultEvent := range chipFaultEvents {
		hwlog.RunLog.Info("generate chip fault event based on chip fault cache change")
		common.SaveDevFaultInfo(chipFaultEvent)
	}
}

// HandleLostNetworkFaultEvents handle network fault events that may be lost by the fault subscription interface
func (tool *AscendTools) HandleLostNetworkFaultEvents(device *common.NpuDevice, initLogicIDs []int32) {
	needHandleLostNetworkFaultCondition := isFirstFlushFault || (common.Int32Tool.Contains(initLogicIDs,
		device.LogicID)) || common.SubscribeFailed || (device.NetworkHealth == v1beta1.Unhealthy &&
		networkMoreThanFiveMin(device))
	if !needHandleLostNetworkFaultCondition {
		return
	}
	tool.generateNetworkFaultEventsBasedOnFaultCacheChange(device)
}

func (tool *AscendTools) generateNetworkFaultEventsBasedOnFaultCacheChange(device *common.NpuDevice) {
	_, errCodes, err := tool.dmgr.GetDeviceAllErrorCode(device.LogicID)
	if err != nil {
		hwlog.RunLog.Errorf("get device fault failed logic: %d, err: %v", device.LogicID, err)
		return
	}
	networkFaultCodes := make([]int64, 0, npuCommon.MaxErrorCodeCount)
	for _, faultCode := range errCodes {
		if !common.NetworkFaultCodes.Has(faultCode) {
			continue
		}
		networkFaultCodes = append(networkFaultCodes, faultCode)
	}

	networkFaultEvents := common.GetChangedDevFaultInfo(device, device.NetworkFaultCodes, networkFaultCodes)
	for _, networkFaultEvent := range networkFaultEvents {
		hwlog.RunLog.Info("generate network fault event based on network fault cache change")
		common.SaveDevFaultInfo(networkFaultEvent)
	}
}
