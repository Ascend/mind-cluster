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

	"ascend-common/common-utils/hwlog"
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

// silent fault detection config bounds and defaults
const (
	// DetectInterval detect loop interval. unit: second
	DetectInterval = 60

	// DefaultMinTaskCards default min task cards
	DefaultMinTaskCards = 16
	// DefaultConsecutiveTimes default consecutive times
	DefaultConsecutiveTimes = 3
	// DefaultHwWindowSeconds default hardware fault window seconds. unit: second
	DefaultHwWindowSeconds = 30
	// DefaultWindowSeconds default detect window seconds. unit: second
	DefaultWindowSeconds = 10800
	// DefaultFaultFreeSeconds default fault free seconds (48 hours). unit: second
	DefaultFaultFreeSeconds = 48 * 60 * 60

	// MinSilentTaskCards min task cards lower bound (exclusive), valid value must be greater than this
	MinSilentTaskCards = 0
	// MaxSilentTaskCards min task cards upper bound (exclusive), valid value must be less than this
	MaxSilentTaskCards = 10000000
	// MinConsecutiveTimes min consecutive times lower bound (exclusive)
	MinConsecutiveTimes = 0
	// MaxConsecutiveTimes min consecutive times upper bound (exclusive)
	MaxConsecutiveTimes = 10000
	// MinHwWindowSeconds min hardware fault window seconds lower bound (exclusive). unit: second
	MinHwWindowSeconds = 3
	// MaxHwWindowSeconds max hardware fault window seconds upper bound (exclusive). unit: second
	MaxHwWindowSeconds = 24 * 60 * 60
	// MinWindowSeconds min detect window seconds lower bound (inclusive). unit: second
	MinWindowSeconds = 30
	// MaxWindowSeconds max detect window seconds upper bound (exclusive). unit: second
	MaxWindowSeconds = 365 * 24 * 60 * 60
	// SilentNotRelease silent fault not auto release
	SilentNotRelease = -1
	// MaxFaultFreeSeconds max fault free seconds upper bound (exclusive). unit: second, -1 is also allowed
	MaxFaultFreeSeconds = 365 * 24 * 60 * 60
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

// GetSilentReleaseSeconds get silent fault release duration. unit: second. -1 means no auto release.
// The value is normalized to a default when absent or invalid during config loading, so 0 is never stored here.
func GetSilentReleaseSeconds() int64 {
	return int64(config.SilentFaultPolicy.Release.FaultFreeSecond)
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

// NormalizeSilentFault validates every parameter and resets each absent (zero) or out-of-range
// value to its default, logging a warning for each parameter that falls back to the default.
func NormalizeSilentFault(policy *SilentFaultPolicy) {
	if policy.Detect.MinTaskCards <= MinSilentTaskCards || policy.Detect.MinTaskCards >= MaxSilentTaskCards {
		hwlog.RunLog.Warnf("silent fault min_task_cards=%d is invalid, use default %d",
			policy.Detect.MinTaskCards, DefaultMinTaskCards)
		policy.Detect.MinTaskCards = DefaultMinTaskCards
	}
	if policy.Detect.ConsecutiveTimes <= MinConsecutiveTimes || policy.Detect.ConsecutiveTimes >= MaxConsecutiveTimes {
		hwlog.RunLog.Warnf("silent fault consecutive_times=%d is invalid, use default %d",
			policy.Detect.ConsecutiveTimes, DefaultConsecutiveTimes)
		policy.Detect.ConsecutiveTimes = DefaultConsecutiveTimes
	}
	if policy.Detect.HardwareFaultWindowSecond <= MinHwWindowSeconds ||
		policy.Detect.HardwareFaultWindowSecond >= MaxHwWindowSeconds {
		hwlog.RunLog.Warnf("silent fault hardware_fault_window_seconds=%d is invalid, use default %d",
			policy.Detect.HardwareFaultWindowSecond, DefaultHwWindowSeconds)
		policy.Detect.HardwareFaultWindowSecond = DefaultHwWindowSeconds
	}
	if policy.Detect.WindowSecond < MinWindowSeconds || policy.Detect.WindowSecond >= MaxWindowSeconds {
		hwlog.RunLog.Warnf("silent fault window_seconds=%d is invalid, use default %d",
			policy.Detect.WindowSecond, DefaultWindowSeconds)
		policy.Detect.WindowSecond = DefaultWindowSeconds
	}
	if !isValidFaultFreeSeconds(policy.Release.FaultFreeSecond) {
		hwlog.RunLog.Warnf("silent fault fault_free_seconds=%d is invalid, use default %d",
			policy.Release.FaultFreeSecond, DefaultFaultFreeSeconds)
		policy.Release.FaultFreeSecond = DefaultFaultFreeSeconds
	}
}

// isValidFaultFreeSeconds reports whether the fault_free_seconds value is valid: -1 means no auto
// release; otherwise it must fall in (0, MaxFaultFreeSeconds).
func isValidFaultFreeSeconds(second int) bool {
	if second == SilentNotRelease {
		return true
	}
	return second > 0 && second < MaxFaultFreeSeconds
}
