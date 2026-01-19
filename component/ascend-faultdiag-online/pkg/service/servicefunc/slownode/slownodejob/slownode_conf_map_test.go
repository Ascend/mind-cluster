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

// Package slownodejob is a DT collection for func in slownode_conf_map
package slownodejob

import (
	"fmt"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"

	"ascend-faultdiag-online/pkg/core/model/enum"
	"ascend-faultdiag-online/pkg/model/slownode"
)

func TestCtxMap(t *testing.T) {
	// insert a context
	var jobName = "job1"
	var job = &slownode.Job{}
	job.JobName = jobName
	job.SlowNode = 1
	ctx := NewJobContext(job, enum.Node)
	slowNodeMap := GetJobCtxMap()
	slowNodeMap.Insert(jobName, ctx)

	// get the context
	value, ok := slowNodeMap.Get(jobName)
	assert.Equal(t, true, ok)
	assert.Equal(t, jobName, value.Job.JobName)
	assert.Equal(t, 1, value.Job.SlowNode)

	// delete the context
	slowNodeMap.Delete(jobName)
	_, ok = slowNodeMap.Get(jobName)
	assert.Equal(t, false, ok)

	// insert job
	slowNodeMap.Insert(jobName, ctx)

	// clear all contexts
	slowNodeMap.Clear()
	_, ok = slowNodeMap.Get(jobName)
	assert.Equal(t, false, ok)
}

func TestGetByJobId(t *testing.T) {
	GetJobCtxMap().Clear()
	ctx := &JobContext{Job: &slownode.Job{}}
	ctx.Job.JobId = "testJobId"
	GetJobCtxMap().Insert("test", ctx)

	instance, ok := GetJobCtxMap().GetByJobId("ttt")
	assert.False(t, ok)
	assert.Nil(t, instance)

	instance, ok = GetJobCtxMap().GetByJobId("testJobId")
	assert.True(t, ok)
	assert.NotNil(t, instance)
}

func TestGetByNodeIp(t *testing.T) {
	var count = 10
	var ip1 = "127.0.0.1"
	var ip2 = "127.0.0.2"
	var ip3 = "127.0.0.3"

	convey.Convey("test GetByNodeIp", t, func() {
		GetJobCtxMap().Clear()
		// insert 10 data with ip 127.0.0.1
		for i := 0; i < count; i++ {
			key := fmt.Sprintf("key1-%d", i)
			GetJobCtxMap().Insert(key, &JobContext{Job: &slownode.Job{
				Servers: []slownode.Server{{Ip: ip1}},
			}})
		}
		// insert 10 data with ip 127.0.0.2
		for i := 0; i < count; i++ {
			key := fmt.Sprintf("key2-%d", i)
			GetJobCtxMap().Insert(key, &JobContext{Job: &slownode.Job{
				Servers: []slownode.Server{{Ip: ip2}},
			}})
		}
		// query by  127.0.0.1
		ctxList := GetJobCtxMap().GetByNodeIp(ip1)
		convey.So(len(ctxList), convey.ShouldEqual, count)
		// query by  127.0.0.2
		ctxList = GetJobCtxMap().GetByNodeIp(ip2)
		convey.So(len(ctxList), convey.ShouldEqual, count)
		// query by  127.0.0.3
		ctxList = GetJobCtxMap().GetByNodeIp(ip3)
		convey.So(ctxList, convey.ShouldBeEmpty)
	})
}
