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
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/api"
	"container-manager/pkg/common"
	"container-manager/pkg/container/domain"
)

func TestGetUsedDevsWithoutAscendRuntimeForDocker(t *testing.T) {
	convey.Convey("test parse docker devices without ascend runtime", t, func() {
		convey.Convey("valid davinci devices", func() {
			resources := container.Resources{
				Devices: []container.DeviceMapping{
					{PathOnHost: "/dev/davinci0", PathInContainer: "/dev/davinci0"},
					{PathOnHost: "/dev/davinci1", PathInContainer: "/dev/davinci1"},
				},
			}
			devs, err := getUsedDevsWithoutAscendRuntimeForDocker(resources)
			convey.So(err, convey.ShouldBeNil)
			convey.So(devs, convey.ShouldResemble, []int32{0, 1})
		})
		convey.Convey("ignore non-davinci devices", func() {
			resources := container.Resources{
				Devices: []container.DeviceMapping{
					{PathOnHost: "/dev/davinci_manager", PathInContainer: "/dev/davinci_manager"},
					{PathOnHost: "/dev/ummu", PathInContainer: "/dev/ummu"},
					{PathOnHost: "/dev/davinci3", PathInContainer: "/dev/davinci3"},
				},
			}
			devs, err := getUsedDevsWithoutAscendRuntimeForDocker(resources)
			convey.So(err, convey.ShouldBeNil)
			convey.So(devs, convey.ShouldResemble, []int32{3})
		})
		convey.Convey("non-numeric davinci path is ignored", func() {
			resources := container.Resources{
				Devices: []container.DeviceMapping{
					{PathOnHost: "/dev/davinciabc", PathInContainer: "/dev/davinciabc"},
				},
			}
			devs, err := getUsedDevsWithoutAscendRuntimeForDocker(resources)
			convey.So(err, convey.ShouldBeNil)
			convey.So(devs, convey.ShouldResemble, []int32{})
		})
		convey.Convey("empty devices", func() {
			devs, err := getUsedDevsWithoutAscendRuntimeForDocker(container.Resources{})
			convey.So(err, convey.ShouldBeNil)
			convey.So(devs, convey.ShouldResemble, []int32{})
		})
	})
}

func TestDockerClientGetJobInfo(t *testing.T) {
	convey.Convey("test docker getJobInfo", t, func() {
		d := &DockerClient{}
		convey.Convey("parse labels", func() {
			info := d.getJobInfo(types.Container{
				ID:     "ctr-1",
				Labels: map[string]string{common.JobLabelID: "job-1", common.JobLabelReplica: "3"},
			}, context.Background())
			convey.So(info, convey.ShouldResemble, domain.JobInfo{JobID: "job-1", JobReplica: 3, EnableRecover: true})
		})
		convey.Convey("wrong type returns recoverable default", func() {
			info := d.getJobInfo(nil, context.Background())
			convey.So(info, convey.ShouldResemble, domain.JobInfo{EnableRecover: true})
		})
	})
}

func TestDockerClientGetUsedDevs(t *testing.T) {
	convey.Convey("test docker getUsedDevs", t, func() {
		d := &DockerClient{}
		convey.Convey("wrong type returns nil", func() {
			devs, err := d.getUsedDevs(nil, context.Background())
			convey.So(err, convey.ShouldBeNil)
			convey.So(devs, convey.ShouldBeNil)
		})
	})
}

func TestDockerClientDoGetUsedDevsExited(t *testing.T) {
	convey.Convey("test docker doGetUsedDevs for exited container", t, func() {
		d := &DockerClient{}
		devs, err := d.doGetUsedDevs(types.Container{ID: "ctr-1", Status: "Exited (0) 1 hour ago"})
		convey.So(err, convey.ShouldBeNil)
		convey.So(devs, convey.ShouldResemble, []int32{})
	})
}

func TestDockerClientClose(t *testing.T) {
	convey.Convey("test docker close", t, func() {
		d := &DockerClient{client: &client.Client{}}
		patch := gomonkey.ApplyMethod(reflect.TypeOf(&client.Client{}), "Close",
			func(*client.Client) error { return nil })
		defer patch.Reset()
		convey.So(d.close(), convey.ShouldBeNil)
	})
}

