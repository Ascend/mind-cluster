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

// Package device a series of device function
package device

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

const (
	containerID = "testContainer"
)

type testCase struct {
	name        string
	devices     string
	containerID string
	expected    []int
}

func TestGetDeviceIDsByCommaStyle(t *testing.T) {
	tests := []testCase{
		{name: "devices contains invalid subString, return deviceId slice after filtering out the invalid strings",
			devices: "0,abc,1", containerID: containerID, expected: []int{0, 1}},
		{name: "devices is an empty string, return empty slice",
			devices: "", containerID: containerID, expected: []int{}},
		{name: "devices is valid string, return deviceId slice",
			devices: "0,1", containerID: containerID, expected: []int{0, 1}},
	}

	for _, test := range tests {
		convey.Convey(test.name, t, func() {
			actual := getDeviceIDsByCommaStyle(test.devices, test.containerID)
			convey.So(actual, convey.ShouldResemble, test.expected)
		})
	}
}

func TestGetDeviceIDsByAscendStyle(t *testing.T) {
	tests := []testCase{
		{name: "devices is an empty string, return empty slice",
			devices: "", containerID: containerID, expected: []int{}},
		{name: "devices contains a valid device, return deviceId slice",
			devices: "dev-0", containerID: containerID, expected: []int{0}},
		{name: "devices contains multiple valid devices, return deviceId slice",
			devices: "dev-0,dev-1", containerID: containerID, expected: []int{0, 1}},
		{name: "devices is a string with invalid device, return empty slice",
			devices: "dev1", containerID: containerID, expected: []int{}},
		{name: "devices is a string with invalid deviceId, return empty slice",
			devices: "dev-a", containerID: containerID, expected: []int{}},
	}

	for _, test := range tests {
		convey.Convey(test.name, t, func() {
			actual := getDeviceIDsByAscendStyle(test.devices, test.containerID)
			convey.So(actual, convey.ShouldResemble, test.expected)
		})
	}
}

func TestGetDeviceIDsByMinusStyle(t *testing.T) {
	tests := []testCase{
		{name: "devices is valid string, min id is less than or equal to max, " +
			"and max is less than or equal to math.MaxInt16, return deviceId slice",
			devices: "0-1", containerID: containerID, expected: []int{0, 1}},
		{name: "devices does not contain '-', return empty slice",
			devices: "1", containerID: containerID, expected: []int{}},
		{name: "devices contains more than one '-', return empty slice",
			devices: "1-2-3", containerID: containerID, expected: []int{}},
		{name: "devices is empty string, return empty slice",
			devices: "", containerID: containerID, expected: []int{}},
		{name: "devices max id cannot be converted to an integer, return empty slice",
			devices: "1-a", containerID: containerID, expected: []int{}},
		{name: "devices min id cannot be converted to an integer, return empty slice",
			devices: "a-2", containerID: containerID, expected: []int{}},
		{name: "devices max id less than min id, return empty slice",
			devices: "2-1", containerID: containerID, expected: []int{}},
		{name: "devices min id or max id is invalid, return empty slice",
			devices: "1-32768", containerID: containerID, expected: []int{}},
	}

	for _, test := range tests {
		convey.Convey(test.name, t, func() {
			actual := getDeviceIDsByMinusStyle(test.devices, test.containerID)
			convey.So(actual, convey.ShouldResemble, test.expected)
		})
	}
}

func TestGetDeviceIDsByCommaMinusStyle(t *testing.T) {
	tests := []testCase{
		{name: "devices only contains '-', return deviceId slice",
			devices: "0-1", containerID: containerID, expected: []int{0, 1}},
		{name: "devices only contains ',', return deviceId slice",
			devices: "0,1", containerID: containerID, expected: []int{0, 1}},
	}

	for _, test := range tests {
		convey.Convey(test.name, t, func() {
			actual := getDeviceIDsByCommaMinusStyle(test.devices, test.containerID)
			convey.So(actual, convey.ShouldResemble, test.expected)
		})
	}
}
