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

// Package config business configuration service for grpc client
package config

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/common"
	"clusterd/pkg/domain/epranktable"
	"clusterd/pkg/interface/grpc/config"
)

const (
	waitTime = 100 * time.Millisecond
)

// BusinessConfigServer business config server
type BusinessConfigServer struct {
	serviceCtx       context.Context
	configPublisher  map[string]*ConfigPublisher[*config.RankTableStream]
	rankTableManager *epranktable.RankTableManager
	lock             sync.RWMutex
	config.UnimplementedConfigServer
}

// NewBusinessConfigServer create a business config server
func NewBusinessConfigServer(ctx context.Context) *BusinessConfigServer {
	server := &BusinessConfigServer{
		serviceCtx:      ctx,
		configPublisher: make(map[string]*ConfigPublisher[*config.RankTableStream]),
		lock:            sync.RWMutex{},
	}
	server.rankTableManager = epranktable.GetEpGlobalRankTableManager()
	server.rankTableManager.HandlerRankTable = server.rankTableChange
	return server
}

// rankTableChange The callback function when the rank table changes
func (c *BusinessConfigServer) rankTableChange(jobId, data string) (bool, error) {
	publisher, ok := c.getPublisher(jobId)
	if !ok {
		return true, errors.New("job not registered")
	}
	hwlog.RunLog.Infof("ranktable changed, jobId=%s", jobId)
	rankTable := &config.RankTableStream{
		JobId:     jobId,
		RankTable: data,
	}
	if isSaved := publisher.SaveData(rankTable); !isSaved {
		return true, errors.New("save data failed")
	}
	return false, nil
}

// Register is task register service
func (c *BusinessConfigServer) Register(ctx context.Context, req *config.ClientInfo) (*config.Status, error) {
	hwlog.RunLog.Infof("business config service receive Register request, jobId=%s, role=%s",
		req.JobId, req.Role)
	publisher, ok := c.getPublisher(req.JobId)
	if ok && publisher != nil {
		publisher.Stop()
		for {
			if _, ok = c.getPublisher(req.JobId); !ok {
				break
			}
			time.Sleep(waitTime)
		}
	}
	c.addPublisher(req.JobId)
	return &config.Status{Code: int32(common.OK), Info: "register success"}, nil
}

// SubscribeRankTable subscribe rank table from ClusterD
func (c *BusinessConfigServer) SubscribeRankTable(request *config.ClientInfo,
	stream config.Config_SubscribeRankTableServer) error {
	hwlog.RunLog.Infof("receive Subscribe ranktable request, jobId=%s, rule=%s",
		request.JobId, request.Role)
	publisher, ok := c.getPublisher(request.JobId)
	if !ok || publisher == nil {
		hwlog.RunLog.Warnf("jobId=%s not registered, role=%s", request.JobId, request.Role)
		return fmt.Errorf("jobId=%s not registered, role=%s", request.JobId, request.Role)
	}
	epranktable.GetRankTableMessageQueue().AddRateLimited(&epranktable.GenerateGlobalRankTableMessage{
		JobId:     request.JobId,
		Namespace: "",
	})
	publisher.ListenDataChange(stream)
	c.deletePublisher(request.JobId)
	return nil
}

func (c *BusinessConfigServer) getPublisher(jobId string) (*ConfigPublisher[*config.RankTableStream], bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	publisher, ok := c.configPublisher[jobId]
	return publisher, ok
}

func (c *BusinessConfigServer) deletePublisher(jobId string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	delete(c.configPublisher, jobId)
}

func (c *BusinessConfigServer) addPublisher(jobId string) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	publisher := NewConfigPublisher[*config.RankTableStream](jobId, c.serviceCtx, constant.RankTableDataType, nil)
	c.configPublisher[jobId] = publisher
}
