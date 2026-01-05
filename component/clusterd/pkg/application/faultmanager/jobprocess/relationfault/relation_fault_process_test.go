// Copyright (c) Huawei Technologies Co., Ltd. 2024-2025. All rights reserved.

// Package relationfault contain relation fault process
package relationfault

import (
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/pkg/errors"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/apimachinery/pkg/util/sets"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
	"clusterd/pkg/domain/faultdomain"
	"clusterd/pkg/domain/job"
	"clusterd/pkg/interface/kube"
)

const (
	node101         = "node-101"
	node102         = "node-102"
	node100         = "node-100"
	node103         = "node-103"
	npu1            = "Ascend-910-1"
	npu2            = "Ascend-910-2"
	switchDevice    = "FF"
	deviceFault     = "deviceFault"
	faultCode0002   = "0x0002"
	faultCode0003   = "0x0003"
	faultCode0001   = "0x0001"
	faultCode0004   = "0x0004"
	triggerCode0001 = "0x0001"
	triggerCode0002 = "0x0002"
	timeOutInterval = 1000
)

func TestNoTriggerNetworkFault(t *testing.T) {
	t.Run("NoTrigger", func(t *testing.T) {
		relationFaults := []string{faultCode0002, faultCode0003}
		strategyList := make([]constant.RelationFaultStrategy, 0)
		strategyList = append(strategyList, constant.RelationFaultStrategy{
			TriggerFault:   triggerCode0001,
			RelationFaults: relationFaults,
			FaultStrategy:  constant.SeparateFaultStrategy,
		})
		networkFaults := make([]*constant.FaultInfo, 0)
		fault1 := constant.FaultInfo{
			NodeName:  node101,
			NPUName:   npu1,
			FaultType: deviceFault,
			FaultCode: faultCode0002,
		}
		fault2 := constant.FaultInfo{
			NodeName:  node101,
			NPUName:   npu2,
			FaultType: deviceFault,
			FaultCode: faultCode0003,
		}
		networkFaults = append(networkFaults, &fault1, &fault2)
		triggerList := make([]constant.FaultInfo, 0)
		faultJob := FaultJob{
			FindNPUUnderSwitch: false,
		}
		relationFaultStrategies = strategyList
		faultJob.RelationFaults = networkFaults
		faultJob.TriggerFault = triggerList
		faultJob.NodeFaultInfoMap = make(map[string][]*constant.FaultInfo)
		faultJob.processNetworkFault()
		if len(faultJob.FaultStrategy.DeviceLvList) != 0 {
			t.Errorf("FaultStrategy is wrong:")
			t.Log(util.ObjToString(faultJob.FaultStrategy))
		}
	})
}

func TestNotAllNetworkInRelationFaults(t *testing.T) {
	t.Run("NotAllNetworkInRelationFaults", func(t *testing.T) {
		strategyList := make([]constant.RelationFaultStrategy, 0)
		strategyList = append(strategyList, constant.RelationFaultStrategy{
			TriggerFault:   triggerCode0001,
			RelationFaults: []string{faultCode0002, faultCode0003},
			FaultStrategy:  constant.SeparateFaultStrategy,
		})
		networkFaults := make([]*constant.FaultInfo, 0)
		fault1 := constant.FaultInfo{
			NodeName:  node100,
			NPUName:   npu1,
			FaultType: deviceFault,
			FaultCode: faultCode0002,
		}
		fault2 := constant.FaultInfo{
			NodeName:  node101,
			NPUName:   npu2,
			FaultType: deviceFault,
			FaultCode: faultCode0004,
		}
		networkFaults = append(networkFaults, &fault1, &fault2)
		triggerList := make([]constant.FaultInfo, 0)
		retryEvent := constant.FaultInfo{
			NodeName:  node102,
			NPUName:   npu2,
			FaultType: deviceFault,
			FaultCode: faultCode0001,
		}
		triggerList = append(triggerList, retryEvent)
		faultJob := FaultJob{
			FindNPUUnderSwitch: false,
		}
		relationFaultStrategies = strategyList
		faultJob.RelationFaults = networkFaults
		faultJob.TriggerFault = triggerList
		faultJob.NodeFaultInfoMap = make(map[string][]*constant.FaultInfo)
		faultJob.processNetworkFault()
		if len(faultJob.FaultStrategy.DeviceLvList) != 0 {
			t.Errorf("FaultStrategy is wrong:")
			t.Log(util.ObjToString(faultJob.FaultStrategy))
		}
	})
}

func equalDeviceStrategy(bm, test map[string][]constant.DeviceStrategy) bool {
	if len(bm) != len(test) {
		return false
	}
	for k, v := range bm {
		if len(v) != len(test[k]) {
			return false
		}
		for i, s := range test[k] {
			if s.NPUName != bm[k][i].NPUName || s.Strategy != bm[k][i].Strategy {
				return false
			}
		}
	}
	return true
}

