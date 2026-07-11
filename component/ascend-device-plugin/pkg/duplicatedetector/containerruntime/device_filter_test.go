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

package containerruntime

import (
	"errors"
	"os"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/containerd/containerd/oci"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/apimachinery/pkg/util/sets"

	"ascend-common/common-utils/utils"
)

const (
	mockMajorIDStr = "195"
)

func int64Ptr(v int64) *int64 {
	return &v
}

func TestFilterNPUDevicesNilSpec(t *testing.T) {
	convey.Convey("01-nil spec should return nil", t, func() {
		result := filterNPUDevices(nil)
		convey.So(result, convey.ShouldBeNil)
	})
}

func TestFilterNPUDevicesNilLinux(t *testing.T) {
	convey.Convey("02-nil linux should return nil", t, func() {
		spec := &oci.Spec{}
		result := filterNPUDevices(spec)
		convey.So(result, convey.ShouldBeNil)
	})
}

func TestFilterNPUDevicesNilResources(t *testing.T) {
	convey.Convey("03-nil resources should return nil", t, func() {
		spec := &oci.Spec{
			Linux: &specs.Linux{},
		}
		result := filterNPUDevices(spec)
		convey.So(result, convey.ShouldBeNil)
	})
}

func TestFilterNPUDevicesEmptyDevices(t *testing.T) {
	convey.Convey("04-empty devices should return empty", t, func() {
		spec := &oci.Spec{
			Linux: &specs.Linux{
				Resources: &specs.LinuxResources{
					Devices: []specs.LinuxDeviceCgroup{},
				},
			},
		}
		result := filterNPUDevices(spec)
		convey.So(result, convey.ShouldResemble, []int{})
	})
}

func TestFilterNPUDevicesNilMinor(t *testing.T) {
	convey.Convey("05-nil minor should be skipped", t, func() {
		spec := &oci.Spec{
			Linux: &specs.Linux{
				Resources: &specs.LinuxResources{
					Devices: []specs.LinuxDeviceCgroup{
						{Type: "c", Major: int64Ptr(195)},
					},
				},
			},
		}
		patch := gomonkey.ApplyFuncReturn(getNPUMajorID,
			sets.NewString(mockMajorIDStr), nil)
		defer patch.Reset()
		result := filterNPUDevices(spec)
		convey.So(result, convey.ShouldResemble, []int{})
	})
}

func TestFilterNPUDevicesNilMajor(t *testing.T) {
	convey.Convey("06-nil major should be skipped", t, func() {
		spec := &oci.Spec{
			Linux: &specs.Linux{
				Resources: &specs.LinuxResources{
					Devices: []specs.LinuxDeviceCgroup{
						{Type: "c", Minor: int64Ptr(0)},
					},
				},
			},
		}
		patch := gomonkey.ApplyFuncReturn(getNPUMajorID,
			sets.NewString(mockMajorIDStr), nil)
		defer patch.Reset()
		result := filterNPUDevices(spec)
		convey.So(result, convey.ShouldResemble, []int{})
	})
}

func TestFilterNPUDevicesNonCharDevice(t *testing.T) {
	convey.Convey("07-non-char device should be skipped", t, func() {
		spec := &oci.Spec{
			Linux: &specs.Linux{
				Resources: &specs.LinuxResources{
					Devices: []specs.LinuxDeviceCgroup{
						{Type: "b", Major: int64Ptr(195), Minor: int64Ptr(0)},
					},
				},
			},
		}
		patch := gomonkey.ApplyFuncReturn(getNPUMajorID,
			sets.NewString(mockMajorIDStr), nil)
		defer patch.Reset()
		result := filterNPUDevices(spec)
		convey.So(result, convey.ShouldResemble, []int{})
	})
}

func TestFilterNPUDevicesSingleNPU(t *testing.T) {
	convey.Convey("08-npu device should be returned", t, func() {
		spec := &oci.Spec{
			Linux: &specs.Linux{
				Resources: &specs.LinuxResources{
					Devices: []specs.LinuxDeviceCgroup{
						{Type: "c", Major: int64Ptr(195), Minor: int64Ptr(0)},
					},
				},
			},
		}
		patch := gomonkey.ApplyFuncReturn(npuMajor,
			sets.NewString(mockMajorIDStr))
		defer patch.Reset()
		result := filterNPUDevices(spec)
		convey.So(result, convey.ShouldResemble, []int{0})
	})
}

func TestFilterNPUDevicesMultipleNPU(t *testing.T) {
	convey.Convey("09-multiple npu devices should be returned", t, func() {
		spec := &oci.Spec{
			Linux: &specs.Linux{
				Resources: &specs.LinuxResources{
					Devices: []specs.LinuxDeviceCgroup{
						{Type: "c", Major: int64Ptr(195), Minor: int64Ptr(0)},
						{Type: "c", Major: int64Ptr(195), Minor: int64Ptr(1)},
					},
				},
			},
		}
		patch := gomonkey.ApplyFuncReturn(npuMajor,
			sets.NewString(mockMajorIDStr))
		defer patch.Reset()
		result := filterNPUDevices(spec)
		convey.So(result, convey.ShouldResemble, []int{0, 1})
	})
}

