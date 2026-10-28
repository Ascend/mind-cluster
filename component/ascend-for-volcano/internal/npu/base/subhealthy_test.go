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

package base

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/cache"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/npu/affinity/chip/topo"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

// shElapsedSeconds is the job-wait snapshot stubbed for every case below
const shElapsedSeconds = 5

// shNodeWithSubHealthChip builds a flat-4 node whose chip 0 is sub-healthy
func shNodeWithSubHealthChip() plugin.NPUNode {
	root := topo.ParseTopology(topo.BuildFlatTopology(4))
	root.Init(&topo.NodeFaultChips{SubHealth: map[int]struct{}{0: {}}}, nil, nil)
	n := createTestNode(nil)
	n.ChipTopo = root
	return n
}

func TestRemoveBackoffCardsSubHealthy(t *testing.T) {
	convey.Convey("should remove sub-healthy chips from nodeTop per the window", t, func() {
		patches := gomonkey.ApplyFuncReturn(cache.SessionElapsed, int64(shElapsedSeconds), true)
		defer patches.Reset()
		node := shNodeWithSubHealthChip()
		nodeTop := []int{0, 1, 2, 3}
		cases := []struct {
			name      string
			afterTime int64
			want      []int
		}{
			{"should remove the sub-healthy chip when the elapsed time is within the window",
				10, []int{1, 2, 3}},
			{"should keep the sub-healthy chip when the elapsed time reaches the window",
				shElapsedSeconds, []int{0, 1, 2, 3}},
			{"should remove the sub-healthy chip when the window is the max tolerance",
				util.MaxSubHealthyToleranceAfterTime, []int{1, 2, 3}},
			{"should keep the sub-healthy chip when released immediately", 0, []int{0, 1, 2, 3}},
		}
		for _, c := range cases {
			convey.Convey(c.name, func() {
				tp := createNPUHandler()
				tp.NPUJob.ToleranceConfig.SubHealthyToleranceAfterTime = c.afterTime
				got := tp.removeBackoffCards(node, nodeTop)
				convey.So(got, convey.ShouldResemble, c.want)
			})
		}
	})
}

func TestSubHealthyToleranceAllows(t *testing.T) {
	convey.Convey("should release sub-healthy resources per the tolerance window", t, func() {
		patches := gomonkey.ApplyFuncReturn(cache.SessionElapsed, int64(shElapsedSeconds), true)
		defer patches.Reset()
		cases := []struct {
			name      string
			handler   *NPUHandler
			afterTime int64
			want      bool
		}{
			{"should allow when the handler is nil", nil, 10, true},
			{"should allow when the job is nil", &NPUHandler{}, 10, true},
			{"should withhold when the elapsed time is within the window", createNPUHandler(), 10, false},
			{"should allow when the elapsed time reaches the window",
				createNPUHandler(), shElapsedSeconds, true},
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
