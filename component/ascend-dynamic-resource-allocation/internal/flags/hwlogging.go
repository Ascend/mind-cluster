/*
 * Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at

 * http://www.apache.org/licenses/LICENSE-2.0

 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package flags

import (
	"context"
	"flag"
	"fmt"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"ascend-dynamic-resource-allocation/pkg/consts"
)

// HWLoggingConfig holds flags for the hwlog run logger.
type HWLoggingConfig struct {
	hwLogConfig hwlog.LogConfig
}

// RegisterFlags register logging config flags using standard flag package
func (l *HWLoggingConfig) RegisterFlags() {
	flag.StringVar(&l.hwLogConfig.LogFileName,
		"logFile",
		api.DefaultLogFile,
		"The log file path, if the file size exceeds 20MB, will be rotate.")

	flag.IntVar(&l.hwLogConfig.LogLevel,
		"logLevel",
		consts.DefaultLogLevel,
		"Log level, -1-debug, 0-info, 1-warning, 2-error, 3-critical(default 0).")

	flag.IntVar(&l.hwLogConfig.MaxAge,
		"maxAge",
		consts.DefaultLogMaxAge,
		"Maximum number of days for backup run log files, range [7, 700] days.")

	flag.IntVar(&l.hwLogConfig.MaxBackups,
		"maxBackups",
		hwlog.DefaultBackups,
		"Maximum number of backup log files, range is (0, 180].")
}

// InitLogModule initializes the hwlog run logger with parameters from flags
func (l *HWLoggingConfig) InitLogModule(ctx context.Context) error {
	hwLogConfig := hwlog.LogConfig{
		LogFileName: l.hwLogConfig.LogFileName,
		LogLevel:    l.hwLogConfig.LogLevel,
		MaxBackups:  l.hwLogConfig.MaxBackups,
		MaxAge:      l.hwLogConfig.MaxAge,
	}
	if err := hwlog.InitRunLogger(&hwLogConfig, ctx); err != nil {
		fmt.Printf("log init failed, error is %v\n", err)
		return err
	}
	return nil
}
