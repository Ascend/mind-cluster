// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package om a series of service function
package om

import (
	"context"
	"strconv"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/interface/grpc/recover"
	"taskd/common/constant"
	"taskd/common/utils"
	"taskd/framework_backend/manager/infrastructure"
	"taskd/framework_backend/manager/infrastructure/storage"
	"taskd/toolkit_backend/net/common"
)

// StressTestPlugin StressTest Plugin
type StressTestPlugin struct {
	pullMsg      []infrastructure.Msg
	workerStatus map[string]*pb.StressTestRankResult
	heartbeat    map[string]heartbeatInfo
	rankOpMap    map[string]*pb.StressOpList
	uuid         string
	jobID        string
	timer        *time.Timer
}

type heartbeatInfo struct {
	heartbeat int64
	dropTime  int64
}

const maxHeartbeatInterval = 30

// Name get pluginName
func (o *StressTestPlugin) Name() string {
	return constant.OMStressTestPluginName
}

// Predicate return the stream request
func (o *StressTestPlugin) Predicate(shot storage.SnapShot) (infrastructure.PredicateResult, error) {
	clusterInfo, ok := shot.ClusterInfos.Clusters[constant.ClusterDRank]
	if !ok {
		return infrastructure.PredicateResult{PluginName: o.Name(), CandidateStatus: constant.UnselectStatus, PredicateStream: nil}, nil
	}
	jobID := clusterInfo.Command[constant.StressTestJobID]
	rankOps := clusterInfo.Command[constant.StressTestRankOPStr]
	uuid := clusterInfo.Command[constant.StressTestUUID]
	if jobID == "" || rankOps == "" || uuid == "" {
		return infrastructure.PredicateResult{PluginName: o.Name(), CandidateStatus: constant.UnselectStatus, PredicateStream: nil}, nil
	}

	// stressing
	if uuid == o.uuid && len(o.workerStatus) != 0 {
		o.updateWorkerStatus(shot)
		return infrastructure.PredicateResult{
			PluginName: o.Name(), CandidateStatus: constant.CandidateStatus, PredicateStream: map[string]string{
				constant.OMStressTestStreamName: ""}}, nil
	}
	// accept new stress
	if uuid != o.uuid {
		o.initPluginStatus(shot)
		return infrastructure.PredicateResult{
			PluginName: o.Name(), CandidateStatus: constant.CandidateStatus, PredicateStream: map[string]string{
				constant.OMStressTestStreamName: ""}}, nil
	}
	// waiting new stress nic
	return infrastructure.PredicateResult{
		PluginName: o.Name(), CandidateStatus: constant.UnselectStatus, PredicateStream: nil}, nil
}

// Release give up token in a stream
func (o *StressTestPlugin) Release() error {
	return nil
}

// Handle business process
func (o *StressTestPlugin) Handle() (infrastructure.HandleResult, error) {
	if len(o.workerStatus) == 0 {
		hwlog.RunLog.Error("worker status is empty")
		o.replyToClusterD(firstRetryTIme, o.workerStatus)
		o.resetPluginStatus()
		return infrastructure.HandleResult{Stage: constant.HandleStageFinal}, nil
	}
	num := 0
	for workerName, result := range o.workerStatus {
		if result != nil && result.RankResult != nil && len(result.RankResult) != 0 {
			num += 1
			hwlog.RunLog.Debugf("rank %s stress test finish", workerName)
		}
	}
	if num == len(o.workerStatus) {
		hwlog.RunLog.Infof("all stress test finish: %v", o.workerStatus)
		o.replyToClusterD(firstRetryTIme, o.workerStatus)
		o.resetPluginStatus()
		return infrastructure.HandleResult{Stage: constant.HandleStageFinal}, nil
	}
	return infrastructure.HandleResult{Stage: constant.HandleStageProcess}, nil
}

func (o *StressTestPlugin) replyToClusterD(retryTime time.Duration, result map[string]*pb.StressTestRankResult) {
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
	_, err = client.ReplyStressTestResult(context.TODO(), &pb.StressTestResult{StressResult: result, JobId: o.jobID})
	if err != nil {
		hwlog.RunLog.Errorf("reply SwitchNicResult err: %v", err)
	}
}

// PullMsg return Msg
func (o *StressTestPlugin) PullMsg() ([]infrastructure.Msg, error) {
	res := o.pullMsg
	o.pullMsg = make([]infrastructure.Msg, 0)
	return res, nil
}

// NewOmStressTestPlugin return New stressTestPlugin
func NewOmStressTestPlugin() infrastructure.ManagerPlugin {
	plugin := &StressTestPlugin{
		pullMsg:      make([]infrastructure.Msg, 0),
		uuid:         "",
		jobID:        "",
		workerStatus: make(map[string]*pb.StressTestRankResult),
		heartbeat:    make(map[string]heartbeatInfo),
		rankOpMap:    make(map[string]*pb.StressOpList),
	}
	return plugin
}

func (o *StressTestPlugin) getWorkerName() []string {
	names := make([]string, 0, len(o.workerStatus))
	for name, _ := range o.workerStatus {
		names = append(names, common.WorkerRole+name)
	}
	return names
}

