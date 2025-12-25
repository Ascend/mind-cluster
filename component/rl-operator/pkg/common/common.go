/*
Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package common

import (
	commonv1 "github.com/kubeflow/common/pkg/apis/common/v1"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"ascend-common/api"
)

type ValidateError struct {
	Reason  string
	Message string
}

func (ve *ValidateError) Error() string {
	return ve.Message
}

type ReQueueError struct {
	Message string
}

func (req *ReQueueError) Error() string {
	return req.Message
}

type ResourceOption struct {
	Kind          *source.Kind
	PredicateFunc predicate.Funcs
}

type ConditionInfo struct {
	CondType        commonv1.JobConditionType
	Reason, Message string
}

// Instance is for annotation
type Instance struct { // Instance
	PodName    string `json:"pod_name"`  // pod Name
	ServerID   string `json:"server_id"` // serverdId
	ServerIP   string `json:"server_ip"` // server ip for A5
	SuperPodId *int32 `json:"super_pod_id"`
	Devices    []Dev  `json:"devices"`      // dev
	RackID     *int32 `json:"rack_id"`      // Rack id for A5
	SeverIndex string `json:"server_index"` // sever index for A5
}

// Dev to hccl
type Dev struct {
	DeviceID      string `json:"device_id"` // hccl deviceId
	DeviceIP      string `json:"device_ip"` // hccl deviceIp
	SuperDeviceID string `json:"super_device_id,omitempty"`
	// rank level info in rank table for A5
	LevelList []api.RankLevel `json:"levelList,omitempty"`
}

type RayLabel struct {
	L0  string `json:"L0"`
	L1  string `json:"L1"`
	L2  string `json:"L2"`
	L3  string `json:"L3,omitempty"`
	NPU int    `json:"-"`
}

type AscendSuperpodBlock struct {
	SpBlock string `json:"SpBlock,omitempty"`
	TpBlock string `json:"TpBlock,omitempty"`
}

type ManagerSetUp interface {
	SetupWithManager(mgr manager.Manager) error
}

var ServicePortKeys = []string{
	RayGcsPortLabelKey,
	RayClientPortLabelKey,
	RayDashboardPortLabelKey,
}
