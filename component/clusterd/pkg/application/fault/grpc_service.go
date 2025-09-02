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

// Package fault service for grpc client
package fault

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/util/uuid"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/application/config"
	"clusterd/pkg/application/faultmanager"
	"clusterd/pkg/application/faultmanager/faultclusterprocess"
	"clusterd/pkg/application/faultmanager/jobprocess/faultrank"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/logs"
	"clusterd/pkg/common/util"
	"clusterd/pkg/domain/common"
	"clusterd/pkg/domain/job"
	common2 "clusterd/pkg/interface/common"
	"clusterd/pkg/interface/grpc/fault"
)

const (
	jobFaultInfoChanCache = 5
	defaultTokenRate      = 10
	defaultBurst          = 10
	defaultMaxQueueLen    = 50
	minJobIdLen           = 8
	maxJobIdLen           = 128
)

var chinesePattern = regexp.MustCompile(`[\x{4e00}-\x{9fa5}]`)

// FaultServer fault server
type FaultServer struct {
	serviceCtx context.Context
	// jobId -> role -> publisher
	faultPublisher map[string]map[string]*config.ConfigPublisher[*fault.FaultMsgSignal]
	lock           sync.RWMutex
	fault.UnimplementedFaultServer
	faultCh chan map[string]constant.JobFaultInfo
	limiter *util.AdvancedRateLimiter
}

// NewFaultServer create a fault server
func NewFaultServer(ctx context.Context) *FaultServer {
	server := &FaultServer{
		serviceCtx:     ctx,
		faultPublisher: make(map[string]map[string]*config.ConfigPublisher[*fault.FaultMsgSignal]),
		lock:           sync.RWMutex{},
		faultCh:        make(chan map[string]constant.JobFaultInfo, jobFaultInfoChanCache),
		limiter:        util.NewAdvancedRateLimiter(defaultTokenRate, defaultBurst, defaultMaxQueueLen),
	}
	if err := faultmanager.RegisterForJobFaultRank(server.faultCh, reflect.TypeOf(server).Name()); err != nil {
		hwlog.RunLog.Error("RegisterForJobFaultRank fail")
	}
	common2.SetPublisher(server)
	go server.checkFaultFromFaultCenter()
	return server
}

// IsSubscribed return job whether subscribe fault message signal from ClusterD
func (s *FaultServer) IsSubscribed(jobId, role string) bool {
	publisher, ok := s.getPublisher(jobId, role)
	return ok && publisher != nil && publisher.IsSubscribed()
}

// Register is task register service
func (s *FaultServer) Register(ctx context.Context, req *fault.ClientInfo) (*fault.Status, error) {
	if req == nil || req.Role == "" {
		hwlog.RunLog.Errorf("Register failed, request: %v", req)
		return nil, errors.New("request is nil or role is empty")
	}
	if req.JobId == "" {
		req.JobId = constant.DefaultJobId
	}
	hwlog.RunLog.Infof("fault service receive Register request, jobId=%s, role=%s",
		req.JobId, req.Role)
	publisher, ok := s.getPublisher(req.JobId, req.Role)
	if !ok || publisher == nil {
		code, err := s.preRegistry(req)
		if err != nil {
			hwlog.RunLog.Errorf("jobId=%s, preCheck err:%v", req.JobId, err)
			return &fault.Status{Code: int32(code), Info: err.Error()}, err
		}
	}
	s.preemptPublisher(req.JobId, req.Role)
	return &fault.Status{Code: int32(common.OK), Info: "register success"}, nil
}

func (s *FaultServer) preRegistry(req *fault.ClientInfo) (common.RespCode, error) {
	if s.serveJobNum() >= constant.MaxServeJobs {
		return common.OutOfMaxServeJobs,
			fmt.Errorf("jobId=%s out of max serve jobs", req.JobId)
	}
	if req.JobId != constant.DefaultJobId {
		_, ok := job.GetJobCache(req.JobId)
		_, err := job.GetNamespaceByJobIdAndAppType(req.JobId, req.Role)
		if !ok && err != nil {
			hwlog.RunLog.Errorf("jobId=%s not exist and is not multi-instance job", req.JobId)
			return common.JobNotExist, fmt.Errorf("jobId=%s not exist and is not multi-instance", req.JobId)
		}
	}
	return common.OK, nil
}

