/* Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package logger for general collector
package logger

import (
	"errors"
	"testing"

	"ascend-common/common-utils/hwlog"
)

// TestInitLogger tests the InitLogger function
func TestInitLogger(t *testing.T) {
	tests := []struct {
		name     string
		platform string
		expected error
	}{
		{
			name:     "Telegraf Platform",
			platform: TelegrafPlatform,
			expected: nil,
		},
		{
			name:     "Prometheus Platform",
			platform: PrometheusPlatform,
			expected: nil,
		},
		{
			name:     "Unsupported Platform",
			platform: "Unsupported",
			expected: errors.New("platform is not supported:Unsupported"),
		},
	}

	HwLogConfig.LogLevel = 0
	HwLogConfig.MaxBackups = hwlog.DefaultMaxBackups
	HwLogConfig.LogFileName = defaultLogFile
	HwLogConfig.MaxAge = hwlog.DefaultMinSaveAge

	var noExistLevel Level = 5
	var args = "mock"
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := InitLogger(tt.platform)
			if tt.expected == nil && err != nil {
				t.Errorf("InitLogger(%s) = %v, want %v", tt.platform, err, tt.expected)
			} else if tt.expected != nil && err.Error() != tt.expected.Error() {
				t.Errorf("InitLogger(%s) = %v, want %v", tt.platform, err, tt.expected)
			}

			Logger.Log(Debug, args)
			Logger.Log(Info, args)
			Logger.Log(Warn, args)
			Logger.Log(noExistLevel, args)
			Logger.LogfWithOptions(Debug, LogOptions{}, "test logf with options %s", "arg")

			Logger.Logf(Debug, args)
			Logger.Logf(Info, args)
			Logger.Logf(Warn, args)
			Logger.Logf(Error, args)
			Logger.Logf(noExistLevel, args)
			Logger.LogfWithOptions(Debug, LogOptions{}, "test logf with options %s", "arg")

		})
	}
}
