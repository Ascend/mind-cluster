// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package recover a series of service function
package recover

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"sync"
	"time"

	kubeflowutil "github.com/kubeflow/common/pkg/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"

	"ascend-common/api"
	"ascend-common/api/ascend-operator/apis/batch/v1"
	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/application/faultmanager"
	"clusterd/pkg/application/faultmanager/cmprocess/recoverinplace"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
	"clusterd/pkg/domain/common"
	"clusterd/pkg/domain/job"
	"clusterd/pkg/domain/pod"
	"clusterd/pkg/domain/podgroup"
	"clusterd/pkg/domain/statistics"
	"clusterd/pkg/domain/superpod"
	"clusterd/pkg/interface/grpc/recover"
	"clusterd/pkg/interface/kube"
)

const (
	retryTimes           = 3
	randomLen            = 16
	reportTimeoutMinutes = 15
	faultFlushSeconds    = 1
	eventChanLength      = 10
	tpBlockStr16         = "16"
	tpBlockStr32         = "32"
	tpBlockStr64         = "64"
)

var (
	saveAndExitActions            = []string{constant.SaveAndExitAction}
	stopTrainActions              = []string{constant.StopAction}
	pauseTrainActions             = []string{constant.PauseTrainAction}
	pauseStartAgentActions        = []string{constant.PauseStartAgent}
	continueStartAgentActions     = []string{constant.ContinueStartAgent}
	globalFaultActions            = []string{constant.OnGlobalRankAction}
	changeStrategyActions         = []string{constant.ChangeStrategyAction}
	hotSwitchActions              = []string{constant.HotSwitchAction}
	stopHotSwitchActions          = []string{constant.StopSwitchAction}
	newPodRunningActions          = []string{constant.NewPodRunningAction}
	preExitProcessActions         = []string{constant.PreExitProcessAction}
	notifyStrategySuccessEventMap = map[string]string{
		constant.ProcessRetryStrategyName:   common.NotifyRetrySuccessEvent,
		constant.ProcessRecoverStrategyName: common.NotifyRecoverSuccessEvent,
		constant.ProcessDumpStrategyName:    common.NotifyDumpSuccessEvent,
		constant.ProcessExitStrategyName:    common.NotifyExitSuccessEvent,
		constant.ProcessContinueTrain:       common.NotifyContinueSuccessEvent,
		constant.ScaleInStrategyName:        common.NotifyScaleInStrategySuccessEvent,
		constant.ScaleOutStrategyName:       common.NotifyScaleOutStrategySuccessEvent,
		constant.ProcessMigration:           common.NotifySuccessEvent,
	}
	validTpBlock = []string{tpBlockStr16, tpBlockStr32, tpBlockStr64}
)

// EventController is recover event controller
type EventController struct {
	jobInfo                     common.JobBaseInfo
	faultFlushing               bool
	keepAliveSecond             int
	uuid                        string
	prePodForScale              map[string]string
	faultPod                    map[string]string
	originPod                   map[string]string
	events                      chan string
	latestStrategy              []string
	latestRecoverResult         []*pb.RecoverStatusRequest
	agentReportStrategies       []string
	platStrategy                string
	signalChan                  chan *pb.ProcessManageSignal
	cacheNormalFault            []*pb.FaultRank
	cacheRetryFault             []*pb.FaultRank
	healthState                 string
	globalSwitchRankIDs         []string
	globalOps                   []bool
	switchNicResponse           chan *pb.SwitchNicResponse
	switchRankList              chan *pb.SwitchRankList
	switchRankResult            chan *pb.SwitchResult
	stressTestNotifyChan        chan *pb.StressTestRankParams
	restartFaultProcess         bool
	recoverInPlacePodFaults     map[string]*constant.PodFaultInfo
	controllerContext           context.Context
	ctxCancelFunc               context.CancelFunc
	serviceContext              context.Context
	state                       *common.StateMachine
	reportStopCompleteChan      chan *pb.StopCompleteRequest
	reportRecoverStrategyChan   chan *pb.RecoverStrategyRequest
	reportStatusChan            chan *pb.RecoverStatusRequest
	scheduleResultChan          chan bool
	isChanClosed                bool
	ctlResetTime                int64
	lock                        sync.RWMutex
	stressTestParam             common.StressTestParam
	stressTestResponse          chan *pb.StressTestResponse
	stressTestResult            chan *pb.StressTestResult
	isolateNodes                sets.String
	newPodStatusMonitorChan     chan corev1.PodPhase
	currentHotSwitchFaultPodId  string
	currentHotSwitchBackupPodId string
}

func catchException() {
	if err := recover(); err != nil {
		hwlog.RunLog.Warnf("catch exception: %v", err)
	}
}

// NewEventController return pointer of EventController
func NewEventController(jobInfo common.JobBaseInfo, keepAlive int, serviceCtx context.Context) *EventController {
	ctl := &EventController{
		jobInfo:                     jobInfo,
		faultFlushing:               false,
		keepAliveSecond:             keepAlive,
		uuid:                        "",
		faultPod:                    make(map[string]string),
		prePodForScale:              make(map[string]string),
		originPod:                   make(map[string]string),
		events:                      make(chan string, eventChanLength),
		latestStrategy:              []string{},
		latestRecoverResult:         []*pb.RecoverStatusRequest{},
		agentReportStrategies:       []string{},
		platStrategy:                "",
		signalChan:                  make(chan *pb.ProcessManageSignal, 1),
		reportStopCompleteChan:      make(chan *pb.StopCompleteRequest, 1),
		reportRecoverStrategyChan:   make(chan *pb.RecoverStrategyRequest, 1),
		reportStatusChan:            make(chan *pb.RecoverStatusRequest, 1),
		scheduleResultChan:          make(chan bool, 1),
		cacheNormalFault:            []*pb.FaultRank{},
		cacheRetryFault:             []*pb.FaultRank{},
		healthState:                 constant.HealthyState,
		globalSwitchRankIDs:         []string{},
		globalOps:                   []bool{},
		switchNicResponse:           make(chan *pb.SwitchNicResponse, eventChanLength),
		switchRankList:              make(chan *pb.SwitchRankList, 1),
		switchRankResult:            make(chan *pb.SwitchResult, 1),
		stressTestNotifyChan:        make(chan *pb.StressTestRankParams, 1),
		restartFaultProcess:         false,
		recoverInPlacePodFaults:     make(map[string]*constant.PodFaultInfo),
		isChanClosed:                false,
		serviceContext:              serviceCtx,
		lock:                        sync.RWMutex{},
		stressTestParam:             make(common.StressTestParam),
		stressTestResponse:          make(chan *pb.StressTestResponse, eventChanLength),
		stressTestResult:            make(chan *pb.StressTestResult, 1),
		isolateNodes:                make(sets.String),
		newPodStatusMonitorChan:     make(chan corev1.PodPhase, 1),
		currentHotSwitchFaultPodId:  "",
		currentHotSwitchBackupPodId: "",
	}
	ctl.updatePodInfo()
	var rules []common.TransRule = ctl.getBaseRules()
	ctl.state = common.NewStateMachine(common.InitState, rules)
	ctl.controllerContext, ctl.ctxCancelFunc = context.WithCancel(ctl.serviceContext)
	return ctl
}

func (ctl *EventController) updatePodInfo() {
	pods := pod.GetPodByJobId(ctl.jobInfo.JobId)
	for _, pod := range pods {
		if pod.Name == "" {
			hwlog.RunLog.Warnf("job pod %s has empty name", ctl.jobInfo.JobId)
			continue
		}
		podRankIndex := pod.Annotations[constant.PodRankIndexAnno]
		ctl.prePodForScale[podRankIndex] = string(pod.UID)
		ctl.originPod[podRankIndex] = string(pod.UID)
	}
}

// GetFaultPod get fault pod
func (ctl *EventController) GetFaultPod() map[string]string {
	ctl.lock.RLock()
	defer ctl.lock.RUnlock()
	faultMap := make(map[string]string, len(ctl.faultPod))
	for k, v := range ctl.faultPod {
		faultMap[k] = v
	}
	return faultMap
}

func (ctl *EventController) mergeFaultPod(faultPod map[string]string) {
	ctl.lock.Lock()
	defer ctl.lock.Unlock()
	for podRank, podId := range faultPod {
		if _, ok := ctl.faultPod[podRank]; !ok {
			ctl.faultPod[podRank] = podId
		}
	}
}

func (ctl *EventController) saveCacheFault(faults []*pb.FaultRank) {
	mergedFaults := common.RemoveSliceDuplicateFaults(faults)
	hwlog.RunLog.Infof("jobId=%s, before append new Fault, normalFaults=%s, retryFaults=%s",
		ctl.jobInfo.JobId, common.Faults2String(ctl.cacheNormalFault), common.Faults2String(ctl.cacheRetryFault))
	for _, fault := range mergedFaults {
		if fault.FaultType == constant.NormalFaultType {
			ctl.cacheNormalFault = append(ctl.cacheNormalFault, fault)
		} else {
			ctl.cacheRetryFault = append(ctl.cacheRetryFault, fault)
		}
	}
	ctl.cacheNormalFault = common.RemoveSliceDuplicateFaults(ctl.cacheNormalFault)
	ctl.cacheRetryFault = common.RemoveSliceDuplicateFaults(ctl.cacheRetryFault)
	hwlog.RunLog.Infof("jobId=%s, after append new Fault, normalFaults=%s, retryFaults=%s",
		ctl.jobInfo.JobId, common.Faults2String(ctl.cacheNormalFault), common.Faults2String(ctl.cacheRetryFault))
}

func (ctl *EventController) reset(stop bool) {
	hwlog.RunLog.Infof("jobId=%s enter reset function", ctl.jobInfo.JobId)
	ctl.lock.Lock()
	defer ctl.lock.Unlock()
	hwlog.RunLog.Infof("jobId=%s's action path = {%s}", ctl.jobInfo.JobId, ctl.state.GetPathGraph())
	cm, err := common.RetryWriteResetCM(ctl.jobInfo.JobName, ctl.jobInfo.Namespace,
		nil, false, constant.ClearOperation)
	if err != nil {
		hwlog.RunLog.Errorf("clear reset configmap error, err=%v", err)
	} else {
		hwlog.RunLog.Infof("clear reset configmap success, %s", cm.Data[constant.ResetInfoCMDataKey])
	}
	if ctl.ctxCancelFunc != nil {
		ctl.ctxCancelFunc()
	}
	ctl.faultFlushing = false
	ctl.restartFaultProcess = false
	ctl.recoverInPlacePodFaults = make(map[string]*constant.PodFaultInfo)
	ctl.uuid = ""
	ctl.latestStrategy = ctl.latestStrategy[:0]
	ctl.faultPod = make(map[string]string)
	ctl.updatePodInfo()
	if !ctl.isChanClosed {
		ctl.closeControllerChan()
		ctl.isChanClosed = true
	}
	ctl.ctlResetTime = time.Now().UnixMilli()
	if stop {
		return
	}
	ctl.initControllerChan()
	ctl.cleanControllerSlice()
	ctl.cleanControllerMapAndSet()
	ctl.healthState = constant.HealthyState
	ctl.platStrategy = ""
	ctl.state.Reset()
	ctl.controllerContext, ctl.ctxCancelFunc = context.WithCancel(ctl.serviceContext)
	ctl.isChanClosed = false
	ctl.currentHotSwitchFaultPodId = ""
	ctl.currentHotSwitchBackupPodId = ""
	go ctl.listenEvent()
	go ctl.keepAlive()
}

func (ctl *EventController) cleanControllerMapAndSet() {
	ctl.stressTestParam = make(common.StressTestParam)
	ctl.isolateNodes = make(sets.String)
}

func (ctl *EventController) initControllerChan() {
	ctl.events = make(chan string, eventChanLength)
	ctl.signalChan = make(chan *pb.ProcessManageSignal, 1)
	ctl.reportStopCompleteChan = make(chan *pb.StopCompleteRequest, 1)
	ctl.reportRecoverStrategyChan = make(chan *pb.RecoverStrategyRequest, 1)
	ctl.reportStatusChan = make(chan *pb.RecoverStatusRequest, 1)
	ctl.newPodStatusMonitorChan = make(chan corev1.PodPhase, 1)
	ctl.scheduleResultChan = make(chan bool, 1)
	ctl.switchNicResponse = make(chan *pb.SwitchNicResponse, eventChanLength)
	ctl.switchRankList = make(chan *pb.SwitchRankList, 1)
	ctl.switchRankResult = make(chan *pb.SwitchResult, 1)
	ctl.stressTestNotifyChan = make(chan *pb.StressTestRankParams, 1)
	ctl.stressTestResponse = make(chan *pb.StressTestResponse, eventChanLength)
	ctl.stressTestResult = make(chan *pb.StressTestResult, 1)
}

func (ctl *EventController) cleanControllerSlice() {
	ctl.cacheRetryFault = ctl.cacheRetryFault[:0]
	ctl.cacheNormalFault = ctl.cacheNormalFault[:0]
	ctl.latestRecoverResult = ctl.latestRecoverResult[:0]
	ctl.agentReportStrategies = ctl.agentReportStrategies[:0]
	ctl.globalSwitchRankIDs = ctl.globalSwitchRankIDs[:0]
	ctl.globalOps = ctl.globalOps[:0]
}

