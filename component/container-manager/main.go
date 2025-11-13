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

// Package main component container-manager main function
package main

import (
	"context"
	"flag"
	"fmt"

	"ascend-common/common-utils/hwlog"
)

const (
	maxAge           = 7
	maxBackups       = 30
	maxLogLineLength = 1024
)

var (
	version bool

	logPath       string
	logLevel      int
	logMaxAge     int
	logMaxBackups int

	// BuildName show component name
	BuildName string
	// BuildVersion show component version
	BuildVersion string
)

func init() {
	flag.BoolVar(&version, "version", false, "Output version information")
	flag.StringVar(&logPath, "logPath", "/var/log/mindx-dl/container-manager/",
		"The log file path. If the file size exceeds 20MB, will be dumped")
	flag.IntVar(&logLevel, "logLevel", 0, "Log level, -1-debug, 0-info, 1-warning, 2-error, 3-critical")
	flag.IntVar(&logMaxAge, "maxAge", maxAge, "Maximum number of days for backup log files, range is [7, 700]")
	flag.IntVar(&logMaxBackups, "maxBackups", maxBackups, "Maximum number of backup log files, range is (0, 30]")
}

func main() {
	flag.Parse()
	if version {
		fmt.Printf("%s version: %s\n", BuildName, BuildVersion)
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := initLog(ctx); err != nil {
		return
	}
}

func initLog(ctx context.Context) error {
	hwLogConfig := hwlog.LogConfig{
		LogFileName:   logPath,
		LogLevel:      logLevel,
		MaxAge:        logMaxAge,
		MaxBackups:    logMaxBackups,
		MaxLineLength: maxLogLineLength,
	}
	if err := hwlog.InitRunLogger(&hwLogConfig, ctx); err != nil {
		fmt.Printf("init log failed, error: %v\n", err)
		return err
	}
	hwlog.RunLog.Info("init log success")
	return nil
}
