// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package recover a series of service function
// om is operation manager
package recover

import (
	"context"
	"fmt"
	"slices"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/application/faultmanager"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/common"
	"clusterd/pkg/domain/job"
	"clusterd/pkg/interface/grpc/recover"
)

var (
	// 0: aic, 1: p2p
	stressOps = []int64{0, 1}
)

// SwitchNicTrack Switch the track of the specified nic
func (s *FaultRecoverService) SwitchNicTrack(ctx context.Context, nics *pb.SwitchNics) (*pb.Status, error) {
	if ok, msg := s.checkNicsParam(nics); !ok {
		hwlog.RunLog.Errorf("check param failed: %s", msg)
		return &pb.Status{
			Code: int32(common.OMParamInvalid),
			Info: msg,
		}, nil
	}
	ctl, _ := s.getController(nics.JobID)
	if !ctl.canDoSwitchingNic() {
		hwlog.RunLog.Errorf("jobId=%s nic is swtiching, or job recovering", nics.JobID)
		return &pb.Status{
			Code: int32(common.OMIsRunning),
			Info: fmt.Sprintf("jobId=%s nic is swtiching, or job recovering", nics.JobID),
		}, nil
	}
	globalSwitchRankIDs, globalOps := s.getGlobalRankIDAndOp(nics)
	ctl.setSwitchNicParam(globalSwitchRankIDs, globalOps)
	ctl.addEvent(common.StartSwitchNic)
	hwlog.RunLog.Infof("jobId=%s nic swtich: %v, %v", nics.JobID, globalSwitchRankIDs, globalOps)
	return &pb.Status{Code: int32(common.OK), Info: "switching operation was successfully distributed"}, nil
}

// SubscribeSwitchNicSignal return the result of switch nic
func (s *FaultRecoverService) SubscribeSwitchNicSignal(req *pb.SwitchNicRequest,
	stream pb.Recover_SubscribeSwitchNicSignalServer) error {
	if req == nil {
		return fmt.Errorf("request is nil")
	}
	hwlog.RunLog.Infof("receive Subscribe signal request, jobID: %s", req.JobID)
	controller, exist := s.getController(req.JobID)
	if !exist {
		return fmt.Errorf("jobId=%s not registed", req.JobID)
	}
	if !controller.isSwitchingNic() {
		return fmt.Errorf("jobId=%s is not swtiching nic", req.JobID)
	}
	controller.listenSwitchNicChannel(stream)
	return nil
}

// SubscribeNotifySwitch notify worker switch nic
func (s *FaultRecoverService) SubscribeNotifySwitch(req *pb.ClientInfo, stream pb.Recover_SubscribeNotifySwitchServer) error {
	if req == nil {
		return fmt.Errorf("request is nil")
	}
	controller, exist := s.getController(req.JobId)
	if !exist {
		hwlog.RunLog.Debugf("jobId=%s not registed, wait job running", req.JobId)
		return fmt.Errorf("jobId=%s not registed, please wait agent register", req.JobId)
	}
	hwlog.RunLog.Infof("receive Subscribe notify switch signal request, jobID: %s", req.JobId)
	controller.listenSwitchNicNotifyChannel(stream)
	return nil
}

// ReplySwitchNicResult reply worker switch nic result
func (s *FaultRecoverService) ReplySwitchNicResult(ctx context.Context, res *pb.SwitchResult) (*pb.Status, error) {
	if res == nil {
		return &pb.Status{Code: int32(common.OMParamInvalid), Info: "request is nil"}, nil
	}
	controller, exist := s.getController(res.JobId)
	if !exist {
		hwlog.RunLog.Errorf("jobId=%s not registed", res.JobId)
		return &pb.Status{
			Code: int32(common.UnRegistry),
			Info: fmt.Sprintf("jobId=%s not registed", res.JobId)}, nil
	}
	controller.setSwitchNicResult(res)
	hwlog.RunLog.Infof("jobId=%s nic swtich result: %v", res.JobId, res.Result)
	return &pb.Status{Code: int32(common.OK), Info: "reply success"}, nil
}

func (s *FaultRecoverService) checkNicsParam(nics *pb.SwitchNics) (bool, string) {
	if nics == nil {
		return false, "nics is nil"
	}
	jobInfo, ok := job.GetJobCache(nics.JobID)
	if !ok {
		return false, fmt.Sprintf("job:%s not exist", nics.JobID)
	}
	if !s.registered(nics.JobID) {
		return false, fmt.Sprintf("job:%s is not registered", nics.JobID)
	}
	if jobInfo.Status != job.StatusJobRunning {
		return false, fmt.Sprintf("job:%s is not running", nics.JobID)
	}
	jobServerMap := s.getNodeDeviceMap(jobInfo.JobRankTable.ServerList)
	deviceInfos := faultmanager.QueryDeviceInfoToReport()
	for node, devs := range nics.NicOps {
		if _, ok := jobServerMap[node]; !ok {
			return false, fmt.Sprintf("node:%s not exist in job:%s", node, nics.JobID)
		}
		if _, ok := deviceInfos[node]; !ok {
			return false, fmt.Sprintf("node:%s not exist in job:%s", node, nics.JobID)
		}
		if msg, ok := s.checkDevsValid(devs, jobServerMap[node], node); !ok {
			return false, msg
		}
		if deviceInfos[node].SuperPodID < 0 {
			return false, fmt.Sprintf("node:%s should operate in superPodID", node)
		}
	}
	return true, ""
}

