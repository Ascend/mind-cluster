/* Copyright(C) 2024. Huawei Technologies Co.,Ltd. All rights reserved.
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

// package parser provides functions to parse device information

package parser

import (
	"bufio"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/containerd/containerd/oci"
	"k8s.io/apimachinery/pkg/util/sets"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
)

const (
	maxEnvLength = 1024

	comma          = ","
	minus          = "-"
	ascend         = "Ascend"
	envSliceLen    = 2
	deviceSliceLen = 2
	formatIntBase  = 10
)

var (
	npuMajorFetchCtrl sync.Once
	npuMajorID        sets.String
)

// ParseAscendDeviceInfo parses the AscendDeviceInfo environment variable
func ParseAscendDeviceInfo(env, containerID string) []int {
	parts := strings.SplitN(env, "=", envSliceLen)
	if len(parts) != envSliceLen {
		hwlog.RunLog.Warnf("Invalid %s format in container %s", api.AscendDeviceInfo, containerID)
		return nil
	}

	devicesStr := parts[1]
	if len(devicesStr) > maxEnvLength {
		hwlog.RunLog.Warnf("%s value too long in container %s", api.AscendDeviceInfo, containerID)
		return nil
	}

	return parseDeviceIDs(devicesStr, containerID)
}

// parseDeviceIDs parses device IDs from various formats
func parseDeviceIDs(devices, containerID string) []int {
	// Handle Ascend style: Ascend910-0,Ascend910-1
	if strings.Contains(devices, ascend) {
		return parseAscendStyle(devices, containerID)
	}
	// Handle comma-minus style: 0-1,3
	if strings.Contains(devices, comma) && strings.Contains(devices, minus) {
		return parseCommaMinusStyle(devices, containerID)
	}
	// Handle minus style: 0-3
	if strings.Contains(devices, minus) {
		return parseMinusStyle(devices, containerID)
	}
	// Handle comma style: 0,1,2
	return parseCommaStyle(devices, containerID)
}

func parseCommaStyle(devices, containerID string) []int {
	devList := strings.Split(devices, comma)
	deviceIDs := make([]int, 0, len(devList))
	for _, devID := range devList {
		id, err := strconv.Atoi(strings.TrimSpace(devID))
		if err != nil {
			hwlog.RunLog.Warnf("Invalid device ID %s in container %s: %v", devID, containerID, err)
			continue
		}
		deviceIDs = append(deviceIDs, id)
	}
	return deviceIDs
}

func parseMinusStyle(devices, containerID string) []int {
	deviceIDs := make([]int, 0)
	rangeParts := strings.Split(devices, minus)
	if len(rangeParts) != deviceSliceLen {
		hwlog.RunLog.Warnf("Invalid device range %s in container %s", devices, containerID)
		return deviceIDs
	}

	start, err := strconv.Atoi(strings.TrimSpace(rangeParts[0]))
	if err != nil {
		hwlog.RunLog.Warnf("Invalid start device ID %s in container %s: %v", rangeParts[0], containerID, err)
		return deviceIDs
	}

	end, err := strconv.Atoi(strings.TrimSpace(rangeParts[1]))
	if err != nil {
		hwlog.RunLog.Warnf("Invalid end device ID %s in container %s: %v", rangeParts[1], containerID, err)
		return deviceIDs
	}

	if start > end {
		hwlog.RunLog.Warnf("Invalid device range %d-%d in container %s: start > end", start, end, containerID)
		return deviceIDs
	}

	if end > math.MaxInt16 {
		hwlog.RunLog.Warnf("End device ID %d exceeds maximum in container %s", end, containerID)
		return deviceIDs
	}

	for i := start; i <= end; i++ {
		deviceIDs = append(deviceIDs, i)
	}
	return deviceIDs
}

func parseCommaMinusStyle(devices, containerID string) []int {
	deviceIDs := make([]int, 0)
	parts := strings.Split(devices, comma)
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.Contains(part, minus) {
			deviceIDs = append(deviceIDs, parseMinusStyle(part, containerID)...)
		} else {
			id, err := strconv.Atoi(part)
			if err != nil {
				hwlog.RunLog.Warnf("Invalid device ID %s in container %s: %v", part, containerID, err)
				continue
			}
			deviceIDs = append(deviceIDs, id)
		}
	}
	return deviceIDs
}

func parseAscendStyle(devices, containerID string) []int {
	deviceIDs := make([]int, 0)
	parts := strings.Split(devices, comma)
	for _, part := range parts {
		part = strings.TrimSpace(part)
		// Format: Ascend910-0 or Ascend310P-1
		if !strings.Contains(part, minus) {
			hwlog.RunLog.Warnf("Invalid Ascend device format %s in container %s", part, containerID)
			continue
		}
		deviceParts := strings.Split(part, minus)
		if len(deviceParts) != deviceSliceLen {
			hwlog.RunLog.Warnf("Invalid Ascend device format %s in container %s", part, containerID)
			continue
		}
		deviceID, err := strconv.Atoi(deviceParts[1])
		if err != nil {
			hwlog.RunLog.Warnf("Invalid device ID %s in container %s: %v", deviceParts[1], containerID, err)
			continue
		}
		deviceIDs = append(deviceIDs, deviceID)
	}
	return deviceIDs
}

func npuMajor() sets.String {
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
func getNPUMajorID() (sets.String, error) {
	const (
		deviceCount   = 2
		maxSearchLine = 512
	)

	path, err := utils.CheckPath("/proc/devices")
	if err != nil {
		return nil, err
	}
	majorID := sets.NewString()
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
		majorID.Insert(fields[0])
	}
	return majorID, nil
}

// FilterNPUDevices filters NPU devices from container detail
func FilterNPUDevices(spec *oci.Spec) []int {
	devIDs := make([]int, 0)
	majorIDs := npuMajor()
	for _, dev := range spec.Linux.Resources.Devices {
		if dev.Minor == nil || dev.Major == nil {
			// do not monitor privileged container
			continue
		}
		if *dev.Minor > math.MaxInt32 {
			return nil
		}
		major := strconv.FormatInt(*dev.Major, formatIntBase)
		if dev.Type == "c" && majorIDs.Has(major) {
			devIDs = append(devIDs, int(*dev.Minor))
		}
	}
	return devIDs
}
