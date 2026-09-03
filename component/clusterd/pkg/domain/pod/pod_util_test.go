// Copyright (c) Huawei Technologies Co., Ltd. 2024-2025. All rights reserved.

//go:build !race

// Package pod a series of pod test function
package pod

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"ascend-common/api"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/node"
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
		convey.Convey("when input info is nil, jobUid should be nil", func() {
			convey.So(GetJobKeyByPod(nil), convey.ShouldEqual, "")
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
		convey.Convey("when pod is nil, jobName should be nil", func() {
			jobName, pgName, namespace := GetPGInfo(nil)
			convey.So(jobName, convey.ShouldEqual, "")
			convey.So(pgName, convey.ShouldEqual, "")
			convey.So(namespace, convey.ShouldEqual, "")
		})
		convey.Convey("get pg info failed, pg key does not exist", func() {
			podDemo := getDemoPod(podName1, podNameSpace1, podUid1)
			podDemo.Annotations = map[string]string{api.PodRankIndexAnno: defaultPodRankIndexKey}
			jobName, pgName, namespace := GetPGInfo(podDemo)
			convey.So(jobName, convey.ShouldEqual, "")
			convey.So(pgName, convey.ShouldEqual, "")
			convey.So(namespace, convey.ShouldEqual, "")
		})
		convey.Convey("get pg info failed, job name label does not exist", func() {
			podDemo := getDemoPod(podName1, podNameSpace1, podUid1)
			podDemo.Labels = map[string]string{"": ""}
			jobName, pgName, namespace := GetPGInfo(podDemo)
			convey.So(jobName, convey.ShouldEqual, "")
			convey.So(pgName, convey.ShouldEqual, "")
			convey.So(namespace, convey.ShouldEqual, "")
		})
	})
}

func TestGetSharedTorIpByPod(t *testing.T) {
	convey.Convey("test GetSharedTorIpByPod", t, func() {
		podsInJob := make(map[string]v1.Pod)
		convey.Convey("when pods is nil, sharedTorIp should be nil", func() {
			sharedTorIp := GetSharedTorIpByPod(podsInJob)
			convey.So(sharedTorIp, convey.ShouldEqual, "")
		})
		convey.Convey("when pods is exists and sharedTorIp annotation is exists, sharedTorIp should be exists",
			func() {
				podDemo1 := getDemoPod(podName1, podNameSpace1, podUid1)
				annotationMap := podDemo1.Annotations
				annotationMap[torTag] = sharedTor
				annotationMap[torIpTag] = sharedIp
				podsInJob[podUid1] = *podDemo1
				sharedTorIp := GetSharedTorIpByPod(podsInJob)
				convey.So(sharedTorIp, convey.ShouldContainSubstring, sharedIp)
			})

	})
}

func TestGetEnvByPod(t *testing.T) {
	convey.Convey("test GetEnvByPod", t, func() {
		podsInJob := make(map[string]v1.Pod)
		convey.Convey("when pods is nil, env should be nil", func() {
			convey.So(GetEnvByPod(podsInJob, envName), convey.ShouldEqual, "")
		})
		podDemo1 := getDemoPod(podName1, podNameSpace1, podUid1)
		podsInJob[podUid1] = *podDemo1
		convey.Convey("when pod's container is nil, env should be nil", func() {
			convey.So(GetEnvByPod(podsInJob, envName), convey.ShouldEqual, "")
		})
		container := v1.Container{}
		podDemo1.Spec.Containers = append(podDemo1.Spec.Containers, container)
		podsInJob[podUid1] = *podDemo1
		convey.Convey("when container's env is nil, env should be nil", func() {
			convey.So(GetEnvByPod(podsInJob, envName), convey.ShouldEqual, "")
		})
		envVar := v1.EnvVar{Name: envName, Value: envValue}
		container.Env = append(container.Env, envVar)
		podDemo1.Spec.Containers = append(podDemo1.Spec.Containers, container)
		podsInJob[podUid1] = *podDemo1
		convey.Convey("when container's env is exists, env should be exists", func() {
			convey.So(GetEnvByPod(podsInJob, envName), convey.ShouldEqual, envValue)
		})
	})
}

