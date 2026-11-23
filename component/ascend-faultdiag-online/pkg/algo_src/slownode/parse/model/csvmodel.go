/* Copyright 2025 Huawei Technologies Co., Ltd.

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

/*
Package model.
*/
package model

// StepGlobalRank global rank structure within the step, output CSV structure
type StepGlobalRank struct {
	// StepIndex is the iteration index
	StepIndex int64 `csv:"step_index"`
	// ZPDevice is the time-consuming metric of the ZP operator on the device side
	ZPDevice int64 `csv:"ZP_device"`
	// ZPHost is the time-consuming metric of the ZP operator on the host side
	ZPHost int64 `csv:"ZP_host"`
	// PPDevice is the time-consuming metric of the PP operator on the device side
	PPDevice int64 `csv:"PP_device"`
	// PPHost is the time-consuming metric of the PP operator on the host side
	PPHost int64 `csv:"PP_host"`
	// DataLoaderHost is the time-consuming metric of the dataloader operator on the host side
	DataLoaderHost int64 `csv:"dataloader_host"`
}

// StepIterateDelay iteration delay of one iteration, output CSV structure
type StepIterateDelay struct {
	// StepTime indicates the index of the iteration
	StepTime int64 `csv:"step time"`
	// Durations is the delay of the iteration, in nanoseconds (ns)
	Durations int64 `csv:"durations"`
}
