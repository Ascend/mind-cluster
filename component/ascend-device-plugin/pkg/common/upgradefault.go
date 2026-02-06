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

// Package common a series of common function
package common

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"ascend-common/common-utils/hwlog"
)

func init() {
	upgradeFaultCacheMgr = UpgradeFaultCacheManager{
		cache:        make(UpgradeFaultReasonMap[LogicId]),
		cacheLock:    sync.Mutex{},
		removedEvent: make(UpgradeFaultReasonMap[LogicId]),
	}
}

var upgradeFaultCacheMgr UpgradeFaultCacheManager

type UpgradeFaultCacheManager struct {
	cache        UpgradeFaultReasonMap[LogicId]
	cacheLock    sync.Mutex
	removedEvent UpgradeFaultReasonMap[LogicId]
}

// SaveUpgradeFaultCache use when device-plugin boot, load reason from cm then save in cache
func SaveUpgradeFaultCache(cache UpgradeFaultReasonMap[LogicId]) {
	upgradeFaultCacheMgr.cacheLock.Lock()
	defer upgradeFaultCacheMgr.cacheLock.Unlock()
	upgradeFaultCacheMgr.cache = cache
}

// InsertUpgradeFaultCache update upgrade fault cache
func InsertUpgradeFaultCache(logicId LogicId, faultTime int64, faultCode, faultLevel string, upgradeType UpgradeTypeEnum) {
	upgradeFaultCacheMgr.cacheLock.Lock()
	defer upgradeFaultCacheMgr.cacheLock.Unlock()

	updated := upgradeFaultCacheMgr.cache.UpdateReason(logicId, faultTime, faultCode, faultLevel, upgradeType)
	if updated {
		hwlog.RunLog.Infof("UpdateUpgradeFaultCache logicId %v, faultTime %v, faultCode %v, faultLevel %v",
			logicId, faultTime, faultCode, faultLevel)
	}
}

// CheckUpgradeFaultCache according to logicId, faultCode, faultLevel, upgradeType
func CheckUpgradeFaultCache(logicId LogicId, faultCode, faultLevel string, upgradeType UpgradeTypeEnum) bool {
	upgradeFaultCacheMgr.cacheLock.Lock()
	defer upgradeFaultCacheMgr.cacheLock.Unlock()
	for key := range upgradeFaultCacheMgr.cache[logicId] {
		if key.FaultCode == faultCode && key.FaultLevel == faultLevel && key.UpgradeType == upgradeType {
			return true
		}
	}
	return false
}

// RemoveManuallySeparateReasonCache when cm remove manually separate npu, the cache should remove reported npu
// but the fault that has not been reported shouldn't be removed
func RemoveManuallySeparateReasonCache(logicIds []LogicId) {
	upgradeFaultCacheMgr.cacheLock.Lock()
	defer upgradeFaultCacheMgr.cacheLock.Unlock()
	for _, id := range logicIds {
		removedReasons := upgradeFaultCacheMgr.cache.removeFaultLevel(id, ManuallySeparateNPU)
		if len(removedReasons) > 0 {
			upgradeFaultCacheMgr.removedEvent.addReasons(id, removedReasons)
			hwlog.RunLog.Infof(
				"remove manually separate reason, logic %v, reason %v", id, removedReasons.toString())
		}
	}
}

// RemoveTimeoutReasonCache when release timeout window reach then reach them from cache
func RemoveTimeoutReasonCache(logic LogicId, faultCode string) {
	upgradeFaultCacheMgr.cacheLock.Lock()
	defer upgradeFaultCacheMgr.cacheLock.Unlock()
	removedReasons := upgradeFaultCacheMgr.cache.removeFaultCode(logic, faultCode)
	if len(removedReasons) > 0 {
		upgradeFaultCacheMgr.removedEvent.addReasons(logic, removedReasons)
		hwlog.RunLog.Infof(
			"remove timeout reason, logic %v, reason %v", logic, removedReasons.toString())
	}
}

// GetAndCleanRemovedReasonEvent get and clean removed reason when notify to k8s event
func GetAndCleanRemovedReasonEvent() UpgradeFaultReasonMap[LogicId] {
	upgradeFaultCacheMgr.cacheLock.Lock()
	defer upgradeFaultCacheMgr.cacheLock.Unlock()
	res := upgradeFaultCacheMgr.removedEvent.copy()
	upgradeFaultCacheMgr.removedEvent = make(UpgradeFaultReasonMap[LogicId])
	return res
}

