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

// Package recoveplugin for taskd manager plugin
package recoveplugin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

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
	agent0Name = "0"
	agent1Name = "1"
	testLen    = 2
)

func getRecoverPlugin() *RecoverPlugin {
	return NewRecoverPlugin().(*RecoverPlugin)
}

func getDemoSnapshot() storage.SnapShot {
	return storage.SnapShot{
		ClusterInfos: &storage.ClusterInfos{
			Clusters: map[string]*storage.ClusterInfo{
				constant.ClusterDRank: {
					Command: map[string]string{},
				},
			},
		},
	}
}

func getSnapshotWithClusterInfo() storage.SnapShot {
	snapshot := getDemoSnapshot()
	faultRanks := map[int]int{0: 0, 1: 1}
	faultRanksStr, err := json.Marshal(faultRanks)
	if err != nil {
		hwlog.RunLog.Error("test marshal faultRanks failed")
	}
	nodeIds := []string{agent0Name}
	nodeIdsStr, err := json.Marshal(nodeIds)
	if err != nil {
		hwlog.RunLog.Error("test marshal faultRanks failed")
	}
	actions := []string{clusterdconstant.ChangeStrategyAction}
	actionsStr, err := json.Marshal(actions)
	if err != nil {
		hwlog.RunLog.Error("test marshal actions failed")
	}
	snapshot.ClusterInfos.Clusters[constant.ClusterDRank].Command[constant.FaultRanks] = string(faultRanksStr)
	snapshot.ClusterInfos.Clusters[constant.ClusterDRank].Command[constant.NodeRankIds] = string(nodeIdsStr)
	snapshot.ClusterInfos.Clusters[constant.ClusterDRank].Command[constant.Actions] = string(actionsStr)
	snapshot.ClusterInfos.Clusters[constant.ClusterDRank].Command[constant.Uuid] = "test-uuid"
	snapshot.ClusterInfos.Clusters[constant.ClusterDRank].Command[constant.ChangeStrategy] = clusterdconstant.ProcessRecoverStrategyName
	return snapshot
}

