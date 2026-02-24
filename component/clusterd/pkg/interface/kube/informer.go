// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package kube a series of kube function
package kube

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"volcano.sh/apis/pkg/apis/batch/v1alpha1"
	"volcano.sh/apis/pkg/apis/scheduling/v1beta1"
	"volcano.sh/apis/pkg/client/informers/externalversions"

	"ascend-common/api"
	ascendv1 "ascend-common/api/ascend-operator/apis/batch/v1"
	ascendexternalversions "ascend-common/api/ascend-operator/client/informers/externalversions"
	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
	"clusterd/pkg/domain/device"
	"clusterd/pkg/domain/dpu"
	"clusterd/pkg/domain/node"
	"clusterd/pkg/domain/pingmeshconfig"
	"clusterd/pkg/domain/publicfault"
	"clusterd/pkg/domain/superpod"
	"clusterd/pkg/domain/switchinfo"
)

var (
	cmDeviceFuncs    = map[string][]func(*constant.DeviceInfo, *constant.DeviceInfo, string){}
	cmDpuFuncs       = map[string][]func(*constant.DpuInfoCM, *constant.DpuInfoCM, string){}
	cmNodeFuncs      = map[string][]func(*constant.NodeInfo, *constant.NodeInfo, string){}
	cmSwitchFuncs    = map[string][]func(*constant.SwitchInfo, *constant.SwitchInfo, string){}
	cmPubFaultFuncs  = map[string][]func(*api.PubFaultInfo, *api.PubFaultInfo, string){}
	acJobFuncs       = map[string][]func(*ascendv1.AscendJob, *ascendv1.AscendJob, string){}
	vcJobFuncs       = map[string][]func(*v1alpha1.Job, *v1alpha1.Job, string){}
	podGroupFuncs    = map[string][]func(*v1beta1.PodGroup, *v1beta1.PodGroup, string){}
	podFuncs         = map[string][]func(*v1.Pod, *v1.Pod, string){}
	nodeFuncs        = map[string][]func(*v1.Node, *v1.Node, string){}
	cmRankTableFuncs = map[string][]func(interface{}, interface{}, string){}
	informerCh       = make(chan struct{})

	// ping mesh configmap deal func
	cmPingMeshCMFuncs = map[string][]func(constant.ConfigPingMesh, constant.ConfigPingMesh, string){}
	indexers          = make(map[schema.GroupVersionKind]cache.Indexer)
)

// PodGVK pod groupVersionKind
func PodGVK() schema.GroupVersionKind {
	return v1.SchemeGroupVersion.WithKind("Pod")
}

// VcJobGVK volcano job groupVersionKind
func VcJobGVK() schema.GroupVersionKind {
	return v1alpha1.SchemeGroupVersion.WithKind("Job")
}

// PodGroupGVK pod groupVersionKind
func PodGroupGVK() schema.GroupVersionKind {
	return v1beta1.SchemeGroupVersion.WithKind("PodGroup")
}

// AcJobGVK ascend job groupVersionKind
func AcJobGVK() schema.GroupVersionKind {
	return ascendv1.SchemeGroupVersion.WithKind("AscendJob")
}

// ListObjects list objects from informer indexer
func ListObjects(gvks ...schema.GroupVersionKind) map[schema.GroupVersionKind][]interface{} {
	objects := make(map[schema.GroupVersionKind][]interface{})
	for _, gvk := range gvks {
		indexer, ok := indexers[gvk]
		if !ok {
			hwlog.RunLog.Warnf("get informer %s failed", gvk)
			continue
		}
		objects[gvk] = indexer.List()
	}

	return objects
}

