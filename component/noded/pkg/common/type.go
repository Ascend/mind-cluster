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

var (
	// ParamOption for option
	ParamOption Option
)

// Option struct definition
type Option struct {
	HeartbeatInterval int
	MonitorPeriod     int
}

// FaultProcessor fault processor responsibility chain interface
type FaultProcessor interface {
	Execute(*FaultDevInfo)
	SetNextFaultProcessor(FaultProcessor)
}

// ConfigProcessor fault processor responsibility chain interface
type ConfigProcessor interface {
	UpdateConfig(*FaultConfig) error
	SetNextConfigProcessor(ConfigProcessor)
}

// FaultDevInfo fault device info
type FaultDevInfo struct {
	FaultDevList      []*FaultDev
	HeartbeatTime     int64
	HeartbeatInterval int
	NodeStatus        string
}

// FaultDev fault device struct
type FaultDev struct {
	DeviceType string
	DeviceId   int64
	FaultCode  []string
	FaultLevel string
}

// FaultConfig fault config
type FaultConfig struct {
	FaultTypeCode *FaultTypeCode
}

// FaultTypeCode fault type code
type FaultTypeCode struct {
	NotHandleFaultCodes   []string
	PreSeparateFaultCodes []string
	SeparateFaultCodes    []string
}

// NodeInfoCM the config map struct of node info
type NodeInfoCM struct {
	NodeInfo  FaultDevInfo
	CheckCode string
}

// FaultEvent the fault event from ipmi
type FaultEvent struct {
	ErrorCode  string
	Severity   int64
	DeviceType string
	DeviceId   int64
}
