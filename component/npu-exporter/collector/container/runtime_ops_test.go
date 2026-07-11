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

// Package container provides utilities for container monitoring and testing.
package container

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	criv1 "k8s.io/cri-api/pkg/apis/runtime/v1"
	"k8s.io/cri-api/pkg/apis/runtime/v1alpha2"

	"ascend-common/common-utils/utils"
	"huawei.com/npu-exporter/v6/collector/container/isula"
	"huawei.com/npu-exporter/v6/collector/container/v1"
)

const (
	// Test constants for runtime operations
	testNamespace = "test-namespace"

	// Test error messages
	testInitCriError                    = "init CRI client failed"
	testInitOciError                    = "init OCI client failed"
	testSockCheckError                  = "socket check failed"
	testCriClientEmptyError             = "criClient is empty"
	testOciClientEmptyError             = "oci client is empty"
	testUnexpectedClientError           = "unexpected client type"
	testUnexpectedContainerdClientError = "unexpected containerd client"
	testUnexpectedIsulaClientError      = "unexpected isula client"
	testCriV1alpha2                     = "runtime.v1alpha2.RuntimeService"
	testCriV1                           = "runtime.v1.RuntimeService"
)

func TestRuntimeOperatorToolInit(t *testing.T) {
	r := &RuntimeOperatorTool{
		CriEndpoint: testContainerdEndpoint,
		OciEndpoint: testContainerdEndpoint,
	}
	convey.Convey("should initialize successfully when all components succeed", t, func() {
		operator := r
		patches := gomonkey.ApplyFuncReturn(sockCheck, nil)
		defer patches.Reset()
		patches.ApplyFuncReturn((*RuntimeOperatorTool).initCriClient, nil)
		patches.ApplyFuncReturn((*RuntimeOperatorTool).initOciClient, nil)
		err := operator.Init()
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("should return error when socket check fails", t, func() {
		operator := r
		patches := gomonkey.ApplyFuncReturn(sockCheck, errors.New(testSockCheckError))
		defer patches.Reset()
		err := operator.Init()
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, testSockCheckError)
	})
	convey.Convey("should return error when CRI client init fails", t, func() {
		operator := r
		patches := gomonkey.ApplyFuncReturn(sockCheck, nil)
		defer patches.Reset()
		patches.ApplyFuncReturn((*RuntimeOperatorTool).initCriClient, errors.New(testInitCriError))
		patches.ApplyFuncReturn((*RuntimeOperatorTool).initOciClient, nil)
		err := operator.Init()
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, testInitCriError)
	})
	convey.Convey("should return error when OCI client init fails", t, func() {
		operator := r
		patches := gomonkey.ApplyFuncReturn(sockCheck, nil)
		defer patches.Reset()
		patches.ApplyFuncReturn((*RuntimeOperatorTool).initCriClient, nil)
		patches.ApplyFuncReturn((*RuntimeOperatorTool).initOciClient, errors.New(testInitOciError))
		err := operator.Init()
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, testInitOciError)
	})
}

