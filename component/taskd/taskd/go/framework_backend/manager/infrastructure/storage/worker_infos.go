/* Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain w copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

// Package storage for taskd manager backend data type
package storage

import (
	"fmt"
	"sync"
	"time"

	"ascend-common/common-utils/hwlog"
	"taskd/common/utils"
	"taskd/toolkit_backend/net/common"
)

// WorkerInfos all worker infos
type WorkerInfos struct {
	Workers   map[string]*WorkerInfo
	AllStatus map[string]string
	RWMutex   sync.RWMutex
}

// WorkerInfo the worker info
type WorkerInfo struct {
	Config     map[string]string
	Actions    map[string]string
	Status     map[string]string
	GlobalRank string
	HeartBeat  time.Time
	FaultInfo  map[string]string
	Pos        *common.Position
	RWMutex    sync.RWMutex
}

func (w *WorkerInfos) registerWorker(workerName string, workerInfo *WorkerInfo) error {
	w.RWMutex.Lock()
	w.Workers[workerName] = workerInfo
	w.RWMutex.Unlock()
	hwlog.RunLog.Infof("register worker name:%v agentInfo:%v", workerName, utils.ObjToString(workerInfo))
	return nil
}

func (w *WorkerInfos) getWorker(workerName string) (*WorkerInfo, error) {
	w.RWMutex.RLock()
	defer w.RWMutex.RUnlock()
	if worker, exists := w.Workers[workerName]; exists {
		return worker, nil
	}
	return nil, fmt.Errorf("worker name is unregistered : %v", workerName)
}

func (w *WorkerInfos) updateWorker(workerName string, newWorker *WorkerInfo) error {
	w.RWMutex.Lock()
	defer w.RWMutex.Unlock()
	w.Workers[workerName] = newWorker
	return nil
}

// DeepCopy return a deep copy of WorkerInfos
func (w *WorkerInfos) DeepCopy() *WorkerInfos {
	w.RWMutex.RLock()
	defer w.RWMutex.RUnlock()
	clone := &WorkerInfos{
		Workers:   make(map[string]*WorkerInfo, len(w.Workers)),
		AllStatus: make(map[string]string, len(w.AllStatus)),
	}
	for k, v := range w.AllStatus {
		clone.AllStatus[k] = v
	}
	for k, v := range w.Workers {
		if v == nil {
			clone.Workers[k] = nil
			continue
		}
		clone.Workers[k] = v.DeepCopy()
	}
	return clone
}

// SetStatusVal set worker status value
func (w *WorkerInfo) SetStatusVal(statusType string, statusVal string) {
	w.RWMutex.Lock()
	defer w.RWMutex.Unlock()
	w.Status[statusType] = statusVal
}

// DeepCopy return a deep copy of WorkerInfo
func (w *WorkerInfo) DeepCopy() *WorkerInfo {
	w.RWMutex.RLock()
	defer w.RWMutex.RUnlock()
	clone := &WorkerInfo{
		Config:     utils.CopyStringMap(w.Config),
		Actions:    utils.CopyStringMap(w.Actions),
		FaultInfo:  utils.CopyStringMap(w.FaultInfo),
		Status:     utils.CopyStringMap(w.Status),
		GlobalRank: w.GlobalRank,
		HeartBeat:  w.HeartBeat,
		RWMutex:    sync.RWMutex{},
	}
	if w.Pos != nil {
		clone.Pos = &common.Position{
			Role:        w.Pos.Role,
			ServerRank:  w.Pos.ServerRank,
			ProcessRank: w.Pos.ProcessRank,
		}
	}
	return clone
}
