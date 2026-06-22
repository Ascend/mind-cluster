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

// Package grus, restore of grus
package grus

import (
	"fmt"
	"golang.org/x/sys/unix"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/opencontainers/runtime-spec/specs-go"

	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
	"ascend-docker-runtime/runtime/common"
	"ascend-docker-runtime/runtime/grus/rootfs"
)

var (
	// unixOpen is a variable holding the unix.Open function to support ut test.
	unixOpen = unix.Open
	// unixClose is a variable holding the unix.Close function to support ut test.
	unixClose = unix.Close
	// unixSetns is a variable holding the unix.Setns function to support ut test.
	unixSetns = unix.Setns
	// netInterfaceAddrs is a variable holding the net.InterfaceAddrs function to support ut test.
	netInterfaceAddrs = net.InterfaceAddrs
)

func getEnvFromSpec(spec *specs.Spec, key string) string {
	for _, env := range spec.Process.Env {
		kv := strings.SplitN(env, "=", common.ENV_KEY_VALUE_MAX_PARTS)
		if len(kv) == common.ENV_KEY_VALUE_MAX_PARTS && kv[0] == key {
			return kv[1]
		}
	}
	return ""
}

func createFlagFile(spec *specs.Spec, rootfs string) error {
	f := getEnvFromSpec(spec, common.GRUS_SNAPSHOT_RESTORED_FLAG)
	hwlog.RunLog.Infof("GRUS_SNAPSHOT_RESTORED_FLAG=%s", f)

	if strings.ToLower(f) != "true" {
		hwlog.RunLog.Infof("flag disabled, skip creating /root/.grusflag")
		return nil
	}

	flagFile := filepath.Clean(filepath.Join(rootfs, common.GRUS_RESTORE_FLAG_FILE))

	hwlog.RunLog.Infof("creating restore flag file: %s", flagFile)

	if err := os.MkdirAll(filepath.Dir(flagFile), common.COMMON_DIR_MODE); err != nil {
		return fmt.Errorf("mkdir flag parent failed: %v", err)
	}

	fd, err := os.OpenFile(flagFile, os.O_CREATE|os.O_RDWR, common.COMMON_FILE_MODE)
	if err != nil {
		return fmt.Errorf("create flag failed: %v", err)
	}
	fd.Close()

	if err = os.Chown(flagFile, 0, 0); err != nil {
		return fmt.Errorf("chown flag failed: %v", err)
	}

	hwlog.RunLog.Infof("flag file created OK: %s", flagFile)
	return nil
}

func prepareLogfile(bundlePath string) error {
	logPath := filepath.Join(bundlePath, "log.json")
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		logFile, err := os.Create(logPath)
		if err != nil {
			hwlog.RunLog.Errorf("log.json does not exist, creating it err: %v", err)
			return err
		}
		defer logFile.Close()
	}
	return nil
}

func rootfsRestore(ckptPath, rootfsPath string) error {
	snapshot := rootfs.GetRootfsSnapshot(common.ContainerdEngineName, common.ContainerdSock)
	if err := snapshot.Restore(ckptPath, rootfsPath, ""); err != nil {
		hwlog.RunLog.Errorf("Cannot apply containerd rootfs diff, error: %v", err)
		return err
	}

	return nil
}

func getContainerNetns(conID string, config *specs.Spec) (string, error) {
	for _, ns := range config.Linux.Namespaces {
		if ns.Type != specs.NetworkNamespace {
			continue
		}
		return ns.Path, nil
	}
	return "", fmt.Errorf("network namespace not found for %s", conID)
}

func getLocalIP() ([]string, error) {
	addrs, err := netInterfaceAddrs()
	if err != nil {
		return nil, err
	}

	// result[0] is IPv4, result[1] is IPv6
	result := make([]string, 2)
	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if !ok || ipNet.IP.IsLoopback() {
			continue
		}

		if ip4 := ipNet.IP.To4(); ip4 != nil {
			if result[0] == "" {
				result[0] = ip4.String()
			}
		} else if result[1] == "" {
			result[1] = ipNet.IP.String()
		}
	}

	if result[0] == "" && result[1] == "" {
		err = fmt.Errorf("failed to obtain any non-loopback IP address (no IPv4 or IPv6)")
		hwlog.RunLog.Errorf("getLocalIP failed: no valid IP address found")
		return result, err
	}

	hwlog.RunLog.Infof("getLocalIP result: IPv4=%q, IPv6=%q", result[0], result[1])
	return result, err
}