func CopyUpgradeFaultCache() UpgradeFaultReasonMap[LogicId] {
	upgradeFaultCacheMgr.cacheLock.Lock()
	defer upgradeFaultCacheMgr.cacheLock.Unlock()
	return upgradeFaultCacheMgr.cache.copy()
}

type UpgradeTypeEnum string

const (
	DurationUpgradeType  UpgradeTypeEnum = "FaultDuration"
	FrequencyUpgradeType UpgradeTypeEnum = "FaultFrequency"
	AutofillUpgradeType  UpgradeTypeEnum = "FaultAutofill"
	validSplitNum                        = 2
	invalidPhyId                         = PhyId(-1)
)

// UpgradeFaultReason indicate the reason of card which is upgrade
type UpgradeFaultReason struct {
	UpgradeTime           int64 `json:"upgrade_time"`
	UpgradeFaultReasonKey `json:",inline"`
}

// UpgradeFaultReasonKey indicate the reason key of card which is upgrade
type UpgradeFaultReasonKey struct {
	FaultCode   string          `json:"fault_code"`
	FaultLevel  string          `json:"fault_level"`
	UpgradeType UpgradeTypeEnum `json:"upgrade_type"`
}

// LogicId used in cache
type LogicId int32

// PhyId used in configmap
type PhyId int32

// ReasonKey the reason key of upgrade fault includes phy id or logic id
type ReasonKey interface {
	LogicId | PhyId
}

type UpgradeFaultReasonSet map[UpgradeFaultReasonKey]UpgradeFaultReason

func (reasonSet UpgradeFaultReasonSet) equals(otherReasonSet UpgradeFaultReasonSet) bool {
	if len(reasonSet) != len(otherReasonSet) {
		return false
	}
	for key, thisVal := range reasonSet {
		thatVal, found := otherReasonSet[key]
		if !found || thisVal != thatVal {
			return false
		}
	}
	return true
}

func (reasonSet UpgradeFaultReasonSet) batchAdd(otherReasonSet UpgradeFaultReasonSet) {
	for reasonKey, reasonVal := range otherReasonSet {
		reasonSet[reasonKey] = reasonVal
	}
}

func (reasonSet UpgradeFaultReasonSet) toList() []UpgradeFaultReason {
	lis := make([]UpgradeFaultReason, 0)
	for _, reasonVal := range reasonSet {
		lis = append(lis, reasonVal)
	}
	return lis
}

func (reasonSet UpgradeFaultReasonSet) toString() string {
	return ObjToString(reasonSet.toList())
}

func ReasonListToSet(reasonList []UpgradeFaultReason) UpgradeFaultReasonSet {
	res := make(UpgradeFaultReasonSet)
	for _, reason := range reasonList {
		key := UpgradeFaultReasonKey{
			FaultCode:   reason.FaultCode,
			FaultLevel:  reason.FaultLevel,
			UpgradeType: reason.UpgradeType,
		}
		oldReason, found := res[key]
		if !found || oldReason.UpgradeTime < reason.UpgradeTime {
			res[key] = reason
		}
	}
	return res
}

func (reasonSet UpgradeFaultReasonSet) checkLevel(faultLevel string) bool {
	for reason := range reasonSet {
		if reason.FaultLevel == faultLevel {
			return true
		}
	}
	return false
}

func (reasonSet UpgradeFaultReasonSet) removeLevel(faultLevel string) UpgradeFaultReasonSet {
	removedReason := make(UpgradeFaultReasonSet)
	for reasonKey, reasonVal := range reasonSet {
		if reasonKey.FaultLevel == faultLevel {
			delete(reasonSet, reasonKey)
			removedReason[reasonKey] = reasonVal
		}
	}
	return removedReason
}

func (reasonSet UpgradeFaultReasonSet) removeFaultCode(faultCode string) UpgradeFaultReasonSet {
	removedReason := make(UpgradeFaultReasonSet)
	for reasonKey, reasonVal := range reasonSet {
		if reasonKey.FaultCode == faultCode {
			delete(reasonSet, reasonKey)
			removedReason[reasonKey] = reasonVal
		}
	}
	return removedReason
}

