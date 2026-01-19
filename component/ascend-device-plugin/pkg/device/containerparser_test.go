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
	"errors"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/containerd/containerd/oci"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/common-utils/utils"
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

// parseDIffEnvFmtTestCase parseDiffEnvFmt test case
type parseDIffEnvFmtTestCase struct {
	Name        string
	devices     string
	containerID string
	WantDevices []int
}

func buildParseIdffEnvFmtTestCases() []parseDIffEnvFmtTestCase {
	return []parseDIffEnvFmtTestCase{
		{
			Name:        "01-parseDiffEnvFmt parse ascend style devices success",
			devices:     "Ascend-0,Ascend-1,Ascend,Ascend-1.1, Ascend-2",
			containerID: "mock container id",
			WantDevices: []int{0, 1, 2},
		},
		{
			Name:        "02-parseDiffEnvFmt parse comma minux style devices success",
			devices:     "0-4,5-5.6,6.2-7,9-8,10-12",
			containerID: "mock container id",
			WantDevices: []int{0, 1, 2, 3, 4, 10, 11, 12},
		},
		{
			Name:        "03-parseDiffEnvFmt parse only comma style devices success",
			devices:     "0,1.1,2,",
			containerID: "mock container id",
			WantDevices: []int{0, 2},
		},
		{
			Name:        "04-parseDiffEnvFmt parse only minux style devices success",
			devices:     "0-3",
			containerID: "mock container id",
			WantDevices: []int{0, 1, 2, 3},
		},
	}
}

// TestParseDiffEnvFmt for test parseDiffEnvFmt
func TestParseDiffEnvFmt(t *testing.T) {
	testCases := buildParseIdffEnvFmtTestCases()
	for _, tt := range testCases {
		t.Run(tt.Name, func(t *testing.T) {
			if devices := parseDiffEnvFmt(tt.devices, tt.containerID); !reflect.DeepEqual(devices, tt.WantDevices) {
				t.Errorf("parseDiffEnvFmt() devices = %v, wantDevices = %v", devices, tt.WantDevices)
			}
		})
	}
}

// filterNPUDevicesTestCase filterNPUDevices test case
type filterNPUDevicesTestCase struct {
	Name        string
	spec        *oci.Spec
	WantDevices []int
	WantErr     error
}

func buildFilterNPUDevicesTestCases() []filterNPUDevicesTestCase {
	var (
		minor = int64(1)
		major = int64(3)
	)
	return []filterNPUDevicesTestCase{
		{
			Name:        "01-filterNPUDevices spec linux is nil, should return nil and error",
			spec:        &oci.Spec{},
			WantDevices: nil,
			WantErr:     errors.New("empty spec info"),
		},
		{
			Name: "02-filterNPUDevices filter success, should return slice and nil ",
			spec: &oci.Spec{
				Linux: &specs.Linux{
					Resources: &specs.LinuxResources{
						Devices: []specs.LinuxDeviceCgroup{
							{Minor: nil},
							{Minor: &minor, Major: &major, Type: ""},
							{Minor: &minor, Major: &major, Type: charDevice},
						},
					},
				},
			},
			WantDevices: []int{1},
			WantErr:     nil,
		},
	}
}

// TestFilterNPUDevices for test filterNPUDevices
func TestFilterNPUDevices(t *testing.T) {
	majorIDs := []string{"3"}
	patch := gomonkey.ApplyFuncReturn(npuMajor, majorIDs)
	defer patch.Reset()
	testCases := buildFilterNPUDevicesTestCases()
	for _, tt := range testCases {
		t.Run(tt.Name, func(t *testing.T) {
			devices, err := filterNPUDevices(tt.spec)
			if !reflect.DeepEqual(devices, tt.WantDevices) {
				t.Errorf("filterNPUDevices() devices = %v, wantDevices = %v", devices, tt.WantDevices)
			}
			if !reflect.DeepEqual(devices, tt.WantDevices) {
				t.Errorf("filterNPUDevices() err = %v, wantErr = %v", err, tt.WantErr)
			}
		})
	}
}

// getContainerValidSpecTestCase getContainerValidSpec test case
type getContainerValidSpecTestCase struct {
	Name     string
	mockSpec *oci.Spec
	mockErr  error
	WantSpec *oci.Spec
	WantErr  error
}

func buildGetContainerValidSpecTestCases() []getContainerValidSpecTestCase {
	return []getContainerValidSpecTestCase{
		{
			Name:     "01-getContainerValidSpec Spec return error, should return nil and error",
			mockSpec: nil,
			mockErr:  errors.New("get spec failed"),
			WantSpec: nil,
			WantErr:  errors.New("get spec failed"),
		},
		{
			Name:     "02-getContainerValidSpec spec Linux is nil, should return nil and error",
			mockSpec: &oci.Spec{},
			mockErr:  nil,
			WantSpec: nil,
			WantErr:  fmt.Errorf("devices in container is too much (%v)", maxDevicesNum),
		},
		{
			Name:     "03-getContainerValidSpec spec process is nil, should return nil and error",
			mockSpec: &oci.Spec{Linux: &specs.Linux{Resources: &specs.LinuxResources{}}},
			mockErr:  nil,
			WantSpec: nil,
			WantErr:  fmt.Errorf("env in container is too much (%v)", maxEnvNum),
		},
		{
			Name:     "04-getContainerValidSpec get valid spec success, should return wantSpec and nil",
			mockSpec: &oci.Spec{Linux: &specs.Linux{Resources: &specs.LinuxResources{}}, Process: &specs.Process{}},
			mockErr:  nil,
			WantSpec: &oci.Spec{Linux: &specs.Linux{Resources: &specs.LinuxResources{}}, Process: &specs.Process{}},
			WantErr:  nil,
		},
	}
}

// TestGetContainerValidSpec for test getContainerValidSpec
func TestGetContainerValidSpec(t *testing.T) {
	testCases := buildGetContainerValidSpecTestCases()
	for _, tt := range testCases {
		t.Run(tt.Name, func(t *testing.T) {
			patch := gomonkey.ApplyFuncReturn((*MockContainer).Spec, tt.mockSpec, tt.mockErr)
			defer patch.Reset()
			spec, err := getContainerValidSpec(MockContainer{}, nil)
			if !reflect.DeepEqual(spec, tt.WantSpec) {
				t.Errorf("getContainerValidSpec() spec = %v, wantSpec = %v", spec, tt.WantSpec)
			}
			if !reflect.DeepEqual(err, tt.WantErr) {
				t.Errorf("getContainerValidSpec() err = %v, wantErr = %v", err, tt.WantErr)
			}
		})
	}
}

// TestGetNPUMajorID for test getNPUMajorID
func TestGetNPUMajorID(t *testing.T) {
	const mode644 = 0644
	tmpDir := t.TempDir()
	tmpFilePath := tmpDir + "devices"
	data := "1 devdrv-cdev\n1234 devdrv-cdev\n123 vdevdrv-cdev\n"
	if err := os.WriteFile(tmpFilePath, []byte(data), mode644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	convey.Convey("test getNpuMajorID", t, func() {
		patch := gomonkey.ApplyFuncReturn(utils.CheckPath, tmpFilePath, nil)
		defer patch.Reset()
		id, err := getNPUMajorID()
		convey.So(err, convey.ShouldBeNil)
		convey.So(id, convey.ShouldResemble, []string{"1", "123"})
	})
}
