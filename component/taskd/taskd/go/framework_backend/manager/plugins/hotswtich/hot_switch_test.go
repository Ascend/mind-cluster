/* Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

// Package hotswitch for taskd manager plugin
package hotswitch

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
	"taskd/common/utils"
	"taskd/framework_backend/manager/infrastructure/storage"
)

const (
	testUuid           = "test-uuid-123"
	testActions        = "test-actions"
	testFaultRanksJSON = `{"0":1,"1":2}`
	testInvalidJSON    = "invalid-json"
	emptyString        = ""
	zeroLength         = 0
	num1               = 1
	num2               = 2
)

func TestMain(m *testing.M) {
	if err := initLog(); err != nil {
		return
	}
	code := m.Run()
	if code != 0 {
		fmt.Printf("exit_code = %v\n", code)
	}
}

func initLog() error {
	logConfig := &hwlog.LogConfig{
		OnlyToStdout: true,
	}
	if err := hwlog.InitRunLogger(logConfig, context.Background()); err != nil {
		fmt.Printf("init hwlog failed, %v\n", err)
		return errors.New("init hwlog failed")
	}
	return nil
}

func createEmptySnapshot() storage.SnapShot {
	return storage.SnapShot{
		ClusterInfos: &storage.ClusterInfos{
			Clusters: make(map[string]*storage.ClusterInfo),
		},
	}
}

func createSnapshotWithClusterInfo(signalType, actions, strategy, uuid, faultRanks string) storage.SnapShot {
	snapshot := storage.SnapShot{
		ClusterInfos: &storage.ClusterInfos{
			Clusters: map[string]*storage.ClusterInfo{
				constant.ClusterDRank: {
					Command: make(map[string]string),
				},
			},
		},
	}
	if signalType != emptyString {
		snapshot.ClusterInfos.Clusters[constant.ClusterDRank].Command[constant.SignalType] = signalType
	}
	if actions != emptyString {
		snapshot.ClusterInfos.Clusters[constant.ClusterDRank].Command[constant.Actions] = actions
	}
	if strategy != emptyString {
		snapshot.ClusterInfos.Clusters[constant.ClusterDRank].Command[constant.ChangeStrategy] = strategy
	}
	if uuid != emptyString {
		snapshot.ClusterInfos.Clusters[constant.ClusterDRank].Command[constant.Uuid] = uuid
	}
	if faultRanks != emptyString {
		snapshot.ClusterInfos.Clusters[constant.ClusterDRank].Command[constant.FaultRanks] = faultRanks
	}
	return snapshot
}

func getHotSwitchPlugin(plugin interface{}) (*HotSwitchPlugin, error) {
	hotSwitchPlugin, ok := plugin.(*HotSwitchPlugin)
	if !ok {
		return nil, fmt.Errorf("type assertion failed: expected *HotSwitchPlugin")
	}
	return hotSwitchPlugin, nil
}

func TestNewHotSwitchPluginShouldReturnInitializedPlugin(t *testing.T) {
	convey.Convey("should return plugin with empty fields when created", t, func() {
		plugin := NewHotSwitchPlugin()
		convey.So(plugin, convey.ShouldNotBeNil)
		hotSwitchPlugin, err := getHotSwitchPlugin(plugin)
		convey.So(err, convey.ShouldBeNil)
		convey.So(hotSwitchPlugin.pullMsgs, convey.ShouldNotBeNil)
		convey.So(len(hotSwitchPlugin.pullMsgs), convey.ShouldEqual, zeroLength)
		convey.So(hotSwitchPlugin.changeStrategy, convey.ShouldEqual, emptyString)
		convey.So(hotSwitchPlugin.faultRanks, convey.ShouldNotBeNil)
		convey.So(len(hotSwitchPlugin.faultRanks), convey.ShouldEqual, zeroLength)
		convey.So(hotSwitchPlugin.actions, convey.ShouldEqual, emptyString)
		convey.So(hotSwitchPlugin.uuid, convey.ShouldEqual, emptyString)
	})
}

func TestNameShouldReturnHotSwitchPluginName(t *testing.T) {
	convey.Convey("should return HotSwitchPluginName when called", t, func() {
		plugin := NewHotSwitchPlugin()
		hotSwitchPlugin, err := getHotSwitchPlugin(plugin)
		convey.So(err, convey.ShouldBeNil)
		convey.So(hotSwitchPlugin.Name(), convey.ShouldEqual, constant.HotSwitchPluginName)
	})
}

func TestPredicateShouldReturnUnselectWhenClusterInfoNotFound(t *testing.T) {
	convey.Convey("should return unselect status when cluster info not found", t, func() {
		plugin := NewHotSwitchPlugin()
		hotSwitchPlugin, err := getHotSwitchPlugin(plugin)
		convey.So(err, convey.ShouldBeNil)
		snapshot := createEmptySnapshot()
		result, err := hotSwitchPlugin.Predicate(snapshot)
		convey.So(err, convey.ShouldBeNil)
		convey.So(result.CandidateStatus, convey.ShouldEqual, constant.UnselectStatus)
		convey.So(result.PredicateStream, convey.ShouldBeNil)
	})
}

func TestPredicateShouldReturnUnselectWhenSignalTypeAndStrategyNotMatch(t *testing.T) {
	cases := []struct {
		name       string
		signalType string
		strategy   string
	}{
		{"should return unselect when signal type not match and strategy not match",
			"other-signal", "other-strategy"},
		{"should return unselect when signal type is empty and strategy not match",
			emptyString, "other-strategy"},
	}
	for _, tc := range cases {
		convey.Convey(tc.name, t, func() {
			plugin := NewHotSwitchPlugin()
			hotSwitchPlugin, err := getHotSwitchPlugin(plugin)
			convey.So(err, convey.ShouldBeNil)
			snapshot := createSnapshotWithClusterInfo(tc.signalType, testActions, tc.strategy, testUuid, testFaultRanksJSON)
			result, err := hotSwitchPlugin.Predicate(snapshot)
			convey.So(err, convey.ShouldBeNil)
			convey.So(result.CandidateStatus, convey.ShouldEqual, constant.UnselectStatus)
		})
	}
}

func TestPredicateShouldReturnUnselectWhenSignalTypeAndActionsNotChanged(t *testing.T) {
	convey.Convey("should return unselect when signal type and actions not changed", t, func() {
		plugin := NewHotSwitchPlugin()
		hotSwitchPlugin, err := getHotSwitchPlugin(plugin)
		convey.So(err, convey.ShouldBeNil)
		hotSwitchPlugin.signalType = clusterdconstant.HotSwitchSignalType
		hotSwitchPlugin.actions = testActions
		snapshot := createSnapshotWithClusterInfo(clusterdconstant.HotSwitchSignalType, testActions,
			clusterdconstant.ProcessMigration, testUuid, testFaultRanksJSON)
		result, err := hotSwitchPlugin.Predicate(snapshot)
		convey.So(err, convey.ShouldBeNil)
		convey.So(result.CandidateStatus, convey.ShouldEqual, constant.UnselectStatus)
	})
}

func TestPredicateShouldReturnUnselectWhenGetClusterInfoFails(t *testing.T) {
	convey.Convey("should return unselect when getClusterInfo fails", t, func() {
		plugin := NewHotSwitchPlugin()
		hotSwitchPlugin, err := getHotSwitchPlugin(plugin)
		convey.So(err, convey.ShouldBeNil)
		snapshot := createSnapshotWithClusterInfo(clusterdconstant.HotSwitchSignalType, testActions,
			clusterdconstant.ProcessMigration, testUuid, testInvalidJSON)
		patch := gomonkey.ApplyFuncReturn(utils.StringToObj[map[int]int], map[int]int{},
			errors.New("parse error"))
		defer patch.Reset()
		result, err := hotSwitchPlugin.Predicate(snapshot)
		convey.So(err, convey.ShouldBeNil)
		convey.So(result.CandidateStatus, convey.ShouldEqual, constant.UnselectStatus)
		convey.So(hotSwitchPlugin.faultRanks, convey.ShouldNotBeNil)
		convey.So(len(hotSwitchPlugin.faultRanks), convey.ShouldEqual, zeroLength)
	})
}

func TestPredicateShouldReturnCandidateWhenHotSwitchSignalType(t *testing.T) {
	convey.Convey("should return candidate when signal type is HotSwitchSignalType", t, func() {
		plugin := NewHotSwitchPlugin()
		hotSwitchPlugin, err := getHotSwitchPlugin(plugin)
		convey.So(err, convey.ShouldBeNil)
		snapshot := createSnapshotWithClusterInfo(clusterdconstant.HotSwitchSignalType, testActions,
			clusterdconstant.ProcessMigration, testUuid, testFaultRanksJSON)
		result, err := hotSwitchPlugin.Predicate(snapshot)
		convey.So(err, convey.ShouldBeNil)
		convey.So(result.CandidateStatus, convey.ShouldEqual, constant.CandidateStatus)
		convey.So(result.PredicateStream, convey.ShouldNotBeNil)
		convey.So(result.PredicateStream[constant.ResumeTrainingAfterFaultStream], convey.ShouldEqual, emptyString)
		convey.So(hotSwitchPlugin.changeStrategy, convey.ShouldEqual, clusterdconstant.ProcessMigration)
		convey.So(hotSwitchPlugin.signalType, convey.ShouldEqual, clusterdconstant.HotSwitchSignalType)
	})
}

func TestPredicateShouldReturnCandidateWhenProcessMigrationStrategy(t *testing.T) {
	convey.Convey("should return candidate when strategy is ProcessMigration", t, func() {
		plugin := NewHotSwitchPlugin()
		hotSwitchPlugin, err := getHotSwitchPlugin(plugin)
		convey.So(err, convey.ShouldBeNil)
		snapshot := createSnapshotWithClusterInfo(emptyString, testActions,
			clusterdconstant.ProcessMigration, testUuid, testFaultRanksJSON)
		result, err := hotSwitchPlugin.Predicate(snapshot)
		convey.So(err, convey.ShouldBeNil)
		convey.So(result.CandidateStatus, convey.ShouldEqual, constant.CandidateStatus)
		convey.So(result.PredicateStream, convey.ShouldNotBeNil)
	})
}

func TestHandleShouldBuildMessageWhenHotSwitchSignalType(t *testing.T) {
	convey.Convey("should build controller message when signal type is HotSwitchSignalType", t, func() {
		plugin := NewHotSwitchPlugin()
		hotSwitchPlugin, err := getHotSwitchPlugin(plugin)
		convey.So(err, convey.ShouldBeNil)
		hotSwitchPlugin.signalType = clusterdconstant.HotSwitchSignalType
		hotSwitchPlugin.actions = testActions
		hotSwitchPlugin.changeStrategy = clusterdconstant.ProcessMigration
		faultRanks := map[int]int{0: 1, 1: 2}
		hotSwitchPlugin.faultRanks = faultRanks
		result, err := hotSwitchPlugin.Handle()
		convey.So(err, convey.ShouldBeNil)
		convey.So(result.Stage, convey.ShouldEqual, constant.HandleStageFinal)
		msgs, _ := hotSwitchPlugin.PullMsg()
		convey.So(len(msgs), convey.ShouldEqual, 1)
		convey.So(msgs[0].Receiver[0], convey.ShouldEqual, constant.ControllerName)
		convey.So(msgs[0].Body.MsgType, convey.ShouldEqual, constant.Action)
		convey.So(msgs[0].Body.Code, convey.ShouldEqual, constant.HotSwitchCode)
	})
}

func TestHandleShouldBuildMessageWhenProcessMigrationStrategy(t *testing.T) {
	convey.Convey("should build controller message when strategy is ProcessMigration", t, func() {
		plugin := NewHotSwitchPlugin()
		hotSwitchPlugin, err := getHotSwitchPlugin(plugin)
		convey.So(err, convey.ShouldBeNil)
		hotSwitchPlugin.changeStrategy = clusterdconstant.ProcessMigration
		hotSwitchPlugin.actions = testActions
		faultRanks := map[int]int{0: 1}
		hotSwitchPlugin.faultRanks = faultRanks
		result, err := hotSwitchPlugin.Handle()
		convey.So(err, convey.ShouldBeNil)
		convey.So(result.Stage, convey.ShouldEqual, constant.HandleStageFinal)
		msgs, _ := hotSwitchPlugin.PullMsg()
		convey.So(len(msgs), convey.ShouldEqual, 1)
	})
}

func TestHandleShouldNotBuildMessageWhenConditionsNotMet(t *testing.T) {
	convey.Convey("should not build message when signal type and strategy not match", t, func() {
		plugin := NewHotSwitchPlugin()
		hotSwitchPlugin, err := getHotSwitchPlugin(plugin)
		convey.So(err, convey.ShouldBeNil)
		hotSwitchPlugin.signalType = emptyString
		hotSwitchPlugin.changeStrategy = emptyString
		result, err := hotSwitchPlugin.Handle()
		convey.So(err, convey.ShouldBeNil)
		convey.So(result.Stage, convey.ShouldEqual, constant.HandleStageFinal)
		msgs, _ := hotSwitchPlugin.PullMsg()
		convey.So(len(msgs), convey.ShouldEqual, zeroLength)
	})
}

func TestPullMsgShouldReturnMessagesAndClear(t *testing.T) {
	convey.Convey("should return messages and clear pullMsgs when called", t, func() {
		plugin := NewHotSwitchPlugin()
		hotSwitchPlugin, err := getHotSwitchPlugin(plugin)
		convey.So(err, convey.ShouldBeNil)
		hotSwitchPlugin.signalType = clusterdconstant.HotSwitchSignalType
		hotSwitchPlugin.actions = testActions
		hotSwitchPlugin.changeStrategy = clusterdconstant.ProcessMigration
		hotSwitchPlugin.faultRanks = map[int]int{0: 1}
		hotSwitchPlugin.Handle()
		msgs, err := hotSwitchPlugin.PullMsg()
		convey.So(err, convey.ShouldBeNil)
		convey.So(len(msgs), convey.ShouldEqual, num1)
		msgs2, _ := hotSwitchPlugin.PullMsg()
		convey.So(len(msgs2), convey.ShouldEqual, zeroLength)
	})
}

func TestReleaseShouldReturnNil(t *testing.T) {
	convey.Convey("should return nil when called", t, func() {
		plugin := NewHotSwitchPlugin()
		hotSwitchPlugin, err := getHotSwitchPlugin(plugin)
		convey.So(err, convey.ShouldBeNil)
		err = hotSwitchPlugin.Release()
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestGetClusterInfoShouldSetFieldsWhenValid(t *testing.T) {
	convey.Convey("should set faultRanks actions uuid when cluster info valid", t, func() {
		plugin := NewHotSwitchPlugin()
		hotSwitchPlugin, err := getHotSwitchPlugin(plugin)
		convey.So(err, convey.ShouldBeNil)
		snapshot := createSnapshotWithClusterInfo(clusterdconstant.HotSwitchSignalType, testActions,
			clusterdconstant.ProcessMigration, testUuid, testFaultRanksJSON)
		err = hotSwitchPlugin.getClusterInfo(snapshot)
		convey.So(err, convey.ShouldBeNil)
		convey.So(len(hotSwitchPlugin.faultRanks), convey.ShouldEqual, num2)
		convey.So(hotSwitchPlugin.actions, convey.ShouldEqual, testActions)
		convey.So(hotSwitchPlugin.uuid, convey.ShouldEqual, testUuid)
	})
}

func TestGetClusterInfoShouldReturnErrorWhenClusterNotFound(t *testing.T) {
	convey.Convey("should return error when cluster info not found", t, func() {
		plugin := NewHotSwitchPlugin()
		hotSwitchPlugin, err := getHotSwitchPlugin(plugin)
		convey.So(err, convey.ShouldBeNil)
		snapshot := createEmptySnapshot()
		err = hotSwitchPlugin.getClusterInfo(snapshot)
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, "cluster info not found")
	})
}

func TestGetClusterInfoShouldReturnErrorWhenFaultRanksInvalid(t *testing.T) {
	convey.Convey("should return error when faultRanks json invalid", t, func() {
		plugin := NewHotSwitchPlugin()
		hotSwitchPlugin, err := getHotSwitchPlugin(plugin)
		convey.So(err, convey.ShouldBeNil)
		snapshot := createSnapshotWithClusterInfo(clusterdconstant.HotSwitchSignalType, testActions,
			clusterdconstant.ProcessMigration, testUuid, testInvalidJSON)
		err = hotSwitchPlugin.getClusterInfo(snapshot)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func TestBuildControllerMessageShouldAppendMessage(t *testing.T) {
	convey.Convey("should append message with correct fields when called", t, func() {
		plugin := NewHotSwitchPlugin()
		hotSwitchPlugin, err := getHotSwitchPlugin(plugin)
		convey.So(err, convey.ShouldBeNil)
		hotSwitchPlugin.actions = testActions
		hotSwitchPlugin.changeStrategy = clusterdconstant.ProcessMigration
		faultRanks := map[int]int{0: 1, 1: 2}
		hotSwitchPlugin.faultRanks = faultRanks
		hotSwitchPlugin.buildControllerMessage()
		convey.So(len(hotSwitchPlugin.pullMsgs), convey.ShouldEqual, 1)
		msg := hotSwitchPlugin.pullMsgs[0]
		convey.So(msg.Receiver[0], convey.ShouldEqual, constant.ControllerName)
		convey.So(msg.Body.MsgType, convey.ShouldEqual, constant.Action)
		convey.So(msg.Body.Code, convey.ShouldEqual, constant.HotSwitchCode)
		convey.So(msg.Body.Extension[constant.Actions], convey.ShouldEqual, testActions)
		convey.So(msg.Body.Extension[constant.ChangeStrategy], convey.ShouldEqual, clusterdconstant.ProcessMigration)
		faultRanksStr := msg.Body.Extension[constant.FaultRanks]
		var result map[int]int
		err = json.Unmarshal([]byte(faultRanksStr), &result)
		convey.So(err, convey.ShouldBeNil)
		convey.So(len(result), convey.ShouldEqual, num2)
		value1, ok := result[0]
		if !ok {
			t.Error("key 0 not exist")
		}
		convey.So(value1, convey.ShouldEqual, num1)
		value2, ok := result[1]
		if !ok {
			t.Error("key 1 not exist")
		}
		convey.So(value2, convey.ShouldEqual, num2)
	})
}

func TestResetPluginInfoShouldClearAllFields(t *testing.T) {
	convey.Convey("should clear all fields when called", t, func() {
		plugin := NewHotSwitchPlugin()
		hotSwitchPlugin, err := getHotSwitchPlugin(plugin)
		convey.So(err, convey.ShouldBeNil)
		hotSwitchPlugin.faultRanks = map[int]int{0: 1}
		hotSwitchPlugin.actions = testActions
		hotSwitchPlugin.signalType = clusterdconstant.HotSwitchSignalType
		hotSwitchPlugin.changeStrategy = clusterdconstant.ProcessMigration
		hotSwitchPlugin.resetPluginInfo()
		convey.So(len(hotSwitchPlugin.faultRanks), convey.ShouldEqual, zeroLength)
		convey.So(hotSwitchPlugin.actions, convey.ShouldEqual, emptyString)
		convey.So(hotSwitchPlugin.signalType, convey.ShouldEqual, emptyString)
		convey.So(hotSwitchPlugin.changeStrategy, convey.ShouldEqual, emptyString)
	})
}