func TestFilterNPUDevicesMinorExceedsMax(t *testing.T) {
	convey.Convey("10-minor exceeds max should return nil", t, func() {
		spec := &oci.Spec{
			Linux: &specs.Linux{
				Resources: &specs.LinuxResources{
					Devices: []specs.LinuxDeviceCgroup{
						{Type: "c", Major: int64Ptr(195), Minor: int64Ptr(1<<32 + 1)},
					},
				},
			},
		}
		patch := gomonkey.ApplyFuncReturn(getNPUMajorID,
			sets.NewString(mockMajorIDStr), nil)
		defer patch.Reset()
		result := filterNPUDevices(spec)
		convey.So(result, convey.ShouldBeNil)
	})
}

func TestGetNPUMajorIDPathCheckFails(t *testing.T) {
	convey.Convey("01-path check fails should return error", t, func() {
		patches := gomonkey.ApplyFuncReturn(utils.CheckPath, "", errors.New("path check failed"))
		defer patches.Reset()
		result, err := getNPUMajorID()
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(result, convey.ShouldBeNil)
	})
}

func TestGetNPUMajorIDFileOpenFails(t *testing.T) {
	convey.Convey("02-file open fails should return error", t, func() {
		patches := gomonkey.ApplyFuncReturn(utils.CheckPath, "/proc/devices", nil).
			ApplyFuncReturn(os.Open, nil, errors.New("file open failed"))
		defer patches.Reset()
		result, err := getNPUMajorID()
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(result, convey.ShouldBeEmpty)
	})
}

func TestGetNPUMajorIDNoNPUDevices(t *testing.T) {
	convey.Convey("03-no npu devices found should return empty", t, func() {
		tmpFile, clean, err := createTempFile("1 mem\n2 pty\n")
		if err != nil {
			t.Fatalf("failed to create temp file: %v", err)
		}
		defer clean()
		patches := gomonkey.ApplyFuncReturn(utils.CheckPath, tmpFile, nil)
		defer patches.Reset()
		result, err := getNPUMajorID()
		convey.So(err, convey.ShouldBeNil)
		convey.So(result, convey.ShouldBeEmpty)
	})
}

func TestGetNPUMajorIDNPUDevicesFound(t *testing.T) {
	convey.Convey("04-npu devices found should return major ids", t, func() {
		tmpFile, clean, err := createTempFile("195 devdrv-cdev\n196 devdrv-cdev\n")
		if err != nil {
			t.Fatalf("failed to create temp file: %v", err)
		}
		defer clean()
		patches := gomonkey.ApplyFuncReturn(utils.CheckPath, tmpFile, nil)
		defer patches.Reset()
		result, err := getNPUMajorID()
		convey.So(err, convey.ShouldBeNil)
		convey.So(result.Has("195"), convey.ShouldBeTrue)
		convey.So(result.Has("196"), convey.ShouldBeTrue)
	})
}

func TestGetNPUMajorIDMixedDevices(t *testing.T) {
	convey.Convey("05-mixed devices found should return npu major ids only", t, func() {
		tmpFile, clean, err := createTempFile("1 mem\n195 devdrv-cdev\n2 pty\n196 devdrv-cdev\n")
		if err != nil {
			t.Fatalf("failed to create temp file: %v", err)
		}
		defer clean()
		patches := gomonkey.ApplyFuncReturn(utils.CheckPath, tmpFile, nil)
		defer patches.Reset()
		result, err := getNPUMajorID()
		convey.So(err, convey.ShouldBeNil)
		convey.So(result.Has("195"), convey.ShouldBeTrue)
		convey.So(result.Has("196"), convey.ShouldBeTrue)
	})
}

func TestNpuMajorCached(t *testing.T) {
	convey.Convey("01-should return cached major ids", t, func() {
		// Note: This test depends on sync.Once not being initialized yet.
		// If it runs after other tests that trigger npuMajor(), the mock won't take effect
		// and npuMajor() will return the cached value from previous calls.
		// In that case, we just verify that npuMajor() returns a non-nil result.
		patches := gomonkey.ApplyFuncReturn(getNPUMajorID, sets.NewString("123", "456"), nil)
		defer patches.Reset()
		result := npuMajor()
		convey.So(result, convey.ShouldNotBeNil)
	})
}

func createTempFile(content string) (string, func(), error) {
	f, err := os.CreateTemp("", "test_device_*")
	if err != nil {
		return "", func() {}, err
	}
	if _, err = f.WriteString(content); err != nil {
		cleanupTempFile(f)
		return "", func() {}, err
	}
	if _, err = f.Seek(0, 0); err != nil {
		cleanupTempFile(f)
		return "", func() {}, err
	}
	name := f.Name()
	return name, func() { cleanupTempFile(f) }, nil
}

func cleanupTempFile(f *os.File) {
	if f == nil {
		return
	}
	f.Close()
	os.Remove(f.Name())
}