func (ctl *EventController) closeControllerChan() {
	close(ctl.events)
	close(ctl.signalChan)
	close(ctl.reportStopCompleteChan)
	close(ctl.reportRecoverStrategyChan)
	close(ctl.reportStatusChan)
	close(ctl.newPodStatusMonitorChan)
	close(ctl.scheduleResultChan)
	close(ctl.switchNicResponse)
	close(ctl.switchRankList)
	close(ctl.switchRankResult)
	close(ctl.stressTestNotifyChan)
	close(ctl.stressTestResponse)
	close(ctl.stressTestResult)
}

func (ctl *EventController) selectKeepAlive(ctx context.Context, sendChan chan *pb.ProcessManageSignal) bool {
	if sendChan == nil {
		return true
	}
	select {
	case <-ctx.Done():
		return true
	case <-time.After(time.Duration(ctl.keepAliveSecond) * time.Second):
		signal := &pb.ProcessManageSignal{
			Uuid:       ctl.uuid,
			JobId:      ctl.jobInfo.JobId,
			SignalType: constant.KeepAliveSignalType,
		}

		var sendResult bool
		func() {
			defer func() {
				if r := recover(); r != nil {
					sendResult = true
				}
			}()
			select {
			case sendChan <- signal:
				sendResult = false
			case <-ctx.Done():
				sendResult = true
			}
		}()

		return sendResult
	}
}

func (ctl *EventController) keepAlive() {
	hwlog.RunLog.Infof("listen new keep-alive, jobId=%s", ctl.jobInfo.JobId)
	ctx, sendChan := ctl.getCtxAndSignalChan()
	exit := false
	for {
		exit = ctl.selectKeepAlive(ctx, sendChan)
		if exit {
			break
		}
	}
}

func (ctl *EventController) annotationWithRetryStrategy() bool {
	return slices.Contains(ctl.jobInfo.MindXConfigStrategies, constant.ProcessRetryStrategyName)
}

func (ctl *EventController) hasRecoverInPlaceStrategy() bool {
	return util.IsSliceContain(constant.ProcessRecoverInPlaceStrategyName, ctl.jobInfo.MindXConfigStrategies)
}

func (ctl *EventController) supportRetryStrategy() bool {
	mindXConfiged := slices.Contains(ctl.jobInfo.MindXConfigStrategies, constant.ProcessRetryStrategyName)
	if !mindXConfiged || !ctl.jobInfo.PlatFormMode {
		return mindXConfiged
	}
	return ctl.platStrategy == constant.ProcessRetryStrategyName
}

func (ctl *EventController) supportRecoverStrategy() bool {
	return ctl.supportTargetRecoverStrategy(constant.ProcessRecoverStrategyName)
}

func (ctl *EventController) supportTargetRecoverStrategy(recoverStrategy string) bool {
	mindXConfiged := slices.Contains(ctl.jobInfo.MindXConfigStrategies, recoverStrategy)
	if !mindXConfiged {
		return false
	}
	agentSupport := ctl.agentSupportStrategy(constant.ProcessRecoverStrategyName)
	if !agentSupport || !ctl.jobInfo.PlatFormMode {
		return agentSupport
	}
	return ctl.platStrategy == constant.ProcessRecoverStrategyName
}

func (ctl *EventController) supportTargetStrategy(recoverStrategy string) bool {
	if !ctl.configTargetStrategy(recoverStrategy) {
		return false
	}
	if recoverStrategy == constant.ElasticTrainingStrategyName {
		recoverStrategy = constant.ScaleInStrategyName
	}
	for _, strategy := range ctl.agentReportStrategies {
		if strategy == recoverStrategy {
			return true
		}
	}
	return false
}

func (ctl *EventController) configTargetStrategy(recoverStrategy string) bool {
	configed := false
	for _, strategy := range ctl.jobInfo.MindXConfigStrategies {
		if strategy == recoverStrategy {
			configed = true
			break
		}
	}
	return configed
}

func (ctl *EventController) supportRestartProcessStrategy() bool {
	return ctl.restartFaultProcess && ctl.supportTargetRecoverStrategy(constant.ProcessRecoverInPlaceStrategyName)
}

func (ctl *EventController) supportDumpStrategy() bool {
	mindXConfiged := slices.Contains(ctl.jobInfo.MindXConfigStrategies, constant.ProcessDumpStrategyName)
	if !mindXConfiged {
		return false
	}
	agentSupport := ctl.agentSupportStrategy(constant.ProcessDumpStrategyName)
	if !ctl.jobInfo.PlatFormMode || !agentSupport {
		return agentSupport
	}
	return ctl.platStrategy == constant.ProcessDumpStrategyName
}

func (ctl *EventController) onlySupportDumpStrategy() bool {
	if !ctl.jobInfo.ProcessRecoverEnable {
		hwlog.RunLog.Infof("jobId=%s ProcessRecoverEnable=%v, should not dump",
			ctl.jobInfo.JobId, ctl.jobInfo.ProcessRecoverEnable)
		return false
	}
	if ctl.jobInfo.PlatFormMode {
		if ctl.platStrategy != constant.ProcessDumpStrategyName {
			hwlog.RunLog.Infof("jobId=%s platStrategy=%v, should not dump", ctl.jobInfo.JobId, ctl.platStrategy)
			return false
		}
		if !util.IsSliceContain(ctl.platStrategy, ctl.jobInfo.MindXConfigStrategies) {
			hwlog.RunLog.Infof("jobId=%s strategy=%v should not dump",
				ctl.jobInfo.JobId, ctl.jobInfo.MindXConfigStrategies)
			return false
		}
		return true
	}
	// MindXConfigStrategies have been sorted by priority defined by recoverStrategyPriorityMap
	mindXConfiged := len(ctl.jobInfo.MindXConfigStrategies) > 0 &&
		ctl.jobInfo.MindXConfigStrategies[0] == constant.ProcessDumpStrategyName
	if !mindXConfiged {
		hwlog.RunLog.Infof("jobId=%s strategy=%v not only support dump",
			ctl.jobInfo.JobId, ctl.jobInfo.MindXConfigStrategies)
		return false
	}
	return true
}

func (ctl *EventController) shouldDumpWhenOccurFault() bool {
	if !ctl.onlySupportDumpStrategy() {
		hwlog.RunLog.Infof("jobId=%s config not only support dump strategy, should not dump",
			ctl.jobInfo.JobId)
		return false
	}
	if ctl.healthState == constant.UnHealthyState ||
		(ctl.healthState == constant.SubHealthyState && ctl.jobInfo.GraceExit) {
		return true
	}
	hwlog.RunLog.Infof("jobId=%s healthState=%v graceExit=%v, should not dump",
		ctl.jobInfo.JobId, ctl.healthState, ctl.jobInfo.GraceExit)
	return false
}

func (ctl *EventController) addEvent(event string) {
	if !ctl.state.RuleCheck(ctl.state.GetState(), event) {
		hwlog.RunLog.Warnf("event add fail, order mix, jobId=%s, uuid=%s, event=%s",
			ctl.jobInfo.JobId, ctl.uuid, event)
		return
	}
	ctx, ch := ctl.getCtxAndEventChan()
	if ch == nil {
		hwlog.RunLog.Errorf("jobId=%s, event chan is nil", ctl.jobInfo.JobId)
		return
	}
	defer catchException()
	select {
	case <-ctx.Done():
		hwlog.RunLog.Warnf("event add fail, controller context canceled, jobId=%s, uuid=%s, event=%s",
			ctl.jobInfo.JobId, ctl.uuid, event)
	case ch <- event:
		hwlog.RunLog.Infof("jobId=%s uuid=%s state is %s, event=%s enqueue success",
			ctl.jobInfo.JobId, ctl.uuid, ctl.state.GetState(), event)
	default:
		hwlog.RunLog.Infof("add event=%s timeout, reset state machine", event)
		ctl.reset(false)
	}
}

func (ctl *EventController) getCtxAndEventChan() (context.Context, chan string) {
	ctl.lock.RLock()
	defer ctl.lock.RUnlock()
	return ctl.controllerContext, ctl.events
}

func (ctl *EventController) getCtlResetTime() int64 {
	ctl.lock.RLock()
	defer ctl.lock.RUnlock()
	return ctl.ctlResetTime
}

func (ctl *EventController) selectEventChan(ctx context.Context, eventChan chan string) bool {
	if eventChan == nil {
		return true
	}
	select {
	case <-ctx.Done():
		hwlog.RunLog.Infof("controller context done, break listen event, jobId=%s", ctl.jobInfo.JobId)
		return true
	case event, ok := <-eventChan:
		if ok {
			nextEvent, code, err := ctl.trigger(event)
			hwlog.RunLog.Infof("jobId=%s's action path = {%s}", ctl.jobInfo.JobId, ctl.state.GetPathGraph())
			if err != nil {
				hwlog.RunLog.Errorf("jobId=%s trigger error, code=%d, err=%v", ctl.jobInfo.JobId, code, err)
				ctl.reset(false)
				return true
			}
			if nextEvent != "" {
				ctl.addEvent(nextEvent)
			}
			return false
		} else {
			hwlog.RunLog.Infof("event channel closed, break listen event, jobId=%s", ctl.jobInfo.JobId)
			return true
		}
	}
}

func (ctl *EventController) listenEvent() {
	hwlog.RunLog.Infof("start listen a new event, jobId=%s", ctl.jobInfo.JobId)
	ctx, eventChan := ctl.getCtxAndEventChan()
	exit := false
	for {
		exit = ctl.selectEventChan(ctx, eventChan)
		if exit {
			break
		}
	}
}

func (ctl *EventController) getCtxAndSignalChan() (context.Context, chan *pb.ProcessManageSignal) {
	ctl.lock.RLock()
	defer ctl.lock.RUnlock()
	return ctl.controllerContext, ctl.signalChan
}

func (ctl *EventController) handleSendResult(signal *pb.ProcessManageSignal, err error) {
	if signal.SignalType == constant.KeepAliveSignalType {
		return
	}
	if signal.SignalType == constant.WaitStartAgentSignalType || signal.SignalType == constant.PreExitProcessSignalType {
		return
	}
	if signal.SignalType == constant.KillMasterSignalType {
		ctl.addEvent(common.FinishEvent)
		return
	}
	if signal.ExtraParams == constant.DumpExit {
		return
	}
	if err != nil {
		defer catchException()
		ctl.replyOMResponse("om failed, send signal failed")
		ctl.addEvent(common.NotifyFailEvent)
		return
	}
	if signal.SignalType == constant.FaultNodesExitSignalType {
		ctl.addEvent(common.NotifyFaultNodesExitSuccessEvent)
		return
	}
	if signal.SignalType != constant.ChangeStrategySignalType {
		ctl.addEvent(common.NotifySuccessEvent)
		return
	}
	if event, ok := notifyStrategySuccessEventMap[signal.ChangeStrategy]; ok {
		ctl.addEvent(event)
	} else {
		hwlog.RunLog.Errorf("unsupported strategy=%s, jobId=%s",
			signal.ChangeStrategy, signal.JobId)
	}
}

func (ctl *EventController) selectSendChannel(ctx context.Context, sendChan chan *pb.ProcessManageSignal,
	stream pb.Recover_SubscribeProcessManageSignalServer) bool {
	if sendChan == nil || stream == nil {
		return true
	}
	select {
	case <-ctx.Done():
		hwlog.RunLog.Infof("context done, jobId=%s break listen sendChan", ctl.jobInfo.JobId)
		return true
	case <-stream.Context().Done():
		ctl.reset(false)
		hwlog.RunLog.Infof("stream context done, jobId=%s break listen sendChan", ctl.jobInfo.JobId)
		return true
	case signal, ok := <-sendChan:
		if ok {
			err := common.SendRetry(stream, signal, retryTimes)
			ctl.handleSendResult(signal, err)
			return false
		} else {
			hwlog.RunLog.Infof("sendChan closed, jobId=%s break listen sendChan", ctl.jobInfo.JobId)
			return true
		}
	}
}

func (ctl *EventController) listenSendChannel(stream pb.Recover_SubscribeProcessManageSignalServer) {
	ctl.reset(false)
	ctx, sendChan := ctl.getCtxAndSignalChan()
	hwlog.RunLog.Infof("start listen a new send channel, jobId=%s", ctl.jobInfo.JobId)
	exit := false
	for {
		exit = ctl.selectSendChannel(ctx, sendChan, stream)
		if exit {
			break
		}
	}
}

func (ctl *EventController) signalEnqueue(signal *pb.ProcessManageSignal) (string, common.RespCode, error) {
	ctx, sendChan := ctl.getCtxAndSignalChan()
	if sendChan == nil {
		hwlog.RunLog.Errorf("jobId=%s, sendChan is nil", ctl.jobInfo.JobId)
		return "", common.SignalQueueBusy, errors.New("sendChan is nil")
	}
	defer catchException()
	select {
	case sendChan <- signal:
		hwlog.RunLog.Infof("signal enqueue, jobId=%s, uuid=%s, signalType=%s, strategy=%s, faults=%s",
			signal.JobId, signal.Uuid, signal.SignalType, signal.ChangeStrategy,
			common.Faults2String(signal.FaultRanks))
		return "", common.OK, nil
	case <-ctx.Done():
		hwlog.RunLog.Warnf("controller context canceled, jobId=%s, uuid=%s", ctl.jobInfo.JobId, ctl.uuid)
		return "", common.ControllerEventCancel, nil
	case <-time.After(time.Second):
		info := fmt.Sprintf("add signal time-out for jobId=%s, "+
			"program may running in chaos", signal.JobId)
		hwlog.RunLog.Error(info)
		return "", common.SignalQueueBusy, errors.New(info)
	}
}

