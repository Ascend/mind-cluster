// Copyright (c) Huawei Technologies Co., Ltd. 2024-2025. All rights reserved.

// Package profile provides utils for profile
package profile

import (
	"encoding/json"
	"errors"
	"fmt"

	"k8s.io/api/core/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/interface/grpc/profiling"
	"clusterd/pkg/interface/kube"
)

// DataTraceController data trace controller is used to manage datatrace
type DataTraceController struct {
	JobNamespace string
	JobName      string // join jobNamespace and jobName with splash, such as default/test-pytorch
}

// NewDataTraceController creates a new DataTraceController
func NewDataTraceController(jobNs, jobName string) *DataTraceController {
	return &DataTraceController{
		JobNamespace: jobNs,
		JobName:      jobName,
	}
}

// IsDataTraceCmExist return whether the cm exist
func (dtc *DataTraceController) IsDataTraceCmExist() (*v1.ConfigMap, error) {
	if dtc == nil {
		return nil, errors.New("data trace controller is nil")
	}
	cm, err := kube.GetConfigMap(DataTraceCmPrefix+dtc.JobName, dtc.JobNamespace)
	if k8serr.IsNotFound(err) {
		hwlog.RunLog.Warnf("comfigmap [%s/%s] is not in this cluster", dtc.JobNamespace, dtc.JobName)
		return nil, err
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get configmap [%s/%s]", dtc.JobNamespace, dtc.JobName)
	}
	return cm, nil
}

// UpdateDataTraceCm to update the datatrace configmap with in parameters
func (dtc *DataTraceController) UpdateDataTraceCm(inParam *profiling.ProfilingSwitch) error {
	if inParam == nil {
		return errors.New("the incoming param is nil")
	}
	dataTraceCm, err := kube.GetConfigMap(DataTraceCmPrefix+dtc.JobName, dtc.JobNamespace)
	if err != nil {
		return fmt.Errorf(" failed to get cm,%v", err)
	}
	data := dataTraceCm.Data[DataTraceCmProfilingSwitchKey]
	var dataTraceParam SwitchStruct
	if err := json.Unmarshal([]byte(data), &dataTraceParam); err != nil {
		hwlog.RunLog.Error("the content of current data-trace cm is not serialize legally, will cover with new")
	}
	dtc.setDataTraceData(&dataTraceParam, inParam)
	newProParam, err := json.Marshal(dataTraceParam)
	if err != nil {
		return fmt.Errorf("failed to marshal content, err: %v", err)
	}
	if dataTraceCm.Data == nil {
		dataTraceCm.Data = make(map[string]string)
	}
	dataTraceCm.Data[DataTraceCmProfilingSwitchKey] = string(newProParam)
	if _, err := kube.UpdateConfigMap(dataTraceCm); err != nil {
		return fmt.Errorf("failed to update comfigmap [%s/%s], err: %v", dtc.JobNamespace, dtc.JobName, err)
	}
	return nil
}

// CreateDataTraceCm creates a new data trace for the given profile
func (dtc *DataTraceController) CreateDataTraceCm(inParam *profiling.ProfilingSwitch) error {
	if inParam == nil {
		return errors.New("the incoming param is nil")
	}
	cm := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      DataTraceCmPrefix + dtc.JobName,
			Namespace: dtc.JobNamespace,
			Labels:    map[string]string{"reset": "true"},
		},
	}
	var dataTraceParam SwitchStruct
	dtc.setDataTraceData(&dataTraceParam, inParam)
	newProParam, err := json.Marshal(dataTraceParam)
	if err != nil {
		return fmt.Errorf("failed to marshal content, err: %v", err)
	}
	if cm.Data == nil {
		cm.Data = make(map[string]string)
	}
	cm.Data[DataTraceCmProfilingSwitchKey] = string(newProParam)
	if _, err := kube.CreateConfigMap(cm); err != nil {
		return fmt.Errorf("failed to create comfigmap [%s/%s], err: %v", dtc.JobNamespace, dtc.JobName, err)
	}
	hwlog.RunLog.Infof("configmap [%s/%s] has been created", dtc.JobNamespace, dtc.JobName)
	return nil
}

func (dtc *DataTraceController) setDataTraceData(dataTraceParam *SwitchStruct,
	inParam *profiling.ProfilingSwitch) {
	dataTraceParam.CommunicationOperator = inParam.CommunicationOperator
	dataTraceParam.FP = inParam.FP
	dataTraceParam.Step = inParam.Step
	dataTraceParam.SaveCheckpoint = inParam.SaveCheckpoint
	dataTraceParam.DataLoader = inParam.DataLoader
}
