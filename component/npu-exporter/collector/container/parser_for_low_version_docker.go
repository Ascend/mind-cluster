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

// Package container for monitoring containers' npu allocation
package container

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"ascend-common/common-utils/utils"
	"ascend-common/devmanager/common"
	"huawei.com/npu-exporter/v6/utils/logger"
)

const (
	procMountInfoColSep              = " "
	cgroupControllerDevices          = "devices"
	expectSystemdCgroupPathCols      = 3
	expectProcMountInfoColNum        = 10
	systemdSliceHierarchySep         = "-"
	suffixSlice                      = ".slice"
	suffixScope                      = ".scope"
	defaultSlice                     = "system.slice"
	devicesList                      = "devices.list"
	expectDevicesListColNum          = 3
	expectDeviceIDNum                = 2
	cgroupIndex                      = 4
	mountPointIdx                    = 3
	cgroupPrePath                    = 1
	cgroupSuffixPath                 = 2
	excludePermissionsForDevicesList = 333
)

var (
	// errUnknownCgroupsPathType cgroups path format not recognized
	errUnknownCgroupsPathType = errors.New("unknown cgroupsPath type")
	// errParseFail parsing devices.list fail
	errParseFail = errors.New("parsing fail")
	// errNoCgroupController no such cgroup controller
	errNoCgroupController = errors.New("no cgroup controller")
	// errNoCgroupHierarchy cgroup path not found
	errNoCgroupHierarchy = errors.New("no cgroup hierarchy")

	procMountInfoGet sync.Once
	procMountInfo    string
)

func (operator *RuntimeOperatorTool) initDockerClient() bool {
	dcli := createDockerCli()
	version, err := dcli.getDockerVersion()
	if err != nil {
		logger.Warnf("cannot get docker version info by docker api,will use oci client to get container info, "+
			"err: %v", err)
		return false
	}
	isLowerVersion := isLowerDockerVersion(version)
	logger.Debugf("current docker version: %v, isLowerVersion: %v", version, isLowerVersion)
	// use lower docker version
	if isLowerVersion {
		logger.Infof("docker is %#v version. use http method to get container info", version)
		operator.dockerCli = dcli
		operator.LowerDockerVersion = isLowerVersion
		return true
	}
	return false
}

func (dp *DevicesParser) parseDevicesForLowDockerVersion(c *CommonContainer, rs chan<- DevicesInfo) error {
	if rs == nil {
		return errors.New("empty result channel")
	}
	if c == nil {
		return errors.New("container info is nil")
	}

	deviceInfo := DevicesInfo{}
	defer func(di *DevicesInfo) {
		rs <- *di
	}(&deviceInfo)
	if len(c.Id) > maxCgroupPath {
		return fmt.Errorf("the containerId (%s) is too long", c.Id)
	}
	p, err := dp.RuntimeOperator.CgroupPath(c.Id)
	if err != nil {
		return contactError(err, fmt.Sprintf("getting cgroup path of container(%#v) fail", c.Id))
	}

	p, err = GetCgroupPath(cgroupControllerDevices, p)
	if err != nil {
		return contactError(err, "parsing cgroup path from spec fail")
	}
	devicesIDs, hasAscend, err := ScanForAscendDevices(filepath.Join(p, devicesList), c.Id)
	logger.Debugf("filter npu devices %#v in container (%s)", devicesIDs, c.Id)
	if errors.Is(err, errNoCgroupHierarchy) {
		return nil
	} else if err != nil {
		return contactError(err, fmt.Sprintf("parsing Ascend devices of container %s fail", c.Id))
	}

	if !hasAscend {
		return nil
	}
	if deviceInfo, err = makeUpDeviceInfo(c); err == nil {
		deviceInfo.Devices = devicesIDs
		return nil
	}
	logger.Error(err)
	return err
}

// GetCgroupPath the method of caculate cgroup path of device.list
func GetCgroupPath(controller, specCgroupsPath string) (string, error) {
	devicesControllerPath, err := getCgroupControllerPath(controller)
	if err != nil {
		return "", contactError(err, "getting mount point of cgroup devices subsystem fail")
	}

	hierarchy, err := toCgroupHierarchy(specCgroupsPath)
	if err != nil {
		return "", contactError(err, "parsing cgroups path of spec to cgroup hierarchy fail")
	}

	return filepath.Join(devicesControllerPath, hierarchy), nil
}

