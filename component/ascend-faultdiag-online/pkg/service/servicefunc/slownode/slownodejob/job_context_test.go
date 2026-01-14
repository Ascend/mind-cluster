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
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"

	"ascend-common/common-utils/hwlog"
	"ascend-faultdiag-online/pkg/core/model/enum"
	"ascend-faultdiag-online/pkg/model/slownode"
	"ascend-faultdiag-online/pkg/utils/grpc"
	"ascend-faultdiag-online/pkg/utils/k8s"
)

func init() {
	config := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	err := hwlog.InitRunLogger(&config, nil)
	if err != nil {
		fmt.Println(err)
	}
}

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

func TestStartAndStop(t *testing.T) {
	job := &slownode.Job{}
	ctx := NewJobContext(job, enum.Node)
	var slowRankIds = []int{0, 1, 2}
	var rescheduleCount = 1
	convey.Convey("test Start and stop", t, func() {
		convey.Convey("start job", func() {
			ctx.Start()
			convey.So(ctx.IsRunning(), convey.ShouldBeTrue)
			convey.So(ctx.Step(), convey.ShouldEqual, InitialStep)
			convey.So(ctx.IsDegradation, convey.ShouldBeFalse)
			convey.So(ctx.IsStartedHeavyProfiling(), convey.ShouldBeFalse)
			convey.So(ctx.cluster.NeedReport(), convey.ShouldBeFalse)
			convey.So(ctx.GetSlowRankIds(), convey.ShouldBeEmpty)
			ctx.SetNeedReport(true)
			ctx.SetSlowRankIds(slowRankIds)
			ctx.SetRescheduleCount(rescheduleCount)
			convey.So(ctx.cluster.NeedReport(), convey.ShouldBeTrue)
			convey.So(reflect.DeepEqual(ctx.GetSlowRankIds(), slowRankIds), convey.ShouldBeTrue)
			convey.So(ctx.GetRescheduleCount(), convey.ShouldEqual, rescheduleCount)

		})
		convey.Convey("stop job", func() {
			ctx.Stop()
			convey.So(ctx.IsRunning(), convey.ShouldBeFalse)
			convey.So(ctx.Step(), convey.ShouldEqual, InitialStep)
			convey.So(ctx.IsDegradation, convey.ShouldBeFalse)
			convey.So(ctx.IsStartedHeavyProfiling(), convey.ShouldBeFalse)
			convey.So(ctx.cluster.NeedReport(), convey.ShouldBeFalse)
			convey.So(ctx.GetSlowRankIds(), convey.ShouldBeEmpty)
			convey.So(ctx.GetRescheduleCount(), convey.ShouldEqual, rescheduleCount)
		})
	})
}

func TestAddAlgoRecord(t *testing.T) {
	convey.Convey("test AddAlgoRecord", t, func() {
		convey.Convey("should do nothing when receiver is nil", func() {
			var c *cluster = nil
			convey.So(func() {
				c.AddAlgoRecord(&slownode.ClusterAlgoResult{})
			}, convey.ShouldNotPanic)
		})

		convey.Convey("should do nothing when result is nil", func() {
			c := &cluster{}
			c.AddAlgoRecord(nil)
			convey.So(c.AlgoRes, convey.ShouldBeEmpty)
		})

		convey.Convey("should add record when within capacity", func() {
			c := &cluster{}
			result := &slownode.ClusterAlgoResult{}
			c.AddAlgoRecord(result)
			convey.So(c.AlgoRes, convey.ShouldHaveLength, 1)
			convey.So(c.AlgoRes[0], convey.ShouldEqual, result)
		})

		convey.Convey("should truncate oldest records when exceeding capacity", func() {
			c := &cluster{}
			extra := 10
			for i := 0; i <= recordsCapacity+extra; i++ {
				res := &slownode.ClusterAlgoResult{}
				res.JobId = strconv.Itoa(i)
				c.AddAlgoRecord(res)
			}
			convey.So(c.AlgoRes, convey.ShouldHaveLength, recordsCapacity)

			// the latest one should be recordsCapacity + 1
			convey.So(c.AlgoRes[len(c.AlgoRes)-1].JobId, convey.ShouldEqual, strconv.Itoa(recordsCapacity+extra))
		})
	})
}

func TestUpdateTrainingJobStatus(t *testing.T) {
	convey.Convey("test UpdateTrainingJobStatus", t, func() {
		convey.Convey("should do nothing when receiver is nil", func() {
			var c *cluster = nil
			convey.So(func() {
				c.UpdateTrainingJobStatus(enum.IsRunning)
			}, convey.ShouldNotPanic)
		})

		convey.Convey("should update TrainingJobStatus correctly", func() {
			c := &cluster{}
			convey.So(c.TrainingJobStatus, convey.ShouldEqual, "")
			c.UpdateTrainingJobStatus(enum.IsCompleted)
			convey.So(c.TrainingJobStatus, convey.ShouldEqual, enum.IsCompleted)
			c.UpdateTrainingJobStatus(enum.IsFailed)
			convey.So(c.TrainingJobStatus, convey.ShouldEqual, enum.IsFailed)
		})

		convey.Convey("should handle empty status string", func() {
			c := &cluster{}
			c.UpdateTrainingJobStatus("")
			convey.So(c.TrainingJobStatus, convey.ShouldEqual, "")
		})
	})
}

