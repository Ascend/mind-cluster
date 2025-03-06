// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package main a series of main function
package main

import (
	"context"
	"flag"
	"fmt"
	"syscall"
	"time"

	"golang.org/x/time/rate"
	"google.golang.org/grpc"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/application/faultmanager"
	"clusterd/pkg/application/jobv2"
	"clusterd/pkg/application/pingmesh"
	"clusterd/pkg/application/publicfault"
	"clusterd/pkg/application/resource"
	"clusterd/pkg/application/statistics"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/logs"
	"clusterd/pkg/common/util"
	sv "clusterd/pkg/interface/grpc"
	"clusterd/pkg/interface/grpc/service"
	pubfaultsvc "clusterd/pkg/interface/grpc/service-pubfault"
	"clusterd/pkg/interface/kube"
)

const (
	defaultLogFile = "/var/log/mindx-dl/clusterd/clusterd.log"
)

var (
	hwLogConfig = &hwlog.LogConfig{LogFileName: defaultLogFile, MaxLineLength: constant.MaxLogLineLength}
	// BuildVersion build version
	BuildVersion string
	// BuildName build name
	BuildName         string
	version           bool
	server            *sv.ClusterInfoMgrServer
	limiter           = rate.NewLimiter(rate.Every(time.Second), constant.QpsLimit)
	keepAliveInterval = 5
)

func limitQPS(ctx context.Context, req interface{},
	info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	if !limiter.Allow() {
		hwlog.RunLog.Errorf("qps exceeded, method=%s", info.FullMethod)
		return nil, fmt.Errorf("qps exceeded, method=%s", info.FullMethod)
	}
	return handler(ctx, req)
}

func startInformer(ctx context.Context) {
	// starting informer requires after adding processing functions
	addResourceFunc()
	addJobFunc()
	kube.AddNodeFunc(constant.PingMesh, pingmesh.NodeCollector)
	kube.InitCMInformer()
	kube.InitPubFaultCMInformer()
	kube.InitPodAndNodeInformer()
	kube.InitPodGroupInformer()
	go pingmesh.TickerCheckSuperPodDevice(ctx)
	// specific functions requires after informer
	addFuncAfterInformer()

	go jobv2.Handler(ctx)
	go jobv2.Checker(ctx)
	go resource.Report(ctx)
	dealPubFault(ctx)
}

func dealPubFault(ctx context.Context) {
	go publicfault.WatchPubFaultCustomFile(ctx)
	go publicfault.PubFaultNeedDelete.DealDelete(ctx)
}

func addJobFunc() {
	kube.AddPodGroupFunc(constant.Job, jobv2.PodGroupCollector)
	kube.AddPodFunc(constant.Job, jobv2.PodCollector)
}

func addResourceFunc() {
	kube.AddCmSwitchFunc(constant.Resource, faultmanager.SwitchInfoCollector)
	kube.AddCmNodeFunc(constant.Resource, faultmanager.NodeCollector)
	kube.AddCmDeviceFunc(constant.Resource, faultmanager.DeviceInfoCollector)
	kube.AddNodeFunc(constant.Resource, statistics.UpdateNodeSNAndNameCache)
}

func addFuncAfterInformer() {
	kube.AddCmPubFaultFunc(constant.Resource, faultmanager.PubFaultCollector)
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
	if err := logs.InitJobEventLogger(ctx); err != nil {
		hwlog.RunLog.Errorf("JobEventLog init failed, error is %v", err)
		return
	}
	if !checkParameters() {
		return
	}
	if err := initK8sServer(); err != nil {
		hwlog.RunLog.Errorf("init k8s servers failed, error: %v", err)
		return
	}
	initGrpcServer(ctx)
	faultmanager.GlobalFaultProcessCenter.Work(ctx)
	startInformer(ctx)
	initStatisticModule(ctx)
	signalCatch(cancel)
}

func initStatisticModule(ctx context.Context) {
	go statistics.GlobalJobCollectMgr.JobCollector(ctx)
	go statistics.GlobalJobOutputMgr.JobOutput(ctx)

	// fault relation
	go statistics.StatisticFault.UpdateFault(ctx)
	statistics.StatisticFault.LoadFaultData()
}

func initGrpcServer(ctx context.Context) {
	server = sv.NewClusterInfoMgrServer([]grpc.ServerOption{grpc.MaxRecvMsgSize(constant.MaxGRPCRecvMsgSize),
		grpc.MaxConcurrentStreams(constant.MaxGRPCConcurrentStreams),
		grpc.UnaryInterceptor(limitQPS)})
	recoverService := service.NewFaultRecoverService(keepAliveInterval, ctx)
	pubFaultSvc := pubfaultsvc.NewPubFaultService(ctx)
	dataTraceSvc := &service.ProfilingSwitchManager{}
	if err := server.Start(recoverService, pubFaultSvc, dataTraceSvc); err != nil {
		hwlog.RunLog.Errorf("clusterd grpc server start failed, error: %v", err)
	}
}

func initK8sServer() error {
	if err := kube.InitClientK8s(); err != nil {
		return fmt.Errorf("new client config err: %v", err)
	}
	vcK8sClient, err := kube.InitClientVolcano()
	if err != nil {
		return fmt.Errorf("new volcano client config err: %v", err)
	}
	if !kube.CheckVolcanoExist(vcK8sClient.ClientSet) {
		return fmt.Errorf("volcano not exist, please deploy volcano")
	}
	return nil
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
		if server != nil {
			server.Stop(false)
		}
		cancel()
	}
}