func equalNodeStrategy(bm, test map[string]string) bool {
	if len(bm) != len(test) {
		return false
	}
	for k, v := range bm {
		if len(v) != len(test[k]) {
			return false
		}
		if bm[k] != test[k] {
			return false
		}
	}
	return true
}

func TestRightNodeAndDeviceSeparate(t *testing.T) {
	t.Run("rightNodeAndDeviceSeparate", func(t *testing.T) {
		strategyList := make([]constant.RelationFaultStrategy, 0)
		strategyList = append(strategyList, constant.RelationFaultStrategy{
			TriggerFault: triggerCode0001, RelationFaults: []string{faultCode0002, faultCode0003},
			FaultStrategy: constant.SeparateFaultStrategy,
		})
		networkFaults := make([]*constant.FaultInfo, 0)
		fault1 := constant.FaultInfo{NodeName: node100, NPUName: npu1, FaultType: deviceFault, FaultCode: faultCode0002}
		fault2 := constant.FaultInfo{NodeName: node101, NPUName: npu2, FaultType: deviceFault, FaultCode: faultCode0003}
		fault3 := constant.FaultInfo{NodeName: node103, NPUName: switchDevice,
			FaultType: constant.SwitchFault, FaultCode: faultCode0003}
		networkFaults = append(networkFaults, &fault1, &fault2, &fault3)

		triggerList := make([]constant.FaultInfo, 0)
		retryEvent := constant.FaultInfo{NPUName: npu2, FaultType: deviceFault, FaultCode: triggerCode0001}
		triggerList = append(triggerList, retryEvent)

		faultJob := FaultJob{FindNPUUnderSwitch: false}
		relationFaultStrategies = strategyList
		faultJob.RelationFaults = networkFaults
		faultJob.TriggerFault = triggerList
		faultJob.NodeFaultInfoMap = make(map[string][]*constant.FaultInfo)
		faultJob.processNetworkFault()

		deviceList := map[string][]constant.DeviceStrategy{
			node100: {{NPUName: npu1, Strategy: constant.SeparateFaultStrategy}},
			node101: {{NPUName: npu2, Strategy: constant.SeparateFaultStrategy}},
			node103: {},
		}
		nodeList := map[string]string{node103: constant.SeparateFaultStrategy}
		if !equalDeviceStrategy(deviceList, faultJob.FaultStrategy.DeviceLvList) ||
			!equalNodeStrategy(nodeList, faultJob.FaultStrategy.NodeLvList) {
			t.Errorf("FaultStrategy is wrong:")
			t.Log(util.ObjToString(faultJob.FaultStrategy))
		}
	})
}

