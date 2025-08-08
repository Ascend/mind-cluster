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

// Package slownodejob is a DT collection for func in slownode_context
package slownodejob

import (
	"reflect"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"

	"ascend-faultdiag-online/pkg/core/model/enum"
	"ascend-faultdiag-online/pkg/model/slownode"
)

func TestSlowNode(t *testing.T) {
	var job = &slownode.Job{
		SlowNode: 1,
	}
	job.JobName = "job1"
	ctx := NewJobContext(job, enum.Node)

	// test job start
	assert.Equal(t, false, ctx.IsRunning())
	ctx.Start()
	assert.Equal(t, true, ctx.IsRunning())

	// test job stop
	ctx.Stop()
	assert.Equal(t, false, ctx.IsRunning())

	// test update
	assert.Equal(t, 1, ctx.Job.SlowNode)
	ctx.Update(&slownode.Job{SlowNode: 0})
	assert.Equal(t, 0, ctx.Job.SlowNode)

}

func TestAllNodesReported(t *testing.T) {
	convey.Convey("test allNodesReported", t, func() {
		ctx := &JobContext{
			Job: &slownode.Job{
				Servers: []slownode.Server{
					{
						Ip: "192.168.0.1",
					},
					{
						Ip: "192.168.0.2",
					},
				},
			},
		}
		convey.Convey("all nodes reported. got True", func() {
			ctx.AddReportedNodeIp("192.168.0.1")
			ctx.AddReportedNodeIp("192.168.0.2")
			convey.So(ctx.AllNodesReported(), convey.ShouldBeTrue)
		})
		convey.Convey("same node report, got True", func() {
			ctx.AddReportedNodeIp("192.168.0.1")
			ctx.AddReportedNodeIp("192.168.0.2")
			ctx.AddReportedNodeIp("192.168.0.2")
			ctx.AddReportedNodeIp("192.168.0.2")
			ctx.AddReportedNodeIp("192.168.0.2")
			convey.So(ctx.AllNodesReported(), convey.ShouldBeTrue)
		})
		convey.Convey("not equal, got False", func() {
			ctx.AddReportedNodeIp("192.168.0.2")
			convey.So(reflect.DeepEqual(ctx.GetReportedNodeIps(), []string{"192.168.0.2"}), convey.ShouldBeTrue)
			convey.So(ctx.AllNodesReported(), convey.ShouldBeFalse)
		})
	})
}
