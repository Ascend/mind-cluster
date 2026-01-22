/* Copyright(C) 2025-2025. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package metrics for general collector
package metrics

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/api"
)

// TestRoceCollectorIsSupported test RoceCollector IsSupported
func TestRoceCollectorIsSupported(t *testing.T) {
	n := mockNewNpuCollector()
	cases := []testCase{
		buildTestCase("NetworkCollector: testIsSupported on Ascend910A3", &RoceCollector{}, api.Ascend910A3, true),
		buildTestCase("NetworkCollector: testIsSupported on Ascend910A5", &RoceCollector{}, api.Ascend910A5, false),
	}

	for _, c := range cases {
		patches := gomonkey.NewPatches()
		convey.Convey(c.name, t, func() {
			defer patches.Reset()
			patches.ApplyMethodReturn(n.Dmgr, "IsTrainingCard", true)
			patches.ApplyMethodReturn(n.Dmgr, "GetDevType", c.deviceType)
			isSupported := c.collectorType.IsSupported(n)
			convey.So(isSupported, convey.ShouldEqual, c.expectValue)
		})
	}
}
