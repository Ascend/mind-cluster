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
	"taskd/framework_backend/manager/infrastructure/storage"
	"taskd/toolkit_backend/net/common"
)

var workerPos = &common.Position{
	Role:        common.WorkerRole,
	ServerRank:  "0",
	ProcessRank: "0",
}

// TestWorkerHandler test worker handler
func TestWorkerHandler(t *testing.T) {
	mpc := &MsgProcessor{}
	convey.Convey("TestWorkerHandler get worker fail return error", t, func() {
		dp := newDataPool()
		msg := createBaseMessage(workerPos, constant.STATUS, 0, "")
		err := mpc.workerHandler(dp, msg)
		convey.So(err.Error(), convey.ShouldContainSubstring, "Worker0")
	})
	msg := createBaseMessage(workerPos, constant.REGISTER, 0, "")
	convey.Convey("TestWorkerHandler register worker message success return nil", t, func() {
		dp := newDataPool()
		err := mpc.workerHandler(dp, msg)
		convey.So(len(dp.Snapshot.WorkerInfos.Workers), convey.ShouldEqual, 1)
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("TestWorkerHandler msg type status profiling code return nil", t, func() {
		dp := newDataPool()
		_ = mpc.workerRegister(dp, msg)
		msg := createBaseMessage(workerPos, constant.STATUS, constant.ProfilingAllCloseCode, "")
		err := mpc.workerHandler(dp, msg)
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("TestWorkerHandler msg type status default code return error", t, func() {
		dp := newDataPool()
		_ = mpc.workerRegister(dp, msg)
		msg := createBaseMessage(workerPos, constant.STATUS, 0, "")
		err := mpc.workerHandler(dp, msg)
		convey.So(err.Error(), convey.ShouldEqual, "unknown message status code: 0")
	})
	convey.Convey("TestWorkerHandler msg type keep alive return nil", t, func() {
		dp := newDataPool()
		_ = mpc.workerRegister(dp, msg)
		msg := createBaseMessage(workerPos, constant.KeepAlive, 0, "")
		err := mpc.workerHandler(dp, msg)
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("TestWorkerHandler msg type default return error", t, func() {
		dp := newDataPool()
		_ = mpc.workerRegister(dp, msg)
		msg := createBaseMessage(workerPos, "default", 0, "")
		err := mpc.workerHandler(dp, msg)
		convey.So(err.Error(), convey.ShouldContainSubstring, "default")
	})
}

const commDomainError = 1403
const defaultDomainError = 1430

// TestProfilingStatus test profiling status
func TestProfilingStatus(t *testing.T) {
	workerInfo := &storage.WorkerInfo{Status: make(map[string]string)}
	convey.Convey("TestProfilingStatus profiling all close code return nil", t, func() {
		msg := createBaseMessage(workerPos, "default", constant.ProfilingAllCloseCode, "")
		err := profilingStatus(msg, workerInfo)
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("TestProfilingStatus comm domain err code return error", t, func() {
		msg := createBaseMessage(workerPos, "default", commDomainError, "")
		err := profilingStatus(msg, workerInfo)
		convey.So(err.Error(), convey.ShouldContainSubstring, "3")
	})
	convey.Convey("TestProfilingStatus default domain err code return error", t, func() {
		msg := createBaseMessage(workerPos, "default", defaultDomainError, "")
		err := profilingStatus(msg, workerInfo)
		convey.So(err.Error(), convey.ShouldContainSubstring, "3")
	})
}
