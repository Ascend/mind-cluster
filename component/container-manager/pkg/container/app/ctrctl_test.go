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
	"errors"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/containerd/containerd"
	"github.com/docker/docker/api/types"
	"github.com/smartystreets/goconvey/convey"

	"container-manager/pkg/common"
	"container-manager/pkg/container/domain"
	"container-manager/pkg/devmgr"
	faultdomain "container-manager/pkg/fault/domain"
	resetdomain "container-manager/pkg/reset/domain"
)

func TestInitAndControl(t *testing.T) {
	convey.Convey("test initAndControl", t, func() {
		oldStrategy := common.ParamOption.CtrStrategy
		defer func() { common.ParamOption.CtrStrategy = oldStrategy }()

		convey.Convey("update ctr related info failed", func() {
			patch := gomonkey.ApplyPrivateMethod(reflect.TypeOf(&CtrCtl{}), "updateCtrRelatedInfo",
				func(*CtrCtl) error { return errors.New("x") })
			defer patch.Reset()
			cm := &CtrCtl{}
			cm.initAndControl()
		})
		convey.Convey("init ring failed under ring strategy", func() {
			common.ParamOption.CtrStrategy = common.RingStrategy
			patchU := gomonkey.ApplyPrivateMethod(reflect.TypeOf(&CtrCtl{}), "updateCtrRelatedInfo",
				func(*CtrCtl) error { return nil })
			defer patchU.Reset()
			patchR := gomonkey.ApplyPrivateMethod(reflect.TypeOf(&CtrCtl{}), "initRingInfo",
				func(*CtrCtl) error { return errors.New("x") })
			defer patchR.Reset()
			cm := &CtrCtl{}
			cm.initAndControl()
		})
		convey.Convey("init ring failed tolerated under single strategy", func() {
			common.ParamOption.CtrStrategy = common.SingleStrategy
			patchU := gomonkey.ApplyPrivateMethod(reflect.TypeOf(&CtrCtl{}), "updateCtrRelatedInfo",
				func(*CtrCtl) error { return nil })
			defer patchU.Reset()
			patchR := gomonkey.ApplyPrivateMethod(reflect.TypeOf(&CtrCtl{}), "initRingInfo",
				func(*CtrCtl) error { return errors.New("x") })
			defer patchR.Reset()
			patchP := gomonkey.ApplyPrivateMethod(reflect.TypeOf(&CtrCtl{}), "pauseCtr", func(*CtrCtl, bool) {})
			defer patchP.Reset()
			patchRS := gomonkey.ApplyPrivateMethod(reflect.TypeOf(&CtrCtl{}), "resumeCtr", func(*CtrCtl, bool) {})
			defer patchRS.Reset()
			cm := &CtrCtl{}
			cm.initAndControl()
		})
		convey.Convey("full success", func() {
			common.ParamOption.CtrStrategy = common.NeverStrategy
			patchU := gomonkey.ApplyPrivateMethod(reflect.TypeOf(&CtrCtl{}), "updateCtrRelatedInfo",
				func(*CtrCtl) error { return nil })
			defer patchU.Reset()
			patchR := gomonkey.ApplyPrivateMethod(reflect.TypeOf(&CtrCtl{}), "initRingInfo",
				func(*CtrCtl) error { return nil })
			defer patchR.Reset()
			cm := &CtrCtl{}
			cm.initAndControl()
		})
	})
}

