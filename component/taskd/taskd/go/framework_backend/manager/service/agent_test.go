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

var agentPos = &common.Position{
	Role:        common.AgentRole,
	ServerRank:  "0",
	ProcessRank: "0",
}

// TestAgentHandler test agent handler
func TestAgentHandler(t *testing.T) {
	mpc := &MsgProcessor{}
	dp := newDataPool()
	convey.Convey("TestAgentHandler get agent fail return error", t, func() {
		msg := createBaseMessage(agentPos, constant.STATUS, 0, "")
		err := mpc.agentHandler(dp, msg)
		convey.So(err.Error(), convey.ShouldContainSubstring, "Agent0")
	})
	msg := createBaseMessage(agentPos, constant.REGISTER, 0, "")
	convey.Convey("TestAgentHandler register agent message success return nil", t, func() {
		err := mpc.agentHandler(dp, msg)
		convey.So(len(dp.Snapshot.AgentInfos.Agents), convey.ShouldEqual, 1)
		convey.So(err, convey.ShouldBeNil)
	})
	_ = mpc.agentRegister(dp, msg)
	convey.Convey("TestAgentHandler msg type status return nil", t, func() {
		msg := createBaseMessage(agentPos, constant.STATUS, constant.RestartTimeCode, "")
		err := mpc.agentHandler(dp, msg)
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("TestAgentHandler msg type keep alive return nil", t, func() {
		msg := createBaseMessage(agentPos, constant.KeepAlive, 0, "")
		err := mpc.agentHandler(dp, msg)
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("TestAgentHandler msg type default return error", t, func() {
		msg := createBaseMessage(agentPos, "default", 0, "")
		err := mpc.agentHandler(dp, msg)
		convey.So(err.Error(), convey.ShouldContainSubstring, "default")
	})
}

// TestAgentStatus test agent status
func TestAgentStatus(t *testing.T) {
	mpc := &MsgProcessor{}
	dp := newDataPool()
	msg := createBaseMessage(agentPos, constant.REGISTER, 0, "")
	agentName := common.AgentRole + "0"
	_ = mpc.agentRegister(dp, msg)
	agentInfo, _ := dp.GetAgent(agentName)
	convey.Convey("TestAgentStatus restart time code return nil", t, func() {
		msg := createBaseMessage(agentPos, "default", constant.RestartTimeCode, "")
		err := mpc.agentStatus(dp, msg, agentName, agentInfo)
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("TestAgentStatus fault rank code return nil", t, func() {
		msg := createBaseMessage(agentPos, "default", constant.FaultRankCode, "")
		err := mpc.agentStatus(dp, msg, agentName, agentInfo)
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("TestAgentStatus exit agent code return nil", t, func() {
		msg := createBaseMessage(agentPos, "default", constant.ExitAgentCode, "")
		err := mpc.agentStatus(dp, msg, agentName, agentInfo)
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("TestAgentHandler msg code default return error", t, func() {
		msg := createBaseMessage(agentPos, "default", 0, "")
		err := mpc.agentStatus(dp, msg, agentName, agentInfo)
		convey.So(err.Error(), convey.ShouldContainSubstring, "0")
	})
}
