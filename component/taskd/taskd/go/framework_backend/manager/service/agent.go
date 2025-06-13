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

// Package service is to provide other service tools, i.e. clusterd
package service

import (
	"fmt"
	"sync"
	"time"

	"taskd/common/constant"
	"taskd/framework_backend/manager/infrastructure/storage"
)

func (mpc *MsgProcessor) agentHandler(dataPool *storage.DataPool, data storage.BaseMessage) error {
	if data.Body.MsgType == constant.REGISTER {
		return mpc.agentRegister(dataPool, data)
	}
	agentName := data.Header.Src.Role + data.Header.Src.ServerRank
	agentInfo, err := dataPool.GetAgent(agentName)
	if err != nil {
		return err
	}
	switch data.Body.MsgType {
	case constant.STATUS:
		err = mpc.agentStatus(dataPool, data, agentName, agentInfo)
	case constant.KeepAlive:
		agentInfo.HeartBeat = time.Now()
	default:
		return fmt.Errorf("unknown message type: %v", data.Body.MsgType)
	}
	err = dataPool.UpdateAgent(agentName, agentInfo)
	return err
}

func (mpc *MsgProcessor) agentRegister(dataPool *storage.DataPool, data storage.BaseMessage) error {
	agentInfo := &storage.AgentInfo{
		Status:    map[string]string{constant.REGISTER: constant.REGISTER},
		NodeRank:  data.Header.Src.ServerRank,
		Pos:       data.Header.Src,
		HeartBeat: time.Now(),
		RWMutex:   sync.RWMutex{},
	}
	agentName := data.Header.Src.Role + data.Header.Src.ServerRank
	err := dataPool.RegisterAgent(agentName, agentInfo)
	return err
}

func (mpc *MsgProcessor) agentStatus(dataPool *storage.DataPool, data storage.BaseMessage, agentName string,
	agentInfo *storage.AgentInfo) error {
	switch data.Body.Code {
	case constant.RestartTimeCode:
		agentInfo.Status[constant.ReportRestartTime] = data.Body.Message
	case constant.FaultRankCode:
		agentInfo.Status[constant.ReportFaultRank] = data.Body.Message
		dataPool.Snapshot.AgentInfos.AllStatus[agentName] = data.Body.Message
	case constant.ExitAgentCode:
		agentInfo.Status[constant.Exit] = data.Body.Message
		dataPool.Snapshot.AgentInfos.AllStatus[agentName] = data.Body.Message
	default:
		return fmt.Errorf("unknown message status code: %v", data.Body.Code)
	}
	return nil
}
