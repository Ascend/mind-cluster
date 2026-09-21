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
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/containerd/containerd"
	"github.com/docker/docker/client"
	"github.com/smartystreets/goconvey/convey"

	"container-manager/pkg/common"
	"container-manager/pkg/container/domain"
	"container-manager/pkg/coordinator"
)

// fakeClient implements ContainerClient for workflow unit tests.
type fakeClient struct {
	startCalls []string
	stopCalls  []string
	startErr   error
	stopErr    error
	closeErr   error

	allContainers    interface{}
	allContainersErr error
	usedDevs         []int32
	usedDevsErr      error
	jobInfo          domain.JobInfo
}

func (f *fakeClient) init() error  { return nil }
func (f *fakeClient) close() error { return f.closeErr }
func (f *fakeClient) getAllContainers() (interface{}, error) {
	return f.allContainers, f.allContainersErr
}
func (f *fakeClient) getUsedDevs(containerObj interface{}, ctx context.Context) ([]int32, error) {
	return f.usedDevs, f.usedDevsErr
}
func (f *fakeClient) getJobInfo(containerObj interface{}, ctx context.Context) domain.JobInfo {
	return f.jobInfo
}
func (f *fakeClient) doStart(containerID, ns string) error {
	f.startCalls = append(f.startCalls, containerID)
	return f.startErr
}
func (f *fakeClient) doStop(containerID, ns string) error {
	f.stopCalls = append(f.stopCalls, containerID)
	return f.stopErr
}

// mockCoord implements coordinator.DistributedCoord.
type mockCoord struct {
	stopJobIDs  []string
	startJobIDs []string
	stopErr     error
	startErr    error
}

func (m *mockCoord) RequestStopJobs(jobIDs, ctrIds []string) error {
	m.stopJobIDs = append(m.stopJobIDs, jobIDs...)
	return m.stopErr
}
func (m *mockCoord) RequestStartJobs(jobIDs, ctrIds []string) error {
	m.startJobIDs = append(m.startJobIDs, jobIDs...)
	return m.startErr
}

var _ coordinator.DistributedCoord = (*mockCoord)(nil)

