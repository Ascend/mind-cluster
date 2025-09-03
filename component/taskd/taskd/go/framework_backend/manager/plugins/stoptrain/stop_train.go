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

// Package stoptrain for stop train plugin
package stoptrain

import (
	"errors"
	"strconv"

	"ascend-common/common-utils/hwlog"
	clusterdconstant "clusterd/pkg/common/constant"
	"taskd/common/constant"
	"taskd/common/utils"
	"taskd/framework_backend/manager/infrastructure"
	"taskd/framework_backend/manager/infrastructure/storage"
	pluginutils "taskd/framework_backend/manager/plugins/utils"
)

// stopTrainingPlugin stop train plugin define
type stopTrainingPlugin struct {
	hasToken        bool
	shot            storage.SnapShot
	signalInfo      *pluginutils.SignalInfo
	HasSendMessages map[string]string
}

// New creates an object
func New() infrastructure.ManagerPlugin {
	return &stopTrainingPlugin{
		HasSendMessages: make(map[string]string),
	}
}

// Name returns plugin name
func (s *stopTrainingPlugin) Name() string {
	return constant.StopTrainPluginName
}

// Predicate check whether apply token
func (s *stopTrainingPlugin) Predicate(shot storage.SnapShot) (infrastructure.PredicateResult, error) {
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
		hwlog.RunLog.Errorf("getSignalInfo error: %v", err)
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
	return infrastructure.PredicateResult{CandidateStatus: constant.UnselectStatus}, nil
}

// Release releases token
func (s *stopTrainingPlugin) Release() error {
	return nil
}

// Handle handles stream events
func (s *stopTrainingPlugin) Handle() (infrastructure.HandleResult, error) {
	hwlog.RunLog.Infof("plugin[%s] enter handle", s.Name())
	s.hasToken = true
	if s.signalInfo == nil {
		err := s.getSignalInfo()
		if err != nil {
			hwlog.RunLog.Errorf("getSignalInfo error: %v", err)
			return infrastructure.HandleResult{Stage: constant.HandleStageException}, nil
		}
	}
	if s.signalInfo.SignalType == clusterdconstant.GlobalFaultSignalType {
		hwlog.RunLog.Info("get global fault signal, need to release token")
		s.hasToken = false
		s.HasSendMessages = make(map[string]string)
		return infrastructure.HandleResult{Stage: constant.HandleStageFinal}, nil
	}
	return infrastructure.HandleResult{Stage: constant.HandleStageProcess}, nil
}

// PullMsg returns messages to other module
func (s *stopTrainingPlugin) PullMsg() ([]infrastructure.Msg, error) {
	if s.signalInfo == nil {
		hwlog.RunLog.Warn("signalInfo is nil")
		return nil, nil
	}
	if _, ok := s.HasSendMessages[s.signalInfo.Uuid]; ok {
		hwlog.RunLog.Debugf("the signal info has dealed, signal info: %v", s.signalInfo)
		return nil, nil
	}
	msgs := make([]infrastructure.Msg, 0)
	if s.signalInfo.SignalType == clusterdconstant.StopTrainSignalType || s.signalInfo.
		SignalType == clusterdconstant.FaultNodesExitSignalType || s.signalInfo.
		SignalType == clusterdconstant.GlobalFaultSignalType {
		msgs = append(msgs, s.signalInfo.GetMsgs()...)
	}
	s.HasSendMessages[s.signalInfo.Uuid] = ""
	hwlog.RunLog.Infof("pull msgs: %+v", msgs)
	return msgs, nil
}

func (s *stopTrainingPlugin) getSignalInfo() error {
	if s.shot.ClusterInfos == nil {
		return errors.New("no cluster info")
	}
	clusterInfo, err := s.shot.ClusterInfos.GetCluster(constant.ClusterDRank)
	if err != nil {
		hwlog.RunLog.Errorf("Get clusterD info failed: %s", err.Error())
		return err
	}
	if clusterInfo == nil {
		return errors.New("cluster info is nil")
	}
	s.signalInfo = &pluginutils.SignalInfo{
		SignalType:     clusterInfo.Command[constant.SignalType],
		ChangeStrategy: clusterInfo.Command[constant.ChangeStrategy],
		ExtraParams:    clusterInfo.Command[constant.ExtraParams],
		Uuid:           clusterInfo.Command[constant.Uuid],
		Command:        clusterInfo.Command,
	}
	if s.signalInfo.SignalType == "" {
		return nil
	}
	s.signalInfo.Timeout, err = strconv.ParseInt(clusterInfo.Command[constant.Timeout], constant.TenBase, constant.BitSize64)
	if err != nil {
		hwlog.RunLog.Errorf("ParseInt failed: %s", err.Error())
		return err
	}
	s.signalInfo.Actions, err = utils.StringToObj[[]string](clusterInfo.Command[constant.Actions])
	if err != nil {
		hwlog.RunLog.Errorf("unmarshal actions failed: %s", err.Error())
		return err
	}
	s.signalInfo.FaultRanks, err = utils.StringToObj[map[int]int](clusterInfo.Command[constant.FaultRanks])
	if err != nil {
		hwlog.RunLog.Errorf("unmarshal FaultRanks failed: %s", err.Error())
		return err
	}
	s.signalInfo.NodeRankIds, err = utils.StringToObj[[]string](clusterInfo.Command[constant.NodeRankIds])
	if err != nil {
		hwlog.RunLog.Errorf("unmarshal NodeRankIds failed: %s", err.Error())
		return err
	}
	return nil
}
