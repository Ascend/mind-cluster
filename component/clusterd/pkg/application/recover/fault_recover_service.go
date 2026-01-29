// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package recover a series of service function
package recover

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"

	"k8s.io/api/core/v1"
	"volcano.sh/apis/pkg/apis/scheduling/v1beta1"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/application/faultmanager"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
	"clusterd/pkg/domain/common"
	"clusterd/pkg/domain/faultdomain"
	"clusterd/pkg/domain/pod"
	"clusterd/pkg/domain/podgroup"
	"clusterd/pkg/interface/grpc/recover"
	"clusterd/pkg/interface/kube"
)

const (
	normalFaultValue = "software"
	retryFaultValue  = "retry-fault"
	maxFaultPodLen   = 10
)

// FaultRecoverService is a service for fault recover
type FaultRecoverService struct {
	keepAliveInterval    int
	serviceCtx           context.Context
	eventCtl             map[string]*EventController
	initJob              map[string]common.JobBaseInfo
	lock                 sync.RWMutex
	faultCh              chan map[string]constant.JobFaultInfo
	podEventCh           chan *v1.Pod
	isJobReschedulingMap map[string]bool
	currentFaults        map[string]map[string]bool
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
	s.isJobReschedulingMap = make(map[string]bool)
	s.currentFaults = make(map[string]map[string]bool)

	filterLevel := []string{constant.NotHandleFault, constant.PreSeparateNPU}
	if err := faultmanager.RegisterForJobFaultRank(s.faultCh, filterLevel, reflect.TypeOf(*s).Name()); err != nil {
		hwlog.RunLog.Errorf("RegisterForJobFaultRank fail")
	}
	// delete EventController cache added by register interface according delete event of podGroup when job is deleted.
	kube.AddPodGroupFunc(constant.FaultRecover, func(_ *v1beta1.PodGroup, pg *v1beta1.PodGroup, op string) {
		if op == constant.UpdateOperator {
			s.resetByJobRescheduling(pg)
		} else if op == constant.DeleteOperator {
			jobID := podgroup.GetJobKeyByPG(pg)
			delete(s.isJobReschedulingMap, jobID)
			s.DeleteJob(jobID)
		}
	})
	go s.checkFaultFromFaultCenter()
	go s.podStatusMonitor()
	return s
}

func (s *FaultRecoverService) resetByJobRescheduling(pg *v1beta1.PodGroup) {
	jobID := podgroup.GetJobKeyByPG(pg)
	controller, exist := s.getController(jobID)
	if !exist || controller == nil {
		hwlog.RunLog.Warnf("cannot find target controller by job<%s>", jobID)
		return
	}
	if isPodGroupJobRescheduling(pg) {
		s.setJobReschedulingMap(jobID)
		isJobReschedulingAnnotation := map[string]interface{}{
			constant.IsJobRescheduling: "true",
		}
		_, err := kube.RetryPatchPodGroupAnnotations(controller.jobInfo.PgName, controller.jobInfo.Namespace,
			retryTimes, isJobReschedulingAnnotation)
		if err != nil {
			hwlog.RunLog.Errorf("failed to patch pg when job is rescheduling, err=%v, pgName=%s",
				err, controller.jobInfo.PgName)
		}
		return
	}
	s.resetStateMachineAfterJobRescheduling(jobID, pg, controller)
}

// setJobReschedulingMap label the job has just been job rescheduling
func (s *FaultRecoverService) setJobReschedulingMap(jobId string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.isJobReschedulingMap[jobId] = true
}

// resetStateMachineByJobRescheduling reset state machine to init state after job rescheduling
func (s *FaultRecoverService) resetStateMachineAfterJobRescheduling(jobId string, pg *v1beta1.PodGroup,
	controller *EventController) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.isJobReschedulingMap[jobId] {
		if _, ok := pg.Annotations[constant.IsJobRescheduling]; !ok {
			controller.reset(false)
			delete(s.isJobReschedulingMap, jobId)
		}
	}
}

