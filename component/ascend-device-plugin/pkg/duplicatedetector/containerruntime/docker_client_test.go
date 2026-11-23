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

// Package containerruntime is the client for interacting with docker and containerd runtime

package containerruntime

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/containerd/containerd"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	dockererrdefs "github.com/docker/docker/errdefs"
	"github.com/smartystreets/goconvey/convey"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	dtypes "Ascend-device-plugin/pkg/duplicatedetector/types"
	"github.com/docker/docker/api/types/events"
)

func TestDockerClient_ParseAllContainers(t *testing.T) {
	convey.Convey("TestDockerClient_ParseAllContainers", t, func() {
		ctx := context.Background()
		mockClient := &dockerClient{
			client:    &client.Client{},
			ociClient: &ociClient{client: &containerd.Client{}},
		}
		convey.Convey("01-list containers failed will return error", func() {
			patch := gomonkey.ApplyMethodReturn(mockClient.client, "ContainerList", nil, errors.New("list containers failed"))
			defer patch.Reset()
			_, err := mockClient.ParseAllContainers(ctx)
			convey.So(err.Error(), convey.ShouldEqual, "failed to list containers: list containers failed")
		})
		convey.Convey("02-ParseSingleContainer failed will return empty container info", func() {
			patch := gomonkey.ApplyMethodReturn(mockClient.client, "ContainerList", []types.Container{
				{
					ID: "test-container-id",
				},
			}, nil).ApplyMethodReturn(mockClient, "ParseSingleContainer", nil, errors.New("parse single container failed"))
			defer patch.Reset()
			info, err := mockClient.ParseAllContainers(ctx)
			convey.So(err, convey.ShouldBeNil)
			convey.So(len(info), convey.ShouldEqual, 0)
		})
		convey.Convey("03-ParseSingleContainer success will return container info", func() {
			patch := gomonkey.ApplyMethodReturn(mockClient.client, "ContainerList", []types.Container{
				{
					ID: "test-container-id",
				},
			}, nil).ApplyMethodReturn(mockClient, "ParseSingleContainer", &dtypes.ContainerNPUInfo{}, nil)
			defer patch.Reset()
			info, err := mockClient.ParseAllContainers(ctx)
			convey.So(err, convey.ShouldBeNil)
			convey.So(len(info), convey.ShouldEqual, 1)

		})
	})
}

func TestDockerClient_ParseSingleContainer(t *testing.T) {
	convey.Convey("TestDockerClient_ParseSingleContainer", t, func() {
		ctx := context.Background()
		mockClient := &dockerClient{
			client:    &client.Client{},
			ociClient: &ociClient{client: &containerd.Client{}},
		}
		convey.Convey("01-ContainerInspect failed will return error", func() {
			patch := gomonkey.ApplyMethodReturn(mockClient.client, "ContainerInspect", nil, errors.New("inspect container failed"))
			defer patch.Reset()
			_, err := mockClient.ParseSingleContainer(ctx, "test-container-id")
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldContainSubstring, "inspect container failed")
		})
		convey.Convey("02-ParseSingleContainer failed will return error", func() {
			containerJSON := types.ContainerJSON{Config: &container.Config{}}
			patch := gomonkey.ApplyMethodReturn(mockClient.client, "ContainerInspect", containerJSON, nil).
				ApplyMethodReturn(mockClient.ociClient, "ParseSingleContainer", nil,
					errors.New("parse single container failed"))
			defer patch.Reset()
			_, err := mockClient.ParseSingleContainer(ctx, "test-container-id")
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldContainSubstring, "parse single container failed")
		})
		convey.Convey("03-success will return container info", func() {
			containerJSON := types.ContainerJSON{
				Config: &container.Config{
					Labels: map[string]string{
						"io.kubernetes.pod.name":       "test-pod",
						"io.kubernetes.pod.namespace":  "test-ns",
						"io.kubernetes.container.name": "test-container",
					},
				},
			}
			patch := gomonkey.ApplyMethodReturn(mockClient.client, "ContainerInspect", containerJSON, nil).
				ApplyMethodReturn(mockClient.ociClient, "ParseSingleContainer",
					&dtypes.ContainerNPUInfo{ID: "test-container-id"}, nil)
			defer patch.Reset()
			info, err := mockClient.ParseSingleContainer(ctx, "test-container-id")
			convey.So(err, convey.ShouldBeNil)
			convey.So(info.PodName, convey.ShouldEqual, "test-pod")
			convey.So(info.PodNS, convey.ShouldEqual, "test-ns")
			convey.So(info.Name, convey.ShouldEqual, "test-container")
		})
	})
}

