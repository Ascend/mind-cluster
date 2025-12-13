// Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.

// Package profile provides utils for profile
package profile

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/interface/grpc/profiling"
	"clusterd/pkg/interface/kube"
)

func init() {
	hwlog.InitRunLogger(&hwlog.LogConfig{OnlyToStdout: true}, context.Background())
}

func TestNewDataTraceController(t *testing.T) {
	tests := []struct {
		name     string
		jobNs    string
		jobName  string
		wantNs   string
		wantName string
	}{
		{
			name:     "normal case",
			jobNs:    "test-ns",
			jobName:  "test-job",
			wantNs:   "test-ns",
			wantName: "test-job",
		},
		{
			name:     "empty ns and name",
			jobNs:    "",
			jobName:  "",
			wantNs:   "",
			wantName: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dtc := NewDataTraceController(tt.jobNs, tt.jobName)
			assert.Equal(t, tt.wantNs, dtc.JobNamespace)
			assert.Equal(t, tt.wantName, dtc.JobName)
		})
	}
}

type TestCaseForIsDataTraceCmExist struct {
	name      string
	dtc       *DataTraceController
	mockExist bool
	mockErr   error
	wantErr   bool
	wantCm    *v1.ConfigMap
}

func createTestCasesForIsDataTraceCmExist() []TestCaseForIsDataTraceCmExist {
	return []TestCaseForIsDataTraceCmExist{
		{
			name:      "controller is nil",
			dtc:       nil,
			mockExist: false,
			mockErr:   nil,
			wantErr:   true,
			wantCm:    nil,
		},
		{
			name:      "cm not found",
			dtc:       NewDataTraceController("ns", "job"),
			mockExist: false,
			mockErr:   errors.NewNotFound(schema.GroupResource{Resource: "configmaps"}, DataTraceCmPrefix+"job"),
			wantErr:   true,
			wantCm:    nil,
		},
		{
			name:      "cm exists",
			dtc:       NewDataTraceController("ns", "job"),
			mockExist: true,
			mockErr:   nil,
			wantErr:   false,
			wantCm: &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      DataTraceCmPrefix + "job",
					Namespace: "ns",
				},
			},
		},
	}
}

func TestIsDataTraceCmExist(t *testing.T) {
	tests := createTestCasesForIsDataTraceCmExist()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var patches *gomonkey.Patches
			if tt.dtc != nil {
				patches = gomonkey.ApplyFunc(kube.GetConfigMap, func(cmName, ns string) (*v1.ConfigMap, error) {
					if tt.mockExist {
						return tt.wantCm, tt.mockErr
					}
					return nil, tt.mockErr
				})
				defer patches.Reset()
			}
			cm, err := tt.dtc.IsDataTraceCmExist()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantCm.Name, cm.Name)
				assert.Equal(t, tt.wantCm.Namespace, cm.Namespace)
			}
		})
	}
}

type TestCaseForCreateDataTraceCm struct {
	name          string
	dtc           *DataTraceController
	switchParam   *profiling.ProfilingSwitch
	owner         metav1.OwnerReference
	mockCreateErr error
	wantErr       bool
}

func createTestCasesForCreateDataTraceCm() []TestCaseForCreateDataTraceCm {
	owner := metav1.OwnerReference{
		APIVersion: "v1",
		Kind:       "Job",
		Name:       "test-job",
		UID:        "123456",
	}

	switchParam := &profiling.ProfilingSwitch{
		CommunicationOperator: "on",
		FP:                    "on",
		Step:                  "on",
		SaveCheckpoint:        "on",
		DataLoader:            "on",
	}
	return []TestCaseForCreateDataTraceCm{
		{
			name:          "switch param is nil",
			dtc:           NewDataTraceController("ns", "job"),
			switchParam:   nil,
			owner:         owner,
			mockCreateErr: nil,
			wantErr:       true,
		},
		{
			name:          "create cm success",
			dtc:           NewDataTraceController("ns", "job"),
			switchParam:   switchParam,
			owner:         owner,
			mockCreateErr: nil,
			wantErr:       false,
		},
		{
			name:          "create cm failed",
			dtc:           NewDataTraceController("ns", "job"),
			switchParam:   switchParam,
			owner:         owner,
			mockCreateErr: fmt.Errorf("create failed"),
			wantErr:       true,
		},
	}
}

