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
	"os"
	"strings"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/require"

	"ascend-common/api"
	"ascend-common/rdma"
	"ascend-docker-runtime/mindxcheckutils"
)

func TestRDMAHookPureRequestAndNODRV(t *testing.T) {
	for _, visible := range []string{"missing", "", "void", " ", "0"} {
		for _, options := range []string{"", "VIRTUAL,NODRV"} {
			for _, mounts := range []string{"", "custom,rdma_driver,custom"} {
				for _, ub := range []string{"", "True", "False"} {
					t.Run("npu="+visible+"/options="+options+"/mounts="+mounts+"/ub="+ub, func(t *testing.T) {
						checkRDMAHookRequest(t, visible, options, mounts, ub)
					})
				}
			}
		}
	}
}

func checkRDMAHookRequest(t *testing.T, visible, options, mounts, ub string) {
	t.Helper()
	env := []string{api.AscendRdmaDevicesEnv + "=hrn5_2,hrn5_0,hrn5_1,hrn5_2", api.AscendUBDrvMountEnv + "=" + ub,
		api.AscendRuntimeMountsEnv + "=" + mounts, api.AscendAllowLinkEnv + "=False",
		api.AscendRuntimeOptionsEnv + "=" + options}
	if visible != "missing" {
		env = append(env, api.AscendVisibleDevicesEnv+"="+visible)
	}
	patches := prepareRDMAHook(t, env)
	discoverRdmaMountConfigs = func(dir string, hcaNames []string) ([]string, error) {
		require.Equal(t, configDir, dir)
		require.Equal(t, []string{"hrn5_2", "hrn5_0", "hrn5_1"}, hcaNames)
		return []string{"rdma_driver", "rdma_extra"}, nil
	}
	var reads []string
	patches.ApplyFunc(readMountConfig, func(dir, name string) ([]string, []string, error) {
		require.Equal(t, configDir, dir)
		reads = append(reads, "ordinary:"+name)
		return []string{"/host/driver.so:/container/driver.so"}, nil, nil
	}).ApplyFunc(rdma.ReadMountEntries, func(dir, name string) ([]string, []string, error) {
		require.Equal(t, configDir, dir)
		reads = append(reads, "strict:"+name)
		return []string{"/host/driver.so:/container/driver.so", "/host/driver.so:/container/extra.so"},
			[]string{"/host/config", "/host/config:/host/config"}, nil
	}).ApplyFunc(addUBMount, func(files, dirs []string) ([]string, []string) {
		reads = append(reads, "topo")
		return files, dirs
	})
	doExec = func(_ string, args, _ []string) error {
		want := []string{"/usr/local/bin/ascend-docker-cli", "--allow-link", "True", "--pid", "123",
			"--rootfs", "/rootfs", "--mount-file", "/host/driver.so:/container/driver.so",
			"--mount-file", "/host/driver.so:/container/extra.so", "--mount-dir", "/host/config"}
		if options != "" {
			want = append(want, "--options", options)
		}
		require.Equal(t, want, args)
		reads = append(reads, "exec")
		return nil
	}
	require.NoError(t, DoPrestartHook())
	require.Equal(t, wantRDMAHookReads(visible, mounts, ub), reads)
}

func wantRDMAHookReads(visible, mounts, ub string) []string {
	want := []string{}
	if visible == "0" {
		if mounts == "" {
			want = append(want, "ordinary:base")
		} else {
			want = append(want, "ordinary:custom", "strict:rdma_driver")
		}
		if ub != "False" {
			want = append(want, "ordinary:ub_driver")
		}
	}
	if visible != "0" || mounts == "" {
		want = append(want, "strict:rdma_driver")
	}
	want = append(want, "strict:rdma_extra")
	if visible == "0" {
		want = append(want, "topo")
	}
	return append(want, "exec")
}

func TestRDMAHookFailureStopsExecution(t *testing.T) {
	for _, stage := range []string{"discovery", "strict"} {
		for _, options := range []string{"", "NODRV"} {
			t.Run(stage+"/"+options, func(t *testing.T) {
				checkRDMAHookFailure(t, stage, options)
			})
		}
	}
}

func checkRDMAHookFailure(t *testing.T, stage, options string) {
	t.Helper()
	patches := prepareRDMAHook(t, []string{api.AscendRdmaDevicesEnv + "=hrn5_0",
		api.AscendUBDrvMountEnv + "=False", api.AscendRuntimeOptionsEnv + "=" + options})
	cause := errors.New("missing host RDMA dependency")
	discoverRdmaMountConfigs = func(string, []string) ([]string, error) {
		if stage == "discovery" {
			return nil, cause
		}
		return []string{"rdma_driver"}, nil
	}
	strictCalls := 0
	patches.ApplyFunc(readMountConfig, func(string, string) ([]string, []string, error) {
		require.Equal(t, "strict", stage, "discovery failure must precede any list read")
		return nil, nil, nil
	}).ApplyFunc(rdma.ReadMountEntries, func(string, string) ([]string, []string, error) {
		strictCalls++
		return nil, nil, cause
	})
	require.ErrorIs(t, DoPrestartHook(), cause)
	wantStrictCalls := 0
	if stage == "strict" {
		wantStrictCalls = 1
	}
	require.Equal(t, wantStrictCalls, strictCalls)
}

