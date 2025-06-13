// Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.

// Package epranktable for generating global rank table
package epranktable

import (
	"fmt"

	"k8s.io/api/core/v1"
	"k8s.io/client-go/util/workqueue"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/job"
)

// epGlobalRankTableManager ep global rank table manager
var epGlobalRankTableManager *RankTableManager = nil

// RankTableManager ep rank table manager
type RankTableManager struct {
	// rankTableQueue message queue for storing generate global ranktable messages
	rankTableQueue workqueue.RateLimitingInterface
	// HandlerRankTable a callback function that push global ranktable information
	HandlerRankTable func(string, string) (bool, error)
}

// GenerateGlobalRankTableMessage generate global rank table message
type GenerateGlobalRankTableMessage struct {
	// the task ID of the reasoning task, globally unique
	JobId string
	// the namespace of the reasoning task
	Namespace string
}

func init() {
	// Initialize speed limiter (exponential backoff strategy)
	rateLimiter := workqueue.NewItemExponentialFailureRateLimiter(
		constant.QueueInitDelay, // Initial delay time
		constant.QueueMaxDelay,  // Maximum delay time
	)
	epGlobalRankTableManager = &RankTableManager{
		rankTableQueue: workqueue.NewRateLimitingQueue(rateLimiter),
	}
}

// GetEpGlobalRankTableManager get ep global rank table manager
func GetEpGlobalRankTableManager() *RankTableManager {
	return epGlobalRankTableManager
}

// GetRankTableMessageQueue get rank table message queue
func GetRankTableMessageQueue() workqueue.RateLimitingInterface {
	return epGlobalRankTableManager.rankTableQueue
}

// ConsumerForQueue a consumer for queue
func (rm *RankTableManager) ConsumerForQueue() {
	for {
		quit := rm.handleGenerateGlobalRankTableMessage()
		if quit {
			break
		}
	}
}

func (rm *RankTableManager) handleGenerateGlobalRankTableMessage() bool {
	item, quit := rm.rankTableQueue.Get()
	if quit {
		return true
	}
	defer rm.rankTableQueue.Done(item)

	// determine if the maximum number of retries has been exceeded
	if rm.rankTableQueue.NumRequeues(item) >= constant.MaxRetryTime {
		hwlog.RunLog.Errorf("max retries exceeded, will forget item: %v", item)
		rm.rankTableQueue.Forget(item)
		return false
	}

	message, err := rm.getMessageInfo(item)
	if err != nil {
		hwlog.RunLog.Errorf("get message info failed, err: %v", err)
		return false
	}

	// get pd deployment mode
	pdDeploymentMode, err := job.GetPdDeploymentMode(message.JobId, message.Namespace, constant.ServerAppType)
	if err != nil {
		hwlog.RunLog.Errorf("get pd deployment mode failed, err: %v", err)
		rm.rankTableQueue.AddRateLimited(item)
		return false
	}

	// global ranktable information that needs to be pushed to grpc
	var globalRankTableInfo string
	// indicate whether a retry is required
	var retry bool
	globalRankTableInfo, retry = GeneratePdDeployModeRankTable(message, pdDeploymentMode)

	if retry {
		rm.rankTableQueue.AddRateLimited(item)
	} else {
		rm.pushGlobalRankTable(message, globalRankTableInfo)
	}
	return false
}

// GeneratePdDeployModeRankTable generate single node or cross node pd deploy mode rank table
func GeneratePdDeployModeRankTable(message *GenerateGlobalRankTableMessage, pdDeploymentMode string) (string, bool) {
	a2RankTableList, err := GetA2RankTableList(message)
	if err != nil {
		hwlog.RunLog.Errorf("get a2 rank table list failed, err: %v", err)
		return "", true
	}
	serverGroup0, err := GenerateServerGroup0Or1(message, constant.CoordinatorAppType)
	if err != nil {
		hwlog.RunLog.Errorf("generate server group 0 failed, err: %v", err)
		return "", true
	}
	serverGroup1, err := GenerateServerGroup0Or1(message, constant.ControllerAppType)
	if err != nil {
		hwlog.RunLog.Errorf("generate server group 1 failed, err: %v", err)
		return "", true
	}
	var globalRankTableInfo string
	globalRankTableInfo, err = getGlobalRankTableInfo(a2RankTableList, serverGroup0, serverGroup1, pdDeploymentMode)
	if err != nil {
		hwlog.RunLog.Errorf("get global rank table info failed, err: %v", err)
		return "", true
	}
	hwlog.RunLog.Infof("generate global rank table info success, jobId: %s", message.JobId)
	hwlog.RunLog.Debugf("global rank table info: %s", globalRankTableInfo)
	return globalRankTableInfo, false
}

// getMessageInfo get message info
func (rm *RankTableManager) getMessageInfo(item interface{}) (*GenerateGlobalRankTableMessage, error) {
	// received the message to generate a global ranking table
	message, ok := item.(*GenerateGlobalRankTableMessage)
	if !ok {
		rm.rankTableQueue.Forget(item)
		return message, fmt.Errorf("cannot convert to GenerateGlobalRankTableMessage:%v", item)
	}

	// When grpc is first registered, a message will be sent stating that the namespace is empty
	// and needs to be obtained based on the jobId
	var err error
	if message.Namespace == "" {
		message.Namespace, err = job.GetNamespaceByJobIdAndAppType(message.JobId, constant.ServerAppType)
		if err != nil {
			rm.rankTableQueue.AddRateLimited(item)
			return message, err
		}
	}

	return message, nil
}

// pushGlobalRankTable push global rank table by grpc
func (rm *RankTableManager) pushGlobalRankTable(message *GenerateGlobalRankTableMessage, globalRankTableInfo string) {
	jobId := message.JobId
	if rm.HandlerRankTable == nil {
		hwlog.RunLog.Warnf("grpc HandlerRankTable is nil")
		rm.rankTableQueue.Forget(message)
		return
	}
	_, err := rm.HandlerRankTable(jobId, globalRankTableInfo)
	if err != nil {
		hwlog.RunLog.Errorf("push global rank table to grpc failed, jobId: %s, will retry", jobId)
		rm.rankTableQueue.AddRateLimited(message)
		return
	}
	// push successful
	rm.rankTableQueue.Forget(message)
}

// InformerHandler collects generate global ranktable message and add to queue
func InformerHandler(oldObj, newObj interface{}, operator string) {
	var jobId = ""
	var namespace = ""
	var exist bool

	switch newObj.(type) {
	case *v1.ConfigMap:
		changedCm, _ := newObj.(*v1.ConfigMap)
		jobId, exist = changedCm.Labels[constant.MindIeJobIdLabelKey]
		if !exist {
			hwlog.RunLog.Errorf("jobId is not exist in labels")
			return
		}
		namespace = changedCm.Namespace
	case *v1.Pod:
		changedPod, _ := newObj.(*v1.Pod)
		jobId, exist = changedPod.Labels[constant.MindIeJobIdLabelKey]
		if !exist {
			hwlog.RunLog.Errorf("jobId is not exist in labels")
			return
		}
		namespace = changedPod.Namespace
	default:
		hwlog.RunLog.Errorf("unknown object type")
		return
	}
	epGlobalRankTableManager.rankTableQueue.Add(&GenerateGlobalRankTableMessage{
		JobId:     jobId,
		Namespace: namespace,
	})
}
