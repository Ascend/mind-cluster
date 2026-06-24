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
	"time"

	"ascend-common/common-utils/hwlog"
	"taskd/common/constant"
	"taskd/common/utils"
	"taskd/framework_backend/manager/infrastructure"
	"taskd/framework_backend/manager/infrastructure/storage"
)

const retryIntervalProfilingCmd = 30 * time.Second

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

func (s *workerExecStatus) checkStatusMeetCmd() bool {
	if s.cmd.DefaultDomainAble && s.defaultDomainState != constant.ProfilingWorkerOpenedState {
		return false
	}
	if !s.cmd.DefaultDomainAble && s.defaultDomainState != constant.ProfilingWorkerClosedState {
		return false
	}
	if s.cmd.CommDomainAble && s.commDomainState != constant.ProfilingWorkerOpenedState {
		return false
	}
	if !s.cmd.CommDomainAble && s.commDomainState != constant.ProfilingWorkerClosedState {
		return false
	}
	return true
}

// PfPlugin Profiling Plugin
type PfPlugin struct {
	watchFile    atomic.Bool
	shot         storage.SnapShot
	cmd          constant.ProfilingDomainCmd
	workerStatus workerExecStatus
	pullMsg      []infrastructure.Msg
	workerNum    int
	retry        retryState
}

// retryState tracks how long the worker status has been inconsistent with the cmd.
type retryState struct {
	notMeetStart time.Time
	cmdChanged   bool
}

// Name get pluginName
func (p *PfPlugin) Name() string {
	return constant.ProfilingPluginName
}

// Predicate Profiling Plugin whether it can resolve SnapShot
func (p *PfPlugin) Predicate(shot storage.SnapShot) (infrastructure.PredicateResult, error) {
	hwlog.RunLog.Debugf("%s shot: %v", p.Name(), shot)
	p.workerNum = shot.WorkerNum
	p.initWorkerStatusMap(shot)
	cmd, errCmd := p.getProfilingCmd(shot)
	p.handleProfilingResult(shot)

	p.retry.update(errCmd == nil, p.workerStatus)
	hwlog.RunLog.Debugf("cmd: %v, workerStatus: %v, retry: %v, errCmd: %v", cmd, p.workerStatus, p.retry, errCmd)

	if errCmd != nil && !p.retry.exceedLimit() {
		hwlog.RunLog.Infof("%s Predicate failed, errCmd: %v", p.Name(), errCmd)
		return infrastructure.PredicateResult{
			PluginName:      p.Name(),
			CandidateStatus: constant.UnselectStatus,
			PredicateStream: nil,
		}, nil
	}
	hwlog.RunLog.Infof("%s Predicate sucess", p.Name())
	p.shot = shot
	if errCmd == nil {
		p.cmd = cmd
		hwlog.RunLog.Infof("%s checkout cmd %v", p.Name(), cmd)
	}
	return infrastructure.PredicateResult{
		PluginName:      p.Name(),
		CandidateStatus: constant.CandidateStatus,
		PredicateStream: map[string]string{
			constant.ProfilingStream: "",
		},
	}, nil
}

// update tracks how long the worker status has been inconsistent with the cmd.
// When the cmd is unchanged and the worker status does not meet the cmd, the timer starts;
// otherwise it resets.
func (r *retryState) update(cmdChanged bool, status workerExecStatus) {
	r.cmdChanged = cmdChanged
	if cmdChanged {
		r.notMeetStart = time.Time{}
		return
	}
	if !status.checkStatusMeetCmd() {
		if r.notMeetStart.IsZero() {
			r.notMeetStart = time.Now()
		}
		return
	}
	r.notMeetStart = time.Time{}
}

// exceedLimit reports whether the retry interval has been exceeded since the worker status
// became inconsistent with the cmd.
func (r *retryState) exceedLimit() bool {
	if r.notMeetStart.IsZero() {
		return false
	}
	return time.Since(r.notMeetStart) > retryIntervalProfilingCmd
}

