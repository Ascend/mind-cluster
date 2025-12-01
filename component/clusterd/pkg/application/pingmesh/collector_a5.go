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

/*
Package pingmesh a series of function handle ping mesh configmap create/update/delete.
*/
package pingmesh

import (
	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
)

func isValidDeviceType(nodeDevice *api.NodeDevice) bool {
	if nodeDevice.ServerType == api.VersionA3 {
		hwlog.RunLog.Infof("device type is %v", api.VersionA3)
		return true
	}
	if nodeDevice.ServerType == api.VersionA5 {
		hwlog.RunLog.Infof("device type is %v, acceleratorType is %s", api.VersionA5, nodeDevice.AcceleratorType)
		return nodeDevice.AcceleratorType == api.A5PodType || nodeDevice.AcceleratorType == api.Ascend800ia5SuperPod
	}
	hwlog.RunLog.Infof("current device type is invalid")
	return false
}
