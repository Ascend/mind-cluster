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
	"math"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/containers"
	"github.com/containerd/containerd/oci"
	"github.com/containerd/typeurl/v2"
	specs "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/api"
	"container-manager/pkg/common"
	"container-manager/pkg/container/domain"
)

// mockContainer implements containerd.Container.
type mockContainer struct {
	id        string
	labels    map[string]string
	labelsErr error
	spec      *oci.Spec
	specErr   error
}

func (m *mockContainer) ID() string { return m.id }
func (m *mockContainer) Info(context.Context, ...containerd.InfoOpts) (containers.Container, error) {
	return containers.Container{}, nil
}
func (m *mockContainer) Delete(context.Context, ...containerd.DeleteOpts) error { return nil }
func (m *mockContainer) NewTask(context.Context, cio.Creator, ...containerd.NewTaskOpts) (containerd.Task, error) {
	return nil, nil
}
func (m *mockContainer) Spec(context.Context) (*oci.Spec, error) { return m.spec, m.specErr }
func (m *mockContainer) Task(context.Context, cio.Attach) (containerd.Task, error) {
	return nil, nil
}
func (m *mockContainer) Image(context.Context) (containerd.Image, error) { return nil, nil }
func (m *mockContainer) Labels(context.Context) (map[string]string, error) {
	return m.labels, m.labelsErr
}
func (m *mockContainer) SetLabels(context.Context, map[string]string) (map[string]string, error) {
	return nil, nil
}
func (m *mockContainer) Extensions(context.Context) (map[string]typeurl.Any, error)      { return nil, nil }
func (m *mockContainer) Update(context.Context, ...containerd.UpdateContainerOpts) error { return nil }
func (m *mockContainer) Checkpoint(context.Context, string, ...containerd.CheckpointOpts) (containerd.Image, error) {
	return nil, nil
}

var _ containerd.Container = (*mockContainer)(nil)

func TestNewContainerdClient(t *testing.T) {
	convey.Convey("test new containerd client", t, func() {
		c := NewContainerdClient()
		convey.So(c, convey.ShouldNotBeNil)
		convey.So(c.stoppedContainer, convey.ShouldNotBeNil)
	})
}

func TestContainerdClientInit(t *testing.T) {
	convey.Convey("test containerd init", t, func() {
		convey.Convey("connect failed", func() {
			patch := gomonkey.ApplyFunc(containerd.New, func(string, ...containerd.ClientOpt) (*containerd.Client, error) {
				return nil, errors.New("connect failed")
			})
			defer patch.Reset()
			c := NewContainerdClient()
			convey.So(c.init(), convey.ShouldNotBeNil)
		})
		convey.Convey("connect success", func() {
			patch := gomonkey.ApplyFunc(containerd.New, func(string, ...containerd.ClientOpt) (*containerd.Client, error) {
				return &containerd.Client{}, nil
			})
			defer patch.Reset()
			c := NewContainerdClient()
			convey.So(c.init(), convey.ShouldBeNil)
			convey.So(c.client, convey.ShouldNotBeNil)
		})
	})
}

func TestContainerdClientDoStart(t *testing.T) {
	convey.Convey("test containerd doStart when not stopped", t, func() {
		c := NewContainerdClient()
		convey.So(c.doStart("ctr-1", "ns"), convey.ShouldNotBeNil)
	})
}

func TestContainerdClientDoStopLoadFailed(t *testing.T) {
	convey.Convey("test containerd doStop load failed", t, func() {
		c := &ContainerdClient{client: &containerd.Client{}}
		patch := gomonkey.ApplyMethod(reflect.TypeOf(&containerd.Client{}), "LoadContainer",
			func(_ *containerd.Client, _ context.Context, _ string) (containerd.Container, error) {
				return nil, errors.New("load failed")
			})
		defer patch.Reset()
		convey.So(c.doStop("ctr-1", "ns"), convey.ShouldNotBeNil)
	})
}

func TestContainerdClientGetJobInfo(t *testing.T) {
	convey.Convey("test containerd getJobInfo", t, func() {
		c := NewContainerdClient()
		convey.Convey("wrong type returns recoverable default", func() {
			info := c.getJobInfo(nil, context.Background())
			convey.So(info, convey.ShouldResemble, domain.JobInfo{EnableRecover: true})
		})
		convey.Convey("parse labels", func() {
			mc := &mockContainer{id: "c1", labels: map[string]string{
				common.JobLabelID: "job-1", common.JobLabelReplica: "2",
			}}
			info := c.getJobInfo(mc, context.Background())
			convey.So(info, convey.ShouldResemble, domain.JobInfo{JobID: "job-1", JobReplica: 2, EnableRecover: true})
		})
		convey.Convey("labels failed returns recoverable default", func() {
			mc := &mockContainer{id: "c2", labelsErr: errors.New("labels failed")}
			info := c.getJobInfo(mc, context.Background())
			convey.So(info, convey.ShouldResemble, domain.JobInfo{EnableRecover: true})
		})
	})
}

