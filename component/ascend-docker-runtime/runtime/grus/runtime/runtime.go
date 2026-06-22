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

// Package runtime, interface of runtime
package runtime

import "ascend-common/common-utils/hwlog"

const (
	// CONTAINER_ID_FILE container id file
	CONTAINER_ID_FILE = "container.id"
	// CRIU_IMG_DIR image dir
	CRIU_IMG_DIR = "image"
)

// StateInfo represents the state information of a container
type StateInfo struct {
	// ID container ID
	ID string
	// Pid process ID
	Pid int
	// Status container status
	Status string
	// Bundle path to the bundle directory
	Bundle string
	// Rootfs path to the root filesystem
	Rootfs string
}

// RuntimeAPI defines the interface for container runtime operations
type RuntimeAPI interface {
	// Init initializes the runtime with the specified root path
	Init(rootPath string)
	// Pause pauses the container with the specified ID
	Pause(id string) error
	// Resume resumes the paused container with the specified ID
	Resume(id string) error
	// State returns the state information of the container with the specified ID
	State(id string) (*StateInfo, error)
	// Checkpoint creates a checkpoint of the container at the specified path
	Checkpoint(ckptPath, id string) error
	// Restore restores the container from a checkpoint at the specified path
	Restore(ckptPath, id, ns string, externalEnvs []string) error
}

var runtimes map[string]RuntimeAPI

func init() {
	runtimes = make(map[string]RuntimeAPI)
	runtimes[RuncName] = NewRuntimeRunc()
	runtimes[MockName] = NewRuntimeMock()
}

// GetRuntime returns the appropriate runtime instance based on the binary name
func GetRuntime(binaryName, rootPath string) RuntimeAPI {
	if r, ok := runtimes[binaryName]; ok {
		r.Init(rootPath)
		return r
	}

	hwlog.RunLog.Infof("invalid runtime config: %s, ignore error and use default runc runtime", binaryName)
	defaultRuntime := NewRuntimeRunc()
	defaultRuntime.Init(rootPath)
	return defaultRuntime
}
