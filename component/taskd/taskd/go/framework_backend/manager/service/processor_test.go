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
	"testing"

	"github.com/smartystreets/goconvey/convey"

	"taskd/common/constant"
	"taskd/toolkit_backend/net/common"
)

// TestMsgProcessor test message processor
func TestMsgProcessor(t *testing.T) {
	mpc := &MsgProcessor{}
	dp := newDataPool()
	convey.Convey("TestMsgProcessor msg processor handler success return nil", t, func() {
		msg := createBaseMessage(&common.Position{Role: common.MgrRole, ServerRank: "0"}, "", 0, "")
		err := mpc.MsgProcessor(dp, msg)
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("TestMsgProcessor msg role default return error", t, func() {
		msg := createBaseMessage(&common.Position{Role: "default", ServerRank: "0"}, "", 0, "")
		err := mpc.MsgProcessor(dp, msg)
		convey.So(err.Error(), convey.ShouldContainSubstring, "default")
	})
	convey.Convey("TestMsgProcessor worker handler get worker fail return error", t, func() {
		msg := createBaseMessage(&common.Position{Role: common.WorkerRole, ProcessRank: "0"}, constant.STATUS, 0, "")
		err := mpc.MsgProcessor(dp, msg)
		convey.So(err.Error(), convey.ShouldContainSubstring, "Worker0")
	})
}
