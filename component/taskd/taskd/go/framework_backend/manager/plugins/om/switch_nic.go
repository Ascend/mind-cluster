// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package om a series of service function
package om

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/interface/grpc/recover"
	"taskd/common/constant"
	"taskd/common/utils"
	"taskd/framework_backend/manager/infrastructure"
	"taskd/framework_backend/manager/infrastructure/storage"
)

const (
	maxRegRetryTime  = 60
	firstRetryTIme   = 0
	switchNicTimeout = 120
)

// OmPlugin Profiling Plugin
type OmPlugin struct {
	pullMsg      []infrastructure.Msg
	workerStatus map[string]string
	uuid         string
	jobID        string
	timer        *time.Timer
}

// Name get pluginName
func (o *OmPlugin) Name() string {
	return constant.OMPluginName
}

// Predicate return the stream request
func (o *OmPlugin) Predicate(shot storage.SnapShot) (infrastructure.PredicateResult, error) {
	clusterInfo, ok := shot.ClusterInfos.Clusters[constant.ClusterDRank]
	if !ok {
		return infrastructure.PredicateResult{PluginName: o.Name(), CandidateStatus: constant.UnselectStatus, PredicateStream: nil}, nil
	}
	jobID := clusterInfo.Command[constant.SwitchJobID]
	ranks := clusterInfo.Command[constant.GlobalRankKey]
	ops := clusterInfo.Command[constant.GlobalOpKey]
	uuid := clusterInfo.Command[constant.SwitchNicUUID]
	if jobID == "" || ranks == "" || ops == "" || uuid == "" {
		return infrastructure.PredicateResult{PluginName: o.Name(), CandidateStatus: constant.UnselectStatus, PredicateStream: nil}, nil
	}

	// switching nic
	if uuid == o.uuid && len(o.workerStatus) != 0 {
		o.updateWorkerStatus(shot)
		return infrastructure.PredicateResult{
			PluginName: o.Name(), CandidateStatus: constant.CandidateStatus, PredicateStream: map[string]string{
				constant.OMStreamName: ""}}, nil
	}
	// accept new switch nic
	if uuid != o.uuid {
		o.initPluginStatus(shot)
		return infrastructure.PredicateResult{
			PluginName: o.Name(), CandidateStatus: constant.CandidateStatus, PredicateStream: map[string]string{
				constant.OMStreamName: ""}}, nil
	}
	// waiting new switch nic
	return infrastructure.PredicateResult{
		PluginName: o.Name(), CandidateStatus: constant.UnselectStatus, PredicateStream: nil}, nil
}

// Release give up token in a stream
func (o *OmPlugin) Release() error {
	return nil
}

// Handle business process
func (o *OmPlugin) Handle() (infrastructure.HandleResult, error) {
	if len(o.workerStatus) == 0 {
		hwlog.RunLog.Error("worker status is empty")
		o.replyToClusterD(firstRetryTIme, false)
		o.resetPluginStatus()
		return infrastructure.HandleResult{
			Stage: constant.HandleStageFinal,
		}, nil
	}

	num := 0
	for workerName, status := range o.workerStatus {
		if status == constant.SwitchFail {
			hwlog.RunLog.Infof("rank %s switch failed", workerName)
			o.replyToClusterD(firstRetryTIme, false)
			o.resetPluginStatus()
			return infrastructure.HandleResult{
				Stage: constant.HandleStageFinal,
			}, nil
		}
		if status == constant.SwitchOK {
			num += 1
			hwlog.RunLog.Debugf("rank %s switch ok", workerName)
		}
	}
	if num == len(o.workerStatus) {
		hwlog.RunLog.Infof("all rank switch success")
		o.replyToClusterD(firstRetryTIme, true)
		o.resetPluginStatus()
		return infrastructure.HandleResult{
			Stage: constant.HandleStageFinal,
		}, nil
	}
	return infrastructure.HandleResult{
		Stage: constant.HandleStageProcess,
	}, nil
}