// GetObject get object from informer indexer
func GetObject(gvk schema.GroupVersionKind, name string) (interface{}, error) {
	if indexer, ok := indexers[gvk]; ok {
		item, exist, err := indexer.GetByKey(name)
		if err != nil {
			return nil, err
		}
		if !exist {
			return nil, fmt.Errorf("get %s %s failed", gvk, name)
		}
		return item, nil
	}
	return nil, fmt.Errorf("gvk %s not found", gvk)
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
	cmDpuFuncs = map[string][]func(*constant.DpuInfoCM, *constant.DpuInfoCM, string){}
	cmNodeFuncs = map[string][]func(*constant.NodeInfo, *constant.NodeInfo, string){}
	cmSwitchFuncs = map[string][]func(*constant.SwitchInfo, *constant.SwitchInfo, string){}
	cmPubFaultFuncs = map[string][]func(*api.PubFaultInfo, *api.PubFaultInfo, string){}
	acJobFuncs = map[string][]func(*ascendv1.AscendJob, *ascendv1.AscendJob, string){}
	vcJobFuncs = map[string][]func(*v1alpha1.Job, *v1alpha1.Job, string){}
	podGroupFuncs = map[string][]func(*v1beta1.PodGroup, *v1beta1.PodGroup, string){}
	podFuncs = map[string][]func(*v1.Pod, *v1.Pod, string){}
	nodeFuncs = map[string][]func(*v1.Node, *v1.Node, string){}
	cmRankTableFuncs = map[string][]func(interface{}, interface{}, string){}
	cmPingMeshCMFuncs = map[string][]func(constant.ConfigPingMesh, constant.ConfigPingMesh, string){}
}

// AddACJobFunc add acJob func
func AddACJobFunc(business string, func1 ...func(*ascendv1.AscendJob, *ascendv1.AscendJob, string)) {
	if _, ok := acJobFuncs[business]; !ok {
		acJobFuncs[business] = []func(*ascendv1.AscendJob, *ascendv1.AscendJob, string){}
	}

	acJobFuncs[business] = append(acJobFuncs[business], func1...)
}

// AddVCJobFunc add vcJob func
func AddVCJobFunc(business string, func1 ...func(*v1alpha1.Job, *v1alpha1.Job, string)) {
	if _, ok := vcJobFuncs[business]; !ok {
		vcJobFuncs[business] = []func(*v1alpha1.Job, *v1alpha1.Job, string){}
	}

	vcJobFuncs[business] = append(vcJobFuncs[business], func1...)
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

// AddCmDpuFunc add dpu func, map by business
func AddCmDpuFunc(business string, func1 ...func(*constant.DpuInfoCM, *constant.DpuInfoCM, string)) {
	if _, ok := cmDpuFuncs[business]; !ok {
		cmDpuFuncs[business] = []func(*constant.DpuInfoCM, *constant.DpuInfoCM, string){}
	}

	cmDpuFuncs[business] = append(cmDpuFuncs[business], func1...)
}

// AddCmConfigPingMeshFunc add configmap func of pingmesh config
func AddCmConfigPingMeshFunc(business string,
	func1 ...func(constant.ConfigPingMesh, constant.ConfigPingMesh, string)) {
	if _, ok := cmPingMeshCMFuncs[business]; !ok {
		cmPingMeshCMFuncs[business] = []func(constant.ConfigPingMesh, constant.ConfigPingMesh, string){}
	}
	cmPingMeshCMFuncs[business] = append(cmPingMeshCMFuncs[business], func1...)
}

// AddCmPubFaultFunc add public fault deal func, map by business
func AddCmPubFaultFunc(business string, func1 ...func(*api.PubFaultInfo, *api.PubFaultInfo, string)) {
	if _, ok := cmPubFaultFuncs[business]; !ok {
		cmPubFaultFuncs[business] = []func(*api.PubFaultInfo, *api.PubFaultInfo, string){}
	}

	cmPubFaultFuncs[business] = append(cmPubFaultFuncs[business], func1...)
}

// AddCmRankTableFunc add rank table cm func, map by business
func AddCmRankTableFunc(business string, func1 ...func(interface{}, interface{}, string)) {
	if _, ok := cmRankTableFuncs[business]; !ok {
		cmRankTableFuncs[business] = []func(interface{}, interface{}, string){}
	}

	cmRankTableFuncs[business] = append(cmRankTableFuncs[business], func1...)
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

var podInformer cache.SharedIndexInformer
var nodeInformer cache.SharedIndexInformer

// InitPodAndNodeInformer init pod informer
func InitPodAndNodeInformer() {
	hwlog.RunLog.Info("start to init pod and node informer")
	factory := informers.NewSharedInformerFactoryWithOptions(k8sClient.ClientSet, 0)
	podInformer = factory.Core().V1().Pods().Informer()
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
	indexers[PodGVK()] = podInformer.GetIndexer()

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
	indexers[v1.SchemeGroupVersion.WithKind("Node")] = nodeInformer.GetIndexer()
	factory.Start(informerCh)
	factory.WaitForCacheSync(wait.NeverStop)
	initClusterDevice()
}

func initClusterDevice() {
	nodes := getNodesFromInformer()
	hwlog.RunLog.Infof("init cluster node length=%d", len(nodes))
	for _, n := range nodes {
		node.SaveNodeToCache(n)
		nodeDevice, superPodID := node.GetNodeDeviceAndSuperPodID(n)
		if nodeDevice == nil || superPodID == "" {
			continue
		}
		superPodID = strings.Trim(superPodID, " ")
		spIdIntValue, err := strconv.Atoi(superPodID)
		if spIdIntValue < 0 || err != nil {
			continue
		}
		superpod.SaveNode(superPodID, nodeDevice)
	}
}

func getNodesFromInformer() []*v1.Node {
	nodes := nodeInformer.GetStore().List()
	if len(nodes) == 0 {
		hwlog.RunLog.Warn("get empty node from informer")
		return nil
	}
	res := make([]*v1.Node, 0, len(nodes))
	for _, obj := range nodes {
		nodeResource, ok := obj.(*v1.Node)
		if !ok {
			hwlog.RunLog.Error("convert to Node error")
			continue
		}
		res = append(res, nodeResource)
	}
	return res
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

var cmInformer cache.SharedIndexInformer

// GetCmInformer get cm informer
func GetCmInformer() cache.SharedIndexInformer {
	return cmInformer
}

// InitCMInformer init configmap informer
func InitCMInformer() {
	hwlog.RunLog.Info("start to init CM informer")
	informerFactory := informers.NewSharedInformerFactoryWithOptions(k8sClient.ClientSet, 0,
		informers.WithTweakListOptions(func(options *metav1.ListOptions) {
			options.LabelSelector = api.CIMCMLabelKey + "=" + constant.CmConsumerValue
		}))
	cmInformer = informerFactory.Core().V1().ConfigMaps().Informer()

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
	AddRankTableEventHandler(&cmInformer)
	addPingMeshConfigEventHandler(&cmInformer)
	informerFactory.Start(informerCh)
}

func addPingMeshConfigEventHandler(cmInformer *cache.SharedIndexInformer) {
	(*cmInformer).AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: checkConfigMapIsPingMeshInfo,
		Handler: cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				cmPingMeshConfigHandler(nil, obj, constant.AddOperator)
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				if !reflect.DeepEqual(oldObj, newObj) {
					cmPingMeshConfigHandler(oldObj, newObj, constant.UpdateOperator)
				}
			},
			DeleteFunc: func(obj interface{}) {
				cmPingMeshConfigHandler(nil, obj, constant.DeleteOperator)
			},
		},
	})
}

