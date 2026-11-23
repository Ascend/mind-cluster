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

package containerruntime

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/containerd/containerd"
	apievents "github.com/containerd/containerd/api/events"
	"github.com/containerd/containerd/events"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/containerd/oci"
	"github.com/containerd/typeurl/v2"
	"github.com/smartystreets/goconvey/convey"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"Ascend-device-plugin/pkg/duplicatedetector/types"
)

func TestContainerdClient_ParseAllContainers01(t *testing.T) {
	convey.Convey("TestContainerdClient_ParseAllContainers", t, func() {
		ctx := context.Background()
		mockClient := &containerdClient{
			ociClient: &ociClient{client: &containerd.Client{}},
		}
		mns := &mockNamespaceService{}
		convey.Convey("01-list namespaces failed will return error", func() {
			patch := gomonkey.ApplyMethodReturn(mockClient.client, "NamespaceService",
				mns).ApplyMethodReturn(mns, "List", []string{}, errors.New("list namespaces failed"))
			defer patch.Reset()
			_, err := mockClient.ParseAllContainers(ctx)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldContainSubstring, "failed to list containers")
		})
		convey.Convey("02-list empty namespace will return empty info", func() {
			patch := gomonkey.ApplyMethodReturn(mockClient.client, "NamespaceService",
				mns).ApplyMethodReturn(mns, "List", []string{}, nil)
			defer patch.Reset()
			info, err := mockClient.ParseAllContainers(ctx)
			convey.So(err, convey.ShouldBeNil)
			convey.So(len(info), convey.ShouldEqual, 0)
		})
		convey.Convey("03-parse single container failed will continue", func() {
			patch := gomonkey.ApplyMethodReturn(mockClient.client, "NamespaceService",
				mns).ApplyMethodReturn(mns, "List", []string{"moby"}, nil).
				ApplyMethodReturn(mockClient.client, "Containers", nil, errors.New("list containers failed"))
			defer patch.Reset()
			info, err := mockClient.ParseAllContainers(ctx)
			convey.So(err, convey.ShouldBeNil)
			convey.So(len(info), convey.ShouldEqual, 0)
		})
	})
}

func TestContainerdClient_ParseAllContainers02(t *testing.T) {
	convey.Convey("TestContainerdClient_ParseAllContainers", t, func() {
		ctx := context.Background()
		mockClient := &containerdClient{
			ociClient: &ociClient{client: &containerd.Client{}},
		}
		mns := &mockNamespaceService{}
		convey.Convey("04-no containerd find will return empty container info", func() {
			patch := gomonkey.ApplyMethodReturn(mockClient.client, "NamespaceService",
				mns).ApplyMethodReturn(mns, "List", []string{"moby"}, nil).
				ApplyMethodReturn(mockClient.client, "Containers", []containerd.Container{}, nil)
			defer patch.Reset()
			info, err := mockClient.ParseAllContainers(ctx)
			convey.So(err, convey.ShouldBeNil)
			convey.So(len(info), convey.ShouldEqual, 0)
		})
		convey.Convey("05-parse failed will return empty container info", func() {
			patch := gomonkey.ApplyMethodReturn(mockClient.client, "NamespaceService", mns).
				ApplyMethodReturn(mns, "List", []string{"moby"}, nil).
				ApplyMethodReturn(mockClient.client, "Containers", []containerd.Container{&mockContainer{}}, nil).
				ApplyMethodReturn(mockClient.ociClient, "ParseSingleContainer", nil, errors.New("parse failed"))
			defer patch.Reset()
			info, err := mockClient.ParseAllContainers(ctx)
			convey.So(err, convey.ShouldBeNil)
			convey.So(len(info), convey.ShouldEqual, 0)
		})
		convey.Convey("06-success will return container info", func() {
			patch := gomonkey.ApplyMethodReturn(mockClient.client, "NamespaceService", mns).
				ApplyMethodReturn(mns, "List", []string{"moby"}, nil).
				ApplyMethodReturn(mockClient.client, "Containers", []containerd.Container{&mockContainer{}}, nil).
				ApplyMethodReturn(mockClient.ociClient, "ParseSingleContainer", &types.ContainerNPUInfo{}, nil)
			defer patch.Reset()
			info, err := mockClient.ParseAllContainers(ctx)
			convey.So(err, convey.ShouldBeNil)
			convey.So(len(info), convey.ShouldEqual, 1)
		})
	})
}

type mockContainer struct {
	containerd.Container
}

func (m *mockContainer) ID() string {
	return "test-container-id"
}

func (m *mockContainer) Spec(ctx context.Context) (*oci.Spec, error) {
	return &oci.Spec{}, nil
}

func (m *mockContainer) Labels(ctx context.Context) (map[string]string, error) {
	return nil, nil
}

