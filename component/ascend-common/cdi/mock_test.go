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

package cdi

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"golang.org/x/sys/unix"

	"ascend-common/common-utils/hwlog"
	"ascend-common/cdi/mount"

	cdispec "tags.cncf.io/container-device-interface/specs-go"
)

func init() {
	hwlog.InitRunLogger(&hwlog.LogConfig{OnlyToStdout: true}, context.Background())
}

// Ascend NPU device major numbers — mock reference values for device generation tests.
// These mirror the real Ascend kernel driver major numbers and are used in mockStatFunc
// and in test assertions (edits_test.go, inject_test.go).
const (
	devMajorDavinci     = 234 // /dev/davinci*
	devMajorManager     = 235 // /dev/davinci_manager*
	devMajorDvppCmdlist = 236 // /dev/dvpp_cmdlist
	devMajorDevmmSvm    = 237 // /dev/devmm_svm
	devMajorHisiHdc     = 238 // /dev/hisi_hdc
	devMajorSvm0        = 239 // /dev/svm0 (Ascend310B)
	devMajorTsAisle     = 240 // /dev/ts_aisle (Ascend310B)
	devMajorUpgrade     = 241 // /dev/upgrade (Ascend310B)
	devMajorSys         = 242 // /dev/sys (Ascend310B)
	devMajorVdec        = 243 // /dev/vdec (Ascend310B)
	devMajorVpc         = 244 // /dev/vpc (Ascend310B)
	devMajorPngd        = 245 // /dev/pngd (Ascend310B)
	devMajorVenc        = 246 // /dev/venc (Ascend310B)
	devMajorLogDrv      = 247 // /dev/log_drv (Ascend310B)
	devMajorAcodec      = 248 // /dev/acodec (Ascend310B)
	devMajorAi          = 249 // /dev/ai (Ascend310B)
	devMajorAo          = 250 // /dev/ao (Ascend310B)
	devMajorVo          = 251 // /dev/vo (Ascend310B)
	devMajorHdmi        = 252 // /dev/hdmi (Ascend310B)
	devMajorUburma      = 253 // /dev/uburma/*
	devMajorUmmu        = 254 // /dev/ummu/*
)

// deviceNodePerm is the filesystem permission mask for mocked device nodes.
const deviceNodePerm = 0660

// ---------------------------------------------------------------------------
// Shared test types
// ---------------------------------------------------------------------------

// mockFileInfo implements os.FileInfo for tests.
type mockFileInfo struct {
	name string
	mode os.FileMode
}

func (m *mockFileInfo) Name() string       { return m.name }
func (m *mockFileInfo) Size() int64        { return 0 }
func (m *mockFileInfo) Mode() os.FileMode  { return m.mode }
func (m *mockFileInfo) ModTime() time.Time { return time.Time{} }
func (m *mockFileInfo) IsDir() bool        { return false }
func (m *mockFileInfo) Sys() interface{}   { return nil }

// mockDirEntry implements os.DirEntry for tests.
type mockDirEntry struct {
	name  string
	isDir bool
}

func (m *mockDirEntry) Name() string               { return m.name }
func (m *mockDirEntry) IsDir() bool                { return m.isDir }
func (m *mockDirEntry) Type() os.FileMode          { return 0 }
func (m *mockDirEntry) Info() (os.FileInfo, error) { return &mockFileInfo{name: m.name}, nil }

// mockMountProvider implements mount.Provider for tests.
type mockMountProvider struct {
	mounts []*cdispec.Mount
	err    error
}

func (m *mockMountProvider) GetMounts() ([]*cdispec.Mount, error) {
	return m.mounts, m.err
}

// ---------------------------------------------------------------------------
// Mock function implementations (portable — no unix imports)
// ---------------------------------------------------------------------------

// mockOSStatFunc returns a non-nil FileInfo for any /dev/ path.
func mockOSStatFunc(path string) (os.FileInfo, error) {
	if !strings.HasPrefix(path, "/dev/") {
		return nil, os.ErrNotExist
	}
	return &mockFileInfo{name: filepath.Base(path), mode: os.ModeDevice | os.ModeCharDevice}, nil
}

// mockReadDirEmpty returns no entries (simulates no UB devices present).
func mockReadDirEmpty(_ string) ([]os.DirEntry, error) {
	return nil, nil
}

// mockStatFunc creates a unix.Stat-compatible function that assigns well-known
// Ascend major/minor numbers based on the device path.
func mockStatFunc(path string, st *unix.Stat_t) error {
	major, minor := mockDeviceMajorMinor(path)
	if major == 0 {
		return os.ErrNotExist
	}
	st.Rdev = unix.Mkdev(major, minor)
	st.Mode = uint32(os.ModeDevice | os.ModeCharDevice | deviceNodePerm)
	return nil
}

