// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package kube a series of kube function
package kube

import (
	"fmt"
	"reflect"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"volcano.sh/apis/pkg/apis/scheduling/v1beta1"
	"volcano.sh/apis/pkg/client/informers/externalversions"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
	"clusterd/pkg/domain/device"
	"clusterd/pkg/domain/node"
	"clusterd/pkg/domain/publicfault"
	"clusterd/pkg/domain/switchinfo"
)

var (
	cmDeviceFuncs   = map[string][]func(*constant.DeviceInfo, *constant.DeviceInfo, string){}
	cmNodeFuncs     = map[string][]func(*constant.NodeInfo, *constant.NodeInfo, string){}
	cmSwitchFuncs   = map[string][]func(*constant.SwitchInfo, *constant.SwitchInfo, string){}
	cmPubFaultFuncs = map[string][]func(*api.PubFaultInfo, *api.PubFaultInfo, string){}
	podGroupFuncs   = map[string][]func(*v1beta1.PodGroup, *v1beta1.PodGroup, string){}
	podFuncs        = map[string][]func(*v1.Pod, *v1.Pod, string){}
	nodeFuncs       = map[string][]func(*v1.Node, *v1.Node, string){}
	informerCh      = make(chan struct{})
)

// JobService a interface with DeleteJob method
type JobService interface {
	// DeleteJob unregistry job
	DeleteJob(jobId string)
}

// StopInformer stop informer when loss-leader
func StopInformer() {
	if informerCh != nil {
		close(informerCh)
		return
	}
	hwlog.RunLog.Warn("channel is nil will not close it")
}

// CleanFuncs clean funcs when loss-leader
func CleanFuncs() {
	cmDeviceFuncs = map[string][]func(*constant.DeviceInfo, *constant.DeviceInfo, string){}
	cmNodeFuncs = map[string][]func(*constant.NodeInfo, *constant.NodeInfo, string){}
	cmSwitchFuncs = map[string][]func(*constant.SwitchInfo, *constant.SwitchInfo, string){}
	cmPubFaultFuncs = map[string][]func(*api.PubFaultInfo, *api.PubFaultInfo, string){}
	podGroupFuncs = map[string][]func(*v1beta1.PodGroup, *v1beta1.PodGroup, string){}
	podFuncs = map[string][]func(*v1.Pod, *v1.Pod, string){}
	nodeFuncs = map[string][]func(*v1.Node, *v1.Node, string){}
}

// AddPodGroupFunc add podGroup func
func AddPodGroupFunc(business string, func1 ...func(*v1beta1.PodGroup, *v1beta1.PodGroup, string)) {
	if _, ok := podGroupFuncs[business]; !ok {
		podGroupFuncs[business] = []func(*v1beta1.PodGroup, *v1beta1.PodGroup, string){}
	}

	podGroupFuncs[business] = append(podGroupFuncs[business], func1...)
}

// AddPodFunc add pod func
func AddPodFunc(business string, func1 ...func(*v1.Pod, *v1.Pod, string)) {
	if _, ok := podFuncs[business]; !ok {
		podFuncs[business] = []func(*v1.Pod, *v1.Pod, string){}
	}

	podFuncs[business] = append(podFuncs[business], func1...)
}

// AddNodeFunc add node func
func AddNodeFunc(business string, func1 ...func(*v1.Node, *v1.Node, string)) {
	if _, ok := nodeFuncs[business]; !ok {
		nodeFuncs[business] = []func(*v1.Node, *v1.Node, string){}
	}

	nodeFuncs[business] = append(nodeFuncs[business], func1...)
}

// AddCmDeviceFunc add device func, map by business
func AddCmDeviceFunc(business string, func1 ...func(*constant.DeviceInfo, *constant.DeviceInfo, string)) {
	if _, ok := cmDeviceFuncs[business]; !ok {
		cmDeviceFuncs[business] = []func(*constant.DeviceInfo, *constant.DeviceInfo, string){}
	}

	cmDeviceFuncs[business] = append(cmDeviceFuncs[business], func1...)
}

// AddCmSwitchFunc add switch func
func AddCmSwitchFunc(business string, func1 ...func(*constant.SwitchInfo, *constant.SwitchInfo, string)) {
	if _, ok := cmSwitchFuncs[business]; !ok {
		cmSwitchFuncs[business] = []func(*constant.SwitchInfo, *constant.SwitchInfo, string){}
	}

	cmSwitchFuncs[business] = append(cmSwitchFuncs[business], func1...)
}

// AddCmNodeFunc add node func, map by business
func AddCmNodeFunc(business string, func1 ...func(*constant.NodeInfo, *constant.NodeInfo, string)) {
	if _, ok := cmNodeFuncs[business]; !ok {
		cmNodeFuncs[business] = []func(*constant.NodeInfo, *constant.NodeInfo, string){}
	}

	cmNodeFuncs[business] = append(cmNodeFuncs[business], func1...)
}

