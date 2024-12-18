// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package service a series of service function
package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/application/faultmanager"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/job"
	"clusterd/pkg/domain/podgroup"
	"clusterd/pkg/interface/grpc/common"
	"clusterd/pkg/interface/grpc/pb"
)

var globalFaultBeaconSecond = 2

// FaultRecoverService is a service for fault recover
type FaultRecoverService struct {
	keepAliveInterval int
	serviceCtx        context.Context
	eventCtl          map[string]*EventController
	initJob           map[string]common.JobBaseInfo
	lock              sync.RWMutex
	pb.UnimplementedRecoverServer
}

// NewFaultRecoverService return a new instance of FaultRecoverService
func NewFaultRecoverService(keepAlive int, ctx context.Context) *FaultRecoverService {
	s := &FaultRecoverService{}
	s.keepAliveInterval = keepAlive
	s.serviceCtx = ctx
	s.eventCtl = make(map[string]*EventController)
	s.initJob = make(map[string]common.JobBaseInfo)
	go s.checkFaultFromFaultCenter()
	return s
}

func (s *FaultRecoverService) getController(jobId string) (*EventController, bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	ctl, exist := s.eventCtl[jobId]
	return ctl, exist
}

func (s *FaultRecoverService) notifyFaultInfoForJob(faultInfo faultmanager.JobFaultInfo) {
	controller, exist := s.getController(faultInfo.JobId)
	if !exist || controller == nil {
		hwlog.RunLog.Errorf("jobId=%s not exist", faultInfo.JobId)
		return
	}
	hwlog.RunLog.Infof("get fault info from fault center=%v", faultInfo)
	var grpcFormatFaults []*pb.FaultRank
	for _, info := range faultInfo.FaultList {
		fault := &pb.FaultRank{
			RankId: info.RankId,
		}
		fault.FaultType = constant.NormalFaultType
		if info.DoStepRetry {
			fault.FaultType = constant.UceFaultType
		}
		grpcFormatFaults = append(grpcFormatFaults, fault)
	}
	hwlog.RunLog.Infof("jobId=%s, fault center fault info change format to grpcFormat, faults=%s",
		controller.jobInfo.JobId, common.Faults2String(grpcFormatFaults))
	controller.saveCacheFault(grpcFormatFaults)
	controller.addEvent(common.FaultOccurEvent)
}

func (s *FaultRecoverService) dealWithJobFaultInfo(jobFaultInfoList []faultmanager.JobFaultInfo) {
	wg := sync.WaitGroup{}
	wg.Add(len(jobFaultInfoList))
	for _, jobFaultInfo := range jobFaultInfoList {
		tmpInfo := jobFaultInfo
		go func() {
			defer wg.Done()
			s.notifyFaultInfoForJob(tmpInfo)
		}()
	}
	wg.Wait()
}

func (s *FaultRecoverService) checkFault() {
	if faultmanager.GlobalFaultProcessCenter == nil {
		hwlog.RunLog.Warnf("global center is nil, try it after %d second", globalFaultBeaconSecond)
		return
	}
	allJobFaultInfo := faultmanager.GlobalFaultProcessCenter.QueryJobsFaultInfo(faultmanager.NotHandleFault)
	var registeredJobInfo []faultmanager.JobFaultInfo
	for jobId, jobFaultInfo := range allJobFaultInfo {
		if !s.registered(jobId) {
			continue
		}
		if len(jobFaultInfo.FaultList) <= 0 {
			continue
		}
		registeredJobInfo = append(registeredJobInfo, jobFaultInfo)
	}
	s.dealWithJobFaultInfo(registeredJobInfo)
}

func (s *FaultRecoverService) checkFaultFromFaultCenter() {
	ticker := time.NewTicker(time.Duration(globalFaultBeaconSecond) * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-s.serviceCtx.Done():
			return
		case <-ticker.C:
			hwlog.RunLog.Debug("ticker check npu fault from global center")
			s.checkFault()
		}
	}
}

func (s *FaultRecoverService) recordInit(jobInfo common.JobBaseInfo) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.initJob[jobInfo.JobId] = jobInfo
}

