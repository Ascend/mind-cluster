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

// Package device a series of device function
package device

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/api/services/tasks/v1"
	"github.com/containerd/containerd/api/types/task"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/apimachinery/pkg/util/sets"

	"Ascend-device-plugin/pkg/common"
	npuCommon "ascend-common/devmanager/common"
)

// TestGetUseChips for test GetUsedChips
func TestGetUseChips(t *testing.T) {
	convey.Convey("test GetUsedChips", t, func() {
		tool := mockAscendTools()
		convey.Convey("when neither process nor containerd use chips, result should be empty", func() {
			common.ParamOption.PresetVDevice = false
			mock := gomonkey.ApplyPrivateMethod(reflect.TypeOf(&tool), "getChipsUsedByProcess",
				func(_ *AscendTools) sets.String {
					return sets.NewString()
				}).ApplyPrivateMethod(reflect.TypeOf(&tool), "getChipsUsedByContainerd",
				func(_ *AscendTools) sets.String {
					return sets.NewString()
				})
			defer mock.Reset()
			res := tool.GetUsedChips()
			convey.So(len(res), convey.ShouldEqual, 0)
		})
		convey.Convey("when both process and containerd use chips, result should be the union of the two",
			func() {
				common.ParamOption.PresetVDevice = false
				mock := gomonkey.ApplyPrivateMethod(reflect.TypeOf(&tool), "getChipsUsedByProcess",
					func(_ *AscendTools) sets.String {
						return sets.NewString().Insert(ascend910FakeID0, ascend910FakeID1)
					}).ApplyPrivateMethod(reflect.TypeOf(&tool), "getChipsUsedByContainerd",
					func(_ *AscendTools) sets.String {
						return sets.NewString().Insert(ascend910FakeID1, ascend910FakeID2)
					})
				defer mock.Reset()
				const expectLen = 3
				res := tool.GetUsedChips()
				convey.So(len(res), convey.ShouldEqual, expectLen)
			})
	})
}

// TestGetChipsUsedByProcess for test getChipsUsedByProcess
func TestGetChipsUsedByProcess(t *testing.T) {
	convey.Convey("test getChipsUsedByProcess", t, func() {
		tool := mockAscendTools()
		convey.Convey("when presetVDevice is false, used chips should be empty", func() {
			common.ParamOption.PresetVDevice = false
			res := tool.getChipsUsedByProcess()
			convey.So(len(res), convey.ShouldEqual, 0)
		})
		common.ParamOption.PresetVDevice = true
		convey.Convey("when get device list failed, used chips should be empty", func() {
			err := fmt.Errorf("failed to get device list")
			mockDeviceList := mockGetDeviceList(0, nil, err)
			defer mockDeviceList.Reset()
			res := tool.getChipsUsedByProcess()
			convey.So(len(res), convey.ShouldEqual, 0)

		})
		convey.Convey("when device list is empty, used chips should be empty", func() {
			mockDeviceList := mockGetDeviceList(0, []int32{}, nil)
			defer mockDeviceList.Reset()
			res := tool.getChipsUsedByProcess()
			convey.So(len(res), convey.ShouldEqual, 0)
		})
		mockDeviceList := mockGetDeviceList(1, []int32{0}, nil)
		defer mockDeviceList.Reset()
		convey.Convey("when get process info failed, used chips should be empty", func() {
			err := fmt.Errorf("failed to get device process info")
			mockDevProcessInfo := mockGetDevProcessInfo(nil, err)
			defer mockDevProcessInfo.Reset()
			res := tool.getChipsUsedByProcess()
			convey.So(len(res), convey.ShouldEqual, 0)
		})
		convey.Convey("when device process num is 0, used chips should be empty", func() {
			mockDevProcessInfo := mockGetDevProcessInfo(&npuCommon.DevProcessInfo{ProcNum: 0}, nil)
			defer mockDevProcessInfo.Reset()
			res := tool.getChipsUsedByProcess()
			convey.So(len(res), convey.ShouldEqual, 0)
		})
		convey.Convey("when device process num is not 0, used chips should not be empty", func() {
			mockDevProcessInfo := mockGetDevProcessInfo(&npuCommon.DevProcessInfo{ProcNum: 1}, nil)
			defer mockDevProcessInfo.Reset()
			res := tool.getChipsUsedByProcess()
			convey.So(len(res), convey.ShouldEqual, 1)
		})
	})
}

