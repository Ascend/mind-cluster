/* Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package schedulingexception is for collecting scheduling exception

package schedulingexception

const (
	defaultCheckInterval = 10
	cmName               = "scheduling-exception-report"
	cmNamespace          = "cluster-system"
	invalidIndex         = -1
)

const (
	jobEnqueueFailedReason    = "JobEnqueueFailed"
	jobValidateFailedReason   = "JobValidateFailed"
	nodePredicateFailedReason = "NodePredicateFailed"
	batchOrderFailedReason    = "BatchOrderFailed"
	notEnoughResourcesReason  = "NotEnoughResources"
)

type jobStatus string

const (
	jobStatusEmpty       jobStatus = "JobEmptyStatus"
	jobStatusInitialized jobStatus = "JobInitialized"
	jobStatusFailed      jobStatus = "JobFailed"
	podGroupCreated      jobStatus = "PodGroupCreated"
	podGroupPending      jobStatus = "PodGroupPending"
	podGroupInqueue      jobStatus = "PodGroupInqueue"
	podGroupUnknown      jobStatus = "PodGroupUnknown"
	podGroupRunning      jobStatus = "PodGroupRunning"
)
