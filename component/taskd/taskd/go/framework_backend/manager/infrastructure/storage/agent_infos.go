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

	"ascend-common/common-utils/hwlog"
	"taskd/common/utils"
	"taskd/toolkit_backend/net/common"
)

// AgentInfos all agent infos
type AgentInfos struct {
	Agents    map[string]*AgentInfo
	AllStatus map[string]string
	RWMutex   sync.RWMutex
}

// AgentInfo the agent info
type AgentInfo struct {
	Config    map[string]string
	Actions   map[string]string
	Status    map[string]string
	NodeRank  string
	HeartBeat time.Time
	FaultInfo map[string]string
	Pos       *common.Position
	RWMutex   sync.RWMutex
}

func (a *AgentInfos) registerAgent(agentName string, agentInfo *AgentInfo) error {
	a.RWMutex.Lock()
	a.Agents[agentName] = agentInfo
	a.RWMutex.Unlock()
	hwlog.RunLog.Infof("register agent name:%v agentInfo:%v", agentName, utils.ObjToString(agentInfo))
	return nil
}

func (a *AgentInfos) getAgent(agentName string) (*AgentInfo, error) {
	a.RWMutex.RLock()
	defer a.RWMutex.RUnlock()
	if agent, exists := a.Agents[agentName]; exists {
		return agent, nil
	}
	return nil, fmt.Errorf("agent name is unregistered : %v", agentName)
}

func (a *AgentInfos) updateAgent(agentName string, newAgent *AgentInfo) error {
	a.RWMutex.Lock()
	defer a.RWMutex.Unlock()
	a.Agents[agentName] = newAgent
	return nil
}

// SetAllStatusVal set agent all status value
func (a *AgentInfos) SetAllStatusVal(agentName string, status string) error {
	a.RWMutex.Lock()
	defer a.RWMutex.Unlock()
	a.AllStatus[agentName] = status
	return nil
}

// DeepCopy return a deep copy of AgentInfos
func (a *AgentInfos) DeepCopy() *AgentInfos {
	a.RWMutex.RLock()
	defer a.RWMutex.RUnlock()
	clone := &AgentInfos{
		Agents:    make(map[string]*AgentInfo, len(a.Agents)),
		AllStatus: make(map[string]string, len(a.AllStatus)),
		RWMutex:   sync.RWMutex{},
	}
	for k, v := range a.AllStatus {
		clone.AllStatus[k] = v
	}
	for k, v := range a.Agents {
		if v == nil {
			clone.Agents[k] = nil
			continue
		}
		clone.Agents[k] = v.DeepCopy()
	}
	return clone
}

// GetStatusVal get agent status value
func (a *AgentInfo) GetStatusVal(key string) (string, bool) {
	a.RWMutex.RLock()
	defer a.RWMutex.RUnlock()
	val, ok := a.Status[key]
	return val, ok
}

// SetStatusVal set agent status value
func (a *AgentInfo) SetStatusVal(key, value string) {
	a.RWMutex.Lock()
	defer a.RWMutex.Unlock()
	a.Status[key] = value
}

// DeepCopy return a deep copy of AgentInfo
func (a *AgentInfo) DeepCopy() *AgentInfo {
	a.RWMutex.RLock()
	defer a.RWMutex.RUnlock()
	clone := &AgentInfo{
		Config:    utils.CopyStringMap(a.Config),
		Actions:   utils.CopyStringMap(a.Actions),
		Status:    utils.CopyStringMap(a.Status),
		FaultInfo: utils.CopyStringMap(a.FaultInfo),
		NodeRank:  a.NodeRank,
		HeartBeat: a.HeartBeat,
		RWMutex:   sync.RWMutex{},
	}
	if a.Pos != nil {
		clone.Pos = &common.Position{
			Role:        a.Pos.Role,
			ServerRank:  a.Pos.ServerRank,
			ProcessRank: a.Pos.ProcessRank,
		}
	}
	return clone
}
