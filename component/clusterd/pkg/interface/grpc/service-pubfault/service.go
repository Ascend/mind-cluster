// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package pubfaultsvc for public fault service
package pubfaultsvc

import (
	"context"

	"ascend-common/api"
	"clusterd/pkg/application/publicfault"
	"clusterd/pkg/interface/grpc/common"
	pb2 "clusterd/pkg/interface/grpc/pb-publicfault"
)

// PubFaultService a service for public fault
type PubFaultService struct {
	serviceCtx context.Context
	pb2.UnimplementedPubFaultServer
}

// NewPubFaultService new PubFaultService
func NewPubFaultService(ctx context.Context) *PubFaultService {
	return &PubFaultService{serviceCtx: ctx}
}

// SendPublicFault send public fault to clusterd
func (s *PubFaultService) SendPublicFault(ctx context.Context, req *pb2.PublicFaultRequest) (*pb2.RespStatus, error) {
	pubFaultInfo := constructPubFaultInfo(req)
	if err := publicfault.PubFaultCollector(pubFaultInfo); err != nil {
		if err.Error() == "limiter work by resource failed" {
			return &pb2.RespStatus{
				Code: int32(common.InvalidReqRate),
				Info: err.Error(),
			}, nil
		}
		return &pb2.RespStatus{
			Code: int32(common.InvalidReqParam),
			Info: err.Error(),
		}, nil
	}
	return &pb2.RespStatus{
		Code: int32(common.OK),
		Info: "public fault send successfully",
	}, nil
}

func constructPubFaultInfo(req *pb2.PublicFaultRequest) *api.PubFaultInfo {
	var faults = []api.Fault{}
	for _, reqFault := range req.Faults {
		var influence []api.Influence
		for _, reqInfluence := range reqFault.Influence {
			influence = append(influence, api.Influence{
				NodeName:  reqInfluence.NodeName,
				NodeSN:    reqInfluence.NodeSN,
				DeviceIds: reqInfluence.DeviceIds,
			})
		}
		faults = append(faults, api.Fault{
			FaultId:       reqFault.FaultId,
			FaultType:     reqFault.FaultType,
			FaultCode:     reqFault.FaultCode,
			FaultTime:     reqFault.FaultTime,
			Assertion:     reqFault.Assertion,
			FaultLocation: reqFault.FaultLocation,
			Influence:     influence,
			Description:   reqFault.Description,
		})
	}
	return &api.PubFaultInfo{
		Id:        req.Id,
		TimeStamp: req.Timestamp,
		Version:   req.Version,
		Resource:  req.Resource,
		Faults:    faults,
	}
}
