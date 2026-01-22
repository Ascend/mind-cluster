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

// Package jobinfo is used to return job info by subscribe
package jobinfo

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/time/rate"
	"k8s.io/apimachinery/pkg/util/uuid"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/logs"
	"clusterd/pkg/common/util"
	"clusterd/pkg/domain/common"
	jobstorage "clusterd/pkg/domain/job"
	"clusterd/pkg/interface/grpc/job"
)

const (
	ten = 10
)

var (
	clientWhiteList = make(map[string]bool)
)

// ClientState to indicate client state
type clientState struct {
	clientChan chan job.JobSummarySignal
	role       string
	mu         sync.RWMutex
	closed     bool
	ctx        context.Context
	cancelCtx  context.CancelFunc
	ctxCount   int32
}

// JobServer job info server
type JobServer struct {
	job.UnimplementedJobServer
	clients map[string]*clientState
	mu      sync.RWMutex
	limiter *rate.Limiter
}

func init() {
	clientWhiteList = map[string]bool{"CCAgent": true, "DefaultUser1": true, "DefaultUser2": true, "FdAgent": true}
}

// NewJobServer create a new job info server
func NewJobServer(ctx context.Context) *JobServer {
	jobserver := &JobServer{
		clients: make(map[string]*clientState),
		limiter: rate.NewLimiter(rate.Every(time.Second/constant.RequestNumPerSecondLimit),
			constant.RequestNumPerSecondLimit),
	}
	jobserver.startBroadcasting(ctx)
	return jobserver
}

