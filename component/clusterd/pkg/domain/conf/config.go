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
	"fmt"

	"clusterd/pkg/common/constant"
)

const (
	// MinFaultWindowHours minimum fault window hours
	MinFaultWindowHours = 1
	// MaxFaultWindowHours maximum fault window hours
	MaxFaultWindowHours = 720
	// MinFaultThreshold minimum fault threshold
	MinFaultThreshold = 1
	// MaxFaultThreshold maximum fault threshold
	MaxFaultThreshold = 50
	// MinFaultFreeHours minimum fault free hours
	MinFaultFreeHours = 1
	// MaxFaultFreeHours maximum fault free hours
	MaxFaultFreeHours = 240
	// NotRelease not release fault
	NotRelease = -1
)

var config GlobalConfig

// GlobalConfig global config
type GlobalConfig struct {
	ManuallySeparatePolicy
}

// ManuallySeparatePolicy manually separate policy config
type ManuallySeparatePolicy struct {
	Enabled  bool `yaml:"enabled"`
	Separate struct {
		FaultWindowHours int `yaml:"fault_window_hours"`
		FaultThreshold   int `yaml:"fault_threshold"`
	} `yaml:"separate"`
	Release struct {
		FaultFreeHours int `yaml:"fault_free_hours"`
	} `yaml:"release"`
}

// GetManualEnabled get manually separate enabled
func GetManualEnabled() bool {
	return config.ManuallySeparatePolicy.Enabled
}

// GetSeparateWindow get manually separate fault window duration. unit: millisecond
func GetSeparateWindow() int64 {
	return int64(config.ManuallySeparatePolicy.Separate.FaultWindowHours * constant.HoursToMilliseconds)
}

// GetSeparateThreshold get manually separate fault threshold
func GetSeparateThreshold() int {
	return config.ManuallySeparatePolicy.Separate.FaultThreshold
}

// GetReleaseDuration get manually separate release duration. unit: millisecond
func GetReleaseDuration() int64 {
	if IsReleaseEnable() {
		return int64(config.ManuallySeparatePolicy.Release.FaultFreeHours * constant.HoursToMilliseconds)
	}
	return NotRelease
}

// SetManualSeparatePolicy set manually separate policy config
func SetManualSeparatePolicy(policy ManuallySeparatePolicy) {
	config.ManuallySeparatePolicy = policy
}

// IsReleaseEnable check manually separate release enable
func IsReleaseEnable() bool {
	return config.ManuallySeparatePolicy.Release.FaultFreeHours != NotRelease
}

// Check manually separate policy config
func Check(policy ManuallySeparatePolicy) error {
	if policy.Separate.FaultWindowHours < MinFaultWindowHours || policy.Separate.FaultWindowHours > MaxFaultWindowHours {
		return fmt.Errorf("fault_window_hours must be in [%d, %d]", MinFaultWindowHours, MaxFaultWindowHours)
	}
	if policy.Separate.FaultThreshold < MinFaultThreshold || policy.Separate.FaultThreshold > MaxFaultThreshold {
		return fmt.Errorf("fault_threshold must be in [%d, %d]", MinFaultThreshold, MaxFaultThreshold)
	}
	if policy.Release.FaultFreeHours == NotRelease {
		return nil
	}
	if policy.Release.FaultFreeHours < MinFaultFreeHours || policy.Release.FaultFreeHours > MaxFaultFreeHours {
		return fmt.Errorf("fault_free_hours must be in [%d, %d] or %d", MinFaultFreeHours, MaxFaultFreeHours, NotRelease)
	}
	return nil
}
