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

// Package roceping for ping by icmp in RoCE mesh net between super pods in A5
package roceping

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

func TestGenerate(t *testing.T) {
	convey.Convey("TestGenerate", t, func() {
		gen := NewGenerator("node", "", "")
		convey.Convey("01-should return nil when get ping list file path failed", func() {
			patch := gomonkey.ApplyFuncReturn(slownet.GetRoCEPingListFilePath, "", errors.New("invalid path"))
			defer patch.Reset()
			res := gen.Generate(map[string]types.SuperDeviceIDs{})
			convey.So(res, convey.ShouldBeNil)
		})

		convey.Convey("02-should return nil when read limit bytes failed", func() {
			patch := gomonkey.ApplyFuncReturn(slownet.GetRoCEPingListFilePath, "string", nil)
			defer patch.Reset()
			patchValid := gomonkey.ApplyFuncReturn(slownet.CheckIsExistAndValid, nil)
			defer patchValid.Reset()
			readPatch := gomonkey.ApplyFuncReturn(utils.ReadLimitBytes, nil, errors.New("read failed"))
			defer readPatch.Reset()
			res := gen.Generate(map[string]types.SuperDeviceIDs{})
			convey.So(res, convey.ShouldBeNil)
		})

		convey.Convey("03-should return nil when unmarshal failed", func() {
			patch := gomonkey.ApplyFuncReturn(slownet.GetRoCEPingListFilePath, "string", nil)
			defer patch.Reset()
			patchValid := gomonkey.ApplyFuncReturn(slownet.CheckIsExistAndValid, nil)
			defer patchValid.Reset()
			readPatch := gomonkey.ApplyFuncReturn(utils.ReadLimitBytes, make([]byte, maxFileSize), nil)
			defer readPatch.Reset()
			unMarshalPatch := gomonkey.ApplyFuncReturn(json.Unmarshal, errors.New("invalid input data"))
			defer unMarshalPatch.Reset()
			res := gen.Generate(map[string]types.SuperDeviceIDs{})
			convey.So(res, convey.ShouldBeNil)
		})
		convey.Convey("04-should return success when input data is valid", func() {
			testData, err := makeTestData()
			convey.So(err, convey.ShouldBeNil)
			patch := gomonkey.ApplyFuncReturn(slownet.GetRoCEPingListFilePath, "string", nil)
			defer patch.Reset()
			patchValid := gomonkey.ApplyFuncReturn(slownet.CheckIsExistAndValid, nil)
			defer patchValid.Reset()
			readPatch := gomonkey.ApplyFuncReturn(utils.ReadLimitBytes, testData, nil)
			defer readPatch.Reset()
			res := gen.Generate(map[string]types.SuperDeviceIDs{
				"node": {"0": "000", "1": "111"},
			})
			convey.So(res, convey.ShouldNotBeNil)
		})
	})
}

func makeTestData() ([]byte, error) {
	const testInt = 1
	const testString = "2"
	testPingListInfos := make([]types.PingListInfo, 0)
	testPingListInfos = append(testPingListInfos, types.PingListInfo{
		TaskId:   testString,
		TaskType: testString,
		PingList: []types.PingItem{{
			SrcType:      testInt,
			DstType:      testInt,
			PktSize:      testInt,
			SrcCardPhyId: testInt,
			SrcAddr:      testString,
			DstAddr:      testString,
		},
		},
	})
	return json.Marshal(testPingListInfos)
}
