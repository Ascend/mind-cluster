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

// Package main a main package for cgo api
package main

import "C"
import (
	"context"
	"fmt"

	"taskd/common/constant"
	"ascend-common/common-utils/hwlog"
)

func init() {
	hwLogConfig := hwlog.LogConfig{
		LogFileName:   constant.DefaultLogFile,
		LogLevel:      constant.DefaultLogLevel,
		MaxBackups:    constant.DefaultMaxBackups,
		MaxAge:        constant.DefaultMaxAge,
		MaxLineLength: constant.DefaultMaxLineLength,
	}
	if err := hwlog.InitRunLogger(&hwLogConfig, context.Background()); err != nil {
		fmt.Printf("hwlog init failed, error is %v\n", err)
	}
}

func main() {
}
