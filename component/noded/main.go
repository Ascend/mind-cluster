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

// Touch entrypoint so CI rebuilds this component after ascend-common boardId updates.

import (
	"context"
	"flag"
	"fmt"
	"syscall"

	"ascend-common/api"
	"ascend-common/common-utils/agreement"
	"ascend-common/common-utils/healthz"
	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
	ver "ascend-common/common-utils/version"
	fdol "ascend-faultdiag-online"
	"nodeD/pkg/common"
	snapshot "nodeD/pkg/containersnapshot"
	"nodeD/pkg/control"
	"nodeD/pkg/device"
	"nodeD/pkg/kubeclient"
	"nodeD/pkg/monitoring"
	"nodeD/pkg/monitoring/config"
	"nodeD/pkg/pingmesh"
	"nodeD/pkg/processmanager"
	"nodeD/pkg/releaser"
	"nodeD/pkg/reporter"
	"nodeD/pkg/watcher/configmap"
)

const (
	defaultLogFile = "/var/log/mindx-dl/noded/noded.log"
	// defaultHeatBeatInterval is the default report interval
	defaultReportInterval = 5
	// defaultIpmiMonitorPeriod is the default ipmi monitor period
	defaultIpmiMonitorPeriod = 60
	// defaultDpcMonitorPeriod is the default dpc monitor period
	defaultDpcMonitorPeriod = 5
	// defaultDtfsMonitorPeriod is the default dtfs monitor period
	defaultDtfsMonitorPeriod = 5
	// maxReportInterval is the max report interval
	maxReportInterval = 300
	// minReportInterval is the min report interval
	minReportInterval = 0
	// maxIpmiMonitorPeriod is the max ipmi monitor period
	maxIpmiMonitorPeriod = 600
	// minIpmiMonitorPeriod is the min ipmi monitor period
	minIpmiMonitorPeriod = 60
	// maxDpcMonitorPeriod is the max dpc monitor period
	maxDpcMonitorPeriod = 300
	// minDpcMonitorPeriod is the min dpc monitor period
	minDpcMonitorPeriod = 1
	// maxDtfsMonitorPeriod is the max dtfs monitor period
	maxDtfsMonitorPeriod = 300
	// minDtfsMonitorPeriod is the min dtfs monitor period
	minDtfsMonitorPeriod = 1
	// maxLineLength is max length of each log line
	maxLineLength          = 512
	fdConfigPath           = "/usr/local/fdConfig.yaml"
	containerSnapshotLabel = "/usr/local/Dockerfile-container-snapshot"
)

var (
	hwLogConfig = &hwlog.LogConfig{
		LogFileName:   defaultLogFile,
		MaxLineLength: maxLineLength,
	}
	controller      = &control.ControllerManager{}
	configManager   = &config.FaultConfigurator{}
	monitorManager  = &monitoring.MonitorManager{}
	reportManager   = &reporter.ReportManager{}
	pingmeshManager *pingmesh.Manager
	podMonitor      = &snapshot.PodMonitor{}
	version         bool
	// BuildVersion build version
	BuildVersion string
	// BuildName build name
	BuildName string
	// reportInterval report Interval
	reportInterval int
	// ipmiMonitorPeriod ipmi monitoring period
	ipmiMonitorPeriod int
	// monitorPeriod is the deprecated alias of ipmiMonitorPeriod, kept for compatibility
	monitorPeriod int
	// dpcMonitorPeriod dpc monitoring period
	dpcMonitorPeriod int
	// dtfsMonitorPeriod dtfs monitoring period
	dtfsMonitorPeriod int
	// resultMaxAge pingmesh result max age
	resultMaxAge int
	// deviceResetTimeout device reset timeout duration
	deviceResetTimeout int
	hzFlags            = healthz.RegisterFlags()
)

