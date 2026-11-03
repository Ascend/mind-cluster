/*
Copyright(C)2026. Huawei Technologies Co.,Ltd. All rights reserved.

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

package rescheduling

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/cache"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

const (
	// shElapsedSeconds is the job-wait snapshot stubbed for every case below
	shElapsedSeconds = 5
	// shJobName is the job owning the tasks in the cases below
	shJobName = api.JobID("vcjob/sh-job")
	// shNodeName is the node checked in the cases below
	shNodeName = "sh-node"
)

// shSchedulerJob builds a job with the given sub-healthy strategy and tolerance
// window in seconds
func shSchedulerJob(strategy string, afterTime int64) plugin.SchedulerJob {
	return plugin.SchedulerJob{SchedulerJobAttr: util.SchedulerJobAttr{
		ComJob: util.ComJob{Name: shJobName},
		NPUJob: &util.NPUJob{SubHealthyStrategy: strategy,
			ToleranceConfig: util.JobToleranceConfig{SubHealthyToleranceAfterTime: afterTime}},
	}}
}

// shReScheduler builds a scheduler that owns one fault node and one job
func shReScheduler(strategy string, afterTime int64, switchSub bool) *ReScheduler {
	return &ReScheduler{
		DealReSchedulerCache: &DealReSchedulerCache{FaultNodes: map[string]*FaultNode{
			shNodeName: {HasSwitchSubHealthFault: switchSub},
		}},
		Jobs: map[api.JobID]plugin.SchedulerJob{shJobName: shSchedulerJob(strategy, afterTime)},
	}
}

func TestSubHealthyToleranceAllows(t *testing.T) {
	convey.Convey("should release the sub-healthy resources per the job's tolerance window", t, func() {
		patches := gomonkey.ApplyFuncReturn(cache.SessionElapsed, int64(shElapsedSeconds), true)
		defer patches.Reset()
		cases := []struct {
			name      string
			afterTime int64
			want      bool
		}{
			{"should not allow when the elapsed time is within the window", 10, false},
			{"should allow when the elapsed time reaches the window", shElapsedSeconds, true},
			{"should allow immediately when the after-time is 0", 0, true},
			{"should not allow when the after-time is the max tolerance",
				util.MaxSubHealthyToleranceAfterTime, false},
		}
		for _, c := range cases {
			convey.Convey(c.name, func() {
				job := shSchedulerJob(util.SubHealthyIgnore, c.afterTime)
				convey.So(subHealthyToleranceAllows(job), convey.ShouldEqual, c.want)
			})
		}
		convey.Convey("should allow when the job has no npu job", func() {
			convey.So(subHealthyToleranceAllows(plugin.SchedulerJob{}), convey.ShouldBeTrue)
		})
	})
}

func TestCheckNodeNPUByTaskSubHealthyWindow(t *testing.T) {
	convey.Convey("should gate the switch sub-healthy node per the tolerance window", t, func() {
		patches := gomonkey.ApplyFuncReturn(cache.SessionElapsed, int64(shElapsedSeconds), true)
		defer patches.Reset()
		cases := []struct {
			name      string
			strategy  string
			switchSub bool
			afterTime int64
			wantErr   string
		}{
			{"should withhold a switch sub-healthy node within the window",
				util.SubHealthyIgnore, true, 10, "sub-healthy"},
			{"should release a switch sub-healthy node after the window",
				util.SubHealthyIgnore, true, shElapsedSeconds, ""},
			{"should release a healthy node within the window",
				util.SubHealthyIgnore, false, 10, ""},
			{"should reject a switch sub-healthy node by strategy when it is not ignore",
				util.SubHealthyForceExit, true, 10, "NodePredicate failed"},
		}
		for _, c := range cases {
			convey.Convey(c.name, func() {
				rdma := gomonkey.ApplyFuncReturn(util.IsRdmaTask, false)
				defer rdma.Reset()
				reScheduler := shReScheduler(c.strategy, c.afterTime, c.switchSub)
				node := &plugin.NPUNode{CommonNode: plugin.CommonNode{Name: shNodeName,
					Annotation: map[string]string{}}}
				task := &api.TaskInfo{Job: shJobName}
				err := reScheduler.CheckNodeNPUByTask(task, node)
				if c.wantErr == "" {
					convey.So(err, convey.ShouldBeNil)
					return
				}
				convey.So(err, convey.ShouldNotBeNil)
				convey.So(err.Error(), convey.ShouldContainSubstring, c.wantErr)
			})
		}
	})
}
