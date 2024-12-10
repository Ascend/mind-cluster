/* Copyright(C) 2022. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package main is the main entry.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"ascend-common/common-utils/hwlog"
	"ascend-docker-runtime/hook/process"
	"ascend-docker-runtime/mindxcheckutils"
)

const (
	loggingPrefix = "ascend-docker-hook"
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			log.Fatal(err)
		}
	}()
	log.SetPrefix(loggingPrefix)

	ctx, _ := context.WithCancel(context.Background())
	if err := process.InitLogModule(ctx); err != nil {
		log.Fatal(err)
	}
	logPrefixWords, err := mindxcheckutils.GetLogPrefix()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := mindxcheckutils.ChangeRuntimeLogMode("hook-run-"); err != nil {
			fmt.Println("defer changeFileMode function failed")
		}
	}()
	hwlog.RunLog.Infof("%v ascend docker hook starting, try to setup container", logPrefixWords)
	if !mindxcheckutils.StringChecker(strings.Join(os.Args, " "), 0,
		process.MaxCommandLength, mindxcheckutils.DefaultWhiteList+" ") {
		hwlog.RunLog.Errorf("%v ascend docker hook failed", logPrefixWords)
		log.Fatal("command error")
	}
	if err := process.DoPrestartHook(); err != nil {
		hwlog.RunLog.Errorf("%v ascend docker hook failed: %#v", logPrefixWords, err)
		log.Fatal(fmt.Errorf("failed in runtime.doProcess: %#v", err))
	}
}