func (s *FaultRecoverService) inited(jobId string) (common.JobBaseInfo, bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	info, ok := s.initJob[jobId]
	return info, ok
}

func getJobBaseInfo(jobId string) (common.JobBaseInfo, common.RespCode, error) {
	jobName, pgName, namespace := podgroup.GetPGFromCacheOrPod(jobId)
	if jobName == "" || pgName == "" || namespace == "" {
		hwlog.RunLog.Errorf("get pg from cache error, jobName=%s, pgName=%s, namespace=%s",
			jobName, pgName, namespace)
		return common.JobBaseInfo{}, common.OperatePodGroupError,
			fmt.Errorf("job(uid=%s) one of jobName, pgName, ns is empty", jobId)
	}
	config, code, err := common.GetRecoverBaseInfo(pgName, namespace)
	if err != nil {
		hwlog.RunLog.Errorf("get recover base info err: %v, pgName=%s, nameSpace=%s",
			err, pgName, namespace)
		return common.JobBaseInfo{}, code,
			fmt.Errorf("get job(uid=%s) base info err:%v", jobId, err)
	}
	if config.ProcessRecoverEnable == false {
		hwlog.RunLog.Errorf("process recover enable does not open, jobId=%s", jobId)
		return common.JobBaseInfo{}, common.ProcessRecoverEnableOff,
			fmt.Errorf("job(uid=%s) process-recover-enable not open:%v", jobId, err)
	}
	return common.JobBaseInfo{
		JobId:         jobId,
		JobName:       jobName,
		PgName:        pgName,
		Namespace:     namespace,
		RecoverConfig: config,
	}, common.OK, nil
}

// Init put process recover enable switch to init state
func (s *FaultRecoverService) Init(ctx context.Context, req *pb.ClientInfo) (*pb.Status, error) {
	reqInfo := fmt.Sprintf("role=%s, jobId=%s", req.Role, req.JobId)
	hwlog.RunLog.Infof("service receive Init request, %s", reqInfo)
	if _, ok := s.inited(req.JobId); ok {
		return &pb.Status{
			Code: int32(common.OK),
			Info: fmt.Sprintf("job(uid=%s) init success", req.JobId),
		}, nil
	}
	baseInfo, code, err := getJobBaseInfo(req.JobId)
	if err != nil {
		return &pb.Status{
			Code: int32(code),
			Info: err.Error(),
		}, nil
	}
	_, err = common.ChangeProcessRecoverEnableMode(baseInfo, constant.ProcessRecoverInit)
	if err != nil {
		hwlog.RunLog.Errorf("change process-recover-enable=init err:%v, jobId=%s", err, req.JobId)
		return &pb.Status{
			Code: int32(common.OperatePodGroupError),
			Info: fmt.Sprintf("job(uid=%s) process-recover-enable init err:%v", req.JobId, err),
		}, nil
	}
	hwlog.RunLog.Infof("job(uid=%s) init success", req.JobId)
	s.recordInit(baseInfo)
	return &pb.Status{
		Code: int32(common.OK),
		Info: fmt.Sprintf("job(uid=%s) init success", req.JobId),
	}, nil
}

func (s *FaultRecoverService) serveJobNum() int {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return len(s.eventCtl)
}

func (s *FaultRecoverService) preRegistry(req *pb.ClientInfo) (common.RespCode, error) {
	if _, ok := job.GetJobCache(req.JobId); !ok {
		return common.JobNotExist, fmt.Errorf("jobId=%s not exist, reuse registry", req.JobId)
	}
	if s.serveJobNum() >= constant.MaxServeJobs {
		return common.OutOfMaxServeJobs,
			fmt.Errorf("out of max serve jobs, reuse registry for jobId=%s", req.JobId)
	}
	return common.OK, nil
}

func (s *FaultRecoverService) registry(jobInfo common.JobBaseInfo) {
	s.lock.Lock()
	defer s.lock.Unlock()
	controller, ok := s.eventCtl[jobInfo.JobId]
	if !ok {
		controller = NewEventController(jobInfo, s.keepAliveInterval, s.serviceCtx)
		s.eventCtl[jobInfo.JobId] = controller
	}
}

