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

// Package app test for config
package app

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/common-utils/utils"
	"container-manager/pkg/common"
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
)

func TestLoadFaultCodeFromFile(t *testing.T) {
	prepareFaultCfg(t)
	common.ParamOption.FaultCfgPath = testFilePath
	convey.Convey("test function 'loadFaultCodeFromFile' success", t, func() {
		fileData, err := utils.LoadFile(testFilePath)
		convey.So(err, convey.ShouldBeNil)
		p1 := gomonkey.ApplyFuncReturn(utils.LoadFile, fileData, nil)
		defer p1.Reset()
		err = loadFaultCodeFromFile()
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("test function 'loadFaultCodeFromFile' failed, load file error", t, func() {
		p1 := gomonkey.ApplyFuncReturn(utils.LoadFile, nil, testErr)
		defer p1.Reset()
		err := loadFaultCodeFromFile()
		convey.So(err, convey.ShouldResemble, testErr)
	})
	convey.Convey("test function 'loadFaultCodeFromFile' failed, unmarshal error", t, func() {
		fileData, err := utils.LoadFile(testFilePath)
		convey.So(err, convey.ShouldBeNil)
		p1 := gomonkey.ApplyFuncReturn(utils.LoadFile, fileData, nil).
			ApplyFuncReturn(json.Unmarshal, testErr)
		defer p1.Reset()
		err = loadFaultCodeFromFile()
		expErr := fmt.Errorf("unmarshal custom fault code byte failed: %v", testErr)
		convey.So(err, convey.ShouldResemble, expErr)
	})
	convey.Convey("test function 'loadFaultCodeFromFile' success, no custom fault code file", t, func() {
		common.ParamOption.FaultCfgPath = ""
		err := loadFaultCodeFromFile()
		convey.So(err, convey.ShouldBeNil)
	})
}

func prepareFaultCfg(t *testing.T) {
	err := os.WriteFile(testFilePath, []byte(testFaultCfg), mode644)
	if err != nil {
		t.Error(err)
	}
}