func TestUpdateCtrRelatedInfo(t *testing.T) {
	convey.Convey("test updateCtrRelatedInfo", t, func() {
		convey.Convey("get all containers failed", func() {
			cm := &CtrCtl{client: &fakeClient{allContainersErr: errors.New("x")}}
			convey.So(cm.updateCtrRelatedInfo(), convey.ShouldNotBeNil)
		})
		convey.Convey("unexpected runtime type", func() {
			cm := &CtrCtl{client: &fakeClient{allContainers: "bad"}}
			convey.So(cm.updateCtrRelatedInfo(), convey.ShouldBeNil)
		})
		convey.Convey("docker containers with used devs", func() {
			cache := domain.GetCtrInfo()
			devCache := domain.NewDevCache([]int32{0, 1})
			fc := &fakeClient{
				allContainers: []types.Container{{ID: "d-1"}},
				usedDevs:      []int32{0},
				jobInfo:       domain.JobInfo{JobID: "job-d"},
			}
			cm := &CtrCtl{ctrInfoMap: cache, devInfoMap: devCache, client: fc}
			convey.So(cm.updateCtrRelatedInfo(), convey.ShouldBeNil)
			convey.So(cache.GetCtrNs("d-1"), convey.ShouldEqual, "default")
			convey.So(cache.GetCtrUsedDevs("d-1"), convey.ShouldResemble, []int32{0})
		})
		convey.Convey("docker used devs error", func() {
			cache := domain.GetCtrInfo()
			devCache := domain.NewDevCache([]int32{0, 1})
			fc := &fakeClient{allContainers: []types.Container{{ID: "d-2"}}, usedDevsErr: errors.New("x")}
			cm := &CtrCtl{ctrInfoMap: cache, devInfoMap: devCache, client: fc}
			convey.So(cm.updateCtrRelatedInfo(), convey.ShouldBeNil)
			convey.So(cache.GetCtrNs("d-2"), convey.ShouldEqual, "")
		})
	})
}

func TestUpdateForContainerd(t *testing.T) {
	convey.Convey("test updateForContainerd", t, func() {
		convey.Convey("with used devs", func() {
			cache := domain.GetCtrInfo()
			devCache := domain.NewDevCache([]int32{0, 1})
			fc := &fakeClient{usedDevs: []int32{0}, jobInfo: domain.JobInfo{JobID: "job-cd"}}
			cm := &CtrCtl{ctrInfoMap: cache, devInfoMap: devCache, client: fc}

			ctrIds := cm.updateForContainerd(map[string][]containerd.Container{
				"default": {&mockContainer{id: "cd-1"}},
			})
			convey.So(ctrIds, convey.ShouldResemble, []string{"cd-1"})
			convey.So(cache.GetCtrNs("cd-1"), convey.ShouldEqual, "default")
			convey.So(cache.GetCtrUsedDevs("cd-1"), convey.ShouldResemble, []int32{0})
		})
		convey.Convey("used devs error", func() {
			fc := &fakeClient{usedDevsErr: errors.New("x")}
			cm := &CtrCtl{client: fc}
			ctrIds := cm.updateForContainerd(map[string][]containerd.Container{
				"default": {&mockContainer{id: "cd-2"}},
			})
			convey.So(ctrIds, convey.ShouldResemble, []string{"cd-2"})
		})
		convey.Convey("no used devs", func() {
			fc := &fakeClient{usedDevs: []int32{}}
			cm := &CtrCtl{client: fc}
			ctrIds := cm.updateForContainerd(map[string][]containerd.Container{
				"default": {&mockContainer{id: "cd-3"}},
			})
			convey.So(ctrIds, convey.ShouldResemble, []string{"cd-3"})
		})
	})
}

func TestCtrControl(t *testing.T) {
	convey.Convey("test ctrControl", t, func() {
		oldStrategy := common.ParamOption.CtrStrategy
		defer func() { common.ParamOption.CtrStrategy = oldStrategy }()

		patchP := gomonkey.ApplyPrivateMethod(reflect.TypeOf(&CtrCtl{}), "pauseCtr", func(*CtrCtl, bool) {})
		defer patchP.Reset()
		patchR := gomonkey.ApplyPrivateMethod(reflect.TypeOf(&CtrCtl{}), "resumeCtr", func(*CtrCtl, bool) {})
		defer patchR.Reset()

		convey.Convey("single strategy", func() {
			common.ParamOption.CtrStrategy = common.SingleStrategy
			(&CtrCtl{}).ctrControl()
		})
		convey.Convey("ring strategy", func() {
			common.ParamOption.CtrStrategy = common.RingStrategy
			(&CtrCtl{}).ctrControl()
		})
		convey.Convey("unknown strategy", func() {
			common.ParamOption.CtrStrategy = "unknown"
			(&CtrCtl{}).ctrControl()
		})
	})
}

