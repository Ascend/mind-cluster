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

// Package policy is used for processing superpod information
package policy

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
)

// TestSpliceAlgorithmInput test for func spliceAlgorithmInput
func TestSpliceAlgorithmInput(t *testing.T) {
	convey.Convey("Test spliceAlgorithmInput", t, func() {
		convey.Convey("should return nil when arg key is not a number", func() {
			var npu2DFullMesh []string
			npuOutOfRackPath := map[string][]string{"key": {"1", "2"}}
			ret := spliceAlgorithmInput(npu2DFullMesh, npuOutOfRackPath)
			convey.So(ret, convey.ShouldBeNil)
		})

		convey.Convey("should return not nil when fullMesh is not empty", func() {
			npu2DFullMesh := []string{"1"}
			npuOutOfRackPath := map[string][]string{}
			ret := spliceAlgorithmInput(npu2DFullMesh, npuOutOfRackPath)
			convey.So(ret != nil, convey.ShouldBeTrue)
		})

		convey.Convey("should return nil when arg makeAlgoArg invalid", func() {
			npu2DFullMesh := []string{"1"}
			npuOutOfRackPath := map[string][]string{"key": {"1", "2"}}
			patch := gomonkey.ApplyFunc(makeAlgoArg, func(argMap map[string]any, npu2DFullMesh []string,
				npuOutOfRackPath map[string][]string) bool {
				return false
			})
			defer patch.Reset()
			ret := spliceAlgorithmInput(npu2DFullMesh, npuOutOfRackPath)
			convey.So(ret, convey.ShouldBeNil)
		})

		convey.Convey("should return nil when arg makeAlgoArg valid", func() {
			npu2DFullMesh := []string{"1"}
			npuOutOfRackPath := map[string][]string{"key": {"1", "2"}}
			patch := gomonkey.ApplyFunc(makeAlgoArg, func(argMap map[string]any, npu2DFullMesh []string,
				npuOutOfRackPath map[string][]string) bool {
				return true
			})
			defer patch.Reset()
			ret := spliceAlgorithmInput(npu2DFullMesh, npuOutOfRackPath)
			convey.So(ret, convey.ShouldNotBeNil)
		})
	})
}

// TestMakeAlgoArg test for func makeAlgoArg
func TestMakeAlgoArg(t *testing.T) {
	convey.Convey("Test spliceAlgorithmInput", t, func() {
		convey.Convey("should return nil when arg makeAlgoArg invalid", func() {
			argMap := make(map[string]any)
			npu2DFullMesh := []string{"1"}
			npuOutOfRackPath := map[string][]string{"abc": {"1", "22"}}
			ret := makeAlgoArg(argMap, npu2DFullMesh, npuOutOfRackPath)
			convey.So(ret, convey.ShouldBeFalse)
		})

		convey.Convey("should return nil when arg makeAlgoArg valid", func() {
			argMap := make(map[string]any)
			npu2DFullMesh := []string{"1"}
			npuOutOfRackPath := map[string][]string{"1": {"1", "22"}}
			ret := makeAlgoArg(argMap, npu2DFullMesh, npuOutOfRackPath)
			resultArgMap := make(map[string]any)
			resultArgMap["npu_npu"] = npu2DFullMesh
			npuNetPlanes := make(map[string]any)
			npuNetPlanes["netplane_0"] = []string{"1", "22"}
			resultArgMap["npu_netplane"] = npuNetPlanes
			convey.So(ret, convey.ShouldBeTrue)
			convey.So(argMap, convey.ShouldResemble, resultArgMap)
		})

		convey.Convey("should make A3 arg and return true when input exist netplane_0", func() {
			argMap := make(map[string]any)
			npuOutOfRackPath := map[string][]string{"netplane_0": {"0", "1"}}
			npu2DFullMesh := make([]string, 0)
			ret := makeAlgoArg(argMap, npu2DFullMesh, npuOutOfRackPath)
			resultArgMap := make(map[string]any)
			npuNetPlanes := make(map[string]any)
			npuNetPlanes["netplane_0"] = []string{"0", "1"}
			resultArgMap["npu_npu"] = npu2DFullMesh
			resultArgMap["npu_netplane"] = npuNetPlanes
			convey.So(ret, convey.ShouldBeTrue)
			convey.So(argMap, convey.ShouldResemble, resultArgMap)
		})
	})
}
