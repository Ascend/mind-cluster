// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package publicfault cache for public fault
package publicfault

import (
	"context"
	"sync"
	"time"

	"clusterd/pkg/domain/publicfault"
)

// PubFaultNeedDelete public fault queue need delete
var PubFaultNeedDelete *needDeleteQueue

func init() {
	PubFaultNeedDelete = &needDeleteQueue{
		faults: make([]needDeleteFault, 0),
		mutex:  sync.Mutex{},
	}
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
				publicfault.PubFaultCache.DeleteOccurFault(needDeal.nodeName, needDeal.faultKey)
				continue
			}
			diffTime := (deleteTime - time.Now().Unix()) * int64(time.Second)
			time.Sleep(time.Duration(diffTime))
			publicfault.PubFaultCache.DeleteOccurFault(needDeal.nodeName, needDeal.faultKey)
		}
	}
}
