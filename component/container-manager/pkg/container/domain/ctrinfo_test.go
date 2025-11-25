package domain

import (
	"sort"
	"sync"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"container-manager/pkg/common"
)

var mockCtrCache = &CtrCache{}

func resetCtrCache() {
	ctrMap := make(map[string]*ctrInfo)
	for _, id := range []string{ctrId0, ctrId1, ctrId2, ctrId3} {
		ctrMap[id] = &ctrInfo{
			Id: id,
			Ns: testCtrNs,
		}
	}
	ctrMap[ctrId0].CtrsOnRing = []string{ctrId0, ctrId1}
	ctrMap[ctrId1].CtrsOnRing = []string{ctrId0, ctrId1}
	ctrMap[ctrId2].CtrsOnRing = []string{ctrId2, ctrId3}
	ctrMap[ctrId3].CtrsOnRing = []string{ctrId2, ctrId3}
	ctrMap[ctrId0].UsedDevs = []int32{devId0, devId1}
	ctrMap[ctrId1].UsedDevs = []int32{devId0, devId1}
	ctrMap[ctrId2].UsedDevs = []int32{devId2, devId3}
	ctrMap[ctrId3].UsedDevs = []int32{devId2, devId3}
	mockCtrCache = &CtrCache{
		ctrInfoMap: ctrMap,
		mutex:      sync.Mutex{},
	}
}

func TestGetCtrUsedDevs(t *testing.T) {
	convey.Convey("test method 'ResetDevStatus'", t, func() {
		resetCtrCache()
		usedDevs := mockCtrCache.GetCtrUsedDevs(ctrId0)
		convey.So(usedDevs, convey.ShouldResemble, []int32{devId0, devId1})
		usedDevs = mockCtrCache.GetCtrUsedDevs(ctrId4)
		convey.So(usedDevs, convey.ShouldResemble, []int32{})
	})
}

func TestSetCtrsStatus(t *testing.T) {
	convey.Convey("test method 'SetCtrsStatus', 'GetCtrStatusAndStartTime' and 'GetCtrsByStatus'", t, func() {
		resetCtrCache()
		mockCtrCache.SetCtrsStatus(ctrId0, common.StatusPaused)
		convey.So(mockCtrCache.ctrInfoMap[ctrId0].Status, convey.ShouldEqual, common.StatusPaused)
		status, _ := mockCtrCache.GetCtrStatusAndStartTime(ctrId0)
		convey.So(status, convey.ShouldEqual, common.StatusPaused)
		res := mockCtrCache.GetCtrsByStatus(common.StatusPaused)
		convey.So(res, convey.ShouldResemble, []string{ctrId0})

		mockCtrCache.SetCtrsStatus(ctrId4, common.StatusPaused)
		res = mockCtrCache.GetCtrsByStatus(common.StatusPaused)
		convey.So(res, convey.ShouldResemble, []string{ctrId0})
	})
}

func TestSetCtrsOnRing(t *testing.T) {
	convey.Convey("test method 'SetCtrsOnRing', 'GetCtrsOnRing'", t, func() {
		resetCtrCache()
		ctrsOnRing := []string{ctrId0, ctrId2, ctrId3}
		mockCtrCache.SetCtrsOnRing(ctrsOnRing)
		convey.So(mockCtrCache.GetCtrsOnRing(ctrId0), convey.ShouldResemble, ctrsOnRing)
		convey.So(mockCtrCache.GetCtrsOnRing(ctrId2), convey.ShouldResemble, ctrsOnRing)
		convey.So(mockCtrCache.GetCtrsOnRing(ctrId3), convey.ShouldResemble, ctrsOnRing)
		mockCtrCache.SetCtrsOnRing([]string{ctrId4})
		convey.So(mockCtrCache.GetCtrsOnRing(ctrId4), convey.ShouldResemble, []string{})
	})
}

func TestSetCtrInfo(t *testing.T) {
	convey.Convey("test method 'SetCtrInfo'", t, func() {
		resetCtrCache()
		usedDevs := []int32{devId0, devId1}
		mockCtrCache.SetCtrInfo(ctrId4, testCtrNs, usedDevs)
		convey.So(mockCtrCache.GetCtrUsedDevs(ctrId4), convey.ShouldResemble, usedDevs)
		convey.So(mockCtrCache.GetCtrNs(ctrId4), convey.ShouldEqual, testCtrNs)
		convey.So(mockCtrCache.GetCtrNs(""), convey.ShouldEqual, "")
	})
}

func TestGetCtrStatusOnRing(t *testing.T) {
	convey.Convey("test method 'GetCtrRelatedDevs'", t, func() {
		resetCtrCache()
		usedDevs := mockCtrCache.GetCtrRelatedDevs([]string{ctrId0, ctrId1})
		expUsedDevs := []int32{devId0, devId1}
		convey.So(usedDevs, convey.ShouldResemble, expUsedDevs)
		usedDevs = mockCtrCache.GetCtrRelatedDevs([]string{ctrId0, ctrId3})
		sort.Slice(usedDevs, func(i, j int) bool {
			return usedDevs[i] < usedDevs[j]
		})
		expUsedDevs = []int32{devId0, devId1, devId2, devId3}
		convey.So(usedDevs, convey.ShouldResemble, expUsedDevs)
	})
}

func TestRemoveDeletedCtrForCtrInfo(t *testing.T) {
	convey.Convey("test method 'RemoveDeletedCtr'", t, func() {
		resetCtrCache()
		mockCtrCache.RemoveDeletedCtr([]string{ctrId0, ctrId1, ctrId2})
		convey.So(len(mockCtrCache.ctrInfoMap), convey.ShouldEqual, len3)
		mockCtrCache.RemoveDeletedCtr([]string{ctrId0, ctrId1, ctrId4})
		convey.So(len(mockCtrCache.ctrInfoMap), convey.ShouldEqual, len2)
		mockCtrCache.RemoveDeletedCtr([]string{})
		convey.So(len(mockCtrCache.ctrInfoMap), convey.ShouldEqual, len0)
	})
}

func TestDeepCopyForCtrInfo(t *testing.T) {
	convey.Convey("test method 'DeepCopy'", t, func() {
		resetCtrCache()
		cpCtr, err := mockCtrCache.DeepCopy()
		convey.So(err, convey.ShouldBeNil)
		convey.So(len(cpCtr), convey.ShouldEqual, len4)

		p1 := gomonkey.ApplyFuncReturn(common.DeepCopy, testErr)
		defer p1.Reset()
		_, err = mockCtrCache.DeepCopy()
		convey.So(err, convey.ShouldResemble, testErr)
	})
}
