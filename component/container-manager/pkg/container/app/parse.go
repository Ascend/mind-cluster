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

// Package app function for parsing container used devices
package app

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
	"github.com/containerd/containerd/api/services/tasks/v1"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/containerd/oci"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
)

const (
	sliceLen8      = 8
	charDevice     = "c"
	maxDevicesNum  = 100000
	maxEnvNum      = 10000
	minus          = "-"
	comma          = ","
	ascendEnvPart  = 2
	maxEnvLength   = 1024
	deviceCount    = 2
	maxSearchLine  = 512
	base           = 10
	deviceListFile = "/proc/devices"
)

var (
	minusStyle = func(s string) bool {
		return strings.Contains(s, minus)
	}
	commaMinusStyle = func(s string) bool {
		return strings.Contains(s, minus) && strings.Contains(s, comma)
	}
	ascendStyle = func(s string) bool {
		return strings.Contains(s, api.Ascend)
	}

	npuMajorFetchCtrl sync.Once
	npuMajorID        []string
	majorIdRegex      = regexp.MustCompile("^[0-9]{1,3}\\s[v]?devdrv-cdev$")
)

func (cm *CtrCtl) updateCtrRelatedInfo() error {
	nss, err := cm.client.NamespaceService().List(context.Background())
	if err != nil {
		hwlog.RunLog.Errorf("failed to get namespace list, error: %v", err)
		return errors.New("failed to get namespace list")
	}
	var ctrIds []string
	for _, ns := range nss {
		ctx := namespaces.WithNamespace(context.Background(), ns)
		// list running containers
		taskList, err := cm.client.TaskService().List(ctx, &tasks.ListTasksRequest{})
		if err != nil {
			hwlog.RunLog.Errorf("failed to get container list by ns %s, error: %v", ns, err)
			continue
		}
		if len(taskList.Tasks) == 0 {
			continue
		}
		for _, taskInfo := range taskList.Tasks {
			containerObj, err := cm.client.LoadContainer(ctx, taskInfo.ID)
			if err != nil {
				hwlog.RunLog.Errorf("failed to load container %s, error: %v", taskInfo.ID, err)
				continue
			}
			usedDevs, err := getUsedDevs(containerObj, ctx)
			if err != nil {
				hwlog.RunLog.Errorf("get container %s used devs failed: %v", containerObj.ID(), err)
				continue
			}
			if len(usedDevs) == 0 {
				// only ctr of used dev need save to cache
				hwlog.RunLog.Debugf("container %s not used devs", containerObj.ID())
				continue
			} else {
				hwlog.RunLog.Debugf("container %s used devs: %v", containerObj.ID(), usedDevs)
			}
			ctrIds = append(ctrIds, containerObj.ID())
			cm.setCtrRelatedInfo(containerObj.ID(), ns, usedDevs)
		}
	}
	cm.removeDeletedCtr(ctrIds)
	return nil
}

func getUsedDevs(containerObj containerd.Container, ctx context.Context) ([]int32, error) {
	spec, err := getCtrValidSpec(containerObj, ctx)
	if err != nil {
		return nil, fmt.Errorf("get container %s valid spec failed, error: %v", containerObj.ID(), err)
	}
	envs := spec.Process.Env
	// for containerd, env with the same name will be appended directly and will not be overwritten.
	// To avoid the presence of environment variables ASCEND_VISIBLE_DEVICES in the image, iterate from back to front
	for i := len(envs) - 1; i >= 0; i-- {
		env := envs[i]
		if strings.Contains(env, api.AscendDeviceInfo) {
			usedDevs, err := getUsedDevsWithAscendRuntime(env)
			if err != nil {
				return nil, fmt.Errorf("parse env %s failed, error: %v", api.AscendDeviceInfo, err)
			}
			return usedDevs, nil
		}
	}
	hwlog.RunLog.Warnf("get used devs by env %s failed, not used ascend docker runtime", api.AscendDeviceInfo)
	usedDevs, err := getUsedDevsWithoutAscendRuntime(spec)
	if err != nil {
		return nil, fmt.Errorf("get container %s device ids failed, error: %v", containerObj.ID(), err)
	}
	return usedDevs, nil
}

func getCtrValidSpec(containerObj containerd.Container, ctx context.Context) (*oci.Spec, error) {
	spec, err := containerObj.Spec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get container spec:%v", err)
	}
	if spec.Linux == nil || spec.Linux.Resources == nil || len(spec.Linux.Resources.Devices) > maxDevicesNum {
		return nil, fmt.Errorf("devices in container is too much (%v) or empty", maxDevicesNum)
	}
	if spec.Process == nil || len(spec.Process.Env) > maxEnvNum {
		return nil, fmt.Errorf("env in container is too much (%v) or empty", maxEnvNum)
	}
	return spec, nil
}

