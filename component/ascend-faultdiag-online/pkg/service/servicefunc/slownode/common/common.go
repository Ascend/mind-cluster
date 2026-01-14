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

// Package common provides some common fun
package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"ascend-faultdiag-online/pkg/model/slownode"
)

// ConvertMaptoStruct convert the callback data to struct
func ConvertMaptoStruct[T slownode.NodeAlgoResult | slownode.ClusterAlgoResult](
	data map[string]map[string]any, target *T) error {
	if len(data) == 0 {
		return fmt.Errorf("callback data is empty: %v", data)
	}
	if target == nil {
		return errors.New("target struct pointer is nil")
	}
	for _, v := range data {
		if len(v) == 0 {
			return fmt.Errorf("callback data is empty: %v", data)
		}
		for _, result := range v {
			dataBytes, err := json.Marshal(result)
			if err != nil {
				return err
			}
			if err = json.Unmarshal(dataBytes, target); err != nil {
				return err
			}
		}
	}
	return nil
}

// AreServersEqual return true if server1 equals server2
func AreServersEqual(servers1, servers2 []slownode.Server) bool {
	if len(servers1) != len(servers2) {
		return false
	}
	var sortFunc = func(servers []slownode.Server) {
		sort.Slice(servers, func(i, j int) bool {
			if servers[i].Sn != servers[j].Sn {
				return servers[i].Sn < servers[j].Sn
			}
			if servers[i].Ip != servers[j].Ip {
				return servers[i].Ip < servers[j].Ip
			}
			return len(servers[i].RankIds) < len(servers[j].RankIds)
		})
	}
	sortFunc(servers1)
	sortFunc(servers2)
	return reflect.DeepEqual(servers1, servers2)
}

// NodeRankValidator validate the node rank
func NodeRankValidator(nodeRank string) error {
	if strings.TrimSpace(nodeRank) == "" {
		return errors.New("node rank is empty")
	}
	// do not include space, /, \
	if strings.ContainsAny(nodeRank, " /\\") {
		return errors.New("contains invalid character: ' ', '/', '\\'")
	}
	return nil
}

// JobIdValidator validate the job id
func JobIdValidator(jobId string) error {
	// job id can only contain letters, numbers, and hyphens
	if strings.TrimSpace(jobId) == "" {
		return errors.New("job id is empty")
	}
	for _, char := range jobId {
		if !(char >= 'a' && char <= 'z') &&
			!(char >= 'A' && char <= 'Z') &&
			!(char >= '0' && char <= '9') &&
			!(char == '-') {
			return fmt.Errorf("contains invalid character: %c", char)
		}
	}
	return nil
}
