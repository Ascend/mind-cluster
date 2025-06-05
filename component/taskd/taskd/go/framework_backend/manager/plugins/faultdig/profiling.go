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
package faultdig

import (
	"fmt"
	"sync/atomic"

	"ascend-common/common-utils/hwlog"
	"taskd/common/constant"
	"taskd/common/utils"
	"taskd/framework_backend/manager/infrastructure"
	"taskd/framework_backend/manager/infrastructure/storage"
)

type workerExecStatus struct {
	workers            map[string]constant.ProfilingResult
	cmd                constant.ProfilingDomainCmd
	defaultDomainState constant.ProfilingWorkerState
	commDomainState    constant.ProfilingWorkerState
}

func (s *workerExecStatus) calcNewState() (constant.ProfilingWorkerState, constant.ProfilingWorkerState) {
	defaultDomainResCnt := make(map[constant.ProfilingExecRes]int)
	commDomainResCnt := make(map[constant.ProfilingExecRes]int)
	for _, result := range s.workers {
		defaultDomainResCnt[result.DefaultDomain]++
		commDomainResCnt[result.CommDomain]++
	}
	workerNum := len(s.workers)
	defaultDomainState := s.calcState(defaultDomainResCnt, workerNum, s.cmd.DefaultDomainAble)
	commDomainState := s.calcState(commDomainResCnt, workerNum, s.cmd.CommDomainAble)
	return defaultDomainState, commDomainState
}

func (s *workerExecStatus) calcState(
	domainResCnt map[constant.ProfilingExecRes]int, workerNum int, enable bool) constant.ProfilingWorkerState {
	if domainResCnt[constant.ProfilingExpStatus] != 0 {
		return constant.ProfilingWorkerExceptionState
	}
	if enable {
		if domainResCnt[constant.ProfilingOnStatus] == workerNum {
			return constant.ProfilingWorkerOpenedState
		}
		return constant.ProfilingWorkerWaitOpenState
	}
	if domainResCnt[constant.ProfilingOffStatus] == workerNum {
		return constant.ProfilingWorkerClosedState
	}
	return constant.ProfilingWorkerWaitCloseState
}

// PfPlugin Profiling Plugin
type PfPlugin struct {
	watchFile    atomic.Bool
	shot         storage.SnapShot
	cmd          constant.ProfilingDomainCmd
	report       map[string]constant.ProfilingResult
	workerStatus workerExecStatus
	pullMsg      []infrastructure.Msg
}

// Name get pluginName
func (p *PfPlugin) Name() string {
	return constant.ProfilingPluginName
}

// Predicate Profiling Plugin whether it can resolve SnapShot
func (p *PfPlugin) Predicate(shot storage.SnapShot) (infrastructure.PredicateResult, error) {
	hwlog.RunLog.Debugf("%s shot: %v", p.Name(), shot)
	p.initWorkerStatusMap(shot)
	cmd, errCmd := p.getProfilingCmd(shot)
	res, errRes := p.getProfilingResult(shot)
	if errCmd != nil && errRes != nil {
		hwlog.RunLog.Debugf("%s Predicate failed, errCmd: %v, errRes: %v", p.Name(), errCmd, errRes)
		return infrastructure.PredicateResult{
			PluginName:      p.Name(),
			CandidateStatus: "unselected",
			PredicateStream: nil,
		}, nil
	}
	hwlog.RunLog.Infof("%s Predicate sucess", p.Name())
	p.shot = shot
	if errCmd == nil {
		p.cmd = cmd
		hwlog.RunLog.Infof("%s checkout cmd %v", p.Name(), cmd)
	}
	if errRes == nil {
		p.report = res
		hwlog.RunLog.Infof("%s checkout res %v", p.Name(), res)
	}
	return infrastructure.PredicateResult{
		PluginName:      p.Name(),
		CandidateStatus: constant.CandidateStatus,
		PredicateStream: map[string]string{
			constant.ProfilingStream: "",
		},
	}, nil
}

func (p *PfPlugin) initWorkerStatusMap(shot storage.SnapShot) {
	for workerName, _ := range shot.WorkerInfos.Workers {
		if _, found := p.workerStatus.workers[workerName]; !found {
			p.workerStatus.workers[workerName] = constant.ProfilingResult{
				DefaultDomain: constant.NewProfilingExecRes(constant.Off),
				CommDomain:    constant.NewProfilingExecRes(constant.Off),
			}
		}
	}
}

func (p *PfPlugin) getProfilingCmd(shot storage.SnapShot) (constant.ProfilingDomainCmd, error) {
	var switchOff = constant.ProfilingDomainCmd{
		DefaultDomainAble: false,
		CommDomainAble:    false,
	}
	var defaultDomainCmd = ""
	var commDomainCmd = ""
	// If taskd register clusterD, then get profiling cmd from clusterD
	clusterD, found := shot.ClusterInfos.Clusters[constant.ClusterDRank]
	hwlog.RunLog.Debugf("clusterd: %v", clusterD)
	if found {
		defaultDomainCmd = clusterD.Command[constant.DefaultDomainCmd]
		commDomainCmd = clusterD.Command[constant.CommDomainCmd]
	} else { // If taskd does not register clusterD, then get profiling cmd from taskd manager
		hwlog.RunLog.Debug("cannot find cmd from clusterD, find profiling cmd in taskd manager")
		taskD, found := shot.ClusterInfos.Clusters[constant.TaskDRank]
		if found {
			defaultDomainCmd = taskD.Command[constant.DefaultDomainCmd]
			commDomainCmd = taskD.Command[constant.CommDomainCmd]
		}
	}
	if defaultDomainCmd == "" || commDomainCmd == "" {
		return switchOff, fmt.Errorf("get domain cmd fail")
	}
	newCmd, err := utils.ParseProfilingDomainCmd(defaultDomainCmd, commDomainCmd)
	if err != nil {
		return switchOff, err
	}
	if newCmd == p.cmd {
		return switchOff, fmt.Errorf("get domain cmd is equal to last cmd")
	}
	return newCmd, nil
}

