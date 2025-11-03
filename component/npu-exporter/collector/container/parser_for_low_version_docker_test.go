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

// Package container for monitoring containers' npu allocation
package container

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/common-utils/utils"
)

const (
	miHeader = "434 425 0:23 / /sys/fs/cgroup/devices rw,nosuid,nodev,noexec,relatime shared:213" +
		" - cgroup cgroup rw,devices\n"
	ctrlDevices    = "devices"
	cgroupRootPath = "/sys/fs/cgroup/devices"
	sysdSlice      = "kubepods-besteffort-pod123456.slice"
	cid            = "abcdef0123456789"
	devLine0       = "c 236:0 rwm\n"
	devLine1       = "c 236:1 rw\n"
	devWildcard    = "c *:* m\n"
)

type mockFile struct {
	*bytes.Reader
	closeErr error
}

func (m *mockFile) Close() error { return m.closeErr }

func TestScanForAscendDevices(t *testing.T) {
	convey.Convey("should collect minors when match ascend major ids", t, func() {
		dev, clean, err := mkTemp(devLine0 + devLine1)
		convey.So(err, convey.ShouldBeNil)
		defer clean()
		p := gomonkey.ApplyFuncReturn(npuMajor, []string{"236"}).
			ApplyFuncReturn(utils.CheckPath, dev, nil)
		defer p.Reset()

		nums, ok, err := ScanForAscendDevices(dev, cid)
		convey.So(err, convey.ShouldBeNil)
		convey.So(ok, convey.ShouldBeTrue)
		convey.So(nums, convey.ShouldResemble, []int{0, 1})
	})
	convey.Convey("should stop when wildcard appears", t, func() {
		dev, clean, err := mkTemp(devWildcard)
		convey.So(err, convey.ShouldBeNil)
		defer clean()
		p := gomonkey.ApplyFuncReturn(npuMajor, []string{"236"}).
			ApplyFuncReturn(utils.CheckPath, dev, nil)
		defer p.Reset()

		nums, ok, err := ScanForAscendDevices(dev, cid)
		convey.So(err, convey.ShouldBeNil)
		convey.So(ok, convey.ShouldBeFalse)
		convey.So(nums, convey.ShouldBeEmpty)
	})
}

func TestToCgroupHierarchy(t *testing.T) {
	convey.Convey("should return hierarchy when cgroupfs path provided", t, func() {
		cgroupsPath := "/docker/test-container"
		result, err := toCgroupHierarchy(cgroupsPath)
		convey.So(err, convey.ShouldBeNil)
		convey.So(result, convey.ShouldEqual, cgroupsPath)
	})
	convey.Convey("should return error when systemd cgroup path parsing fails", t, func() {
		cgroupsPath := "systemd:docker:test-container"
		result, err := toCgroupHierarchy(cgroupsPath)
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, "parsing fail")
		convey.So(result, convey.ShouldBeEmpty)
	})
	convey.Convey("should return error when unknown cgroup path type", t, func() {
		cgroupsPath := "unknown:format"
		result, err := toCgroupHierarchy(cgroupsPath)
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, "parsing fail")
		convey.So(result, convey.ShouldBeEmpty)
	})
}

func TestParseSystemdCgroup(t *testing.T) {
	convey.Convey("should return empty when systemd cgroup parsing fails", t, func() {
		cgroup := "systemd:docker:test-container"
		result := parseSystemdCgroup(cgroup)
		convey.So(result, convey.ShouldBeEmpty)
	})
	convey.Convey("should return empty when invalid cgroup format", t, func() {
		cgroup := "invalid:format"
		result := parseSystemdCgroup(cgroup)
		convey.So(result, convey.ShouldBeEmpty)
	})
}

func TestParseSlice(t *testing.T) {
	convey.Convey("should return default slice when empty slice provided", t, func() {
		slice := ""
		result := parseSlice(slice)
		convey.So(result, convey.ShouldEqual, defaultSlice)
	})
	convey.Convey("should return path when valid slice provided", t, func() {
		slice := "docker.slice"
		result := parseSlice(slice)
		convey.So(result, convey.ShouldNotBeEmpty)
		convey.So(result, convey.ShouldContainSubstring, "docker")
	})
	convey.Convey("should return empty when invalid slice format", t, func() {
		slice := "invalid/slice"
		result := parseSlice(slice)
		convey.So(result, convey.ShouldBeEmpty)
	})
}

func TestGetUnit(t *testing.T) {
	convey.Convey("should return name with scope when not slice", t, func() {
		prefix := "docker"
		name := "test-container"
		result := getUnit(prefix, name)
		convey.So(result, convey.ShouldEqual, "docker-test-container.scope")
	})
	convey.Convey("should return name when already slice", t, func() {
		prefix := "docker"
		name := "test.slice"
		result := getUnit(prefix, name)
		convey.So(result, convey.ShouldEqual, "test.slice")
	})
}

func TestIsLowerDockerVersion(t *testing.T) {
	convey.Convey("should return true when version is lower", t, func() {
		version := "1.13.1"
		result := isLowerDockerVersion(version)
		convey.So(result, convey.ShouldBeTrue)
	})
	convey.Convey("should return false when version is higher", t, func() {
		version := "20.10.0"
		result := isLowerDockerVersion(version)
		convey.So(result, convey.ShouldBeFalse)
	})
	convey.Convey("should return false when version is empty", t, func() {
		version := ""
		result := isLowerDockerVersion(version)
		convey.So(result, convey.ShouldBeFalse)
	})
	convey.Convey("should return false when version format is invalid", t, func() {
		version := "invalid.version"
		result := isLowerDockerVersion(version)
		convey.So(result, convey.ShouldBeFalse)
	})
}

func TestGetCgroupControllerPathShouldReadMountinfo(t *testing.T) {
	convey.Convey("should parse mountinfo and return devices controller path", t, func() {
		mi, clean, err := mkTemp(miHeader)
		convey.So(err, convey.ShouldBeNil)
		defer clean()
		p1 := gomonkey.ApplyFuncReturn(utils.CheckPath, mi, nil)
		defer p1.Reset()
		path, err := getCgroupControllerPath(ctrlDevices)
		convey.So(err, convey.ShouldBeNil)
		convey.So(path, convey.ShouldEqual, cgroupRootPath)
	})
}

func TestGetCgroupPathShouldJoinControllerAndHierarchy(t *testing.T) {
	convey.Convey("should join controller path with hierarchy when systemd", t, func() {
		mi, clean, err := mkTemp(miHeader)
		convey.So(err, convey.ShouldBeNil)
		defer clean()
		p1 := gomonkey.ApplyFuncReturn(utils.CheckPath, mi, nil)
		defer p1.Reset()
		cg := sysdSlice + ":docker:" + cid
		got, err := GetCgroupPath(ctrlDevices, cg)
		convey.So(err, convey.ShouldBeNil)
		convey.So(got, convey.ShouldEqual, filepath.Join(cgroupRootPath, parseSystemdCgroup(cg)))
	})
}
