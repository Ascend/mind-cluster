/*
Copyright(C)2025. Huawei Technologies Co.,Ltd. All rights reserved.

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

// Package pingmesh a series of function handle ping mesh configmap create/update/delete
package pingmesh

import (
	"ascend-common/common-utils/hwlog"
	"ascend-common/devmanager/common"
	"clusterd/pkg/common/constant"
)

func isValidConfigPingMesh(cfg constant.ConfigPingMesh) bool {
	if cfg == nil || len(cfg) == 0 {
		hwlog.RunLog.Error("ping mesh config is empty")
		return false
	}

	for i, item := range cfg {
		if item == nil {
			hwlog.RunLog.Errorf("invalid config item for %s, which is empty", i)
			return false
		}
		if item.Activate != constant.RasNetDetectOnStr && item.Activate != constant.RasNetDetectOffStr {
			hwlog.RunLog.Errorf("invalid config for %s, active is neither %s nor %s",
				i, constant.RasNetDetectOnStr, constant.RasNetDetectOffStr)
			return false
		}
		if item.TaskInterval < common.MinTaskInterval || item.TaskInterval > common.MaxTaskInterval {
			hwlog.RunLog.Errorf("task interval %d is invalid, task_interval should be between %d and %d",
				item.TaskInterval, common.MinTaskInterval, common.MaxTaskInterval)
			return false
		}
	}
	return true
}
