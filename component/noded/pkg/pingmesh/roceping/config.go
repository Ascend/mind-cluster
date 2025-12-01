/*
Copyright(C)2025. Huawei Technologies Co.,Ltd. All rights reserved.

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

// Package roceping for ping by icmp in RoCE mesh net between super pods in A5
package roceping

import (
	"errors"
	"fmt"
	"net"
	"sync"
	"time"
)

// IcmpBodyData for icmp body data struct
type IcmpBodyData struct {
	Uid      string
	SeqId    int
	SendTime string
}

// IcmpPacketInfo for the received icmp packet info
type IcmpPacketInfo struct {
	Data        []byte
	Peer        net.Addr
	ReceiveTime time.Time
}

// IcmpPingMeshStatistics refers to the statistic result of ping task
type IcmpPingMeshStatistics struct {
	SrcAddr    string
	DstAddr    string
	SucPktNum  int64
	FailPktNum int64
	MaxTime    int64
	MinTime    int64
	AvgTime    float64
	Rtt        []int64
	ActionTime []int64
	// TP95Time 处于95%位置的响应时间
	TP95Time     int64
	ReplyStatNum int64
	PingTotalNum int64
}

// NewIcmpPingMeshStatistics for create IcmpPingMeshStatistics instance
func NewIcmpPingMeshStatistics(srcAddr, dstAddr string) *IcmpPingMeshStatistics {
	return &IcmpPingMeshStatistics{
		SrcAddr:      srcAddr,
		DstAddr:      dstAddr,
		SucPktNum:    0,
		FailPktNum:   0,
		MaxTime:      -1,
		MinTime:      -1,
		AvgTime:      -1,
		Rtt:          make([]int64, 0),
		ActionTime:   make([]int64, 0),
		TP95Time:     -1,
		ReplyStatNum: 0,
		PingTotalNum: 0,
	}
}

// IcmpPingMeshOperate refers to the operation of icmp ping parameters
type IcmpPingMeshOperate struct {
	SrcAddr      string
	DstAddr      string
	PktSendNum   int
	PktInterval  int // ms
	Timeout      int // s
	TaskInterval int // s
	TaskId       int
	DstIpAddr    *net.IPAddr
}

// NewOperator for icmp ping action parameters
func NewOperator(dstAddr string, srcAddr string, taskInterval int) *IcmpPingMeshOperate {
	return &IcmpPingMeshOperate{
		SrcAddr:      srcAddr,
		DstAddr:      dstAddr,
		PktSendNum:   1,
		PktInterval:  10,           // millisecond
		Timeout:      1,            // second
		TaskInterval: taskInterval, // second
	}
}

type resultInfo struct {
	SourceAddr   string `json:"source_addr"`
	TargetAddr   string `json:"target_addr"`
	SucPktNum    uint   `json:"suc_pkt_num"`
	FailPktNum   uint   `json:"fail_pkt_num"`
	MaxTime      int    `json:"max_time"`
	MinTime      int    `json:"min_time"`
	AvgTime      int    `json:"avg_time"`
	TP95Time     int    `json:"tp95_time"`
	ReplyStatNum int    `json:"reply_stat_num"`
	PingTotalNum int    `json:"ping_total_num"`
}

type statisticData struct {
	result string
	record []string
}

// SeqIdSet for sequence id set
type SeqIdSet struct {
	sync.RWMutex
	m map[int]int64
}

// NewSeqIdSet for create the SeqIdSet instance
func NewSeqIdSet() *SeqIdSet {
	return &SeqIdSet{
		m: make(map[int]int64),
	}
}

// Add for adding k into set
func (s *SeqIdSet) Add(k int, v int64) {
	s.Lock()
	defer s.Unlock()
	s.m[k] = v
}

// Delete for deleting k from set
func (s *SeqIdSet) Delete(k int) {
	s.Lock()
	defer s.Unlock()
	delete(s.m, k)
}

// Contains for checking is k exist
func (s *SeqIdSet) Contains(k int) bool {
	s.RLock()
	defer s.RUnlock()
	_, ok := s.m[k]
	return ok
}

// Len for count elements
func (s *SeqIdSet) Len() int64 {
	s.RLock()
	defer s.RUnlock()
	return int64(len(s.m))
}

// DeleteLessValues for delete elements less than target, return deleted count and items
func (s *SeqIdSet) DeleteLessValues(target int64) (int64, []int) {
	s.RLock()
	if len(s.m) == 0 {
		s.RUnlock()
		return 0, nil
	}

	toBeDelete := make([]int, 0)
	for k, v := range s.m {
		if v < target {
			toBeDelete = append(toBeDelete, k)
		}
	}
	s.RUnlock()

	n := int64(len(toBeDelete))
	if n == 0 {
		return 0, nil
	}

	s.Lock()
	defer s.Unlock()
	for _, k := range toBeDelete {
		delete(s.m, k)
	}
	return n, toBeDelete
}

// WaitingPktCache for cache the pkt sent before received
type WaitingPktCache struct {
	cache map[string]*SeqIdSet
	lock  *sync.RWMutex
}

// NewWaitingPktCache for create the WaitingPktCache instance
func NewWaitingPktCache() *WaitingPktCache {
	return &WaitingPktCache{
		cache: make(map[string]*SeqIdSet),
		lock:  &sync.RWMutex{},
	}
}

// Len for counting elements
func (w *WaitingPktCache) Len() int64 {
	w.lock.RLock()
	defer w.lock.RUnlock()
	totalCnt := int64(0)
	for _, seqSet := range w.cache {
		totalCnt += seqSet.Len()
	}
	return totalCnt
}

func (w *WaitingPktCache) calcMayLostPktCnt() (int64, map[string][]int) {
	w.lock.RLock()
	defer w.lock.RUnlock()
	curTime := time.Now().UnixMilli()
	target := curTime - waitTimeoutMilliSec
	mayLostCnt := int64(0)
	evictedItems := make(map[string][]int)
	for uid, seqSet := range w.cache {
		timedOutCnt, timedOutSeqId := seqSet.DeleteLessValues(target)
		if timedOutCnt == 0 {
			continue
		}
		mayLostCnt += timedOutCnt
		evictedItems[uid] = append(evictedItems[uid], timedOutSeqId...)
	}
	return mayLostCnt, evictedItems
}

func (w *WaitingPktCache) addPktSeqIdToWaitingSet(uid string, seqId int, timestamp int64) {
	w.lock.Lock()
	defer w.lock.Unlock()
	if _, exist := w.cache[uid]; !exist {
		w.cache[uid] = NewSeqIdSet()
	}
	w.cache[uid].Add(seqId, timestamp)
}

func (w *WaitingPktCache) isPacketValid(dataInfo *IcmpBodyData) error {
	if dataInfo == nil {
		return errors.New("data is empty")
	}

	w.lock.RLock()
	waitingSeqIdSet, exist := w.cache[dataInfo.Uid]
	if !exist {
		w.lock.RUnlock()
		return fmt.Errorf("uid of the echo reply is not registered, uid: %v", dataInfo.Uid)
	}
	w.lock.RUnlock()
	if !waitingSeqIdSet.Contains(dataInfo.SeqId) {
		return fmt.Errorf("seqId of the echo reply is not in waiting set, uid: %v, seqId: %v",
			dataInfo.Uid, dataInfo.SeqId)
	}
	return nil
}

func (w *WaitingPktCache) deletePacketInWaitingSet(dataInfo *IcmpBodyData) error {
	if dataInfo == nil {
		return errors.New("data is empty")
	}

	w.lock.Lock()
	defer w.lock.Unlock()
	waitingSeqIdSet, exist := w.cache[dataInfo.Uid]
	if !exist {
		return nil
	}
	if !waitingSeqIdSet.Contains(dataInfo.SeqId) {
		return nil
	}

	waitingSeqIdSet.Delete(dataInfo.SeqId)
	if waitingSeqIdSet.Len() == 0 {
		delete(w.cache, dataInfo.Uid)
	}
	return nil
}
