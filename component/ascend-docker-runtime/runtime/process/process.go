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

// Package process does what ascend-docker-runtime is supposed to do before runc being executed.
package process

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/containerd/containerd/oci"
	"github.com/opencontainers/runtime-spec/specs-go"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"ascend-docker-runtime/mindxcheckutils"
	"ascend-docker-runtime/runtime/common"
	"ascend-docker-runtime/runtime/dcmi"
	"ascend-docker-runtime/runtime/grus"
)

const (
	runLogPath          = api.RunTimeRunLogPath
	hookDefaultFilePath = "/usr/local/bin/ascend-docker-hook"
	// MaxCommandLength is the max length of command.
	MaxCommandLength = 65535
	hookCli          = "ascend-docker-hook"
	destroyHookCli   = "ascend-docker-destroy"
	envLength        = 2
	kvPairSize       = 2
	borderNum        = 2

	// ENV for device-plugin to identify ascend-docker-runtime
	useAscendDocker      = api.AscendDockerRuntimeEnv + "=True"
	ascendVisibleDevices = api.AscendVisibleDevicesEnv
	ascendRuntimeOptions = api.AscendRuntimeOptionsEnv
	ldLibraryPathKey     = "LD_LIBRARY_PATH"

	// void indicates that the NPU card does not need to be mounted
	void = "void"

	mountByRuntimeForDPEnv = "MOUNT_BY_RUNTIME_FOR_DP"

	dockerSockHostPath      = "/run/docker.sock"
	dockerSockContainerPath = "/run/docker.sock"
	dockerDirHostPath       = "/run/docker"
	dockerDirContainerPath  = "/run/docker"
	containerdHostPath      = "/run/containerd"
	containerdContainerPath = "/run/containerd"
)

type dpMountConfig struct {
	hostPath      string
	containerPath string
	readOnly      bool
}

var dpMountConfigs = []dpMountConfig{
	{dockerSockHostPath, dockerSockContainerPath, true},
	{dockerDirHostPath, dockerDirContainerPath, true},
	{containerdHostPath, containerdContainerPath, true},
}

var (
	hookCliPath     = hookCli
	hookDefaultFile = hookDefaultFilePath
	deviceRegx      = fmt.Sprintf(`^(?:%s(%s|%s|%s|%s)|%s)-(\d+)$`, api.Ascend, api.Ascend910No,
		api.Ascend310BNo, api.Ascend310PNo, api.Ascend310No, api.NPULowerCase)

	// Device lists for different chip types
	ascend910A5ManagerDevices = []string{hisiHdc}
	defaultManagerDevices     = []string{devmmSvm, hisiHdc}

	// managerDevicesMap maps device types to their corresponding manager devices
	managerDevicesMap = map[string][]string{
		Ascend910A5: ascend910A5ManagerDevices,
	}
	// ascendDriverLibPaths contains the Ascend driver library paths to be added to LD_LIBRARY_PATH
	ascendDriverLibPaths = []string{
		"/usr/local/Ascend/driver/lib64/common",
		"/usr/local/Ascend/driver/lib64/driver",
	}
)

const (
	// Atlas200ISoc Product name
	Atlas200ISoc = "Atlas 200I SoC A1"
	// Atlas200 Product name
	Atlas200 = "Atlas 200 Model 3000"
	// Ascend310 ascend 310 chip
	Ascend310 = api.Ascend310
	// Ascend310P ascend 310P chip
	Ascend310P = api.Ascend310P
	// Ascend310B ascend 310B chip
	Ascend310B = api.Ascend310B
	// Ascend910 ascend 910 chip
	Ascend910 = api.Ascend910
	// Ascend910A5 asecnd 910a5 chip
	Ascend910A5 = api.Ascend910A5
	ascend      = api.Ascend
	npu         = api.NPULowerCase

	devicePath           = "/dev/"
	davinciName          = "davinci"
	virtualDavinciName   = "vdavinci"
	davinciManager       = "davinci_manager"
	davinciManagerDocker = "davinci_manager_docker"
	devmmSvm             = "devmm_svm"
	hisiHdc              = "hisi_hdc"
	svm0                 = "svm0"
	tsAisle              = "ts_aisle"
	upgrade              = "upgrade"
	sys                  = "sys"
	vdec                 = "vdec"
	vpc                  = "vpc"
	pngd                 = "pngd"
	venc                 = "venc"
	dvppCmdList          = "dvpp_cmdlist"
	logDrv               = "log_drv"
	acodec               = "acodec"
	ai                   = "ai"
	ao                   = "ao"
	vo                   = "vo"
	hdmi                 = "hdmi"
	uburma               = "uburma"
	ummu                 = "ummu"
)

