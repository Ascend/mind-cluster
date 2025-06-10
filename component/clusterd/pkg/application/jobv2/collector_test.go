// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

//go:build !race

// Package jobv2 a series of job test function
package jobv2

import (
	"context"
	"sync"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"volcano.sh/apis/pkg/apis/scheduling/v1beta1"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/pod"
	"clusterd/pkg/domain/podgroup"
)

const (
	jobName1     = "job1"
	jobNameSpace = "default"
	jobUid1      = "123"

	podName1      = "pod1"
	podName2      = "pod2"
	podNameSpace1 = "default"
	podUid1       = "123"
	podUid2       = "456"
	podRank1      = "1"

	vcJobKey = "job"
)

func init() {
	hwlog.InitRunLogger(&hwlog.LogConfig{OnlyToStdout: true}, context.Background())
}

func TestPodGroupCollector(t *testing.T) {
	convey.Convey("test PodGroupCollector", t, func() {
		oldPGInfo := getDemoPodGroup(jobName1, jobNameSpace, jobUid1)
		newPGInfo := getDemoPodGroup(jobName1, jobNameSpace, jobUid1)
		convey.Convey("test add podGroup message", func() {
			uniqueQueue = sync.Map{}
			PodGroupCollector(oldPGInfo, newPGInfo, queueOperatorAdd)
			convey.So(podgroup.GetPodGroup(jobUid1).Namespace, convey.ShouldNotBeBlank)
			value, ok := uniqueQueue.Load(podgroup.GetJobKeyByPG(newPGInfo))
			convey.So(ok, convey.ShouldEqual, true)
			convey.So(value, convey.ShouldEqual, queueOperatorAdd)
		})
		convey.Convey("test update podGroup message", func() {
			uniqueQueue = sync.Map{}
			PodGroupCollector(oldPGInfo, newPGInfo, queueOperatorUpdate)
			value, ok := uniqueQueue.Load(podgroup.GetJobKeyByPG(newPGInfo))
			convey.So(ok, convey.ShouldEqual, true)
			convey.So(value, convey.ShouldEqual, queueOperatorUpdate)
		})
		convey.Convey("test delete podGroup message", func() {
			uniqueQueue = sync.Map{}
			PodGroupCollector(oldPGInfo, newPGInfo, queueOperatorDelete)
			convey.So(podgroup.GetPodGroup(jobUid1).Namespace, convey.ShouldBeBlank)
			value, ok := uniqueQueue.Load(podgroup.GetJobKeyByPG(newPGInfo))
			convey.So(ok, convey.ShouldEqual, true)
			convey.So(value, convey.ShouldEqual, queueOperatorPreDelete)
		})
		convey.Convey("test error podGroup message", func() {
			uniqueQueue = sync.Map{}
			PodGroupCollector(oldPGInfo, newPGInfo, "illegal")
			_, ok := uniqueQueue.Load(podgroup.GetJobKeyByPG(newPGInfo))
			convey.So(ok, convey.ShouldEqual, false)
		})
	})
}

func TestPodIsControllerOrCoordinator(t *testing.T) {
	convey.Convey("test PodIsControllerOrCoordinator", t, func() {
		convey.Convey("when pod is controller or coordinator", func() {
			demoPod := getDemoPod(podName1, podNameSpace1, podUid1, podRank1)
			demoPod.ObjectMeta.Labels = map[string]string{
				constant.MindIeAppTypeLabelKey: constant.ControllerAppType,
			}
			result := checkPodIsControllerOrCoordinator(demoPod)
			convey.So(result, convey.ShouldBeTrue)
			demoPod.ObjectMeta.Labels = map[string]string{
				constant.MindIeAppTypeLabelKey: constant.CoordinatorAppType,
			}
			result = checkPodIsControllerOrCoordinator(demoPod)
			convey.So(result, convey.ShouldBeTrue)
		})
		convey.Convey("when pod is neither controller nor coordinator", func() {
			demoPod := getDemoPod(podName1, podNameSpace1, podUid1, podRank1)
			demoPod.ObjectMeta.Labels = map[string]string{
				constant.MindIeAppTypeLabelKey: constant.ServerAppType,
			}
			result := checkPodIsControllerOrCoordinator(demoPod)
			convey.So(result, convey.ShouldBeFalse)
		})
		convey.Convey("when pod has no app type label", func() {
			demoPod := getDemoPod(podName1, podNameSpace1, podUid1, podRank1)
			result := checkPodIsControllerOrCoordinator(demoPod)
			convey.So(result, convey.ShouldBeFalse)
		})
		convey.Convey("when object is not a pod", func() {
			nonPod := "not a pod"
			result := checkPodIsControllerOrCoordinator(nonPod)
			convey.So(result, convey.ShouldBeFalse)
		})
	})
}

