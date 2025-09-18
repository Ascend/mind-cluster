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
	"fmt"
	"sync"

	"k8s.io/apimachinery/pkg/util/uuid"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/domain/common"
	"clusterd/pkg/interface/grpc/job"
)

const (
	msgCacheNumPerClient = 10
	maxClientNum         = 20
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
}

// JobServer job info server
type JobServer struct {
	job.UnimplementedJobServer
	clients map[string]*clientState
	mu      sync.RWMutex
}

func init() {
	clientWhiteList = map[string]bool{"CCAgent": true, "DefaultUser1": true, "DefaultUser2": true, "FdAgent": true}
}

// NewJobServer create a new job info server
func NewJobServer(ctx context.Context) *JobServer {
	jobserver := &JobServer{
		clients: make(map[string]*clientState),
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
	if len(s.clients) >= maxClientNum {
		hwlog.RunLog.Warn("too many clients registered for job info")
		return &job.Status{
			Code:     common.RateLimitedCode,
			Info:     "too many clients registered",
			ClientId: "",
		}, nil
	}

	clientId := string(uuid.NewUUID())
	s.clients[clientId] = &clientState{
		clientChan: make(chan job.JobSummarySignal, msgCacheNumPerClient),
		role:       req.Role,
		closed:     false,
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
		hwlog.RunLog.Errorf("invalid clientId, please register first")
		return fmt.Errorf("invalid clientId: %s, please register first", req.ClientId)
	}
	s.mu.Unlock()
	ctx := stream.Context()
	defer func() {
		s.mu.Lock()
		delete(s.clients, req.ClientId)
		s.mu.Unlock()
		cltState.safeCloseChannel()
		hwlog.RunLog.Infof("client %s disconnected, role: %s", req.ClientId, cltState.role)
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case jobInfo, ok := <-cltState.clientChan:
			if !ok {
				return fmt.Errorf("client channel closed")
			}
			jobInfo.Uuid = string(uuid.NewUUID())
			if err := stream.Send(&jobInfo); err != nil {
				hwlog.RunLog.Errorf("error sending to client %s: %v", req.ClientId, err)
				return fmt.Errorf("error sending to client %s: %v", req.ClientId, err)
			}
			hwlog.RunLog.Debugf("Sent job summary signal to client %s", req.ClientId)
		}
	}
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

func (cs *clientState) safeCloseChannel() {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	if !cs.closed {
		close(cs.clientChan)
		cs.closed = true
		hwlog.RunLog.Debug("Channel closed for client")
	}
}