func getSnapshotWithSaveAndExit() storage.SnapShot {
	snapshot := getDemoSnapshot()
	nodeIds := []string{agent0Name}
	nodeIdsStr, err := json.Marshal(nodeIds)
	if err != nil {
		hwlog.RunLog.Error("test marshal nodeIds failed")
	}
	snapshot.ClusterInfos.Clusters[constant.ClusterDRank].Command[constant.SignalType] = clusterdconstant.SaveAndExitSignalType
	snapshot.ClusterInfos.Clusters[constant.ClusterDRank].Command[constant.Uuid] = "save-exit-uuid"
	snapshot.ClusterInfos.Clusters[constant.ClusterDRank].Command[constant.NodeRankIds] = string(nodeIdsStr)
	return snapshot
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

func TestNewRecoverPlugin(t *testing.T) {
	convey.Convey("new recover plugin should not nil", t, func() {
		plugin := NewRecoverPlugin()
		convey.ShouldNotBeNil(plugin)
		recoverPlugin, ok := plugin.(*RecoverPlugin)
		convey.ShouldBeTrue(ok)
		convey.ShouldEqual(len(recoverPlugin.pullMsgs), 0)
		convey.ShouldEqual(recoverPlugin.processStatus, "")
	})
}

func TestName(t *testing.T) {
	convey.Convey("plugin name should be RecoverPluginName", t, func() {
		plugin := getRecoverPlugin()
		convey.ShouldEqual(plugin.Name(), constant.RecoverPluginName)
	})
}

func TestResetPluginInfo(t *testing.T) {
	convey.Convey("reset plugin info should set all fields to default", t, func() {
		plugin := getRecoverPlugin()
		plugin.processStatus = "processing"
		plugin.faultNode = []string{agent0Name}
		plugin.faultRanks = []string{"0"}
		plugin.actions = []string{constant.RestartController}
		plugin.msgSend = true
		plugin.recoverInPlace = true
		plugin.saveAndExit = true
		plugin.resetPluginInfo()
		convey.ShouldEqual(plugin.processStatus, "")
		convey.ShouldEqual(len(plugin.faultNode), 0)
		convey.ShouldEqual(len(plugin.faultRanks), 0)
		convey.ShouldEqual(len(plugin.actions), 0)
		convey.ShouldBeFalse(plugin.msgSend)
		convey.ShouldBeFalse(plugin.recoverInPlace)
		convey.ShouldBeFalse(plugin.saveAndExit)
	})
}

func TestGetClusterInfo(t *testing.T) {
	convey.Convey("get cluster info should parse data correctly", t, func() {
		plugin := getRecoverPlugin()
		snapshot := getSnapshotWithClusterInfo()
		err := plugin.getClusterInfo(snapshot)
		convey.ShouldBeNil(err)
		convey.ShouldEqual(len(plugin.faultRanks), testLen)
		convey.ShouldEqual(len(plugin.faultNode), 1)
		convey.ShouldEqual(len(plugin.actions), 1)
		convey.ShouldEqual(plugin.uuid, "test-uuid")
	})

	convey.Convey("get cluster info should return error when cluster not found", t, func() {
		plugin := getRecoverPlugin()
		snapshot := getDemoSnapshot()
		delete(snapshot.ClusterInfos.Clusters, constant.ClusterDRank)
		err := plugin.getClusterInfo(snapshot)
		convey.ShouldNotBeNil(err)
		convey.ShouldEqual(err.Error(), "cluster info not found")
	})

	convey.Convey("get cluster info should handle recover in place correctly", t, func() {
		plugin := getRecoverPlugin()
		snapshot := getSnapshotWithClusterInfo()
		snapshot.ClusterInfos.Clusters[constant.ClusterDRank].Command[constant.ExtraParams] = clusterdconstant.ProcessRecoverInPlaceStrategyName
		err := plugin.getClusterInfo(snapshot)
		convey.ShouldBeNil(err)
		convey.ShouldBeTrue(plugin.recoverInPlace)
	})
}

func TestCheckSaveAndExit(t *testing.T) {
	convey.Convey("check save and exit should return true when signal type is save and exit", t, func() {
		plugin := getRecoverPlugin()
		snapshot := getSnapshotWithSaveAndExit()
		result := plugin.checkSaveAndExit(snapshot)
		convey.ShouldBeTrue(result)
		convey.ShouldBeTrue(plugin.saveAndExit)
		convey.ShouldEqual(plugin.uuid, "save-exit-uuid")
		convey.ShouldEqual(len(plugin.faultNode), 1)
	})

	convey.Convey("check save and exit should return false when cluster info not found", t, func() {
		plugin := getRecoverPlugin()
		snapshot := getDemoSnapshot()
		delete(snapshot.ClusterInfos.Clusters, constant.ClusterDRank)
		result := plugin.checkSaveAndExit(snapshot)
		convey.ShouldBeFalse(result)
	})

	convey.Convey("check save and exit should return false when uuid is same", t, func() {
		plugin := getRecoverPlugin()
		plugin.uuid = "save-exit-uuid"
		snapshot := getSnapshotWithSaveAndExit()
		result := plugin.checkSaveAndExit(snapshot)
		convey.ShouldBeFalse(result)
	})
}

func TestBuildControllerMessage(t *testing.T) {
	convey.Convey("build controller message should add message to pullMsgs", t, func() {
		plugin := getRecoverPlugin()
		plugin.recoverStrategy = clusterdconstant.ProcessRecoverStrategyName
		plugin.actions = []string{constant.RestartController}
		plugin.buildControllerMessage()
		convey.ShouldEqual(len(plugin.pullMsgs), 1)
		convey.ShouldEqual(plugin.pullMsgs[0].Receiver[0], constant.ControllerName)
		convey.ShouldEqual(plugin.pullMsgs[0].Body.Code, constant.ProcessManageRecoverSignal)
	})
}

func TestBuildSaveAndExitMessage(t *testing.T) {
	convey.Convey("build save and exit message should add messages to pullMsgs", t, func() {
		plugin := getRecoverPlugin()
		plugin.faultNode = []string{agent0Name}
		plugin.buildSaveAndExitMessage()
		convey.ShouldEqual(len(plugin.pullMsgs), testLen)
		convey.ShouldEqual(plugin.pullMsgs[0].Receiver[0], constant.ControllerName)
		convey.ShouldEqual(plugin.pullMsgs[1].Receiver[0], common.AgentRole+agent0Name)
	})
}

func TestPullMsg(t *testing.T) {
	convey.Convey("pull msg should return all messages and clear pullMsgs", t, func() {
		plugin := getRecoverPlugin()
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

func TestRelease(t *testing.T) {
	convey.Convey("release should return nil", t, func() {
		plugin := getRecoverPlugin()
		err := plugin.Release()
		convey.ShouldBeNil(err)
	})
}

func TestPredicate(t *testing.T) {
	convey.Convey("predicate should return candidate when process status is not empty", t, func() {
		plugin := getRecoverPlugin()
		plugin.processStatus = "processing"
		snapshot := getDemoSnapshot()
		result, err := plugin.Predicate(snapshot)
		convey.ShouldBeNil(err)
		convey.ShouldEqual(result.CandidateStatus, constant.CandidateStatus)
		convey.ShouldNotBeNil(result.PredicateStream)
	})

	convey.Convey("predicate should return candidate when save and exit", t, func() {
		plugin := getRecoverPlugin()
		snapshot := getSnapshotWithSaveAndExit()
		result, err := plugin.Predicate(snapshot)
		convey.ShouldBeNil(err)
		convey.ShouldEqual(result.CandidateStatus, constant.CandidateStatus)
		convey.ShouldBeTrue(plugin.saveAndExit)
	})

	convey.Convey("predicate should return unselect when strategy not in recover/retry/dump/continue", t, func() {
		plugin := getRecoverPlugin()
		snapshot := getDemoSnapshot()
		snapshot.ClusterInfos.Clusters[constant.ClusterDRank].Command[constant.ChangeStrategy] = "unknown"
		result, err := plugin.Predicate(snapshot)
		convey.ShouldBeNil(err)
		convey.ShouldEqual(result.CandidateStatus, constant.UnselectStatus)
	})

	convey.Convey("predicate should return candidate when strategy is recover", t, func() {
		plugin := getRecoverPlugin()
		snapshot := getSnapshotWithClusterInfo()
		result, err := plugin.Predicate(snapshot)
		convey.ShouldBeNil(err)
		convey.ShouldEqual(result.CandidateStatus, constant.CandidateStatus)
		convey.ShouldNotBeNil(result.PredicateStream)
	})

	convey.Convey("predicate should return unselect when getClusterInfo fails", t, func() {
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		patches.ApplyPrivateMethod(getRecoverPlugin(), "getClusterInfo",
			func(*RecoverPlugin, storage.SnapShot) error {
				return errors.New("test error")
			})
		plugin := getRecoverPlugin()
		snapshot := getSnapshotWithClusterInfo()
		result, err := plugin.Predicate(snapshot)
		convey.ShouldBeNil(err)
		convey.ShouldEqual(result.CandidateStatus, constant.UnselectStatus)
	})
}

func TestHandle(t *testing.T) {
	convey.Convey("handle should process save and exit correctly", t, func() {
		plugin := getRecoverPlugin()
		plugin.saveAndExit = true
		plugin.faultNode = []string{agent0Name}
		result, err := plugin.Handle()
		convey.ShouldBeNil(err)
		convey.ShouldEqual(result.Stage, constant.HandleStageFinal)
		convey.ShouldBeFalse(plugin.saveAndExit)
		convey.ShouldBeFalse(plugin.doAction)
	})

	convey.Convey("handle should process dump strategy correctly", t, func() {
		plugin := getRecoverPlugin()
		plugin.recoverStrategy = clusterdconstant.ProcessDumpStrategyName
		result, err := plugin.Handle()
		convey.ShouldBeNil(err)
		convey.ShouldEqual(result.Stage, constant.HandleStageFinal)
		convey.ShouldBeFalse(plugin.doAction)
	})

	convey.Convey("handle should process msgSend true correctly", t, func() {
		plugin := getRecoverPlugin()
		plugin.msgSend = true
		result, err := plugin.Handle()
		convey.ShouldBeNil(err)
		convey.ShouldEqual(result.Stage, constant.HandleStageFinal)
		convey.ShouldBeFalse(plugin.doAction)
	})

	convey.Convey("handle should process recover in place correctly", t, func() {
		plugin := getRecoverPlugin()
		plugin.recoverInPlace = true
		plugin.faultNode = []string{agent0Name}
		plugin.faultRanks = []string{"0"}
		result, err := plugin.Handle()
		convey.ShouldBeNil(err)
		convey.ShouldEqual(result.Stage, constant.HandleStageFinal)
		convey.ShouldBeFalse(plugin.doAction)
		convey.ShouldEqual(len(plugin.pullMsgs), 1)
	})

	convey.Convey("handle should process normal case correctly", t, func() {
		plugin := getRecoverPlugin()
		plugin.faultNode = []string{agent0Name}
		result, err := plugin.Handle()
		convey.ShouldBeNil(err)
		convey.ShouldEqual(result.Stage, constant.HandleStageProcess)
		convey.ShouldBeTrue(plugin.msgSend)
		convey.ShouldEqual(len(plugin.pullMsgs), 1)
	})
}
