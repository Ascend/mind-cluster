//go:build !volcano_v115

/*
Copyright(C)2026. Huawei Technologies Co.,Ltd. All rights reserved.

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

package main

import (
	"k8s.io/klog/v2"
	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
)

func addSimulateFns(ssn *framework.Session, tp *huaweiNPUPlugin) {
	klog.V(util.LogInfoLev).Infof("pre-v1.15: skip simulate fns, topologyAwarePreempt re-check disabled")
}

func topologyAwarePreemptActive(ssn *framework.Session) bool {
	return false
}