// only support k8s
func getNetnsIP(nsPath string) ([]string, error) {
	fd, err := unixOpen(nsPath, unix.O_RDONLY|unix.O_CLOEXEC, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to open network namespace: %v", err)
	}
	defer unixClose(fd)

	// Save current network namespace
	currentNs, err := unixOpen("/proc/self/ns/net", unix.O_RDONLY|unix.O_CLOEXEC, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to save current network namespace: %v", err)
	}
	defer unixClose(currentNs)

	// Enter new network namespace
	if err = unixSetns(fd, unix.CLONE_NEWNET); err != nil {
		return nil, fmt.Errorf("failed to set network namespace: %v", err)
	}

	addr, err2 := getLocalIP()

	// Restore original network namespace
	if err = unixSetns(currentNs, unix.CLONE_NEWNET); err != nil {
		return nil, fmt.Errorf("failed to restore original network namespace: %v", err)
	}

	return addr, err2
}

func getContainerNewIP(conID string, config *specs.Spec) ([]string, error) {
	netNs, err := getContainerNetns(conID, config)
	if err != nil {
		//host network mode
		hwlog.RunLog.Errorf("not support host network mode, err: %v", err)
		return nil, nil
	}
	if netNs == "" {
		//new net ns, just ignore it
		hwlog.RunLog.Infof("container %s with new network namespace, skipping IP setup", conID)
		return nil, nil
	}

	ip4Addr, err := getNetnsIP(netNs)
	if err != nil {
		hwlog.RunLog.Errorf("get ns ip fail: %v", err)
		return nil, err
	}

	return ip4Addr, nil
}

func readRuntimeLog(logPath string) string {
	file, err := os.Open(logPath)
	if err != nil {
		return ""
	}
	defer file.Close()
	lreader := io.LimitReader(file, common.READ_MAX_LEN)
	data := make([]byte, common.READ_MAX_LEN)
	if n, err := lreader.Read(data); err == nil && n > 0 {
		return strings.TrimSpace(string(data))
	}
	return ""
}

// validateSnapshotImagePath validates GRUS_SNAPSHOT_IMAGE_PATH per strict policy:
// - "" or non-absolute path → skip snapshot (return "", nil)
// - absolute path but not exist → ERROR (return "", err)
// - absolute path and exists as dir → OK (return cleanPath, nil)
// - absolute path but is a file → ERROR (since it's not a valid snapshot dir)
func validateSnapshotImagePath(imagePath string, containerID string) (string, error) {
	if imagePath == "" {
		hwlog.RunLog.Infof("Image path not set, skipping snapshot restore for container %s", containerID)
		return "", nil
	}

	if !filepath.IsAbs(imagePath) {
		hwlog.RunLog.Infof("Image path is not absolute: %q, skipping snapshot restore for container %s", imagePath, containerID)
		return "", nil
	}

	cleanPath := filepath.Clean(imagePath)
	info, err := os.Stat(cleanPath)
	if err != nil {
		if os.IsNotExist(err) {
			errMsg := fmt.Sprintf("snapshot image path does not exist: %s", cleanPath)
			hwlog.RunLog.Errorf("FATAL: %s (container %s)", errMsg, containerID)
			return "", fmt.Errorf("invalid snapshot config: %s", errMsg)
		}
		hwlog.RunLog.Errorf("Failed to stat snapshot path %s for container %s: %v", cleanPath, containerID, err)
		return "", fmt.Errorf("failed to access snapshot path %s: %v", cleanPath, err)
	}

	if !info.IsDir() {
		errMsg := fmt.Sprintf("snapshot image path is not a directory: %s", cleanPath)
		hwlog.RunLog.Errorf("FATAL: %s (container %s)", errMsg, containerID)
		return "", fmt.Errorf("invalid snapshot config: %s", errMsg)
	}

	return cleanPath, nil
}

