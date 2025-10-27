/* Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package container provides utilities for container monitoring and testing.
package container

import (
	"errors"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
)

const (
	// Test endpoint constants
	testContainerdEndpoint = "unix:///run/containerd.sock"
	testDockerEndpoint     = "unix:///run/docker.sock"

	device0              = 0
	device1              = 1
	device2              = 2
	device3              = 3
	testDeviceRange      = "0-2"
	testDeviceComma      = "0,1,2"
	testDeviceCommaRange = "0-1,2-3"
	testAscendDevice0    = "Ascend-0"
	testAscendDevices    = "Ascend-0,Ascend-1"
	testMixedDevices     = "0-1,3"

	// Test error constants
	testOriginalError  = "original error"
	testErrorMessage   = "test message"
	testContactedError = "original error->test message"

	// Test path constants
	testDevicePattern = "/dev/npu([0-9]+)"

	// Test duration constants
	testZeroDuration = 0
)

func TestMakeDevicesParser(t *testing.T) {
	testCases := []struct {
		name     string
		opts     CntNpuMonitorOpts
		expected *DevicesParser
	}{
		{name: "should create parser when options are valid for containerd",
			opts: CntNpuMonitorOpts{CriEndpoint: testContainerdEndpoint, EndpointType: EndpointTypeContainerd,
				OciEndpoint: testContainerdEndpoint, UserBackUp: false},
			expected: &DevicesParser{RuntimeOperator: &RuntimeOperatorTool{UseBackup: false,
				CriEndpoint: testContainerdEndpoint, OciEndpoint: testContainerdEndpoint}, Timeout: testZeroDuration}},
		{name: "should create parser when options are valid for docker",
			opts: CntNpuMonitorOpts{CriEndpoint: testDockerEndpoint, EndpointType: EndpointTypeDockerd,
				OciEndpoint: testDockerEndpoint, UserBackUp: true},
			expected: &DevicesParser{RuntimeOperator: &RuntimeOperatorTool{UseBackup: true,
				CriEndpoint: testDockerEndpoint, OciEndpoint: testDockerEndpoint}, Timeout: testZeroDuration}},
		{name: "should create parser when options are valid for isula",
			opts: CntNpuMonitorOpts{CriEndpoint: testContainerdEndpoint, EndpointType: EndpointTypeIsula,
				OciEndpoint: testContainerdEndpoint, UserBackUp: true},
			expected: &DevicesParser{RuntimeOperator: &RuntimeOperatorTool{UseBackup: true,
				CriEndpoint: testContainerdEndpoint, OciEndpoint: testContainerdEndpoint}, Timeout: testZeroDuration}},
	}

	for _, tc := range testCases {
		convey.Convey(tc.name, t, func() {
			result := MakeDevicesParser(tc.opts)
			convey.So(result, convey.ShouldNotBeNil)
			convey.So(result.RuntimeOperator, convey.ShouldNotBeNil)
			convey.So(result.Timeout, convey.ShouldEqual, tc.expected.Timeout)
		})
	}
}

func TestDevicesParserInit(t *testing.T) {
	convey.Convey("TestDevicesParserInit", t, func() {
		convey.Convey("should initialize successfully when runtime operator init succeeds", func() {
			dp := &DevicesParser{
				RuntimeOperator: &RuntimeOperatorTool{},
			}

			patches := gomonkey.ApplyMethodReturn(dp.RuntimeOperator, "Init", nil)
			defer patches.Reset()

			err := dp.Init()
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("should return error when initialization fails", func() {
			dp := &DevicesParser{
				RuntimeOperator: &RuntimeOperatorTool{},
			}
			patches := gomonkey.ApplyMethodReturn(dp.RuntimeOperator, "Init", errors.New("init failed"))
			defer patches.Reset()
			err := dp.Init()
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldContainSubstring, "init failed")
		})
	})
}