// GetDeviceTypeByChipName get device type by chipName
func GetDeviceTypeByChipName(chipName string) string {
	if strings.Contains(chipName, api.Ascend310BNo) {
		return Ascend310B
	}
	if strings.Contains(chipName, api.Ascend310PNo) {
		return Ascend310P
	}
	// This uses HasPrefix because A5 chips have specific prefix pattern: Ascend950XX
	if strings.HasPrefix(chipName, api.Ascend910A5Prefix) {
		return Ascend910A5
	}
	if strings.Contains(chipName, api.Ascend310No) {
		return Ascend310
	}
	if strings.Contains(chipName, api.Ascend910No) {
		return Ascend910
	}
	return ""
}

func getArgs() (*common.Args, error) {
	args := &common.Args{}

	for i, param := range os.Args {
		if param == "--bundle" || param == "-b" {
			if len(os.Args)-i <= 1 {
				return nil, fmt.Errorf("bundle option needs an argument")
			}
			args.Bundle = os.Args[i+1]
		} else if param == "--image-path" {
			if len(os.Args)-i <= 1 {
				return nil, fmt.Errorf("image-path option needs an argument")
			}
			args.CkptPath = os.Args[i+1]
		} else if param == "--root" {
			if len(os.Args)-i <= 1 {
				return nil, fmt.Errorf("root option needs an argument")
			}
			args.Root = os.Args[i+1]
		} else if param == "create" || param == "start" || param == "checkpoint" || param == "resume" {
			args.Cmd = param
		}
	}
	args.ContainerID = os.Args[len(os.Args)-1]
	return args, nil
}

func updateRoot(a *common.Args) {
	if a.Root != "" {
		return
	}
	// container engine not support namespace
	if _, err := os.Stat(common.ContainerdRunRoot + "/" + a.ContainerID); err == nil {
		a.Root = common.ContainerdRunRoot
		return
	}

	a.Root = fmt.Sprintf("%s/default", common.ContainerdRunRoot)
	_, err := os.Stat(a.Root + "/" + a.ContainerID)
	if os.IsNotExist(err) {
		a.Root = fmt.Sprintf("%s/k8s.io", common.ContainerdRunRoot)
	}
}

// InitLogModule initializes some logging configuration.
func InitLogModule(ctx context.Context) error {
	const backups = 2
	const logMaxAge = 365
	const fileMaxSize = 2
	runLogConfig := hwlog.LogConfig{
		LogFileName: runLogPath,
		LogLevel:    0,
		MaxBackups:  backups,
		MaxAge:      logMaxAge,
		OnlyToFile:  true,
		FileMaxSize: fileMaxSize,
	}
	if err := hwlog.InitRunLogger(&runLogConfig, ctx); err != nil {
		fmt.Printf("hwlog init failed, error is %v", err)
		return err
	}
	return nil
}

func addAscendDockerEnv(spec *specs.Spec) {
	if spec == nil || spec.Process == nil || spec.Process.Env == nil {
		return
	}
	spec.Process.Env = append(spec.Process.Env, useAscendDocker)
}

func addAscendLibraryPath(spec *specs.Spec) {
	if spec == nil || spec.Process == nil || spec.Process.Env == nil {
		return
	}

	existingPaths := collectExistingLibPaths()
	if len(existingPaths) == 0 {
		return
	}

	for i := len(spec.Process.Env) - 1; i >= 0; i-- {
		env := spec.Process.Env[i]
		parts := strings.SplitN(env, "=", kvPairSize)
		if len(parts) != kvPairSize || parts[0] != ldLibraryPathKey {
			continue
		}
		newPaths := filterNewLibPaths(existingPaths, parts[1])
		if len(newPaths) == 0 {
			return
		}
		spec.Process.Env[i] = ldLibraryPathKey + "=" + strings.Join(newPaths, ":") + ":" + parts[1]
		return
	}

	ascendLibPath := strings.Join(existingPaths, ":")
	spec.Process.Env = append(spec.Process.Env, ldLibraryPathKey+"="+ascendLibPath)
}

