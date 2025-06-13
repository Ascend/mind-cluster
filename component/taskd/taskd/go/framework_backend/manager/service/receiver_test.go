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
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"taskd/common/constant"
	"taskd/common/utils"
	"taskd/framework_backend/manager/infrastructure/storage"
	"taskd/toolkit_backend/net"
	"taskd/toolkit_backend/net/common"
)

type testReturn struct {
	msg     *common.Message
	msgBody storage.MsgBody
	err     error
}

const fiveHundred = 500

// TestReceiveMsg test receive message
func TestReceiveMsg(t *testing.T) {
	mrc := &MsgReceiver{}
	mq := &storage.MsgQueue{Queue: make([]storage.BaseMessage, 0)}
	tool, _ := net.InitNetwork(&common.TaskNetConfig{
		Pos:        common.Position{Role: common.MgrRole, ServerRank: "0", ProcessRank: "-1"},
		ListenAddr: constant.DefaultIP + constant.MgrPort})
	mockMsg := &common.Message{Uuid: "test_uuid", BizType: "test_biz_type", Src: workerPos,
		Dst: workerPos, Body: utils.ObjToString(&storage.MsgBody{})}
	patch := gomonkey.ApplyFuncReturn(tool.ReceiveMessage, mockMsg)
	defer patch.Reset()
	convey.Convey("TestReceiveMsg test enqueue success wait exit return nil", t, func() {
		ctx, cancel := context.WithTimeout(context.Background(), fiveHundred*time.Millisecond)
		defer cancel()
		testReturn := &testReturn{}
		go func() {
			testReturn.msg, testReturn.msgBody, testReturn.err = mrc.ReceiveMsg(mq, tool, ctx)
		}()
		convey.So(testReturn.msg, convey.ShouldBeNil)
		convey.So(testReturn.msgBody, convey.ShouldResemble, storage.MsgBody{})
		convey.So(testReturn.err, convey.ShouldBeNil)
	})
}