func TestDockerClientDoStopAndStart(t *testing.T) {
	convey.Convey("test docker doStop and doStart", t, func() {
		d := &DockerClient{client: &client.Client{}}
		patchStop := gomonkey.ApplyMethod(reflect.TypeOf(&client.Client{}), "ContainerStop",
			func(_ *client.Client, _ context.Context, _ string, _ container.StopOptions) error { return nil })
		defer patchStop.Reset()
		convey.So(d.doStop("c", "ns"), convey.ShouldBeNil)

		patchStart := gomonkey.ApplyMethod(reflect.TypeOf(&client.Client{}), "ContainerStart",
			func(_ *client.Client, _ context.Context, _ string, _ container.StartOptions) error { return nil })
		defer patchStart.Reset()
		convey.So(d.doStart("c", "ns"), convey.ShouldBeNil)
	})
}

func TestDockerClientGetAllContainers(t *testing.T) {
	convey.Convey("test docker getAllContainers", t, func() {
		convey.Convey("list failed", func() {
			d := &DockerClient{client: &client.Client{}}
			patch := gomonkey.ApplyMethod(reflect.TypeOf(&client.Client{}), "ContainerList",
				func(_ *client.Client, _ context.Context, _ container.ListOptions) ([]types.Container, error) {
					return nil, errors.New("x")
				})
			defer patch.Reset()
			ctrs, err := d.getAllContainers()
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(ctrs, convey.ShouldBeNil)
		})
		convey.Convey("success", func() {
			d := &DockerClient{client: &client.Client{}}
			patch := gomonkey.ApplyMethod(reflect.TypeOf(&client.Client{}), "ContainerList",
				func(_ *client.Client, _ context.Context, _ container.ListOptions) ([]types.Container, error) {
					return []types.Container{{ID: "d1"}}, nil
				})
			defer patch.Reset()
			ctrs, err := d.getAllContainers()
			convey.So(err, convey.ShouldBeNil)
			convey.So(ctrs, convey.ShouldResemble, []types.Container{{ID: "d1"}})
		})
	})
}

func TestDockerClientDoGetUsedDevs(t *testing.T) {
	convey.Convey("test docker doGetUsedDevs", t, func() {
		convey.Convey("inspect failed", func() {
			d := &DockerClient{client: &client.Client{}}
			patch := gomonkey.ApplyMethod(reflect.TypeOf(&client.Client{}), "ContainerInspect",
				func(_ *client.Client, _ context.Context, _ string) (types.ContainerJSON, error) {
					return types.ContainerJSON{}, errors.New("x")
				})
			defer patch.Reset()
			_, err := d.doGetUsedDevs(types.Container{ID: "c1"})
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("env found", func() {
			d := &DockerClient{client: &client.Client{}}
			patch := gomonkey.ApplyMethod(reflect.TypeOf(&client.Client{}), "ContainerInspect",
				func(_ *client.Client, _ context.Context, _ string) (types.ContainerJSON, error) {
					return types.ContainerJSON{
						Config: &container.Config{Env: []string{api.AscendDeviceInfo + "=0,3"}},
					}, nil
				})
			defer patch.Reset()
			devs, err := d.doGetUsedDevs(types.Container{ID: "c2"})
			convey.So(err, convey.ShouldBeNil)
			convey.So(devs, convey.ShouldResemble, []int32{0, 3})
		})
		convey.Convey("env parse error", func() {
			d := &DockerClient{client: &client.Client{}}
			patch := gomonkey.ApplyMethod(reflect.TypeOf(&client.Client{}), "ContainerInspect",
				func(_ *client.Client, _ context.Context, _ string) (types.ContainerJSON, error) {
					return types.ContainerJSON{
						Config: &container.Config{Env: []string{api.AscendDeviceInfo + "=a,b"}},
					}, nil
				})
			defer patch.Reset()
			_, err := d.doGetUsedDevs(types.Container{ID: "c3"})
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("no ascend env uses resources", func() {
			d := &DockerClient{client: &client.Client{}}
			patch := gomonkey.ApplyMethod(reflect.TypeOf(&client.Client{}), "ContainerInspect",
				func(_ *client.Client, _ context.Context, _ string) (types.ContainerJSON, error) {
					return types.ContainerJSON{
						Config: &container.Config{Env: []string{"PATH=/usr/bin"}},
						ContainerJSONBase: &types.ContainerJSONBase{
							HostConfig: &container.HostConfig{Resources: container.Resources{
								Devices: []container.DeviceMapping{
									{PathOnHost: "/dev/davinci0", PathInContainer: "/dev/davinci0"},
								},
							}},
						},
					}, nil
				})
			defer patch.Reset()
			devs, err := d.doGetUsedDevs(types.Container{ID: "c4"})
			convey.So(err, convey.ShouldBeNil)
			convey.So(devs, convey.ShouldResemble, []int32{0})
		})
	})
}
