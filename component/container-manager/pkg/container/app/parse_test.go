/* Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package app test for container module
package app

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/smartystreets/goconvey/convey"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"container-manager/pkg/common"
	"container-manager/pkg/container/domain"
)

func TestMain(m *testing.M) {
	if err := initLog(); err != nil {
		return
	}
	code := m.Run()
	fmt.Printf("exit_code = %v\n", code)
}

func initLog() error {
	logConfig := &hwlog.LogConfig{
		OnlyToStdout: true,
	}
	if err := hwlog.InitRunLogger(logConfig, context.Background()); err != nil {
		fmt.Printf("init hwlog failed, %v\n", err)
		return errors.New("init hwlog failed")
	}
	return nil
}

func TestGetUsedDevsWithAscendRuntime(t *testing.T) {
	convey.Convey("test parse ascend visible devices env", t, func() {
		convey.Convey("comma style", func() {
			devs, err := getUsedDevsWithAscendRuntime(api.AscendDeviceInfo + "=0,1,2")
			convey.So(err, convey.ShouldBeNil)
			convey.So(devs, convey.ShouldResemble, []int32{0, 1, 2})
		})
		convey.Convey("minus style", func() {
			devs, err := getUsedDevsWithAscendRuntime(api.AscendDeviceInfo + "=0-3")
			convey.So(err, convey.ShouldBeNil)
			convey.So(devs, convey.ShouldResemble, []int32{0, 1, 2, 3})
		})
		convey.Convey("comma minus style", func() {
			devs, err := getUsedDevsWithAscendRuntime(api.AscendDeviceInfo + "=0-2,4")
			convey.So(err, convey.ShouldBeNil)
			convey.So(devs, convey.ShouldResemble, []int32{0, 1, 2, 4})
		})
		convey.Convey("ascend style", func() {
			devs, err := getUsedDevsWithAscendRuntime(api.AscendDeviceInfo + "=Ascend910-0,Ascend910-1")
			convey.So(err, convey.ShouldBeNil)
			convey.So(devs, convey.ShouldResemble, []int32{0, 1})
		})
		convey.Convey("npu style", func() {
			devs, err := getUsedDevsWithAscendRuntime(api.AscendDeviceInfo + "=npu-0,npu-1")
			convey.So(err, convey.ShouldBeNil)
			convey.So(devs, convey.ShouldResemble, []int32{0, 1})
		})
	})

	convey.Convey("test invalid ascend visible devices env", t, func() {
		convey.Convey("env without delimiter", func() {
			_, err := getUsedDevsWithAscendRuntime("invalid-env")
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("env with wrong key", func() {
			_, err := getUsedDevsWithAscendRuntime("OTHER=0,1")
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("minus style min bigger than max", func() {
			_, err := getUsedDevsWithAscendRuntime(api.AscendDeviceInfo + "=5-3")
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("ascend style missing id", func() {
			_, err := getUsedDevsWithAscendRuntime(api.AscendDeviceInfo + "=Ascend910")
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

func TestParseJobLabels(t *testing.T) {
	convey.Convey("test parse job labels", t, func() {
		convey.Convey("valid labels", func() {
			info := parseJobLabels(map[string]string{
				common.JobLabelID:      "job-1",
				common.JobLabelReplica: "2",
			})
			convey.So(info, convey.ShouldResemble, domain.JobInfo{JobID: "job-1", JobReplica: 2, EnableRecover: true})
		})
		convey.Convey("disable recover", func() {
			info := parseJobLabels(map[string]string{
				common.JobLabelID:            "job-1",
				common.JobLabelReplica:       "2",
				common.JobLabelEnableRecover: "false",
			})
			convey.So(info.EnableRecover, convey.ShouldBeFalse)
		})
		convey.Convey("invalid replica set to zero", func() {
			info := parseJobLabels(map[string]string{
				common.JobLabelID:      "job-1",
				common.JobLabelReplica: "invalid",
			})
			convey.So(info.JobID, convey.ShouldEqual, "job-1")
			convey.So(info.JobReplica, convey.ShouldEqual, 0)
		})
		convey.Convey("empty labels", func() {
			info := parseJobLabels(map[string]string{})
			convey.So(info.JobID, convey.ShouldEqual, "")
			convey.So(info.JobReplica, convey.ShouldEqual, 0)
			convey.So(info.EnableRecover, convey.ShouldBeTrue)
		})
	})
}