func (s *FaultServer) serveJobNum() int {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return len(s.faultPublisher)
}

// SubscribeFaultMsgSignal subscribe fault message signal from ClusterD
func (s *FaultServer) SubscribeFaultMsgSignal(request *fault.ClientInfo,
	stream fault.Fault_SubscribeFaultMsgSignalServer) error {
	if request == nil || request.Role == "" {
		hwlog.RunLog.Errorf("Register failed, request: %v", request)
		return errors.New("request is nil or role is empty")
	}
	event := "subscribe fault msg signal"
	logs.RecordLog(request.Role, event, constant.Start)
	res := constant.Failed
	defer func() {
		logs.RecordLog(request.Role, event, res)
	}()

	if request.JobId == "" {
		request.JobId = constant.DefaultJobId
	}
	requestInfo := fmt.Sprintf("jobId=%s, role=%s", request.JobId, request.Role)
	hwlog.RunLog.Infof("receive Subscribe fault message signal request, %s", requestInfo)
	faultPublisher, exist := s.getPublisher(request.JobId, request.Role)
	if !exist || faultPublisher == nil {
		hwlog.RunLog.Warnf("jobId=%s not registered, role=%s", request.JobId, request.Role)
		return fmt.Errorf("jobId=%s not registered, role=%s", request.JobId, request.Role)
	}
	faultPublisher.ListenDataChange(stream)
	s.deletePublisher(request.JobId, request.Role, faultPublisher.GetCreateTime())
	hwlog.RunLog.Infof("jobId=%s stop subscribe fault message signal, createTime=%v",
		request.JobId, faultPublisher.GetCreateTime().UnixNano())
	res = constant.Success
	return nil
}

// isValidJobId check jobId is valid
func isValidJobId(jobId string) bool {
	if len(jobId) < minJobIdLen || len(jobId) > maxJobIdLen {
		return false
	}
	if chinesePattern.MatchString(jobId) {
		return false
	}
	return true
}

// GetFaultMsgSignal return cluster fault
func (s *FaultServer) GetFaultMsgSignal(ctx context.Context, request *fault.ClientInfo) (*fault.FaultQueryResult, error) {
	event := "get fault info"
	logs.RecordLog(request.Role, event, constant.Start)
	res := constant.Failed
	defer func() {
		logs.RecordLog(request.Role, event, res)
	}()

	hwlog.RunLog.Infof("job %s role %s call get faults", request.JobId, request.Role)
	if !s.limiter.Allow(ctx) {
		return &fault.FaultQueryResult{
			Code:        common.RateLimitedCode,
			Info:        "rate limited, there is too many requests, please retry later",
			FaultSignal: nil,
		}, errors.New("rate limited, there is too many requests, please retry later")
	}
	jobId := request.GetJobId()
	if jobId == "" {
		return s.getClusterFaultInfo(), nil
	}
	if !isValidJobId(jobId) {
		errMsg := fmt.Sprintf("job with jobId: %v is invalid", jobId)
		return &fault.FaultQueryResult{
			Code:        common.InvalidReqParam,
			Info:        errMsg,
			FaultSignal: nil,
		}, errors.New(errMsg)
	}
	jobFaultInfo := faultrank.JobFaultRankProcessor.GetJobFaultRankInfos()
	faultInfo, ok := jobFaultInfo[jobId]
	if !ok {
		return &fault.FaultQueryResult{
			Code:        int32(common.SuccessCode),
			Info:        fmt.Sprintf("job with jobId: %v not found fault info", jobId),
			FaultSignal: nil,
		}, nil
	}
	faultMsg := faultDeviceToSortedFaultMsgSignal(jobId, faultInfo.FaultDevice)
	res = constant.Success
	return &fault.FaultQueryResult{
		Code:        int32(common.SuccessCode),
		Info:        "all info returned",
		FaultSignal: faultMsg,
	}, nil
}

