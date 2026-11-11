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

package process

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/containerd/containerd/oci"
	"github.com/opencontainers/runtime-spec/specs-go"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"ascend-common/rdma"
	"ascend-docker-runtime/mindxcheckutils"
	"ascend-docker-runtime/runtime/common"
	"ascend-docker-runtime/runtime/dcmi"
)

func TestCDIRDMANoticePreservesRequests(t *testing.T) {
	for _, npuIDs := range [][]int{nil, {1}} {
		for _, tc := range []struct {
			value string
			warn  bool
		}{
			{"", false}, {" ", false}, {"void", false}, {" void ", false},
			{"hrn5_0", true}, {"../hrn5_0", true}, {"a,,b", true},
		} {
			t.Run(fmt.Sprintf("NPU=%v/RDMA=%q", npuIDs, tc.value), func(t *testing.T) {
				calls, warnings := 0, []string{}
				patches := gomonkey.ApplyFuncReturn(loadConfig, &Config{InjectionMode: cdiInjectionMode}).
					ApplyFuncReturn(checkVisibleDevice, npuIDs, nil).
					ApplyFunc(processDevicesCDI, func(_ *specs.Spec, ids []int) error {
						calls++
						if !reflect.DeepEqual(ids, npuIDs) {
							t.Fatalf("CDI received NPU IDs %v, want %v", ids, npuIDs)
						}
						return nil
					}).ApplyFunc(addAscendDockerEnv, func(*specs.Spec) {}).
					ApplyFuncReturn(isMountByRuntimeForDP, false).
					ApplyMethodFunc(hwlog.RunLog, "Warnf", func(format string, args ...interface{}) {
						warnings = append(warnings, fmt.Sprintf(format, args...))
					})
				defer patches.Reset()
				spec := &specs.Spec{Process: &specs.Process{Env: []string{api.AscendRdmaDevicesEnv + "=" + tc.value}}}
				if err := processDevicesAndHooks(spec); err != nil || calls != 1 || (len(warnings) != 0) != tc.warn {
					t.Fatalf("CDI calls=%d warnings=%v error=%v", calls, warnings, err)
				}
				if tc.warn && (len(warnings) != 1 || !strings.Contains(warnings[0], api.AscendRdmaDevicesEnv) ||
					!strings.Contains(warnings[0], "legacy")) {
					t.Fatalf("missing actionable CDI warning: %v", warnings)
				}
			})
		}
	}
}

func TestCDIPureRDMAIgnoresRequestWithoutProbing(t *testing.T) {
	patches := gomonkey.ApplyFuncReturn(loadConfig, &Config{InjectionMode: cdiInjectionMode}).
		ApplyFunc(addAscendDockerEnv, func(*specs.Spec) {}).ApplyFuncReturn(isMountByRuntimeForDP, false).
		ApplyFunc(dcmi.GetMatchingNpuWorker, func() (dcmi.WorkerInterface, error) {
			t.Fatal("pure RDMA in CDI mode must not initialize DCMI")
			return nil, testError
		}).ApplyFunc(rdma.ParseRDMADevices, func(string) ([]string, error) {
		t.Fatal("CDI must not parse an unsupported RDMA request")
		return nil, testError
	}).ApplyFunc(rdma.ResolveRdmaDeviceNode, func(string) (string, error) {
		t.Fatal("CDI must not probe RDMA nodes")
		return "", testError
	})
	defer patches.Reset()
	for _, value := range []string{"hrn5_0", "../hrn5_0", "a,,b"} {
		spec := &specs.Spec{Process: &specs.Process{Env: []string{
			api.AscendVisibleDevicesEnv + "=void", api.AscendRdmaDevicesEnv + "=" + value}}}
		if err := processDevicesAndHooks(spec); err != nil || spec.Linux != nil || spec.Hooks != nil {
			t.Fatalf("ignored RDMA value %q changed CDI behavior: spec=%+v error=%v", value, spec, err)
		}
	}
}

