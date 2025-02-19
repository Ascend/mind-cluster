// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package publicfault cache utils for public fault
package publicfault

import (
	"context"
	"errors"
	"sync"
	"time"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
)

var (
	// PubFaultCache public fault cache
	PubFaultCache *pubFaultCache
	// PubFaultNeedDelete public fault queue need delete
	PubFaultNeedDelete *needDeleteQueue
)

func init() {
	PubFaultCache = &pubFaultCache{
		faultCache: make(map[string]map[string]*constant.PubFaultCache),
		mutex:      sync.Mutex{},
	}
	PubFaultNeedDelete = &needDeleteQueue{
		faults: make([]needDeleteFault, 0),
		mutex:  sync.Mutex{},
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

type needDeleteQueue struct {
	faults []needDeleteFault
	mutex  sync.Mutex
}

type needDeleteFault struct {
	deleteTime int64
	nodeName   string
	faultKey   string
}

// Push to push new item to needDeleteQueue
func (q *needDeleteQueue) Push(deleteTime int64, nodeName, faultKey string) {
	newItem := needDeleteFault{
		deleteTime: deleteTime,
		nodeName:   nodeName,
		faultKey:   faultKey,
	}
	q.mutex.Lock()
	defer q.mutex.Unlock()
	q.faults = append(q.faults, newItem)
}

// Pop to pop item from needDeleteQueue
func (q *needDeleteQueue) Pop() needDeleteFault {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if len(q.faults) == 0 {
		return needDeleteFault{}
	}
	removed := q.faults[0]
	// remove the head item
	q.faults = q.faults[1:]
	return removed
}

// Len length of needDeleteQueue
func (q *needDeleteQueue) Len() int {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	return len(q.faults)
}

// DealDelete deal public fault need delete
func (q *needDeleteQueue) DealDelete(ctx context.Context) {
	const duration = 500 * time.Millisecond
	for {
		select {
		case <-ctx.Done():
			return
		default:
			if q.Len() == 0 {
				time.Sleep(duration)
				continue
			}
			needDeal := q.Pop()
			deleteTime := needDeal.deleteTime
			if deleteTime <= time.Now().Unix() {
				PubFaultCache.DeleteOccurFault(needDeal.nodeName, needDeal.faultKey)
				continue
			}
			time.Sleep(time.Duration(time.Now().Unix() - deleteTime))
			PubFaultCache.DeleteOccurFault(needDeal.nodeName, needDeal.faultKey)
		}
	}
}
