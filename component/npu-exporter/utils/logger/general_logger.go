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
	"fmt"

	"ascend-common/common-utils/hwlog"
)

const (
	maxLogLineLength = 1024
	defaultLogFile   = "/var/log/mindx-dl/npu-exporter/npu-exporter.log"
)

type generalLogger struct {
}

// DynamicConfigure configures the logger
func (c *generalLogger) DynamicConfigure(Config) {
}

// Log logs with specified level
func (c *generalLogger) Log(level Level, args ...interface{}) {
	fn, ok := logFuncs[level]
	if !ok {
		hwlog.RunLog.Warnf("unknown log level: %v", level)
		return
	}

	fn(nil, args...)
}

// Logf logs with specified level and format
func (c *generalLogger) Logf(level Level, format string, args ...interface{}) {
	fn, ok := logfFuncs[level]
	if !ok {
		hwlog.RunLog.Warnf("unknown log level: %v", level)
		return
	}

	fn(nil, format, args...)
}

func (c *generalLogger) LogfWithOptions(level Level, opts LogOptions, format string, args ...interface{}) {

	if opts.MaxCounts == 0 {
		opts.MaxCounts = hwlog.ProblemOccurMaxNumbers
	}

	if needPrint, extraErrLog := hwlog.IsNeedPrintWithSpecifiedCounts(opts.Domain, opts.ID, opts.MaxCounts); needPrint {
		format = fmt.Sprintf("%s %s", format, extraErrLog)
		fn, ok := logfFuncs[level]
		if !ok {
			hwlog.RunLog.Warnf("unknown log level: %v", level)
			return
		}

		fn(nil, format, args...)
	}
}
