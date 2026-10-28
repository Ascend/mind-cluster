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

package plugin

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
)

const (
	// shTestJobName test job name for the sub-healthy backoff cases
	shTestJobName = "sh-job"
	// shDevListPre is the npu name prefix used by the -Fault annotation cases
	shDevListPre = util.NPU910CardNamePre
	// shSubHealthFaultAnno a -Fault annotation whose fault_handling is SubHealthFault on chip 0 and 1
	shSubHealthFaultAnno = `[{"npu_name":"Ascend910-0,Ascend910-1","fault_handling":"SubHealthFault"}]`
	// shResetHandlingAnno a -Fault annotation whose fault_handling is not SubHealthFault
	shResetHandlingAnno = `[{"npu_name":"Ascend910-0","fault_handling":"Reset"}]`
	// shEmptyNPUNameAnno a -Fault annotation carrying SubHealthFault but no npu_name
	shEmptyNPUNameAnno = `[{"fault_handling":"SubHealthFault"}]`
	// shMixedFaultAnno a -Fault annotation carrying SilentFault on chip 0 and SubHealthFault on chip 1
	shMixedFaultAnno = `[{"npu_name":"Ascend910-0","fault_handling":"SilentFault"},` +
		`{"npu_name":"Ascend910-1","fault_handling":"SubHealthFault"}]`
	// shInvalidFaultAnno a malformed -Fault annotation
	shInvalidFaultAnno = "not-json"
)

// shBuildJob builds a SchedulerJob whose only inputs are the annotations
func shBuildJob(anno map[string]string) *SchedulerJob {
	return &SchedulerJob{SchedulerJobAttr: util.SchedulerJobAttr{
		ComJob: util.ComJob{Name: shTestJobName, Annotation: anno},
		NPUJob: &util.NPUJob{},
	}}
}

func TestSetSubHealthyTolerance(t *testing.T) {
	convey.Convey("should set the sub-healthy tolerance from the job annotation", t, func() {
		cases := []struct {
			name     string
			anno     map[string]string
			wantTime int64
		}{
			{"should use default when the annotation is missing", nil,
				util.DefaultSubHealthyToleranceAfterTime},
			{"should allow immediately when the annotation is 0",
				map[string]string{util.SubHealthyToleranceAnnoKey: "0"}, 0},
			{"should clamp to lower bound when the annotation is negative",
				map[string]string{util.SubHealthyToleranceAnnoKey: "-1"}, 0},
			{"should clamp to max when the annotation exceeds the upper bound",
				map[string]string{util.SubHealthyToleranceAnnoKey: "301"},
				util.MaxSubHealthyToleranceAfterTime},
			{"should clamp to lower bound when the annotation is below zero",
				map[string]string{util.SubHealthyToleranceAnnoKey: "-2"}, 0},
			{"should use default when the annotation is not a number",
				map[string]string{util.SubHealthyToleranceAnnoKey: "abc"},
				util.DefaultSubHealthyToleranceAfterTime},
		}
		for _, c := range cases {
			convey.Convey(c.name, func() {
				sJob := shBuildJob(c.anno)
				sJob.setSubHealthyTolerance()
				got := sJob.NPUJob.ToleranceConfig.SubHealthyToleranceAfterTime
				convey.So(got, convey.ShouldEqual, c.wantTime)
			})
		}
	})
}

func TestClampSubHealthyTolerance(t *testing.T) {
	convey.Convey("should clamp the sub-healthy tolerance into its bounds", t, func() {
		cases := []struct {
			name      string
			tolerance int64
			want      int64
		}{
			{"should clamp to lower bound when tolerance is -1", -1, 0},
			{"should clamp to lower bound when tolerance is -2", -2, 0},
			{"should keep the value when tolerance is 0", 0, 0},
			{"should keep the value when tolerance is 10", 10, 10},
			{"should keep the max when tolerance is the upper bound", 300, 300},
			{"should cap the value when tolerance exceeds the upper bound", 301, 300},
		}
		for _, c := range cases {
			convey.Convey(c.name, func() {
				convey.So(clampSubHealthyTolerance("", c.tolerance), convey.ShouldEqual, c.want)
			})
		}
	})
}

func TestFaultCardsFromAnno(t *testing.T) {
	convey.Convey("should classify the silent and sub-healthy chips from the -Fault annotation", t, func() {
		cases := []struct {
			name          string
			anno          string
			wantSilent    map[int]struct{}
			wantSubHealth map[int]struct{}
		}{
			{"should return chip 0 and 1 as sub-healthy when fault_handling is SubHealthFault",
				shSubHealthFaultAnno, map[int]struct{}{}, map[int]struct{}{0: {}, 1: {}}},
			{"should return nothing when fault_handling is neither SilentFault nor SubHealthFault",
				shResetHandlingAnno, map[int]struct{}{}, map[int]struct{}{}},
			{"should return nothing when npu_name is empty",
				shEmptyNPUNameAnno, map[int]struct{}{}, map[int]struct{}{}},
			{"should return nothing when the annotation is malformed",
				shInvalidFaultAnno, map[int]struct{}{}, map[int]struct{}{}},
			{"should classify each chip by its fault_handling when marks are mixed",
				shMixedFaultAnno, map[int]struct{}{0: {}}, map[int]struct{}{1: {}}},
		}
		for _, c := range cases {
			convey.Convey(c.name, func() {
				silent, subHealth := faultCardsFromAnno(c.anno, shDevListPre)
				convey.So(silent, convey.ShouldResemble, c.wantSilent)
				convey.So(subHealth, convey.ShouldResemble, c.wantSubHealth)
			})
		}
	})
}
