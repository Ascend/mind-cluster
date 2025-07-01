// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package publicfault for public fault service
package publicfault

import (
	"context"

	"ascend-common/api"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/logs"
	"clusterd/pkg/domain/common"
	"clusterd/pkg/interface/grpc/pubfault"
)

// PubFaultService a service for public fault
type PubFaultService struct {
	serviceCtx context.Context
	pubfault.UnimplementedPubFaultServer
}

// NewPubFaultService new PubFaultService
func NewPubFaultService(ctx context.Context) *PubFaultService {
	return &PubFaultService{serviceCtx: ctx}
}

// SendPublicFault send public fault to clusterd
func (s *PubFaultService) SendPublicFault(ctx context.Context,
	req *pubfault.PublicFaultRequest) (*pubfault.RespStatus, error) {
	event := "send public fault"
	logs.RecordLog(req.Resource, event, constant.Start)
	res := constant.Failed
	defer func() {
		logs.RecordLog(req.Resource, event, res)
	}()

	pubFaultInfo := constructPubFaultInfo(req)
	if err := PubFaultCollector(pubFaultInfo); err != nil {
		if err.Error() == "limiter work by resource failed" {
			return &pubfault.RespStatus{
				Code: int32(common.InvalidReqRate),
				Info: err.Error(),
			}, nil
		}
		return &pubfault.RespStatus{
			Code: int32(common.InvalidReqParam),
			Info: err.Error(),
		}, nil
	}
	res = constant.Success
	return &pubfault.RespStatus{
		Code: int32(common.OK),
		Info: "public fault send successfully",
	}, nil
}

func constructPubFaultInfo(req *pubfault.PublicFaultRequest) *api.PubFaultInfo {
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
