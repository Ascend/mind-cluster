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
	"time"

	"ascend-common/common-utils/hwlog"
	"taskd/common/constant"
	"taskd/toolkit_backend/net"
	"taskd/toolkit_backend/net/common"
)

// MsgSender about send message
type MsgSender struct {
	RequestChan chan SendGrpcMsg
}

// SendGrpcMsg   send msg with grpc server
type SendGrpcMsg struct {
	Uuid    string
	MsgType string
	MsgBody string
	Dst     *common.Position
}

// SendMessage send message with grpc server
func (msd *MsgSender) SendMessage(tool *net.NetInstance, req SendGrpcMsg) {
	for resendTimes := 0; resendTimes < constant.MaxResendTimes; resendTimes++ {
		ack, err := tool.SyncSendMessage(req.Uuid, req.MsgType, req.MsgBody, req.Dst)
		if err != nil {
			hwlog.RunLog.Errorf("mgr send message failed %v", err)
			time.Sleep(time.Second * constant.ResendSeconds)
			continue
		}
		hwlog.RunLog.Infof("mgr send message success  %v", ack)
		return
	}
	hwlog.RunLog.Errorf("mgr resend message more times than limit %v", constant.MaxResendTimes)
}
