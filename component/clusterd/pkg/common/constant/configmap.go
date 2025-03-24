// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package constant a series of para
package constant

const (
	// DLNamespace is the namespace of MindX DL
	DLNamespace = "mindx-dl"
	// ClusterNamespace mind cluster cm namespace
	ClusterNamespace = "cluster-system"
	// KubeNamespace is the namespace of k8s
	KubeNamespace = "kube-system"
	// DeviceInfoPrefix is prefix of device info name, which is reported by device-plugin
	DeviceInfoPrefix = "mindx-dl-deviceinfo-"
	// NodeInfoPrefix is prefix of node info name, which is reported by nodeD
	NodeInfoPrefix = "mindx-dl-nodeinfo-"
	// SwitchInfoPrefix is prefix of switch info name, which is reported by device-plugin
	SwitchInfoPrefix = "mindx-dl-switchinfo-"
	// MindIeRanktablePrefix is prefix of mindie ranktable cm name
	MindIeRanktablePrefix = "rings-config-"
	// StatisticFaultCMName statistic fault configmap name
	StatisticFaultCMName = "statistic-fault-info"
	// DevInfoCMKey mindx-dl-deviceinfo configmap key
	DevInfoCMKey = "DeviceInfoCfg"
	// PubFaultCMKey public fault configmap key
	PubFaultCMKey = "PublicFault"
	// SwitchInfoCmKey is the key name of data of switchinfo configmap
	SwitchInfoCmKey = "SwitchInfoCfg"
	// NodeInfoCMKey mindx-dl-nodeinfo configmap key
	NodeInfoCMKey = "NodeInfo"
	// StatisticPubFaultKey configmap statistic-fault-info key Fault
	StatisticPubFaultKey = "PublicFaults"
	// StatisticFaultNumKey configmap statistic-fault-info key FaultNum
	StatisticFaultNumKey = "FaultNum"
	// StatisticFaultDescKey configmap statistic-fault-info key Description
	StatisticFaultDescKey = "Description"
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
	// CmConsumerPubFault cm label for public fault
	CmConsumerPubFault = "mc-consumer-publicfault"
	// CmStatisticFault cm label for fault statistic
	CmStatisticFault = "mc-statistic-fault"
	// CmConsumer who uses these configmap
	CmConsumer = "mx-consumer-volcano"
	// CmConsumerValue the value only for true
	CmConsumerValue = "true"
)
