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
	"taskd/common/constant"
)

// InitHwLogger init hwlogger
func InitHwLogger(logFileName string, ctx context.Context) error {
	hwLogConfig := GetLoggerConfigWithFileName(logFileName)
	if err := hwlog.InitRunLogger(&hwLogConfig, context.Background()); err != nil {
		fmt.Printf("hwlog init failed, error is %v\n", err)
		return err
	}
	return nil
}

// GetLoggerConfigWithFileName get logger config with log file name
func GetLoggerConfigWithFileName(logFileName string) hwlog.LogConfig {
	var logFile string
	logFilePath := os.Getenv(constant.LogFilePathEnv)
	if logFilePath == "" {
		logFile = constant.DefaultLogFilePath + logFileName
	} else {
		logFile = filepath.Join(logFilePath, logFileName)
	}
	hwLogConfig := hwlog.LogConfig{
		LogFileName:   logFile,
		LogLevel:      constant.DefaultLogLevel,
		MaxBackups:    constant.DefaultMaxBackups,
		MaxAge:        constant.DefaultMaxAge,
		MaxLineLength: constant.DefaultMaxLineLength,
		// do not print to screen to avoid influence training log
		OnlyToFile: true,
	}
	return hwLogConfig
}
