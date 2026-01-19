/*
Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.

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
Package scaling is using for scale AscendJobs.
*/

package scaling

// Rule is the struct of the scaling rule.
type Rule struct {
	Version            string `json:"version"`
	ElasticScalingList []Item `json:"elastic_scaling_list"`
}

// Item is the definition of groups.
type Item struct {
	GroupList []Group `json:"group_list"`
}

// Group is the destination of the group.
type Group struct {
	GroupName         string `json:"group_name"`
	GroupNum          string `json:"group_num"`
	ServerNumPerGroup string `json:"server_num_per_group"`
}

type groupInfo struct {
	groupNum          int
	serverNumPerGroup int
}

const (
	scalingRuleKey   = "mind-cluster/scaling-rule"
	groupNameKey     = "mind-cluster/group-name"
	jobGroupNameKey  = "jobID"
	configmapRuleKey = "elastic_scaling.json"
	invalidIndex     = -1
)
