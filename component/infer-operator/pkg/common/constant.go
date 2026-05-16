/*
Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.

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
	"time"
)

const (
	// LabelKeyPrefix is the prefix of the labels
	LabelKeyPrefix = "infer.huawei.com/"
	HuaweiPrefix   = "huawei.com/"
	// OperatorNameKey is the key of the operator name
	OperatorNameKey = LabelKeyPrefix + "ascend-infer-operator"

	// VolcanoPodGroupCrdName is the name of volcano PodGroup CRD
	VolcanoPodGroupCrdName = "podgroups.scheduling.volcano.sh"
	// InstanceSetKind is InstanceSet kind in its gkv
	InstanceSetKind = "InstanceSet"
	// InferServiceSetControllerName is the name of the infer serviceset controller
	InferServiceSetControllerName = "inferserviceset-controller"
	// InferServiceControllerName is the name of the infer service controller
	InferServiceControllerName = "inferservice-controller"
	// InstanceSetControllerName is the name of the instance set controller
	InstanceSetControllerName = "instanceset-controller"
	// InferReschedulingControllerName is the name of the infer rescheduling controller
	InferReschedulingControllerName = "infer-rescheduling-controller"
	// DefaultReEnqueueInterval is the default re-enqueue interval when reconcile failed
	DefaultReEnqueueInterval = time.Second
	// NonRetriableRequeInterval is the non-retriable re-enqueue interval when reconcile failed
	NonRetriableRequeInterval = time.Minute

	// InferServiceNameLabelKey is the label key of the infer service name
	InferServiceNameLabelKey = LabelKeyPrefix + "inferservice-name"
	// InstanceSetNameLabelKey is the label key of the instance set name
	InstanceSetNameLabelKey = LabelKeyPrefix + "instanceset-name"
	// InstanceIndexLabelKey is the label key of the instance index
	InstanceIndexLabelKey = LabelKeyPrefix + "instanceset-index"
	// GangScheduleLabelKey is the label key of the gang schedule
	GangScheduleLabelKey = LabelKeyPrefix + "gang-schedule"
	// GroupNameAnnotationKey is the annotation key of the gang schedule group name
	GroupNameAnnotationKey = "scheduling.k8s.io/group-name"
	// InferServiceSetNameLabelKey is the label key of the infer serviceset name
	InferServiceSetNameLabelKey = LabelKeyPrefix + "inferserviceset-name"
	// InferServiceIndexLabelKey is the label key of the infer service index
	InferServiceIndexLabelKey = LabelKeyPrefix + "inferservice-index"

	// InstanceIndexEnvKey is env key used to identify instance index
	InstanceIndexEnvKey = "INSTANCE_INDEX"
	// InstanceRoleEnvKey is env key used to identify instance role
	InstanceRoleEnvKey = "INSTANCE_ROLE"
	// InferServiceIndexEnvKey is env key used to identify infer service index
	InferServiceIndexEnvKey = "INFER_SERVICE_INDEX"
	// InferServiceNameEnvKey is env key used to identify infer service set name
	InferServiceNameEnvKey = "INFER_SERVICE_NAME"

	// ValidateErrorReason is the reason of the validate error condition
	ValidateErrorReason = "ValidateError"
	// ServiceCreateReason is the reason of service create error condition
	ServiceCreateReason = "ServiceCreateError"
	// InferServiceSetReadyReason is the reason of the infer serviceset ready condition
	InferServiceSetReadyReason = "InferServiceSetReady"
	// InferServiceReadyReason is the reason of the infer service ready condition
	InferServiceReadyReason = "InferServiceReady"
	// InstanceSetReadyReason is the reason of the instance set ready condition
	InstanceSetReadyReason = "InstanceSetReady"
	// InstanceReadyReason is the reason of the instance ready condition
	InstanceReadyReason = "InstanceReady"
	// AllWorkloadReadyReason is the reason of instanceSet ready condition
	AllWorkloadReadyReason = "AllWorkLoadReady"
	// AllWorkloadReadyMessage is the message of instanceSet ready condition
	AllWorkloadReadyMessage = "All WorkLoad replicas are ready"
	// WorkloadNotReadyReason is the reason of instanceSet's workload not ready condition
	WorkloadNotReadyReason = "ReplicasNotReady"
	DefaultReplicas        = int32(1)
	// TrueBool is the value of the true boolean
	TrueBool = "true"
	// FalseBool is the value of the false boolean
	FalseBool = "false"
	// DefaultPortName is the default port name
	DefaultPortName = "infer"
	// DefaultPort is the default port
	DefaultPort = 8080
	// DefaultPriority is the default priority
	DefaultPriority = int32(32)
	// InferServiceNameSplitNum is the least number of segments after splitting infer service name
	InferServiceNameSplitNum = 2
	// BaseDec is the base for integer conversion
	BaseDec = 10
	// BitSize is the bit size for integer conversion
	BitSize = 32
)

// InferServiceSetConditionType is the type of the infer serviceset condition
type InferServiceSetConditionType string

const (
	// InferServiceSetReady means the infer serviceset is available
	InferServiceSetReady InferServiceSetConditionType = "Ready"
)

// InferServiceConditionType is the type of the infer service condition
type InferServiceConditionType string

const (
	// InferServiceReady means the infer service is available
	InferServiceReady InferServiceConditionType = "Ready"
)

// InstanceSetConditionType is the type of the instance set condition
type InstanceSetConditionType string

const (
	// InstanceSetReady means the instanceset is available
	InstanceSetReady InstanceSetConditionType = "Ready"
)

const (
	// DeleteOperator informer operator
	DeleteOperator = "delete"
	// AddOperator informer operator
	AddOperator = "add"
	// UpdateOperator informer operator
	UpdateOperator = "update"
)

const (
	// NamespacePath is the path of the namespace file in the service account directory
	NamespacePath = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
	// DefaultNamespace is the default namespace when unable to get namespace from service account file
	DefaultNamespace = "mindx-dl"
	// InferOperatorCfgName is the name of infer-operator-configmap
	InferOperatorCfgName = "infer-operator-config"
)

const (
	// MaxInferServiceReplicas is the max replicas of the infer service
	MaxInferServiceReplicas = 64
	// MaxRoleTypeCount is the max role type count of the infer service
	MaxRoleTypeCount = 32
	// MaxRoleReplicas is the max replicas of the role type of the infer service
	MaxRoleReplicas = 256
)

const (
	SchedulingStrategySequential       = "Sequential"
	SchedulingStrategyPriority         = "Priority"
	PrioritySchedulingStrategyLabelKey = LabelKeyPrefix + "priority-scheduling-strategy"
)
