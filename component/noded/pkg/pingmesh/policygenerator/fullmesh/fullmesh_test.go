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
Package fullmesh is one of policy generator for pingmesh
*/

package fullmesh

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/api/slownet"
	"ascend-common/common-utils/utils"
	"nodeD/pkg/pingmesh/types"
	_ "nodeD/pkg/testtool"
)

const (
	testInt    = 1
	testString = "2"
)

func TestGenerate(t *testing.T) {
	convey.Convey("TestGenerate", t, func() {
		gen := New("node", "", "")
		convey.Convey("01-case GetPingListFilePath error, should return nil", func() {
			gen := New("node", "", "")
			gen.local = "local"
			patch := gomonkey.ApplyFuncReturn(slownet.GetPingListFilePath, "string", errors.New(""))
			defer patch.Reset()
			res := gen.Generate(map[string]types.SuperDeviceIDs{"local": {}})
			convey.So(res, convey.ShouldEqual, nil)
		})
		convey.Convey("02-case ReadLimitBytes return error, should return nil", func() {
			gen := New("node", "", "")
			gen.local = "local"
			patch := gomonkey.ApplyFuncReturn(slownet.GetPingListFilePath, "string", nil).
				ApplyFuncReturn(utils.ReadLimitBytes, make([]byte, maxFileSize, maxFileSize), errors.New(""))
			defer patch.Reset()
			res := gen.Generate(map[string]types.SuperDeviceIDs{"local": {}})
			convey.So(res, convey.ShouldBeNil)
		})
		convey.Convey("03-case with out local should return nil", func() {
			patch := gomonkey.ApplyFuncReturn(slownet.GetPingListFilePath, "string", nil).
				ApplyFuncReturn(utils.ReadLimitBytes, make([]byte, maxFileSize, maxFileSize), nil)
			defer patch.Reset()
			res := gen.Generate(map[string]types.SuperDeviceIDs{})
			convey.So(res, convey.ShouldBeNil)
		})
		convey.Convey("04-case with local should return local", func() {
			testPingListInfos := make([]types.PingListInfo, 0)
			testPingListInfos = append(testPingListInfos, types.PingListInfo{
				TaskId: testString, TaskType: testString,
				PingList: []types.PingItem{{
					SrcType: testInt, DstType: testInt,
					PktSize: testInt, SrcCardPhyId: testInt,
					SrcAddr: testString, DstAddr: testString,
				},
				},
			})
			testData, err := json.Marshal(testPingListInfos)
			convey.So(err, convey.ShouldBeNil)
			patch := gomonkey.ApplyFuncReturn(slownet.GetPingListFilePath, "string", nil).
				ApplyFuncReturn(utils.ReadLimitBytes, testData, nil)
			defer patch.Reset()
			res := gen.Generate(map[string]types.SuperDeviceIDs{
				"node":  {"0": "000", "1": "111"},
				"node1": {"0": "001", "1": "112"},
			})
			convey.So(res, convey.ShouldNotBeNil)
		})
	})
}

func TestGetDestAddrMap(t *testing.T) {
	convey.Convey("Given a GeneratorImp instance", t, func() {
		convey.Convey("When the receiver is nil", func() {
			var g *GeneratorImp
			result := g.GetDestAddrMap()
			convey.So(result, convey.ShouldBeNil)
		})

		convey.Convey("When DestAddrMap is empty", func() {
			g := &GeneratorImp{DestAddrMap: make(map[string][]types.PingItem)}
			result := g.GetDestAddrMap()
			convey.So(result, convey.ShouldBeEmpty)
		})
	})
}
