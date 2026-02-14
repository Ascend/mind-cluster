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
	ptFramework            = "pytorch"
	envName                = "testEnv"
	envValue               = "true"

	jobUid1  = "123"
	jobUid2  = "456"
	jobName1 = "job1"
	vcJobKey = "job"
	pgName1  = "pg1"
	sharedIp = "127.0.0.1"

	nodeName1 = "node1"
	nodeName2 = "node2"
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
		convey.Convey("when pod info is nil, cache should remain unchanged", func() {
			oldLen := len(podManager.podMap)
			SavePod(nil)
			newLen := len(podManager.podMap)
			convey.So(oldLen, convey.ShouldEqual, newLen)
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
		convey.Convey("when pod info is nil, cache should remain unchanged", func() {
			oldLen := len(podManager.podMap)
			DeletePod(nil)
			newLen := len(podManager.podMap)
			convey.So(oldLen, convey.ShouldEqual, newLen)
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
