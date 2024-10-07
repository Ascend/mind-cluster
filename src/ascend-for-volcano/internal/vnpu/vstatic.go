/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*
Package vnpu is using for HuaWei Ascend pin vnpu allocation.
*/
package vnpu

import (
	"errors"
	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

// CheckNodeNPUByTask pass for static
func (tp *StaticVNPU) CheckNodeNPUByTask(task *api.TaskInfo, node plugin.NPUNode, _ util.VResource) error {
	if tp == nil || task == nil {
		klog.V(util.LogDebugLev).Infof("CheckNodeNPUByTask failed: %s", util.ArgumentError)
		return errors.New(util.ArgumentError)
	}
	klog.V(util.LogInfoLev).Infof("static vnpu task<%s> node<%s> CheckNodeNPUByTask pass", task.Name, node.Name)
	return nil
}

// ScoreBestNPUNodes pass for static
func (tp *StaticVNPU) ScoreBestNPUNodes(task *api.TaskInfo, _ []*api.NodeInfo, scoreMap map[string]float64) error {
	if tp == nil || task == nil {
		klog.V(util.LogDebugLev).Infof("ScoreBestNPUNodes failed: %s", util.ArgumentError)
		return errors.New(util.ArgumentError)
	}
	klog.V(util.LogInfoLev).Infof("static vnpu task<%s> scoreMap<%#v> ScoreBestNPUNodes pass",
		task.Name, scoreMap)
	return nil
}

// UseAnnotation pass for static
func (tp *StaticVNPU) UseAnnotation(task *api.TaskInfo, node plugin.NPUNode, _ util.VResource,
	_ VTemplate) *plugin.NPUNode {
	if tp == nil || task == nil {
		klog.V(util.LogDebugLev).Infof("UseAnnotation failed: %s", util.ArgumentError)
		return &node
	}
	klog.V(util.LogInfoLev).Infof("static vnpu task<%s> node<%s> UseAnnotation pass", task.Name, node.Name)
	return &node
}