func getCgroupControllerPath(controller string) (string, error) {
	procMountInfoGet.Do(func() {
		pid := os.Getpid()
		procMountInfo = "/proc/" + strconv.Itoa(pid) + "/mountinfo"
	})
	path, err := utils.CheckPath(procMountInfo)
	if err != nil {
		return "", err
	}
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() {
		err = f.Close()
		if err != nil {
			logger.Error(err)
		}
	}()

	// parsing the /proc/self/mountinfo file content to find the mount point of specified
	// cgroup subsystem.
	// the format of the file is described in proc man page.
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		split := strings.Split(scanner.Text(), procMountInfoColSep)
		l := len(split)
		if l < expectProcMountInfoColNum {
			return "", contactError(errParseFail,
				fmt.Sprintf("mount info record has less than %d columns", expectProcMountInfoColNum))
		}

		// finding cgroup mount point, ignore others
		// if text is
		// 		434 425 0:23 / /sys/fs/cgroup/devices rw,nosuid,nodev,noexec,relatime shared:213 - cgroup cgroup rw,devices
		// then return /sys/fs/cgroup/devices
		if split[l-mountPointIdx] != "cgroup" {
			continue
		}

		// finding the specified cgroup controller
		for _, opt := range strings.Split(split[l-1], ",") {
			if opt == controller {
				// returns the path of specified cgroup controller in fs
				return split[cgroupIndex], nil
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return "", errNoCgroupController
}

// ScanForAscendDevices scan ascend devices from device.list file
func ScanForAscendDevices(devicesListFile string, id string) ([]int, bool, error) {
	minorNumbers := make([]int, 0, sliceLen8)
	majorIDs := npuMajor()

	// full devicesListFile path like:
	// /sys/fs/cgroup/devices/kubeepods.slice/kubeepods-besteffort.slice/kube***.slice/docker-<id>.scope/devices.list
	logger.Debugf("majorIDs:%v,	id:%v,	devicesList:%v", majorIDs, id, devicesListFile)
	if len(majorIDs) == 0 {
		return nil, false, fmt.Errorf("majorID is null")
	}

	f, err := os.Open(devicesListFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, errNoCgroupHierarchy
		}
		return nil, false, contactError(err, fmt.Sprintf("error while opening devices cgroup file %q",
			utils.MaskPrefix(strings.TrimPrefix(devicesListFile, unixPrefix+"://"))))
	}
	defer func() {
		err = f.Close()
		if err != nil {
			logger.Error(err)
		}
	}()
	_, err = utils.CheckPath(devicesListFile)
	if err != nil {
		return nil, false, err
	}

	s := bufio.NewScanner(f)
	// sample file content:
	// 	c 1:5 rwm
	// 	c 1:3 rwm
	// 	c 237:0 rwm
	// 	c 234:0 rwm
	// 	c 236:0 rwm
	// 	c 236:1 rw
	// 	c 236:2 rw
	// 	c 236:3 rw
	// 	c 236:4 rw
	// 	c *:* m
	// 	b *:* m
	for s.Scan() {
		text := s.Text()
		fields := strings.Fields(text)
		if len(fields) != expectDevicesListColNum {
			return nil, false, fmt.Errorf("cgroup entry %q must have three whitespace-separated fields", text)
		}

		majorMinor := strings.Split(fields[1], ":")
		if len(majorMinor) != expectDeviceIDNum {
			return nil, false, fmt.Errorf("second field of cgroup entry %q should have one colon", text)
		}

		if fields[0] != charDevice || !contains(majorIDs, majorMinor[0]) {
			continue
		}
		if majorMinor[1] == "*" {
			return nil, false, nil
		}
		minorNumber, err := strconv.Atoi(majorMinor[1])
		if err != nil {
			return nil, false, fmt.Errorf("cgroup entry %q: minor number is not integer", text)
		}
		// the max NPU cards supported number is 64
		if minorNumber < common.HiAIMaxCardNum {
			minorNumbers = append(minorNumbers, minorNumber)
		}
	}

	return minorNumbers, len(minorNumbers) > 0, nil
}

func toCgroupHierarchy(cgroupsPath string) (string, error) {
	// /kubepods-besteffort-pod****.slice/containerID
	// kubepods-besteffort-pod****.slice:docker:containerID
	var hierarchy string
	if strings.HasPrefix(cgroupsPath, "/") {
		// as cgroupfs
		hierarchy = cgroupsPath
	} else if strings.ContainsRune(cgroupsPath, ':') {
		// as systemd cgroup
		hierarchy = parseSystemdCgroup(cgroupsPath)
	} else {
		return "", errUnknownCgroupsPathType
	}
	if hierarchy == "" {
		return "", contactError(errParseFail, fmt.Sprintf("failed to parse cgroupsPath value %s", cgroupsPath))
	}
	return hierarchy, nil
}

func parseSystemdCgroup(cgroup string) string {
	// kubepods-besteffort-pod****.slice:docker:containerID
	pathsArr := strings.Split(cgroup, ":")
	if len(pathsArr) != expectSystemdCgroupPathCols {
		logger.Error("systemd cgroup path must have three parts separated by colon")
		return ""
	}

	slicePath := parseSlice(pathsArr[0])
	if slicePath == "" {
		logger.Error("failed to parse the slice part of the cgroupsPath")
		return ""
	}
	return filepath.Join(slicePath, getUnit(pathsArr[cgroupPrePath], pathsArr[cgroupSuffixPath]))
}

func parseSlice(slice string) string {
	if slice == "" {
		return defaultSlice
	}

	if len(slice) < len(suffixSlice) || !strings.HasSuffix(slice, suffixSlice) || strings.ContainsRune(slice, '/') {
		logger.Errorf("invalid slice %s when parsing slice part of systemd cgroup path", slice)
		return ""
	}

	sliceMain := strings.TrimSuffix(slice, suffixSlice)
	if sliceMain == systemdSliceHierarchySep {
		return "/"
	}

	b := new(strings.Builder)
	prefix := ""
	for _, part := range strings.Split(sliceMain, systemdSliceHierarchySep) {
		if part == "" {
			logger.Errorf("slice %s contains invalid content of continuous double dashes", slice)
			return ""
		}
		_, err := b.WriteRune('/')
		_, err = b.WriteString(prefix)
		_, err = b.WriteString(part)
		_, err = b.WriteString(suffixSlice)
		if err != nil {
			return "" // err is always nil
		}
		prefix += part + "-"
	}

	return b.String()
}

func getUnit(prefix, name string) string {
	if strings.HasSuffix(name, suffixSlice) {
		return name
	}
	return prefix + "-" + name + suffixScope
}

// IsLowerDockerVersion return cgroup path for lower docker version
func (operator *RuntimeOperatorTool) IsLowerDockerVersion() bool {
	return operator.LowerDockerVersion && operator.EndpointType == EndpointTypeDockerd && operator.dockerCli != nil
}

// CgroupPath return tue cgroup path from spec of specified container
func (operator *RuntimeOperatorTool) CgroupPath(id string) (string, error) {
	return getCgroupPathForLowerDocker(operator, id)
}

func getCgroupPathForLowerDocker(operator *RuntimeOperatorTool, containerID string) (string, error) {
	rs, err := operator.inspectContainer(containerID)
	if err != nil {
		return "", err
	}
	var cgroupPath string
	// if rs.HostConfig.CgroupParent is /kubepods-besteffort-pod****.slice
	// then return /kubepods-besteffort-pod****.slice/containerID
	if strings.HasPrefix(rs.HostConfig.CgroupParent, "/") {
		cgroupPath = fmt.Sprintf("%s/%s", rs.HostConfig.CgroupParent, containerID)
		return cgroupPath, nil
	}
	// if rs.HostConfig.CgroupParent is kubepods-besteffort-pod****.slice
	// then return kubepods-besteffort-pod****.slice:docker:containerID
	cgroupPath = fmt.Sprintf("%s:docker:%s", rs.HostConfig.CgroupParent, containerID)
	return cgroupPath, nil
}

func isLowerDockerVersion(version string) bool {
	if version == "" {
		logger.Info("docker version is empty and set lower version to false")
		return false
	}

	if strings.HasPrefix(version, "1.") {
		return true
	}
	return false
}