// shouldPullMsg reports whether the profiling cmd should be (re)sent to workers.
// It is true when the cmd has just changed, or when the retry interval has been exceeded.
func (r *retryState) shouldPullMsg() bool {
	return r.cmdChanged || r.exceedLimit()
}

// markPulled records that the profiling cmd has been sent to workers,
// resetting the retry timer for the next potential retry cycle.
func (r *retryState) markPulled() {
	r.notMeetStart = time.Now()
	r.cmdChanged = false
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
		return p.workerStatus.cmd, fmt.Errorf("get domain cmd fail")
	}
	newCmd, err := utils.ParseProfilingDomainCmd(defaultDomainCmd, commDomainCmd)
	if err != nil {
		return p.workerStatus.cmd, err
	}
	if newCmd == p.workerStatus.cmd {
		return p.workerStatus.cmd, fmt.Errorf("get domain cmd is equal to last cmd")
	}
	return newCmd, nil
}

func (p *PfPlugin) handleProfilingResult(shot storage.SnapShot) {
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
	p.handleWorkerRes(result)
}

// Release do nothing now
func (p *PfPlugin) Release() error {
	return nil
}

// Handle resolve snapshot
func (p *PfPlugin) Handle() (infrastructure.HandleResult, error) {
	p.handleNewCmd()
	return infrastructure.HandleResult{
		Stage: constant.HandleStageFinal,
	}, nil
}

func (p *PfPlugin) handleWorkerRes(res map[string]constant.ProfilingResult) {
	for workerName, workerStatus := range p.workerStatus.workers {
		newStatus := res[workerName]
		if workerStatus != newStatus {
			hwlog.RunLog.Infof("update worker %s profiling status %v to %v", workerName, workerStatus, newStatus)
			p.workerStatus.workers[workerName] = newStatus
		}
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
		if p.changeCmd(p.cmd) {
			p.workerStatus.cmd = p.cmd
		}
	}
}

// PullMsg return Msg
func (p *PfPlugin) PullMsg() ([]infrastructure.Msg, error) {
	switch {
	case p.retry.shouldPullMsg():
		hwlog.RunLog.Infof("shouldPullMsg Profiling PullMsg: %s", utils.ObjToString(p.pullMsg))
		p.retry.markPulled()
		return p.pullMsg, nil
	case p.workerStatus.checkStatusMeetCmd():
		hwlog.RunLog.Info("status meet cmd")
		p.pullMsg = make([]infrastructure.Msg, 0)
		return nil, nil
	default:
		hwlog.RunLog.Info("waiting...")
		return nil, nil
	}
}

// NewProfilingPlugin return New ProfilingPlugin
func NewProfilingPlugin() infrastructure.ManagerPlugin {
	plugin := &PfPlugin{
		watchFile: atomic.Bool{},
		shot:      storage.SnapShot{},
		cmd:       constant.ProfilingDomainCmd{},
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

func (p *PfPlugin) changeCmd(cmd constant.ProfilingDomainCmd) bool {
	p.pullMsg = make([]infrastructure.Msg, 0)
	workers := p.getAllWorkerName()
	hwlog.RunLog.Infof("changeCmd: %v, workers: %v,p.workerNum=%v", cmd, workers, p.workerNum)
	if len(workers) < p.workerNum {
		return false
	}
	p.pullMsg = append(p.pullMsg, infrastructure.Msg{
		Receiver: p.getAllWorkerName(),
		Body: storage.MsgBody{
			MsgType: constant.Action,
			Code:    utils.ProfilingCmdToBizCode(cmd),
		},
	})
	return true
}

func (p *PfPlugin) getAllWorkerName() []string {
	names := make([]string, 0, len(p.workerStatus.workers))
	for name := range p.workerStatus.workers {
		names = append(names, name)
	}
	return names
}
