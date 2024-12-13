package faultmanager

import (
	"testing"

	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
)

func TestFaultJobProcessNetworkFault(t *testing.T) {
	t.Run("processNetworkFault_noTrigger", func(t *testing.T) {
		strategyList := make([]RelationFaultStrategy, 0)
		strategyList = append(strategyList, RelationFaultStrategy{
			TriggerFault:   "0x0001",
			RelationFaults: []string{"0x0002", "0x0003"},
			FaultStrategy:  constant.SeparateFaultStrategy,
		})
		networkFaults := make([]*faultInfo, 0)
		fault1 := faultInfo{
			NodeName:  "node-101",
			NPUName:   "Ascend-910-1",
			FaultType: "devicefault",
			FaultCode: "0x0002",
		}
		fault2 := faultInfo{
			NodeName:  "node-101",
			NPUName:   "Ascend-910-2",
			FaultType: "devicefault",
			FaultCode: "0x0003",
		}
		networkFaults = append(networkFaults, &fault1, &fault2)
		retryEventList := make([]faultInfo, 0)
		faultJob := FaultJob{
			FindNPUUnderSwitch: false,
		}
		relationFaultStrategies = strategyList
		faultJob.RelationFaults = networkFaults
		faultJob.TriggerFault = retryEventList
		faultJob.processNetworkFault()
		t.Log(util.ObjToString(faultJob.FaultStrategy))
	})

	t.Run("processNetworkFault_not_all_network_in_RelationFaults", func(t *testing.T) {
		strategyList := make([]RelationFaultStrategy, 0)
		strategyList = append(strategyList, RelationFaultStrategy{
			TriggerFault:   "0x0001",
			RelationFaults: []string{"0x0002", "0x0003"},
			FaultStrategy:  constant.SeparateFaultStrategy,
		})
		networkFaults := make([]*faultInfo, 0)
		fault1 := faultInfo{
			NodeName:  "node-100",
			NPUName:   "Ascend-910-1",
			FaultType: "devicefault",
			FaultCode: "0x0002",
		}
		fault2 := faultInfo{
			NodeName:  "node-101",
			NPUName:   "Ascend-910-2",
			FaultType: "devicefault",
			FaultCode: "0x0004",
		}
		networkFaults = append(networkFaults, &fault1, &fault2)
		retryEventList := make([]faultInfo, 0)
		retryEvent := faultInfo{
			NodeName:  "node-102",
			NPUName:   "Ascend-910-2",
			FaultType: "retryEvent",
			FaultCode: "0x0001",
		}
		retryEventList = append(retryEventList, retryEvent)
		faultJob := FaultJob{
			FindNPUUnderSwitch: false,
		}
		relationFaultStrategies = strategyList
		faultJob.RelationFaults = networkFaults
		faultJob.TriggerFault = retryEventList
		faultJob.processNetworkFault()
		t.Log(util.ObjToString(faultJob.FaultStrategy))

	})

	t.Run("processNetworkFault_right_node_and_device_separate", func(t *testing.T) {
		strategyList := make([]RelationFaultStrategy, 0)
		strategyList = append(strategyList, RelationFaultStrategy{
			TriggerFault:   "0x0001",
			RelationFaults: []string{"0x0002", "0x0003"},
			FaultStrategy:  constant.SeparateFaultStrategy,
		})
		networkFaults := make([]*faultInfo, 0)
		fault1 := faultInfo{
			NodeName:  "node-100",
			NPUName:   "Ascend-910-1",
			FaultType: "devicefault",
			FaultCode: "0x0002",
		}
		fault2 := faultInfo{
			NodeName:  "node-101",
			NPUName:   "Ascend-910-2",
			FaultType: "devicefault",
			FaultCode: "0x0003",
		}
		fault3 := faultInfo{
			NodeName:  "node-103",
			NPUName:   "Ascend-910-2",
			FaultType: constant.SwitchFault,
			FaultCode: "0x0003",
		}
		networkFaults = append(networkFaults, &fault1, &fault2, &fault3)
		retryEventList := make([]faultInfo, 0)
		retryEvent := faultInfo{
			NPUName:   "Ascend-910-2",
			FaultType: "retryEvent",
			FaultCode: "0x0001",
		}
		retryEventList = append(retryEventList, retryEvent)
		faultJob := FaultJob{
			FindNPUUnderSwitch: false,
		}
		relationFaultStrategies = strategyList
		faultJob.RelationFaults = networkFaults
		faultJob.TriggerFault = retryEventList
		faultJob.processNetworkFault()
		t.Log(util.ObjToString(faultJob.FaultStrategy))
	})

	t.Run("processNetworkFault_right_node_separate_device_subHealth", func(t *testing.T) {
		strategyList := make([]RelationFaultStrategy, 0)
		strategyList = append(strategyList, RelationFaultStrategy{
			TriggerFault:   "Trigger0x0001",
			RelationFaults: []string{"0x0002", "0x0003"},
			FaultStrategy:  constant.SubHealthFaultStrategy,
		})
		strategyList = append(strategyList, RelationFaultStrategy{
			TriggerFault:   "Trigger0x0002",
			RelationFaults: []string{"0x0004"},
			FaultStrategy:  constant.SeparateFaultStrategy,
		})
		networkFaults := make([]*faultInfo, 0)
		fault1 := faultInfo{
			NodeName:  "node-100",
			NPUName:   "Ascend-910-1",
			FaultType: "devicefault",
			FaultCode: "0x0002",
		}
		fault2 := faultInfo{
			NodeName:  "node-101",
			NPUName:   "Ascend-910-2",
			FaultType: "devicefault",
			FaultCode: "0x0003",
		}
		nodeFault := faultInfo{
			NodeName:  "node-103",
			NPUName:   "Ascend-910-2",
			FaultType: constant.SwitchFault,
			FaultCode: "0x0004",
		}

		networkFaults = append(networkFaults, &fault1, &fault2, &nodeFault)
		retryEventList := make([]faultInfo, 0)
		retryEvent := faultInfo{
			NPUName:   "Ascend-910-2",
			FaultType: "retryEvent",
			FaultCode: "Trigger0x0001",
		}
		retryEvent2 := faultInfo{
			NPUName:   "Ascend-910-2",
			FaultType: "retryEvent",
			FaultCode: "Trigger0x0002",
		}
		retryEventList = append(retryEventList, retryEvent, retryEvent2)
		faultJob := FaultJob{
			FindNPUUnderSwitch: false,
		}
		relationFaultStrategies = strategyList
		faultJob.RelationFaults = networkFaults
		faultJob.TriggerFault = retryEventList
		faultJob.processNetworkFault()

		t.Log(util.ObjToString(faultJob.FaultStrategy))

	})

	t.Run("processNetworkFault_right_node_separate_device_both has sub-health and separate", func(t *testing.T) {
		strategyList := make([]RelationFaultStrategy, 0)
		strategyList = append(strategyList, RelationFaultStrategy{
			TriggerFault:   "Trigger0x0001",
			RelationFaults: []string{"0x0002", "0x0003"},
			FaultStrategy:  constant.SubHealthFaultStrategy,
		})
		strategyList = append(strategyList, RelationFaultStrategy{
			TriggerFault:   "Trigger0x0002",
			RelationFaults: []string{"0x0004"},
			FaultStrategy:  constant.SeparateFaultStrategy,
		})
		networkFaults := make([]*faultInfo, 0)
		fault1 := faultInfo{
			NodeName:  "node-100",
			NPUName:   "Ascend-910-1",
			FaultType: "devicefault",
			FaultCode: "0x0002",
		}
		fault2 := faultInfo{
			NodeName:  "node-101",
			NPUName:   "Ascend-910-2",
			FaultType: "devicefault",
			FaultCode: "0x0003",
		}
		fault4 := faultInfo{
			NodeName:  "node-100",
			NPUName:   "Ascend-910-1",
			FaultType: "devicefault",
			FaultCode: "0x0004",
		}
		nodeFault := faultInfo{
			NodeName:  "node-103",
			NPUName:   "Ascend-910-2",
			FaultType: constant.SwitchFault,
			FaultCode: "0x0004",
		}

		networkFaults = append(networkFaults, &fault1, &fault2, &nodeFault, &fault4)
		retryEventList := make([]faultInfo, 0)
		retryEvent := faultInfo{
			NPUName:   "Ascend-910-2",
			FaultType: "retryEvent",
			FaultCode: "Trigger0x0001",
		}
		retryEvent2 := faultInfo{
			NPUName:   "Ascend-910-2",
			FaultType: "retryEvent",
			FaultCode: "Trigger0x0002",
		}
		retryEventList = append(retryEventList, retryEvent, retryEvent2)
		faultJob := FaultJob{
			FindNPUUnderSwitch: false,
		}
		relationFaultStrategies = strategyList
		faultJob.RelationFaults = networkFaults
		faultJob.TriggerFault = retryEventList
		faultJob.processNetworkFault()
		t.Log(util.ObjToString(faultJob.FaultStrategy))

	})
}
