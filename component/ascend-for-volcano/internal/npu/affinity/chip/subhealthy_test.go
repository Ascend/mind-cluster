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

package chip

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/cache"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/npu/affinity/chip/topo"
)

// shElapsedSeconds is the job-wait snapshot stubbed for every case below
const (
	shElapsedSeconds = 5
	shTestNodeName   = "sh-node"
)

func TestSubHealthyToleranceAllows(t *testing.T) {
	convey.Convey("should release sub-healthy resources per the tolerance window", t, func() {
		patches := gomonkey.ApplyFuncReturn(cache.SessionElapsed, int64(shElapsedSeconds), true)
		defer patches.Reset()
		cases := []struct {
			name      string
			handler   *chipHandler
			afterTime int64
			want      bool
		}{
			{"should allow when the handler is nil", nil, 10, true},
			{"should allow when the job is nil", &chipHandler{}, 10, true},
			{"should withhold when the elapsed time is within the window",
				newTestHandler(), 10, false},
			{"should allow when the elapsed time reaches the window",
				newTestHandler(), shElapsedSeconds, true},
			{"should allow immediately when the after-time is 0", newTestHandler(), 0, true},
			{"should withhold when the after-time is the max tolerance",
				newTestHandler(), util.MaxSubHealthyToleranceAfterTime, false},
		}
		for _, c := range cases {
			convey.Convey(c.name, func() {
				if c.handler != nil && c.handler.NPUJob != nil {
					c.handler.NPUJob.ToleranceConfig.SubHealthyToleranceAfterTime = c.afterTime
				}
				convey.So(c.handler.subHealthyToleranceAllows(), convey.ShouldEqual, c.want)
			})
		}
	})
}

func TestExcludedChipIDsSubHealthy(t *testing.T) {
	convey.Convey("should withhold sub-healthy chips per the tolerance window", t, func() {
		patches := gomonkey.ApplyFuncReturn(cache.SessionElapsed, int64(shElapsedSeconds), true)
		defer patches.Reset()
		node := flatNode(shTestNodeName, 4)
		node.ChipTopo.Init(&topo.NodeFaultChips{SubHealth: map[int]struct{}{0: {}}}, nil, nil)
		cases := []struct {
			name      string
			afterTime int64
			want      map[int]struct{}
		}{
			{"should withhold chip 0 when the elapsed time is within the window",
				10, map[int]struct{}{0: {}}},
			{"should release chip 0 when the elapsed time reaches the window", 5, nil},
			{"should withhold chip 0 when the window is the max tolerance",
				util.MaxSubHealthyToleranceAfterTime, map[int]struct{}{0: {}}},
			{"should release chip 0 immediately when the after-time is 0", 0, nil},
		}
		for _, c := range cases {
			convey.Convey(c.name, func() {
				tp := newTestHandler()
				tp.NPUJob.ToleranceConfig = util.JobToleranceConfig{
					SubHealthyToleranceAfterTime: c.afterTime}
				convey.So(tp.excludedChipIDs(node), convey.ShouldResemble, c.want)
			})
		}
	})
}

func TestExcludedChipIDsSilentAndSubHealthy(t *testing.T) {
	convey.Convey("should gate silent-fault and sub-healthy chips independently", t, func() {
		patches := gomonkey.ApplyFuncReturn(cache.SessionElapsed, int64(shElapsedSeconds), true)
		defer patches.Reset()
		node := flatNode(shTestNodeName, 4)
		node.ChipTopo.Init(&topo.NodeFaultChips{
			Silent: map[int]struct{}{1: {}}, SubHealth: map[int]struct{}{0: {}}}, nil, nil)
		convey.Convey("should withhold only the sub-healthy chip when silent is released", func() {
			tp := newTestHandler()
			tp.NPUJob.ToleranceConfig = util.JobToleranceConfig{
				SilentUnhealthyToleranceAfterTime: 0,
				SubHealthyToleranceAfterTime:      10}
			convey.So(tp.excludedChipIDs(node), convey.ShouldResemble, map[int]struct{}{0: {}})
		})
		convey.Convey("should withhold both chips when both windows are active", func() {
			tp := newTestHandler()
			tp.NPUJob.ToleranceConfig = util.JobToleranceConfig{
				SilentUnhealthyToleranceAfterTime: util.ToleranceForbidden,
				SubHealthyToleranceAfterTime:      10}
			want := map[int]struct{}{0: {}, 1: {}}
			convey.So(tp.excludedChipIDs(node), convey.ShouldResemble, want)
		})
	})
}