func TestIsSingleDevNeedPause(t *testing.T) {
	convey.Convey("test isSingleDevNeedPause", t, func() {
		convey.Convey("device in resetting", func() {
			cm := &CtrCtl{}
			patch := gomonkey.ApplyMethod(reflect.TypeOf(&resetdomain.NpuInResetCache{}), "IsNpuInReset",
				func(*resetdomain.NpuInResetCache, int32) bool { return true })
			defer patch.Reset()
			convey.So(cm.isSingleDevNeedPause(0), convey.ShouldBeTrue)
		})
		convey.Convey("get device error code failed", func() {
			cm := &CtrCtl{}
			patchR := gomonkey.ApplyMethod(reflect.TypeOf(&resetdomain.NpuInResetCache{}), "IsNpuInReset",
				func(*resetdomain.NpuInResetCache, int32) bool { return false })
			defer patchR.Reset()
			patchD := gomonkey.ApplyMethod(reflect.TypeOf(&devmgr.HwDevMgr{}), "GetDeviceErrCode",
				func(*devmgr.HwDevMgr, int32) (int32, []int64, error) { return 0, nil, errors.New("x") })
			defer patchD.Reset()
			convey.So(cm.isSingleDevNeedPause(0), convey.ShouldBeTrue)
		})
		convey.Convey("fault level need pause", func() {
			cm := &CtrCtl{}
			patchR := gomonkey.ApplyMethod(reflect.TypeOf(&resetdomain.NpuInResetCache{}), "IsNpuInReset",
				func(*resetdomain.NpuInResetCache, int32) bool { return false })
			defer patchR.Reset()
			patchD := gomonkey.ApplyMethod(reflect.TypeOf(&devmgr.HwDevMgr{}), "GetDeviceErrCode",
				func(*devmgr.HwDevMgr, int32) (int32, []int64, error) { return 0, []int64{1}, nil })
			defer patchD.Reset()
			patchL := gomonkey.ApplyFunc(faultdomain.GetFaultLevelByCode,
				func([]int64) string { return common.RestartRequest })
			defer patchL.Reset()
			convey.So(cm.isSingleDevNeedPause(0), convey.ShouldBeTrue)
		})
		convey.Convey("no pause needed", func() {
			cm := &CtrCtl{}
			patchR := gomonkey.ApplyMethod(reflect.TypeOf(&resetdomain.NpuInResetCache{}), "IsNpuInReset",
				func(*resetdomain.NpuInResetCache, int32) bool { return false })
			defer patchR.Reset()
			patchD := gomonkey.ApplyMethod(reflect.TypeOf(&devmgr.HwDevMgr{}), "GetDeviceErrCode",
				func(*devmgr.HwDevMgr, int32) (int32, []int64, error) { return 0, nil, nil })
			defer patchD.Reset()
			convey.So(cm.isSingleDevNeedPause(0), convey.ShouldBeFalse)
		})
	})
}

func TestIsDevsNeedPause(t *testing.T) {
	convey.Convey("test isDevsNeedPause", t, func() {
		cm := &CtrCtl{devInfoMap: domain.NewDevCache([]int32{0, 1})}
		patch := gomonkey.ApplyPrivateMethod(reflect.TypeOf(&CtrCtl{}), "isSingleDevNeedPause",
			func(_ *CtrCtl, id int32) bool { return id == 0 })
		defer patch.Reset()
		convey.So(cm.isDevsNeedPause([]int32{0, 1}), convey.ShouldBeTrue)
		convey.So(cm.isDevsNeedPause([]int32{1}), convey.ShouldBeFalse)
	})
}

