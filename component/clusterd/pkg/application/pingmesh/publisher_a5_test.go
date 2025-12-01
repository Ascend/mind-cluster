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

/*
Package pingmesh a series of function handle ping mesh configmap create/update/delete.
*/
package pingmesh

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"

	"ascend-common/api"
)

const (
	testSuperPodID1 = 1
	testSuperPodID2 = 2
	testSuperPodID3 = 3
	testSuperPodID4 = 4
)

func TestFilterA5SuperPodIDFromMap(t *testing.T) {
	convey.Convey("Test filterA5SuperPodIDFromMap", t, func() {
		convey.Convey("When input map is nil", func() {
			result := filterA5SuperPodIDFromMap(nil)
			convey.So(result, convey.ShouldBeEmpty)
		})
		convey.Convey("When input map is empty", func() {
			result := filterA5SuperPodIDFromMap(map[int]string{})
			convey.So(result, convey.ShouldBeEmpty)
		})
		convey.Convey("When all items are A5 pods", func() {
			input := map[int]string{
				testSuperPodID1: api.A5PodType,
				testSuperPodID2: api.A5PodType,
				testSuperPodID3: api.A5PodType,
			}
			result := filterA5SuperPodIDFromMap(input)
			convey.So(result, convey.ShouldContain, testSuperPodID1)
			convey.So(result, convey.ShouldContain, testSuperPodID2)
			convey.So(result, convey.ShouldContain, testSuperPodID3)
		})
		convey.Convey("When no items are A5 pods", func() {
			input := map[int]string{
				testSuperPodID1: "other_type1",
				testSuperPodID2: "other_type2",
				testSuperPodID3: "other_type3",
			}
			result := filterA5SuperPodIDFromMap(input)
			convey.So(result, convey.ShouldBeEmpty)
		})
		convey.Convey("When mixed A5 and non-A5 pods", func() {
			input := map[int]string{
				testSuperPodID1: api.A5PodType,
				testSuperPodID2: "other_type",
				testSuperPodID3: api.A5PodType,
				testSuperPodID4: "another_type",
			}
			result := filterA5SuperPodIDFromMap(input)
			convey.So(result, convey.ShouldContain, testSuperPodID1)
			convey.So(result, convey.ShouldContain, testSuperPodID3)
		})
	})
}
