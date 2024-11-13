// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package job a series of job function
package job

import (
	"fmt"
	"reflect"
	"time"

	"huawei.com/npu-exporter/v6/common-utils/hwlog"
	apiCoreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"volcano.sh/apis/pkg/client/clientset/versioned"
)

// NewAgent to create an agent
func NewAgent(kubeClientSet kubernetes.Interface, config *Config, vcClient *versioned.Clientset) (*Agent, error) {
	// create pod informer factory
	temp, newErr := labels.NewRequirement(Key910, selection.In, []string{Val910B, Val910})
	if newErr != nil {
		hwlog.RunLog.Errorf("Newagent %s", newErr)
		return nil, newErr
	}

	labelSelector := temp.String()
	podInformerFactory := informers.NewSharedInformerFactoryWithOptions(kubeClientSet,
		time.Second*defaultResyncTime, informers.WithTweakListOptions(func(options *metav1.ListOptions) {
			options.LabelSelector = labelSelector
		}))

	agent := &Agent{
		podsInformer:  podInformerFactory.Core().V1().Pods().Informer(),
		podsIndexer:   podInformerFactory.Core().V1().Pods().Informer().GetIndexer(),
		KubeClientSet: kubeClientSet,
		BsWorker:      make(map[string]PodWorker),
		Config:        config,
		vcClient:      vcClient,
	}

	agent.podsInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			agent.doWork(obj, EventAdd)
		},
		UpdateFunc: func(old, new interface{}) {
			if !reflect.DeepEqual(old, new) {
				agent.doWork(new, EventUpdate)
			}
		},
		DeleteFunc: func(obj interface{}) {
			agent.doWork(obj, EventDelete)
		},
	})

	podInformerFactory.Start(wait.NeverStop)

	return agent, nil
}

func (agent *Agent) doWork(obj interface{}, eventType string) {
	podKeyInfo, err := getPodInfo(obj, eventType)
	if err != nil {
		hwlog.RunLog.Errorf("get pod key error %s", err)
		return
	}
	// get pod obj from lister
	tmpObj, podExist, err := agent.podsIndexer.GetByKey(podKeyInfo.namespace + "/" + podKeyInfo.name)
	if err != nil {
		hwlog.RunLog.Errorf("syncing '%s' failed: failed to get obj from indexer", podKeyInfo)
		return
	}
	podCacheAgent := agent.GetBsWorker(podKeyInfo.jobId)
	hwlog.RunLog.Debugf("worker: %+v", podCacheAgent)
	if podCacheAgent == nil {
		if !podExist {
			hwlog.RunLog.Warnf("syncing '%s' terminated: current obj is no longer exist",
				podKeyInfo.podInfo2String())
			return
		}
		// if someone create a single 910 pod without a job, how to handle?
		hwlog.RunLog.Infof("syncing '%s' delayed: corresponding job worker may be uninitialized",
			podKeyInfo.podInfo2String())
		return
	}
	if podKeyInfo.eventType == EventDelete {
		if err = podCacheAgent.handlePodDelEvent(podKeyInfo); err != nil {
			// only logs need to be recorded.
			hwlog.RunLog.Errorf("handleDeleteEvent error, error is %s", err)
		}
		return
	}
	// if worker exist but pod not exist, try again except delete event
	if !podExist {
		return
	}
	pod, ok := tmpObj.(*apiCoreV1.Pod)
	if !ok {
		hwlog.RunLog.Errorf("pod transform failed")
		return
	}
	// if worker exist && pod exist, need check some special scenarios
	hwlog.RunLog.Debugf("successfully synced '%s'", podKeyInfo)
	podCacheAgent.doPodWork(pod, podKeyInfo)
	if podKeyInfo.eventType == EventUpdate {
		if err = podCacheAgent.UpdateCMWhenJobEnd(podKeyInfo); err != nil {
			hwlog.RunLog.Errorf("UpdateCMWhenJobEnd error, error is %s", err)
		}
	}
}