func TestLegacyPureRDMABypassesDCMIAndPreservesVolumes(t *testing.T) {
	patches := patchRDMACharacterNodes(true)
	defer patches.Reset()
	registered := 0
	patches.ApplyFunc(addHook, func(*specs.Spec) error { registered++; return nil }).
		ApplyFunc(dcmi.GetMatchingNpuWorker, func() (dcmi.WorkerInterface, error) {
			t.Fatal("a pure RDMA request must not initialize an NPU worker")
			return nil, errors.New("unexpected DCMI")
		}).ApplyFunc(addAscendLibraryPath, func(*specs.Spec) {})
	spec := &specs.Spec{
		Process: &specs.Process{Env: []string{api.AscendRdmaDevicesEnv + "=hrn5_0,hrn5_3,hrn5_0", ascendRuntimeOptions + "=NODRV"}},
		Mounts:  []specs.Mount{{Source: "/user/driver", Destination: "/driver", Type: "bind"}},
	}
	if err := processDevicesLegacy(spec, nil); err != nil {
		t.Fatal(err)
	}
	if registered != 1 || len(spec.Linux.Devices) != 3 || len(spec.Mounts) != 1 ||
		spec.Mounts[0].Source != "/user/driver" || spec.Hooks != nil {
		t.Fatalf("pure RDMA changed unrelated state: %+v", spec)
	}
	for i, nodePath := range []string{"/dev/infiniband/uverbs0", "/dev/infiniband/uverbs3", "/dev/infiniband/rdma_cm"} {
		if spec.Linux.Devices[i].Path != nodePath {
			t.Fatalf("NODRV dropped requested node %s", nodePath)
		}
	}
}

func TestLegacyHookRegistrationMatrix(t *testing.T) {
	for _, tc := range []struct {
		name      string
		npuIDs    []int
		hasRDMA   bool
		wantHooks int
	}{
		{"NPU", []int{1}, false, 1}, {"RDMA", nil, true, 1},
		{"combined", []int{1}, true, 1}, {"empty", nil, false, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			hooks, npus := 0, 0
			patches := gomonkey.ApplyFuncReturn(processRDMARequest, tc.hasRDMA, nil).
				ApplyFunc(addHook, func(*specs.Spec) error { hooks++; return nil }).
				ApplyFunc(processNPURequest, func(_ *specs.Spec, ids []int) error {
					npus++
					if hooks != 1 || !reflect.DeepEqual(ids, tc.npuIDs) {
						t.Fatal("NPU handling ran before hook registration")
					}
					return nil
				}).ApplyFunc(addAscendLibraryPath, func(*specs.Spec) {})
			defer patches.Reset()
			if err := processDevicesLegacy(&specs.Spec{Process: &specs.Process{}}, tc.npuIDs); err != nil {
				t.Fatal(err)
			}
			if hooks != tc.wantHooks || (npus != 0) != (len(tc.npuIDs) != 0) {
				t.Fatalf("unexpected calls: hooks=%d NPU=%d", hooks, npus)
			}
		})
	}
}

func TestLegacyRejectsMissingProcess(t *testing.T) {
	for _, npuIDs := range [][]int{nil, {0}} {
		spec := &specs.Spec{}
		err := processDevicesLegacy(spec, npuIDs)
		if err == nil || err.Error() != "empty process info" || spec.Hooks != nil || spec.Linux != nil {
			t.Fatalf("missing process must fail before device or hook changes: %v, spec=%+v", err, spec)
		}
	}
}

func TestLegacyHookFailureStopsBeforeDCMI(t *testing.T) {
	for _, hasRDMA := range []bool{false, true} {
		patches := gomonkey.ApplyFuncReturn(processRDMARequest, hasRDMA, nil).
			ApplyFuncReturn(addHook, testError).
			ApplyFunc(dcmi.GetMatchingNpuWorker, func() (dcmi.WorkerInterface, error) {
				t.Fatal("failed hook registration initialized DCMI")
				return nil, testError
			}).ApplyFunc(dcmi.CreateVDevice, func(dcmi.WorkerInterface, *specs.Spec, []int) (dcmi.VDeviceInfo, error) {
			t.Fatal("failed hook registration created a vNPU")
			return dcmi.VDeviceInfo{}, testError
		})
		err := processDevicesLegacy(&specs.Spec{Process: &specs.Process{}}, []int{1})
		patches.Reset()
		if !errors.Is(err, testError) {
			t.Fatalf("lost hook failure: %v", err)
		}
	}
}

