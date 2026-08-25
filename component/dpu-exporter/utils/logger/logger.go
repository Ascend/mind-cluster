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

// Package logger provides a thin logging wrapper for dpu-exporter.
// It delegates to the ascend-common hwlog package for consistent log output
// across all mind-cluster components.
package logger

import (
	"context"
	"fmt"
	"log"

	"ascend-common/common-utils/hwlog"
)

const (
	maxLogLineLength = 1024
	// DefaultLogFile is the default log file path for dpu-exporter
	DefaultLogFile = "/var/log/mindx-dl/dpu-exporter/dpu-exporter.log"
)

// HwLogConfig is the shared log configuration used by hwlog
var HwLogConfig = &hwlog.LogConfig{
	LogFileName:   DefaultLogFile,
	ExpiredTime:   hwlog.DefaultExpiredTime,
	CacheSize:     hwlog.DefaultCacheSize,
	MaxAge:        hwlog.DefaultMinSaveAge,
	MaxBackups:    hwlog.DefaultBackups,
	MaxLineLength: maxLogLineLength,
}

// InitLogger initializes the hwlog run logger
func InitLogger() error {
	if err := hwlog.InitRunLogger(HwLogConfig, context.Background()); err != nil {
		return fmt.Errorf("hwlog init failed: %w", err)
	}
	return nil
}

// Debug logs at debug level
func Debug(args ...interface{}) {
	hwlog.RunLog.Debug(args...)
}

// Info logs at info level
func Info(args ...interface{}) {
	hwlog.RunLog.Info(args...)
}

// Warn logs at warn level
func Warn(args ...interface{}) {
	hwlog.RunLog.Warn(args...)
}

// Error logs at error level
func Error(args ...interface{}) {
	hwlog.RunLog.Error(args...)
}

// Debugf logs at debug level with format
func Debugf(format string, args ...interface{}) {
	hwlog.RunLog.Debugf(format, args...)
}

// Infof logs at info level with format
func Infof(format string, args ...interface{}) {
	hwlog.RunLog.Infof(format, args...)
}

// Warnf logs at warn level with format
func Warnf(format string, args ...interface{}) {
	hwlog.RunLog.Warnf(format, args...)
}

// Errorf logs at error level with format
func Errorf(format string, args ...interface{}) {
	hwlog.RunLog.Errorf(format, args...)
}

// SetStdLog configures the standard log package to use hwlog writer
func SetStdLog() {
	log.SetOutput(&hwlog.SelfLogWriter{})
}