func (s *FaultRecoverService) checkDevsValid(switchDev *pb.DeviceList, devs []string, node string) (string, bool) {
	if len(switchDev.Dev) != len(switchDev.Op) {
		return "dev and op length is not equal", false
	}
	if len(switchDev.Dev) == 0 || len(switchDev.Op) == 0 {
		return "dev or op is empty", false
	}
	for _, dev := range switchDev.Dev {
		if !slices.Contains(devs, dev) {
			return fmt.Sprintf("device:%s not exist in node:%s", dev, node), false
		}
	}
	return "", true
}

func (s *FaultRecoverService) getNodeDeviceMap(serverList []constant.ServerHccl) map[string][]string {
	serverMap := make(map[string][]string)
	for _, server := range serverList {
		devs := make([]string, 0)
		for _, dev := range server.DeviceList {
			devs = append(devs, dev.DeviceID)
		}
		serverMap[server.ServerName] = devs
	}
	return serverMap
}

func (s *FaultRecoverService) getGlobalRankIDAndOp(nics *pb.SwitchNics) ([]string, []bool) {
	globalRankIDs := make([]string, 0)
	globalOps := make([]bool, 0)
	jobInfo, ok := job.GetJobCache(nics.JobID)
	if !ok {
		hwlog.RunLog.Errorf("get job cache failed, jobId=%s", nics.JobID)
		return globalRankIDs, globalOps
	}
	serverMap := make(map[string]map[string]string)
	for _, server := range jobInfo.JobRankTable.ServerList {
		devs := make(map[string]string)
		for _, dev := range server.DeviceList {
			devs[dev.DeviceID] = dev.RankID
		}
		serverMap[server.ServerName] = devs
	}

	for node, devs := range nics.NicOps {
		for _, dev := range devs.Dev {
			globalRankIDs = append(globalRankIDs, serverMap[node][dev])
		}
		for _, op := range devs.Op {
			globalOps = append(globalOps, op)
		}
	}
	return globalRankIDs, globalOps
}

// StressTest stress test of the specified node
func (s *FaultRecoverService) StressTest(ctx context.Context, params *pb.StressTestParam) (*pb.Status, error) {
	if ok, msg := s.checkStressTestParam(params); !ok {
		hwlog.RunLog.Errorf("check param failed: %s", msg)
		return &pb.Status{
			Code: int32(common.OMParamInvalid),
			Info: msg,
		}, nil
	}
	ctl, _ := s.getController(params.JobID)
	if !ctl.canDoStressTest() {
		hwlog.RunLog.Errorf("jobId=%s om is running, or job recovering", params.JobID)
		return &pb.Status{
			Code: int32(common.OMIsRunning),
			Info: fmt.Sprintf("jobId=%s om is running, or job recovering", params.JobID),
		}, nil
	}
	globalRankIDs := s.getNodeRankOpsMap(params)
	ctl.setStressTestParam(globalRankIDs)
	ctl.addEvent(common.StartStressTest)
	hwlog.RunLog.Infof("jobId=%s stress test param: %v, global ranks: %v", params.JobID, params, globalRankIDs)
	return &pb.Status{Code: int32(common.OK), Info: "stress test operation was successfully distributed"}, nil
}

// SubscribeNotifyExecStressTest notify worker stress test
func (s *FaultRecoverService) SubscribeNotifyExecStressTest(req *pb.ClientInfo, stream pb.Recover_SubscribeNotifyExecStressTestServer) error {
	if req == nil {
		return fmt.Errorf("request is nil")
	}
	controller, exist := s.getController(req.JobId)
	if !exist {
		hwlog.RunLog.Debugf("jobId=%s not registed, wait job running", req.JobId)
		return fmt.Errorf("jobId=%s not registed, please wait agent register", req.JobId)
	}
	hwlog.RunLog.Infof("receive Subscribe notify stress test signal request, jobID: %s", req.JobId)
	controller.listenStressTestNotifyChannel(stream)
	return nil
}