func (s *FaultServer) getClusterFaultInfo() *fault.FaultQueryResult {
	faultMsg := faultclusterprocess.ClusterFaultCenter.GatherClusterFaultInfo()
	sort.Slice(faultMsg.NodeFaultInfo, func(i, j int) bool {
		return faultMsg.NodeFaultInfo[i].NodeIP < faultMsg.NodeFaultInfo[j].NodeIP
	})
	faultMsg.JobId = constant.DefaultJobId
	faultMsg.Uuid = string(uuid.NewUUID())
	return &fault.FaultQueryResult{
		Code:        int32(common.SuccessCode),
		Info:        "Succeed",
		FaultSignal: faultMsg,
	}
}

func (s *FaultServer) preemptPublisher(jobId, role string) *config.ConfigPublisher[*fault.FaultMsgSignal] {
	s.lock.Lock()
	defer s.lock.Unlock()
	roleMap, ok := s.faultPublisher[jobId]
	if !ok || roleMap == nil {
		roleMap = make(map[string]*config.ConfigPublisher[*fault.FaultMsgSignal])
	}
	publisher, ok := roleMap[role]
	if ok && publisher != nil {
		publisher.Stop()
	}
	newPublisher := config.NewConfigPublisher[*fault.FaultMsgSignal](jobId, s.serviceCtx,
		constant.FaultMsgDataType, compareFaultMsg)
	roleMap[role] = newPublisher
	s.faultPublisher[jobId] = roleMap
	return newPublisher
}

func (s *FaultServer) getPublisher(jobId, role string) (*config.ConfigPublisher[*fault.FaultMsgSignal], bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	roleMap, ok := s.faultPublisher[jobId]
	if !ok || roleMap == nil {
		return nil, false
	}
	publisher, ok := roleMap[role]
	return publisher, ok
}

func (s *FaultServer) getPublisherListByJobId(jobId string) []*config.ConfigPublisher[*fault.FaultMsgSignal] {
	s.lock.RLock()
	defer s.lock.RUnlock()
	roleMap := s.faultPublisher[jobId]
	publisherList := make([]*config.ConfigPublisher[*fault.FaultMsgSignal], 0, len(roleMap))
	for _, publisher := range roleMap {
		publisherList = append(publisherList, publisher)
	}
	return publisherList
}

func (s *FaultServer) getAllPublisherList() []*config.ConfigPublisher[*fault.FaultMsgSignal] {
	s.lock.RLock()
	defer s.lock.RUnlock()
	publisherList := make([]*config.ConfigPublisher[*fault.FaultMsgSignal], 0, len(s.faultPublisher))
	for _, roleMap := range s.faultPublisher {
		for _, publisher := range roleMap {
			publisherList = append(publisherList, publisher)
		}
	}
	return publisherList
}

func (s *FaultServer) deletePublisher(jobId, role string, createTime time.Time) {
	s.lock.Lock()
	defer s.lock.Unlock()
	roleMap, ok := s.faultPublisher[jobId]
	if !ok || roleMap == nil {
		return
	}
	publisher, ok := roleMap[role]
	if !ok || publisher == nil || !createTime.Equal(publisher.GetCreateTime()) {
		return
	}
	delete(roleMap, role)
	if len(roleMap) == 0 {
		delete(s.faultPublisher, jobId)
	}
}

func (s *FaultServer) addPublisher(jobId, role string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	publisher := config.NewConfigPublisher[*fault.FaultMsgSignal](jobId, s.serviceCtx,
		constant.FaultMsgDataType, compareFaultMsg)
	roleMap, ok := s.faultPublisher[jobId]
	if !ok || roleMap == nil {
		roleMap = make(map[string]*config.ConfigPublisher[*fault.FaultMsgSignal])
	}
	roleMap[role] = publisher
	s.faultPublisher[jobId] = roleMap
}
