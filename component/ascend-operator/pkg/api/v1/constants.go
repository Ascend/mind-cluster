/*
Copyright(C) 2023. Huawei Technologies Co.,Ltd. All rights reserved.
*/

package v1

import (
	commonv1 "github.com/kubeflow/common/pkg/apis/common/v1"
)

const (
	// FrameworkKey the key of the laebl
	FrameworkKey = "framework"

	// Kind is the kind name
	Kind = "AscendJob"
	// Plural is the Plural for AscendJob.
	Plural = "ascendjobs"
	// Singular is the singular for AscendJob.
	Singular = "ascendjob"

	// DefaultContainerName the default container name for AscendJob.
	DefaultContainerName = "ascend"
	// DefaultPortName is name of the port used to communicate between other process.
	DefaultPortName = "ascendjob-port"
	// DefaultPort is default value of the port.
	DefaultPort = 2222

	// MindSporeFrameworkName is the name of ML Framework
	MindSporeFrameworkName = "mindspore"
	// MindSporeReplicaTypeScheduler is the type for Scheduler of distribute ML
	MindSporeReplicaTypeScheduler commonv1.ReplicaType = "Scheduler"

	// PytorchFrameworkName is the name of ML Framework
	PytorchFrameworkName = "pytorch"
	// PytorchReplicaTypeMaster is the type for Scheduler of distribute ML
	PytorchReplicaTypeMaster commonv1.ReplicaType = "Master"

	// TensorflowFrameworkName is the name of ML Framework
	TensorflowFrameworkName = "tensorflow"
	// TensorflowReplicaTypeChief is the type for Scheduler of distribute ML
	TensorflowReplicaTypeChief commonv1.ReplicaType = "Chief"

	// ReplicaTypeWorker this is also used for non-distributed AscendJob
	ReplicaTypeWorker commonv1.ReplicaType = "Worker"

	// DefaultRestartPolicy is default RestartPolicy for MSReplicaSpec.
	DefaultRestartPolicy = commonv1.RestartPolicyNever
)
