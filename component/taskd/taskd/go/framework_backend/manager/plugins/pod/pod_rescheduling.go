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
package podrescheduling

import (
	"encoding/json"
	"errors"
	"strconv"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	commonutils "ascend-common/common-utils/utils"
	clusterd_constant "clusterd/pkg/common/constant"
	"taskd/common/constant"
	"taskd/common/utils"
	"taskd/framework_backend/manager/infrastructure"
	"taskd/framework_backend/manager/infrastructure/storage"
	"taskd/toolkit_backend/net/common"
)

// PodReschedulingPlugin pod rescheduling plugin
type PodReschedulingPlugin struct {
	pullMsgs         []infrastructure.Msg
	processStatus    string
	faultAgentStatus map[string]bool
	faultOccur       bool
	restartTimes     int
	exitNum          int
	uuid             string
	exitStratrgy     bool
	actions          []string
	handleMap        map[string]string
	oldRetryTimes    int
	newRetryTimes    int
	isRetried        bool
}

var (
	defaultTimes = -1
	unselectMsg  = infrastructure.PredicateResult{
		PluginName:      constant.PodReschedulingPluginName,
		CandidateStatus: constant.UnselectStatus, PredicateStream: nil}
)

// NewPodReschedulingPlugin new pod rescheduling plugin
func NewPodReschedulingPlugin() infrastructure.ManagerPlugin {
	return &PodReschedulingPlugin{
		pullMsgs:         make([]infrastructure.Msg, 0),
		faultAgentStatus: make(map[string]bool),
		restartTimes:     defaultTimes,
		actions:          make([]string, 0),
		handleMap:        make(map[string]string),
	}
}

// Name name of pod rescheduling plugin
func (pod *PodReschedulingPlugin) Name() string {
	return constant.PodReschedulingPluginName
}

// Handle handle pod rescheduling plugin
func (pod *PodReschedulingPlugin) Handle() (infrastructure.HandleResult, error) {
	pod.processStatus = constant.HandleStageProcess
	if pod.uuid != "" {
		pod.handleMap[pod.uuid] = constant.HandleDone
	}
	exitReceiver := make([]string, 0)
	restartReceiver := make([]string, 0)
	for agentName, faultStatus := range pod.faultAgentStatus {
		if faultStatus {
			exitReceiver = append(exitReceiver, agentName)
		} else {
			restartReceiver = append(restartReceiver, agentName)
		}
	}
	if pod.isRetried {
		pod.restartTimes -= 1
		pod.oldRetryTimes = pod.newRetryTimes
		pod.addHandleMsgs(exitReceiver, restartReceiver)
		hwlog.RunLog.Infof("pod rescheduling plugin handle isRetried, restart times: %d, exit receiver: %v, restart receiver: %v",
			pod.restartTimes, exitReceiver, restartReceiver)
		pod.resetPluginInfo()
		return infrastructure.HandleResult{Stage: constant.HandleStageFinal}, nil
	}

	if len(exitReceiver) == 0 {
		pod.resetPluginInfo()
		retryTime, err := pod.getCmRetryTims()
		if err != nil {
			hwlog.RunLog.Errorf("get cm retry time failed, err: %v", err)
			return infrastructure.HandleResult{Stage: constant.HandleStageFinal}, nil
		}
		pod.oldRetryTimes = retryTime
		return infrastructure.HandleResult{Stage: constant.HandleStageFinal}, nil
	}
	if pod.exitNum != 0 {
		hwlog.RunLog.Infof("pod rescheduling plugin handle, exit num: %d", pod.exitNum)
		return infrastructure.HandleResult{Stage: constant.HandleStageProcess}, nil
	}
	pod.restartTimes -= 1
	hwlog.RunLog.Infof("pod rescheduling plugin handle, restart times: %d", pod.restartTimes)
	hwlog.RunLog.Infof("pod rescheduling plugin handle, exit receiver: %v", exitReceiver)
	hwlog.RunLog.Infof("pod rescheduling plugin handle, restart receiver: %v", restartReceiver)
	pod.addHandleMsgs(exitReceiver, restartReceiver)
	pod.exitNum = len(exitReceiver)

	return infrastructure.HandleResult{Stage: constant.HandleStageProcess}, nil
}

