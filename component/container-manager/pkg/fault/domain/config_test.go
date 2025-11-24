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

// Package domain test for config
package domain

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/common-utils/utils"
	"container-manager/pkg/common"
)

const (
	testFilePath = "./testCfg.json"
	mode644      = 0644
	len5         = 5
)

var (
	testFaultCfg = `
{
  "NotHandleFaultCodes":[
    "80E21007","80E38003","80F78006","80C98006","80CB8006"
  ],
  "RestartRequestCodes":[
    "80C98008","80C98002","80C98003","80C98009","80CB8002"
  ],
  "RestartBusinessCodes":[
    "8C204E00","A8028802","A4302003","A4302004","A4302005"
  ],
  "FreeRestartNPUCodes":[
    "8C0E4E00","8C104E00","8C0C4E00","8C044E00","8C064E00"
  ],
  "RestartNPUCodes":[
    "8C03A000","8C1FA006","40F84E00","80E24E00","80E21E01"
  ],
  "SeparateNPUCodes":[
    "80E3A201","80E18402","80E0020B","817F8002","816F8002"
  ]
}
`
	nodeHandleCodes      = []int64{0x80E21007, 0x80E38003, 0x80F78006, 0x80C98006, 0x80CB8006}
	restartRequestCodes  = []int64{0x80C98008, 0x80C98002, 0x80C98003, 0x80C98009, 0x80CB8002}
	restartBusinessCodes = []int64{0x8C204E00, 0xA8028802, 0xA4302003, 0xA4302004, 0xA4302005}
	freeRestartNPUCodes  = []int64{0x8C0E4E00, 0x8C104E00, 0x8C0C4E00, 0x8C044E00, 0x8C064E00}
	restartNPUCodes      = []int64{0x8C03A000, 0x8C1FA006, 0x40F84E00, 0x80E24E00, 0x80E21E01}
	separateNPUCodes     = []int64{0x80E3A201, 0x80E18402, 0x80E0020B, 0x817F8002, 0x816F8002}
)

func prepareFaultCfg(t *testing.T) {
	err := os.WriteFile(testFilePath, []byte(testFaultCfg), mode644)
	if err != nil {
		t.Error(err)
	}
}

func resetFaultCodeCache() {
	faultCodeCfg = faultCodeCfgCache{
		NotHandleFaultCodes:  make(map[int64]struct{}),
		RestartRequestCodes:  make(map[int64]struct{}),
		RestartBusinessCodes: make(map[int64]struct{}),
		RestartNPUCodes:      make(map[int64]struct{}),
		FreeRestartNPUCodes:  make(map[int64]struct{}),
		SeparateNPUCodes:     make(map[int64]struct{}),
	}
}

func TestSaveFaultCodesToCache(t *testing.T) {
	prepareFaultCfg(t)
	convey.Convey("test function 'SaveFaultCodesToCache' success", t, func() {
		resetFaultCodeCache()
		fileData, err := utils.LoadFile(testFilePath)
		convey.So(err, convey.ShouldBeNil)
		p1 := gomonkey.ApplyFuncReturn(utils.LoadFile, fileData, nil)
		defer p1.Reset()
		var faultCodes FaultCodeFromFile
		err = json.Unmarshal(fileData, &faultCodes)
		convey.So(err, convey.ShouldBeNil)
		SaveFaultCodesToCache(faultCodes)
		convey.So(err, convey.ShouldBeNil)
		convey.So(len(faultCodeCfg.NotHandleFaultCodes), convey.ShouldResemble, len5)
		convey.So(len(faultCodeCfg.RestartRequestCodes), convey.ShouldResemble, len5)
		convey.So(len(faultCodeCfg.RestartBusinessCodes), convey.ShouldResemble, len5)
		for _, code := range nodeHandleCodes {
			_, ok := faultCodeCfg.NotHandleFaultCodes[code]
			convey.So(ok, convey.ShouldBeTrue)
		}
	})
}

func TestGetFaultLevelByCode(t *testing.T) {
	tests := []struct {
		name       string
		faultCodes []int64
		expected   string
	}{
		{
			name:       "01 the most severe fault level is NotHandleFault",
			faultCodes: nodeHandleCodes,
			expected:   common.NotHandleFault,
		},
		{
			name:       "02 the most severe fault level is RestartRequest",
			faultCodes: append(restartRequestCodes, nodeHandleCodes...),
			expected:   common.RestartRequest,
		},
		{
			name:       "03 the most severe fault level is RestartBusiness",
			faultCodes: append(restartBusinessCodes, restartRequestCodes...),
			expected:   common.RestartBusiness,
		},
		{
			name:       "04 the most severe fault level is FreeRestartNPU",
			faultCodes: append(freeRestartNPUCodes, nodeHandleCodes...),
			expected:   common.FreeRestartNPU,
		},
		{
			name:       "05 the most severe fault level is RestartNPU",
			faultCodes: append(restartNPUCodes, restartBusinessCodes...),
			expected:   common.RestartNPU,
		},
		{
			name:       "06 the most severe fault level is SeparateNPU",
			faultCodes: append(separateNPUCodes, restartNPUCodes...),
			expected:   common.SeparateNPU,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetFaultLevelByCode(tt.faultCodes)
			if result != tt.expected {
				t.Errorf("GetFaultLevelByCode() = %v, want %v", result, tt.expected)
			}
		})
	}
}
