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

import (
	"fmt"
	"strconv"
)

func convertRule(rule *Rule) ([]map[string]*groupInfo, error) {
	destList := make([]map[string]*groupInfo, len(rule.ElasticScalingList))
	for i, item := range rule.ElasticScalingList {
		dest := make(map[string]*groupInfo, len(item.GroupList))
		for _, grp := range item.GroupList {
			groupNum, err := strconv.Atoi(grp.GroupNum)
			if err != nil {
				return nil, err
			}
			serverNumPerGroup, err := strconv.Atoi(grp.ServerNumPerGroup)
			if err != nil {
				return nil, err
			}
			dest[grp.GroupName] = &groupInfo{
				groupNum:          groupNum,
				serverNumPerGroup: serverNumPerGroup,
			}
		}
		destList[i] = dest
	}
	if err := checkRule(destList); err != nil {
		return nil, err
	}

	return destList, nil
}

// {GroupList: []Group{{GroupName: "group0", GroupNum: "2", ServerNumPerGroup: "4"}}}, pre
// {GroupList: []Group{{GroupName: "group0", GroupNum: "1", ServerNumPerGroup: "8"}}}, now
// now's GroupName should exit is pre
// now's GroupNum and ServerNumPerGroup should not greater than pre
func checkRule(destList []map[string]*groupInfo) error {
	for i := 1; i < len(destList); i++ {
		for name, now := range destList[i] {
			pre, ok := destList[i-1][name]
			if !ok {
				return fmt.Errorf("the group %s is not exist in previous state", name)
			}

			if now.serverNumPerGroup > pre.serverNumPerGroup {
				return fmt.Errorf("the server_num_per_group in index<%d> can not be increased compared to the"+
					" previous state", i)
			}
			if now.groupNum > pre.groupNum {
				return fmt.Errorf("the group_num in index<%d> can not be increased compared to the previous "+
					"state", i)
			}
		}
	}
	return nil
}

func getDestOfCurrentState(rule []map[string]*groupInfo, cur map[string]int) int {
	for i := len(rule) - 1; i >= 0; i-- {
		for name, dest := range rule[i] {
			if cur[name] < dest.groupNum {
				return i
			}
		}
	}
	return invalidIndex
}
