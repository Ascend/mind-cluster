/* Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package plugins for custom metrics
package plugins

import (
	"sync"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/api"
	"ascend-common/devmanager"
	"huawei.com/npu-exporter/v6/collector/common"
	"huawei.com/npu-exporter/v6/collector/container"
)

const (
	testCacheTime  = 5
	testUpdateTime = 5
)

type isSupportedTestCase struct {
	name     string
	devType  string
	expected bool
}

func getIsSupportedTestCases() []isSupportedTestCase {
	return []isSupportedTestCase{
		{
			name:     "should return true when devType is Ascend310P",
			devType:  api.Ascend310P,
			expected: true,
		},
		{
			name:     "should return true when devType is Ascend910A",
			devType:  api.Ascend910A,
			expected: true,
		},
		{
			name:     "should return true when devType is Ascend910B",
			devType:  api.Ascend910B,
			expected: true,
		},
		{
			name:     "should return true when devType is Ascend910A3",
			devType:  api.Ascend910A3,
			expected: true,
		},
		{
			name:     "should return false when devType is Ascend910A5",
			devType:  api.Ascend910A5,
			expected: false,
		},
		{
			name:     "should return false when devType is Ascend310",
			devType:  api.Ascend310,
			expected: false,
		},
		{
			name:     "should return false when devType is unsupported",
			devType:  "UnsupportedType",
			expected: false,
		},
	}
}

func newTestNpuCollector() *common.NpuCollector {
	return common.NewNpuCollector(
		testCacheTime,
		testUpdateTime,
		&container.DevicesParser{},
		&devmanager.DeviceManager{},
	)
}

func TestIsSupportedForCardNum(t *testing.T) {
	n := newTestNpuCollector()
	testCases := getIsSupportedTestCases()

	for _, tc := range testCases {
		convey.Convey(tc.name, t, func() {
			patches := gomonkey.NewPatches()
			defer patches.Reset()
			patches.ApplyMethodReturn(n.Dmgr, "GetDevType", tc.devType)

			collector := &MachineCardNumPluginInfoCollector{Cache: sync.Map{}}
			result := collector.IsSupported(n)
			convey.So(result, convey.ShouldEqual, tc.expected)
		})
	}
}