func checkConfigMapIsPingMeshInfo(obj interface{}) bool {
	return util.IsNSAndNameMatched(obj, constant.PingMeshCMNamespace, constant.PingMeshConfigCm)
}

func cmPingMeshConfigHandler(oldObj interface{}, newObj interface{}, operator string) {
	var oldInfo constant.ConfigPingMesh
	var newInfo constant.ConfigPingMesh
	var err error
	if oldObj != nil {
		oldInfo, err = pingmeshconfig.ParseFaultNetworkInfoCM(oldObj)
		if err != nil {
			hwlog.RunLog.Errorf("parse old cm error: %v", err)
			return
		}
	}
	newInfo, err = pingmeshconfig.ParseFaultNetworkInfoCM(newObj)
	if err != nil {
		hwlog.RunLog.Errorf("parse new cm error: %v", err)
		return
	}
	index := 0
	for _, cmFuncs := range cmPingMeshCMFuncs {
		// different businesses use different data sources
		oldInfoForBusiness := oldInfo
		newInfoForBusiness := newInfo
		if index > 0 {
			oldInfoForBusiness = pingmeshconfig.DeepCopy(oldInfo)
			newInfoForBusiness = pingmeshconfig.DeepCopy(newInfo)
		}
		for _, cmFunc := range cmFuncs {
			cmFunc(oldInfoForBusiness, newInfoForBusiness, operator)
		}
		index++
	}
}

// AddRankTableEventHandler add rank table event handler for cmInformer
func AddRankTableEventHandler(cmInformer *cache.SharedIndexInformer) {
	(*cmInformer).AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: checkConfigMapIsEpRankTableInfo,
		Handler: cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				cmRankTableHandler(nil, obj, constant.AddOperator)
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				if !reflect.DeepEqual(oldObj, newObj) {
					cmRankTableHandler(oldObj, newObj, constant.UpdateOperator)
				}
			},
			DeleteFunc: func(obj interface{}) {
				cmRankTableHandler(nil, obj, constant.DeleteOperator)
			},
		},
	})
}