func (o *StressTestPlugin) handleWorkerHeartbeat(name string, info *storage.WorkerInfo) bool {
	heartbeat := o.heartbeat[name]
	if info.HeartBeat.Unix() != heartbeat.heartbeat {
		hwlog.RunLog.Debugf("name: worker:%s  heartbeat: %v", name, heartbeat)
		heartbeat.heartbeat = info.HeartBeat.Unix()
		heartbeat.dropTime = 0
		o.heartbeat[name] = heartbeat
		return true
	}
	if info.Status[constant.StressTest] != "" {
		return true
	}
	heartbeat.dropTime += 1
	o.heartbeat[name] = heartbeat
	if heartbeat.dropTime+1 > maxHeartbeatInterval {
		hwlog.RunLog.Errorf("worker %s heartbeat timeout, last heartbeat time: %d", name, heartbeat)
		o.workerStatus[name] = &pb.StressTestRankResult{RankResult: map[string]*pb.StressTestOpResult{}}
		for _, op := range o.rankOpMap[name].Ops {
			o.workerStatus[name].RankResult[strconv.Itoa(int(op))] = &pb.StressTestOpResult{
				Code:   constant.StressTestTimeout,
				Result: "worker heartbeat timeout",
			}
		}
	}
	return false
}

func (o *StressTestPlugin) updateWorkerStatus(shot storage.SnapShot) {
	for name, info := range shot.WorkerInfos.Workers {
		name = strings.TrimPrefix(name, common.WorkerRole)
		if st, ok := o.workerStatus[name]; !ok || len(st.RankResult) != 0 {
			continue
		}
		if !o.handleWorkerHeartbeat(name, info) {
			continue
		}
		if info.Status[constant.StressTestUUID] != o.uuid {
			continue
		}
		rankResultStr := info.Status[constant.StressTest]
		if rankResultStr == "" {
			continue
		}
		rankResult, err := utils.StringToObj[*pb.StressTestRankResult](rankResultStr)
		if err != nil {
			hwlog.RunLog.Errorf("failed to unmarshal err: %v, rankResultStr: %s", err, rankResultStr)
			continue
		}
		o.workerStatus[name] = rankResult
	}
	hwlog.RunLog.Debugf("update worker status: %v, worker heartbeat: %v", o.workerStatus, o.heartbeat)
}

func (o *StressTestPlugin) resetPluginStatus() {
	o.workerStatus = make(map[string]*pb.StressTestRankResult)
	o.heartbeat = make(map[string]heartbeatInfo)
	o.rankOpMap = make(map[string]*pb.StressOpList)
	if o.timer != nil {
		o.timer.Stop()
	}
	o.timer = nil

}

func (o *StressTestPlugin) initPluginStatus(shot storage.SnapShot) {
	clusterInfo := shot.ClusterInfos.Clusters[constant.ClusterDRank]
	rankOps := clusterInfo.Command[constant.StressTestRankOPStr]
	rankOpMap, err := utils.StringToObj[map[string]*pb.StressOpList](rankOps)
	if err != nil {
		o.replyToClusterD(firstRetryTIme, o.workerStatus)
		hwlog.RunLog.Errorf("failed to unmarshal err: %v", err)
		return
	}
	o.rankOpMap = rankOpMap
	for rank, _ := range o.rankOpMap {
		o.workerStatus[rank] = &pb.StressTestRankResult{}
		o.heartbeat[rank] = heartbeatInfo{heartbeat: 0, dropTime: 0}
	}
	o.uuid = clusterInfo.Command[constant.StressTestUUID]
	o.jobID = clusterInfo.Command[constant.StressTestJobID]
	o.pullMsg = append(o.pullMsg, infrastructure.Msg{
		Receiver: o.getWorkerName(),
		Body: storage.MsgBody{
			MsgType: constant.Action,
			Code:    constant.StressTestCode,
			Extension: map[string]string{
				constant.StressTestRankOPStr: clusterInfo.Command[constant.StressTestRankOPStr],
				constant.StressTestUUID:      clusterInfo.Command[constant.StressTestUUID],
			},
		},
	})
	if o.timer != nil {
		o.timer.Stop()
	}
	o.timer = time.AfterFunc(stressTestTimeout*time.Minute, func() {
		hwlog.RunLog.Warn("wait stress test timeout, reset plugin status")
		for workerName, result := range o.workerStatus {
			if len(result.RankResult) != 0 {
				continue
			}
			for _, op := range o.rankOpMap[workerName].Ops {
				o.workerStatus[workerName].RankResult[strconv.Itoa(int(op))] = &pb.StressTestOpResult{
					Code:   constant.StressTestTimeout,
					Result: "worker heartbeat timeout",
				}
			}
		}
		o.replyToClusterD(firstRetryTIme, o.workerStatus)
		o.resetPluginStatus()
	})
	hwlog.RunLog.Infof("recv new option, workerstate: %v, jobID: %v, uuid:%v", o.workerStatus, o.jobID, o.uuid)
	hwlog.RunLog.Infof("Stress test PullMsg: %s", utils.ObjToString(o.pullMsg))
}
