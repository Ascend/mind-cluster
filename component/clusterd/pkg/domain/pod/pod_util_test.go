// Copyright (c) Huawei Technologies Co., Ltd. 2024-2025. All rights reserved.

//go:build !race

// Package pod a series of pod test function
package pod

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"ascend-common/api"
)

const (
	testJobId            = "testJobId"
	testPodRank          = "0"
	testIllegalCardRank  = "-1"
	testDeviceNumPerNode = 8
	testPodName          = "testPodName"
	testPodUid           = "testPodUid"
)

func TestGetJobKeyByPod(t *testing.T) {
	convey.Convey("test GetJobKeyByPod", t, func() {
		podDemo1 := getDemoPod(podName1, podNameSpace1, podUid1)
		convey.Convey("when pod getOwnerReferences exists, jobUid should be exits", func() {
			convey.So(GetJobKeyByPod(podDemo1), convey.ShouldEqual, jobUid1)
		})
		convey.Convey("when pod getOwnerReferences is nil, jobUid should be nil", func() {
			podDemo1.OwnerReferences = []metav1.OwnerReference{}
			convey.So(GetJobKeyByPod(podDemo1), convey.ShouldEqual, "")
		})
	})
}

func TestGetPodKey(t *testing.T) {
	convey.Convey("test GetPodKey", t, func() {
		podDemo1 := getDemoPod(podName1, podNameSpace1, podUid1)
		convey.Convey("when pod is nil, podKey should be nil", func() {
			convey.So(GetPodKey(nil), convey.ShouldEqual, "")
		})
		convey.Convey("when pod exists, podKey should be exits", func() {
			convey.So(GetPodKey(podDemo1), convey.ShouldEqual, podUid1)
		})
	})
}

func TestGetPGInfo(t *testing.T) {
	convey.Convey("test GetPGInfo", t, func() {
		podDemo1 := getDemoPod(podName1, podNameSpace1, podUid1)
		convey.Convey("when pod is nil, jobName should be nil", func() {
			jobName, pgName, namespace := GetPGInfo(nil)
			convey.So(jobName, convey.ShouldEqual, "")
			convey.So(pgName, convey.ShouldEqual, "")
			convey.So(namespace, convey.ShouldEqual, "")
		})
		convey.Convey("when pod exists, podKey should be exits", func() {
			jobName, pgName, namespace := GetPGInfo(podDemo1)
			convey.So(jobName, convey.ShouldEqual, jobName1)
			convey.So(pgName, convey.ShouldEqual, pgName1)
			convey.So(namespace, convey.ShouldEqual, podNameSpace1)
		})
	})
}

func TestGetSharedTorIpByPod(t *testing.T) {
	convey.Convey("test GetSharedTorIpByPod", t, func() {
		podJobMap := make(map[string]v1.Pod)
		convey.Convey("when pods is nil, sharedTorIp should be nil", func() {
			sharedTorIp := GetSharedTorIpByPod(podJobMap)
			convey.So(sharedTorIp, convey.ShouldEqual, "")
		})
		convey.Convey("when pods is exists and sharedTorIp annotation is exists, sharedTorIp should be exists",
			func() {
				podDemo1 := getDemoPod(podName1, podNameSpace1, podUid1)
				annotationMap := podDemo1.Annotations
				annotationMap[torTag] = sharedTor
				annotationMap[torIpTag] = sharedIp
				podJobMap[podUid1] = *podDemo1
				sharedTorIp := GetSharedTorIpByPod(podJobMap)
				convey.So(sharedTorIp, convey.ShouldContainSubstring, sharedIp)
			})

	})
}

func TestGetEnvByPod(t *testing.T) {
	convey.Convey("test GetEnvByPod", t, func() {
		podJobMap := make(map[string]v1.Pod)
		convey.Convey("when pods is nil, env should be nil", func() {
			convey.So(GetEnvByPod(podJobMap, envName), convey.ShouldEqual, "")
		})
		podDemo1 := getDemoPod(podName1, podNameSpace1, podUid1)
		podJobMap[podUid1] = *podDemo1
		convey.Convey("when pod's container is nil, env should be nil", func() {
			convey.So(GetEnvByPod(podJobMap, envName), convey.ShouldEqual, "")
		})
		container := v1.Container{}
		podDemo1.Spec.Containers = append(podDemo1.Spec.Containers, container)
		podJobMap[podUid1] = *podDemo1
		convey.Convey("when container's env is nil, env should be nil", func() {
			convey.So(GetEnvByPod(podJobMap, envName), convey.ShouldEqual, "")
		})
		envVar := v1.EnvVar{Name: envName, Value: envValue}
		container.Env = append(container.Env, envVar)
		podDemo1.Spec.Containers = append(podDemo1.Spec.Containers, container)
		podJobMap[podUid1] = *podDemo1
		convey.Convey("when container's env is exists, env should be exists", func() {
			convey.So(GetEnvByPod(podJobMap, envName), convey.ShouldEqual, envValue)
		})
	})
}

