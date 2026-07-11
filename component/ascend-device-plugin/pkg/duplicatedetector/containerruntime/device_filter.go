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

package containerruntime

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

	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
)

const (
	formatIntBase = 10
)

var (
	npuMajorFetchCtrl sync.Once
	npuMajorID        sets.String
)

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

// getNPUMajorID query the MajorID of NPU devices
func getNPUMajorID() (sets.String, error) {
	const (
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
	if err := s.Err(); err != nil {
		return majorID, err
	}
	return majorID, nil
}

// filterNPUDevices filters NPU devices from container detail
func filterNPUDevices(spec *oci.Spec) []int {
	if spec == nil || spec.Linux == nil || spec.Linux.Resources == nil {
		return nil
	}
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
