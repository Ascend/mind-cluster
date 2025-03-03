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
	"context"
	"errors"
	"fmt"

	"github.com/influxdata/telegraf"

	"ascend-common/common-utils/hwlog"
)

// the method mapping table (avoid rebuilding with every call)
var (
	logFuncs  = map[Level]logFunc{}
	logfFuncs = map[Level]logfFunc{}
)

const (
	// Debug Level
	Debug Level = iota - 1
	// Info Level
	Info
	// Warn Level
	Warn
	// Error Level
	Error

	// PrometheusPlatform Prometheus platform
	PrometheusPlatform = "Prometheus"
	// TelegrafPlatform Telegraf platform
	TelegrafPlatform = "Telegraf"
)

// HwLogConfig default log file
var HwLogConfig = &hwlog.LogConfig{
	LogFileName:   defaultLogFile,
	ExpiredTime:   hwlog.DefaultExpiredTime,
	CacheSize:     hwlog.DefaultCacheSize,
	MaxLineLength: maxLogLineLength,
}

// Level log level
type Level int

// logFunc log function
type logFunc func(ctx context.Context, args ...interface{})

// logfFunc logf function
type logfFunc func(ctx context.Context, format string, args ...interface{})

var (
	// Logger Unified log printer
	Logger UnifiedLogger
)

// InitLogger initialize the log manager
func InitLogger(platform string) error {

	if platform == TelegrafPlatform {
		Logger = &telegrafLogger{}
		HwLogConfig.LogFileName = defaultTelegrafLogPath
	} else if platform == PrometheusPlatform {
		Logger = &generalLogger{}
	} else {
		return errors.New("platform is not supported:" + platform)
	}

	if err := hwlog.InitRunLogger(HwLogConfig, context.Background()); err != nil {
		fmt.Printf("hwlog init failed, error is %v\n", err)
		return err
	}

	logFuncs = map[Level]logFunc{
		Debug: hwlog.RunLog.DebugWithCtx,
		Info:  hwlog.RunLog.InfoWithCtx,
		Warn:  hwlog.RunLog.WarnWithCtx,
		Error: hwlog.RunLog.ErrorWithCtx,
	}

	logfFuncs = map[Level]logfFunc{
		Debug: hwlog.RunLog.DebugfWithCtx,
		Info:  hwlog.RunLog.InfofWithCtx,
		Warn:  hwlog.RunLog.WarnfWithCtx,
		Error: hwlog.RunLog.ErrorfWithCtx,
	}
	return nil
}

// LogOptions options for log
type LogOptions struct {
	Domain    string
	ID        interface{}
	MaxCounts int
}

// Config config for telegraf
type Config struct {
	Acc telegraf.Accumulator
}

// UnifiedLogger unified logger interface
type UnifiedLogger interface {
	// DynamicConfigure configure the logger
	DynamicConfigure(Config)

	Log(level Level, args ...interface{})
	Logf(level Level, format string, args ...interface{})
	LogfWithOptions(level Level, opts LogOptions, format string, args ...interface{})
}
