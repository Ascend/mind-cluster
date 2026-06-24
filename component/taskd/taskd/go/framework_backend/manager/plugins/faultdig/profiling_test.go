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

// Package faultdig for taskd manager plugin
package faultdig

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/common-utils/hwlog"
	"taskd/common/constant"
	"taskd/framework_backend/manager/infrastructure"
	"taskd/framework_backend/manager/infrastructure/storage"
)

const (
	worker0Name = "worker0"
	worker1Name = "worker1"
)

func getProfilingPlugin() *PfPlugin {
	return NewProfilingPlugin().(*PfPlugin)
}

func getDemoSnapshot() storage.SnapShot {
	return storage.SnapShot{
		WorkerInfos: &storage.WorkerInfos{
			Workers: map[string]*storage.WorkerInfo{
				worker0Name: {
					Status: map[string]string{
						constant.DefaultDomainStatus: constant.On,
						constant.CommDomainStatus:    constant.On,
					},
				},
			},
		},
		ClusterInfos: &storage.ClusterInfos{
			Clusters: map[string]*storage.ClusterInfo{
				constant.ClusterDRank: {
					Command: map[string]string{
						constant.DefaultDomainCmd: "true",
						constant.CommDomainCmd:    "true",
					},
				},
			},
		},
	}
}

func getDemoWorkerStatus() *workerExecStatus {
	return &workerExecStatus{
		workers: map[string]constant.ProfilingResult{
			worker0Name: {
				DefaultDomain: constant.ProfilingOnStatus,
				CommDomain:    constant.ProfilingOnStatus,
			},
			worker1Name: {
				DefaultDomain: constant.ProfilingOnStatus,
				CommDomain:    constant.ProfilingExpStatus,
			},
		},
		cmd: constant.ProfilingDomainCmd{
			DefaultDomainAble: true,
			CommDomainAble:    true,
		},
		defaultDomainState: constant.ProfilingWorkerClosedState,
		commDomainState:    constant.ProfilingWorkerClosedState,
	}
}

func TestMain(m *testing.M) {
	if err := setup(); err != nil {
		return
	}
	code := m.Run()
	fmt.Printf("exit_code = %v\n", code)
}

func setup() error {
	return initLog()
}

func initLog() error {
	logConfig := &hwlog.LogConfig{
		OnlyToStdout: true,
	}
	if err := hwlog.InitRunLogger(logConfig, context.Background()); err != nil {
		fmt.Printf("init hwlog failed, %v\n", err)
		return fmt.Errorf("init hwlog failed")
	}
	return nil
}

func TestGetAllWorkerName(t *testing.T) {
	convey.Convey("get worker name from workerStatus.workers should right", t, func() {
		plugin := getProfilingPlugin()
		plugin.workerStatus.workers[worker0Name] = constant.ProfilingResult{}
		names := []string{worker0Name}
		convey.ShouldEqual(plugin.getAllWorkerName(), names)
	})
}

func TestChangeCmd(t *testing.T) {
	cmd := constant.ProfilingDomainCmd{
		DefaultDomainAble: true,
		CommDomainAble:    true,
	}
	convey.Convey("when change cmd, then pullMsg should not be empty", t, func() {
		plugin := getProfilingPlugin()
		plugin.changeCmd(cmd)
		convey.ShouldEqual(len(plugin.pullMsg), 1)
	})
	convey.Convey("when change cmd and retry.cmdChanged is true, then PullMsg returns pullMsg", t, func() {
		plugin := getProfilingPlugin()
		plugin.changeCmd(cmd)
		plugin.retry.cmdChanged = true
		msg, err := plugin.PullMsg()
		convey.ShouldBeNil(err)
		convey.ShouldEqual(len(msg), 1)
		convey.ShouldBeFalse(plugin.retry.cmdChanged)
		convey.ShouldBeFalse(plugin.retry.notMeetStart.IsZero())
	})
	convey.Convey("when change cmd but retry not triggered, then PullMsg returns nil", t, func() {
		plugin := getProfilingPlugin()
		plugin.changeCmd(cmd)
		plugin.retry = retryState{notMeetStart: time.Now().Add(-2 * time.Second)}
		plugin.workerStatus = workerExecStatus{
			cmd:                cmd,
			defaultDomainState: constant.ProfilingWorkerWaitOpenState,
			commDomainState:    constant.ProfilingWorkerWaitOpenState,
		}
		msg, err := plugin.PullMsg()
		convey.ShouldBeNil(err)
		convey.ShouldBeNil(msg)
	})
}

