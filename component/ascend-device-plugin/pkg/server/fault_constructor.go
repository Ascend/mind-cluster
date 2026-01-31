/* Copyright(C) 2024. Huawei Technologies Co.,Ltd. All rights reserved.
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
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"k8s.io/api/core/v1"

	"Ascend-device-plugin/pkg/common"
	"Ascend-device-plugin/pkg/kubeclient"
	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
	npuCommon "ascend-common/devmanager/common"
)

const (
	// FaultEventCMName name of npu fault event configmap
	FaultEventCMName = "mindx-dl-npu-fault-event"
	// FaultEventCMNameSpace namespace of npu fault event configmap
	FaultEventCMNameSpace = "kube-system"
	// FaultEventFileKey key of loading npu faults
	FaultEventFileKey = "npuFaultCM.json"
	// FaultEventCMPollSecInterval interval of polling npu fault event configmap, unit:second
	FaultEventCMPollSecInterval = 1
	// FaultCacheSaveToDPMillInterval interval of saving cached npu fault to DP, unit:millisecond
	FaultCacheSaveToDPMillInterval = 500
	// ReInjectAllFaultsDefaultValue default value of re-injecting all faults in configmap
	ReInjectAllFaultsDefaultValue = 1
	// FaultEventFileAbsPath file absolute path of injecting fault event with file
	FaultEventFileAbsPath = "/user/inject/fault/npuFaultFile.json"
)

var (
	// faultCacheLock is used for devFaultCache which may be used concurrence
	faultCacheLock   sync.Mutex
	devFaultCache    []npuCommon.DevFaultInfo
	switchFaultCache []common.SwitchFaultEvent
)

type FaultInfo struct {
	EventID    string
	LogicID    int32
	Severity   int8
	Assertion  int8
	TimeOffset []int64
}

type SwitchFaultInfo struct {
	AssembledFaultCode string
	Assertion          uint
	SwitchChipId       uint
	SwitchPortId       uint
	TimeOffset         []int64
}

type FaultDebugConfig struct {
	Node         string // When injecting faults through local files, this field does not work
	PollInterval int64
	ReInject     int
	Faults       []FaultInfo
	SwitchFaults []SwitchFaultInfo
}

func (hdm *HwDevManager) constructNpuFaultByCm(ctx context.Context) {
	hwlog.RunLog.Infof("start construct npu fault from cm or file")
	if err := hdm.createFaultFile(); err != nil {
		hwlog.RunLog.Errorf("create fault file fail, err: %v", err)
	} else {
		go hdm.loadFaultEventFromFile(ctx)
	}
	go hdm.pollFaultEventFromCm(ctx)
	go hdm.saveCachedFaultToDP(ctx)
}

func (hdm *HwDevManager) createFaultFile() error {
	dir := filepath.Dir(FaultEventFileAbsPath)
	if !utils.IsExist(dir) {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return fmt.Errorf("mkdir fail, err: %v", err)
		}
	}
	defaultConfig := &FaultDebugConfig{
		PollInterval: FaultEventCMPollSecInterval,
		ReInject:     0,
	}
	return hdm.updateFaultInjectFile(defaultConfig)
}

func (hdm *HwDevManager) loadFaultEventFromFile(ctx context.Context) {
	for {
		select {
		case _, ok := <-ctx.Done():
			if !ok {
				hwlog.RunLog.Info("stop signal channel closed")
			}
			hwlog.RunLog.Info("load fault event from file stop")
			return
		default:
			interval := int64(FaultEventCMPollSecInterval)
			config := hdm.readAndInjectFaultFromFile()
			if config != nil && config.PollInterval > 0 {
				interval = config.PollInterval
			}
			time.Sleep(time.Duration(interval) * time.Second)
		}
	}
}

func (hdm *HwDevManager) pollFaultEventFromCm(ctx context.Context) {
	for {
		select {
		case _, ok := <-ctx.Done():
			if !ok {
				hwlog.RunLog.Info("stop signal channel closed")
			}
			hwlog.RunLog.Info("poll fault event from cm stop")
			return
		default:
			interval := int64(FaultEventCMPollSecInterval)
			config := hdm.pollAndInjectFaultFromCm()
			if config != nil && config.PollInterval > 0 {
				interval = config.PollInterval
			}
			time.Sleep(time.Duration(interval) * time.Second)
		}
	}
}

func (hdm *HwDevManager) saveCachedFaultToDP(ctx context.Context) {
	for {
		select {
		case _, ok := <-ctx.Done():
			if !ok {
				hwlog.RunLog.Info("stop signal channel closed")
			}
			hwlog.RunLog.Info("save cached fault to dp stop")
			return
		default:
			hdm.injectDevFaultToDp()
			hdm.injectSwitchFaultToDp()
			time.Sleep(time.Duration(FaultCacheSaveToDPMillInterval) * time.Millisecond)
		}
	}
}

func (hdm *HwDevManager) readAndInjectFaultFromFile() *FaultDebugConfig {
	config, err := readFaultDebugFileJson()
	if err != nil {
		hwlog.RunLog.ErrorfWithLimit(FaultEventFileAbsPath, 1, "cannot load fault from '%s' file, reason: %v", FaultEventFileAbsPath, err)
		return nil
	}
	if config.ReInject != ReInjectAllFaultsDefaultValue {
		return config
	}

	hwlog.RunLog.Infof("ReInject value is '%d' in file, start saving to DP", config.ReInject)
	// reset devFaultCache
	hdm.updateDevFaultCache(config.Faults)
	config.ReInject = 0

	hdm.updateFaultInjectFile(config)
	return config
}

func (hdm *HwDevManager) pollAndInjectFaultFromCm() *FaultDebugConfig {

	configMap, err := hdm.manager.GetKubeClient().GetConfigMap(FaultEventCMName, FaultEventCMNameSpace)
	if err != nil {
		hwlog.RunLog.ErrorfWithLimit(FaultEventCMName, 2, "cannot find '%s' configmap, reason: %v", FaultEventCMName, err)
		return nil
	}

	config, err := parseFaultDebugConfigJson(configMap)
	if err != nil || config == nil {
		hwlog.RunLog.Error(err)
		return nil
	}

	if config.ReInject != ReInjectAllFaultsDefaultValue {
		return config
	}
	hwlog.RunLog.Infof("ReInject value is '%d' in CM, start saving to DP", config.ReInject)

	node, err := kubeclient.GetNodeNameFromEnv()
	if err != nil || node == "" {
		hwlog.RunLog.Errorf("cannot get node from env, reason: %v", err)
		return config
	}

	if node != config.Node {
		hwlog.RunLog.Infof("dont have node '%s' in configmap, target nodes: %s", node, config.Node)
		return config
	}

	// reset devFaultCache
	hdm.updateDevFaultCache(config.Faults)
	hdm.updateSwitchFaultCache(config.SwitchFaults)
	config.ReInject = 0

	hdm.updateConfigMap(config, configMap)

	return config
}

func (hdm *HwDevManager) updateSwitchFaultCache(faultInfos []SwitchFaultInfo) {
	tempSwitchFaultCache := make([]common.SwitchFaultEvent, 0)
	now := time.Now()

	// save npu device fault
	for _, fault := range faultInfos {
		if len(fault.TimeOffset) == 0 {
			fault.TimeOffset = append(fault.TimeOffset, 0)
		}
		for _, offset := range fault.TimeOffset {
			raisedTime := now.Add(time.Duration(offset) * time.Second)

			switchFault := common.SwitchFaultEvent{
				AssembledFaultCode: fault.AssembledFaultCode,
				Assertion:          fault.Assertion,
				SwitchChipId:       fault.SwitchChipId,
				SwitchPortId:       fault.SwitchPortId,
				AlarmRaisedTime:    raisedTime.UnixMilli(),
			}
			tempSwitchFaultCache = append(tempSwitchFaultCache, switchFault)
			hwlog.RunLog.Infof("add switch fault to dp cache, switchFault: %v, AssembledFaultCode code: %v",
				switchFault, fault.AssembledFaultCode)
		}
	}

	faultCacheLock.Lock()
	hwlog.RunLog.Infof("update switch cache fault data finished, pre fault cnt: %d, latest fault count: %d",
		len(switchFaultCache), len(tempSwitchFaultCache))
	switchFaultCache = tempSwitchFaultCache
	faultCacheLock.Unlock()
}

func (hdm *HwDevManager) updateDevFaultCache(faultInfos []FaultInfo) {
	tempDevFaultCache := make([]npuCommon.DevFaultInfo, 0)
	now := time.Now()

	// save npu device fault
	for _, fault := range faultInfos {
		eventId, err := convertFaultCodeHexToInt(fault.EventID)
		if err != nil {
			hwlog.RunLog.Errorf("get fault code fail, reason: %v", err)
			continue
		}
		if len(fault.TimeOffset) == 0 {
			fault.TimeOffset = append(fault.TimeOffset, 0)
		}
		for _, offset := range fault.TimeOffset {
			rasedTime := now.Add(time.Duration(offset) * time.Second)

			devFault := npuCommon.DevFaultInfo{
				EventID:         eventId,
				LogicID:         fault.LogicID,
				Severity:        fault.Severity,
				Assertion:       fault.Assertion,
				AlarmRaisedTime: rasedTime.UnixMilli(),
			}
			tempDevFaultCache = append(tempDevFaultCache, devFault)
			hwlog.RunLog.Infof("add npu fault to dp cache, devFaultInfo: %v, hex code: %v",
				devFault, strconv.FormatInt(devFault.EventID, common.Hex))
		}
	}

	faultCacheLock.Lock()
	hwlog.RunLog.Infof("update cache fault data finished, pre fault cnt: %d, latest fault count: %d",
		len(devFaultCache), len(tempDevFaultCache))
	devFaultCache = tempDevFaultCache
	faultCacheLock.Unlock()
}

func (hdm *HwDevManager) injectDevFaultToDp() {
	nowTime := time.Now().UnixMilli()
	newDevFaultCache := make([]npuCommon.DevFaultInfo, 0)
	for _, devFault := range devFaultCache {
		if nowTime >= devFault.AlarmRaisedTime {
			common.SaveDevFaultInfo(devFault)
			continue
		}
		newDevFaultCache = append(newDevFaultCache, devFault)
	}
	faultCacheLock.Lock()
	defer faultCacheLock.Unlock()
	devFaultCache = newDevFaultCache
}

func (hdm *HwDevManager) injectSwitchFaultToDp() {
	nowTime := time.Now().UnixMilli()
	newSwitchFaultCache := make([]common.SwitchFaultEvent, 0)
	for _, switchFault := range switchFaultCache {
		if nowTime >= switchFault.AlarmRaisedTime {
			doInjectSwitchFaultToDp(switchFault)
			continue
		}
		newSwitchFaultCache = append(newSwitchFaultCache, switchFault)
	}
	faultCacheLock.Lock()
	defer faultCacheLock.Unlock()
	switchFaultCache = newSwitchFaultCache
}

func doInjectSwitchFaultToDp(switchFault common.SwitchFaultEvent) {
	// for recovered fault, delete them from current fault codes
	if int8(switchFault.Assertion) == npuCommon.FaultRecover {
		newFaultCodes := make([]common.SwitchFaultEvent, 0)
		for _, errInfo := range common.GetSwitchFaultCode() {
			// only in faultEvent and recoverEvent major info is the same it will be thought as recover
			if !isFaultRecoveredEvent(errInfo, switchFault) {
				newFaultCodes = append(newFaultCodes, errInfo)
			} else {
				hwlog.RunLog.Infof("switch fault recover, errInfo: %v", errInfo)
			}
		}
		common.SetSwitchFaultCode(newFaultCodes)
		return
	}
	currentFault := common.GetSwitchFaultCode()
	common.SetSwitchFaultCode(append(currentFault, switchFault))
}

func isFaultRecoveredEvent(faultEvent, recoverEvent common.SwitchFaultEvent) bool {
	if int8(recoverEvent.Assertion) != npuCommon.FaultRecover || recoverEvent.Assertion == faultEvent.Assertion {
		return false
	}
	faultEventInfo := fmt.Sprintf("AssembledFaultCode:%v,SwitchChipId:%v,SwitchPortId:%v",
		faultEvent.AssembledFaultCode, faultEvent.SwitchChipId, faultEvent.SwitchPortId)
	recoveredEventInfo := fmt.Sprintf("AssembledFaultCode:%v,SwitchChipId:%v,SwitchPortId:%v",
		recoverEvent.AssembledFaultCode, recoverEvent.SwitchChipId, recoverEvent.SwitchPortId)
	return faultEventInfo == recoveredEventInfo
}

func (hdm *HwDevManager) updateConfigMap(config *FaultDebugConfig, configMap *v1.ConfigMap) {
	configBytes, err := json.Marshal(*config)
	if err != nil {
		hwlog.RunLog.Errorf("marshal FaultDebugConfig fail, data: %v reason: %v", config, err)
		return
	}
	configMap.Data[FaultEventFileKey] = string(configBytes)
	_, err = hdm.manager.GetKubeClient().UpdateConfigMap(configMap)
	if err != nil {
		hwlog.RunLog.Errorf("update '%s' configmap fail, reason: %v", FaultEventCMName, err)
	}
}

func (hdm *HwDevManager) updateFaultInjectFile(config *FaultDebugConfig) error {
	configBytes, err := json.Marshal(*config)
	if err != nil {
		hwlog.RunLog.Errorf("marshal FaultDebugConfig fail, data: %v err: %v", config, err)
		return fmt.Errorf("marshal FaultDebugConfig fail, data: %v err: %v", config, err)
	}
	f, err := os.OpenFile(FaultEventFileAbsPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		hwlog.RunLog.Errorf("open fault file failed, reason: %v", err)
		return fmt.Errorf("open fault file failed, reason: %v", err)
	}
	defer f.Close()
	if _, err = f.WriteString(string(configBytes)); err != nil {
		hwlog.RunLog.Errorf("write fault file failed, reason: %v", err)
		return fmt.Errorf("write fault file failed, reason: %v", err)
	}
	return nil
}

func parseFaultDebugConfigJson(configMap *v1.ConfigMap) (*FaultDebugConfig, error) {
	jsonStr, ok := configMap.Data[FaultEventFileKey]
	if !ok {
		return nil, fmt.Errorf("cannot find data '%s' in CM'", FaultEventFileKey)
	}
	return convertByteToFaultDebugConfig([]byte(jsonStr))
}

func readFaultDebugFileJson() (*FaultDebugConfig, error) {
	faultCodeBytes, err := utils.LoadFile(FaultEventFileAbsPath)
	if err != nil {
		return nil, fmt.Errorf("load fault event json file failed, path: %v, reason: %v", FaultEventFileAbsPath, err)
	}
	if faultCodeBytes == nil {
		return nil, errors.New("the file does not exist or for other reasons, the read data is empty")
	}
	return convertByteToFaultDebugConfig(faultCodeBytes)
}

func convertByteToFaultDebugConfig(bytes []byte) (*FaultDebugConfig, error) {
	configInfo := &FaultDebugConfig{
		PollInterval: FaultEventCMPollSecInterval,
	}
	if err := json.Unmarshal(bytes, configInfo); err != nil {
		return nil, fmt.Errorf("cannot unmarshal json data, data: %s, reason: %v", string(bytes), err)
	}
	return configInfo, nil
}

func convertFaultCodeHexToInt(hexStr string) (int64, error) {
	hexStr = strings.TrimPrefix(hexStr, "0x")
	codes := common.StringTool.HexStringToInt([]string{hexStr})
	if len(codes) == 0 {
		return -1, fmt.Errorf("convert fault code hex string '%s' to int failed", hexStr)
	}
	return codes[0], nil
}
