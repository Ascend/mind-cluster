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
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"taskd/common/constant"
	"taskd/toolkit_backend/net"
	"taskd/toolkit_backend/net/common"
)

// TestSendMessage test manager grpc send message
func TestSendMessage(t *testing.T) {
	msd := &MsgSender{}
	tool, _ := net.InitNetwork(&common.TaskNetConfig{
		Pos:        common.Position{Role: common.MgrRole, ServerRank: "0", ProcessRank: "-1"},
		ListenAddr: constant.DefaultIP + constant.MgrPort})
	req := SendGrpcMsg{Uuid: "test_uuid", MsgType: "test_msgType", MsgBody: "test_msgBody", Dst: workerPos}
	convey.Convey("TestSendMessage test grpc send message failed", t, func() {
		patch := gomonkey.ApplyFuncReturn(tool.SyncSendMessage, &common.Ack{}, fmt.Errorf("test error"))
		defer patch.Reset()
		convey.So(capturePanic(func() { msd.SendMessage(tool, req) }), convey.ShouldBeNil)
	})
	convey.Convey("TestSendMessage test grpc send message failed", t, func() {
		patch := gomonkey.ApplyFuncReturn(tool.SyncSendMessage, &common.Ack{}, nil)
		defer patch.Reset()
		convey.So(capturePanic(func() { msd.SendMessage(tool, req) }), convey.ShouldBeNil)
	})
}
