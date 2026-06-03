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

// Package common implements common utils for ascend device plugin
package common

import (
	"encoding/json"
	"sync"

	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
)

// HangDetectionConfig hang detection configuration
type HangDetectionConfig struct {
	HangDetection HangDetection `json:"HangDetection"`
}

// HangDetection hang detection settings
type HangDetection struct {
	Enabled   bool          `json:"Enabled"`
	Threshold HangThreshold `json:"Threshold"`
}

// HangThreshold hang detection threshold settings
type HangThreshold struct {
	AICoreUtilization int32 `json:"AICoreUtilization"`
	HbmMemoryDelta    int32 `json:"HbmMemoryDelta"`
	TrafficDelta      int32 `json:"TrafficDelta"`
	CPUTimeDelta      int32 `json:"CPUTimeDelta"`
	DetectDuration    int32 `json:"DetectDuration"`
}

const (
	hangDetectionConfigPath = "/usr/local/hangDetectionConfig.json"
	defaultUtilizationLow   = 5   // unit: %
	defaultMemoryLow        = 0   // unit: %
	defaultTrafficLow       = 100 // unit: pkt/min
	defaultCPUTimeLow       = 5   // unit: second
	defaultDetectDuration   = 5   // unit: times
)

var hangConfig HangDetectionConfig = HangDetectionConfig{
	HangDetection: HangDetection{
		Enabled: true,
		Threshold: HangThreshold{
			AICoreUtilization: defaultUtilizationLow,
			HbmMemoryDelta:    defaultMemoryLow,
			TrafficDelta:      defaultTrafficLow,
			CPUTimeDelta:      defaultCPUTimeLow,
			DetectDuration:    defaultDetectDuration,
		},
	},
}
var configLock sync.RWMutex

// LoadHangDetectionConfigFromFile loads hang detection config from file
func LoadHangDetectionConfigFromFile() {
	data, err := utils.LoadFile(hangDetectionConfigPath)
	if err != nil {
		hwlog.RunLog.Errorf("load hang detection config file failed: %v，set default config", err)
		return
	}
	loadHangDetectionConfigFromBytes(data)
}

// loadHangDetectionConfigFromBytes loads hang detection config from bytes
func loadHangDetectionConfigFromBytes(data []byte) {
	var cfg HangDetectionConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		hwlog.RunLog.Errorf("unmarshal hang detection config failed: %v, set default config", err)
		return
	}
	if cfg.HangDetection.Threshold.AICoreUtilization < 0 {
		hwlog.RunLog.Warnf("utilization threshold < 0, set default threshold: %d%%", defaultUtilizationLow)
		cfg.HangDetection.Threshold.AICoreUtilization = defaultUtilizationLow
	}
	if cfg.HangDetection.Threshold.HbmMemoryDelta < 0 {
		hwlog.RunLog.Warnf("memory threshold < 0, set default threshold: %d%%", defaultMemoryLow)
		cfg.HangDetection.Threshold.HbmMemoryDelta = defaultMemoryLow
	}
	if cfg.HangDetection.Threshold.TrafficDelta < 0 {
		hwlog.RunLog.Warnf("traffic threshold < 0, set default threshold: %d pkt/min", defaultTrafficLow)
		cfg.HangDetection.Threshold.TrafficDelta = defaultTrafficLow
	}
	if cfg.HangDetection.Threshold.CPUTimeDelta < 0 {
		hwlog.RunLog.Warnf("cpu time threshold < 0, set default threshold: %d s", defaultCPUTimeLow)
		cfg.HangDetection.Threshold.CPUTimeDelta = defaultCPUTimeLow
	}
	if cfg.HangDetection.Threshold.DetectDuration <= 0 {
		hwlog.RunLog.Warnf("detect duration threshold <= 0, set default value: %d times", defaultDetectDuration)
		cfg.HangDetection.Threshold.DetectDuration = defaultDetectDuration
	}
	configLock.Lock()
	defer configLock.Unlock()
	hangConfig = cfg
	hwlog.RunLog.Infof("hang detection config loaded: %v", cfg.HangDetection)
}

// IsHangDetectionEnabled returns whether hang detection is enabled
func IsHangDetectionEnabled() bool {
	configLock.RLock()
	defer configLock.RUnlock()
	return hangConfig.HangDetection.Enabled
}

// GetHangDetectionThreshold returns a copy of current hang detection threshold
func GetHangDetectionThreshold() HangThreshold {
	configLock.RLock()
	defer configLock.RUnlock()
	return hangConfig.HangDetection.Threshold
}