func getDemoPodGroup(jobName, nameSpace, jobUid string) *v1beta1.PodGroup {
	podGroupInfo := &v1beta1.PodGroup{}
	podGroupInfo.Name = jobName
	podGroupInfo.Namespace = nameSpace
	isControlle := true
	owner := metav1.OwnerReference{
		Name:       jobName,
		Controller: &isControlle,
		Kind:       vcJobKey,
		UID:        types.UID(jobUid)}
	podGroupInfo.SetOwnerReferences([]metav1.OwnerReference{owner})
	return podGroupInfo
}

func TestPodCollector(t *testing.T) {
	convey.Convey("test PodCollector", t, func() {
		oldPod := getDemoPod(podName1, podNameSpace1, podUid1, podRank1)
		newPod := getDemoPod(podName1, podNameSpace1, podUid1, podRank1)
		convey.Convey("test add pod message", func() {
			PodCollector(oldPod, newPod, constant.AddOperator)
			convey.So(len(pod.GetPodByJobId(jobUid1)), convey.ShouldEqual, 1)
			convey.So(len(pod.GetSimplePodByJobId(jobUid1)), convey.ShouldEqual, 1)
			value, ok := uniqueQueue.Load(jobUid1)
			convey.So(ok, convey.ShouldEqual, true)
			convey.So(value, convey.ShouldEqual, queueOperatorUpdate)
		})
		convey.Convey("test update pod message", func() {
			PodCollector(oldPod, newPod, constant.UpdateOperator)
			convey.So(len(pod.GetPodByJobId(jobUid1)), convey.ShouldEqual, 1)
			convey.So(len(pod.GetSimplePodByJobId(jobUid1)), convey.ShouldEqual, 1)
			value, ok := uniqueQueue.Load(jobUid1)
			convey.So(ok, convey.ShouldEqual, true)
			convey.So(value, convey.ShouldEqual, queueOperatorUpdate)
		})
		convey.Convey("test delete pod message", func() {
			PodCollector(oldPod, newPod, constant.DeleteOperator)
			convey.So(len(pod.GetPodByJobId(jobUid1)), convey.ShouldEqual, 0)
			convey.So(len(pod.GetSimplePodByJobId(jobUid1)), convey.ShouldEqual, 0)
			value, ok := uniqueQueue.Load(jobUid1)
			convey.So(ok, convey.ShouldEqual, true)
			convey.So(value, convey.ShouldEqual, queueOperatorUpdate)
		})
		convey.Convey("test error pod message", func() {
			uniqueQueue = sync.Map{}
			PodCollector(oldPod, newPod, "illegal")
			_, ok := uniqueQueue.Load(jobUid1)
			convey.So(ok, convey.ShouldEqual, false)
		})
	})
}

func getDemoPod(name, nameSpace, podUid, podRank string) *v1.Pod {
	p := &v1.Pod{}
	p.Name = name
	p.Namespace = nameSpace
	p.UID = types.UID(podUid)
	isControlle := true
	owner := metav1.OwnerReference{
		Name:       jobName1,
		Controller: &isControlle,
		Kind:       vcJobKey,
		UID:        types.UID(jobUid1)}
	p.SetOwnerReferences([]metav1.OwnerReference{owner})
	p.Annotations = make(map[string]string)
	p.Annotations[api.PodRankIndexAnno] = podRank
	return p
}
func getDemoPodWithStatus(name, nameSpace, podUid string, status v1.PodPhase) v1.Pod {
	p := v1.Pod{}
	p.Name = name
	p.Namespace = nameSpace
	p.UID = types.UID(podUid)
	isControlle := true
	owner := metav1.OwnerReference{
		Name:       jobName1,
		Controller: &isControlle,
		Kind:       vcJobKey,
		UID:        types.UID(jobUid1)}
	p.SetOwnerReferences([]metav1.OwnerReference{owner})
	p.Status.Phase = status
	return p
}