func (ctl *EventController) trigger(event string) (string, common.RespCode, error) {
	return ctl.state.Trigger(event)
}

func (ctl *EventController) handleFinish() (string, common.RespCode, error) {
	ctl.reset(false)
	return "", common.OK, nil
}

func (ctl *EventController) handleNotifyWaitFaultFlushing() (string, common.RespCode, error) {
	if ctl.jobInfo.PlatFormMode {
		hwlog.RunLog.Infof("start wait plat strategy ready, jobId=%s, pgName=%s",
			ctl.jobInfo.JobId, ctl.jobInfo.PgName)
		strategy, err := WaitPlatFormStrategyReady(ctl.jobInfo.PgName, ctl.jobInfo.Namespace)
		hwlog.RunLog.Infof("finish wait plat strategy, jobId=%s, pgName=%s, strategy=%s, err=%v",
			ctl.jobInfo.JobId, ctl.jobInfo.PgName, strategy, err)
		if err != nil {
			return common.WaitPlatStrategyTimeoutEvent, common.WaitPlatStrategyTimeout, nil
		}
		ctl.platStrategy = strategy
	}
	if ctl.shouldDumpWhenOccurFault() {
		hwlog.RunLog.Infof("should dump, jobId=%s, strategies=%v, platFormMode=%v, plat strategy=%s",
			ctl.jobInfo.JobId, ctl.jobInfo.MindXConfigStrategies, ctl.jobInfo.PlatFormMode, ctl.platStrategy)
		ctl.agentReportStrategies = append(ctl.agentReportStrategies, constant.ProcessDumpStrategyName)
		return common.DumpForFaultEvent, common.OK, nil
	}
	cm, err := common.RetryWriteResetCM(ctl.jobInfo.JobName, ctl.jobInfo.Namespace, nil, false,
		constant.NotifyFaultFlushingOperation)
	if err != nil {
		hwlog.RunLog.Errorf("notify agent faultFlushing error, err=%v", err)
		return common.NotifyFailEvent, common.OperateConfigMapError, nil
	}
	hwlog.RunLog.Infof("write configmap FaultFlushing success, %s", cm.Data[constant.ResetInfoCMDataKey])
	return common.NotifyFinishEvent, common.OK, nil
}

func (ctl *EventController) handleFaultClear() (string, common.RespCode, error) {
	cm, err := common.RetryWriteResetCM(ctl.jobInfo.JobName, ctl.jobInfo.Namespace, nil, false,
		constant.ClearOperation)
	if err != nil {
		hwlog.RunLog.Errorf("clear reset configmap error, err=%v", err)
		return common.ClearConfigMapFaultFailEvent, common.OperateConfigMapError, nil
	}
	hwlog.RunLog.Infof("clear reset configmap success, %s", cm.Data[constant.ResetInfoCMDataKey])
	return common.ClearConfigMapFaultSuccessEvent, common.OK, nil
}

func (ctl *EventController) handleNotifyStopTrain() (string, common.RespCode, error) {
	ctl.uuid = common.NewEventId(randomLen)
	signal := &pb.ProcessManageSignal{
		Uuid:           ctl.uuid,
		JobId:          ctl.jobInfo.JobId,
		SignalType:     constant.StopTrainSignalType,
		Actions:        stopTrainActions,
		ChangeStrategy: "",
	}
	signal.FaultRanks = append(ctl.cacheRetryFault, ctl.cacheNormalFault...)
	return ctl.signalEnqueue(signal)
}

func (ctl *EventController) handleNotifyElagantDump() (string, common.RespCode, error) {
	if ctl.uuid == "" {
		ctl.uuid = common.NewEventId(randomLen)
	}
	signal := &pb.ProcessManageSignal{
		Uuid:           ctl.uuid,
		JobId:          ctl.jobInfo.JobId,
		SignalType:     constant.SaveAndExitSignalType,
		Actions:        saveAndExitActions,
		ChangeStrategy: "",
	}
	_, allFaultRanks, err := ctl.updateCacheFaultAndPod()
	if err != nil {
		hwlog.RunLog.Errorf("update cache info fail, jobId=%s err=%v", ctl.jobInfo.JobId, err)
		return "", common.ServerInnerError, err
	}
	signal.NodeRankIds, err = common.GetNodeRankIdsByRankIds(ctl.jobInfo.JobId, allFaultRanks)
	if err != nil {
		hwlog.RunLog.Errorf("jobId=%s, GetNodeRankIdsByRankIds err:%v", ctl.jobInfo.JobId, err)
		return common.NotifyFailEvent, common.OperateConfigMapError, nil
	}
	cm, err := common.WriteResetInfoToCM(ctl.jobInfo.JobName, ctl.jobInfo.Namespace,
		allFaultRanks, false, constant.NotifyFaultListOperation)
	if err != nil {
		err = fmt.Errorf("notify agent faultList error, jobId=%s, err=%v", ctl.jobInfo.JobId, err)
		hwlog.RunLog.Error(err)
		return "", common.OperateConfigMapError, err
	}
	hwlog.RunLog.Infof("write configmap faultList success, jobId=%s, cm data: %s", ctl.jobInfo.JobId,
		cm.Data[constant.ResetInfoCMDataKey])
	return ctl.signalEnqueue(signal)
}

func (ctl *EventController) getCtxAndStopCompleteChan() (context.Context, chan *pb.StopCompleteRequest) {
	ctl.lock.RLock()
	defer ctl.lock.RUnlock()
	return ctl.controllerContext, ctl.reportStopCompleteChan
}

func (ctl *EventController) handleWaitReportStopComplete() (string, common.RespCode, error) {
	ctx, reportChan := ctl.getCtxAndStopCompleteChan()
	if reportChan == nil {
		hwlog.RunLog.Infof("jobId=%s, reportChan is nil", ctl.jobInfo.JobId)
		return "", common.OK, fmt.Errorf("jobId=%s, reportChan is nil", ctl.jobInfo.JobId)
	}
	select {
	case <-ctx.Done():
		hwlog.RunLog.Warnf("controller context canceled, jobId=%s, uuid=%s",
			ctl.jobInfo.JobId, ctl.uuid)
		return "", common.ControllerEventCancel, nil
	case req := <-reportChan:
		if req.Status.Code == common.ProcessNotReady {
			return common.ProcessNotReadyEvent, common.ClientError, nil
		}
		return common.ReceiveReportEvent, common.OK, nil
	case <-time.After(time.Duration(reportTimeoutMinutes) * time.Minute):
		hwlog.RunLog.Errorf("wait report stop complete timeout, jobId=%s, uuid=%s", ctl.jobInfo.JobId, ctl.uuid)
		return common.ReportTimeoutEvent, common.WaitReportTimeout, nil
	}
}

func (ctl *EventController) handleWaitFlushFinish() (string, common.RespCode, error) {
	ctx, _ := ctl.getCtxAndEventChan()
	retryFaults, normalFaults := ctl.takeRetryFault2NormalFault()
	if len(retryFaults) > 0 && len(normalFaults) == 0 && ctl.annotationWithRetryStrategy() {
		hwlog.RunLog.Infof("jobId=%s occur uce error, will not sleep for fault flushing",
			ctl.jobInfo.JobId)
		return common.FaultFlushFinishedEvent, common.OK, nil
	}
	select {
	case <-time.After(time.Duration(faultFlushSeconds) * time.Second):
		return common.FaultFlushFinishedEvent, common.OK, nil
	case <-ctx.Done():
		return "", common.OK, nil
	}
}

func (ctl *EventController) normalFaultAssociateSameNodeRank() ([]*pb.FaultRank, []string) {
	var faultRankIds []string
	for _, fault := range ctl.cacheNormalFault {
		faultRankIds = append(faultRankIds, fault.RankId)
	}
	var restartRanks []string = nil
	if ctl.restartFaultProcess {
		faultRankIds, restartRanks = ctl.recoverInPlaceFaultAssociateSameNodeRank(faultRankIds)
	}
	allFaultRankIds := common.GetFaultRankIdsInSameNode(faultRankIds, pod.GetPodDeviceNumByJobId(ctl.jobInfo.JobId))
	allFaultRankIds = append(allFaultRankIds, restartRanks...)
	removeSameRankIds := util.RemoveSliceDuplicateElement(allFaultRankIds)
	var res []*pb.FaultRank
	for _, rank := range removeSameRankIds {
		res = append(res, &pb.FaultRank{
			RankId:    rank,
			FaultType: constant.NormalFaultType,
		})
	}
	return res, removeSameRankIds
}

func (ctl *EventController) writeConfirmFaultAndWaitPlatResultFault(faults []*pb.FaultRank) ([]*pb.FaultRank, error) {
	allFaultRanks := common.RemoveSliceDuplicateFaults(faults)
	err := UpdateProcessConfirmFault(ctl.jobInfo.PgName, ctl.jobInfo.Namespace, allFaultRanks)
	if err != nil {
		hwlog.RunLog.Errorf("update process confirm fault err: %v, jobId=%s", err, ctl.jobInfo.JobId)
		return nil, fmt.Errorf("update process confirm fault err: %v, jobId=%s", err, ctl.jobInfo.JobId)
	}
	hwlog.RunLog.Infof("start wait plat result fault, jobId=%s, pgName=%s",
		ctl.jobInfo.JobId, ctl.jobInfo.PgName)
	platFaultResult, err := WaitProcessResultFault(ctl.jobInfo.PgName, ctl.jobInfo.Namespace)
	hwlog.RunLog.Infof("finish wait plat result fault, jobId=%s, faults=%s, err=%v",
		ctl.jobInfo.JobId, common.Faults2String(platFaultResult), err)
	if err != nil {
		hwlog.RunLog.Errorf("wait process result fault err: %v, jobId=%s", err, ctl.jobInfo.JobId)
		return nil, fmt.Errorf("wait process result fault err: %v, jobId=%s", err, ctl.jobInfo.JobId)
	}
	allFaultRanks = common.RemoveSliceDuplicateFaults(append(allFaultRanks, platFaultResult...))
	ctl.platStrategy, err = platFormStrategy(ctl.jobInfo.PgName, ctl.jobInfo.Namespace, true)
	hwlog.RunLog.Infof("plat confirm strategy=%s, jobId=%s, err=%v", ctl.platStrategy, ctl.jobInfo.JobId, err)
	if err != nil {
		return nil, fmt.Errorf("confirm plat strategy err:%v", err)
	}
	return allFaultRanks, nil
}

func (ctl *EventController) takeRetryFault2NormalFault() ([]*pb.FaultRank, []*pb.FaultRank) {
	ctl.lock.Lock()
	defer ctl.lock.Unlock()
	n := len(ctl.latestRecoverResult)
	if n > 0 && ctl.latestRecoverResult[n-1].Strategy == constant.ProcessRetryStrategyName {
		ctl.cacheNormalFault = append(ctl.cacheNormalFault, ctl.cacheRetryFault...)
		ctl.cacheRetryFault = ctl.cacheRetryFault[:0]
	}
	if len(ctl.cacheNormalFault) > 0 {
		ctl.cacheNormalFault = append(ctl.cacheNormalFault, ctl.cacheRetryFault...)
		ctl.cacheRetryFault = ctl.cacheRetryFault[:0]
	}
	if !ctl.supportRetryStrategy() {
		ctl.cacheNormalFault = append(ctl.cacheNormalFault, ctl.cacheRetryFault...)
		ctl.cacheRetryFault = ctl.cacheRetryFault[:0]
	}
	if !ctl.hasSameRetryFault() {
		ctl.cacheNormalFault = append(ctl.cacheNormalFault, ctl.cacheRetryFault...)
		ctl.cacheRetryFault = ctl.cacheRetryFault[:0]
	}
	return ctl.cacheRetryFault, ctl.cacheNormalFault
}

func (ctl *EventController) setCacheFault(uceFaults, normalFaults []*pb.FaultRank) {
	ctl.lock.Lock()
	defer ctl.lock.Unlock()
	ctl.cacheRetryFault = uceFaults
	ctl.cacheNormalFault = normalFaults
}

func (ctl *EventController) notifyFaultForRetryFaultCase(retryFaults,
	normalFaults []*pb.FaultRank) (string, common.RespCode, error) {
	hwlog.RunLog.Infof("jobId=%s enter notifyFaultForRetryFaultCase function", ctl.jobInfo.JobId)
	signal := &pb.ProcessManageSignal{
		Uuid:       ctl.uuid,
		JobId:      ctl.jobInfo.JobId,
		SignalType: constant.GlobalFaultSignalType,
		Actions:    globalFaultActions,
	}
	if ctl.jobInfo.PlatFormMode {
		return ctl.notifyFaultForRetryFaultCasePlatFormMode(signal, retryFaults, normalFaults)
	}
	hwlog.RunLog.Infof("jobId=%s, uce error case", ctl.jobInfo.JobId)
	signal.FaultRanks = retryFaults
	signal = ctl.notifyHCCLRoutingTimeout(signal)
	return ctl.signalEnqueue(signal)
}