func TestTriggerMerge(t *testing.T) {
	convey.Convey("test TriggerMerge", t, func() {
		var warnCalled bool

		convey.Convey("should do nothing when receiver is nil", func() {
			warnCalled = false
			var c *cluster = nil
			convey.So(func() {
				c.TriggerMerge()
			}, convey.ShouldNotPanic)
			convey.So(warnCalled, convey.ShouldBeFalse)
		})

		convey.Convey("should send signal successfully when channel is ready", func() {
			warnCalled = false
			ch := make(chan struct{}, 1)
			c := &cluster{MergeParallelGroupInfoSignal: ch}
			c.TriggerMerge()

			select {
			case <-ch:
			default:
				assert.Fail(t, "not receive signal")
			}
			convey.So(warnCalled, convey.ShouldBeFalse)
		})
	})
}

func TestLogPrefix(t *testing.T) {
	convey.Convey("test LogPrefix", t, func() {
		expectedNilPrefix := "[FD-OL SLOWNODE]job(nil)"

		convey.Convey("should return nil prefix when ctx is nil", func() {
			var ctx *JobContext = nil
			prefix := ctx.LogPrefix()
			convey.So(prefix, convey.ShouldEqual, expectedNilPrefix)
		})

		convey.Convey("should return nil prefix when ctx.Job is nil", func() {
			ctx := &JobContext{Job: nil}
			prefix := ctx.LogPrefix()
			convey.So(prefix, convey.ShouldEqual, expectedNilPrefix)
		})

		convey.Convey("should return formatted prefix when Job fields are valid", func() {
			job := &slownode.Job{}
			job.JobName = "name1"
			job.JobId = "job-12345"
			job.Namespace = "default"
			ctx := &JobContext{Job: job}
			prefix := ctx.LogPrefix()
			expected := "[FD-OL SLOWNODE]job(name=name1, namespace=default, jobId=job-12345)"
			convey.So(prefix, convey.ShouldEqual, expected)
		})
	})
}

