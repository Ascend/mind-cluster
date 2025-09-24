// Copyright (c) Huawei Technologies Co., Ltd. 2024-2025. All rights reserved.

// Package podrescheduling for taskd manager plugin test
package podrescheduling

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
	clusterdconstant "clusterd/pkg/common/constant"
	"taskd/common/constant"
	"taskd/framework_backend/manager/infrastructure"
	"taskd/framework_backend/manager/infrastructure/storage"
	"taskd/toolkit_backend/net/common"
)

var (
	reportRestartTime = 3
	retryTime         = 5
	retryTime1        = 3
	msgLen            = 5
)

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
func TestNewPodReschedulingPlugin(t *testing.T) {
	plugin := NewPodReschedulingPlugin()
	assert.NotNil(t, plugin)
	podPlugin, ok := plugin.(*PodReschedulingPlugin)
	assert.True(t, ok)
	assert.Empty(t, podPlugin.pullMsgs)
	assert.Empty(t, podPlugin.faultAgentStatus)
	assert.Equal(t, -1, podPlugin.restartTimes)
}

func TestPodReschedulingPlugin_Name(t *testing.T) {
	plugin, ok := NewPodReschedulingPlugin().(*PodReschedulingPlugin)
	if !ok {
		hwlog.RunLog.Errorf("NewPodReschedulingPlugin failed, expect *PodReschedulingPlugin, got %T", plugin)
		return
	}
	name := plugin.Name()
	assert.Equal(t, constant.PodReschedulingPluginName, name)
}

func TestPodReschedulingPlugin_Release(t *testing.T) {
	plugin, ok := NewPodReschedulingPlugin().(*PodReschedulingPlugin)
	if !ok {
		hwlog.RunLog.Errorf("NewPodReschedulingPlugin failed, expect *PodReschedulingPlugin, got %T", plugin)
		return
	}
	err := plugin.Release()
	assert.NoError(t, err)
}

func TestPodReschedulingPlugin_ResetPluginInfo(t *testing.T) {
	plugin := &PodReschedulingPlugin{
		processStatus:    "processing",
		faultAgentStatus: map[string]bool{"agent1": true, "agent2": false},
		exitNum:          5,
		faultOccur:       true,
		exitStratrgy:     true,
		actions:          []string{"action1", "action2"},
		isRetried:        true,
	}

	plugin.resetPluginInfo()
	assert.Empty(t, plugin.processStatus)
	assert.Empty(t, plugin.faultAgentStatus)
	assert.Zero(t, plugin.exitNum)
	assert.False(t, plugin.faultOccur)
	assert.False(t, plugin.exitStratrgy)
	assert.Empty(t, plugin.actions)
	assert.False(t, plugin.isRetried)
}

func TestPodReschedulingPlugin_UpdatePluginInfo(t *testing.T) {
	plugin, ok := NewPodReschedulingPlugin().(*PodReschedulingPlugin)
	if !ok {
		hwlog.RunLog.Errorf("NewPodReschedulingPlugin failed, expect *PodReschedulingPlugin, got %T", plugin)
		return
	}
	shot := storage.SnapShot{
		AgentInfos: &storage.AgentInfos{
			Agents: map[string]*storage.AgentInfo{
				"agent1": {Status: map[string]string{constant.ReportFaultRank: "rank1"}},
				"agent2": {Status: map[string]string{}},
			},
		},
	}

	plugin.updatePluginInfo(shot)
	assert.True(t, plugin.faultAgentStatus["agent1"])
	assert.False(t, plugin.faultAgentStatus["agent2"])
}

func TestPodReschedulingPlugin_CheckFaultrecover(t *testing.T) {
	plugin, ok := NewPodReschedulingPlugin().(*PodReschedulingPlugin)
	if !ok {
		hwlog.RunLog.Errorf("NewPodReschedulingPlugin failed, expect *PodReschedulingPlugin, got %T", plugin)
		return
	}

	shot := storage.SnapShot{MgrInfos: nil}
	result := plugin.checkFaultrecover(shot)
	assert.False(t, result)

	shot = storage.SnapShot{
		MgrInfos: &storage.MgrInfo{Status: map[string]string{}},
	}
	result = plugin.checkFaultrecover(shot)
	assert.False(t, result)

	shot = storage.SnapShot{
		MgrInfos: &storage.MgrInfo{Status: map[string]string{constant.FaultRecover: "true"}},
	}
	result = plugin.checkFaultrecover(shot)
	assert.True(t, result)
}

