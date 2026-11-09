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

package npu_nic

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/agiledragon/gomonkey/v2"

	"ascend-common/api"
	"ascend-common/devmanager"
	"ascend-common/devmanager/common"
)

type fakeDcManager struct {
	devmanager.DeviceManagerMock
	logicIDs []int32
	topo     map[string]common.TopoType
	topoErr  error
	devType  string
}

func (f *fakeDcManager) GetDevType() string {
	if f.devType == "" {
		return api.Ascend910A5
	}
	return f.devType
}

func (f *fakeDcManager) GetDeviceList() (int32, []int32, error) {
	return int32(len(f.logicIDs)), f.logicIDs, nil
}

func (f *fakeDcManager) GetTopoInfoByDeviceIDAndNicName(logicID int32,
	nicName string) (common.TopoType, error) {
	if f.topoErr != nil {
		return common.TopoTypeMax, f.topoErr
	}
	topoType, ok := f.topo[fmt.Sprintf("%d-%s", logicID, nicName)]
	if !ok {
		return common.TopoTypeSys, nil
	}
	return topoType, nil
}

// fakeDirEntry is a minimal os.DirEntry used to mock the sysfs directory.
type fakeDirEntry struct {
	name string
}

func (e fakeDirEntry) Name() string { return e.name }

func (e fakeDirEntry) IsDir() bool { return true }

func (e fakeDirEntry) Type() os.FileMode { return os.ModeDir }

func (e fakeDirEntry) Info() (os.FileInfo, error) { return nil, nil }

// prepareInfinibandDir mocks os.ReadDir so that common.InfinibandClassPath resolves to a fake sysfs
// tree, the key is the rdma device name and the value is its ethernet interface name, an empty
// ethernet interface name creates a device without the device/net directory.
func prepareInfinibandDir(t *testing.T, devices map[string]string) {
	t.Helper()
	root := make([]os.DirEntry, 0, len(devices))
	entries := make(map[string][]os.DirEntry, len(devices)+1)
	for ibName, ethName := range devices {
		root = append(root, fakeDirEntry{name: ibName})
		if ethName == "" {
			continue
		}
		netDir := filepath.Join(common.InfinibandClassPath, ibName, common.InfinibandDeviceDirName,
			common.InfinibandNetDirName)
		entries[netDir] = []os.DirEntry{fakeDirEntry{name: ethName}}
	}
	entries[common.InfinibandClassPath] = root

	patch := gomonkey.ApplyFunc(os.ReadDir, func(name string) ([]os.DirEntry, error) {
		dirEntries, ok := entries[name]
		if !ok {
			return nil, &os.PathError{Op: "open", Path: name, Err: os.ErrNotExist}
		}
		return dirEntries, nil
	})
	t.Cleanup(patch.Reset)
}

func TestListNics(t *testing.T) {
	prepareInfinibandDir(t, map[string]string{
		"hrn0_10": "eth2",
		"hrn0_2":  "eth10",
		"hrn1_0":  "eth0",
	})
	nics, err := ListNics()
	if err != nil {
		t.Fatalf("ListNics failed: %v", err)
	}
	expect := []string{"eth0", "eth10", "eth2"}
	if len(nics) != len(expect) {
		t.Fatalf("expect %v, got %v", expect, nics)
	}
	for i := range expect {
		if nics[i] != expect[i] {
			t.Fatalf("expect %v, got %v", expect, nics)
		}
	}
}

// TestListNicsWithoutNameValidation verifies that the rdma device name is not validated.
func TestListNicsWithoutNameValidation(t *testing.T) {
	prepareInfinibandDir(t, map[string]string{"rxe0": "rxe0net"})
	nics, err := ListNics()
	if err != nil {
		t.Fatalf("ListNics failed: %v", err)
	}
	if len(nics) != 1 || nics[0] != "rxe0net" {
		t.Fatalf("expect [rxe0net], got %v", nics)
	}
}

func TestListNicsWithoutEthName(t *testing.T) {
	prepareInfinibandDir(t, map[string]string{"hrn0_0": "", "rxe0": ""})
	if _, err := ListNics(); err == nil {
		t.Fatal("expect error when the nic has no ethernet interface")
	}
}

