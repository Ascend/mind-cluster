/*
Copyright(C)2025. Huawei Technologies Co.,Ltd. All rights reserved.

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
Package fullmesh is one of policy generator for pingmeshv1
*/

package fullmesh

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"

	"nodeD/pkg/pingmeshv1/types"
	_ "nodeD/pkg/testtool"
)

func TestGenerate(t *testing.T) {
	convey.Convey("TestGenerate", t, func() {
		gen := New("node")
		convey.Convey("01-addrs with out local should return nil", func() {
			res := gen.Generate(map[string]types.SuperDeviceIDs{})
			convey.So(res, convey.ShouldBeNil)
		})
		convey.Convey("02-addrs with local should return local", func() {
			res := gen.Generate(map[string]types.SuperDeviceIDs{
				"node":  {"0": "000", "1": "111"},
				"node1": {"0": "001", "1": "112"},
			})
			convey.So(res, convey.ShouldNotBeNil)
			convey.So(res["0"][0], convey.ShouldEqual, "111")
			convey.So(res["0"][1], convey.ShouldEqual, "001")
			convey.So(res["1"][0], convey.ShouldEqual, "000")
			convey.So(res["1"][1], convey.ShouldEqual, "112")
		})
	})
}