func TestRightNodeDeviceSubHealth(t *testing.T) {
	t.Run("device SubHealth Node Separate", func(t *testing.T) {
		strategyList := make([]constant.RelationFaultStrategy, 0)
		strategyList = append(strategyList, constant.RelationFaultStrategy{
			TriggerFault: triggerCode0001, RelationFaults: []string{faultCode0002, faultCode0003},
			FaultStrategy: constant.SubHealthFaultStrategy})
		strategyList = append(strategyList, constant.RelationFaultStrategy{
			TriggerFault: triggerCode0002, RelationFaults: []string{faultCode0004},
			FaultStrategy: constant.SeparateFaultStrategy,
		})
		networkFaults := make([]*constant.FaultInfo, 0)
		fault1 := constant.FaultInfo{NodeName: node100, NPUName: npu1, FaultType: deviceFault, FaultCode: faultCode0002}
		fault2 := constant.FaultInfo{NodeName: node101, NPUName: npu2, FaultType: deviceFault, FaultCode: faultCode0003}
		nodeFault := constant.FaultInfo{NodeName: node103, NPUName: switchDevice, FaultType: constant.SwitchFault,
			FaultCode: faultCode0004}
		networkFaults = append(networkFaults, &fault1, &fault2, &nodeFault)

		triggerList := make([]constant.FaultInfo, 0)
		trigger1 := constant.FaultInfo{NPUName: npu2, FaultType: deviceFault, FaultCode: triggerCode0001}
		trigger2 := constant.FaultInfo{NPUName: npu2, FaultType: deviceFault, FaultCode: triggerCode0002}
		triggerList = append(triggerList, trigger1, trigger2)

		faultJob := FaultJob{FindNPUUnderSwitch: false}
		relationFaultStrategies = strategyList
		faultJob.RelationFaults = networkFaults
		faultJob.TriggerFault = triggerList
		faultJob.NodeFaultInfoMap = make(map[string][]*constant.FaultInfo)
		faultJob.processNetworkFault()

		deviceList := map[string][]constant.DeviceStrategy{
			node100: {{NPUName: npu1, Strategy: constant.SubHealthFaultStrategy}},
			node101: {{NPUName: npu2, Strategy: constant.SubHealthFaultStrategy}},
			node103: {},
		}
		nodeList := map[string]string{node103: constant.SeparateFaultStrategy}
		if !equalDeviceStrategy(deviceList, faultJob.FaultStrategy.DeviceLvList) ||
			!equalNodeStrategy(nodeList, faultJob.FaultStrategy.NodeLvList) {
			t.Errorf("FaultStrategy is wrong:")
			t.Log(util.ObjToString(faultJob.FaultStrategy))
		}

	})
}
func TestRightNodeDeviceSubHealthAndSeparate(t *testing.T) {
	t.Run("rightNodeSeparateDeviceBothHasSub-healthAndSeparate", func(t *testing.T) {
		strategyList := make([]constant.RelationFaultStrategy, 0)
		strategyList = append(strategyList, constant.RelationFaultStrategy{
			TriggerFault: triggerCode0001, RelationFaults: []string{faultCode0002, faultCode0003},
			FaultStrategy: constant.SubHealthFaultStrategy,
		})
		strategyList = append(strategyList, constant.RelationFaultStrategy{
			TriggerFault: triggerCode0002, RelationFaults: []string{faultCode0004},
			FaultStrategy: constant.SeparateFaultStrategy,
		})
		networkFaults := make([]*constant.FaultInfo, 0)
		fault1 := constant.FaultInfo{NodeName: node100, NPUName: npu1, FaultType: deviceFault, FaultCode: faultCode0002}
		fault2 := constant.FaultInfo{NodeName: node101, NPUName: npu2, FaultType: deviceFault, FaultCode: faultCode0003}
		fault4 := constant.FaultInfo{NodeName: node100, NPUName: npu1, FaultType: deviceFault, FaultCode: faultCode0004}
		nodeFault := constant.FaultInfo{NodeName: node103, NPUName: switchDevice,
			FaultType: constant.SwitchFault, FaultCode: faultCode0004}

		networkFaults = append(networkFaults, &fault1, &fault2, &nodeFault, &fault4)
		triggerList := make([]constant.FaultInfo, 0)
		trigger1 := constant.FaultInfo{NPUName: npu2, FaultType: deviceFault, FaultCode: triggerCode0001}
		trigger2 := constant.FaultInfo{NPUName: npu2, FaultType: deviceFault, FaultCode: triggerCode0002}
		triggerList = append(triggerList, trigger1, trigger2)

		faultJob := FaultJob{
			FindNPUUnderSwitch: false,
		}
		relationFaultStrategies = strategyList
		faultJob.RelationFaults = networkFaults
		faultJob.TriggerFault = triggerList
		faultJob.NodeFaultInfoMap = make(map[string][]*constant.FaultInfo)
		faultJob.processNetworkFault()

		deviceList := map[string][]constant.DeviceStrategy{
			node100: {{NPUName: npu1, Strategy: constant.SeparateFaultStrategy}},
			node101: {{NPUName: npu2, Strategy: constant.SubHealthFaultStrategy}},
			node103: {},
		}
		nodeList := map[string]string{node103: constant.SeparateFaultStrategy}
		if !equalDeviceStrategy(deviceList, faultJob.FaultStrategy.DeviceLvList) ||
			!equalNodeStrategy(nodeList, faultJob.FaultStrategy.NodeLvList) {
			t.Errorf("FaultStrategy is wrong:")
			t.Log(util.ObjToString(faultJob.FaultStrategy))
		}

	})
}

func TestGetPodStrategiesMapsByJobId(t *testing.T) {
	convey.Convey("Test getPodStrategiesMapsByJobId", t, func() {
		fJobCenter := &relationFaultProcessor{
			faultJobs: map[string]*FaultJob{
				"job1": {PodStrategiesMaps: map[string]string{
					"pod1": constant.SeparateFaultStrategy,
				}},
			}}
		convey.Convey("01-job id not in map keys, should return nil", func() {
			mockJobId := "job2"
			resultMap := fJobCenter.GetPodStrategiesMapsByJobId(mockJobId)
			convey.So(resultMap, convey.ShouldBeNil)
		})
		convey.Convey("02-job id in map keys, should return map", func() {
			mockJobId := "job1"
			resultMap := fJobCenter.GetPodStrategiesMapsByJobId(mockJobId)
			convey.So(resultMap, convey.ShouldNotBeNil)
			mockPod := "pod1"
			convey.So(resultMap[mockPod] == constant.SeparateFaultStrategy, convey.ShouldBeTrue)
			mockPod = "pod2"
			convey.So(resultMap[mockPod] == "", convey.ShouldBeTrue)
		})
	})

}

func TestProcess(t *testing.T) {
	convey.Convey("Test Process", t, func() {
		processor := &relationFaultProcessor{}

		testInvalidInfoType(processor)
		testValidInfoType(processor)
	})
}

