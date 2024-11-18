// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package grpc a series of grpc function
package grpc

import (
	"net"
	"time"

	"google.golang.org/grpc"
	"huawei.com/npu-exporter/v6/common-utils/hwlog"
	"huawei.com/npu-exporter/v6/common-utils/limiter"

	"clusterd/pkg/common/constant"
	"clusterd/pkg/interface/grpc/pb"
	"clusterd/pkg/interface/grpc/service"
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

// Start the grpc server
func (server *ClusterInfoMgrServer) Start() error {
	listen, err := net.Listen("tcp", constant.GrpcPort)
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
	pb.RegisterRecoverServer(server.grpcServer, service.NewFaultRecoverService())

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
