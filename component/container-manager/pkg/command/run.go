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
	"errors"
	"flag"
	"fmt"

	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
	"container-manager/pkg/common"
	app2 "container-manager/pkg/container/app"
	"container-manager/pkg/devmgr"
	"container-manager/pkg/fault/app"
	app3 "container-manager/pkg/reset/app"
	"container-manager/pkg/workflow"
)

const (
	maxAge           = 7
	maxBackups       = 30
	maxLogLineLength = 1024
	defaultSockPath  = "/run/containerd/containerd.sock"
)

type runCmd struct {
	logPath       string
	logLevel      int
	logMaxAge     int
	logMaxBackups int
	ctrStrategy   string
	sockPath      string
	runtimeType   string
}

// RunCmd cmd 'run'
func RunCmd() Command {
	return &runCmd{}
}

// Name cmd name
func (cmd *runCmd) Name() string {
	return "run"
}

// Description cmd description
func (cmd *runCmd) Description() string {
	return "Run container-manager"
}

// BindFlag bind flag. If not needed, return false directly
func (cmd *runCmd) BindFlag() bool {
	flag.StringVar(&cmd.logPath, "logPath", "/var/log/mindx-dl/container-manager/container-manager.log",
		"The log file path. If the file size exceeds 20MB, will be dumped")
	flag.IntVar(&cmd.logLevel, "logLevel", 0, "Log level, -1-debug, 0-info, 1-warning, 2-error, 3-critical")
	flag.IntVar(&cmd.logMaxAge, "maxAge", maxAge, "Maximum number of days for backup log files, range is [7, 700]")
	flag.IntVar(&cmd.logMaxBackups, "maxBackups", maxBackups, "Maximum number of backup log files, range is (0, 30]")
	flag.StringVar(&cmd.runtimeType, "runtimeMode", common.DockerType, "Container Runtime type")
	flag.StringVar(&cmd.sockPath, "sockPath", defaultSockPath, "Container Runtime sock file path")
	flag.StringVar(&cmd.ctrStrategy, "ctrStrategy", common.NeverStrategy, "Retracting strategy for faulty containers")
	return true
}

// CheckParam check param
func (cmd *runCmd) CheckParam() error {
	checker := newRunCmdArgsChecker(*cmd)
	return checker.Check()
}

func newRunCmdArgsChecker(cmd runCmd) *runCmdArgsChecker {
	return &runCmdArgsChecker{
		runtimeType: cmd.runtimeType,
		sockPath:    cmd.sockPath,
		ctrStrategy: cmd.ctrStrategy,
	}
}

type runCmdArgsChecker struct {
	runtimeType string
	sockPath    string
	ctrStrategy string
}

// Check param checker
func (c *runCmdArgsChecker) Check() error {
	var checkFuncs = []func() error{
		c.checkRuntimeType,
		c.checkSockPath,
		c.checkCtrStrategy,
	}
	for _, checkFun := range checkFuncs {
		if err := checkFun(); err != nil {
			return err
		}
	}
	return nil
}

func (c *runCmdArgsChecker) checkRuntimeType() error {
	if !utils.Contains([]string{common.DockerType, common.ContainerDType}, c.runtimeType) {
		return fmt.Errorf("invalid runtimeType, should be between %s and %s", common.DockerType, common.ContainerDType)
	}
	return nil
}

func (c *runCmdArgsChecker) checkSockPath() error {
	if !utils.IsExist(c.sockPath) {
		return errors.New("socket file not exist")
	}
	_, err := utils.CheckPath(c.sockPath)
	if err != nil {
		return fmt.Errorf("invalid sockPath, %v", err)
	}
	return nil
}

func (c *runCmdArgsChecker) checkCtrStrategy() error {
	if !utils.Contains([]string{common.NeverStrategy, common.SingleStrategy, common.RingStrategy}, c.ctrStrategy) {
		return fmt.Errorf("invalid ctrStrategy, should be between %s, %s and %s",
			common.NeverStrategy, common.SingleStrategy, common.RingStrategy)
	}
	return nil
}

// InitLog init log
func (cmd *runCmd) InitLog(ctx context.Context) error {
	hwLogConfig := hwlog.LogConfig{
		LogFileName:   cmd.logPath,
		LogLevel:      cmd.logLevel,
		MaxAge:        cmd.logMaxAge,
		MaxBackups:    cmd.logMaxBackups,
		MaxLineLength: maxLogLineLength,
	}
	if err := hwlog.InitRunLogger(&hwLogConfig, ctx); err != nil {
		return err
	}
	hwlog.RunLog.Info("init log success")
	return nil
}

// Execute execute cmd
func (cmd *runCmd) Execute(ctx context.Context) error {
	cmd.setParameters()
	if err := devmgr.NewHwDevMgr(); err != nil {
		hwlog.RunLog.Errorf("new dev manager failed, error: %v", err)
		return errors.New("new dev manager failed")
	}
	faultMgr := app.NewFaultMgr()
	ctrCtl, err := app2.NewCtrCtl()
	if err != nil {
		hwlog.RunLog.Errorf("new container controller failed, error: %v", err)
		return errors.New("new container controller failed")
	}
	resetMgr := app3.NewResetMgr()

	moduleMgr := workflow.NewModuleMgr()
	moduleMgr.Register(devmgr.DevMgr)
	moduleMgr.Register(faultMgr)
	moduleMgr.Register(ctrCtl)
	moduleMgr.Register(resetMgr)
	if err = moduleMgr.Init(); err != nil {
		return err
	}
	moduleMgr.Work(ctx)
	moduleMgr.ShutDown()
	return nil
}

func (cmd *runCmd) setParameters() {
	common.ParamOption = common.Option{
		RuntimeMode: cmd.runtimeType,
		SockPath:    cmd.sockPath,
		CtrStrategy: cmd.ctrStrategy,
	}
}