func TestPodReschedulingPlugin_FirstGetRestartTime(t *testing.T) {
	plugin, ok := NewPodReschedulingPlugin().(*PodReschedulingPlugin)
	if !ok {
		hwlog.RunLog.Errorf("NewPodReschedulingPlugin failed, expect *PodReschedulingPlugin, got %T", plugin)
		return
	}
	shot := storage.SnapShot{
		AgentInfos: &storage.AgentInfos{
			Agents: map[string]*storage.AgentInfo{
				"agent1": {Status: map[string]string{constant.ReportRestartTime: "3"}},
			},
		},
	}

	plugin.firstGetRestartTime(shot)
	assert.Equal(t, reportRestartTime, plugin.restartTimes)

	plugin.restartTimes = -1
	shot.AgentInfos.Agents["agent1"].Status[constant.ReportRestartTime] = "invalid"
	plugin.firstGetRestartTime(shot)
	assert.Equal(t, 0, plugin.restartTimes)
}

func TestPodReschedulingPlugin_CheckExitStrategy(t *testing.T) {
	plugin, ok := NewPodReschedulingPlugin().(*PodReschedulingPlugin)
	if !ok {
		hwlog.RunLog.Errorf("NewPodReschedulingPlugin failed, expect *PodReschedulingPlugin, got %T", plugin)
		return
	}
	nodeIds := []string{"0", "1"}
	actions := []string{"action1"}
	nodeJson, err := json.Marshal(nodeIds)
	if err != nil {
		hwlog.RunLog.Errorf("marshal nodeIds failed, err: %v", err)
		return
	}
	actionJson, err := json.Marshal(actions)
	if err != nil {
		hwlog.RunLog.Errorf("marshal actions failed, err: %v", err)
		return
	}

	shot := storage.SnapShot{
		ClusterInfos: &storage.ClusterInfos{
			Clusters: map[string]*storage.ClusterInfo{
				constant.ClusterDRank: {
					Command: map[string]string{
						constant.Uuid:           "test-uuid",
						constant.ChangeStrategy: clusterdconstant.ProcessExitStrategyName,
						constant.NodeRankIds:    string(nodeJson),
						constant.Actions:        string(actionJson),
					},
				},
			},
		},
		AgentInfos: &storage.AgentInfos{
			Agents: map[string]*storage.AgentInfo{
				common.AgentRole + "0": {},
				common.AgentRole + "1": {},
			},
		},
	}

	result, err := plugin.checkExitStrategy(shot)
	assert.NoError(t, err)
	assert.Equal(t, constant.CandidateStatus, result.CandidateStatus)
	assert.Equal(t, "test-uuid", plugin.uuid)
	assert.True(t, plugin.faultAgentStatus[common.AgentRole+"0"])
	assert.True(t, plugin.faultAgentStatus[common.AgentRole+"1"])
}

func TestPodReschedulingPlugin_CheckResetConfig(t *testing.T) {
	plugin := &PodReschedulingPlugin{
		oldRetryTimes: 3,
	}

	configData := api.ResetCmInfo{RetryTime: 5}
	configBytes, err := json.Marshal(configData)
	if err != nil {
		hwlog.RunLog.Errorf("marshal reset config failed, err: %v", err)
		return
	}
	mock := gomonkey.ApplyFuncReturn(utils.LoadFile, configBytes, nil)
	defer mock.Reset()

	result := plugin.checkResetConfig()
	assert.True(t, result)
	assert.Equal(t, retryTime, plugin.newRetryTimes)
	assert.True(t, plugin.isRetried)
}

func TestPodReschedulingPlugin_CheckResetConfig_FileError(t *testing.T) {
	plugin, ok := NewPodReschedulingPlugin().(*PodReschedulingPlugin)
	if !ok {
		hwlog.RunLog.Errorf("NewPodReschedulingPlugin failed, expect *PodReschedulingPlugin, got %T", plugin)
		return
	}

	mock := gomonkey.ApplyFuncReturn(utils.LoadFile, nil, errors.New("file not found"))
	defer mock.Reset()

	result := plugin.checkResetConfig()
	assert.False(t, result)
}

func TestPodReschedulingPlugin_CheckResetConfig_UnmarshalError(t *testing.T) {
	plugin, ok := NewPodReschedulingPlugin().(*PodReschedulingPlugin)
	if !ok {
		hwlog.RunLog.Errorf("NewPodReschedulingPlugin failed, expect *PodReschedulingPlugin, got %T", plugin)
		return
	}

	mock := gomonkey.ApplyFuncReturn(utils.LoadFile, []byte("invalid json"), nil)
	defer mock.Reset()

	result := plugin.checkResetConfig()
	assert.False(t, result)
}

