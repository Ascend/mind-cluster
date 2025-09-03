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

// Package constant a package for constant
package constant

// ProfilingDomainCmd Profiling Domain Cmd
type ProfilingDomainCmd struct {
	DefaultDomainAble bool
	CommDomainAble    bool
}

// ProfilingExecRes Profiling Execute Result of Worker
type ProfilingExecRes struct {
	status string
}

// ProfilingResult Profiling Execute Result of Worker, include DefaultDomain and CommDomain
type ProfilingResult struct {
	DefaultDomain ProfilingExecRes
	CommDomain    ProfilingExecRes
}

// ProfilingSwitch is the struct for serialization and deserialization of profiling switches
type ProfilingSwitch struct {
	CommunicationOperator string
	Step                  string
	SaveCheckpoint        string
	FP                    string
	DataLoader            string
}

// ProfilingWorkerState Profiling Worker State
type ProfilingWorkerState struct {
	state string
}

// ControllerMessage define the message from controller
type ControllerMessage struct {
	// Actions indicate the action from clusterd
	Actions []string `json:"actions,omitempty"`
	// Action indicate the action from controller
	Action string `json:"action,omitempty"`
	// Code indicate the controller return code
	Code int `json:"code,omitempty"`
	// msg indicate the controller return message
	Msg string `json:"msg,omitempty"`
	// Strategy indicate the clusterd stratege
	Strategy string `json:"strategy,omitempty"`
	// Strategy_list indicate the controller strategies
	StrategyList []string `json:"strategy_list,omitempty"`
	// Fault_ranks indicate the fault ranks infomation
	FaultRanks map[int]int `json:"fault_ranks,omitempty"`
	// Params indicate the controller params
	Params string `json:"params,omitempty"`
}
