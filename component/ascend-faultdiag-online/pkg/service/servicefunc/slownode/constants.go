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

// Package slownode a constants parameters for slownode
package slownode

const (
	// DeleteOperator informer operator
	DeleteOperator = "delete"
	// AddOperator informer operator
	AddOperator = "add"
	// UpdateOperator informer operator
	UpdateOperator = "update"
	// SlowNodeOn start slow node feature
	SlowNodeOn = 1
	// SlowNodeOff stop slow node feature
	SlowNodeOff = 0
	// FileMode write file mode
	FileMode = 0644

	defaultNamespace = "default"
	jobSummaryPrefix = "job-summary-"
	maxRetryCount    = 5

	keyJobName = "training.kubeflow.org/job-name"

	keyJobId     = "job_id"
	keyJobStatus = "job_status"
	isRunning    = "running"

	// NodeLevelDetectionResult is the path where store the tp/pp/ slow node algo result
	nodeLevelDetectionResult = "NodeLevelDetectionResult"

	slownodeAlgoResultSuffix = "_Result.json"
	slownodeAlgoResultPrefix = "slownode_"
	parallelGroupSuffix      = "_parallel_group.json"

	start = "start"
	stop  = "stop"

	success = "success"
)