// ReplyStressTestResult reply worker stress test result
func (s *FaultRecoverService) ReplyStressTestResult(ctx context.Context, res *pb.StressTestResult) (*pb.Status, error) {
	if res == nil {
		return &pb.Status{Code: int32(common.OMParamInvalid), Info: "request is nil"}, nil
	}
	controller, exist := s.getController(res.JobId)
	if !exist {
		hwlog.RunLog.Errorf("jobId=%s not registed", res.JobId)
		return &pb.Status{
			Code: int32(common.UnRegistry),
			Info: fmt.Sprintf("jobId=%s not registed", res.JobId)}, nil
	}
	controller.setStressTestResult(res)
	hwlog.RunLog.Infof("jobId=%s stress test result: %v", res.JobId, res.StressResult)
	return &pb.Status{Code: int32(common.OK), Info: "reply success"}, nil
}

// SubscribeStressTestResponse return the result of stress test
func (s *FaultRecoverService) SubscribeStressTestResponse(req *pb.StressTestRequest,
	stream pb.Recover_SubscribeStressTestResponseServer) error {
	if req == nil {
		return fmt.Errorf("request is nil")
	}
	hwlog.RunLog.Infof("receive Subscribe signal request, jobID: %s", req.JobID)
	controller, exist := s.getController(req.JobID)
	if !exist {
		return fmt.Errorf("jobId=%s not registed", req.JobID)
	}
	if !controller.isStressTest() {
		return fmt.Errorf("jobId=%s is not stress test", req.JobID)
	}
	controller.listenStressTestChannel(stream)
	return nil
}

func (s *FaultRecoverService) getNodeRankOpsMap(param *pb.StressTestParam) common.StressTestParam {
	nodeRankOpsMap := make(common.StressTestParam)
	jobInfo, ok := job.GetJobCache(param.JobID)
	if !ok {
		hwlog.RunLog.Errorf("get job cache failed, jobId=%s", param.JobID)
		return nil
	}
	nodeRankMap := s.getNodeRankMap(jobInfo.JobRankTable.ServerList)
	if len(param.AllNodesOps) != 0 {
		for node, ranks := range nodeRankMap {
			nodeRankOpsMap[node] = make(map[string][]int64, len(ranks))
			for _, rank := range ranks {
				nodeRankOpsMap[node][rank] = param.AllNodesOps
			}
		}
		return nodeRankOpsMap
	}
	for nodeName, _ := range param.StressParam {
		nodeRankOpsMap[nodeName] = map[string][]int64{}
		for _, rank := range nodeRankMap[nodeName] {
			nodeRankOpsMap[nodeName][rank] = param.StressParam[nodeName].Ops
		}
	}
	return nodeRankOpsMap
}

func (s *FaultRecoverService) checkStressTestParam(params *pb.StressTestParam) (bool, string) {
	if params == nil {
		return false, "param is nil"
	}
	jobInfo, ok := job.GetJobCache(params.JobID)
	if !ok {
		return false, fmt.Sprintf("job:%s not exist", params.JobID)
	}
	if !s.registered(params.JobID) {
		return false, fmt.Sprintf("job:%s is not registered", params.JobID)
	}
	if len(params.AllNodesOps) != 0 {
		if ok, msg := s.validateStressTestOps(params.AllNodesOps, "AllNodes"); !ok {
			return false, msg
		}
		return true, ""
	}
	jobServerMap := s.getNodeRankMap(jobInfo.JobRankTable.ServerList)
	if len(params.StressParam) == 0 {
		return false, "stress test node is nil"
	}
	for node, ops := range params.StressParam {
		if ops == nil {
			return false, fmt.Sprintf("node:%s stress test ops is nil", node)
		}
		if len(ops.Ops) == 0 {
			return false, fmt.Sprintf("node:%s stress test ops is 0", node)
		}
		if _, ok := jobServerMap[node]; !ok {
			return false, fmt.Sprintf("node:%s not exist in job:%s", node, params.JobID)
		}
		if ok, msg := s.validateStressTestOps(ops.Ops, node); !ok {
			return false, msg
		}
	}
	return true, ""
}

func (s *FaultRecoverService) validateStressTestOps(ops []int64, node string) (bool, string) {
	opMap := make(map[int64]struct{})
	for _, op := range ops {
		opMap[op] = struct{}{}
		if !slices.Contains(stressOps, op) {
			return false, fmt.Sprintf("op:%v not exist in support operation:%v", op, stressOps)
		}
	}
	if len(opMap) != len(ops) {
		return false, fmt.Sprintf("node:%s stress test ops should not repeat", node)
	}
	return true, ""
}

func (ctl *EventController) canDoStressTest() bool {
	return ctl.state.GetState() == common.InitState
}

func (s *FaultRecoverService) getNodeRankMap(serverList []constant.ServerHccl) map[string][]string {
	serverMap := make(map[string][]string)
	for _, server := range serverList {
		ranks := make([]string, 0)
		for _, dev := range server.DeviceList {
			ranks = append(ranks, dev.RankID)
		}
		serverMap[server.ServerName] = ranks
	}
	return serverMap
}
