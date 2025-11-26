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
	"errors"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"

	"ascend-common/common-utils/utils"
	"container-manager/pkg/common"
)

const (
	totalWidth = 96
	textWidth  = 65
	sideWidth  = 2
)

var (
	border     = "+" + strings.Repeat("=", totalWidth-sideWidth) + "+"
	ctrIDRegex = regexp.MustCompile(`^[a-f0-9]{64}$`)
	format     = "| %-25s: %-65s |\n"
)

type statusCmd struct {
	containerID string
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
	flag.StringVar(&cmd.containerID, "containerID", "", "Container ID for displaying information")
	return true
}

// CheckParam check param
func (cmd *statusCmd) CheckParam() error {
	if match := ctrIDRegex.MatchString(cmd.containerID); !match {
		return errors.New("invalid containerID")
	}
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
	for _, info := range contexts {
		if info.CtrId == cmd.containerID {
			fmt.Println(border)
			printInfo("Container ID", info.CtrId)
			printInfo("Container Status", info.Status)
			printInfo("Container Description", info.Description)
			fmt.Println(border)
			return nil
		}
	}
	fmt.Printf("container id <%s> is not existed\n", cmd.containerID)
	return nil
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
		format = "| %25s  %-65s |\n"
		label = ""
	}
	fmt.Printf(format, label, text)
}