func main() {
	flag.Parse()
	info := ver.Get()
	if version {
		fmt.Printf("version=%s commit=%s branch=%s os=%s arch=%s goVersion=%s\n",
			info.Version, info.GitCommit, info.GitBranch, info.BuildOS, info.BuildArch, info.GoVersion)
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	// init hwlog
	if err := hwlog.InitRunLogger(hwLogConfig, ctx); err != nil {
		fmt.Printf("hwlog init failed, error is %v\n", err)
		return
	}
	if err := hzFlags.Serve(ctx); err != nil {
		hwlog.RunLog.Errorf("failed to start healthz server: %v", err)
		return
	}
	if !checkParameters() {
		return
	}
	hwlog.RunLog.Infof("%s starting and the version is %s", BuildName, BuildVersion)
	if utils.IsExist(containerSnapshotLabel) {
		hwlog.RunLog.Info("dependent image build for container snapshot")
	}
	setParameters()
	if err := createWorkers(); err != nil {
		hwlog.RunLog.Errorf("create workers failed, err is %v", err)
		return
	}
	if err := initFunction(ctx); err != nil {
		hwlog.RunLog.Errorf("init function failed, err is %v", err)
		return
	}
	go monitorManager.Run(ctx)
	go configmap.GetCmWatcher().Watch(ctx.Done())
	initPodInformer()
	if err := ver.ReportVersionToNodeAnnotation(kubeclient.GetK8sClient(), info, "noded"); err != nil {
		hwlog.RunLog.Error(err)
	}
	if pingmeshManager != nil {
		go pingmeshManager.Run(ctx)
	}
	fdol.StartFDOnline(fdConfigPath, []string{"slowNode"}, "node")
	signalCatch(cancel)
}

func init() {
	agreement.PrintAgreement()
	flag.BoolVar(&version, "version", false, "the version of the program")
	flag.IntVar(&reportInterval, "reportInterval", defaultReportInterval,
		"Min interval of report node status")
	flag.IntVar(&ipmiMonitorPeriod, "ipmiMonitorPeriod", defaultIpmiMonitorPeriod, fmt.Sprintf(
		"Monitoring period of the ipmi monitor, range [%d,%d] seconds",
		minIpmiMonitorPeriod, maxIpmiMonitorPeriod))
	flag.IntVar(&monitorPeriod, "monitorPeriod", defaultIpmiMonitorPeriod, fmt.Sprintf(
		"(Deprecated) Monitoring period of the ipmi monitor, range [%d,%d] seconds, "+
			"replaced by -ipmiMonitorPeriod and takes precedence when both are set",
		minIpmiMonitorPeriod, maxIpmiMonitorPeriod))
	flag.IntVar(&dpcMonitorPeriod, "dpcMonitorPeriod", defaultDpcMonitorPeriod, fmt.Sprintf(
		"Monitoring period of the dpc monitor, range [%d,%d] seconds",
		minDpcMonitorPeriod, maxDpcMonitorPeriod))
	flag.IntVar(&dtfsMonitorPeriod, "dtfsMonitorPeriod", defaultDtfsMonitorPeriod, fmt.Sprintf(
		"Monitoring period of the dtfs monitor, range [%d,%d] seconds",
		minDtfsMonitorPeriod, maxDtfsMonitorPeriod))
	// hwlog configuration
	flag.IntVar(&hwLogConfig.LogLevel, "logLevel", 0,
		"Log level, -1-debug, 0-info, 1-warning, 2-error, 3-critical(default 0)")
	flag.IntVar(&hwLogConfig.MaxAge, "maxAge", hwlog.DefaultMinSaveAge,
		"Maximum number of days for backup run log files, range [7, 700] days")
	flag.StringVar(&hwLogConfig.LogFileName, "logFile", defaultLogFile,
		"Run log file path. if the file size exceeds 20MB, will be rotated")
	flag.IntVar(&hwLogConfig.MaxBackups, "maxBackups", hwlog.DefaultBackups,
		"Maximum number of backup operation logs, range is (0, 180]")
	flag.IntVar(&resultMaxAge, "resultMaxAge", pingmesh.DefaultResultMaxAge,
		"Maximum number of days for backup run pingmesh result files, range [7, 700] days")
	flag.IntVar(&deviceResetTimeout, api.DeviceResetTimeout, api.DefaultDeviceResetTimeout,
		"when noded starts, if the number of chips is insufficient, the maximum duration to wait for "+
			"the driver to report all chips, unit second, range [10, 600]")
}

func checkParameters() bool {
	// reportInterval is a half-open range (0,300], so it is checked explicitly
	if reportInterval <= minReportInterval || reportInterval > maxReportInterval {
		hwlog.RunLog.Errorf("report interval %d out of range (0,300]", reportInterval)
		return false
	}
	// monitorPeriod is deprecated but takes precedence over ipmiMonitorPeriod when both are set
	if isFlagPassed("monitorPeriod") {
		if !checkRange("monitor period", monitorPeriod, minIpmiMonitorPeriod, maxIpmiMonitorPeriod) {
			return false
		}
	} else if !checkRange("ipmi monitor period", ipmiMonitorPeriod, minIpmiMonitorPeriod, maxIpmiMonitorPeriod) {
		return false
	}
	if !checkRange("dpc monitor period", dpcMonitorPeriod, minDpcMonitorPeriod, maxDpcMonitorPeriod) {
		return false
	}
	if !checkRange("dtfs monitor period", dtfsMonitorPeriod, minDtfsMonitorPeriod, maxDtfsMonitorPeriod) {
		return false
	}
	if !checkRange("resultMaxAge", resultMaxAge, pingmesh.MinResultMaxAge, pingmesh.MaxResultMaxAge) {
		return false
	}
	return checkRange("deviceResetTimeout", deviceResetTimeout, api.MinDeviceResetTimeout, api.MaxDeviceResetTimeout)
}

// checkRange logs an error and returns false if value is outside [min, max].
func checkRange(name string, value, min, max int) bool {
	if value < min || value > max {
		hwlog.RunLog.Errorf("%s %d out of range [%d,%d]", name, value, min, max)
		return false
	}
	return true
}

// isFlagPassed reports whether the named command-line flag was explicitly set.
func isFlagPassed(name string) bool {
	passed := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			passed = true
		}
	})
	return passed
}

