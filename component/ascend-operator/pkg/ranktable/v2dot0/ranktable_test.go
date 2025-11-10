/*
Copyright(C) 2024. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*
Package v2dot0 is used to test generate ranktable in v2.0
*/
package v2dot0

import (
	"encoding/json"
	"sync"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"ascend-common/api"
	"ascend-operator/pkg/api/v1"
	"ascend-operator/pkg/ranktable/common"
)

// TestGetNetInfoByDefault test case for GetNetInfoByDefault
func TestGetNetInfoByDefault(t *testing.T) {
	convey.Convey("TestGetNetInfoByDefault", t, func() {
		convey.Convey("01-should return empty when portAddrTypes is empty", func() {
			r := &RankTable{
				portAddrTypes: &sync.Map{},
			}
			expected := NetInfo{}
			res, err := r.GetNetInfoByDefault()
			convey.So(err, convey.ShouldBeNil)
			convey.So(res, convey.ShouldResemble, expected)
		})
		convey.Convey("02-should return roce info when portAddrTypes contain roce info", func() {
			r := &RankTable{
				portAddrTypes: &sync.Map{},
			}
			r.portAddrTypes.Store(v1.PortAddrTypeRoCE, struct{}{})
			expected := NetInfo{
				PortAddrType: v1.PortAddrTypeRoCE, ScaleOutType: v1.ScaleOutTypeRoCE, RankAddrType: v1.RankAddrTypeIP,
			}
			res, err := r.GetNetInfoByDefault()
			convey.So(err, convey.ShouldBeNil)
			convey.So(res, convey.ShouldResemble, expected)
		})
		convey.Convey("03-should return uboe info when portAddrTypes contain uboe info", func() {
			r := &RankTable{
				portAddrTypes: &sync.Map{},
			}
			r.portAddrTypes.Store(v1.PortAddrTypeUBoE, struct{}{})
			expected := NetInfo{
				PortAddrType: v1.PortAddrTypeUBoE, ScaleOutType: v1.ScaleOutTypeUBoE, RankAddrType: v1.RankAddrTypeIP,
			}
			res, err := r.GetNetInfoByDefault()
			convey.So(err, convey.ShouldBeNil)
			convey.So(res, convey.ShouldResemble, expected)
		})
		convey.Convey("04-should return ubg info when portAddrTypes contain ubg info", func() {
			r := &RankTable{
				portAddrTypes: &sync.Map{},
			}
			r.portAddrTypes.Store(v1.PortAddrTypeUBG, struct{}{})
			expected := NetInfo{
				PortAddrType: v1.PortAddrTypeUBG, ScaleOutType: v1.ScaleOutTypeUBoE, RankAddrType: v1.RankAddrTypeEID,
			}
			res, err := r.GetNetInfoByDefault()
			convey.So(err, convey.ShouldBeNil)
			convey.So(res, convey.ShouldResemble, expected)
		})
	})
}

// TestGetNetInfoByCustom test case for GetNetInfoByCustom
func TestGetNetInfoByCustom(t *testing.T) {
	convey.Convey("TestGetNetInfoByCustom", t, func() {
		convey.Convey("00-should return empty when portAddrTypes is empty", func() {
			r := &RankTable{
				customScaleOutType: v1.ScaleOutTypeRoCE,
				portAddrTypes:      &sync.Map{},
			}
			expected := NetInfo{}
			res, err := r.GetNetInfoByCustom()
			convey.So(err, convey.ShouldBeNil)
			convey.So(res, convey.ShouldResemble, expected)
		})
		convey.Convey("01-should return error when scaleout-type is invalid", func() {
			r := &RankTable{
				customScaleOutType: "XXX",
				portAddrTypes:      &sync.Map{},
			}
			expected := NetInfo{}
			res, err := r.GetNetInfoByCustom()
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(res, convey.ShouldResemble, expected)
		})
		testGetNetInfoByCustomWhenScaleOutTypeExists()
	})
}