func TestLegacyNPUAndRDMASelectionAreIndependent(t *testing.T) {
	patches := patchRDMACharacterNodes(false)
	defer patches.Reset()
	npuCalls := 0
	patches.ApplyFuncReturn(addHook, nil).
		ApplyFunc(processNPURequest, func(_ *specs.Spec, ids []int) error {
			npuCalls++
			if !reflect.DeepEqual(ids, []int{1}) {
				t.Fatalf("HCA names entered NPU injection: %v", ids)
			}
			return nil
		}).ApplyFunc(addAscendLibraryPath, func(*specs.Spec) {})
	for _, source := range []string{"", "/host/rdma/driver.so", "/usr/local/Ascend/driver"} {
		spec := &specs.Spec{Process: &specs.Process{Env: []string{
			api.AscendRdmaDevicesEnv + "=hrn5_3", ascendRuntimeOptions + "=NODRV"}}}
		if source != "" {
			spec.Mounts = []specs.Mount{{Source: source, Destination: source, Type: "bind"}}
		}
		originalMounts := append([]specs.Mount(nil), spec.Mounts...)
		if err := processDevicesLegacy(spec, []int{1}); err != nil ||
			len(spec.Linux.Devices) != 1 || spec.Linux.Devices[0].Path != "/dev/infiniband/uverbs3" ||
			!reflect.DeepEqual(spec.Mounts, originalMounts) {
			t.Fatalf("independent NODRV requests/volumes failed: %v, spec=%+v", err, spec)
		}
	}
	if npuCalls != 3 {
		t.Fatalf("NPU calls: %d", npuCalls)
	}
}

func TestLegacyEmptyAndInvalidRDMADoNotProbeNodes(t *testing.T) {
	patches := gomonkey.ApplyFunc(oci.DeviceFromPath, func(string) (*specs.LinuxDevice, error) {
		t.Fatal("disabled or invalid RDMA must not probe nodes")
		return nil, errors.New("unexpected probe")
	}).ApplyFunc(rdma.ResolveRdmaDeviceNode, func(string) (string, error) {
		t.Fatal("disabled or invalid RDMA must not probe sysfs")
		return "", os.ErrNotExist
	}).ApplyFunc(addHook, func(*specs.Spec) error { t.Fatal("empty/invalid request registered hook"); return nil }).
		ApplyFunc(addAscendLibraryPath, func(*specs.Spec) {})
	defer patches.Reset()
	for _, value := range []string{"", " ", "void", "-1", "129", "0,,1", "../hrn5_0"} {
		spec := &specs.Spec{Process: &specs.Process{Env: []string{api.AscendRdmaDevicesEnv + "=" + value}}}
		err := processDevicesLegacy(spec, nil)
		invalid := value != "" && value != " " && value != "void"
		if (err != nil) != invalid || spec.Linux != nil || spec.Hooks != nil {
			t.Fatalf("value %q: error=%v, spec=%+v", value, err, spec)
		}
	}
}

func TestRDMAFailureDoesNotWritePartialSpec(t *testing.T) {
	patches := patchRDMACharacterNodes(false)
	defer patches.Reset()
	wrote := false
	spec := &specs.Spec{Process: &specs.Process{Env: []string{api.AscendRdmaDevicesEnv + "=hrn5_0,hrn5_3"}}}
	failedPath := "/dev/infiniband/uverbs3"
	patches.ApplyFuncReturn(mindxcheckutils.CheckPath, nil).
		ApplyFuncReturn(readSpecFile, spec, nil).
		ApplyFuncReturn(loadConfig, &Config{InjectionMode: "legacy"}).
		ApplyFunc(oci.DeviceFromPath, func(nodePath string) (*specs.LinuxDevice, error) {
			if nodePath == failedPath {
				return nil, os.ErrPermission
			}
			return fakeRDMADevice(nodePath, true)
		}).ApplyFunc(common.WriteSpecFile, func(string, *specs.Spec) error { wrote = true; return nil })
	for _, nodePath := range []string{"/dev/infiniband/uverbs3", "/dev/infiniband/rdma_cm"} {
		failedPath = nodePath
		spec.Linux = nil
		if err := modifySpecFile("/bundle/config.json"); err == nil || wrote || !strings.Contains(err.Error(), nodePath) {
			t.Fatalf("failure saved a partial spec: err=%v wrote=%v", err, wrote)
		}
	}
}

