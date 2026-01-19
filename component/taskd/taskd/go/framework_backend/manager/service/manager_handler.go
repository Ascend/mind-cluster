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
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/interface/grpc/recover"
	"taskd/common/constant"
	"taskd/common/utils"
	"taskd/framework_backend/manager/infrastructure/storage"
)

const (
	maxRegRetryTime = 10
)

func (mpc *MsgProcessor) managerHandler(dataPool *storage.DataPool, msg storage.BaseMessage) error {
	hwlog.RunLog.Infof("managerHandler, msg: %v", msg)
	mgrInfo, err := dataPool.GetMgr()
	if err != nil {
		return err
	}
	switch msg.Body.MsgType {
	case constant.Action:
		if msg.Body.Code == constant.RestartTimeCode {
			mgrInfo.Status[constant.ReportRestartTime] = msg.Body.Message
			return nil
		}
		if msg.Body.Code == constant.ProcessManageRecoverSignal {
			mgrInfo.Status[constant.Actions] = msg.Body.Extension[constant.Actions]
			mgrInfo.Status[constant.SignalType] = msg.Body.Extension[constant.SignalType]
		}
		if msg.Body.Code == constant.FaultRecoverCode {
			mgrInfo.Status[constant.FaultRecover] = msg.Body.Message
		}
		if msg.Body.Code == constant.ReplyToClusterDCode {
			go mpc.replyToClusterD(msg.Body.Extension)
		}
	default:
		return fmt.Errorf("unknown message type: %v", msg.Body.MsgType)
	}
	err = dataPool.UpdateMgr(mgrInfo)
	return err
}

func (mpc *MsgProcessor) replyToClusterD(result map[string]string) {
	hwlog.RunLog.Infof("reply to clusterD result: %v", result)
	addr, err := utils.GetClusterdAddr()
	if err != nil {
		hwlog.RunLog.Errorf("get clusterd address err: %v", err)
		return
	}
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		hwlog.RunLog.Errorf("init clusterd connect err: %v", err)
		return
	}
	client := pb.NewRecoverClient(conn)
	if result[constant.StressTestResultStr] != "" {
		err = mpc.replyStressTestMsg(result[constant.StressTestResultStr], client)
	}
	if result[constant.SwitchNicResultStr] != "" {
		err = mpc.replySwitchNicMsg(result[constant.SwitchNicResultStr], client)
	}
	if err != nil {
		hwlog.RunLog.Errorf("reply result:%v to clusterd err: %v", result, err)
		return
	}
	hwlog.RunLog.Infof("reply to clusterD result success: %v", result)
}

func (mpc *MsgProcessor) replyStressTestMsg(result string, client pb.RecoverClient) error {
	rankResult, err := utils.StringToObj[*pb.StressTestResult](result)
	if err != nil {
		return fmt.Errorf("parse stress result: %v, err: %v", result, err)
	}
	for i := 0; i < maxRegRetryTime; i++ {
		_, err = client.ReplyStressTestResult(context.TODO(), rankResult)
		if err == nil {
			return nil
		}
		time.Sleep(1 * time.Second)
		err = fmt.Errorf("reply StressTestResult err: %v", err)
	}
	return err
}

func (mpc *MsgProcessor) replySwitchNicMsg(result string, client pb.RecoverClient) error {
	rankResult, err := utils.StringToObj[*pb.SwitchResult](result)
	if err != nil {
		return fmt.Errorf("parse switch nic resul: %v, err: %v", result, err)
	}
	for i := 0; i < maxRegRetryTime; i++ {
		_, err = client.ReplySwitchNicResult(context.TODO(), rankResult)
		if err == nil {
			return nil
		}
		time.Sleep(1 * time.Second)
		err = fmt.Errorf("reply SwitchNicResult err: %v", err)
	}
	return err
}
