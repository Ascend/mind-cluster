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

// Package grus, checkpoint of grus
package grus

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/opencontainers/runtime-spec/specs-go"

	"ascend-common/common-utils/hwlog"
	"ascend-docker-runtime/runtime/common"
	"ascend-docker-runtime/runtime/grus/rootfs"
	"ascend-docker-runtime/runtime/grus/runtime"
)

var getRuntime = runtime.GetRuntime

type runtimeClient struct {
	client  runtime.RuntimeAPI
	bundle  string
	conSpec *specs.Spec
}

func initRuntimeClient(containerID, root string) (*runtimeClient, error) {
	result := &runtimeClient{}
	result.client = getRuntime(common.RuntimeNameRunc, root)

	con, err := result.client.State(containerID)
	if err != nil {
		return nil, err
	}
	result.bundle = con.Bundle
	spec, err := readOCIConfig(result.bundle)
	if err != nil {
		return nil, err
	}
	result.conSpec = spec
	return result, nil
}

func (c *runtimeClient) pause(containerID string) error {
	if c.client == nil {
		return fmt.Errorf("runtime client not init")
	}
	return c.client.Pause(containerID)
}

func (c *runtimeClient) resume(containerID string) error {
	if c.client == nil {
		return fmt.Errorf("runtime client not init")
	}
	return c.client.Resume(containerID)
}

func (c *runtimeClient) state(containerID string) (*runtime.StateInfo, error) {
	if c.client == nil {
		return nil, fmt.Errorf("runtime client not init")
	}
	return c.client.State(containerID)
}

func findDevShmMountSource(mounts []specs.Mount) string {
	for _, m := range mounts {
		if m.Destination == common.DEV_SHM_PATH && strings.Contains(m.Source, "/") {
			return m.Source
		}
	}
	return ""
}

func parseCRIULogLevel(raw string) string {
	effectiveLevel := common.LogLevel
	if raw == "" {
		return effectiveLevel
	}
	level, err := strconv.Atoi(raw)
	if err != nil || level < common.CRIULogLevelMin || level > common.CRIULogLevelMax {
		hwlog.RunLog.Errorf("Invalid CRIU_LOG_LEVEL '%s', falling back to default '1'", raw)
		return effectiveLevel
	}
	return raw
}

func (c *runtimeClient) updateCkptEnvs(checkpointPath string) error {
	dst := fmt.Sprintf("%s/%s", checkpointPath, common.ROOTFS_EXTERNAL_DIFF)
	if c.conSpec == nil {
		return fmt.Errorf("container spec (conSpec) is nil")
	}
	// Set Grus-specific CRIU environment variables
	if err := os.Setenv("CRIU_CALL_BY_GRUS", "1"); err != nil {
		hwlog.RunLog.Errorf("Failed to set CRIU_CALL_BY_GRUS=1: %v", err)
		return err
	}
	hwlog.RunLog.Infof("CRIU_CALL_BY_GRUS=%s", "1")

	criuLogLevelRaw := os.Getenv(common.CRIU_LOG_LEVEL)
	criuLogLevel := parseCRIULogLevel(criuLogLevelRaw)
	if err := os.Setenv(common.CRIU_LOG_LEVEL, criuLogLevel); err != nil {
		hwlog.RunLog.Errorf("Failed to set CRIU_LOG_LEVEL=%s: %v", criuLogLevel, err)
		return fmt.Errorf("failed to set CRIU_LOG_LEVEL: %v", err)
	}
	hwlog.RunLog.Infof("CRIU_LOG_LEVEL has been set to %s", criuLogLevel)

	shmSource := findDevShmMountSource(c.conSpec.Mounts)
	if shmSource != "" {
		hwlog.RunLog.Infof("Set SNAPSHOT_LINK_REMAP_SRC=%s, SNAPSHOT_LINK_REMAP_DST=%s", shmSource, dst)
		if err := os.Setenv("SNAPSHOT_LINK_REMAP_SRC", shmSource); err != nil {
			hwlog.RunLog.Errorf("Failed to set SNAPSHOT_LINK_REMAP_SRC=%s: %v", shmSource, err)
			return err
		}
		if err := os.Setenv("SNAPSHOT_LINK_REMAP_DST", dst); err != nil {
			hwlog.RunLog.Errorf("Failed to set SNAPSHOT_LINK_REMAP_DST=%s: %v", dst, err)
			return err
		}
	} else {
		hwlog.RunLog.Infof("Skipping /dev/shm remap: no external shm mountpoint found")
	}
	return nil
}