func TestInitRankTableByPod(t *testing.T) {
	convey.Convey("test InitRankTableByPod", t, func() {
		convey.Convey("when replicas is 0, podJobMap is empty, completedPodNum should be 0", func() {
			rankTable, completedPodNum := InitRankTableByPod(map[string]v1.Pod{}, 0)
			convey.So(completedPodNum, convey.ShouldEqual, 0)
			convey.So(rankTable.ServerCount, convey.ShouldEqual, "")
		})
		convey.Convey("when replicas is 1, podJobMap is empty, completedPodNum should be 0", func() {
			rankTable, completedPodNum := InitRankTableByPod(map[string]v1.Pod{}, 1)
			convey.So(completedPodNum, convey.ShouldEqual, 0)
			convey.So(rankTable.ServerCount, convey.ShouldEqual, "0")
		})
		convey.Convey("when replicas is 1, podJobMap is completed, completedPodNum should be 1", func() {
			podJobMap := make(map[string]v1.Pod)
			podDemo1 := getDemoPod(podName1, podNameSpace1, podUid1)
			podJobMap[podUid1] = *podDemo1
			rankTable, completedPodNum := InitRankTableByPod(podJobMap, 1)
			convey.So(completedPodNum, convey.ShouldEqual, 1)
			convey.So(rankTable.ServerCount, convey.ShouldEqual, "1")
		})
		convey.Convey("when replicas is 1, podJobMap nodeRank is illegal, completedPodNum should be 0", func() {
			podJobMap := make(map[string]v1.Pod)
			podDemo1 := getDemoPod(podName1, podNameSpace1, podUid1)
			podDemo1.Annotations[podRankIndexKey] = errorPodRankIndexKey
			podJobMap[podUid1] = *podDemo1
			rankTable, completedPodNum := InitRankTableByPod(podJobMap, 1)
			convey.So(completedPodNum, convey.ShouldEqual, 0)
			convey.So(rankTable.ServerCount, convey.ShouldEqual, "0")
		})
	})
}

func TestGetPodDeviceNumByJobId(t *testing.T) {
	convey.Convey("test GetPodDeviceNumByJobId", t, func() {
		convey.Convey("when podMap is nil, deviceNum should be 0", func() {
			convey.So(GetPodDeviceNumByJobId(jobUid1), convey.ShouldEqual, 0)
		})
		convey.Convey("when podMap is exists, deviceNum should be 1", func() {
			podDemo1 := getDemoPod(podName1, podNameSpace1, podUid1)
			SavePod(podDemo1)
			defer DeletePod(podDemo1)
			convey.So(GetPodDeviceNumByJobId(jobUid1), convey.ShouldEqual, 1)
		})
	})
}

func TestGetPodByRankIndex(t *testing.T) {
	convey.Convey("test GetPodByRankIndex", t, func() {
		convey.Convey("when podMap is nil, pod should be nil", func() {
			convey.So(GetPodByRankIndex(jobUid1, defaultPodRankIndexKey).Name, convey.ShouldEqual, "")
		})
		convey.Convey("when podMap is exists,but rankIndex is error, pod should be nil", func() {
			podDemo1 := getDemoPod(podName1, podNameSpace1, podUid1)
			SavePod(podDemo1)
			defer DeletePod(podDemo1)
			convey.So(GetPodByRankIndex(jobUid1, errorPodRankIndexKey).Name, convey.ShouldEqual, "")
		})
		convey.Convey("when podMap is exists,but rankIndex is right, pod should be exists", func() {
			podDemo1 := getDemoPod(podName1, podNameSpace1, podUid1)
			SavePod(podDemo1)
			defer DeletePod(podDemo1)
			convey.So(GetPodByRankIndex(jobUid1, defaultPodRankIndexKey).Name, convey.ShouldEqual, podName1)
		})
	})
}

func TestGetModelFramework(t *testing.T) {
	convey.Convey("test GetModelFramework", t, func() {
		convey.Convey("when podMap is nil, framework should be nil", func() {
			convey.So(GetModelFramework(map[string]v1.Pod{}), convey.ShouldEqual, "")
		})
		convey.Convey("when podMap is exists, framework should be exists", func() {
			podJobMap := make(map[string]v1.Pod)
			podDemo1 := getDemoPod(podName1, podNameSpace1, podUid1)
			podDemo1.Labels[podLabelKey] = ptFramework
			podJobMap[podUid1] = *podDemo1
			convey.So(GetModelFramework(podJobMap), convey.ShouldEqual, ptFramework)
		})
	})
}

