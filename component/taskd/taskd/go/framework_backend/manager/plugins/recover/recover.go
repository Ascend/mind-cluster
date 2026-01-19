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
package recoveplugin

import (
	"errors"
	"strconv"

	"ascend-common/common-utils/hwlog"
	clusterdconstant "clusterd/pkg/common/constant"
	"taskd/common/constant"
	"taskd/common/utils"
	"taskd/framework_backend/manager/infrastructure"
	"taskd/framework_backend/manager/infrastructure/storage"
	"taskd/toolkit_backend/net/common"
)

// RecoverPlugin recover plugin
type RecoverPlugin struct {
	pullMsgs        []infrastructure.Msg
	processStatus   string
	recoverStrategy string
	faultNode       []string
	faultRanks      []string
	actions         []string
	msgSend         bool
	recoverInPlace  bool
	uuid            string
	saveAndExit     bool
	doAction        bool
}

// NewRecoverPlugin new recover plugin
func NewRecoverPlugin() infrastructure.ManagerPlugin {
	return &RecoverPlugin{
		pullMsgs:        make([]infrastructure.Msg, 0),
		processStatus:   "",
		recoverStrategy: "",
		faultNode:       make([]string, 0),
		faultRanks:      make([]string, 0),
		actions:         make([]string, 0),
		msgSend:         false,
		recoverInPlace:  false,
		uuid:            "",
		saveAndExit:     false,
	}
}

// Name name of recover plugin
func (rp *RecoverPlugin) Name() string {
	return constant.RecoverPluginName
}

// Predicate predicate of recover plugin
func (rp *RecoverPlugin) Predicate(shot storage.SnapShot) (infrastructure.PredicateResult, error) {
	if rp.processStatus != "" {
		hwlog.RunLog.Infof("recover plugin process status not empty, status:%v", rp.processStatus)
		return infrastructure.PredicateResult{PluginName: rp.Name(),
			CandidateStatus: constant.CandidateStatus,
			PredicateStream: map[string]string{constant.ResumeTrainingAfterFaultStream: ""}}, nil
	}
	rp.resetPluginInfo()
	clusterInfo, ok := shot.ClusterInfos.Clusters[constant.ClusterDRank]
	if !ok {
		hwlog.RunLog.Debug("cluster info not found")
		return infrastructure.PredicateResult{
			PluginName: rp.Name(), CandidateStatus: constant.UnselectStatus, PredicateStream: nil}, nil
	}
	if rp.checkSaveAndExit(shot) {
		return infrastructure.PredicateResult{
			PluginName: rp.Name(), CandidateStatus: constant.CandidateStatus, PredicateStream: map[string]string{
				constant.ResumeTrainingAfterFaultStream: ""}}, nil
	}
	strategy := clusterInfo.Command[constant.ChangeStrategy]
	if rp.uuid == clusterInfo.Command[constant.Uuid] && strategy == rp.recoverStrategy && !rp.doAction {
		rp.resetPluginInfo()
		hwlog.RunLog.Debugf("recover strategy not change, uuid:%v strategy:%v", rp.uuid, rp.recoverStrategy)
		return infrastructure.PredicateResult{
			PluginName: rp.Name(), CandidateStatus: constant.UnselectStatus, PredicateStream: nil}, nil
	}
	if strategy == clusterdconstant.ProcessRetryStrategyName ||
		strategy == clusterdconstant.ProcessRecoverStrategyName ||
		strategy == clusterdconstant.ProcessDumpStrategyName ||
		strategy == clusterdconstant.ProcessContinueTrain {
		hwlog.RunLog.Infof("recover strategy in recover/retry/dump/continue, strtegy:%v", strategy)
		rp.recoverStrategy = strategy
		rp.doAction = true
	} else {
		hwlog.RunLog.Debugf("recover strategy not in recover/retry/dump/continue, strtegy:%v", strategy)
		return infrastructure.PredicateResult{
			PluginName: rp.Name(), CandidateStatus: constant.UnselectStatus, PredicateStream: nil}, nil
	}
	err := rp.getClusterInfo(shot)
	if err != nil {
		rp.resetPluginInfo()
		hwlog.RunLog.Errorf("getClusterInfo error: %v", err)
		return infrastructure.PredicateResult{
			PluginName: rp.Name(), CandidateStatus: constant.UnselectStatus, PredicateStream: nil}, nil
	}
	hwlog.RunLog.Infof("recover plugin candidate faultranks: %v, nodeids: %v", rp.faultRanks, rp.faultNode)
	return infrastructure.PredicateResult{
		PluginName: rp.Name(), CandidateStatus: constant.CandidateStatus, PredicateStream: map[string]string{
			constant.ResumeTrainingAfterFaultStream: ""}}, nil
}