func testInvalidInfoType(processor *relationFaultProcessor) {
	convey.Convey("When info type is invalid", func() {
		patches := gomonkey.ApplyFunc(hwlog.RunLog.Errorf, func(format string, args ...interface{}) {})
		defer patches.Reset()

		info := "invalid-info-type"
		result := processor.Process(info)
		convey.So(result, convey.ShouldEqual, info)
	})
}

func testValidInfoType(processor *relationFaultProcessor) {
	convey.Convey("When info type is valid", func() {
		content := constant.AllConfigmapContent{
			DeviceCm: map[string]*constant.AdvanceDeviceFaultCm{},
			SwitchCm: map[string]*constant.SwitchInfo{},
			NodeCm:   map[string]*constant.NodeInfo{},
		}

		patches := gomonkey.ApplyMethod(processor, "InitFaultJobs", func(_ *relationFaultProcessor) {})
		defer patches.Reset()

		processor.faultJobs = map[string]*FaultJob{"test": &FaultJob{}}

		result := processor.Process(content)
		convey.So(result, convey.ShouldBeNil)
		convey.So(processor.deviceInfoCm, convey.ShouldResemble, content.DeviceCm)
		convey.So(processor.switchInfoCm, convey.ShouldResemble, content.SwitchCm)
		convey.So(processor.nodeInfoCm, convey.ShouldResemble, content.NodeCm)
	})
}

func TestFaultJobMethods(t *testing.T) {
	convey.Convey("Test FaultJob methods", t, func() {
		fJob := &FaultJob{
			AllFaultCode:        sets.NewString(),
			ProcessingFaultCode: sets.NewString(),
			RelationFaults:      []*constant.FaultInfo{},
			PodNames:            make(map[string]string),
			PodStrategiesMaps:   make(map[string]string),
			NodeFaultInfoMap:    map[string][]*constant.FaultInfo{},
		}

		testInitFaultJobAttr(fJob)
		testPreStartProcess(fJob)
		testPreStopProcess(fJob)
		testInitByDeviceFault(fJob)
	})
}

func testInitFaultJobAttr(fJob *FaultJob) {
	convey.Convey("When initializing FaultJob attributes", func() {
		fJob.initFaultJobAttr()
		convey.So(fJob.FaultStrategy, convey.ShouldResemble, constant.FaultStrategy{})
		convey.So(fJob.TriggerFault, convey.ShouldBeNil)
		convey.So(fJob.AllFaultCode, convey.ShouldNotBeNil)
		convey.So(fJob.SeparateNodes, convey.ShouldNotBeNil)
		convey.So(fJob.PodNames, convey.ShouldNotBeNil)
		convey.So(fJob.ProcessingFaultCode, convey.ShouldNotBeNil)
		convey.So(fJob.PodStrategiesMaps, convey.ShouldNotBeNil)
	})
}

func testPreStartProcess(fJob *FaultJob) {
	convey.Convey("When running preStartProcess", func() {
		fJob.AllFaultCode.Insert("fault1")
		fJob.RelationFaults = []*constant.FaultInfo{
			{FaultUid: "fault1"},
			{FaultUid: "fault2"},
		}

		patches := gomonkey.ApplyFunc(hwlog.RunLog.Infof, func(format string, args ...interface{}) {})
		defer patches.Reset()
		patches.ApplyFunc(util.ObjToString, func(obj interface{}) string {
			return "test-fault-info"
		})

		fJob.preStartProcess()
		convey.So(fJob.RelationFaults, convey.ShouldHaveLength, 1)
		convey.So(fJob.TMOutRelationFaults, convey.ShouldHaveLength, 1)
		convey.So(fJob.RelationFaults[0].FaultUid, convey.ShouldEqual, "fault1")
		convey.So(fJob.TMOutRelationFaults[0].FaultUid, convey.ShouldEqual, "fault2")
	})
}

func testPreStopProcess(fJob *FaultJob) {
	convey.Convey("When running preStopProcess", func() {
		fJob.RelationFaults = []*constant.FaultInfo{
			{FaultUid: "fault1", ExecutedStrategy: constant.SeparateFaultStrategy},
			{FaultUid: "fault2", FaultTime: time.Now().UnixMilli() - 10000, DealMaxTime: 1},
		}

		patches := gomonkey.ApplyFunc(hwlog.RunLog.Infof, func(format string, args ...interface{}) {})
		defer patches.Reset()

		fJob.preStopProcess()
		convey.So(fJob.RelationFaults, convey.ShouldHaveLength, 0)
	})
}

