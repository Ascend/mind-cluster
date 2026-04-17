/* Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package storage for taskd manager backend data type
package storage

import (
	"fmt"
	"sync"
	"time"

	"taskd/common/constant"
	"taskd/toolkit_backend/net/common"
)

// DataPool is used to store message info
type DataPool struct {
	Snapshot *SnapShot
	RWMutex  sync.RWMutex
}

// SnapShot is used to store agent worker cluster infos
type SnapShot struct {
	AgentInfos   *AgentInfos
	WorkerInfos  *WorkerInfos
	ClusterInfos *ClusterInfos
	MgrInfos     *MgrInfo
	// WorkerNum is the number of worker
	WorkerNum int
}

// MsgQueue the queue store message
type MsgQueue struct {
	Queue []BaseMessage
	Mutex sync.Mutex
}

// NewMsg new a base message to enqueue
func (mq *MsgQueue) NewMsg(uuid string, bizType string, src *common.Position, msgBody MsgBody) BaseMessage {
	return BaseMessage{
		Header: MsgHeader{
			Uuid:      uuid,
			BizType:   bizType,
			Src:       src,
			Timestamp: time.Now(),
		},
		Body: msgBody,
	}
}

// Enqueue message enter the queue
func (mq *MsgQueue) Enqueue(msg BaseMessage) error {
	mq.Mutex.Lock()
	defer mq.Mutex.Unlock()
	if len(mq.Queue) >= constant.MaxMsgQueueLength {
		return fmt.Errorf("message queue is full")
	}

	mq.Queue = append(mq.Queue, msg)
	return nil
}

// Dequeue message departure
func (mq *MsgQueue) Dequeue() (BaseMessage, error) {
	mq.Mutex.Lock()
	defer mq.Mutex.Unlock()
	if len(mq.Queue) == 0 {
		return BaseMessage{}, fmt.Errorf("message queue is empty")
	}
	msg := mq.Queue[0]
	mq.Queue = mq.Queue[1:]
	return msg, nil
}

// RegisterAgent register agent in the data pool
func (d *DataPool) RegisterAgent(agentName string, agentInfo *AgentInfo) error {
	err := d.Snapshot.AgentInfos.registerAgent(agentName, agentInfo)
	return err
}

// RegisterWorker register worker in the data pool
func (d *DataPool) RegisterWorker(workerName string, workerInfo *WorkerInfo) error {
	err := d.Snapshot.WorkerInfos.registerWorker(workerName, workerInfo)
	return err
}

// RegisterCluster register cluster in the data pool
func (d *DataPool) RegisterCluster(clusterName string) *ClusterInfo {
	return d.Snapshot.ClusterInfos.registerCluster(clusterName)
}

// UpdateAgent update agent info in the data pool
func (d *DataPool) UpdateAgent(agentName string, agentInfo *AgentInfo) error {
	if d == nil || d.Snapshot == nil || d.Snapshot.AgentInfos == nil || d.Snapshot.AgentInfos.Agents == nil {
		return fmt.Errorf("agents is not initialized")
	}
	err := d.Snapshot.AgentInfos.updateAgent(agentName, agentInfo)
	return err
}

// UpdateWorker update worker info in the data pool
func (d *DataPool) UpdateWorker(workerName string, workerInfo *WorkerInfo) error {
	if d == nil || d.Snapshot == nil || d.Snapshot.WorkerInfos == nil || d.Snapshot.WorkerInfos.Workers == nil {
		return fmt.Errorf("workers is not initialized")
	}
	err := d.Snapshot.WorkerInfos.updateWorker(workerName, workerInfo)
	return err
}

// UpdateCluster update cluster info in the data pool
func (d *DataPool) UpdateCluster(clusterName string, clusterInfo *ClusterInfo) error {
	if d == nil || d.Snapshot == nil || d.Snapshot.ClusterInfos == nil || d.Snapshot.ClusterInfos.Clusters == nil {
		return fmt.Errorf("clusters is not initialized")
	}
	err := d.Snapshot.ClusterInfos.updateCluster(clusterName, clusterInfo)
	return err
}

// GetAgent return agent info about agent name
func (d *DataPool) GetAgent(agentName string) (*AgentInfo, error) {
	if d == nil || d.Snapshot == nil || d.Snapshot.AgentInfos == nil {
		return nil, fmt.Errorf("agents is not initialized")
	}
	return d.Snapshot.AgentInfos.getAgent(agentName)
}

// GetWorker return worker info about worker name
func (d *DataPool) GetWorker(workerName string) (*WorkerInfo, error) {
	if d == nil || d.Snapshot == nil || d.Snapshot.WorkerInfos == nil {
		return nil, fmt.Errorf("workers is not initialized")
	}
	return d.Snapshot.WorkerInfos.getWorker(workerName)
}

// GetCluster return cluster info about cluster name
func (d *DataPool) GetCluster(clusterName string) (*ClusterInfo, error) {
	if d == nil || d.Snapshot == nil || d.Snapshot.ClusterInfos == nil {
		return nil, fmt.Errorf("clusters is not initialized")
	}
	return d.Snapshot.ClusterInfos.getCluster(clusterName)
}

// GetPos return worker or agent position
func (d *DataPool) GetPos(infoType, name string) (*common.Position, error) {
	switch infoType {
	case common.AgentRole:
		agent, err := d.GetAgent(name)
		if err != nil || agent == nil || agent.Pos == nil {
			return nil, fmt.Errorf("agent name is unregistered : %v", name)
		}
		return agent.Pos, nil
	case common.WorkerRole:
		worker, err := d.GetWorker(name)
		if err != nil || worker == nil || worker.Pos == nil {
			return nil, fmt.Errorf("worker name is unregistered : %v", name)
		}
		return worker.Pos, nil
	default:
		return nil, fmt.Errorf("invalid info type")
	}
}

// GetSnapShot get data pool snapshot
func (d *DataPool) GetSnapShot() (*SnapShot, error) {
	d.RWMutex.RLock()
	defer d.RWMutex.RUnlock()
	return d.Snapshot.deepCopy()
}

// UpdateMgr update mgr info in the data pool
func (d *DataPool) UpdateMgr(mgrInfo *MgrInfo) error {
	if d == nil || d.Snapshot == nil || d.Snapshot.MgrInfos == nil || d.Snapshot.MgrInfos.Status == nil {
		return fmt.Errorf("mgr is not initialized")
	}
	err := d.Snapshot.MgrInfos.updateMgr(mgrInfo)
	return err
}

// GetMgr return mgr info about mgr name
func (d *DataPool) GetMgr() (*MgrInfo, error) {
	if d == nil || d.Snapshot == nil || d.Snapshot.MgrInfos == nil || d.Snapshot.MgrInfos.Status == nil {
		return nil, fmt.Errorf("mgr is not initialized")
	}
	return d.Snapshot.MgrInfos, nil
}

func (s *SnapShot) deepCopy() (*SnapShot, error) {
	if s == nil {
		return nil, fmt.Errorf("snapshot is null")
	}
	clone := &SnapShot{}
	if s.AgentInfos != nil {
		clone.AgentInfos = deepCopyAgent(s.AgentInfos)
	}
	if s.WorkerInfos != nil {
		clone.WorkerInfos = deepCopyWorker(s.WorkerInfos)
	}
	if s.ClusterInfos != nil {
		clone.ClusterInfos = deepCopyCluster(s.ClusterInfos)
	}
	if s.MgrInfos != nil {
		clone.MgrInfos = deepCopyMgr(s.MgrInfos)
	}
	clone.WorkerNum = s.WorkerNum
	return clone, nil
}

func deepCopyAgent(agentInfos *AgentInfos) *AgentInfos {
	return agentInfos.DeepCopy()
}

func deepCopyWorker(workerInfos *WorkerInfos) *WorkerInfos {
	return workerInfos.DeepCopy()
}

func deepCopyCluster(clusterInfos *ClusterInfos) *ClusterInfos {
	return clusterInfos.DeepCopy()
}

func deepCopyMgr(mgrInfos *MgrInfo) *MgrInfo {
	return mgrInfos.DeepCopy()
}
