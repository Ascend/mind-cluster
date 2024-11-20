/* Copyright(C) 2023. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package main
package main

import (
	"context"
	"flag"
	"fmt"

	"huawei.com/npu-exporter/v5/common-utils/hwlog"

	"nodeD/pkg"
)

const (
	defaultLogFile = "/var/log/mindx-dl/noded/noded.log"
)

var (
	hwLogConfig = &hwlog.LogConfig{LogFileName: defaultLogFile}
	version     bool
	// BuildVersion build version
	BuildVersion string
	// BuildName build name
	BuildName string
	// heartbeatInterval send Heartbeat Interval
	heartbeatInterval int
)

func main() {
	flag.Parse()

	if version {
		fmt.Printf("%s version: %s \n", BuildName, BuildVersion)
		return
	}

	// init hwlog
	if err := hwlog.InitRunLogger(hwLogConfig, context.Background()); err != nil {
		fmt.Printf("hwlog init failed, error is %v\n", err)
		return
	}

	hwlog.RunLog.Infof("%s starting and the version is %s", BuildName, BuildVersion)

	if err := pkg.ValidHeartbeatInterval(heartbeatInterval); err != nil {
		hwlog.RunLog.Errorf("validate heartbeat interval failed: %v", err)
		return
	}

	if err := pkg.SendHeartbeat(heartbeatInterval); err != nil {
		hwlog.RunLog.Errorf("send heartbeat failed: %v", err)
		return
	}
}

func init() {
	flag.BoolVar(&version, "version", false, "the version of the program")

	flag.IntVar(&heartbeatInterval, "heartbeatInterval", pkg.DefaultHeartbeatInterval,
		"Interval of sending heartbeat")

	// hwlog configuration
	flag.IntVar(&hwLogConfig.LogLevel, "logLevel", 0,
		"Log level, -1-debug, 0-info, 1-warning, 2-error, 3-critical(default 0)")
	flag.IntVar(&hwLogConfig.MaxAge, "maxAge", hwlog.DefaultMinSaveAge,
		"Maximum number of days for backup run log files, range [7, 700] days")
	flag.StringVar(&hwLogConfig.LogFileName, "logFile", defaultLogFile,
		"Run log file path. if the file size exceeds 20MB, will be rotated")
	flag.IntVar(&hwLogConfig.MaxBackups, "maxBackups", hwlog.DefaultMaxBackups,
		"Maximum number of backup operation logs, range is (0, 30]")
}
