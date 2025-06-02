// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package grpc a series of grpc function
package grpc

import (
	"context"
	"errors"
	"net"
	"os"
	"time"

	"google.golang.org/grpc"

	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/limiter"
	"clusterd/pkg/application/config"
	"clusterd/pkg/application/fault"
	"clusterd/pkg/application/jobinfo"
	"clusterd/pkg/application/profiling"
	"clusterd/pkg/application/publicfault"
	"clusterd/pkg/application/recover"
	"clusterd/pkg/common/constant"
	grpcconfig "clusterd/pkg/interface/grpc/config"
	grpcfault "clusterd/pkg/interface/grpc/fault"
	"clusterd/pkg/interface/grpc/job"
	pbprofiling "clusterd/pkg/interface/grpc/profiling"
	"clusterd/pkg/interface/grpc/pubfault"
	"clusterd/pkg/interface/grpc/recover"
)

var (
	keepAliveInterval = 5
)

// ClusterInfoMgrServer is a server of clusterd
type ClusterInfoMgrServer struct {
	grpcServer *grpc.Server
	opts       []grpc.ServerOption
}

// NewClusterInfoMgrServer get a pointer of ClusterInfoMgrServer Object
func NewClusterInfoMgrServer(opts []grpc.ServerOption) *ClusterInfoMgrServer {
	server := &ClusterInfoMgrServer{}
	server.opts = append([]grpc.ServerOption(nil), opts...)
	return server
}

func isIPValid(ipStr string) error {
	parsedIp := net.ParseIP(ipStr)
	if parsedIp == nil {
		return errors.New("parse to ip failed")
	}
	if parsedIp.To4() == nil {
		return errors.New("not a valid ipv4 ip")
	}
	if parsedIp.Equal(net.IPv4bcast) {
		return errors.New("cannot be broadcast ip")
	}
	if parsedIp.IsUnspecified() {
		return errors.New("is all zeros ip")
	}
	return nil
}

// Start the grpc server
func (server *ClusterInfoMgrServer) Start(ctx context.Context) error {
	recoverSvc := recover.NewFaultRecoverService(keepAliveInterval, ctx)
	pubFaultSvc := publicfault.NewPubFaultService(ctx)
	dataTraceSvc := profiling.NewSwitchManager(ctx)
	configSvc := config.NewBusinessConfigServer(ctx)
	faultSvc := fault.NewFaultServer(ctx)
	jobSvc := jobinfo.NewJobServer(ctx)

	ipStr := os.Getenv("POD_IP")
	if err := isIPValid(ipStr); err != nil {
		return err
	}
	listenAddress := ipStr + constant.GrpcPort
	listen, err := net.Listen("tcp", listenAddress)
	if err != nil {
		hwlog.RunLog.Errorf("cluster info server listen failed, err: %#v", err)
		return err
	}
	limitedListener, err := limiter.LimitListener(listen, constant.MaxConcurrentLimit,
		constant.MaxIPConnectionLimit, constant.CacheSize)
	if err != nil {
		hwlog.RunLog.Errorf("create limit listener failed, err: %#v", err)
		return err
	}
	server.grpcServer = grpc.NewServer(server.opts...)
	pb.RegisterRecoverServer(server.grpcServer, recoverSvc)
	pubfault.RegisterPubFaultServer(server.grpcServer, pubFaultSvc)
	pbprofiling.RegisterTrainingDataTraceServer(server.grpcServer, dataTraceSvc)
	grpcconfig.RegisterConfigServer(server.grpcServer, configSvc)
	grpcfault.RegisterFaultServer(server.grpcServer, faultSvc)
	job.RegisterJobServer(server.grpcServer, jobSvc)

	go func() {
		if err := server.grpcServer.Serve(limitedListener); err != nil {
			hwlog.RunLog.Errorf("cluster info server crashed, err: %#v", err)
		}
	}()

	// Wait for grpc server ready
	for len(server.grpcServer.GetServiceInfo()) <= 0 {
		time.Sleep(time.Second)
	}
	hwlog.RunLog.Infof("cluster info server start listen...")
	return nil
}

// Stop grpc server
func (server *ClusterInfoMgrServer) Stop(grace bool) {
	if server.grpcServer == nil {
		return
	}
	if grace {
		server.grpcServer.GracefulStop()
	} else {
		server.grpcServer.Stop()
	}
}
