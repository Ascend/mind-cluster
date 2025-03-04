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
Package contextdata 全局上下文信息
*/
package contextdata

import (
	"ascend-faultdiag-online/pkg/model/cluster"
	"ascend-faultdiag-online/pkg/model/enum"
	"ascend-faultdiag-online/pkg/model/node"
)

// Environment 环境信息
type Environment struct {
	NodeStatus    *node.Status    // 节点状态， node时使用
	ClusterStatus *cluster.Status // 集群状态， cluster时使用
}

// NewEnvironment 创建环境变量实例
func NewEnvironment() *Environment {
	return &Environment{
		NodeStatus:    queryNodeStatus(),
		ClusterStatus: queryClusterStatus(),
	}
}

// queryNodeStatus 查询节点信息
func queryNodeStatus() *node.Status {
	return &node.Status{
		ChipType: enum.Ascend910A2, // TODO 补全收集服务器信息
	}
}

// queryClusterStatus 查询集群信息
func queryClusterStatus() *cluster.Status {
	return &cluster.Status{} // TODO 补全收集集群信息
}