func (ctl *EventController) notifyFaultForRetryFaultCasePlatFormMode(signal *pb.ProcessManageSignal, retryFaults,
	normalFaults []*pb.FaultRank) (string, common.RespCode, error) {
	allFaults, err := ctl.writeConfirmFaultAndWaitPlatResultFault(retryFaults)
	if err != nil {
		hwlog.RunLog.Errorf("interacte with plat error, err=%v", err)
		return common.WriteConfirmFaultOrWaitResultFaultTimeoutEvent,
			common.WriteConfirmFaultOrWaitPlatResultFault, nil
	}
	hwlog.RunLog.Infof("jobId=%s, plat merge faults=%s", ctl.jobInfo.JobId, common.Faults2String(allFaults))
	if !common.IsRetryFault(allFaults) {
		retryFaults = retryFaults[:0]
		_, allFaultRanks, err := ctl.updateCacheFaultAndPod()
		if err != nil {
			hwlog.RunLog.Errorf("update cache info fail, jobId=%s err=%v", ctl.jobInfo.JobId, err)
			return "", common.ServerInnerError, err
		}
		// Note: RetryWriteResetCM is only used for exiting processes in elastic scenarios.
		// The elastic component will be sunset in the future.
		cm, err := common.RetryWriteResetCM(ctl.jobInfo.JobName, ctl.jobInfo.Namespace,
			allFaultRanks, ctl.restartFaultProcess, constant.NotifyFaultListOperation)
		if err != nil {
			hwlog.RunLog.Errorf("notify agent faultList error, err=%v", err)
			return common.NotifyFailEvent, common.OperateConfigMapError, nil
		}
		signal.FaultRanks = normalFaults
		signal.NodeRankIds, err = common.GetNodeRankIdsByRankIds(ctl.jobInfo.JobId, allFaultRanks)
		if err != nil {
			hwlog.RunLog.Errorf("jobId=%s, GetNodeRankIdsByRankIds err:%v", ctl.jobInfo.JobId, err)
			return common.NotifyFailEvent, common.OperateConfigMapError, nil
		}
		if !ctl.restartFaultProcess {
			signal.Actions = append(signal.Actions, constant.FaultNodesExitAction)
			ctl.filterPodsNodeRankIds(signal)
		} else {
			signal.Actions = append(signal.Actions, constant.FaultNodesRestartAction)
		}
		hwlog.RunLog.Infof("write configmap faultList success, %s", cm.Data[constant.ResetInfoCMDataKey])
	} else {
		hwlog.RunLog.Infof("jobId=%s, uce error case", ctl.jobInfo.JobId)
		signal.FaultRanks = retryFaults
	}
	signal = ctl.notifyHCCLRoutingTimeout(signal)
	return ctl.signalEnqueue(signal)
}

func (ctl *EventController) notifyFaultForNormalFaultCase(uceFaults, normalFaults []*pb.FaultRank) (
	string, common.RespCode, error) {
	hwlog.RunLog.Infof("jobId=%s enter notifyFaultForNormalFaultCase function", ctl.jobInfo.JobId)
	if ctl.jobInfo.PlatFormMode {
		hwlog.RunLog.Infof("jobId=%s enter notifyFaultForNormalFaultCase function", ctl.jobInfo.JobId)
		allFaults, err := ctl.writeConfirmFaultAndWaitPlatResultFault(normalFaults)
		if err != nil {
			hwlog.RunLog.Errorf("write confirm fault or wait plat result fault timeout, err=%v", err)
			return common.WriteConfirmFaultOrWaitResultFaultTimeoutEvent,
				common.WriteConfirmFaultOrWaitPlatResultFault, nil
		}
		hwlog.RunLog.Infof("jobId=%s, plat merge faults=%s", ctl.jobInfo.JobId, common.Faults2String(allFaults))
		ctl.setCacheFault(nil, allFaults)
		normalFaults = allFaults
	}
	allFaults, allFaultRanks, err := ctl.updateCacheFaultAndPod()
	if err != nil {
		hwlog.RunLog.Errorf("update cache info fail, jobId=%s err=%v", ctl.jobInfo.JobId, err)
		return "", common.ServerInnerError, err
	}
	signal := &pb.ProcessManageSignal{
		Uuid:           ctl.uuid,
		JobId:          ctl.jobInfo.JobId,
		SignalType:     constant.GlobalFaultSignalType,
		Actions:        globalFaultActions,
		ChangeStrategy: "",
	}
	signal.FaultRanks = append(signal.FaultRanks, allFaults...)
	hwlog.RunLog.Infof("jobId=%s write reset json, restartFaultProcess: %v, faultRanks: %v",
		ctl.jobInfo.JobId, ctl.restartFaultProcess, allFaultRanks)
	// Note: WriteResetInfoToCM is only used for exiting processes in elastic scenarios.
	// The elastic component will be sunset in the future.
	cm, err := common.WriteResetInfoToCM(ctl.jobInfo.JobName, ctl.jobInfo.Namespace,
		allFaultRanks, ctl.restartFaultProcess, constant.NotifyFaultListOperation)
	if err != nil {
		hwlog.RunLog.Errorf("notify agent faultList error, err=%v", err)
		return common.NotifyFailEvent, common.OperateConfigMapError, nil
	}
	hwlog.RunLog.Infof("write configmap faultList success, %s", cm.Data[constant.ResetInfoCMDataKey])
	signal.Actions = append(signal.Actions, constant.FaultNodesExitAction)
	signal.NodeRankIds, err = common.GetNodeRankIdsByRankIds(ctl.jobInfo.JobId, allFaultRanks)
	if err != nil {
		hwlog.RunLog.Errorf("jobId=%s, GetNodeRankIdsByRankIds err:%v", ctl.jobInfo.JobId, err)
		return common.NotifyFailEvent, common.OperateConfigMapError, nil
	}
	if ctl.restartFaultProcess {
		_, signal.NodeRankIds = ctl.splitRecoverInPlacePodInfoToRestartProcessOrRestartPodList()
	}
	ctl.filterPodsNodeRankIds(signal)
	return ctl.signalEnqueue(signal)
}

func (ctl *EventController) updateCacheFaultAndPod() ([]*pb.FaultRank, []string, error) {
	allFaults, allFaultRanks, err := ctl.getAllFaultRanks()
	if err != nil {
		return allFaults, allFaultRanks, err
	}
	ctl.setCacheFault(nil, allFaults)
	faultPod, err := common.GetPodMap(ctl.jobInfo.JobId, allFaultRanks)
	if err != nil {
		hwlog.RunLog.Errorf("jobId=%s, get pod map err:%v", ctl.jobInfo.JobId, err)
		return allFaults, allFaultRanks, err
	}
	ctl.mergeFaultPod(faultPod)
	hwlog.RunLog.Infof("jobId=%s, fault pod = %v", ctl.jobInfo.JobId, ctl.GetFaultPod())
	return allFaults, allFaultRanks, nil
}

func (ctl *EventController) handleNotifyGlobalFault() (string, common.RespCode, error) {
	if !job.GetJobIsExists(ctl.jobInfo.JobId) {
		return "", common.JobNotExist, fmt.Errorf("jobId=%s not exist", ctl.jobInfo.JobId)
	}

	retryFaults, normalFaults := ctl.takeRetryFault2NormalFault()
	// if len(ctl.cacheRetryFault) still bigger than 0 after takeRetryFault2NormalFault
	// that means job support retry strategy, and it's first time choose strategy case only have uce fault
	if len(retryFaults) > 0 {
		return ctl.notifyFaultForRetryFaultCase(retryFaults, normalFaults)
	}
	if isReleased := ctl.dealWithForceRelease(); !isReleased {
		// "sleep" is used to ensure that the time taken is consistent with the situation of resource release
		time.Sleep(constant.NotReleaseSleepTime)
		hwlog.RunLog.Infof("jobId=%s do not release sources", ctl.jobInfo.JobId)
	}
	// strategy includes recover-in-place and fault can be restored in place, then check if fault disappears.
	// If fault does not disappear after 15 seconds, can not choose recover-in-place strategy
	ctl.updateRestartFaultProcess()
	if ctl.restartFaultProcess && ctl.hasRecoverInPlaceStrategy() {
		restartPods := ctl.waitNormalFaultRecovery()
		hwlog.RunLog.Warnf("jobId:%s restartPods:%v", ctl.jobInfo.JobId, restartPods)
		ctl.updateRecoverInPlaceInfoByRestartPods(restartPods)
		faultmanager.CallbackForReportNoRetryInfo(ctl.jobInfo.JobId, restartPods, time.Now().UnixMilli())
	}
	return ctl.notifyFaultForNormalFaultCase(retryFaults, normalFaults)
}

func (ctl *EventController) dealWithForceRelease() bool {
	if ctl.isA5Job() {
		hwlog.RunLog.Infof("jobId=%s is npu job, skip SDID-based force release", ctl.jobInfo.JobId)
		return true
	}
	hwlog.RunLog.Infof("jobId=%s force release", ctl.jobInfo.JobId)
	faultPods := ctl.GetFaultPod()
	hwlog.RunLog.Infof("jobId=%s force release nodes %v", ctl.jobInfo.JobId, faultPods)
	faultInfos := job.GetJobFaultSdIdAndNodeName(ctl.jobInfo.JobId, faultPods)
	if faultInfos == nil {
		hwlog.RunLog.Warnf("jobId=%s force release faultInfos is nil ", ctl.jobInfo.JobId)
		return false
	}
	kube.CreateOrUpdateSuperPodFaultInfo(ctl.jobInfo.JobId, faultInfos)
	// wait noded force release ipc resource
	time.Sleep(constant.ReleaseTimeOut)
	hwlog.RunLog.Warnf("jobId=%s force release finish", ctl.jobInfo.JobId)
	return true
}

func (ctl *EventController) isA5Job() bool {
	pods := pod.GetPodByJobId(ctl.jobInfo.JobId)
	clusterDevices := superpod.ListClusterDevice()
	for _, podInfo := range pods {
		if podInfo.Spec.NodeName == "" {
			continue
		}
		if isNodeDeviceA5(podInfo.Spec.NodeName, clusterDevices) {
			return true
		}
	}
	return false
}
func isNodeDeviceA5(nodeName string, clusterDevices []*api.SuperPodDevice) bool {
	for _, superPodDevice := range clusterDevices {
		if nodeDevice, exists := superPodDevice.NodeDeviceMap[nodeName]; exists {
			return nodeDevice.ServerType == api.VersionNPU
		}
	}
	return false
}

func (ctl *EventController) firstChooseStrategy() string {
	hwlog.RunLog.Infof("first choose strategy, jobId=%s, configed: %v, reported: %v", ctl.jobInfo.JobId,
		ctl.jobInfo.MindXConfigStrategies, ctl.agentReportStrategies)
	if ctl.supportRetryStrategy() && len(ctl.cacheNormalFault) <= 0 {
		return constant.ProcessRetryStrategyName
	}
	if ctl.supportRestartProcessStrategy() {
		return constant.ProcessRecoverInPlaceStrategyName
	}
	return ctl.chooseForRestartProcessFail()
}

func (ctl *EventController) chooseForRestartProcessFail() string {
	if ctl.supportRecoverStrategy() {
		return constant.ProcessRecoverStrategyName
	}
	if ctl.canChooseScaleInStrategy() {
		return constant.ScaleInStrategyName
	}
	if ctl.supportDumpStrategy() {
		return constant.ProcessDumpStrategyName
	}
	return constant.ProcessExitStrategyName
}

func (ctl *EventController) canChooseScaleInStrategy() bool {
	if _, ok := ctl.faultPod[constant.RankZeroNodeId]; ok {
		return false
	}
	if ctl.supportTargetStrategy(constant.ElasticTrainingStrategyName) && !ctl.configTargetStrategy(
		constant.JobReschedulingStrategyName) && !ctl.configTargetStrategy(
		constant.PodReschedulingStrategyName) {
		return true
	} else if ctl.supportTargetStrategy(constant.ElasticTrainingStrategyName) && (ctl.configTargetStrategy(
		constant.JobReschedulingStrategyName) || ctl.configTargetStrategy(
		constant.PodReschedulingStrategyName)) && !ctl.whetherHasEnoughResource() {
		return true
	}
	return false
}

func (ctl *EventController) chooseForRetryFail() string {
	if ctl.supportRestartProcessStrategy() {
		return constant.ProcessRecoverInPlaceStrategyName
	}
	if ctl.supportRecoverStrategy() {
		return constant.ProcessRecoverStrategyName
	}
	if ctl.canChooseScaleInStrategy() {
		return constant.ScaleInStrategyName
	}
	if ctl.supportDumpStrategy() {
		return constant.ProcessDumpStrategyName
	}
	return constant.ProcessExitStrategyName
}

func (ctl *EventController) chooseForRecoverFail() string {
	if ctl.supportDumpStrategy() {
		return constant.ProcessDumpStrategyName
	}
	return constant.ProcessExitStrategyName
}

func (ctl *EventController) agentSupportStrategy(strategy string) bool {
	return slices.Contains(ctl.agentReportStrategies, strategy)
}

func (ctl *EventController) chooseStrategy() (string, error) {
	ctl.lock.RLock()
	n := len(ctl.latestRecoverResult)
	ctl.lock.RUnlock()
	if n == 0 {
		strategy := ctl.firstChooseStrategy()
		if strategy == constant.ProcessRetryStrategyName &&
			!ctl.agentSupportStrategy(constant.ProcessRetryStrategyName) {
			hwlog.RunLog.Warnf("uce repair not enabled by controller, jobId=%s", ctl.jobInfo.JobId)
			ctl.takeRetryFault2NormalFault()
			_, _, err := ctl.updateCacheFaultAndPod()
			if err != nil {
				hwlog.RunLog.Errorf("update cache info fail, jobId=%s err=%v", ctl.jobInfo.JobId, err)
				return "", err
			}
			return ctl.chooseForRecoverFail(), nil // dump or exit
		}
		return strategy, nil
	}
	res := ctl.latestRecoverResult[n-1]
	if res.Strategy == constant.ProcessRetryStrategyName {
		return ctl.chooseForRetryFail(), nil
	} else if res.Strategy == constant.ProcessRecoverStrategyName || res.Strategy == constant.ScaleInStrategyName ||
		res.Strategy == constant.ScaleOutStrategyName {
		return ctl.chooseForRecoverFail(), nil
	}
	return constant.ProcessExitStrategyName, nil
}