func TestProcessFaultStrategies(t *testing.T) {
	convey.Convey("Test processFaultStrategies", t, func() {
		fJob := &FaultJob{
			FaultStrategy: constant.FaultStrategy{
				DeviceLvList: map[string][]constant.DeviceStrategy{
					"node1": {{Strategy: "strategy1"}},
				},
				NodeLvList: map[string]string{},
			},
			PodNames:          map[string]string{"node1": "pod1"},
			PodStrategiesMaps: map[string]string{},
			NameSpace:         "test-namespace",
			NodeFaultInfoMap:  map[string][]*constant.FaultInfo{},
		}

		testDeepCopyError(fJob)
		testPatchPodLabelError(fJob)
		testSuccess(fJob)
	})
}

func testDeepCopyError(fJob *FaultJob) {
	convey.Convey("When deep copy fails", func() {
		patches := gomonkey.ApplyFunc(util.DeepCopy, func(dst, src interface{}) error {
			return errors.New("deep copy error")
		})
		defer patches.Reset()
		patches.ApplyFunc(hwlog.RunLog.Errorf, func(format string, args ...interface{}) {})

		fJob.processFaultStrategies()
		convey.So(fJob.PodStrategiesMaps, convey.ShouldBeEmpty)
	})
}

func testPatchPodLabelError(fJob *FaultJob) {
	convey.Convey("When patch pod label fails", func() {
		patches := gomonkey.ApplyFunc(util.DeepCopy, func(dst, src interface{}) error {
			return nil
		})
		defer patches.Reset()
		patches.ApplyFunc(kube.RetryPatchPodLabels, func(podName, namespace string, retryTimes int, labels map[string]string) error {
			return errors.New("patch pod label error")
		})
		patches.ApplyFunc(hwlog.RunLog.Errorf, func(format string, args ...interface{}) {})

		fJob.processFaultStrategies()
		convey.So(fJob.PodStrategiesMaps["pod1"], convey.ShouldEqual, "strategy1")
	})
}

func testSuccess(fJob *FaultJob) {
	convey.Convey("When all operations succeed", func() {
		patches := gomonkey.ApplyFunc(util.DeepCopy, func(dst, src interface{}) error {
			return nil
		})
		defer patches.Reset()
		patches.ApplyFunc(kube.RetryPatchPodLabels, func(podName, namespace string, retryTimes int, labels map[string]string) error {
			return nil
		})
		patches.ApplyFunc(hwlog.RunLog.Debugf, func(format string, args ...interface{}) {})

		fJob.processFaultStrategies()
		convey.So(fJob.PodStrategiesMaps["pod1"], convey.ShouldEqual, "strategy1")
	})
}

func testInitByDeviceFault(fJob *FaultJob) {
	convey.Convey("When initializing by device fault", func() {
		cardName := "server-type-device1"
		nodeFaultInfo := &constant.AdvanceDeviceFaultCm{
			DeviceType: "server-type",
			FaultDeviceList: map[string][]constant.DeviceFault{
				cardName: {
					{
						NPUName: "npu1",
						FaultTimeAndLevelMap: map[string]constant.FaultTimeAndLevel{
							"fault1": {FaultLevel: "level1"},
						},
					},
				},
			},
			CardUnHealthy: []string{cardName},
		}
		serverList := constant.ServerHccl{
			ServerName: "node1",
			DeviceList: []constant.Device{
				{DeviceID: "device1", RankID: "rank1"},
			},
		}

		fJob.initByDeviceFault(nodeFaultInfo, serverList)
	})
}

func TestInitFaultInfoByDeviceFault(t *testing.T) {
	convey.Convey("Test initFaultInfoByDeviceFault", t, func() {
		fJob := &FaultJob{
			AllFaultCode:        sets.NewString(),
			ProcessingFaultCode: sets.NewString(),
			RelationFaults:      []*constant.FaultInfo{},
			NodeFaultInfoMap:    map[string][]*constant.FaultInfo{},
		}

		testAssociateFault(fJob)
		testNonAssociateFault(fJob)
	})
}

func testAssociateFault(fJob *FaultJob) {
	convey.Convey("When fault is associate fault and card is healthy", func() {
		associateFault := "fault1"
		deviceName := "device1"
		nodeName := "node1"
		faultList := []constant.DeviceFault{
			{
				NPUName: deviceName,
				FaultTimeAndLevelMap: map[string]constant.FaultTimeAndLevel{
					associateFault: {FaultLevel: "level1"},
				},
			},
		}

		patches := gomonkey.ApplyFunc(isAssociateFault, func(faultCode string) bool {
			return faultCode == associateFault
		})
		defer patches.Reset()

		fJob.initFaultInfoByDeviceFault(faultList, nodeName, "", false)
		convey.So(fJob.AllFaultCode.Has(nodeName+"-"+deviceName+"-"+associateFault), convey.ShouldBeTrue)
	})
}

