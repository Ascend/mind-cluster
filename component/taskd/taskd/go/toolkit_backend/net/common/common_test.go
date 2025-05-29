/*
Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.

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
// Package common defines common constants and types used by the toolkit backend.
package common

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestRoleLevel(t *testing.T) {
	convey.Convey("Test RoleLevel function", t, func() {
		convey.Convey("Valid roles should return correct level", func() {
			convey.So(RoleLevel(MgrRole), convey.ShouldEqual, MgrLevel)
			convey.So(RoleLevel(ProxyRole), convey.ShouldEqual, ProxyLevel)
			convey.So(RoleLevel(AgentRole), convey.ShouldEqual, AgentLevel)
			convey.So(RoleLevel(WorkerRole), convey.ShouldEqual, WorkerLevel)
		})
		convey.Convey("Invalid role should return -1", func() {
			convey.So(RoleLevel("InvalidRole"), convey.ShouldEqual, -1)
		})
	})
}

func TestRoleHasProcessProperty(t *testing.T) {
	convey.Convey("Test RoleHasProcessProperty function", t, func() {
		convey.Convey("Worker role should have process property", func() {
			convey.So(RoleHasProcessProperty(WorkerRole), convey.ShouldBeTrue)
		})
		convey.Convey("Other roles should not have process property", func() {
			convey.So(RoleHasProcessProperty(MgrRole), convey.ShouldBeFalse)
			convey.So(RoleHasProcessProperty(ProxyRole), convey.ShouldBeFalse)
			convey.So(RoleHasProcessProperty(AgentRole), convey.ShouldBeFalse)
		})
		convey.Convey("Invalid role should return false", func() {
			convey.So(RoleHasProcessProperty("InvalidRole"), convey.ShouldBeFalse)
		})
	})
}

func TestRoleRecvBuffer(t *testing.T) {
	convey.Convey("Test RoleRecvBuffer function", t, func() {
		convey.Convey("Valid roles should return correct buffer size", func() {
			convey.So(RoleRecvBuffer(MgrRole), convey.ShouldEqual, MgrBufSize)
			convey.So(RoleRecvBuffer(ProxyRole), convey.ShouldEqual, ProxyBufSize)
			convey.So(RoleRecvBuffer(AgentRole), convey.ShouldEqual, AgentBufSize)
			convey.So(RoleRecvBuffer(WorkerRole), convey.ShouldEqual, WorkerBufSize)
		})
		convey.Convey("Invalid role should return -1", func() {
			convey.So(RoleRecvBuffer("InvalidRole"), convey.ShouldEqual, -1)
		})
	})
}

func TestRoleWorkerNum(t *testing.T) {
	convey.Convey("Test RoleWorkerNum function", t, func() {
		convey.Convey("Valid roles should return correct worker number", func() {
			convey.So(RoleWorkerNum(MgrRole), convey.ShouldEqual, MgrGrNum)
			convey.So(RoleWorkerNum(ProxyRole), convey.ShouldEqual, ProxyGrNum)
			convey.So(RoleWorkerNum(AgentRole), convey.ShouldEqual, AgentGrNum)
			convey.So(RoleWorkerNum(WorkerRole), convey.ShouldEqual, WorkerGrNum)
		})
		convey.Convey("Invalid role should return -1", func() {
			convey.So(RoleWorkerNum("InvalidRole"), convey.ShouldEqual, -1)
		})
	})
}