func collectExistingLibPaths() []string {
	var paths []string
	for _, libPath := range ascendDriverLibPaths {
		if _, err := os.Stat(libPath); err == nil {
			paths = append(paths, libPath)
		}
	}
	return paths
}

func filterNewLibPaths(paths []string, existingLD string) []string {
	existingSet := make(map[string]struct{}, len(paths))
	for _, p := range strings.Split(existingLD, ":") {
		existingSet[p] = struct{}{}
	}
	filtered := make([]string, 0, len(paths))
	for _, p := range paths {
		if _, ok := existingSet[p]; !ok {
			filtered = append(filtered, p)
		}
	}
	return filtered
}

func isMountByRuntimeForDP(env []string) bool {
	for i := len(env) - 1; i >= 0; i-- {
		words := strings.SplitN(env[i], "=", kvPairSize)
		if len(words) != kvPairSize {
			continue
		}
		if words[0] == mountByRuntimeForDPEnv {
			return true
		}
	}
	return false
}

func addDPMountsToSpec(spec *specs.Spec) {
	if spec == nil {
		return
	}
	hwlog.RunLog.Warn("add device-plugin mount points")
	for _, cfg := range dpMountConfigs {
		if _, err := os.Stat(cfg.hostPath); err != nil {
			hwlog.RunLog.Warnf("dp mount host path %s does not exist, skip", cfg.hostPath)
			continue
		}

		alreadyMounted := false
		for _, m := range spec.Mounts {
			if m.Destination == cfg.containerPath {
				alreadyMounted = true
				break
			}
		}
		if alreadyMounted {
			hwlog.RunLog.Warnf("dp mount host path %s has been mounted, skip", cfg.hostPath)
			continue
		}

		options := []string{"rbind", "rprivate"}
		if cfg.readOnly {
			options = append(options, "ro")
		}

		spec.Mounts = append(spec.Mounts, specs.Mount{
			Destination: cfg.containerPath,
			Source:      cfg.hostPath,
			Type:        "bind",
			Options:     options,
		})
	}
}

func addHook(w dcmi.WorkerInterface, spec *specs.Spec, deviceIdList *[]int) error {
	if deviceIdList == nil {
		return nil
	}
	currentExecPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot get the path of docker-runtime: %v", err)
	}

	hookCliPath = path.Join(path.Dir(currentExecPath), hookCli)
	if _, err := mindxcheckutils.RealFileChecker(hookCliPath, true, false, mindxcheckutils.DefaultSize); err != nil {
		return err
	}
	if _, err = os.Stat(hookCliPath); err != nil {
		return fmt.Errorf("cannot find docker-hook executable file at %s: %v", hookCliPath, err)
	}

	if spec.Hooks == nil {
		spec.Hooks = &specs.Hooks{}
	}

	needUpdate := true
	if len(spec.Hooks.Prestart) > MaxCommandLength {
		return fmt.Errorf("too many items in Prestart ")
	}
	for _, hook := range spec.Hooks.Prestart {
		if strings.Contains(hook.Path, hookCli) {
			needUpdate = false
			break
		}
	}
	if needUpdate {
		spec.Hooks.Prestart = append(spec.Hooks.Prestart, specs.Hook{
			Path: hookCliPath,
			Args: []string{hookCliPath},
		})
	}

	if len(spec.Process.Env) > MaxCommandLength {
		return fmt.Errorf("too many items in Env ")
	}

	if strings.Contains(getValueByKey(spec.Process.Env, ascendRuntimeOptions), "VIRTUAL") {
		return nil
	}

	vdevice, err := dcmi.CreateVDevice(w, spec, *deviceIdList)
	if err != nil {
		return err
	}
	hwlog.RunLog.Infof("vnpu split done: vdevice: %v", vdevice.VdeviceID)

	if vdevice.VdeviceID != -1 {
		if err = updateEnvAndPostHook(spec, vdevice, deviceIdList); err != nil {
			return fmt.Errorf("update evn and post hook failed: %v ", err)
		}
	}

	return nil
}

func removeDuplication(devices []int) []int {
	list := make([]int, 0, len(devices))
	prev := -1

	for _, device := range devices {
		if device == prev {
			continue
		}

		list = append(list, device)
		prev = device
	}

	return list
}