func TestCreateDataTraceCm(t *testing.T) {
	tests := createTestCasesForCreateDataTraceCm()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var patches *gomonkey.Patches
			if tt.switchParam != nil {
				cmName := DataTraceCmPrefix + tt.dtc.JobName
				expectedCm := &v1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:            cmName,
						Namespace:       tt.dtc.JobNamespace,
						Labels:          map[string]string{"reset": "true"},
						OwnerReferences: []metav1.OwnerReference{tt.owner},
					},
				}
				var dataTraceParam SwitchStruct
				tt.dtc.setDataTraceData(&dataTraceParam, tt.switchParam)
				newProParam, err := json.Marshal(dataTraceParam)
				assert.NoError(t, err)
				expectedCm.Data = map[string]string{
					DataTraceCmProfilingSwitchKey: string(newProParam),
				}
				patches = gomonkey.ApplyFunc(kube.CreateConfigMap, func(cm *v1.ConfigMap) (*v1.ConfigMap, error) {
					assert.Equal(t, expectedCm.Name, cm.Name)
					assert.Equal(t, expectedCm.Namespace, cm.Namespace)
					assert.Equal(t, expectedCm.Labels, cm.Labels)
					assert.Equal(t, expectedCm.OwnerReferences, cm.OwnerReferences)
					assert.Equal(t, expectedCm.Data, cm.Data)
					return cm, tt.mockCreateErr
				})
				defer patches.Reset()
			}
			err := tt.dtc.CreateDataTraceCm(tt.switchParam, tt.owner)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDataTraceController_setDataTraceData(t *testing.T) {
	dtc := NewDataTraceController("ns", "job")
	inParam := &profiling.ProfilingSwitch{
		CommunicationOperator: "on",
		FP:                    "on",
		Step:                  "on",
		SaveCheckpoint:        "on",
		DataLoader:            "on",
	}
	var dataTraceParam SwitchStruct
	dtc.setDataTraceData(&dataTraceParam, inParam)
	assert.Equal(t, inParam.CommunicationOperator, dataTraceParam.CommunicationOperator)
	assert.Equal(t, inParam.FP, dataTraceParam.FP)
	assert.Equal(t, inParam.Step, dataTraceParam.Step)
	assert.Equal(t, inParam.SaveCheckpoint, dataTraceParam.SaveCheckpoint)
	assert.Equal(t, inParam.DataLoader, dataTraceParam.DataLoader)
}

type TestCaseForUpdateDataTraceCm struct {
	name    string
	dtc     *DataTraceController
	param   *profiling.ProfilingSwitch
	owner   metav1.OwnerReference
	mock    func() *gomonkey.Patches
	wantErr bool
}

func createFirstTestCasesForUpdateDataTraceCm(
	owner metav1.OwnerReference, switchParam *profiling.ProfilingSwitch) []TestCaseForUpdateDataTraceCm {
	return []TestCaseForUpdateDataTraceCm{
		{
			name:    "param is nil",
			dtc:     NewDataTraceController("ns", "name"),
			param:   nil,
			owner:   owner,
			mock:    func() *gomonkey.Patches { return gomonkey.NewPatches() },
			wantErr: true,
		},
		{
			name:  "get cm error",
			dtc:   NewDataTraceController("ns", "name"),
			param: switchParam,
			owner: owner,
			mock: func() *gomonkey.Patches {
				patches := gomonkey.NewPatches()
				patches.ApplyFunc(kube.GetConfigMap, func(name, namespace string) (*v1.ConfigMap, error) {
					return nil, fmt.Errorf("get cm error")
				})
				return patches
			},
			wantErr: true,
		},
	}
}

