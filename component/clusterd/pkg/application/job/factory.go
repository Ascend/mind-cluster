// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package job a series of job function
package job

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"huawei.com/npu-exporter/v6/common-utils/hwlog"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"volcano.sh/apis/pkg/apis/scheduling/v1beta1"
	"volcano.sh/apis/pkg/client/clientset/versioned"

	"clusterd/pkg/common/util"
)

func initCM(kubeClientSet kubernetes.Interface, job *jobModel) {
	data := make(map[string]string, 8)
	label := make(map[string]string, 2)

	data[ConfigmapKey] = DataValue
	data[JobName] = job.JobName
	data[ConfigmapOperator] = OperatorAdd
	data[FrameWork] = ModelFramework
	data[JobId] = job.JobUid
	data[DeleteTime] = "0"
	data[cmIndex] = "0"
	data[JobStatus] = StatusJobPending
	data[AddTime] = getUnixTime2String()
	label[Key910] = Val910
	label[ConfigmapLabel] = "true"
	putCM := &v1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("%s-%s", ConfigmapPrefix, job.JobName),
		Namespace: job.Namespace, Labels: label}, Data: data}
	if err := util.CreateOrUpdateConfigMap(kubeClientSet, putCM, fmt.Sprintf("%s-%s", ConfigmapPrefix, job.JobName),
		job.Namespace); err != nil {
		hwlog.RunLog.Errorf("initCM CreateOrUpdateConfigMap error: %s", err)
	}
}

// SyncJob handle jobs according to event type
func SyncJob(obj interface{}, eventType string, indexer cache.Indexer, agent *Agent) error {
	metaData, err := meta.Accessor(obj)
	if err != nil {
		hwlog.RunLog.Errorf("object has no meta: %v", err)
		return err
	}
	jobName, jobUid := getPGJobInfo(metaData)
	pgName := metaData.GetName()
	pgUid := string(metaData.GetUID())
	namespace := metaData.GetNamespace()
	modelFramework := metaData.GetLabels()
	ModelFramework = modelFramework[FrameWork]
	key := metaData.GetName() + "/" + eventType
	if len(metaData.GetNamespace()) > 0 {
		key = metaData.GetNamespace() + "/" + metaData.GetName() + "/" + eventType
	}
	hwlog.RunLog.Debugf("SyncJob start, current key is %v", key)
	_, exists, err := indexer.GetByKey(namespace + "/" + metaData.GetName())
	if err != nil {
		hwlog.RunLog.Errorf("failed to get obj from indexer: %s", key)
		return err
	}
	pg, isPg := obj.(*v1beta1.PodGroup)
	if !isPg {
		return fmt.Errorf("create job model by obj is not pg, %s ", key)
	}
	version, err := strconv.Atoi(pg.ResourceVersion)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get ResourceVersion: %s", pg.ResourceVersion)
		return err
	}
	model := &jobModel{key: key,
		Info: Info{JobUid: jobUid, Version: int32(version),
			CreationTimestamp: pg.CreationTimestamp, Namespace: namespace, JobName: jobName,
			PGName: pgName, PGUid: pgUid,
			Key:     namespace + "/" + metaData.GetName(),
			JobType: getPGType(pg)},
		replicas: pg.Spec.MinMember, devices: pg.Spec.MinResources}

	if !exists {
		if eventType == EventDelete {
			model.DeleteWorker(namespace, jobName, jobUid, agent)
			hwlog.RunLog.Infof("not exist + delete, eventType is %s, current key is %s", eventType, key)
			return nil
		}
		return fmt.Errorf("undefined condition, eventType is %s, current key is %s", eventType, key)
	}
	if err = HandlePGAddOrUpdateEvent(eventType, agent, model); err != nil {
		hwlog.RunLog.Errorf("handle pg add or update event failed, key: %s, err: %v", key, err)
		return err
	}
	return nil
}

func getPGType(pg *v1beta1.PodGroup) string {
	if pg == nil || len(pg.OwnerReferences) == 0 {
		return ""
	}
	return pg.OwnerReferences[0].Kind
}

// HandlePGAddOrUpdateEvent handle add or update event for pod group
func HandlePGAddOrUpdateEvent(eventType string, agent *Agent, model *jobModel) error {
	switch eventType {
	case EventAdd:
		hwlog.RunLog.Infof("exist + add, current job is %s/%s", model.Namespace, model.JobName)
		return model.AddEvent(agent)
	case EventUpdate:
		return model.EventUpdate(agent)
	default:
		return fmt.Errorf("undefined condition, eventType is %s", eventType)
	}
}