func parseDevices(visibleDevices string) ([]int, error) {
	devices := make([]int, 0)

	for _, value := range strings.Split(visibleDevices, ",") {
		deviceFromValue, err := getDeviceListFromVisibleValue(value)
		if err != nil {
			hwlog.RunLog.Errorf("failed to get devices from %v value, error: %v", api.AscendVisibleDevicesEnv, err)
			return nil, err
		}
		devices = append(devices, deviceFromValue...)
	}

	sort.Slice(devices, func(i, j int) bool { return i < j })
	return removeDuplication(devices), nil
}

func getDeviceListFromVisibleValue(visibleValue string) ([]int, error) {
	maxDevice := 128
	devices := make([]int, 0)
	visibleValue = strings.TrimSpace(visibleValue)
	if strings.Contains(visibleValue, "-") {
		borders := strings.Split(visibleValue, "-")
		if len(borders) != borderNum {
			return nil, fmt.Errorf("invalid device range: %s", visibleValue)
		}

		borders[0] = strings.TrimSpace(borders[0])
		borders[1] = strings.TrimSpace(borders[1])

		left, err := strconv.Atoi(borders[0])
		if err != nil || left < 0 {
			return nil, fmt.Errorf("invalid left boarder range parameter: %s", borders[0])
		}

		right, err := strconv.Atoi(borders[1])
		if err != nil || right > maxDevice {
			return nil, fmt.Errorf("invalid right boarder range parameter: %s", borders[1])
		}

		if left > right {
			return nil, fmt.Errorf("left boarder (%d) should not be larger than the right one(%d)", left, right)
		}

		for n := left; n <= right; n++ {
			devices = append(devices, n)
		}
		return devices, nil
	}
	n, err := strconv.Atoi(visibleValue)
	if err != nil {
		return nil, fmt.Errorf("invalid single device parameter: %s", visibleValue)
	}
	devices = append(devices, n)

	return devices, nil
}

func parseAscendDevices(visibleDevices string) ([]int, error) {
	devicesList := strings.Split(visibleDevices, ",")
	devices := make([]int, 0, len(devicesList))

	for _, d := range devicesList {
		matchGroups := regexp.MustCompile(deviceRegx).FindStringSubmatch(strings.TrimSpace(d))
		if matchGroups == nil {
			return nil, fmt.Errorf("invalid device format: %s", d)
		}
		n, err := strconv.Atoi(matchGroups[2])
		if err != nil {
			return nil, fmt.Errorf("invalid device id: %s", d)
		}

		devices = append(devices, n)
	}

	sort.Slice(devices, func(i, j int) bool { return i < j })
	return removeDuplication(devices), nil
}

func getValueByKey(data []string, name string) string {
	for _, envLine := range data {
		words := strings.SplitN(envLine, "=", kvPairSize)
		if len(words) != kvPairSize {
			hwlog.RunLog.Error("environment error")
			return ""
		}

		if words[0] == name {
			return words[1]
		}
	}

	return ""
}

func getValueByDeviceKey(data []string) string {
	res := ""
	for i := len(data) - 1; i >= 0; i-- {
		words := strings.SplitN(data[i], "=", kvPairSize)
		if len(words) != kvPairSize {
			hwlog.RunLog.Error("environment error")
			return ""
		}

		if words[0] == ascendVisibleDevices {
			res = words[1]
			break
		}
	}
	if res == "" {
		hwlog.RunLog.Errorf("%v env variable is empty, will not mount any ascend device",
			api.AscendVisibleDevicesEnv)
	}
	return res
}

func getMountPath(dHostPath string, deviceType string) (string, error) {
	switch deviceType {
	case virtualDavinciName:
		vDeviceNumber := regexp.MustCompile("[0-9]+").FindAllString(dHostPath, -1)
		if len(vDeviceNumber) != 1 {
			return "", fmt.Errorf("invalid vdavinci path: %s", dHostPath)
		}
		return devicePath + davinciName + vDeviceNumber[0], nil
	case davinciManagerDocker:
		return devicePath + davinciManager, nil
	default: // do nothing
		return dHostPath, nil
	}
}

func addDeviceToSpec(spec *specs.Spec, dHostPath string, dContainerPath string) error {
	device, err := oci.DeviceFromPath(dHostPath)
	if err != nil {
		return fmt.Errorf("failed to get %s info : %v", dHostPath, err)
	}

	device.Path = dContainerPath

	spec.Linux.Devices = append(spec.Linux.Devices, *device)
	newDeviceCgroup := specs.LinuxDeviceCgroup{
		Allow:  true,
		Type:   device.Type,
		Major:  &device.Major,
		Minor:  &device.Minor,
		Access: "rwm",
	}
	spec.Linux.Resources.Devices = append(spec.Linux.Resources.Devices, newDeviceCgroup)
	return nil
}

