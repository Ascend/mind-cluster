/*
Copyright(C)2025. Huawei Technologies Co.,Ltd. All rights reserved.

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

// Package slownodejob provides a sync map for storing slow node and the context
package slownodejob

import (
	"sync"
)

// ctxMap is a sync map that stores slow node feature function config.
type ctxMap struct {
	data sync.Map // {<jobId>: {*context.FaultDiagContext, *SlowNode}}
}

// Insert inserts a key-value pair into the slowNodeConfMap.
func (s *ctxMap) Insert(key string, value *JobContext) {
	s.data.Store(key, value)
}

// Get retrieves the value associated with the key from the slowNodeConfMap.
func (s *ctxMap) Get(key string) (*JobContext, bool) {
	value, ok := s.data.Load(key)
	if !ok {
		return nil, false
	}
	ctx, ok := value.(*JobContext)
	if !ok {
		return nil, false
	}
	return ctx, true
}

// Delete removes the key-value pair from the slowNodeConfMap.
func (s *ctxMap) Delete(key string) {
	s.data.Delete(key)
}

// Clear clears all key-value pairs from the slowNodeConfMap.
func (s *ctxMap) Clear() {
	s.data.Range(func(key, ctx any) bool {
		s.data.Delete(key)
		return true
	})
}

// GetByJobId get the ctx by jobId, key for cluster: namespace/name, key for node: jobId, this only for cluster
func (c *ctxMap) GetByJobId(jobId string) (*JobContext, bool) {
	var instance *JobContext
	var exist = false
	c.data.Range(func(key, ctx any) bool {
		jobCtx, ok := ctx.(*JobContext)
		if ok && jobCtx.Job.JobId == jobId {
			instance = jobCtx
			exist = true
			return false
		}
		return true
	})
	return instance, exist
}

// GetByNodeIp get the ctx by nodeIp, return a array includes all ctx which has nodeIp
func (c *ctxMap) GetByNodeIp(nodeIp string) []*JobContext {
	var jobCtxList []*JobContext
	c.data.Range(func(key, ctx any) bool {
		jobCtx, ok := ctx.(*JobContext)
		if !ok {
			return true
		}
		for _, server := range jobCtx.Job.Servers {
			if server.Ip == nodeIp {
				jobCtxList = append(jobCtxList, jobCtx)
				break
			}
		}
		return true
	})
	return jobCtxList
}

var (
	jobCtxMap *ctxMap
	once      sync.Once
)

// GetJobCtxMap returns a instance of ctxMap.
func GetJobCtxMap() *ctxMap {
	once.Do(func() {
		jobCtxMap = &ctxMap{}
	})
	return jobCtxMap
}