func getUsedDevsWithAscendRuntime(env string) ([]int32, error) {
	devInfo := strings.Split(env, "=")
	if len(devInfo) != ascendEnvPart {
		return nil, fmt.Errorf("env %s is invalid", devInfo)
	}
	idsStr := devInfo[1]
	if len(idsStr) > maxEnvLength {
		return []int32{}, errors.New("env length invalid")
	}
	// parse 4 env value format
	if ascendStyle(idsStr) { // eg. Ascend910-0, Ascend-1
		return getDevIdsByAscendStyle(idsStr)
	}
	if commaMinusStyle(idsStr) { // eg. 0-2,4
		return getDevIdsByCommaMinusStyle(idsStr)
	}
	if minusStyle(idsStr) { // eg. 0-3
		return getDevIdsByMinusStyle(idsStr)
	}
	return getDevIdsByCommaStyle(idsStr) // eg. 0,1,2,3
}

func getDevIdsByAscendStyle(idsStr string) ([]int32, error) {
	devList := strings.Split(idsStr, comma)
	ids := make([]int32, 0, len(devList))
	for _, ascendId := range devList {
		deviceName := strings.Split(ascendId, minus)
		if len(deviceName) != ascendEnvPart {
			return ids, errors.New("ascend style env format error")
		}
		if !strings.HasPrefix(deviceName[0], api.Ascend) {
			return ids, fmt.Errorf("ascend style env must start with %s", api.Ascend)
		}
		id, err := strconv.Atoi(deviceName[1])
		if err != nil {
			return ids, errors.New("dev id cannot convert to int")
		}
		ids = append(ids, int32(id))
	}
	return ids, nil
}

func getDevIdsByCommaMinusStyle(idsStr string) ([]int32, error) {
	var ids []int32
	devList := strings.Split(idsStr, comma)
	for _, minusId := range devList {
		if minusStyle(minusId) {
			minusStyleIds, err := getDevIdsByMinusStyle(minusId)
			if err != nil {
				return ids, err
			}
			ids = append(ids, minusStyleIds...)
			continue
		}
		commaStyleIds, err := getDevIdsByCommaStyle(minusId)
		if err != nil {
			return ids, err
		}
		ids = append(ids, commaStyleIds...)
	}
	return ids, nil
}

func getDevIdsByMinusStyle(idsStr string) ([]int32, error) {
	ids := make([]int32, 0)
	idRange := strings.Split(idsStr, minus)
	if len(idRange) != ascendEnvPart {
		return ids, errors.New("minus style env format error")
	}
	minId, err := strconv.Atoi(idRange[0])
	if err != nil {
		return ids, errors.New("min dev id cannot convert to int")
	}
	maxId, err := strconv.Atoi(idRange[1])
	if err != nil {
		return ids, errors.New("max dev id cannot convert to int")
	}
	if minId > maxId {
		return ids, errors.New("min id bigger than max id")
	}
	if maxId > math.MaxInt16 {
		return ids, errors.New("invalid max id")
	}
	for id := minId; id <= maxId; id++ {
		ids = append(ids, int32(id))
	}
	return ids, nil
}

func getDevIdsByCommaStyle(idsStr string) ([]int32, error) {
	devList := strings.Split(idsStr, comma)
	ids := make([]int32, 0, len(devList))
	for _, devID := range devList {
		id, err := strconv.Atoi(devID)
		if err != nil {
			return ids, errors.New("dev id cannot convert to int")
		}
		ids = append(ids, int32(id))
	}
	return ids, nil
}

func getUsedDevsWithoutAscendRuntime(spec *oci.Spec) ([]int32, error) {
	if spec.Linux == nil || spec.Linux.Resources == nil {
		return nil, errors.New("empty spec info")
	}

	phyIds := make([]int32, 0, sliceLen8)
	majorIDs := npuMajor()
	for _, dev := range spec.Linux.Resources.Devices {
		if dev.Minor == nil || dev.Major == nil {
			continue
		}
		if *dev.Minor > math.MaxInt32 {
			return nil, fmt.Errorf("get wrong device ID (%v)", dev.Minor)
		}
		major := strconv.FormatInt(*dev.Major, base)
		if dev.Type == charDevice && utils.Contains(majorIDs, major) {
			phyIds = append(phyIds, int32(*dev.Minor))
		}
	}
	return phyIds, nil
}

func npuMajor() []string {
	npuMajorFetchCtrl.Do(func() {
		var err error
		npuMajorID, err = getNPUMajorId()
		if err != nil {
			return
		}
	})
	return npuMajorID
}

func getNPUMajorId() ([]string, error) {
	path, err := utils.CheckPath(deviceListFile)
	if err != nil {
		return nil, err
	}
	majorId := make([]string, 0, deviceCount)
	f, err := os.Open(path)
	if err != nil {
		return majorId, err
	}
	defer func() {
		if err = f.Close(); err != nil {
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
		if !majorIdRegex.MatchString(text) {
			continue
		}
		fields := strings.Fields(text)
		majorId = append(majorId, fields[0])
	}
	return majorId, nil
}
