// Copyright (c) Huawei Technologies Co., Ltd. 2024-2025. All rights reserved.

//go:build !race

// Package pod a series of pod test function
package pod

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
)

func init() {
	hwlog.InitRunLogger(&hwlog.LogConfig{OnlyToStdout: true}, context.Background())
}

const (
	podName1               = "pod1"
	podName2               = "pod2"
	podNameSpace1          = "default"
	podUid1                = "123"
	podUid2                = "456"
	defaultPodRankIndexKey = "0"
	errorPodRankIndexKey   = "-1"
	defaultPodDeviceKey    = `{"server_id":"127.0.0.1","devices":[{"device_id":"0"}]}`
	podDeviceKey2          = `{"server_id":"127.0.0.1","devices":[{"device_id":"2"}]}`
	podDeviceKey5          = `{"server_id":"127.0.0.1","devices":[{"device_id":"5"}]}`
	ptFramework            = "pytorch"
	envName                = "testEnv"
	envValue               = "true"

	dev0     = "0"
	dev2     = "2"
	dev5     = "5"
	jobUid1  = "123"
	jobUid2  = "456"
	jobUid3  = "789"
	jobName1 = "job1"
	jobName2 = "job2"
	jobName3 = "job3"
	vcJobKey = "job"
	pgName1  = "pg1"
	sharedIp = "127.0.0.1"

	nodeName1 = "node1"
	nodeName2 = "node2"
	nodeName3 = "node3"
	nodeIp1   = "192.168.1.1"
	nodeSn1   = "sn1"

	len3 = 3
)

func TestSavePod(t *testing.T) {
	convey.Convey("test SavePod", t, func() {
		podDemo1 := getDemoPod(podName1, podNameSpace1, podUid1)
		convey.Convey("when pod cache less than maxPodNum, cache should save success", func() {
			SavePod(podDemo1)
			defer DeletePod(podDemo1)
			podMap := GetPodByJobId(jobUid1)
			convey.So(len(podMap), convey.ShouldEqual, 1)
			convey.So(len(GetSimplePodByJobId(jobUid1)), convey.ShouldEqual, 1)
		})
	})
}

func TestDeletePod(t *testing.T) {
	convey.Convey("test DeletePod", t, func() {
		podDemo1 := getDemoPod(podName1, podNameSpace1, podUid1)
		convey.Convey("when pod cache less than maxPodNum, cache should save success", func() {
			SavePod(podDemo1)
			DeletePod(podDemo1)
			podMap := GetPodByJobId(jobUid1)
			convey.So(len(podMap), convey.ShouldEqual, 0)
			convey.So(len(GetSimplePodByJobId(jobUid1)), convey.ShouldEqual, 0)
		})
	})
}

func getDemoPod(name, nameSpace, podUid string) *v1.Pod {
	p := &v1.Pod{}
	p.Name = name
	p.Namespace = nameSpace
	p.UID = types.UID(podUid)
	p.Spec.NodeName = nodeName1
	isControlle := true
	owner := metav1.OwnerReference{
		Name:       jobName1,
		Controller: &isControlle,
		Kind:       vcJobKey,
		UID:        types.UID(jobUid1)}
	p.SetOwnerReferences([]metav1.OwnerReference{owner})
	annotation := map[string]string{
		podGroupKey:          pgName1,
		api.PodRankIndexAnno: defaultPodRankIndexKey,
		api.Pod910DeviceAnno: defaultPodDeviceKey,
	}
	p.SetAnnotations(annotation)
	label := map[string]string{
		vcJobNameKey: jobName1,
	}
	p.SetLabels(label)
	return p
}

func TestGetPodsByNodeName(t *testing.T) {
	convey.Convey("test GetPodsByNodeName", t, func() {
		podDemo1 := getDemoPod(podName1, podNameSpace1, podUid1)
		convey.Convey("the pod exist on node", func() {
			SavePod(podDemo1)
			defer DeletePod(podDemo1)
			pods, exist := GetPodsByNodeName(nodeName1)
			convey.So(exist, convey.ShouldBeTrue)
			convey.So(len(pods), convey.ShouldEqual, 1)
		})
		convey.Convey("the pod does not exist on node", func() {
			SavePod(podDemo1)
			defer DeletePod(podDemo1)
			pods, exist := GetPodsByNodeName(nodeName2)
			convey.So(exist, convey.ShouldBeFalse)
			convey.So(len(pods), convey.ShouldEqual, 0)
		})
	})
}

func TestGetPodByPodId(t *testing.T) {
	convey.Convey("test GetPodByPodId", t, func() {
		podDemo1 := getDemoPod(podName1, podNameSpace1, podUid1)
		convey.Convey("the pod exist", func() {
			SavePod(podDemo1)
			defer DeletePod(podDemo1)
			pod, exist := GetPodByPodId(podUid1)
			convey.So(exist, convey.ShouldBeTrue)
			convey.So(pod.Name, convey.ShouldEqual, podName1)
		})
		convey.Convey("the pod does not exist", func() {
			SavePod(podDemo1)
			defer DeletePod(podDemo1)
			pod, exist := GetPodByPodId(podUid2)
			convey.So(exist, convey.ShouldBeFalse)
			convey.So(pod.Name, convey.ShouldEqual, "")
		})
	})
}

func TestGetPodByJobIdAndPodName(t *testing.T) {
	convey.Convey("test GetPodByJobIdAndPodName", t, func() {
		podDemo1 := getDemoPod(podName1, podNameSpace1, podUid1)
		convey.Convey("the pod exist", func() {
			SavePod(podDemo1)
			defer DeletePod(podDemo1)
			pod, exist := GetPodByJobIdAndPodName(jobUid1, podName1)
			convey.So(exist, convey.ShouldBeTrue)
			convey.So(pod.UID, convey.ShouldEqual, podUid1)
		})
		convey.Convey("the pod does not exist", func() {
			SavePod(podDemo1)
			defer DeletePod(podDemo1)
			pod, exist := GetPodByJobIdAndPodName(jobUid2, podName2)
			convey.So(exist, convey.ShouldBeFalse)
			convey.So(pod.UID, convey.ShouldEqual, "")
		})
	})
}

