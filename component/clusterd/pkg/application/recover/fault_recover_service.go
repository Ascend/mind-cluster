// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package recover a series of service function
package recover

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"k8s.io/api/core/v1"
	"volcano.sh/apis/pkg/apis/scheduling/v1beta1"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/application/faultmanager"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/common"
	"clusterd/pkg/domain/faultdomain"
	"clusterd/pkg/domain/job"
	"clusterd/pkg/domain/pod"
	"clusterd/pkg/domain/podgroup"
	"clusterd/pkg/interface/grpc/recover"
	"clusterd/pkg/interface/kube"
)

const (
	normalFaultValue = "software"
	retryFaultValue  = "retry-fault"
)

// FaultRecoverService is a service for fault recover
type FaultRecoverService struct {
	keepAliveInterval int
	serviceCtx        context.Context
	eventCtl          map[string]*EventController
	initJob           map[string]common.JobBaseInfo
	lock              sync.RWMutex
	faultCh           chan map[string]constant.JobFaultInfo
	podEventCh        chan *v1.Pod
	pb.UnimplementedRecoverServer
}

// NewFaultRecoverService return a new instance of FaultRecoverService
func NewFaultRecoverService(keepAlive int, ctx context.Context) *FaultRecoverService {
	s := &FaultRecoverService{}
	s.keepAliveInterval = keepAlive
	s.serviceCtx = ctx
	s.eventCtl = make(map[string]*EventController)
	s.initJob = make(map[string]common.JobBaseInfo)
	s.faultCh = make(chan map[string]constant.JobFaultInfo, 5)
	if err := faultmanager.RegisterForJobFaultRank(s.faultCh, reflect.TypeOf(*s).Name()); err != nil {
		hwlog.RunLog.Errorf("RegisterForJobFaultRank fail")
	}
	// delete EventController cache added by register interface according delete event of podGroup when job is deleted.
	kube.AddPodGroupFunc(constant.FaultRecover, func(_ *v1beta1.PodGroup, pg *v1beta1.PodGroup, op string) {
		if op != constant.DeleteOperator {
			return
		}
		s.DeleteJob(podgroup.GetJobKeyByPG(pg))
	})
	go s.checkFaultFromFaultCenter()
	go s.podStatusMonitor()
	return s
}

func catchAndSetExceptionInfo(code *int32, info *string, ctl *EventController) {
	if r := recover(); r != nil {
		*code = int32(common.ServerInnerError)
		*info = fmt.Sprintf("jobId=%s, uuid=%s, chan closed",
			ctl.jobInfo.JobId, ctl.uuid)
	}
}

func (s *FaultRecoverService) podStatusMonitor() {
	for {
		select {
		case <-s.serviceCtx.Done():
			return
		case podInfo := <-pod.GetRunningEventChan():
			ctl := s.getControllerByPod(podInfo)
			if ctl == nil {
				continue
			}
			hwlog.RunLog.Infof("get running pod info from pod informmer =%v", podInfo.Name)
			ctl.currentHotSwitchBackupPodId = string(podInfo.UID)
			// use to cancel pod timeout monitor
			ctl.ChangePodStatus(v1.PodRunning)
		case podInfo := <-pod.GetDeletedEventChan():
			ctl := s.getControllerByPod(podInfo)
			if ctl == nil {
				continue
			}
			status, exists := podInfo.Annotations[api.PodTypeKey]
			if exists && status == api.PodTypeBackup {
				// new pod deleted
				hwlog.RunLog.Infof("new pod %s has been deleted", podInfo.Name)
			} else {
				// old pod deleted
				hwlog.RunLog.Infof("old pod %s has been deleted", podInfo.Name)
				ctl.addEvent(common.OldPodDeletedEvent)
			}
		}
	}
}

