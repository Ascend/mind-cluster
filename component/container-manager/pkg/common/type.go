/* Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package common a series common struct
package common

// NPUInfo phy related, static info
type NPUInfo struct {
	IP         string
	PhyID      int32
	LogicID    int32
	CardID     int32
	DeviceID   int32
	DevsOnRing []int32
}

// DevFaultInfo device's fault info
type DevFaultInfo struct {
	EventID       int64
	LogicID       int32
	ModuleType    int8 // ModuleType prototype is dcmi node_type
	ModuleID      int8 // ModuleID prototype is dcmi node_id
	SubModuleType int8 // SubModuleType prototype is dcmi sub_node_type
	SubModuleID   int8 // SubModuleID prototype is dcmi sub_node_id
	Assertion     int8
	PhyID         int32
	FaultLevel    string
	ReceiveTime   int64
}

var (
	// ParamOption for option
	ParamOption Option
)

// Option option param
type Option struct {
	CtrStrategy string
	SockPath    string
}

// CtrStatusInfo container status info for displaying
type CtrStatusInfo struct {
	CtrId           string
	Status          string
	StatusStartTime int64
	Description     string
}
