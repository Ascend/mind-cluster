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

	"taskd/common/constant"
	"taskd/framework_backend/manager/infrastructure/storage"
	"taskd/toolkit_backend/net/common"
)

// MsgProcessor about process message
type MsgProcessor struct {
}

// MsgProcessor process message
func (mpc *MsgProcessor) MsgProcessor(dataPool *storage.DataPool, msg storage.BaseMessage) error {
	handlers := map[string]func(*storage.DataPool, storage.BaseMessage) error{
		common.MgrRole:       mpc.managerHandler,
		common.WorkerRole:    mpc.workerHandler,
		common.AgentRole:     mpc.agentHandler,
		constant.ClusterRole: mpc.clusterHandler,
	}
	handler, exists := handlers[msg.Header.Src.Role]
	if !exists {
		return fmt.Errorf("unknown role: %v", msg.Header.Src.Role)
	}
	if err := handler(dataPool, msg); err != nil {
		return fmt.Errorf("%s message error: %v", msg.Header.Src.Role, err)
	}
	return nil
}
