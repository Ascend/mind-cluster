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

// Package jobrescheduling for taskd manager plugin
package jobrescheduling

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/common-utils/hwlog"
	clusterdconstant "clusterd/pkg/common/constant"
	"taskd/common/constant"
	"taskd/framework_backend/manager/infrastructure"
	"taskd/framework_backend/manager/infrastructure/storage"
	"taskd/toolkit_backend/net/common"
)

const (
	agent0Name   = "agent0"
	agent1Name   = "agent1"
	threeMinutes = 3 * time.Minute
	sixMinutes   = 6 * time.Minute
)

func getJobReschedulingPlugin() *JobReschedulingPlugin {
	return NewJobReschedulingPlugin().(*JobReschedulingPlugin)
}

func getDemoSnapshot() storage.SnapShot {
	return storage.SnapShot{
		AgentInfos: &storage.AgentInfos{
			Agents: map[string]*storage.AgentInfo{
				common.AgentRole + "0": {
					Status: map[string]string{},
				},
			},
		},
		ClusterInfos: &storage.ClusterInfos{
			Clusters: map[string]*storage.ClusterInfo{
				constant.ClusterDRank: {
					Command: map[string]string{},
				},
			},
		},
		MgrInfos: &storage.MgrInfo{
			Status: map[string]string{},
		},
	}
}

func getSnapshotWithAgent0Fault() storage.SnapShot {
	snapshot := getDemoSnapshot()
	snapshot.AgentInfos.Agents[common.AgentRole+"0"].Status[constant.ReportFaultRank] = "0"
	return snapshot
}

