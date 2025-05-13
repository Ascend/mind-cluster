// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package faultdomain contain fault date structure
package faultdomain

import (
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/api/core/v1"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
)

const (
	jobId                      = "JobId"
	nodeName                   = "Node"
	time100Seconds             = int64(100000)
	time120Seconds             = int64(120000)
	time1Seconds               = int64(1000)
	deviceId                   = "0"
	rankID                     = "8"
	cmName                     = "mindx-dl-deviceinfo-" + nodeName
	deviceName                 = constant.Ascend910 + "-" + deviceId
	originalDeviceFaultCodeCnt = 2
)

var (
	jobServerMap = constant.JobServerInfoMap{
		InfoMap: map[string]map[string]constant.ServerHccl{
			jobId: {
				nodeName: {
					DeviceList: []constant.Device{{
						DeviceID: deviceId,
						RankID:   rankID,
					}},
					ServerName: nodeName,
				},
			},
		},
	}
	originalDeviceCm = &constant.DeviceInfo{
		CmName: cmName,
		DeviceInfoNoName: constant.DeviceInfoNoName{
			DeviceList: map[string]string{
				"huawei.com/Ascend910-Fault": `
[
  {
    "fault_type": "CardUnhealthy",
	"fault_code": "80E01801      ,  80C98009    ",
    "fault_time_and_level_map":
      {
        "80E01801": {"fault_time":100000, "fault_level": "RestartBusiness"}, 
        "80C98009": {"fault_time":120000, "fault_level": "NotHandleFault"}
      },"npu_name": "Ascend910-0"
  }
]`,
			},
		},
	}
)

func TestMain(m *testing.M) {
	hwlog.InitRunLogger(&hwlog.LogConfig{OnlyToStdout: true}, nil)
	m.Run()
}

func TestSplitDeviceFault(t *testing.T) {
	t.Run("TestSplitDeviceFault", func(t *testing.T) {
		npuName := "Ascend910-0"
		var faultInfo = constant.DeviceFault{
			NPUName:   npuName,
			FaultCode: "0x1,0x2",
			FaultTimeAndLevelMap: map[string]constant.FaultTimeAndLevel{
				"0x1": {FaultLevel: constant.NotHandleFault, FaultTime: 1},
				"0x2": {FaultLevel: constant.SubHealthFault, FaultTime: 1},
			},
		}

		got := splitDeviceFault(faultInfo, "node1")
		want := []constant.DeviceFault{
			{
				NPUName:              npuName,
				FaultCode:            "0x1",
				FaultLevel:           constant.NotHandleFault,
				LargeModelFaultLevel: constant.NotHandleFault,
				FaultHandling:        constant.NotHandleFault,
				FaultTimeAndLevelMap: map[string]constant.FaultTimeAndLevel{
					"0x1": {FaultLevel: constant.NotHandleFault, FaultTime: 1},
				},
			}, {
				NPUName:              npuName,
				FaultCode:            "0x2",
				FaultLevel:           constant.SubHealthFault,
				LargeModelFaultLevel: constant.SubHealthFault,
				FaultHandling:        constant.SubHealthFault,
				FaultTimeAndLevelMap: map[string]constant.FaultTimeAndLevel{
					"0x2": {FaultLevel: constant.SubHealthFault, FaultTime: 1},
				},
			},
		}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("splitDeviceFault() = %v, want %v", got, want)
		}
	})
}

// TestSplitDeviceFaultWithManuallySeparateFaultLevel should split out a DeviceFault as middle data,
// when dp report constant.ManuallySeparateNPU
func TestSplitDeviceFaultWithManuallySeparateFaultLevel(t *testing.T) {
	t.Run("TestSplitDeviceFaultWithManuallySeparateFaultLevel", func(t *testing.T) {
		npuName := "Ascend910-0"
		var faultInfo = constant.DeviceFault{
			NPUName:              npuName,
			FaultCode:            "",
			FaultLevel:           constant.ManuallySeparateNPU,
			FaultTimeAndLevelMap: map[string]constant.FaultTimeAndLevel{},
		}

		got := splitDeviceFault(faultInfo, "node1")
		want := []constant.DeviceFault{
			{
				NPUName:              npuName,
				FaultCode:            constant.ManuallySeparateNPU,
				FaultLevel:           constant.ManuallySeparateNPU,
				LargeModelFaultLevel: constant.ManuallySeparateNPU,
				FaultHandling:        constant.ManuallySeparateNPU,
				FaultTimeAndLevelMap: map[string]constant.FaultTimeAndLevel{
					constant.ManuallySeparateNPU: {
						FaultTime:  constant.UnknownFaultTime,
						FaultLevel: constant.ManuallySeparateNPU,
					},
				},
			},
		}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("splitDeviceFault() = %v, want %v", got, want)
		}
	})
}