func TestStartAllProfiling(t *testing.T) {
	convey.Convey("test StartAllProfiling", t, func() {
		var ctx *JobContext
		convey.Convey("ctx or job is nil", func() {
			err := ctx.StartAllProfiling()
			convey.So(err.Error(), convey.ShouldEqual, "ctx is nil or ctx.Job is nil")
			ctx = &JobContext{}
			err = ctx.StartAllProfiling()
			convey.So(err.Error(), convey.ShouldEqual, "ctx is nil or ctx.Job is nil")
		})

		ctx = &JobContext{}
		ctx.Job = &slownode.Job{}

		convey.Convey("get client failed", func() {
			patch := gomonkey.ApplyFuncReturn(grpc.GetClient, nil, errors.New("mock grpc client failed"))
			defer patch.Reset()
			err := ctx.StartAllProfiling()
			convey.So(err.Error(), convey.ShouldEqual, "mock grpc client failed")
		})

		patches := gomonkey.ApplyFuncReturn(grpc.GetClient, &grpc.Client{}, nil)
		defer patches.Reset()

		convey.Convey("start faileds", func() {
			patch := gomonkey.ApplyMethodReturn(&grpc.Client{}, "StartAllProfiling", errors.New("mock start failed"))
			defer patch.Reset()
			err := ctx.StartAllProfiling()
			convey.So(err.Error(), convey.ShouldEqual, "mock start failed")
		})

		patches.ApplyMethodReturn(&grpc.Client{}, "StartAllProfiling", nil)
		convey.Convey("start success", func() {
			err := ctx.StartAllProfiling()
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestStopAllProfiling(t *testing.T) {
	convey.Convey("test StopAllProfiling", t, func() {
		var ctx *JobContext
		convey.Convey("ctx or job is nil", func() {
			convey.So(func() {
				ctx.StartHeavyProfiling()
			}, convey.ShouldNotPanic)
		})

		ctx = &JobContext{}
		ctx.Job = &slownode.Job{}

		convey.Convey("get client failed", func() {
			patch := gomonkey.ApplyFuncReturn(grpc.GetClient, nil, errors.New("mock grpc client failed"))
			defer patch.Reset()
			ctx.StopAllProfiling()
		})

		patches := gomonkey.ApplyFuncReturn(grpc.GetClient, &grpc.Client{}, nil)
		defer patches.Reset()

		convey.Convey("stop faileds", func() {
			patch := gomonkey.ApplyMethodReturn(&grpc.Client{}, "StopAllProfiling", errors.New("mock start failed"))
			defer patch.Reset()
			ctx.StopAllProfiling()
		})

		patches.ApplyMethodReturn(&grpc.Client{}, "StopAllProfiling", nil)
		convey.Convey("stop success", func() {
			ctx.StopAllProfiling()
		})
	})
}

func TestStartHeavyProfiling(t *testing.T) {
	convey.Convey("test StartHeavyProfiling", t, func() {
		var ctx *JobContext
		convey.Convey("ctx or job is nil", func() {
			convey.So(func() {
				ctx.StartHeavyProfiling()
			}, convey.ShouldNotPanic)
		})

		ctx = &JobContext{}
		ctx.Job = &slownode.Job{}

		convey.Convey("get client failed", func() {
			patch := gomonkey.ApplyFuncReturn(grpc.GetClient, nil, errors.New("mock grpc client failed"))
			defer patch.Reset()
			ctx.StartHeavyProfiling()
			convey.So(ctx.isStartedHeavyProfiling, convey.ShouldBeFalse)
		})

		patches := gomonkey.ApplyFuncReturn(grpc.GetClient, &grpc.Client{}, nil)
		defer patches.Reset()

		convey.Convey("stop faileds", func() {
			patch := gomonkey.ApplyMethodReturn(&grpc.Client{}, "StartHeavyProfiling", errors.New("mock start failed"))
			defer patch.Reset()
			ctx.StartHeavyProfiling()
			convey.So(ctx.isStartedHeavyProfiling, convey.ShouldBeFalse)
		})

		patches.ApplyMethodReturn(&grpc.Client{}, "StartHeavyProfiling", nil)
		convey.Convey("stop success", func() {
			ctx.StartHeavyProfiling()
			convey.So(ctx.isStartedHeavyProfiling, convey.ShouldBeTrue)
		})
	})
}

func TestSoptHeavyProfiling(t *testing.T) {
	convey.Convey("test StopHeavyProfiling", t, func() {
		var ctx *JobContext
		convey.Convey("ctx or job is nil", func() {
			convey.So(func() {
				ctx.StartHeavyProfiling()
			}, convey.ShouldNotPanic)
		})

		ctx = &JobContext{}
		ctx.Job = &slownode.Job{}

		convey.Convey("get client failed", func() {
			patch := gomonkey.ApplyFuncReturn(grpc.GetClient, nil, errors.New("mock grpc client failed"))
			defer patch.Reset()
			ctx.StopHeavyProfiling()
		})

		patches := gomonkey.ApplyFuncReturn(grpc.GetClient, &grpc.Client{}, nil)
		defer patches.Reset()

		convey.Convey("stop faileds", func() {
			patch := gomonkey.ApplyMethodReturn(&grpc.Client{}, "StopHeavyProfiling", errors.New("mock start failed"))
			defer patch.Reset()
			ctx.StopHeavyProfiling()
		})

		patches.ApplyMethodReturn(&grpc.Client{}, "StopHeavyProfiling", nil)
		convey.Convey("stop success", func() {
			ctx.StopHeavyProfiling()
			convey.So(ctx.isStartedHeavyProfiling, convey.ShouldBeFalse)
			convey.So(ctx.AlgoRes, convey.ShouldHaveLength, 0)
		})
	})
}

func TestRemoveAllCM(t *testing.T) {
	convey.Convey("test RemoveAllCM", t, func() {
		var ctx *JobContext
		convey.Convey("ctx or job is nil", func() {
			convey.So(func() {
				ctx.RemoveAllCM()
			}, convey.ShouldNotPanic)
		})

		ctx = &JobContext{}
		ctx.Job = &slownode.Job{}
		convey.Convey("get k8s client failed", func() {
			patch := gomonkey.ApplyFuncReturn(k8s.GetClient, nil, errors.New("mock k8s get client failed"))
			defer patch.Reset()
			convey.So(func() {
				ctx.RemoveAllCM()
			}, convey.ShouldNotPanic)
		})

		// insert some cm into ctx.AllCMNames
		ctx.AllCMNames.Store("test1", struct{}{})

		// insert wrong type data
		ctx.AllCMNames.Store(struct{}{}, struct{}{})

		patches := gomonkey.ApplyFuncReturn(k8s.GetClient, &k8s.Client{}, nil)
		patches.ApplyMethodReturn(&k8s.Client{}, "DeleteConfigMap", nil)
		defer patches.Reset()
		convey.Convey("remove cm successfully", func() {
			convey.So(func() {
				ctx.RemoveAllCM()
			}, convey.ShouldNotPanic)
		})
	})
}
