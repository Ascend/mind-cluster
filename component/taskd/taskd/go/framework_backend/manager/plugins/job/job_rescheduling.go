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
	"ascend-common/common-utils/hwlog"
	clusterd_constant "clusterd/pkg/common/constant"
	"taskd/common/constant"
	"taskd/common/utils"
	"taskd/framework_backend/manager/infrastructure"
	"taskd/framework_backend/manager/infrastructure/storage"
	"taskd/toolkit_backend/net/common"
)

// JobReschedulingPlugin job rescheduling plugin
type JobReschedulingPlugin struct {
	pullMsgs      []infrastructure.Msg
	faultOccur    bool
	processStatus string
	killMaster    bool
}

var (
	agent0ExitMsg = infrastructure.Msg{
		Receiver: []string{common.AgentRole + "0"},
		Body: storage.MsgBody{
			MsgType: constant.Action,
			Code:    constant.ExitAgentCode,
		},
	}
)

// NewJobReschedulingPlugin new job rescheduling plugin
func NewJobReschedulingPlugin() infrastructure.ManagerPlugin {
	return &JobReschedulingPlugin{
		pullMsgs:      make([]infrastructure.Msg, 0),
		faultOccur:    false,
		processStatus: "",
		killMaster:    false,
	}
}

// Name name of job rescheduling plugin
func (job *JobReschedulingPlugin) Name() string {
	return constant.JobReschedulingPluginName
}

// Handle handle job rescheduling plugin
func (job *JobReschedulingPlugin) Handle() (infrastructure.HandleResult, error) {
	job.processStatus = constant.HandleStageProcess
	if job.killMaster {
		job.pullMsgs = append(job.pullMsgs, infrastructure.Msg{
			Receiver: []string{constant.ControllerName},
			Body: storage.MsgBody{
				MsgType: constant.Action,
				Code:    0,
				Extension: map[string]string{
					constant.Actions: utils.ObjToString([]string{constant.DestroyController}),
				},
			},
		})
		job.pullMsgs = append(job.pullMsgs, agent0ExitMsg)
		job.processStatus = ""
		job.resetPluginInfo()
		return infrastructure.HandleResult{Stage: constant.HandleStageFinal}, nil
	}
	if !job.faultOccur {
		hwlog.RunLog.Info("JobReschedulingPlugin not fault occur")
		job.resetPluginInfo()
		return infrastructure.HandleResult{
			Stage: constant.HandleStageFinal,
		}, nil
	}

	hwlog.RunLog.Info("JobReschedulingPlugin handle fault")
	job.pullMsgs = append(job.pullMsgs, agent0ExitMsg)
	job.processStatus = ""
	return infrastructure.HandleResult{Stage: constant.HandleStageFinal}, nil
}

// Handle handle job rescheduling plugin
func (job *JobReschedulingPlugin) PullMsg() ([]infrastructure.Msg, error) {
	msgs := job.pullMsgs
	job.pullMsgs = make([]infrastructure.Msg, 0)
	return msgs, nil
}

// Predicate predicate job rescheduling plugin
func (job *JobReschedulingPlugin) Predicate(shot storage.SnapShot) (infrastructure.PredicateResult, error) {
	if job.processStatus != "" {
		hwlog.RunLog.Infof("JobReschedulingPlugin Predicate processStatus:%v", job.processStatus)
		job.updatePluginInfo(shot)
		return infrastructure.PredicateResult{PluginName: job.Name(),
			CandidateStatus: constant.CandidateStatus,
			PredicateStream: map[string]string{
				constant.ResumeTrainingAfterFaultStream: "",
			}}, nil
	}
	job.resetPluginInfo()
	if job.checkKillMaster(shot) {
		if job.killMaster {
			hwlog.RunLog.Info("JobReschedulingPlugin checkKillMaster kill master")
			return infrastructure.PredicateResult{PluginName: job.Name(),
				CandidateStatus: constant.CandidateStatus,
				PredicateStream: map[string]string{
					constant.ResumeTrainingAfterFaultStream: "",
				}}, nil
		} else {
			return infrastructure.PredicateResult{
				PluginName: job.Name(), CandidateStatus: constant.UnselectStatus, PredicateStream: nil}, nil
		}
	}

	for agentName, agent := range shot.AgentInfos.Agents {
		if agentName == common.AgentRole+"0" && agent.Status[constant.ReportFaultRank] != "" {
			job.faultOccur = true
			hwlog.RunLog.Info("agent 0 fault job job rescheduling")
			return infrastructure.PredicateResult{PluginName: job.Name(),
				CandidateStatus: constant.CandidateStatus,
				PredicateStream: map[string]string{
					constant.ResumeTrainingAfterFaultStream: "",
				}}, nil
		}
	}
	hwlog.RunLog.Debug("JobReschedulingPlugin not fault occur")
	return infrastructure.PredicateResult{
		PluginName: job.Name(), CandidateStatus: constant.UnselectStatus, PredicateStream: nil}, nil
}

// Release release job rescheduling plugin
func (job *JobReschedulingPlugin) Release() error {
	return nil
}

func (job *JobReschedulingPlugin) resetPluginInfo() {
	job.faultOccur = false
	job.processStatus = ""
	job.killMaster = false
}

func (job *JobReschedulingPlugin) updatePluginInfo(shot storage.SnapShot) {
	for agentName, agent := range shot.AgentInfos.Agents {
		if agentName == common.AgentRole+"0" && agent.Status[constant.ReportFaultRank] != "" {
			job.killMaster = true
		}
	}
	clusterInfo, ok := shot.ClusterInfos.Clusters[constant.ClusterDRank]
	if ok {
		if clusterInfo.Command[constant.SignalType] == clusterd_constant.KillMasterSignalType {
			job.killMaster = true
		}
	}
}

func (job *JobReschedulingPlugin) checkKillMaster(shot storage.SnapShot) bool {
	if shot.MgrInfos == nil {
		hwlog.RunLog.Debug("mgr info is empty")
		return false
	}
	if shot.MgrInfos.Status[constant.FaultRecover] == "" {
		hwlog.RunLog.Debug("fault recover status is empty")
		return false
	}
	clusterInfo, ok := shot.ClusterInfos.Clusters[constant.ClusterDRank]
	if ok {
		if clusterInfo.Command[constant.SignalType] == clusterd_constant.KillMasterSignalType {
			hwlog.RunLog.Info("kill master signal type")
			job.killMaster = true
		}
	}
	job.checktRank0Fault(shot)
	return true
}

func (job *JobReschedulingPlugin) checktRank0Fault(shot storage.SnapShot) {
	if shot.AgentInfos == nil {
		hwlog.RunLog.Debug("agent info is empty")
		return
	}
	agent0Info, ok := shot.AgentInfos.Agents[common.AgentRole+"0"]
	if !ok {
		hwlog.RunLog.Info("JobReschedulingPlugin checkRank0Fault agent 0 not exist")
		return
	}
	if agent0Info.Status[constant.ReportFaultRank] != "" {
		hwlog.RunLog.Info("JobReschedulingPlugin checktRank0Fault agent 0 fault")
		if shot.MgrInfos != nil && shot.MgrInfos.Status[constant.FaultRecover] != "" {
			hwlog.RunLog.Infof("JobReschedulingPlugin fault recover %v, wait clusterD notify...",
				shot.MgrInfos.Status[constant.FaultRecover])
			return
		}
		job.killMaster = true
	}
}
