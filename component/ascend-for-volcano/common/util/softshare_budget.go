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

import "math"

// softShareCoreScaling inflates SoftShareDevCount for chip1softsharedev Fit.
// Default 1 keeps the historical 100-point budget. Ascend Device Plugin
// softShareCoreScaling must advertise the same ratio.
var softShareCoreScaling = 1.0

// SoftShareCoreScalingArg is the volcano plugin argument key for soft-share
// compute oversell (must match device-plugin -softShareCoreScaling).
const SoftShareCoreScalingArg = "softShareCoreScaling"

// SetSoftShareCoreScaling updates the soft-share oversell ratio used by Fit.
// Values below 1 are clamped to 1 so hami-core-style under-provisioning is not
// introduced by misconfiguration.
func SetSoftShareCoreScaling(v float64) {
	if math.IsNaN(v) || math.IsInf(v, 0) || v < 1 {
		softShareCoreScaling = 1
		return
	}
	softShareCoreScaling = v
}

// SoftShareCoreScaling returns the current soft-share oversell ratio.
func SoftShareCoreScaling() float64 {
	return softShareCoreScaling
}

// SoftShareAicoreBudget is the percentage capacity used when packing soft-share
// tasks onto one card. Per-task aicoreQuota remains in [MinAicoreQuota,
// MaxAicoreQuota]; only the card budget may exceed the base.
func SoftShareAicoreBudget() int {
	budget := int(math.Round(float64(SoftShareDevCount) * softShareCoreScaling))
	if budget < MaxAicoreQuota {
		return MaxAicoreQuota
	}
	return budget
}
