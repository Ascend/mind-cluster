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
	reserveIndexFromEnd             = 5
	actionPosition                  = 0
	srcFilePosition                 = 1
	destFilePosition                = 2
	runtimeFilePosition             = 3
	rmCommandLength                 = 5
	addCommandLength                = 6
	maxFileSize                     = 1024 * 1024 * 10
	perm                os.FileMode = 0600
	configVersion1                  = 1
	configVersion2                  = 2
	configVersion3                  = 3
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
	InstallSceneIsula = "isula"
	runtimeName       = "ascend"
	// default runtime type for containerd
	v2RuncRuntimeType         = "io.containerd.runc.v2"
	defaultRuntimeValue       = "runc"
	version1RuntimePluginName = "cri"
	version2RuntimePluginName = "io.containerd.grpc.v1.cri"
	version3RuntimePluginName = "io.containerd.cri.v1.runtime"
	containerdKey             = "containerd"
	runtimesKey               = "runtimes"
	pluginsKey                = "plugins"
	optionsKey                = "options"
	binaryNameKey             = "BinaryName"
	defaultRuntimeNameKey     = "default_runtime_name"
	systemdCgroupKey          = "SystemdCgroup"
)
