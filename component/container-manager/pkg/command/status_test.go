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

// Package command test for status command
package command

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/common-utils/utils"
	"container-manager/pkg/common"
)

const (
	testFilePath = "./testStatus.json"
	mode644      = 0644

	invalidCtrID = "12345678"
	validCtrID   = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
)

var (
	testStatusInfo = `
[
  {
    "ctrId": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
    "status": "resuming",
    "statusStartTime": 1700000000,
    "description": "The device has been recovered, but the container failed to be resumed. Please manually pull up the container"
  },
  {
    "ctrId": "8f3a7b2c9e1d4f6a8b5c3d7e2f1a9b6c8d4e7f2a5b9c1d3e6f8a2b4c7d9e1f3a5",
    "status": "paused",
    "statusStartTime": 1699996400,
    "description": "Device hot reset may fail. Please check of device status and recovery are required"
  }
]
`
)

func TestStatusCmd(t *testing.T) {
	convey.Convey("test cmd 'status' basic methods", t, func() {
		cmd := StatusCmd()
		convey.So(cmd.Name(), convey.ShouldEqual, "status")
		convey.So(cmd.Description(), convey.ShouldEqual,
			"Display container status information and container abnormal information")
		err := cmd.InitLog(context.Background())
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("test cmd 'status' methods CheckParam", t, func() {
		stCmd := statusCmd{containerID: validCtrID}
		convey.So(stCmd.CheckParam(), convey.ShouldBeNil)

		stCmd = statusCmd{containerID: invalidCtrID}
		convey.So(stCmd.CheckParam(), convey.ShouldBeNil)
	})
}

func TestStatusCmdExecute(t *testing.T) {
	prepareStatusInfo(t)
	cmd := StatusCmd()
	cmd.BindFlag()
	if err := flag.Set("containerID", validCtrID); err != nil {
		t.Errorf("set flag err: %v", err)
	}
	flag.Parse()

	convey.Convey("test method 'Execute' success", t, func() {
		fileData, err := utils.LoadFile(testFilePath)
		convey.So(err, convey.ShouldBeNil)
		p1 := gomonkey.ApplyFuncReturn(os.Stat, nil, nil).
			ApplyFuncReturn(utils.LoadFile, fileData, nil)
		defer p1.Reset()
		err = cmd.Execute(context.Background())
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("test method 'Execute' failed, file is not exist", t, func() {
		p1 := gomonkey.ApplyFuncReturn(os.Stat, nil, testErr)
		defer p1.Reset()
		err := cmd.Execute(context.Background())
		expErr := fmt.Errorf("get file %s info failed", common.StatusInfoFile)
		convey.So(err, convey.ShouldResemble, expErr)
	})
	convey.Convey("test method 'Execute' failed, load file error", t, func() {
		p1 := gomonkey.ApplyFuncReturn(os.Stat, nil, nil).
			ApplyFuncReturn(utils.LoadFile, nil, testErr)
		defer p1.Reset()
		err := cmd.Execute(context.Background())
		expErr := fmt.Errorf("read file %s failed", common.StatusInfoFile)
		convey.So(err, convey.ShouldResemble, expErr)
	})
	convey.Convey("test method 'Execute' failed, unmarshal error", t, func() {
		fileData, err := utils.LoadFile(testFilePath)
		convey.So(err, convey.ShouldBeNil)
		p1 := gomonkey.ApplyFuncReturn(os.Stat, nil, nil).
			ApplyFuncReturn(utils.LoadFile, fileData, nil).
			ApplyFuncReturn(json.Unmarshal, testErr)
		defer p1.Reset()
		err = cmd.Execute(context.Background())
		expErr := fmt.Errorf("unmarshal status info failed")
		convey.So(err, convey.ShouldResemble, expErr)
	})
}

func prepareStatusInfo(t *testing.T) {
	err := os.WriteFile(testFilePath, []byte(testStatusInfo), mode644)
	if err != nil {
		t.Errorf("write test status file err: %v", err)
	}
}
