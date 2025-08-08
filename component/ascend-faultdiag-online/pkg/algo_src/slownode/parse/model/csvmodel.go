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

/*
Package model.
*/
package model

// StepGlobalRank step内global rank结构 输出csv结构
type StepGlobalRank struct {
	StepIndex      int64 `csv:"step_index"`
	ZPDevice       int64 `csv:"ZP_device"`
	ZPHost         int64 `csv:"ZP_host"`
	PPDevice       int64 `csv:"PP_device"`
	PPHost         int64 `csv:"PP_host"`
	DataLoaderHost int64 `csv:"dataloader_host"`
}

// StepIterateDelay 一个迭代的迭代时延，输出csv结构
type StepIterateDelay struct {
	// StepTime 第几个迭代
	StepTime int64 `csv:"step time"`
	// Durations 该迭代的时延，单位ns
	Durations int64 `csv:"durations"`
}