func TestDockerClient_WatchContainerEvents01(t *testing.T) {
	convey.Convey("TestDockerClient_WatchContainerEvents", t, func() {
		mockClient := &dockerClient{
			client:    &client.Client{},
			ociClient: &ociClient{client: &containerd.Client{}},
		}
		convey.Convey("01-receive start event", func() {
			eventChan := make(chan events.Message, 1)
			errChan := make(chan error, 1)
			eventChan <- events.Message{
				Action: "start",
				Actor: events.Actor{
					ID: "test-container-id",
				},
			}
			var receivedEvent *dtypes.ContainerEvent
			handler := func(event dtypes.ContainerEvent) {
				receivedEvent = &event
			}
			patch := gomonkey.ApplyMethod(mockClient.client, "Events",
				func(_ *client.Client, _ context.Context, _ types.EventsOptions) (<-chan events.Message, <-chan error) {
					return eventChan, errChan
				}).ApplyMethod(mockClient.client, "Close",
				func(_ *client.Client) error {
					return nil
				})
			defer patch.Reset()
			ctx, cancel := context.WithCancel(context.Background())
			go func() {
				time.Sleep(timeout)
				cancel()
			}()
			mockClient.WatchContainerEvents(ctx, handler)
			convey.So(receivedEvent, convey.ShouldNotBeNil)
			convey.So(receivedEvent.Type, convey.ShouldEqual, dtypes.ContainerEventCreate)
			convey.So(receivedEvent.ContainerID, convey.ShouldEqual, "test-container-id")
			convey.So(receivedEvent.Namespace, convey.ShouldEqual, dockerNamespace)
		})
	})
}

func TestDockerClient_WatchContainerEvents02(t *testing.T) {
	convey.Convey("TestDockerClient_WatchContainerEvents", t, func() {
		mockClient := &dockerClient{
			client:    &client.Client{},
			ociClient: &ociClient{client: &containerd.Client{}},
		}
		convey.Convey("03-receive error", func() {
			eventChan := make(chan events.Message, 1)
			errChan := make(chan error, 1)
			errChan <- errors.New("event error")
			var receivedEvent *dtypes.ContainerEvent
			handler := func(event dtypes.ContainerEvent) {
				receivedEvent = &event
			}
			patch := gomonkey.ApplyMethod(mockClient.client, "Events",
				func(_ *client.Client, _ context.Context, _ types.EventsOptions) (<-chan events.Message, <-chan error) {
					return eventChan, errChan
				}).ApplyMethod(mockClient.client, "Close",
				func(_ *client.Client) error {
					return nil
				})
			defer patch.Reset()
			ctx, cancel := context.WithCancel(context.Background())
			go func() {
				time.Sleep(timeout)
				cancel()
			}()
			mockClient.WatchContainerEvents(ctx, handler)
			convey.So(receivedEvent, convey.ShouldBeNil)
		})
	})
}