func isPodGroupJobRescheduling(pg *v1beta1.PodGroup) bool {
	return pg.Status.Running == 0 && pg.Status.Succeeded == 0 && pg.Status.Failed == 0
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
	if len(faultInfo.FaultList) > 0 || len(faultInfo.FaultDevice) > 0 {
		hwlog.RunLog.Infof("get fault info from fault center,jobId:%s,faultList=%v", faultInfo.JobId, faultInfo.FaultList)
		hwlog.RunLog.Infof("get fault info from fault center,jobId:%s,faultDevice=%v", faultInfo.JobId, faultInfo.FaultDevice)
	}
	if s.skipHandleSubHealthyFaults(controller, &faultInfo) {
		hwlog.RunLog.ErrorfWithLimit(constant.SubHealthyState, controller.jobInfo.JobId+"SkipHandleSubHealthy",
			"jobId=%s skip handle subHealthy faults", faultInfo.JobId)
		return
	}
	hwlog.ResetErrCnt(constant.SubHealthyState, controller.jobInfo.JobId+"SkipHandleSubHealthy")
	subHealthyHotSwitch := faultInfo.HealthyState == constant.SubHealthyState && controller.jobInfo.HotSwitch
	grpcFormatFaults := s.getGrpcFormatFaults(faultInfo, controller)
	addedFaults, addedFaultRanks := s.getAdditionFault(faultInfo.JobId, grpcFormatFaults)
	if len(grpcFormatFaults) == 0 {
		hwlog.RunLog.Debugf("job %s has no new faults", faultInfo.JobId)
		return
	}
	hwlog.RunLog.Infof("jobId=%s, fault center fault info change format to grpcFormat, faults=%s",
		controller.jobInfo.JobId, common.Faults2String(grpcFormatFaults))
	controller.saveCacheFault(grpcFormatFaults)
	controller.healthState = faultInfo.HealthyState
	controller.updateRestartProcessOrPodInfoByHardwareFault(faultInfo.FaultList)
	if subHealthyHotSwitch {
		if len(grpcFormatFaults) > 0 {
			controller.addEvent(common.BeginHotSwitchEvent)
		}
		return
	}
	if len(addedFaults) == 0 {
		hwlog.RunLog.Infof("jobId=%s, no new faults added, skip additional processing", faultInfo.JobId)
		return
	}
	hwlog.RunLog.Infof("jobId=%s, new faults detected, enter additional processing, faultRanks=%v",
		faultInfo.JobId, addedFaultRanks)
	onlyRetryFault, supportRetry := s.getRetryStatus(addedFaults, controller)
	removeGrpcFault, faultNodes := s.getFaultAndFaultNodes(addedFaultRanks, controller)
	if !supportRetry || !onlyRetryFault && len(faultNodes) > 0 {
		s.sendPreExitSignal(controller, removeGrpcFault, faultNodes)
		return
	}
	controller.addEvent(common.FaultOccurEvent)
}

func (s *FaultRecoverService) sendPreExitSignal(controller *EventController,
	removeGrpcFault []*pb.FaultRank, faultNodes []string) {
	signal := &pb.ProcessManageSignal{
		Uuid:           common.NewEventId(randomLen),
		JobId:          controller.jobInfo.JobId,
		SignalType:     constant.PreExitProcessSignalType,
		Actions:        preExitProcessActions,
		ChangeStrategy: "",
		FaultRanks:     removeGrpcFault,
		NodeRankIds:    faultNodes,
	}
	controller.signalEnqueue(signal)
}