func TestPodReschedulingPlugin_CheckResetConfig_NegativeRetryTime(t *testing.T) {
	plugin, ok := NewPodReschedulingPlugin().(*PodReschedulingPlugin)
	if !ok {
		hwlog.RunLog.Errorf("NewPodReschedulingPlugin failed, expect *PodReschedulingPlugin, got %T", plugin)
		return
	}

	configData := api.ResetCmInfo{RetryTime: -1}
	configBytes, err := json.Marshal(configData)
	if err != nil {
		hwlog.RunLog.Errorf("marshal reset config failed, err: %v", err)
		return
	}
	mock := gomonkey.ApplyFuncReturn(utils.LoadFile, configBytes, nil)
	defer mock.Reset()

	result := plugin.checkResetConfig()
	assert.False(t, result)
}

func TestPodReschedulingPlugin_Predicate_NoFault(t *testing.T) {
	plugin, ok := NewPodReschedulingPlugin().(*PodReschedulingPlugin)
	if !ok {
		hwlog.RunLog.Errorf("NewPodReschedulingPlugin failed, expect *PodReschedulingPlugin, got %T", plugin)
		return
	}
	shot := storage.SnapShot{
		AgentInfos: &storage.AgentInfos{
			Agents: map[string]*storage.AgentInfo{
				"agent1": {Status: map[string]string{}},
			},
		},
	}

	result, err := plugin.Predicate(shot)
	assert.NoError(t, err)
	assert.Equal(t, constant.UnselectStatus, result.CandidateStatus)
}

func TestPodReschedulingPlugin_Predicate_WithFault(t *testing.T) {
	plugin, ok := NewPodReschedulingPlugin().(*PodReschedulingPlugin)
	if !ok {
		hwlog.RunLog.Errorf("NewPodReschedulingPlugin failed, expect *PodReschedulingPlugin, got %T", plugin)
		return
	}
	shot := storage.SnapShot{
		AgentInfos: &storage.AgentInfos{
			Agents: map[string]*storage.AgentInfo{
				"agent1": {Status: map[string]string{constant.ReportFaultRank: "rank1"}},
			},
		},
	}

	result, err := plugin.Predicate(shot)
	assert.NoError(t, err)
	assert.Equal(t, constant.CandidateStatus, result.CandidateStatus)
}

func TestPodReschedulingPlugin_Handle(t *testing.T) {
	plugin := &PodReschedulingPlugin{
		restartTimes:     2,
		faultAgentStatus: map[string]bool{"agent1": true, "agent2": false},
		exitNum:          0,
	}

	result, err := plugin.Handle()
	assert.NoError(t, err)
	assert.Equal(t, constant.HandleStageProcess, result.Stage)
	assert.Equal(t, 1, plugin.restartTimes)
	assert.Equal(t, 1, plugin.exitNum)
}

func TestPodReschedulingPlugin_Handle_IsRetried(t *testing.T) {
	plugin := &PodReschedulingPlugin{
		restartTimes:     2,
		newRetryTimes:    3,
		oldRetryTimes:    1,
		faultAgentStatus: map[string]bool{"agent1": true, "agent2": false},
		isRetried:        true,
	}

	result, err := plugin.Handle()
	assert.NoError(t, err)
	assert.Equal(t, constant.HandleStageFinal, result.Stage)
	assert.Equal(t, 1, plugin.restartTimes)
	assert.Equal(t, retryTime1, plugin.oldRetryTimes)
}

func TestPodReschedulingPlugin_PullMsg(t *testing.T) {
	plugin := &PodReschedulingPlugin{
		pullMsgs: []infrastructure.Msg{
			{Receiver: []string{"receiver1"}},
		},
	}

	msgs, err := plugin.PullMsg()
	assert.NoError(t, err)
	assert.Len(t, msgs, 1)
	assert.Equal(t, "receiver1", msgs[0].Receiver[0])
	assert.Empty(t, plugin.pullMsgs)
}

func TestPodReschedulingPlugin_AddHandleMsgs(t *testing.T) {
	plugin := &PodReschedulingPlugin{
		restartTimes: 2,
		uuid:         "test-uuid",
		actions:      []string{"action1"},
	}

	plugin.addHandleMsgs([]string{"agent1"}, []string{"agent2"})
	assert.Len(t, plugin.pullMsgs, msgLen)
}