func TestRDMAHookEmptyRequestKeepsOrdinaryCollections(t *testing.T) {
	for _, value := range []string{"", " ", "void"} {
		t.Run(value, func(t *testing.T) {
			patches := prepareRDMAHook(t, []string{api.AscendVisibleDevicesEnv + "=0",
				api.AscendRdmaDevicesEnv + "=" + value, api.AscendUBDrvMountEnv + "=False",
				api.AscendRuntimeMountsEnv + "=rdma_driver,custom"})
			var reads []string
			patches.ApplyFunc(readMountConfig, func(_ string, name string) ([]string, []string, error) {
				reads = append(reads, name)
				return nil, nil, nil
			}).ApplyFunc(rdma.ReadMountEntries, func(string, string) ([]string, []string, error) {
				t.Fatal("empty RDMA request must not use strict reading")
				return nil, nil, nil
			})
			executed := false
			doExec = func(_ string, args, _ []string) error {
				executed = true
				require.Contains(t, strings.Join(args, " "), "--allow-link False")
				return nil
			}
			require.NoError(t, DoPrestartHook())
			require.True(t, executed)
			require.Equal(t, []string{"rdma_driver", "custom"}, reads)
		})
	}
}

func TestRDMAHookInvalidRequestAndNoDevices(t *testing.T) {
	for _, value := range []string{"129", "1,,2", "2-0"} {
		t.Run(value, func(t *testing.T) {
			prepareRDMAHook(t, []string{api.AscendRdmaDevicesEnv + "=" + value})
			require.ErrorContains(t, DoPrestartHook(), api.AscendRdmaDevicesEnv)
		})
	}
	for _, visible := range []string{"", "void", " "} {
		t.Run("no devices/"+visible, func(t *testing.T) {
			patches := prepareRDMAHook(t, []string{api.AscendVisibleDevicesEnv + "=" + visible,
				api.AscendRdmaDevicesEnv + "=void"})
			patches.ApplyFunc(readConfigsOfDir, func(string, []string, map[string]bool) ([]string, []string, error) {
				t.Fatal("no device request must return before mount processing")
				return nil, nil, nil
			})
			require.NoError(t, DoPrestartHook())
		})
	}
}

func TestSelectHookMountConfigsWithoutRDMAHasNoDiscovery(t *testing.T) {
	prepareRDMAHook(t, nil)
	for _, configNames := range [][]string{nil, {}, {"rdma_driver", "custom", "custom"}} {
		actual, strictNames, err := selectHookMountConfigs(configNames, nil)
		require.NoError(t, err)
		require.Equal(t, configNames, actual)
		require.Nil(t, strictNames)
	}
}

func TestHookDeduplicatesAfterUBWithoutRDMA(t *testing.T) {
	for _, ub := range []string{"", "True", "False"} {
		t.Run("ub="+ub, func(t *testing.T) { checkHookDeduplicatesAfterUB(t, ub) })
	}
}

func checkHookDeduplicatesAfterUB(t *testing.T, ub string) {
	t.Helper()
	patches := prepareRDMAHook(t, []string{api.AscendVisibleDevicesEnv + "=0", api.AscendUBDrvMountEnv + "=" + ub})
	var reads []string
	patches.ApplyFunc(readMountConfig, func(_ string, name string) ([]string, []string, error) {
		reads = append(reads, name)
		return []string{"/host/driver.so"}, []string{"/host/config"}, nil
	})
	patches.ApplyFunc(addUBMount, func(files, dirs []string) ([]string, []string) {
		return append(files, "/host/driver.so:/host/driver.so"), append(dirs, "/host/config:/host/config")
	})
	executed := false
	doExec = func(_ string, args, _ []string) error {
		executed = true
		allowLink := "True"
		if ub == "False" {
			allowLink = "False"
		}
		require.Equal(t, []string{"/usr/local/bin/ascend-docker-cli", "--allow-link", allowLink, "--pid", "123",
			"--rootfs", "/rootfs", "--mount-file", "/host/driver.so", "--mount-dir", "/host/config"}, args)
		return nil
	}
	require.NoError(t, DoPrestartHook())
	require.True(t, executed)
	want := []string{"base"}
	if ub != "False" {
		want = append(want, "ub_driver")
	}
	require.Equal(t, want, reads)
}

// These tests replace package-level functions and must not run in parallel.
func prepareRDMAHook(t *testing.T, env []string) *gomonkey.Patches {
	t.Helper()
	originalConfig, originalDiscover, originalExec := getContainerConfig, discoverRdmaMountConfigs, doExec
	t.Cleanup(func() {
		getContainerConfig, discoverRdmaMountConfigs, doExec = originalConfig, originalDiscover, originalExec
	})
	getContainerConfig = func() (*containerConfig, error) {
		return &containerConfig{Pid: 123, Rootfs: "/rootfs", Env: env}, nil
	}
	discoverRdmaMountConfigs = func(string, []string) ([]string, error) {
		t.Error("unexpected RDMA hardware discovery")
		return nil, errors.New("unexpected discovery")
	}
	doExec = func(string, []string, []string) error {
		t.Error("unexpected CLI execution")
		return errors.New("unexpected exec")
	}
	patches := gomonkey.NewPatches().
		ApplyFuncReturn(os.Executable, "/usr/local/bin/ascend-docker-hook", nil).
		ApplyFuncReturn(os.Stat, mockFileInfo{mode: os.ModeDir}, nil).
		ApplyFuncReturn(mindxcheckutils.RealFileChecker, "", nil).
		ApplyFuncReturn(mindxcheckutils.ChangeRuntimeLogMode, nil).
		ApplyFunc(addUBMount, func(files, dirs []string) ([]string, []string) { return files, dirs })
	t.Cleanup(patches.Reset)
	return patches
}