func (pod *PodReschedulingPlugin) addHandleMsgs(exitReciver []string, restartReceiver []string) {
	if pod.uuid != "" {
		pod.pullMsgs = append(pod.pullMsgs, infrastructure.Msg{
			Receiver: []string{constant.ControllerName},
			Body: storage.MsgBody{
				MsgType: constant.Action,
				Extension: map[string]string{
					constant.Actions:        utils.ObjToString(pod.actions),
					constant.ChangeStrategy: clusterd_constant.ProcessExitStrategyName,
				},
			},
		})
	}
	pod.pullMsgs = append(pod.pullMsgs, infrastructure.Msg{
		Receiver: []string{common.MgrRole},
		Body: storage.MsgBody{
			MsgType: constant.Action,
			Code:    constant.RestartTimeCode,
			Message: strconv.Itoa(pod.restartTimes),
		},
	})
	pod.pullMsgs = append(pod.pullMsgs, infrastructure.Msg{
		Receiver: exitReciver,
		Body: storage.MsgBody{
			MsgType:   constant.Action,
			Code:      constant.ExitAgentCode,
			Extension: map[string]string{},
		},
	})
	pod.pullMsgs = append(pod.pullMsgs, infrastructure.Msg{
		Receiver: restartReceiver,
		Body: storage.MsgBody{
			MsgType: constant.Action,
			Code:    constant.RestartAgentCode,
			Message: strconv.Itoa(pod.restartTimes),
		},
	})
	pod.pullMsgs = append(pod.pullMsgs, infrastructure.Msg{
		Receiver: []string{constant.ControllerName},
		Body: storage.MsgBody{
			MsgType: constant.Action,
			Code:    0,
			Extension: map[string]string{
				constant.Actions: utils.ObjToString([]string{constant.RestartController}),
			},
		},
	})
}

// Handle handle pod rescheduling plugin
func (pod *PodReschedulingPlugin) PullMsg() ([]infrastructure.Msg, error) {
	msgs := pod.pullMsgs
	pod.pullMsgs = make([]infrastructure.Msg, 0)
	return msgs, nil
}

// Predicate predicate job rescheduling plugin
func (pod *PodReschedulingPlugin) Predicate(shot storage.SnapShot) (infrastructure.PredicateResult, error) {
	if pod.processStatus != "" {
		pod.updatePluginInfo(shot)
		return infrastructure.PredicateResult{PluginName: pod.Name(),
			CandidateStatus: constant.CandidateStatus,
			PredicateStream: map[string]string{
				constant.ResumeTrainingAfterFaultStream: "",
			}}, nil
	}
	pod.resetPluginInfo()
	if pod.checkFaultrecover(shot) {
		return pod.checkExitStrategy(shot)
	}
	pod.firstGetRestartTime(shot)

	for agentName, agentInfo := range shot.AgentInfos.Agents {
		pod.faultAgentStatus[agentName] = false
		if agentName == common.AgentRole+"0" && agentInfo.Status[constant.ReportFaultRank] != "" {
			hwlog.RunLog.Info("agent 0 fault, pod rescheduling plugin unselect")
			return infrastructure.PredicateResult{
				PluginName: pod.Name(), CandidateStatus: constant.UnselectStatus, PredicateStream: nil}, nil
		}
		if agentInfo.Status[constant.ReportFaultRank] != "" {
			pod.faultAgentStatus[agentName] = true
			pod.faultOccur = true
		}
	}
	hwlog.RunLog.Debugf("pod rescheduling plugin predicate, fault agent status: %v", pod.faultAgentStatus)
	if pod.faultOccur {
		hwlog.RunLog.Info("pod.faultOccur is true, pluin candidate")
		return infrastructure.PredicateResult{PluginName: pod.Name(),
			CandidateStatus: constant.CandidateStatus,
			PredicateStream: map[string]string{
				constant.ResumeTrainingAfterFaultStream: "",
			}}, nil
	}
	if pod.checkResetConfig() {
		hwlog.RunLog.Info("reset json changed, pluin candidate")
		return infrastructure.PredicateResult{PluginName: pod.Name(),
			CandidateStatus: constant.CandidateStatus,
			PredicateStream: map[string]string{constant.ResumeTrainingAfterFaultStream: ""}}, nil
	}
	return infrastructure.PredicateResult{
		PluginName: pod.Name(), CandidateStatus: constant.UnselectStatus, PredicateStream: nil}, nil
}

// Release release pod rescheduling plugin
func (pod *PodReschedulingPlugin) Release() error {
	return nil
}

func (pod *PodReschedulingPlugin) resetPluginInfo() {
	pod.processStatus = ""
	pod.faultAgentStatus = make(map[string]bool)
	pod.exitNum = 0
	pod.faultOccur = false
	pod.exitStratrgy = false
	pod.actions = make([]string, 0)
	pod.isRetried = false
}

func (pod *PodReschedulingPlugin) updatePluginInfo(shot storage.SnapShot) {
	for agentName, agentInfo := range shot.AgentInfos.Agents {
		if agentInfo.Status[constant.ReportFaultRank] != "" {
			pod.faultAgentStatus[agentName] = true
		} else {
			pod.faultAgentStatus[agentName] = false
		}
	}
}