func TestInitRankTableByPod(t *testing.T) {
	convey.Convey("test InitRankTableByPod", t, func() {
		convey.Convey("when replicas is 0, podsInJob is empty, completedPodNum should be 0", func() {
			rankTable, completedPodNum := ConstructRankTableByPod(map[string]v1.Pod{}, 0)
			convey.So(completedPodNum, convey.ShouldEqual, 0)
			convey.So(rankTable.ServerCount, convey.ShouldEqual, "")
		})
		convey.Convey("when replicas is 1, podsInJob is empty, completedPodNum should be 0", func() {
			rankTable, completedPodNum := ConstructRankTableByPod(map[string]v1.Pod{}, 1)
			convey.So(completedPodNum, convey.ShouldEqual, 0)
			convey.So(rankTable.ServerCount, convey.ShouldEqual, "0")
		})
		convey.Convey("when replicas is 1, podsInJob is completed, completedPodNum should be 1", func() {
			podsInJob := make(map[string]v1.Pod)
			podDemo1 := getDemoPod(podName1, podNameSpace1, podUid1)
			podsInJob[podUid1] = *podDemo1
			rankTable, completedPodNum := ConstructRankTableByPod(podsInJob, 1)
			convey.So(completedPodNum, convey.ShouldEqual, 1)
			convey.So(rankTable.ServerCount, convey.ShouldEqual, "1")
		})
		convey.Convey("when replicas is 1, podsInJob nodeRank is illegal, completedPodNum should be 0", func() {
			podsInJob := make(map[string]v1.Pod)
			podDemo1 := getDemoPod(podName1, podNameSpace1, podUid1)
			podDemo1.Annotations[api.PodRankIndexAnno] = errorPodRankIndexKey
			podsInJob[podUid1] = *podDemo1
			rankTable, completedPodNum := ConstructRankTableByPod(podsInJob, 1)
			convey.So(completedPodNum, convey.ShouldEqual, 0)
			convey.So(rankTable.ServerCount, convey.ShouldEqual, "0")
		})
	})
}

func TestGetPodDeviceNumByJobId(t *testing.T) {
	convey.Convey("test GetPodDeviceNumByJobId", t, func() {
		convey.Convey("when podsInJob is nil, deviceNum should be 0", func() {
			convey.So(GetPodDeviceNumByJobId(jobUid1), convey.ShouldEqual, 0)
		})
		convey.Convey("when podsInJob is exists, deviceNum should be 1", func() {
			podDemo1 := getDemoPod(podName1, podNameSpace1, podUid1)
			SavePod(podDemo1)
			defer DeletePod(podDemo1)
			convey.So(GetPodDeviceNumByJobId(jobUid1), convey.ShouldEqual, 1)
		})
	})
}

func TestGetPodByRankIndex(t *testing.T) {
	convey.Convey("test GetPodByRankIndex", t, func() {
		convey.Convey("when podsInJob is nil, pod should be nil", func() {
			convey.So(GetPodByRankIndex(jobUid1, defaultPodRankIndexKey).Name, convey.ShouldEqual, "")
		})
		convey.Convey("when podsInJob is exists,but rankIndex is error, pod should be nil", func() {
			podDemo1 := getDemoPod(podName1, podNameSpace1, podUid1)
			SavePod(podDemo1)
			defer DeletePod(podDemo1)
			convey.So(GetPodByRankIndex(jobUid1, errorPodRankIndexKey).Name, convey.ShouldEqual, "")
		})
		convey.Convey("when podsInJob is exists,but rankIndex is right, pod should be exists", func() {
			podDemo1 := getDemoPod(podName1, podNameSpace1, podUid1)
			SavePod(podDemo1)
			defer DeletePod(podDemo1)
			convey.So(GetPodByRankIndex(jobUid1, defaultPodRankIndexKey).Name, convey.ShouldEqual, podName1)
		})
	})
}

