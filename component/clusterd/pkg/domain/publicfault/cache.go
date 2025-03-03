// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package publicfault cache utils for public fault
package publicfault

import (
	"errors"
	"sync"
	"time"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
)

// PubFaultCache public fault cache
var PubFaultCache *pubFaultCache

func init() {
	PubFaultCache = &pubFaultCache{
		faultCache: make(map[string]map[string]*constant.PubFaultCache),
		mutex:      sync.Mutex{},
	}
}

type pubFaultCache struct {
	// key: node name; value: {faultResource+faultId:fault}
	faultCache map[string]map[string]*constant.PubFaultCache
	mutex      sync.Mutex
}

// AddPubFaultToCache add new public fault to cache
func (pc *pubFaultCache) AddPubFaultToCache(newFault *constant.PubFaultCache, nodeName, faultKey string) {
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

// DeleteOccurFault delete occur from cache
func (pc *pubFaultCache) DeleteOccurFault(nodeName, faultKey string) {
	pc.mutex.Lock()
	defer pc.mutex.Unlock()
	delete(pc.faultCache[nodeName], faultKey)
}

// GetPubFault get public fault from cache
func (pc *pubFaultCache) GetPubFault() map[string]map[string]*constant.PubFaultCache {
	pc.mutex.Lock()
	defer pc.mutex.Unlock()
	return pc.faultCache
}

// GetPubFaultByNodeName get public fault from cache by node name
func (pc *pubFaultCache) GetPubFaultByNodeName(nodeName string) (map[string]*constant.PubFaultCache, bool) {
	pc.mutex.Lock()
	defer pc.mutex.Unlock()
	nodeFault, nodeExisted := pc.faultCache[nodeName]
	return nodeFault, nodeExisted
}

// DeepCopy deep copy fault cache
func (pc *pubFaultCache) DeepCopy() (map[string]map[string]*constant.PubFaultCache, error) {
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
func (pc *pubFaultCache) FaultExisted(nodeName, faultKey string) (bool, int64) {
	pc.mutex.Lock()
	nodeFault, nodeExist := pc.faultCache[nodeName]
	fault, faultExist := nodeFault[faultKey]
	pc.mutex.Unlock()
	if !nodeExist || !faultExist || fault.Assertion != constant.AssertionOccur {
		return false, 0
	}
	return true, fault.FaultAddTime
}