func (s *FaultRecoverService) registered(jobId string) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()
	_, ok := s.eventCtl[jobId]
	return ok
}

// Register is task register service
func (s *FaultRecoverService) Register(ctx context.Context, req *pb.ClientInfo) (*pb.Status, error) {
	reqInfo := fmt.Sprintf("role=%s, jobId=%s", req.Role, req.JobId)
	hwlog.RunLog.Infof("service receive Register request, %s", reqInfo)
	if s.registered(req.JobId) {
		return &pb.Status{Code: int32(common.OK), Info: "register success"}, nil
	}
	jobInfo, ok := s.inited(req.JobId)
	if !ok {
		hwlog.RunLog.Errorf("jobId=%s not inited", req.JobId)
		return &pb.Status{
			Code: int32(common.UnInit),
			Info: fmt.Sprintf("jobId=%s not inited", req.JobId),
		}, nil
	}
	code, err := s.preRegistry(req)
	if err != nil {
		hwlog.RunLog.Errorf("jobId=%s, preCheck err:%v", req.JobId, err)
		return &pb.Status{Code: int32(code), Info: err.Error()}, nil
	}
	_, err = common.ChangeProcessRecoverEnableMode(jobInfo, constant.ProcessRecoverEnable)
	if err != nil {
		hwlog.RunLog.Errorf("jobId=%s, change process recover enable to on state err: %v",
			req.JobId, err)
		return &pb.Status{
			Code: int32(common.OperatePodGroupError),
			Info: fmt.Sprintf("jobId=%s register err:%v", req.JobId, err),
		}, nil
	}
	s.registry(jobInfo)
	hwlog.RunLog.Infof("jobId=%s register success", req.JobId)
	return &pb.Status{Code: int32(common.OK), Info: "register success"}, nil
}

// SubscribeProcessManageSignal subscribe process manage signal from ClusterD
func (s *FaultRecoverService) SubscribeProcessManageSignal(request *pb.ClientInfo,
	stream pb.Recover_SubscribeProcessManageSignalServer) error {
	requestInfo := fmt.Sprintf("taskId=%s, rule=%s", request.JobId, request.Role)
	hwlog.RunLog.Infof("receive Subscribe signal request, %s", requestInfo)
	controller, exist := s.getController(request.JobId)
	if !exist {
		return fmt.Errorf("jobId=%s not registed", request.JobId)
	}
	controller.listenSendChannel(stream)
	return nil
}

// ReportStopComplete report stop process result
func (s *FaultRecoverService) ReportStopComplete(ctx context.Context,
	request *pb.StopCompleteRequest) (*pb.Status, error) {
	hwlog.RunLog.Infof("receive ReportStopComplete, jobId=%s, code=%d, info=%s",
		request.JobId, request.Status.Code, request.Status.Info)
	controller, exist := s.getController(request.JobId)
	if !exist {
		hwlog.RunLog.Errorf("jobId=%s not registed", request.JobId)
		return &pb.Status{
			Code: int32(common.UnRegistry),
			Info: fmt.Sprintf("jobId=%s not registed", request.JobId),
		}, nil
	}
	controller.reportStopCompleteChan <- request
	return &pb.Status{
		Code: int32(common.OK),
		Info: fmt.Sprintf("jobId=%s, uuid=%s, receive ReportStopComplete",
			controller.jobInfo.JobId, controller.uuid),
	}, nil
}

// ReportRecoverStrategy report supported recover strategy to ClusterD
func (s *FaultRecoverService) ReportRecoverStrategy(ctx context.Context,
	request *pb.RecoverStrategyRequest) (*pb.Status, error) {
	hwlog.RunLog.Infof("receive ReportRecoverStrategy, jobId=%s, strategy=%v",
		request.JobId, request.Strategies)
	controller, exist := s.getController(request.JobId)
	if !exist {
		hwlog.RunLog.Errorf("jobId=%s not registed", request.JobId)
		return &pb.Status{
			Code: int32(common.UnRegistry),
			Info: fmt.Sprintf("jobId=%s not registed", request.JobId),
		}, nil
	}
	controller.reportRecoverStrategyChan <- request
	return &pb.Status{
		Code: int32(common.OK),
		Info: fmt.Sprintf("jobId=%s, uuid=%s, receive ReportRecoverStrategy",
			controller.jobInfo.JobId, controller.uuid),
	}, nil
}

