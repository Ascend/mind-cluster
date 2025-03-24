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

// Package busconfig business configuration service for grpc client
package busconfig

import (
	"context"
	"sync"
	"time"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/interface/grpc/config"
)

const (
	retryTimes     = 3
	chanBufferSize = 1000
)

// ConfigPublisher save the rank table and send it to the client
type ConfigPublisher struct {
	jobId          string
	rankTableChan  chan *config.RankTableStream
	ctxContext     context.Context
	ctxCancelFunc  context.CancelFunc
	serviceContext context.Context
	lock           sync.RWMutex
}

// NewConfigPublisher create a config publisher
func NewConfigPublisher(jobId string, serviceCtx context.Context) *ConfigPublisher {
	publisher := &ConfigPublisher{
		jobId:          jobId,
		rankTableChan:  make(chan *config.RankTableStream, chanBufferSize),
		serviceContext: serviceCtx,
		lock:           sync.RWMutex{},
	}
	publisher.ctxContext, publisher.ctxCancelFunc = context.WithCancel(publisher.serviceContext)
	return publisher
}

func (c *ConfigPublisher) listenRankTableChange(stream config.Config_SubscribeRankTableServer) {
	hwlog.RunLog.Infof("start listen a new rankTableChan, jobId=%s", c.jobId)
	for {
		if !c.selectChanAndContext(stream) {
			break
		}
	}
}

func (c *ConfigPublisher) selectChanAndContext(stream config.Config_SubscribeRankTableServer) bool {
	select {
	case <-c.ctxContext.Done():
		hwlog.RunLog.Warnf("context canceled, jobId=%s", c.jobId)
		return false
	case <-stream.Context().Done():
		hwlog.RunLog.Warnf("stream is closed, do not send ranktable jobId=%s", c.jobId)
		return false
	case data, ok := <-c.rankTableChan:
		if ok {
			sendRankTable(stream, data)
			return true
		} else {
			hwlog.RunLog.Warnf("rankTableChan closed, jobId=%s break listen rankTableChan", c.jobId)
			return false
		}
	}
}

func sendRankTable(stream config.Config_SubscribeRankTableServer, data *config.RankTableStream) {
	for i := 0; i < retryTimes; i++ {
		err := stream.Send(data)
		if err == nil {
			hwlog.RunLog.Infof("send ranktable success, jobId=%s", data.JobId)
			return
		}
		hwlog.RunLog.Errorf("send ranktable failed, jobId=%s, error= %v", data.JobId, err)
		if i < retryTimes-1 {
			time.Sleep(time.Second)
		}
	}
}

// SaveData save data to rankTableChan
func (c *ConfigPublisher) SaveData(jobId, data string) bool {
	saved := true
	defer func() {
		if r := recover(); r != nil {
			saved = false
			hwlog.RunLog.Errorf("panic occured when saving rank table, err: %v", r)
		}
	}()
	if len(c.rankTableChan) >= chanBufferSize {
		hwlog.RunLog.Warnf("rankTableChan is full, do not send rank table")
		return false
	}
	rankTable := &config.RankTableStream{
		JobId:     jobId,
		RankTable: data,
	}
	c.rankTableChan <- rankTable
	return saved
}

func (c *ConfigPublisher) stop() {
	hwlog.RunLog.Infof("jobId=%s enter publisher stop function", c.jobId)
	if c.ctxCancelFunc != nil {
		c.ctxCancelFunc()
	}
	close(c.rankTableChan)
}
