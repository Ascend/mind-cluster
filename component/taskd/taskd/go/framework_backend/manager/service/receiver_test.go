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
	"encoding/json"
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/common-utils/hwlog"
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
	hwLogConfig := hwlog.LogConfig{OnlyToStdout: true}
	hwlog.InitRunLogger(&hwLogConfig, context.Background())
	customLog := hwlog.SetCustomLogger(hwlog.RunLog)
	mrc := &MsgReceiver{}
	mq := &storage.MsgQueue{Queue: make([]storage.BaseMessage, 0)}
	tool, _ := net.InitNetwork(&common.TaskNetConfig{
		Pos:        common.Position{Role: common.MgrRole, ServerRank: "0", ProcessRank: "-1"},
		ListenAddr: constant.DefaultIP + constant.MgrPort}, customLog)
	mockMsg := &common.Message{Uuid: "test_uuid", BizType: "test_biz_type", Src: workerPos,
		Dst: workerPos, Body: utils.ObjToString(&storage.MsgBody{})}
	patch := gomonkey.ApplyMethod(tool, "ReceiveMessage", func(*net.NetInstance) *common.Message {
		defer func() { mockMsg = nil }()
		return mockMsg
	})
	defer patch.Reset()
	testReturn := &testReturn{}
	convey.Convey("TestReceiveMsg test enqueue failed return error", t, func() {
		mq := &storage.MsgQueue{Queue: make([]storage.BaseMessage, constant.MaxMsgQueueLength)}
		testReturn.msg, testReturn.msgBody, testReturn.err = mrc.ReceiveMsg(mq, tool)
		convey.So(testReturn.msg, convey.ShouldEqual, &common.Message{Uuid: "test_uuid", BizType: "test_biz_type",
			Src: workerPos, Dst: workerPos, Body: utils.ObjToString(&storage.MsgBody{})})
		convey.So(testReturn.msgBody, convey.ShouldResemble, storage.MsgBody{})
		convey.So(testReturn.err, convey.ShouldEqual, fmt.Errorf("message queue is full"))
	})
	convey.Convey("TestReceiveMsg test receive message failed log error return nil", t, func() {
		patch := gomonkey.ApplyMethodReturn(tool, "ReceiveMessage", nil)
		defer patch.Reset()
		testReturn.msg, testReturn.msgBody, testReturn.err = mrc.ReceiveMsg(mq, tool)
		convey.So(testReturn.msg, convey.ShouldBeNil)
		convey.So(testReturn.msgBody, convey.ShouldResemble, storage.MsgBody{})
		convey.So(testReturn.err, convey.ShouldBeNil)
	})
	convey.Convey("TestReceiveMsg test unmarshal failed log error return nil", t, func() {
		mockUnmarshal := gomonkey.ApplyFuncReturn(json.Unmarshal, fmt.Errorf("test error"))
		defer mockUnmarshal.Reset()
		testReturn.msg, testReturn.msgBody, testReturn.err = mrc.ReceiveMsg(mq, tool)
		convey.So(testReturn.msg, convey.ShouldBeNil)
		convey.So(testReturn.msgBody, convey.ShouldResemble, storage.MsgBody{})
		convey.So(testReturn.err, convey.ShouldBeNil)
	})
	convey.Convey("TestReceiveMsg test enqueue success return nil", t, func() {
		testReturn.msg, testReturn.msgBody, testReturn.err = mrc.ReceiveMsg(mq, tool)
		convey.So(testReturn.msg, convey.ShouldBeNil)
		convey.So(testReturn.msgBody, convey.ShouldResemble, storage.MsgBody{})
		convey.So(testReturn.err, convey.ShouldBeNil)
	})
}
