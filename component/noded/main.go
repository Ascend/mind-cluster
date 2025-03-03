/* Copyright(C) 2023-2024. Huawei Technologies Co.,Ltd. All rights reserved.
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
	"syscall"

	"ascend-common/common-utils/hwlog"
	"nodeD/pkg/common"
	"nodeD/pkg/config"
	"nodeD/pkg/control"
	"nodeD/pkg/kubeclient"
	"nodeD/pkg/monitoring"
	"nodeD/pkg/pingmesh"
	"nodeD/pkg/reporter"
)

const (
	defaultLogFile = "/var/log/mindx-dl/noded/noded.log"
	// defaultHeatBeatInterval is the default report interval
	defaultReportInterval = 5
	// defaultMonitorPeriod is the default plugin monitor period
	defaultMonitorPeriod = 60
	// maxReportInterval is the max report interval
	maxReportInterval = 300
	// minReportInterval is the min report interval
	minReportInterval = 0
	// maxMonitorPeriod is the max plugin monitor period
	maxMonitorPeriod = 600
	// minMonitorPeriod is the min plugin monitor period
	minMonitorPeriod = 60
	// maxLineLength is max length of each log line
	maxLineLength = 512
)

var (
	hwLogConfig = &hwlog.LogConfig{
		LogFileName:   defaultLogFile,
		MaxLineLength: maxLineLength,
	}
	controller     = &control.NodeController{}
	configManager  = &config.FaultConfigurator{}
	monitorManager = &monitoring.MonitorManager{}
	reportManager  = &reporter.ReportManager{}
	version        bool
	// BuildVersion build version
	BuildVersion string
	// BuildName build name
	BuildName string
	// reportInterval report Interval
	reportInterval int
	// monitorPeriod monitoring period
	monitorPeriod int
	// resultMaxAge pingmesh result max age
	resultMaxAge int
)

func main() {
	flag.Parse()

	if version {
		fmt.Printf("%s version: %s \n", BuildName, BuildVersion)
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	// init hwlog
	if err := hwlog.InitRunLogger(hwLogConfig, ctx); err != nil {
		fmt.Printf("hwlog init failed, error is %v\n", err)
		return
	}
	if !checkParameters() {
		return
	}
	hwlog.RunLog.Infof("%s starting and the version is %s", BuildName, BuildVersion)
	setParameters()
	if err := createWorkers(); err != nil {
		hwlog.RunLog.Errorf("create workers failed, err is %v", err)
		return
	}
	if err := initFunction(); err != nil {
		hwlog.RunLog.Errorf("init function failed, err is %v", err)
		return
	}
	go configManager.Run(ctx)
	go monitorManager.Run(ctx)
	signalCatch(cancel)
}

func init() {
	flag.BoolVar(&version, "version", false, "the version of the program")
	flag.IntVar(&reportInterval, "reportInterval", defaultReportInterval,
		"Min interval of report node status")
	flag.IntVar(&monitorPeriod, "monitorPeriod", defaultMonitorPeriod, "Monitoring period of monitor ,"+
		"range [60,600] seconds")
	// hwlog configuration
	flag.IntVar(&hwLogConfig.LogLevel, "logLevel", 0,
		"Log level, -1-debug, 0-info, 1-warning, 2-error, 3-critical(default 0)")
	flag.IntVar(&hwLogConfig.MaxAge, "maxAge", hwlog.DefaultMinSaveAge,
		"Maximum number of days for backup run log files, range [7, 700] days")
	flag.StringVar(&hwLogConfig.LogFileName, "logFile", defaultLogFile,
		"Run log file path. if the file size exceeds 20MB, will be rotated")
	flag.IntVar(&hwLogConfig.MaxBackups, "maxBackups", hwlog.DefaultMaxBackups,
		"Maximum number of backup operation logs, range is (0, 30]")
	flag.IntVar(&resultMaxAge, "resultMaxAge", pingmesh.DefaultResultMaxAge,
		"Maximum number of days for backup run pingmesh result files, range [7, 700] days")
}

func checkParameters() bool {
	if reportInterval <= minReportInterval || reportInterval > maxReportInterval {
		hwlog.RunLog.Errorf("report interval %d out of range (0,300]", reportInterval)
		return false
	}
	if monitorPeriod < minMonitorPeriod || monitorPeriod > maxMonitorPeriod {
		hwlog.RunLog.Errorf("monitor period %d out of range [60,600]", monitorPeriod)
		return false
	}
	if resultMaxAge < pingmesh.MinResultMaxAge || resultMaxAge > pingmesh.MaxResultMaxAge {
		hwlog.RunLog.Errorf("resultMaxAge %d out of range [%d,%d]", resultMaxAge, pingmesh.MinResultMaxAge,
			pingmesh.MaxResultMaxAge)
		return false
	}
	return true
}

func setParameters() {
	common.ParamOption = common.Option{
		ReportInterval: reportInterval,
		MonitorPeriod:  monitorPeriod,
	}
}

func createWorkers() error {
	// init k8s client
	clientK8s, err := kubeclient.NewClientK8s()
	if err != nil {
		hwlog.RunLog.Infof("init k8s client failed when start, err is %v", err)
		return err
	}

	// init workers
	configManager = config.NewFaultConfigurator(clientK8s)
	controller = control.NewNodeController(clientK8s)
	monitorManager = monitoring.NewMonitorManager(clientK8s)
	reportManager = reporter.NewReporterManager(clientK8s)

	// build the connections between workers
	monitorManager.SetNextFaultProcessor(controller)
	controller.SetNextFaultProcessor(reportManager)
	configManager.SetNextConfigProcessor(controller)
	return nil
}

func initFunction() error {
	if err := configManager.Init(); err != nil {
		hwlog.RunLog.Errorf("init config manager failed when start, err is %v", err)
		return err
	}
	if err := controller.Init(); err != nil {
		hwlog.RunLog.Errorf("init controller failed when start, err is %v", err)
		return err
	}
	if err := monitorManager.Init(); err != nil {
		hwlog.RunLog.Errorf("init monitor manager failed when start, err is %v", err)
		return err
	}
	if err := reportManager.Init(); err != nil {
		hwlog.RunLog.Errorf("init reporter manager failed when start, err is %v", err)
		return err
	}
	return nil
}

func signalCatch(cancel context.CancelFunc) {
	osSignalChan := common.NewSignalWatcher(syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGKILL)
	if osSignalChan == nil {
		hwlog.RunLog.Error("create stop signal channel failed")
		return
	}
	select {
	case sig, sigEnd := <-osSignalChan:
		if !sigEnd {
			hwlog.RunLog.Info("catch system stop signal channel is closed")
			return
		}
		hwlog.RunLog.Infof("receive system signal: %s, NodeD shutting down", sig.String())
		cancel()
		configManager.Stop()
		monitorManager.Stop()
	}
}