// Register to register a new watching client
func (s *JobServer) Register(ctx context.Context, req *job.ClientInfo) (*job.Status, error) {
	hwlog.RunLog.Infof("role: %v call Register", req.Role)
	if !clientWhiteList[req.Role] {
		hwlog.RunLog.Warnf("role:%v is not in whitelist:%#v", req.Role, clientWhiteList)
		return &job.Status{
			Code:     int32(common.UnRegistry),
			Info:     fmt.Sprintf("role:%v is not in whitelist:%#v", req.Role, clientWhiteList),
			ClientId: "",
		}, fmt.Errorf("role:%v is not in whitelist", req.Role)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	roleClientCount := 0
	for _, client := range s.clients {
		if client.role == req.Role {
			roleClientCount++
		}
	}
	if roleClientCount >= constant.MaxClientPerRole {
		errMsg := fmt.Sprintf("role %s has reached max client limit (%d), cannot register new client",
			req.Role, constant.MaxClientPerRole)
		hwlog.RunLog.Warnf(errMsg)
		return &job.Status{
			Code:     int32(common.RateLimitedCode),
			Info:     errMsg,
			ClientId: "",
		}, fmt.Errorf(errMsg)
	}

	clientId := string(uuid.NewUUID())
	s.clients[clientId] = &clientState{
		clientChan: make(chan job.JobSummarySignal, constant.MsgCacheNumPerClient),
		role:       req.Role,
		closed:     false,
		ctx:        nil,
		cancelCtx:  nil,
		ctxCount:   0,
	}
	hwlog.RunLog.Infof("client registered: %s, role: %s", clientId, req.Role)
	return &job.Status{
		Code:     int32(common.SuccessCode),
		Info:     "registration successful",
		ClientId: clientId,
	}, nil
}

// SubscribeJobSummarySignal to subscribe all job info
func (s *JobServer) SubscribeJobSummarySignal(req *job.ClientInfo,
	stream job.Job_SubscribeJobSummarySignalServer) error {
	hwlog.RunLog.Infof("role: %v call SubscribeJobSummarySignal, clientId: %s", req.Role, req.ClientId)
	s.mu.Lock()
	cltState, exists := s.clients[req.ClientId]
	if !exists {
		s.mu.Unlock()
		hwlog.RunLog.Errorf("invalid clientId: %s, please register first", req.ClientId)
		return fmt.Errorf("invalid clientId: %s, please register first", req.ClientId)
	}
	s.manageClientContext(cltState, stream.Context(), req.ClientId)
	s.mu.Unlock()
	defer s.cleanupClientContext(cltState, req.ClientId)
	if !s.limiter.Allow() {
		hwlog.RunLog.Errorf("rate limited, there is too many requests, please retry later")
		return fmt.Errorf("rate limited, there is too many requests, please retry later")
	}
	// Get and push all the jobs if role is FdAgent
	if cltState.role == constant.RoleFdAgent {
		allJobs := jobstorage.GetAllJobCache()
		for _, jobInfo := range allJobs {
			jobSummary := BuildJobSignalFromJobInfo(
				jobInfo, util.ObjToString(jobInfo.JobRankTable), constant.AddOperator)
			if err := stream.Send(&jobSummary); err != nil {
				errMsg := fmt.Sprintf("job: %v error sending to client %s: %v", jobSummary.JobId, req.ClientId, err)
				hwlog.RunLog.Error(errMsg)
				return errors.New(errMsg)
			}
		}
	}

	for {
		select {
		case <-cltState.ctx.Done():
			return cltState.ctx.Err()
		case jobInfo, ok := <-cltState.clientChan:
			if !ok {
				return fmt.Errorf("client channel closed")
			}
			if err := s.handleSingleJobInfo(&jobInfo); err != nil {
				return fmt.Errorf("handle large npu job failed: %v", err)
			}
			if err := stream.Send(&jobInfo); err != nil {
				hwlog.RunLog.Errorf("error sending to client %s: %v", req.ClientId, err)
				return fmt.Errorf("error sending to client %s: %v", req.ClientId, err)
			}
			hwlog.RunLog.Debugf("sent job summary signal to client %s", req.ClientId)
		}
	}
}

// SubscribeJobSummarySignalList to subscribe job infos
func (s *JobServer) SubscribeJobSummarySignalList(req *job.ClientInfo,
	stream job.Job_SubscribeJobSummarySignalListServer) error {
	hwlog.RunLog.Infof("role: %v call SubscribeJobSummarySignalList, clientId: %s", req.Role, req.ClientId)
	s.mu.Lock()
	cltState, exists := s.clients[req.ClientId]
	if !exists {
		s.mu.Unlock()
		hwlog.RunLog.Errorf("invalid clientId: %s, please register first", req.ClientId)
		return fmt.Errorf("invalid clientId: %s, please register first", req.ClientId)
	}
	s.manageClientContext(cltState, stream.Context(), req.ClientId)
	s.mu.Unlock()
	defer s.cleanupClientContext(cltState, req.ClientId)
	if !s.limiter.Allow() {
		hwlog.RunLog.Errorf("rate limited, there is too many requests, please retry later")
		return fmt.Errorf("rate limited, there is too many requests, please retry later")
	}
	allBatchJobSummarySignals, allBatchJobIds := GetAllBatchJobSummarySignals()
	reportTime := time.Now().UnixMilli()
	for i, batchJobSummarySignal := range allBatchJobSummarySignals {
		jobInfos := job.JobSummarySignalList{
			JobSummarySignals: batchJobSummarySignal,
			ReportTime:        strconv.FormatInt(reportTime, ten),
			JobTotalNum:       int32(len(batchJobSummarySignal))}
		if err := stream.Send(&jobInfos); err != nil {
			hwlog.RunLog.Errorf("send full job summary signal to client %s (role: %s) failed: %v",
				req.ClientId, req.Role, err)
			return err
		}
		logs.JobEventLog.Infof("subscribeJobSummarySignalList report all jobs, batchJobIds: %v, "+
			"reportTime: %v, jobTotalNum: %v", allBatchJobIds[i], jobInfos.ReportTime, jobInfos.JobTotalNum)
	}
	return s.handleStreamJobSend(cltState, req, stream)
}

func (s *JobServer) manageClientContext(cltState *clientState, streamCtx context.Context, clientId string) {
	cltState.mu.Lock()
	defer cltState.mu.Unlock()
	if cltState.ctx != nil && cltState.cancelCtx != nil {
		hwlog.RunLog.Infof("client %s has old ctx, cancel it first", clientId)
		cltState.cancelCtx()
	}
	newCtx, newCancel := context.WithCancel(streamCtx)
	cltState.ctx = newCtx
	cltState.cancelCtx = newCancel
	atomic.AddInt32(&cltState.ctxCount, 1)
	hwlog.RunLog.Infof("client %s init new context, ctx count increase to %d",
		clientId, atomic.LoadInt32(&cltState.ctxCount))
}

func (s *JobServer) cleanupClientContext(cltState *clientState, clientId string) {
	currentCount := atomic.AddInt32(&cltState.ctxCount, -1)
	hwlog.RunLog.Infof("client %s ctx count decrease to %d", clientId, currentCount)
	if currentCount <= 0 {
		s.mu.Lock()
		delete(s.clients, clientId)
		s.mu.Unlock()
		cltState.safeCloseClientResources()
		hwlog.RunLog.Infof("client %s ctx count is 0, delete it, role: %s", clientId, cltState.role)
	}
}

func (s *JobServer) handleStreamJobSend(cltState *clientState, req *job.ClientInfo,
	stream job.Job_SubscribeJobSummarySignalListServer) error {
	for {
		select {
		case <-cltState.ctx.Done():
			return cltState.ctx.Err()
		case jobInfo, ok := <-cltState.clientChan:
			if !ok {
				return fmt.Errorf("client channel closed")
			}
			if err := s.handleSingleJobInfo(&jobInfo); err != nil {
				return fmt.Errorf("handle large npu job failed: %v", err)
			}
			jobInfos := &job.JobSummarySignalList{
				JobSummarySignals: []*job.JobSummarySignal{&jobInfo},
				ReportTime:        strconv.FormatInt(time.Now().UnixMilli(), ten),
				JobTotalNum:       1,
			}
			if err := stream.Send(jobInfos); err != nil {
				hwlog.RunLog.Errorf("error sending to client %s: %v", req.ClientId, err)
				return fmt.Errorf("error sending to client %s: %v", req.ClientId, err)
			}
			hwlog.RunLog.Debugf("sent job summary signal to client %s", req.ClientId)
			logs.JobEventLog.Infof("subscribeJobSummarySignalList report job, jobId: %v, "+
				"reportTime: %v, jobTotalNum: %v", jobInfo.JobId, jobInfos.ReportTime, jobInfos.JobTotalNum)
		}
	}
}

func (s *JobServer) handleSingleJobInfo(jobInfo *job.JobSummarySignal) error {
	if jobInfo == nil {
		return fmt.Errorf("invalid job info")
	}
	jobNPUNum := strings.Count(jobInfo.HcclJson, "rank_id")
	if jobNPUNum > constant.MaxNPUsPerBatch {
		jobInfo.HcclJson = ""
		hwlog.RunLog.Warnf("job %s NPU num(%d) exceed max threshold(%d), set HcclJson empty",
			jobInfo.JobId, jobNPUNum, constant.MaxNPUsPerBatch)
	}
	return nil
}

func (s *JobServer) startBroadcasting(ctx context.Context) {
	if jobUpdateChan == nil {
		jobUpdateChan = make(chan job.JobSummarySignal, jobUpdateChanCache)
	}
	go func() {
		for {
			select {
			case <-ctx.Done():
				hwlog.RunLog.Info("job info service stop broadcasting")
				return
			case jobSignal := <-jobUpdateChan:
				s.broadcastJobUpdate(jobSignal)
			}
		}
	}()
}

func (s *JobServer) broadcastJobUpdate(jobSignal job.JobSummarySignal) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var wg sync.WaitGroup
	wg.Add(len(s.clients))
	for clientId, ch := range s.clients {
		if ch == nil || ch.closed {
			hwlog.RunLog.Debugf("client %s chan may be closed", clientId)
			continue
		}
		select {
		case ch.clientChan <- jobSignal:
			hwlog.RunLog.Debugf("broadcasted to client %s", clientId)
		default:
			hwlog.RunLog.Warnf("client %s channel buffer is full, dropping message", clientId)
		}
		wg.Done()
	}
	wg.Wait()
}