func addAscend310BManagerDevice(spec *specs.Spec) error {
	var Ascend310BManageDevices = []string{
		svm0,
		tsAisle,
		upgrade,
		sys,
		vdec,
		vpc,
		pngd,
		venc,
		logDrv,
		acodec,
		ai,
		ao,
		vo,
		hdmi,
	}

	for _, device := range Ascend310BManageDevices {
		dPath := devicePath + device
		if err := addDeviceToSpec(spec, dPath, dPath); err != nil {
			hwlog.RunLog.Warnf("failed to add %s to spec : %v", dPath, err)
		}
	}

	davinciManagerPath := devicePath + davinciManagerDocker
	if _, err := os.Stat(davinciManagerPath); err != nil {
		hwlog.RunLog.Warnf("failed to get davinci manager docker, err: %v", err)
		davinciManagerPath = devicePath + davinciManager
		if _, err := os.Stat(davinciManagerPath); err != nil {
			return fmt.Errorf("failed to get davinci manager, err: %v", err)
		}
	}
	dContainerPath, err := getMountPath(davinciManagerPath, davinciManagerDocker)
	if err != nil {
		return fmt.Errorf("failed to get virtual davinci name : %v", err)
	}
	return addDeviceToSpec(spec, davinciManagerPath, dContainerPath)
}

// getCommonManagerDevices returns chip-specific manager device list.
func getCommonManagerDevices(devType string) []string {
	if devices, ok := managerDevicesMap[devType]; ok {
		return devices
	}
	return defaultManagerDevices
}

// addCommonManagerDevice adds common manager devices to spec based on device type
func addCommonManagerDevice(spec *specs.Spec, devType string) error {
	devices := getCommonManagerDevices(devType)
	for _, device := range devices {
		dPath := devicePath + device
		if err := addDeviceToSpec(spec, dPath, dPath); err != nil {
			return fmt.Errorf("failed to add common manage device to spec : %v", err)
		}
	}

	return nil
}

func addManagerDevice(w dcmi.WorkerInterface, spec *specs.Spec) error {
	chipName, err := w.GetChipName()
	if err != nil {
		return fmt.Errorf("get chip name error: %v", err)
	}
	devType := GetDeviceTypeByChipName(chipName)
	if devType == Ascend910A5 {
		hwlog.RunLog.Info("device type is npu")
	} else {
		hwlog.RunLog.Infof("device type is: %s", devType)
	}
	if devType != "" && devType != Ascend910A5 {
		dPath := devicePath + dvppCmdList
		if err := addDeviceToSpec(spec, dPath, dPath); err != nil {
			hwlog.RunLog.Warnf("failed to add dvpp_cmdlist to spec : %v", err)
		}
	}
	if devType == Ascend310B {
		return addAscend310BManagerDevice(spec)
	}
	dPath := devicePath + davinciManager
	if err := addDeviceToSpec(spec, dPath, dPath); err != nil {
		return fmt.Errorf("add davinci_manager to spec error: %v", err)
	}

	productType, err := w.GetProductType()
	if err != nil {
		return fmt.Errorf("parse product type error: %v", err)
	}
	hwlog.RunLog.Infof("product type is %s", productType)

	switch productType {
	// do nothing
	case Atlas200ISoc, Atlas200:
	default:
		if err = addCommonManagerDevice(spec, devType); err != nil {
			return fmt.Errorf("add common manage device error: %v", err)
		}
	}

	return nil
}

func addUBDevice(spec *specs.Spec) error {
	uburmaPath := devicePath + uburma
	if _, err := os.Stat(uburmaPath); err == nil {
		if err := addDevicesInDir(spec, uburmaPath); err != nil {
			return err
		}
		hwlog.RunLog.Infof("uburma devices exist, add uburma devices to spec")
	}

	ummuPath := devicePath + ummu
	if _, err := os.Stat(ummuPath); err == nil {
		if err := addDevicesInDir(spec, ummuPath); err != nil {
			return err
		}
		hwlog.RunLog.Infof("ummu devices exist, add ummu devices to spec")
	}
	return nil
}