// GetBsWorker return a bs Worker
func (agent *Agent) GetBsWorker(bsKey string) PodWorker {
	agent.RwMutex.RLock()
	defer agent.RwMutex.RUnlock()
	if worker, exist := agent.BsWorker[bsKey]; exist {
		return worker
	}
	return nil
}

// BsExist is to check whether bsKey exist
func (agent *Agent) BsExist(bsKey string) bool {
	agent.RwMutex.RLock()
	defer agent.RwMutex.RUnlock()
	if _, exist := agent.BsWorker[bsKey]; exist {
		return true
	}
	return false
}

// BsLength return BsWorker length
func (agent *Agent) BsLength(bsKey string) int {
	agent.RwMutex.RLock()
	defer agent.RwMutex.RUnlock()
	return len(agent.BsWorker)
}

// SetBsWorker is set bs worker
func (agent *Agent) SetBsWorker(bsKey string, worker PodWorker) {
	agent.RwMutex.Lock()
	defer agent.RwMutex.Unlock()
	agent.BsWorker[bsKey] = worker
}

// DeleteBsWorker delete bs worker by key
func (agent *Agent) DeleteBsWorker(bsKey string) {
	agent.RwMutex.Lock()
	defer agent.RwMutex.Unlock()
	delete(agent.BsWorker, bsKey)
}

// UpdateJobDeviceStatus update node's device healthy status
func (agent *Agent) UpdateJobDeviceStatus(nodeName string, networkUnhealthyCards, unHealthyCards string) {
	agent.RwMutex.RLock()
	defer agent.RwMutex.RUnlock()
	for _, worker := range agent.BsWorker {
		worker.UpdateJobDeviceHealthyStatus(nodeName, networkUnhealthyCards, unHealthyCards)
	}
}

// UpdateJobNodeStatus update node's node healthy status
func (agent *Agent) UpdateJobNodeStatus(nodeName string, healthy bool) {
	agent.RwMutex.RLock()
	defer agent.RwMutex.RUnlock()
	for _, worker := range agent.BsWorker {
		worker.UpdateJobNodeHealthyStatus(nodeName, healthy)
	}
}

// TODO 如何判断job是uce容忍的
func (agent *Agent) JobTolerateUceFault(jobId string) bool {
	return true
}

func (agent *Agent) GetNodeAndDeviceFromJobIdAndRankId(jobId, rankId string) (string, string, error) {
	agent.RwMutex.RLock()
	defer agent.RwMutex.RUnlock()
	worker, ok := agent.BsWorker[jobId]
	if !ok {
		return "", "", fmt.Errorf("not find node and device from jobId %v", jobId)
	}
	rankTable := worker.GetWorkerInfo().CMData
	for _, server := range rankTable.GetServerList() {
		for _, dev := range server.DeviceList {
			if dev.RankID == rankId {
				return server.ServerName, dev.DeviceID, nil
			}
		}
	}
	return "", "", fmt.Errorf("not find node and device from jobId %v and rankid %v", jobId, rankId)
}

func getWorkName(labels map[string]string) string {
	if label, ok := labels["volcano.sh/job-name"]; ok {
		return label
	}
	if label, ok := labels["job-name"]; ok {
		return label
	}
	return ""
}

func getPodInfo(obj interface{}, eventType string) (*podIdentifier, error) {
	metaData, err := meta.Accessor(obj)
	if err != nil {
		return nil, fmt.Errorf("object has no meta: %v", err)
	}
	labelsMaps := metaData.GetLabels()
	jobId := ""
	for _, owner := range metaData.GetOwnerReferences() {
		if *owner.Controller {
			jobId = string(owner.UID)
			break
		}
	}
	podPathInfo := &podIdentifier{
		namespace: metaData.GetNamespace(),
		name:      metaData.GetName(),
		jobName:   getWorkName(labelsMaps),
		eventType: eventType,
		UID:       string(metaData.GetUID()),
		jobId:     jobId,
	}
	return podPathInfo, nil
}

func (p *podIdentifier) podInfo2String() string {
	return fmt.Sprintf("namespace:%s,name:%s,jobName:%s,eventType:%s", p.namespace, p.name, p.jobName, p.eventType)
}
