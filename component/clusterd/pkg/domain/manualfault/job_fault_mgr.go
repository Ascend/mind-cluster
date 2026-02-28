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

// Package manualfault cache for hardware frequency fault with job
package manualfault

import (
	"sync"
	"time"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/podgroup"
)

// JobFaultMgr an instance of JobFaultManager
var JobFaultMgr *JobFaultManager

// JobFaultManager is the job fault manager
type JobFaultManager struct {
	jobFault      map[string]*faultInfo
	slidingWindow int64 // unit: millisecond
	mutex         sync.RWMutex
}

type faultInfo struct {
	faults []*Fault
}

// Fault is the hardware frequency fault detail
type Fault struct {
	Code        string
	JobId       string
	NodeName    string
	DevName     string
	ReceiveTime int64 // unit: millisecond
}

func init() {
	InitJobFaultManager(constant.DefaultSlidingWindow)
}

// InitJobFaultManager init the job fault manager
func InitJobFaultManager(windowSize int64) {
	JobFaultMgr = &JobFaultManager{
		jobFault:      make(map[string]*faultInfo),
		slidingWindow: windowSize * constant.SecondsToMilliseconds,
		mutex:         sync.RWMutex{},
	}
}

func (m *JobFaultManager) GetFaultsByJobId(jobId string) []*Fault {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	faults, ok := m.jobFault[jobId]
	if !ok || faults == nil {
		return nil
	}
	return faults.faults
}

// AddFault add fault to cache, and start to deal fault
func (m *JobFaultManager) AddFault(newFault *Fault) {
	if newFault == nil {
		return
	}
	if newFault.JobId == "" {
		Counter.AddFault(newFault)
		return
	}

	hwlog.RunLog.Infof("fault enters the process of determining software fault: %+v", newFault)
	m.mutex.Lock()
	fault, ok := m.jobFault[newFault.JobId]
	if !ok || fault == nil {
		newJobFault := &faultInfo{
			faults: []*Fault{newFault},
		}
		m.jobFault[newFault.JobId] = newJobFault
		m.mutex.Unlock()
		go m.dealJobFault(newFault.JobId)
		return
	}

	fault.faults = append(fault.faults, newFault)
	m.mutex.Unlock()
}

func (m *JobFaultManager) dealJobFault(jobId string) {
	for {
		pg := podgroup.GetPodGroup(jobId)
		if pg.Name == "" {
			m.safeDeleteByJobId(jobId)
			return
		}
		m.mutex.Lock()
		jobFault, ok := m.jobFault[jobId]
		if !ok || jobFault == nil {
			delete(m.jobFault, jobId)
			m.mutex.Unlock()
			return
		}

		faults := jobFault.faults
		if len(faults) == 0 {
			delete(m.jobFault, jobId)
			m.mutex.Unlock()
			return
		}
		for {
			if len(faults) == 0 {
				break
			}
			dealTime := faults[0].ReceiveTime + m.slidingWindow
			if dealTime > time.Now().UnixMilli() {
				break
			}
			isSftFault := firstItemIsSfwFault(faults)
			if !isSftFault {
				Counter.AddFault(faults[0])
			}
			// after deal first fault, delete from cache
			faults = m.deleteSameWithFirstFault(faults, isSftFault)
			jobFault.faults = faults
		}
		m.mutex.Unlock()
		time.Sleep(1 * time.Second)
	}
}

func (m *JobFaultManager) safeDeleteByJobId(jobId string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	delete(m.jobFault, jobId)
}

// deleteSameWithFirstFault delete the first fault and the faults which are same with the first fault
func (m *JobFaultManager) deleteSameWithFirstFault(faults []*Fault, isSftFault bool) []*Fault {
	fault0 := faults[0]
	var faultsAfterDelete []*Fault
	for idx, fault := range faults {
		if idx == 0 {
			continue
		}
		if fault.Code == fault0.Code && fault.ReceiveTime <= fault0.ReceiveTime+m.slidingWindow {
			// within 30 seconds, dev 1 occur faults 1 for 2 times, and dev 2 occur fault 1 for 1 times,
			// which is a software fault
			if isSftFault {
				continue
			}
			// within 30 seconds, dev 1 occur faults 1 for 2 times, and no other devs occur fault 1, which is a hardware fault
			if fault.NodeName == fault0.NodeName && fault.DevName == fault0.DevName {
				hwlog.RunLog.Infof("fault: %+v, is not software fault", fault)
				Counter.AddFault(fault)
				continue
			}
		}
		faultsAfterDelete = append(faultsAfterDelete, fault)
	}
	return faultsAfterDelete
}

func firstItemIsSfwFault(faults []*Fault) bool {
	if len(faults) == 0 {
		return false
	}
	fault0 := faults[0]
	for idx, fault := range faults {
		if idx == 0 {
			continue
		}
		if fault.Code == fault0.Code && (fault.NodeName != fault0.NodeName || fault.DevName != fault0.DevName) {
			hwlog.RunLog.Infof("fault: %+v, is software fault", fault0)
			return true
		}
	}
	hwlog.RunLog.Infof("fault: %+v, is not software fault", fault0)
	return false
}