func TestHandle(t *testing.T) {
	convey.Convey("when handle finish, when release token", t, func() {
		plugin := getProfilingPlugin()
		plugin.cmd = constant.ProfilingDomainCmd{DefaultDomainAble: true, CommDomainAble: true}
		plugin.workerStatus.cmd = plugin.cmd
		plugin.workerStatus.workers[worker0Name] = constant.ProfilingResult{
			DefaultDomain: constant.ProfilingOffStatus,
			CommDomain:    constant.ProfilingOffStatus,
		}
		handle, err := plugin.Handle()
		convey.ShouldBeNil(err)
		convey.ShouldEqual(handle.Stage, constant.HandleStageFinal)
	})
}

func TestHandleProfilingResult(t *testing.T) {
	convey.Convey("when snapshot has new result, then worker status is updated", t, func() {
		snapshot := getDemoSnapshot()
		plugin := getProfilingPlugin()
		plugin.handleProfilingResult(snapshot)
		convey.ShouldEqual(len(plugin.workerStatus.workers), 1)
	})
}

func TestGetProfilingCmd(t *testing.T) {
	convey.Convey("when predicate snapshot with new cmd and result, then should candidate", t, func() {
		snapshot := getDemoSnapshot()
		plugin := getProfilingPlugin()
		cmd, err := plugin.getProfilingCmd(snapshot)
		convey.ShouldBeNil(err)
		convey.ShouldBeTrue(cmd.DefaultDomainAble)
		convey.ShouldBeTrue(cmd.CommDomainAble)
	})
}

func TestPredicate(t *testing.T) {
	snapshot := getDemoSnapshot()
	plugin := getProfilingPlugin()
	convey.Convey("when predicate snapshot with new cmd and result, then should candidate", t, func() {
		predicate, err := plugin.Predicate(snapshot)
		convey.ShouldBeNil(err)
		convey.ShouldEqual(predicate.CandidateStatus, constant.CandidateStatus)
	})

	convey.Convey("when predicate snapshot, both cmd and result are fail, then should unselect", t, func() {
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		patches.ApplyPrivateMethod(plugin, "getProfilingCmd",
			func(*PfPlugin, storage.SnapShot) (constant.ProfilingDomainCmd, error) {
				return constant.ProfilingDomainCmd{}, fmt.Errorf("error")
			}).ApplyPrivateMethod(plugin, "handleProfilingResult",
			func(*PfPlugin, storage.SnapShot) {})
		predicate, err := plugin.Predicate(snapshot)
		convey.ShouldBeNil(err)
		convey.ShouldEqual(predicate.CandidateStatus, constant.UnselectStatus)
	})
}

func TestCalcNewState(t *testing.T) {
	convey.Convey("when all worker default on, then state is on; if some is exp, then state is exp", t, func() {
		status := getDemoWorkerStatus()
		s1, s2 := status.calcNewState()
		convey.ShouldEqual(s1, constant.ProfilingOnStatus)
		convey.ShouldEqual(s2, constant.ProfilingExpStatus)
	})
}

