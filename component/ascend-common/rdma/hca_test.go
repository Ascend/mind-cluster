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

package rdma

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"ascend-common/api"
)

func TestParseRDMADevices(t *testing.T) {
	for _, value := range []string{"", " ", " void "} {
		names, err := ParseRDMADevices(value)
		if err != nil || len(names) != 0 {
			t.Fatalf("disabled %q: %v %v", value, names, err)
		}
	}
	names, err := ParseRDMADevices(" hrn5_3, mlx5_999,hrn5_3,hrn5_0 ")
	if err != nil || !reflect.DeepEqual(names, []string{"hrn5_3", "mlx5_999", "hrn5_0"}) {
		t.Fatalf("names %v: %v", names, err)
	}
	for _, value := range []string{
		"0", "0-7", "-1", "+1", "hrn5_0,", ",hrn5_0", "hrn5_0,,hrn5_1", "void,hrn5_0", ".",
		"../hrn5_0", "/hrn5_0", "a\\b", "hrn5_*", "hrn5_?", "hrn5_[0]", "hrn5_ 0", "h\x00ca",
		"hrn5_0:1", "hca@0", "hca;0", "hca=0", "hca'0", "hca\"0", "hca\u4e00", "hca\t0", "hca\n0",
		"1hca0", "_hca0", ".hca0", "hca0.", "hca..0",
		"hca-0", "hca.0",
	} {
		got, err := ParseRDMADevices(value)
		if err == nil || got != nil || !strings.Contains(err.Error(), api.AscendRdmaDevicesEnv) {
			t.Fatalf("invalid %q: %v %v", value, got, err)
		}
	}
}

func TestParseRDMADevicesAllowedNames(t *testing.T) {
	want := []string{"hrn5_0", "mlx5_999", "rxe0", "HCA_0"}
	got, err := ParseRDMADevices(strings.Join(want, ","))
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("allowed names %v: %v", got, err)
	}
}

func TestParseRDMADevicesUniqueLimit(t *testing.T) {
	names := make([]string, 512)
	for i := range names {
		names[i] = fmt.Sprintf("hrn5_%d", i)
	}
	value := strings.Join(names, ",")
	got, err := ParseRDMADevices(value + "," + value)
	if err != nil || len(got) != 512 {
		t.Fatalf("512 unique names: %d %v", len(got), err)
	}
	if got, err = ParseRDMADevices(value + ",hrn5_512"); err == nil || got != nil {
		t.Fatal("513 unique names must fail")
	}
}

func TestResolveRdmaDeviceNode(t *testing.T) {
	root := t.TempDir()
	replaceRdmaSysfs(t, root)
	verbs := filepath.Join(root, "mlx5_999", "device", "infiniband_verbs")
	if err := os.MkdirAll(filepath.Join(verbs, "uverbs17"), 0755); err != nil {
		t.Fatal(err)
	}
	got, err := ResolveRdmaDeviceNode("mlx5_999")
	if err != nil || got != "/dev/infiniband/uverbs17" {
		t.Fatalf("mapping %s: %v", got, err)
	}

	if err := os.Symlink(filepath.Join(root, "mlx5_999"), filepath.Join(root, "hrn5_0")); err != nil {
		t.Fatal(err)
	}
	if got, err = ResolveRdmaDeviceNode("hrn5_0"); err != nil || got != "/dev/infiniband/uverbs17" {
		t.Fatalf("symlink %s: %v", got, err)
	}

	if err := os.Mkdir(filepath.Join(verbs, "uverbs18"), 0755); err != nil {
		t.Fatal(err)
	}
	if _, err = ResolveRdmaDeviceNode("hrn5_0"); err == nil || !strings.Contains(err.Error(), "got 2") {
		t.Fatalf("ambiguous: %v", err)
	}
}

func TestResolveRdmaDeviceNodeFailures(t *testing.T) {
	root := t.TempDir()
	replaceRdmaSysfs(t, root)
	for _, name := range []string{"missing", "../escape", "0-7", "void"} {
		if _, err := ResolveRdmaDeviceNode(name); err == nil {
			t.Fatalf("accepted %q", name)
		}
	}

	hca := filepath.Join(root, "hrn5_0")
	if err := os.WriteFile(hca, []byte("file"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := ResolveRdmaDeviceNode("hrn5_0"); err == nil {
		t.Fatal("non-directory HCA accepted")
	}

	if err := os.Remove(hca); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(hca, 0755); err != nil {
		t.Fatal(err)
	}
	if _, err := ResolveRdmaDeviceNode("hrn5_0"); err == nil {
		t.Fatal("missing verbs accepted")
	}
	verbs := filepath.Join(hca, "device", "infiniband_verbs")
	if err := os.MkdirAll(verbs, 0755); err != nil {
		t.Fatal(err)
	}
	if _, err := ResolveRdmaDeviceNode("hrn5_0"); err == nil || !strings.Contains(err.Error(), "got 0") {
		t.Fatalf("empty verbs: %v", err)
	}

	if err := os.Symlink(filepath.Join(root, "missing"), filepath.Join(verbs, "uverbs0")); err != nil {
		t.Fatal(err)
	}
	if _, err := ResolveRdmaDeviceNode("hrn5_0"); err == nil {
		t.Fatal("broken mapping accepted")
	}
}

func replaceRdmaSysfs(t *testing.T, root string) {
	t.Helper()
	original := rdmaSysfsDir
	rdmaSysfsDir = root
	t.Cleanup(func() { rdmaSysfsDir = original })
}
