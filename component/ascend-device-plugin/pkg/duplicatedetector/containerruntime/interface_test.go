/* Copyright(C) 2024. Huawei Technologies Co.,Ltd. All rights reserved.
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

package containerruntime

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/api/services/tasks/v1"
	"github.com/containerd/containerd/api/types/task"
	"github.com/smartystreets/goconvey/convey"

	"Ascend-device-plugin/pkg/duplicatedetector/types"
	"Ascend-device-plugin/pkg/kubeclient"
	"ascend-common/common-utils/hwlog"
)

func init() {
	hwLogConfig := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	hwlog.InitRunLogger(&hwLogConfig, context.Background())
}

const (
	timeout = 100 * time.Millisecond
)

func TestNewClient_NilConfig(t *testing.T) {
	_, err := NewClient(nil)
	if err == nil {
		t.Error("expected error for nil config")
	}
}

func TestNewClient_DockerRuntime(t *testing.T) {
	config := &types.DetectorConfig{
		CriEndpoint: defaultDockerAddress,
		RuntimeType: kubeclient.DockerRuntime,
	}
	patch := gomonkey.ApplyFuncReturn(autoDetectOciEndpoint, defaultContainerdAddr,
		nil).ApplyFuncReturn(checkSockFile,
		nil).ApplyFuncReturn(NewDockerClient,
		&dockerClient{}, nil)
	defer patch.Reset()
	_, err := NewClient(config)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNewClient_ContainerdRuntime(t *testing.T) {
	config := &types.DetectorConfig{
		CriEndpoint: "/run/containerd/containerd.sock",
		RuntimeType: kubeclient.ContainerdRuntime,
	}
	patch := gomonkey.ApplyFuncReturn(autoDetectOciEndpoint, defaultContainerdAddr,
		nil).ApplyFuncReturn(checkSockFile, nil).ApplyFuncReturn(NewContainerdClient,
		&containerdClient{}, nil)
	defer patch.Reset()
	_, err := NewClient(config)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNewClient_UnsupportedRuntime(t *testing.T) {
	config := &types.DetectorConfig{
		CriEndpoint: "/var/run/docker.sock",
		RuntimeType: "unsupported",
	}
	patch := gomonkey.ApplyFuncReturn(autoDetectOciEndpoint, defaultContainerdAddr,
		nil)
	defer patch.Reset()
	_, err := NewClient(config)
	if err == nil {
		t.Error("expected error for unsupported runtime")
	}
}

func TestAutoDetectOciEndpoint(t *testing.T) {
	patch := gomonkey.ApplyFuncReturn(os.Stat, nil, errors.New("file not found"))
	defer patch.Reset()
	endpoint, err := autoDetectOciEndpoint()
	if err == nil || endpoint != "" {
		t.Error("expected empty endpoint")
	}
}

func TestAutoDetectOciEndpoint_WithContainerd(t *testing.T) {
	patch := gomonkey.ApplyFuncReturn(os.Stat, nil, nil)
	defer patch.Reset()

	endpoint, err := autoDetectOciEndpoint()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if endpoint != defaultContainerdAddr {
		t.Errorf("expected default containerd endpoint")
	}
}

func TestNewDockerClient_EmptyEndpoint(t *testing.T) {
	_, err := NewDockerClient("", "")
	if err == nil {
		t.Error("expected error")
	}
}

func TestNewContainerdClient_EmptyEndpoint(t *testing.T) {
	_, err := NewContainerdClient("", "")
	if err == nil {
		t.Error("expected error")
	}
}

func TestNewContainerdClient_ValidEndpoint(t *testing.T) {
	patch := gomonkey.ApplyFuncReturn(checkSockFile, nil)
	patch.ApplyFuncReturn(containerd.New, &containerd.Client{}, nil)
	defer patch.Reset()
	_, err := NewContainerdClient("/run/containerd/containerd.sock", "")
	if err != nil {
		t.Errorf("expected error: %v", err)
	}
}

func TestParseSingleContainer01(t *testing.T) {
	convey.Convey("TestParseSingleContainer", t, func() {
		mockClient := &ociClient{
			client: &containerd.Client{},
		}
		mtc := &mockTasksClient{}
		mockResponse := &tasks.GetResponse{
			Process: &task.Process{},
		}
		containerID := "containerID"
		convey.Convey("01-get task failed will return error", func() {
			patch := gomonkey.ApplyMethodReturn(mockClient.client, "TaskService", mtc).
				ApplyMethodReturn(mtc, "Get", nil, errors.New("get task failed"))
			defer patch.Reset()
			_, err := mockClient.ParseSingleContainer(context.Background(), "")
			convey.So(err.Error(), convey.ShouldEqual, "get task failed")
		})
		convey.Convey("02-get task process failed will return error", func() {
			patch := gomonkey.ApplyMethodReturn(mockClient.client, "TaskService", mtc).
				ApplyMethodReturn(mtc, "Get", mockResponse, nil).
				ApplyMethodReturn(mockResponse, "GetProcess", nil)
			defer patch.Reset()

			_, err := mockClient.ParseSingleContainer(context.Background(), containerID)
			convey.So(err.Error(), convey.ShouldEqual, "task not found for container containerID")
		})
		convey.Convey("03-load container failed will return error", func() {
			patch := gomonkey.ApplyMethodReturn(mockClient.client, "TaskService", mtc).
				ApplyMethodReturn(mtc, "Get", mockResponse, nil).
				ApplyMethodReturn(mockClient.client, "LoadContainer", nil, errors.New("load container failed"))
			defer patch.Reset()
			_, err := mockClient.ParseSingleContainer(context.Background(), containerID)
			convey.So(err.Error(), convey.ShouldEqual, "load container failed")
		})
	})
}

func TestParseSingleContainer02(t *testing.T) {
	convey.Convey("TestParseSingleContainer", t, func() {
		mockClient := &ociClient{
			client: &containerd.Client{},
		}
		mtc := &mockTasksClient{}
		mockResponse := &tasks.GetResponse{
			Process: &task.Process{},
		}
		mc := &mockContainer{}
		containerID := "containerID"
		convey.Convey("04-get container spec failed will return error", func() {
			patch := gomonkey.ApplyMethodReturn(mockClient.client, "TaskService", mtc).
				ApplyMethodReturn(mtc, "Get", mockResponse, nil).
				ApplyMethodReturn(mockClient.client, "LoadContainer", mc, nil).
				ApplyMethodReturn(mc, "Spec", nil, errors.New("get container spec failed"))
			defer patch.Reset()
			_, err := mockClient.ParseSingleContainer(context.Background(), containerID)
			convey.So(err.Error(), convey.ShouldEqual, "failed to get container spec: get container spec failed")
		})
		convey.Convey("05-get container label failed will return error", func() {
			patch := gomonkey.ApplyMethodReturn(mockClient.client, "TaskService", mtc).
				ApplyMethodReturn(mtc, "Get", mockResponse, nil).
				ApplyMethodReturn(mockClient.client, "LoadContainer", mc, nil).
				ApplyMethodReturn(mc, "Labels", nil, errors.New("get container label failed"))
			defer patch.Reset()
			_, err := mockClient.ParseSingleContainer(context.Background(), containerID)
			convey.So(err.Error(), convey.ShouldEqual, "failed to get container labels: get container label failed")
		})
		convey.Convey("06-get container info success will return correct info", func() {
			patch := gomonkey.ApplyMethodReturn(mockClient.client, "TaskService", mtc).
				ApplyMethodReturn(mtc, "Get", mockResponse, nil).
				ApplyMethodReturn(mockClient.client, "LoadContainer", mc, nil)
			defer patch.Reset()
			info, err := mockClient.ParseSingleContainer(context.Background(), containerID)
			convey.So(err, convey.ShouldBeNil)
			convey.So(info.ID, convey.ShouldEqual, "test-container-id")
		})
	})
}

type mockTasksClient struct {
	tasks.TasksClient
}