// mockDeviceMajorMinor returns the (major, minor) pair for a mock device path.
func mockDeviceMajorMinor(path string) (uint32, uint32) {
	// Per-device davinci / vdavinci: extract numeric ID from path suffix.
	if (strings.HasPrefix(path, "/dev/davinci") || strings.HasPrefix(path, "/dev/vdavinci")) &&
		!strings.Contains(path, "manager") {
		idStr := path
		if strings.HasPrefix(path, "/dev/vdavinci") {
			idStr = strings.TrimPrefix(path, "/dev/vdavinci")
		} else {
			idStr = strings.TrimPrefix(path, "/dev/davinci")
		}
		var id int
		if n, _ := fmt.Sscanf(idStr, "%d", &id); n == 1 {
			return devMajorDavinci, uint32(id)
		}
		return devMajorDavinci, 0
	}

	// Fixed manager / utility devices: major is predetermined, minor is 0.
	deviceMajors := map[string]uint32{
		"davinci_manager": devMajorManager, "dvpp_cmdlist": devMajorDvppCmdlist,
		"devmm_svm": devMajorDevmmSvm, "hisi_hdc": devMajorHisiHdc,
		"/svm0": devMajorSvm0, "/ts_aisle": devMajorTsAisle, "/upgrade": devMajorUpgrade,
		"/sys": devMajorSys, "/vdec": devMajorVdec, "/vpc": devMajorVpc,
		"/pngd": devMajorPngd, "/venc": devMajorVenc, "/log_drv": devMajorLogDrv,
		"/acodec": devMajorAcodec, "/ai": devMajorAi, "/ao": devMajorAo,
		"/vo": devMajorVo, "/hdmi": devMajorHdmi,
		"uburma": devMajorUburma, "ummu": devMajorUmmu,
	}
	for key, maj := range deviceMajors {
		if strings.Contains(path, key) {
			return maj, 0
		}
	}
	return 0, 0
}

// ---------------------------------------------------------------------------
// Mock setup
// ---------------------------------------------------------------------------

// setupPlatformMocks overrides the package-level function variables
// (unixStat, osStat, osReadDir) for testing. Returns a cleanup
// function that MUST be deferred.
func setupPlatformMocks() func() {
	origStat := unixStat
	origOSStat := osStat
	origReadDir := osReadDir

	unixStat = mockStatFunc
	osStat = mockOSStatFunc
	osReadDir = mockReadDirEmpty

	return func() {
		unixStat = origStat
		osStat = origOSStat
		osReadDir = origReadDir
	}
}

// setupMocks provides a mock environment for tests that require device
// generation. It overrides the package-level function variables
// (unixStat, osStat, osReadDir).
func setupMocks() func() {
	return setupPlatformMocks()
}

// ---------------------------------------------------------------------------
// Assert helpers
// ---------------------------------------------------------------------------

func assertDevicesContain(t *testing.T, devices []*cdispec.DeviceNode, expected ...string) {
	t.Helper()
	devicePaths := make(map[string]bool, len(devices))
	for _, d := range devices {
		devicePaths[d.Path] = true
	}
	for _, path := range expected {
		if !devicePaths[path] {
			t.Errorf("expected device %q not found in generated nodes", path)
		}
	}
}

func assertNoDevice(devices []*cdispec.DeviceNode, path string) error {
	for _, d := range devices {
		if d.Path == path {
			return fmt.Errorf("unexpected device %q found", path)
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// HCCL topology helpers
// ---------------------------------------------------------------------------

func saveRestoreHCCLItems() func() {
	orig := make([]mount.TopologyItem, len(mount.TopologyItems))
	copy(orig, mount.TopologyItems)
	return func() { mount.TopologyItems = orig }
}

// ---------------------------------------------------------------------------
// Shorthand helpers
// ---------------------------------------------------------------------------

// writeClaimSpec calls GenerateClaimSpec with the standard Ascend910 device
// type and a mock provider.  It is a shorthand to reduce test boilerplate.
func writeClaimSpec(t *testing.T, claimUID string, deviceIDs []int, productType string, provider mount.Provider) (string, []string) {
	t.Helper()
	specName, ids, err := GenerateClaimSpec(ClaimSpecConfig{DeviceConfig: DeviceConfig{DeviceIDs: deviceIDs, DevType: "Ascend910", ProductType: productType}, ClaimUID: claimUID, Provider: provider})
	if err != nil {
		t.Fatalf("GenerateClaimSpec: %v", err)
	}
	return specName, ids
}

func newMockProvider(mounts []*cdispec.Mount) *mockMountProvider {
	return &mockMountProvider{mounts: mounts}
}

func newMockProviderWithErr(err error) *mockMountProvider {
	return &mockMountProvider{err: err}
}