func TestLegacyPureRDMAKeepsEnvironmentLimit(t *testing.T) {
	patches := patchRDMACharacterNodes(false)
	defer patches.Reset()
	patches.ApplyFuncReturn(mindxcheckutils.RealFileChecker, "", nil)
	env := make([]string, MaxCommandLength+1)
	env[0] = api.AscendRdmaDevicesEnv + "=hrn5_0"
	spec := &specs.Spec{Process: &specs.Process{Env: env}}
	if err := processDevicesLegacy(spec, nil); err == nil || !strings.Contains(err.Error(), "too many items in Env") {
		t.Fatalf("pure RDMA bypassed the existing environment limit: %v", err)
	}
}

func TestRDMASelectedNodesAndExactCgroup(t *testing.T) {
	patches := patchRDMACharacterNodes(false)
	defer patches.Reset()
	spec := &specs.Spec{}
	if err := addRDMADevices(spec, []string{"hrn5_0", "hrn5_3"}); err != nil {
		t.Fatal(err)
	}
	if len(spec.Linux.Devices) != 2 || len(spec.Linux.Resources.Devices) != 2 {
		t.Fatalf("unexpected count: %+v", spec.Linux)
	}
	for i, nodePath := range []string{"/dev/infiniband/uverbs0", "/dev/infiniband/uverbs3"} {
		node, rule := spec.Linux.Devices[i], spec.Linux.Resources.Devices[i]
		if node.Path != nodePath || node.Type != "c" || node.Major != 231 ||
			rule.Major == nil || rule.Minor == nil || *rule.Major != node.Major || *rule.Minor != node.Minor ||
			!rule.Allow || rule.Type != "c" || rule.Access != "rwm" {
			t.Fatalf("node and permission do not match: node=%+v rule=%+v", node, rule)
		}
	}
	if spec.Linux.Devices[0].Minor != 5 || spec.Linux.Devices[1].Minor != 41 || len(spec.Mounts) != 0 {
		t.Fatal("must use actual device numbers, without a directory bind")
	}
}

func TestRDMAOptionalCMNodeAndExactCgroup(t *testing.T) {
	patches := patchRDMACharacterNodes(true)
	defer patches.Reset()
	reads := make(map[string]int)
	patches.ApplyFunc(oci.DeviceFromPath, func(nodePath string) (*specs.LinuxDevice, error) {
		reads[nodePath]++
		return fakeRDMADevice(nodePath, true)
	})
	spec := &specs.Spec{}
	if err := addRDMADevices(spec, []string{"hrn5_0", "hrn5_3"}); err != nil {
		t.Fatal(err)
	}
	if len(spec.Linux.Devices) != 3 || len(spec.Linux.Resources.Devices) != 3 {
		t.Fatalf("unexpected count: %+v", spec)
	}
	for nodePath, count := range reads {
		if count != 1 {
			t.Fatalf("read %s %d times", nodePath, count)
		}
	}
	node, rule := spec.Linux.Devices[2], spec.Linux.Resources.Devices[2]
	if node.Path != "/dev/infiniband/rdma_cm" || node.Type != "c" || node.Major != 10 || node.Minor != 42 ||
		rule.Major == nil || rule.Minor == nil || *rule.Major != node.Major || *rule.Minor != node.Minor ||
		!rule.Allow || rule.Type != "c" || rule.Access != "rwm" || len(spec.Mounts) != 0 {
		t.Fatalf("rdma_cm must use its actual device numbers: node=%+v rule=%+v", node, rule)
	}
}

func TestRDMAOptionalCMRejectsInvalidNodes(t *testing.T) {
	for _, tc := range []struct {
		name     string
		nodeType string
		err      error
	}{
		{"regular file", "", oci.ErrNotADevice}, {"directory", "", oci.ErrNotADevice},
		{"block device", "b", nil}, {"FIFO", "p", nil},
		{"permission denied", "", os.ErrPermission}, {"device number", "", errors.New("read device number failed")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			patches := patchRDMACharacterNodes(false)
			defer patches.Reset()
			patches.ApplyFunc(oci.DeviceFromPath, func(nodePath string) (*specs.LinuxDevice, error) {
				if nodePath == "/dev/infiniband/rdma_cm" {
					return &specs.LinuxDevice{Path: nodePath, Type: tc.nodeType}, tc.err
				}
				return fakeRDMADevice(nodePath, false)
			})
			err := addRDMADevices(&specs.Spec{}, []string{"hrn5_0"})
			if err == nil || !strings.Contains(err.Error(), "rdma_cm") ||
				!strings.Contains(err.Error(), api.AscendRdmaDevicesEnv) || (tc.err != nil && !errors.Is(err, tc.err)) {
				t.Fatalf("invalid optional node must fail with context: %v", err)
			}
		})
	}
}

