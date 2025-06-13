// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package publicfault cache utils for public fault
package publicfault

import (
	"errors"
	"strings"
	"sync"
	"time"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
)

// PubFaultCache public fault cache
var PubFaultCache *cache

func init() {
	PubFaultCache = &cache{
		faultCache: make(map[string]map[string]*constant.PubFaultCache),
		mutex:      sync.Mutex{},
	}
}

// cache public fault cache
type cache struct {
	// faultCache key is node name, value is {faultResource+faultId:fault}
	faultCache map[string]map[string]*constant.PubFaultCache
	mutex      sync.Mutex
}

// AddPubFaultToCache add new public fault to cache. After adding, notify statistic module
func (pc *cache) AddPubFaultToCache(newFault *constant.PubFaultCache, nodeName, faultKey string) {
	const maxPubFaultCacheNum = 50000
	if pc.GetPubFaultNum() >= maxPubFaultCacheNum {
		hwlog.RunLog.Errorf("add public fault to cache failed, "+
			"public fault number in cache exceeds the upper limit %d", maxPubFaultCacheNum)
		return
	}
	newFault.FaultAddTime = time.Now().Unix()
	pc.mutex.Lock()
	defer pc.mutex.Unlock()
	nodeFault, nodeExist := pc.faultCache[nodeName]
	if !nodeExist {
		pc.faultCache[nodeName] = make(map[string]*constant.PubFaultCache)
		pc.faultCache[nodeName][faultKey] = newFault
		return
	}
	nodeFault[faultKey] = newFault
}

// DeleteOccurFault delete occur from cache. After deleting, notify statistic module
func (pc *cache) DeleteOccurFault(nodeName, faultKey string) {
	pc.mutex.Lock()
	defer pc.mutex.Unlock()
	delete(pc.faultCache[nodeName], faultKey)
}

// GetPubFault get public fault from cache
func (pc *cache) GetPubFault() map[string]map[string]*constant.PubFaultCache {
	pc.mutex.Lock()
	defer pc.mutex.Unlock()
	return pc.faultCache
}

// GetPubFaultNum get public fault number in cache
func (pc *cache) GetPubFaultNum() int {
	pc.mutex.Lock()
	defer pc.mutex.Unlock()
	faultNum := 0
	for _, faultsCache := range pc.faultCache {
		faultNumPerNode := len(faultsCache)
		faultNum += faultNumPerNode
	}
	return faultNum
}

// GetPubFaultByNodeName get public fault from cache by node name
func (pc *cache) GetPubFaultByNodeName(nodeName string) (map[string]*constant.PubFaultCache, bool) {
	pc.mutex.Lock()
	defer pc.mutex.Unlock()
	nodeFault, nodeExisted := pc.faultCache[nodeName]
	return nodeFault, nodeExisted
}

// DeepCopy deep copy fault cache
func (pc *cache) DeepCopy() (map[string]map[string]*constant.PubFaultCache, error) {
	pc.mutex.Lock()
	defer pc.mutex.Unlock()
	result := new(map[string]map[string]*constant.PubFaultCache)
	if err := util.DeepCopy(result, pc.faultCache); err != nil {
		hwlog.RunLog.Errorf("deep copy public fault in cache failed, error: %v", err)
		return nil, errors.New("deep copy public fault in cache failed")
	}
	return *result, nil
}

// FaultExisted if occur existed, means fault existed, return fault add time
func (pc *cache) FaultExisted(nodeName, faultKey string) (bool, int64) {
	pc.mutex.Lock()
	nodeFault, nodeExist := pc.faultCache[nodeName]
	fault, faultExist := nodeFault[faultKey]
	pc.mutex.Unlock()
	if !nodeExist || !faultExist || fault.Assertion != constant.AssertionOccur {
		return false, 0
	}
	return true, fault.FaultAddTime
}

// GetPubFaultsForCM get public faults for configmap statistic-fault-info
func (pc *cache) GetPubFaultsForCM() (map[string][]constant.NodeFault, int) {
	pc.mutex.Lock()
	defer pc.mutex.Unlock()
	pubFaults := make(map[string][]constant.NodeFault, len(pc.faultCache))
	pubFaultsNum := 0
	for nodeName, faultsCache := range pc.faultCache {
		if len(faultsCache) == 0 {
			continue
		}
		nodeFaults := make([]constant.NodeFault, 0, len(faultsCache))
		for faultKey, fault := range faultsCache {
			nodeFaults = append(nodeFaults, constant.NodeFault{
				FaultResource: getResourceFromFaultKey(faultKey, fault.FaultId),
				FaultDevIds:   fault.FaultDevIds,
				FaultId:       fault.FaultId,
				FaultType:     fault.FaultType,
				FaultCode:     fault.FaultCode,
				FaultLevel:    fault.FaultLevel,
				FaultTime:     fault.FaultTime,
			})
		}
		pubFaults[nodeName] = nodeFaults
		pubFaultsNum += len(nodeFaults)
	}
	return pubFaults, pubFaultsNum
}

// LoadFaultToCache load public fault to cache
func (pc *cache) LoadFaultToCache(faults map[string][]constant.NodeFault) {
	for nodeName, nodeFaults := range faults {
		for _, nodeFault := range nodeFaults {
			faultKey := nodeFault.FaultResource + nodeFault.FaultId
			faultCache := &constant.PubFaultCache{
				FaultDevIds: nodeFault.FaultDevIds,
				FaultId:     nodeFault.FaultId,
				FaultType:   nodeFault.FaultType,
				FaultCode:   nodeFault.FaultCode,
				FaultLevel:  nodeFault.FaultLevel,
				FaultTime:   nodeFault.FaultTime,
				Assertion:   constant.AssertionOccur,
			}
			pc.AddPubFaultToCache(faultCache, nodeName, faultKey)
		}
	}
}

func getResourceFromFaultKey(faultKey, faultId string) string {
	return strings.Replace(faultKey, faultId, "", -1)
}