func cmRankTableHandler(oldObj interface{}, newObj interface{}, operator string) {
	for _, cmFuncs := range cmRankTableFuncs {
		for _, cmFunc := range cmFuncs {
			cmFunc(oldObj, newObj, operator)
		}
	}
}

func cmDeviceHandler(oldObj interface{}, newObj interface{}, operator string) {
	var oldDevInfo *constant.DeviceInfo
	var newDevInfo *constant.DeviceInfo
	var err error
	var oldCm *v1.ConfigMap
	if oldObj != nil {
		oldCmTemp, ok := oldObj.(*v1.ConfigMap)
		if !ok {
			hwlog.RunLog.Error("oldObj not device configmap")
			return
		}
		oldCm = oldCmTemp
		oldDevInfo, err = device.ParseDeviceInfoCM(oldCm)
		if err != nil {
			hwlog.RunLog.Errorf("parse old cm error: %v", err)
			return
		}
	}
	newCm, ok := newObj.(*v1.ConfigMap)
	if !ok {
		hwlog.RunLog.Error("newObj not device configmap")
		return
	}
	newDevInfo, err = device.ParseDeviceInfoCM(newCm)
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
	if _, ok := newCm.Data[api.SwitchInfoCMDataKey]; ok {
		cmSwitchHandler(oldCm, newCm, operator)
	}
	if _, ok := newCm.Data[api.DpuInfoCMDataKey]; ok {
		cmDpuHandler(oldCm, newCm, operator)
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

func cmSwitchHandler(oldCm *v1.ConfigMap, newCm *v1.ConfigMap, operator string) {
	var oldSwitchInfo *constant.SwitchInfo
	var newSwitchInfo *constant.SwitchInfo
	var err error
	if oldCm != nil {
		oldSwitchInfo, err = switchinfo.ParseSwitchInfoCM(oldCm)
		if err != nil {
			hwlog.RunLog.Errorf("parse old cm error: %v", err)
			return
		}
	}
	newSwitchInfo, err = switchinfo.ParseSwitchInfoCM(newCm)
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

func cmDpuHandler(oldCm *v1.ConfigMap, newCm *v1.ConfigMap, operator string) {
	var newDpuList *constant.DpuInfoCM
	var err error

	newDpuList, err = dpu.ParseDpuInfoCM(newCm)
	if err != nil {
		hwlog.RunLog.Errorf("%s parse new CM error: %v", api.DpuLogPrefix, err)
		return
	}
	index := 0
	for _, cmFuncs := range cmDpuFuncs {
		newDpuListForBusiness := newDpuList
		if index > 0 {
			newDpuListForBusiness = dpu.DeepCopy(newDpuList)
		}
		for _, cmFunc := range cmFuncs {
			cmFunc(&constant.DpuInfoCM{}, newDpuListForBusiness, operator)
		}
		index++
	}
}

// checkConfigMapIsDeviceInfo check if configmap is device info
func checkConfigMapIsDeviceInfo(obj interface{}) bool {
	return util.IsNSAndNameMatched(obj, api.KubeNS, constant.DeviceInfoPrefix)
}

// checkConfigMapIsNodeInfo check if configmap is node info
func checkConfigMapIsNodeInfo(obj interface{}) bool {
	return util.IsNSAndNameMatched(obj, api.DLNamespace, constant.NodeInfoPrefix)
}

// checkConfigMapIsEpRankTableInfo check if configmap is ep ranktable info
func checkConfigMapIsEpRankTableInfo(obj interface{}) bool {
	cm, ok := obj.(*v1.ConfigMap)
	if !ok {
		hwlog.RunLog.Errorf("Cannot convert to ConfigMap:%v", obj)
		return false
	}
	return strings.HasPrefix(cm.Name, constant.MindIeRanktablePrefix)
}

func checkConfigMapIsSwitchInfo(obj interface{}) bool {
	return util.IsNSAndNameMatched(obj, api.DLNamespace, constant.SwitchInfoPrefix)
}

// InitPubFaultCMInformer init cm informer for public fault
func InitPubFaultCMInformer() {
	hwlog.RunLog.Info("start to init PubFault CM informer")
	informerFactory := informers.NewSharedInformerFactoryWithOptions(k8sClient.ClientSet, 0,
		informers.WithTweakListOptions(func(options *metav1.ListOptions) {
			options.LabelSelector = api.PubFaultCMLabelKey + "=" + constant.CmConsumerValue
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

// InitACJobInformer is to init acJob informer
func InitACJobInformer() {
	hwlog.RunLog.Info("start to init ACjob informer")
	opClient := GetOperatorClient().ClientSet
	factory := ascendexternalversions.NewSharedInformerFactory(opClient, 0)
	acJobInformer := factory.Batch().V1().Jobs().Informer()

	acJobInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			acJobHandler(nil, obj, constant.AddOperator)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			if !reflect.DeepEqual(oldObj, newObj) {
				acJobHandler(oldObj, newObj, constant.UpdateOperator)
			}
		},
		DeleteFunc: func(obj interface{}) {
			acJobHandler(nil, obj, constant.DeleteOperator)
		},
	})
	indexers[AcJobGVK()] = acJobInformer.GetIndexer()
	factory.Start(wait.NeverStop)
	factory.WaitForCacheSync(wait.NeverStop)
}

func acJobHandler(oldObj interface{}, newObj interface{}, operator string) {
	newJob, ok := newObj.(*ascendv1.AscendJob)
	if !ok {
		return
	}
	var oldJob *ascendv1.AscendJob
	if oldObj != nil {
		oldJob, ok = oldObj.(*ascendv1.AscendJob)
		if !ok {
			return
		}
	}
	index := 0
	for _, ascendJobFunc := range acJobFuncs {
		// different businesses use different data sources
		oldJobForBusiness := oldJob
		newJobForBusiness := newJob
		if oldJob != nil && newJob != nil && index > 0 {
			oldJob = oldJob.DeepCopy()
			newJob = newJob.DeepCopy()
		}
		for _, ajFunc := range ascendJobFunc {
			ajFunc(oldJobForBusiness, newJobForBusiness, operator)
		}
		index++
	}
}

// InitVCJobInformer is to init vcJob informer
func InitVCJobInformer() {
	hwlog.RunLog.Info("start to init VCjob informer")
	vcClient := GetClientVolcano().ClientSet
	factory := externalversions.NewSharedInformerFactory(vcClient, 0)
	vcJobInformer := factory.Batch().V1alpha1().Jobs()

	vcJobInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			vcJobHandler(nil, obj, constant.AddOperator)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			if !reflect.DeepEqual(oldObj, newObj) {
				vcJobHandler(oldObj, newObj, constant.UpdateOperator)
			}
		},
		DeleteFunc: func(obj interface{}) {
			vcJobHandler(nil, obj, constant.DeleteOperator)
		},
	})
	indexers[VcJobGVK()] = vcJobInformer.Informer().GetIndexer()
	factory.Start(wait.NeverStop)
	factory.WaitForCacheSync(wait.NeverStop)
}

func vcJobHandler(oldObj interface{}, newObj interface{}, operator string) {
	newVCjob, ok := newObj.(*v1alpha1.Job)
	if !ok {
		return
	}
	var oldVCjob *v1alpha1.Job
	if oldObj != nil {
		oldVCjob, ok = oldObj.(*v1alpha1.Job)
		if !ok {
			return
		}
	}
	index := 0
	for _, vcJobFunc := range vcJobFuncs {
		// different businesses use different data sources
		oldVCjobForBusiness := oldVCjob
		newVCjobForBusiness := newVCjob
		if oldVCjob != nil && newVCjob != nil && index > 0 {
			oldVCjob = oldVCjob.DeepCopy()
			newVCjob = newVCjob.DeepCopy()
		}
		for _, pgFunc := range vcJobFunc {
			pgFunc(oldVCjobForBusiness, newVCjobForBusiness, operator)
		}
		index++
	}
}

// InitPodGroupInformer is to init pod group informer
func InitPodGroupInformer() {
	hwlog.RunLog.Info("start to init PodGroup informer")
	vcClient := GetClientVolcano().ClientSet
	factory := externalversions.NewSharedInformerFactory(vcClient, 0)
	PodGroupInformer := factory.Scheduling().V1beta1().PodGroups().Informer()

	PodGroupInformer.SetWatchErrorHandler(func(r *cache.Reflector, err error) {
		hwlog.RunLog.Warnf("pg informer watcher err: %s", err.Error())
	})
	PodGroupInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
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
	indexers[PodGroupGVK()] = PodGroupInformer.GetIndexer()
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