func (ctl *EventController) handleNotifyDecidedStrategy() (string, common.RespCode, error) {
	signal := &pb.ProcessManageSignal{
		Uuid:       ctl.uuid,
		JobId:      ctl.jobInfo.JobId,
		SignalType: constant.ChangeStrategySignalType,
		Actions:    changeStrategyActions,
		FaultRanks: ctl.cacheNormalFault,
	}
	var err error
	signal.ChangeStrategy, err = ctl.chooseStrategy()
	if err != nil {
		hwlog.RunLog.Errorf("jobId=%s, chooseStrategy err:%v", ctl.jobInfo.JobId, err)
		return "", common.ServerInnerError, err
	}
	return ctl.handleDecidedStrategyAfterChoose(signal)
}

func (ctl *EventController) handleDecidedStrategyAfterChoose(signal *pb.ProcessManageSignal) (string,
	common.RespCode, error) {
	if signal.ChangeStrategy == constant.ProcessDumpStrategyName &&
		ctl.isSwitchingNic() && ctl.jobInfo.Framework == constant.MsFramework {
		// In order to correctly switch from the state machine of mindIO to the state machine for failure recovery,
		// after the switch nic failed, need to notify the dump first. After mindIO fails to return dump,
		// it goes through the failure recovery state machine again.
		signal.ChangeStrategy = constant.ProcessExitStrategyName
	}
	if ctl.jobInfo.PlatFormMode && signal.ChangeStrategy == constant.ProcessRecoverStrategyName {
		hwlog.RunLog.Infof("start wait plat rankTable ready, jobId=%s, pgName=%s",
			ctl.jobInfo.JobId, ctl.jobInfo.PgName)
		err := WaitRankTableReady(ctl.jobInfo.PgName, ctl.jobInfo.Namespace)
		hwlog.RunLog.Infof("finish wait plat rankTable ready, jobId=%s, pgName=%s, err=%v",
			ctl.jobInfo.JobId, ctl.jobInfo.PgName, err)
		if err != nil {
			ctl.platStrategy, err = platFormStrategy(ctl.jobInfo.PgName, ctl.jobInfo.Namespace, true)
			if err != nil {
				hwlog.RunLog.Errorf("flush platform strategy failed, err:%v", err)
				return "", common.ServerInnerError, err
			}
			signal.ChangeStrategy = ctl.chooseForRecoverFail()
		}
	}
	if signal.ChangeStrategy == constant.ProcessRetryStrategyName && ctl.shouldWaitHcclRoutingConvergence() {
		hwlog.RunLog.Infof("jobId=%s, should wait hccl routing convergence", ctl.jobInfo.JobId)
		if !ctl.waitHCCLRoutingConvergence() {
			return common.WaitHCCLRoutingConvergenceFail, common.HCCLRoutingConvergenceFail, nil
		}
	}
	if ctl.restartFaultProcess {
		return ctl.handleRestartFaultProcess(signal)
	}
	hwlog.RunLog.Infof("jobId=%s, choose strategy:%s", ctl.jobInfo.JobId, signal.ChangeStrategy)
	if signal.ChangeStrategy == constant.ScaleInStrategyName {
		signal.ExtraParams = `{"scale-in-strategy": "DP"}`
	} else if signal.ChangeStrategy == constant.ProcessRetryStrategyName {
		signal.FaultRanks = ctl.cacheRetryFault
	}
	var err error
	signal.NodeRankIds, err = common.GetNodeRankIdsByFaultRanks(ctl.jobInfo.JobId, signal.FaultRanks)
	if err != nil {
		hwlog.RunLog.Errorf("jobId=%s, GetNodeRankIdsByFaultRanks err:%v", ctl.jobInfo.JobId, err)
		return "", common.ServerInnerError, err
	}
	ctl.filterRecoverPodsNodeRankIds(signal)
	return ctl.signalEnqueue(signal)
}

func (ctl *EventController) filterRecoverPodsNodeRankIds(signal *pb.ProcessManageSignal) {
	if signal.ChangeStrategy != constant.ProcessRecoverStrategyName {
		return
	}
	ctl.filterPodsNodeRankIds(signal)
}

func (ctl *EventController) filterPodsNodeRankIds(signal *pb.ProcessManageSignal) {
	filteredNodeRankIds := make([]string, 0, len(signal.NodeRankIds))
	for _, nodeRankId := range signal.NodeRankIds {
		currentPod := pod.GetPodByRankIndex(ctl.jobInfo.JobId, nodeRankId)
		if currentPod.Name == "" {
			continue
		}
		hwlog.RunLog.Debugf("node rank is %s, currentPod uuid is %s, originPod uuid is %s",
			nodeRankId, currentPod.UID, ctl.originPod[nodeRankId])
		prePodUID, exists := ctl.originPod[nodeRankId]
		if exists && string(currentPod.UID) != prePodUID {
			hwlog.RunLog.Infof("jobId=%s, pod for rank %s has changed, original UID: %s, current UID: %s",
				ctl.jobInfo.JobId, nodeRankId, prePodUID, currentPod.UID)
			continue
		}
		filteredNodeRankIds = append(filteredNodeRankIds, nodeRankId)
	}
	hwlog.RunLog.Infof("jobId=%s, filteredNodeRankIds: %v", ctl.jobInfo.JobId, filteredNodeRankIds)
	signal.NodeRankIds = filteredNodeRankIds
}

func (ctl *EventController) handleRestartFaultProcess(signal *pb.ProcessManageSignal) (string,
	common.RespCode, error) {
	hwlog.RunLog.Infof("jobId:%s, enter handleRestartFaultProcess func, choose strategy:%s",
		ctl.jobInfo.JobId, signal.ChangeStrategy)
	restartProcessPods, restartPodPods := ctl.splitRecoverInPlacePodInfoToRestartProcessOrRestartPodList()
	if signal.ChangeStrategy == constant.ProcessRecoverInPlaceStrategyName {
		allFaults, allFaultRanks, err := ctl.updateCacheFaultAndPod()
		if err != nil {
			hwlog.RunLog.Errorf("update cache info fail, jobId=%s err=%v", ctl.jobInfo.JobId, err)
			return "", common.ServerInnerError, err
		}
		hwlog.RunLog.Infof("jobId=%s write reset json for restart process, faultRanks: %v",
			ctl.jobInfo.JobId, allFaultRanks)
		cm, err := common.WriteResetInfoToCM(ctl.jobInfo.JobName, ctl.jobInfo.Namespace,
			allFaultRanks, ctl.restartFaultProcess, constant.NotifyFaultListOperation)
		if err != nil {
			hwlog.RunLog.Errorf("notify agent faultList error, err=%v", err)
			return common.NotifyFailEvent, common.OperateConfigMapError, nil
		}
		hwlog.RunLog.Infof("write configmap faultList success, %s", cm.Data[constant.ResetInfoCMDataKey])
		signal.ExtraParams = constant.ProcessRecoverInPlaceStrategyName
		signal.ChangeStrategy = constant.ProcessRecoverStrategyName
		signal.Actions = append(signal.Actions, constant.FaultNodesRestartAction)
		signal.FaultRanks = allFaults
		signal.NodeRankIds = restartProcessPods
	} else if signal.ChangeStrategy != constant.ProcessRetryStrategyName {
		hwlog.RunLog.Warnf("choose strategy: %s, not restart fault process", signal.ChangeStrategy)
		restartPodPods = append(restartPodPods, restartProcessPods...)
		ctl.updateRecoverInPlaceInfoByRestartPods(restartPodPods)
		ctl.restartFaultProcess = false
		allFaults, _, err := ctl.updateCacheFaultAndPod()
		if err != nil {
			hwlog.RunLog.Errorf("update cache info fail, jobId=%s err=%v", ctl.jobInfo.JobId, err)
			return "", common.ServerInnerError, err
		}
		signal.FaultRanks = allFaults
		signal.Actions = append(signal.Actions, constant.FaultNodesExitAction)
		signal.NodeRankIds = restartPodPods
		ctl.filterPodsNodeRankIds(signal)
		faultmanager.CallbackForReportNoRetryInfo(ctl.jobInfo.JobId, restartPodPods, time.Now().UnixMilli())
	}
	return ctl.signalEnqueue(signal)
}

func (ctl *EventController) waitNormalFaultRecovery() []string {
	startTime := time.Now().Unix()
	subHealthStrategy := podgroup.GetSubHealthStrategyByJobKey(ctl.jobInfo.JobId)
	faultPodRanks, done := make([]string, 0), false
	for i := 0; i < constant.JobFaultDisappearRetryTimes; i++ {
		time.Sleep(constant.JobFaultCheckPeriod * time.Second)
		faultPodRanks, done = recoverinplace.RecoverInplaceProcessor.GetJobUnRecoverdPodRanks(ctl.jobInfo.JobId,
			subHealthStrategy)
		if done {
			hwlog.RunLog.Infof("fault disappear, jobId:%s", ctl.jobInfo.JobId)
			return faultPodRanks
		}
	}
	timeUse := time.Now().Unix() - startTime
	timeout := constant.JobFaultDisappearRetryTimes * constant.JobFaultCheckPeriod
	hwlog.RunLog.Warnf("wait fault disappear timeout, jobId:%s timeUse=%d > %d second",
		ctl.jobInfo.JobId, timeUse, timeout)
	return faultPodRanks
}

func (ctl *EventController) getStrategyResult() ([]string, []*pb.RecoverStatusRequest) {
	ctl.lock.RLock()
	defer ctl.lock.RUnlock()
	return ctl.latestStrategy, ctl.latestRecoverResult
}

func (ctl *EventController) extractRecoverResult() (common.RecoverResult, error) {
	latestStrategy, latestResult := ctl.getStrategyResult()
	n := len(latestStrategy)
	if n == 0 {
		hwlog.RunLog.Errorf("unexpected case, jobId=%s, have not decide strategy", ctl.jobInfo.JobId)
		return common.RecoverResult{
			Strategy:       "",
			Code:           common.ServerInnerError,
			RecoverSuccess: false,
		}, fmt.Errorf("unexpected case, jobId=%s, have not decide strategy", ctl.jobInfo.JobId)
	}
	if latestStrategy[n-1] == constant.ProcessExitStrategyName {
		return common.RecoverResult{
			Strategy:       constant.ProcessExitStrategyName,
			Code:           common.OK,
			RecoverSuccess: true,
		}, nil
	}

	lengthResult := len(latestResult)
	strategy := latestResult[lengthResult-1].Strategy
	code := latestResult[lengthResult-1].Status.Code
	recoverSuccess := latestResult[lengthResult-1].Status.Code == int32(common.OK)
	return common.RecoverResult{
		Strategy:       strategy,
		Code:           common.RespCode(code),
		RecoverSuccess: recoverSuccess,
		IsolateRankIds: latestResult[lengthResult-1].IsolateRankIds,
	}, nil
}

func (ctl *EventController) removeAgentStrategy(strategy string) {
	ctl.lock.Lock()
	defer ctl.lock.Unlock()
	var res []string
	for _, s := range ctl.agentReportStrategies {
		if s != strategy {
			res = append(res, s)
		}
	}
	ctl.agentReportStrategies = res
}

func (ctl *EventController) updateFixResult(strategy, value string) {
	newRecoverStatusAnnotation := map[string]interface{}{
		constant.ProcessRecoverStatusKey: value,
	}
	_, err := kube.RetryPatchPodGroupAnnotations(ctl.jobInfo.PgName, ctl.jobInfo.Namespace,
		retryTimes, newRecoverStatusAnnotation)
	if err != nil {
		hwlog.RunLog.Errorf("failed to patch pg when update fix result, err:%v, pgName=%s",
			err, ctl.jobInfo.PgName)
		return
	}
	hwlog.RunLog.Infof("fix result annatate success, %s=%s, pgName=%s",
		constant.ProcessRecoverStatusKey, value, ctl.jobInfo.PgName)
}

