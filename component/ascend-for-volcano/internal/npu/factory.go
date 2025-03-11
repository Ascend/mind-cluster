/*
Copyright(C)2025. Huawei Technologies Co.,Ltd. All rights reserved.

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

/*
Package npu is using for HuaWei Ascend pin affinity schedule.
*/
package npu

import (
	"fmt"
	"strings"

	"k8s.io/klog"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/npu/ascend310/card310x4"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/npu/ascend310/chip310x4"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/npu/ascend310p/card310px2"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/npu/ascend310p/chip310px2"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/npu/ascend910/ascend910a3/module910a3x16"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/npu/ascend910/ascend910a3/superpod"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/npu/ascend910/ascend910b/module910bx16"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/npu/ascend910/asend910old/module910x8"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/npu/base"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

var (
	card910Factory  = map[string]base.AscendHandler{}
	card310Factory  = map[string]base.AscendHandler{}
	card310pFactory = map[string]base.AscendHandler{}
)

const (
	// accelerator310Key accelerator key of old infer card
	accelerator310Key     = "npu-310-strategy"
	chipAcceleratorValue  = "chip"
	cardAcceleratorValue  = "card"
	card310pMaxNodeNPUNum = 64
)

const (
	duoKeyLabel = "duo"
	trueStr     = "true"
)

const (
	card910x2Name    = "huawei.com/Ascend910card"
	half910x4Name    = "huawei.com/Ascend910half"
	module910bx8Name = "huawei.com/Ascend910module-910b-8"
)

func init() {
	initCard910Factory()
	initCard310Factory()
	initCard310PFactory()
}

func initCard310Factory() {
	card310Factory[chip310x4.SchedulerName] = chip310x4.New(chip310x4.SchedulerName)
	card310Factory[card310x4.SchedulerName] = card310x4.New(card310x4.SchedulerName)
}

func initCard310PFactory() {
	card310pFactory[chip310px2.SchedulerName] = chip310px2.New(chip310px2.SchedulerName)
	card310pFactory[card310px2.SchedulerName] = card310px2.New(card310px2.SchedulerName)
}

func initCard910Factory() {
	card910Factory[card910x2Name] = base.New(util.NPU910CardName,
		base.WithAnnoPreVal(util.NPU910CardNamePre), base.WithMaxNodeNum(util.NPUIndex2))
	card910Factory[module910bx8Name] = base.New(util.NPU910CardName,
		base.WithAnnoPreVal(util.NPU910CardNamePre), base.WithMaxNodeNum(util.NPUIndex8), base.WithNetworkFault(true))
	card910Factory[half910x4Name] = base.New(util.NPU910CardName,
		base.WithAnnoPreVal(util.NPU910CardNamePre),
		base.WithMaxNodeNum(util.NPUIndex4),
		base.WithNetworkFault(true),
		base.WithNpuInvalidMap(map[int]struct{}{util.NPUIndex3: {}}))
	card910Factory[module910bx16.SchedulerName] = module910bx16.New(module910bx16.SchedulerName)
	card910Factory[module910x8.SchedulerName] = module910x8.New(module910x8.SchedulerName)
	card910Factory[superpod.SchedulerName] = superpod.New(superpod.SchedulerName)
	card910Factory[module910a3x16.SchedulerName] = module910a3x16.New(module910a3x16.SchedulerName)
}

// InitPolicyHandler init npu affinity policy handler
func InitPolicyHandler(attr util.SchedulerJobAttr, env plugin.ScheduleEnv) (plugin.SchedulerPluginNeed, bool) {
	pluginName := attr.GetPluginNameByReq()
	switch pluginName {
	case util.NPU910CardName:
		return init910CardPolicyHandler(attr)
	case util.NPU310CardName:
		return init310CardPolicyHandler(attr)
	case util.NPU310PCardName:
		return init310PCardPolicyHandler(attr)
	default:
		return nil, false
	}
}

func init310CardPolicyHandler(attr util.SchedulerJobAttr) (plugin.SchedulerPluginNeed, bool) {
	v, ok := attr.Label[accelerator310Key]
	if !ok {
		v = chipAcceleratorValue
	}
	value, ok := card310Factory[attr.ReqNPUName+v]
	if !ok {
		return nil, false
	}
	return value, true
}

func init910CardPolicyHandler(attr util.SchedulerJobAttr) (plugin.SchedulerPluginNeed, bool) {
	handlerName := get910CardHandlerName(attr)
	value, ok := card910Factory[handlerName]
	if !ok {
		err := fmt.Errorf("not support %s", handlerName)
		klog.V(util.LogErrorLev).Infof("Init910CardPolicyHandler err: %s", err.Error())
		return nil, false
	}
	return value, ok
}

func get910CardHandlerName(attr util.SchedulerJobAttr) string {
	_, ok := attr.Annotation[superpod.SuperPodAnnoKey]
	if ok {
		return superpod.SchedulerName
	}
	v, ok := attr.Selector[util.AcceleratorType]
	if !ok {
		v = util.ModuleAcceleratorType
		return util.NPU910CardName + util.ModuleAcceleratorType
	}
	if strings.Contains(v, cardAcceleratorValue) {
		return util.NPU910CardName + cardAcceleratorValue
	}
	return util.NPU910CardName + v
}

func init310PCardPolicyHandler(attr util.SchedulerJobAttr) (plugin.SchedulerPluginNeed, bool) {
	duo, ok := attr.Label[duoKeyLabel]
	if !ok {
		klog.V(util.LogInfoLev).Info("not 300I duo")
		return base.New(util.NPU310PCardName,
			base.WithAnnoPreVal(util.NPU310PCardNamePre), base.WithMaxNodeNum(card310pMaxNodeNPUNum)), false
	}
	v, ok := attr.Label[accelerator310Key]
	if !ok {
		v = chipAcceleratorValue
	}
	if duo == trueStr {
		duo = duoKeyLabel
	}
	value, ok := card310pFactory[attr.ReqNPUName+duo+v]
	if !ok {
		return nil, false
	}
	return value, true
}