func (s *FaultRecoverService) getControllerByPod(podInfo *v1.Pod) *EventController {
	jobId := pod.GetJobKeyByPod(podInfo)
	ctl, exist := s.getController(jobId)
	if !exist || ctl == nil {
		hwlog.RunLog.Errorf("jobId=%s not exist", jobId)
		return nil
	}
	return ctl
}

func (s *FaultRecoverService) getController(jobId string) (*EventController, bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	ctl, exist := s.eventCtl[jobId]
	return ctl, exist
}

func (s *FaultRecoverService) notifyFaultInfoForJob(faultInfo constant.JobFaultInfo) {
	controller, exist := s.getController(faultInfo.JobId)
	if !exist || controller == nil {
		hwlog.RunLog.Errorf("jobId=%s not exist", faultInfo.JobId)
		return
	}
	hwlog.RunLog.Infof("get fault info from fault center=%v", faultInfo)
	if s.skipHandleSubHealthyFaults(controller, faultInfo) {
		hwlog.RunLog.ErrorfWithLimit(constant.SubHealthyState, controller.jobInfo.JobId,
			"jobId=%s skip handle subHealthy faults", faultInfo.JobId)
		return
	}
	hwlog.ResetErrCnt(constant.SubHealthyState, controller.jobInfo.JobId)

	var grpcFormatFaults []*pb.FaultRank = nil
	subHealthyHotSwitch := faultInfo.HealthyState == constant.SubHealthyState && controller.jobInfo.HotSwitch
	hasNewFaultPod := false
	for _, info := range faultInfo.FaultList {
		if info.PodUid == "" || info.PodRank == "" {
			hwlog.RunLog.Warnf("invalid pod info, podId=%s, podRank=%s", info.PodUid, info.PodRank)
			continue
		}
		faultPod := make(map[string]string)
		faultPod[info.PodRank] = info.PodUid
		_, ok := controller.faultPod[info.PodRank]
		lastestStrategies, _ := controller.getStrategyResult()
		if len(lastestStrategies) > 0 &&
			lastestStrategies[len(lastestStrategies)-1] == constant.ScaleInStrategyName && ok {
			hwlog.RunLog.Debugf("job %s fault pod has deal", controller.jobInfo.JobId)
			continue
		}
		hasNewFaultPod = true
		controller.mergeFaultPod(faultPod)
		hwlog.RunLog.Debugf("mergeFaultPod: %v", faultPod)
		fault := &pb.FaultRank{
			RankId: info.RankId,
		}
		fault.FaultType = constant.NormalFaultType
		if info.DoStepRetry && faultdomain.IsUceFault(info.FaultCode) {
			fault.FaultType = constant.UceFaultType
		}
		if info.DoStepRetry && faultdomain.IsHcclRetryFault(info.FaultCode) {
			fault.FaultType = constant.HcclFaultType
		}
		grpcFormatFaults = append(grpcFormatFaults, fault)
	}
	if !hasNewFaultPod {
		return
	}
	hwlog.RunLog.Infof("jobId=%s, fault center fault info change format to grpcFormat, faults=%s",
		controller.jobInfo.JobId, common.Faults2String(grpcFormatFaults))
	controller.saveCacheFault(grpcFormatFaults)
	controller.healthState = faultInfo.HealthyState
	controller.restartFaultProcess = common.CanRestartFaultProcess(faultInfo.JobId, faultInfo.FaultList)
	dealWithForceRelease(controller, faultInfo, grpcFormatFaults)
	if subHealthyHotSwitch {
		handleSubHealthyFault(controller, grpcFormatFaults)
		return
	}
	controller.addEvent(common.FaultOccurEvent)
}

func handleSubHealthyFault(ctl *EventController, faults []*pb.FaultRank) {
	faultPods := ctl.GetFaultPod()
	if _, exists := faultPods[api.MasterPodRank]; exists {
		// log max 3 times
		hwlog.RunLog.ErrorfWithLimit(constant.SubHealthyState, ctl.jobInfo.JobId+api.MasterPodRank,
			"fault occour on master node, skip this fault, podRank: %v,faults: %v", api.MasterPodRank, faults)
		return
	}
	hwlog.ResetErrCnt(constant.SubHealthyState, ctl.jobInfo.JobId+api.MasterPodRank)

	if len(faults) > 0 {
		ctl.addEvent(common.BeginHotSwitchEvent)
	}
}

