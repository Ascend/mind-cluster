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

// Package hotswitch for taskd manager plugin
package hotswitch

import (
	"errors"

	"ascend-common/common-utils/hwlog"
	clusterdconstant "clusterd/pkg/common/constant"
	"taskd/common/constant"
	"taskd/common/utils"
	"taskd/framework_backend/manager/infrastructure"
	"taskd/framework_backend/manager/infrastructure/storage"
)

// HotSwitchPlugin hot switch plugin
type HotSwitchPlugin struct {
	pullMsgs       []infrastructure.Msg
	signalType     string
	changeStrategy string
	faultRanks     map[int]int
	actions        string
	uuid           string
}

// NewHotSwitchPlugin new hot switch plugin
func NewHotSwitchPlugin() infrastructure.ManagerPlugin {
	return &HotSwitchPlugin{
		pullMsgs:       make([]infrastructure.Msg, 0),
		changeStrategy: "",
		faultRanks:     make(map[int]int),
		actions:        "",
		uuid:           "",
	}
}

// Name name of hot switch plugin
func (rp *HotSwitchPlugin) Name() string {
	return constant.HotSwitchPluginName
}

// Predicate predicate hot switch plugin
func (rp *HotSwitchPlugin) Predicate(shot storage.SnapShot) (infrastructure.PredicateResult, error) {
	clusterInfo, ok := shot.ClusterInfos.Clusters[constant.ClusterDRank]
	if !ok {
		hwlog.RunLog.Debug("cluster info not found")
		return infrastructure.PredicateResult{
			PluginName: rp.Name(), CandidateStatus: constant.UnselectStatus, PredicateStream: nil}, nil
	}
	signalType := clusterInfo.Command[constant.SignalType]
	actions := clusterInfo.Command[constant.Actions]
	strategy := clusterInfo.Command[constant.ChangeStrategy]

	if signalType != clusterdconstant.HotSwitchSignalType && strategy != clusterdconstant.ProcessMigration {
		return infrastructure.PredicateResult{
			PluginName: rp.Name(), CandidateStatus: constant.UnselectStatus, PredicateStream: nil}, nil
	}
	if signalType == rp.signalType && actions == rp.actions {
		hwlog.RunLog.Debugf("strategy not change,will not send msg, uuid:%v signalType:%v,strategy:%v",
			rp.uuid, rp.signalType, rp.changeStrategy)
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
	rp.changeStrategy = strategy
	rp.signalType = signalType
	return infrastructure.PredicateResult{
		PluginName: rp.Name(), CandidateStatus: constant.CandidateStatus, PredicateStream: map[string]string{
			constant.ResumeTrainingAfterFaultStream: ""}}, nil
}

// Handle handle hot switch plugin
func (rp *HotSwitchPlugin) Handle() (infrastructure.HandleResult, error) {
	if rp.signalType == clusterdconstant.HotSwitchSignalType ||
		rp.changeStrategy == clusterdconstant.ProcessMigration {
		rp.buildControllerMessage()
	}
	return infrastructure.HandleResult{Stage: constant.HandleStageFinal}, nil
}

// PullMsg pull message
func (rp *HotSwitchPlugin) PullMsg() ([]infrastructure.Msg, error) {
	msgs := rp.pullMsgs
	rp.pullMsgs = make([]infrastructure.Msg, 0)
	return msgs, nil
}

// Release release hot switch plugin
func (rp *HotSwitchPlugin) Release() error {
	return nil
}

func (rp *HotSwitchPlugin) resetPluginInfo() {
	rp.faultRanks = make(map[int]int)
	rp.actions = ""
	rp.signalType = ""
	rp.changeStrategy = ""
}

func (rp *HotSwitchPlugin) getClusterInfo(shot storage.SnapShot) error {
	clusterInfo, ok := shot.ClusterInfos.Clusters[constant.ClusterDRank]
	if !ok {
		return errors.New("cluster info not found")
	}
	faultRanks, err := utils.StringToObj[map[int]int](clusterInfo.Command[constant.FaultRanks])
	if err != nil {
		return err
	}
	rp.faultRanks = faultRanks
	rp.actions = clusterInfo.Command[constant.Actions]
	rp.uuid = clusterInfo.Command[constant.Uuid]
	return nil
}

func (rp *HotSwitchPlugin) buildControllerMessage() {
	rp.pullMsgs = append(rp.pullMsgs, infrastructure.Msg{
		Receiver: []string{constant.ControllerName},
		Body: storage.MsgBody{
			MsgType: constant.Action,
			Code:    constant.HotSwitchCode,
			Extension: map[string]string{
				constant.Actions:        rp.actions,
				constant.ChangeStrategy: rp.changeStrategy,
				constant.FaultRanks:     utils.ObjToString(rp.faultRanks),
			},
		},
	})
	hwlog.RunLog.Infof("build hotswitch message, rp.pullMsgs: %v", rp.pullMsgs)
}