func testNonAssociateFault(fJob *FaultJob) {
	convey.Convey("When fault is not associate fault", func() {
		faultList := []constant.DeviceFault{
			{
				NPUName: "npu1",
				FaultTimeAndLevelMap: map[string]constant.FaultTimeAndLevel{
					"fault1": {FaultLevel: "level1"},
				},
				ForceAdd: false,
			},
		}

		patches := gomonkey.ApplyFunc(isAssociateFault, func(faultCode string) bool {
			return false
		})
		defer patches.Reset()

		fJob.initFaultInfoByDeviceFault(faultList, "node1", "rank1", false)
		convey.So(fJob.AllFaultCode.Has("node1-npu1-fault1"), convey.ShouldBeFalse)
		convey.So(fJob.RelationFaults, convey.ShouldHaveLength, 0)
	})
}

func TestInitBySwitchFault(t *testing.T) {
	convey.Convey("Test initBySwitchFault", t, func() {
		fJob := &FaultJob{
			AllFaultCode:        sets.NewString(),
			ProcessingFaultCode: sets.NewString(),
			RelationFaults:      []*constant.FaultInfo{},
			SeparateNodes:       sets.NewString(),
			NodeFaultInfoMap:    map[string][]*constant.FaultInfo{},
		}

		testNilSwitchInfo(fJob)
		testUnhealthyNode(fJob)
		testSwitchAssociateFault(fJob)
	})
}

func testNilSwitchInfo(fJob *FaultJob) {
	convey.Convey("When switch info is nil", func() {
		fJob.initBySwitchFault(nil, constant.ServerHccl{ServerName: "node1"})
		convey.So(fJob.SeparateNodes.Has("node1"), convey.ShouldBeFalse)
	})
}

func testUnhealthyNode(fJob *FaultJob) {
	convey.Convey("When node is unhealthy", func() {
		switchInfo := &constant.SwitchInfo{
			SwitchFaultInfo: constant.SwitchFaultInfo{NodeStatus: constant.UnHealthyState},
		}
		fJob.initBySwitchFault(switchInfo, constant.ServerHccl{ServerName: "node1"})
		convey.So(fJob.SeparateNodes.Has("node1"), convey.ShouldBeTrue)
	})
}

func testSwitchAssociateFault(fJob *FaultJob) {
	convey.Convey("When switch fault is associate fault", func() {
		switchInfo := &constant.SwitchInfo{
			SwitchFaultInfo: constant.SwitchFaultInfo{
				NodeStatus:           constant.HealthyState,
				FaultInfo:            []constant.SimpleSwitchFaultInfo{{AssembledFaultCode: "fault1", ForceAdd: false}},
				FaultLevel:           "level1",
				FaultTimeAndLevelMap: map[string]constant.FaultTimeAndLevel{"fault1": {FaultLevel: "level1"}},
			}}

		patches := gomonkey.ApplyFunc(isAssociateFault, func(faultCode string) bool {
			return faultCode == "fault1"
		}).ApplyGlobalVar(&relationFaultTypeMap, sets.NewString("fault1"))
		defer patches.Reset()

		fJob.initBySwitchFault(switchInfo, constant.ServerHccl{ServerName: "node1"})
		convey.So(fJob.AllFaultCode.Has("node1-"+constant.AllCardId+"-fault1"), convey.ShouldBeTrue)
		convey.So(fJob.RelationFaults, convey.ShouldHaveLength, 1)
	})
}

func TestValidateFaultDurationConfig(t *testing.T) {
	convey.Convey("Test validateFaultDurationConfig", t, func() {
		testEmptyFaultCode()
		testInvalidTimeOutInterval()
		testValidConfig()
	})
}

func testEmptyFaultCode() {
	convey.Convey("When fault code is empty", func() {
		faultConfig := constant.FaultDuration{
			FaultCode:       "",
			TimeOutInterval: timeOutInterval,
		}

		patches := gomonkey.ApplyFunc(hwlog.RunLog.Error, func(args ...interface{}) {})
		defer patches.Reset()

		result := validateFaultDurationConfig(faultConfig)
		convey.So(result, convey.ShouldBeFalse)
	})
}

func testInvalidTimeOutInterval() {
	convey.Convey("When time out interval is invalid", func() {
		faultConfig := constant.FaultDuration{
			FaultCode:       "fault1",
			TimeOutInterval: -1,
		}

		patches := gomonkey.ApplyFunc(hwlog.RunLog.Error, func(args ...interface{}) {})
		defer patches.Reset()

		result := validateFaultDurationConfig(faultConfig)
		convey.So(result, convey.ShouldBeFalse)
	})
}

func testValidConfig() {
	convey.Convey("When fault config is valid", func() {
		faultConfig := constant.FaultDuration{
			FaultCode:       "fault1",
			TimeOutInterval: timeOutInterval,
		}

		result := validateFaultDurationConfig(faultConfig)
		convey.So(result, convey.ShouldBeTrue)
	})
}

