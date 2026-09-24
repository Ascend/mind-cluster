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
	SilentFaultPolicy
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

// silent fault detection config bound and defaults
const (
	// DetectInterval detect loop interval. unit: second
	DetectInterval = 60

	// MinSilentTaskCards min task cards lower bound (exclusive), valid value must be greater than this
	MinSilentTaskCards = 0
	// MinConsecutiveTimes min consecutive times lower bound (exclusive)
	MinConsecutiveTimes = 0
	// MinHwWindowSeconds min hardware fault window seconds lower bound (exclusive). unit: second
	MinHwWindowSeconds = 3
	// MinWindowSeconds min detect window seconds lower bound (exclusive). unit: second
	MinWindowSeconds = 30
	// SilentNotRelease silent fault not auto release
	SilentNotRelease = -1
	// DefaultSilentReleaseSeconds default silent fault release duration (48 hours). unit: second
	DefaultSilentReleaseSeconds = 48 * 60 * 60
)

// SilentFaultPolicy silent fault policy config
type SilentFaultPolicy struct {
	Enabled bool `yaml:"enabled"`
	Detect  struct {
		MinTaskCards              int `yaml:"min_task_cards"`
		ConsecutiveTimes          int `yaml:"consecutive_times"`
		HardwareFaultWindowSecond int `yaml:"hardware_fault_window_seconds"`
		WindowSecond              int `yaml:"window_seconds"`
	} `yaml:"detect"`
	Release struct {
		FaultFreeSecond int `yaml:"fault_free_seconds"`
	} `yaml:"release"`
}

// GetSilentFaultEnabled get silent fault detection enabled
func GetSilentFaultEnabled() bool {
	return config.SilentFaultPolicy.Enabled
}

// GetMinTaskCards get min task cards
func GetMinTaskCards() int {
	return config.SilentFaultPolicy.Detect.MinTaskCards
}

// GetConsecutiveTimes get consecutive times
func GetConsecutiveTimes() int {
	return config.SilentFaultPolicy.Detect.ConsecutiveTimes
}

// GetHwWindowSeconds get hardware fault window. unit: second
func GetHwWindowSeconds() int64 {
	return int64(config.SilentFaultPolicy.Detect.HardwareFaultWindowSecond)
}

// GetWindowSeconds get detect window. unit: second
func GetWindowSeconds() int64 {
	return int64(config.SilentFaultPolicy.Detect.WindowSecond)
}

// GetSilentReleaseSeconds get silent fault release duration. unit: second. -1 means no auto release, 0 means default 48 hours
func GetSilentReleaseSeconds() int64 {
	second := config.SilentFaultPolicy.Release.FaultFreeSecond
	if second == 0 {
		return DefaultSilentReleaseSeconds
	}
	return int64(second)
}

// GetDetectInterval get detect loop interval. unit: second
func GetDetectInterval() int64 {
	return DetectInterval
}

// SetSilentFaultPolicy set silent fault policy config
func SetSilentFaultPolicy(policy SilentFaultPolicy) {
	config.SilentFaultPolicy = policy
}

// SilentFaultDetectChanged reports whether the detection parameters (Detect) differ from the
// currently loaded policy. Only the detection parameters affect judgment; the release config
// (auto release) is ignored. Used to clear cached detection events on hot-reload.
func SilentFaultDetectChanged(newPolicy SilentFaultPolicy) bool {
	return config.SilentFaultPolicy.Detect != newPolicy.Detect
}

// CheckSilentFault check silent fault policy config
func CheckSilentFault(policy SilentFaultPolicy) error {
	if policy.Detect.MinTaskCards <= MinSilentTaskCards {
		return fmt.Errorf("min_task_cards must be greater than %d", MinSilentTaskCards)
	}
	if policy.Detect.ConsecutiveTimes <= MinConsecutiveTimes {
		return fmt.Errorf("consecutive_times must be greater than %d", MinConsecutiveTimes)
	}
	if policy.Detect.HardwareFaultWindowSecond <= MinHwWindowSeconds {
		return fmt.Errorf("hardware_fault_window_seconds must be greater than %d", MinHwWindowSeconds)
	}
	if policy.Detect.WindowSecond <= MinWindowSeconds {
		return fmt.Errorf("window_seconds must be greater than %d", MinWindowSeconds)
	}
	if policy.Release.FaultFreeSecond == SilentNotRelease {
		return nil
	}
	// 0 means use default 48 hours; positive value means explicit release seconds
	if policy.Release.FaultFreeSecond < 0 {
		return fmt.Errorf("fault_free_seconds must be %d (no release) or a non-negative number, 0 means default 48 hours", SilentNotRelease)
	}
	return nil
}