func (ctl *EventController) handleCheckRecoverResult() (string, common.RespCode, error) {
	result, err := ctl.extractRecoverResult()
	if err != nil {
		return "", result.Code, err
	}
	hwlog.RunLog.Infof("handleCheckRecoverResult result: %+v", result)
	switch result.Strategy {
	case constant.ProcessRetryStrategyName:
		if result.RecoverSuccess {
			ctl.updateFixResult(result.Strategy, constant.RetrySuccess)
			return common.RecoverSuccessEvent, common.OK, nil
		}
		ctl.updateFixResult(result.Strategy, constant.RetryFailed)
		if result.Code == common.RecoverableRetryError {
			return common.RecoverableRetryErrorEvent, common.RecoverableRetryError, nil
		}
		ctl.removeAgentStrategy(constant.ProcessRecoverStrategyName)
		if result.Code == common.ClientError {
			ctl.removeAgentStrategy(constant.ProcessDumpStrategyName)
		}
		return common.UnRecoverableRetryErrorEvent, common.UnRecoverableRetryError, nil
	case constant.ProcessRecoverStrategyName:
		if result.RecoverSuccess {
			go kube.RecoverFaultJobInfoCm(ctl.jobInfo.JobId, "")
			ctl.updateFixResult(result.Strategy, constant.RecoverSuccess)
			return common.RecoverSuccessEvent, common.OK, nil
		}
		ctl.updateFixResult(result.Strategy, constant.RecoverFailed)
		return common.RecoverFailEvent, common.ClientError, nil
	case constant.ScaleInStrategyName, constant.ScaleOutStrategyName:
		return ctl.handleCheckScaleStrategyRecoverResult(result)
	case constant.ProcessDumpStrategyName, constant.ProcessExitStrategyName:
		if result.Strategy == constant.ProcessExitStrategyName {
			ctl.updateFixResult(result.Strategy, constant.ExitCompleted)
			return common.CheckResultFinishEvent, common.OK, nil
		}
		if result.RecoverSuccess {
			ctl.updateFixResult(result.Strategy, constant.DumpSuccess)
		} else {
			ctl.updateFixResult(result.Strategy, constant.DumpFailed)
		}
		ctl.sendDumpExit()
		return common.CheckResultFinishEvent, common.OK, nil
	default:
		return "", common.ServerInnerError, fmt.Errorf("unexpected case, strategy=%s "+
			"not support, jobId=%s", result.Strategy, ctl.jobInfo.JobId)
	}
}

func (ctl *EventController) handleCheckScaleStrategyRecoverResult(result common.RecoverResult) (string, common.RespCode,
	error) {
	switch result.Strategy {
	case constant.ScaleInStrategyName:
		if !result.RecoverSuccess && result.Code == common.ExitIsolateRanksCode {
			hwlog.RunLog.Infof("job[%s] need notify isolate nodes to exit, ranks: %v",
				ctl.jobInfo.JobId, result.IsolateRankIds)
			signal := &pb.ProcessManageSignal{
				Uuid:       ctl.uuid,
				JobId:      ctl.jobInfo.JobId,
				SignalType: constant.FaultNodesExitSignalType,
				Actions:    []string{constant.FaultNodesExitAction},
			}
			nodeRankIds, err := common.GetNodeRankIdsByRankIds(ctl.jobInfo.JobId, result.IsolateRankIds)
			if err != nil {
				hwlog.RunLog.Errorf("get node rank ids failed, err:%v", err)
				return common.NotifyFailEvent, common.ClientError, nil
			}
			nodeRankIds = common.RemoveDuplicateNodeRanks(nodeRankIds, ctl.faultPod)
			for _, nodeRankId := range nodeRankIds {
				ctl.faultPod[nodeRankId] = ctl.prePodForScale[nodeRankId]
			}
			signal.NodeRankIds = nodeRankIds
			_, respCode, err := ctl.signalEnqueue(signal)
			if err != nil || respCode != common.OK {
				hwlog.RunLog.Errorf("failed to signalEnqueue, signal: %v, error:%v", signal, err)
				return common.NotifyFailEvent, common.ClientError, nil
			}
			return "", common.OK, nil
		}
		if result.RecoverSuccess {
			hwlog.RunLog.Infof("job[%s] scale-in successfully", ctl.jobInfo.JobId)
			return common.ScaleInSuccessEvent, common.OK, nil
		}
		hwlog.RunLog.Infof("job[%s] scale-in failed", ctl.jobInfo.JobId)
		return common.RecoverFailEvent, common.ClientError, nil
	case constant.ScaleOutStrategyName:
		if result.RecoverSuccess {
			hwlog.RunLog.Infof("job[%s] scale-out successfully", ctl.jobInfo.JobId)
			return common.ScaleOutSuccessEvent, common.OK, nil
		}
		hwlog.RunLog.Infof("job[%s] scale-out failed", ctl.jobInfo.JobId)
		return common.RecoverFailEvent, common.ClientError, nil
	default:
		return "", common.ServerInnerError, fmt.Errorf("unexpected case, strategy=%s "+
			"not support, jobId=%s", result.Strategy, ctl.jobInfo.JobId)
	}
}

func (ctl *EventController) handleKillPod() (string, common.RespCode, error) {
	if !job.GetJobIsExists(ctl.jobInfo.JobId) {
		return "", common.JobNotExist, fmt.Errorf("jobId=%s not exist", ctl.jobInfo.JobId)
	}
	ctl.takeRetryFault2NormalFault()
	_, allFaultRanks, err := ctl.updateCacheFaultAndPod()
	if err != nil {
		hwlog.RunLog.Errorf("update cache info fail, jobId=%s err=%v", ctl.jobInfo.JobId, err)
		return "", common.ServerInnerError, err
	}
	cm, err := common.RetryWriteResetCM(ctl.jobInfo.JobName, ctl.jobInfo.Namespace,
		allFaultRanks, false, constant.NotifyFaultListOperation)
	if err != nil {
		hwlog.RunLog.Errorf("notify kill pod fail, err=%v", err)
		return "", common.OperateConfigMapError, fmt.Errorf("jobId=%s write cm err:%v",
			ctl.jobInfo.JobId, err)
	}
	hwlog.RunLog.Infof("write configmap faultList success, %s", cm.Data[constant.ResetInfoCMDataKey])
	return common.FinishKillPodEvent, common.OK, nil
}

func (ctl *EventController) handleFaultRetry() (string, common.RespCode, error) {
	if _, err := common.ChangeProcessRecoverEnableMode(ctl.jobInfo, constant.ProcessRecoverPause); err != nil {
		hwlog.RunLog.Errorf("failed to change the process rescheduling label pause %s of pg %s, "+
			"prepare notify agent kill master through grpc channel",
			constant.ProcessRecoverPause, ctl.jobInfo.PgName)
		return common.ChangeProcessSchedulingModePauseErrorEvent, common.OperatePodGroupError, nil
	}
	hwlog.RunLog.Infof("change process rescheduling label %s success,"+
		" pgName=%s, uuid=%s", constant.ProcessRecoverPause, ctl.jobInfo.PgName, ctl.uuid)
	ctl.sendDumpExit()
	scheduleSuccess := false
	for i := 1; i <= constant.CheckPGRunningRetryTimes/2; i++ {
		time.Sleep(time.Second * constant.SleepSecondBeforeCheckPGRunning)
		if job.GetJobIsRunning(ctl.jobInfo.JobId) {
			scheduleSuccess = true
			break
		}
	}

	if !scheduleSuccess {
		hwlog.RunLog.Errorf("jobId=%s schedule timeout, "+
			"prepare notify agent kill master through grpc channel", ctl.jobInfo.JobId)
		return common.ScheduleTimeoutEvent, common.ScheduleTimeout, nil
	}

	if _, err := common.ChangeProcessRecoverEnableMode(ctl.jobInfo, constant.ProcessRecoverEnable); err != nil {
		hwlog.RunLog.Errorf("failed to change the process rescheduling label on %s of pg %s, "+
			"prepare notify agent kill master through grpc channel",
			constant.ProcessRecoverEnable, ctl.jobInfo.PgName)
		return common.ChangeProcessSchedulingModeEnableErrorEvent, common.OperatePodGroupError, nil
	}
	hwlog.RunLog.Infof("change process rescheduling label %s success,"+
		" jobId=%s, uuid=%s", constant.ProcessRecoverEnable, ctl.jobInfo.JobId, ctl.uuid)
	return common.FinishEvent, common.OK, nil
}

func (ctl *EventController) handleKillJob() (string, common.RespCode, error) {
	ctx, sendChan := ctl.getCtxAndSignalChan()
	if sendChan == nil {
		return "", common.ServerInnerError,
			fmt.Errorf("jobId=%s, sendChan is nil", ctl.jobInfo.JobId)
	}
	signal := &pb.ProcessManageSignal{
		Uuid:           ctl.uuid,
		JobId:          ctl.jobInfo.JobId,
		SignalType:     constant.KillMasterSignalType,
		Actions:        nil,
		FaultRanks:     nil,
		ChangeStrategy: "",
	}
	defer catchException()
	select {
	case sendChan <- signal:
		hwlog.RunLog.Infof("kill master signal enqueue, jobId=%s, uuid=%s", ctl.jobInfo.JobId, ctl.uuid)
		return "", common.OK, nil
	case <-ctx.Done():
		hwlog.RunLog.Warnf("controller context canceled, jobId=%s, uuid=%s", ctl.jobInfo.JobId, ctl.uuid)
		return "", common.ControllerEventCancel, nil
	case <-time.After(time.Minute):
		hwlog.RunLog.Warnf("add signal=%s time-out, program may running in chaos", ctl.jobInfo.JobId)
		return common.FinishEvent, common.OK, nil
	}
}

func (ctl *EventController) getCtxAndReportRecoverStrategyChan() (context.Context, chan *pb.RecoverStrategyRequest) {
	ctl.lock.RLock()
	defer ctl.lock.RUnlock()
	return ctl.controllerContext, ctl.reportRecoverStrategyChan
}

func (ctl *EventController) handleWaitReportRecoverStrategy() (string, common.RespCode, error) {
	ctx, strategyChan := ctl.getCtxAndReportRecoverStrategyChan()
	if strategyChan == nil {
		hwlog.RunLog.Errorf("jobId=%s, strategyChan is nil", ctl.jobInfo.JobId)
		return "", common.ServerInnerError, fmt.Errorf("jobId=%s, strategyChan is nil", ctl.jobInfo.JobId)
	}
	go func() {
		ctl.listenScheduleResult()
	}()
	select {
	case req, ok := <-strategyChan:
		if !ok {
			hwlog.RunLog.Warnf("strategyChan closed, jobId=%s, uuid=%s", ctl.jobInfo.JobId, ctl.uuid)
			return "", common.OK, nil
		}
		ctl.agentReportStrategies = req.Strategies
		return common.ReceiveReportEvent, common.OK, nil
	case <-ctx.Done():
		hwlog.RunLog.Warnf("controller context canceled, jobId=%s, uuid=%s", ctl.jobInfo.JobId, ctl.uuid)
		return "", common.ControllerEventCancel, nil
	case <-time.After(time.Duration(reportTimeoutMinutes) * time.Minute):
		hwlog.RunLog.Errorf("wait report recover strategy timeout, jobId=%s", ctl.jobInfo.JobId)
		return common.ReportTimeoutEvent, common.WaitReportTimeout, nil
	}
}

func (ctl *EventController) getCtxAndScheduleResultChan() (context.Context, chan bool) {
	ctl.lock.RLock()
	defer ctl.lock.RUnlock()
	return ctl.controllerContext, ctl.scheduleResultChan
}

func (ctl *EventController) pgStatusEnqueue(pgRunning bool) {
	ctx, ch := ctl.getCtxAndScheduleResultChan()
	if ch == nil {
		hwlog.RunLog.Warnf("resultCh is nil, jobId=%s, uuid=%s", ctl.jobInfo.JobId, ctl.uuid)
		return
	}
	defer catchException()
	select {
	case ch <- pgRunning:
		hwlog.RunLog.Infof("schedule result enqueue success, jobId=%s, value=%s",
			ctl.jobInfo.JobId, strconv.FormatBool(pgRunning))
	case <-ctx.Done():
		hwlog.RunLog.Warnf("controller context canceled, jobId=%s, uuid=%s", ctl.jobInfo.JobId, ctl.uuid)
	case <-time.After(time.Second):
		hwlog.RunLog.Errorf("schedule result enqueue timeout, jobId=%s, uuid=%s", ctl.jobInfo.JobId, ctl.uuid)
	}
}

func (ctl *EventController) listenScheduleResult() {
	preReSetTime := ctl.getCtlResetTime()
	if !(ctl.restartFaultProcess && ctl.configTargetStrategy(constant.ProcessRecoverInPlaceStrategyName)) &&
		!ctl.configTargetStrategy(constant.ProcessRecoverStrategyName) {
		hwlog.RunLog.Infof("job %s does not support recover or recover-in-place strategy, "+
			"not need to listen schedule result", ctl.jobInfo.JobId)
		return
	}
	if len(ctl.latestStrategy) > 0 && ctl.latestStrategy[len(ctl.latestStrategy)-1] == constant.
		ScaleInStrategyName {
		hwlog.RunLog.Infof("job %s has fault again when scale-in running, not need to listen schedule result",
			ctl.jobInfo.JobId)
		return
	}
	podReschedulingTimeout := constant.DefaultWaitRescheduleTimeout
	pgInfo := podgroup.GetPodGroup(ctl.jobInfo.JobId)
	if pgInfo.Annotations != nil && pgInfo.Annotations[constant.WaitRescheduleTimeoutKey] != "" {
		if timeout, err := strconv.Atoi(pgInfo.Annotations[constant.
			WaitRescheduleTimeoutKey]); err == nil && timeout <= constant.
			DefaultWaitRescheduleTimeout && timeout >= constant.MinWaitRescheduleTimeout {
			podReschedulingTimeout = timeout
		} else {
			hwlog.RunLog.Warnf("failed to convert wait_reschedule_timeout to int, value is %s",
				pgInfo.Annotations[constant.WaitRescheduleTimeoutKey])
		}
	}
	pgRunning := true
	start := time.Now().Unix()
	hwlog.RunLog.Infof("job %s begin listenScheduleResult, timeout: %v", ctl.jobInfo.JobId, podReschedulingTimeout)
	for !podgroup.JudgeIsRunningByJobKey(ctl.jobInfo.JobId) || !ctl.checkWhetherRestartPodPodsChanged() ||
		(!ctl.restartFaultProcess && !ctl.checkWhetherPodChanged()) {
		time.Sleep(time.Second * constant.SleepSecondBeforeCheckPGRunning)
		if ctl.getCtlResetTime() != preReSetTime {
			return
		}
		if len(ctl.latestStrategy) > 0 && ctl.latestStrategy[len(ctl.latestStrategy)-1] == constant.
			ScaleInStrategyName {
			hwlog.RunLog.Infof("job[%s] has choose elastic training, "+
				"not need to continue listen reschedule result", ctl.jobInfo.JobName)
			return
		}
		if time.Now().Unix()-start > int64(podReschedulingTimeout) {
			pgRunning = false
			break
		}
	}
	hwlog.RunLog.Infof("job %s ScheduleResult %v", ctl.jobInfo.JobId, pgRunning)
	ctl.pgStatusEnqueue(pgRunning)
}