func (p *PfPlugin) getProfilingResult(shot storage.SnapShot) (map[string]constant.ProfilingResult, error) {
	result := make(map[string]constant.ProfilingResult)
	for workerName, workerInfo := range shot.WorkerInfos.Workers {
		defaultDomainStat := workerInfo.Status[constant.DefaultDomainStatus]
		commDomainStat := workerInfo.Status[constant.CommDomainStatus]
		if defaultDomainStat == "" || commDomainStat == "" {
			continue
		}
		defaultDomainRes := constant.NewProfilingExecRes(defaultDomainStat)
		commDomainRes := constant.NewProfilingExecRes(commDomainStat)
		orgWorkerRes := p.workerStatus.workers[workerName]
		if defaultDomainRes != orgWorkerRes.DefaultDomain ||
			commDomainRes != orgWorkerRes.CommDomain {
			result[workerName] = constant.ProfilingResult{
				DefaultDomain: defaultDomainRes,
				CommDomain:    commDomainRes,
			}
		}
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("no worker report profiling exec result")
	}
	return result, nil
}

// Release do nothing now
func (p *PfPlugin) Release() error {
	return nil
}

// Handle resolve snapshot
func (p *PfPlugin) Handle() (infrastructure.HandleResult, error) {
	p.handleWorkerRes()
	p.handleNewCmd()
	return infrastructure.HandleResult{
		Stage: constant.HandleStageFinal,
	}, nil
}

func (p *PfPlugin) handleWorkerRes() {
	for workerName, result := range p.report {
		p.workerStatus.workers[workerName] = result
	}
	defaultDomainState, commDomainState := p.workerStatus.calcNewState()
	if p.workerStatus.defaultDomainState != defaultDomainState ||
		p.workerStatus.commDomainState != commDomainState {
		p.notifyStateChange(defaultDomainState, commDomainState)
	}
}

func (p *PfPlugin) notifyStateChange(
	curDefaultDomainState constant.ProfilingWorkerState, curCommDomainState constant.ProfilingWorkerState) {
	hwlog.RunLog.Infof("pre DefaultDomainState %v, pre CommDomainState %v, "+
		"cur DefaultDomainState %v, cur CommDomainState %v", p.workerStatus.defaultDomainState,
		p.workerStatus.commDomainState, curDefaultDomainState, curCommDomainState)
	p.workerStatus.defaultDomainState = curDefaultDomainState
	p.workerStatus.commDomainState = curCommDomainState
}

func (p *PfPlugin) handleNewCmd() {
	if p.workerStatus.cmd != p.cmd {
		p.workerStatus.cmd = p.cmd
		p.changeCmd(p.cmd)
	}
}

// PullMsg return Msg
func (p *PfPlugin) PullMsg() ([]infrastructure.Msg, error) {
	hwlog.RunLog.Infof("Profiling PullMsg: %s", utils.ObjToString(p.pullMsg))
	res := p.pullMsg
	p.pullMsg = make([]infrastructure.Msg, 0)
	return res, nil
}

// NewProfilingPlugin return New ProfilingPlugin
func NewProfilingPlugin() infrastructure.ManagerPlugin {
	plugin := &PfPlugin{
		watchFile: atomic.Bool{},
		shot:      storage.SnapShot{},
		cmd:       constant.ProfilingDomainCmd{},
		report:    nil,
		workerStatus: workerExecStatus{
			workers: make(map[string]constant.ProfilingResult),
			cmd: constant.ProfilingDomainCmd{
				DefaultDomainAble: false,
				CommDomainAble:    false,
			},
			defaultDomainState: constant.NewWorkerProfilingState(constant.Closed),
			commDomainState:    constant.NewWorkerProfilingState(constant.Closed),
		},
	}
	return plugin
}

func (p *PfPlugin) changeCmd(cmd constant.ProfilingDomainCmd) {
	hwlog.RunLog.Infof("changeCmd: %v", cmd)
	p.pullMsg = make([]infrastructure.Msg, 0)
	p.pullMsg = append(p.pullMsg, infrastructure.Msg{
		Receiver: p.getAllWorkerName(),
		Body: storage.MsgBody{
			MsgType: constant.Action,
			Code:    utils.ProfilingCmdToBizCode(cmd),
		},
	})
}

func (p *PfPlugin) getAllWorkerName() []string {
	names := make([]string, 0, len(p.workerStatus.workers))
	for name := range p.workerStatus.workers {
		names = append(names, name)
	}
	return names
}
