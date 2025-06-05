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
	if worker, exists := w.Workers[workerName]; exists {
		return worker.getWorker()
	}
	return nil, fmt.Errorf("worker name is unregistered : %v", workerName)
}

func (w *WorkerInfo) getWorker() (*WorkerInfo, error) {
	w.RWMutex.RLock()
	defer w.RWMutex.RUnlock()
	return &WorkerInfo{
		Config:     w.Config,
		Actions:    w.Actions,
		Status:     w.Status,
		GlobalRank: w.GlobalRank,
		HeartBeat:  w.HeartBeat,
		FaultInfo:  w.FaultInfo,
		Pos:        w.Pos,
		RWMutex:    sync.RWMutex{},
	}, nil
}

func (w *WorkerInfos) updateWorker(workerName string, newWorker *WorkerInfo) error {
	w.Workers[workerName].RWMutex.Lock()
	defer w.Workers[workerName].RWMutex.Unlock()
	w.Workers[workerName] = &WorkerInfo{
		Config:     newWorker.Config,
		Actions:    newWorker.Actions,
		Status:     newWorker.Status,
		GlobalRank: newWorker.GlobalRank,
		HeartBeat:  newWorker.HeartBeat,
		FaultInfo:  newWorker.FaultInfo,
		Pos:        newWorker.Pos,
	}
	return nil
}