func (pod *PodReschedulingPlugin) checkFaultrecover(shot storage.SnapShot) bool {
	if shot.MgrInfos == nil {
		hwlog.RunLog.Info("mgr info is empty")
		return false
	}
	if shot.MgrInfos.Status[constant.FaultRecover] == "" {
		hwlog.RunLog.Debug("fault recover status is empty")
		return false
	}
	return true
}

func (pod *PodReschedulingPlugin) firstGetRestartTime(shot storage.SnapShot) {
	var err error
	for _, agentInfo := range shot.AgentInfos.Agents {
		if agentInfo.Status[constant.ReportRestartTime] != "" && pod.restartTimes == -1 {
			hwlog.RunLog.Infof("pod rescheduling first set plugin restart times: %v", agentInfo.Status[constant.ReportRestartTime])
			pod.restartTimes, err = strconv.Atoi(agentInfo.Status[constant.ReportRestartTime])
			if err != nil {
				hwlog.RunLog.Error("firstGetRestartTime strconv.Atoi failed")
				return
			}
			break
		}
	}
}

func (pod *PodReschedulingPlugin) checkExitStrategy(shot storage.SnapShot) (infrastructure.PredicateResult, error) {
	pod.firstGetRestartTime(shot)
	clusterInfo, ok := shot.ClusterInfos.Clusters[constant.ClusterDRank]
	if !ok {
		hwlog.RunLog.Debug("cluster info not found")
		return infrastructure.PredicateResult{
			PluginName: pod.Name(), CandidateStatus: constant.UnselectStatus, PredicateStream: nil}, nil
	}
	clusterUuid := clusterInfo.Command[constant.Uuid]
	if pod.uuid == clusterUuid && pod.handleMap[clusterUuid] == constant.HandleDone {
		hwlog.RunLog.Debug("cluster uuid not change")
		return infrastructure.PredicateResult{
			PluginName: pod.Name(), CandidateStatus: constant.UnselectStatus, PredicateStream: nil}, nil
	}

	if clusterInfo.Command[constant.ChangeStrategy] != clusterd_constant.ProcessExitStrategyName {
		hwlog.RunLog.Debug("change strategy not process exit strategy")
		return infrastructure.PredicateResult{
			PluginName: pod.Name(), CandidateStatus: constant.UnselectStatus, PredicateStream: nil}, nil
	}
	pod.exitStratrgy = true
	exitAgent, err := utils.StringToObj[[]string](clusterInfo.Command[constant.NodeRankIds])
	if err != nil {
		pod.resetPluginInfo()
		hwlog.RunLog.Error("getExitStrategy string to obj failed")
		return infrastructure.PredicateResult{
			PluginName: pod.Name(), CandidateStatus: constant.UnselectStatus, PredicateStream: nil}, nil
	}
	for agentName, _ := range shot.AgentInfos.Agents {
		pod.faultAgentStatus[agentName] = false
	}
	for _, agentName := range exitAgent {
		pod.faultAgentStatus[common.AgentRole+agentName] = true
	}
	actions, err := utils.StringToObj[[]string](clusterInfo.Command[constant.Actions])
	if err != nil {
		pod.resetPluginInfo()
		hwlog.RunLog.Error("getExitStrategy string to obj failed")
		return infrastructure.PredicateResult{
			PluginName: pod.Name(), CandidateStatus: constant.UnselectStatus, PredicateStream: nil}, nil
	}
	pod.uuid = clusterUuid
	pod.actions = actions
	hwlog.RunLog.Infof("pod rescheduling plugin check exit strategy, uuid:%v", pod.uuid)
	return infrastructure.PredicateResult{PluginName: pod.Name(),
		CandidateStatus: constant.CandidateStatus,
		PredicateStream: map[string]string{
			constant.ResumeTrainingAfterFaultStream: "",
		}}, nil
}

func (pod *PodReschedulingPlugin) checkResetConfig() bool {
	retryTime, err := pod.getCmRetryTims()
	if err != nil {
		hwlog.RunLog.Errorf("get cm retry time failed, err: %v", err)
		return false
	}
	if pod.oldRetryTimes != retryTime {
		hwlog.RunLog.Infof("retry time changes, old: %v, new: %v", pod.oldRetryTimes, retryTime)
		pod.newRetryTimes = retryTime
		pod.isRetried = true
		return true
	}
	return false
}

func (pod *PodReschedulingPlugin) getCmRetryTims() (int, error) {
	configBytes, err := commonutils.LoadFile(constant.ResetConfigPath)
	if err != nil {
		return 0, err
	}
	var result api.ResetCmInfo
	err = json.Unmarshal(configBytes, &result)
	if err != nil {
		return 0, err
	}
	if result.RetryTime < 0 {
		return 0, errors.New("retry time is negative")
	}
	return result.RetryTime, nil
}
