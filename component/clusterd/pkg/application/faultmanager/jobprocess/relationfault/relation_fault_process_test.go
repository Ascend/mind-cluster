// Copyright (c) Huawei Technologies Co., Ltd. 2024-2025. All rights reserved.

// Package relationfault contain relation fault process
package relationfault

import (
	"testing"

	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
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
		fault3 := constant.FaultInfo{NodeName: node103, NPUName: switchDevice, FaultType: constant.SwitchFault, FaultCode: faultCode0003}
		networkFaults = append(networkFaults, &fault1, &fault2, &fault3)

		triggerList := make([]constant.FaultInfo, 0)
		retryEvent := constant.FaultInfo{NPUName: npu2, FaultType: deviceFault, FaultCode: triggerCode0001}
		triggerList = append(triggerList, retryEvent)

		faultJob := FaultJob{FindNPUUnderSwitch: false}
		relationFaultStrategies = strategyList
		faultJob.RelationFaults = networkFaults
		faultJob.TriggerFault = triggerList
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
		nodeFault := constant.FaultInfo{NodeName: node103, NPUName: switchDevice, FaultType: constant.SwitchFault, FaultCode: faultCode0004}

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