func TestRDMARejectsMissingAndNonCharacterNodes(t *testing.T) {
	for _, tc := range []struct {
		nodeType string
		err      error
	}{
		{"b", nil}, {"p", nil}, {"", oci.ErrNotADevice}, {"", os.ErrNotExist}, {"", os.ErrPermission},
	} {
		patches := patchRDMACharacterNodes(false).
			ApplyFuncReturn(oci.DeviceFromPath, &specs.LinuxDevice{Type: tc.nodeType}, tc.err)
		err := addRDMADevices(&specs.Spec{}, []string{"hrn5_3"})
		patches.Reset()
		if err == nil || !strings.Contains(err.Error(), "uverbs3") ||
			!strings.Contains(err.Error(), api.AscendRdmaDevicesEnv) || (tc.err != nil && !errors.Is(err, tc.err)) {
			t.Fatalf("invalid selected node must fail with context: %v", err)
		}
	}
}

func TestRDMAMappingOrderDedupAndFailure(t *testing.T) {
	patches := patchRDMACharacterNodes(false)
	defer patches.Reset()
	calls, nodeReads := []string{}, 0
	patches.ApplyFunc(rdma.ResolveRdmaDeviceNode, func(hcaName string) (string, error) {
		calls = append(calls, hcaName)
		if hcaName == "missing" {
			return "", os.ErrNotExist
		}
		return "/dev/infiniband/uverbs17", nil
	}).ApplyFunc(oci.DeviceFromPath, func(nodePath string) (*specs.LinuxDevice, error) {
		if strings.Contains(nodePath, "uverbs") {
			nodeReads++
		}
		return fakeRDMADevice(nodePath, false)
	})
	spec := &specs.Spec{}
	if err := addRDMADevices(spec, []string{"mlx5_999", "hrn5_0"}); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(calls, []string{"mlx5_999", "hrn5_0"}) || len(spec.Linux.Devices) != 1 ||
		nodeReads != 1 || spec.Linux.Devices[0].Path != "/dev/infiniband/uverbs17" || len(spec.Linux.Resources.Devices) != 1 {
		t.Fatalf("mapping or dedup failed: calls=%v reads=%d spec=%+v", calls, nodeReads, spec)
	}
	spec.Process = &specs.Process{Env: []string{api.AscendRdmaDevicesEnv + "=mlx5_999,missing"}}
	patches.ApplyFuncReturn(mindxcheckutils.CheckPath, nil).ApplyFuncReturn(readSpecFile, spec, nil).
		ApplyFuncReturn(loadConfig, &Config{InjectionMode: "legacy"}).
		ApplyFunc(common.WriteSpecFile, func(string, *specs.Spec) error { t.Fatal("saved failed mapping"); return nil })
	if err := modifySpecFile("/bundle/config.json"); !errors.Is(err, os.ErrNotExist) || !strings.Contains(err.Error(), "missing") {
		t.Fatalf("missing selected HCA: %v", err)
	}
}

func patchRDMACharacterNodes(includeCM bool) *gomonkey.Patches {
	return gomonkey.ApplyFunc(rdma.ResolveRdmaDeviceNode, func(hcaName string) (string, error) {
		return "/dev/infiniband/uverbs" + strings.TrimPrefix(hcaName, "hrn5_"), nil
	}).ApplyFunc(oci.DeviceFromPath, func(nodePath string) (*specs.LinuxDevice, error) {
		return fakeRDMADevice(nodePath, includeCM)
	})
}

func fakeRDMADevice(nodePath string, includeCM bool) (*specs.LinuxDevice, error) {
	if nodePath == "/dev/infiniband/rdma_cm" {
		if !includeCM {
			return nil, os.ErrNotExist
		}
		return &specs.LinuxDevice{Path: nodePath, Type: "c", Major: 10, Minor: 42}, nil
	}
	minor := int64(5)
	if strings.HasSuffix(nodePath, "3") {
		minor = 41
	}
	return &specs.LinuxDevice{Path: nodePath, Type: "c", Major: 231, Minor: minor}, nil
}
