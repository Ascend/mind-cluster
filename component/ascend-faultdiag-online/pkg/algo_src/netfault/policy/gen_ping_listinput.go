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

// Package policy is used for processing superpod infromation
package policy

import (
	"fmt"
	"strconv"

	"ascend-common/common-utils/hwlog"
)

func makeAlgoArg(argMap map[string]any, npu2DFullMesh []string, npuOutOfRackPath map[string][]string) bool {
	npuNetPlanes := make(map[string]any)

	k, exist := npuOutOfRackPath["netplane_0"]

	//A3构造参数
	if exist {
		delete(npuOutOfRackPath, "netplane_0")
		if npuOutOfRackPath == nil {
			npuOutOfRackPath = make(map[string][]string)
		}
		npuOutOfRackPath["1"] = k
	}
	for netPlaneIdStr, npuNetPlanLinks := range npuOutOfRackPath {
		netPlaneId, err := strconv.Atoi(netPlaneIdStr)
		if err != nil {
			hwlog.RunLog.Errorf("the err info is: %v", err)
			return false
		}
		netPlaneName := fmt.Sprintf("netplane_%d", netPlaneId-1)
		npuNetPlanes[netPlaneName] = npuNetPlanLinks
	}
	if argMap == nil {
		hwlog.RunLog.Info("invalid algo input map")
		return false
	}
	if len(npu2DFullMesh) == 0 {
		argMap["npu_npu"] = make([]string, 0)
	} else {
		argMap["npu_npu"] = npu2DFullMesh
	}
	argMap["npu_netplane"] = npuNetPlanes
	return true
}

func spliceAlgorithmInput(npu2DFullMesh []string, npuOutOfRackPath map[string][]string) map[string]any {
	if len(npuOutOfRackPath) == 0 && len(npu2DFullMesh) == 0 {
		hwlog.RunLog.Errorf("Invalid input")
		return nil
	}
	InputJsonMap := make(map[string]any)
	if !makeAlgoArg(InputJsonMap, npu2DFullMesh, npuOutOfRackPath) {
		return nil
	}
	InputJsonMap["npu_superpod"] = map[string]any{}
	return InputJsonMap
}
