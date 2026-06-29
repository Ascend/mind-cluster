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

	"k8s.io/api/core/v1"

	ascendapi "ascend-common/api"
)

const (
	// LabelKeyPrefix is the prefix of the labels
	LabelKeyPrefix = "infer.huawei.com/"
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
	// RoleNameLabelKey is the label key of the role name (e.g. prefill, decode)
	RoleNameLabelKey = LabelKeyPrefix + "role-name"

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
	PriorityLabelKey                   = LabelKeyPrefix + "priority"
	SchedulingStrategyParallel         = "Parallel"
	SchedulingStrategyPriority         = "Priority"
	PrioritySchedulingStrategyLabelKey = LabelKeyPrefix + "priority-scheduling-strategy"
)

const (
	// ScalingPolicyTypeHPA means the scaling policy is HorizontalPodAutoscaler
	ScalingPolicyTypeHPA = "HPA"
)

const (
	// FaultSchedulingLabelKey describe resource deleting policy (force/grace)
	FaultSchedulingLabelKey = "fault-scheduling"
	// ExternalForceReschedulingValue describe external force rescheduling mode
	ExternalForceReschedulingValue = "external-force"
	// ExternalGraceReschedulingValue describe external grace rescheduling mode
	ExternalGraceReschedulingValue = "external-grace"
	// DefaultTerminationGracePeriodSeconds is the Kubernetes default grace period for pod termination
	DefaultTerminationGracePeriodSeconds = 30
	// PodStatusAnnotationKey describe pod status of infer service
	PodStatusAnnotationKey = ascendapi.ResourceNamePrefix + "pod-status"
	// CommonUnhealthyStatus describe common unhealthy status of infer service pod
	CommonUnhealthyStatus = "Unhealthy"
	// PodFailed the state of failed pod
	PodFailed = "pod-failed"
	// FaultRetryTimesLabelKey describe the retry times of the business fault
	FaultRetryTimesLabelKey = "fault-retry-times"
	// DeletingTrigger describe the trigger of deleting workloads
	DeletingTriggerAnnotationKey = LabelKeyPrefix + "deleting-trigger"
	// FaultRetryTimesCleanupInterval is the interval of cleanup fault retry times map
	FaultRetryTimesCleanupInterval = 24 * time.Hour
)

const (
	// ContainerSnapshotLabelKey is the label key of the container snapshot
	ContainerSnapshotLabelKey = LabelKeyPrefix + "container-snapshot"
	// ActiveLabelKey is the label key to indicate the pod is active
	ActiveLabelKey = LabelKeyPrefix + "active"
	// HostSnapshotAnnotationKey is the annotation key of the host snapshot
	HostSnapshotFlagAnnotationKey = "host_snapshot_finish_flag"
	// SnapshotModeAnnotationKey is the annotation key of the snapshot mode
	SnapshotModeAnnotationKey = "snapshot_mode"
	// HostSnapshotLoadPathEnvKey is env key used to identify host snapshot load path
	HostSnapshotLoadPathEnvKey = "host_snapshot_load_path"
	// HostSnapshotSavePathEnvKey is env key used to identify host snapshot save path
	HostSnapshotSavePathEnvKey = "host_snapshot_save_path"
	// HostSnapshotDirPathEnvKey is env key used to identify host snapshot dir path
	HostSnapshotDirPathEnvKey = "host_snapshot_dir_path"
	// HostSnapshotPathEnvKey is env key used to identify host snapshot path
	HostSnapshotPathEnvKey = "host_snapshot_path"
	// NpuSnapshotPathEnvKey is env key used to identify npu snapshot path
	NpuSnapshotPathEnvKey = "npu_snapshot_path"
	// PodNameEnvKey is env key used to identify pod name
	PodNameEnvKey = "pod_name"
	// GrusSnapshotRestoredFlag is env key used to identify snapshot restore flag
	GrusSnapshotRestoredFlag = "GRUS_SNAPSHOT_RESTORED_FLAG"
	// GrusSnapshotRestoredFlagKey is configmap key used to identify snapshot restore flag
	GrusSnapshotRestoredFlagKey = "GrusSnapshotRestoredFlag"
	// SnapshotMetadataPrefix snapshot meta data configmap name prefix
	SnapshotMetadataPrefix = "snapshot-metadata-"

	// SnapshotCheckInterval is the interval for checking snapshot status
	SnapshotCheckInterval = 5 * time.Second
	// SnapshotTimeout is the timeout for snapshot operation
	SnapshotTimeout = 40 * time.Minute
	// SnapshotStatusFileName is the name of the snapshot status file
	SnapshotStatusFileName = "snapshot_status.json"

	// HostSnapshotVolumnsName is the name of the host snapshot volume
	HostSnapshotVolumnsName = "host-snapshot"
	// NpuSnapshotVolumnsName is the name of the npu snapshot volume
	NpuSnapshotVolumnsName = "npu-snapshot"
)

const (
	// SnapshotStatusSuccess indicates snapshot was created successfully
	SnapshotStatusSuccess = "success"
	// SnapshotStatusFailed indicates snapshot creation failed
	SnapshotStatusFailed = "failed"
)

const (
	// PodSnapshotReadyGate is the name of the pod readiness gate for snapshot
	PodSnapshotReadyGate = LabelKeyPrefix + "snapshot-ready"
	// PodSnapshotReadyConditionType is the condition type for snapshot readiness
	PodSnapshotReadyConditionType = v1.PodConditionType(PodSnapshotReadyGate)
	// SnapshotConfigMapSuffix is the suffix for snapshot configmap name
	SnapshotConfigMapSuffix = "snapshot-env"

	// SnapshotLoadMode is the snapshot load mode
	SnapshotLoadMode = "load"
	// SnapshotSaveMode is the snapshot save mode
	SnapshotSaveMode = "save"
)
