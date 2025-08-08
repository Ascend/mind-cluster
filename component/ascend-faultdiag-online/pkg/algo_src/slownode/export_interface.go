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

// Package slownode is the main entry
package slownode

import (
	"ascend-common/common-utils/hwlog"
	"ascend-faultdiag-online/pkg/algo_src/slownode/externalbridge"
	"ascend-faultdiag-online/pkg/core/model"
)

const (
	defaultCusterLogFile = "/var/log/mindx-dl/clusterd/slownode.log"
	defaultNodeLogFile   = "/var/log/mindx-dl/noded/slownode.log"
	maxAge               = 7
	maxBackups           = 7
)

// Execute for a uniform interface
func Execute(input model.Input) int {
	hwlog.RunLog.Infof("[SLOWNODE]execute got req struct data: %+v", input)
	return externalbridge.Execute(&input)
}

// GetType to return algorithm type
func GetType() string {
	return "slownode"
}

// GetVersion to return algorithm version
func GetVersion() string {
	return "1.0.0"
}
