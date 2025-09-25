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
	"fmt"
	"os"
	"path/filepath"

	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
	"taskd/common/constant"
)

// InitHwLogger init hwlogger
func InitHwLogger(logFileName string, ctx context.Context) error {
	hwLogConfig, err := GetLoggerConfigWithFileName(logFileName)
	if err != nil {
		return fmt.Errorf("hwlog init failed, error is %v", err)
	}
	if err = hwlog.InitRunLogger(hwLogConfig, context.Background()); err != nil {
		return fmt.Errorf("hwlog init failed, error is %v", err)
	}
	return nil
}

// GetLoggerConfigWithFileName get logger config with log file name
func GetLoggerConfigWithFileName(logFileName string) (*hwlog.LogConfig, error) {
	if len(logFileName) == 0 {
		return nil, fmt.Errorf("logFileName is empty")
	}
	var logFile string
	logFilePath := os.Getenv(constant.LogFilePathEnv)
	if logFilePath == "" {
		logFile = constant.DefaultLogFilePath + logFileName
	} else {
		logFile = filepath.Join(logFilePath, logFileName)
	}
	absLogPath, err := filepath.Abs(logFile)
	if err != nil {
		return nil, fmt.Errorf("get abs log file path error: %v", err)
	}
	checkPath, err := utils.CheckPath(absLogPath)
	if err != nil {
		return nil, fmt.Errorf("check log file path error: %v", err)
	}
	hwLogConfig := &hwlog.LogConfig{
		LogFileName:   checkPath,
		LogLevel:      constant.DefaultLogLevel,
		MaxBackups:    constant.DefaultMaxBackups,
		MaxAge:        constant.DefaultMaxAge,
		MaxLineLength: constant.DefaultMaxLineLength,
		// do not print to screen to avoid influence training log
		OnlyToFile: true,
	}
	return hwLogConfig, nil
}
