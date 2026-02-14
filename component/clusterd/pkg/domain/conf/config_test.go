/* Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package conf global config base func
package conf

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"

	"clusterd/pkg/common/constant"
)

const (
	defaultFaultWindowHours = 24
	defaultFaultThreshold   = 3
	defaultFaultFreeHours   = 48
)

var validPolicy = ManuallySeparatePolicy{
	Separate: struct {
		FaultWindowHours int `yaml:"fault_window_hours"`
		FaultThreshold   int `yaml:"fault_threshold"`
	}{
		FaultWindowHours: defaultFaultWindowHours,
		FaultThreshold:   defaultFaultThreshold,
	},
	Release: struct {
		FaultFreeHours int `yaml:"fault_free_hours"`
	}{
		FaultFreeHours: defaultFaultFreeHours,
	},
}

// TestCheck tests the Check function for ManuallySeparatePolicy
func TestCheck(t *testing.T) {
	convey.Convey("test func Check", t, func() {
		buildCase := append(buildTestCase1(), buildTestCase2()...)
		buildCase = append(buildCase, buildTestCase3()...)
		buildCase = append(buildCase, buildTestCase4()...)
		for _, tt := range buildCase {
			convey.Convey(tt.name, func() {
				err := Check(tt.policy)
				if tt.expectError {
					convey.So(err.Error(), convey.ShouldContainSubstring, tt.errorContains)
				} else {
					convey.So(err, convey.ShouldBeNil)
				}
			})
		}
	})
}

func buildTestCase1() []struct {
	name          string
	policy        ManuallySeparatePolicy
	expectError   bool
	errorContains string
} {
	return []struct {
		name          string
		policy        ManuallySeparatePolicy
		expectError   bool
		errorContains string
	}{
		{
			name:          "Valid configuration",
			policy:        validPolicy,
			expectError:   false,
			errorContains: "",
		},
		{
			name: "FaultWindowHours below minimum",
			policy: func() ManuallySeparatePolicy {
				p := validPolicy
				p.Separate.FaultWindowHours = MinFaultWindowHours - 1
				return p
			}(),
			expectError:   true,
			errorContains: "fault_window_hours must be in",
		},
		{
			name: "FaultWindowHours above maximum",
			policy: func() ManuallySeparatePolicy {
				p := validPolicy
				p.Separate.FaultWindowHours = MaxFaultWindowHours + 1
				return p
			}(),
			expectError:   true,
			errorContains: "fault_window_hours must be in",
		},
	}
}

func buildTestCase2() []struct {
	name          string
	policy        ManuallySeparatePolicy
	expectError   bool
	errorContains string
} {
	return []struct {
		name          string
		policy        ManuallySeparatePolicy
		expectError   bool
		errorContains string
	}{
		{
			name: "FaultThreshold below minimum",
			policy: func() ManuallySeparatePolicy {
				p := validPolicy
				p.Separate.FaultThreshold = MinFaultThreshold - 1
				return p
			}(),
			expectError:   true,
			errorContains: "fault_threshold must be in",
		},
		{
			name: "FaultThreshold above maximum",
			policy: func() ManuallySeparatePolicy {
				p := validPolicy
				p.Separate.FaultThreshold = MaxFaultThreshold + 1
				return p
			}(),
			expectError:   true,
			errorContains: "fault_threshold must be in",
		},
		{
			name: "FaultFreeHours at minimum boundary",
			policy: func() ManuallySeparatePolicy {
				p := validPolicy
				p.Release.FaultFreeHours = MinFaultFreeHours
				return p
			}(),
			expectError:   false,
			errorContains: "",
		},
	}
}

func buildTestCase3() []struct {
	name          string
	policy        ManuallySeparatePolicy
	expectError   bool
	errorContains string
} {
	return []struct {
		name          string
		policy        ManuallySeparatePolicy
		expectError   bool
		errorContains string
	}{
		{
			name: "FaultFreeHours at maximum boundary",
			policy: func() ManuallySeparatePolicy {
				p := validPolicy
				p.Release.FaultFreeHours = MaxFaultFreeHours
				return p
			}(),
			expectError:   false,
			errorContains: "",
		},
		{
			name: "FaultFreeHours below minimum",
			policy: func() ManuallySeparatePolicy {
				p := validPolicy
				p.Release.FaultFreeHours = MinFaultFreeHours - 1
				return p
			}(),
			expectError:   true,
			errorContains: "fault_free_hours must be in",
		},
		{
			name: "FaultFreeHours above maximum",
			policy: func() ManuallySeparatePolicy {
				p := validPolicy
				p.Release.FaultFreeHours = MaxFaultFreeHours + 1
				return p
			}(),
			expectError:   true,
			errorContains: "fault_free_hours must be in",
		},
	}
}

func buildTestCase4() []struct {
	name          string
	policy        ManuallySeparatePolicy
	expectError   bool
	errorContains string
} {
	return []struct {
		name          string
		policy        ManuallySeparatePolicy
		expectError   bool
		errorContains string
	}{
		{
			name: "FaultFreeHours is NotRelease",
			policy: func() ManuallySeparatePolicy {
				p := validPolicy
				p.Release.FaultFreeHours = NotRelease
				return p
			}(),
			expectError:   false,
			errorContains: "",
		},
	}
}

// TestSetAndGet tests the Set and Get functions
func TestSetAndGet(t *testing.T) {
	convey.Convey("test func set and get", t, func() {
		p := validPolicy
		SetManualSeparatePolicy(p)

		enabled := GetManualEnabled()
		convey.So(enabled, convey.ShouldEqual, p.Enabled)

		faultWindowHours := GetSeparateWindow()
		convey.So(faultWindowHours, convey.ShouldEqual, int64(p.Separate.FaultWindowHours*constant.HoursToMilliseconds))

		faultThreshold := GetSeparateThreshold()
		convey.So(faultThreshold, convey.ShouldEqual, p.Separate.FaultThreshold)

		releaseDuration := GetReleaseDuration()
		convey.So(releaseDuration, convey.ShouldEqual, int64(p.Release.FaultFreeHours*constant.HoursToMilliseconds))
	})
	convey.Convey("test not release", t, func() {
		p := validPolicy
		p.Release.FaultFreeHours = NotRelease
		SetManualSeparatePolicy(p)

		releaseDuration := GetReleaseDuration()
		convey.So(releaseDuration, convey.ShouldEqual, NotRelease)
	})
}
