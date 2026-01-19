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

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/common-utils/hwlog"
	"taskd/common/constant"
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
	plugin := getProfilingPlugin()
	cmd := constant.ProfilingDomainCmd{
		DefaultDomainAble: true,
		CommDomainAble:    true,
	}
	convey.Convey("when change cmd, then pullMsg should not be empty", t, func() {
		plugin.changeCmd(cmd)
		convey.ShouldBeTrue(len(plugin.pullMsg), 1)
	})
	convey.Convey("when pull, then pullMsg should be empty", t, func() {
		msg, err := plugin.PullMsg()
		convey.ShouldBeNil(err)
		convey.ShouldBeTrue(len(msg), 1)
		convey.ShouldBeTrue(len(plugin.pullMsg), 0)
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
		plugin.report[worker0Name] = constant.ProfilingResult{
			DefaultDomain: constant.ProfilingOnStatus,
			CommDomain:    constant.ProfilingExpStatus,
		}
		handle, err := plugin.Handle()
		convey.ShouldBeNil(err)
		convey.ShouldEqual(handle.Stage, constant.HandleStageFinal)
	})
}

func TestGetProfilingResult(t *testing.T) {
	convey.Convey("when snapshot has new result, then result is not empty", t, func() {
		snapshot := getDemoSnapshot()
		plugin := getProfilingPlugin()
		result, err := plugin.getProfilingResult(snapshot)
		convey.ShouldBeNil(err)
		convey.ShouldEqual(len(result), 1)
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
			}).ApplyPrivateMethod(plugin, "getProfilingResult",
			func(*PfPlugin, storage.SnapShot) (map[string]constant.ProfilingResult, error) {
				return nil, fmt.Errorf("error")
			})
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