func TestContainerdClientGetUsedDevs(t *testing.T) {
	convey.Convey("test containerd getUsedDevs", t, func() {
		c := NewContainerdClient()
		convey.Convey("wrong type returns nil", func() {
			devs, err := c.getUsedDevs(nil, context.Background())
			convey.So(err, convey.ShouldBeNil)
			convey.So(devs, convey.ShouldBeNil)
		})
		convey.Convey("parse env", func() {
			mc := &mockContainer{id: "c1", spec: &oci.Spec{
				Process: &specs.Process{Env: []string{api.AscendDeviceInfo + "=0,1"}},
				Linux:   &specs.Linux{Resources: &specs.LinuxResources{}},
			}}
			devs, err := c.getUsedDevs(mc, context.Background())
			convey.So(err, convey.ShouldBeNil)
			convey.So(devs, convey.ShouldResemble, []int32{0, 1})
		})
	})
}

func TestGetCtrValidSpec(t *testing.T) {
	convey.Convey("test getCtrValidSpec", t, func() {
		ctx := context.Background()
		convey.Convey("spec failed", func() {
			mc := &mockContainer{specErr: errors.New("spec failed")}
			_, err := getCtrValidSpec(mc, ctx)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("empty linux", func() {
			mc := &mockContainer{spec: &oci.Spec{}}
			_, err := getCtrValidSpec(mc, ctx)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("too many devices", func() {
			mc := &mockContainer{spec: &oci.Spec{
				Linux: &specs.Linux{Resources: &specs.LinuxResources{
					Devices: make([]specs.LinuxDeviceCgroup, maxDevicesNum+1),
				}},
			}}
			_, err := getCtrValidSpec(mc, ctx)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("empty process", func() {
			mc := &mockContainer{spec: &oci.Spec{
				Linux: &specs.Linux{Resources: &specs.LinuxResources{}},
			}}
			_, err := getCtrValidSpec(mc, ctx)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("valid spec", func() {
			mc := &mockContainer{spec: &oci.Spec{
				Process: &specs.Process{},
				Linux:   &specs.Linux{Resources: &specs.LinuxResources{}},
			}}
			spec, err := getCtrValidSpec(mc, ctx)
			convey.So(err, convey.ShouldBeNil)
			convey.So(spec, convey.ShouldNotBeNil)
		})
	})
}

func TestContainerdClientDoGetUsedDevs(t *testing.T) {
	convey.Convey("test containerd doGetUsedDevs", t, func() {
		c := NewContainerdClient()
		ctx := context.Background()
		convey.Convey("spec failed", func() {
			mc := &mockContainer{id: "c1", specErr: errors.New("spec failed")}
			_, err := c.doGetUsedDevs(mc, ctx)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("parse env from last to first", func() {
			mc := &mockContainer{id: "c2", spec: &oci.Spec{
				Process: &specs.Process{Env: []string{
					"PATH=/usr/bin", api.AscendDeviceInfo + "=0,2",
				}},
				Linux: &specs.Linux{Resources: &specs.LinuxResources{}},
			}}
			devs, err := c.doGetUsedDevs(mc, ctx)
			convey.So(err, convey.ShouldBeNil)
			convey.So(devs, convey.ShouldResemble, []int32{0, 2})
		})
		convey.Convey("no ascend env", func() {
			mc := &mockContainer{id: "c3", spec: &oci.Spec{
				Process: &specs.Process{Env: []string{}},
				Linux:   &specs.Linux{Resources: &specs.LinuxResources{}},
			}}
			devs, err := c.doGetUsedDevs(mc, ctx)
			convey.So(err, convey.ShouldBeNil)
			convey.So(devs, convey.ShouldResemble, []int32{})
		})
	})
}

func TestGetUsedDevsWithoutAscendRuntime(t *testing.T) {
	convey.Convey("test containerd getUsedDevsWithoutAscendRuntime", t, func() {
		convey.Convey("empty spec info", func() {
			_, err := getUsedDevsWithoutAscendRuntime(&oci.Spec{})
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("minor id overflow", func() {
			major := int64(0)
			minor := int64(math.MaxInt32) + 1
			spec := &oci.Spec{
				Linux: &specs.Linux{
					Resources: &specs.LinuxResources{
						Devices: []specs.LinuxDeviceCgroup{{Type: charDevice, Major: &major, Minor: &minor}},
					},
				},
			}
			_, err := getUsedDevsWithoutAscendRuntime(spec)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("match char device major", func() {
			patch := gomonkey.ApplyFunc(npuMajor, func() []string { return []string{"234"} })
			defer patch.Reset()

			major := int64(234)
			minor := int64(3)
			spec := &oci.Spec{
				Linux: &specs.Linux{
					Resources: &specs.LinuxResources{
						Devices: []specs.LinuxDeviceCgroup{{Type: charDevice, Major: &major, Minor: &minor}},
					},
				},
			}
			devs, err := getUsedDevsWithoutAscendRuntime(spec)
			convey.So(err, convey.ShouldBeNil)
			convey.So(devs, convey.ShouldResemble, []int32{3})
		})
	})
}