func testGetNetInfoByCustomWhenScaleOutTypeExists() {
	convey.Convey("02-return roce info when scaleout-type=roce and portAddrTypes contain roce info", func() {
		r := &RankTable{
			customScaleOutType: v1.ScaleOutTypeRoCE,
			portAddrTypes:      &sync.Map{},
		}
		r.portAddrTypes.Store(v1.PortAddrTypeRoCE, struct{}{})
		expected := NetInfo{
			PortAddrType: v1.PortAddrTypeRoCE, ScaleOutType: v1.ScaleOutTypeRoCE, RankAddrType: v1.RankAddrTypeIP,
		}
		res, err := r.GetNetInfoByCustom()
		convey.So(err, convey.ShouldBeNil)
		convey.So(res, convey.ShouldResemble, expected)
	})
	convey.Convey("03-return uboe info when scaleout-type=uboe and portAddrTypes contain uboe info", func() {
		r := &RankTable{
			customScaleOutType: v1.ScaleOutTypeUBoE,
			portAddrTypes:      &sync.Map{},
		}
		r.portAddrTypes.Store(v1.PortAddrTypeUBoE, struct{}{})
		expected := NetInfo{
			PortAddrType: v1.PortAddrTypeUBoE, ScaleOutType: v1.ScaleOutTypeUBoE, RankAddrType: v1.RankAddrTypeIP,
		}
		res, err := r.GetNetInfoByCustom()
		convey.So(err, convey.ShouldBeNil)
		convey.So(res, convey.ShouldResemble, expected)
	})
	convey.Convey("04-return ubg info when scaleout-type=uboe and portAddrTypes contain ubg info", func() {
		r := &RankTable{
			customScaleOutType: v1.ScaleOutTypeUBoE,
			portAddrTypes:      &sync.Map{},
		}
		r.portAddrTypes.Store(v1.PortAddrTypeUBG, struct{}{})
		expected := NetInfo{
			PortAddrType: v1.PortAddrTypeUBG, ScaleOutType: v1.ScaleOutTypeUBoE, RankAddrType: v1.RankAddrTypeEID,
		}
		res, err := r.GetNetInfoByCustom()
		convey.So(err, convey.ShouldBeNil)
		convey.So(res, convey.ShouldResemble, expected)
	})
}

// Construct a fake Instance JSON
func fakeInstanceJSON() string {
	inst := common.Instance{
		PodName: "test-pod",
		Devices: []common.Dev{
			{
				DeviceID: "0",
				DeviceIP: "192.168.0.1",
				LevelList: []api.RankLevel{
					{Level: level0, Info: map[string]api.LevelElement{
						v1.PortAddrTypeUB: {NetLayer: level0, NetInstanceID: "L0"},
					}},
					{Level: level1, Info: map[string]api.LevelElement{
						v1.PortAddrTypeUB: {NetLayer: level1, NetInstanceID: "L1"},
					}},
					{Level: level2, Info: map[string]api.LevelElement{
						v1.ScaleOutTypeUBoE: {NetLayer: level2, NetInstanceID: "L2-UBOE"},
					}},
					{Level: level3, Info: map[string]api.LevelElement{
						v1.PortAddrTypeRoCE: {NetLayer: level3, NetInstanceID: "L3-ROCE"},
					}},
				},
			},
		},
	}
	b, err := json.Marshal(inst)
	if err != nil {
		return ""
	}
	return string(b)
}

func fakeAscendJob() *v1.AscendJob {
	return &v1.AscendJob{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{},
		},
	}
}

// TestNewRankTable verifies that creating a new RankTable from a minimal AscendJob
// returns a non-nil RankTable with its RankList properly initialized.
func TestNewRankTable(t *testing.T) {
	// Construct a minimal AscendJob
	job := fakeAscendJob()

	rt := New(job)
	if rt == nil {
		t.Fatal("expected RankTable, got nil")
	}
	if rt.RankList == nil {
		t.Error("expected RankList initialized")
	}
}