func TestInitRingInfo(t *testing.T) {
	convey.Convey("test initRingInfo", t, func() {
		convey.Convey("deep copy failed", func() {
			cm := &CtrCtl{ctrInfoMap: domain.GetCtrInfo(), devInfoMap: domain.NewDevCache([]int32{0})}
			patch := gomonkey.ApplyFuncReturn(common.DeepCopy, errors.New("x"))
			defer patch.Reset()
			convey.So(cm.initRingInfo(), convey.ShouldNotBeNil)
		})
		convey.Convey("get phy id on ring failed", func() {
			cm := &CtrCtl{ctrInfoMap: domain.GetCtrInfo(), devInfoMap: domain.NewDevCache([]int32{0})}
			patch := gomonkey.ApplyMethod(reflect.TypeOf(&devmgr.HwDevMgr{}), "GetPhyIdOnRing",
				func(*devmgr.HwDevMgr, int32) ([]int32, error) { return nil, errors.New("x") })
			defer patch.Reset()
			convey.So(cm.initRingInfo(), convey.ShouldNotBeNil)
		})
		convey.Convey("success", func() {
			cm := &CtrCtl{ctrInfoMap: domain.GetCtrInfo(), devInfoMap: domain.NewDevCache([]int32{0, 1})}
			patch := gomonkey.ApplyMethod(reflect.TypeOf(&devmgr.HwDevMgr{}), "GetPhyIdOnRing",
				func(_ *devmgr.HwDevMgr, _ int32) ([]int32, error) { return []int32{0, 1}, nil })
			defer patch.Reset()
			convey.So(cm.initRingInfo(), convey.ShouldBeNil)
		})
	})
}

func TestPauseCtr(t *testing.T) {
	convey.Convey("test pauseCtr", t, func() {
		convey.Convey("recoverable ctr group is paused", func() {
			cache := domain.GetCtrInfo()
			devCache := domain.NewDevCache([]int32{0, 1})
			fc := &fakeClient{}
			cm := &CtrCtl{ctrInfoMap: cache, devInfoMap: devCache, client: fc}
			cm.setCtrRelatedInfo("pause-grp", "ns", []int32{0}, domain.JobInfo{EnableRecover: true})
			devCache.SetDevStatus(0, common.StatusNeedPause)

			cm.pauseCtr(false)
			status, _ := cache.GetCtrStatusAndStartTime("pause-grp")
			convey.So(status, convey.ShouldEqual, common.StatusPaused)
		})
		convey.Convey("non-recoverable ctr group is skipped", func() {
			cache := domain.GetCtrInfo()
			devCache := domain.NewDevCache([]int32{0, 1})
			cm := &CtrCtl{ctrInfoMap: cache, devInfoMap: devCache, client: &fakeClient{}}
			cm.setCtrRelatedInfo("pause-nr", "ns", []int32{0}, domain.JobInfo{EnableRecover: false})
			devCache.SetDevStatus(0, common.StatusNeedPause)

			cm.pauseCtr(false)
			status, _ := cache.GetCtrStatusAndStartTime("pause-nr")
			convey.So(status, convey.ShouldEqual, common.StatusRunning)
		})
	})
}

func TestPauseCtrsInGroups(t *testing.T) {
	convey.Convey("test pauseCtrsInGroups", t, func() {
		cache := domain.GetCtrInfo()
		devCache := domain.NewDevCache([]int32{0, 1})
		convey.Convey("coordinate success", func() {
			cache.SetCtrInfo("pg-dist", "ns", []int32{0}, domain.JobInfo{JobID: "job-d", JobReplica: 2})
			coord := &mockCoord{}
			fc := &fakeClient{}
			cm := &CtrCtl{ctrInfoMap: cache, devInfoMap: devCache, client: fc, coord: coord}

			convey.So(cm.pauseCtrsInGroups([]string{"pg-dist"}), convey.ShouldBeTrue)
			convey.So(coord.stopJobIDs, convey.ShouldResemble, []string{"job-d"})
		})
		convey.Convey("coordinate failed", func() {
			cache.SetCtrInfo("pg-dist-2", "ns", []int32{0}, domain.JobInfo{JobID: "job-d2", JobReplica: 2})
			coord := &mockCoord{stopErr: errors.New("x")}
			cm := &CtrCtl{ctrInfoMap: cache, devInfoMap: devCache, client: &fakeClient{}, coord: coord}

			convey.So(cm.pauseCtrsInGroups([]string{"pg-dist-2"}), convey.ShouldBeFalse)
		})
	})
}

