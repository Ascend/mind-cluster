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
Package npu provides Huawei Ascend NPU pin affinity scheduling functionality.
*/
package npu

import (
	"strings"

	"k8s.io/klog"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/npu/ascend310/card310x4"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/npu/ascend310/chip310x4"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/npu/ascend310p/card310px2"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/npu/ascend310p/chip310px2"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/npu/ascend310p/vnpu"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/npu/ascend910/ascend800ia5/ascend800ia5stacking"
	ascend800ia5superpod "volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/npu/ascend910/ascend800ia5/superpod"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/npu/ascend910/ascend910a3/module910a3x16"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/npu/ascend910/ascend910a3/superpod"
	superpoda5 "volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/npu/ascend910/ascend910a5/superpod"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/npu/ascend910/ascend910b/module910bx16"
	vnpu2 "volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/npu/ascend910/ascend910b/vnpu"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/npu/ascend910/ascend910old/module910x8"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/npu/base"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

var (
	card910Factory  = map[string]func() base.AscendHandler{}
	card310Factory  = map[string]func() base.AscendHandler{}
	card310pFactory = map[string]func() base.AscendHandler{}

	// policyHandlerMap maps scheduling policies to handler names
	policyHandlerMap = map[string]string{util.SchedulePolicyA3x16: module910a3x16.SchedulerName}
)

const (
	chipAcceleratorValue  = "chip"
	cardAcceleratorValue  = "card"
	card310pMaxNodeNPUNum = 64
)

const (
	duoKeyLabel = "duo"
	trueStr     = "true"
)

const (
	card910x2Name    = util.HwPreName + util.Ascend910 + "card"
	half910x4Name    = util.HwPreName + util.Ascend910 + "half"
	module910bx8Name = util.HwPreName + util.Ascend910 + util.Module910bx8AcceleratorType
)

func init() {
	initCard910Factory()
	initCard310Factory()
	initCard310PFactory()
}

func initCard310Factory() {
	card310Factory[chip310x4.SchedulerName] =
		func() base.AscendHandler { return chip310x4.New(chip310x4.SchedulerName) }
	card310Factory[card310x4.SchedulerName] =
		func() base.AscendHandler { return card310x4.New(card310x4.SchedulerName) }
}

func initCard310PFactory() {
	card310pFactory[chip310px2.SchedulerName] =
		func() base.AscendHandler { return chip310px2.New(chip310px2.SchedulerName) }
	card310pFactory[card310px2.SchedulerName] =
		func() base.AscendHandler { return card310px2.New(card310px2.SchedulerName) }
}

func initCard910Factory() {
	card910Factory[card910x2Name] = func() base.AscendHandler {
		return base.New(util.NPU910CardName,
			base.WithAnnoPreVal(util.NPU910CardNamePre), base.WithMaxNodeNum(util.NPUIndex8))
	}
	card910Factory[module910bx8Name] = func() base.AscendHandler {
		return base.New(util.NPU910CardName, base.WithAnnoPreVal(util.NPU910CardNamePre),
			base.WithMaxNodeNum(util.NPUIndex8), base.WithNetworkFault(true))
	}
	card910Factory[util.Ascend800ia5x8SchedulerName] = func() base.AscendHandler {
		return base.New(util.NPU910CardName, base.WithAnnoPreVal(util.NPU910CardNamePre),
			base.WithMaxNodeNum(util.NPUIndex8), base.WithNetworkFault(true))
	}
	card910Factory[util.Ascend800ta5x8TrainSchedulerName] = func() base.AscendHandler {
		return base.New(util.NPU910CardName, base.WithAnnoPreVal(util.NPU910CardNamePre),
			base.WithMaxNodeNum(util.NPUIndex8), base.WithNetworkFault(true))
	}
	card910Factory[half910x4Name] = func() base.AscendHandler {
		return base.New(util.NPU910CardName,
			base.WithAnnoPreVal(util.NPU910CardNamePre),
			base.WithMaxNodeNum(util.NPUIndex4),
			base.WithNetworkFault(true),
			base.WithNpuInvalidMap(map[int]struct{}{util.NPUIndex3: {}}))
	}
	card910Factory[module910bx16.SchedulerName] =
		func() base.AscendHandler { return module910bx16.New(module910bx16.SchedulerName) }
	card910Factory[module910x8.SchedulerName] =
		func() base.AscendHandler { return module910x8.New(module910x8.SchedulerName) }
	card910Factory[superpod.SchedulerName] =
		func() base.AscendHandler { return superpod.New(superpod.SchedulerName) }
	card910Factory[module910a3x16.SchedulerName] =
		func() base.AscendHandler { return module910a3x16.New(module910a3x16.SchedulerName) }
	card910Factory[superpoda5.SuperPodx8SchedulerName] =
		func() base.AscendHandler { return module910a3x16.New(superpoda5.SuperPodx8SchedulerName) }
	card910Factory[ascend800ia5superpod.InferSchedulerName] =
		func() base.AscendHandler { return ascend800ia5superpod.New(ascend800ia5superpod.InferSchedulerName) }
	card910Factory[ascend800ia5superpod.TrainSchedulerName] =
		func() base.AscendHandler { return ascend800ia5superpod.New(ascend800ia5superpod.InferSchedulerName) }
	card910Factory[ascend800ia5stacking.SchedulerName] =
		func() base.AscendHandler { return ascend800ia5stacking.New(ascend800ia5stacking.SchedulerName) }
}

// InitPolicyHandler initializes the NPU affinity policy handler
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
	v, ok := attr.Label[util.Accelerator310Key]
	if !ok {
		v = chipAcceleratorValue
	}
	handlerFunc, ok := card310Factory[attr.ReqNPUName+v]
	if !ok {
		return nil, false
	}
	return handlerFunc(), true
}

