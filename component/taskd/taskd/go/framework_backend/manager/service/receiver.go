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
	"encoding/json"

	"ascend-common/common-utils/hwlog"
	"taskd/framework_backend/manager/infrastructure/storage"
	"taskd/toolkit_backend/net"
	"taskd/toolkit_backend/net/common"
)

// MsgReceiver about receive message
type MsgReceiver struct {
}

// ReceiveMsg receive msg from grpc server
func (mrc *MsgReceiver) ReceiveMsg(mq *storage.MsgQueue, tool *net.NetInstance) (
	*common.Message, storage.MsgBody, error) {
	msg := tool.ReceiveMessage()
	if msg == nil {
		return nil, storage.MsgBody{}, nil
	}
	hwlog.RunLog.Debugf("mgr recv message: %v", msg)
	var msgBody storage.MsgBody
	err := json.Unmarshal([]byte(msg.Body), &msgBody)
	if err != nil {
		hwlog.RunLog.Errorf("unmarshal failed: %v", err)
		return nil, storage.MsgBody{}, nil
	}
	data := mq.NewMsg(msg.Uuid, msg.BizType, msg.Src, msgBody)
	err = mq.Enqueue(data)
	hwlog.RunLog.Debugf("enqueue msg: %v", data)
	if err != nil {
		hwlog.RunLog.Errorf("enqueue failed: %v", err)
		return msg, msgBody, err
	}
	return nil, storage.MsgBody{}, nil
}
