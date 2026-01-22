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

// Package command run command
package command

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"ascend-common/common-utils/utils"
	"container-manager/pkg/common"
)

const (
	totalWidth = 100
	textWidth  = 65
	sideWidth  = 2

	pausingDuration  = 30
	pausedDuration   = 400
	resumingDuration = 30
)

var (
	border = "+" + strings.Repeat("=", totalWidth-sideWidth) + "+"
	format = "| %-27s: %-67s |\n"

	ctrStatusInfos = map[string]struct {
		statusDuration int64
		description    string
	}{
		common.StatusPausing: {
			pausingDuration,
			"Container pause may fail. Please manually delete the container",
		},
		common.StatusPaused: {
			pausedDuration,
			"Device hot reset may fail. Please check of device status and recovery are required",
		},
		common.StatusResuming: {
			resumingDuration,
			"The device has been recovered, but the container failed to be resumed. Please manually pull up the container",
		},
	}
)

type statusCmd struct {
}

// StatusCmd cmd 'status'
func StatusCmd() Command {
	return &statusCmd{}
}

// Name cmd name
func (cmd *statusCmd) Name() string {
	return "status"
}

// Description cmd description
func (cmd *statusCmd) Description() string {
	return "Display container status information and container abnormal information"
}

// BindFlag bind flag. If not needed, return false directly
func (cmd *statusCmd) BindFlag() bool {
	return false
}

// CheckParam check param
func (cmd *statusCmd) CheckParam() error {
	return nil
}

// InitLog init log
func (cmd *statusCmd) InitLog(ctx context.Context) error {
	return nil
}

// Execute execute cmd
func (cmd *statusCmd) Execute(ctx context.Context) error {
	if _, err := os.Stat(common.StatusInfoFile); err != nil {
		fmt.Printf("get file %s info failed, error: %v\n", common.StatusInfoFile, err)
		return fmt.Errorf("get file %s info failed", common.StatusInfoFile)
	}
	bytes, err := utils.LoadFile(common.StatusInfoFile)
	if err != nil {
		fmt.Printf("read file %s failed, error: %v\n", common.StatusInfoFile, err)
		return fmt.Errorf("read file %s failed", common.StatusInfoFile)
	}
	var contexts []common.CtrStatusInfo
	if err = json.Unmarshal(bytes, &contexts); err != nil {
		fmt.Printf("unmarshal status info failed, error: %v\n", err)
		return fmt.Errorf("unmarshal status info failed")
	}
	for idx, info := range contexts {
		info.Description = getDesc(info.Status, info.StatusStartTime)
		fmt.Println(border)
		printInfo("Container ID", info.CtrId)
		printInfo("Container Status", info.Status)
		printInfo("Container Status Start Time", time.Unix(info.StatusStartTime, 0).Format("2006-01-02 15:04:05"))
		printInfo("Container Description", info.Description)
		if idx == len(contexts)-1 {
			fmt.Println(border)
		}
	}
	return nil
}

func getDesc(status string, startTime int64) string {
	if status == common.StatusRunning {
		return common.DescNormal
	}
	infos, ok := ctrStatusInfos[status]
	if !ok {
		return common.DescUnknown
	}
	if time.Now().Unix()-startTime > infos.statusDuration {
		return infos.description
	}
	return common.DescNormal
}

func printInfo(label, text string) {
	if len(text) <= textWidth {
		fmt.Printf(format, label, text)
		return
	}

	words := strings.Fields(text)
	current, first := "", true
	for _, word := range words {
		if len(current)+len(word)+len(" ") > textWidth && current != "" {
			printLine(label, current, first)
			current, first = word, false
		} else {
			if current != "" {
				current += " "
			}
			current += word
		}
	}
	printLine(label, current, first)
}

func printLine(label, text string, first bool) {
	if !first {
		format = "| %27s  %-67s |\n"
		label = ""
	}
	fmt.Printf(format, label, text)
}