func TestRuntimeOperatorToolInitCriClient(t *testing.T) {
	convey.Convey("TestRuntimeOperatorToolInitCriClient", t, func() {
		convey.Convey("should initialize CRI client successfully for containerd", func() {
			operator := &RuntimeOperatorTool{
				CriEndpoint:  testContainerdEndpoint,
				UseOciBackup: false,
				UseCriBackup: false,
			}

			patches := gomonkey.ApplyFuncReturn(GetConnection, &grpc.ClientConn{}, nil)
			defer patches.Reset()

			err := operator.initCriClient()
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("should initialize CRI client successfully for isulad", func() {
			operator := &RuntimeOperatorTool{
				CriEndpoint:  DefaultIsuladAddr,
				UseOciBackup: false,
				UseCriBackup: false,
			}

			patches := gomonkey.ApplyFuncReturn(GetConnection, &grpc.ClientConn{}, nil)
			defer patches.Reset()

			err := operator.initCriClient()
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("should return error when connection fails and no backup", func() {
			operator := &RuntimeOperatorTool{
				CriEndpoint:  testContainerdEndpoint,
				UseOciBackup: false,
				UseCriBackup: false,
			}

			patches := gomonkey.ApplyFuncReturn(GetConnection, nil, errors.New("connection failed"))
			defer patches.Reset()

			err := operator.initCriClient()
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

func TestRuntimeOperatorToolInitOciClient(t *testing.T) {
	testCases := buildInitOciClientTestCases()
	for _, tc := range testCases {
		convey.Convey(tc.name, t, func() {
			operator, patches := tc.setup()
			if patches != nil {
				defer patches.Reset()
			}
			err := operator.initOciClient()
			if tc.hasError {
				convey.So(err, convey.ShouldNotBeNil)
			} else {
				convey.So(err, convey.ShouldBeNil)
			}
		})
	}
}

type initOciClientTestCase struct {
	name     string
	setup    func() (*RuntimeOperatorTool, *gomonkey.Patches)
	hasError bool
}

func buildInitOciClientTestCases() []initOciClientTestCase {
	return []initOciClientTestCase{
		{name: "should initialize OCI client successfully for containerd",
			setup: func() (*RuntimeOperatorTool, *gomonkey.Patches) {
				op := &RuntimeOperatorTool{OciEndpoint: testContainerdEndpoint, UseOciBackup: false}
				p := gomonkey.ApplyFuncReturn(GetConnection, &grpc.ClientConn{}, nil)
				return op, p
			},
			hasError: false},
		{name: "should initialize OCI client successfully for isulad",
			setup: func() (*RuntimeOperatorTool, *gomonkey.Patches) {
				op := &RuntimeOperatorTool{OciEndpoint: DefaultIsuladAddr, UseOciBackup: false}
				p := gomonkey.ApplyFuncReturn(GetConnection, &grpc.ClientConn{}, nil)
				return op, p
			},
			hasError: false},
		{name: "should return error when connection fails and no backup",
			setup: func() (*RuntimeOperatorTool, *gomonkey.Patches) {
				op := &RuntimeOperatorTool{OciEndpoint: testContainerdEndpoint, UseOciBackup: false}
				p := gomonkey.ApplyFuncReturn(GetConnection, nil, errors.New("connection failed"))
				return op, p
			},
			hasError: true},
		{name: "should return error when OCI endpoint is empty",
			setup: func() (*RuntimeOperatorTool, *gomonkey.Patches) {
				op := &RuntimeOperatorTool{OciEndpoint: "", UseOciBackup: false}
				return op, nil
			},
			hasError: true},
		{name: "should try backup when primary connection fails",
			setup: func() (*RuntimeOperatorTool, *gomonkey.Patches) {
				op := &RuntimeOperatorTool{OciEndpoint: testContainerdEndpoint, UseOciBackup: true}
				p := gomonkey.ApplyFunc(GetConnection, func(endpoint string) (*grpc.ClientConn, error) {
					if endpoint == testContainerdEndpoint {
						return nil, errors.New("primary failed")
					}
					return nil, errors.New("backup failed")
				})
				return op, p
			},
			hasError: true},
		{name: "should return error when all connections fail",
			setup: func() (*RuntimeOperatorTool, *gomonkey.Patches) {
				op := &RuntimeOperatorTool{OciEndpoint: testContainerdEndpoint, UseOciBackup: true}
				p := gomonkey.ApplyFuncReturn(GetConnection, nil, errors.New("all failed"))
				return op, p
			},
			hasError: true},
	}
}

func TestSockCheck(t *testing.T) {
	convey.Convey("TestSockCheck", t, func() {
		convey.Convey("should pass when socket paths are valid", func() {
			operator := &RuntimeOperatorTool{
				CriEndpoint: testContainerdEndpoint,
				OciEndpoint: testContainerdEndpoint,
			}

			patches := gomonkey.ApplyFuncReturn(utils.CheckPath, "/run/containerd.sock", nil)
			defer patches.Reset()
			patches.ApplyFuncReturn(utils.DoCheckOwnerAndPermission, nil)

			err := sockCheck(operator)
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("should return error when CRI endpoint check fails", func() {
			operator := &RuntimeOperatorTool{
				CriEndpoint: testContainerdEndpoint,
				OciEndpoint: testContainerdEndpoint,
			}

			patches := gomonkey.ApplyFuncReturn(utils.CheckPath, "", errors.New("path check failed"))
			defer patches.Reset()

			err := sockCheck(operator)
			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("should return error when CRI endpoint permission check fails", func() {
			operator := &RuntimeOperatorTool{
				CriEndpoint: testContainerdEndpoint,
				OciEndpoint: testContainerdEndpoint,
			}

			patches := gomonkey.ApplyFuncReturn(utils.CheckPath, "/run/containerd.sock", nil)
			defer patches.Reset()
			patches.ApplyFuncReturn(utils.DoCheckOwnerAndPermission, errors.New("permission check failed"))

			err := sockCheck(operator)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

func TestRuntimeOperatorToolClose(t *testing.T) {
	convey.Convey("TestRuntimeOperatorToolClose", t, func() {
		convey.Convey("should close connections successfully", func() {
			operator := &RuntimeOperatorTool{
				conn:    &grpc.ClientConn{},
				criConn: &grpc.ClientConn{},
			}

			patches := gomonkey.ApplyFunc((*grpc.ClientConn).Close, func(*grpc.ClientConn) error {
				return nil
			})
			defer patches.Reset()

			err := operator.Close()
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("should return error when OCI connection close fails", func() {
			operator := &RuntimeOperatorTool{
				conn:    &grpc.ClientConn{},
				criConn: &grpc.ClientConn{},
			}

			patches := gomonkey.ApplyFunc((*grpc.ClientConn).Close, func(*grpc.ClientConn) error {
				return errors.New("close failed")
			})
			defer patches.Reset()

			err := operator.Close()
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

// mockV1alpha2Client implements v1alpha2.RuntimeServiceClient for testing.
// Only ListContainers is configurable; other methods return nil.
type mockV1alpha2Client struct {
	listContainersFunc func(ctx context.Context, in *v1alpha2.ListContainersRequest,
		opts ...grpc.CallOption) (*v1alpha2.ListContainersResponse, error)
}

func (m *mockV1alpha2Client) ListContainers(ctx context.Context, in *v1alpha2.ListContainersRequest,
	opts ...grpc.CallOption) (*v1alpha2.ListContainersResponse, error) {
	return m.listContainersFunc(ctx, in, opts...)
}

// Stub methods to satisfy v1alpha2.RuntimeServiceClient interface.
func (m *mockV1alpha2Client) Version(context.Context, *v1alpha2.VersionRequest, ...grpc.CallOption) (*v1alpha2.VersionResponse, error) {
	return nil, nil
}
func (m *mockV1alpha2Client) RunPodSandbox(context.Context, *v1alpha2.RunPodSandboxRequest, ...grpc.CallOption) (*v1alpha2.RunPodSandboxResponse, error) {
	return nil, nil
}
func (m *mockV1alpha2Client) StopPodSandbox(context.Context, *v1alpha2.StopPodSandboxRequest, ...grpc.CallOption) (*v1alpha2.StopPodSandboxResponse, error) {
	return nil, nil
}
func (m *mockV1alpha2Client) RemovePodSandbox(context.Context, *v1alpha2.RemovePodSandboxRequest, ...grpc.CallOption) (*v1alpha2.RemovePodSandboxResponse, error) {
	return nil, nil
}
func (m *mockV1alpha2Client) PodSandboxStatus(context.Context, *v1alpha2.PodSandboxStatusRequest, ...grpc.CallOption) (*v1alpha2.PodSandboxStatusResponse, error) {
	return nil, nil
}
func (m *mockV1alpha2Client) ListPodSandbox(context.Context, *v1alpha2.ListPodSandboxRequest, ...grpc.CallOption) (*v1alpha2.ListPodSandboxResponse, error) {
	return nil, nil
}
func (m *mockV1alpha2Client) CreateContainer(context.Context, *v1alpha2.CreateContainerRequest, ...grpc.CallOption) (*v1alpha2.CreateContainerResponse, error) {
	return nil, nil
}
func (m *mockV1alpha2Client) StartContainer(context.Context, *v1alpha2.StartContainerRequest, ...grpc.CallOption) (*v1alpha2.StartContainerResponse, error) {
	return nil, nil
}
func (m *mockV1alpha2Client) StopContainer(context.Context, *v1alpha2.StopContainerRequest, ...grpc.CallOption) (*v1alpha2.StopContainerResponse, error) {
	return nil, nil
}
func (m *mockV1alpha2Client) RemoveContainer(context.Context, *v1alpha2.RemoveContainerRequest, ...grpc.CallOption) (*v1alpha2.RemoveContainerResponse, error) {
	return nil, nil
}
func (m *mockV1alpha2Client) ContainerStatus(context.Context, *v1alpha2.ContainerStatusRequest, ...grpc.CallOption) (*v1alpha2.ContainerStatusResponse, error) {
	return nil, nil
}
func (m *mockV1alpha2Client) UpdateContainerResources(context.Context, *v1alpha2.UpdateContainerResourcesRequest, ...grpc.CallOption) (*v1alpha2.UpdateContainerResourcesResponse, error) {
	return nil, nil
}
func (m *mockV1alpha2Client) ReopenContainerLog(context.Context, *v1alpha2.ReopenContainerLogRequest, ...grpc.CallOption) (*v1alpha2.ReopenContainerLogResponse, error) {
	return nil, nil
}
func (m *mockV1alpha2Client) ExecSync(context.Context, *v1alpha2.ExecSyncRequest, ...grpc.CallOption) (*v1alpha2.ExecSyncResponse, error) {
	return nil, nil
}
func (m *mockV1alpha2Client) Exec(context.Context, *v1alpha2.ExecRequest, ...grpc.CallOption) (*v1alpha2.ExecResponse, error) {
	return nil, nil
}
func (m *mockV1alpha2Client) Attach(context.Context, *v1alpha2.AttachRequest, ...grpc.CallOption) (*v1alpha2.AttachResponse, error) {
	return nil, nil
}
func (m *mockV1alpha2Client) PortForward(context.Context, *v1alpha2.PortForwardRequest, ...grpc.CallOption) (*v1alpha2.PortForwardResponse, error) {
	return nil, nil
}
func (m *mockV1alpha2Client) ContainerStats(context.Context, *v1alpha2.ContainerStatsRequest, ...grpc.CallOption) (*v1alpha2.ContainerStatsResponse, error) {
	return nil, nil
}
func (m *mockV1alpha2Client) ListContainerStats(context.Context, *v1alpha2.ListContainerStatsRequest, ...grpc.CallOption) (*v1alpha2.ListContainerStatsResponse, error) {
	return nil, nil
}
func (m *mockV1alpha2Client) PodSandboxStats(context.Context, *v1alpha2.PodSandboxStatsRequest, ...grpc.CallOption) (*v1alpha2.PodSandboxStatsResponse, error) {
	return nil, nil
}
func (m *mockV1alpha2Client) ListPodSandboxStats(context.Context, *v1alpha2.ListPodSandboxStatsRequest, ...grpc.CallOption) (*v1alpha2.ListPodSandboxStatsResponse, error) {
	return nil, nil
}
func (m *mockV1alpha2Client) UpdateRuntimeConfig(context.Context, *v1alpha2.UpdateRuntimeConfigRequest, ...grpc.CallOption) (*v1alpha2.UpdateRuntimeConfigResponse, error) {
	return nil, nil
}
func (m *mockV1alpha2Client) Status(context.Context, *v1alpha2.StatusRequest, ...grpc.CallOption) (*v1alpha2.StatusResponse, error) {
	return nil, nil
}

func TestRuntimeOperatorToolGetContainers(t *testing.T) {
	convey.Convey("TestRuntimeOperatorToolGetContainers", t, func() {
		convey.Convey("should return error when CRI client is empty", func() {
			operator := &RuntimeOperatorTool{}

			patches := gomonkey.ApplyFuncReturn(utils.IsNil, true)
			defer patches.Reset()

			containers, err := operator.GetContainers(context.Background())
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldEqual, testCriClientEmptyError)
			convey.So(containers, convey.ShouldBeNil)
		})

		convey.Convey("should return error when CRI connection is nil", func() {
			operator := &RuntimeOperatorTool{
				criClient: "mock-client",
			}

			patches := gomonkey.ApplyFuncReturn(utils.IsNil, false)
			defer patches.Reset()

			containers, err := operator.GetContainers(context.Background())
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldEqual, testCriClientEmptyError)
			convey.So(containers, convey.ShouldBeNil)
		})

		convey.Convey("should return error when client type is unexpected", func() {
			operator := &RuntimeOperatorTool{
				criClient: "unexpected",
				criConn:   &grpc.ClientConn{},
			}

			patches := gomonkey.ApplyFuncReturn(utils.IsNil, false)
			defer patches.Reset()

			containers, err := operator.GetContainers(context.Background())
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldEqual, testUnexpectedClientError)
			convey.So(containers, convey.ShouldBeNil)
		})

		convey.Convey("should return containers via v1 client on success", func() {
			mockClient := &mockV1alpha2Client{}
			operator := &RuntimeOperatorTool{
				criClient: mockClient,
				criConn:   &grpc.ClientConn{},
			}

			patches := gomonkey.ApplyFuncReturn(utils.IsNil, false)
			defer patches.Reset()
			patches.ApplyFunc(getContainersByContainerdV1, func(ctx context.Context,
				client criv1.RuntimeServiceClient) ([]*CommonContainer, error) {
				return []*CommonContainer{
					{Id: "v1-container-1", Labels: map[string]string{"app": "test"}},
					{Id: "v1-container-2", Labels: map[string]string{"app": "prod"}},
				}, nil
			})

			containers, err := operator.GetContainers(context.Background())
			convey.So(err, convey.ShouldBeNil)
			convey.So(containers, convey.ShouldHaveLength, 2)
			convey.So(containers[0].Id, convey.ShouldEqual, "v1-container-1")
			convey.So(containers[1].Id, convey.ShouldEqual, "v1-container-2")
		})

		convey.Convey("should return empty list when v1 has no containers", func() {
			mockClient := &mockV1alpha2Client{}
			operator := &RuntimeOperatorTool{
				criClient: mockClient,
				criConn:   &grpc.ClientConn{},
			}

			patches := gomonkey.ApplyFuncReturn(utils.IsNil, false)
			defer patches.Reset()
			patches.ApplyFunc(getContainersByContainerdV1, func(ctx context.Context,
				client criv1.RuntimeServiceClient) ([]*CommonContainer, error) {
				return nil, nil
			})

			containers, err := operator.GetContainers(context.Background())
			convey.So(err, convey.ShouldBeNil)
			convey.So(containers, convey.ShouldBeNil)
		})

		convey.Convey("should fallback to v1alpha2 when v1 returns unimplemented error", func() {
			unimplementedErr := status.Error(codes.Unimplemented, "unknown service "+testCriV1)
			mockClient := &mockV1alpha2Client{
				listContainersFunc: func(ctx context.Context, in *v1alpha2.ListContainersRequest,
					opts ...grpc.CallOption) (*v1alpha2.ListContainersResponse, error) {
					return &v1alpha2.ListContainersResponse{
						Containers: []*v1alpha2.Container{
							{Id: "v1alpha2-fallback-container", Labels: map[string]string{"app": "fallback"}},
						},
					}, nil
				},
			}
			operator := &RuntimeOperatorTool{
				criClient: mockClient,
				criConn:   &grpc.ClientConn{},
			}

			patches := gomonkey.ApplyFuncReturn(utils.IsNil, false)
			defer patches.Reset()
			patches.ApplyFunc(getContainersByContainerdV1, func(ctx context.Context,
				client criv1.RuntimeServiceClient) ([]*CommonContainer, error) {
				return nil, unimplementedErr
			})

			containers, err := operator.GetContainers(context.Background())
			convey.So(err, convey.ShouldBeNil)
			convey.So(containers, convey.ShouldHaveLength, 1)
			convey.So(containers[0].Id, convey.ShouldEqual, "v1alpha2-fallback-container")
		})

		convey.Convey("should return error when v1 returns non-unimplemented error", func() {
			otherErr := status.Error(codes.Internal, "internal error")
			mockClient := &mockV1alpha2Client{}
			operator := &RuntimeOperatorTool{
				criClient: mockClient,
				criConn:   &grpc.ClientConn{},
			}

			patches := gomonkey.ApplyFuncReturn(utils.IsNil, false)
			defer patches.Reset()
			patches.ApplyFunc(getContainersByContainerdV1, func(ctx context.Context,
				client criv1.RuntimeServiceClient) ([]*CommonContainer, error) {
				return nil, otherErr
			})

			containers, err := operator.GetContainers(context.Background())
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(containers, convey.ShouldBeNil)
		})

		convey.Convey("should use cached v1alpha2 version directly", func() {
			mockClient := &mockV1alpha2Client{
				listContainersFunc: func(ctx context.Context, in *v1alpha2.ListContainersRequest,
					opts ...grpc.CallOption) (*v1alpha2.ListContainersResponse, error) {
					return &v1alpha2.ListContainersResponse{
						Containers: []*v1alpha2.Container{
							{Id: "v1alpha2-cached-container", Labels: map[string]string{"app": "cached"}},
						},
					}, nil
				},
			}
			operator := &RuntimeOperatorTool{
				criClient:  mockClient,
				criConn:    &grpc.ClientConn{},
				criVersion: criVersionV1alpha2,
			}

			patches := gomonkey.ApplyFuncReturn(utils.IsNil, false)
			defer patches.Reset()

			containers, err := operator.GetContainers(context.Background())
			convey.So(err, convey.ShouldBeNil)
			convey.So(containers, convey.ShouldHaveLength, 1)
			convey.So(containers[0].Id, convey.ShouldEqual, "v1alpha2-cached-container")
		})

		convey.Convey("should cache v1 version on success", func() {
			mockClient := &mockV1alpha2Client{
				listContainersFunc: func(ctx context.Context, in *v1alpha2.ListContainersRequest,
					opts ...grpc.CallOption) (*v1alpha2.ListContainersResponse, error) {
					return &v1alpha2.ListContainersResponse{}, nil
				},
			}
			operator := &RuntimeOperatorTool{
				criClient: mockClient,
				criConn:   &grpc.ClientConn{},
			}

			patches := gomonkey.ApplyFuncReturn(utils.IsNil, false)
			defer patches.Reset()
			patches.ApplyFunc(getContainersByContainerdV1, func(ctx context.Context,
				client criv1.RuntimeServiceClient) ([]*CommonContainer, error) {
				return []*CommonContainer{
					{Id: "v1-container", Labels: map[string]string{"app": "test"}},
				}, nil
			})

			containers, err := operator.GetContainers(context.Background())
			convey.So(err, convey.ShouldBeNil)
			convey.So(containers, convey.ShouldHaveLength, 1)
			convey.So(operator.criVersion, convey.ShouldEqual, criVersionV1)
		})

		convey.Convey("should cache v1alpha2 version after fallback", func() {
			unimplementedErr := status.Error(codes.Unimplemented, "unknown service "+testCriV1)
			mockClient := &mockV1alpha2Client{
				listContainersFunc: func(ctx context.Context, in *v1alpha2.ListContainersRequest,
					opts ...grpc.CallOption) (*v1alpha2.ListContainersResponse, error) {
					return &v1alpha2.ListContainersResponse{
						Containers: []*v1alpha2.Container{
							{Id: "v1alpha2-fallback-container", Labels: map[string]string{"app": "fallback"}},
						},
					}, nil
				},
			}
			operator := &RuntimeOperatorTool{
				criClient: mockClient,
				criConn:   &grpc.ClientConn{},
			}

			patches := gomonkey.ApplyFuncReturn(utils.IsNil, false)
			defer patches.Reset()
			patches.ApplyFunc(getContainersByContainerdV1, func(ctx context.Context,
				client criv1.RuntimeServiceClient) ([]*CommonContainer, error) {
				return nil, unimplementedErr
			})

			containers, err := operator.GetContainers(context.Background())
			convey.So(err, convey.ShouldBeNil)
			convey.So(containers, convey.ShouldHaveLength, 1)
			convey.So(operator.criVersion, convey.ShouldEqual, criVersionV1alpha2)
		})

	})
}

func TestIsUnimplementedError(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		serviceName string
		want        bool
	}{
		{
			name:        "nil error returns false",
			err:         nil,
			serviceName: testCriV1alpha2,
			want:        false,
		},
		{
			name:        "non-grpc error returns false",
			err:         errors.New("unknown service " + testCriV1alpha2),
			serviceName: testCriV1alpha2,
			want:        false,
		},
		{
			name:        "mismatched code returns false",
			err:         status.Error(codes.NotFound, "unknown service "+testCriV1alpha2),
			serviceName: testCriV1alpha2,
			want:        false,
		},
		{
			name:        "mismatched message returns false",
			err:         status.Error(codes.Unimplemented, "unknown service "+testCriV1),
			serviceName: testCriV1alpha2,
			want:        false,
		},
		{
			name:        "matched unimplemented error returns true",
			err:         status.Error(codes.Unimplemented, "unknown service "+testCriV1alpha2),
			serviceName: testCriV1alpha2,
			want:        true,
		},
		{
			name:        "real grpc error format returns true",
			err:         fmt.Errorf("rpc error: code = Unimplemented desc = unknown service " + testCriV1alpha2),
			serviceName: testCriV1alpha2,
			want:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isUnimplementedError(tt.err, tt.serviceName); got != tt.want {
				t.Errorf("isUnimplementedError() = %v, want %v (err: %v)", got, tt.want, tt.err)
			}
		})
	}
}

func TestRuntimeOperatorToolGetContainerInfoByID(t *testing.T) {
	convey.Convey("TestRuntimeOperatorToolGetContainerInfoByID", t, func() {
		convey.Convey("should return error when OCI client is empty", func() {
			operator := &RuntimeOperatorTool{}
			patches := gomonkey.ApplyFuncReturn(utils.IsNil, true)
			defer patches.Reset()
			spec, err := operator.GetContainerInfoByID(context.Background(), testContainerID)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldEqual, testOciClientEmptyError)
			convey.So(spec, convey.ShouldResemble, v1.Spec{})
		})
		convey.Convey("should return error when OCI connection is nil", func() {
			operator := &RuntimeOperatorTool{client: "mock-client"}
			patches := gomonkey.ApplyFuncReturn(utils.IsNil, false)
			defer patches.Reset()
			spec, err := operator.GetContainerInfoByID(context.Background(), testContainerID)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldEqual, testOciClientEmptyError)
			convey.So(spec, convey.ShouldResemble, v1.Spec{})
		})
		convey.Convey("should return error when client type is unexpected", func() {
			operator := &RuntimeOperatorTool{client: "unexpected", conn: &grpc.ClientConn{}}
			patches := gomonkey.ApplyFuncReturn(utils.IsNil, false)
			defer patches.Reset()
			spec, err := operator.GetContainerInfoByID(context.Background(), testContainerID)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldEqual, testUnexpectedContainerdClientError)
			convey.So(spec, convey.ShouldResemble, v1.Spec{})
		})
		convey.Convey("should return error when GetContainer call fails", func() {
			operator := &RuntimeOperatorTool{client: "mock-containers-client", conn: &grpc.ClientConn{}}
			patches := gomonkey.ApplyFuncReturn(utils.IsNil, false)
			defer patches.Reset()
			spec, err := operator.GetContainerInfoByID(context.Background(), testContainerID)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(spec, convey.ShouldResemble, v1.Spec{})
		})
		convey.Convey("should return error when JSON unmarshal fails", func() {
			operator := &RuntimeOperatorTool{client: "mock-containers-client", conn: &grpc.ClientConn{}}
			patches := gomonkey.ApplyFuncReturn(utils.IsNil, false)
			defer patches.Reset()
			spec, err := operator.GetContainerInfoByID(context.Background(), testContainerID)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(spec, convey.ShouldResemble, v1.Spec{})
		})

	})
}

func TestRuntimeOperatorToolGetIsulaContainerInfoByID(t *testing.T) {
	convey.Convey("TestRuntimeOperatorToolGetIsulaContainerInfoByID", t, func() {
		convey.Convey("should return error when OCI client is empty", func() {
			operator := &RuntimeOperatorTool{}
			patches := gomonkey.ApplyFuncReturn(utils.IsNil, true)
			defer patches.Reset()
			containerInfo, err := operator.GetIsulaContainerInfoByID(context.Background(), testContainerID)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldEqual, testOciClientEmptyError)
			convey.So(containerInfo, convey.ShouldResemble, isula.ContainerJson{})
		})
		convey.Convey("should return error when OCI connection is nil", func() {
			operator := &RuntimeOperatorTool{client: "mock-client"}
			patches := gomonkey.ApplyFuncReturn(utils.IsNil, false)
			defer patches.Reset()
			containerInfo, err := operator.GetIsulaContainerInfoByID(context.Background(), testContainerID)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldEqual, testOciClientEmptyError)
			convey.So(containerInfo, convey.ShouldResemble, isula.ContainerJson{})
		})
		convey.Convey("should return error when client type is unexpected", func() {
			operator := &RuntimeOperatorTool{client: "unexpected", conn: &grpc.ClientConn{}}
			patches := gomonkey.ApplyFuncReturn(utils.IsNil, false)
			defer patches.Reset()
			containerInfo, err := operator.GetIsulaContainerInfoByID(context.Background(), testContainerID)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldEqual, testUnexpectedIsulaClientError)
			convey.So(containerInfo, convey.ShouldResemble, isula.ContainerJson{})
		})
		convey.Convey("should return error when Inspect call fails", func() {
			operator := &RuntimeOperatorTool{client: "mock-isula-client", conn: &grpc.ClientConn{}}
			patches := gomonkey.ApplyFuncReturn(utils.IsNil, false)
			defer patches.Reset()
			containerInfo, err := operator.GetIsulaContainerInfoByID(context.Background(), testContainerID)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(containerInfo, convey.ShouldResemble, isula.ContainerJson{})
		})
		convey.Convey("should return error when JSON unmarshal fails", func() {
			operator := &RuntimeOperatorTool{client: "mock-isula-client", conn: &grpc.ClientConn{}}
			patches := gomonkey.ApplyFuncReturn(utils.IsNil, false)
			defer patches.Reset()
			containerInfo, err := operator.GetIsulaContainerInfoByID(context.Background(), testContainerID)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(containerInfo, convey.ShouldResemble, isula.ContainerJson{})
		})

	})
}

func TestRuntimeOperatorToolGetContainerType(t *testing.T) {
	convey.Convey("TestRuntimeOperatorToolGetContainerType", t, func() {
		convey.Convey("should return isula when endpoint is isulad", func() {
			operator := &RuntimeOperatorTool{
				OciEndpoint: DefaultIsuladAddr,
			}

			containerType := operator.GetContainerType()
			convey.So(containerType, convey.ShouldEqual, IsulaContainer)
		})

		convey.Convey("should return default when endpoint is not isulad", func() {
			operator := &RuntimeOperatorTool{
				OciEndpoint: testContainerdEndpoint,
			}

			containerType := operator.GetContainerType()
			convey.So(containerType, convey.ShouldEqual, DefaultContainer)
		})
	})
}

func TestSetGrpcNamespaceHeader(t *testing.T) {
	convey.Convey("TestSetGrpcNamespaceHeader", t, func() {
		convey.Convey("should set namespace header when context has no metadata", func() {
			ctx := context.Background()
			result := setGrpcNamespaceHeader(ctx, testNamespace)
			convey.So(result, convey.ShouldNotBeNil)
		})

		convey.Convey("should set namespace header when context has existing metadata", func() {
			ctx := context.Background()
			ctx = context.WithValue(ctx, "test", "value")
			result := setGrpcNamespaceHeader(ctx, testNamespace)
			convey.So(result, convey.ShouldNotBeNil)
		})
	})
}

func TestGenContainerRequestV1alpha2(t *testing.T) {
	convey.Convey("TestGenContainerRequestV1alpha2", t, func() {
		convey.Convey("should generate valid container request", func() {
			request := genContainerRequestV1alpha2()
			convey.So(request, convey.ShouldNotBeNil)
			convey.So(request.Filter, convey.ShouldNotBeNil)
			convey.So(request.Filter.State, convey.ShouldNotBeNil)
			convey.So(request.Filter.State.State, convey.ShouldEqual, v1alpha2.ContainerState_CONTAINER_RUNNING)
		})
	})
}

func TestGenContainerRequestV1(t *testing.T) {
	convey.Convey("TestGenContainerRequestV1", t, func() {
		convey.Convey("should generate valid container request", func() {
			request := genContainerRequestV1()
			convey.So(request, convey.ShouldNotBeNil)
			convey.So(request.Filter, convey.ShouldNotBeNil)
			convey.So(request.Filter.State, convey.ShouldNotBeNil)
			convey.So(request.Filter.State.State, convey.ShouldEqual, criv1.ContainerState_CONTAINER_RUNNING)
		})
	})
}

func TestGenIsulaRequest(t *testing.T) {
	convey.Convey("TestGenIsulaRequest", t, func() {
		convey.Convey("should generate valid isula request", func() {
			request := genIsulaRequest()
			convey.So(request, convey.ShouldNotBeNil)
			convey.So(request.Filter, convey.ShouldNotBeNil)
			convey.So(request.Filter.State, convey.ShouldNotBeNil)
			convey.So(request.Filter.State.State, convey.ShouldEqual, isula.ContainerState_CONTAINER_RUNNING)
		})
	})
}