// TestMergeSameTypeDeviceFault should be merged, when fault type is same
func TestMergeSameTypeDeviceFault(t *testing.T) {
	t.Run("Test_mergeDeviceFault", func(t *testing.T) {
		npuName := "Ascend910-0"
		split := []constant.DeviceFault{
			{
				FaultType:  constant.CardUnhealthy,
				NPUName:    npuName,
				FaultCode:  "0x1",
				FaultLevel: constant.NotHandleFault,
				FaultTimeAndLevelMap: map[string]constant.FaultTimeAndLevel{
					"0x1": {FaultLevel: constant.NotHandleFault, FaultTime: 1},
					"0x2": {FaultLevel: constant.SubHealthFault, FaultTime: 1},
				},
			},
			{
				FaultType:  constant.CardUnhealthy,
				NPUName:    npuName,
				FaultCode:  "0x2",
				FaultLevel: constant.SubHealthFault,
				FaultTimeAndLevelMap: map[string]constant.FaultTimeAndLevel{
					"0x1": {FaultLevel: constant.NotHandleFault, FaultTime: 1},
					"0x2": {FaultLevel: constant.SubHealthFault, FaultTime: 1},
				},
			},
		}
		want := []constant.DeviceFault{
			{
				FaultType:            constant.CardUnhealthy,
				NPUName:              npuName,
				FaultCode:            "0x1,0x2",
				FaultLevel:           constant.SubHealthFault,
				LargeModelFaultLevel: constant.SubHealthFault,
				FaultHandling:        constant.SubHealthFault,
				FaultTimeAndLevelMap: map[string]constant.FaultTimeAndLevel{
					"0x1": {FaultLevel: constant.NotHandleFault, FaultTime: 1},
					"0x2": {FaultLevel: constant.SubHealthFault, FaultTime: 1},
				},
			},
		}
		got, err := mergeDeviceFault(split)
		if err != nil {
			t.Errorf("mergeDeviceFault() error = %v", err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("mergeDeviceFault() got = %v, want %v", util.ObjToString(got), util.ObjToString(want))
		}
	})
}

// TestMergeDifferentTypeDeviceFault should not be merged, when fault type isn't same
func TestMergeDifferentTypeDeviceFault(t *testing.T) {
	t.Run("Test_mergeDeviceFault", func(t *testing.T) {
		npuName := "Ascend910-0"
		split := []constant.DeviceFault{
			{
				FaultType:            constant.CardUnhealthy,
				NPUName:              npuName,
				FaultCode:            "0x1",
				FaultLevel:           constant.NotHandleFault,
				LargeModelFaultLevel: constant.NotHandleFault,
				FaultHandling:        constant.NotHandleFault,
				FaultTimeAndLevelMap: map[string]constant.FaultTimeAndLevel{
					"0x1": {FaultLevel: constant.NotHandleFault, FaultTime: 1},
				},
			},
			{
				FaultType:            constant.CardNetworkUnhealthy,
				NPUName:              npuName,
				FaultCode:            "0x2",
				FaultLevel:           constant.SubHealthFault,
				LargeModelFaultLevel: constant.SubHealthFault,
				FaultHandling:        constant.SubHealthFault,
				FaultTimeAndLevelMap: map[string]constant.FaultTimeAndLevel{
					"0x2": {FaultLevel: constant.SubHealthFault, FaultTime: 1},
				},
			},
		}
		got, err := mergeDeviceFault(split)
		if err != nil {
			t.Errorf("mergeDeviceFault() error = %v", err)
		}
		sort.Slice(got, func(i, j int) bool {
			return got[i].FaultType > got[j].FaultType
		})
		if !reflect.DeepEqual(got, split) {
			t.Errorf("mergeDeviceFault() got = %v, want %v", util.ObjToString(got), util.ObjToString(split))
		}
	})
}

// TestMergeManuallySeparateNPUTypeDeviceFault should combine other fault info and constant.ManuallySeparateNPU.
func TestMergeManuallySeparateNPUTypeDeviceFault(t *testing.T) {
	t.Run("TestMergeManuallySeparateNPUTypeDeviceFault", func(t *testing.T) {
		npuName := "Ascend910-0"
		split := []constant.DeviceFault{
			{
				FaultType:            constant.CardUnhealthy,
				NPUName:              npuName,
				FaultCode:            constant.ManuallySeparateNPU,
				FaultLevel:           constant.ManuallySeparateNPU,
				FaultTimeAndLevelMap: map[string]constant.FaultTimeAndLevel{},
			},
			{
				FaultType:  constant.CardUnhealthy,
				NPUName:    npuName,
				FaultCode:  constant.UceFaultCode,
				FaultLevel: constant.RestartBusiness,
				FaultTimeAndLevelMap: map[string]constant.FaultTimeAndLevel{
					constant.UceFaultCode: {
						FaultTime:  constant.UnknownFaultTime,
						FaultLevel: constant.RestartBusiness,
					},
				},
			},
		}
		got, err := mergeDeviceFault(split)
		if err != nil {
			t.Errorf("mergeDeviceFault() error = %v", err)
		}
		want := []constant.DeviceFault{
			{
				FaultType:            constant.CardUnhealthy,
				NPUName:              npuName,
				FaultCode:            constant.UceFaultCode,
				FaultLevel:           constant.ManuallySeparateNPU,
				LargeModelFaultLevel: constant.ManuallySeparateNPU,
				FaultHandling:        constant.ManuallySeparateNPU,
				FaultTimeAndLevelMap: map[string]constant.FaultTimeAndLevel{
					constant.UceFaultCode: {
						FaultTime:  constant.UnknownFaultTime,
						FaultLevel: constant.RestartBusiness,
					},
				},
			},
		}
		if !reflect.DeepEqual(got, want) {
			t.Error("TestMergeManuallySeparateNPUTypeDeviceFault fail")
		}
	})
}

// TestGetAdvanceDeviceCm should get advanceDeviceCm from originalDeviceCm
func TestGetAdvanceDeviceCm(t *testing.T) {
	advanceDeviceCm := GetAdvanceDeviceCm(originalDeviceCm)
	if len(advanceDeviceCm.FaultDeviceList[deviceName]) != originalDeviceFaultCodeCnt {
		t.Errorf("TestGetAdvanceDeviceCm failed")
		return
	}
	faultTimeAndLevel, ok := advanceDeviceCm.FaultDeviceList[deviceName][0].FaultTimeAndLevelMap[constant.UceFaultCode]
	if !ok || faultTimeAndLevel.FaultTime != time100Seconds ||
		faultTimeAndLevel.FaultLevel != constant.RestartBusiness {
		t.Errorf("TestGetAdvanceDeviceCm failed")
		return
	}
}

// TestValidBusinessUceReportInfo valid business uce report info
func TestValidBusinessUceReportInfo(t *testing.T) {
	t.Run("TestValidBusinessUceReportInfo", func(t *testing.T) {
		reportInfo := &constant.ReportInfo{
			RecoverTime:  time100Seconds - time1Seconds,
			CompleteTime: 0,
		}
		mockTime := time.Time{}
		mockUnixMilli := gomonkey.ApplyPrivateMethod(mockTime, "UnixMilli", func() int64 {
			return time100Seconds
		})
		mockNow := gomonkey.ApplyFunc(time.Now, func() time.Time {
			return mockTime
		})
		defer func() {
			mockNow.Reset()
			mockUnixMilli.Reset()
		}()
		if !ValidBusinessUceReportInfo(reportInfo) {
			t.Error("TestValidBusinessUceReportInfo fail")
		}
		reportInfo.RecoverTime = 0
		if ValidBusinessUceReportInfo(reportInfo) {
			t.Error("TestValidBusinessUceReportInfo fail")
		}
	})
}

// TestCanDoStepRetry check uceDeviceInfo can do step retry
func TestCanDoStepRetry(t *testing.T) {
	uceDeviceInfo := &constant.UceDeviceInfo{
		DeviceName:   deviceName,
		FaultTime:    time100Seconds,
		RecoverTime:  time100Seconds + time1Seconds,
		CompleteTime: 0,
	}
	t.Run("TestCanDoStepRetry", func(t *testing.T) {
		mockTime := time.Time{}
		mockUnixMilli := gomonkey.ApplyPrivateMethod(mockTime, "UnixMilli", func() int64 {
			return time120Seconds
		})
		mockNow := gomonkey.ApplyFunc(time.Now, func() time.Time {
			return mockTime
		})
		defer func() {
			mockNow.Reset()
			mockUnixMilli.Reset()
		}()
		if !CanDoStepRetry(uceDeviceInfo) {
			t.Error("TestCanDoStepRetry fail")
		}
	})
}

// TestGetContainedElementIdx should return id of the item from slice
func TestGetContainedElementIdx(t *testing.T) {
	arr := []string{"1", "2"}
	t.Run("TestGetContainedElementIdx", func(t *testing.T) {
		if got := GetContainedElementIdx("1", arr); got != 0 {
			t.Error("GetContainedElementIdx() fail")
		}
		if got := GetContainedElementIdx("3", arr); got != -1 {
			t.Error("GetContainedElementIdx() fail")
		}
	})
}

// TestGetFaultTime should return fault time from DeviceFault
func TestGetFaultTime(t *testing.T) {
	fault := constant.DeviceFault{
		FaultCode: constant.UceFaultCode,
		FaultTimeAndLevelMap: map[string]constant.FaultTimeAndLevel{
			constant.UceFaultCode: {
				FaultTime:  time100Seconds,
				FaultLevel: constant.RestartBusiness,
			},
		},
	}
	t.Run("TestGetFaultTime", func(t *testing.T) {
		if got := GetFaultTime(fault, ""); got != time100Seconds {
			t.Error("GetFaultTime fail")
		}
		fault.FaultTimeAndLevelMap = make(map[string]constant.FaultTimeAndLevel)
		if got := GetFaultTime(fault, ""); got != constant.DeviceNotFault {
			t.Error("GetFaultTime fail")
		}
	})
}

// TestFaultCodeJudge check fault code is right
func TestFaultCodeJudge(t *testing.T) {
	t.Run("TestFaultCodeJudgeAic", func(t *testing.T) {
		if got := IsUceAccompanyFault(constant.AicFaultCode); got == false {
			t.Error("TestFaultCodeJudgeAic fail")
		}
	})
	t.Run("TestFaultCodeJudgeLinkDownFault", func(t *testing.T) {
		if got := IsLinkDownFault(constant.LinkDownFaultCode); got == false {
			t.Error("TestFaultCodeJudgeLinkDownFault fail")
		}
	})
	t.Run("TestFaultCodeJudgeCqeFault", func(t *testing.T) {
		if got := IsCqeFault(constant.DevCqeFaultCode); got == false {
			t.Error("TestFaultCodeJudgeCqeFault fail")
		}
	})
	t.Run("TestFaultCodeJudgeUceFault", func(t *testing.T) {
		if got := IsUceFault(constant.UceFaultCode); got == false {
			t.Error("TestFaultCodeJudgeUceFault fail")
		}
	})
}

// TestGetAdvanceDeviceCmForNodeMap should get AdvanceDeviceCm
func TestGetAdvanceDeviceCmForNodeMap(t *testing.T) {
	deviceInfoCms := map[string]*constant.DeviceInfo{
		cmName: originalDeviceCm,
	}
	t.Run("TestGetAdvanceDeviceConfigmap", func(t *testing.T) {
		got := GetAdvanceFaultCm[*constant.AdvanceDeviceFaultCm](deviceInfoCms)
		if len(got[nodeName].FaultDeviceList[deviceName]) != originalDeviceFaultCodeCnt {
			t.Error("TestGetAdvanceDeviceConfigmap fail")
		}
	})
}

// TestGetNodeAndDeviceFromJobIdAndRankId should return right node and device according to the jobId and rankID
func TestGetNodeAndDeviceFromJobIdAndRankId(t *testing.T) {
	t.Run("TestGetNodeAndDeviceFromJobIdAndRankId", func(t *testing.T) {
		serverName, device, err := GetNodeAndDeviceFromJobIdAndRankId(jobId, rankID, jobServerMap)
		if serverName != nodeName || device != deviceId || err != nil {
			t.Error("TestGetNodeAndDeviceFromJobIdAndRankId fail")
		}
	})
}

// TestIsNodeReady check node is ready
func TestIsNodeReady(t *testing.T) {
	node := &v1.Node{
		Status: v1.NodeStatus{
			Conditions: []v1.NodeCondition{{
				Type:   v1.NodeReady,
				Status: v1.ConditionTrue,
			}},
		},
	}
	t.Run("TestIsNodeReady", func(t *testing.T) {
		if !IsNodeReady(node) {
			t.Error("TestIsNodeReady fail")
		}
	})
}

func getCardNotHandleFaults() []constant.DeviceFault {
	return append([]constant.DeviceFault{}, constant.DeviceFault{
		FaultType:  constant.CardUnhealthy,
		FaultLevel: constant.NotHandleFault,
	})
}

func getCardNotHandleAndPublicFaultSeparateNPU() []constant.DeviceFault {
	return append([]constant.DeviceFault{}, constant.DeviceFault{
		FaultType:  constant.CardUnhealthy,
		FaultLevel: constant.NotHandleFault,
	}, constant.DeviceFault{
		FaultType:  constant.PublicFaultType,
		FaultLevel: constant.SeparateNPU,
	})
}

func getCardNetworkNotHandleAndPublicFaultSeparateNPU() []constant.DeviceFault {
	return append([]constant.DeviceFault{}, constant.DeviceFault{
		FaultType:  constant.CardNetworkUnhealthy,
		FaultLevel: constant.NotHandleFault,
	}, constant.DeviceFault{
		FaultType:  constant.PublicFaultType,
		FaultLevel: constant.SeparateNPU,
	})
}

func getCardNotHandleAndPublicFaultSubHealth() []constant.DeviceFault {
	return append([]constant.DeviceFault{}, constant.DeviceFault{
		FaultType:  constant.CardUnhealthy,
		FaultLevel: constant.NotHandleFault,
	}, constant.DeviceFault{
		FaultType:  constant.PublicFaultType,
		FaultLevel: constant.SubHealthFault,
	})
}

// TestIsFaultDeletable check faults in specified fault type are NotHandleFault and SubHealthFault
func TestIsFaultDeletable(t *testing.T) {
	t.Run("TestIsFaultDeletable", func(t *testing.T) {
		deletableFaultLevels := []string{constant.NotHandleFault, constant.SubHealthFault}
		faults := getCardNotHandleFaults()
		if !isFaultDeletable(faults, []string{constant.CardUnhealthy}, deletableFaultLevels) {
			t.Error("when only NotHandleFault in CardUnhealthy then should remove")
		}
		if !isFaultDeletable(faults, []string{constant.CardNetworkUnhealthy}, deletableFaultLevels) {
			t.Error("when no fault in CardNetworkUnhealthy then should remove from CardNetworkUnhealthy")
		}

		faults = getCardNotHandleAndPublicFaultSeparateNPU()
		if isFaultDeletable(faults, []string{constant.CardUnhealthy, constant.PublicFaultType}, deletableFaultLevels) {
			t.Error("when PublicFaultType is SeparateNPU then should not remove from CardUnhealthy")
		}

		faults = getCardNetworkNotHandleAndPublicFaultSeparateNPU()
		if !isFaultDeletable(faults, []string{constant.CardNetworkUnhealthy}, deletableFaultLevels) {
			t.Error("when PublicFaultType is SeparateNPU and CardNetworkUnhealthy is NotHandleFault " +
				"then should remove from CardNetworkUnhealthy")
		}

		faults = append([]constant.DeviceFault{}, constant.DeviceFault{FaultType: constant.PublicFaultType})
		if isFaultDeletable(faults, []string{constant.CardUnhealthy, constant.PublicFaultType}, deletableFaultLevels) {
			t.Error("when PublicFaultType is SeparateNPU then should not remove from CardUnhealthy")
		}
		faults = make([]constant.DeviceFault, 0)
		if !isFaultDeletable(faults, []string{constant.CardUnhealthy, constant.PublicFaultType}, deletableFaultLevels) {
			t.Error("when no faults then should remove from CardUnhealthy")
		}
		faults = getCardNotHandleAndPublicFaultSubHealth()
		if !isFaultDeletable(faults, []string{constant.CardUnhealthy, constant.PublicFaultType}, deletableFaultLevels) {
			t.Error("when SubHealthFault and NotHandleFault faults then should remove from CardUnhealthy")
		}
	})
}

// TestGetAdvanceFaultForNode test get advance fault info for node
func TestGetAdvanceFaultForNode(t *testing.T) {
	patches := gomonkey.NewPatches()
	defer patches.Reset()

	convey.Convey("Test GetAdvanceFaultForNode", t, func() {
		convey.Convey("Case 1: DeviceInfo input", func() {
			deviceInfo := &constant.DeviceInfo{}
			expected := &constant.AdvanceDeviceFaultCm{}

			patches.ApplyFunc(GetAdvanceDeviceCm, func(*constant.DeviceInfo) *constant.AdvanceDeviceFaultCm {
				return expected
			})

			result := GetAdvanceFaultForNode(deviceInfo)
			convey.So(result, convey.ShouldEqual, expected)
		})

		convey.Convey("Case 2: NodeInfo input", func() {
			nodeInfo := &constant.NodeInfo{}
			result := GetAdvanceFaultForNode(nodeInfo)
			convey.So(result, convey.ShouldEqual, nodeInfo)
		})

		convey.Convey("Case 3: SwitchInfo input", func() {
			switchInfo := &constant.SwitchInfo{}
			result := GetAdvanceFaultForNode(switchInfo)
			convey.So(result, convey.ShouldEqual, switchInfo)
		})

		convey.Convey("Case 4: AdvanceDeviceFaultCm input", func() {
			advanceCm := &constant.AdvanceDeviceFaultCm{}
			result := GetAdvanceFaultForNode(advanceCm)
			convey.So(result, convey.ShouldEqual, advanceCm)
		})
	})
}

func TestAdvanceFaultCmToOriginalFaultCm(t *testing.T) {
	convey.Convey("Test AdvanceFaultCmToOriginalFaultCm", t, func() {
		node1 := "node1"
		node2 := "node2"
		mockAdvanceCm1 := &constant.AdvanceDeviceFaultCm{
			DeviceType:          "",
			CmName:              "CmName-" + node1,
			SuperPodID:          0,
			ServerIndex:         0,
			FaultDeviceList:     make(map[string][]constant.DeviceFault),
			AvailableDeviceList: []string{"xxx"},
			Recovering:          []string{"xxx"},
			CardUnHealthy:       []string{"xxx"},
			NetworkUnhealthy:    []string{"xxx"},
			UpdateTime:          0,
		}
		mockAdvanceCm2 := new(constant.AdvanceDeviceFaultCm)
		util.DeepCopy(mockAdvanceCm1, mockAdvanceCm2)
		mockAdvanceCm2.CmName = "CmName-" + node2

		convey.Convey("should convert map correctly", func() {
			input := map[string]constant.ConfigMapInterface{
				node1: mockAdvanceCm1,
				node2: mockAdvanceCm2,
			}

			result := AdvanceFaultMapToOriginalFaultMap[*constant.DeviceInfo](input)

			convey.So(len(result), convey.ShouldEqual, len(input))
		})
	})
}

func TestGetSortedKeys(t *testing.T) {
	patches := gomonkey.NewPatches()
	defer patches.Reset()

	convey.Convey("Test getSortedKeys", t, func() {
		convey.Convey("should return sorted keys for string-struct map", func() {
			type testStruct struct{ val int }
			input := map[string]testStruct{
				"zebra": {1},
				"lion":  {2},
				"ape":   {3},
			}
			expected := []string{"ape", "lion", "zebra"}
			result := getSortedKeys(input)
			convey.So(result, convey.ShouldResemble, expected)
		})

		convey.Convey("should handle empty map", func() {
			input := map[string]float64{}
			result := getSortedKeys(input)
			convey.So(result, convey.ShouldBeEmpty)
		})
	})
}

func TestCompareFaultTimeAndLevel(t *testing.T) {
	convey.Convey("should compare by FaultTime first", t, func() {
		a := constant.FaultTimeAndLevel{FaultTime: 100, FaultLevel: "low"}
		b := constant.FaultTimeAndLevel{FaultTime: 200, FaultLevel: "high"}
		convey.So(compareFaultTimeAndLevel(a, b), convey.ShouldBeLessThan, 0)
	})

	convey.Convey("should compare by FaultLevel when FaultTime equal", t, func() {
		a := constant.FaultTimeAndLevel{FaultTime: 100, FaultLevel: "low"}
		b := constant.FaultTimeAndLevel{FaultTime: 100, FaultLevel: "high"}
		convey.So(compareFaultTimeAndLevel(a, b), convey.ShouldBeGreaterThan, 0)
	})

	convey.Convey("should return 0 when both equal", t, func() {
		a := constant.FaultTimeAndLevel{FaultTime: 100, FaultLevel: "medium"}
		b := constant.FaultTimeAndLevel{FaultTime: 100, FaultLevel: "medium"}
		convey.So(compareFaultTimeAndLevel(a, b), convey.ShouldEqual, 0)
	})
}

func TestCompareDeviceFault(t *testing.T) {
	convey.Convey("Test compareDeviceFault", t, func() {
		baseFault := constant.DeviceFault{
			FaultType:            constant.CardUnhealthy,
			NPUName:              "npu0",
			LargeModelFaultLevel: constant.SubHealthFault,
			FaultLevel:           constant.SubHealthFault,
			FaultHandling:        constant.SubHealthFault,
			FaultCode:            constant.AicFaultCode,
			FaultTimeAndLevelMap: map[string]constant.FaultTimeAndLevel{
				constant.AicFaultCode: {
					FaultTime:  0,
					FaultLevel: constant.SubHealthFault,
				},
			},
		}

		convey.Convey("should compare by FaultType first", func() {
			f1 := baseFault
			f2 := baseFault
			convey.So(compareDeviceFault(f1, f2), convey.ShouldEqual, 0)
		})

		convey.Convey("should compare by NPUName when FaultType equal", func() {
			f1 := baseFault
			f2 := baseFault
			f2.NPUName = "npu1"
			convey.So(compareDeviceFault(f1, f2), convey.ShouldBeLessThan, 0)
		})

		convey.Convey("should compare FaultTimeAndLevelMap when all fields equal", func() {
			f1 := baseFault
			f1.FaultTimeAndLevelMap = map[string]constant.FaultTimeAndLevel{
				"key1": {FaultTime: 100, FaultLevel: "low"},
			}
			f2 := baseFault
			f2.FaultTimeAndLevelMap = map[string]constant.FaultTimeAndLevel{
				"key1": {FaultTime: 200, FaultLevel: "low"},
			}
			convey.So(compareDeviceFault(f1, f2), convey.ShouldBeLessThan, 0)
		})
	})
}

func TestSortDataForAdvanceDeviceInfo(t *testing.T) {
	deviceInfo := &constant.AdvanceDeviceFaultCm{
		AvailableDeviceList: []string{"d3", "d1", "d2"},
		CardUnHealthy:       []string{"c2", "c1"},
		NetworkUnhealthy:    []string{"n2", "n1"},
		Recovering:          []string{"r2", "r1"},
		FaultDeviceList: map[string][]constant.DeviceFault{
			"list1": {{FaultType: "typeB"}, {FaultType: "typeA"}},
		},
	}

	expDeviceInfo := &constant.AdvanceDeviceFaultCm{
		AvailableDeviceList: []string{"d1", "d2", "d3"},
		CardUnHealthy:       []string{"c1", "c2"},
		NetworkUnhealthy:    []string{"n1", "n2"},
		Recovering:          []string{"r1", "r2"},
		FaultDeviceList: map[string][]constant.DeviceFault{
			"list1": {{FaultType: "typeA"}, {FaultType: "typeB"}},
		},
	}
	convey.Convey("should be equal", t, func() {
		SortDataForAdvanceDeviceInfo(deviceInfo)
		convey.So(deviceInfo, convey.ShouldResemble, expDeviceInfo)
	})
}

func TestMergeCode(t *testing.T) {
	patches := gomonkey.NewPatches()
	defer patches.Reset()
	convey.Convey("Test mergeCode", t, func() {
		testDevice := "device1"
		orgFaults := []constant.DeviceFault{{FaultCode: "1001"}}
		mergedFaults := []constant.DeviceFault{{FaultCode: "merged"}}
		advanceCm := &constant.AdvanceDeviceFaultCm{
			FaultDeviceList: map[string][]constant.DeviceFault{
				testDevice: orgFaults,
			},
		}
		patches.ApplyFunc(mergeDeviceFault, func([]constant.DeviceFault) ([]constant.DeviceFault, error) {
			return mergedFaults, nil
		})
		convey.Convey("should skip empty fault lists", func() {
			mergeCode(advanceCm)
			convey.So(advanceCm.FaultDeviceList[testDevice], convey.ShouldResemble, mergedFaults)
		})
	})
}
