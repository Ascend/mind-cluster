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
	"time"

	v1 "github.com/kubeflow/common/pkg/apis/common/v1"
)

// kind and name
const (
	RayHeadType              = "Head"
	RayClusterControllerName = "raycluster-controller"
	RayHeadControllerName    = "rayhead-controller"
	RayWorkerControllerName  = "rayworker-controller"
	VerlJobControllerName    = "verljob-controller"
	VerlJobResourceKind      = "VerlJob"
	RayClusterResourceKind   = "RayCluster"
	RayHeadResourceKind      = "RayHead"
	RayWorkerResourceKind    = "RayWorker"
	RayScriptVolumeName      = "ray-env-script"
	// GangSchedulerName gang scheduler name.
	GangSchedulerName        = "volcano"
	RayClusterIdentification = "cluster"
	RayHeadIdentification    = "head"
	RayWorkerIdentification  = "worker"
	VerlJobIdentification    = "verl"
)

// config fields or label Name
const (
	AutoSubmitLabelKey       = "autoSubmit"
	VerlPathLabelKey         = "verlPath"
	VerlExecLabelKey         = "verlExec"
	VerlConfigLabelKey       = "verlConfig"
	CheckRayStatusLabelKey   = "checkRayStatus"
	RayAddrEnvKey            = "RAY_ADDRESS"
	RayLabelEnvKey           = "RAY_LABEL"
	RayResourcesEnvKey       = "RAY_RESOURCES"
	AscendSuperpodEnvKey     = "ASCEND_SUPERPOD_BLOCK_SIZE"
	RayGcsPortLabelKey       = "rayGcsPort"
	RayClientPortLabelKey    = "rayClientPort"
	RayDashboardPortLabelKey = "rayDashboardPort"
	ServiceNameLabelKey      = "serviceName"
	RayPortParamKey          = "port"
	RayAddressParamKey       = "address"
	RayDashHostParamKey      = "dashboard-host"
	RayDashPortParamKey      = "dashboard-port"
	RayNodeManagerParamKey   = "node-manager-port"
	RayObjectManagerParamKey = "object-manager-port"
	RayLabelParamKey         = "labels"
	RayResourceParamKey      = "resources"
	VerlReplicaType          = "VerlJobTemp"
	ConfigMapRayEnvKey       = "set_ray_env_%s.sh"
	RayInfoConfigMapName     = "ray-info-%s"
	// VolcanoTaskSpecKey volcano.sh/task-spec key used in pod annotation when EnableGangScheduling is true
	VolcanoTaskSpecKey               = "volcano.sh/task-spec"
	GangSchedulingPodGroupAnnotation = "scheduling.k8s.io/group-name"
	OwnerLabel                       = "owner"
	SpBlockAnnotationKey             = "sp-block"
	TpBlockAnnotationKey             = "tp-block"
	// PodVersionLabel version of the current pod, if the value is 0, the pod is created for the first time.
	// If the value is n (n > 0), the pod is rescheduled for the nth time.
	PodVersionLabel = "version"
)

// Reasons
const (
	ArgumentValidateReason = "ConfigValidateFailed"
	// FailedDeleteJobReason is added in an ascendjob when it is deleted failed.
	FailedDeleteJobReason = "FailedDeleteJob"
	// SuccessfulDeleteJobReason is added in an ascendjob when it is deleted successful.
	SuccessfulDeleteJobReason = "SuccessfulDeleteJob"
	// PodTemplateRestartPolicyReason is the reason of a job that set podTemplate restartPolicy.
	PodTemplateRestartPolicyReason = "SettedPodTemplateRestartPolicy"
	// JobSchedulerNameReason is the warning reason when other scheduler name is set in job with gang-scheduling enabled
	JobSchedulerNameReason = "SettedJobSchedulerName"
	// PodTemplateSchedulerNameReason is the warning reason when other scheduler name is set
	// in pod templates with gang-scheduling enabled
	PodTemplateSchedulerNameReason = "SettedPodTemplateSchedulerName"
	StatusUpdateFailedReason       = "StatusUpdateFailed"
)

// default values
const (
	DefaultRayGcsPort            = "6379"
	DefaultRayClientPort         = "10001"
	DefaultRayDashboardPort      = "8265"
	DefaultRestartPolicy         = v1.RestartPolicyNever
	DefaultMinMember             = 1
	DefaultNodeManagerPort       = "8001"
	DefaultObjectManagerPort     = "8002"
	DefaultDashboardHost         = "0.0.0.0"
	DefaultPodIndex              = "0"
	DefaultContainerName         = "ascend"
	DefaultRayCheckContainerName = "ray-check"
	DefaultCluster               = "cluster_0"
	SuperPodPrefix               = "superpod_"
	RackPrefix                   = "rack_"
	NodePrefix                   = "node_"
	DefaultLabelL1               = "virtual"
	// UnsetBackoffLimits default Re-scheduling Times of job, it stands for Unlimited.
	UnsetBackoffLimits = -1
	// DefaultPodVersion is the default version of pod.
	DefaultPodVersion = 0
	// Decimal stands for base-10.
	Decimal         = 10
	RayEnvMountPath = "/etc/profile.d"
)

// default config value
const (
	WorkQueueBaseDelay      = 5 * time.Millisecond
	WorkQueueMaxDelay       = 20 * time.Second
	WorkQueueQps            = 10
	WorkQueueBurst          = 100
	MaxReplicas             = 15000 // maximum limit of replicas
	ConfigMapRetry          = 3
	ConfigMapRetrySleepTime = 50 * time.Millisecond
	MinPort                 = 0
	MaxPort                 = 65535
)
