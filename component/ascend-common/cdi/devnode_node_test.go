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

// Package cdi — devnode_node_test.go
//
// Tests for devnode.go low-level helpers:
// resolveDavinciManager, appendDeviceNode,
// addDevicesFromDir, addUBDevicesFromDir,
// and GenerateSharedNodes integration paths.
package cdi

import (
	"errors"
	"os"
	"strings"
	"testing"

	"golang.org/x/sys/unix"

	"ascend-common/api"
)

func TestResolveDavinciManager_310B_DockerVariant_Missing(t *testing.T) {
	cleanup := setupMocks()
	defer cleanup()

	origOSStat := osStat
	osStat = func(path string) (os.FileInfo, error) {
		if strings.Contains(path, "davinci_manager_docker") {
			return nil, os.ErrNotExist
		}
		return mockOSStatFunc(path)
	}
	defer func() { osStat = origOSStat }()

	dev, err := resolveDavinciManager(api.Ascend310B)
	if err != nil {
		t.Fatalf("expected fallback to succeed: %v", err)
	}
	if dev.Path != "/dev/davinci_manager" {
		t.Errorf("Path = %q, want /dev/davinci_manager", dev.Path)
	}
	if dev.HostPath != "/dev/davinci_manager" {
		t.Errorf("HostPath = %q, want /dev/davinci_manager", dev.HostPath)
	}
}

func TestResolveDavinciManager_310B_DockerVariant_Exists(t *testing.T) {
	cleanup := setupMocks()
	defer cleanup()

	// mockOSStatFunc returns a non-nil FileInfo for /dev/ paths, so the docker
	// variant "exists".
	dev, err := resolveDavinciManager(api.Ascend310B)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dev.Path != "/dev/davinci_manager" {
		t.Errorf("Path = %q, want /dev/davinci_manager (container path)", dev.Path)
	}
	if dev.HostPath != "/dev/davinci_manager_docker" {
		t.Errorf("HostPath = %q, want /dev/davinci_manager_docker (host path)", dev.HostPath)
	}
}

func TestGenerateSharedNodes_DvppCmdList_StatFail(t *testing.T) {
	cleanup := setupMocks()
	defer cleanup()

	origStat := unixStat
	unixStat = func(path string, st *unix.Stat_t) error {
		if strings.Contains(path, "dvpp_cmdlist") {
			return os.ErrNotExist
		}
		return mockStatFunc(path, st)
	}
	defer func() { unixStat = origStat }()

	nodes, err := GenerateSharedNodes(api.Ascend910, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(nodes) != 3 {
		t.Errorf("expected 3 nodes after dvpp_cmdlist stat failure, got %d", len(nodes))
	}
	if err := assertNoDevice(nodes, "/dev/dvpp_cmdlist"); err != nil {
		t.Error(err)
	}
}

func TestAddDevicesFromDir_ReadError(t *testing.T) {
	origReadDir := osReadDir
	osReadDir = func(_ string) ([]os.DirEntry, error) {
		return nil, errors.New("read error")
	}
	defer func() { osReadDir = origReadDir }()

	_, err := addDevicesFromDir("/dev/uburma")
	if err == nil {
		t.Fatal("expected error from osReadDir")
	}
}

func TestAddDevicesFromDir_StatFailure(t *testing.T) {
	cleanup := setupMocks()
	defer cleanup()

	origReadDir := osReadDir
	osReadDir = mockReadDirWithUB
	defer func() { osReadDir = origReadDir }()

	origStat := unixStat
	unixStat = func(path string, st *unix.Stat_t) error {
		if strings.Contains(path, "ub1") {
			return os.ErrNotExist
		}
		return mockStatFunc(path, st)
	}
	defer func() { unixStat = origStat }()

	_, err := addDevicesFromDir("/dev/uburma")
	if err == nil {
		t.Fatal("expected error from statDevice failure in addDevicesFromDir")
	}
}

func TestAddUBDevicesFromDir_ReadError(t *testing.T) {
	cleanup := setupMocks()
	defer cleanup()

	origReadDir := osReadDir
	osReadDir = func(dir string) ([]os.DirEntry, error) {
		if strings.Contains(dir, "uburma") {
			return nil, errors.New("read error")
		}
		return nil, nil
	}
	defer func() { osReadDir = origReadDir }()

	nodes, err := GenerateSharedNodes(api.Ascend910, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(nodes) != 4 {
		t.Errorf("got %d nodes, want 4 (UB error non-fatal)", len(nodes))
	}
}

func TestGenerateSharedNodes_310BSpecials_StatFailure(t *testing.T) {
	cleanup := setupMocks()
	defer cleanup()

	origStat := unixStat
	unixStat = func(path string, st *unix.Stat_t) error {
		if strings.Contains(path, "/svm0") {
			return os.ErrNotExist
		}
		return mockStatFunc(path, st)
	}
	defer func() { unixStat = origStat }()

	nodes, err := GenerateSharedNodes(api.Ascend310B, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(nodes) != 15 {
		t.Errorf("got %d nodes, want 15 (one stat failure)", len(nodes))
	}
	if err := assertNoDevice(nodes, "/dev/svm0"); err != nil {
		t.Error(err)
	}
}

func TestGenerateSharedNodes_CommonManagers_StatFailure(t *testing.T) {
	cleanup := setupMocks()
	defer cleanup()

	origStat := unixStat
	unixStat = func(path string, st *unix.Stat_t) error {
		if strings.Contains(path, "hisi_hdc") {
			return os.ErrNotExist
		}
		return mockStatFunc(path, st)
	}
	defer func() { unixStat = origStat }()

	nodes, err := GenerateSharedNodes(api.Ascend910, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(nodes) != 3 {
		t.Errorf("got %d nodes, want 3 (hisi_hdc stat failed)", len(nodes))
	}
	if err := assertNoDevice(nodes, "/dev/hisi_hdc"); err != nil {
		t.Error(err)
	}
	assertDevicesContain(t, nodes, "/dev/devmm_svm")
}
