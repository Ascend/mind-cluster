package common

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"ascend-common/common-utils/hwlog"
)

func init() {
	upgradeFaultCacheMgr = UpgradeFaultCacheManager{
		cache: make(UpgradeFaultReasonMap[LogicId]),
		lock:  sync.Mutex{},
	}
}

var upgradeFaultCacheMgr UpgradeFaultCacheManager

type UpgradeFaultCacheManager struct {
	cache UpgradeFaultReasonMap[LogicId]
	lock  sync.Mutex
}

// SaveUpgradeFaultCache use when device-plugin boot, load reason from cm then save in cache
func SaveUpgradeFaultCache(cache UpgradeFaultReasonMap[LogicId]) {
	upgradeFaultCacheMgr.lock.Lock()
	defer upgradeFaultCacheMgr.lock.Unlock()
	upgradeFaultCacheMgr.cache = cache
}

// InsertUpgradeFaultCache update upgrade fault cache
func InsertUpgradeFaultCache(logicId LogicId, faultTime int64, faultCode, faultLevel string) {
	upgradeFaultCacheMgr.lock.Lock()
	defer upgradeFaultCacheMgr.lock.Unlock()
	hwlog.RunLog.Infof("UpdateUpgradeFaultCache logicId %v, faultTime %v, faultCode %v, faultLevel %v",
		logicId, faultTime, faultCode, faultLevel)
	upgradeFaultCacheMgr.cache.UpdateReason(logicId, faultTime, faultCode, faultLevel)
}

// RemoveManuallySeparateReasonCache when cm remove manually separate npu, the cache should remove reported npu
// but the fault that has not been reported shouldn't be removed
func RemoveManuallySeparateReasonCache(logicIds []LogicId) {
	upgradeFaultCacheMgr.lock.Lock()
	defer upgradeFaultCacheMgr.lock.Unlock()
	for _, id := range logicIds {
		upgradeFaultCacheMgr.cache[id].removeLevel(ManuallySeparateNPU)
	}
	hwlog.RunLog.Infof("RemoveManuallySeparateReasonCache logicIds %v", logicIds)
}

func CopyUpgradeFaultCache() UpgradeFaultReasonMap[LogicId] {
	upgradeFaultCacheMgr.lock.Lock()
	defer upgradeFaultCacheMgr.lock.Unlock()
	return upgradeFaultCacheMgr.cache.DeepCopy()
}

// UpgradeFaultReason indicate the reason of card which is upgrade
type UpgradeFaultReason struct {
	UpgradeTime int64
	FaultCode   string
	FaultLevel  string
}

// LogicId used in cache
type LogicId int32

// PhyId used in configmap
type PhyId int32

// ReasonKey the reason key of upgrade fault includes phy id or logic id
type ReasonKey interface {
	LogicId | PhyId
}

type UpgradeFaultReasonSet map[UpgradeFaultReason]struct{}

func (reasonSet UpgradeFaultReasonSet) equals(otherReasonSet UpgradeFaultReasonSet) bool {
	if len(reasonSet) != len(otherReasonSet) {
		return false
	}
	for thisReason := range reasonSet {
		_, found := otherReasonSet[thisReason]
		if !found {
			return false
		}
	}
	return true
}

func (reasonSet UpgradeFaultReasonSet) toList() []UpgradeFaultReason {
	lis := make([]UpgradeFaultReason, 0)
	for reason := range reasonSet {
		lis = append(lis, reason)
	}
	return lis
}

func ReasonListToSet(reasonList []UpgradeFaultReason) UpgradeFaultReasonSet {
	res := make(UpgradeFaultReasonSet)
	for _, reason := range reasonList {
		res[reason] = struct{}{}
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

func (reasonSet UpgradeFaultReasonSet) removeLevel(faultLevel string) {
	for reason := range reasonSet {
		if reason.FaultLevel == faultLevel {
			delete(reasonSet, reason)
		}
	}
}

func (reasonSet UpgradeFaultReasonSet) copy() UpgradeFaultReasonSet {
	res := make(UpgradeFaultReasonSet)
	for reason := range reasonSet {
		res[reason] = struct{}{}
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
		if !found || len(thisReasons) != len(otherReasons) {
			return false
		}
		if !thisReasons.equals(otherReasons) {
			return false
		}
	}
	return true
}

func (reasonMap UpgradeFaultReasonMap[ReasonKey]) GetKeys() []ReasonKey {
	ReasonKeys := make([]ReasonKey, 0)
	for deviceKey, _ := range reasonMap {
		ReasonKeys = append(ReasonKeys, deviceKey)
	}
	return ReasonKeys
}

func (reasonMap UpgradeFaultReasonMap[ReasonKey]) DeepCopy() UpgradeFaultReasonMap[ReasonKey] {
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
	if len(split) != 2 {
		return -1, fmt.Errorf("get phyid from %s failed", deviceName)
	}
	phyId, atoiErr := strconv.Atoi(split[1])
	if atoiErr != nil {
		return -1, fmt.Errorf("get phyid from splited %s failed", split[1])
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

// generateManuallySeparateNPU covert ReasonCm to ManuallySeparateNPU string
func (reasonMap UpgradeFaultReasonMap[PhyId]) generateManuallySeparateNPU(deviceTypePrefix string) string {
	deviceNames := make([]string, 0)
	phyIdToDeviceName := func(phyId PhyId) string {
		return deviceTypePrefix + strconv.Itoa(int(phyId))
	}
	for phyId, reasonSet := range reasonMap {
		if reasonSet.checkLevel(ManuallySeparateNPU) {
			deviceName := phyIdToDeviceName(phyId)
			deviceNames = append(deviceNames, deviceName)
		}
	}
	return strings.Join(deviceNames, ",")
}

// FixManuallySeparateReason fix the manually separate NPU reason according to the ManuallySeparateNPU value
// When configmap ManuallySeparateNPU changed
func (reasonMap UpgradeFaultReasonMap[PhyId]) FixManuallySeparateReason(manuallySeparateList []string) error {
	shouldManuallySeparateList := make(map[PhyId]struct{})
	for _, deviceName := range manuallySeparateList {
		phyId, err := deviceNameToPhyId(deviceName)
		if err != nil {
			return fmt.Errorf("FixManuallySeparateReason deviceNameToPhyId error: %v", err)
		}
		shouldManuallySeparateList[PhyId(phyId)] = struct{}{}
	}
	for phyId, reasonSet := range reasonMap {
		if _, found := shouldManuallySeparateList[phyId]; !found {
			reasonSet.removeLevel(ManuallySeparateNPU)
		}
	}
	return nil
}

// UpdateReason update reason cache
func (reasonMap UpgradeFaultReasonMap[LogicId]) UpdateReason(
	logicId LogicId, faultTime int64, faultCode, faultLevel string) {
	reasonSet, found := reasonMap[logicId]
	if !found {
		reasonSet = make(UpgradeFaultReasonSet)
	}
	reason := UpgradeFaultReason{
		UpgradeTime: faultTime,
		FaultCode:   faultCode,
		FaultLevel:  faultLevel,
	}
	reasonSet[reason] = struct{}{}
	reasonMap[logicId] = reasonSet
}
