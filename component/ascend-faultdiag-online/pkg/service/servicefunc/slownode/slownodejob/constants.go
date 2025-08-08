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

// Package slownodejob a constants parameters for slownode
package slownodejob

// Step is the int type for process step
type Step int

const (
	// InitialStep is the first step for all job
	InitialStep Step = 0
	// ClusterStep1 is start all profiling
	ClusterStep1 Step = 1
	// ClusterStep2 is start slow node algo
	ClusterStep2 Step = 2
	// NodeStep1 is start data parse
	NodeStep1 Step = 1
	// NodeStep2 is report data profiling result
	NodeStep2 Step = 2
	// NodeStep3 is start slow node algo
	NodeStep3 Step = 3
)

const (
	recordsCapacity = 10
	channelCapacity = 5
)