func TestRetryStateUpdate(t *testing.T) {
	convey.Convey("when cmdChanged is true, then notMeetStart is reset and cmdChanged is set", t, func() {
		r := retryState{notMeetStart: time.Now().Add(-3 * time.Second)}
		status := workerExecStatus{
			cmd:                constant.ProfilingDomainCmd{DefaultDomainAble: true, CommDomainAble: true},
			defaultDomainState: constant.ProfilingWorkerWaitOpenState,
			commDomainState:    constant.ProfilingWorkerWaitOpenState,
		}
		r.update(true, status)
		convey.ShouldBeTrue(r.notMeetStart.IsZero())
		convey.ShouldBeTrue(r.cmdChanged)
	})

	convey.Convey("when cmdChanged is false and status not meet cmd, then notMeetStart is set if zero", t, func() {
		r := retryState{}
		status := workerExecStatus{
			cmd:                constant.ProfilingDomainCmd{DefaultDomainAble: true, CommDomainAble: true},
			defaultDomainState: constant.ProfilingWorkerWaitOpenState,
			commDomainState:    constant.ProfilingWorkerWaitOpenState,
		}
		r.update(false, status)
		convey.ShouldBeFalse(r.notMeetStart.IsZero())
		convey.ShouldBeFalse(r.cmdChanged)
	})

	convey.Convey("when cmdChanged is false and status not meet cmd, then notMeetStart is kept if already set", t, func() {
		originalStart := time.Now().Add(-2 * time.Second)
		r := retryState{notMeetStart: originalStart}
		status := workerExecStatus{
			cmd:                constant.ProfilingDomainCmd{DefaultDomainAble: true, CommDomainAble: true},
			defaultDomainState: constant.ProfilingWorkerWaitOpenState,
			commDomainState:    constant.ProfilingWorkerWaitOpenState,
		}
		r.update(false, status)
		convey.ShouldEqual(r.notMeetStart, originalStart)
	})

	convey.Convey("when cmdChanged is false and status meets cmd, then notMeetStart is reset", t, func() {
		r := retryState{notMeetStart: time.Now().Add(-3 * time.Second)}
		status := workerExecStatus{
			cmd:                constant.ProfilingDomainCmd{DefaultDomainAble: true, CommDomainAble: true},
			defaultDomainState: constant.ProfilingWorkerOpenedState,
			commDomainState:    constant.ProfilingWorkerOpenedState,
		}
		r.update(false, status)
		convey.ShouldBeTrue(r.notMeetStart.IsZero())
	})
}

func TestRetryStateExceedLimit(t *testing.T) {
	convey.Convey("when notMeetStart is zero, then exceedLimit returns false", t, func() {
		r := retryState{}
		convey.ShouldBeFalse(r.exceedLimit())
	})

	convey.Convey("when notMeetStart is within retry interval, then exceedLimit returns false", t, func() {
		r := retryState{notMeetStart: time.Now().Add(-3 * time.Second)}
		convey.ShouldBeFalse(r.exceedLimit())
	})

	convey.Convey("when notMeetStart exceeds retry interval, then exceedLimit returns true", t, func() {
		r := retryState{notMeetStart: time.Now().Add(-6 * time.Second)}
		convey.ShouldBeTrue(r.exceedLimit())
	})
}

func TestRetryStateShouldPullMsg(t *testing.T) {
	convey.Convey("when cmdChanged is true, then shouldPullMsg returns true", t, func() {
		r := retryState{cmdChanged: true}
		convey.ShouldBeTrue(r.shouldPullMsg())
	})

	convey.Convey("when cmdChanged is false and notMeetStart is zero, then shouldPullMsg returns false", t, func() {
		r := retryState{}
		convey.ShouldBeFalse(r.shouldPullMsg())
	})

	convey.Convey("when cmdChanged is false and within retry interval, then shouldPullMsg returns false", t, func() {
		r := retryState{notMeetStart: time.Now().Add(-3 * time.Second)}
		convey.ShouldBeFalse(r.shouldPullMsg())
	})

	convey.Convey("when cmdChanged is false and retry interval exceeded, then shouldPullMsg returns true", t, func() {
		r := retryState{notMeetStart: time.Now().Add(-6 * time.Second)}
		convey.ShouldBeTrue(r.shouldPullMsg())
	})
}