func (c *runtimeClient) checkpoint(ckptPath, containerID string) error {
	if c.client == nil {
		return fmt.Errorf("runtime client not init")
	}

	if err := c.updateCkptEnvs(ckptPath); err != nil {
		return err
	}
	if err := c.client.Checkpoint(ckptPath, containerID); err != nil {
		return err
	}
	return checkDumpLogForNpuError(ckptPath)
}

func rootfsCheckpoint(a *common.Args, ns string) error {
	snapshot := rootfs.GetRootfsSnapshot(common.ContainerdEngineName, common.ContainerdSock)
	digest, err := snapshot.Checkpoint(a.CkptPath, a.ContainerID, ns)
	if err != nil {
		hwlog.RunLog.Errorf("Cannot create containerd rootfs diff, error: %v", err)
		return err
	}

	digestFile := filepath.Join(a.CkptPath, common.ROOTFS_DIFF_DIGEST)
	if err := os.WriteFile(digestFile, []byte(digest), common.COMMON_FILE_MODE); err != nil {
		hwlog.RunLog.Errorf("Create rootfs digest file: %s, err: %v", digest, err)
		return err
	}

	return nil
}

func checkDumpLogForNpuError(ckptPath string) error {
	logPath := filepath.Join(ckptPath, "image", "work", "dump.log")
	file, err := os.Open(logPath)
	if err != nil {
		hwlog.RunLog.Errorf("dump.log not found at %s, skip NPU error check", logPath)
		return nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, common.NPU_PLUGIN_DUMP_ERR) {
			hwlog.RunLog.Infof("NPU dump error detected in CRIU log: %s", line)
			return fmt.Errorf("detected NPU plugin dump failure in CRIU log: %s", common.NPU_PLUGIN_DUMP_ERR)
		}
	}

	if err := scanner.Err(); err != nil {
		hwlog.RunLog.Errorf("error reading dump.log: %v", err)
	}

	return nil
}

func Scheckpoint(a *common.Args) (err error) {
	hwlog.RunLog.Infof("begin checkpoint for %s", a.ContainerID)

	// step 1: init runtime client
	client, err := initRuntimeClient(a.ContainerID, a.Root)
	if err != nil {
		hwlog.RunLog.Errorf("init container err: %v", err)
		return
	}

	// step 2: pause container
	if err = client.pause(a.ContainerID); err != nil {
		hwlog.RunLog.Errorf("pause container err: %v", err)
		return
	}

	defer func() {
		// step 5: resume container
		if terr := client.resume(a.ContainerID); terr != nil {
			hwlog.RunLog.Errorf("resume container: %s failed: %v", a.ContainerID, terr)
			if err != nil {
				err = fmt.Errorf("original err: %v and resume err: %v", err, terr)
			}
		}
	}()

	//step 3: use criu to runtime checkpoint container
	if err = client.checkpoint(a.CkptPath, a.ContainerID); err != nil {
		hwlog.RunLog.Errorf("runtime checkpoint err: %v", err)
		return
	}

	//step 4: checkpoint rw layer of container
	runtimeNs := filepath.Base(a.Root)
	if err = rootfsCheckpoint(a, runtimeNs); err != nil {
		hwlog.RunLog.Errorf("rootfs checkpoint err: %v", err)
		return
	}

	return nil
}

// Sresume if grus-agent killed, resume which in scheckpoint will not do, so need extra resume
// ResumeContainer resumes a paused container after a checkpoint timeout or failure.
func Sresume(a *common.Args) error {
	hwlog.RunLog.Infof("begin resume for %s", a.ContainerID)
	// init runtime client
	client, err := initRuntimeClient(a.ContainerID, a.Root)
	if err != nil {
		hwlog.RunLog.Errorf("init container err: %v", err)
		return fmt.Errorf("failed to initilaize runtime client: %v", err)
	}
	// get container's status
	status, err := client.state(a.ContainerID)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get container status: %v", err)
		return fmt.Errorf("failed to get container status: %v", err)
	}
	// if container not paused, skip resume
	if status.Status != "paused" {
		hwlog.RunLog.Errorf("container %s not paused (status=%v), skip resume", a.ContainerID, status)
		return nil
	}
	// do resume
	if err = client.resume(a.ContainerID); err != nil {
		hwlog.RunLog.Errorf("resume container err: %v", err)
		return fmt.Errorf("failed to resume container: %v", err)
	}

	hwlog.RunLog.Infof("resume container %s success", a.ContainerID)
	return nil
}
