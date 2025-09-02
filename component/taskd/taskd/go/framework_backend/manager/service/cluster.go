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
	"strconv"
	"time"

	"ascend-common/common-utils/hwlog"
	"taskd/common/constant"
	"taskd/framework_backend/manager/infrastructure/storage"
)

func (mpc *MsgProcessor) clusterHandler(dataPool *storage.DataPool, data storage.BaseMessage) error {
	clusterName := data.Header.Src.ServerRank
	clusterInfo, err := dataPool.GetCluster(clusterName)
	if err != nil {
		clusterInfo = dataPool.RegisterCluster(clusterName)
	}
	switch data.Body.MsgType {
	case constant.Action:
		err := mpc.clusterAction(data, clusterInfo)
		if err != nil {
			hwlog.RunLog.Errorf("clusterHandler error : %v", err)
			return err
		}
	case constant.KeepAlive:
		clusterInfo.HeartBeat = time.Now()
	default:
		return fmt.Errorf("unknown message type: %v", data.Body.MsgType)
	}
	err = dataPool.UpdateCluster(clusterName, clusterInfo)
	return err
}

func (mpc *MsgProcessor) clusterAction(data storage.BaseMessage, clusterInfo *storage.ClusterInfo) error {
	switch data.Body.Code {
	case constant.SwitchNicCode:
		clusterInfo.Command[constant.GlobalRankKey] = data.Body.Extension[constant.GlobalRankKey]
		clusterInfo.Command[constant.GlobalOpKey] = data.Body.Extension[constant.GlobalOpKey]
		clusterInfo.Command[constant.SwitchNicUUID] = data.Header.Uuid
		clusterInfo.Command[constant.SwitchJobID] = data.Body.Extension[constant.SwitchJobID]
	case constant.StressTestCode:
		clusterInfo.Command[constant.StressTestRankOPStr] = data.Body.Extension[constant.StressTestRankOPStr]
		clusterInfo.Command[constant.StressTestUUID] = data.Header.Uuid
		clusterInfo.Command[constant.StressTestJobID] = data.Body.Extension[constant.StressTestJobID]
	default:
		defaultDomainCmd, commDomainCmd, err := profilingCmd(data.Body.Code)
		if err != nil {
			return err
		}
		clusterInfo.Command[constant.DefaultDomainCmd] = defaultDomainCmd
		clusterInfo.Command[constant.CommDomainCmd] = commDomainCmd
	}
	return nil
}

func profilingCmd(actionCode int32) (string, string, error) {
	profilingSwitch, err := BizCodeToProfilingCmd(actionCode)
	if err != nil {
		return "", "", err
	}
	defaultDomainCmd := strconv.FormatBool(profilingSwitch.DefaultDomainAble)
	commDomainCmd := strconv.FormatBool(profilingSwitch.CommDomainAble)
	return defaultDomainCmd, commDomainCmd, nil
}

// BizCodeToProfilingCmd convert code to ProfilingDomainCmd
func BizCodeToProfilingCmd(code int32) (constant.ProfilingDomainCmd, error) {
	profilingCmdMap := map[int32]constant.ProfilingDomainCmd{
		constant.ProfilingAllCloseCmdCode: {
			DefaultDomainAble: false,
			CommDomainAble:    false,
		},
		constant.ProfilingDefaultDomainOnCode: {
			DefaultDomainAble: true,
			CommDomainAble:    false,
		},
		constant.ProfilingCommDomainOnCode: {
			DefaultDomainAble: false,
			CommDomainAble:    true,
		},
		constant.ProfilingAllOnCmdCode: {
			DefaultDomainAble: true,
			CommDomainAble:    true,
		},
	}
	if config, exists := profilingCmdMap[code]; exists {
		return config, nil
	}
	return constant.ProfilingDomainCmd{}, fmt.Errorf("cannot convert code %d to ProfilingDomainCmd", code)
}
