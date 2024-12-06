// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package constant a series of para
package constant

const (
	// DLNamespace is the namespace of MindX DL
	DLNamespace = "mindx-dl"
	// KubeNamespace is the namespace of k8s
	KubeNamespace = "kube-system"
	// DeviceInfoPrefix is prefix of device info name, which is reported by device-plugin
	DeviceInfoPrefix = "mindx-dl-deviceinfo-"
	// NodeInfoPrefix is prefix of node info name, which is reported by nodeD
	NodeInfoPrefix = "mindx-dl-nodeinfo-"
	// SwitchInfoPrefix is prefix of switch info name, which is reported by device-plugin
	SwitchInfoPrefix = "mindx-dl-switchinfo-"
	// DevInfoCMKey mindx-dl-deviceinfo configmap key
	DevInfoCMKey = "DeviceInfoCfg"
	// SwitchInfoCmKey is the key name of data of switchinfo configmap
	SwitchInfoCmKey = "SwitchInfoCfg"
	// NodeInfoCMKey mindx-dl-nodeinfo configmap key
	NodeInfoCMKey = "NodeInfo"
	// ClusterDeviceInfo the name of cluster device info config map
	ClusterDeviceInfo = "cluster-info-device-"
	// ClusterNodeInfo the name of cluster node info config map
	ClusterNodeInfo = "cluster-info-node-"
	// ClusterSwitchInfo the name of cluster switchinfo 1520 info config map
	ClusterSwitchInfo = "cluster-info-switch-"
)

const (
	// CmConsumerCIM who uses these configmap
	CmConsumerCIM = "mx-consumer-cim"
	// CmConsumer who uses these configmap
	CmConsumer = "mx-consumer-volcano"
	// CmConsumerValue the value only for true
	CmConsumerValue = "true"
)

type ConfigMapInterface interface {
	GetCmName() string
}

func (cm *DeviceInfo) GetCmName() string {
	return cm.CmName
}

func (cm *SwitchInfo) GetCmName() string {
	return cm.CmName
}

func (cm *NodeInfo) GetCmName() string {
	return cm.CmName
}
