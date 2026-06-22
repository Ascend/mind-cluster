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

// Package common, constant of runtime
package common

const (
	ENV_KEY_VALUE_MAX_PARTS = 2

	ROOTFS_EXTERNAL_DIFF = "rootfs-external-diff.tar"
	ROOTFS_DIFF_DIGEST   = "rootfs-diff.digest"

	GRUS_SNAPSHOT_IMAGE_PATH    = "host_snapshot_path"
	POD_NAME                    = "pod_name"
	GRUS_SNAPSHOT_RESTORED_FLAG = "GRUS_SNAPSHOT_RESTORED_FLAG"
	GRUS_RESTORE_FLAG_FILE      = "/root/.grusflag"

	CRIU_LOG_LEVEL        = "CRIU_LOG_LEVEL"
	CRIULogLevelMin       = 0
	CRIULogLevelMax       = 4
	NPU_PLUGIN_DUMP_ERR   = "[npu-plugin fini-dump err]"
	INETSK_LOCAL_IPV4_KEY = "INETSK_LOCAL_IPV4_KEY"

	DEV_SHM_PATH = "/dev/shm"

	COMMON_DIR_MODE  = 0750
	COMMON_FILE_MODE = 0640
	SAFE_CONFIG_MODE = 0600

	READ_MAX_LEN = 4096

	ContainerdRunRoot    = "/run/containerd/runc"
	RuntimeNameRunc      = "runc"
	ContainerdEngineName = "containerd"
	ContainerdSock       = "/run/containerd/containerd.sock"

	RESTORE_ROOT_DIR = "/var/log/ascend-docker-runtime/restore"

	dockerRuncFile = "docker-runc"
	runcFile       = "runc"
)

var (
	DockerRuncName = dockerRuncFile
	RuncName       = runcFile
)
