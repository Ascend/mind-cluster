// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package kube a series of kube function
package kube

import (
	"context"
	"reflect"
	"strings"

	"huawei.com/npu-exporter/v6/common-utils/hwlog"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"volcano.sh/apis/pkg/client/clientset/versioned"
	"volcano.sh/apis/pkg/client/informers/externalversions"
	"volcano.sh/apis/pkg/client/informers/externalversions/scheduling/v1beta1"

	"clusterd/pkg/application/job"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
	"clusterd/pkg/domain/device"
	"clusterd/pkg/domain/node"
	"clusterd/pkg/domain/switchinfo"
)

var (
	podSyncFuncs  []func(*v1.Pod, *v1.Pod, string)
	podFuncs      []func(*v1.Pod, *v1.Pod, string)
	cmDeviceFuncs = map[string][]func(*constant.DeviceInfo, *constant.DeviceInfo, string){}
	cmNodeFuncs   = map[string][]func(*constant.NodeInfo, *constant.NodeInfo, string){}
	cmSwitchFuncs = map[string][]func(*constant.SwitchInfo, *constant.SwitchInfo, string){}
	informerCh    = make(chan struct{})
	// JobMgr is a mgr of job
	JobMgr *job.Agent
	// PGInformer is pod group informer
	PGInformer v1beta1.PodGroupInformer
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

// InitPodInformer init pod informer
func InitPodInformer() {
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
	factory.Start(informerCh)
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
	for _, podFunc := range podFuncs {
		go podFunc(newPod, oldPod, operator)
	}
	for _, podFunc := range podSyncFuncs {
		podFunc(newPod, oldPod, operator)
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
		if index > 0 {
			oldDevInfo = device.DeepCopy(oldDevInfo)
			newDevInfo = device.DeepCopy(newDevInfo)
		}
		for _, cmFunc := range cmFuncs {
			cmFunc(oldDevInfo, newDevInfo, operator)
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
		if index > 0 {
			oldNodeInfo = node.DeepCopy(oldNodeInfo)
			newNodeInfo = node.DeepCopy(newNodeInfo)
		}
		for _, cmFunc := range cmFuncs {
			cmFunc(oldNodeInfo, newNodeInfo, operator)
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
		if index > 0 {
			oldSwitchInfo, err = switchinfo.DeepCopy(oldSwitchInfo)
			if err != nil {
				return
			}
			newSwitchInfo, err = switchinfo.DeepCopy(newSwitchInfo)
			if err != nil {
				return
			}
		}
		for _, cmFunc := range cmFuncs {
			cmFunc(oldSwitchInfo, newSwitchInfo, operator)
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

func checkVolcanoExist(vcClient *versioned.Clientset) bool {
	_, err := vcClient.SchedulingV1beta1().PodGroups(constant.DefaultNamespace).Get(context.Background(),
		constant.TestName, metav1.GetOptions{})
	if err != nil && strings.Contains(err.Error(), constant.NoResourceOnServer) {
		return false
	}
	return true
}

// InitPGInformer is to init pod group informer
func InitPGInformer(ctx context.Context, jobSrv JobService) {
	vcClient := GetClientVolcano().ClientSet
	factory := externalversions.NewSharedInformerFactory(vcClient, 0)
	PGInformer = factory.Scheduling().V1beta1().PodGroups()

	if !checkVolcanoExist(vcClient) {
		hwlog.RunLog.Warn("Volcano not exist, please deploy Volcano and restart ClusterD.")
		return
	}

	cacheIndexer := PGInformer.Informer().GetIndexer()
	var err error
	JobMgr, err = job.NewAgent(k8sClient.ClientSet, job.NewConfig(), vcClient)
	if err != nil {
		hwlog.RunLog.Errorf("create agent err: %v", err)
		return
	}
	go job.HandleDeleteJobSummaryCM(ctx, k8sClient.ClientSet, vcClient)
	PGInformer.Informer().SetWatchErrorHandler(func(r *cache.Reflector, err error) {
		hwlog.RunLog.Warnf("pg informer watcher err: %s", err.Error())
	})
	PGInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if err = job.SyncJob(obj, constant.AddOperator, cacheIndexer, JobMgr); err != nil {
				hwlog.RunLog.Errorf("error to syncing EventAdd: %v", err)
			}
		},
		UpdateFunc: func(old, new interface{}) {
			if reflect.DeepEqual(old, new) {
				return
			}
			if err = job.SyncJob(new, constant.UpdateOperator, cacheIndexer, JobMgr); err != nil {
				hwlog.RunLog.Errorf("error to syncing EventUpdate: %v", err)
			}
		},
		DeleteFunc: func(obj interface{}) {
			if err = job.SyncJob(obj, constant.DeleteOperator, cacheIndexer, JobMgr); err != nil {
				hwlog.RunLog.Errorf("error to syncing EventDelete: %v", err)
			}
			deleteJobForJobService(jobSrv, obj)
		},
	})
	factory.Start(wait.NeverStop)
}

func deleteJobForJobService(jobSrv JobService, obj interface{}) {
	metaData, err := meta.Accessor(obj)
	if err != nil {
		hwlog.RunLog.Errorf("object has no meta: %v", err)
		return
	}
	ownerReferences := metaData.GetOwnerReferences()
	var jobUid string
	for _, v := range ownerReferences {
		if string(v.Kind) == constant.JobRefKind || string(v.Kind) == constant.AscendJobRefKind {
			jobUid = string(v.UID)
			break
		}
	}
	jobSrv.DeleteJob(jobUid)
}