// AddCmPubFaultFunc add public fault deal func, map by business
func AddCmPubFaultFunc(business string, func1 ...func(*api.PubFaultInfo, *api.PubFaultInfo, string)) {
	if _, ok := cmPubFaultFuncs[business]; !ok {
		cmPubFaultFuncs[business] = []func(*api.PubFaultInfo, *api.PubFaultInfo, string){}
	}

	cmPubFaultFuncs[business] = append(cmPubFaultFuncs[business], func1...)
}

// GetNodeFromIndexer get node from informer indexer
func GetNodeFromIndexer(name string) (*v1.Node, error) {
	item, exist, err := nodeInformer.GetIndexer().GetByKey(name)
	if err != nil || !exist {
		return nil, fmt.Errorf("get node %s from informer failed, err: %v, exist: %v", name, err, exist)
	}
	n, ok := item.(*v1.Node)
	if !ok {
		return nil, fmt.Errorf("get node %s from informer failed, item: %v", name, item)
	}
	return n, nil
}

var nodeInformer cache.SharedIndexInformer

// InitPodAndNodeInformer init pod informer
func InitPodAndNodeInformer() {
	factory := informers.NewSharedInformerFactoryWithOptions(k8sClient.ClientSet, 0)
	podInformer := factory.Core().V1().Pods().Informer()
	podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			podHandler(nil, obj, constant.AddOperator)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			if !reflect.DeepEqual(oldObj, newObj) {
				podHandler(oldObj, newObj, constant.UpdateOperator)
			}
		},
		DeleteFunc: func(obj interface{}) {
			podHandler(nil, obj, constant.DeleteOperator)
		},
	})

	nodeInformer = factory.Core().V1().Nodes().Informer()
	nodeInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			nodeHandler(nil, obj, constant.AddOperator)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			if !reflect.DeepEqual(oldObj, newObj) {
				nodeHandler(oldObj, newObj, constant.UpdateOperator)
			}
		},
		DeleteFunc: func(obj interface{}) {
			nodeHandler(nil, obj, constant.DeleteOperator)
		},
	},
	)
	factory.Start(informerCh)
	factory.WaitForCacheSync(wait.NeverStop)
}

func nodeHandler(oldObj, newObj interface{}, operator string) {
	newNode, ok := newObj.(*v1.Node)
	if !ok {
		hwlog.RunLog.Error("new obj is not node type")
		return
	}
	var oldNode *v1.Node
	if oldObj != nil {
		oldNode, ok = oldObj.(*v1.Node)
		if !ok {
			hwlog.RunLog.Error("old obj is not node type")
			return
		}
	}
	index := 0
	for _, nodeFunc := range nodeFuncs {
		// different businesses use different data sources
		oldNodeForBusiness := oldNode
		newNodeForBusiness := newNode
		if oldNode != nil && newNode != nil && index > 0 {
			oldNodeForBusiness = oldNode.DeepCopy()
			newNodeForBusiness = newNode.DeepCopy()
		}
		for _, nfunc := range nodeFunc {
			nfunc(oldNodeForBusiness, newNodeForBusiness, operator)
		}
		index++
	}
}

func podHandler(oldObj interface{}, newObj interface{}, operator string) {
	newPod, ok := newObj.(*v1.Pod)
	if !ok {
		return
	}
	var oldPod *v1.Pod
	if oldObj != nil {
		oldPod, ok = oldObj.(*v1.Pod)
		if !ok {
			return
		}
	}
	index := 0
	for _, podFunc := range podFuncs {
		// different businesses use different data sources
		oldPodForBusiness := oldPod
		newPodForBusiness := newPod
		if oldPod != nil && newPod != nil && index > 0 {
			oldPodForBusiness = oldPod.DeepCopy()
			newPodForBusiness = newPod.DeepCopy()
		}
		for _, pfunc := range podFunc {
			pfunc(oldPodForBusiness, newPodForBusiness, operator)
		}
		index++
	}
}

