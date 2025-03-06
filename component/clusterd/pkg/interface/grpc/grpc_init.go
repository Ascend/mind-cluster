// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package grpc a series of grpc function
package grpc

import (
	pb_profiling "clusterd/pkg/interface/grpc/pb-profiling"
	"errors"
	"net"
	"os"
	"time"

	"google.golang.org/grpc"

	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/limiter"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/interface/grpc/pb"
	pb2 "clusterd/pkg/interface/grpc/pb-publicfault"
	"clusterd/pkg/interface/grpc/service"
	pubfaultsvc "clusterd/pkg/interface/grpc/service-pubfault"
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
func (server *ClusterInfoMgrServer) Start(recoverSvc *service.FaultRecoverService,
	pubFaultSvc *pubfaultsvc.PubFaultService, dataTraceSvc *service.ProfilingSwitchManager) error {
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
	pb2.RegisterPubFaultServer(server.grpcServer, pubFaultSvc)
	pb_profiling.RegisterTrainingDataTraceServer(server.grpcServer, dataTraceSvc)

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