func TestRetryStateMarkPulled(t *testing.T) {
	convey.Convey("markPulled resets notMeetStart to now and cmdChanged to false", t, func() {
		r := retryState{
			notMeetStart: time.Now().Add(-6 * time.Second),
			cmdChanged:   true,
		}
		beforeMark := time.Now()
		r.markPulled()
		convey.ShouldBeFalse(r.notMeetStart.Before(beforeMark.Add(-1 * time.Millisecond)))
		convey.ShouldBeFalse(r.cmdChanged)
	})
}

func TestPullMsg(t *testing.T) {
	convey.Convey("when shouldPullMsg is true via cmdChanged, then returns pullMsg and marks pulled", t, func() {
		plugin := getProfilingPlugin()
		plugin.pullMsg = []infrastructure.Msg{{Receiver: []string{worker0Name}}}
		plugin.retry = retryState{cmdChanged: true}
		msg, err := plugin.PullMsg()
		convey.ShouldBeNil(err)
		convey.ShouldEqual(len(msg), 1)
		convey.ShouldBeFalse(plugin.retry.cmdChanged)
		convey.ShouldBeFalse(plugin.retry.notMeetStart.IsZero())
	})

	convey.Convey("when shouldPullMsg is true via exceedLimit, then returns pullMsg and resets timer", t, func() {
		plugin := getProfilingPlugin()
		plugin.pullMsg = []infrastructure.Msg{{Receiver: []string{worker0Name}}}
		plugin.retry = retryState{notMeetStart: time.Now().Add(-6 * time.Second)}
		msg, err := plugin.PullMsg()
		convey.ShouldBeNil(err)
		convey.ShouldEqual(len(msg), 1)
	})

	convey.Convey("when status meets cmd, then clears pullMsg and returns nil", t, func() {
		plugin := getProfilingPlugin()
		plugin.pullMsg = []infrastructure.Msg{{Receiver: []string{worker0Name}}}
		plugin.retry = retryState{}
		plugin.workerStatus = workerExecStatus{
			cmd:                constant.ProfilingDomainCmd{DefaultDomainAble: false, CommDomainAble: false},
			defaultDomainState: constant.ProfilingWorkerClosedState,
			commDomainState:    constant.ProfilingWorkerClosedState,
		}
		msg, err := plugin.PullMsg()
		convey.ShouldBeNil(err)
		convey.ShouldBeNil(msg)
		convey.ShouldEqual(len(plugin.pullMsg), 0)
	})

	convey.Convey("when status not meet cmd and shouldPullMsg is false, then returns nil without clearing", t, func() {
		plugin := getProfilingPlugin()
		plugin.pullMsg = []infrastructure.Msg{{Receiver: []string{worker0Name}}}
		plugin.retry = retryState{notMeetStart: time.Now().Add(-2 * time.Second)}
		plugin.workerStatus = workerExecStatus{
			cmd:                constant.ProfilingDomainCmd{DefaultDomainAble: true, CommDomainAble: true},
			defaultDomainState: constant.ProfilingWorkerWaitOpenState,
			commDomainState:    constant.ProfilingWorkerWaitOpenState,
		}
		msg, err := plugin.PullMsg()
		convey.ShouldBeNil(err)
		convey.ShouldBeNil(msg)
		convey.ShouldEqual(len(plugin.pullMsg), 1)
	})
}

func TestPredicateRetryExceedLimit(t *testing.T) {
	convey.Convey("when getProfilingCmd fails but retry exceeds interval, then should still candidate", t, func() {
		plugin := getProfilingPlugin()
		plugin.retry = retryState{notMeetStart: time.Now().Add(-6 * time.Second)}
		snapshot := getDemoSnapshot()
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		patches.ApplyPrivateMethod(plugin, "getProfilingCmd",
			func(*PfPlugin, storage.SnapShot) (constant.ProfilingDomainCmd, error) {
				return constant.ProfilingDomainCmd{}, fmt.Errorf("cmd unchanged")
			}).ApplyPrivateMethod(plugin, "handleProfilingResult",
			func(*PfPlugin, storage.SnapShot) {})
		predicate, err := plugin.Predicate(snapshot)
		convey.ShouldBeNil(err)
		convey.ShouldEqual(predicate.CandidateStatus, constant.CandidateStatus)
	})
}
