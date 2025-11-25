package domain

import (
	"sync"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"container-manager/pkg/common"
)

var (
	mockDevCache = &DevCache{}

	mockFault1 = common.DevFaultInfo{
		EventID:    eventId0,
		PhyID:      devId0,
		FaultLevel: common.SeparateNPU,
	}
	mockFault2 = common.DevFaultInfo{
		EventID:    eventId1,
		PhyID:      devId0,
		FaultLevel: common.RestartNPU,
	}
	mockFault3 = common.DevFaultInfo{
		EventID:    eventId1,
		PhyID:      devId1,
		FaultLevel: common.NotHandleFault,
	}
)

func resetDevCache() {
	devMap := make(map[int32]*devInfo)
	for _, id := range []int32{devId0, devId1, devId2, devId3} {
		devMap[id] = &devInfo{
			CtrIds: []string{},
			Status: common.StatusIgnorePause,
		}
	}
	devMap[devId0].DevsOnRing = []int32{devId0, devId1}
	devMap[devId1].DevsOnRing = []int32{devId0, devId1}
	devMap[devId2].DevsOnRing = []int32{devId2, devId3}
	devMap[devId3].DevsOnRing = []int32{devId2, devId3}
	mockDevCache = &DevCache{
		devInfoMap: devMap,
		mutex:      sync.Mutex{},
	}
}

func TestResetDevStatus(t *testing.T) {
	convey.Convey("test method 'ResetDevStatus'", t, func() {
		resetDevCache()
		mockDevCache.ResetDevStatus()
		for _, info := range mockDevCache.devInfoMap {
			convey.So(info.Status, convey.ShouldEqual, common.StatusIgnorePause)
		}
	})
}

func TestSetCtrRelatedInfo(t *testing.T) {
	convey.Convey("test method 'SetCtrRelatedInfo'", t, func() {
		resetDevCache()
		mockDevCache.SetCtrRelatedInfo(ctrId0, []int32{devId0, devId1, devId2})
		for id, info := range mockDevCache.devInfoMap {
			if id == devId0 || id == devId1 || id == devId2 {
				convey.So(info.CtrIds, convey.ShouldResemble, []string{ctrId0})
			}
		}
		mockDevCache.SetCtrRelatedInfo(ctrId1, []int32{devId0, devId2})
		for id, info := range mockDevCache.devInfoMap {
			if id == devId0 || id == devId2 {
				convey.So(info.CtrIds, convey.ShouldResemble, []string{ctrId0, ctrId1})
			}
			if id == devId1 {
				convey.So(info.CtrIds, convey.ShouldResemble, []string{ctrId0})
			}
		}
	})
}

func TestRemoveDeletedCtr(t *testing.T) {
	convey.Convey("test method 'RemoveDeletedCtr'", t, func() {
		resetDevCache()
		mockDevCache.SetCtrRelatedInfo(ctrId0, []int32{devId0, devId1, devId2})
		mockDevCache.SetCtrRelatedInfo(ctrId1, []int32{devId0, devId2, devId3})
		mockDevCache.SetCtrRelatedInfo(ctrId2, []int32{devId3})
		mockDevCache.RemoveDeletedCtr([]string{ctrId0})
		for id, info := range mockDevCache.devInfoMap {
			if id == devId0 || id == devId2 {
				convey.So(info.CtrIds, convey.ShouldResemble, []string{ctrId0})
			}
			if id == devId1 {
				convey.So(info.CtrIds, convey.ShouldResemble, []string{ctrId0})
			}
			if id == devId3 {
				convey.So(info.CtrIds, convey.ShouldResemble, []string{})
			}
		}
	})
}

