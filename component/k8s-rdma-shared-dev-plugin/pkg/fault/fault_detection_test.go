/*
   Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package fault for fault check and fault report
package fault

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
	"github.com/Mellanox/k8s-rdma-shared-dev-plugin/pkg/resources/common"
)

func init() {
	_ = hwlog.InitRunLogger(&hwlog.LogConfig{OnlyToStdout: true}, context.Background())
}

func TestValidateSysfsPath(t *testing.T) {
	convey.Convey("When path starts with /sys/", t, func() {
		convey.So(validateSysfsPath("/sys/class/net"), convey.ShouldBeTrue)
	})
	convey.Convey("When path does not start with /sys/", t, func() {
		convey.So(validateSysfsPath("/etc/passwd"), convey.ShouldBeFalse)
	})
	convey.Convey("When path is empty", t, func() {
		convey.So(validateSysfsPath(""), convey.ShouldBeFalse)
	})
}

func TestLoadFaultConfigSuccess(t *testing.T) {
	jsonData := `{"faults":[{"name":"ub_port_down","description":"test","faultcode":"21000022","faultlevel":"SeparateDPU","check_method":"check_ub_port"}]}`

	patches := gomonkey.ApplyFunc(utils.LoadFile, func(path string) ([]byte, error) {
		return []byte(jsonData), nil
	})
	defer patches.Reset()

	convey.Convey("When fault config file is valid", t, func() {
		config, err := LoadFaultConfig()
		convey.Convey("Then config should be loaded correctly", func() {
			convey.So(err, convey.ShouldBeNil)
			convey.So(config, convey.ShouldNotBeNil)
			convey.So(len(config.Faults), convey.ShouldEqual, 1)
			convey.So(config.Faults[0].Name, convey.ShouldEqual, "ub_port_down")
			convey.So(config.Faults[0].FaultCode, convey.ShouldEqual, "21000022")
			convey.So(config.Faults[0].FaultLevel, convey.ShouldEqual, "SeparateDPU")
			convey.So(config.Faults[0].CheckMethod, convey.ShouldEqual, "check_ub_port")
		})
	})
}

func TestLoadFaultConfigFileError(t *testing.T) {
	patches := gomonkey.ApplyFunc(utils.LoadFile, func(path string) ([]byte, error) {
		return nil, errors.New("file not found")
	})
	defer patches.Reset()

	convey.Convey("When fault config file cannot be read", t, func() {
		_, err := LoadFaultConfig()
		convey.Convey("Then error should be returned", func() {
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldContainSubstring, "failed to read fault config file")
		})
	})
}

func TestLoadFaultConfigNilData(t *testing.T) {
	patches := gomonkey.ApplyFunc(utils.LoadFile, func(path string) ([]byte, error) {
		return nil, nil
	})
	defer patches.Reset()

	convey.Convey("When fault config file returns nil data", t, func() {
		_, err := LoadFaultConfig()
		convey.Convey("Then error should be returned", func() {
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldContainSubstring, "fault config file not found")
		})
	})
}

func TestLoadFaultConfigInvalidJSON(t *testing.T) {
	patches := gomonkey.ApplyFunc(utils.LoadFile, func(path string) ([]byte, error) {
		return []byte("invalid json"), nil
	})
	defer patches.Reset()

	convey.Convey("When fault config file contains invalid JSON", t, func() {
		_, err := LoadFaultConfig()
		convey.Convey("Then error should be returned", func() {
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldContainSubstring, "failed to unmarshal fault config")
		})
	})
}

func TestRunShellCommandSuccess(t *testing.T) {
	patches := gomonkey.ApplyMethodReturn(&exec.Cmd{}, "CombinedOutput", []byte("true:port is up"), nil)
	defer patches.Reset()

	convey.Convey("When shell command succeeds with valid output", t, func() {
		result, details := runShellCommand("echo 'true:port is up'")
		convey.Convey("Then result should be true", func() {
			convey.So(result, convey.ShouldEqual, "true")
			convey.So(details, convey.ShouldContainSubstring, "port is up")
		})
	})
}

func TestRunShellCommandFailure(t *testing.T) {
	patches := gomonkey.ApplyMethodReturn(&exec.Cmd{}, "CombinedOutput", []byte(""), errors.New("exit status 1"))
	defer patches.Reset()

	convey.Convey("When shell command fails", t, func() {
		result, details := runShellCommand("false")
		convey.Convey("Then result should be false", func() {
			convey.So(result, convey.ShouldEqual, "false")
			convey.So(details, convey.ShouldContainSubstring, "command failed")
		})
	})
}

func TestRunShellCommandInvalidOutput(t *testing.T) {
	patches := gomonkey.ApplyMethodReturn(&exec.Cmd{}, "CombinedOutput", []byte("no_colon_here"), nil)
	defer patches.Reset()

	convey.Convey("When shell command output has no colon separator", t, func() {
		result, details := runShellCommand("echo 'no_colon_here'")
		convey.Convey("Then result should be false with invalid format", func() {
			convey.So(result, convey.ShouldEqual, "false")
			convey.So(details, convey.ShouldContainSubstring, "invalid output format")
		})
	})
}

func TestRunShellFunction(t *testing.T) {
	patches := gomonkey.ApplyMethodReturn(&exec.Cmd{}, "CombinedOutput", []byte("false:no fault"), nil)
	defer patches.Reset()

	convey.Convey("When calling runShellFunction", t, func() {
		result, details := runShellFunction("/etc/rdma-plugin/fault_detection.sh", "check_ub_port", "mlx5_0")
		convey.Convey("Then it should execute and return result", func() {
			convey.So(result, convey.ShouldEqual, "false")
			convey.So(details, convey.ShouldContainSubstring, "no fault")
		})
	})
}

func TestRunCheckMethodNotFound(t *testing.T) {
	fault := FaultConfig{
		Name:        "test_fault",
		CheckMethod: "nonexistent_method",
	}

	convey.Convey("When check method does not exist", t, func() {
		result, details := runCheck(fault, "mlx5_0")
		convey.Convey("Then result should be false", func() {
			convey.So(result, convey.ShouldEqual, "false")
			convey.So(details, convey.ShouldContainSubstring, "not found")
		})
	})
}

func TestRunCheckMethodFound(t *testing.T) {
	fault := FaultConfig{
		Name:        "ub_port_down",
		CheckMethod: CheckUbPort,
	}

	patches := gomonkey.ApplyMethodReturn(&exec.Cmd{}, "CombinedOutput", []byte("false:no fault detected"), nil)
	defer patches.Reset()

	convey.Convey("When check method exists", t, func() {
		result, details := runCheck(fault, "mlx5_0")
		convey.Convey("Then it should execute the check", func() {
			convey.So(result, convey.ShouldEqual, "false")
			convey.So(details, convey.ShouldContainSubstring, "no fault detected")
		})
	})
}

func TestRunFaultChecksDpuCardDrop(t *testing.T) {
	config := &FaultConfigList{
		Faults: []FaultConfig{
			{
				Name:        "dpu_card_drop",
				CheckMethod: CheckDpuCardDrop,
				FaultCode:   "22000026",
				FaultLevel:  "SeparateDPU",
			},
		},
	}

	patches := gomonkey.ApplyMethodReturn(&exec.Cmd{}, "CombinedOutput", []byte("false:no drop"), nil)
	defer patches.Reset()

	convey.Convey("When checking dpu_card_drop fault", t, func() {
		results := RunFaultChecks(config, []string{"mlx5_0"})
		convey.Convey("Then it should run check with empty Hca", func() {
			convey.So(len(results), convey.ShouldEqual, 1)
			convey.So(results[0].Hca, convey.ShouldEqual, "")
		})
	})
}

func TestRunFaultChecksRegularFault(t *testing.T) {
	config := &FaultConfigList{
		Faults: []FaultConfig{
			{
				Name:        "ub_port_down",
				CheckMethod: CheckUbPort,
				FaultCode:   "21000022",
				FaultLevel:  "SeparateDPU",
			},
		},
	}

	patches := gomonkey.ApplyMethodReturn(&exec.Cmd{}, "CombinedOutput", []byte("false:port ok"), nil)
	defer patches.Reset()

	convey.Convey("When checking regular fault with multiple HCAs", t, func() {
		results := RunFaultChecks(config, []string{"mlx5_0", "mlx5_1"})
		convey.Convey("Then it should run check for each HCA", func() {
			convey.So(len(results), convey.ShouldEqual, 2)
			convey.So(results[0].Hca, convey.ShouldEqual, "mlx5_0")
			convey.So(results[1].Hca, convey.ShouldEqual, "mlx5_1")
		})
	})
}

func TestRunFaultChecksEmptyHcaList(t *testing.T) {
	config := &FaultConfigList{
		Faults: []FaultConfig{
			{
				Name:        "ub_port_down",
				CheckMethod: CheckUbPort,
				FaultCode:   "21000022",
				FaultLevel:  "SeparateDPU",
			},
		},
	}

	convey.Convey("When HCA list is empty", t, func() {
		results := RunFaultChecks(config, []string{})
		convey.Convey("Then no results should be returned for regular faults", func() {
			convey.So(len(results), convey.ShouldEqual, 0)
		})
	})
}

func TestCheckHcaPortActive(t *testing.T) {
	patches := gomonkey.ApplyFunc(utils.ReadLimitBytesWithSymlink, func(path string, limit int, validate func(string) bool) ([]byte, error) {
		if strings.Contains(path, "state") && !strings.Contains(path, "phys_state") {
			return []byte("ACTIVE"), nil
		}
		return []byte("LinkUp"), nil
	})
	defer patches.Reset()

	convey.Convey("When HCA port is ACTIVE and phys_state is LinkUp", t, func() {
		result, details := checkHcaPort("mlx5_0")
		convey.Convey("Then result should be false (no fault)", func() {
			convey.So(result, convey.ShouldEqual, "false")
			convey.So(details, convey.ShouldContainSubstring, "ACTIVE")
			convey.So(details, convey.ShouldContainSubstring, "LinkUp")
		})
	})
}

func TestCheckHcaPortNotActive(t *testing.T) {
	patches := gomonkey.ApplyFunc(utils.ReadLimitBytesWithSymlink, func(path string, limit int, validate func(string) bool) ([]byte, error) {
		if strings.Contains(path, "state") && !strings.Contains(path, "phys_state") {
			return []byte("DOWN"), nil
		}
		return []byte("Disabled"), nil
	})
	defer patches.Reset()

	convey.Convey("When HCA port is not ACTIVE", t, func() {
		result, details := checkHcaPort("mlx5_0")
		convey.Convey("Then result should be true (fault detected)", func() {
			convey.So(result, convey.ShouldEqual, "true")
			convey.So(details, convey.ShouldContainSubstring, "DOWN")
		})
	})
}

func TestCheckHcaPortStateReadError(t *testing.T) {
	patches := gomonkey.ApplyFunc(utils.ReadLimitBytesWithSymlink, func(path string, limit int, validate func(string) bool) ([]byte, error) {
		return nil, errors.New("read error")
	})
	defer patches.Reset()

	convey.Convey("When reading port state fails", t, func() {
		result, details := checkHcaPort("mlx5_0")
		convey.Convey("Then result should be false with UNKNOWN state", func() {
			convey.So(result, convey.ShouldEqual, "false")
			convey.So(details, convey.ShouldContainSubstring, "UNKNOWN")
		})
	})
}

func TestCheckHcaPortPhysStateReadError(t *testing.T) {
	patches := gomonkey.ApplyFunc(utils.ReadLimitBytesWithSymlink, func(path string, limit int, validate func(string) bool) ([]byte, error) {
		if strings.Contains(path, "phys_state") {
			return nil, errors.New("read error")
		}
		return []byte("ACTIVE"), nil
	})
	defer patches.Reset()

	convey.Convey("When reading phys_state fails", t, func() {
		result, details := checkHcaPort("mlx5_0")
		convey.Convey("Then result should be false with UNKNOWN phys_state", func() {
			convey.So(result, convey.ShouldEqual, "false")
			convey.So(details, convey.ShouldContainSubstring, "ACTIVE")
			convey.So(details, convey.ShouldContainSubstring, "UNKNOWN")
		})
	})
}

func TestCheckBondMemberNoEthName(t *testing.T) {
	patches := gomonkey.ApplyFunc(GetHcaEthName, func(hca string) string {
		return ""
	})
	defer patches.Reset()

	convey.Convey("When eth name cannot be found for HCA", t, func() {
		result, details := checkBondMember("mlx5_0")
		convey.Convey("Then result should be false", func() {
			convey.So(result, convey.ShouldEqual, "false")
			convey.So(details, convey.ShouldContainSubstring, "cannot get eth name")
		})
	})
}

func TestCheckBondMemberBondNotFound(t *testing.T) {
	patches := gomonkey.ApplyFunc(GetHcaEthName, func(hca string) string {
		return "enp0s1"
	})
	patches.ApplyFunc(os.ReadDir, func(name string) ([]os.DirEntry, error) {
		return []os.DirEntry{}, nil
	})
	defer patches.Reset()

	convey.Convey("When no bond contains the eth name", t, func() {
		result, details := checkBondMember("mlx5_0")
		convey.Convey("Then result should be false", func() {
			convey.So(result, convey.ShouldEqual, "false")
			convey.So(details, convey.ShouldContainSubstring, "no bond contains eth")
		})
	})
}

func TestCheckBondMemberReadDirError(t *testing.T) {
	patches := gomonkey.ApplyFunc(GetHcaEthName, func(hca string) string {
		return "enp0s1"
	})
	patches.ApplyFunc(os.ReadDir, func(name string) ([]os.DirEntry, error) {
		return nil, errors.New("read dir error")
	})
	defer patches.Reset()

	convey.Convey("When reading net directory fails", t, func() {
		result, details := checkBondMember("mlx5_0")
		convey.Convey("Then result should be false", func() {
			convey.So(result, convey.ShouldEqual, "false")
			convey.So(details, convey.ShouldContainSubstring, "bond not found")
		})
	})
}

func TestFindBondByEthNameReadDirError(t *testing.T) {
	patches := gomonkey.ApplyFunc(os.ReadDir, func(name string) ([]os.DirEntry, error) {
		return nil, errors.New("permission denied")
	})
	defer patches.Reset()

	convey.Convey("When reading sys/class/net fails", t, func() {
		_, _, err := findBondByEthName("enp0s1")
		convey.Convey("Then error should be returned", func() {
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldContainSubstring, "permission denied")
		})
	})
}

func TestFindBondByEthNameNoBonds(t *testing.T) {
	entries := []os.DirEntry{
		&mockDirEntry{name: "eth0", isDir: true},
		&mockDirEntry{name: "lo", isDir: true},
	}

	patches := gomonkey.ApplyFunc(os.ReadDir, func(name string) ([]os.DirEntry, error) {
		return entries, nil
	})
	defer patches.Reset()

	convey.Convey("When no bond interfaces exist", t, func() {
		bondName, _, err := findBondByEthName("eth0")
		convey.Convey("Then no bond should be found", func() {
			convey.So(err, convey.ShouldBeNil)
			convey.So(bondName, convey.ShouldEqual, "")
		})
	})
}

func TestFindBondByEthNameBondFound(t *testing.T) {
	netEntries := []os.DirEntry{
		&mockDirEntry{name: "bond0", isDir: true},
		&mockDirEntry{name: "eth0", isDir: true},
	}

	patches := gomonkey.ApplyFunc(os.ReadDir, func(name string) ([]os.DirEntry, error) {
		if strings.Contains(name, "bonding") {
			return nil, nil
		}
		return netEntries, nil
	})
	patches.ApplyFunc(utils.ReadLimitBytesWithSymlink, func(path string, limit int, validate func(string) bool) ([]byte, error) {
		if strings.Contains(path, "slaves") {
			return []byte("enp0s1 enp0s2"), nil
		}
		return nil, nil
	})
	defer patches.Reset()

	convey.Convey("When bond containing the eth name is found", t, func() {
		bondName, slaves, err := findBondByEthName("enp0s1")
		convey.Convey("Then bond info should be returned", func() {
			convey.So(err, convey.ShouldBeNil)
			convey.So(bondName, convey.ShouldEqual, "bond0")
			convey.So(slaves, convey.ShouldResemble, []string{"enp0s1", "enp0s2"})
		})
	})
}

func TestGetBondSlavesSuccess(t *testing.T) {
	patches := gomonkey.ApplyFunc(utils.ReadLimitBytesWithSymlink, func(path string, limit int, validate func(string) bool) ([]byte, error) {
		return []byte("enp0s1 enp0s2"), nil
	})
	defer patches.Reset()

	convey.Convey("When reading bond slaves succeeds", t, func() {
		slaves, err := getBondSlaves("bond0")
		convey.Convey("Then slaves should be parsed correctly", func() {
			convey.So(err, convey.ShouldBeNil)
			convey.So(slaves, convey.ShouldResemble, []string{"enp0s1", "enp0s2"})
		})
	})
}

func TestGetBondSlavesReadError(t *testing.T) {
	patches := gomonkey.ApplyFunc(utils.ReadLimitBytesWithSymlink, func(path string, limit int, validate func(string) bool) ([]byte, error) {
		return nil, errors.New("read error")
	})
	defer patches.Reset()

	convey.Convey("When reading bond slaves fails", t, func() {
		_, err := getBondSlaves("bond0")
		convey.Convey("Then error should be returned", func() {
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

func TestCheckBondSlavesStateOneDown(t *testing.T) {
	patches := gomonkey.ApplyFunc(utils.ReadLimitBytesWithSymlink, func(path string, limit int, validate func(string) bool) ([]byte, error) {
		if strings.Contains(path, "enp0s1") {
			return []byte("down"), nil
		}
		return []byte("up"), nil
	})
	defer patches.Reset()

	convey.Convey("When one bond member is down", t, func() {
		result, details := checkBondSlavesState("bond0", []string{"enp0s1", "enp0s2"}, "mlx5_0")
		convey.Convey("Then fault should be detected", func() {
			convey.So(result, convey.ShouldEqual, "true")
			convey.So(details, convey.ShouldContainSubstring, "one member")
			convey.So(details, convey.ShouldContainSubstring, "down")
		})
	})
}

func TestCheckBondSlavesStateAllDown(t *testing.T) {
	patches := gomonkey.ApplyFunc(utils.ReadLimitBytesWithSymlink, func(path string, limit int, validate func(string) bool) ([]byte, error) {
		return []byte("down"), nil
	})
	defer patches.Reset()

	convey.Convey("When all bond members are down", t, func() {
		result, details := checkBondSlavesState("bond0", []string{"enp0s1", "enp0s2"}, "mlx5_0")
		convey.Convey("Then result should be false (all down, not partial fault)", func() {
			convey.So(result, convey.ShouldEqual, "false")
			convey.So(details, convey.ShouldContainSubstring, "all members down")
		})
	})
}

func TestCheckBondSlavesStateAllUp(t *testing.T) {
	patches := gomonkey.ApplyFunc(utils.ReadLimitBytesWithSymlink, func(path string, limit int, validate func(string) bool) ([]byte, error) {
		return []byte("up"), nil
	})
	defer patches.Reset()

	convey.Convey("When all bond members are up", t, func() {
		result, details := checkBondSlavesState("bond0", []string{"enp0s1", "enp0s2"}, "mlx5_0")
		convey.Convey("Then no fault should be detected", func() {
			convey.So(result, convey.ShouldEqual, "false")
			convey.So(details, convey.ShouldContainSubstring, "no bond member failure")
		})
	})
}

func TestIsEthPortDown(t *testing.T) {
	convey.Convey("When operstate is down", t, func() {
		patches := gomonkey.ApplyFunc(utils.ReadLimitBytesWithSymlink, func(path string, limit int, validate func(string) bool) ([]byte, error) {
			return []byte("down"), nil
		})
		defer patches.Reset()
		convey.So(isEthPortDown("enp0s1"), convey.ShouldBeTrue)
	})

	convey.Convey("When operstate is up", t, func() {
		patches := gomonkey.ApplyFunc(utils.ReadLimitBytesWithSymlink, func(path string, limit int, validate func(string) bool) ([]byte, error) {
			return []byte("up"), nil
		})
		defer patches.Reset()
		convey.So(isEthPortDown("enp0s1"), convey.ShouldBeFalse)
	})

	convey.Convey("When reading operstate fails", t, func() {
		patches := gomonkey.ApplyFunc(utils.ReadLimitBytesWithSymlink, func(path string, limit int, validate func(string) bool) ([]byte, error) {
			return nil, errors.New("read error")
		})
		defer patches.Reset()
		convey.So(isEthPortDown("enp0s1"), convey.ShouldBeFalse)
	})
}

func TestCheckDpuCardDrop(t *testing.T) {
	patches := gomonkey.ApplyMethodReturn(&exec.Cmd{}, "CombinedOutput", []byte("false:no drop detected"), nil)
	defer patches.Reset()

	convey.Convey("When checking DPU card drop", t, func() {
		result, details := checkDpuCardDrop("mlx5_0")
		convey.Convey("Then it should return result from shell function", func() {
			convey.So(result, convey.ShouldEqual, "false")
			convey.So(details, convey.ShouldContainSubstring, "no drop detected")
		})
	})
}

func TestCheckUbPort(t *testing.T) {
	patches := gomonkey.ApplyMethodReturn(&exec.Cmd{}, "CombinedOutput", []byte("true:ub port down"), nil)
	defer patches.Reset()

	convey.Convey("When checking UB port", t, func() {
		result, details := checkUbPort("mlx5_0")
		convey.Convey("Then it should return result from shell function", func() {
			convey.So(result, convey.ShouldEqual, "true")
			convey.So(details, convey.ShouldContainSubstring, "ub port down")
		})
	})
}

func TestCheckUbLane(t *testing.T) {
	patches := gomonkey.ApplyMethodReturn(&exec.Cmd{}, "CombinedOutput", []byte("false:lane ok"), nil)
	defer patches.Reset()

	convey.Convey("When checking UB lane", t, func() {
		result, details := checkUbLane("mlx5_0")
		convey.Convey("Then it should return result from shell function", func() {
			convey.So(result, convey.ShouldEqual, "false")
			convey.So(details, convey.ShouldContainSubstring, "lane ok")
		})
	})
}

func TestStartFaultDetectionContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	getHcaList := func() []string { return []string{"mlx5_0"} }
	rediscoverCh := make(chan struct{}, 1)

	convey.Convey("When context is cancelled", t, func() {
		cancel()
		StartFaultDetection(ctx, getHcaList, rediscoverCh, 10)
		convey.So(true, convey.ShouldBeTrue)
	})
}

func TestStartFaultDetectionRediscover(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	getHcaList := func() []string { return []string{"mlx5_0"} }
	rediscoverCh := make(chan struct{}, 1)

	patches := gomonkey.ApplyFunc(LoadFaultConfig, func() (*FaultConfigList, error) {
		cancel()
		return nil, errors.New("config error")
	})
	defer patches.Reset()

	convey.Convey("When rediscover signal is received", t, func() {
		rediscoverCh <- struct{}{}
		StartFaultDetection(ctx, getHcaList, rediscoverCh, 10)
		convey.So(true, convey.ShouldBeTrue)
	})
}

func TestStartFaultDetectionLoadConfigError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	getHcaList := func() []string { return []string{"mlx5_0"} }
	rediscoverCh := make(chan struct{}, 1)

	callCount := 0
	patches := gomonkey.ApplyFunc(LoadFaultConfig, func() (*FaultConfigList, error) {
		callCount++
		if callCount > 1 {
			cancel()
		}
		return nil, errors.New("config error")
	})
	defer patches.Reset()

	ticker := time.NewTicker(10 * time.Millisecond)
	patches.ApplyFunc(time.NewTicker, func(d time.Duration) *time.Ticker {
		return ticker
	})

	convey.Convey("When LoadFaultConfig returns error on ticker", t, func() {
		StartFaultDetection(ctx, getHcaList, rediscoverCh, 10)
		convey.So(true, convey.ShouldBeTrue)
	})
}

func TestStartFaultDetectionSuccess(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	getHcaList := func() []string { return []string{"mlx5_0"} }
	rediscoverCh := make(chan struct{}, 1)

	config := &FaultConfigList{
		Faults: []FaultConfig{
			{Name: "hca_port_down", CheckMethod: CheckHcaPort, FaultCode: "21000024", FaultLevel: "SeparateDPU"},
		},
	}

	callCount := 0
	patches := gomonkey.ApplyFunc(LoadFaultConfig, func() (*FaultConfigList, error) {
		callCount++
		if callCount > 1 {
			cancel()
		}
		return config, nil
	})
	patches.ApplyFunc(RunFaultChecks, func(config *FaultConfigList, hcas []string) []FaultResult {
		cancel()
		return []FaultResult{}
	})
	patches.ApplyFunc(BuildDPUInfoCfg, func(results []FaultResult) DpuInfoCfg {
		return DpuInfoCfg{}
	})
	defer patches.Reset()

	ticker := time.NewTicker(10 * time.Millisecond)
	patches.ApplyFunc(time.NewTicker, func(d time.Duration) *time.Ticker {
		return ticker
	})

	convey.Convey("When fault detection runs successfully", t, func() {
		StartFaultDetection(ctx, getHcaList, rediscoverCh, 10)
		convey.So(true, convey.ShouldBeTrue)
	})

	select {
	case <-faultResultChan:
	default:
	}
}

func TestStartFaultDetectionChannelFull(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	getHcaList := func() []string { return []string{"mlx5_0"} }
	rediscoverCh := make(chan struct{}, 1)

	select {
	case <-faultResultChan:
	default:
	}

	faultResultChan <- DpuInfoCfg{}

	config := &FaultConfigList{
		Faults: []FaultConfig{
			{Name: "hca_port_down", CheckMethod: CheckHcaPort, FaultCode: "21000024", FaultLevel: "SeparateDPU"},
		},
	}

	patches := gomonkey.ApplyFunc(LoadFaultConfig, func() (*FaultConfigList, error) {
		return config, nil
	})
	patches.ApplyFunc(RunFaultChecks, func(config *FaultConfigList, hcas []string) []FaultResult {
		return []FaultResult{}
	})
	patches.ApplyFunc(BuildDPUInfoCfg, func(results []FaultResult) DpuInfoCfg {
		cancel()
		return DpuInfoCfg{}
	})
	defer patches.Reset()

	ticker := time.NewTicker(10 * time.Millisecond)
	patches.ApplyFunc(time.NewTicker, func(d time.Duration) *time.Ticker {
		return ticker
	})

	convey.Convey("When fault result channel is full", t, func() {
		StartFaultDetection(ctx, getHcaList, rediscoverCh, 10)
		convey.So(true, convey.ShouldBeTrue)
	})

	select {
	case <-faultResultChan:
	default:
	}
}

type mockDirEntry struct {
	name  string
	isDir bool
}

func (m *mockDirEntry) Name() string               { return m.name }
func (m *mockDirEntry) IsDir() bool                { return m.isDir }
func (m *mockDirEntry) Type() os.FileMode          { return 0 }
func (m *mockDirEntry) Info() (os.FileInfo, error) { return nil, nil }

func TestCheckBondMemberBondWithSlavesReadError(t *testing.T) {
	patches := gomonkey.ApplyFunc(GetHcaEthName, func(hca string) string {
		return "enp0s1"
	})
	patches.ApplyFunc(os.ReadDir, func(name string) ([]os.DirEntry, error) {
		if strings.Contains(name, common.SysClassNet) {
			return []os.DirEntry{&mockDirEntry{name: "bond0", isDir: true}}, nil
		}
		return nil, nil
	})
	patches.ApplyFunc(utils.ReadLimitBytesWithSymlink, func(path string, limit int, validate func(string) bool) ([]byte, error) {
		if strings.Contains(path, "slaves") {
			return nil, errors.New("read slaves error")
		}
		return nil, nil
	})
	defer patches.Reset()

	convey.Convey("When reading bond slaves fails", t, func() {
		result, details := checkBondMember("mlx5_0")
		convey.Convey("Then result should be false", func() {
			convey.So(result, convey.ShouldEqual, "false")
			convey.So(details, convey.ShouldContainSubstring, "no bond contains eth")
		})
	})
}

func TestCheckBondMemberBondWithWrongSlaveCount(t *testing.T) {
	patches := gomonkey.ApplyFunc(GetHcaEthName, func(hca string) string {
		return "enp0s1"
	})
	patches.ApplyFunc(os.ReadDir, func(name string) ([]os.DirEntry, error) {
		if strings.Contains(name, common.SysClassNet) {
			return []os.DirEntry{&mockDirEntry{name: "bond0", isDir: true}}, nil
		}
		return nil, nil
	})
	patches.ApplyFunc(utils.ReadLimitBytesWithSymlink, func(path string, limit int, validate func(string) bool) ([]byte, error) {
		if strings.Contains(path, "slaves") {
			return []byte("enp0s1"), nil
		}
		return nil, nil
	})
	defer patches.Reset()

	convey.Convey("When bond has wrong number of slaves", t, func() {
		result, details := checkBondMember("mlx5_0")
		convey.Convey("Then result should be false", func() {
			convey.So(result, convey.ShouldEqual, "false")
			convey.So(details, convey.ShouldContainSubstring, "no bond contains eth")
		})
	})
}

func TestCheckBondMemberSuccess(t *testing.T) {
	patches := gomonkey.ApplyFunc(GetHcaEthName, func(hca string) string {
		return "enp0s1"
	})
	patches.ApplyFunc(os.ReadDir, func(name string) ([]os.DirEntry, error) {
		if strings.Contains(name, common.SysClassNet) {
			return []os.DirEntry{&mockDirEntry{name: "bond0", isDir: true}}, nil
		}
		return nil, nil
	})
	patches.ApplyFunc(utils.ReadLimitBytesWithSymlink, func(path string, limit int, validate func(string) bool) ([]byte, error) {
		if strings.Contains(path, "slaves") {
			return []byte("enp0s1 enp0s2"), nil
		}
		if strings.Contains(path, "operstate") {
			return []byte("up"), nil
		}
		return nil, nil
	})
	defer patches.Reset()

	convey.Convey("When bond member check succeeds with all members up", t, func() {
		result, details := checkBondMember("mlx5_0")
		convey.Convey("Then no fault should be detected", func() {
			convey.So(result, convey.ShouldEqual, "false")
			convey.So(details, convey.ShouldContainSubstring, "no bond member failure")
		})
	})
}
