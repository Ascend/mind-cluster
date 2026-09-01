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
	"net"
	"os"
	"strconv"

	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
	"container-manager/pkg/common"
	app2 "container-manager/pkg/container/app"
	coordapp "container-manager/pkg/coordinator/app"
	"container-manager/pkg/devmgr"
	"container-manager/pkg/fault/app"
	app3 "container-manager/pkg/reset/app"
	"container-manager/pkg/workflow"
)

const (
	maxAge                         = 7
	maxLogLineLength               = 1024
	defaultSockPath                = "/run/docker.sock"
	faultCodeFileUmask os.FileMode = 0137 // file mode: 0640
)

type runCmd struct {
	logPath               string
	logLevel              int
	logMaxAge             int
	logMaxBackups         int
	ctrStrategy           string
	sockPath              string
	runtimeType           string
	faultCfgPath          string
	leaderIp              string
	leaderPort            int
	nodeID                string
	leaderAddrs           string
	eventSyncInterval     int
	scheduledSyncInterval int
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
	flag.IntVar(&cmd.logMaxBackups, "maxBackups", hwlog.DefaultBackups, "Maximum number of backup log files, range is (0, 180]")
	flag.StringVar(&cmd.runtimeType, "runtimeType", common.DockerType, "Container Runtime type")
	flag.StringVar(&cmd.sockPath, "sockPath", defaultSockPath, "Container Runtime sock file path")
	flag.StringVar(&cmd.ctrStrategy, "ctrStrategy", common.NeverStrategy, "Retracting strategy for faulty containers")
	flag.StringVar(&cmd.faultCfgPath, "faultConfigPath", "", "Custom fault config file path")
	flag.StringVar(&cmd.leaderIp, "leaderIp", "", "IP address this node's leader gRPC server binds to. Non-empty starts this node as a leader")
	flag.IntVar(&cmd.leaderPort, "leaderPort", common.DefaultPort, "Port this node's leader gRPC server listens on")
	flag.StringVar(&cmd.nodeID, "nodeID", "", "Unique node identifier. Defaults to hostname when empty")
	flag.StringVar(&cmd.leaderAddrs, "leaderAddrs", "", "Comma-separated list of leader ip:port, at most 2. Empty disables the coordinator")
	flag.IntVar(&cmd.eventSyncInterval, "eventSyncInterval", common.ChangedInterval, "Interval in seconds for event-driven data sync checks")
	flag.IntVar(&cmd.scheduledSyncInterval, "scheduledSyncInterval", common.SyncInterval, "Interval in seconds for scheduled full data sync")
	return true
}

// CheckParam check param
func (cmd *runCmd) CheckParam() error {
	checker := newRunCmdArgsChecker(*cmd)
	return checker.Check()
}

func newRunCmdArgsChecker(cmd runCmd) *runCmdArgsChecker {
	return &runCmdArgsChecker{
		runtimeType:           cmd.runtimeType,
		sockPath:              cmd.sockPath,
		ctrStrategy:           cmd.ctrStrategy,
		faultCfgPath:          cmd.faultCfgPath,
		leaderIp:              cmd.leaderIp,
		leaderPort:            cmd.leaderPort,
		nodeID:                cmd.nodeID,
		leaderAddrs:           cmd.leaderAddrs,
		eventSyncInterval:     cmd.eventSyncInterval,
		scheduledSyncInterval: cmd.scheduledSyncInterval,
	}
}

type runCmdArgsChecker struct {
	runtimeType           string
	sockPath              string
	ctrStrategy           string
	faultCfgPath          string
	leaderIp              string
	leaderPort            int
	nodeID                string
	leaderAddrs           string
	eventSyncInterval     int
	scheduledSyncInterval int
}

// Check param checker
func (c *runCmdArgsChecker) Check() error {
	var checkFuncs = []func() error{
		c.checkRuntimeType,
		c.checkSockPath,
		c.checkCtrStrategy,
		c.checkFaultConfigPath,
		c.checkLeaderIP,
		c.checkLeaderPort,
		c.checkLeaderAddrs,
		c.checkInterval,
		c.checkNodeID,
	}
	for _, checkFun := range checkFuncs {
		if err := checkFun(); err != nil {
			return err
		}
	}
	return nil
}

func (c *runCmdArgsChecker) checkLeaderIP() error {
	if c.leaderIp != "" && net.ParseIP(c.leaderIp) == nil {
		return fmt.Errorf("invalid leaderIp: %s", c.leaderIp)
	}
	return nil
}