func TestGetNeedResumeCtrInGroups(t *testing.T) {
	convey.Convey("test getNeedResumeCtrInGroups", t, func() {
		cache := domain.GetCtrInfo()
		devCache := domain.NewDevCache([]int32{0, 1})
		cache.SetCtrInfo("rn-1", "ns", []int32{0}, domain.JobInfo{})
		cache.SetCtrsOnRing([]string{"rn-1"})

		convey.Convey("devs need pause", func() {
			cm := &CtrCtl{ctrInfoMap: cache, devInfoMap: devCache}
			patch := gomonkey.ApplyPrivateMethod(reflect.TypeOf(&CtrCtl{}), "isDevsNeedPause",
				func(*CtrCtl, []int32) bool { return true })
			defer patch.Reset()
			convey.So(cm.getNeedResumeCtrInGroups("rn-1", false), convey.ShouldBeNil)
			convey.So(cm.getNeedResumeCtrInGroups("rn-1", true), convey.ShouldBeNil)
		})
		convey.Convey("devs not need pause", func() {
			cm := &CtrCtl{ctrInfoMap: cache, devInfoMap: devCache}
			patch := gomonkey.ApplyPrivateMethod(reflect.TypeOf(&CtrCtl{}), "isDevsNeedPause",
				func(*CtrCtl, []int32) bool { return false })
			defer patch.Reset()
			convey.So(cm.getNeedResumeCtrInGroups("rn-1", false), convey.ShouldResemble, []string{"rn-1"})
			convey.So(cm.getNeedResumeCtrInGroups("rn-1", true), convey.ShouldResemble, []string{"rn-1"})
		})
	})
}

func TestResumeCtr(t *testing.T) {
	convey.Convey("test resumeCtr", t, func() {
		cache := domain.GetCtrInfo()
		devCache := domain.NewDevCache([]int32{0, 1})
		cache.SetCtrInfo("resume-main", "ns", []int32{0}, domain.JobInfo{})
		cache.SetCtrsStatus("resume-main", common.StatusPaused)
		cm := &CtrCtl{ctrInfoMap: cache, devInfoMap: devCache, client: &fakeClient{}, coord: &mockCoord{}}

		patch := gomonkey.ApplyPrivateMethod(reflect.TypeOf(&CtrCtl{}), "isSingleDevNeedPause",
			func(*CtrCtl, int32) bool { return false })
		defer patch.Reset()

		cm.resumeCtr(false)
		status, _ := cache.GetCtrStatusAndStartTime("resume-main")
		convey.So(status, convey.ShouldEqual, common.StatusRunning)
	})
}

func TestResumeCtrsInGroups(t *testing.T) {
	convey.Convey("test resumeCtrsInGroups", t, func() {
		cache := domain.GetCtrInfo()
		devCache := domain.NewDevCache([]int32{0, 1})
		convey.Convey("local ctr", func() {
			cache.SetCtrInfo("rg-local", "ns", []int32{0}, domain.JobInfo{JobID: ""})
			fc := &fakeClient{}
			cm := &CtrCtl{ctrInfoMap: cache, devInfoMap: devCache, client: fc, coord: &mockCoord{}}
			cm.resumeCtrsInGroups([]string{"rg-local"}, &domain.PeerGateGroups{})
			convey.So(fc.startCalls, convey.ShouldContain, "rg-local")
		})
		convey.Convey("distributed ctr with gate none", func() {
			cache.SetCtrInfo("rg-dist", "ns", []int32{0}, domain.JobInfo{JobID: "job-rd", JobReplica: 2})
			coord := &mockCoord{}
			fc := &fakeClient{}
			cm := &CtrCtl{ctrInfoMap: cache, devInfoMap: devCache, client: fc, coord: coord}
			cm.resumeCtrsInGroups([]string{"rg-dist"}, &domain.PeerGateGroups{None: []string{"rg-dist"}})
			convey.So(coord.startJobIDs, convey.ShouldResemble, []string{"job-rd"})
		})
	})
}

func TestGetLocalContainersAndHasDataChanged(t *testing.T) {
	convey.Convey("test GetLocalContainers and HasDataChanged", t, func() {
		cm := &CtrCtl{ctrInfoMap: domain.GetCtrInfo()}
		_ = cm.GetLocalContainers()
		_ = cm.HasDataChanged()
	})
}