func (o *OmPlugin) replyToClusterD(retryTime time.Duration, result bool) {
	if retryTime >= maxRegRetryTime {
		hwlog.RunLog.Error("init clusterd connect meet max retry time")
		return
	}
	time.Sleep(retryTime * time.Second)
	addr, err := utils.GetClusterdAddr()
	if err != nil {
		hwlog.RunLog.Errorf("get clusterd address err: %v", err)
		return
	}
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		hwlog.RunLog.Errorf("init clusterd connect err: %v", err)
		o.replyToClusterD(retryTime+1, result)
		return
	}
	client := pb.NewRecoverClient(conn)
	_, err = client.ReplySwitchNicResult(context.TODO(), &pb.SwitchResult{Result: result, JobId: o.jobID})
	if err != nil {
		hwlog.RunLog.Errorf("reply SwitchNicResult err: %v", err)
	}
}

// PullMsg return Msg
func (o *OmPlugin) PullMsg() ([]infrastructure.Msg, error) {
	res := o.pullMsg
	o.pullMsg = make([]infrastructure.Msg, 0)
	return res, nil
}

// NewOmPlugin return New ProfilingPlugin
func NewOmPlugin() infrastructure.ManagerPlugin {
	plugin := &OmPlugin{
		pullMsg:      make([]infrastructure.Msg, 0),
		uuid:         "",
		jobID:        "",
		workerStatus: make(map[string]string),
	}
	return plugin
}

func (o *OmPlugin) getAllWorkerName() []string {
	names := make([]string, 0, len(o.workerStatus))
	for name, _ := range o.workerStatus {
		names = append(names, name)
	}
	return names
}

func (o *OmPlugin) updateWorkerStatus(shot storage.SnapShot) {
	for name, info := range shot.WorkerInfos.Workers {
		if info.Status[constant.SwitchNicUUID] != o.uuid {
			continue
		}
		o.workerStatus[name] = info.Status[constant.SwitchNic]
	}
	hwlog.RunLog.Debugf("update worker status: %v", o.workerStatus)
}

func (o *OmPlugin) resetPluginStatus() {
	o.workerStatus = make(map[string]string)
	if o.timer != nil {
		o.timer.Stop()
	}
	o.timer = nil

}

func (o *OmPlugin) initPluginStatus(shot storage.SnapShot) {
	for workerName, _ := range shot.WorkerInfos.Workers {
		o.workerStatus[workerName] = ""
	}
	clusterInfo := shot.ClusterInfos.Clusters[constant.ClusterDRank]
	o.uuid = clusterInfo.Command[constant.SwitchNicUUID]
	o.jobID = clusterInfo.Command[constant.SwitchJobID]
	o.pullMsg = append(o.pullMsg, infrastructure.Msg{
		Receiver: o.getAllWorkerName(),
		Body: storage.MsgBody{
			MsgType: constant.SwitchNic,
			Extension: map[string]string{
				constant.GlobalRankKey: clusterInfo.Command[constant.GlobalRankKey],
				constant.GlobalOpKey:   clusterInfo.Command[constant.GlobalOpKey],
				constant.SwitchNicUUID: clusterInfo.Command[constant.SwitchNicUUID],
			},
		},
	})
	if o.timer != nil {
		o.timer.Stop()
	}
	o.timer = time.AfterFunc(switchNicTimeout*time.Minute, func() {
		hwlog.RunLog.Warn("wait switch timeout, reset plugin status")
		o.replyToClusterD(firstRetryTIme, false)
		o.resetPluginStatus()
	})
	hwlog.RunLog.Infof("recv new option, workerstate: %v, jobID: %v, uuid:%v", o.workerStatus, o.jobID, o.uuid)
	hwlog.RunLog.Infof("Switch PullMsg: %s", utils.ObjToString(o.pullMsg))
}
