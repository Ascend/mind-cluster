/*
   Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.
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

package utils

import (
	"encoding/json"
	"errors"
	"os"
	"sync"
	"testing"

	ascutils "ascend-common/common-utils/utils"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
)

func resetMappingState() *gomonkey.Patches {
	p := gomonkey.NewPatches()
	p.ApplyGlobalVar(&dpuToNpuOnce, sync.Once{})
	dpuToNpuIDs = nil
	return p
}

func TestInitNpuNicMappingFileErrorAndSuccess(t *testing.T) {
	convey.Convey("When os.ReadFile fails, mapping stays empty and GetAffectedNPU returns empty slice", t, func() {
		resetPatches := resetMappingState()
		defer resetPatches.Reset()

		patches := gomonkey.ApplyFunc(ascutils.RealFileChecker,
			func(_ string, _, _ bool, _ int64) (string, error) {
				return "", errors.New("file not exist")
			})
		defer patches.Reset()

		InitNpuNicMapping()
		convey.So(GetAffectedNPU("enp0s1"), convey.ShouldBeEmpty)
	})

	convey.Convey("When config parses, primary DPU reverse index is built", t, func() {
		resetPatches := resetMappingState()
		defer resetPatches.Reset()

		data := []byte(`{"npuNics":[{"npuId":0,"nicNames":["enp0s1","enp0s2"]},{"npuId":1,"nicNames":["enp0s3"]}]}`)
		patches := gomonkey.ApplyFunc(ascutils.RealFileChecker,
			func(_ string, _, _ bool, _ int64) (string, error) {
				return npuNicMappingConfigPath, nil
			})
		defer patches.Reset()
		rfPatches := gomonkey.ApplyFunc(os.ReadFile, func(_ string) ([]byte, error) {
			return data, nil
		})
		defer rfPatches.Reset()

		InitNpuNicMapping()
		convey.So(GetAffectedNPU("enp0s1"), convey.ShouldResemble, []int{0})
		convey.So(GetAffectedNPU("enp0s3"), convey.ShouldResemble, []int{1})
		convey.So(GetAffectedNPU("enp0s2"), convey.ShouldBeEmpty)
	})
}

func TestInitNpuNicMappingArrayFormatSelectByForm(t *testing.T) {
	convey.Convey("When config is array format and card_type is A5Server, "+
		"Server entry is selected and reverse index built", t, func() {
		resetPatches := resetMappingState()
		defer resetPatches.Reset()

		cfg := []byte(`[
			{"productType":"PoD","npuNics":[{"npuId":0,"nicNames":["ens0f0","ens0f2"]}]},
			{"productType":"Server","npuNics":[
				{"npuId":0,"nicNames":["ens0f0"]},
				{"npuId":4,"nicNames":["ens2f0"]}
			]}
		]`)
		patches := gomonkey.ApplyFunc(ascutils.RealFileChecker,
			func(_ string, _, _ bool, _ int64) (string, error) {
				return npuNicMappingConfigPath, nil
			})
		defer patches.Reset()
		rfPatches := gomonkey.ApplyFunc(os.ReadFile, func(path string) ([]byte, error) {
			if path == npuNicMappingConfigPath {
				return cfg, nil
			}
			return nil, errors.New("unexpected read: " + path)
		})
		defer rfPatches.Reset()
		rlbPatches := gomonkey.ApplyFunc(ascutils.ReadLimitBytesWithSymlink,
			func(_ string, _ int, _ func(string) bool) ([]byte, error) {
				return []byte("A5Server"), nil
			})
		defer rlbPatches.Reset()

		InitNpuNicMapping()
		convey.So(GetAffectedNPU("ens0f0"), convey.ShouldResemble, []int{0})
		convey.So(GetAffectedNPU("ens2f0"), convey.ShouldResemble, []int{4})
		convey.So(GetAffectedNPU("ens0f2"), convey.ShouldBeEmpty)
	})
}

func TestInitNpuNicMappingParseErrorAndEmptyEth(t *testing.T) {
	convey.Convey("When JSON is invalid, mapping stays empty", t, func() {
		resetPatches := resetMappingState()
		defer resetPatches.Reset()

		patches := gomonkey.ApplyFunc(ascutils.RealFileChecker,
			func(_ string, _, _ bool, _ int64) (string, error) {
				return npuNicMappingConfigPath, nil
			})
		defer patches.Reset()
		rfPatches := gomonkey.ApplyFunc(os.ReadFile, func(_ string) ([]byte, error) {
			return []byte("{invalid json"), nil
		})
		defer rfPatches.Reset()

		InitNpuNicMapping()
		convey.So(GetAffectedNPU("enp0s1"), convey.ShouldBeEmpty)
	})

	convey.Convey("When ethName is empty or unknown, GetAffectedNPU returns empty slice", t, func() {
		resetPatches := resetMappingState()
		defer resetPatches.Reset()

		data, _ := json.Marshal(NpuNicMapping{
			NpuNics: []NpuNicItem{{NpuId: 2, NicNames: []string{"enp0s9"}}},
		})
		patches := gomonkey.ApplyFunc(ascutils.RealFileChecker,
			func(_ string, _, _ bool, _ int64) (string, error) {
				return npuNicMappingConfigPath, nil
			})
		defer patches.Reset()
		rfPatches := gomonkey.ApplyFunc(os.ReadFile, func(_ string) ([]byte, error) { return data, nil })
		defer rfPatches.Reset()

		InitNpuNicMapping()
		convey.So(GetAffectedNPU(""), convey.ShouldResemble, []int{})
		convey.So(GetAffectedNPU("unknown-eth"), convey.ShouldResemble, []int{})
		convey.So(GetAffectedNPU("enp0s9"), convey.ShouldResemble, []int{2})
	})
}
