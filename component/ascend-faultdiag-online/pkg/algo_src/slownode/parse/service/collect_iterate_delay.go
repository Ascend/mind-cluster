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

// Package service provides some functions relevant to the iteration delay
package service

import (
	"fmt"

	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/model"
)

// CollectIterateDelay 收集迭代时延，将用于劣化感知算法
func CollectIterateDelay(startEndNsList []*model.StepStartEndNs) ([]*model.StepIterateDelay, error) {
	var iterateDelayInfo = make([]*model.StepIterateDelay, 0, len(startEndNsList))
	for _, data := range startEndNsList {
		if data == nil {
			continue
		}
		duration := data.EndNs - data.StartNs
		if duration < 0 {
			return nil, fmt.Errorf("the iteration delay is less than 0, 'startNs' is: %d, 'endNs' is: %d",
				data.StartNs, data.EndNs)
		}
		iterateDelayInfo = append(iterateDelayInfo, &model.StepIterateDelay{StepTime: data.Id, Durations: duration})
	}
	return iterateDelayInfo, nil
}