func (reasonSet UpgradeFaultReasonSet) copy() UpgradeFaultReasonSet {
	res := make(UpgradeFaultReasonSet)
	for reasonKey, reasonVal := range reasonSet {
		res[reasonKey] = reasonVal
	}
	return res
}

type UpgradeFaultReasonMap[T ReasonKey] map[T]UpgradeFaultReasonSet

func (reasonMap UpgradeFaultReasonMap[ReasonKey]) Equals(otherReasonMap UpgradeFaultReasonMap[ReasonKey]) bool {
	if len(reasonMap) != len(otherReasonMap) {
		return false
	}
	for id, thisReasons := range reasonMap {
		otherReasons, found := otherReasonMap[id]
		if !found || !thisReasons.equals(otherReasons) {
			return false
		}
	}
	return true
}

func (reasonMap UpgradeFaultReasonMap[LogicId]) addReasons(logicId LogicId, otherReasons UpgradeFaultReasonSet) {
	reasons, found := reasonMap[logicId]
	if !found {
		reasons = make(UpgradeFaultReasonSet)
	}
	reasons.batchAdd(otherReasons)
	reasonMap[logicId] = reasons
}

func (reasonMap UpgradeFaultReasonMap[ReasonKey]) removeFaultLevel(
	id ReasonKey, faultLevel string) UpgradeFaultReasonSet {
	reasons, found := reasonMap[id]
	removedReasons := make(UpgradeFaultReasonSet)
	if !found {
		return removedReasons
	}
	removedReasons = reasons.removeLevel(faultLevel)
	if len(reasons) == 0 {
		delete(reasonMap, id)
	}
	return removedReasons
}

func (reasonMap UpgradeFaultReasonMap[LogicId]) removeFaultCode(
	logicId LogicId, faultCode string) UpgradeFaultReasonSet {
	reasons, found := reasonMap[logicId]
	if !found {
		return make(UpgradeFaultReasonSet)
	}
	removedReasons := reasons.removeFaultCode(faultCode)
	if len(reasons) == 0 {
		delete(reasonMap, logicId)
	}
	return removedReasons
}

func (reasonMap UpgradeFaultReasonMap[ReasonKey]) GetKeys() []ReasonKey {
	ReasonKeys := make([]ReasonKey, 0, len(reasonMap))
	for deviceKey := range reasonMap {
		ReasonKeys = append(ReasonKeys, deviceKey)
	}
	return ReasonKeys
}

func (reasonMap UpgradeFaultReasonMap[ReasonKey]) copy() UpgradeFaultReasonMap[ReasonKey] {
	ret := make(UpgradeFaultReasonMap[ReasonKey])
	for id, reason := range reasonMap {
		ret[id] = reason.copy()
	}
	return ret
}

// ConvertCacheToCm reasonCache convert to reasonCm
func (reasonMap UpgradeFaultReasonMap[LogicId]) ConvertCacheToCm(
	logicToPhyConvertFunc func(int32) (int32, error)) (UpgradeFaultReasonMap[PhyId], error) {
	reasonCm := make(UpgradeFaultReasonMap[PhyId])

	for logicId, reasons := range reasonMap {
		phyId, err := logicToPhyConvertFunc(int32(logicId))
		if err != nil {
			return nil, fmt.Errorf("convert logicId %v to phyId error: %v", logicId, err)
		}
		reasonCm[PhyId(phyId)] = reasons.copy()
	}
	return reasonCm, nil
}

// ConvertCmToCache reasonCache convert to reasonCm
func (reasonMap UpgradeFaultReasonMap[PhyId]) ConvertCmToCache(
	phyToLogicConvertFunc func(int32) (int32, error)) (UpgradeFaultReasonMap[LogicId], error) {
	reasonCache := make(UpgradeFaultReasonMap[LogicId])

	for phyId, reasons := range reasonMap {
		logicId, err := phyToLogicConvertFunc(int32(phyId))
		if err != nil {
			return nil, fmt.Errorf("convert phyId %v to logicId error: %v", phyId, err)
		}
		reasonCache[LogicId(logicId)] = reasons.copy()
	}
	return reasonCache, nil
}