func getSnapshotWithKillMasterSignal() storage.SnapShot {
	snapshot := getDemoSnapshot()
	snapshot.MgrInfos.Status[constant.FaultRecover] = "true"
	snapshot.ClusterInfos.Clusters[constant.ClusterDRank].Command[constant.SignalType] = clusterdconstant.KillMasterSignalType
	return snapshot
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

func TestNewJobReschedulingPlugin(t *testing.T) {
	convey.Convey("new job rescheduling plugin should not nil", t, func() {
		plugin := NewJobReschedulingPlugin()
		convey.ShouldNotBeNil(plugin)
	})
}

func TestName(t *testing.T) {
	convey.Convey("plugin name should be jobReschedulingPlugin", t, func() {
		plugin := getJobReschedulingPlugin()
		convey.ShouldEqual(plugin.Name(), constant.JobReschedulingPluginName)
	})
}

func TestResetPluginInfo(t *testing.T) {
	convey.Convey("reset plugin info should set all fields to default", t, func() {
		plugin := getJobReschedulingPlugin()
		plugin.faultOccur = true
		plugin.processStatus = "processing"
		plugin.killMaster = true
		plugin.resetPluginInfo()
		convey.ShouldBeFalse(plugin.faultOccur)
		convey.ShouldEqual(plugin.processStatus, "")
		convey.ShouldBeFalse(plugin.killMaster)
	})
}

func TestUpdatePluginInfo(t *testing.T) {
	convey.Convey("update plugin info when agent0 has fault", t, func() {
		plugin := getJobReschedulingPlugin()
		snapshot := getSnapshotWithAgent0Fault()
		plugin.updatePluginInfo(snapshot)
		convey.ShouldBeTrue(plugin.killMaster)
	})

	convey.Convey("update plugin info when has kill master signal", t, func() {
		plugin := getJobReschedulingPlugin()
		snapshot := getSnapshotWithKillMasterSignal()
		plugin.updatePluginInfo(snapshot)
		convey.ShouldBeTrue(plugin.killMaster)
	})
}

func TestCheckKillMaster(t *testing.T) {
	convey.Convey("check kill master should return true when mgr info and fault recover exist", t, func() {
		plugin := getJobReschedulingPlugin()
		snapshot := getSnapshotWithKillMasterSignal()
		result := plugin.checkKillMaster(snapshot)
		convey.ShouldBeTrue(result)
		convey.ShouldBeTrue(plugin.killMaster)
	})

	convey.Convey("check kill master should return false when mgr info is nil", t, func() {
		plugin := getJobReschedulingPlugin()
		snapshot := getDemoSnapshot()
		snapshot.MgrInfos = nil
		result := plugin.checkKillMaster(snapshot)
		convey.ShouldBeFalse(result)
	})

	convey.Convey("check kill master should return false when fault recover is empty", t, func() {
		plugin := getJobReschedulingPlugin()
		snapshot := getDemoSnapshot()
		snapshot.MgrInfos.Status[constant.FaultRecover] = ""
		result := plugin.checkKillMaster(snapshot)
		convey.ShouldBeFalse(result)
	})
}

func TestCheckRank0Fault(t *testing.T) {
	convey.Convey("check rank0 fault should set kill master when agent0 has fault", t, func() {
		plugin := getJobReschedulingPlugin()
		snapshot := getSnapshotWithAgent0Fault()
		plugin.checkRank0Fault(snapshot)
		convey.ShouldBeTrue(plugin.killMaster)
	})

	convey.Convey("check rank0 fault should not set kill master when agent0 not exist", t, func() {
		plugin := getJobReschedulingPlugin()
		snapshot := getDemoSnapshot()
		delete(snapshot.AgentInfos.Agents, common.AgentRole+"0")
		plugin.checkRank0Fault(snapshot)
		convey.ShouldBeFalse(plugin.killMaster)
	})

	convey.Convey("check rank0 fault should return when agent info is nil", t, func() {
		plugin := getJobReschedulingPlugin()
		snapshot := getDemoSnapshot()
		snapshot.AgentInfos = nil
		plugin.checkRank0Fault(snapshot)
		convey.ShouldBeFalse(plugin.killMaster)
	})
}

func TestHandle(t *testing.T) {
	convey.Convey("handle when kill master is true should return destroy controller message", t, func() {
		plugin := getJobReschedulingPlugin()
		plugin.killMaster = true
		handleResult, err := plugin.Handle()
		convey.ShouldBeNil(err)
		convey.ShouldEqual(handleResult.Stage, constant.HandleStageFinal)
		convey.ShouldBeFalse(plugin.killMaster)
		convey.ShouldBeFalse(plugin.faultOccur)
		convey.ShouldEqual(plugin.processStatus, "")
	})

	convey.Convey("handle when fault occur is false should reset plugin info", t, func() {
		plugin := getJobReschedulingPlugin()
		plugin.faultOccur = false
		handleResult, err := plugin.Handle()
		convey.ShouldBeNil(err)
		convey.ShouldEqual(handleResult.Stage, constant.HandleStageFinal)
		convey.ShouldBeFalse(plugin.killMaster)
		convey.ShouldBeFalse(plugin.faultOccur)
		convey.ShouldEqual(plugin.processStatus, "")
	})

	convey.Convey("handle when fault occur is true should return exit agent message", t, func() {
		plugin := getJobReschedulingPlugin()
		plugin.faultOccur = true
		handleResult, err := plugin.Handle()
		convey.ShouldBeNil(err)
		convey.ShouldEqual(handleResult.Stage, constant.HandleStageFinal)
		convey.ShouldEqual(plugin.processStatus, "")
	})
}

func TestPullMsg(t *testing.T) {
	convey.Convey("pull msg should return all messages and clear pullMsgs", t, func() {
		plugin := getJobReschedulingPlugin()
		msg := infrastructure.Msg{
			Receiver: []string{"agent0"},
			Body: storage.MsgBody{
				MsgType: constant.Action,
				Code:    constant.ExitAgentCode,
			},
		}
		plugin.pullMsgs = append(plugin.pullMsgs, msg)
		msgs, err := plugin.PullMsg()
		convey.ShouldBeNil(err)
		convey.ShouldEqual(len(msgs), 1)
		convey.ShouldEqual(msgs[0].Receiver[0], "agent0")
		convey.ShouldEqual(len(plugin.pullMsgs), 0)
	})
}

func TestPredicate(t *testing.T) {
	convey.Convey("predicate when process status is not empty should return candidate", t, func() {
		plugin := getJobReschedulingPlugin()
		plugin.processStatus = "processing"
		snapshot := getDemoSnapshot()
		predicateResult, err := plugin.Predicate(snapshot)
		convey.ShouldBeNil(err)
		convey.ShouldEqual(predicateResult.CandidateStatus, constant.CandidateStatus)
		convey.ShouldNotBeNil(predicateResult.PredicateStream)
	})

	convey.Convey("predicate when check kill master returns true and kill master is true should return candidate", t, func() {
		plugin := getJobReschedulingPlugin()
		snapshot := getSnapshotWithKillMasterSignal()
		predicateResult, err := plugin.Predicate(snapshot)
		convey.ShouldBeNil(err)
		convey.ShouldEqual(predicateResult.CandidateStatus, constant.CandidateStatus)
		convey.ShouldNotBeNil(predicateResult.PredicateStream)
	})

	convey.Convey("predicate when agent0 has fault should return candidate", t, func() {
		plugin := getJobReschedulingPlugin()
		snapshot := getSnapshotWithAgent0Fault()
		predicateResult, err := plugin.Predicate(snapshot)
		convey.ShouldBeNil(err)
		convey.ShouldBeTrue(plugin.faultOccur)
		convey.ShouldEqual(predicateResult.CandidateStatus, constant.CandidateStatus)
		convey.ShouldNotBeNil(predicateResult.PredicateStream)
	})

	convey.Convey("predicate when no fault occur should return unselect", t, func() {
		predicatePatches := gomonkey.NewPatches()
		defer predicatePatches.Reset()
		predicatePatches.ApplyPrivateMethod(getJobReschedulingPlugin(), "checkKillMaster",
			func(*JobReschedulingPlugin, storage.SnapShot) bool {
				return false
			})

		plugin := getJobReschedulingPlugin()
		snapshot := getDemoSnapshot()
		predicateResult, err := plugin.Predicate(snapshot)
		convey.ShouldBeNil(err)
		convey.ShouldEqual(predicateResult.CandidateStatus, constant.UnselectStatus)
		convey.ShouldBeNil(predicateResult.PredicateStream)
	})
}

func TestRelease(t *testing.T) {
	convey.Convey("release should return nil", t, func() {
		plugin := getJobReschedulingPlugin()
		err := plugin.Release()
		convey.ShouldBeNil(err)
	})
}

func TestIsRank0FaultTimeout(t *testing.T) {
	tests := []struct {
		name       string
		agent0Info *storage.AgentInfo
		want       bool
	}{
		{
			name: "report_fault_time_not_exist",
			agent0Info: &storage.AgentInfo{
				Status: map[string]string{},
			},
			want: false,
		},
		{
			name: "report_fault_time_parse_error",
			agent0Info: &storage.AgentInfo{
				Status: map[string]string{
					constant.ReportFaultTime: "invalid_time",
				},
			},
			want: false,
		},
		{
			name: "fault_time_within_timeout",
			agent0Info: &storage.AgentInfo{
				Status: map[string]string{
					constant.ReportFaultTime: strconv.FormatInt(time.Now().Add(-threeMinutes).Unix(), constant.Base),
				},
			},
			want: true,
		},
		{
			name: "fault_time_exceed_timeout",
			agent0Info: &storage.AgentInfo{
				Status: map[string]string{
					constant.ReportFaultTime: strconv.FormatInt(time.Now().Add(-sixMinutes).Unix(), constant.Base),
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := skipRank0FaultForNow(tt.agent0Info); got != tt.want {
				t.Errorf("isRank0FaultTimeout() = %v, want %v", got, tt.want)
			}
		})
	}
}