func setParameters() {
	ipmiPeriod := ipmiMonitorPeriod
	if isFlagPassed("monitorPeriod") {
		hwlog.RunLog.Warn("monitorPeriod is deprecated, please use ipmiMonitorPeriod instead")
		ipmiPeriod = monitorPeriod
	}
	common.ParamOption = common.Option{
		ReportInterval:     reportInterval,
		IpmiMonitorPeriod:  ipmiPeriod,
		DpcMonitorPeriod:   dpcMonitorPeriod,
		DtfsMonitorPeriod:  dtfsMonitorPeriod,
		DeviceResetTimeout: deviceResetTimeout,
	}
}

func createWorkers() error {
	// init k8s client
	clientK8s, err := kubeclient.NewClientK8s()
	if err != nil {
		hwlog.RunLog.Errorf("init k8s client failed when start, error: %v", err)
		return err
	}

	configmap.InitCmWatcher(clientK8s)
	if err = device.InitDeviceManager(); err != nil {
		hwlog.RunLog.Errorf("init device manager failed when start, error: %v", err)
		return err
	}
	// init workers
	controller = control.NewControlManager(clientK8s)
	monitorManager = monitoring.NewMonitorManager(clientK8s)
	reportManager = reporter.NewReporterManager(clientK8s)
	pingmeshManager = pingmesh.NewManager(&pingmesh.Config{
		ResultMaxAge: resultMaxAge,
		KubeClient:   clientK8s,
	})
	releaser.InitReleaser()
	configmap.GetCmWatcher().Init()
	// build the connections between workers
	monitorManager.SetNextFaultProcessor(controller)
	controller.SetNextFaultProcessor(reportManager)
	return nil
}

func initFunction(ctx context.Context) error {
	if err := processmanager.InitPlugin(ctx); err != nil {
		hwlog.RunLog.Errorf("init controller failed when start, err is %v", err)
		return err
	}
	go func() {
		if err := controller.InitNodeSNInfo(); err != nil {
			hwlog.RunLog.Warnf("init node annotation failed when start, err is %v", err)
		}
	}()
	return nil
}

func initPodInformer() {
	if kubeclient.GetK8sClient() == nil {
		hwlog.RunLog.Errorf("k8s client is nil")
		return
	}
	podMonitor = snapshot.NewPodMonitor(kubeclient.GetK8sClient())
	go podMonitor.Monitoring()
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
		podMonitor.Stop()
	}
}
