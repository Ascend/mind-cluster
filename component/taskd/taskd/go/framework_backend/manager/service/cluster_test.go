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

var clusterPos = &common.Position{
	Role:        constant.ClusterRole,
	ServerRank:  constant.ClusterDRank,
	ProcessRank: "0",
}

// TestClusterHandler test cluster handler
func TestClusterHandler(t *testing.T) {
	mpc := &MsgProcessor{}
	dp := newDataPool()
	convey.Convey("TestClusterHandler msg type action, code 0 return error", t, func() {
		msg := createBaseMessage(clusterPos, constant.Action, 0, "")
		err := mpc.clusterHandler(dp, msg)
		convey.So(err.Error(), convey.ShouldContainSubstring, "0")
	})
	convey.Convey("TestClusterHandler msg type action, profiling code return nil", t, func() {
		msg := createBaseMessage(clusterPos, constant.Action, constant.ProfilingAllCloseCmdCode, "")
		err := mpc.clusterHandler(dp, msg)
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("TestClusterHandler msg type keep alive return nil", t, func() {
		msg := createBaseMessage(&common.Position{Role: constant.ClusterRole, ServerRank: "0"},
			constant.KeepAlive, 0, "")
		err := mpc.clusterHandler(dp, msg)
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("TestClusterHandler msg type default return error", t, func() {
		msg := createBaseMessage(&common.Position{Role: constant.ClusterRole, ServerRank: "0"},
			"default", 0, "")
		err := mpc.clusterHandler(dp, msg)
		convey.So(err.Error(), convey.ShouldContainSubstring, "default")
	})
}
