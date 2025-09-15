// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package recover a series of service function
package recover

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/common"
	"clusterd/pkg/interface/grpc/recover"
)

type hotSwitchTestCase struct {
	name                  string
	controller            *EventController
	mockFaultRanks        []*pb.FaultRank
	mockSignalEnqueueResp string
	mockSignalEnqueueCode common.RespCode
	mockSignalEnqueueErr  error
	expectedEvent         string
	expectedCode          common.RespCode
	expectedError         error
}

type signalTestConfig struct {
	signalType string
	actions    []string
	strategy   string
}

func buildHotSwitchTestCases() []hotSwitchTestCase {
	return []hotSwitchTestCase{
		{
			name:                  "successful hot switch preparation",
			controller:            &EventController{jobInfo: common.JobBaseInfo{JobId: "test-job-1"}, uuid: "test-uuid-1"},
			mockFaultRanks:        []*pb.FaultRank{{RankId: "0"}},
			mockSignalEnqueueResp: "success",
			mockSignalEnqueueCode: common.OK,
			mockSignalEnqueueErr:  nil,
			expectedEvent:         "success",
			expectedCode:          common.OK,
			expectedError:         nil,
		},
		{
			name:                  "hot switch preparation with error",
			controller:            &EventController{jobInfo: common.JobBaseInfo{JobId: "test-job-2"}, uuid: "test-uuid-2"},
			mockFaultRanks:        []*pb.FaultRank{{RankId: "1"}},
			mockSignalEnqueueResp: "",
			mockSignalEnqueueCode: common.ServerInnerError,
			mockSignalEnqueueErr:  nil,
			expectedEvent:         "",
			expectedCode:          common.ServerInnerError,
			expectedError:         nil,
		},
	}
}

func buildSignalTestCases(signalType string, actions []string, strategy string) []hotSwitchTestCase {
	return []hotSwitchTestCase{
		{
			name:                  "successful " + signalType + " handling",
			controller:            &EventController{jobInfo: common.JobBaseInfo{JobId: "test-job-1"}, uuid: "test-uuid-1"},
			mockSignalEnqueueResp: "success",
			mockSignalEnqueueCode: common.OK,
			mockSignalEnqueueErr:  nil,
			expectedEvent:         "success",
			expectedCode:          common.OK,
			expectedError:         nil,
		},
		{
			name:                  signalType + " handling with error",
			controller:            &EventController{jobInfo: common.JobBaseInfo{JobId: "test-job-2"}, uuid: "test-uuid-2"},
			mockSignalEnqueueResp: "",
			mockSignalEnqueueCode: common.ServerInnerError,
			mockSignalEnqueueErr:  nil,
			expectedEvent:         "",
			expectedCode:          common.ServerInnerError,
			expectedError:         nil,
		},
	}
}

func runSignalTest(t *testing.T, tests []hotSwitchTestCase, testFunc func(*EventController) (string, common.RespCode, error), config signalTestConfig) {
	for _, tt := range tests {
		convey.Convey(tt.name, t, func() {
			patch := gomonkey.ApplyPrivateMethod((*EventController)(nil), "signalEnqueue",
				func(_ *EventController, signal *pb.ProcessManageSignal) (string, common.RespCode, error) {
					convey.So(signal.Uuid, convey.ShouldEqual, tt.controller.uuid)
					convey.So(signal.JobId, convey.ShouldEqual, tt.controller.jobInfo.JobId)
					convey.So(signal.SignalType, convey.ShouldEqual, config.signalType)
					convey.So(signal.Actions, convey.ShouldResemble, config.actions)
					convey.So(signal.ChangeStrategy, convey.ShouldEqual, config.strategy)
					return tt.mockSignalEnqueueResp, tt.mockSignalEnqueueCode, tt.mockSignalEnqueueErr
				})
			defer patch.Reset()
			event, code, err := testFunc(tt.controller)
			convey.So(event, convey.ShouldEqual, tt.expectedEvent)
			convey.So(code, convey.ShouldEqual, tt.expectedCode)
			convey.So(err, convey.ShouldEqual, tt.expectedError)
		})
	}
}