func TestDockerClient_WatchContainerEvents03(t *testing.T) {
	convey.Convey("TestDockerClient_WatchContainerEvents", t, func() {
		mockClient := &dockerClient{
			client:    &client.Client{},
			ociClient: &ociClient{client: &containerd.Client{}},
		}
		convey.Convey("03-receive error", func() {
			eventChan := make(chan events.Message, 1)
			errChan := make(chan error, 1)
			errChan <- errors.New("event error")
			var receivedEvent *dtypes.ContainerEvent
			handler := func(event dtypes.ContainerEvent) {
				receivedEvent = &event
			}
			patch := gomonkey.ApplyMethod(mockClient.client, "Events",
				func(_ *client.Client, _ context.Context, _ types.EventsOptions) (<-chan events.Message, <-chan error) {
					return eventChan, errChan
				}).ApplyMethod(mockClient.client, "Close",
				func(_ *client.Client) error {
					return nil
				})
			defer patch.Reset()
			ctx, cancel := context.WithCancel(context.Background())
			go func() {
				time.Sleep(timeout)
				cancel()
			}()
			mockClient.WatchContainerEvents(ctx, handler)
			convey.So(receivedEvent, convey.ShouldBeNil)
		})
	})
}

func TestNewDockerClient(t *testing.T) {
	convey.Convey("TestNewDockerClient", t, func() {
		convey.Convey("01-empty cri endpoint will use default", func() {
			patch := gomonkey.ApplyFuncReturn(checkSockFile, nil).
				ApplyFuncReturn(client.NewClientWithOpts, &client.Client{}, nil).
				ApplyFuncReturn(containerd.New, &containerd.Client{}, nil)
			defer patch.Reset()
			_, err := NewDockerClient("", "/run/containerd/containerd.sock")
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("02-invalid cri endpoint will return error", func() {
			patch := gomonkey.ApplyFuncReturn(checkSockFile, errors.New("invalid socket"))
			defer patch.Reset()
			_, err := NewDockerClient("/invalid/socket", "/run/containerd/containerd.sock")
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldContainSubstring, "invalid cri endpoint")
		})
		convey.Convey("03-invalid oci endpoint will return error", func() {
			patch := gomonkey.ApplyFuncSeq(checkSockFile, []gomonkey.OutputCell{
				{Values: gomonkey.Params{nil}},
				{Values: gomonkey.Params{errors.New("invalid socket")}},
			})
			defer patch.Reset()
			_, err := NewDockerClient("/run/docker.sock", "/invalid/socket")
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldContainSubstring, "invalid oci endpoint")
		})
	})
}

func TestDockerClient_ParseSingleContainerNotRunning(t *testing.T) {
	convey.Convey("TestDockerClient_ParseSingleContainer", t, func() {
		ctx := context.Background()
		mockClient := &dockerClient{
			client:    &client.Client{},
			ociClient: &ociClient{client: &containerd.Client{}},
		}
		convey.Convey("04-not running container keeps the original error for callers", func() {
			containerJSON := types.ContainerJSON{Config: &container.Config{}}
			patch := gomonkey.ApplyMethodReturn(mockClient.client, "ContainerInspect", containerJSON, nil).
				ApplyMethodReturn(mockClient.ociClient, "ParseSingleContainer", nil,
					status.Error(codes.NotFound, "task test-container-id not found"))
			defer patch.Reset()
			_, err := mockClient.ParseSingleContainer(ctx, "test-container-id")
			convey.So(IsNotFound(err), convey.ShouldBeTrue)
			convey.So(err.Error(), convey.ShouldContainSubstring, "task test-container-id not found")
		})
		convey.Convey("05-inspect not found should still be classified after wrapping", func() {
			patch := gomonkey.ApplyMethodReturn(mockClient.client, "ContainerInspect", nil,
				dockererrdefs.NotFound(errors.New("no such container")))
			defer patch.Reset()
			_, err := mockClient.ParseSingleContainer(ctx, "test-container-id")
			convey.So(IsNotFound(err), convey.ShouldBeTrue)
			convey.So(err.Error(), convey.ShouldContainSubstring, "no such container")
		})
	})
}