// Handle handle recover plugin
func (rp *RecoverPlugin) Handle() (infrastructure.HandleResult, error) {
	rp.processStatus = constant.HandleStageProcess
	if rp.saveAndExit {
		rp.buildSaveAndExitMessage()
		rp.doAction = false
		rp.resetPluginInfo()
		return infrastructure.HandleResult{Stage: constant.HandleStageFinal}, nil
	}
	if rp.recoverStrategy == clusterdconstant.ProcessDumpStrategyName ||
		rp.recoverStrategy == clusterdconstant.ProcessRetryStrategyName ||
		rp.recoverStrategy == clusterdconstant.ProcessContinueTrain {
		rp.buildControllerMessage()
		rp.doAction = false
		rp.resetPluginInfo()
		return infrastructure.HandleResult{Stage: constant.HandleStageFinal}, nil
	}
	if rp.msgSend {
		rp.buildControllerMessage()
		rp.doAction = false
		rp.resetPluginInfo()
		return infrastructure.HandleResult{Stage: constant.HandleStageFinal}, nil
	}
	if rp.recoverInPlace {
		rp.buildControllerMessage()
		rp.doAction = false
		for _, node := range rp.faultNode {
			rp.pullMsgs = append(rp.pullMsgs, infrastructure.Msg{
				Receiver: []string{common.AgentRole + node},
				Body: storage.MsgBody{
					MsgType: constant.Action,
					Code:    constant.RestartWorkersCode,
					Message: utils.ObjToString(rp.faultRanks),
				},
			})
		}
		rp.resetPluginInfo()
		return infrastructure.HandleResult{Stage: constant.HandleStageFinal}, nil
	}
	for _, node := range rp.faultNode {
		rp.pullMsgs = append(rp.pullMsgs, infrastructure.Msg{
			Receiver: []string{common.AgentRole + node},
			Body: storage.MsgBody{
				MsgType: constant.Action,
				Code:    constant.ExitAgentCode,
			},
		})
	}
	rp.msgSend = true
	return infrastructure.HandleResult{Stage: constant.HandleStageProcess}, nil
}

func (rp *RecoverPlugin) buildControllerMessage() {
	rp.pullMsgs = append(rp.pullMsgs, infrastructure.Msg{
		Receiver: []string{constant.ControllerName},
		Body: storage.MsgBody{
			MsgType: constant.Action,
			Code:    constant.ProcessManageRecoverSignal,
			Extension: map[string]string{
				constant.Actions:        utils.ObjToString(rp.actions),
				constant.ChangeStrategy: rp.recoverStrategy,
			},
		},
	})
	hwlog.RunLog.Infof("build controller message, rp.pullMsgs: %v", rp.pullMsgs)
}

// Handle handle recover plugin
func (rp *RecoverPlugin) PullMsg() ([]infrastructure.Msg, error) {
	msgs := rp.pullMsgs
	rp.pullMsgs = make([]infrastructure.Msg, 0)
	return msgs, nil
}

// Release release recover plugin
func (pod *RecoverPlugin) Release() error {
	return nil
}

func (rp *RecoverPlugin) resetPluginInfo() {
	rp.processStatus = ""
	rp.faultNode = make([]string, 0)
	rp.faultRanks = make([]string, 0)
	rp.actions = make([]string, 0)
	rp.msgSend = false
	rp.recoverInPlace = false
	rp.saveAndExit = false
}

func (rp *RecoverPlugin) getClusterInfo(shot storage.SnapShot) error {
	clusterInfo, ok := shot.ClusterInfos.Clusters[constant.ClusterDRank]
	if !ok {
		return errors.New("cluster info not found")
	}
	faultRanks, err := utils.StringToObj[map[int]int](clusterInfo.Command[constant.FaultRanks])
	if err != nil {
		return err
	}
	for rankId, _ := range faultRanks {
		rp.faultRanks = append(rp.faultRanks, strconv.Itoa(rankId))
	}
	nodeId, err := utils.StringToObj[[]string](clusterInfo.Command[constant.NodeRankIds])
	if err != nil {
		return err
	}
	for _, node := range nodeId {
		rp.faultNode = append(rp.faultNode, node)
	}
	actions, err := utils.StringToObj[[]string](clusterInfo.Command[constant.Actions])
	if err != nil {
		return err
	}
	rp.actions = actions
	rp.recoverInPlace = clusterInfo.Command[constant.ExtraParams] == clusterdconstant.ProcessRecoverInPlaceStrategyName
	hwlog.RunLog.Infof("recover in place: %v, rp.actions = %v", rp.recoverInPlace, actions)
	rp.uuid = clusterInfo.Command[constant.Uuid]
	return nil
}

func (rp *RecoverPlugin) checkSaveAndExit(shot storage.SnapShot) bool {
	clusterInfo, ok := shot.ClusterInfos.Clusters[constant.ClusterDRank]
	if !ok {
		hwlog.RunLog.Info("cluster info not found")
		return false
	}
	if clusterInfo.Command[constant.SignalType] == clusterdconstant.SaveAndExitSignalType &&
		rp.uuid != clusterInfo.Command[constant.Uuid] {
		hwlog.RunLog.Infof("receiver save and exit signal, uuid:%v", clusterInfo.Command[constant.Uuid])
		rp.uuid = clusterInfo.Command[constant.Uuid]
		rp.saveAndExit = true
		nodeId, err := utils.StringToObj[[]string](clusterInfo.Command[constant.NodeRankIds])
		if err != nil {
			hwlog.RunLog.Errorf("checkSaveAndExit err:%v", err)
			return false
		}
		for _, node := range nodeId {
			rp.faultNode = append(rp.faultNode, node)
		}
		return true
	}
	return false
}

func (rp *RecoverPlugin) buildSaveAndExitMessage() {
	actions := []string{constant.SaveAndExit}
	rp.pullMsgs = append(rp.pullMsgs, infrastructure.Msg{
		Receiver: []string{constant.ControllerName},
		Body: storage.MsgBody{
			MsgType: constant.Action,
			Code:    0,
			Extension: map[string]string{
				constant.Actions: utils.ObjToString(actions),
			},
		},
	})
	for _, node := range rp.faultNode {
		rp.pullMsgs = append(rp.pullMsgs, infrastructure.Msg{
			Receiver: []string{common.AgentRole + node},
			Body: storage.MsgBody{
				MsgType: constant.Action,
				Code:    constant.ExitAgentCode,
			},
		})
	}
}