func dealWithForceRelease(controller *EventController, faultInfo constant.JobFaultInfo, grpcFormatFaults []*pb.FaultRank) {
	if len(grpcFormatFaults) == 0 {
		return
	}
	faultNodes, ok := isNeedForceReleaseFault(faultInfo)
	if !ok {
		return
	}
	hwlog.RunLog.Infof("jobId=%s force release", controller.jobInfo.JobId)
	faultInfos := job.GetJobFaultSdIdAndNodeName(controller.jobInfo.JobId, faultNodes)
	if faultInfos == nil {
		return
	}
	kube.CreateOrUpdateSuperPodFaultInfo(controller.jobInfo.JobId, faultInfos)
	time.Sleep(constant.ReleaseTimeOut)
}

func isNeedForceReleaseFault(faultInfo constant.JobFaultInfo) (map[string]struct{}, bool) {
	faultNodes := make(map[string]struct{})
	for _, faultDevice := range faultInfo.FaultDevice {
		if strings.Contains(faultDevice.FaultCode, constant.CardDropFault) ||
			faultDevice.DeviceType == constant.FaultTypeNode || faultDevice.DeviceType == constant.SwitchFaultType {
			faultNodes[faultDevice.ServerName] = struct{}{}
		}
	}
	if len(faultNodes) == 0 {
		return faultNodes, false
	}
	return faultNodes, true
}

func (s *FaultRecoverService) skipHandleSubHealthyFaults(ctl *EventController,
	faultInfo constant.JobFaultInfo) bool {
	// not sub health fault, cannot be skipped
	if faultInfo.HealthyState != constant.SubHealthyState {
		return false
	}
	// hotswitch and pytorch scene，cannot be skipped
	if ctl.jobInfo.HotSwitch {
		if ctl.jobInfo.Framework != constant.PtFramework {
			hwlog.RunLog.Warnf("subhealthy hotswitch only support pytorch framework,current framework is:%v ", ctl.jobInfo.Framework)
			ctl.jobInfo.HotSwitch = false
			ctl.jobInfo.SubHealthyStrategy = constant.SubHealthyIngore
			return true
		}

		const maxFaultPodLen = 10
		faultPods := ctl.GetFaultPod()
		if len(faultPods) > maxFaultPodLen {
			hwlog.RunLog.Warnf("too much fault pods,change subhealthy strategy to ignore,fault pods len: %v", len(faultPods))
			ctl.jobInfo.HotSwitch = false
			ctl.jobInfo.SubHealthyStrategy = constant.SubHealthyIngore
			return true
		}
		return false
	}
	// sub health and not hotswitch scene , if graceExit is false, skip
	if !ctl.jobInfo.GraceExit {
		return true
	}
	// sub health and not hotswitch scene , if not onlyDump strategy, skip
	if !ctl.onlySupportDumpStrategy() {
		return true
	}
	return false
}

func isMasterSubhealthHotSwitchFault(subHealthyHotSwitch bool, info constant.FaultRank) bool {
	if subHealthyHotSwitch && info.PodRank == api.MasterPodRank {
		// log max 3 times
		hwlog.RunLog.ErrorfWithLimit(constant.SubHealthyState, info.PodRank,
			"fault occour on master node,skip this fault, podRank: %v,rankId: %v", info.PodRank, info.RankId)
		return true
	}
	hwlog.ResetErrCnt(constant.SubHealthyState, info.PodRank)
	return false
}

