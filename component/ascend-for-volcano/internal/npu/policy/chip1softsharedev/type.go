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

// Package chip1softsharedev is using for HuaWei chip1softsharedev schedule.
package chip1softsharedev

import (
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/npu/ascend910/ascend910b"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/rescheduling"
)

type chip1softsharedev struct {
	ascend910b.Base910b
	reHandle        *rescheduling.ReScheduler
	netUnhealthyKey string
}

type softShareDevResource struct {
	aicoreQuota      int
	hbmQuota         int
	schedulingPolicy string
}

type secondPriorityNPU struct {
	npuID        int
	remainAicore int
	remainHbm    int
}

const (
	SchedulePolicySoftShareDev = "softsharedev"
	nodeNPUNumber              = 16
	networkUnhealthyNPU        = "huawei.com/Ascend910-NetworkUnhealthy"
)