func (cs *clientState) safeCloseClientResources() {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	if !cs.closed {
		if cs.cancelCtx != nil {
			cs.cancelCtx()
			cs.ctx = nil
			cs.cancelCtx = nil
		}
		close(cs.clientChan)
		cs.closed = true
		hwlog.RunLog.Debug("Channel closed for client")
	}
}

func processJobSliceForBatchSignals(jobMap map[string]constant.JobInfo, maxNPUs int) (
	[][]*job.JobSummarySignal, [][]string) {
	allBatchJobSummarySignals, allBatchJobIds := make([][]*job.JobSummarySignal, 0), make([][]string, 0)
	batchSignals, batchJobIds := make([]*job.JobSummarySignal, 0), make([]string, 0)
	accumulatedNPUs := 0
	for _, jobInfo := range jobMap {
		jobNPUNum := calcJobNPUNum(jobInfo)
		if jobNPUNum > maxNPUs {
			hwlog.RunLog.Infof("job %s NPU num(%d) exceed max threshold(%d), set HcclJson empty",
				jobInfo.Key, jobNPUNum, maxNPUs)
			if len(batchSignals) > 0 {
				allBatchJobSummarySignals = append(allBatchJobSummarySignals, batchSignals)
				allBatchJobIds = append(allBatchJobIds, batchJobIds)
				batchSignals, batchJobIds, accumulatedNPUs = []*job.JobSummarySignal{}, []string{}, 0
			}
			jobSummary := BuildJobSignalFromJobInfo(jobInfo, "", constant.AddOperator)
			allBatchJobSummarySignals = append(allBatchJobSummarySignals, []*job.JobSummarySignal{&jobSummary})
			allBatchJobIds = append(allBatchJobIds, []string{jobSummary.JobId})
			continue
		}
		if accumulatedNPUs+jobNPUNum > maxNPUs && len(batchSignals) > 0 {
			hwlog.RunLog.Infof("accumulated NPU num(%d) exceed max threshold(%d), job %s",
				jobNPUNum, maxNPUs, jobInfo.Key)
			allBatchJobSummarySignals = append(allBatchJobSummarySignals, batchSignals)
			allBatchJobIds = append(allBatchJobIds, batchJobIds)
			batchSignals, batchJobIds, accumulatedNPUs = []*job.JobSummarySignal{}, []string{}, 0
		}
		hcclJson := util.ObjToString(jobInfo.JobRankTable)
		jobSummary := BuildJobSignalFromJobInfo(jobInfo, hcclJson, constant.AddOperator)
		batchSignals = append(batchSignals, &jobSummary)
		batchJobIds = append(batchJobIds, jobSummary.JobId)
		accumulatedNPUs += jobNPUNum
	}
	if len(batchSignals) > 0 {
		allBatchJobSummarySignals = append(allBatchJobSummarySignals, batchSignals)
		allBatchJobIds = append(allBatchJobIds, batchJobIds)
	}
	return allBatchJobSummarySignals, allBatchJobIds
}

// GetAllBatchJobSummarySignals get all batch job summary signals
func GetAllBatchJobSummarySignals() ([][]*job.JobSummarySignal, [][]string) {
	allBatchJobSummarySignals := make([][]*job.JobSummarySignal, 0)
	allBatchJobIds := make([][]string, 0)
	jobMap := jobstorage.GetAllJobCache()
	if len(jobMap) == 0 {
		return allBatchJobSummarySignals, allBatchJobIds
	}
	return processJobSliceForBatchSignals(jobMap, constant.MaxNPUsPerBatch)
}