// TestAddPodAndGather checks that adding a Pod updates the RankTable correctly.
func TestAddPodAndGather(t *testing.T) {
	// Construct a fake Pod
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
			UID:       "uid-123",
			Annotations: map[string]string{
				api.Pod910DeviceAnno: fakeInstanceJSON(),
				api.PodRankIndexAnno: "0",
			},
		},
		Status: corev1.PodStatus{
			PodIP: "10.0.0.1",
		},
	}

	rt := New(fakeAscendJob())

	// Call AddPod
	if err := rt.AddPod(pod); err != nil {
		t.Fatalf("AddPod failed: %v", err)
	}

	// Gather RankList
	rt.GatherServerList()
	if rt.RankCount != 1 {
		t.Errorf("expected RankCount=1, got %d", rt.RankCount)
	}
	if len(rt.RankList) != 1 {
		t.Fatalf("expected RankList length=1, got %d", len(rt.RankList))
	}
}

// TestShouldInclude verifies the inclusion logic for different levels and address types.
func TestShouldInclude(t *testing.T) {
	if !shouldInclude(level2, v1.ScaleOutTypeUBoE, "") {
		t.Error("expected level=2 with UBOE to be included")
	}
	if shouldInclude(level2, v1.PortAddrTypeUB, "") {
		t.Error("expected level=2 with FOO to be excluded")
	}
	if !shouldInclude(level3, v1.ScaleOutTypeRoCE, "") {
		t.Error("expected level=3 with ROCE to be included")
	}
	if shouldInclude(level3, v1.ScaleOutTypeUBoE, "") {
		t.Error("expected level=3 with UBOE to be excluded")
	}
}

// TestGetElement checks that getElement retrieves the correct LevelElement from a device.
func TestGetElement(t *testing.T) {
	dev := common.Dev{
		DeviceID: "0",
		LevelList: []api.RankLevel{
			{Level: 0, Info: map[string]api.LevelElement{
				"UB": {NetLayer: 0, NetInstanceID: "L0"},
			}},
		},
	}
	elem, ok := getElement(dev, 0, "UB")
	if !ok || elem.NetInstanceID != "L0" {
		t.Errorf("expected element L0, got %+v", elem)
	}
}

// TestDeletePod checks that calling DeletePod reinitializes the internal maps in RankTable.
func TestDeletePod(t *testing.T) {
	rt := New(fakeAscendJob())
	rt.DeletePod()
	if rt.ranks == nil || rt.portAddrTypes == nil {
		t.Error("expected maps reinitialized")
	}
}

// genRankListForTest is a test-only helper that allows specifying portAddrType directly.
func genRankListForTest(inst *common.Instance, portAddrType string) *common.Rank {
	rank := &common.Rank{}
	curDevice := inst.Devices[0]

	for level := 0; level < 4; level++ {
		if !shouldInclude(level, portAddrType, "") {
			continue
		}
		elem, ok := getElement(curDevice, level, portAddrType)
		if !ok {
			// Fallback to default values when element is not found
			elem = api.LevelElement{
				NetLayer:      level,
				NetInstanceID: "CLUSTER",
				NetType:       "CLOS",
				NetAttr:       "",
			}
		}
		rank.LevelList = append(rank.LevelList, elem)
	}
	return rank
}

func newTestInstanceForGenRank() common.Instance {
	return common.Instance{
		PodName: "test-pod",
		Devices: []common.Dev{
			{
				DeviceID: "0",
				DeviceIP: "192.168.0.1",
				LevelList: []api.RankLevel{
					{Level: level0, Info: map[string]api.LevelElement{
						v1.PortAddrTypeUB: {NetLayer: level0, NetInstanceID: "L0"},
					}},
					{Level: level1, Info: map[string]api.LevelElement{
						v1.PortAddrTypeUB: {NetLayer: level1, NetInstanceID: "L1"},
					}},
					{Level: level2, Info: map[string]api.LevelElement{
						v1.PortAddrTypeUBoE: {NetLayer: level2, NetInstanceID: "L2-UBOE"},
						v1.PortAddrTypeUBG:  {NetLayer: level2, NetInstanceID: "L2-UBG"},
					}},
					{Level: level3, Info: map[string]api.LevelElement{
						v1.PortAddrTypeRoCE: {NetLayer: level3, NetInstanceID: "L3-ROCE"},
					}},
				},
			},
		},
	}
}