var testErr = fmt.Errorf("test error")

func testGetChipsUsedByContainerdCase1(tool AscendTools) {
	convey.Convey("when containerd client is nil, result should be empty", func() {
		tool.containerdClient = nil
		res := tool.getChipsUsedByContainerd()
		convey.So(len(res), convey.ShouldEqual, 0)
	})
	convey.Convey("when list namespaces failed, result should be empty", func() {
		mock := gomonkey.ApplyMethodReturn(tool.containerdClient.NamespaceService(),
			"List", nil, testErr)
		defer mock.Reset()
		res := tool.getChipsUsedByContainerd()
		convey.So(len(res), convey.ShouldEqual, 0)
	})
}

func testGetChipsUsedByContainerdCase2(tool AscendTools) {
	convey.Convey("when list tasks failed, result should be empty", func() {
		mock := gomonkey.ApplyMethodReturn(tool.containerdClient.TaskService(),
			"List", nil, testErr)
		defer mock.Reset()
		res := tool.getChipsUsedByContainerd()
		convey.So(len(res), convey.ShouldEqual, 0)
	})
	convey.Convey("when taskList is empty, result should be empty", func() {
		mock := gomonkey.ApplyMethodReturn(tool.containerdClient.TaskService(),
			"List", &tasks.ListTasksResponse{Tasks: []*task.Process{}}, nil)
		defer mock.Reset()
		res := tool.getChipsUsedByContainerd()
		convey.So(len(res), convey.ShouldEqual, 0)
	})
}

func testGetChipsUsedByContainerdCase3(tool AscendTools) {
	convey.Convey("when container use chip by ascend runtime, result should not be empty", func() {
		mock := gomonkey.ApplyPrivateMethod(reflect.TypeOf(&tool), "getDeviceWithAscendRuntime",
			func(_ *AscendTools, _ containerd.Container, _ context.Context) sets.String {

				return sets.NewString().Insert(ascend910FakeID0)
			})
		defer mock.Reset()
		res := tool.getChipsUsedByContainerd()
		convey.So(len(res), convey.ShouldEqual, 1)
	})
	convey.Convey("when container not use chip by ascend runtime, result should not be empty", func() {
		mock := gomonkey.ApplyPrivateMethod(reflect.TypeOf(&tool), "getDeviceWithAscendRuntime",
			func(_ *AscendTools, _ containerd.Container, _ context.Context) sets.String {
				return sets.NewString()
			}).ApplyPrivateMethod(reflect.TypeOf(&tool), "getDeviceWithoutAscendRuntime",
			func(_ *AscendTools, _ containerd.Container, _ context.Context) sets.String {
				return sets.NewString().Insert(ascend910FakeID0)
			})
		defer mock.Reset()
		res := tool.getChipsUsedByContainerd()
		convey.So(len(res), convey.ShouldEqual, 1)
	})
}

// TestGetChipsUsedByContainerd for test getChipsUsedByContainerd
func TestGetChipsUsedByContainerd(t *testing.T) {
	convey.Convey("test getChipsUsedByContainerd", t, func() {
		tool := mockAscendTools()
		testGetChipsUsedByContainerdCase1(tool)
		namespaceList := []string{"default"}
		mockListNameSpace := gomonkey.ApplyMethodReturn(tool.containerdClient.NamespaceService(),
			"List", namespaceList, nil)
		defer mockListNameSpace.Reset()
		testGetChipsUsedByContainerdCase2(tool)
		mockListTask := gomonkey.ApplyMethodReturn(tool.containerdClient.TaskService(),
			"List", &tasks.ListTasksResponse{Tasks: []*task.Process{&task.Process{}}}, nil)
		defer mockListTask.Reset()
		convey.Convey("when load container failed, result should be empty", func() {
			mock := gomonkey.ApplyMethodReturn(tool.containerdClient,
				"LoadContainer", nil, testErr)
			defer mock.Reset()
			res := tool.getChipsUsedByContainerd()
			convey.So(len(res), convey.ShouldEqual, 0)
		})
		mockLoadContainer := gomonkey.ApplyMethodReturn(tool.containerdClient,
			"LoadContainer", *new(containerd.Container), nil)
		defer mockLoadContainer.Reset()
		testGetChipsUsedByContainerdCase3(tool)
	})
}
