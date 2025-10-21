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

package process

import "os"

const commonTemplate = `{
        "runtimes":     {
                "ascend":       {
                        "path": "%s",
                        "runtimeArgs":  []
                }
        },
        "default-runtime":      "ascend"
}`

const noDefaultTemplate = `{
        "runtimes":     {
                "ascend":       {
                        "path": "%s",
                        "runtimeArgs":  []
                }
        }
}`

const (
	reserveIndexFromEnd                = 5
	actionPosition                     = 0
	srcFilePosition                    = 1
	destFilePosition                   = 2
	runtimeFilePosition                = 3
	rmCommandLength                    = 8
	addCommandLength                   = 9
	maxFileSize                        = 1024 * 1024 * 10
	cgroupInfoIndexFromEnd             = 3
	osNameIndexFromEnd                 = 2
	osVersionIndexFromEnd              = 1
	perm                   os.FileMode = 0600
)

const (
	addCommand        = "add"
	rmCommand         = "rm"
	defaultRuntimeKey = "default-runtime"
	// InstallSceneDocker is a 'docker' string of scene
	InstallSceneDocker = "docker"
	// InstallSceneContainerd is a 'containerd' string of scene
	InstallSceneContainerd = "containerd"
	// InstallSceneIsula is a 'isula' string of scene
	InstallSceneIsula          = "isula"
	v1NeedChangeKeyRuntime     = "runtime"
	v1NeedChangeKeyRuntimeType = "runtime_type"
	v1RuntimeType              = "io.containerd.runtime.v1.linux"
	// default runtime type for containerd
	v2RuncRuntimeType                = "io.containerd.runc.v2"
	defaultRuntimeValue              = "runc"
	v1RuntimeTypeFirstLevelPlugin    = "io.containerd.grpc.v1.cri"
	containerdKey                    = "containerd"
	runtimesKey                      = "runtimes"
	runcKey                          = "runc"
	runcOptionsKey                   = "options"
	binaryNameKey                    = "BinaryName"
	cgroupV2InfoStr                  = "cgroup2fs"
	openEulerStr                     = "openEuler"
	openEulerVersionForV2RuntimeType = "24.03"
)

const (
	notFindPluginLogStr       = "can not find plugin %v, plugins is: %+v"
	notFindPluginErrorStr     = "can not find plugin: %v"
	convertConfigFailLogStr   = "can not convert config %v, config is: %+v"
	convertConfigFailErrorStr = "can not convert config %v, config is: %+v"
	convertTreeFailLogStr     = "failed to convert map to tree, error: %v"
	getMapFaileLogStr         = "failed to get map, key: %v, error: %v"
)