func assertRankList(t *testing.T, got []api.LevelElement, expectedLayers []int, expectedIDs []string) {
	t.Helper()

	minLen := len(got)
	if len(expectedLayers) < minLen {
		minLen = len(expectedLayers)
	}
	if len(expectedIDs) < minLen {
		minLen = len(expectedIDs)
	}

	for i := 0; i < minLen; i++ {
		if got[i].NetLayer != expectedLayers[i] {
			t.Errorf("expected NetLayer=%d at index %d, got %d",
				expectedLayers[i], i, got[i].NetLayer)
		}
		if got[i].NetInstanceID != expectedIDs[i] {
			t.Errorf("expected NetInstanceID=%s at index %d, got %s",
				expectedIDs[i], i, got[i].NetInstanceID)
		}
	}

	if len(got) != len(expectedLayers) || len(got) != len(expectedIDs) {
		t.Fatalf("length mismatch: got=%d, expectedLayers=%d, expectedIDs=%d",
			len(got), len(expectedLayers), len(expectedIDs))
	}
}

// TestGenRankList_DifferentPortAddrTypes verifies that genRankListForTest
// produces the expected RankList for different port address types.
func TestGenRankList_DifferentPortAddrTypes(t *testing.T) {
	inst := newTestInstanceForGenRank()

	tests := []struct {
		name           string
		portAddrType   string
		expectedLayers []int
		expectedIDs    []string
	}{
		{"UBOE allowed at L2", v1.PortAddrTypeUBoE,
			[]int{0, 1, 2}, []string{"L0", "L1", "L2-UBOE"}},
		{"UBG allowed at L2", v1.PortAddrTypeUBG,
			[]int{0, 1, 2}, []string{"L0", "L1", "L2-UBG"}},
		{"ROCE allowed at L3", v1.PortAddrTypeRoCE,
			[]int{0, 1, 3}, []string{"L0", "L1", "L3-ROCE"}},
		{"Other type skip L2/L3", v1.PortAddrTypeUB,
			[]int{0, 1}, []string{"L0", "L1"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rank := genRankListForTest(&inst, tt.portAddrType)
			assertRankList(t, rank.LevelList, tt.expectedLayers, tt.expectedIDs)
		})
	}
}

// TestAddPodErrors checks that AddPod returns errors for invalid Pod inputs.
func TestAddPodErrors(t *testing.T) {
	rt := New(fakeAscendJob())

	// nil Pod
	if err := rt.AddPod(nil); err == nil {
		t.Error("expected error when pod is nil")
	}

	// Pod without IP
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "no-ip",
			Annotations: map[string]string{
				api.Pod910DeviceAnno: "{}",
				api.PodRankIndexAnno: "0",
			},
		},
	}
	if err := rt.AddPod(pod); err == nil {
		t.Error("expected error when PodIP is empty")
	}
}

// TestGenRankListInvalidDeviceID checks that GenRankList returns an error when the DeviceID is not numeric.
func TestGenRankListInvalidDeviceID(t *testing.T) {
	inst := common.Instance{
		Devices: []common.Dev{
			{DeviceID: "not-a-number"},
		},
	}
	var rank common.Rank
	if err := GenRankList(&rank, &inst, 0); err == nil {
		t.Error("expected error when DeviceID is not numeric")
	}
}

// TestGatherServerListSorting checks that GatherServerList sorts the RankList by RankID.
func TestGatherServerListSorting(t *testing.T) {
	rt := New(fakeAscendJob())
	// Simulate two Ranks
	rt.ranks.Store("uid1", []*common.Rank{{RankID: 2}, {RankID: 1}})
	rt.GatherServerList()
	if rt.RankList[0].RankID != 1 {
		t.Errorf("expected first RankID=1, got %d", rt.RankList[0].RankID)
	}
}
