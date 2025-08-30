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
	"io"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/logs"
	"clusterd/pkg/interface/grpc/config"
	"clusterd/pkg/interface/grpc/fault"
	"clusterd/pkg/interface/grpc/profiling"
)

const (
	retryTimes     = 3
	waitSendTime   = 3 * time.Second
	chanBufferSize = 1000
)

type signalType interface {
	*config.RankTableStream | *fault.FaultMsgSignal | *profiling.DataStatusRes
}

type grpcServerStreamType[T signalType] interface {
	Send(T) error
	grpc.ServerStream
}

// jobDataForChan job data for chan
type jobDataForChan[T signalType] struct {
	jobId string
	data  T
}

// ConfigPublisher save data and send it to the client
type ConfigPublisher[T signalType] struct {
	jobId          string
	role           string
	dataType       string
	sendChan       chan *jobDataForChan[T]
	sentData       map[string]T
	subscribe      bool
	compareFunc    func(T, T) bool
	ctxContext     context.Context
	ctxCancelFunc  context.CancelFunc
	serviceContext context.Context
	isChanClosed   bool
	createTime     time.Time
	lock           sync.RWMutex
}

// NewConfigPublisher create a config publisher
func NewConfigPublisher[T signalType](jobId string, serviceCtx context.Context, dataType string,
	compareFunc func(T, T) bool) *ConfigPublisher[T] {
	publisher := &ConfigPublisher[T]{
		jobId:          jobId,
		dataType:       dataType,
		sendChan:       make(chan *jobDataForChan[T], chanBufferSize),
		sentData:       make(map[string]T),
		subscribe:      false,
		compareFunc:    compareFunc,
		serviceContext: serviceCtx,
		isChanClosed:   false,
		createTime:     time.Now(),
		lock:           sync.RWMutex{},
	}
	publisher.ctxContext, publisher.ctxCancelFunc = context.WithCancel(publisher.serviceContext)
	return publisher
}

func (c *ConfigPublisher[T]) ListenDataChange(stream grpcServerStreamType[T]) {
	hwlog.RunLog.Infof("start listen a new %s sendChan, jobId=%s, createTime=%v",
		c.dataType, c.jobId, c.createTime.UnixNano())
	c.SetSubscribe(true)
	for {
		if !c.selectChanAndContext(stream) {
			break
		}
	}
	c.SetSubscribe(false)
}

func (c *ConfigPublisher[T]) selectChanAndContext(stream grpcServerStreamType[T]) bool {
	select {
	case <-c.ctxContext.Done():
		hwlog.RunLog.Warnf("context canceled, jobId=%s", c.jobId)
		return false
	case <-stream.Context().Done():
		hwlog.RunLog.Warnf("stream is closed, do not send %s, jobId=%s", c.dataType, c.jobId)
		return false
	case data, ok := <-c.sendChan:
		if ok {
			if data == nil || (c.compareFunc != nil && c.compareFunc(data.data, c.GetSentData(data.jobId))) {
				return true
			}
			sendSuccess, stillListen := sendDataToClient(stream, data.data, c.jobId, c.dataType)
			if sendSuccess {
				c.SetSentData(data.jobId, data.data)
			}
			return stillListen
		} else {
			hwlog.RunLog.Warnf("%s sendChan closed, jobId=%s break listen sendChan", c.dataType, c.jobId)
			return false
		}
	}
}

func sendDataToClient[T signalType](stream grpcServerStreamType[T], data T, jobId, dataType string) (bool, bool) {
	timer := time.NewTimer(time.Second)
	defer timer.Stop()
	for i := 0; i < retryTimes; i++ {
		err := sendWithTimeout(stream, data)
		if err == nil {
			hwlog.RunLog.Infof("send %s success, jobId=%s", dataType, jobId)
			logs.GrpcEventLogger.Infof("send %s success, jobId=%s, data=%v", dataType, jobId, data)
			return true, true
		}
		if err == io.EOF {
			hwlog.RunLog.Warnf("send %s failed, client cancel connection, jobId=%s", dataType, jobId)
			return false, false
		}
		hwlog.RunLog.Errorf("send %s failed, jobId=%s, error= %v", dataType, jobId, err)
		if i >= retryTimes-1 {
			break
		}
		timer.Reset(time.Second)
		select {
		case <-timer.C:
			continue
		case <-stream.Context().Done():
			hwlog.RunLog.Warnf("stream is closed, do not send %s, jobId=%s", dataType, jobId)
			return false, false
		}
	}
	return false, true
}

func sendWithTimeout[T signalType](stream grpcServerStreamType[T], data T) error {
	errChan := make(chan error, 1)
	go func() {
		errChan <- stream.Send(data)
	}()
	timer := time.NewTimer(waitSendTime)
	defer timer.Stop()

	select {
	case err := <-errChan:
		return err
	case <-timer.C:
		return status.Error(codes.DeadlineExceeded, "send data timeout")
	}
}

// SaveData save data to sendChan
func (c *ConfigPublisher[T]) SaveData(jobId string, data T) bool {
	saved := true
	defer func() {
		if r := recover(); r != nil {
			saved = false
			hwlog.RunLog.Errorf("panic occured when saving %s, jobId=%s err=%v", c.dataType, c.jobId, r)
		}
	}()
	if len(c.sendChan) >= chanBufferSize {
		hwlog.RunLog.Warnf("sendChan is full, do not send %s jobId=%s", c.dataType, c.jobId)
		return false
	}
	c.sendChan <- &jobDataForChan[T]{jobId: jobId, data: data}
	return saved
}

func (c *ConfigPublisher[T]) Stop() {
	hwlog.RunLog.Infof("jobId=%s enter %s stop function", c.jobId, c.dataType)
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.isChanClosed {
		return
	}
	if c.ctxCancelFunc != nil {
		c.ctxCancelFunc()
	}
	close(c.sendChan)
	c.isChanClosed = true
}

// SetSubscribe set subscribe when client subscribe to or unsubscribe from the service
func (c *ConfigPublisher[T]) SetSubscribe(isSubscribed bool) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.subscribe = isSubscribed
}

// IsSubscribed return whether the client has subscribed or not
func (c *ConfigPublisher[T]) IsSubscribed() bool {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.subscribe
}

// SetSentData store successfully sent data
func (c *ConfigPublisher[T]) SetSentData(jobId string, data T) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.sentData[jobId] = data
}

// GetSentData return the latest successfully sent data
func (c *ConfigPublisher[T]) GetSentData(jobId string) T {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.sentData[jobId]
}

// GetAllSentJobIdList get all sent job id list
func (c *ConfigPublisher[T]) GetAllSentJobIdList() []string {
	jobIdList := make([]string, 0, len(c.sentData))
	c.lock.RLock()
	defer c.lock.RUnlock()
	for jobId, _ := range c.sentData {
		jobIdList = append(jobIdList, jobId)
	}
	return jobIdList
}

// ClearDeletedJobIdList clear job key that already be deleted
func (c *ConfigPublisher[T]) ClearDeletedJobIdList(jobKeyList []string) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	for _, jobId := range jobKeyList {
		delete(c.sentData, jobId)
	}
}

// GetSentChan return sendChan
func (c *ConfigPublisher[T]) GetSentChan() chan *jobDataForChan[T] {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.sendChan
}

// GetCreateTime return publisher create time
func (c *ConfigPublisher[T]) GetCreateTime() time.Time {
	return c.createTime
}

// GetJobId return job id
func (c *ConfigPublisher[T]) GetJobId() string {
	if c == nil {
		return ""
	}
	return c.jobId
}