func addDevicesInDir(spec *specs.Spec, dirPath string) error {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("read device dir %s err:%v", dirPath, err)
	}

	for _, entry := range entries {
		fullDevicePath := filepath.Join(dirPath, entry.Name())
		if entry.IsDir() {
			continue
		}
		if err := addDeviceToSpec(spec, fullDevicePath, fullDevicePath); err != nil {
			return fmt.Errorf("add %s to spec error: %#v", fullDevicePath, err)
		}
	}
	return nil
}

func checkVisibleDevice(spec *specs.Spec) ([]int, error) {
	if spec.Process == nil {
		return nil, errors.New("empty process info")
	}
	visibleDevices := getValueByDeviceKey(spec.Process.Env)
	if visibleDevices == "" || visibleDevices == void {
		return nil, nil
	}

	if strings.Contains(visibleDevices, ascend) || strings.Contains(visibleDevices, npu) {
		devices, err := parseAscendDevices(visibleDevices)
		if err != nil {
			return nil, fmt.Errorf("failed to parse ascend device : %v", err)
		}
		hwlog.RunLog.Infof("ascend devices is: %v", devices)
		return devices, err
	}
	devices, err := parseDevices(visibleDevices)
	if err != nil {
		return nil, fmt.Errorf("failed to parse device : %v", err)
	}
	hwlog.RunLog.Infof("devices is: %v", devices)
	return devices, err
}

func addDevice(w dcmi.WorkerInterface, spec *specs.Spec, deviceIdList []int) error {
	deviceName := davinciName
	if strings.Contains(getValueByKey(spec.Process.Env, ascendRuntimeOptions), "VIRTUAL") {
		deviceName = virtualDavinciName
	}
	for _, deviceId := range deviceIdList {
		dPath := devicePath + deviceName + strconv.Itoa(deviceId)
		dContainerPath, err := getMountPath(dPath, deviceName)
		if err != nil {
			return fmt.Errorf("failed to get virtual davinci name : %v", err)
		}
		if err := addDeviceToSpec(spec, dPath, dContainerPath); err != nil {
			return fmt.Errorf("failed to add davinci device to spec: %v", err)
		}
	}

	if err := addUBDevice(spec); err != nil {
		hwlog.RunLog.Errorf("failed to add ub device, error: %v", err)
		return fmt.Errorf("failed to add ub device to spec: %v", err)
	}

	if err := addManagerDevice(w, spec); err != nil {
		hwlog.RunLog.Errorf("failed to add manager device, error: %v", err)
		return fmt.Errorf("failed to add Manager device to spec: %v", err)
	}

	return nil
}

func updateEnvAndPostHook(spec *specs.Spec, vdevice dcmi.VDeviceInfo, deviceIdList *[]int) error {
	if deviceIdList == nil {
		return nil
	}
	newEnv := make([]string, 0, len(spec.Process.Env)+1)
	needAddVirtualFlag := true
	*deviceIdList = []int{int(vdevice.VdeviceID)}
	for _, line := range spec.Process.Env {
		words := strings.Split(line, "=")
		if len(words) == envLength && strings.TrimSpace(words[0]) == ascendRuntimeOptions {
			needAddVirtualFlag = false
			if strings.Contains(words[1], "VIRTUAL") {
				newEnv = append(newEnv, line)
				continue
			} else {
				newEnv = append(newEnv, strings.TrimSpace(line)+",VIRTUAL")
				continue
			}
		}
		newEnv = append(newEnv, line)
	}
	if needAddVirtualFlag {
		newEnv = append(newEnv, fmt.Sprintf(api.AscendRuntimeOptionsEnv+"=VIRTUAL"))
	}
	spec.Process.Env = newEnv
	currentExecPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot get the path of docker-destroy: %#v", err)
	}
	postHookCliPath := path.Join(path.Dir(currentExecPath), destroyHookCli)
	_, err = mindxcheckutils.RealFileChecker(postHookCliPath, true, false, mindxcheckutils.DefaultSize)
	if err != nil {
		return fmt.Errorf("failed to check docker-destroy executable file at %s: %#v", postHookCliPath, err)
	}
	spec.Hooks.Poststop = append(spec.Hooks.Poststop, specs.Hook{
		Path: postHookCliPath,
		Args: []string{postHookCliPath, fmt.Sprintf("%d", vdevice.CardID), fmt.Sprintf("%d", vdevice.DeviceID),
			fmt.Sprintf("%d", vdevice.VdeviceID)},
	})
	return nil
}