// CmToString convert ReasonCm to configmap string
func (reasonMap UpgradeFaultReasonMap[PhyId]) CmToString(deviceTypePrefix string) string {
	cm := make(map[string][]UpgradeFaultReason)
	phyIdToDeviceName := func(phyId PhyId) string {
		return deviceTypePrefix + "-" + strconv.Itoa(int(phyId))
	}
	for phyId, reasonSet := range reasonMap {
		cm[phyIdToDeviceName(phyId)] = reasonSet.toList()
	}
	return ObjToString(cm)
}

func deviceNameToPhyId(deviceName string) (PhyId, error) {
	split := strings.Split(deviceName, "-")
	if len(split) != validSplitNum {
		return -1, fmt.Errorf("get phyid from %s failed", deviceName)
	}
	phyId, atoiErr := strconv.Atoi(split[1])
	if atoiErr != nil {
		return invalidPhyId, fmt.Errorf("get phyid from splited %s failed", split[1])
	}
	return PhyId(phyId), nil
}

// StringToReasonCm convert string configmap to reasonCm
func StringToReasonCm(cm string) (UpgradeFaultReasonMap[PhyId], error) {
	cmData := make(map[string][]UpgradeFaultReason)

	err := json.Unmarshal([]byte(cm), &cmData)
	if err != nil {
		return nil, fmt.Errorf("StrToReasonCm unmarshal %s to cmData error: %v", cm, err)
	}
	reasonCm := make(UpgradeFaultReasonMap[PhyId])
	for deviceName, reasons := range cmData {
		phyId, err := deviceNameToPhyId(deviceName)
		if err != nil {
			return nil, fmt.Errorf("StrToReasonCm deviceNameToPhyId error: %v", err)
		}
		reasonCm[phyId] = ReasonListToSet(reasons)
	}
	return reasonCm, nil
}

// FixManuallySeparateReason fix the manually separate NPU reason according to the ManuallySeparateNPU value
// When configmap ManuallySeparateNPU changed
// 1. if Npu is in ManuallySeparateNPU not in UpgradeFaultReason, then UpgradeFaultReason should fill the reason
// 2. if Npu is not in ManuallySeparateNPU but in UpgradeFaultReason, then UpgradeFaultReason should remove the reason
func (reasonMap UpgradeFaultReasonMap[PhyId]) FixManuallySeparateReason(manuallySeparateNPU []PhyId) {
	shouldManuallySeparateList := make(map[PhyId]struct{})
	// 1. insert missing ManuallySeparateNPU
	for _, phyId := range manuallySeparateNPU {
		shouldManuallySeparateList[phyId] = struct{}{}
		reasonSet, found := reasonMap[phyId]
		if !found {
			reasonMap.UpdateReason(phyId, time.Now().UnixMilli(), "", ManuallySeparateNPU, AutofillUpgradeType)
			continue
		}
		exist := reasonSet.checkLevel(ManuallySeparateNPU)
		if !exist {
			reasonMap.UpdateReason(phyId, time.Now().UnixMilli(), "", ManuallySeparateNPU, AutofillUpgradeType)
		}
	}
	// 2. remove redundant ManuallySeparateNPU
	for phyId := range reasonMap {
		if _, found := shouldManuallySeparateList[phyId]; !found {
			reasonMap.removeFaultLevel(phyId, ManuallySeparateNPU)
		}
	}
}

// UpdateReason update reason cache
func (reasonMap UpgradeFaultReasonMap[ReasonKey]) UpdateReason(
	key ReasonKey, faultTime int64, faultCode, faultLevel string, upgradeType UpgradeTypeEnum) bool {
	reasonSet, found := reasonMap[key]
	if !found {
		reasonSet = make(UpgradeFaultReasonSet)
	}

	reasonKey := UpgradeFaultReasonKey{
		FaultCode:   faultCode,
		FaultLevel:  faultLevel,
		UpgradeType: upgradeType,
	}
	reasonVal := UpgradeFaultReason{
		UpgradeTime:           faultTime,
		UpgradeFaultReasonKey: reasonKey,
	}
	oldReasonVal, found := reasonSet[reasonKey]
	updated := false
	if !found || oldReasonVal != reasonVal {
		reasonSet[reasonKey] = reasonVal
		updated = true
	}

	reasonMap[key] = reasonSet
	return updated
}