func (s *FaultRecoverService) dealWithJobFaultInfo(jobFaultInfoList []constant.JobFaultInfo) {
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

func (s *FaultRecoverService) checkFault(allJobFaultInfo map[string]constant.JobFaultInfo) {
	var registeredJobInfo []constant.JobFaultInfo
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
	for {
		select {
		case <-s.serviceCtx.Done():
			return
		case allJobFaultInfo := <-s.faultCh:
			s.checkFault(allJobFaultInfo)
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
	pg := podgroup.GetPodGroup(jobId)
	framework := podgroup.GetModelFramework(&pg)
	return common.JobBaseInfo{
		JobId:         jobId,
		JobName:       jobName,
		PgName:        pgName,
		Namespace:     namespace,
		RecoverConfig: config,
		Framework:     framework,
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
	if ok := podgroup.CheckPodGroupExist(req.JobId); !ok {
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
	code := int32(common.OK)
	info := fmt.Sprintf("jobId=%s, uuid=%s, receive ReportStopComplete",
		controller.jobInfo.JobId, controller.uuid)
	func() {
		defer catchAndSetExceptionInfo(&code, &info, controller)
		controller.reportStopCompleteChan <- request
	}()

	return &pb.Status{Code: code, Info: info}, nil
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

	code := int32(common.OK)
	info := fmt.Sprintf("jobId=%s, uuid=%s, receive ReportRecoverStrategy",
		controller.jobInfo.JobId, controller.uuid)
	func() {
		defer catchAndSetExceptionInfo(&code, &info, controller)
		controller.reportRecoverStrategyChan <- request
	}()

	return &pb.Status{Code: code, Info: info}, nil
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

	code := int32(common.OK)
	info := fmt.Sprintf("jobId=%s, uuid=%s, receive ReportRecoverStatus",
		controller.jobInfo.JobId, controller.uuid)
	func() {
		defer catchAndSetExceptionInfo(&code, &info, controller)
		controller.reportStatusChan <- request
	}()

	return &pb.Status{Code: code, Info: info}, nil
}

func giveSoftFault2FaultCenter(jobId string, faults []*pb.FaultRank) {
	t := time.Now().UnixMilli()
	var infos []constant.ReportRecoverInfo
	for _, fault := range faults {
		infos = append(infos, constant.ReportRecoverInfo{
			JobId:       jobId,
			Rank:        fault.RankId,
			RecoverTime: t,
			FaultType:   fault.FaultType,
		})
	}
	faultmanager.CallbackForReportRetryInfo(infos)
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
	var err error
	faultReason := getFaultReason(request.FaultRanks)
	faultPod, err := common.LabelFaultPod(request.JobId,
		common.Faults2Ranks(request.FaultRanks), controller.GetFaultPod(), faultReason)
	controller.mergeFaultPod(faultPod)
	if err != nil {
		hwlog.RunLog.Errorf("failed to label soft fault label, err:%v, jobId=%s",
			err, request.JobId)
	}
	if !common.IsRetryFault(request.FaultRanks) {
		// when config only support dump strategy, in order to be able to dump directly, set healthState to UnHealthy
		controller.healthState = constant.UnHealthyState
		controller.restartFaultProcess = common.CanRestartFaultProcess(request.JobId, nil)
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
	controller.reset(true)
	delete(s.eventCtl, jobId)
	if s.initJob != nil {
		delete(s.initJob, jobId)
	}
}

func getFaultReason(faults []*pb.FaultRank) string {
	if common.IsRetryFault(faults) {
		return retryFaultValue
	}
	return normalFaultValue
}

// HealthCheck report connection health check
func (s *FaultRecoverService) HealthCheck(ctx context.Context, request *pb.ClientInfo) (*pb.Status, error) {
	hwlog.RunLog.Debugf("receive HeathCheck from client, jobId: %s", request.JobId)
	return &pb.Status{
		Code: int32(common.OK),
		Info: fmt.Sprintf("jobId=%s, receive HeathCheck", request.JobId),
	}, nil
}