func TestInitFaultJobs(t *testing.T) {
	convey.Convey("Test InitFaultJobs", t, func() {
		processor := &relationFaultProcessor{
			switchInfoCm: map[string]*constant.SwitchInfo{},
			faultJobs:    make(map[string]*FaultJob),
		}

		testEmptyServerList(processor)
		testInitFaultJob(processor)
	})
}

func testEmptyServerList(processor *relationFaultProcessor) {
	convey.Convey("When server list is empty", func() {
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		deviceInfo := map[string]*constant.AdvanceDeviceFaultCm{"node1": {SuperPodID: 1}}
		processor.deviceInfoCm = deviceInfo

		patches.ApplyFunc(job.GetJobServerInfoMap, func() constant.JobServerInfoMap {
			return constant.JobServerInfoMap{
				InfoMap: map[string]map[string]constant.ServerHccl{
					"job1": {},
				},
			}
		})
		patches.ApplyFunc(hwlog.RunLog.Debugf, func(format string, args ...interface{}) {})

		processor.InitFaultJobs()
		convey.So(processor.faultJobs, convey.ShouldBeEmpty)
	})
}

func testInitFaultJob(processor *relationFaultProcessor) {
	convey.Convey("When initializing fault job", func() {
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		deviceInfo := map[string]*constant.AdvanceDeviceFaultCm{"node1": {SuperPodID: 1}}
		processor.deviceInfoCm = deviceInfo

		patches.ApplyFunc(job.GetJobServerInfoMap, func() constant.JobServerInfoMap {
			return constant.JobServerInfoMap{
				InfoMap: map[string]map[string]constant.ServerHccl{
					"job1": {
						"node1": {
							ServerName:   "node1",
							PodID:        "pod1",
							PodNameSpace: "namespace1",
						},
					},
				},
			}
		})
		patches.ApplyFunc(hwlog.RunLog.Debugf, func(format string, args ...interface{}) {})
		patches.ApplyFunc(util.ObjToString, func(obj interface{}) string {
			return "test-fault-job"
		})

		processor.InitFaultJobs()
		convey.So(processor.faultJobs, convey.ShouldNotBeEmpty)
		convey.So(processor.faultJobs["job1"].IsA3Job, convey.ShouldBeTrue)
		convey.So(processor.faultJobs["job1"].PodNames["node1"], convey.ShouldEqual, "pod1")
	})
}

func TestGetRelationFaultInfo(t *testing.T) {
	convey.Convey("Test GetRelationFaultInfo", t, func() {
		fJobCenter := &relationFaultProcessor{
			faultJobs: map[string]*FaultJob{
				"job1": {NodeFaultInfoMap: map[string][]*constant.FaultInfo{
					"node1": {{FaultCode: faultCode0001}},
				}},
			}}
		convey.Convey("01-relation fault not exit, should return nil", func() {
			faultList := fJobCenter.GetRelationFaultInfo("job2", "node1")
			convey.So(faultList, convey.ShouldBeNil)
		})
		convey.Convey("02-relation fault exit, should return list", func() {
			faultList := fJobCenter.GetRelationFaultInfo("job1", "node1")
			convey.So(faultList, convey.ShouldResemble, []*constant.FaultInfo{{FaultCode: faultCode0001}})
		})
	})

}

func TestUpdateNodeFaultInfoMap(t *testing.T) {
	convey.Convey("Test updateNodeFaultInfoMap", t, func() {
		fJob := &FaultJob{
			NodeFaultInfoMap: map[string][]*constant.FaultInfo{},
		}
		fault := &constant.FaultInfo{NodeName: "node1"}
		fJob.updateNodeFaultInfoMap(fault)
		convey.So(len(fJob.NodeFaultInfoMap) == 1, convey.ShouldBeTrue)
	})

}

const (
	fakeFaultTime1 = 500
	fakeFaultTime2 = 1000
	fakeFaultTime3 = 2000
	time1          = 1
)