func createSecondTestCasesForUpdateDataTraceCm(
	owner metav1.OwnerReference, switchParam *profiling.ProfilingSwitch) []TestCaseForUpdateDataTraceCm {
	return []TestCaseForUpdateDataTraceCm{
		{
			name:  "unmarshal data error",
			dtc:   NewDataTraceController("ns", "name"),
			param: switchParam,
			owner: owner,
			mock: func() *gomonkey.Patches {
				patches := gomonkey.NewPatches()
				cm := &v1.ConfigMap{Data: map[string]string{DataTraceCmProfilingSwitchKey: "invalid json"}}
				patches.ApplyFunc(kube.GetConfigMap, func(name, namespace string) (*v1.ConfigMap, error) {
					return cm, nil
				})
				patches.ApplyFunc(kube.UpdateConfigMap, func(cm *v1.ConfigMap) (*v1.ConfigMap, error) {
					return cm, nil
				})
				return patches
			},
			wantErr: false,
		},
		{
			name:  "marshal error",
			dtc:   NewDataTraceController("ns", "name"),
			param: switchParam,
			owner: owner,
			mock: func() *gomonkey.Patches {
				patches := gomonkey.NewPatches()
				cm := &v1.ConfigMap{Data: map[string]string{DataTraceCmProfilingSwitchKey: "{}"}}
				patches.ApplyFunc(kube.GetConfigMap, func(name, namespace string) (*v1.ConfigMap, error) {
					return cm, nil
				})
				patches.ApplyFunc(json.Marshal, func(v interface{}) ([]byte, error) {
					return nil, fmt.Errorf("marshal error")
				})
				return patches
			},
			wantErr: true,
		},
	}
}

func createThirdTestCasesForUpdateDataTraceCm(
	owner metav1.OwnerReference, switchParam *profiling.ProfilingSwitch) []TestCaseForUpdateDataTraceCm {
	return []TestCaseForUpdateDataTraceCm{
		{
			name:  "update cm error",
			dtc:   NewDataTraceController("ns", "name"),
			param: switchParam,
			owner: owner,
			mock: func() *gomonkey.Patches {
				patches := gomonkey.NewPatches()
				cm := &v1.ConfigMap{Data: map[string]string{DataTraceCmProfilingSwitchKey: "{}"}}
				patches.ApplyFunc(kube.GetConfigMap, func(name, namespace string) (*v1.ConfigMap, error) {
					return cm, nil
				})
				patches.ApplyFunc(kube.UpdateConfigMap, func(cm *v1.ConfigMap) (*v1.ConfigMap, error) {
					return nil, fmt.Errorf("update error")
				})
				return patches
			},
			wantErr: true,
		},
		{
			name:  "success",
			dtc:   NewDataTraceController("ns", "name"),
			param: switchParam,
			owner: owner,
			mock: func() *gomonkey.Patches {
				patches := gomonkey.NewPatches()
				cm := &v1.ConfigMap{Data: map[string]string{DataTraceCmProfilingSwitchKey: "{}"}}
				patches.ApplyFunc(kube.GetConfigMap, func(name, namespace string) (*v1.ConfigMap, error) {
					return cm, nil
				})
				patches.ApplyFunc(kube.UpdateConfigMap, func(cm *v1.ConfigMap) (*v1.ConfigMap, error) {
					return cm, nil
				})
				return patches
			},
			wantErr: false,
		},
	}
}

func TestUpdateDataTraceCm(t *testing.T) {
	owner := metav1.OwnerReference{}
	switchParam := &profiling.ProfilingSwitch{
		CommunicationOperator: "on", FP: "on", Step: "on", SaveCheckpoint: "on", DataLoader: "on"}
	tests := createFirstTestCasesForUpdateDataTraceCm(owner, switchParam)
	tests = append(tests, createSecondTestCasesForUpdateDataTraceCm(owner, switchParam)...)
	tests = append(tests, createThirdTestCasesForUpdateDataTraceCm(owner, switchParam)...)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patches := tt.mock()
			defer patches.Reset()

			err := tt.dtc.UpdateDataTraceCm(tt.param, tt.owner)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
