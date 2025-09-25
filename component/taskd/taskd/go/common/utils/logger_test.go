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

// Package utils is to provide go runtime utils
package utils

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"

	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
	"taskd/common/constant"
)

var (
	mockErr             = errors.New("mock error")
	testAbsoluteLogPath = "/tmp/test.log"
	testRelativeLogPath = "./test.log"
)

func buildLoggerNormalTestCases() []struct {
	name           string
	logFileName    string
	envValue       string
	setupMocks     func() *gomonkey.Patches
	expectError    bool
	expectedErrMsg string
	validateResult func(*testing.T, *hwlog.LogConfig)
} {
	return []struct {
		name           string
		logFileName    string
		envValue       string
		setupMocks     func() *gomonkey.Patches
		expectError    bool
		expectedErrMsg string
		validateResult func(*testing.T, *hwlog.LogConfig)
	}{
		{
			name:        "normal case, log env is nil, use default path",
			logFileName: testAbsoluteLogPath,
			envValue:    "",
			setupMocks: func() *gomonkey.Patches {
				patches := gomonkey.ApplyFuncReturn(filepath.Abs, "/default/path/test.log", nil).
					ApplyFuncReturn(utils.CheckPath, "/default/path/test.log", nil)
				return patches
			},
			expectError: false,
			validateResult: func(t *testing.T, config *hwlog.LogConfig) {
				assert.Equal(t, "/default/path/test.log", config.LogFileName)
			},
		},
		{
			name:        "normal case, log env is not nil, use define path",
			logFileName: testAbsoluteLogPath,
			envValue:    "/custom/log/path",
			setupMocks: func() *gomonkey.Patches {
				patches := gomonkey.ApplyFuncReturn(filepath.Abs, "/custom/log/path/test.log", nil).
					ApplyFuncReturn(utils.CheckPath, "/custom/log/path/test.log", nil)
				return patches
			},
			expectError: false,
			validateResult: func(t *testing.T, config *hwlog.LogConfig) {
				assert.Equal(t, "/custom/log/path/test.log", config.LogFileName)
			},
		},
	}
}

func buildLoggerAbnormalTestCases() []struct {
	name           string
	logFileName    string
	envValue       string
	setupMocks     func() *gomonkey.Patches
	expectError    bool
	expectedErrMsg string
	validateResult func(*testing.T, *hwlog.LogConfig)
} {
	return []struct {
		name           string
		logFileName    string
		envValue       string
		setupMocks     func() *gomonkey.Patches
		expectError    bool
		expectedErrMsg string
		validateResult func(*testing.T, *hwlog.LogConfig)
	}{
		{
			name:           "abnormal case, input is empty",
			setupMocks:     func() *gomonkey.Patches { return nil },
			expectError:    true,
			expectedErrMsg: "logFileName is empty",
			validateResult: nil,
		},
		{
			name:        "abnormal case, get absolute path failed",
			logFileName: testRelativeLogPath,
			setupMocks: func() *gomonkey.Patches {
				patches := gomonkey.ApplyFuncReturn(filepath.Abs, "", fmt.Errorf("failed to get absolute path"))
				return patches
			},
			expectError:    true,
			expectedErrMsg: "get abs log file path error: failed to get absolute path",
			validateResult: nil,
		},
		{
			name:        "abnormal case, check path failed",
			logFileName: testAbsoluteLogPath,
			setupMocks: func() *gomonkey.Patches {
				patches := gomonkey.ApplyFuncReturn(filepath.Abs, "/default/path/test.log", nil).
					ApplyFuncReturn(utils.CheckPath, "", fmt.Errorf("path check failed"))
				return patches
			},
			expectError:    true,
			expectedErrMsg: "check log file path error: path check failed",
			validateResult: nil,
		},
	}
}

func dealEnv(t *testing.T, envValue string) error {
	if envValue != "" {
		t.Setenv(constant.LogFilePathEnv, envValue)
	} else {
		err := os.Unsetenv(constant.LogFilePathEnv)
		if err != nil {
			return errors.New("unset env failed")
		}
	}
	return nil
}
func TestGetLoggerConfigWithFileName(t *testing.T) {
	for _, tt := range append(buildLoggerNormalTestCases(), buildLoggerAbnormalTestCases()...) {
		t.Run(tt.name, func(t *testing.T) {
			if err := dealEnv(t, tt.envValue); err != nil {
				return
			}

			var patches *gomonkey.Patches
			if tt.setupMocks != nil {
				patches = tt.setupMocks()
			}
			if patches != nil {
				defer patches.Reset()
			}

			logConfig, err := GetLoggerConfigWithFileName(tt.logFileName)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, logConfig)
				if tt.expectedErrMsg != "" {
					assert.Contains(t, err.Error(), tt.expectedErrMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, logConfig)
				if tt.validateResult != nil {
					tt.validateResult(t, logConfig)
				}
			}
		})
	}
}

func TestInitHwLogger(t *testing.T) {
	t.Run("normal test, init hwlog success", func(t *testing.T) {
		patches := gomonkey.ApplyFuncReturn(GetLoggerConfigWithFileName, &hwlog.LogConfig{}, nil).
			ApplyFuncReturn(hwlog.InitRunLogger, nil)
		defer patches.Reset()

		err := InitHwLogger("test.log", context.Background())
		assert.NoError(t, err, "Expected no error when initializing hw logger successfully")
	})

	t.Run("abnormal test, get config failed", func(t *testing.T) {
		patches := gomonkey.ApplyFuncReturn(GetLoggerConfigWithFileName, &hwlog.LogConfig{}, mockErr)
		defer patches.Reset()

		err := InitHwLogger("invalid.log", nil)
		assert.Error(t, err, "Expected error when GetLoggerConfigWithFileName fails")
		assert.Contains(t, err.Error(), "hwlog init failed", "Error message should contain 'hwlog init failed'")
		assert.Contains(t, err.Error(), mockErr.Error(), "Error message should contain the original error")
	})

	t.Run("abnormal test, init runlog failed", func(t *testing.T) {
		patches := gomonkey.ApplyFuncReturn(GetLoggerConfigWithFileName, &hwlog.LogConfig{}, nil).
			ApplyFuncReturn(hwlog.InitRunLogger, mockErr)
		defer patches.Reset()

		err := InitHwLogger("test.log", nil)
		assert.Error(t, err, "Expected error when hwlog.InitRunLogger fails")
		assert.Contains(t, err.Error(), "hwlog init failed", "Error message should contain 'hwlog init failed'")
		assert.Contains(t, err.Error(), mockErr.Error(), "Error message should contain the original error")
	})
}