func modifySpecFile(path string) error {
	if err := mindxcheckutils.CheckPath(path, true); err != nil {
		hwlog.RunLog.Error(err)
		return err
	}
	spec, err := readSpecFile(path)
	if err != nil {
		hwlog.RunLog.Errorf("failed to read spec file, error: %v", err)
		return err
	}

	if err := processDevicesAndHooks(spec); err != nil {
		hwlog.RunLog.Error(err)
		return err
	}

	return common.WriteSpecFile(path, spec)
}

// readSpecFile handles file reading and parsing
func readSpecFile(path string) (*specs.Spec, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("spec file does not exist %s: %v", path, err)
	}
	jsonFile, err := os.OpenFile(path, os.O_RDWR, stat.Mode())
	if err != nil {
		return nil, fmt.Errorf("cannot open oci spec file %s: %v", path, err)
	}
	defer jsonFile.Close()
	if err = mindxcheckutils.CheckFileInfo(jsonFile, mindxcheckutils.DefaultSize); err != nil {
		hwlog.RunLog.Error(err)
		return nil, err
	}
	jsonContent, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read oci spec file %s: %v", path, err)
	}

	var spec specs.Spec
	if err = json.Unmarshal(jsonContent, &spec); err != nil {
		return nil, fmt.Errorf("failed to unmarshal oci spec file %s: %v", path, err)
	}
	return &spec, nil
}

// processDevicesAndHooks handles device detection and hook addition
func processDevicesAndHooks(spec *specs.Spec) error {
	devices, err := checkVisibleDevice(spec)
	if err != nil {
		return fmt.Errorf("failed to check %v parameter, err: %v", api.AscendVisibleDevicesEnv, err)
	}

	if len(devices) != 0 {
		npuWorker, err := dcmi.GetMatchingNpuWorker()
		if err != nil {
			return err
		}
		if err = addHook(npuWorker, spec, &devices); err != nil {
			return fmt.Errorf("failed to inject hook, err: %v", err)
		}
		if err = addDevice(npuWorker, spec, devices); err != nil {
			return fmt.Errorf("failed to add device to env: %v", err)
		}
	}
	addAscendDockerEnv(spec)
	addAscendLibraryPath(spec)
	if isMountByRuntimeForDP(spec.Process.Env) {
		addDPMountsToSpec(spec)
	}
	return nil
}

// DoProcess does what ascend-docker-runtime is supposed to do before runc being executed.
func DoProcess() error {
	args, err := getArgs()
	if err != nil {
		return fmt.Errorf("failed to get args: %v", err)
	}
	updateRoot(args)
	if args.Cmd == common.CmdCheckpoint || args.Cmd == common.CmdResume {
		if err = mindxcheckutils.ChangeRuntimeLogMode("runtime-checkpoint-"); err != nil {
			hwlog.RunLog.Errorf("change log for checkpoint failed: %v", err)
		}
	}

	cmdHandlers := map[string]func() error{
		common.CmdStart: func() error {
			return grus.Sstart(args)
		},
		common.CmdCreate: func() error {
			if args.Bundle == "" {
				hwlog.RunLog.Warn("get bundleDirPath is empty,try get current working dir from pwd ")
				args.Bundle, err = os.Getwd()
				if err != nil {
					return fmt.Errorf("failed to get current working dir: %v", err)
				}
			}

			specFilePath := args.Bundle + "/config.json"

			if err = modifySpecFile(specFilePath); err != nil {
				return fmt.Errorf("failed to modify spec file %s: %v", specFilePath, err)
			}

			err = grus.Screate(args)
			if err == nil {
				return nil
			}
			hwlog.RunLog.Errorf("screate err: %v", err)
			// only when valid image path/directory is empty or validation failed, continue process
			if err.Error() != "valid image path is empty" && err.Error() != "valid image path directory is empty" &&
				!strings.Contains(err.Error(), "image path validation failed") {
				return err
			}
			return common.ExecRunc()
		},
		common.CmdCheckpoint: func() error {
			return grus.Scheckpoint(args)
		},
		common.CmdResume: func() error {
			return grus.Sresume(args)
		},
	}

	if handler, ok := cmdHandlers[args.Cmd]; ok {
		return handler()
	}

	return common.ExecRunc()
}
