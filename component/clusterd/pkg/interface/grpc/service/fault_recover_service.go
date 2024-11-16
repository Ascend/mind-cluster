// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package service a series of service function
package service

import (
	"context"

	"clusterd/pkg/interface/grpc/pb"
)

// FaultRecoverService is a service for fault recover
type FaultRecoverService struct {
	keepAliveInterval int
	serviceCtx        context.Context
	pb.UnimplementedRecoverServer
}

// NewFaultRecoverService return a new instance of FaultRecoverService
func NewFaultRecoverService(keepAlive int, ctx context.Context) *FaultRecoverService {
	s := &FaultRecoverService{}
	s.keepAliveInterval = keepAlive
	s.serviceCtx = ctx
	return s
}

// DeleteJob clear registered resources
func (s *FaultRecoverService) DeleteJob(jobId string) {

}
