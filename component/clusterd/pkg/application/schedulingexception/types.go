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

type jobExceptionInfo struct {
	JobName   string          `json:"jobName"`
	JobType   string          `json:"jobType"`
	NameSpace string          `json:"nameSpace"`
	Condition conditionDetail `json:"conditions"`
}

type conditionDetail struct {
	Status  jobStatus `json:"status"`
	Reason  string    `json:"reason"`
	Message string    `json:"message"`
}

func (c conditionDetail) Equal(c2 conditionDetail) bool {
	return c.Status == c2.Status && c.Reason == c2.Reason && c.Message == c2.Message
}

type conditionIndices struct {
	jobEnqueueFailedIndex     int
	jobValidFailedIndex       int
	predicatedNodesErrorIndex int
	batchOrderFailedIndex     int
	notEnoughResourcesIndex   int
}
