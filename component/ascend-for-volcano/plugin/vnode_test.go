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

package plugin

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
)

const (
	chipNameAscend910BPro  = "Ascend910B-Pro"
	chipNameAscend310PFull = "Ascend310P-Custom"
	chipNameAscend310Full  = "Ascend310-Custom"
	chipNameAscend950Full  = "Ascend950-Custom"
	chipNameUnknown        = "UnknownChip"
)

func TestDeriveAcceleratorFromChipName(t *testing.T) {
	type testCase struct {
		name        string
		chipName    string
		want        string
		wantDerived bool
	}
	tests := []testCase{
		{name: "should return Ascend910 when chipName starts with Ascend910",
			chipName: chipNameAscend910,
			want:     util.Ascend910, wantDerived: true},
		{name: "should return Ascend910 when chipName contains 910B",
			chipName: chipNameAscend910B,
			want:     util.Ascend910, wantDerived: true},
		{name: "should return Ascend910 when chipName is Ascend910B-Pro",
			chipName: chipNameAscend910BPro,
			want:     util.Ascend910, wantDerived: true},
		{name: "should return Ascend310P when chipName starts with Ascend310P",
			chipName: chipNameAscend310P,
			want:     util.Ascend310P, wantDerived: true},
		{name: "should return Ascend310P when chipName starts with Ascend310P with suffix",
			chipName: chipNameAscend310PFull,
			want:     util.Ascend310P, wantDerived: true},
		{name: "should return Ascend310 when chipName starts with Ascend310",
			chipName: chipNameAscend310,
			want:     util.Ascend310, wantDerived: true},
		{name: "should return Ascend310 when chipName starts with Ascend310 with suffix",
			chipName: chipNameAscend310Full,
			want:     util.Ascend310, wantDerived: true},
		{name: "should return npu when chipName starts with Ascend950",
			chipName: chipNameAscend950,
			want:     util.NPULowerCase, wantDerived: true},
		{name: "should return npu when chipName starts with Ascend950 with suffix",
			chipName: chipNameAscend950Full,
			want:     util.NPULowerCase, wantDerived: true},
		{name: "should return empty and derived when chipName is 310B",
			chipName: chipNameAscend310B,
			want:     "", wantDerived: true},
		{name: "should return empty and not derived when chipName is empty",
			chipName: "",
			want:     "", wantDerived: false},
		{name: "should return empty and not derived when chipName is unknown",
			chipName: chipNameUnknown,
			want:     "", wantDerived: false},
	}
	for _, tt := range tests {
		convey.Convey(tt.name, t, func() {
			got, derived := deriveAcceleratorFromChipName(tt.chipName)
			convey.So(got, convey.ShouldEqual, tt.want)
			convey.So(derived, convey.ShouldEqual, tt.wantDerived)
		})
	}
}