func TestGetModelFramework(t *testing.T) {
	convey.Convey("test GetModelFramework", t, func() {
		convey.Convey("when podsInJob is nil, framework should be nil", func() {
			convey.So(GetModelFramework(map[string]v1.Pod{}), convey.ShouldEqual, "")
		})
		convey.Convey("when podsInJob is exists, framework should be exists", func() {
			podsInJob := make(map[string]v1.Pod)
			podDemo1 := getDemoPod(podName1, podNameSpace1, podUid1)
			podDemo1.Labels[podLabelKey] = ptFramework
			podsInJob[podUid1] = *podDemo1
			convey.So(GetModelFramework(podsInJob), convey.ShouldEqual, ptFramework)
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
				delete(podDemo1.Annotations, api.Pod910DeviceAnno)
				convey.So(DeviceAllocateIsCompleted(*podDemo1), convey.ShouldBeFalse)
			})
	})
}

func TestGetPGByPod(t *testing.T) {
	convey.Convey("test GetPGByPod", t, func() {
		convey.Convey("when podsInJob is nil, jobName should be nil", func() {
			jobName, pgName, namespace := GetPGByPod(jobUid1)
			convey.So(jobName, convey.ShouldEqual, "")
			convey.So(pgName, convey.ShouldEqual, "")
			convey.So(namespace, convey.ShouldEqual, "")
		})
		convey.Convey("when podsInJob is exists, jobName should be exists", func() {
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

func TestConstructServersByJobKey(t *testing.T) {
	convey.Convey("test ConstructServersByJobKey", t, func() {
		fakePod1 := v1.Pod{
			Spec:   v1.PodSpec{NodeName: ""},
			Status: v1.PodStatus{},
		}
		fakePod2 := v1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: podName1, Namespace: podNameSpace1},
			Spec:       v1.PodSpec{NodeName: nodeName1},
			Status:     v1.PodStatus{},
		}
		podMap := map[string]v1.Pod{podName1: fakePod1, podName2: fakePod2}
		mockGetPodByJobId := gomonkey.ApplyFunc(GetPodByJobId, func(string) map[string]v1.Pod {
			return podMap
		}).ApplyFuncReturn(node.GetNodeIpByName, nodeIp1).
			ApplyFuncReturn(node.GetNodeSNByName, nodeSn1)
		defer mockGetPodByJobId.Reset()
		ret := ConstructServersByJobKey(jobName1)
		convey.ShouldResemble(ret, map[string]constant.ServerHccl{
			nodeName1: {
				ServerID:     nodeIp1,
				PodID:        podName1,
				PodNameSpace: podNameSpace1,
				ServerName:   nodeName1,
				ServerSN:     nodeSn1,
			},
		})
	})
}

type podCreateParam struct {
	podName       string
	podNameSpace  string
	podUid        string
	nodeName      string
	podStatus     v1.PodPhase
	podAnnotation map[string]string
}

func TestGetUsedDevicesByNodeName(t *testing.T) {
	convey.Convey("test GetUsedDevicesByNodeName", t, func() {
		param := podCreateParam{
			podName:      podName1,
			podNameSpace: podNameSpace1,
			podUid:       podUid1,
			nodeName:     nodeName1,
			podStatus:    v1.PodFailed,
		}
		convey.Convey("when podsInJob is nil, deviceNum should be 0", func() {
			convey.So(GetUsedDevicesByNodeName(nodeName1).Len(), convey.ShouldEqual, 0)
		})
		convey.Convey("when podsInJob exists but nodename not match, deviceNum should be 0", func() {
			podDemo := createAndSavePod(param)
			defer DeletePod(podDemo)
			convey.So(GetUsedDevicesByNodeName(nodeName2).Len(), convey.ShouldEqual, 0)
		})
		convey.Convey("when podsInJob exists but pod status is failed, deviceNum should be 0", func() {
			podDemo := createAndSavePod(param)
			defer DeletePod(podDemo)
			convey.So(GetUsedDevicesByNodeName(nodeName1).Len(), convey.ShouldEqual, 0)
		})
		param.podStatus = v1.PodRunning
		convey.Convey("when podsInJob exists but pod annotation not match, deviceNum should be 0", func() {
			podDemo := createAndSavePod(param)
			defer DeletePod(podDemo)
			convey.So(GetUsedDevicesByNodeName(nodeName1).Len(), convey.ShouldEqual, 0)
		})
		param.podAnnotation = map[string]string{api.PodAnnotationAscendReal: "Ascend910-0"}
		convey.Convey("when podsInJob exists and pod annotation matches, deviceNum should be 1", func() {
			podDemo := createAndSavePod(param)
			defer DeletePod(podDemo)
			convey.So(GetUsedDevicesByNodeName(nodeName1).Len(), convey.ShouldEqual, 1)
		})
	})
}

func TestHasVolcanoTaskOnNode(t *testing.T) {
	convey.Convey("test HasVolcanoTaskOnNode", t, func() {
		convey.Convey("when no pod is on the node, result should be false", func() {
			convey.So(HasVolcanoTaskOnNode(nodeName1), convey.ShouldBeFalse)
		})
		convey.Convey("when pod has no podgroup annotation, result should be false", func() {
			podDemo := createAndSavePod(podCreateParam{
				podName:       podName1,
				podNameSpace:  podNameSpace1,
				podUid:        podUid1,
				nodeName:      nodeName1,
				podStatus:     v1.PodRunning,
				podAnnotation: map[string]string{},
			})
			defer DeletePod(podDemo)
			convey.So(HasVolcanoTaskOnNode(nodeName1), convey.ShouldBeFalse)
		})
		convey.Convey("when pod is on another node, result should be false", func() {
			podDemo := createAndSavePod(podCreateParam{
				podName:       podName1,
				podNameSpace:  podNameSpace1,
				podUid:        podUid1,
				nodeName:      nodeName1,
				podStatus:     v1.PodRunning,
				podAnnotation: map[string]string{podGroupKey: pgName1},
			})
			defer DeletePod(podDemo)
			convey.So(HasVolcanoTaskOnNode(nodeName2), convey.ShouldBeFalse)
		})
		convey.Convey("when pod has podgroup annotation and is running, result should be true", func() {
			podDemo := createAndSavePod(podCreateParam{
				podName:       podName1,
				podNameSpace:  podNameSpace1,
				podUid:        podUid1,
				nodeName:      nodeName1,
				podStatus:     v1.PodRunning,
				podAnnotation: map[string]string{podGroupKey: pgName1},
			})
			defer DeletePod(podDemo)
			convey.So(HasVolcanoTaskOnNode(nodeName1), convey.ShouldBeTrue)
		})
		convey.Convey("when pod has podgroup annotation but is failed, result should be false", func() {
			podDemo := createAndSavePod(podCreateParam{
				podName:       podName1,
				podNameSpace:  podNameSpace1,
				podUid:        podUid1,
				nodeName:      nodeName1,
				podStatus:     v1.PodFailed,
				podAnnotation: map[string]string{podGroupKey: pgName1},
			})
			defer DeletePod(podDemo)
			convey.So(HasVolcanoTaskOnNode(nodeName1), convey.ShouldBeFalse)
		})
	})
}

func createAndSavePod(param podCreateParam) *v1.Pod {
	pod := getDemoPod(param.podName, param.podNameSpace, param.podUid)
	pod.Spec.NodeName = param.nodeName
	pod.Status.Phase = param.podStatus
	if param.podAnnotation != nil {
		pod.SetAnnotations(param.podAnnotation)
	}
	SavePod(pod)
	return pod
}

const (
	fakeMinAvailableNum = 3
)

func TestGetMinMember(t *testing.T) {
	convey.Convey("test GetMinMember", t, func() {
		convey.Convey("when podsInJob is nil, minMember should be 0", func() {
			minMember := GetMinMember(map[string]v1.Pod{})
			convey.So(minMember, convey.ShouldEqual, 0)
		})
		convey.Convey("when podsInJob exists but no pod has MinAvailableKey annotation, "+
			"minMember should be 0", func() {
			podsInJob := make(map[string]v1.Pod)
			podDemo1 := getDemoPod(podName1, podNameSpace1, podUid1)
			delete(podDemo1.Annotations, api.MinAvailableKey)
			podsInJob[podUid1] = *podDemo1
			minMember := GetMinMember(podsInJob)
			convey.So(minMember, convey.ShouldEqual, 0)
		})
		convey.Convey("when podsInJob exists and one pod has valid MinAvailableKey annotation, "+
			"minMember should be the value", func() {
			podsInJob := make(map[string]v1.Pod)
			podDemo1 := getDemoPod(podName1, podNameSpace1, podUid1)
			podDemo1.Annotations[api.MinAvailableKey] = "3"
			podsInJob[podUid1] = *podDemo1
			minMember := GetMinMember(podsInJob)
			convey.So(minMember, convey.ShouldEqual, fakeMinAvailableNum)
		})
		convey.Convey("when podsInJob exists and one pod has invalid MinAvailableKey annotation, "+
			"minMember should be 0", func() {
			podsInJob := make(map[string]v1.Pod)
			podDemo1 := getDemoPod(podName1, podNameSpace1, podUid1)
			podDemo1.Annotations[api.MinAvailableKey] = "invalid"
			podsInJob[podUid1] = *podDemo1
			minMember := GetMinMember(podsInJob)
			convey.So(minMember, convey.ShouldEqual, 0)
		})
	})
}

func TestJudgePodOnDiffNode(t *testing.T) {
	podMap := make(map[string]v1.Pod)
	podMap["uid1"] = v1.Pod{
		Spec: v1.PodSpec{
			NodeName: "node-1",
		},
	}
	podMap["uid2"] = v1.Pod{
		Spec: v1.PodSpec{
			NodeName: "node-1",
		},
	}
	flg := JudgePodOnDiffNode(podMap)
	if flg {
		t.Errorf("when pod on same node then should return false")
	}

	podMap["uid2"] = v1.Pod{
		Spec: v1.PodSpec{
			NodeName: "node-2",
		},
	}
	flg = JudgePodOnDiffNode(podMap)
	if !flg {
		t.Errorf("when pod on same node then should return true")
	}
}

func TestGetPodByRankIndexInPods(t *testing.T) {
	convey.Convey("test GetPodByRankIndexInPods", t, func() {
		convey.Convey("when podsInJob exists and pod with matching rank index exists", func() {
			podsInJob := make(map[string]v1.Pod)
			podDemo1 := getDemoPod(podName1, podNameSpace1, podUid1)
			podsInJob[podUid1] = *podDemo1
			result := GetPodByRankIndexInPods(defaultPodRankIndexKey, podsInJob)
			convey.So(result.Name, convey.ShouldEqual, podName1)
			convey.So(result.Namespace, convey.ShouldEqual, podNameSpace1)
		})
		convey.Convey("when podsInJob exists but no pod with matching rank index exists", func() {
			podsInJob := make(map[string]v1.Pod)
			podDemo1 := getDemoPod(podName1, podNameSpace1, podUid1)
			podsInJob[podUid1] = *podDemo1
			result := GetPodByRankIndexInPods("1", podsInJob)
			convey.So(result.Name, convey.ShouldEqual, "")
			convey.So(result.Namespace, convey.ShouldEqual, "")
		})
	})
}

func TestGetContainerIdsbyPod(t *testing.T) {
	convey.Convey("test getContainerIdsbyPod", t, func() {
		convey.Convey("when pod is nil, return empty map", func() {
			containerIdsMap := getContainerIdsbyPod(nil)
			convey.So(len(containerIdsMap), convey.ShouldEqual, 0)
		})

		convey.Convey("when pod has no container statuses, return empty map", func() {
			podDemo := getDemoPod(podName1, podNameSpace1, podUid1)
			podDemo.Status.ContainerStatuses = []v1.ContainerStatus{}
			containerIdsMap := getContainerIdsbyPod(podDemo)
			convey.So(len(containerIdsMap), convey.ShouldEqual, 0)
		})

		convey.Convey("when pod has container statuses, return map with container ids", func() {
			podDemo := getDemoPod(podName1, podNameSpace1, podUid1)
			podDemo.Status.ContainerStatuses = []v1.ContainerStatus{
				{
					Name:        "container1",
					ContainerID: "docker://container1-id",
				},
				{
					Name:        "container2",
					ContainerID: "docker://container2-id",
				},
			}
			containerIdsMap := getContainerIdsbyPod(podDemo)
			convey.So(len(containerIdsMap), convey.ShouldEqual, len(podDemo.Status.ContainerStatuses))
			convey.So(containerIdsMap["container1"], convey.ShouldEqual, "docker://container1-id")
			convey.So(containerIdsMap["container2"], convey.ShouldEqual, "docker://container2-id")
		})
	})
}

func TestGetPodUsedDev(t *testing.T) {
	convey.Convey("test GetPodUsedDev", t, func() {
		convey.Convey("when pod or annotation is nil, used devs is empty", func() {
			nilPod := v1.Pod{}
			usedDev := GetPodUsedDev(nilPod)
			convey.So(len(usedDev), convey.ShouldEqual, 0)
		})
		convey.Convey("get pod used dev failed, annotation does not exist", func() {
			podDemo := getDemoPod(podName1, podNameSpace1, podUid1)
			podDemo.Annotations = map[string]string{podGroupKey: pgName1}
			usedDev := GetPodUsedDev(*podDemo)
			convey.So(len(usedDev), convey.ShouldEqual, 0)
		})
		convey.Convey("get pod used dev successfully", func() {
			podDemo := getDemoPod(podName1, podNameSpace1, podUid1)
			usedDev := GetPodUsedDev(*podDemo)
			convey.So(len(usedDev), convey.ShouldEqual, 1)
			convey.So(usedDev, convey.ShouldResemble, []string{"0"})
		})
		convey.Convey("get pod used dev successfully, pod has 3 devices", func() {
			const podDeviceKey = `{"server_id":"127.0.0.1","devices":[{"device_id":"0"},{"device_id":"1"},{"device_id":"2"}]}`
			podDemo := getDemoPod(podName1, podNameSpace1, podUid1)
			podDemo.Annotations[api.Pod910DeviceAnno] = podDeviceKey
			usedDev := GetPodUsedDev(*podDemo)
			convey.So(len(usedDev), convey.ShouldEqual, len3)
			convey.So(usedDev, convey.ShouldResemble, []string{"0", "1", "2"})
		})
		convey.Convey("get pod used dev failed, parse annotation failed", func() {
			const podDeviceKey = `error device key`
			podDemo := getDemoPod(podName1, podNameSpace1, podUid1)
			podDemo.Annotations[api.Pod910DeviceAnno] = podDeviceKey
			usedDev := GetPodUsedDev(*podDemo)
			convey.So(len(usedDev), convey.ShouldEqual, 0)
		})
	})
}

func TestCreateDevNameJobMap(t *testing.T) {
	convey.Convey("test CreateDevNameJobMap and GetJobIdByDev", t, func() {
		convey.Convey("test CreateDevNameJobMap and GetJobIdByDev success", testCreateDevNameJobMapAndGet)
	})
}

func testCreateDevNameJobMapAndGet() {
	podManager.nodePodMap = make(map[string]map[string]v1.Pod)
	// node1, job1, dev0
	podDemo1 := getDemoPod(podName1, podNameSpace1, podUid1)
	// node1, job2, dev2
	podDemo2 := getDemoPod(podName2, podNameSpace1, podUid2)
	annotation := map[string]string{api.Pod910DeviceAnno: podDeviceKey2}
	podDemo2.SetAnnotations(annotation)
	podDemo2.OwnerReferences[0].UID = jobUid2
	podDemo2.OwnerReferences[0].Name = jobName2
	// node2, job1, dev0
	podDemo3 := getDemoPod(podName1, podNameSpace1, podUid1)
	podDemo3.Spec.NodeName = nodeName2
	// node2, job3, dev5
	podDemo4 := getDemoPod(podName2, podNameSpace1, podUid2)
	podDemo4.Spec.NodeName = nodeName2
	annotation = map[string]string{api.Pod910DeviceAnno: podDeviceKey5}
	podDemo4.SetAnnotations(annotation)
	podDemo4.OwnerReferences[0].UID = jobUid3
	podDemo4.OwnerReferences[0].Name = jobName3
	SavePod(podDemo1)
	SavePod(podDemo2)
	SavePod(podDemo3)
	SavePod(podDemo4)
	defer func() {
		DeletePod(podDemo1)
		DeletePod(podDemo2)
		DeletePod(podDemo3)
		DeletePod(podDemo4)
	}()

	devJobMap := CreateDevNameJobMap(nodeName1, api.Ascend910)
	jobName := GetJobIdByDev(devJobMap, fmt.Sprintf("%s-%s", api.Ascend910, dev0))
	convey.So(jobName, convey.ShouldEqual, jobUid1)
	jobName = GetJobIdByDev(devJobMap, fmt.Sprintf("%s-%s", api.Ascend910, dev2))
	convey.So(jobName, convey.ShouldEqual, jobUid2)
	// no job on dev5
	jobName = GetJobIdByDev(devJobMap, fmt.Sprintf("%s-%s", api.Ascend910, dev5))
	convey.So(jobName, convey.ShouldEqual, "")

	devJobMap = CreateDevNameJobMap(nodeName2, api.Ascend910)
	jobName = GetJobIdByDev(devJobMap, fmt.Sprintf("%s-%s", api.Ascend910, dev0))
	convey.So(jobName, convey.ShouldEqual, jobUid1)
	jobName = GetJobIdByDev(devJobMap, fmt.Sprintf("%s-%s", api.Ascend910, dev5))
	convey.So(jobName, convey.ShouldEqual, jobUid3)

	// no job on node3
	devJobMap = CreateDevNameJobMap(nodeName3, api.Ascend910)
	jobName = GetJobIdByDev(devJobMap, fmt.Sprintf("%s-%s", api.Ascend910, dev0))
	convey.So(jobName, convey.ShouldEqual, "")
}

func TestCheckPodIsSoftShareDev(t *testing.T) {
	convey.Convey("test CheckPodIsNotSoftShareDev", t, func() {
		convey.Convey("when podId does not exist in storage, should return true", func() {
			convey.So(CheckPodIsNotSoftShareDev("not-exist-pod-id"), convey.ShouldBeTrue)
		})
		convey.Convey("when pod has no annotations, should return true", func() {
			podDemo := getDemoPod(podName1, podNameSpace1, podUid1)
			podDemo.SetAnnotations(nil)
			SavePod(podDemo)
			defer DeletePod(podDemo)
			convey.So(CheckPodIsNotSoftShareDev(podUid1), convey.ShouldBeTrue)
		})
		convey.Convey("when pod annotation does not contain SchedulePolicyAnnoKey, should return true", func() {
			podDemo := getDemoPod(podName1, podNameSpace1, podUid1)
			podDemo.SetAnnotations(map[string]string{podGroupKey: pgName1})
			SavePod(podDemo)
			defer DeletePod(podDemo)
			convey.So(CheckPodIsNotSoftShareDev(podUid1), convey.ShouldBeTrue)
		})
		convey.Convey("when pod SchedulePolicyAnnoKey is not Chip1SoftShareDev, should return true", func() {
			podDemo := getDemoPod(podName1, podNameSpace1, podUid1)
			podDemo.SetAnnotations(map[string]string{
				api.SchedulePolicyAnnoKey: "other-policy",
			})
			SavePod(podDemo)
			defer DeletePod(podDemo)
			convey.So(CheckPodIsNotSoftShareDev(podUid1), convey.ShouldBeTrue)
		})
		convey.Convey("when pod SchedulePolicyAnnoKey is Chip1SoftShareDev, should return false", func() {
			podDemo := getDemoPod(podName1, podNameSpace1, podUid1)
			podDemo.SetAnnotations(map[string]string{
				api.SchedulePolicyAnnoKey: api.Chip1SoftShareDev,
			})
			SavePod(podDemo)
			defer DeletePod(podDemo)
			convey.So(CheckPodIsNotSoftShareDev(podUid1), convey.ShouldBeFalse)
		})
	})
}

func TestGetServerInfoDoesNotLeakLevelList(t *testing.T) {
	convey.Convey("test getServerInfo does not leak levelList", t, func() {
		podDev := constant.PodDevice{
			PodName: "test-pod",
			Devices: []constant.Device{fakeDeviceWithLevelList("0")},
		}
		podDemo := getDemoPod(podName1, podNameSpace1, podUid1)
		server := getServerInfo(podDev, *podDemo, map[string]string{}, 0)
		convey.So(len(server.DeviceList), convey.ShouldEqual, 1)
		convey.So(server.DeviceList[0].LevelList, convey.ShouldBeNil)
		b, err := json.Marshal(server)
		convey.So(err, convey.ShouldBeNil)
		convey.So(strings.Contains(string(b), `"levelList"`), convey.ShouldBeFalse)
	})
}
