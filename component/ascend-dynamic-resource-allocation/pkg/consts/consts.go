/*
 * Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at

 * http://www.apache.org/licenses/LICENSE-2.0

 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package consts

import "os"

// Driver related constants.
const (
	// DriverPluginCheckpointFile is the checkpoint file name for the DRA plugin.
	DriverPluginCheckpointFile = "checkpoint.json"
	// DriverName is the DRA driver name registered with kubelet.
	DriverName = "npu.huawei.com"
	// DefaultCDIRoot is the default directory where CDI spec files are generated.
	DefaultCDIRoot = "/var/run/cdi"
)

// Path and Environment related constants.
const(
	// DevPath is the host device directory.
	DevPath = "/dev"
	// DefaultDirMode is the default permission for created directories.
	DefaultDirMode = os.FileMode(0750)
	// NodeNameEnv is the environment variable holding the node name.
	NodeNameEnv = "NODE_NAME"
)

// AscendProductName is the product name for Ascend devices.
const (
	// Ascend910ReleasedName is the public released product name for Ascend 910.
	Ascend910ReleasedName = "Ascend910"
	// Ascend950ReleasedName is the released name for Ascend 950.
	Ascend950ReleasedName = "npu"
)


// KubeClientConfig related constants.
const (
	// DefaultKubeAPIQPS is the default QPS for Kubernetes API client.
	DefaultKubeAPIQPS = 5
	// DefaultKubeAPIBurst is the default burst for Kubernetes API client.
	DefaultKubeAPIBurst = 10
)

// Log related constants.
const (
	// DefaultLogMaxAge is the default maximum number of days for backup log files.
	DefaultLogMaxAge = 7
	// DefaultLogLevel is the default log level.
	DefaultLogLevel = 0
)
