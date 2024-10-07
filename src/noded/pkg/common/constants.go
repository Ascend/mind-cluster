/* Copyright(C) 2024. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package common for common function
package common

const (
	// FaultConfigCMName the name of fault config map
	FaultConfigCMName = "mindx-dl-node-fault-config"
	// FaultConfigCMNameSpace the name space of fault config map
	FaultConfigCMNameSpace = "mindx-dl"
	// NodeInfoCMNamePrefix the name prefix of node info config map
	NodeInfoCMNamePrefix = "mindx-dl-nodeinfo-"
	// NodeInfoCMNameSpace the name space of node info config map
	NodeInfoCMNameSpace = "mindx-dl"
	// NodeInfoCMDataKey the data key of node info config map
	NodeInfoCMDataKey = "NodeInfo"
	// ENVNodeNameKey the env key to get node name
	ENVNodeNameKey = "NODE_NAME"
	// RegexNodeNameKey the regex key of node name
	RegexNodeNameKey = "nodeName"
	// RegexFaultCodeKey the regex key of code str
	RegexFaultCodeKey = "faultCode"
	// MetaDataNameSpace the metadata key of name space
	MetaDataNameSpace = "metadata.namespace"
	// MetaDataName the metadata key of name
	MetaDataName = "metadata.name"

	// CmConsumer who uses these configmap
	CmConsumer = "mx-consumer-cim"
	// CmConsumerValue the value only for true
	CmConsumerValue = "true"
)

const (
	// FaultConfigKey the fault config json file key of nodeD configuration
	FaultConfigKey = "NodeDConfiguration.json"
	// FaultConfigFilePath the fault config file path
	FaultConfigFilePath = "/usr/local/NodeDConfiguration.json"
)

const (
	// KubeEnvMaxLength max k8s env length
	KubeEnvMaxLength = 230
)

// device fault level
const (
	// NotHandleFaultLevel the level of not handle fault
	NotHandleFaultLevel = iota
	// PreSeparateFaultLevel the level of pre-separate fault
	PreSeparateFaultLevel
	// SeparateFaultLevel the level of separate fault
	SeparateFaultLevel
)

// device fault name str
const (
	// NotHandleFault not handle fault
	NotHandleFault = "NotHandleFault"
	// PreSeparateFault pre-separate fault
	PreSeparateFault = "PreSeparateFault"
	// SeparateFault separate fault
	SeparateFault = "SeparateFault"
)

// node fault level
const (
	// NodeHealthyLevel the level of unhealthy node
	NodeHealthyLevel = iota
	// NodeSubHealthyLevel the level of sub-healthy node
	NodeSubHealthyLevel
	// NodeUnHealthyLevel the level of unhealthy node
	NodeUnHealthyLevel
)

// node fault name str
const (
	// NodeHealthy healthy node
	NodeHealthy = "Healthy"
	// PreSeparate sub-healthy node
	PreSeparate = "PreSeparate"
	// NodeUnHealthy unhealthy node
	NodeUnHealthy = "UnHealthy"
)

const (
	// HexBase Hexadecimal base number
	HexBase = 16
	// HexByteBase Hexadecimal base number in bytes
	HexByteBase = HexBase * HexBase
	// TwoByteSliceLength two byte slice length
	TwoByteSliceLength = 2
	// MaxSixTeenBitIntValue the int value of max sixteen bit data
	MaxSixTeenBitIntValue = 65535
	// zeroByte zero byte
	zeroByte = 0x00
)

const (
	// TotalEventsStartIndex total events number byte start index
	TotalEventsStartIndex = 4
	// TotalEventsEndIndex total events number byte end index
	TotalEventsEndIndex = 6
	// TotalEventsByteLength total events number byte length
	TotalEventsByteLength = 2
	// MsgEventsIndex msg events number byte index
	MsgEventsIndex = 6
	// EventFieldStartIndex msg events filed byte start index
	EventFieldStartIndex = 8
	// SingleEventBytes the byte length of an event
	SingleEventBytes = 15
	// ErrorCodeStartIndex error code byte start index
	ErrorCodeStartIndex = 0
	// ErrorCodeEndIndex error code byte end index
	ErrorCodeEndIndex = 4
	// SeverityIndex serverity byte index
	SeverityIndex = 8
	// DeviceTypeIndex device type byte index
	DeviceTypeIndex = 9
	// DeviceIdIndex device id byte index
	DeviceIdIndex = 10
)

const (
	// UnknownDevice unknown device type
	UnknownDevice = "UNKNOWN"
)