func TestUpdateDevStatus(t *testing.T) {
	convey.Convey("test method 'UpdateDevStatus'", t, func() {
		resetDevCache()
		faultCache := map[int32][]*common.DevFaultInfo{
			devId0: {&mockFault1, &mockFault2},
			devId1: {&mockFault3},
		}
		mockDevCache.UpdateDevStatus(faultCache)
		for id, info := range mockDevCache.devInfoMap {
			if id == devId0 {
				convey.So(info.Status, convey.ShouldEqual, common.StatusNeedPause)
			}
			if id == devId1 {
				convey.So(info.Status, convey.ShouldEqual, common.StatusIgnorePause)
			}
		}
		mockDevCache.UpdateDevStatus(nil)
		for id, info := range mockDevCache.devInfoMap {
			if id == devId0 {
				convey.So(info.Status, convey.ShouldEqual, common.StatusNeedPause)
			}
			if id == devId1 {
				convey.So(info.Status, convey.ShouldEqual, common.StatusIgnorePause)
			}
		}
	})
}

func TestGetNeedPausedCtr(t *testing.T) {
	convey.Convey("test method 'GetNeedPausedCtr'", t, func() {
		resetDevCache()
		// devId2 is ok
		faultCache := map[int32][]*common.DevFaultInfo{
			devId0: {&mockFault1, &mockFault2},
			devId1: {&mockFault3},
			devId2: {&mockFault3},
		}
		mockDevCache.UpdateDevStatus(faultCache)
		mockDevCache.SetCtrRelatedInfo(ctrId0, []int32{devId0, devId1, devId2})
		mockDevCache.SetCtrRelatedInfo(ctrId1, []int32{devId1})
		mockDevCache.SetCtrRelatedInfo(ctrId2, []int32{devId2})
		ctrs := mockDevCache.GetNeedPausedCtr(false)
		convey.So(ctrs, convey.ShouldResemble, []string{ctrId0})
		ctrs = mockDevCache.GetNeedPausedCtr(true)
		convey.So(ctrs, convey.ShouldResemble, []string{ctrId0, ctrId1})
	})
}

func TestSetDevStatus(t *testing.T) {
	convey.Convey("test method 'SetDevStatus'", t, func() {
		resetDevCache()
		mockDevCache.SetDevStatus(devId0, common.StatusResuming)
		convey.So(mockDevCache.devInfoMap[devId0].Status, convey.ShouldEqual, common.StatusResuming)
	})
}

func TestSetDevsOnRing(t *testing.T) {
	convey.Convey("test method 'SetDevsOnRing'", t, func() {
		resetDevCache()
		mockRing := []int32{devId0, devId1, devId2, devId3}
		mockDevCache.SetDevsOnRing(devId0, mockRing)
		convey.So(mockDevCache.devInfoMap[devId0].DevsOnRing, convey.ShouldResemble, mockRing)
	})
}

func TestGetDevsRelatedCtrs(t *testing.T) {
	convey.Convey("test method 'GetDevsRelatedCtrs'", t, func() {
		resetDevCache()
		mockDevCache.SetCtrRelatedInfo(ctrId0, []int32{devId0, devId1, devId2})
		ctrs := mockDevCache.GetDevsRelatedCtrs(devId0)
		convey.So(ctrs, convey.ShouldResemble, []string{ctrId0})
		ctrs = mockDevCache.GetDevsRelatedCtrs(devId1)
		convey.So(ctrs, convey.ShouldResemble, []string{ctrId0})
		ctrs = mockDevCache.GetDevsRelatedCtrs(devId2)
		convey.So(ctrs, convey.ShouldResemble, []string{ctrId0})
	})
}

func TestDeepCopy(t *testing.T) {
	convey.Convey("test method 'DeepCopy'", t, func() {
		resetDevCache()
		cpDev, err := mockDevCache.DeepCopy()
		convey.So(err, convey.ShouldBeNil)
		convey.So(len(cpDev), convey.ShouldEqual, len4)

		p1 := gomonkey.ApplyFuncReturn(common.DeepCopy, testErr)
		defer p1.Reset()
		_, err = mockDevCache.DeepCopy()
		convey.So(err, convey.ShouldResemble, testErr)
	})
}
