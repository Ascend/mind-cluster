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

// Package cdi — devnode_shared_test.go
//
// Tests for GenerateSharedNodes and its devType/productType variants.
package cdi

import (
	"os"
	"strings"
	"testing"

	"golang.org/x/sys/unix"
	cdispec "tags.cncf.io/container-device-interface/specs-go"
)

// ascend310BSpecialPaths mirrors ascend310BManagerDevices with /dev/ prefix.
var ascend310BSpecialPaths = []string{
	"/dev/svm0", "/dev/ts_aisle", "/dev/upgrade", "/dev/sys",
	"/dev/vdec", "/dev/vpc", "/dev/pngd", "/dev/venc",
	"/dev/log_drv", "/dev/acodec", "/dev/ai", "/dev/ao",
	"/dev/vo", "/dev/hdmi",
}

func mockReadDirWithUB(dir string) ([]os.DirEntry, error) {
	if strings.Contains(dir, "uburma") {
		return []os.DirEntry{
			&mockDirEntry{name: "ub0"},
			&mockDirEntry{name: "ub1"},
		}, nil
	}
	return nil, nil
}

func checkNo310BSpecials(t *testing.T, nodes []*cdispec.DeviceNode) {
	t.Helper()
	for _, s := range ascend310BSpecialPaths {
		if err := assertNoDevice(nodes, s); err != nil {
			t.Error(err)
		}
	}
}

func assert310BSpecials(t *testing.T, nodes []*cdispec.DeviceNode) {
	t.Helper()
	assertDevicesContain(t, nodes, ascend310BSpecialPaths...)
}

func TestGenerateSharedNodes_EmptyDevType(t *testing.T) {
	_, err := GenerateSharedNodes("", "")
	if err == nil {
		t.Fatal("expected error for empty devType")
	}
}

func TestGenerateSharedNodes_Ascend910(t *testing.T) {
	cleanup := setupMocks()
	defer cleanup()

	nodes, err := GenerateSharedNodes(Ascend910, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertDevicesContain(t, nodes,
		"/dev/davinci_manager",
		"/dev/dvpp_cmdlist",
		"/dev/devmm_svm",
		"/dev/hisi_hdc",
	)
	checkNo310BSpecials(t, nodes)
	if len(nodes) != 4 {
		t.Errorf("got %d nodes, want 4", len(nodes))
	}
}

func TestGenerateSharedNodes_Ascend310B(t *testing.T) {
	cleanup := setupMocks()
	defer cleanup()

	nodes, err := GenerateSharedNodes(Ascend310B, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertDevicesContain(t, nodes,
		"/dev/davinci_manager",
		"/dev/dvpp_cmdlist",
	)
	assert310BSpecials(t, nodes)
	if err := assertNoDevice(nodes, "/dev/devmm_svm"); err != nil {
		t.Error(err)
	}
	if err := assertNoDevice(nodes, "/dev/hisi_hdc"); err != nil {
		t.Error(err)
	}
	if len(nodes) != 16 {
		t.Errorf("got %d nodes, want 16", len(nodes))
	}
}

func TestGenerateSharedNodes_Ascend910A5(t *testing.T) {
	cleanup := setupMocks()
	defer cleanup()

	nodes, err := GenerateSharedNodes(Ascend910A5, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertDevicesContain(t, nodes,
		"/dev/davinci_manager",
		"/dev/hisi_hdc",
	)
	if err := assertNoDevice(nodes, "/dev/dvpp_cmdlist"); err != nil {
		t.Error(err)
	}
	if err := assertNoDevice(nodes, "/dev/devmm_svm"); err != nil {
		t.Error(err)
	}
	if len(nodes) != 2 {
		t.Errorf("got %d nodes, want 2", len(nodes))
	}
}

func TestGenerateSharedNodes_Atlas200(t *testing.T) {
	for _, pt := range []string{Atlas200ISoc, Atlas200} {
		t.Run(pt, func(t *testing.T) {
			cleanup := setupMocks()
			defer cleanup()

			nodes, err := GenerateSharedNodes(Ascend910, pt)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			assertDevicesContain(t, nodes,
				"/dev/davinci_manager",
				"/dev/dvpp_cmdlist",
			)
			if err := assertNoDevice(nodes, "/dev/devmm_svm"); err != nil {
				t.Error(err)
			}
			if err := assertNoDevice(nodes, "/dev/hisi_hdc"); err != nil {
				t.Error(err)
			}
			if len(nodes) != 2 {
				t.Errorf("got %d nodes, want 2", len(nodes))
			}
		})
	}
}

func TestGenerateSharedNodes_UBDevicesPresent(t *testing.T) {
	cleanup := setupMocks()
	defer cleanup()

	origReadDir := osReadDir
	osReadDir = mockReadDirWithUB
	defer func() { osReadDir = origReadDir }()

	nodes, err := GenerateSharedNodes(Ascend910, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertDevicesContain(t, nodes,
		"/dev/davinci_manager",
		"/dev/dvpp_cmdlist",
		"/dev/devmm_svm",
		"/dev/hisi_hdc",
		"/dev/uburma/ub0",
		"/dev/uburma/ub1",
	)
	if len(nodes) != 6 {
		t.Errorf("got %d nodes, want 6", len(nodes))
	}
}

func TestGenerateSharedNodes_ManagerError(t *testing.T) {
	cleanup := setupMocks()
	defer cleanup()

	origStat := unixStat
	unixStat = func(path string, st *unix.Stat_t) error {
		if strings.Contains(path, "davinci_manager") && !strings.Contains(path, "docker") {
			return os.ErrNotExist
		}
		return mockStatFunc(path, st)
	}
	defer func() { unixStat = origStat }()

	_, err := GenerateSharedNodes(Ascend910, "")
	if err == nil {
		t.Fatal("expected error from resolveDavinciManager via GenerateSharedNodes")
	}
}