func (c *runCmdArgsChecker) checkLeaderPort() error {
	if c.leaderPort < common.MinPort || c.leaderPort > common.MaxPort {
		return fmt.Errorf("leaderPort must be in [%d, %d], got %d", common.MinPort, common.MaxPort, c.leaderPort)
	}
	return nil
}

func (c *runCmdArgsChecker) checkNodeID() error {
	if _, err := common.GetNodeId(c.nodeID); err != nil {
		return fmt.Errorf("invalid nodeID, error %v", err)
	}
	return nil
}

func (c *runCmdArgsChecker) checkInterval() error {
	if c.eventSyncInterval < common.MinChangedInterval || c.eventSyncInterval > common.MaxChangedInterval {
		return fmt.Errorf("eventSyncInterval must be in [%d, %d], got %d",
			common.MinChangedInterval, common.MaxChangedInterval, c.eventSyncInterval)
	}
	if c.scheduledSyncInterval < common.MinSyncInterval || c.scheduledSyncInterval > common.MaxSyncInterval {
		return fmt.Errorf("scheduledSyncInterval must be in [%d, %d], got %d",
			common.MinSyncInterval, common.MaxSyncInterval, c.scheduledSyncInterval)
	}
	return nil
}

func (c *runCmdArgsChecker) checkLeaderAddrs() error {
	addrs, err := common.ParseAddrs(c.leaderAddrs)
	if err != nil {
		return fmt.Errorf("leaderAddrs error %v", err)
	}
	if len(addrs) > common.MaxLeaderNum {
		return fmt.Errorf("leaderAddrs must have at most %d addresses, got %d", common.MaxLeaderNum, len(addrs))
	}
	return nil
}

func (c *runCmdArgsChecker) checkRuntimeType() error {
	if !utils.Contains([]string{common.DockerType, common.ContainerDType}, c.runtimeType) {
		return fmt.Errorf("invalid runtimeType, should be in [%s, %s]", common.DockerType, common.ContainerDType)
	}
	return nil
}

func (c *runCmdArgsChecker) checkSockPath() error {
	_, err := utils.CheckPath(c.sockPath)
	if err != nil {
		return fmt.Errorf("invalid sockPath, %v", err)
	}
	return nil
}

func (c *runCmdArgsChecker) checkCtrStrategy() error {
	if !utils.Contains([]string{common.NeverStrategy, common.SingleStrategy, common.RingStrategy}, c.ctrStrategy) {
		return fmt.Errorf("invalid ctrStrategy, should be in [%s, %s, %s]",
			common.NeverStrategy, common.SingleStrategy, common.RingStrategy)
	}
	return nil
}

func (c *runCmdArgsChecker) checkFaultConfigPath() error {
	if c.faultCfgPath == "" {
		return nil
	}
	_, err := utils.CheckPath(c.faultCfgPath)
	if err != nil {
		return fmt.Errorf("invalid faultConfigPath, %v", err)
	}
	uid, err := utils.GetCurrentUid()
	if err != nil {
		return fmt.Errorf("get current uid failed, %v", err)
	}
	if err = utils.DoCheckOwnerAndPermission(c.faultCfgPath, faultCodeFileUmask, uid); err != nil {
		return fmt.Errorf("invalid faultConfigPath permission, %v", err)
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
	coordinator := coordapp.NewCoordinator(ctx, ctrCtl)
	ctrCtl.BindCoordinator(coordinator) // for CtrCtl to call RequestStop/Start

	moduleMgr := workflow.NewModuleMgr()
	moduleMgr.Register(devmgr.DevMgr)
	moduleMgr.Register(faultMgr)
	moduleMgr.Register(ctrCtl)
	moduleMgr.Register(resetMgr)
	moduleMgr.Register(coordinator)

	if err = moduleMgr.Init(); err != nil {
		return err
	}
	moduleMgr.Work(ctx)
	moduleMgr.ShutDown()
	return nil
}

func (cmd *runCmd) setParameters() {
	common.ParamOption = common.Option{
		RuntimeType:           cmd.runtimeType,
		SockPath:              cmd.sockPath,
		CtrStrategy:           cmd.ctrStrategy,
		FaultCfgPath:          cmd.faultCfgPath,
		ListenAddr:            common.GetListenAddr(cmd.leaderIp, strconv.Itoa(cmd.leaderPort)),
		EventSyncInterval:     cmd.eventSyncInterval,
		ScheduledSyncInterval: cmd.scheduledSyncInterval,
	}
	common.ParamOption.LocalNodeID, _ = common.GetNodeId(cmd.nodeID)
	common.ParamOption.LeaderAddrs, _ = common.ParseAddrs(cmd.leaderAddrs)
}