func (s *FaultRecoverService) getAdditionFault(jobId string,
	grpcFormatFaults []*pb.FaultRank) (map[string]*pb.FaultRank, []string) {
	newFaults := make(map[string]bool)
	faultMap := make(map[string]*pb.FaultRank)
	for _, fault := range grpcFormatFaults {
		faultKey := fmt.Sprintf("%s_%s", fault.RankId, fault.FaultType)
		newFaults[faultKey] = true
		faultMap[faultKey] = fault
	}
	s.lock.Lock()
	defer s.lock.Unlock()
	currentFaultMap, exists := s.currentFaults[jobId]
	if !exists {
		currentFaultMap = make(map[string]bool)
	}
	s.currentFaults[jobId] = newFaults
	addedFaults := make(map[string]*pb.FaultRank)
	var faultRanks []string
	for faultKey := range newFaults {
		if !currentFaultMap[faultKey] {
			addedFaults[faultKey] = faultMap[faultKey]
			faultRanks = append(faultRanks, faultMap[faultKey].RankId)
		}
	}
	return addedFaults, faultRanks
}

func (s *FaultRecoverService) getRetryStatus(addedFaults map[string]*pb.FaultRank,
	controller *EventController) (bool, bool) {
	onlyRetryFault := true
	for _, fault := range addedFaults {
		if fault.FaultType != constant.UceFaultType && fault.FaultType != constant.HcclFaultType {
			onlyRetryFault = false
			break
		}
	}
	supportRetry := controller.supportRetryStrategy()
	hwlog.RunLog.Infof("jobId=%s, onlyRetryFault=%v, supportRetry=%v",
		controller.jobInfo.JobId, onlyRetryFault, supportRetry)
	return onlyRetryFault, supportRetry
}

func (s *FaultRecoverService) getFaultAndFaultNodes(faultRanks []string,
	controller *EventController) ([]*pb.FaultRank, []string) {
	removeSameRankIds := util.RemoveSliceDuplicateElement(faultRanks)
	var removeGrpcFault []*pb.FaultRank
	for _, rank := range removeSameRankIds {
		removeGrpcFault = append(removeGrpcFault, &pb.FaultRank{
			RankId:    rank,
			FaultType: constant.NormalFaultType,
		})
	}

	faultNodes, err := common.GetNodeRankIdsByRankIds(controller.jobInfo.JobId, removeSameRankIds)
	if err != nil {
		hwlog.RunLog.Errorf("jobId=%s, get node rank ids by rank ids failed, err=%v", controller.jobInfo.JobId, err)
	}
	hwlog.RunLog.Infof("jobId=%s, removeGrpcFault=%v, faultNodes=%v",
		controller.jobInfo.JobId, removeGrpcFault, faultNodes)
	return removeGrpcFault, faultNodes
}