func TestContainerdClient_WatchContainerEvents01(t *testing.T) {
	convey.Convey("TestContainerdClient_WatchContainerEvents", t, func() {
		mockClient := &containerdClient{
			ociClient: &ociClient{client: &containerd.Client{}},
		}
		convey.Convey("01-receive start event", func() {
			mes := &mockEventService{
				eventChan: make(chan *events.Envelope, 1),
				errChan:   make(chan error, 1),
			}
			patch := gomonkey.ApplyMethodReturn(mockClient.client, "EventService", mes)
			defer patch.Reset()
			startEvent := &apievents.TaskStart{
				ContainerID: "test-container-id",
			}
			any, err := typeurl.MarshalAny(startEvent)
			if err != nil {
				t.Fatal(err)
			}
			mes.eventChan <- &events.Envelope{
				Event:     any,
				Namespace: "test-ns",
			}
			var receivedEvent *types.ContainerEvent
			handler := func(event types.ContainerEvent) {
				receivedEvent = &event
			}
			ctx, cancel := context.WithCancel(context.Background())
			go func() {
				time.Sleep(timeout)
				cancel()
			}()
			mockClient.WatchContainerEvents(ctx, handler)
			convey.So(receivedEvent, convey.ShouldNotBeNil)
			convey.So(receivedEvent.Type, convey.ShouldEqual, types.ContainerEventCreate)
			convey.So(receivedEvent.ContainerID, convey.ShouldEqual, "test-container-id")
			convey.So(receivedEvent.Namespace, convey.ShouldEqual, "test-ns")
		})
	})
}

func TestContainerdClient_WatchContainerEvents02(t *testing.T) {
	convey.Convey("TestContainerdClient_WatchContainerEvents", t, func() {
		mockClient := &containerdClient{
			ociClient: &ociClient{client: &containerd.Client{}},
		}
		convey.Convey("02-receive exit event", func() {
			mes := &mockEventService{
				eventChan: make(chan *events.Envelope, 1),
				errChan:   make(chan error, 1),
			}
			patch := gomonkey.ApplyMethodReturn(mockClient.client, "EventService", mes)
			defer patch.Reset()
			deleteEvent := &apievents.TaskExit{
				ContainerID: "test-container-id",
				ID:          "test-container-id",
			}
			any, _ := typeurl.MarshalAny(deleteEvent)
			mes.eventChan <- &events.Envelope{
				Event:     any,
				Namespace: "test-ns",
			}
			var receivedEvent *types.ContainerEvent
			handler := func(event types.ContainerEvent) {
				receivedEvent = &event
			}
			ctx, cancel := context.WithCancel(context.Background())
			go func() {
				time.Sleep(timeout)
				cancel()
			}()
			mockClient.WatchContainerEvents(ctx, handler)
			convey.So(receivedEvent, convey.ShouldNotBeNil)
			convey.So(receivedEvent.Type, convey.ShouldEqual, types.ContainerEventDestroy)
			convey.So(receivedEvent.ContainerID, convey.ShouldEqual, "test-container-id")
			convey.So(receivedEvent.Namespace, convey.ShouldEqual, "test-ns")
		})
	})
}

