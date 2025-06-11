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
	"clusterd/pkg/common/util"
	"clusterd/pkg/domain/common"
	"clusterd/pkg/domain/job"
	"clusterd/pkg/interface/grpc/fault"
)

const (
	jobFaultInfoChanCache = 5
	defaultTokenRate      = 10
	defaultBurst          = 10
	defaultMaxQueueLen    = 50
)

// FaultServer fault server
type FaultServer struct {
	serviceCtx     context.Context
	faultPublisher map[string]*config.ConfigPublisher[*fault.FaultMsgSignal]
	lock           sync.RWMutex
	fault.UnimplementedFaultServer
	faultCh chan map[string]constant.JobFaultInfo
	limiter *util.AdvancedRateLimiter
}

// NewFaultServer create a fault server
func NewFaultServer(ctx context.Context) *FaultServer {
	server := &FaultServer{
		serviceCtx:     ctx,
		faultPublisher: make(map[string]*config.ConfigPublisher[*fault.FaultMsgSignal]),
		lock:           sync.RWMutex{},
		faultCh:        make(chan map[string]constant.JobFaultInfo, jobFaultInfoChanCache),
		limiter:        util.NewAdvancedRateLimiter(defaultTokenRate, defaultBurst, defaultMaxQueueLen),
	}
	if err := faultmanager.RegisterForJobFaultRank(server.faultCh, reflect.TypeOf(server).Name()); err != nil {
		hwlog.RunLog.Error("RegisterForJobFaultRank fail")
	}
	go server.checkFaultFromFaultCenter()
	return server
}

// Register is task register service
func (s *FaultServer) Register(ctx context.Context, req *fault.ClientInfo) (*fault.Status, error) {
	hwlog.RunLog.Infof("fault service receive Register request, jobId=%s, role=%s",
		req.JobId, req.Role)
	publisher, ok := s.getPublisher(req.JobId)
	if !ok || publisher == nil {
		code, err := s.preRegistry(req)
		if err != nil {
			hwlog.RunLog.Errorf("jobId=%s, preCheck err:%v", req.JobId, err)
			return &fault.Status{Code: int32(code), Info: err.Error()}, err
		}
	}
	s.preemptPublisher(req.JobId)
	return &fault.Status{Code: int32(common.OK), Info: "register success"}, nil
}

func (s *FaultServer) preRegistry(req *fault.ClientInfo) (common.RespCode, error) {
	_, ok := job.GetJobCache(req.JobId)
	_, err := job.GetNamespaceByJobIdAndAppType(req.JobId, req.Role)
	if !ok && err != nil {
		hwlog.RunLog.Errorf("jobId=%s not exist and is not multi-instance job", req.JobId)
		return common.JobNotExist, fmt.Errorf("jobId=%s not exist and is not multi-instance", req.JobId)
	}
	if s.serveJobNum() >= constant.MaxServeJobs {
		return common.OutOfMaxServeJobs,
			fmt.Errorf("jobId=%s out of max serve jobs", req.JobId)
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
	requestInfo := fmt.Sprintf("jobId=%s, role=%s", request.JobId, request.Role)
	hwlog.RunLog.Infof("receive Subscribe fault message signal request, %s", requestInfo)
	faultPublisher, exist := s.getPublisher(request.JobId)
	if !exist || faultPublisher == nil {
		hwlog.RunLog.Warnf("jobId=%s not registered, role=%s", request.JobId, request.Role)
		return fmt.Errorf("jobId=%s not registered, role=%s", request.JobId, request.Role)
	}
	faultPublisher.ListenDataChange(stream)
	s.deletePublisher(request.JobId, faultPublisher.GetCreateTime())
	hwlog.RunLog.Infof("jobId=%s stop subscribe fault message signal, createTime=%v",
		request.JobId, faultPublisher.GetCreateTime().UnixNano())
	return nil
}

// GetFaultMsgSignal return cluster fault
func (s *FaultServer) GetFaultMsgSignal(ctx context.Context, request *fault.ClientInfo) (*fault.FaultQueryResult, error) {
	hwlog.RunLog.Infof("role: %#v call get faults", request)
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
	jobFaultInfo := faultrank.JobFaultRankProcessor.GetJobFaultRankInfos()
	faultInfo, ok := jobFaultInfo[jobId]
	if !ok {
		return &fault.FaultQueryResult{
			Code:        int32(common.SuccessCode),
			Info:        fmt.Sprintf("job with jobId: %v not found", jobId),
			FaultSignal: nil,
		}, nil
	}
	faultMsg := faultDeviceToSortedFaultMsgSignal(jobId, faultInfo.FaultDevice)
	faultMsg.Uuid = string(uuid.NewUUID())
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
	faultMsg.JobId = "-1"
	faultMsg.Uuid = string(uuid.NewUUID())
	return &fault.FaultQueryResult{
		Code:        int32(common.SuccessCode),
		Info:        "Succeed",
		FaultSignal: faultMsg,
	}
}

func (s *FaultServer) preemptPublisher(jobId string) *config.ConfigPublisher[*fault.FaultMsgSignal] {
	s.lock.Lock()
	defer s.lock.Unlock()
	publisher, ok := s.faultPublisher[jobId]
	if ok && publisher != nil {
		publisher.Stop()
	}
	newPublisher := config.NewConfigPublisher[*fault.FaultMsgSignal](jobId,
		s.serviceCtx, constant.FaultMsgDataType, compareFaultMsg)
	s.faultPublisher[jobId] = newPublisher
	return newPublisher
}

func (s *FaultServer) getPublisher(jobId string) (*config.ConfigPublisher[*fault.FaultMsgSignal], bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	publisher, ok := s.faultPublisher[jobId]
	return publisher, ok
}

func (s *FaultServer) deletePublisher(jobId string, createTime time.Time) {
	s.lock.Lock()
	defer s.lock.Unlock()
	publisher, ok := s.faultPublisher[jobId]
	if !ok || publisher == nil || !createTime.Equal(publisher.GetCreateTime()) {
		return
	}
	delete(s.faultPublisher, jobId)
}

func (s *FaultServer) addPublisher(jobId string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	publisher := config.NewConfigPublisher[*fault.FaultMsgSignal](jobId,
		s.serviceCtx, constant.FaultMsgDataType, compareFaultMsg)
	s.faultPublisher[jobId] = publisher
}
