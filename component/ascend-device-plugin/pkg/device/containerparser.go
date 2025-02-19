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

// Package device a series of device function
package device

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/oci"

	"Ascend-device-plugin/pkg/common"
	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
)

const (
	sliceLen8     = 8
	ascendEnvPart = 2
	charDevice    = "c"
	minus         = "-"
	comma         = ","
	ascend        = "Ascend"
	maxEnvLength  = 1024
	maxDevicesNum = 100000
	maxEnvNum     = 10000
)

var (
	envErrDescribe = func(ctrID, devID, env string, err error) string {
		return fmt.Sprintf("container (%s) has an invalid device ID (%s) in %s, err is %v", ctrID, devID, env, err)
	}
	minusStyle = func(s string) bool {
		return strings.Contains(s, minus)
	}
	commaMinusStyle = func(s string) bool {
		return strings.Contains(s, minus) && strings.Contains(s, comma)
	}
	ascendStyle = func(s string) bool {
		return strings.Contains(s, ascend)
	}
	npuMajorID        []string
	npuMajorFetchCtrl sync.Once
)

func parseDiffEnvFmt(devices, containerID string) []int {
	if len(devices) > maxEnvLength {
		return []int{}
	}
	if ascendStyle(devices) {
		return getDeviceIDsByAscendStyle(devices, containerID)
	}
	if commaMinusStyle(devices) {
		return getDeviceIDsByCommaMinusStyle(devices, containerID)
	}
	if minusStyle(devices) {
		return getDeviceIDsByMinusStyle(devices, containerID)
	}
	return getDeviceIDsByCommaStyle(devices, containerID)
}

func getDeviceIDsByCommaStyle(devices, containerID string) []int {
	devList := strings.Split(devices, comma)
	devicesIDs := make([]int, 0, len(devList))
	for _, devID := range devList {
		id, err := strconv.Atoi(devID)
		if err != nil {
			hwlog.RunLog.Errorf("container (%s) has an invalid device ID (%v) in %s, error is %s", containerID,
				devID, common.AscendVisibleDevicesEnv, err)
			continue
		}
		devicesIDs = append(devicesIDs, id)
	}
	return devicesIDs
}

func getDeviceIDsByAscendStyle(devices, containerID string) []int {
	devList := strings.Split(devices, comma)
	deviceIDs := make([]int, 0, len(devList))
	for _, subDevice := range devList {
		deviceName := strings.Split(subDevice, minus)
		if len(deviceName) != ascendEnvPart {
			hwlog.RunLog.Error("deviceName slice length not equal to 2")
			continue
		}
		id, err := strconv.Atoi(deviceName[1])
		if err != nil {
			hwlog.RunLog.Errorf(envErrDescribe(containerID, deviceName[1], common.AscendVisibleDevicesEnv, err))
			continue
		}
		deviceIDs = append(deviceIDs, id)
	}
	return deviceIDs
}

func getDeviceIDsByMinusStyle(devices, containerID string) []int {
	deviceIDs := make([]int, 0)
	devIDRange := strings.Split(devices, minus)
	if len(devIDRange) != ascendEnvPart {
		hwlog.RunLog.Errorf(envErrDescribe(containerID, "range", common.AscendVisibleDevicesEnv, nil))
		return deviceIDs
	}
	minDevID, err := strconv.Atoi(devIDRange[0])
	if err != nil {
		hwlog.RunLog.Errorf(envErrDescribe(containerID, devIDRange[0], common.AscendVisibleDevicesEnv, err))
		return deviceIDs
	}
	maxDevID, err := strconv.Atoi(devIDRange[1])
	if err != nil {
		hwlog.RunLog.Errorf(envErrDescribe(containerID, devIDRange[1], common.AscendVisibleDevicesEnv, err))
		return deviceIDs
	}
	if minDevID > maxDevID {
		hwlog.RunLog.Errorf(envErrDescribe(containerID, "", common.AscendVisibleDevicesEnv,
			errors.New("min id bigger than max id")))
		return deviceIDs
	}
	if maxDevID > math.MaxInt16 {
		hwlog.RunLog.Errorf(envErrDescribe(containerID, "", common.AscendVisibleDevicesEnv,
			errors.New("max id invalid")))
		return deviceIDs
	}
	for deviceID := minDevID; deviceID <= maxDevID; deviceID++ {
		deviceIDs = append(deviceIDs, deviceID)
	}
	return deviceIDs
}

