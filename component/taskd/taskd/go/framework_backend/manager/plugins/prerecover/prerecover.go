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

// Package prerecover for preparation before recovery,such as stoptrain, globalfault, etc
package prerecover

import (
	"errors"

	"ascend-common/common-utils/hwlog"
	clusterdconstant "clusterd/pkg/common/constant"
	"taskd/common/constant"
	"taskd/framework_backend/manager/infrastructure"
	"taskd/framework_backend/manager/infrastructure/storage"
	pluginutils "taskd/framework_backend/manager/plugins/utils"
)

// preRecoverPlugin preRecover plugin define
type preRecoverPlugin struct {
	hasToken        bool
	shot            storage.SnapShot
	signalInfo      *pluginutils.SignalInfo
	HasSendMessages map[string]string
	lastUuid        string
}

// New creates an object
func New() infrastructure.ManagerPlugin {
	return &preRecoverPlugin{
		HasSendMessages: make(map[string]string),
		lastUuid:        "",
	}
}

// Name returns plugin name
func (s *preRecoverPlugin) Name() string {
	return constant.StopTrainPluginName
}

// Predicate check whether apply token
func (s *preRecoverPlugin) Predicate(shot storage.SnapShot) (infrastructure.PredicateResult, error) {
	s.shot = shot
	s.signalInfo = nil
	if s.hasToken {
		return infrastructure.PredicateResult{
			PluginName:      s.Name(),
			CandidateStatus: constant.CandidateStatus,
			PredicateStream: map[string]string{constant.ResumeTrainingAfterFaultStream: ""},
		}, nil
	}
	if err := s.getSignalInfo(); err != nil {
		hwlog.RunLog.Debugf("getSignalInfo error: %v", err)
		return infrastructure.PredicateResult{PluginName: s.Name(), CandidateStatus: constant.UnselectStatus}, nil
	}
	if s.signalInfo.SignalType == clusterdconstant.StopTrainSignalType {
		hwlog.RunLog.Info("get stop_train signal, apply for the token")
		return infrastructure.PredicateResult{
			PluginName:      s.Name(),
			CandidateStatus: constant.CandidateStatus,
			PredicateStream: map[string]string{constant.ResumeTrainingAfterFaultStream: ""},
		}, nil
	}
	if (s.signalInfo.SignalType == clusterdconstant.PreExitProcessSignalType ||
		s.signalInfo.SignalType == clusterdconstant.GlobalFaultSignalType) && s.signalInfo.Uuid != s.lastUuid {
		hwlog.RunLog.Infof("get %s signal, apply for the token", s.signalInfo.SignalType)
		return infrastructure.PredicateResult{
			PluginName:      s.Name(),
			CandidateStatus: constant.CandidateStatus,
			PredicateStream: map[string]string{constant.ResumeTrainingAfterFaultStream: ""},
		}, nil
	}
	return infrastructure.PredicateResult{CandidateStatus: constant.UnselectStatus}, nil
}

// Release releases token
func (s *preRecoverPlugin) Release() error {
	return nil
}

// Handle handles stream events
func (s *preRecoverPlugin) Handle() (infrastructure.HandleResult, error) {
	hwlog.RunLog.Infof("plugin[%s] enter handle", s.Name())
	s.hasToken = true
	if s.signalInfo == nil {
		err := s.getSignalInfo()
		if err != nil {
			hwlog.RunLog.Errorf("getSignalInfo error: %v", err)
			return infrastructure.HandleResult{Stage: constant.HandleStageException}, nil
		}
	}
	if s.signalInfo.SignalType == clusterdconstant.GlobalFaultSignalType ||
		s.signalInfo.SignalType == clusterdconstant.ChangeStrategySignalType {
		s.lastUuid = s.signalInfo.Uuid
		hwlog.RunLog.Info("get release token signal, need to release token")
		s.hasToken = false
		s.HasSendMessages = make(map[string]string)
		return infrastructure.HandleResult{Stage: constant.HandleStageFinal}, nil
	}
	if s.signalInfo.SignalType == clusterdconstant.PreExitProcessSignalType {
		s.lastUuid = s.signalInfo.Uuid
		hwlog.RunLog.Info("get pre_exit_process signal, need to exit fault process")
		s.hasToken = false
		return infrastructure.HandleResult{Stage: constant.HandleStageFinal}, nil
	}
	return infrastructure.HandleResult{Stage: constant.HandleStageProcess}, nil
}

// PullMsg returns messages to other module
func (s *preRecoverPlugin) PullMsg() ([]infrastructure.Msg, error) {
	if s.signalInfo == nil {
		hwlog.RunLog.Warn("signalInfo is nil")
		return nil, nil
	}
	if _, ok := s.HasSendMessages[s.signalInfo.SignalType+s.signalInfo.Command[constant.Actions]]; ok {
		hwlog.RunLog.Debugf("the signal info has dealed, signal info: %v", s.signalInfo)
		return nil, nil
	}
	msgs := make([]infrastructure.Msg, 0)
	if s.signalInfo.SignalType == clusterdconstant.PreExitProcessSignalType {
		msgs = append(msgs, s.signalInfo.GetMsgs()...)
		s.lastUuid = s.signalInfo.Uuid
		hwlog.RunLog.Infof("pull msgs: %+v", msgs)
		return msgs, nil
	}
	if s.signalInfo.SignalType == clusterdconstant.StopTrainSignalType || s.signalInfo.
		SignalType == clusterdconstant.FaultNodesExitSignalType || s.signalInfo.
		SignalType == clusterdconstant.GlobalFaultSignalType {
		msgs = append(msgs, s.signalInfo.GetMsgs()...)
	}
	s.HasSendMessages[s.signalInfo.SignalType+s.signalInfo.Command[constant.Actions]] = ""
	hwlog.RunLog.Infof("pull msgs: %+v", msgs)
	return msgs, nil
}

func (s *preRecoverPlugin) getSignalInfo() error {
	if s.shot.ClusterInfos == nil {
		return errors.New("no cluster info")
	}
	clusterInfo, err := s.shot.ClusterInfos.GetCluster(constant.ClusterDRank)
	if err != nil {
		hwlog.RunLog.Debugf("Get clusterD info failed: %s", err.Error())
		return err
	}
	s.signalInfo, err = pluginutils.ParseSignalInfo(clusterInfo)
	if err != nil {
		hwlog.RunLog.Errorf("ParseSignalInfo failed: %v", err)
		return err
	}
	return nil
}
