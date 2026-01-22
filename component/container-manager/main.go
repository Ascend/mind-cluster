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
	"os"
	"sort"
	"strings"

	"container-manager/pkg/command"
)

const (
	cmdIndex    = 1
	cmdArgIndex = 2
)

var (
	h       bool
	help    bool
	v       bool
	version bool

	// BuildName show component name
	BuildName string
	// BuildVersion show component version
	BuildVersion string

	curCmd command.Command
	cmdMap = make(map[string]command.Command)
)

func setCurCmd(cmd command.Command) {
	curCmd = cmd
}

func main() {
	initCmd()
	if !dealArgs() {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := curCmd.CheckParam(); err != nil {
		fmt.Printf("cmd '%s' check param failed: %v\n", curCmd.Name(), err)
		return
	}
	if err := curCmd.InitLog(ctx); err != nil {
		fmt.Printf("cmd '%s' init log failed, error: %v\n", curCmd.Name(), err)
		return
	}
	if err := curCmd.Execute(ctx); err != nil {
		fmt.Printf("cmd '%s' execute failed, error: %v\n", curCmd.Name(), err)
		return
	}
}

func initCmd() {
	registerCmd(command.StatusCmd())
}

func registerCmd(cmd command.Command) {
	if _, ok := cmdMap[cmd.Name()]; !ok {
		cmdMap[cmd.Name()] = cmd
	}
}

func dealArgs() bool {
	flag.Usage = printHelp
	if len(os.Args) <= cmdIndex {
		printHelp()
		return false
	}
	if len(os.Args[cmdIndex]) == 0 {
		fmt.Println("the required parameter is missing")
		return false
	}
	if os.Args[cmdIndex][0] == '-' {
		dealOptionFlag()
		return false
	}
	return dealCmdFlag()
}

func dealCmdFlag() bool {
	cmd, ok := cmdMap[os.Args[cmdIndex]]
	if !ok {
		fmt.Printf("unknown command: %s\n", os.Args[cmdIndex])
		printHelp()
		return false
	}
	setCurCmd(cmd)
	if !curCmd.BindFlag() {
		return true
	}
	flag.Usage = flag.PrintDefaults
	if err := flag.CommandLine.Parse(os.Args[cmdArgIndex:]); err != nil {
		fmt.Printf("parse cmd args failed, error: %v\n", err)
		return false
	}
	return true
}

func dealOptionFlag() {
	flag.BoolVar(&h, "h", false, "Print help information")
	flag.BoolVar(&help, "help", false, "Print help information")
	flag.BoolVar(&v, "v", false, "Print version information")
	flag.BoolVar(&version, "version", false, "Print version information")
	flag.Parse()
	if h || help {
		printCmdUsage()
		return
	}
	if v || version {
		fmt.Printf("%s version: %s\n", BuildName, BuildVersion)
		return
	}
}

func printHelp() {
	fmt.Println("use '-help' for help information")
}

func printCmdUsage() {
	descriptions := make([]string, 0, len(cmdMap))
	for _, cmd := range cmdMap {
		descriptions = append(descriptions, fmt.Sprintf("\t%-10s\t%s", cmd.Name(), cmd.Description()))
	}
	sort.Strings(descriptions)
	fmt.Printf(`Container Manager, supports fault management and automatic recovery.

Usage: [OPTIONS...] COMMAND

Options:
	-h,-help	Print help information
	-v,-version	Print version information

Commands:
%s
`, strings.Join(descriptions, "\n"))
}