func TestContainerdClient_WatchContainerEvents03(t *testing.T) {
	convey.Convey("TestContainerdClient_WatchContainerEvents", t, func() {
		mockClient := &containerdClient{
			ociClient: &ociClient{client: &containerd.Client{}},
		}
		convey.Convey("03-receive error", func() {
			mes := &mockEventService{
				eventChan: make(chan *events.Envelope, 1),
				errChan:   make(chan error, 1),
			}
			patch := gomonkey.ApplyMethodReturn(mockClient.client, "EventService", mes)
			defer patch.Reset()
			mes.errChan <- errors.New("event error")
			var receivedEvent *types.ContainerEvent
			handler := func(event types.ContainerEvent) {
				receivedEvent = &event
			}
			ctx, cancel := context.WithCancel(context.Background())
			go func() {
				time.Sleep(timeout)
				cancel()
			}()
			mockClient.WatchContainerEvents(ctx, handler)
			convey.So(receivedEvent, convey.ShouldBeNil)
		})
		convey.Convey("04-receive nil event", func() {
			mes := &mockEventService{
				eventChan: make(chan *events.Envelope, 1),
				errChan:   make(chan error, 1),
			}
			patch := gomonkey.ApplyMethodReturn(mockClient.client, "EventService", mes)
			defer patch.Reset()
			mes.eventChan <- &events.Envelope{
				Event:     nil,
				Namespace: "test-ns",
			}
			var receivedEvent *types.ContainerEvent
			handler := func(event types.ContainerEvent) {
				receivedEvent = &event
			}
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

type mockNamespaceService struct {
	namespaces.Store
}

func (m *mockNamespaceService) List(ctx context.Context) ([]string, error) {
	return []string{"default", "test-ns"}, nil
}

type mockEventService struct {
	eventChan chan *events.Envelope
	errChan   chan error
	containerd.EventService
}

func (m *mockEventService) Subscribe(ctx context.Context, filters ...string) (<-chan *events.Envelope, <-chan error) {
	return m.eventChan, m.errChan
}

func TestNewContainerdClient(t *testing.T) {
	convey.Convey("TestNewContainerdClient", t, func() {
		convey.Convey("01-empty cri endpoint will use oci endpoint", func() {
			patch := gomonkey.ApplyFuncReturn(checkSockFile, nil).
				ApplyFuncReturn(containerd.New, &containerd.Client{}, nil)
			defer patch.Reset()
			_, err := NewContainerdClient("", "/run/containerd/containerd.sock")
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("02-invalid oci endpoint will return error", func() {
			patch := gomonkey.ApplyFuncReturn(checkSockFile, errors.New("invalid socket"))
			defer patch.Reset()
			_, err := NewContainerdClient("", "/invalid/socket")
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("03-invalid cri endpoint will return error", func() {
			patch := gomonkey.ApplyFuncSeq(checkSockFile, []gomonkey.OutputCell{
				{Values: gomonkey.Params{nil}},
				{Values: gomonkey.Params{errors.New("invalid socket")}},
			})
			defer patch.Reset()
			_, err := NewContainerdClient("/invalid/socket", "/run/containerd/containerd.sock")
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("04-containerd.New failed will return error", func() {
			patch := gomonkey.ApplyFuncReturn(checkSockFile, nil).
				ApplyFuncReturn(containerd.New, nil, errors.New("create client failed"))
			defer patch.Reset()
			_, err := NewContainerdClient("/run/containerd/containerd.sock", "/run/containerd/containerd.sock")
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldContainSubstring, "create client failed")
		})
	})
}

func TestContainerdClient_ParseAllContainers03(t *testing.T) {
	convey.Convey("TestContainerdClient_ParseAllContainers", t, func() {
		ctx := context.Background()
		mockClient := &containerdClient{
			ociClient: &ociClient{client: &containerd.Client{}},
		}
		mns := &mockNamespaceService{}
		convey.Convey("07-not running container should be skipped", func() {
			patch := gomonkey.ApplyMethodReturn(mockClient.client, "NamespaceService", mns).
				ApplyMethodReturn(mns, "List", []string{"moby"}, nil).
				ApplyMethodReturn(mockClient.client, "Containers", []containerd.Container{&mockContainer{}}, nil).
				ApplyMethodReturn(mockClient.ociClient, "ParseSingleContainer", nil,
					status.Error(codes.NotFound, "task test-container-id not found"))
			defer patch.Reset()
			info, err := mockClient.ParseAllContainers(ctx)
			convey.So(err, convey.ShouldBeNil)
			convey.So(len(info), convey.ShouldEqual, 0)
		})
		convey.Convey("08-cancelled scan should stop and report interruption", func() {
			cancelCtx, cancel := context.WithCancel(ctx)
			cancel()
			patch := gomonkey.ApplyMethodReturn(mockClient.client, "NamespaceService", mns).
				ApplyMethodReturn(mns, "List", []string{"moby"}, nil).
				ApplyMethodReturn(mockClient.client, "Containers", []containerd.Container{&mockContainer{}}, nil).
				ApplyMethodReturn(mockClient.ociClient, "ParseSingleContainer", nil,
					status.Error(codes.Canceled, "context canceled"))
			defer patch.Reset()
			info, err := mockClient.ParseAllContainers(cancelCtx)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(info, convey.ShouldBeNil)
		})
		convey.Convey("09-grpc canceled error should interrupt even if ctx is alive", func() {
			patch := gomonkey.ApplyMethodReturn(mockClient.client, "NamespaceService", mns).
				ApplyMethodReturn(mns, "List", []string{"moby"}, nil).
				ApplyMethodReturn(mockClient.client, "Containers", []containerd.Container{&mockContainer{}}, nil).
				ApplyMethodReturn(mockClient.ociClient, "ParseSingleContainer", nil,
					status.Error(codes.Canceled, "rpc error: code = Canceled"))
			defer patch.Reset()
			info, err := mockClient.ParseAllContainers(ctx)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(info, convey.ShouldBeNil)
		})
	})
}

func TestIsScanCanceled(t *testing.T) {
	convey.Convey("test isScanCanceled", t, func() {
		convey.Convey("01-cancelled context should be treated as interrupted", func() {
			cancelCtx, cancel := context.WithCancel(context.Background())
			cancel()
			convey.So(isScanCanceled(cancelCtx, errors.New("any")), convey.ShouldBeTrue)
		})
		convey.Convey("02-bare context sentinel should be treated as interrupted", func() {
			convey.So(isScanCanceled(context.Background(), context.Canceled), convey.ShouldBeTrue)
			convey.So(isScanCanceled(context.Background(), context.DeadlineExceeded), convey.ShouldBeTrue)
		})
		convey.Convey("03-grpc status error should be treated as interrupted", func() {
			convey.So(isScanCanceled(context.Background(), status.Error(codes.Canceled, "x")),
				convey.ShouldBeTrue)
			convey.So(isScanCanceled(context.Background(),
				fmt.Errorf("wrap: %w", status.Error(codes.DeadlineExceeded, "x"))), convey.ShouldBeTrue)
		})
		convey.Convey("04-other errors should not be treated as interrupted", func() {
			convey.So(isScanCanceled(context.Background(), errors.New("boom")), convey.ShouldBeFalse)
			convey.So(isScanCanceled(context.Background(), status.Error(codes.NotFound, "x")),
				convey.ShouldBeFalse)
		})
	})
}