func TestQuery(t *testing.T) {
	prepareInfinibandDir(t, map[string]string{"hrn0_0": "eth0", "hrn0_1": "eth1"})
	dcMgr := &fakeDcManager{
		logicIDs: []int32{0, 1},
		topo: map[string]common.TopoType{
			"0-hrn0_0": common.TopoTypeUB,
			"0-hrn0_1": common.TopoTypeHccs,
			"1-hrn0_0": common.TopoTypeUB,
			"1-hrn0_1": common.TopoTypeUB,
		},
	}
	affinity, err := Query(dcMgr)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	if len(affinity[0]) != 1 || affinity[0][0] != "eth0" {
		t.Fatalf("unexpected affinity of logicID 0: %v", affinity[0])
	}
	// The primary nic of the second npu is the next nic in the shared order.
	if len(affinity[1]) != 2 || affinity[1][0] != "eth1" || affinity[1][1] != "eth0" {
		t.Fatalf("unexpected affinity of logicID 1: %v", affinity[1])
	}
}

func TestQueryNotSupported(t *testing.T) {
	prepareInfinibandDir(t, map[string]string{"hrn0_0": "eth0"})
	dcMgr := &fakeDcManager{
		logicIDs: []int32{0},
		topoErr:  fmt.Errorf("not supported, error code: %s", common.NotSupportErrorCode),
	}
	if _, err := Query(dcMgr); err == nil {
		t.Fatal("expect error when the driver does not support the topo query")
	}
}

func TestQueryWithoutAffinity(t *testing.T) {
	prepareInfinibandDir(t, map[string]string{"hrn0_0": "eth0"})
	dcMgr := &fakeDcManager{
		logicIDs: []int32{0},
		topo:     map[string]common.TopoType{"0-hrn0_0": common.TopoTypePix},
	}
	if _, err := Query(dcMgr); err == nil {
		t.Fatal("expect error when no ub affinity is found")
	}
	if _, err := Query(nil); err == nil {
		t.Fatal("expect error when device manager is nil")
	}
}

func TestMatchPrimaryNics(t *testing.T) {
	affinity := map[int32][]string{
		0: {"hrn0_0"},
		1: {"hrn0_0", "hrn1_0"},
		2: {"hrn0_0"},
	}
	primary := matchPrimaryNics(affinity)
	expected := map[int32]string{0: "hrn0_0", 1: "hrn1_0", 2: "hrn0_0"}
	for npuID, nicName := range expected {
		if primary[npuID] != nicName {
			t.Fatalf("npu %d: expect primary %s, got %s", npuID, nicName, primary[npuID])
		}
	}
}

func TestMatchPrimaryNicsWrapAround(t *testing.T) {
	// The nic list is hrn0_0, hrn1_0 in the shared order.
	affinity := map[int32][]string{
		0: {"hrn1_0"},
		1: {"hrn0_0"},
	}
	primary := matchPrimaryNics(affinity)
	if primary[0] != "hrn1_0" {
		t.Fatalf("expect npu 0 matches hrn1_0, got %s", primary[0])
	}
	// The cursor is after hrn1_0, so the scan wraps around and reaches hrn0_0.
	if primary[1] != "hrn0_0" {
		t.Fatalf("expect npu 1 matches hrn0_0, got %s", primary[1])
	}
}

func TestMatchPrimaryNicsWithoutAffinity(t *testing.T) {
	if primary := matchPrimaryNics(nil); len(primary) != 0 {
		t.Fatalf("expect empty result for nil affinity, got %v", primary)
	}
	if primary := matchPrimaryNics(map[int32][]string{0: nil}); len(primary) != 0 {
		t.Fatalf("expect empty result when no nic is connected, got %v", primary)
	}
}

func TestSortWithPrimaryNics(t *testing.T) {
	affinity := map[int32][]string{
		0: {"hrn0_1", "hrn0_0"},
		1: {"hrn1_0"},
	}
	ordered := sortWithPrimaryNics(affinity)
	if len(ordered[0]) != 2 || ordered[0][0] != "hrn0_0" {
		t.Fatalf("expect primary hrn0_0 in front, got %v", ordered[0])
	}
	if len(ordered[1]) != 1 || ordered[1][0] != "hrn1_0" {
		t.Fatalf("expect primary hrn1_0 in front, got %v", ordered[1])
	}
}