// InitCMInformer init configmap informer
func InitCMInformer() {
	informerFactory := informers.NewSharedInformerFactoryWithOptions(k8sClient.ClientSet, 0,
		informers.WithTweakListOptions(func(options *metav1.ListOptions) {
			options.LabelSelector = constant.CmConsumerCIM + "=" + constant.CmConsumerValue
		}))
	cmInformer := informerFactory.Core().V1().ConfigMaps().Informer()

	cmInformer.AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: checkConfigMapIsDeviceInfo,
		Handler: cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				cmDeviceHandler(nil, obj, constant.AddOperator)
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				if !reflect.DeepEqual(oldObj, newObj) {
					cmDeviceHandler(oldObj, newObj, constant.UpdateOperator)
				}
			},
			DeleteFunc: func(obj interface{}) {
				cmDeviceHandler(nil, obj, constant.DeleteOperator)
			},
		},
	})

	cmInformer.AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: checkConfigMapIsNodeInfo,
		Handler: cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				cmNodeHandler(nil, obj, constant.AddOperator)
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				if !reflect.DeepEqual(oldObj, newObj) {
					cmNodeHandler(oldObj, newObj, constant.UpdateOperator)
				}
			},
			DeleteFunc: func(obj interface{}) {
				cmNodeHandler(nil, obj, constant.DeleteOperator)
			},
		},
	})

	informerFactory.Start(informerCh)
}

func cmDeviceHandler(oldObj interface{}, newObj interface{}, operator string) {
	var oldDevInfo *constant.DeviceInfo
	var newDevInfo *constant.DeviceInfo
	var err error
	if oldObj != nil {
		oldDevInfo, err = device.ParseDeviceInfoCM(oldObj)
		if err != nil {
			hwlog.RunLog.Errorf("parse old cm error: %v", err)
			return
		}
	}
	newDevInfo, err = device.ParseDeviceInfoCM(newObj)
	if err != nil {
		hwlog.RunLog.Errorf("parse new cm error: %v", err)
		return
	}
	index := 0
	for _, cmFuncs := range cmDeviceFuncs {
		// different businesses use different data sources
		oldDevInfoForBusiness := oldDevInfo
		newDevInfoForBusiness := newDevInfo
		if index > 0 {
			oldDevInfoForBusiness = device.DeepCopy(oldDevInfo)
			newDevInfoForBusiness = device.DeepCopy(newDevInfo)
		}
		for _, cmFunc := range cmFuncs {
			cmFunc(oldDevInfoForBusiness, newDevInfoForBusiness, operator)
		}
		index++
	}
	if deviceCm, ok := newObj.(*v1.ConfigMap); ok {
		if _, ok := deviceCm.Data[constant.SwitchInfoCmKey]; ok {
			cmSwitchHandler(oldObj, newObj, operator)
		}
	}
}

func cmNodeHandler(oldObj interface{}, newObj interface{}, operator string) {
	var oldNodeInfo *constant.NodeInfo
	var newNodeInfo *constant.NodeInfo
	var err error
	if oldObj != nil {
		oldNodeInfo, err = node.ParseNodeInfoCM(oldObj)
		if err != nil {
			hwlog.RunLog.Errorf("parse old cm error: %v", err)
			return
		}
	}
	newNodeInfo, err = node.ParseNodeInfoCM(newObj)
	if err != nil {
		hwlog.RunLog.Errorf("parse new cm error: %v", err)
		return
	}
	index := 0
	for _, cmFuncs := range cmNodeFuncs {
		// different businesses use different data sources
		oldNodeInfoForBusiness := oldNodeInfo
		newNodeInfoForBusiness := newNodeInfo
		if index > 0 {
			oldNodeInfoForBusiness = node.DeepCopy(oldNodeInfo)
			newNodeInfoForBusiness = node.DeepCopy(newNodeInfo)
		}
		for _, cmFunc := range cmFuncs {
			cmFunc(oldNodeInfoForBusiness, newNodeInfoForBusiness, operator)
		}
		index++
	}
}

func cmSwitchHandler(oldObj interface{}, newObj interface{}, operator string) {
	var oldSwitchInfo *constant.SwitchInfo
	var newSwitchInfo *constant.SwitchInfo
	var err error
	if oldObj != nil {
		oldSwitchInfo, err = switchinfo.ParseSwitchInfoCM(oldObj)
		if err != nil {
			hwlog.RunLog.Errorf("parse old cm error: %v", err)
			return
		}
	}
	newSwitchInfo, err = switchinfo.ParseSwitchInfoCM(newObj)
	if err != nil {
		hwlog.RunLog.Errorf("parse new cm error: %v", err)
		return
	}
	index := 0
	for _, cmFuncs := range cmSwitchFuncs {
		// different businesses use different data sources
		oldSwitchInfoForBusiness := oldSwitchInfo
		newSwitchInfoForBusiness := newSwitchInfo
		if index > 0 {
			oldSwitchInfoForBusiness, err = switchinfo.DeepCopy(oldSwitchInfo)
			if err != nil {
				return
			}
			newSwitchInfoForBusiness, err = switchinfo.DeepCopy(newSwitchInfo)
			if err != nil {
				return
			}
		}
		for _, cmFunc := range cmFuncs {
			cmFunc(oldSwitchInfoForBusiness, newSwitchInfoForBusiness, operator)
		}
		index++
	}
}