func TestFaultJobIsMeetTMOutTriggerFault(t *testing.T) {
	convey.Convey("Test isMeetTMOutTriggerFault function", t, func() {
		now := time.Now().UnixMilli()
		kilo := constant.Kilo

		convey.Convey("Case 1: Should return true when trigger fault time is within allowed range", func() {
			fJob := &FaultJob{TMOutTriggerFault: []constant.FaultInfo{{FaultUid: "trigger-001", FaultTime: now}}}
			fault := &constant.FaultInfo{
				FaultUid: "fault-001", FaultTime: now - int64(fakeFaultTime1*kilo), DealMaxTime: fakeFaultTime2}
			result := fJob.isMeetTMOutTriggerFault(fault)
			convey.So(result, convey.ShouldBeTrue)
			convey.So(len(fJob.TMOutTriggerFault), convey.ShouldEqual, 0)
		})

		convey.Convey("Case 2: Should return false when trigger fault time is before fault time", func() {
			fJob := &FaultJob{TMOutTriggerFault: []constant.FaultInfo{
				{FaultUid: "trigger-002", FaultTime: now - int64(fakeFaultTime2*kilo)}}}
			fault := &constant.FaultInfo{FaultUid: "fault-002", FaultTime: now, DealMaxTime: fakeFaultTime2}
			result := fJob.isMeetTMOutTriggerFault(fault)
			convey.So(result, convey.ShouldBeFalse)
			convey.So(len(fJob.TMOutTriggerFault), convey.ShouldEqual, 1)
		})

		convey.Convey("Case 3: Should return false when time difference exceeds DealMaxTime", func() {
			fJob := &FaultJob{TMOutTriggerFault: []constant.FaultInfo{{FaultUid: "trigger-003", FaultTime: now}}}
			fault := &constant.FaultInfo{FaultUid: "fault-003", FaultTime: now - int64(fakeFaultTime3*kilo),
				DealMaxTime: fakeFaultTime2}
			result := fJob.isMeetTMOutTriggerFault(fault)
			convey.So(result, convey.ShouldBeFalse)
			convey.So(len(fJob.TMOutTriggerFault), convey.ShouldEqual, 1)
		})

		convey.Convey("Case 4: Should filter out non-matching triggers and return true", func() {
			fJob := &FaultJob{TMOutTriggerFault: []constant.FaultInfo{
				{FaultUid: "trigger-004", FaultTime: now + int64(fakeFaultTime2*kilo)},
				{FaultUid: "trigger-005", FaultTime: now}}}
			fault := &constant.FaultInfo{FaultUid: "fault-004", FaultTime: now -
				int64(fakeFaultTime1*kilo), DealMaxTime: fakeFaultTime2}
			result := fJob.isMeetTMOutTriggerFault(fault)
			convey.So(result, convey.ShouldBeTrue)
			convey.So(len(fJob.TMOutTriggerFault), convey.ShouldEqual, 1)
			convey.So(fJob.TMOutTriggerFault[0].FaultUid, convey.ShouldEqual, "trigger-004")
		})
	})
}

func TestGetCQETriggerFault(t *testing.T) {
	convey.Convey("TestGetCQETriggerFault", t, func() {
		convey.Convey("when TriggerFault is nil then len of CQETriggerFault is 0", func() {
			fJob := &FaultJob{}
			result := fJob.getCQETriggerFault()
			convey.So(len(result), convey.ShouldEqual, 0)
		})

		convey.Convey("when TriggerFault contains CQE faults then len of CQETriggerFault is 2", func() {
			fJob := &FaultJob{
				TriggerFault: []constant.FaultInfo{
					{FaultCode: constant.DevCqeFaultCode},
					{FaultCode: constant.HostCqeFaultCode},
					{FaultCode: "other fault code"},
				},
			}

			patches := gomonkey.ApplyFunc(faultdomain.IsCqeFault, func(faultCode string) bool {
				return faultCode == constant.DevCqeFaultCode || faultCode == constant.HostCqeFaultCode
			})
			defer patches.Reset()

			result := fJob.getCQETriggerFault()
			convey.So(len(result), convey.ShouldEqual, constant.GroupIdOffset)
			convey.So(result[0].FaultCode, convey.ShouldEqual, constant.DevCqeFaultCode)
			convey.So(result[1].FaultCode, convey.ShouldEqual, constant.HostCqeFaultCode)
		})

		convey.Convey("when TriggerFault contains no CQE faults then len of CQETriggerFault is 0", func() {
			fJob := &FaultJob{
				TriggerFault: []constant.FaultInfo{
					{FaultCode: "other fault code 1"},
					{FaultCode: "other fault code 2"},
				},
			}
			patches := gomonkey.ApplyFunc(faultdomain.IsCqeFault, func(faultCode string) bool {
				return false
			})
			defer patches.Reset()

			result := fJob.getCQETriggerFault()
			convey.So(len(result), convey.ShouldEqual, 0)
		})
	})
}

func TestExecDeviceFaultTMOut(t *testing.T) {
	convey.Convey("test execDeviceFaultTMOut", t, func() {
		convey.Convey("update node fault info map", func() {
			fJob := &FaultJob{
				FaultStrategy:    constant.FaultStrategy{NodeLvList: map[string]string{}},
				NodeFaultInfoMap: map[string][]*constant.FaultInfo{},
			}
			fault := &constant.FaultInfo{NodeName: node101}
			patches := gomonkey.ApplyPrivateMethod(fJob, "isMeetTMOutTriggerFault",
				func(_ *FaultJob, fault *constant.FaultInfo) bool { return true })
			defer patches.Reset()
			fJob.execDeviceFaultTMOut(fault)
			convey.So(fJob.NodeFaultInfoMap[node101], convey.ShouldResemble, []*constant.FaultInfo{
				{NodeName: node101, ExecutedStrategy: constant.SeparateFaultStrategy}})
		})
	})
}
