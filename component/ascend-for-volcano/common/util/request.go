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

package util

// ScheduleMode indicates how strictly a task's NPU topology request must be satisfied.
type ScheduleMode string

const (
	// HardScheduleMode enforces strict topology locality: the requested chips
	// must fit within a single topology domain, otherwise the node is rejected.
	HardScheduleMode ScheduleMode = "hard"
	// SoftScheduleMode relaxes the topology constraint: the node fits as long as
	// it has enough usable chips, or enough chips that can be reclaimed by eviction.
	SoftScheduleMode ScheduleMode = "soft"
)

// Request describes a task's NPU resource request used for topology fitting
// and chip selection.
type Request struct {
	ReqNPUName string
	ReqNPUNum  int
	Mode       ScheduleMode
	// AllowNetUnhealthy tolerates parameter-plane network faults: when true,
	// chips whose parameter plane is unhealthy may still be fitted or selected.
	// Mirrors the huawei.com/parameterplane.unhealthy-tolerance annotation.
	AllowNetUnhealthy bool
}

// ParseScheduleMode converts a schedule.mode annotation value to a ScheduleMode.
// Only the literal "hard" selects HardScheduleMode; every other value
// (including empty or invalid) selects SoftScheduleMode.
func ParseScheduleMode(val string) ScheduleMode {
	if ScheduleMode(val) == HardScheduleMode {
		return HardScheduleMode
	}
	return SoftScheduleMode
}
