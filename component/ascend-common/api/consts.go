// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package api common const
package api

// Env
const (
	NodeNameEnv = "NODE_NAME"

	// PtWorldSizeEnv the total number of npu used for the task for PyTorch
	PtWorldSizeEnv = "WORLD_SIZE"
	// PtLocalWorldSizeEnv number of npu used per pod for PyTorch
	PtLocalWorldSizeEnv = "LOCAL_WORLD_SIZE"
	// PtLocalRankEnv logic id List of npu used by pod for PyTorch
	PtLocalRankEnv = "LOCAL_RANK"

	// TfWorkerSizeEnv the total number of npu used for the task for TensorFlow
	TfWorkerSizeEnv = "CM_WORKER_SIZE"
	// TfLocalWorkerEnv number of npu used per pod for TensorFlow
	TfLocalWorkerEnv = "CM_LOCAL_WORKER"

	// MsWorkerNumEnv the total number of npu used for the task for MindSpore
	MsWorkerNumEnv = "MS_WORKER_NUM"
	// MsLocalWorkerEnv number of npu used per pod for MindSpore
	MsLocalWorkerEnv = "MS_LOCAL_WORKER"
)

// NPU
const (
	ResourceNamePrefix = "huawei.com/"
)

// NameSpace
const (
	DLNamespace = "mindx-dl"
	ClusterNS   = "cluster-system"
	KubeNS      = "kube-system"
)

// Node
const (
	// NPUChipMemoryLabel label value is npu chip memory
	NPUChipMemoryLabel = "mind-cluster/npu-chip-memory"

	// NodeSNAnnotation annotation value is node sn
	NodeSNAnnotation = "product-serial-number"
	// BaseDevInfoAnno annotation value is device base info
	BaseDevInfoAnno = "baseDeviceInfos"
)

// Pod
const (
	// PodUsedHardwareTypeAnno annotation value is the hardware type that real used in pod
	PodUsedHardwareTypeAnno = "mind-cluster/hardware-type"
	// Pod910DeviceAnno annotation value is for generating 910 hccl rank table
	Pod910DeviceAnno = "ascend.kubectl.kubernetes.io/ascend-910-configuration"
	// PodRankIndexAnno annotation value is rank index of the pod
	PodRankIndexAnno = "hccl/rankIndex"
	// SuperPodIDAnno annotation key of the super pod id
	SuperPodIDAnno = "super-pod-id"
)

// PodGroup
const (
	// AtlasTaskLabel label value task kind, eg. ascend-910, ascend-{xxx}b
	AtlasTaskLabel = "ring-controller.atlas"
)

// ConfigMap
const (
	// DeviceInfoCMDataKey device-info-cm data key, record device info
	DeviceInfoCMDataKey = "DeviceInfoCfg"
	// SwitchInfoCMDataKey device-info-cm data key, record switch info
	SwitchInfoCMDataKey = "SwitchInfoCfg"
	// NodeInfoCMDataKey node-info-cm data key, record node info
	NodeInfoCMDataKey = "NodeInfo"
	// PubFaultCMDataKey public fault cm data key, record public fault info
	PubFaultCMDataKey = "PublicFault"

	// CIMCMLabelKey cm label key, who uses these cms
	CIMCMLabelKey = "mx-consumer-cim"
	// PubFaultCMLabelKey public fault cm label key
	PubFaultCMLabelKey = "mc-consumer-publicfault"
)