// checkConfigMapIsDeviceInfo check if configmap is device info
func checkConfigMapIsDeviceInfo(obj interface{}) bool {
	return util.IsNSAndNameMatched(obj, constant.KubeNamespace, constant.DeviceInfoPrefix)
}

// checkConfigMapIsNodeInfo check if configmap is node info
func checkConfigMapIsNodeInfo(obj interface{}) bool {
	return util.IsNSAndNameMatched(obj, constant.DLNamespace, constant.NodeInfoPrefix)
}

func checkConfigMapIsSwitchInfo(obj interface{}) bool {
	return util.IsNSAndNameMatched(obj, constant.DLNamespace, constant.SwitchInfoPrefix)
}

// InitPubFaultCMInformer init cm informer for public fault
func InitPubFaultCMInformer() {
	informerFactory := informers.NewSharedInformerFactoryWithOptions(k8sClient.ClientSet, 0,
		informers.WithTweakListOptions(func(options *metav1.ListOptions) {
			options.LabelSelector = constant.CmConsumerPubFault + "=" + constant.CmConsumerValue
		}))
	cmInformer := informerFactory.Core().V1().ConfigMaps().Informer()

	cmInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			cmPubFaultHandler(nil, obj, constant.AddOperator)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			if !reflect.DeepEqual(oldObj, newObj) {
				cmPubFaultHandler(oldObj, newObj, constant.UpdateOperator)
			}
		},
		DeleteFunc: func(obj interface{}) {
			cmPubFaultHandler(nil, obj, constant.DeleteOperator)
		},
	})

	informerFactory.Start(informerCh)
}

func cmPubFaultHandler(oldObj, newObj interface{}, operator string) {
	var oldPubFaultInfo, newPubFaultInfo *api.PubFaultInfo
	var err error
	if oldObj != nil {
		oldPubFaultInfo, err = publicfault.ParsePubFaultCM(oldObj)
		if err != nil {
			hwlog.RunLog.Errorf("parse old public fault cm error: %v", err)
			return
		}
	}
	newPubFaultInfo, err = publicfault.ParsePubFaultCM(newObj)
	if err != nil {
		hwlog.RunLog.Errorf("parse new cm error: %v", err)
		return
	}
	index := 0
	for _, cmFuncs := range cmPubFaultFuncs {
		// index = 0, use original obj; index > 0, use deepcopy obj, keep the original obj
		// different cmFuncs use different data
		oldPubFaultForBusiness := oldPubFaultInfo
		newPubFaultForBusiness := newPubFaultInfo
		if index > 0 {
			oldPubFaultForBusiness = publicfault.DeepCopy(oldPubFaultInfo)
			newPubFaultForBusiness = publicfault.DeepCopy(newPubFaultInfo)
		}
		for _, cmFunc := range cmFuncs {
			cmFunc(oldPubFaultForBusiness, newPubFaultForBusiness, operator)
		}
		index++
	}
}

// InitPodGroupInformer is to init pod group informer
func InitPodGroupInformer() {
	vcClient := GetClientVolcano().ClientSet
	factory := externalversions.NewSharedInformerFactory(vcClient, 0)
	PodGroupInformer := factory.Scheduling().V1beta1().PodGroups()

	PodGroupInformer.Informer().SetWatchErrorHandler(func(r *cache.Reflector, err error) {
		hwlog.RunLog.Warnf("pg informer watcher err: %s", err.Error())
	})
	PodGroupInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			podGroupHandler(nil, obj, constant.AddOperator)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			if !reflect.DeepEqual(oldObj, newObj) {
				podGroupHandler(oldObj, newObj, constant.UpdateOperator)
			}
		},
		DeleteFunc: func(obj interface{}) {
			podGroupHandler(nil, obj, constant.DeleteOperator)
		},
	})
	factory.Start(wait.NeverStop)
}

func podGroupHandler(oldObj interface{}, newObj interface{}, operator string) {
	newPodGroup, ok := newObj.(*v1beta1.PodGroup)
	if !ok {
		return
	}
	var oldPodGroup *v1beta1.PodGroup
	if oldObj != nil {
		oldPodGroup, ok = oldObj.(*v1beta1.PodGroup)
		if !ok {
			return
		}
	}
	index := 0
	for _, podGroupFunc := range podGroupFuncs {
		// different businesses use different data sources
		oldPodGroupForBusiness := oldPodGroup
		newPodGroupForBusiness := newPodGroup
		if oldPodGroup != nil && newPodGroup != nil && index > 0 {
			oldPodGroup = oldPodGroup.DeepCopy()
			newPodGroup = newPodGroup.DeepCopy()
		}
		for _, pgFunc := range podGroupFunc {
			pgFunc(oldPodGroupForBusiness, newPodGroupForBusiness, operator)
		}
		index++
	}
}