func (s *FaultRecoverService) getGrpcFormatFaults(faultInfo constant.JobFaultInfo, controller *EventController) []*pb.FaultRank {
	grpcFormatFaults := make([]*pb.FaultRank, 0)
	for _, info := range faultInfo.FaultList {
		if info.PodUid == "" || info.PodRank == "" {
			hwlog.RunLog.Warnf("invalid pod info, podId=%s, podRank=%s", info.PodUid, info.PodRank)
			continue
		}
		if faultInfo.HealthyState != constant.SubHealthyState && info.FaultLevel == constant.SubHealthFault {
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
	return grpcFormatFaults
}

func (s *FaultRecoverService) skipHandleSubHealthyFaults(ctl *EventController, faultInfo *constant.JobFaultInfo) bool {
	// not sub health fault, cannot be skipped
	if faultInfo.HealthyState != constant.SubHealthyState {
		return false
	}
	// hotswitch and pytorch scene，cannot be skipped
	if ctl.jobInfo.HotSwitch {
		if len(faultInfo.FaultList) <= 0 {
			return false
		}
		return skipHandleSubHealthyHotSwitch(ctl, faultInfo)
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

func skipHandleSubHealthyHotSwitch(ctl *EventController, faultInfo *constant.JobFaultInfo) bool {
	if ctl.jobInfo.Framework != constant.PtFramework && ctl.jobInfo.Framework != constant.MsFramework {
		hwlog.RunLog.Warnf("subhealthy hotswitch only support pytorch and mindspore framework,current is:%v ",
			ctl.jobInfo.Framework)
		ctl.jobInfo.HotSwitch = false
		ctl.jobInfo.SubHealthyStrategy = constant.SubHealthyIngore
		return true
	}
	// If it is in another state machine process, skip for this time
	if ctl.state.GetState() != common.InitState {
		hwlog.RunLog.ErrorfWithLimit(constant.SubHealthyState, ctl.jobInfo.JobId+"InAnotherState",
			"not handle subhealth hotswitch when state machine is in another state, jobId=%s", ctl.jobInfo.JobId)
		return true
	}
	hwlog.ResetErrCnt(constant.SubHealthyState, ctl.jobInfo.JobId+"InAnotherState")
	newFaultRankList := make([]constant.FaultRank, 0)
	faultPods := map[string]struct{}{}
	currentHandlePod := ""
	for _, info := range faultInfo.FaultList {
		if info.PodUid == "" || info.PodRank == "" {
			hwlog.RunLog.Warnf("invalid pod info, podId=%s, podRank=%s", info.PodUid, info.PodRank)
			continue
		}
		if info.PodRank == api.MasterPodRank {
			continue
		}
		faultPods[info.PodRank] = struct{}{}
		if currentHandlePod == "" {
			currentHandlePod = info.PodUid
		}
		if info.PodUid != currentHandlePod {
			continue
		}
		newFaultRankList = append(newFaultRankList, info)
	}
	if len(newFaultRankList) == 0 {
		// only master has subhealth fault, log max 3 times
		hwlog.RunLog.ErrorfWithLimit(constant.SubHealthyState, ctl.jobInfo.JobId+api.MasterPodRank,
			"fault occour on master node, skip this fault, podRank: %v", api.MasterPodRank)
		return true
	}
	if len(faultPods) > maxFaultPodLen {
		hwlog.RunLog.Warnf("too much fault pods,change subhealthy strategy to ignore, "+
			"maxFaultPod: %v, current fault pods len: %v", maxFaultPodLen, len(faultPods))
		ctl.jobInfo.HotSwitch = false
		ctl.jobInfo.SubHealthyStrategy = constant.SubHealthyIngore
		return true
	}
	faultInfo.FaultList = newFaultRankList
	hwlog.ResetErrCnt(constant.SubHealthyState, ctl.jobInfo.JobId+api.MasterPodRank)
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
		s.preHandleFaultInfo(jobId, &jobFaultInfo)
		registeredJobInfo = append(registeredJobInfo, jobFaultInfo)
	}
	s.dealWithJobFaultInfo(registeredJobInfo)
}

func (s *FaultRecoverService) preHandleFaultInfo(jobId string, faultInfo *constant.JobFaultInfo) {
	if len(faultInfo.FaultList) <= 0 {
		return
	}
	const logDomain = "preHandleFaultInfo"
	currentServerMap, podToServerMap := getJobServerInfos(jobId)
	currentFaultServerMap := make(map[string]bool)
	newFaultDevice := make([]constant.FaultDevice, 0)
	newFaultList := make([]constant.FaultRank, 0)
	for _, device := range faultInfo.FaultDevice {
		if !currentServerMap[device.ServerName] {
			hwlog.RunLog.WarnfWithLimit(logDomain, jobId+device.ServerName,
				"jobId=%s, fault device [%s] is not in current server list", jobId, device.ServerName)
			continue
		}
		hwlog.ResetErrCnt(logDomain, jobId+device.ServerName)
		currentFaultServerMap[device.ServerName] = true
		newFaultDevice = append(newFaultDevice, device)
	}
	if len(newFaultDevice) != len(faultInfo.FaultDevice) {
		hwlog.RunLog.WarnfWithLimit(logDomain, jobId+"faultDeviceChanged",
			"jobId=%s, FaultDevice has changed, old:[%v],new:[%v]", jobId, faultInfo.FaultDevice, newFaultDevice)
	} else {
		hwlog.ResetErrCnt(logDomain, jobId+"faultDeviceChanged")
	}
	faultInfo.FaultDevice = newFaultDevice
	for _, fault := range faultInfo.FaultList {
		actualServerName, ok := podToServerMap[fault.PodUid]
		if !ok { // pending pod, ignore
			continue
		}
		logUniqueKey := jobId + fault.PodUid + "notRunningOnFaultServer"
		if _, ok := currentFaultServerMap[actualServerName]; !ok {
			hwlog.RunLog.WarnfWithLimit(logDomain, logUniqueKey,
				"jobId=%s, pod [%s/%s] is not running on fault server", jobId, actualServerName, fault.PodUid)
			continue
		}
		hwlog.ResetErrCnt(logDomain, logUniqueKey)
		newFaultList = append(newFaultList, fault)
	}
	if len(newFaultList) != len(faultInfo.FaultList) {
		hwlog.RunLog.WarnfWithLimit(logDomain, jobId+"faultListChanged",
			"jobId=%s, FaultList has changed, old:[%v],new:[%v]", jobId, faultInfo.FaultList, newFaultList)
	} else {
		hwlog.ResetErrCnt(logDomain, jobId+"faultListChanged")
	}
	faultInfo.FaultList = newFaultList
}

func getJobServerInfos(jobId string) (map[string]bool, map[string]string) {
	podsInJob := pod.GetSimplePodByJobId(jobId)
	var currentServerMap = make(map[string]bool)
	var podToServerMap = make(map[string]string)
	for _, pod := range podsInJob {
		if pod.NodeName != "" {
			currentServerMap[pod.NodeName] = true
			podToServerMap[pod.PodUid] = pod.NodeName
			continue
		}
		hwlog.RunLog.Warnf("jobId=%s pod [%v] is not scheduled", jobId, pod.PodUid)
	}
	hwlog.RunLog.Debugf("jobId=%s, currentServerMap: %v", jobId, currentServerMap)
	hwlog.RunLog.Debugf("jobId=%s, podToServerMap: %v", jobId, podToServerMap)
	return currentServerMap, podToServerMap
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
	if config.ProcessRecoverEnable == false && config.HotSwitch == false {
		hwlog.RunLog.Errorf("process recover enable and subhealthy hotswtich does not open, jobId=%s", jobId)
		return common.JobBaseInfo{}, common.ProcessRecoverEnableOff,
			fmt.Errorf("job(uid=%s) process-recover-enable and subhealthy hotswtich not open:%v", jobId, err)
	}
	pg, err := kube.RetryGetPodGroup(pgName, namespace, constant.GetPodGroupTimes)
	if err != nil {
		hwlog.RunLog.Warnf("get podGroup err: %v, pgName=%s, nameSpace=%s", err, pgName, namespace)
	}
	return common.JobBaseInfo{
		JobId:         jobId,
		JobName:       jobName,
		PgName:        pgName,
		Namespace:     namespace,
		RecoverConfig: config,
		Framework:     podgroup.GetModelFramework(pg),
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
	faultReason := getFaultReason(append(controller.cacheRetryFault, controller.cacheNormalFault...))
	faultPod, err := common.LabelFaultPod(request.JobId,
		common.Faults2Ranks(request.FaultRanks), controller.GetFaultPod(), faultReason)
	controller.mergeFaultPod(faultPod)
	if err != nil {
		hwlog.RunLog.Errorf("failed to label soft fault label, err:%v, jobId=%s",
			err, request.JobId)
	}
	controller.updateRestartProcessOrPodInfoBySoftFault(request.FaultRanks)
	if !common.IsRetryFault(request.FaultRanks) {
		// when config only support dump strategy, in order to be able to dump directly, set healthState to UnHealthy
		controller.healthState = constant.UnHealthyState
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
	delete(s.currentFaults, jobId)
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