func TestNotifyPrepareHotSwitch(t *testing.T) {
	tests := buildHotSwitchTestCases()
	for _, tt := range tests {
		convey.Convey(tt.name, t, func() {
			patch := gomonkey.NewPatches()
			defer patch.Reset()
			patch.ApplyPrivateMethod(tt.controller, "normalFaultAssociateSameNodeRank",
				func(_ *EventController) ([]*pb.FaultRank, map[string]string) {
					return tt.mockFaultRanks, nil
				}).ApplyPrivateMethod(tt.controller, "signalEnqueue",
				func(_ *EventController, signal *pb.ProcessManageSignal) (string, common.RespCode, error) {
					convey.So(signal.Uuid, convey.ShouldEqual, tt.controller.uuid)
					convey.So(signal.JobId, convey.ShouldEqual, tt.controller.jobInfo.JobId)
					convey.So(signal.SignalType, convey.ShouldEqual, constant.HotSwitchSignalType)
					convey.So(signal.Actions, convey.ShouldResemble, hotSwitchActions)
					convey.So(signal.FaultRanks, convey.ShouldResemble, tt.mockFaultRanks)
					convey.So(signal.ChangeStrategy, convey.ShouldEqual, "")
					return tt.mockSignalEnqueueResp, tt.mockSignalEnqueueCode, tt.mockSignalEnqueueErr
				})

			event, code, err := tt.controller.notifyPrepareHotSwitch()

			convey.So(event, convey.ShouldEqual, tt.expectedEvent)
			convey.So(code, convey.ShouldEqual, tt.expectedCode)
			convey.So(err, convey.ShouldEqual, tt.expectedError)
		})
	}
}

func TestNotifyNewPodFailedHandler(t *testing.T) {
	tests := buildSignalTestCases("new pod failed", stopHotSwitchActions, "")
	config := signalTestConfig{
		signalType: constant.HotSwitchSignalType,
		actions:    stopHotSwitchActions,
		strategy:   "",
	}
	testFunc := func(controller *EventController) (string, common.RespCode, error) {
		return controller.notifyNewPodFailedHandler()
	}
	runSignalTest(t, tests, testFunc, config)
}

func TestNotifyNewPodRunningHandler(t *testing.T) {
	tests := buildSignalTestCases("new pod running", newPodRunningActions, "")
	config := signalTestConfig{
		signalType: constant.HotSwitchSignalType,
		actions:    newPodRunningActions,
		strategy:   "",
	}
	testFunc := func(controller *EventController) (string, common.RespCode, error) {
		return controller.notifyNewPodRunningHandler()
	}
	runSignalTest(t, tests, testFunc, config)
}

func TestNotifyRestartTrain(t *testing.T) {
	tests := buildSignalTestCases("restart train", changeStrategyActions, constant.ProcessMigration)
	config := signalTestConfig{
		signalType: constant.ChangeStrategySignalType,
		actions:    changeStrategyActions,
		strategy:   constant.ProcessMigration,
	}
	testFunc := func(controller *EventController) (string, common.RespCode, error) {
		return controller.notifyRestartTrain()
	}
	runSignalTest(t, tests, testFunc, config)
}

func TestNotifyDumpForHotSwitch(t *testing.T) {
	tests := buildSignalTestCases("dump", changeStrategyActions, constant.ProcessMigration)
	config := signalTestConfig{
		signalType: constant.ChangeStrategySignalType,
		actions:    changeStrategyActions,
		strategy:   constant.ProcessDumpStrategyName,
	}
	testFunc := func(controller *EventController) (string, common.RespCode, error) {
		return controller.notifyDump()
	}
	runSignalTest(t, tests, testFunc, config)
}
