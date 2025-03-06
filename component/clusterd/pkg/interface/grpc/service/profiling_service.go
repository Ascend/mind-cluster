// Copyright (c) Huawei Technologies Co., Ltd. 2024-2025. All rights reserved.

package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/api/errors"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/domain/profile"
	"clusterd/pkg/interface/grpc/pb-profiling"
	"clusterd/pkg/interface/kube"
)

const (
	// PartsOfJobNs parts num of in param
	PartsOfJobNs = 2
)

// ProfilingSwitchManager represents profiling switch manager
type ProfilingSwitchManager struct {
	pb_profiling.UnimplementedTrainingDataTraceServer
}

const (
	// ErrInvalidParam is returned when a parameter is invalid
	ErrInvalidParam = 300
	// ErrNotFound is returned when a parameter is not found
	ErrNotFound = 404
	// OK all good
	OK = 200
	// ErrServerFault is returned when server error occurs
	ErrServerFault = 500
)

// ModifyTrainingDataTraceSwitch to modify the profiling marker status by updating the cm
func (ps *ProfilingSwitchManager) ModifyTrainingDataTraceSwitch(ctx context.Context,
	in *pb_profiling.DataTypeReq) (*pb_profiling.DataTypeRes, error) {
	jobNsName := in.GetJobNsName()
	jobNameInfo := strings.Split(jobNsName, "/")
	if len(jobNameInfo) != PartsOfJobNs {
		return &pb_profiling.DataTypeRes{Message: "the format of jobNsName is not namespace/jobName",
			Code: ErrInvalidParam}, fmt.Errorf("the format of jobNsName is not namespace/jobName")
	}
	jobNs, jobName := jobNameInfo[0], jobNameInfo[1]
	dtc := profiling.NewDataTraceController(jobNs, jobName)
	if cm, err := kube.GetConfigMap(profiling.DataTraceCmPrefix+dtc.JobName,
		dtc.JobNamespace); cm == nil || err != nil {
		if !errors.IsNotFound(err) {
			return &pb_profiling.DataTypeRes{Message: fmt.Sprintf("failed to found comfigmap:[%s/%s]",
				dtc.JobNamespace, dtc.JobName), Code: ErrNotFound}, err
		}
		if err := dtc.CreateDataTraceCm(in.ProfilingSwitch); err != nil {
			return &pb_profiling.DataTypeRes{Message: fmt.Sprintf("failed to create comfigmap:[%s/%s] "+
				"while cm is not exist", dtc.JobNamespace, dtc.JobName), Code: ErrServerFault}, err
		}
		return &pb_profiling.DataTypeRes{Message: fmt.Sprintf("comfigmap:[%s/%s] has been created"+
			" and param is updated to change profiling marker status", dtc.JobNamespace, dtc.JobName), Code: OK}, nil
	}
	if err := dtc.UpdateDataTraceCm(in.ProfilingSwitch); err != nil {
		return &pb_profiling.DataTypeRes{Message: fmt.Sprintf("failed to update comfigmap:[%s/%s]",
			dtc.JobNamespace, dtc.JobName), Code: ErrServerFault}, err
	}
	response := &pb_profiling.DataTypeRes{Message: "successfully changed profiling marker enable status", Code: OK}
	return response, nil
}

// GetTrainingDataTraceSwitch get  current profiling marker status
func (ps *ProfilingSwitchManager) GetTrainingDataTraceSwitch(ctx context.Context,
	in *pb_profiling.DataStatusReq) (*pb_profiling.DataStatusRes, error) {
	jobNsName := in.GetJobNsName()
	jobNameInfo := strings.Split(jobNsName, "/")
	if len(jobNameInfo) != PartsOfJobNs {
		return &pb_profiling.DataStatusRes{Message: "the format of jobNsName is not namespace/jobName",
			Code: ErrInvalidParam}, nil
	}
	jobNs, jobName := jobNameInfo[0], jobNameInfo[1]
	dtc := profiling.NewDataTraceController(jobNs, jobName)
	cm, err := dtc.IsDataTraceCmExist()
	if cm == nil || err != nil {
		hwlog.RunLog.Errorf("can not find data trace configmap[%s/%s]", dtc.JobNamespace, dtc.JobName)
		return &pb_profiling.DataStatusRes{
			Message: fmt.Sprintf("failed to found comfigmap:[%s/%s]", dtc.JobNamespace, dtc.JobName),
			Code:    ErrNotFound}, err
	}
	data, ok := cm.Data[profiling.DataTraceCmProfilingSwitchKey]
	if !ok {
		hwlog.RunLog.Infof("data trace configmap[%s/%s] has no %s field",
			dtc.JobNamespace, dtc.JobName, profiling.DataTraceCmProfilingSwitchKey)
		return &pb_profiling.DataStatusRes{Message: "data trace configmap does not contain the 'profilingSwitch' field",
			Code: ErrNotFound}, nil
	}
	var resProfile pb_profiling.ProfilingSwitch
	if err := json.Unmarshal([]byte(data), &resProfile); err != nil {
		hwlog.RunLog.Errorf("failed to unmarshal configmap[%s/%s], err:%v", dtc.JobNamespace, dtc.JobName, err)
		return &pb_profiling.DataStatusRes{
			Message: fmt.Sprintf("failed to convert comfigmap:[%s/%s]", dtc.JobNamespace, dtc.JobName),
			Code:    ErrServerFault}, err
	}
	return &pb_profiling.DataStatusRes{
		Message:         fmt.Sprintf("successfully get the status of job[%s/%s]", dtc.JobNamespace, dtc.JobName),
		ProfilingSwitch: &resProfile,
		Code:            OK}, nil
}
