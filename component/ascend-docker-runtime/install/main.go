/* Copyright(C) 2021. Huawei Technologies Co.,Ltd. All rights reserved.
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
	"log"
	"os"
	"strings"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"ascend-docker-runtime/install/process"
	"ascend-docker-runtime/mindxcheckutils"
)

const commonTemplate = `{
        "runtimes":     {
                "ascend":       {
                        "path": "%s",
                        "runtimeArgs":  []
                }
        },
        "default-runtime":      "ascend"
}`

const noDefaultTemplate = `{
        "runtimes":     {
                "ascend":       {
                        "path": "%s",
                        "runtimeArgs":  []
                }
        }
}`

const (
	maxCommandLength         = 65535
	logPath                  = api.InstallHelperRunLogPath
	minCommandLength         = 2
	installSceneIndexFromEnd = 2
)

var reserveDefaultRuntime = false

func main() {
	ctx, _ := context.WithCancel(context.Background())
	if err := initLogModule(ctx); err != nil {
		log.Fatal(err)
	}
	logPrefixWords, err := mindxcheckutils.GetLogPrefix()
	if err != nil {
		log.Fatal(err)
	}
	hwlog.RunLog.Infof("%v start running script", logPrefixWords)

	if !mindxcheckutils.StringChecker(strings.Join(os.Args, " "), 0,
		maxCommandLength, mindxcheckutils.DefaultWhiteList+" ") {
		hwlog.RunLog.Errorf("%v check command failed, maybe command contains illegal char", logPrefixWords)
		log.Fatalf("command error, please check %s for detail", logPath)
	}

	const helpMessage = "\tadd <config file path> <new config file path> " +
		"<docker-runtime path> <whether reserve default> <docker or containerd> <cgroup info>\n" +
		"\t rm <config file path> <new config file path> <docker or containerd> <whether reserve default>" +
		" <docker or containerd> <cgroup info>\n" + "\t -h help command"
	helpFlag := flag.Bool("h", false, helpMessage)
	flag.Parse()
	if *helpFlag {
		_, err := fmt.Println(helpMessage)
		log.Fatalf("need help, error: %v", err)
	}
	command := flag.Args()
	if len(command) == 0 {
		log.Fatalf("error param")
	}
	var behavior string
	if len(command) < minCommandLength {
		log.Fatalf("error param")
	}
	installScene := command[len(command)-installSceneIndexFromEnd]
	if installScene == process.InstallSceneDocker || installScene == process.InstallSceneIsula {
		behavior, err = process.DockerProcess(command)
	} else if installScene == process.InstallSceneContainerd {
		behavior, err = process.ContainerdProcess(command)
	} else {
		hwlog.RunLog.Errorf("error param: %v", command[len(command)-1])
		log.Fatalf("error param: %v", command[len(command)-1])
	}
	if err != nil {
		hwlog.RunLog.Errorf("%v run script failed: %v", logPrefixWords, err)
		log.Fatal(fmt.Errorf("error in installation, err is %v", err))
	}
	hwlog.RunLog.Infof("%v run %v success", logPrefixWords, behavior)
}

func initLogModule(ctx context.Context) error {
	const backups = 2
	const logMaxAge = 365
	const fileMaxSize = 2
	logConfig := hwlog.LogConfig{
		LogFileName: logPath,
		LogLevel:    0,
		MaxBackups:  backups,
		MaxAge:      logMaxAge,
		OnlyToFile:  true,
		FileMaxSize: fileMaxSize,
	}
	if err := hwlog.InitRunLogger(&logConfig, ctx); err != nil {
		fmt.Printf("hwlog init failed, error is %v", err)
		return err
	}
	return nil
}
