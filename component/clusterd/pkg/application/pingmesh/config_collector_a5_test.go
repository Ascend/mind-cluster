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

// Package pingmesh a series of function handle ping mesh configmap create/update/delete
package pingmesh

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"

	"clusterd/pkg/common/constant"
)

func TestIsValidConfigPingMesh(t *testing.T) {
	convey.Convey("Testing isValidConfigPingMesh case 1", t, func() {
		cfg := make(map[string]*constant.HccspingMeshItem)
		convey.So(isValidConfigPingMesh(cfg), convey.ShouldBeFalse)
	})

	convey.Convey("Testing isValidConfigPingMesh case 2", t, func() {
		cfg := make(map[string]*constant.HccspingMeshItem)
		cfg["1"] = nil
		convey.So(isValidConfigPingMesh(cfg), convey.ShouldBeFalse)
	})

	convey.Convey("Testing isValidConfigPingMesh case 2", t, func() {
		cfg := make(map[string]*constant.HccspingMeshItem)
		cfg["1"] = &constant.HccspingMeshItem{
			Activate: "1",
		}
		convey.So(isValidConfigPingMesh(cfg), convey.ShouldBeFalse)
	})

	convey.Convey("Testing isValidConfigPingMesh case 3", t, func() {
		cfg := make(map[string]*constant.HccspingMeshItem)
		cfg["1"] = &constant.HccspingMeshItem{
			Activate:     constant.RasNetDetectOnStr,
			TaskInterval: 0,
		}
		convey.So(isValidConfigPingMesh(cfg), convey.ShouldBeFalse)
	})

	convey.Convey("Testing isValidConfigPingMesh case 4", t, func() {
		cfg := make(map[string]*constant.HccspingMeshItem)
		cfg["1"] = &constant.HccspingMeshItem{
			Activate:     constant.RasNetDetectOnStr,
			TaskInterval: 1,
		}
		convey.So(isValidConfigPingMesh(cfg), convey.ShouldBeTrue)
	})
}
