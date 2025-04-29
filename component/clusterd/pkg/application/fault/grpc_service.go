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
	"fmt"
	"reflect"
	"sync"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/application/config"
	"clusterd/pkg/application/faultmanager"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/common"
	"clusterd/pkg/domain/job"
	"clusterd/pkg/interface/grpc/fault"
)

// FaultServer fault server
type FaultServer struct {
	serviceCtx     context.Context
	faultPublisher map[string]*config.ConfigPublisher[*fault.FaultMsgSignal]
	lock           sync.RWMutex
	fault.UnimplementedFaultServer
	faultCh chan map[string]constant.JobFaultInfo
}

// NewFaultServer create a fault server
func NewFaultServer(ctx context.Context) *FaultServer {
	server := &FaultServer{
		serviceCtx:     ctx,
		faultPublisher: make(map[string]*config.ConfigPublisher[*fault.FaultMsgSignal]),
		lock:           sync.RWMutex{},
		faultCh:        make(chan map[string]constant.JobFaultInfo, 5),
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
	if ok && publisher != nil {
		return &fault.Status{Code: int32(common.OK), Info: "register success"}, nil
	}
	code, err := s.preRegistry(req)
	if err != nil {
		hwlog.RunLog.Errorf("jobId=%s, preCheck err:%v", req.JobId, err)
		return &fault.Status{Code: int32(code), Info: err.Error()}, err
	}
	s.addPublisher(req.JobId)
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
	requestInfo := fmt.Sprintf("jobId=%s, rule=%s", request.JobId, request.Role)
	hwlog.RunLog.Infof("receive Subscribe fault message signal request, %s", requestInfo)
	faultPublisher, exist := s.getPublisher(request.JobId)
	if !exist || faultPublisher == nil {
		return fmt.Errorf("jobId=%s not registered, role=%s", request.JobId, request.Role)
	}
	faultPublisher.Stop()
	s.addPublisher(request.JobId)

	faultPublisher, _ = s.getPublisher(request.JobId)
	faultPublisher.ListenDataChange(stream)
	return nil
}

func (s *FaultServer) getPublisher(jobId string) (*config.ConfigPublisher[*fault.FaultMsgSignal], bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	publisher, ok := s.faultPublisher[jobId]
	return publisher, ok
}

func (s *FaultServer) addPublisher(jobId string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	publisher := config.NewConfigPublisher[*fault.FaultMsgSignal](jobId,
		s.serviceCtx, constant.FaultMsgDataType, compareFaultMsg)
	s.faultPublisher[jobId] = publisher
}
