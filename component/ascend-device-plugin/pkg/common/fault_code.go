/* Copyright(C) 2023-2024. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package common a series of common function
package common

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"huawei.com/npu-exporter/v6/common-utils/hwlog"
	"huawei.com/npu-exporter/v6/common-utils/utils"
	"huawei.com/npu-exporter/v6/devmanager/common"
	"k8s.io/apimachinery/pkg/util/sets"
)

const (
	// NotHandleFault not handle fault
	NotHandleFault = "NotHandleFault"
	// RestartRequest restart request
	RestartRequest = "RestartRequest"
	// RestartBusiness restart business
	RestartBusiness = "RestartBusiness"
	// RestartNPU restart NPU
	RestartNPU = "RestartNPU"
	// FreeRestartNPU wait free and restart NPU
	FreeRestartNPU = "FreeRestartNPU"
	// SeparateNPU separate NPU
	SeparateNPU = "SeparateNPU"
	// NormalNPU normal NPU
	NormalNPU = "NormalNPU"
	// NormalNetwork normal network
	NormalNetwork = "NormalNetwork"
	// PreSeparateNPU pre separate NPU
	PreSeparateNPU = "PreSeparateNPU"
	// ManuallySeparateNPU Manually Separate NPU
	ManuallySeparateNPU = "ManuallySeparateNPU"
	// CardUnhealthy fault is caused by card unhealthy
	CardUnhealthy = "CardUnhealthy"
	// CardNetworkUnhealthy  fault is caused by card network unhealthy
	CardNetworkUnhealthy = "CardNetworkUnhealthy"
	// LinkDownFaultCode linkdown fault code
	LinkDownFaultCode = 0x81078603
	// ResetFinishFaultCode reset finish fault code
	ResetFinishFaultCode = 0x8C2FA009
	// CardDropFaultCode card drop fault code
	CardDropFaultCode = 0x40F84E00
	// faultCodeFilePath load the path for fault code
	faultCodeFilePath = "/usr/local/faultCode.json"
	// faultCustomizationFilePath load the path for fault customization
	faultCustomizationFilePath = "/usr/local/faultCustomization.json"
	// switchFaultCodeFilePath is the path for switch fault code file
	switchFaultCodeFilePath = "/usr/local/SwitchFaultCode.json"
	// halfDivisor is the number of 2
	halfDivisor = 2
	// WaitNpuReadyTime is the time used in waiting for npu ready
	WaitNpuReadyTime time.Duration = 30
	// WaitErrorCodeCleanTime is the time used in waiting for clean error code
	WaitErrorCodeCleanTime time.Duration = 30
	// WaitProcessesToZeroTime is the time used in waiting for process to zero
	WaitProcessesToZeroTime time.Duration = 60
	// ResetInterVal is the interval time used in waiting for reset
	ResetInterVal time.Duration = 5
	// PollingInterval is used to poll the dcmi interface interval time
	PollingInterval time.Duration = DefaultPollingInterval
	// SubHealthFault subHealth code
	SubHealthFault = "SubHealthFault"
)

var (
	faultTypeCode FaultTypeCode
	// NotHandleFaultCodes contains all fault code that believed to be not handled, in this case is L1
	NotHandleFaultCodes = make([]string, 0, GeneralMapSize)
	// PreSeparateFaultCodes contains all fault code that believed to be PreSeparate, in this case is L2-L3
	PreSeparateFaultCodes = make([]string, 0, GeneralMapSize)
	// SeparateFaultCodes contains all fault code that believed to be Separate, in this case is L4-L5
	SeparateFaultCodes = make([]string, 0, GeneralMapSize)
	// initLogicIDs need init fault code device. add by train or inference
	initLogicIDs []int32
	// logicIDLock operate initLogicIDs lock
	logicIDLock sync.Mutex
	// recoverFaultMap recover fault event info cache
	recoverFaultMap = make(map[int32][]int64, GeneralMapSize)
	// recoverNetworkFaultMap network recover fault event info cache
	recoverNetworkFaultMap = make(map[int32][]int64, GeneralMapSize)
	// recoverFaultFrequencyMap frequency fault info cache
	recoverFaultFrequencyMap = make(map[int32]string, GeneralMapSize)
	// devFaultInfoMap save the subscribe interface return fault
	devFaultInfoMap = make(map[int32][]common.DevFaultInfo, GeneralMapSize)
	// devFaultInfoMapLock operate devFaultInfoMap lock
	devFaultInfoMapLock sync.Mutex
	// SubscribeFailed subscribe failed flag
	SubscribeFailed bool
	// SwitchSubscribeFailed indicate switch fault subscribe failed result, true is subscribe failed
	SwitchSubscribeFailed bool
	// Synchronize used for synchronizing the fault cache between the main process and the grace tolerance coroutines
	Synchronize bool
	// manuallySeparateNpuMapLock operate manuallySeparateNpuMap lock
	manuallySeparateNpuMapLock sync.Mutex
	// manuallySeparateNpuMap manually separate npu info cache
	manuallySeparateNpuMap = make(map[int32]ManuallyFaultInfo, GeneralMapSize)
	// FaultTypeSet is a set that contains all the fault level
	FaultTypeSet = sets.NewString(NotHandleFault, RestartRequest, RestartBusiness, FreeRestartNPU,
		RestartNPU, PreSeparateNPU, SeparateNPU, ManuallySeparateNPU, SubHealthFault)
	// FaultDurationTypeSet is a set that contains all the fault Duration level
	FaultDurationTypeSet = sets.NewString(NotHandleFault, RestartRequest, RestartBusiness, FreeRestartNPU,
		RestartNPU, PreSeparateNPU, SeparateNPU, SubHealthFault)
	// NetworkFaultCodes is a set that contains all the network fault codes
	NetworkFaultCodes = sets.NewInt64(LinkDownFaultCode)
)

// fault customization
var (
	// WaitProcessReadCMTime is the time used in waiting for process read cm
	WaitProcessReadCMTime time.Duration = DefaultProcessReadCMTime
	// WaitFaultSelfHealingTime for waiting for fault self-healing
	WaitFaultSelfHealingTime time.Duration = DefaultWaitFaultSelfHealingTime
	// WaitDeviceResetTime is the time used in waiting device reset
	WaitDeviceResetTime time.Duration = DefaultWaitDeviceResetTime
	// faultFrequencyMap is the cache saving to occur frequency of a fault, key is event id
	faultFrequencyMap = make(map[string]*FaultFrequencyCache, common.MaxErrorCodeCount)
	// faultFrequencyMapLock is the lock of faultFrequencyMap
	faultFrequencyMapLock sync.Mutex
	// faultDurationMap is the cache saving to occur duration of a fault, key is event id
	faultDurationMap = make(map[string]*FaultDurationCache, common.MaxErrorCodeCount)
	// faultDurationMapLock is the lock of faultDurationMap
	faultDurationMapLock           sync.Mutex
	faultSeverityMap               = make(map[int64]int8, common.MaxErrorCodeCount)
	parseHexFailedMsg              = "parse hex int failed and skip it, string: %s"
	networkFaultConfigureFailedMsg = "%x is a network fault and cannot be configured to %s now, " +
		"fault handling policy is set to NotHandleFault"
	hbmTool = NewHbmFaultManager()
)

// ManuallyFaultInfo save the info of ManuallySeparateNPU
type ManuallyFaultInfo struct {
	LogicID     int32
	FirstHandle bool
	RecordTime  int64
}

// FaultTypeCode group code by type
type FaultTypeCode struct {
	NotHandleFaultCodes        []int64
	RestartRequestCodes        []int64
	RestartBusinessCodes       []int64
	RestartNPUCodes            []int64
	FreeRestartNPUCodes        []int64
	PreSeparateNPUCodes        []int64
	SeparateNPUCodes           []int64
	NotHandleFaultNetworkCodes []int64
	PreSeparateNPUNetworkCodes []int64
	SeparateNPUNetworkCodes    []int64
	SubHealthFaultCodes        []int64
}

// faultFileInfo fault code file data
type faultFileInfo struct {
	NotHandleFaultCodes        []string
	RestartRequestCodes        []string
	RestartBusinessCodes       []string
	RestartNPUCodes            []string
	FreeRestartNPUCodes        []string
	SeparateNPUCodes           []string
	PreSeparateNPUCodes        []string
	NotHandleFaultNetworkCodes []string
	PreSeparateNPUNetworkCodes []string
	SeparateNPUNetworkCodes    []string
	SubHealthFaultCodes        []string
}

// SwitchFaultFileInfo contains all fault code loading from faultconfig configmap or switchfaultconfig.json
type SwitchFaultFileInfo struct {
	NotHandleFaultCodes []string
	SubHealthFaultCodes []string
	ResetFaultCodes     []string
	SeparateFaultCodes  []string
}

// FaultCustomization is the customization info of fault
type FaultCustomization struct {
	GraceTolerance GraceToleranceCustomization
	FaultFrequency []FaultFrequencyCustomization
	FaultDuration  []FaultDurationCustomization
}

// GraceToleranceCustomization is the customization info of grace tolerance
type GraceToleranceCustomization struct {
	WaitProcessReadCMTime    int64
	WaitDeviceResetTime      int64
	WaitFaultSelfHealingTime int64
}

// FaultFrequencyCustomization is the customization info of fault frequency
type FaultFrequencyCustomization struct {
	EventId []string
	FaultFrequency
}

// FaultFrequencyCache is the cache saving the FaultFrequency
type FaultFrequencyCache struct {
	// key: logicID, value: fault occurrence time (unix time)
	Frequency map[int32][]int64
	FaultFrequency
}

// FaultFrequency is the base info of fault frequency
type FaultFrequency struct {
	TimeWindow    int64
	Times         int64
	FaultHandling string
}

// FaultDurationCustomization is the customization info of fault duration
type FaultDurationCustomization struct {
	EventId []string
	FaultDuration
}

// FaultDurationCache is the cache saving the FaultDuration
type FaultDurationCache struct {
	// key: logicID, value: fault duration data
	Duration map[int32]FaultDurationData
	FaultDuration
}

// FaultDurationData saved data during fault duration statistics
type FaultDurationData struct {
	TimeoutStatus            bool
	FaultEventQueue          []common.DevFaultInfo
	FaultDurationTime        int64
	FaultRecoverDurationTime int64
}

// FaultDuration is the base info of fault duration
type FaultDuration struct {
	FaultTimeout   int64
	RecoverTimeout int64
	FaultHandling  string
}

type handleDurationInputPara struct {
	logicID       int32
	eventId       string
	index         int
	timeoutStatus bool
	duration      int64
}

// DevFaultInfoBasedTimeAscend sort fault queue based on alarmRaisedTime in ascending order
type DevFaultInfoBasedTimeAscend []common.DevFaultInfo

// Len is a fixed usage to find the length of type
func (devFault DevFaultInfoBasedTimeAscend) Len() int {
	return len(devFault)
}

// Swap is a fixed usage to switch the index of type
func (devFault DevFaultInfoBasedTimeAscend) Swap(i, j int) {
	if i >= len(devFault) || j >= len(devFault) {
		hwlog.RunLog.Errorf("index out of range, i: %d, j: %d, length: %d", i, j, len(devFault))
		return
	}
	devFault[i], devFault[j] = devFault[j], devFault[i]
}

// Less is fixed usage to check if one is less than the other one of type
func (devFault DevFaultInfoBasedTimeAscend) Less(i, j int) bool {
	if i >= len(devFault) || j >= len(devFault) {
		hwlog.RunLog.Errorf("index out of range, i: %d, j: %d, length: %d", i, j, len(devFault))
		return false
	}
	return devFault[i].AlarmRaisedTime < devFault[j].AlarmRaisedTime
}

// HbmFaultManager manage the accompanying faults of aic error and hbm error
type HbmFaultManager struct {
	HbmOccurTimeCache map[int32]int64
	AicFaultEventQue  map[int32][]common.DevFaultInfo
}

// NewHbmFaultManager return a hbm fault manager
func NewHbmFaultManager() *HbmFaultManager {
	return &HbmFaultManager{
		HbmOccurTimeCache: make(map[int32]int64, GeneralMapSize),
		AicFaultEventQue:  make(map[int32][]common.DevFaultInfo, GeneralMapSize),
	}
}

func (h *HbmFaultManager) updateHbmOccurTime(faultInfo common.DevFaultInfo) {
	h.HbmOccurTimeCache[faultInfo.LogicID] = faultInfo.AlarmRaisedTime
	hwlog.RunLog.Debugf("hbm fault occur, device %d update occur time: %d",
		faultInfo.LogicID, h.HbmOccurTimeCache[faultInfo.LogicID])
}

func (h *HbmFaultManager) aicFaultEventInQue(faultInfo common.DevFaultInfo) {
	_, ok := h.AicFaultEventQue[faultInfo.LogicID]
	if !ok {
		h.AicFaultEventQue[faultInfo.LogicID] = []common.DevFaultInfo{}
	}
	h.AicFaultEventQue[faultInfo.LogicID] = append(h.AicFaultEventQue[faultInfo.LogicID], faultInfo)
	sort.Sort(DevFaultInfoBasedTimeAscend(h.AicFaultEventQue[faultInfo.LogicID]))
	hwlog.RunLog.Debugf("aic/aiv fault event %d in que, device %d new event que:%#v",
		faultInfo.EventID, faultInfo.LogicID, h.AicFaultEventQue[faultInfo.LogicID])
}

func (h *HbmFaultManager) aicFaultEventOutQue(logicId int32) []common.DevFaultInfo {
	var faultInfoList []common.DevFaultInfo
	faultEventQue, ok := h.AicFaultEventQue[logicId]
	if !ok {
		return faultInfoList
	}
	if _, ok := h.HbmOccurTimeCache[logicId]; !ok {
		h.HbmOccurTimeCache[logicId] = 0
	}
	var newFaultEventQue []common.DevFaultInfo
	nowTime := time.Now().Unix()
	for i := 0; i < len(faultEventQue); i++ {
		// The fault aic error occurring ten seconds before and after the occurrence of hbm error should be deleted,
		if Int64Tool.Abs(h.HbmOccurTimeCache[logicId], faultEventQue[i].AlarmRaisedTime) <
			AssociatedFaultDiagnosisTime*TimeMilliseconds {
			hwlog.RunLog.Infof("device %d delete event in fault event que, aic event time %d hbm event time %d",
				logicId, faultEventQue[i].AlarmRaisedTime, h.HbmOccurTimeCache[logicId])
			continue
		}
		// aic error should report if hbm error does not occur within ten seconds,
		// and the event in this outbound queue should also be deleted
		if nowTime-faultEventQue[i].AlarmRaisedTime > AssociatedFaultDiagnosisTime*TimeMilliseconds {
			hwlog.RunLog.Infof("device % delete event in fault event que, aic event time %d now time %d",
				logicId, faultEventQue[i].AlarmRaisedTime, nowTime)
			faultInfoList = append(faultInfoList, faultEventQue[i])
			continue
		}
		newFaultEventQue = append(newFaultEventQue, faultEventQue[i])
	}
	h.AicFaultEventQue[logicId] = newFaultEventQue
	return faultInfoList
}

// LoadFaultCodeFromFile load fault code and fault type from faultCode.json
func LoadFaultCodeFromFile() error {
	faultCodeBytes, err := utils.LoadFile(faultCodeFilePath)
	if err != nil {
		return fmt.Errorf("load fault code json failed: %v", err)
	}
	return LoadFaultCode(faultCodeBytes)
}

// LoadSwitchFaultCodeFromFile load fault code from SwitchFaultCode.json
func LoadSwitchFaultCodeFromFile() error {
	switchFaultsBytes, err := utils.LoadFile(switchFaultCodeFilePath)
	if err != nil {
		return fmt.Errorf("load switch fault code failed: %v", err)
	}
	return LoadSwitchFaultCode(switchFaultsBytes)
}

// LoadFaultCustomizationFromFile load fault customization from faultCustomization.json
func LoadFaultCustomizationFromFile() error {
	faultCodeBytes, err := utils.LoadFile(faultCustomizationFilePath)
	if err != nil {
		return fmt.Errorf("load fault customization json failed: %v", err)
	}
	if err = LoadFaultCustomization(faultCodeBytes); err != nil {
		return err
	}
	return nil
}

// ResetFaultCustomizationCache reset fault customization cache
func ResetFaultCustomizationCache() {
	hwlog.RunLog.Debug("reset fault customization, fault customization cache will be cleared")
	faultFrequencyMapLock.Lock()
	faultFrequencyMap = make(map[string]*FaultFrequencyCache, common.MaxErrorCodeCount)
	faultFrequencyMapLock.Unlock()
	faultDurationMapLock.Lock()
	faultDurationMap = make(map[string]*FaultDurationCache, common.MaxErrorCodeCount)
	faultDurationMapLock.Unlock()
}

// LoadFaultCode loads the fault codes
func LoadFaultCode(faultCodeBytes []byte) error {
	var fileInfo faultFileInfo
	if err := json.Unmarshal(faultCodeBytes, &fileInfo); err != nil {
		return fmt.Errorf("unmarshal fault code byte failed: %v", err)
	}
	faultTypeCode = FaultTypeCode{
		NotHandleFaultCodes:        StringTool.HexStringToInt(fileInfo.NotHandleFaultCodes),
		RestartRequestCodes:        StringTool.HexStringToInt(fileInfo.RestartRequestCodes),
		RestartBusinessCodes:       StringTool.HexStringToInt(fileInfo.RestartBusinessCodes),
		RestartNPUCodes:            StringTool.HexStringToInt(fileInfo.RestartNPUCodes),
		FreeRestartNPUCodes:        StringTool.HexStringToInt(fileInfo.FreeRestartNPUCodes),
		PreSeparateNPUCodes:        StringTool.HexStringToInt(fileInfo.PreSeparateNPUCodes),
		SeparateNPUCodes:           StringTool.HexStringToInt(fileInfo.SeparateNPUCodes),
		NotHandleFaultNetworkCodes: StringTool.HexStringToInt(fileInfo.NotHandleFaultNetworkCodes),
		PreSeparateNPUNetworkCodes: StringTool.HexStringToInt(fileInfo.PreSeparateNPUNetworkCodes),
		SeparateNPUNetworkCodes:    StringTool.HexStringToInt(fileInfo.SeparateNPUNetworkCodes),
		SubHealthFaultCodes:        StringTool.HexStringToInt(fileInfo.SubHealthFaultCodes),
	}

	// It is not clear whether the current network fault is separated from the chip fault. The network fault configured
	// in chip fault is temporarily mapped to network processing policy for processing.
	mappingChipFaultToNetworkFaultCodesSupport()
	mappingChipFaultToNetworkFaultCodesNotSupport()

	return nil
}

func mappingChipFaultToNetworkFaultCodesSupport() {
	for _, faultCode := range faultTypeCode.NotHandleFaultCodes {
		if NetworkFaultCodes.Has(faultCode) {
			faultTypeCode.NotHandleFaultNetworkCodes = append(faultTypeCode.NotHandleFaultNetworkCodes, faultCode)
		}
	}

	for _, faultCode := range faultTypeCode.PreSeparateNPUCodes {
		if NetworkFaultCodes.Has(faultCode) {
			faultTypeCode.PreSeparateNPUNetworkCodes = append(faultTypeCode.PreSeparateNPUNetworkCodes, faultCode)
		}
	}

	for _, faultCode := range faultTypeCode.SeparateNPUCodes {
		if NetworkFaultCodes.Has(faultCode) {
			faultTypeCode.SeparateNPUNetworkCodes = append(faultTypeCode.SeparateNPUNetworkCodes, faultCode)
		}
	}
}

func mappingChipFaultToNetworkFaultCodesNotSupport() {
	for _, faultCode := range faultTypeCode.RestartRequestCodes {
		if NetworkFaultCodes.Has(faultCode) {
			hwlog.RunLog.Warnf(networkFaultConfigureFailedMsg, faultCode, RestartRequest)
			faultTypeCode.NotHandleFaultNetworkCodes = append(faultTypeCode.NotHandleFaultNetworkCodes, faultCode)
		}
	}

	for _, faultCode := range faultTypeCode.RestartBusinessCodes {
		if NetworkFaultCodes.Has(faultCode) {
			hwlog.RunLog.Warnf(networkFaultConfigureFailedMsg, faultCode, RestartBusiness)
			faultTypeCode.NotHandleFaultNetworkCodes = append(faultTypeCode.NotHandleFaultNetworkCodes, faultCode)
		}
	}

	for _, faultCode := range faultTypeCode.RestartNPUCodes {
		if NetworkFaultCodes.Has(faultCode) {
			hwlog.RunLog.Warnf(networkFaultConfigureFailedMsg, faultCode, RestartNPU)
			faultTypeCode.NotHandleFaultNetworkCodes = append(faultTypeCode.NotHandleFaultNetworkCodes, faultCode)
		}
	}

	for _, faultCode := range faultTypeCode.FreeRestartNPUCodes {
		if NetworkFaultCodes.Has(faultCode) {
			hwlog.RunLog.Warnf(networkFaultConfigureFailedMsg, faultCode, FreeRestartNPU)
			faultTypeCode.NotHandleFaultNetworkCodes = append(faultTypeCode.NotHandleFaultNetworkCodes, faultCode)
		}
	}
}

// LoadFaultCustomization loads fault customization
func LoadFaultCustomization(faultCustomizationByte []byte) error {
	var faultCustomization FaultCustomization
	if err := json.Unmarshal(faultCustomizationByte, &faultCustomization); err != nil {
		hwlog.RunLog.Errorf("load fault customization failed, unmarshal err: %v", err)
		return err
	}
	loadGraceToleranceCustomization(faultCustomization.GraceTolerance)
	loadFaultFrequencyCustomization(faultCustomization.FaultFrequency)
	loadFaultDurationCustomization(faultCustomization.FaultDuration)
	return nil
}

// LoadSwitchFaultCode Load SwitchFault Code from bytes of config file or configmap
func LoadSwitchFaultCode(switchFaultCodeByte []byte) error {
	var switchFileInfo SwitchFaultFileInfo
	if err := json.Unmarshal(switchFaultCodeByte, &switchFileInfo); err != nil {
		return fmt.Errorf("failed to unmarsha switch fault code, err: %s", err.Error())
	}

	NotHandleFaultCodes = make([]string, 0, GeneralMapSize)
	PreSeparateFaultCodes = make([]string, 0, GeneralMapSize)
	SeparateFaultCodes = make([]string, 0, GeneralMapSize)
	invalidFormatInfo := "failed to parse %s faultCode:%v, will ignore it," +
		" please check if its format, such as: [0x00f1ff09,155914,cpu,na]"
	for _, code := range switchFileInfo.NotHandleFaultCodes {
		if !isValidSwitchFaultCode(code) {
			hwlog.RunLog.Warnf(invalidFormatInfo, "NotHandleFaultCodes", code)
			continue
		}
		NotHandleFaultCodes = append(NotHandleFaultCodes, code)
	}

	for _, code := range switchFileInfo.SubHealthFaultCodes {
		if !isValidSwitchFaultCode(code) {
			hwlog.RunLog.Warnf(invalidFormatInfo, "SubHealthFaultCodes", code)
			continue
		}
		PreSeparateFaultCodes = append(PreSeparateFaultCodes, code)
	}

	switchFileInfo.SeparateFaultCodes = append(switchFileInfo.SeparateFaultCodes, switchFileInfo.ResetFaultCodes...)
	for _, code := range switchFileInfo.SeparateFaultCodes {
		if !isValidSwitchFaultCode(code) {
			hwlog.RunLog.Warnf(invalidFormatInfo, "SeparateFaultCodes", code)
			continue
		}
		SeparateFaultCodes = append(SeparateFaultCodes, code)
	}

	return nil
}

// isValidSwitchFaultCode to judge is a fault code is valid format as [0x00f1ff09,155914,cpu,na]
func isValidSwitchFaultCode(code string) bool {
	if len(code) > MaxLengthOfFaultCode {
		return false
	}
	if !strings.HasPrefix(code, "[") || !strings.HasSuffix(code, "]") {
		return false
	}
	parts := strings.Split(code, CommaSepDev)
	return len(parts) == PartNumOfFaultCode
}

func loadFaultDurationCustomization(customization []FaultDurationCustomization) {
	handledEventId := make(sets.String, common.MaxErrorCodeCount)
	for _, cus := range customization {
		if !validateFaultDurationCustomization(cus) {
			continue
		}
		for _, id := range cus.EventId {
			id = strings.ToLower(id)
			if handledEventId.Has(id) {
				hwlog.RunLog.Warnf("duplicated event id detected when handling FaultDuration, skip, "+
					"event id: %s", id)
				continue
			}
			handledEventId.Insert(id)
			if cache, ok := faultDurationMap[id]; ok {
				cache.FaultTimeout = cus.FaultTimeout
				cache.RecoverTimeout = cus.RecoverTimeout
				cache.FaultHandling = cus.FaultHandling
				hwlog.RunLog.Debugf("update FaultDuration for event id %s success, FaultTimeout: %d, "+
					"RecoverTimeout: %d, FaultHandling: %s", id, cus.FaultTimeout, cus.RecoverTimeout,
					cus.FaultHandling)
			} else {
				faultDurationMap[id] = &FaultDurationCache{
					Duration: make(map[int32]FaultDurationData, GeneralMapSize),
					FaultDuration: FaultDuration{
						FaultTimeout:   cus.FaultTimeout,
						RecoverTimeout: cus.RecoverTimeout,
						FaultHandling:  cus.FaultHandling,
					},
				}
				hwlog.RunLog.Debugf("insert FaultDuration for event id %s success, FaultTimeout: %d, "+
					"RecoverTimeout: %d, FaultHandling: %s", id, cus.FaultTimeout, cus.RecoverTimeout,
					cus.FaultHandling)
			}
		}
	}
	// delete event id those in cache but not in CM
	cachedEventIds := make([]string, 0, len(faultDurationMap))
	for k := range faultDurationMap {
		cachedEventIds = append(cachedEventIds, k)
	}
	for _, cachedId := range cachedEventIds {
		if !handledEventId.Has(cachedId) && len(cachedId) != 0 {
			delete(faultDurationMap, cachedId)
			hwlog.RunLog.Infof("delete FaultDuration for event id %s", cachedId)
		}
	}
}

func loadGraceToleranceCustomization(customization GraceToleranceCustomization) {
	if customization.WaitDeviceResetTime < MinWaitDeviceResetTime ||
		customization.WaitDeviceResetTime > MaxWaitDeviceResetTime {
		hwlog.RunLog.Errorf("WaitDeviceResetTime(%d) exceed limit(%d~%d), use default(%d)",
			customization.WaitDeviceResetTime, MinWaitDeviceResetTime,
			MaxWaitDeviceResetTime, DefaultWaitDeviceResetTime)
		WaitDeviceResetTime = DefaultWaitDeviceResetTime
	} else {
		hwlog.RunLog.Debugf("modify WaitDeviceResetTime(%d) success", customization.WaitDeviceResetTime)
		WaitDeviceResetTime = time.Duration(customization.WaitDeviceResetTime)
	}
	if customization.WaitProcessReadCMTime < MinWaitProcessReadCMTime || customization.
		WaitProcessReadCMTime > MaxWaitProcessReadCMTime {
		hwlog.RunLog.Errorf("WaitProcessReadCMTime(%d) exceed limit(%d~%d), use default(%d)",
			customization.WaitProcessReadCMTime, MinWaitProcessReadCMTime,
			MaxWaitProcessReadCMTime, DefaultProcessReadCMTime)
		WaitProcessReadCMTime = DefaultProcessReadCMTime
	} else {
		hwlog.RunLog.Debugf("modify WaitProcessReadCMTime(%d) success", customization.WaitProcessReadCMTime)
		WaitProcessReadCMTime = time.Duration(customization.WaitProcessReadCMTime)
	}
	if customization.WaitFaultSelfHealingTime < MinWaitFaultSelfHealingTime ||
		time.Duration(customization.WaitFaultSelfHealingTime) > MaxWaitFaultSelfHealingTime {
		hwlog.RunLog.Errorf("WaitFaultSelfHealingTime(%d) exceed limit(%d~%d), use default(%d)",
			customization.WaitFaultSelfHealingTime,
			MinWaitFaultSelfHealingTime, WaitProcessReadCMTime, DefaultWaitFaultSelfHealingTime)
		WaitFaultSelfHealingTime = DefaultWaitFaultSelfHealingTime
	} else {
		hwlog.RunLog.Debugf("modify WaitFaultSelfHealingTime(%d) success", customization.WaitFaultSelfHealingTime)
		WaitFaultSelfHealingTime = time.Duration(customization.WaitFaultSelfHealingTime)
	}
}

func loadFaultFrequencyCustomization(customizations []FaultFrequencyCustomization) {
	handledEventId := make(sets.String, GeneralMapSize)
	faultFrequencyMapLock.Lock()
	defer faultFrequencyMapLock.Unlock()
	for _, cus := range customizations {
		if !validateFaultFrequencyCustomization(cus) {
			continue
		}
		for _, id := range cus.EventId {
			id = strings.ToLower(id)
			if handledEventId.Has(id) {
				hwlog.RunLog.Warnf("duplicated event id detected when handling FaultFrequency, "+
					"skip, event id: %s", id)
				continue
			}
			handledEventId.Insert(id)
			if cache, ok := faultFrequencyMap[id]; ok {
				cache.TimeWindow = cus.TimeWindow
				cache.Times = cus.Times
				cache.FaultHandling = cus.FaultHandling
				hwlog.RunLog.Debugf("update FaultFrequency for event id %s success, TimeWindow: %d, "+
					"Times: %d, FaultHandling: %s", id, cus.TimeWindow, cus.Times, cus.FaultHandling)
			} else {
				faultFrequencyMap[id] = &FaultFrequencyCache{
					Frequency: make(map[int32][]int64, common.MaxErrorCodeCount),
					FaultFrequency: FaultFrequency{
						TimeWindow:    cus.TimeWindow,
						Times:         cus.Times,
						FaultHandling: cus.FaultHandling,
					},
				}
				hwlog.RunLog.Debugf("insert FaultFrequency for event id %s success, TimeWindow: %d, "+
					"Times: %d, FaultHandling: %s", id, cus.TimeWindow, cus.Times, cus.FaultHandling)
			}
		}
	}
	// delete event id those in cache but not in CM
	cachedEventIds := make([]string, 0, len(faultFrequencyMap))
	for k := range faultFrequencyMap {
		cachedEventIds = append(cachedEventIds, k)
	}
	for _, cachedId := range cachedEventIds {
		if !handledEventId.Has(cachedId) && len(cachedId) != 0 {
			delete(faultFrequencyMap, cachedId)
			hwlog.RunLog.Infof("delete FaultFrequency for event id %s", cachedId)
		}
	}
}

func insertFaultFrequency(logicId int32, eventId int64) {
	faultFrequencyMapLock.Lock()
	defer faultFrequencyMapLock.Unlock()
	eventIdStr := strings.ToLower(strconv.FormatInt(eventId, Hex))
	frequencyCache, ok := faultFrequencyMap[eventIdStr]
	if !ok {
		hwlog.RunLog.Debugf("skip inserting event id %s to fault frequency cache, no config found", eventIdStr)
		return
	}
	_, ok = frequencyCache.Frequency[logicId]
	if !ok {
		frequencyCache.Frequency[logicId] = make([]int64, 0, frequencyCache.Times)
	}
	frequencyCache.Frequency[logicId] = append(frequencyCache.Frequency[logicId], time.Now().Unix())
	hwlog.RunLog.Infof("insert fault frequency success, event id: %s, logic id: %d, unix time: %d, "+
		"occurrence times :%d", eventIdStr, logicId, time.Now().Unix(), len(frequencyCache.Frequency[logicId]))
}

func validateFaultFrequencyCustomization(customization FaultFrequencyCustomization) bool {
	if len(customization.EventId) == 0 {
		hwlog.RunLog.Warnf("empty event id in this FaultFrequency, skip")
		return false
	}
	invalidMsg := "FaultFrequency configuration of this part will be invalid"
	if customization.TimeWindow > MaxFaultFrequencyTimeWindow || customization.TimeWindow < MinFaultFrequencyTimeWindow {
		hwlog.RunLog.Warnf("EventIDs: %v, TimeWindow(%d) in this FaultFrequency exceeds limit(%d~%d). %s",
			customization.EventId, customization.TimeWindow, MinFaultFrequencyTimeWindow, MaxFaultFrequencyTimeWindow,
			invalidMsg)
		return false
	}
	if customization.Times > MaxFaultFrequencyTimes || customization.Times < MinFaultFrequencyTimes {
		hwlog.RunLog.Warnf("EventIDs: %v, Times(%d) in this FaultFrequency exceeds limit(%d~%d). %s",
			customization.EventId, customization.Times, MinFaultFrequencyTimes, MaxFaultFrequencyTimes, invalidMsg)
		return false
	}
	if !FaultTypeSet.Has(customization.FaultHandling) {
		hwlog.RunLog.Warnf("EventIDs: %v, FaultHandling(%s) in this FaultFrequency is unrecognized. "+
			"The supported range of FaultHandling in this FaultFrequency is %v. %s",
			customization.EventId, customization.FaultHandling, FaultTypeSet.List(), invalidMsg)
		return false
	}
	return true
}

func validateFaultDurationCustomization(faultDurationCustomization FaultDurationCustomization) bool {
	if len(faultDurationCustomization.EventId) == 0 {
		hwlog.RunLog.Warnf("empty event id in this FaultDuration, skip")
		return false
	}
	invalidMsg := "FaultDuration configuration of this part will be invalid"
	if faultDurationCustomization.FaultTimeout > MaxFaultTimeout ||
		faultDurationCustomization.FaultTimeout < MinFaultTimeout {
		hwlog.RunLog.Warnf("EventIDs: %v, FaultTimeout(%d) in this FaultDuration exceeds limit(%d~%d). %s",
			faultDurationCustomization.EventId, faultDurationCustomization.FaultTimeout,
			MinFaultTimeout, MaxFaultTimeout, invalidMsg)
		return false
	}
	if faultDurationCustomization.RecoverTimeout > MaxRecoverTimeout ||
		faultDurationCustomization.RecoverTimeout < MinRecoverTimeout {
		hwlog.RunLog.Warnf("EventIDs: %v, RecoverTimeout(%d) in this FaultDuration exceeds limit(%d~%d). %s",
			faultDurationCustomization.EventId, faultDurationCustomization.RecoverTimeout,
			MinRecoverTimeout, MaxRecoverTimeout, invalidMsg)
		return false
	}
	if !FaultDurationTypeSet.Has(faultDurationCustomization.FaultHandling) {
		hwlog.RunLog.Warnf("EventIDs: %v, FaultHandling(%s) in this FaultDuration is unrecognized. "+
			"The supported range of FaultHandling in this FaultDuration is %v. %s", faultDurationCustomization.EventId,
			faultDurationCustomization.FaultHandling, FaultDurationTypeSet.List(), invalidMsg)
		return false
	}
	return true
}

// GetNetworkFaultTypeByCode get network fault type by fault code. if code not record, default PreSeparateNPU
func GetNetworkFaultTypeByCode(faultCodes []int64) string {
	if len(faultCodes) == 0 {
		return NormalNetwork
	}
	if len(faultTypeCode.NotHandleFaultCodes) == 0 && len(faultTypeCode.PreSeparateNPUNetworkCodes) == 0 {
		if err := LoadFaultCodeFromFile(); err != nil {
			return PreSeparateNPU
		}
	}
	switch {
	case Int64Tool.SameElement(faultTypeCode.SeparateNPUNetworkCodes, faultCodes):
		return SeparateNPU
	case Int64Tool.SameElement(faultTypeCode.PreSeparateNPUNetworkCodes, faultCodes):
		return PreSeparateNPU
	case Int64Tool.SameElement(faultTypeCode.NotHandleFaultNetworkCodes, faultCodes):
		return NotHandleFault
	default:
		hwlog.RunLog.Debugf("not record fault code : %v, use default type PreSeparateNPU", faultCodes)
		return PreSeparateNPU
	}
}

// GetFaultType will return the fault type from fault codes,
// fault frequency, fault duration and ManuallySeparateNPU cache
func GetFaultType(faultCodes []int64, logicId int32) string {
	newFaultCodes := make([]int64, 0)
	for _, faultCode := range faultCodes {
		if !NetworkFaultCodes.Has(faultCode) {
			newFaultCodes = append(newFaultCodes, faultCode)
		}
	}

	faultTypes := make([]string, 0, len(FaultTypeSet))
	faultTypes = append(faultTypes, GetFaultTypeByCode(newFaultCodes))
	faultTypes = append(faultTypes, GetFaultTypeFromFaultFrequency(logicId))
	faultTypes = append(faultTypes, GetFaultTypeFromFaultDuration(logicId, ChipFaultMode))
	if QueryManuallyFaultInfoByLogicID(logicId) {
		faultTypes = append(faultTypes, ManuallySeparateNPU)
	}
	return getMostSeriousFaultType(faultTypes)
}

// GetNetworkFaultType will return the fault type from network fault codes, fault duration
func GetNetworkFaultType(faultCodes []int64, logicId int32) string {
	newNetworkFaultCodes := make([]int64, 0)
	for _, faultCode := range faultCodes {
		if NetworkFaultCodes.Has(faultCode) {
			newNetworkFaultCodes = append(newNetworkFaultCodes, faultCode)
		}
	}

	faultTypes := make([]string, 0, len(FaultTypeSet))
	faultTypes = append(faultTypes, GetNetworkFaultTypeByCode(newNetworkFaultCodes))
	faultTypes = append(faultTypes, GetFaultTypeFromFaultDuration(logicId, NetworkFaultMode))
	return getMostSeriousFaultType(faultTypes)
}

// GetFaultTypeByCode get fault type by fault code. if code not record, default SeparateNPU0
func GetFaultTypeByCode(faultCodes []int64) string {
	if len(faultCodes) == 0 {
		return NormalNPU
	}
	switch {
	case Int64Tool.SameElement(faultTypeCode.SeparateNPUCodes, faultCodes):
		return SeparateNPU
	case Int64Tool.SameElement(faultTypeCode.PreSeparateNPUCodes, faultCodes):
		return PreSeparateNPU
	case Int64Tool.SameElement(faultTypeCode.RestartNPUCodes, faultCodes):
		return RestartNPU
	case Int64Tool.SameElement(faultTypeCode.FreeRestartNPUCodes, faultCodes):
		return FreeRestartNPU
	case Int64Tool.SameElement(faultTypeCode.RestartBusinessCodes, faultCodes):
		return RestartBusiness
	case Int64Tool.SameElement(faultTypeCode.RestartRequestCodes, faultCodes):
		return RestartRequest
	case Int64Tool.SameElement(faultTypeCode.NotHandleFaultCodes, faultCodes):
		return NotHandleFault
	case Int64Tool.SameElement(faultTypeCode.SubHealthFaultCodes, faultCodes):
		return SubHealthFault
	default:
		faultType := getFaultTypeBySeverity(faultCodes)
		hwlog.RunLog.Debugf("not record fault code: %v, get fault type by severity: %s", faultCodes, faultType)
		return faultType
	}
}

// GetFaultTypeFromFaultFrequency refreshes the cache of FaultFrequency, delete the faults those not in time window,
// and return the fault level if the occurrence times of fault >= the set value
func GetFaultTypeFromFaultFrequency(logicId int32) string {
	faultTypes := make([]string, 0, len(faultFrequencyMap))
	faultFrequencyMapLock.Lock()
	defer faultFrequencyMapLock.Unlock()
	for eventId, frequencyCache := range faultFrequencyMap {
		_, ok := frequencyCache.Frequency[logicId]
		if !ok {
			continue
		}
		timeWindowStart := time.Now().Unix() - frequencyCache.TimeWindow
		// delete the occurrence times those less than the start of time window
		index := 0
		for _, occurrenceTime := range frequencyCache.Frequency[logicId] {
			if occurrenceTime < timeWindowStart {
				hwlog.RunLog.Infof("delete the expired fault occurrence, event id: %s, logic id: %d, "+
					"time window start: %d, occurrence time: %d", eventId, logicId, timeWindowStart, occurrenceTime)
				index++
			} else {
				break
			}
		}
		frequencyCache.Frequency[logicId] = frequencyCache.Frequency[logicId][index:]
		if int64(len(frequencyCache.Frequency[logicId])) >= frequencyCache.Times {
			hwlog.RunLog.Infof("FaultFrequency detected, event id: %s, logic id: %d, fault occurred times: %d, "+
				"fault level: %s", eventId, logicId, len(frequencyCache.Frequency[logicId]), frequencyCache.FaultHandling)
			if frequencyCache.FaultHandling == ManuallySeparateNPU {
				hwlog.RunLog.Infof("detect ManuallySeparateNPU, logic id: %d", logicId)
				SaveManuallyFaultInfo(logicId)
			}
			faultTypes = append(faultTypes, frequencyCache.FaultHandling)
			// every time when frequency fault detected, record frequency fault to be cleared in cache
			recoverFaultFrequencyMap[logicId] = eventId
		}
	}
	return getMostSeriousFaultType(faultTypes)
}

// GetFaultTypeFromFaultDuration get fault type from fault duration cache
func GetFaultTypeFromFaultDuration(logicId int32, mode string) string {
	if mode != ChipFaultMode && mode != NetworkFaultMode {
		return NormalNPU
	}
	faultDurationMapLock.Lock()
	defer faultDurationMapLock.Unlock()

	faultTypes := make([]string, 0, len(faultDurationMap))
	for eventId, faultDurationCache := range faultDurationMap {
		num, err := strconv.ParseInt(eventId, Hex, 0)
		if err != nil {
			hwlog.RunLog.Errorf(parseHexFailedMsg, eventId)
			continue
		}

		if (mode == ChipFaultMode && NetworkFaultCodes.Has(num)) ||
			(mode == NetworkFaultMode && !NetworkFaultCodes.Has(num)) {
			continue
		}

		faultDurationData, ok := faultDurationCache.Duration[logicId]
		if !ok {
			continue
		}

		if faultDurationData.TimeoutStatus {
			hwlog.RunLog.Debugf("FaultDuration detected, event id: %s, logic id: %d, "+
				"fault duration time: %.2f seconds, "+
				"fault level: %s", eventId, logicId,
				float64(faultDurationData.FaultDurationTime)/SecondMagnificationFloat,
				faultDurationCache.FaultHandling)
			faultTypes = append(faultTypes, faultDurationCache.FaultHandling)
		}
	}
	return getMostSeriousFaultType(faultTypes)
}

func getFaultTypeBySeverity(faultCodes []int64) string {
	for _, code := range faultCodes {
		severity, ok := faultSeverityMap[code]
		if !ok {
			hwlog.RunLog.Warnf("detect unknown fault code and no match severity: %d", code)
			return SeparateNPU
		}
		if severity > FaultSeverityMinor {
			return SeparateNPU
		}
	}
	return NotHandleFault
}

func getMostSeriousFaultType(fautTypes []string) string {
	faultTypeSet := sets.NewString(fautTypes...)
	if faultTypeSet.Has(ManuallySeparateNPU) {
		return ManuallySeparateNPU
	} else if faultTypeSet.Has(SeparateNPU) {
		return SeparateNPU
	} else if faultTypeSet.Has(PreSeparateNPU) {
		return PreSeparateNPU
	} else if faultTypeSet.Has(RestartNPU) {
		return RestartNPU
	} else if faultTypeSet.Has(FreeRestartNPU) {
		return FreeRestartNPU
	} else if faultTypeSet.Has(RestartBusiness) {
		return RestartBusiness
	} else if faultTypeSet.Has(RestartRequest) {
		return RestartRequest
	} else if faultTypeSet.Has(SubHealthFault) {
		return SubHealthFault
	} else if faultTypeSet.Has(NotHandleFault) {
		return NotHandleFault
	}
	return NormalNPU
}

// SetDeviceInit set should init device's logicID
func SetDeviceInit(logicID int32) {
	logicIDLock.Lock()
	initLogicIDs = append(initLogicIDs, logicID)
	logicIDLock.Unlock()
}

// GetAndCleanLogicID get should init device's logicID and clean cache
func GetAndCleanLogicID() []int32 {
	if len(initLogicIDs) == 0 {
		return nil
	}
	logicIDLock.Lock()
	oldInitLogicIDs := initLogicIDs
	initLogicIDs = []int32{}
	logicIDLock.Unlock()
	return oldInitLogicIDs
}

// setAlarmRaisedTime set `AlarmRaisedTime` by device fault code length
func setAlarmRaisedTime(device *NpuDevice) {
	if len(device.FaultCodes) == 0 {
		device.AlarmRaisedTime = 0
	} else if device.AlarmRaisedTime == 0 {
		device.AlarmRaisedTime = time.Now().UnixMilli()
	}
}

// setNetworkAlarmRaisedTime set `NetworkAlarmRaisedTime` by device network fault code length
func setNetworkAlarmRaisedTime(device *NpuDevice) {
	if len(device.NetworkFaultCodes) == 0 {
		device.NetworkAlarmRaisedTime = 0
	} else if device.NetworkAlarmRaisedTime == 0 {
		device.NetworkAlarmRaisedTime = time.Now().UnixMilli()
	}
}

// SetNewFaultAndCacheOnceRecoverFault set new fault code and cache once recover fault
func SetNewFaultAndCacheOnceRecoverFault(logicID int32, faultInfos []common.DevFaultInfo, device *NpuDevice) {
	if device == nil {
		hwlog.RunLog.Error("param device is nil in SetNewFaultAndCacheOnceRecoverFault")
		return
	}

	if _, ok := faultDurationMap[HbmDoubleBitFaultCodeStr]; ok {
		var newFaultInfos []common.DevFaultInfo
		// dealing with Hbm and Aic/Aiv associated faults
		for i := 0; i < len(faultInfos); i++ {
			if faultInfos[i].EventID == HbmDoubleBitFaultCode && faultInfos[i].Assertion != common.FaultRecover {
				hbmTool.updateHbmOccurTime(faultInfos[i])
			}
			if faultInfos[i].EventID == AicBusFaultCode || faultInfos[i].EventID == AivBusFaultCode {
				hbmTool.aicFaultEventInQue(faultInfos[i])
				continue
			}
			newFaultInfos = append(newFaultInfos, faultInfos[i])
		}
		faultInfos = append(newFaultInfos, hbmTool.aicFaultEventOutQue(logicID)...)
	}

	// it must deal with two 'for', because the fault may recover one moment, in this case,
	// the recover message and occur message both in faultInfos, this fault cannot be reports outside.
	for _, faultInfo := range faultInfos {
		if NetworkFaultCodes.Has(faultInfo.EventID) {
			continue
		}
		if faultInfo.Assertion == common.FaultRecover {
			if Int64Tool.Index(device.FaultCodes, faultInfo.EventID) == -1 {
				recoverFaultMap[logicID] = append(recoverFaultMap[logicID], faultInfo.EventID)
			} else {
				device.FaultCodes = Int64Tool.Remove(device.FaultCodes, faultInfo.EventID)
			}
		}
		if faultInfo.Assertion == common.FaultOnce {
			recoverFaultMap[logicID] = append(recoverFaultMap[logicID], faultInfo.EventID)
		}
	}
	for _, faultInfo := range faultInfos {
		if NetworkFaultCodes.Has(faultInfo.EventID) {
			continue
		}
		if faultInfo.Assertion == common.FaultOccur || faultInfo.Assertion == common.FaultOnce {
			device.FaultCodes = append(device.FaultCodes, faultInfo.EventID)
			eventIdStr := strings.ToLower(strconv.FormatInt(faultInfo.EventID, Hex))
			if _, ok := faultDurationMap[eventIdStr]; !ok {
				insertFaultFrequency(device.LogicID, faultInfo.EventID)
			}
		}
	}
	setAlarmRaisedTime(device)
}

// SetNetworkNewFaultAndCacheOnceRecoverFault set new network fault code and cache once recover network fault
func SetNetworkNewFaultAndCacheOnceRecoverFault(logicID int32, faultInfos []common.DevFaultInfo, device *NpuDevice) {
	if device == nil {
		hwlog.RunLog.Error("param device is nil in SetNetworkNewFaultAndCacheOnceRecoverFault")
		return
	}
	// it must deal with two 'for', because the fault may recover one moment, in this case,
	// the recover message and occur message both in faultInfos, this fault cannot be reports outside.
	networkFaultRecoverAndFaultOnceHandle(logicID, faultInfos, device)
	networkFaultOccurAndFaultOnceHandle(faultInfos, device)
	setNetworkAlarmRaisedTime(device)
}

func networkFaultRecoverAndFaultOnceHandle(logicID int32, faultInfos []common.DevFaultInfo, device *NpuDevice) {
	for _, faultInfo := range faultInfos {
		if !NetworkFaultCodes.Has(faultInfo.EventID) {
			continue
		}
		if faultInfo.Assertion == common.FaultRecover {
			if Int64Tool.Index(device.NetworkFaultCodes, faultInfo.EventID) == -1 {
				recoverNetworkFaultMap[logicID] = append(recoverNetworkFaultMap[logicID], faultInfo.EventID)
			} else {
				device.NetworkFaultCodes = Int64Tool.Remove(device.NetworkFaultCodes, faultInfo.EventID)
			}
		}
		if faultInfo.Assertion == common.FaultOnce {
			recoverNetworkFaultMap[logicID] = append(recoverNetworkFaultMap[logicID], faultInfo.EventID)
		}
	}
}

func networkFaultOccurAndFaultOnceHandle(faultInfos []common.DevFaultInfo, device *NpuDevice) {
	for _, faultInfo := range faultInfos {
		if !NetworkFaultCodes.Has(faultInfo.EventID) {
			continue
		}
		if faultInfo.Assertion == common.FaultOccur || faultInfo.Assertion == common.FaultOnce {
			device.NetworkFaultCodes = append(device.NetworkFaultCodes, faultInfo.EventID)
			eventIdStr := strings.ToLower(strconv.FormatInt(faultInfo.EventID, Hex))
			if _, ok := faultDurationMap[eventIdStr]; !ok {
				insertFaultFrequency(device.LogicID, faultInfo.EventID)
			}
		}
	}
}

// DelOnceRecoverFault delete func 'cacheAfterDelFaultCode' record fault code and network fault code in the end of cycle
func DelOnceRecoverFault(groupDevice map[string][]*NpuDevice) {
	for _, devices := range groupDevice {
		for _, device := range devices {
			recoverFaults := recoverFaultMap[device.LogicID]
			for _, recoverFault := range recoverFaults {
				device.FaultCodes = Int64Tool.Remove(device.FaultCodes, recoverFault)
			}
			setAlarmRaisedTime(device)

			recoverNetworkFaults := recoverNetworkFaultMap[device.LogicID]
			for _, recoverNetworkFault := range recoverNetworkFaults {
				device.NetworkFaultCodes = Int64Tool.Remove(device.NetworkFaultCodes, recoverNetworkFault)
			}
			setNetworkAlarmRaisedTime(device)
		}
	}
	recoverFaultMap = make(map[int32][]int64, GeneralMapSize)
	recoverNetworkFaultMap = make(map[int32][]int64, GeneralMapSize)
}

// DelOnceFrequencyFault clear all the fault occurrence time in cache when frequency
// fault detected at the end of each cycle
func DelOnceFrequencyFault() {
	for logicId, eventId := range recoverFaultFrequencyMap {
		frequencyCache, ok := faultFrequencyMap[eventId]
		if !ok {
			hwlog.RunLog.Warnf("eventId %v is not exist in faultFrequencyMap %v", eventId, faultFrequencyMap)
			return
		}
		frequencyCache.Frequency[logicId] = make([]int64, 0, frequencyCache.Times)
		hwlog.RunLog.Infof("logic id %v frequency cache is successfully cleared", logicId)
	}
	recoverFaultFrequencyMap = make(map[int32]string, GeneralMapSize)
}

// SaveDevFaultInfo save device fault info , subscribe interface call back function
func SaveDevFaultInfo(devFaultInfo common.DevFaultInfo) {
	hwlog.RunLog.Infof("receive devFaultInfo: %v, hex code: %v", devFaultInfo,
		strconv.FormatInt(devFaultInfo.EventID, Hex))
	if devFaultInfo.EventID == 0 {
		return
	}
	if devFaultInfo.EventID == ResetFinishFaultCode {
		SetDeviceInit(devFaultInfo.LogicID)
		return
	}
	faultSeverityMap[devFaultInfo.EventID] = devFaultInfo.Severity
	devFaultInfoMapLock.Lock()
	devFaultInfoMap[devFaultInfo.LogicID] = append(devFaultInfoMap[devFaultInfo.LogicID], devFaultInfo)
	devFaultInfoMapLock.Unlock()
}

// GetAndCleanFaultInfo get device fault info and clean cache
func GetAndCleanFaultInfo() map[int32][]common.DevFaultInfo {
	if len(devFaultInfoMap) == 0 {
		return map[int32][]common.DevFaultInfo{}
	}
	devFaultInfoMapLock.Lock()
	oldDevFaultInfoMap := devFaultInfoMap
	devFaultInfoMap = make(map[int32][]common.DevFaultInfo, GeneralMapSize)
	devFaultInfoMapLock.Unlock()
	return oldDevFaultInfoMap
}

// SaveManuallyFaultInfo save manually fault info into manuallySeparateNpuMap
func SaveManuallyFaultInfo(logicID int32) {
	if logicID < MinLogicID || logicID > MaxLogicID {
		hwlog.RunLog.Warnf("logic id %d is not valid, logic id must be in [0, 15]", logicID)
		return
	}
	manFaultInfo := ManuallyFaultInfo{
		LogicID:     logicID,
		FirstHandle: true,
		RecordTime:  time.Now().UnixMilli(),
	}
	manuallySeparateNpuMapLock.Lock()
	defer manuallySeparateNpuMapLock.Unlock()
	manuallySeparateNpuMap[logicID] = manFaultInfo
	hwlog.RunLog.Debugf("received manually fault info, manually separate npu logic id: %d, first handle: %v, "+
		"manually separate device cache is: %v", manFaultInfo.LogicID, manFaultInfo.FirstHandle, manuallySeparateNpuMap)
}

// QueryManuallyFaultInfoByLogicID query manually fault info based on logic id from manuallySeparateNpuMap
func QueryManuallyFaultInfoByLogicID(logicID int32) bool {
	if logicID < MinLogicID || logicID > MaxLogicID {
		hwlog.RunLog.Warnf("logic id %d is invalid, logic id must be in [0, 15]", logicID)
		return false
	}

	manuallySeparateNpuMapLock.Lock()
	_, ok := manuallySeparateNpuMap[logicID]
	manuallySeparateNpuMapLock.Unlock()
	return ok
}

// QueryManuallyFaultNPULogicIDsByHandleStatus query manually fault npu logic ids
// based on handle status from manuallySeparateNpuMap
func QueryManuallyFaultNPULogicIDsByHandleStatus(handleStatus string) []int32 {
	logicIDs := make([]int32, 0, GeneralMapSize)
	if handleStatus != ManuallySeparateNpuFirstHandle && handleStatus != ManuallySeparateNpuHandled &&
		handleStatus != ManuallySeparateNpuAll {
		hwlog.RunLog.Warnf("manually fault npu handle status %v is invalid, it must be in [%v,%v,%v]", handleStatus,
			ManuallySeparateNpuFirstHandle, ManuallySeparateNpuHandled, ManuallySeparateNpuAll)
		return logicIDs
	}

	manuallySeparateNpuMapLock.Lock()
	defer manuallySeparateNpuMapLock.Unlock()

	switch {
	case handleStatus == ManuallySeparateNpuFirstHandle:
		for _, manuallySeparateNpu := range manuallySeparateNpuMap {
			if manuallySeparateNpu.FirstHandle {
				logicIDs = append(logicIDs, manuallySeparateNpu.LogicID)
			}
		}
		break
	case handleStatus == ManuallySeparateNpuHandled:
		for _, manuallySeparateNpu := range manuallySeparateNpuMap {
			if !manuallySeparateNpu.FirstHandle {
				logicIDs = append(logicIDs, manuallySeparateNpu.LogicID)
			}
		}
		break
	default:
		for _, manuallySeparateNpu := range manuallySeparateNpuMap {
			logicIDs = append(logicIDs, manuallySeparateNpu.LogicID)
		}
	}

	return logicIDs
}

// SetManuallyFaultNPUHandled set manually fault NPU handled
func SetManuallyFaultNPUHandled() {
	manuallySeparateNpuMapLock.Lock()
	defer manuallySeparateNpuMapLock.Unlock()

	for logicId, manuallyFaultInfo := range manuallySeparateNpuMap {
		manuallyFaultInfo.FirstHandle = false
		manuallySeparateNpuMap[logicId] = manuallyFaultInfo
	}
}

// DeleteManuallyFaultInfo delete manually fault info from manuallySeparateNpuMap
func DeleteManuallyFaultInfo(logicID int32) {
	if logicID < MinLogicID || logicID > MaxLogicID {
		hwlog.RunLog.Warnf("logic id %d not valid, must be in [0, 15]", logicID)
		return
	}

	manuallySeparateNpuMapLock.Lock()
	defer manuallySeparateNpuMapLock.Unlock()

	if deleteManuallySeparateFaultInfo, ok := manuallySeparateNpuMap[logicID]; ok {
		delete(manuallySeparateNpuMap, logicID)
		hwlog.RunLog.Infof("device logic id %v, manually fault info %v has been removed, manually separate device "+
			"cache: %v", logicID, deleteManuallySeparateFaultInfo, manuallySeparateNpuMap)
	} else {
		hwlog.RunLog.Warnf("device logic id %v manually fault info not exist, no need to remove", logicID)
	}
}

// CountFaultDuration used to calculate each fault duration
func CountFaultDuration(device *NpuDevice, devFaultInfoMap map[int32][]common.DevFaultInfo) {
	// Collect fault events from fault event queue cache to form the fault queue for duration statistics
	collectEachFaultEvent(device.LogicID, devFaultInfoMap[device.LogicID])
	faultDurationMapLock.Lock()
	defer faultDurationMapLock.Unlock()

	for eventId, _ := range faultDurationMap {
		// Sort fault events in the fault queue in ascending order based on fault event AlarmRaisedTime
		sortFaultEventsInAscendingOrder(device.LogicID, eventId)

		// Merge consecutive fault events by fault event assertion in the fault queue
		// and clear first event according to the fault status of the current fault code
		cleanFaultQueue(device.LogicID, eventId)

		// update the fault code timeout status, fault duration time, fault recover duration time
		// and clear fault queue cache through timeout judgment and recovery judgment algorithm
		handleFaultQueue(device.LogicID, eventId)
	}
}

func collectEachFaultEvent(logicId int32, faultInfos []common.DevFaultInfo) {
	faultDurationMapLock.Lock()
	defer faultDurationMapLock.Unlock()

	for _, faultInfo := range faultInfos {
		eventIdStr := strings.ToLower(strconv.FormatInt(faultInfo.EventID, Hex))
		if _, ok := faultDurationMap[eventIdStr]; !ok {
			continue
		}

		if faultDurationMap[eventIdStr].Duration == nil {
			faultDurationMap[eventIdStr].Duration = make(map[int32]FaultDurationData, GeneralMapSize)
		}

		if _, ok := faultDurationMap[eventIdStr].Duration[logicId]; !ok {
			faultDurationMap[eventIdStr].Duration[logicId] = FaultDurationData{
				FaultEventQueue: []common.DevFaultInfo{}, // Initializing the slice
			}
		}
		faultDurationData := faultDurationMap[eventIdStr].Duration[logicId]
		faultDurationData.FaultEventQueue = append(faultDurationData.FaultEventQueue, faultInfo)
		faultDurationMap[eventIdStr].Duration[logicId] = faultDurationData
	}
}

func sortFaultEventsInAscendingOrder(logicID int32, eventId string) {
	if _, ok := faultDurationMap[eventId]; !ok {
		return
	}
	if _, ok := faultDurationMap[eventId].Duration[logicID]; !ok {
		return
	}

	faultQueue := faultDurationMap[eventId].Duration[logicID].FaultEventQueue
	sort.Sort(DevFaultInfoBasedTimeAscend(faultQueue))
}

func cleanFaultQueue(logicID int32, eventId string) {
	if _, ok := faultDurationMap[eventId]; !ok {
		return
	}
	if _, ok := faultDurationMap[eventId].Duration[logicID]; !ok {
		return
	}

	faultDurationData := faultDurationMap[eventId].Duration[logicID]
	mergeContinuousElementBasedAssertion(&faultDurationData.FaultEventQueue)
	clearFirstEventBasedOnFaultStatus(&faultDurationData)
	faultDurationMap[eventId].Duration[logicID] = faultDurationData
	hwlog.RunLog.Debugf("NPU logic id: %d, %s fault timeout status: %v, fault queue after sort and merge: %v",
		logicID, eventId, faultDurationMap[eventId].Duration[logicID].TimeoutStatus,
		faultDurationMap[eventId].Duration[logicID].FaultEventQueue)
}

// mergeContinuousElementBasedAssertion merge continuous element based on assertion
func mergeContinuousElementBasedAssertion(devFaultInfo *[]common.DevFaultInfo) {
	if devFaultInfo == nil || len(*devFaultInfo) == 0 {
		return
	}

	previousEvent := (*devFaultInfo)[0]
	newDevFaultInfo := []common.DevFaultInfo{previousEvent}
	for i := 1; i < len(*devFaultInfo); i++ {
		currentEvent := (*devFaultInfo)[i]
		if currentEvent.Assertion == previousEvent.Assertion {
			continue
		}
		previousEvent = currentEvent
		newDevFaultInfo = append(newDevFaultInfo, currentEvent)
	}
	*devFaultInfo = newDevFaultInfo
}

func clearFirstEventBasedOnFaultStatus(faultDurationData *FaultDurationData) {
	// If the first fault event assertion is fault recover in fault queue when the fault status is healthy,
	// clear the first fault event
	if !faultDurationData.TimeoutStatus && len(faultDurationData.FaultEventQueue) > 0 &&
		faultDurationData.FaultEventQueue[0].Assertion == common.FaultRecover {
		faultDurationData.FaultEventQueue = faultDurationData.FaultEventQueue[1:]
	}

	// If the first fault event assertion is fault occur in fault queue when the fault status is unhealthy,
	// clear the first fault event
	if faultDurationData.TimeoutStatus && len(faultDurationData.FaultEventQueue) > 0 &&
		faultDurationData.FaultEventQueue[0].Assertion == common.FaultOccur {
		faultDurationData.FaultEventQueue = faultDurationData.FaultEventQueue[1:]
	}
}

func handleFaultQueue(logicID int32, eventId string) {
	if _, ok := faultDurationMap[eventId]; !ok {
		return
	}
	if _, ok := faultDurationMap[eventId].Duration[logicID]; !ok {
		return
	}
	faultDurationData := faultDurationMap[eventId].Duration[logicID]
	if len(faultDurationData.FaultEventQueue) == 0 {
		hwlog.RunLog.Debugf("NPU logic id: %v, %v fault queue is empty, no need to handle fault queue",
			logicID, eventId)
		return
	}

	initTimeoutStatus := faultDurationData.TimeoutStatus
	exitTag := false
	for !exitTag {
		faultDurationData = faultDurationMap[eventId].Duration[logicID]
		exitTag = timeoutOrRecoveryAlgorithm(logicID, eventId, !faultDurationData.TimeoutStatus)
	}
	faultDurationData = faultDurationMap[eventId].Duration[logicID]
	hwlog.RunLog.Debugf("NPU logic id: %v, after timeout or recovery algorithm handling, %v fault timeout "+
		"status is %v, fault duration time is %.2f seconds, fault recover duration time is %.2f seconds, "+
		"fault queue is %v", logicID, eventId, faultDurationData.TimeoutStatus,
		float64(faultDurationData.FaultDurationTime)/SecondMagnificationFloat,
		float64(faultDurationData.FaultRecoverDurationTime)/SecondMagnificationFloat,
		faultDurationData.FaultEventQueue)

	if initTimeoutStatus == false && faultDurationData.TimeoutStatus == true {
		num, err := strconv.ParseInt(eventId, Hex, 0)
		if err != nil {
			hwlog.RunLog.Errorf(parseHexFailedMsg, eventId)
			return
		}
		insertFaultFrequency(logicID, num)
	}

	var duration int64
	if faultDurationData.TimeoutStatus {
		duration = faultDurationData.FaultDurationTime
	} else {
		duration = faultDurationData.FaultRecoverDurationTime
	}
	if initTimeoutStatus != faultDurationData.TimeoutStatus {
		hwlog.RunLog.Infof("NPU logic id: %v, after timeout or recovery algorithm handling, %v fault timeout "+
			"status change, now fault timeout status set %v, duration time is %.2f seconds",
			logicID, eventId, faultDurationData.TimeoutStatus, float64(duration)/SecondMagnificationFloat)
	}
}

func timeoutOrRecoveryAlgorithm(logicID int32, eventId string, timeoutStatus bool) bool {
	process := getProcessInFaultDuration(timeoutStatus)
	faultQueueLen := len(faultDurationMap[eventId].Duration[logicID].FaultEventQueue)
	if faultQueueLen == 0 {
		hwlog.RunLog.Debugf("NPU logic id: %v, %v fault queue is empty, no need to do %v judgment", logicID,
			eventId, process)
		return true
	}
	var i int
	var duration int64
	timeoutThreshold := getTimeoutThreshold(eventId, timeoutStatus)
	faultTimeoutMsg := "NPU logic id: %v, in %v judgment, %v duration is %.2f seconds > %v seconds, %v fault " +
		"timeout status set %v"
	faultNotTimeoutMsg := "NPU logic id: %v, in %v judgment, %v duration is %.2f seconds <= %v seconds, %v " +
		"fault timeout status %v doesn't need to change, continue to perform %v judgment"
	for i = 0; i < faultQueueLen/halfDivisor; i++ {
		faultDurationData := faultDurationMap[eventId].Duration[logicID]
		duration = faultDurationData.FaultEventQueue[i*halfDivisor+1].AlarmRaisedTime -
			faultDurationData.FaultEventQueue[i*halfDivisor].AlarmRaisedTime
		if duration <= timeoutThreshold*SecondMagnification {
			continue
		}
		hwlog.RunLog.Debugf(faultTimeoutMsg, logicID, process, process, float64(duration)/SecondMagnificationFloat,
			timeoutThreshold, eventId, timeoutStatus)
		return handleTimeoutCondition(handleDurationInputPara{logicID: logicID, eventId: eventId, index: i,
			timeoutStatus: timeoutStatus, duration: duration})
	}
	if i*halfDivisor+1 == faultQueueLen {
		faultDurationData := faultDurationMap[eventId].Duration[logicID]
		currentHostTime := time.Now().UnixMilli()
		duration = currentHostTime - faultDurationData.FaultEventQueue[i*halfDivisor].AlarmRaisedTime
		if duration <= timeoutThreshold*SecondMagnification {
			hwlog.RunLog.Debugf(faultNotTimeoutMsg, logicID, process, process, float64(duration)/
				SecondMagnificationFloat, timeoutThreshold, eventId, faultDurationData.TimeoutStatus, process)
			return handleNotTimeoutCondition(handleDurationInputPara{logicID: logicID, eventId: eventId, index: i,
				timeoutStatus: timeoutStatus, duration: duration})
		}
		hwlog.RunLog.Debugf(faultTimeoutMsg, logicID, process, process, float64(duration)/SecondMagnificationFloat,
			timeoutThreshold, eventId, timeoutStatus)
		return handleTimeoutCondition(handleDurationInputPara{logicID: logicID, eventId: eventId, index: i,
			timeoutStatus: timeoutStatus, duration: duration})
	}
	if halfDivisor*i == faultQueueLen {
		hwlog.RunLog.Debugf(faultNotTimeoutMsg, logicID, process, process, float64(duration)/SecondMagnificationFloat,
			timeoutThreshold, eventId, faultDurationMap[eventId].Duration[logicID].TimeoutStatus, process)
		return handleNotTimeoutCondition(handleDurationInputPara{logicID: logicID, eventId: eventId, index: i,
			timeoutStatus: timeoutStatus, duration: duration})
	}
	return true
}

func getProcessInFaultDuration(timeoutStatus bool) string {
	if timeoutStatus {
		return TimeoutProcess
	}
	return TimeoutRecoverProcess
}

func getTimeoutThreshold(eventId string, timeoutStatus bool) int64 {
	if _, ok := faultDurationMap[eventId]; !ok {
		return MinFaultTimeout
	}

	if timeoutStatus {
		return faultDurationMap[eventId].FaultDuration.FaultTimeout
	}
	return faultDurationMap[eventId].FaultDuration.RecoverTimeout
}

func handleTimeoutCondition(inputPara handleDurationInputPara) bool {
	faultDurationData := faultDurationMap[inputPara.eventId].Duration[inputPara.logicID]
	faultDurationData.TimeoutStatus = inputPara.timeoutStatus
	faultQueueMsg := "NPU logic id: %v, %v fault queue: %v"
	if inputPara.timeoutStatus {
		faultDurationData.FaultDurationTime = inputPara.duration
		faultDurationMap[inputPara.eventId].Duration[inputPara.logicID] = faultDurationData
		hwlog.RunLog.Debugf(faultQueueMsg, inputPara.logicID, inputPara.eventId, faultDurationData.FaultEventQueue)
		return true
	}
	faultDurationData.FaultRecoverDurationTime = inputPara.duration
	faultDurationData.FaultEventQueue = faultDurationData.FaultEventQueue[halfDivisor*inputPara.index+1:]
	faultDurationMap[inputPara.eventId].Duration[inputPara.logicID] = faultDurationData
	hwlog.RunLog.Debugf(faultQueueMsg, inputPara.logicID, inputPara.eventId, faultDurationData.FaultEventQueue)
	return false
}

func handleNotTimeoutCondition(inputPara handleDurationInputPara) bool {
	faultDurationData := faultDurationMap[inputPara.eventId].Duration[inputPara.logicID]
	if inputPara.timeoutStatus {
		faultDurationData.FaultDurationTime = inputPara.duration
	} else {
		faultDurationData.FaultRecoverDurationTime = inputPara.duration
	}

	faultDurationData.FaultEventQueue = faultDurationData.FaultEventQueue[halfDivisor*inputPara.index:]
	faultDurationMap[inputPara.eventId].Duration[inputPara.logicID] = faultDurationData
	hwlog.RunLog.Debugf("NPU logic id: %v, %v fault queue: %v", inputPara.logicID, inputPara.eventId,
		faultDurationData.FaultEventQueue)
	return true
}

// GetFaultAssertionName get assertion name of fault code
func GetFaultAssertionName(assertion int8) string {
	switch assertion {
	case common.FaultRecover:
		return AssertionRecovery
	case common.FaultOccur:
		return AssertionOccur
	case common.FaultOnce:
		return AssertionNotice
	default:
		return ""
	}
}

// GetChangedDevFaultInfo get device changed fault info
func GetChangedDevFaultInfo(device *NpuDevice, oldErrCodes []int64, newErrCodes []int64) []common.DevFaultInfo {
	devFaultInfo := make([]common.DevFaultInfo, 0, len(newErrCodes))
	for _, newCode := range newErrCodes {
		if Int64Tool.Index(oldErrCodes, newCode) == -1 {
			faultInfo := common.DevFaultInfo{
				EventID:         newCode,
				LogicID:         device.LogicID,
				Assertion:       common.FaultOccur,
				AlarmRaisedTime: time.Now().UnixMilli(),
			}
			devFaultInfo = append(devFaultInfo, faultInfo)
		}
	}
	for _, oldCode := range oldErrCodes {
		if Int64Tool.Index(newErrCodes, oldCode) == -1 {
			faultInfo := common.DevFaultInfo{
				EventID:         oldCode,
				LogicID:         device.LogicID,
				Assertion:       common.FaultRecover,
				AlarmRaisedTime: time.Now().UnixMilli(),
			}
			devFaultInfo = append(devFaultInfo, faultInfo)
		}
	}
	return devFaultInfo
}

// CheckErrorMessage check whether the error message contains a specific string
func CheckErrorMessage(err error, target string) bool {
	return err != nil && strings.Contains(err.Error(), target)
}

// GetTimeoutFaultCodes get timeout fault codes
func GetTimeoutFaultCodes(mode string) []int64 {
	faultCodes := make([]int64, 0)
	if mode != ChipFaultMode && mode != NetworkFaultMode {
		return faultCodes
	}

	faultDurationMapLock.Lock()
	defer faultDurationMapLock.Unlock()

	for eventId, faultDurationCache := range faultDurationMap {
		num, err := strconv.ParseInt(eventId, Hex, 0)
		if err != nil {
			hwlog.RunLog.Errorf(parseHexFailedMsg, eventId)
			continue
		}
		if (mode == ChipFaultMode && NetworkFaultCodes.Has(num)) ||
			(mode == NetworkFaultMode && !NetworkFaultCodes.Has(num)) {
			continue
		}

		for _, faultDurationData := range faultDurationCache.Duration {
			if faultDurationData.TimeoutStatus {
				faultCodes = append(faultCodes, num)
			}
		}
	}

	return faultCodes
}