func Screate(a *common.Args) error {
	if a.Bundle == "" {
		return fmt.Errorf("restore require bundle")
	}
	runtimeNs := filepath.Base(a.Root)

	spec, err := readOCIConfig(a.Bundle)
	if err != nil {
		return err
	}
	crootfs := filepath.Join(a.Bundle, spec.Root.Path)

	imgPath := getEnvFromSpec(spec, common.GRUS_SNAPSHOT_IMAGE_PATH)
	if imgPath == "" {
		return fmt.Errorf("valid image path is empty")
	}
	podName := getEnvFromSpec(spec, common.POD_NAME)
	podIndex, err := utils.GetLastNumberFromString(podName)
	if err != nil {
		return fmt.Errorf("get pod index from pod name failed: %v", err)
	}
	validImgPath, err := validateSnapshotImagePath(filepath.Join(imgPath, podIndex), a.ContainerID)
	if err != nil {
		return fmt.Errorf("image path validation failed: %v", err)
	}
	files, _ := os.ReadDir(validImgPath)
	if len(files) == 0 {
		return fmt.Errorf("valid image path directory is empty")
	}

	// Proceed with snapshot restore
	hwlog.RunLog.Infof("Valid image path found: %s, begin restore for %s", validImgPath, a.ContainerID)
	a.CkptPath = validImgPath

	// Step 1: update ip of container and prepare image path
	newIP, err := getContainerNewIP(a.ContainerID, spec)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get container network IP for %s: %v", a.ContainerID, err)
		return fmt.Errorf("prepare container image path failed: %v", err)
	}
	// Step 2: disabling hardware breakpoints, to improve performance
	externalEnvs := []string{"CRIU_FAULT=130", "CRIU_CALL_BY_GRUS=1"}
	criuLogLevelRaw := getEnvFromSpec(spec, common.CRIU_LOG_LEVEL)
	criuLogLevel := parseCRIULogLevel(criuLogLevelRaw)
	externalEnvs = append(externalEnvs, fmt.Sprintf("%s=%s", common.CRIU_LOG_LEVEL, criuLogLevel))
	hwlog.RunLog.Infof("CRIU_LOG_LEVEL has been set to %s", criuLogLevel)

	// Only set IPv4 env if it's available
	if newIP != nil && newIP[0] != "" {
		externalEnvs = append(externalEnvs, fmt.Sprintf("%s=%s", common.INETSK_LOCAL_IPV4_KEY, newIP[0]))
	}

	shmSource := findDevShmMountSource(spec.Mounts)
	if shmSource != "" {
		dst := fmt.Sprintf("%s/%s", validImgPath, common.ROOTFS_EXTERNAL_DIFF)
		hwlog.RunLog.Infof("SNAPSHOT_LINK_REMAP_SRC=%s, SNAPSHOT_LINK_REMAP_DST=%s", shmSource, dst)
		externalEnvs = append(externalEnvs,
			fmt.Sprintf("SNAPSHOT_LINK_REMAP_SRC=%s", shmSource),
			fmt.Sprintf("SNAPSHOT_LINK_REMAP_DST=%s", dst),
		)
	} else {
		hwlog.RunLog.Infof("Skipping /dev/shm remap: no external shm mount point found")
	}

	// Step 3: apply diff tar to rw layer
	hwlog.RunLog.Infof("Restoring rootfs from image path: %s", validImgPath)
	if err = rootfsRestore(validImgPath, crootfs); err != nil {
		return err
	}

	// Step 4: create restore flag file into rw layer
	if err = createFlagFile(spec, crootfs); err != nil {
		return err
	}

	// Step 5: create log.json
	if err = prepareLogfile(a.Bundle); err != nil {
		hwlog.RunLog.Errorf("Failed to prepare log.json in bundle %s: %v", a.Bundle, err)
		return err
	}

	if newIP != nil && newIP[0] != "" {
		addEnv(spec, common.INETSK_LOCAL_IPV4_KEY, fmt.Sprintf("%s=%s", "POD_IP", newIP[0]))
	}
	if err = common.WriteSpecFile(filepath.Join(a.Bundle, "config.json"), spec); err != nil {
		hwlog.RunLog.Errorf("failed to modify spec file: %v", err)
	}

	// Step 6: call runtime restore container
	client := getRuntime(common.RuntimeNameRunc, common.ContainerdRunRoot)
	if err = client.Restore(validImgPath, a.ContainerID, runtimeNs, externalEnvs); err != nil {
		runtimeLog := readRuntimeLog(filepath.Join(a.Bundle, "log.json"))
		hwlog.RunLog.Errorf("runtime restore for %s failed: %v, log: %s", a.ContainerID, err, runtimeLog)
		return err
	}

	return nil
}
