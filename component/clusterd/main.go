// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package main a series of main function
package main

import (
	"context"
	"flag"
	"fmt"
	"syscall"

	"huawei.com/npu-exporter/v6/common-utils/hwlog"

	"clusterd/pkg/application/faultmanager"
	"clusterd/pkg/application/resource"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
	"clusterd/pkg/interface/kube"
)

const (
	defaultLogFile   = "/var/log/mindx-dl/clusterd/clusterd.log"
	maxLogLineLength = 1023
)

var (
	hwLogConfig = &hwlog.LogConfig{LogFileName: defaultLogFile, MaxLineLength: maxLogLineLength}
	// BuildVersion build version
	BuildVersion string
	// BuildName build name
	BuildName string
	version   bool
)

func startInformer(ctx context.Context) {
	kube.InitCMInformer()
	kube.InitPodInformer()
	kube.InitPGInformer(ctx)
	kube.AddCmNodeFunc(constant.Resource, resource.NodeCollector)
	kube.AddCmDeviceFunc(constant.Resource, resource.DeviceInfoCollector)
	kube.AddCmSwitchFunc(constant.Resource, resource.SwitchInfoCollector)
	go resource.Report()
}

func startFaultManager(ctx context.Context) {
	go faultmanager.Process(ctx)
}

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
	err := kube.InitClientK8s()
	if err != nil {
		hwlog.RunLog.Errorf("new client config err: %v", err)
		return
	}
	err = kube.InitClientVolcano()
	if err != nil {
		hwlog.RunLog.Errorf("new volcano client config err: %v", err)
	}
	// election and running process
	startInformer(ctx)
	startFaultManager(ctx)
	signalCatch(cancel)
}

func init() {
	flag.BoolVar(&version, "version", false, "the version of the program")

	// hwlog configuration
	flag.IntVar(&hwLogConfig.LogLevel, "logLevel", 0,
		"Log level, -1-debug, 0-info, 1-warning, 2-error, 3-critical(default 0)")
	flag.IntVar(&hwLogConfig.MaxAge, "maxAge", hwlog.DefaultMinSaveAge,
		"Maximum number of days for backup run log files, range [7, 700] days")
	flag.StringVar(&hwLogConfig.LogFileName, "logFile", defaultLogFile,
		"Run log file path. if the file size exceeds 20MB, will be rotated")
	flag.IntVar(&hwLogConfig.MaxBackups, "maxBackups", hwlog.DefaultMaxBackups,
		"Maximum number of backup operator logs, range is (0, 30]")
}

func checkParameters() bool {
	return true
}

func signalCatch(cancel context.CancelFunc) {
	osSignalChan := util.NewSignalWatcher(syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGKILL)
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
		hwlog.RunLog.Infof("receive system signal: %s, ClusterD shutting down", sig.String())
		cancel()
	}
}