func (ctl *EventController) appendStrategy(strategy string) {
	ctl.lock.Lock()
	defer ctl.lock.Unlock()
	ctl.latestStrategy = append(ctl.latestStrategy, strategy)
}

func (ctl *EventController) appendRecoverResult(req *pb.RecoverStatusRequest) {
	ctl.lock.Lock()
	defer ctl.lock.Unlock()
	ctl.latestRecoverResult = append(ctl.latestRecoverResult, req)
}

func (ctl *EventController) getCtxAndResultChan() (context.Context, chan *pb.RecoverStatusRequest) {
	ctl.lock.RLock()
	defer ctl.lock.RUnlock()
	return ctl.controllerContext, ctl.reportStatusChan
}

func (ctl *EventController) getCtxAndNewStatusMonitorChan() (context.Context, chan corev1.PodPhase) {
	ctl.lock.RLock()
	defer ctl.lock.RUnlock()
	return ctl.controllerContext, ctl.newPodStatusMonitorChan
}

func (ctl *EventController) handleDecideRetryStrategy() (string, common.RespCode, error) {
	ctl.appendStrategy(constant.ProcessRetryStrategyName)
	ctx, ch := ctl.getCtxAndResultChan()
	if ch == nil {
		return "", common.OK, fmt.Errorf("jobId=%s, result chan is nil", ctl.jobInfo.JobId)
	}
	select {
	case req, ok := <-ch:
		if !ok {
			hwlog.RunLog.Warnf("resultCh closed, jobId=%s, uuid=%s", ctl.jobInfo.JobId, ctl.uuid)
			return "", common.OK, nil
		}
		hwlog.RunLog.Infof("cur state is wait retry strategy status, report strategy=%s result", req.Strategy)
		ctl.appendRecoverResult(req)
		return common.ReceiveReportEvent, common.OK, nil
	case <-ctx.Done():
		hwlog.RunLog.Warnf("controller context canceled, jobId=%s, uuid=%s", ctl.jobInfo.JobId, ctl.uuid)
		return "", common.ControllerEventCancel, nil
	case <-time.After(time.Duration(reportTimeoutMinutes) * time.Minute):
		hwlog.RunLog.Errorf("wait report recover status timeout, jobId=%s", ctl.jobInfo.JobId)
		return common.ReportTimeoutEvent, common.WaitReportTimeout, nil
	}
}

func (ctl *EventController) handleDecideRecoverStrategy() (string, common.RespCode, error) {
	ctl.appendStrategy(constant.ProcessRecoverStrategyName)
	ctx, resultCh := ctl.getCtxAndResultChan()
	_, scheduleCh := ctl.getCtxAndScheduleResultChan()
	if resultCh == nil || scheduleCh == nil {
		hwlog.RunLog.Errorf("jobId=%s, resultCh or scheduleCh is nil", ctl.jobInfo.JobId)
		return "", common.OK, fmt.Errorf("jobId=%s, resultCh or scheduleCh is nil", ctl.jobInfo.JobId)
	}
	timer := time.NewTimer(time.Duration(reportTimeoutMinutes) * time.Minute)
	defer timer.Stop()
	for {
		select {
		case scheduleSuccess, ok := <-scheduleCh:
			if !ok {
				hwlog.RunLog.Warnf("scheduleCh closed, jobId=%s, uuid=%s", ctl.jobInfo.JobId, ctl.uuid)
				return "", common.OK, nil
			}
			if !scheduleSuccess && !ctl.supportTargetStrategy(constant.ElasticTrainingStrategyName) {
				return common.ScheduleTimeoutEvent, common.ScheduleTimeout, nil
			} else if !scheduleSuccess && ctl.supportTargetStrategy(constant.ElasticTrainingStrategyName) {
				return common.NeedTryScaleInStrategyEvent, common.ScheduleTimeout, nil
			}
			if ctl.restartFaultProcess {
				time.Sleep(time.Second * constant.WaitAgentGetFaultRank)
			}
			_, err := common.RetryWriteResetCM(ctl.jobInfo.JobName, ctl.jobInfo.Namespace, nil, false,
				constant.ClearOperation)
			if err != nil {
				hwlog.RunLog.Errorf("clear reset configMap error, err=%v, jobId=%s, uuid=%s", err, ctl.jobInfo.JobId,
					ctl.uuid)
				return common.ClearConfigMapFaultFailEvent, common.OperateConfigMapError, nil
			}
		case req := <-resultCh:
			hwlog.RunLog.Infof("cur state is %s, strategy=%s, code=%d", ctl.state.GetState(), req.Strategy,
				req.Status.Code)
			ctl.appendRecoverResult(req)
			return common.ReceiveReportEvent, common.OK, nil
		case <-ctx.Done():
			hwlog.RunLog.Warnf("controller context canceled, jobId=%s, uuid=%s", ctl.jobInfo.JobId, ctl.uuid)
			return "", common.ControllerEventCancel, nil
		case <-timer.C:
			hwlog.RunLog.Errorf("wait report recover status timeout, jobId=%s", ctl.jobInfo.JobId)
			return common.ReportTimeoutEvent, common.WaitReportTimeout, nil
		}
	}
}

func (ctl *EventController) handleDecideDumpStrategy() (string, common.RespCode, error) {
	ctl.appendStrategy(constant.ProcessDumpStrategyName)
	return ctl.waitReportStatus()
}

func (ctl *EventController) handleDecideExitStrategy() (string, common.RespCode, error) {
	ctl.appendStrategy(constant.ProcessExitStrategyName)
	return common.CheckResultFinishEvent, common.OK, nil
}

func (ctl *EventController) handleListenScheduleResult() (string, common.RespCode, error) {
	if ctl.jobInfo.PlatFormMode {
		// for plat not support pod rescheduling
		return common.ScheduleSuccessEvent, common.OK, nil
	}
	scheduleSuccess := false
	for i := 1; i <= constant.CheckPodReschedulingTimes; i++ {
		time.Sleep(time.Second * constant.SleepSecondBeforeCheckPGRunning)
		if job.GetJobIsRunning(ctl.jobInfo.JobId) &&
			common.FaultPodAllRescheduled(ctl.jobInfo.JobId, ctl.faultPod) {
			scheduleSuccess = true
			break
		}
	}
	if scheduleSuccess || job.GetJobIsRunning(ctl.jobInfo.JobId) {
		return common.ScheduleSuccessEvent, common.OK, nil
	}
	return common.ScheduleTimeoutEvent, common.ScheduleTimeout, nil
}

func (ctl *EventController) handleRestartAllProcess() (string, common.RespCode, error) {
	_, err := common.RetryWriteResetCM(ctl.jobInfo.JobName, ctl.jobInfo.Namespace, nil, false,
		constant.RestartAllProcessOperation)
	if err != nil {
		hwlog.RunLog.Errorf("clear reset configMap error, err=%v, jobId=%s, uuid=%s",
			err, ctl.jobInfo.JobId, ctl.uuid)
		return common.NotifyFailEvent, common.OperateConfigMapError, nil
	}
	return common.NotifySuccessEvent, common.OK, nil
}

func (ctl *EventController) sendDumpExit() {
	signal := &pb.ProcessManageSignal{
		Uuid:           ctl.uuid,
		JobId:          ctl.jobInfo.JobId,
		SignalType:     constant.ChangeStrategySignalType,
		Actions:        changeStrategyActions,
		ChangeStrategy: constant.ProcessExitStrategyName,
		ExtraParams:    constant.DumpExit,
	}
	_, allFaultRanks, updateErr := ctl.updateCacheFaultAndPod()
	if updateErr != nil {
		hwlog.RunLog.Errorf("update cache info fail, jobId=%s err=%v", ctl.jobInfo.JobId, updateErr)
		return
	}
	signal.NodeRankIds, updateErr = common.GetNodeRankIdsByRankIds(ctl.jobInfo.JobId, allFaultRanks)
	if updateErr != nil {
		hwlog.RunLog.Errorf("jobId=%s, GetNodeRankIdsByRankIds err:%v", ctl.jobInfo.JobId, updateErr)
		return
	}
	_, respCode, enqueueErr := ctl.signalEnqueue(signal)
	if enqueueErr != nil || respCode != common.OK {
		hwlog.RunLog.Errorf("failed to signalEnqueue, signal: %v, error:%v", signal, enqueueErr)
		return
	}
}

func (ctl *EventController) handleWaitRestartAllProcess() (string, common.RespCode, error) {
	if ctl.jobInfo.PlatFormMode {
		// for plat not support pod rescheduling
		return common.RestartProcessFinishEvent, common.OK, nil
	}
	ctx, _ := ctl.getCtxAndScheduleResultChan()
	select {
	case <-time.After(time.Minute):
		return common.RestartProcessFinishEvent, common.OK, nil
	case <-ctx.Done():
		time.Sleep(time.Minute)
		hwlog.RunLog.Warnf("controller context canceled, jobId=%s, uuid=%s", ctl.jobInfo.JobId, ctl.uuid)
		return "", common.ControllerEventCancel, nil
	}
}

func (ctl *EventController) waitHCCLRoutingConvergence() bool {
	time.Sleep(constant.HCCLRoutingConvergenceTimeout * time.Second)
	hwlog.RunLog.Info("hccl routing convergence finished")
	return true
}

func (ctl *EventController) hasSameRetryFault() bool {
	var faultType string
	for i, fault := range ctl.cacheRetryFault {
		if i == 0 {
			faultType = fault.FaultType
			continue
		}
		if faultType != fault.FaultType {
			return false
		}
	}
	return true
}

func (ctl *EventController) notifyHCCLRoutingTimeout(signal *pb.ProcessManageSignal) *pb.ProcessManageSignal {
	for _, fault := range signal.FaultRanks {
		if fault.FaultType == constant.HcclFaultType {
			signal.Timeout = constant.HCCLRoutingConvergenceTimeout + constant.StepRetryTimeout
			hwlog.RunLog.Infof("notify hccl routing timeout: %d, jobId=%s", signal.Timeout, ctl.jobInfo.JobId)
			return signal
		}
	}
	return signal
}

func (ctl *EventController) shouldWaitHcclRoutingConvergence() bool {
	for _, fault := range ctl.cacheRetryFault {
		if fault.FaultType == constant.HcclFaultType {
			return true
		}
	}
	return false
}

func (ctl *EventController) handleNotifyScaleInStrategy() (string, common.RespCode, error) {
	signal := &pb.ProcessManageSignal{
		Uuid:           ctl.uuid,
		JobId:          ctl.jobInfo.JobId,
		SignalType:     constant.ChangeStrategySignalType,
		Actions:        changeStrategyActions,
		ChangeStrategy: constant.ScaleInStrategyName,
		FaultRanks:     ctl.cacheNormalFault,
		ExtraParams:    `{"scale-in-strategy": "DP"}`,
	}
	var err error
	signal.NodeRankIds, err = common.GetNodeRankIdsByFaultRanks(ctl.jobInfo.JobId, signal.FaultRanks)
	if err != nil {
		hwlog.RunLog.Warnf("jobId=%s, GetNodeRankIdsByFaultRanks err:%v", ctl.jobInfo.JobId, err)
	}
	return ctl.signalEnqueue(signal)
}

func (ctl *EventController) handleWaitReportScaleInIsolateRanksStatus() (string, common.RespCode, error) {
	ctl.appendStrategy(constant.ScaleInStrategyName)
	return ctl.waitReportStatus()
}

func (ctl *EventController) waitReportStatus() (string, common.RespCode, error) {
	ctx, resultCh := ctl.getCtxAndResultChan()
	if resultCh == nil {
		hwlog.RunLog.Errorf("jobId=%s, resultCh is nil", ctl.jobInfo.JobId)
		return "", common.ServerInnerError, fmt.Errorf("jobId=%s, resultCh is nil", ctl.jobInfo.JobId)
	}
	select {
	case req, ok := <-resultCh:
		if !ok {
			hwlog.RunLog.Warnf("resultCh closed, jobId=%s, uuid=%s", ctl.jobInfo.JobId, ctl.uuid)
			return "", common.OK, nil
		}
		hwlog.RunLog.Infof("cur state is %s, strategy=%s, code=%d", ctl.state.GetState(), req.Strategy, req.Status.Code)
		ctl.appendRecoverResult(req)
		return common.ReceiveReportEvent, common.OK, nil
	case <-ctx.Done():
		hwlog.RunLog.Warnf("controller context canceled, jobId=%s, uuid=%s", ctl.jobInfo.JobId, ctl.uuid)
		return "", common.ControllerEventCancel, nil
	case <-time.After(time.Duration(reportTimeoutMinutes) * time.Minute):
		hwlog.RunLog.Errorf("%s timeout, jobId=%s", ctl.state.GetState(), ctl.jobInfo.JobId)
		return common.ReportTimeoutEvent, common.WaitReportTimeout, nil
	}
}