// ReportRecoverStatus report recover result
func (s *FaultRecoverService) ReportRecoverStatus(ctx context.Context,
	request *pb.RecoverStatusRequest) (*pb.Status, error) {
	hwlog.RunLog.Infof("receive ReportRecoverStatus, jobId=%s, code=%d, info=%s, strategy=%s",
		request.JobId, request.Status.Code, request.Status.Info, request.Strategy)
	controller, exist := s.getController(request.JobId)
	if !exist {
		hwlog.RunLog.Errorf("jobId=%s not registed", request.JobId)
		return &pb.Status{
			Code: int32(common.UnRegistry),
			Info: fmt.Sprintf("jobId=%s not registed", request.JobId),
		}, nil
	}
	controller.reportStatusChan <- request
	return &pb.Status{
		Code: int32(common.OK),
		Info: fmt.Sprintf("jobId=%s, uuid=%s, receive ReportRecoverStatus",
			controller.jobInfo.JobId, controller.uuid),
	}, nil
}

func giveSoftFault2FaultCenter(jobId string, faults []*pb.FaultRank) {
	t := time.Now().UnixMilli()
	var infos []faultmanager.ReportRecoverInfo
	for _, fault := range faults {
		infos = append(infos, faultmanager.ReportRecoverInfo{
			JobId:       jobId,
			Rank:        fault.RankId,
			RecoverTime: t,
		})
	}
	faultmanager.GlobalFaultProcessCenter.CallbackForReportUceInfo(infos)
}

// ReportProcessFault report soft fault ranks to ClusterD
func (s *FaultRecoverService) ReportProcessFault(ctx context.Context,
	request *pb.ProcessFaultRequest) (*pb.Status, error) {
	requestInfo := fmt.Sprintf("jobId=%s, faultRanks={%s}",
		request.JobId, common.Faults2String(request.FaultRanks))
	hwlog.RunLog.Infof("receive ReportProcessFault, info={%s}", requestInfo)
	controller, exist := s.getController(request.JobId)
	if !exist {
		hwlog.RunLog.Errorf("jobId=%s not registed", request.JobId)
		return &pb.Status{
			Code: int32(common.UnRegistry),
			Info: fmt.Sprintf("jobId=%s not registed", request.JobId),
		}, nil
	}
	controller.saveCacheFault(request.FaultRanks)
	if !common.IsUceFault(request.FaultRanks) {
		_, err := common.LabelFaultPod(request.JobId, common.Faults2Ranks(request.FaultRanks))
		if err != nil {
			hwlog.RunLog.Errorf("failed to label soft fault label, err:%v, jobId=%s",
				err, request.JobId)
		}
		controller.addEvent(common.FaultOccurEvent)
	} else {
		if faultmanager.GlobalFaultProcessCenter != nil {
			giveSoftFault2FaultCenter(request.JobId, request.FaultRanks)
		} else {
			hwlog.RunLog.Warnf("global fault center is nil")
		}
	}
	return &pb.Status{
		Code: int32(common.OK),
		Info: "receive ReportProcessFault",
	}, nil
}

// DeleteJob clear registered resources
func (s *FaultRecoverService) DeleteJob(jobId string) {
	hwlog.RunLog.Infof("current serve jobs=%d, prepare delete jobId=%s", len(s.eventCtl), jobId)
	s.lock.Lock()
	defer func() {
		hwlog.RunLog.Infof("after delete serve jobs=%d, jobId=%s", len(s.eventCtl), jobId)
		s.lock.Unlock()
	}()
	if s.eventCtl == nil {
		return
	}
	controller, exist := s.eventCtl[jobId]
	if !exist || controller == nil {
		return
	}
	controller.reset()
	delete(s.eventCtl, jobId)
	if s.initJob != nil {
		delete(s.initJob, jobId)
	}
}
