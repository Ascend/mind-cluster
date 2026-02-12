// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package utils is common utils
package utils

import (
	"fmt"
	"strings"

	commonv1 "github.com/kubeflow/common/pkg/apis/common/v1"
	"github.com/kubeflow/common/pkg/util"

	"ascend-common/common-utils/hwlog"
	v1 "ascend-operator/pkg/api/v1"
)

// GetScaleOutTypeFromJob get the scaleout-type label in lowercase format of ascend job
func GetScaleOutTypeFromJob(job *v1.AscendJob) (string, bool) {
	if job == nil || job.Labels == nil {
		return "", false
	}
	value, exist := job.Labels[v1.ScaleOutTypeLabel]
	return strings.ToUpper(value), exist
}

// IsScaleOutTypeValid get the scaleout-type label in lowercase format of ascend job
func IsScaleOutTypeValid(scaleOutType string) bool {
	return scaleOutType == v1.ScaleOutTypeRoCE || scaleOutType == v1.ScaleOutTypeUBoE
}

// UpdateAcJobFailedWhenInvalidScaleOutType set acjob state to failed when invalid scaleout-type label value
func UpdateAcJobFailedWhenInvalidScaleOutType(job *v1.AscendJob, r commonv1.ControllerInterface) error {
	if job == nil {
		return fmt.Errorf("illegal input: job is nil")
	}
	errMsg := fmt.Sprintf("the value of label %s is invalid, which should be %s or %s",
		v1.ScaleOutTypeLabel, v1.ScaleOutTypeRoCE, v1.ScaleOutTypeUBoE)
	hwlog.RunLog.Error(errMsg)
	err := util.UpdateJobConditions(&job.Status, commonv1.JobFailed, "invalid label config", errMsg)
	if err != nil {
		hwlog.RunLog.Errorf("update job condition error: %v", err)
		return err
	}
	err = r.UpdateJobStatusInApiServer(job, &job.Status)
	if err != nil {
		hwlog.RunLog.Errorf("update job status in api server error: %v", err)
		return err
	}
	return nil
}

// CheckAcJobScaleOutTypeLabel check the scaleout-type label of ascend job
func CheckAcJobScaleOutTypeLabel(job *v1.AscendJob) error {
	scaleOutType, exist := GetScaleOutTypeFromJob(job)
	if !exist || strings.TrimSpace(scaleOutType) == "" {
		return nil
	}
	if IsScaleOutTypeValid(scaleOutType) {
		return nil
	}
	return fmt.Errorf("the value of label %s is invalid, which should be %s or %s", v1.ScaleOutTypeLabel,
		v1.ScaleOutTypeRoCE, v1.ScaleOutTypeUBoE)
}

// CheckAndUpdateAcJobScaleOutTypeLabel check the scaleout-type label of ascend job,
// update the ascend job task state when the value is invalid
func CheckAndUpdateAcJobScaleOutTypeLabel(job *v1.AscendJob, r commonv1.ControllerInterface) error {
	checkErr := CheckAcJobScaleOutTypeLabel(job)
	if checkErr == nil {
		return nil
	}
	if updateErr := UpdateAcJobFailedWhenInvalidScaleOutType(job, r); updateErr != nil {
		hwlog.RunLog.Errorf("failed to update acjob state to failed when %s, err: %s", checkErr.Error(), updateErr)
		return updateErr
	}
	hwlog.RunLog.Error(checkErr.Error())
	return checkErr
}