func (ctl *EventController) handleWaitReportScaleInStatus() (string, common.RespCode, error) {
	return ctl.waitReportStatus()
}

func (ctl *EventController) handleScaleInRunningState() (string, common.RespCode, error) {
	go ctl.waitScaleOut()
	return "", common.OK, nil
}

func (ctl *EventController) waitScaleOut() {
	if ctl == nil || ctl.state == nil {
		return
	}
	for {
		time.Sleep(time.Second * constant.SleepSecondBeforeCheckPGRunning)
		hwlog.RunLog.Debugf("check job[%s] pg running and job completed", ctl.jobInfo.JobId)
		if ctl.state.GetState() != common.ScaleInRunningState {
			hwlog.RunLog.Infof("job[%s] state is not %v now, not need to check pg running status",
				ctl.jobInfo.JobId, common.ScaleInRunningState)
			return
		}
		jobObject := statistics.GetJob(ctl.jobInfo.JobId)
		if jobObject == nil {
			ctl.addEvent(common.FinishEvent)
			return
		}
		if acJobInfo, ok := jobObject.(*v1.AscendJob); ok && (kubeflowutil.IsSucceeded(acJobInfo.
			Status) || kubeflowutil.IsFailed(acJobInfo.Status)) {
			hwlog.RunLog.Infof("job[%s] is succeeded or failed, IsSucceed: %v", ctl.jobInfo.JobId,
				kubeflowutil.IsSucceeded(acJobInfo.
					Status))
			ctl.addEvent(common.FinishEvent)
			return
		}
		if podgroup.JudgeIsRunningByJobKey(ctl.jobInfo.JobId) && ctl.checkWhetherPodChanged() {
			break
		}
	}
	signal := &pb.ProcessManageSignal{
		Uuid:           ctl.uuid,
		JobId:          ctl.jobInfo.JobId,
		SignalType:     constant.ChangeStrategySignalType,
		Actions:        changeStrategyActions,
		ChangeStrategy: constant.ScaleOutStrategyName,
		FaultRanks:     ctl.cacheNormalFault,
		ExtraParams:    `{"scale-out-strategy": "DP"}`,
	}
	var err error
	signal.NodeRankIds, err = common.GetNodeRankIdsByFaultRanks(ctl.jobInfo.JobId, signal.FaultRanks)
	if err != nil {
		hwlog.RunLog.Warnf("jobId=%s, GetNodeRankIdsByFaultRanks err:%v", ctl.jobInfo.JobId, err)
	}
	_, respCode, err := ctl.signalEnqueue(signal)
	if err != nil || respCode != common.OK {
		hwlog.RunLog.Warnf("failed to send scale-out signal, error: %v", err)
		ctl.addEvent(common.NotifyFailEvent)
	}
}

func (ctl *EventController) handleWaitReportScaleOutStatusState() (string, common.RespCode, error) {
	ctl.appendStrategy(constant.ScaleOutStrategyName)
	return ctl.waitReportStatus()
}

func (ctl *EventController) whetherHasEnoughResource() bool {
	podReschedulingTimeout := constant.DefaultWaitRescheduleTimeoutBeforeDeployStrategy
	podInfo := pod.GetPodByRankIndex(ctl.jobInfo.JobId, constant.RankZeroNodeId)
	envValue := pod.GetEnvByPod(map[string]corev1.Pod{podInfo.Name: podInfo}, constant.MindIOWaitTimeKey)
	waitTimeInt, err := strconv.Atoi(envValue)
	if err != nil {
		hwlog.RunLog.Warnf("failed  to convert str to int, error: %v", err)
	} else if err == nil && waitTimeInt >= constant.DifferenceTime && waitTimeInt <= constant.MindIOWaitTimeMax {
		podReschedulingTimeout = waitTimeInt - constant.DifferenceTime
	}
	hwlog.RunLog.Infof("job %s will wait %v seconds to check the fault pod reschedule result", ctl.jobInfo.JobId,
		podReschedulingTimeout)
	start := time.Now().Unix()
	for !podgroup.JudgeIsRunningByJobKey(ctl.jobInfo.JobId) || !ctl.checkWhetherPodChanged() {
		time.Sleep(time.Second * constant.SleepSecondBeforeCheckPGRunning)
		if time.Now().Unix()-start > int64(podReschedulingTimeout) {
			hwlog.RunLog.Infof("job %s has not enough resource to reschedule", ctl.jobInfo.JobId)
			return false
		}
	}
	hwlog.RunLog.Infof("job %s has enough resource to reschedule", ctl.jobInfo.JobId)
	return true
}

func (ctl *EventController) checkWhetherPodChanged() bool {
	hasChanged := false
	for podRank, _ := range ctl.faultPod {
		pod := pod.GetPodByRankIndex(ctl.jobInfo.JobId, podRank)
		if pod.Name == "" || pod.Annotations == nil {
			continue
		}
		if realCard, ok := pod.Annotations[api.PodAnnotationAscendReal]; !ok || realCard == "" {
			continue
		}
		hwlog.RunLog.Debugf("node rank %s prePod id %s now pod id %s", podRank, ctl.prePodForScale[podRank],
			string(pod.UID))
		if ctl.prePodForScale[podRank] != string(pod.UID) {
			hasChanged = true
			ctl.prePodForScale[podRank] = string(pod.UID)
		}
	}
	return hasChanged
}

func (ctl *EventController) ChangePodStatus(status corev1.PodPhase) {
	_, newPodMonitorChan := ctl.getCtxAndNewStatusMonitorChan()
	if newPodMonitorChan == nil {
		hwlog.RunLog.Warnf("newPodMonitorChan is nil")
		return
	}
	newPodMonitorChan <- status
}

func (ctl *EventController) checkWhetherRestartPodPodsChanged() bool {
	if !ctl.restartFaultProcess {
		return true
	}
	_, restartPodPods := ctl.splitRecoverInPlacePodInfoToRestartProcessOrRestartPodList()
	for _, podRank := range restartPodPods {
		currentPod := pod.GetPodByRankIndex(ctl.jobInfo.JobId, podRank)
		if currentPod.Name == "" {
			return false
		}
		if realCard, ok := currentPod.Annotations[api.PodAnnotationAscendReal]; !ok || realCard == "" {
			return false
		}
		prePodUID, exists := ctl.originPod[podRank]
		if exists && string(currentPod.UID) != prePodUID {
			hwlog.RunLog.Infof("jobId=%s, pod for rank %s has changed, original UID: %s, current UID: %s",
				ctl.jobInfo.JobId, podRank, prePodUID, currentPod.UID)
			continue
		}
	}
	return true
}

func (ctl *EventController) updateRestartProcessOrPodInfoBySoftFault(faultRanks []*pb.FaultRank) {
	if !podgroup.JudgeRestartProcessByJobKey(ctl.jobInfo.JobId) {
		return
	}
	deviceNumPerPod := pod.GetPodDeviceNumByJobId(ctl.jobInfo.JobId)
	podRankFaultList, err := common.GetPodRankFaultBySoftFaultRank(faultRanks, deviceNumPerPod)
	if err != nil {
		hwlog.RunLog.Errorf("get pod rank fault failed, err: %v", err)
		ctl.restartFaultProcess = false
		return
	}
	ctl.updateRestartProcessOrPodInfo(podRankFaultList)
}

func (ctl *EventController) updateRestartProcessOrPodInfoByHardwareFault(faultRanks []constant.FaultRank) {
	if !podgroup.JudgeRestartProcessByJobKey(ctl.jobInfo.JobId) {
		return
	}
	podRankFaultList := common.GetPodRankFaultByFaultRank(faultRanks)
	ctl.updateRestartProcessOrPodInfo(podRankFaultList)
}

func (ctl *EventController) updateRestartProcessOrPodInfo(podRankFaultList []*constant.PodFaultInfo) {
	ctl.lock.Lock()
	defer ctl.lock.Unlock()
	for _, fault := range podRankFaultList {
		podRankFault, ok := ctl.recoverInPlacePodFaults[fault.PodRank]
		if !ok {
			ctl.recoverInPlacePodFaults[fault.PodRank] = fault
			continue
		}
		podRankFault.DoRestartInPlace = podRankFault.DoRestartInPlace && fault.DoRestartInPlace
		podRankFault.FaultType = podRankFault.FaultType | fault.FaultType
	}
}

func (ctl *EventController) updateRestartFaultProcess() {
	ctl.lock.Lock()
	defer ctl.lock.Unlock()
	jobFaultType := constant.StatusNone
	for _, podRankFault := range ctl.recoverInPlacePodFaults {
		jobFaultType = jobFaultType | podRankFault.FaultType
		ctl.restartFaultProcess = ctl.restartFaultProcess || podRankFault.DoRestartInPlace
	}
	// strategy does not support recover-inplace or multi pods have software fault, can not do recover-inplace
	if jobFaultType == constant.StatusHasSoftFault && len(ctl.recoverInPlacePodFaults) > 1 {
		hwlog.RunLog.Infof("job %v all fault is software fault", ctl.jobInfo.JobId)
		ctl.restartFaultProcess = false
	}
	hwlog.RunLog.Infof("job %v update restartFaultProcess %v", ctl.jobInfo.JobId, ctl.restartFaultProcess)
}

func (ctl *EventController) updateRecoverInPlaceInfoByRestartPods(restartPods []string) {
	ctl.lock.Lock()
	for _, podRank := range restartPods {
		if podRankFault, ok := ctl.recoverInPlacePodFaults[podRank]; ok {
			podRankFault.DoRestartInPlace = false
			podRankFault.ShouldExitPod = true
		}
	}
	ctl.lock.Unlock()
	ctl.updateRestartFaultProcess()
}

func (ctl *EventController) splitRecoverInPlacePodInfoToRestartProcessOrRestartPodList() ([]string, []string) {
	ctl.lock.RLock()
	defer ctl.lock.RUnlock()
	restartProcessPods, restartPodPods := make([]string, 0), make([]string, 0)
	for podRank, podRankInfo := range ctl.recoverInPlacePodFaults {
		if podRankInfo.ShouldExitPod {
			restartPodPods = append(restartPodPods, podRank)
			continue
		}
		restartProcessPods = append(restartProcessPods, podRank)
	}
	return restartProcessPods, restartPodPods
}

func (ctl *EventController) recoverInPlaceFaultAssociateSameNodeRank(faultRankIds []string) ([]string, []string) {
	deviceNumPerNode := pod.GetPodDeviceNumByJobId(ctl.jobInfo.JobId)
	if deviceNumPerNode == 0 || len(faultRankIds) == 0 {
		return faultRankIds, nil
	}
	_, restartPodPods := ctl.splitRecoverInPlacePodInfoToRestartProcessOrRestartPodList()
	restartPodPodSet := sets.NewString(restartPodPods...)
	faultRanks := util.StringSliceToIntSlice(faultRankIds)
	exitRanks, restartRanks := make([]string, 0), make([]string, 0)
	for i, rankId := range faultRanks {
		if restartPodPodSet.Has(strconv.Itoa(rankId / deviceNumPerNode)) {
			exitRanks = append(exitRanks, faultRankIds[i])
			continue
		}
		restartRanks = append(restartRanks, faultRankIds[i])
	}
	return exitRanks, restartRanks
}

func (ctl *EventController) getAllFaultRanks() ([]*pb.FaultRank, []string, error) {
	var allFaults []*pb.FaultRank
	var allFaultRanks []string
	if ctl.tpBlockRescheduleScene() {
		allFaults, allFaultRanks = ctl.normalFaultAssociateRescheduledPods()
	} else {
		allFaults, allFaultRanks = ctl.normalFaultAssociateSameNodeRank()
	}
	return allFaults, allFaultRanks, nil
}

func (ctl *EventController) tpBlockRescheduleScene() bool {
	pg, err := kube.RetryGetPodGroup(ctl.jobInfo.PgName, ctl.jobInfo.Namespace, constant.GetPodGroupTimes)
	if err != nil {
		hwlog.RunLog.Errorf("kube get pod group err, err=%v", err)
		return false
	}
	tpBlock, ok := pg.Annotations[constant.TpBlock]
	return ok && slices.Contains(validTpBlock, tpBlock) && !ctl.restartFaultProcess
}

// normalFaultAssociateRescheduledPods calculate fault ranks in rescheduled pods
func (ctl *EventController) normalFaultAssociateRescheduledPods() ([]*pb.FaultRank, []string) {
	var rescheduledPods []string
	currentPods := pod.GetPodByJobId(ctl.jobInfo.JobId)
	for uid, podInfo := range currentPods {
		if ctl.originPod[podInfo.Annotations[constant.PodRankIndexAnno]] != uid {
			rescheduledPods = append(rescheduledPods, podInfo.Annotations[constant.PodRankIndexAnno])
		}
	}
	devicePerNode := pod.GetPodDeviceNumByJobId(ctl.jobInfo.JobId)
	allFaultRankIds := common.GetFaultRankIdsByPodRank(rescheduledPods, devicePerNode)
	removeSameRankIds := util.RemoveSliceDuplicateElement(allFaultRankIds)
	var res []*pb.FaultRank
	for _, rank := range removeSameRankIds {
		res = append(res, &pb.FaultRank{
			RankId:    rank,
			FaultType: constant.NormalFaultType,
		})
	}
	return res, removeSameRankIds
}