func getDeviceIDsByCommaMinusStyle(devices, containerID string) []int {
	var deviceIDs []int
	devList := strings.Split(devices, comma)
	for _, subDevices := range devList {
		if minusStyle(subDevices) {
			deviceIDs = append(deviceIDs, getDeviceIDsByMinusStyle(subDevices, containerID)...)
			continue
		}
		deviceIDs = append(deviceIDs, getDeviceIDsByCommaStyle(subDevices, containerID)...)
	}
	return deviceIDs
}

func filterNPUDevices(spec *oci.Spec) ([]int, error) {
	if spec.Linux == nil || spec.Linux.Resources == nil {
		return nil, errors.New("empty spec info")
	}

	const base = 10
	devIDs := make([]int, 0, sliceLen8)
	majorIDs := npuMajor()
	for _, dev := range spec.Linux.Resources.Devices {
		if dev.Minor == nil || dev.Major == nil {
			continue
		}
		if *dev.Minor > math.MaxInt32 {
			return nil, fmt.Errorf("get wrong device ID (%v)", dev.Minor)
		}
		major := strconv.FormatInt(*dev.Major, base)
		if dev.Type == charDevice && contains(majorIDs, major) {
			devIDs = append(devIDs, int(*dev.Minor))
		}
	}

	return devIDs, nil
}

func npuMajor() []string {
	npuMajorFetchCtrl.Do(func() {
		var err error
		npuMajorID, err = getNPUMajorID()
		if err != nil {
			return
		}
	})
	return npuMajorID
}

// query the MajorID of NPU devices
func getNPUMajorID() ([]string, error) {
	const (
		deviceCount   = 2
		maxSearchLine = 512
	)

	path, err := utils.CheckPath("/proc/devices")
	if err != nil {
		return nil, err
	}
	majorID := make([]string, 0, deviceCount)
	f, err := os.Open(path)
	if err != nil {
		return majorID, err
	}
	defer func() {
		err = f.Close()
		if err != nil {
			hwlog.RunLog.Error(err)
		}
	}()
	s := bufio.NewScanner(f)
	count := 0
	for s.Scan() {
		// prevent from searching too many lines
		if count > maxSearchLine {
			break
		}
		count++
		text := s.Text()
		matched, err := regexp.MatchString("^[0-9]{1,3}\\s[v]?devdrv-cdev$", text)
		if err != nil {
			return majorID, err
		}
		if !matched {
			continue
		}
		fields := strings.Fields(text)
		majorID = append(majorID, fields[0])
	}
	return majorID, nil
}

func contains(slice []string, target string) bool {
	for _, v := range slice {
		if v == target {
			return true
		}
	}
	return false
}

func getContainerValidSpec(containerObj containerd.Container, ctx context.Context) (*oci.Spec, error) {
	spec, err := containerObj.Spec(ctx)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get container spec:%v", err)
		return nil, err
	}
	if spec.Linux == nil || spec.Linux.Resources == nil || len(spec.Linux.Resources.Devices) > maxDevicesNum {
		hwlog.RunLog.Errorf("devices in container is too much (%v) or empty", maxDevicesNum)
		return nil, err
	}
	if spec.Process == nil || len(spec.Process.Env) > maxEnvNum {
		hwlog.RunLog.Errorf("env in container is too much (%v) or empty", maxEnvNum)
		return nil, err
	}
	return spec, nil
}
