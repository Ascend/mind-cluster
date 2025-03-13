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
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/api/services/tasks/v1"
	"github.com/containerd/containerd/api/types/task"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/containers"
	"github.com/containerd/containerd/oci"
	"github.com/gogo/protobuf/types"
	"github.com/opencontainers/runtime-spec/specs-go"
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

// TestGetDeviceWithAscendRuntime for test getDeviceWithAscendRuntime
func TestGetDeviceWithAscendRuntime(t *testing.T) {
	tool := mockAscendTools()
	convey.Convey("test getDeviceWithAscendRuntime", t, func() {
		convey.Convey("01-get info failed, should return empty sets", func() {
			patch := gomonkey.ApplyFuncReturn((*MockContainer).Info, nil, errors.New("get info failed"))
			defer patch.Reset()
			chips := tool.getDeviceWithAscendRuntime(MockContainer{}, nil)
			convey.So(chips, convey.ShouldResemble, sets.NewString())
		})
		convey.Convey("02-get spec failed, should return empty sets", func() {
			patch := gomonkey.ApplyFuncReturn(getContainerValidSpec, nil, errors.New("get spec failed"))
			defer patch.Reset()
			chips := tool.getDeviceWithAscendRuntime(MockContainer{}, nil)
			convey.So(chips, convey.ShouldResemble, sets.NewString())
		})
		convey.Convey("03-invalid env, should return empty sets", func() {
			spec := &oci.Spec{Process: &specs.Process{
				Env: []string{"ASCEND_VISIBLE_DEVICES", "FAKE_ENV"},
			}}
			patch := gomonkey.ApplyFuncReturn(getContainerValidSpec, spec, nil)
			defer patch.Reset()
			chips := tool.getDeviceWithAscendRuntime(MockContainer{}, nil)
			convey.So(chips, convey.ShouldResemble, sets.NewString())
		})
		convey.Convey("04-get dev success, should return chip set", func() {
			spec := &oci.Spec{Process: &specs.Process{
				Env: []string{"ASCEND_VISIBLE_DEVICES=0,1", "FAKE_ENV"},
			}}
			patch := gomonkey.ApplyFuncReturn(getContainerValidSpec, spec, nil)
			defer patch.Reset()
			chips := tool.getDeviceWithAscendRuntime(MockContainer{}, nil)
			dev0 := fmt.Sprintf("%s-%d", common.Ascend910, 0)
			dev1 := fmt.Sprintf("%s-%d", common.Ascend910, 1)
			convey.So(chips, convey.ShouldResemble, sets.NewString(dev0, dev1))
		})
	})
}

// TestGetDeviceWithoutAscendRuntime for test getDeviceWithoutAscendRuntime
func TestGetDeviceWithoutAscendRuntime(t *testing.T) {
	tool := mockAscendTools()
	convey.Convey("test getDeviceWithoutAscendRuntime", t, func() {
		convey.Convey("01-get spec failed, should return empty sets", func() {
			patch := gomonkey.ApplyFuncReturn(getContainerValidSpec, nil, errors.New("get spec failed"))
			defer patch.Reset()
			chips := tool.getDeviceWithoutAscendRuntime(MockContainer{}, nil)
			convey.So(chips, convey.ShouldResemble, sets.NewString())
		})
		convey.Convey("02-filter npu devices failed, should return empty sets", func() {
			patch := gomonkey.ApplyFuncReturn(getContainerValidSpec, &oci.Spec{}, nil)
			defer patch.Reset()
			chips := tool.getDeviceWithoutAscendRuntime(MockContainer{}, nil)
			convey.So(chips, convey.ShouldResemble, sets.NewString())
		})
		convey.Convey("03-filter npu devices failed, should return empty sets", func() {
			minor := int64(0)
			major := int64(3)
			spec := &oci.Spec{Linux: &specs.Linux{Resources: &specs.LinuxResources{
				Devices: []specs.LinuxDeviceCgroup{{Minor: &minor, Major: &major, Type: charDevice}}}},
			}
			patch := gomonkey.ApplyFuncReturn(getContainerValidSpec, spec, nil).
				ApplyFuncReturn(npuMajor, []string{"3"})
			defer patch.Reset()
			dev4 := fmt.Sprintf("%s-%d", common.Ascend910, 0)
			chips := tool.getDeviceWithoutAscendRuntime(MockContainer{}, nil)
			convey.So(chips, convey.ShouldResemble, sets.NewString(dev4))
		})
	})
}

// MockContainer mock container implements interface containerd.Container
type MockContainer struct{}

// ID identifies the container
func (m MockContainer) ID() string { return "mockContainer" }

// Info returns the underlying container record type
func (m MockContainer) Info(ctx context.Context, opts ...containerd.InfoOpts) (containers.Container, error) {
	return containers.Container{}, nil
}

// Extensions returns the extensions set on the container
func (m MockContainer) Extensions(ctx context.Context) (map[string]types.Any, error) {
	return map[string]types.Any{}, nil
}

// Labels returns the labels set on the container
func (m MockContainer) Labels(ctx context.Context) (map[string]string, error) {
	return map[string]string{}, nil
}

// SetLabels sets the provided labels for the container and returns the final label set
func (m MockContainer) SetLabels(ctx context.Context, labels map[string]string) (map[string]string, error) {
	return map[string]string{}, nil
}

// Spec returns the OCI runtime specification
func (m MockContainer) Spec(ctx context.Context) (*oci.Spec, error) {
	return &oci.Spec{}, nil
}

// Delete removes the container
func (m MockContainer) Delete(ctx context.Context, opts ...containerd.DeleteOpts) error {
	return nil
}

// Task returns the current task for the container
func (m MockContainer) Task(ctx context.Context, attach cio.Attach) (containerd.Task, error) {
	return nil, nil
}

// Image returns the image that the container is based on
func (m MockContainer) Image(ctx context.Context) (containerd.Image, error) {
	return nil, nil
}

// NewTask creates a new task based on the container metadata
func (m MockContainer) NewTask(ctx context.Context, ioCreate cio.Creator,
	opts ...containerd.NewTaskOpts) (containerd.Task, error) {
	return nil, nil
}

// Update a container
func (m MockContainer) Update(ctx context.Context, opts ...containerd.UpdateContainerOpts) error {
	return nil
}

// Checkpoint creates a checkpoint image of the current container
func (m MockContainer) Checkpoint(ctx context.Context, ref string,
	opts ...containerd.CheckpointOpts) (containerd.Image, error) {
	return nil, nil
}