func TestUnscheduledPod(t *testing.T) {
	convey.Convey("test UnscheduledPod", t, func() {
		podDemo1 := getDemoPod(podName1, podNameSpace1, podUid1)
		podDemo1.Spec.NodeName = ""
		convey.Convey("the pod does not exist", func() {
			SavePod(podDemo1)
			defer DeletePod(podDemo1)
			_, exist := GetPodsByNodeName(podDemo1.Spec.NodeName)
			convey.So(exist, convey.ShouldBeFalse)
		})
	})
}

func TestNonOwnerReference(t *testing.T) {
	convey.Convey("test NonOwnerReference", t, func() {
		podDemo1 := getDemoPod(podName1, podNameSpace1, podUid1)
		podDemo1.SetOwnerReferences([]metav1.OwnerReference{})
		convey.Convey("the pod does not exist", func() {
			SavePod(podDemo1)
			defer DeletePod(podDemo1)
			result := GetSimplePodByJobId(GetJobKeyByPod(podDemo1))
			convey.So(len(result), convey.ShouldBeZeroValue)
		})
	})
}

func TestAddPodInCache(t *testing.T) {
	convey.Convey("test addPodInCache", t, func() {
		podDemo1 := getDemoPod(podName1, podNameSpace1, podUid1)
		convey.Convey("add pod to cache success", func() {
			podKey, jobKey := addPodInCache(podDemo1)
			defer deletePodInCache(podDemo1)
			convey.So(podKey, convey.ShouldEqual, podUid1)
			convey.So(jobKey, convey.ShouldEqual, jobUid1)
			convey.So(len(podManager.podMap), convey.ShouldEqual, 1)
			convey.So(len(podManager.nodePodMap[nodeName1]), convey.ShouldEqual, 1)
			convey.So(len(podManager.jobPodMap[jobUid1]), convey.ShouldEqual, 1)
		})
		convey.Convey("add pod without nodeName to cache success", func() {
			podDemo2 := getDemoPod(podName2, podNameSpace1, podUid2)
			podDemo2.Spec.NodeName = ""
			podKey, jobKey := addPodInCache(podDemo2)
			defer deletePodInCache(podDemo2)
			convey.So(podKey, convey.ShouldEqual, podUid2)
			convey.So(jobKey, convey.ShouldEqual, jobUid1)
			_, exist := podManager.nodePodMap[""]
			convey.So(exist, convey.ShouldBeFalse)
		})
		convey.Convey("add pod without ownerReference to cache success", func() {
			podDemo2 := getDemoPod(podName2, podNameSpace1, podUid2)
			podDemo2.SetOwnerReferences([]metav1.OwnerReference{})
			podKey, jobKey := addPodInCache(podDemo2)
			defer deletePodInCache(podDemo2)
			convey.So(podKey, convey.ShouldEqual, podUid2)
			convey.So(jobKey, convey.ShouldEqual, "")
			_, exist := podManager.jobPodMap[""]
			convey.So(exist, convey.ShouldBeFalse)
		})
	})
}

func TestUpdatePodInCache(t *testing.T) {
	convey.Convey("test updatePodInCache", t, func() {
		podDemo1 := getDemoPod(podName1, podNameSpace1, podUid1)
		podDemo2 := getDemoPod(podName1, podNameSpace1, podUid1)
		podDemo2.Spec.NodeName = nodeName2
		convey.Convey("update pod in cache success", func() {
			addPodInCache(podDemo1)
			defer deletePodInCache(podDemo2)
			podKey, jobKey := updatePodInCache(podDemo1, podDemo2)
			convey.So(podKey, convey.ShouldEqual, podUid1)
			convey.So(jobKey, convey.ShouldEqual, jobUid1)
			convey.So(len(podManager.podMap), convey.ShouldEqual, 1)
			convey.So(podManager.podMap[podUid1].Spec.NodeName, convey.ShouldEqual, nodeName2)
			_, exist := podManager.nodePodMap[nodeName1]
			convey.So(exist, convey.ShouldBeFalse)
			_, exist = podManager.nodePodMap[nodeName2]
			convey.So(exist, convey.ShouldBeTrue)
		})
	})
}

func TestDeletePodInCache(t *testing.T) {
	convey.Convey("test deletePodInCache", t, func() {
		podDemo1 := getDemoPod(podName1, podNameSpace1, podUid1)
		convey.Convey("delete pod from cache success", func() {
			addPodInCache(podDemo1)
			podKey, jobKey := deletePodInCache(podDemo1)
			convey.So(podKey, convey.ShouldEqual, podUid1)
			convey.So(jobKey, convey.ShouldEqual, jobUid1)
			convey.So(len(podManager.podMap), convey.ShouldEqual, 0)
			_, exist := podManager.nodePodMap[nodeName1]
			convey.So(exist, convey.ShouldBeFalse)
			_, exist = podManager.jobPodMap[jobUid1]
			convey.So(exist, convey.ShouldBeFalse)
		})
		convey.Convey("delete pod that does not exist should not panic", func() {
			podDemo2 := getDemoPod(podName2, podNameSpace1, podUid2)
			podKey, jobKey := deletePodInCache(podDemo2)
			convey.So(podKey, convey.ShouldEqual, podUid2)
			convey.So(jobKey, convey.ShouldEqual, jobUid1)
		})
	})
}
