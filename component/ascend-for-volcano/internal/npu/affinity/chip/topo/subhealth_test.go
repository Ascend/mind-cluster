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

package topo

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

const shTestNodeName = "sh-topo-node"

func TestSubHealthChips(t *testing.T) {
	convey.Convey("should expose the recorded sub-healthy chips of the tree", t, func() {
		cases := []struct {
			name      string
			raw       string
			subHealth map[int]struct{}
			want      map[int]struct{}
		}{
			{"should return nothing when no chip is sub-healthy",
				"[[0,1],[2,3]]", nil, map[int]struct{}{}},
			{"should return the sub-healthy chips",
				"[[0,1],[2,3]]", map[int]struct{}{0: {}, 3: {}},
				map[int]struct{}{0: {}, 3: {}}},
			{"should return nothing when the tree is nil", "", nil, map[int]struct{}{}},
		}
		for _, c := range cases {
			convey.Convey(c.name, func() {
				root := (*ChipNode)(nil)
				if c.raw != "" {
					root = ParseTopology(c.raw)
					convey.So(root, convey.ShouldNotBeNil)
				}
				if root != nil {
					root.Init(&NodeFaultChips{SubHealth: c.subHealth}, nil, nil)
				}
				convey.So(root.SubHealthChips(shTestNodeName), convey.ShouldResemble, c.want)
			})
		}
	})
}