func init910CardPolicyHandler(attr util.SchedulerJobAttr) (plugin.SchedulerPluginNeed, bool) {
	if attr.ReqNPUName == util.AscendNPUCore {
		return vnpu2.New(util.NPU910CardName), true
	}
	handlerName := get910CardHandlerName(attr)
	handlerFunc, ok := card910Factory[handlerName]
	if !ok {
		klog.V(util.LogErrorLev).Infof("Handler %s not found in card910Factory", handlerName)
		return nil, false
	}
	return handlerFunc(), true
}

func get910CardHandlerName(attr util.SchedulerJobAttr) string {
	policy, ok := attr.Annotation[util.SchedulePolicyAnnoKey]
	if ok {
		handlerName, ok := policyHandlerMap[policy]
		if ok {
			klog.V(util.LogInfoLev).Infof("get handler name for schedule policy %s success", policy)
			return handlerName
		}
	}
	if handlerName, ok := get910A5CardHandlerName(attr); ok {
		klog.V(util.LogInfoLev).Infof("handler %s found in A5CardFactory", handlerName)
		return handlerName
	}
	if handlerName, ok := get800IA5HandlerName(attr); ok {
		klog.V(util.LogInfoLev).Infof("handler %s found in 800IA5CardFactory", handlerName)
		return handlerName
	}
	if _, ok := attr.Annotation[util.SuperPodAnnoKey]; ok {
		return superpod.SchedulerName
	}
	v, ok := attr.Selector[util.AcceleratorType]
	if !ok {
		return util.NPU910CardName + util.ModuleAcceleratorType
	}
	if strings.Contains(v, cardAcceleratorValue) {
		return util.NPU910CardName + cardAcceleratorValue
	}
	return util.NPU910CardName + v
}

func init310PCardPolicyHandler(attr util.SchedulerJobAttr) (plugin.SchedulerPluginNeed, bool) {
	if attr.ReqNPUName == util.AscendNPUCore {
		return vnpu.New(util.NPU310PCardName), true
	}
	handlerName := get310PCardHandlerName(attr)
	handlerFunc, ok := card310pFactory[handlerName]
	if !ok {
		klog.V(util.LogWarningLev).Infof("Handler %s not found in Factory", handlerName)
		return base.New(util.NPU310PCardName,
			base.WithAnnoPreVal(util.NPU310PCardNamePre), base.WithMaxNodeNum(card310pMaxNodeNPUNum)), true
	}
	return handlerFunc(), true
}

func get310PCardHandlerName(attr util.SchedulerJobAttr) string {
	duo := attr.Label[duoKeyLabel]
	if duo == trueStr {
		klog.V(util.LogInfoLev).Info("Detected as 300I duo configuration")
		duo = duoKeyLabel
	}
	v, ok := attr.Label[util.Accelerator310Key]
	if !ok {
		v = chipAcceleratorValue
	}
	return attr.ReqNPUName + duo + v
}

func get910A5CardHandlerName(attr util.SchedulerJobAttr) (string, bool) {
	acceleratorType, existAcceleratorType := attr.Selector[util.AcceleratorType]
	_, existSpBlock := attr.Annotation[superpoda5.SuperPodAnnoKey]
	if existSpBlock && existAcceleratorType && acceleratorType == superpoda5.SuperPodx8 {
		return util.NPU910CardName + acceleratorType, true
	}
	return "", false
}

func get800IA5HandlerName(attr util.SchedulerJobAttr) (string, bool) {
	const SuperPodAnnoKey = "sp-block"
	acceleratorType, existAcceleratorType := attr.Selector[util.AcceleratorType]
	if !existAcceleratorType {
		return "", false
	}

	_, existSpBlock := attr.Annotation[SuperPodAnnoKey]
	switch acceleratorType {
	case util.Ascend800ia5x8, util.Ascend800ta5x8, ascend800ia5stacking.Ascend800ia5stacking:
		return util.NPU910CardName + acceleratorType, true
	case ascend800ia5superpod.AcceleratorType, ascend800ia5superpod.AcceleratorTypeTrain:
		{
			if existSpBlock {
				return util.NPU910CardName + acceleratorType, true
			} else {
				return "", false
			}
		}
	default:
		return "", false
	}
}
