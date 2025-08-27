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
	if agent, exists := a.Agents[agentName]; exists {
		agentInfo := agent.getAgent()
		a.RWMutex.RUnlock()
		return agentInfo, nil
	}
	a.RWMutex.RUnlock()
	return nil, fmt.Errorf("agent name is unregistered : %v", agentName)
}

func (a *AgentInfo) getAgent() *AgentInfo {
	a.RWMutex.RLock()
	defer a.RWMutex.RUnlock()
	return &AgentInfo{
		Config:    a.Config,
		Actions:   a.Actions,
		Status:    a.Status,
		NodeRank:  a.NodeRank,
		HeartBeat: a.HeartBeat,
		FaultInfo: a.FaultInfo,
		Pos:       a.Pos,
		RWMutex:   sync.RWMutex{},
	}
}

func (a *AgentInfos) updateAgent(agentName string, newAgent *AgentInfo) error {
	a.RWMutex.Lock()
	defer a.RWMutex.Unlock()
	a.Agents[agentName] = &AgentInfo{
		Config:    newAgent.Config,
		Actions:   newAgent.Actions,
		Status:    newAgent.Status,
		NodeRank:  newAgent.NodeRank,
		HeartBeat: newAgent.HeartBeat,
		FaultInfo: newAgent.FaultInfo,
		Pos:       newAgent.Pos,
	}
	return nil
}
