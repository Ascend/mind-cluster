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

// Package slownode start fd and run slownode
package slownode

import (
	"ascend-common/common-utils/hwlog"
	"ascend-faultdiag-online/pkg/global"
	"ascend-faultdiag-online/pkg/model/enum"
	"ascend-faultdiag-online/pkg/utils"
	"ascend-faultdiag-online/pkg/utils/grpc"
)

// StartSlowNode
func StartSlowNode(target enum.DeployMode) {
	switch target {
	case enum.Cluster:
		hwlog.RunLog.Infof("[FD-OL SLOWNODE]start fd-ol slow node in cluster")
		grpcClient, err := grpc.GetClient(utils.GetClusterIP())
		if err != nil {
			hwlog.RunLog.Errorf("[FD-OL SLOWNODE]create grpc client failed: %v", err)
			return
		}
		global.GrpcClient = grpcClient
		registerHandlers(filterSlowNodeFeature, slowNodeFeatureHandler)
		registerHandlers(filterSlowNodeAlgoResult, nodeSlowNodeAlgoHandler)
		registerHandlers(filterDataProfilingResult, nodeDataProfilingHandler)
		AddCMHandler(&jobFuncList, ClusterProcessSlowNodeJob)
		AddCMHandler(&nodeDataProfilingResFuncList, ClusterProcessDataProfilingResult)
		AddCMHandler(&nodeSlowNodeAlgoResFuncList, ClusterProcessSlowNodeAlgoResult)
	case enum.Node:
		hwlog.RunLog.Infof("[FD-OL SLOWNODE]start fd-ol slow node in node")
		registerHandlers(filterNodeSlowNodeJob, nodeSlowNodeJobHandler)
		AddCMHandler(&jobFuncList, NodeProcessSlowNodeJob)
	default:
		hwlog.RunLog.Errorf("[FD-OL SLOWNODE]target is not support, target is %s", target)
		return
	}
	InitCMInformer()
}
