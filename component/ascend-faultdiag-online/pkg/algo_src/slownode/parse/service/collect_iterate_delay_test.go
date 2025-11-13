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

// Package service provides some DT collection for iteration delay
package service

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"

	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/model"
)

func TestCollectIterateDelay(t *testing.T) {
	convey.Convey("test CollectIterateDelay", t, func() {
		// empty input
		result, err := CollectIterateDelay([]*model.StepStartEndNs{})
		convey.So(err, convey.ShouldBeNil)
		convey.So(len(result), convey.ShouldEqual, 0)

		// normal input
		input := []*model.StepStartEndNs{
			nil,
			{Id: 1, StartNs: 1000, EndNs: 2000},
			{Id: 2, StartNs: 2000, EndNs: 5000},
		}
		expected := []*model.StepIterateDelay{
			{StepTime: 1, Durations: 1000},
			{StepTime: 2, Durations: 3000},
		}
		result, err = CollectIterateDelay(input)
		convey.So(err, convey.ShouldBeNil)
		convey.So(result, convey.ShouldResemble, expected)

		// negative duration
		input = []*model.StepStartEndNs{
			{Id: 1, StartNs: 3000, EndNs: 2000},
		}
		result, err = CollectIterateDelay(input)
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(result, convey.ShouldBeNil)
	})
}