func TestNewCtrCtl(t *testing.T) {
	convey.Convey("test NewCtrCtl", t, func() {
		oldRuntime := common.ParamOption.RuntimeType
		defer func() { common.ParamOption.RuntimeType = oldRuntime }()

		convey.Convey("unknown runtime", func() {
			common.ParamOption.RuntimeType = "unknown"
			_, err := NewCtrCtl()
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("docker runtime connect failed", func() {
			common.ParamOption.RuntimeType = common.DockerType
			patch := gomonkey.ApplyFunc(client.NewClientWithOpts, func(...client.Opt) (*client.Client, error) {
				return nil, errors.New("connect failed")
			})
			defer patch.Reset()
			_, err := NewCtrCtl()
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("docker runtime success", func() {
			common.ParamOption.RuntimeType = common.DockerType
			patch := gomonkey.ApplyFunc(client.NewClientWithOpts, func(...client.Opt) (*client.Client, error) {
				return &client.Client{}, nil
			})
			defer patch.Reset()
			ctr, err := NewCtrCtl()
			convey.So(err, convey.ShouldBeNil)
			convey.So(ctr, convey.ShouldNotBeNil)
		})
		convey.Convey("containerd runtime success", func() {
			common.ParamOption.RuntimeType = common.ContainerDType
			patch := gomonkey.ApplyFunc(containerd.New, func(string, ...containerd.ClientOpt) (*containerd.Client, error) {
				return &containerd.Client{}, nil
			})
			defer patch.Reset()
			ctr, err := NewCtrCtl()
			convey.So(err, convey.ShouldBeNil)
			convey.So(ctr, convey.ShouldNotBeNil)
		})
	})
}

func TestWork(t *testing.T) {
	convey.Convey("test Work exits when context is done", t, func() {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		cm := &CtrCtl{}
		cm.Work(ctx)
	})
}

func TestShutDown(t *testing.T) {
	convey.Convey("test ShutDown", t, func() {
		convey.Convey("no error", func() {
			cm := &CtrCtl{client: &fakeClient{}}
			cm.ShutDown()
		})
		convey.Convey("close error", func() {
			cm := &CtrCtl{client: &fakeClient{closeErr: errors.New("close failed")}}
			cm.ShutDown()
		})
	})
}

func TestPartitionCtrs(t *testing.T) {
	convey.Convey("test partitionCtrs", t, func() {
		cache := domain.GetCtrInfo()
		cache.SetCtrInfo("local-id", "ns", nil, domain.JobInfo{JobID: ""})
		cache.SetCtrInfo("single-replica", "ns", nil, domain.JobInfo{JobID: "job-1", JobReplica: 1})
		cache.SetCtrInfo("dist-1", "ns", nil, domain.JobInfo{JobID: "job-2", JobReplica: 2})
		cache.SetCtrInfo("dist-2", "ns", nil, domain.JobInfo{JobID: "job-2", JobReplica: 2})
		cache.SetCtrInfo("dist-0-replica", "ns", nil, domain.JobInfo{JobID: "job-3", JobReplica: 0})

		cm := &CtrCtl{ctrInfoMap: cache}

		local, dist, jobs := cm.partitionCtrs([]string{
			"local-id", "single-replica", "dist-1", "dist-2", "dist-0-replica", "not-exist", "dist-1",
		})
		convey.So(local, convey.ShouldResemble, []string{"local-id", "single-replica"})
		convey.So(dist, convey.ShouldResemble, []string{"dist-1", "dist-2", "dist-0-replica"})
		convey.So(jobs, convey.ShouldResemble, []string{"job-2", "job-3"})
	})
}

func TestNameInitAndBindCoordinator(t *testing.T) {
	convey.Convey("test Name, Init and BindCoordinator", t, func() {
		cm := &CtrCtl{}
		convey.So(cm.Name(), convey.ShouldEqual, "container controller")
		convey.So(cm.Init(), convey.ShouldBeNil)

		coord := &mockCoord{}
		cm.BindCoordinator(coord)
		convey.So(cm.coord, convey.ShouldEqual, coord)
	})
}

func TestCtrControlNeverStrategy(t *testing.T) {
	convey.Convey("test ctrControl under never strategy", t, func() {
		oldStrategy := common.ParamOption.CtrStrategy
		common.ParamOption.CtrStrategy = common.NeverStrategy
		defer func() { common.ParamOption.CtrStrategy = oldStrategy }()

		cm := &CtrCtl{}
		cm.ctrControl()
	})
}

func TestSetCtrRelatedInfo(t *testing.T) {
	convey.Convey("test setCtrRelatedInfo", t, func() {
		cache := domain.GetCtrInfo()
		devCache := domain.NewDevCache([]int32{0, 1, 2})
		cm := &CtrCtl{ctrInfoMap: cache, devInfoMap: devCache}

		cm.setCtrRelatedInfo("ctr-set", "ns-set", []int32{0, 1}, domain.JobInfo{JobID: "job"})
		convey.So(cache.GetCtrNs("ctr-set"), convey.ShouldEqual, "ns-set")
		convey.So(cache.GetCtrUsedDevs("ctr-set"), convey.ShouldResemble, []int32{0, 1})
	})
}

func TestRemoveDeletedCtr(t *testing.T) {
	convey.Convey("test removeDeletedCtr", t, func() {
		cache := domain.GetCtrInfo()
		cache.SetCtrInfo("ctr-rm-keep", "ns", nil, domain.JobInfo{})
		cache.SetCtrInfo("ctr-rm-del", "ns", nil, domain.JobInfo{})
		cm := &CtrCtl{ctrInfoMap: cache, devInfoMap: domain.NewDevCache([]int32{0, 1})}

		cm.removeDeletedCtr([]string{"ctr-rm-keep"})
		convey.So(cache.GetCtrNs("ctr-rm-keep"), convey.ShouldEqual, "ns")
		convey.So(cache.GetCtrNs("ctr-rm-del"), convey.ShouldEqual, "")
	})
}

func TestDoPauseCtrs(t *testing.T) {
	convey.Convey("test doPauseCtrs", t, func() {
		cache := domain.GetCtrInfo()
		cache.SetCtrInfo("ctr-dop", "ns", []int32{0}, domain.JobInfo{})
		fc := &fakeClient{}
		cm := &CtrCtl{ctrInfoMap: cache, client: fc}

		cm.doPauseCtrs([]string{"ctr-dop"})
		status, _ := cache.GetCtrStatusAndStartTime("ctr-dop")
		convey.So(status, convey.ShouldEqual, common.StatusPaused)
		convey.So(fc.stopCalls, convey.ShouldResemble, []string{"ctr-dop"})
	})
}

func TestDoResumeCtrs(t *testing.T) {
	convey.Convey("test doResumeCtrs", t, func() {
		cache := domain.GetCtrInfo()
		cache.SetCtrInfo("ctr-dor", "ns", []int32{0}, domain.JobInfo{})
		fc := &fakeClient{}
		cm := &CtrCtl{ctrInfoMap: cache, client: fc}

		cm.doResumeCtrs([]string{"ctr-dor"})
		status, _ := cache.GetCtrStatusAndStartTime("ctr-dor")
		convey.So(status, convey.ShouldEqual, common.StatusRunning)
		convey.So(fc.startCalls, convey.ShouldResemble, []string{"ctr-dor"})
	})
}

func TestPauseJobContainers(t *testing.T) {
	convey.Convey("test PauseJobContainers", t, func() {
		cache := domain.GetCtrInfo()
		cache.SetCtrInfo("ctr-pjob", "ns", []int32{0}, domain.JobInfo{JobID: "job-p1", JobReplica: 2})
		cm := &CtrCtl{ctrInfoMap: cache}

		err := cm.PauseJobContainers([]string{"job-p1"}, []string{}, "peer-a")
		convey.So(err, convey.ShouldBeNil)
		convey.So(cache.GetCtrPausedByPeer("ctr-pjob"), convey.ShouldEqual, "peer-a")
	})
}

func TestResumeJobContainers(t *testing.T) {
	convey.Convey("test ResumeJobContainers", t, func() {
		cache := domain.GetCtrInfo()
		cache.SetCtrInfo("ctr-rjob", "ns", []int32{0}, domain.JobInfo{JobID: "job-r1", JobReplica: 2})
		cache.SetCtrsStatus("ctr-rjob", common.StatusPaused)
		cm := &CtrCtl{ctrInfoMap: cache}

		err := cm.ResumeJobContainers([]string{"job-r1"}, []string{}, "peer-b")
		convey.So(err, convey.ShouldBeNil)
		convey.So(cache.GetCtrPausedByPeer("ctr-rjob"), convey.ShouldEqual, "peer-b")
	})
}