func TestDeviceAllocateIsCompleted(t *testing.T) {
	convey.Convey("test DeviceAllocateIsCompleted", t, func() {
		convey.Convey("when containers is nil, return false", func() {
			podDemo1 := getDemoPod(podName1, podNameSpace1, podUid1)
			convey.So(DeviceAllocateIsCompleted(*podDemo1), convey.ShouldBeFalse)
		})
		convey.Convey("when resources do not include huawei.com, return true", func() {
			podDemo1 := getDemoPod(podName1, podNameSpace1, podUid1)
			container := v1.Container{}
			podDemo1.Spec.Containers = append(podDemo1.Spec.Containers, container)
			convey.So(DeviceAllocateIsCompleted(*podDemo1), convey.ShouldBeTrue)
		})
		convey.Convey("when resources include huawei.com, annotation podDeviceKey is exists, return true",
			func() {
				podDemo1 := getDemoPod(podName1, podNameSpace1, podUid1)
				container := v1.Container{}
				container.Resources.Limits = v1.ResourceList{
					api.ResourceNamePrefix: resource.Quantity{},
				}
				podDemo1.Spec.Containers = append(podDemo1.Spec.Containers, container)
				convey.So(DeviceAllocateIsCompleted(*podDemo1), convey.ShouldBeTrue)
			})
		convey.Convey("when resources include huawei.com, annotation podDeviceKey is not exists, return false",
			func() {
				podDemo1 := getDemoPod(podName1, podNameSpace1, podUid1)
				container := v1.Container{}
				container.Resources.Limits = v1.ResourceList{
					api.ResourceNamePrefix: resource.Quantity{},
				}
				podDemo1.Spec.Containers = append(podDemo1.Spec.Containers, container)
				delete(podDemo1.Annotations, podDeviceKey)
				convey.So(DeviceAllocateIsCompleted(*podDemo1), convey.ShouldBeFalse)
			})
	})
}

func TestGetPGByPod(t *testing.T) {
	convey.Convey("test GetPGByPod", t, func() {
		convey.Convey("when podMap is nil, jobName should be nil", func() {
			jobName, pgName, namespace := GetPGByPod(jobUid1)
			convey.So(jobName, convey.ShouldEqual, "")
			convey.So(pgName, convey.ShouldEqual, "")
			convey.So(namespace, convey.ShouldEqual, "")
		})
		convey.Convey("when podMap is exists, jobName should be exists", func() {
			podDemo1 := getDemoPod(podName1, podNameSpace1, podUid1)
			SavePod(podDemo1)
			defer DeletePod(podDemo1)
			jobName, pgName, namespace := GetPGByPod(jobUid1)
			convey.So(jobName, convey.ShouldEqual, jobName1)
			convey.So(pgName, convey.ShouldEqual, pgName1)
			convey.So(namespace, convey.ShouldEqual, podNameSpace1)
		})
	})
}

func TestGetPodRankAndPodUid(t *testing.T) {
	convey.Convey("test IsSliceContain", t, func() {
		convey.Convey("case pod.GetPodDeviceNumByJobId small than 1", func() {
			patch := gomonkey.ApplyFuncReturn(GetPodDeviceNumByJobId, 0)
			defer patch.Reset()
			podRank, podUid := GetPodRankAndPodUid(testJobId, "")
			convey.So(podRank, convey.ShouldEqual, "")
			convey.So(podUid, convey.ShouldEqual, "")
		})
		convey.Convey("case illegal card rank", func() {
			patch := gomonkey.ApplyFuncReturn(GetPodDeviceNumByJobId, testDeviceNumPerNode)
			defer patch.Reset()
			podRank, podUid := GetPodRankAndPodUid(testJobId, testIllegalCardRank)
			convey.So(podRank, convey.ShouldEqual, "")
			convey.So(podUid, convey.ShouldEqual, "")
		})
		convey.Convey("case empty pod name", func() {
			patch := gomonkey.ApplyFuncReturn(GetPodDeviceNumByJobId, testDeviceNumPerNode).
				ApplyFuncReturn(GetPodByRankIndex, v1.Pod{})
			defer patch.Reset()
			podRank, podUid := GetPodRankAndPodUid(testJobId, "0")
			convey.So(podRank, convey.ShouldEqual, testPodRank)
			convey.So(podUid, convey.ShouldEqual, "")
		})
		convey.Convey("normal case", func() {
			patch := gomonkey.ApplyFuncReturn(GetPodDeviceNumByJobId, testDeviceNumPerNode).
				ApplyFuncReturn(GetPodByRankIndex, v1.Pod{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{Name: testPodName, UID: testPodUid},
					Spec:       v1.PodSpec{},
					Status:     v1.PodStatus{},
				})
			defer patch.Reset()
			podRank, podUid := GetPodRankAndPodUid(testJobId, "0")
			convey.So(podRank, convey.ShouldEqual, testPodRank)
			convey.So(podUid, convey.ShouldEqual, testPodUid)
		})
	})
}
