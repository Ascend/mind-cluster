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

// Package spacedetector is a DT collection for the func in homogenization_clustering.go
package spacedetector

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
)

func TestRecurseDimensionalClustering(t *testing.T) {
	convey.Convey("test RecurseDimensionalClustering", t, func() {
		// test max loop
		mock := gomonkey.NewPatches()
		defer mock.Reset()
		mock.ApplyFunc(oneDimensionalClustering, func([]float64, float64) ([]int, []float64) {
			return []int{}, nil
		})
		dataList := []float64{1}
		result := recurseDimensionalClustering(dataList, 0)
		convey.So(result, convey.ShouldBeEmpty)
	})
}
