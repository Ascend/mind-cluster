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

// Package common for general constants
package common

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/devmanager/hccn"
)

// TestGetNpuDevNetPortInfos test getNpuDevNetPortInfos function
func TestGetNpuDevNetPortInfos(t *testing.T) {
	convey.Convey("TestGetNpuDevNetPortInfos", t, func() {
		// Use existing mock
		n := mockNewNpuCollector()

		// Setup mocks for success case
		patches := gomonkey.NewPatches()
		defer patches.Reset()

		// Mock device list with one device
		patches.ApplyMethodReturn(n.Dmgr, "GetDeviceList", int32(0), []int32{0}, nil)
		// Mock port info
		patches.ApplyFuncReturn(hccn.GetNpuDevNetPortInfo, map[int][]int{0: {0}}, nil)

		// Test function
		err := getNpuDevNetPortInfos(n)
		convey.So(err, convey.ShouldBeNil)
	})
}