// NewVCClientK8s new vcjob client
func NewVCClientK8s() (*versioned.Clientset, error) {
	clientCfg, err := clientcmd.BuildConfigFromFlags("", "")
	if err != nil {
		hwlog.RunLog.Errorf("build vcClient config err: %v", err)
		return nil, err
	}
	vcClientSet, err := versioned.NewForConfig(clientCfg)
	if err != nil {
		hwlog.RunLog.Errorf("get vcClient err: %v", err)
		return nil, err
	}
	return vcClientSet, nil
}

// NewConfig is to create new config
func NewConfig() *Config {
	myConfig := &Config{
		DryRun:           dryRun,
		DisplayStatistic: displayStatistic,
		CmCheckInterval:  cmCheckInterval,
		CmCheckTimeout:   cmCheckTimeout,
	}
	return myConfig
}

func shouldCmDelete(cm v1.ConfigMap, pgList *v1beta1.PodGroupList) bool {
	cmJobUid := cm.Data[JobId]

	for _, pg := range pgList.Items {
		_, pgJobUid := getPGJobInfo(&pg)
		if pgJobUid == cmJobUid {
			return false
		}
	}
	return true
}

func deleteJobSummaryCM(kubeClientSet kubernetes.Interface, firstTimeDel bool, vcClient *versioned.Clientset) error {
	cms, err := kubeClientSet.CoreV1().ConfigMaps("").List(context.TODO(),
		metav1.ListOptions{LabelSelector: ConfigmapWholeLabel})
	if err != nil {
		return fmt.Errorf("failed to list configmap, err: %v", err)
	}
	for _, cm := range cms.Items {
		value, ok := cm.Data[ConfigmapOperator]
		if !ok {
			continue
		}
		timeNow := time.Now().Unix()
		deleteTime, err := strconv.Atoi(cm.Data[DeleteTime])
		if err != nil {
			return fmt.Errorf("failed to convert delete time to int for configmap %s/%s, err: %v",
				cm.Namespace, cm.Name, err)
		}
		if deleteTime == 0 && firstTimeDel {
			pgList, err := vcClient.SchedulingV1beta1().PodGroups("").List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				hwlog.RunLog.Errorf("get pg list error, err: %v", err)
				continue
			}
			if !shouldCmDelete(cm, pgList) {
				continue
			}
			if err := kubeClientSet.CoreV1().ConfigMaps(cm.Namespace).Delete(context.TODO(),
				cm.Name, metav1.DeleteOptions{}); err != nil {
				hwlog.RunLog.Errorf("failed to delete configmap %s/%s, err: %v", cm.Namespace, cm.Name, err)
				continue
			}
			hwlog.RunLog.Infof("configmap: %v %v deleted", cm.Namespace, cm.Name)
			continue
		}
		interval := int(timeNow) - deleteTime
		if value == OperatorDelete && interval > deleteCMInterval {
			if err := kubeClientSet.CoreV1().ConfigMaps(cm.Namespace).Delete(context.TODO(),
				cm.Name, metav1.DeleteOptions{}); err != nil {
				hwlog.RunLog.Errorf("failed to delete configmap %s/%s, err: %v", cm.Namespace, cm.Name, err)
				continue
			}
			hwlog.RunLog.Infof("delete configmap %s/%s succeed", cm.Namespace, cm.Name)
		}
	}
	return nil
}

// HandleDeleteJobSummaryCM make sure job summary cm could be deleted
func HandleDeleteJobSummaryCM(ctx context.Context, kubeClientSet kubernetes.Interface, vcClient *versioned.Clientset) {
	hwlog.RunLog.Info("delete job summary cm goroutine started")
	firstTimeDel := true
	for {
		select {
		case _, ok := <-ctx.Done():
			if !ok {
				hwlog.RunLog.Info("catch stop signal channel closed")
			}
			hwlog.RunLog.Info("handle delete job-summary configmap stop")
			return
		default:
			if err := deleteJobSummaryCM(kubeClientSet, firstTimeDel, vcClient); err != nil {
				hwlog.RunLog.Errorf("failed to delete job summary configmap, err: %v", err)
			}
			firstTimeDel = false
			time.Sleep(time.Duration(deleteCMCyclicTime) * time.Second)
		}
	}
}
