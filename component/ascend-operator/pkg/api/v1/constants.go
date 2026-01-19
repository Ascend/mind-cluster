/*
Copyright(C) 2023. Huawei Technologies Co.,Ltd. All rights reserved.
*/

package v1

import (
	"github.com/kubeflow/common/pkg/apis/common/v1"
)

const (
	// FrameworkKey the key of the laebl
	FrameworkKey = "framework"

	// DefaultPort is default value of the port.
	DefaultPort = 2222

	// MindSporeFrameworkName is the name of ML Framework
	MindSporeFrameworkName = "mindspore"
	// MindSporeReplicaTypeScheduler is the type for Scheduler of distribute ML
	MindSporeReplicaTypeScheduler v1.ReplicaType = "Scheduler"

	// PytorchFrameworkName is the name of ML Framework
	PytorchFrameworkName = "pytorch"
	// PytorchReplicaTypeMaster is the type for Scheduler of distribute ML
	PytorchReplicaTypeMaster v1.ReplicaType = "Master"

	// TensorflowFrameworkName is the name of ML Framework
	TensorflowFrameworkName = "tensorflow"
	// TensorflowReplicaTypeChief is the type for Scheduler of distribute ML
	TensorflowReplicaTypeChief v1.ReplicaType = "Chief"

	// ReplicaTypeWorker this is also used for non-distributed Job
	ReplicaTypeWorker v1.ReplicaType = "Worker"

	// DefaultRestartPolicy is default RestartPolicy for MSReplicaSpec.
	DefaultRestartPolicy = v1.RestartPolicyNever

	// JodIdLabelKey is Job label key jobID
	JodIdLabelKey = "jobID"
	// AppLabelKey is Job label key app
	AppLabelKey = "app"
)
