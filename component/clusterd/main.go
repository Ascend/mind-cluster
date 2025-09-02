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
	"google.golang.org/grpc/keepalive"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/application/faultmanager"
	"clusterd/pkg/application/fdapi"
	"clusterd/pkg/application/jobv2"
	"clusterd/pkg/application/node"
	"clusterd/pkg/application/pingmesh"
	"clusterd/pkg/application/publicfault"
	"clusterd/pkg/application/resource"
	"clusterd/pkg/application/statistics"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/logs"
	"clusterd/pkg/common/util"
	"clusterd/pkg/domain/epranktable"
	"clusterd/pkg/domain/job"
	sv "clusterd/pkg/interface/grpc"
	"clusterd/pkg/interface/kube"
)

const (
	defaultLogFile       = "/var/log/mindx-dl/clusterd/clusterd.log"
	grpcKeepAliveTimeOut = 5
)

var (
	hwLogConfig = &hwlog.LogConfig{LogFileName: defaultLogFile, MaxLineLength: constant.MaxLogLineLength}
	// BuildVersion build version
	BuildVersion string
	// BuildName build name
	BuildName string
	version   bool
	server    *sv.ClusterInfoMgrServer
	limiter   = rate.NewLimiter(rate.Every(time.Second), constant.QpsLimit)
	useProxy  bool
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
	addEpRankTableFunc()
	kube.AddNodeFunc(constant.PingMesh, pingmesh.NodeCollector)
	kube.InitCMInformer()
	kube.InitPubFaultCMInformer()
	kube.InitPodAndNodeInformer()
	kube.InitACJobInformer()
	kube.InitVCJobInformer()
	kube.InitPodGroupInformer()
	go pingmesh.TickerCheckSuperPodDevice(ctx)
	// specific functions requires after informer
	addFuncAfterInformer()

	// generate global ranktable message handler
	go epranktable.GetEpGlobalRankTableManager().ConsumerForQueue()

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
	kube.AddACJobFunc(constant.Statistics, statistics.ACJobInfoCollector)
	kube.AddVCJobFunc(constant.Statistics, statistics.VCJobInfoCollector)
}

func addEpRankTableFunc() {
	kube.AddPodFunc(constant.EpRankTable, jobv2.EpGlobalRankTableMassageCollector)
	kube.AddCmRankTableFunc(constant.EpRankTable, epranktable.InformerHandler)
}

func addResourceFunc() {
	kube.AddCmSwitchFunc(constant.Resource, faultmanager.SwitchInfoCollector)
	kube.AddCmNodeFunc(constant.Resource, faultmanager.NodeCollector)
	kube.AddCmDeviceFunc(constant.Resource, faultmanager.DeviceInfoCollector)
	// UpdateNodeInfoCache must be before pingmesh
	kube.AddNodeFunc(constant.Resource, node.UpdateNodeInfoCache)
	kube.AddCmConfigPingMeshFunc(constant.Resource, pingmesh.ConfigCollector)
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
	if err := initLogger(ctx); err != nil {
		fmt.Printf("logger init failed: %v\n", err)
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
	fdapi.StartFdOL()
	faultmanager.GlobalFaultProcessCenter.Work(ctx)
	startInformer(ctx)
	initStatisticModule(ctx)
	go job.RefreshFaultJobInfo(ctx)
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
	keepAlive := keepalive.ServerParameters{
		Time:    time.Minute,
		Timeout: grpcKeepAliveTimeOut * time.Second,
	}
	server = sv.NewClusterInfoMgrServer([]grpc.ServerOption{
		grpc.MaxRecvMsgSize(constant.MaxGRPCRecvMsgSize),
		grpc.MaxSendMsgSize(constant.MaxGRPCSendMsgSize),
		grpc.MaxConcurrentStreams(constant.MaxGRPCConcurrentStreams),
		grpc.UnaryInterceptor(limitQPS),
		grpc.KeepaliveParams(keepAlive)})
	if err := server.Start(ctx, useProxy); err != nil {
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
	if _, err := kube.InitOperatorClient(); err != nil {
		return fmt.Errorf("new operator client config err: %v", err)
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
	flag.BoolVar(&useProxy, "useProxy", false, "use local grpc proxy")
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

func initLogger(ctx context.Context) error {
	// init hwlog
	if err := hwlog.InitRunLogger(hwLogConfig, ctx); err != nil {
		return fmt.Errorf("log init failed, error is %v", err)
	}
	if err := logs.InitJobEventLogger(ctx); err != nil {
		hwlog.RunLog.Errorf("JobEventLog init failed, error is %v", err)
		return fmt.Errorf("job event log init failed, error is %v", err)
	}
	if err := logs.InitGrpcEventLogger(ctx); err != nil {
		hwlog.RunLog.Errorf("GrpcEventLog init failed, error is %v", err)
		return fmt.Errorf("grpc event log init failed, error is %v", err)
	}
	return nil
}
