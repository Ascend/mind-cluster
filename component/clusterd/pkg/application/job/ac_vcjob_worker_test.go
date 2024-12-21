// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package job a series of job function
package job

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"

	"clusterd/pkg/common/util"
)

const (
	errRankNum        = 10001
	noErrRankNum      = 9999
	mockReplicasTotal = 2

	mockPodName1      = "test-1"
	mockPodUID1       = "pfw23-01"
	mockPodUID2       = "pfw23-02"
	mockPodAnnotation = `{"pod_name":"pytorch-simple-ccae-test-two-001-worker-0","server_id":"9.88.82.19",
					"super_pod_id":-2,"devices":[{"device_id":"0","device_ip":"192.168.30.30"},
					{"device_id":"1","device_ip":"192.168.30.31"},{"device_id":"2","device_ip":"192.168.30.32"},
                    {"device_id":"3","device_ip":"192.168.30.33"},{"device_id":"4","device_ip":"192.168.30.34"},
                    {"device_id":"5","device_ip":"192.168.30.35"},{"device_id":"6","device_ip":"192.168.30.36"},
                    {"device_id":"7","device_ip":"192.168.30.37"}]}`
	mockPodRankIndex       = "0"
	mockHccLJsonSliceLen   = 2
	mockPositiveModifyStat = 1
	mockNegativeModifyStat = -1
	mockNodeName           = "test-node"
)

// TestNewJobWorker test NewJobWorker
func TestNewJobWorker(t *testing.T) {
	convey.Convey("test NewJobWorker", t, func() {
		agent := mockAgentEmpty()
		job := mockJobInfo()
		var ranktable RankTabler
		jobWorker := NewJobWorker(agent, job, ranktable, mockReplicasTotal)
		convey.So(jobWorker, convey.ShouldNotBeNil)
		convey.So(jobWorker.jobReplicasTotal, convey.ShouldEqual, mockReplicasTotal)
		convey.So(jobWorker.cachedPodNum, convey.ShouldEqual, 0)
		convey.So(jobWorker.CMName, convey.ShouldEqual, fmt.Sprintf("%s-%s", ConfigmapPrefix, mockJobName))
	})
}

// TestPGRunning test NewJobWorker
func TestPGRunning(t *testing.T) {
	convey.Convey("test PGRunning", t, func() {
		job := mockWorkerInfo()
		convey.Convey("job schedule success", func() {
			mockJobScheduleSuccess := gomonkey.ApplyPrivateMethod(reflect.TypeOf(new(WorkerInfo)), "constructionFinished",
				func(_ *WorkerInfo) bool {
					return true
				})
			defer mockJobScheduleSuccess.Reset()
			convey.So(job.PGRunning(), convey.ShouldBeTrue)
		})

		convey.Convey("job schedule fail", func() {
			mockJobScheduleFail := gomonkey.ApplyPrivateMethod(reflect.TypeOf(new(WorkerInfo)), "constructionFinished",
				func(_ *WorkerInfo) bool {
					return false
				})
			defer mockJobScheduleFail.Reset()
			convey.So(job.PGRunning(), convey.ShouldBeFalse)
		})
	})
}

// TestGetBaseInfo
func TestGetBaseInfo(t *testing.T) {
	convey.Convey("test GetBaseInfo", t, func() {
		job := mockWorker()
		info := job.GetBaseInfo()
		convey.So(info.Namespace, convey.ShouldEqual, mockNamespace)
		convey.So(info.Name, convey.ShouldEqual, mockJobName)
		convey.So(info.Key, convey.ShouldEqual, "")
		convey.So(info.Version, convey.ShouldEqual, 0)
		convey.So(info.Uid, convey.ShouldEqual, mockJobUID)
	})
}

func TestGetDeviceNumPerNode(t *testing.T) {
	convey.Convey("test GetDeviceNumPerNode", t, func() {
		job := mockWorker()
		convey.Convey("case server list length is 0", func() {
			job.CMData = mockRankTableInit()
			convey.So(job.GetDeviceNumPerNode(), convey.ShouldEqual, -1)
		})
		convey.Convey("case server list length > 0", func() {
			job.CMData = mockRankTableWithLength1()
			convey.So(job.GetDeviceNumPerNode(), convey.ShouldEqual, 1)
		})
	})
}

// TestDoPreCheck test doPreCheck
func TestDoPreCheck(t *testing.T) {
	convey.Convey("test doPreCheck", t, func() {
		worker := mockWorker()
		podInfo := mockPodIdentifier()
		pod := mockPod()
		convey.Convey("pod is not exixt, return error", func() {
			worker.CreationTimestamp = metav1.Now()
			pod.CreationTimestamp = metav1.NewTime(time.Date(1993, 02, 28, 9, 04, 39, 213, time.Local))
			err := worker.doPreCheck(pod, podInfo)
			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("pod does not init, return error", func() {
			err := worker.doPreCheck(pod, podInfo)
			convey.So(err, convey.ShouldNotBeNil)
		})
		mockIsReferenceJobSameWithWorker := gomonkey.ApplyFunc(isReferenceJobSameWithWorker,
			func(_ *v1.Pod, _ string, _ string) bool {
				return true
			})
		defer mockIsReferenceJobSameWithWorker.Reset()

		convey.Convey("with configmap init, return nil", func() {
			err := worker.doPreCheck(pod, podInfo)
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("if current pod use chip, the device info may not be ready, return error", func() {
			mockIsPodAnnotationsReady := gomonkey.ApplyFunc(isPodAnnotationsReady, func(_ *v1.Pod, _ string) bool {
				return false
			})
			mockContainerUsedChip := gomonkey.ApplyFunc(containerUsedChip, func(_ *v1.Pod) bool {
				return true
			})
			defer mockIsPodAnnotationsReady.Reset()
			defer mockContainerUsedChip.Reset()
			err := worker.doPreCheck(pod, podInfo)
			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("with configmap complete, return error", func() {
			worker.CMData.SetStatus(ConfigmapCompleted)
			err := worker.doPreCheck(pod, podInfo)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

// TestHandler test handler
func TestHandler(t *testing.T) {
	convey.Convey("test handler", t, func() {
		workerInfo := mockWorkerInfo()
		pod := mockPod()
		podInfo := mockPodIdentifier()
		mockHandlePodAddUpdateEvent := gomonkey.ApplyPrivateMethod(reflect.TypeOf(new(WorkerInfo)),
			"handlePodAddUpdateEvent", func(_ *WorkerInfo, _ *podIdentifier, _ *v1.Pod) error {
				return fmt.Errorf("the key of " + PodDeviceKey + " does not exist ")
			})

		defer mockHandlePodAddUpdateEvent.Reset()
		convey.Convey("configmap update failed, return error", func() {
			mockEndConstruction := gomonkey.ApplyPrivateMethod(reflect.TypeOf(workerInfo), "endConstruction",
				func(_ *WorkerInfo, _ *podIdentifier) error {
					return fmt.Errorf("update configmap failed")
				})
			defer mockEndConstruction.Reset()
			err := workerInfo.handler(pod, podInfo)
			convey.So(err, convey.ShouldNotBeNil)
		})
		mockEndConstruction := gomonkey.ApplyPrivateMethod(reflect.TypeOf(workerInfo), "endConstruction",
			func(_ *WorkerInfo, _ *podIdentifier) error {
				return nil
			})
		defer mockEndConstruction.Reset()
		convey.Convey("no chip case, return nil", func() {
			err := workerInfo.handler(pod, podInfo)
			convey.So(err, convey.ShouldBeNil)
		})
		workerInfo.jobReplicasTotal = 1
		convey.Convey("dry run, return nil", func() {
			workerInfo.dryRun = true
			err := workerInfo.handler(pod, podInfo)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("pod update event failed, return error", func() {
			workerInfo.dryRun = false
			err := workerInfo.handler(pod, podInfo)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("undefined event, return nil", func() {
			podInfo.eventType = EventDelete
			err := workerInfo.handler(pod, podInfo)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

// TestConstructionFinished test constructionFinished
func TestConstructionFinished(t *testing.T) {
	b := mockWorkerInfo()
	convey.Convey("test constructionFinished true", t, func() {
		b.cachedPodNum = 1
		b.jobReplicasTotal = 1
		boolCheck := b.constructionFinished()
		convey.So(boolCheck, convey.ShouldBeTrue)
	})

	convey.Convey("test constructionFinished false", t, func() {
		b.cachedPodNum = 1
		b.jobReplicasTotal = 2
		boolCheck := b.constructionFinished()
		convey.So(boolCheck, convey.ShouldBeFalse)
	})
}

// TestHandlePodWithoutChip test handlePodWithoutChip
func TestHandlePodWithoutChip(t *testing.T) {
	convey.Convey("test handlePodWithoutChip", t, func() {
		b := mockWorkerInfo()
		podInfo := mockPodIdentifier()
		pod := mockPod()
		convey.Convey("no chip pod, cached pod plus one", func() {
			b.cachedPodNum = 0
			b.jobReplicasTotal = 3
			b.handlePodWithoutChip(podInfo, pod)
			convey.So(b.cachedPodNum, convey.ShouldEqual, 1)
		})
	})
}

func conveyHandlePodAddUpdateEvent(b *WorkerInfo, podInfo *podIdentifier, pod *v1.Pod) {
	convey.Convey("deviceInfo unmarshal failed", func() {
		patch := gomonkey.ApplyFunc(json.Unmarshal, func(data []byte, v any) error {
			return fmt.Errorf("json unmarshal failed")
		})
		defer patch.Reset()
		err := b.handlePodAddUpdateEvent(podInfo, pod)
		convey.So(err, convey.ShouldNotBeNil)
	})
	convey.Convey("string parse integer failed", func() {
		patch := gomonkey.ApplyFunc(strconv.ParseInt, func(s string, base int, bitSize int) (i int64, err error) {
			return 0, fmt.Errorf("string parse integer failed")
		})
		defer patch.Reset()
		err := b.handlePodAddUpdateEvent(podInfo, pod)
		convey.So(err, convey.ShouldNotBeNil)
	})
	convey.Convey("validateRank execute failed", func() {
		patch := gomonkey.ApplyFunc(validateRank, func(rank int64) error {
			return fmt.Errorf("validateRank execute failed")
		})
		defer patch.Reset()
		err := b.handlePodAddUpdateEvent(podInfo, pod)
		convey.So(err, convey.ShouldNotBeNil)
	})
	convey.Convey("set cachePodInfo failed", func() {
		patch := gomonkey.ApplyMethod(reflect.TypeOf(new(RankTable)), "CachePodInfo",
			func(_ *RankTable, pod *v1.Pod, instance Instance, rankIndex *int) error {
				return fmt.Errorf("deviceInfo failed the validation")
			})
		defer patch.Reset()
		err := b.handlePodAddUpdateEvent(podInfo, pod)
		convey.So(err, convey.ShouldNotBeNil)
	})
	convey.Convey("updateWithFinish failed", func() {
		patch := gomonkey.ApplyPrivateMethod(reflect.TypeOf(new(WorkerInfo)), "updateWithFinish",
			func(_ *WorkerInfo, podInfo *podIdentifier) error {
				return fmt.Errorf("updateWithFinish failed")
			})
		defer patch.Reset()
		err := b.handlePodAddUpdateEvent(podInfo, pod)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

// TestHandlePodAddUpdateEvent test handlePodAddUpdateEvent
func TestHandlePodAddUpdateEvent(t *testing.T) {
	convey.Convey("test handlePodAddUpdateEvent", t, func() {
		b, podInfo, pod := mockWorkerInfo(), mockPodIdentifier(), mockPod()
		convey.Convey("container does not use chip", func() {
			patch := gomonkey.ApplyFunc(containerUsedChip, func(_ *v1.Pod) bool {
				return false
			}).ApplyPrivateMethod(reflect.TypeOf(new(WorkerInfo)), "handlePodWithoutChip",
				func(_ *WorkerInfo, _ *podIdentifier, _ *v1.Pod) { return })
			defer patch.Reset()
			err := b.handlePodAddUpdateEvent(podInfo, pod)
			convey.So(err, convey.ShouldBeNil)
		})
		mockContainerUsedChip := gomonkey.ApplyFunc(containerUsedChip, func(_ *v1.Pod) bool {
			return true
		})
		defer mockContainerUsedChip.Reset()
		convey.Convey("pod annotation does not exist", func() {
			err := b.handlePodAddUpdateEvent(podInfo, pod)
			convey.So(err, convey.ShouldNotBeNil)
		})
		pod.Spec.NodeName = mockNodeName
		pod.Annotations[PodDeviceKey] = mockPodAnnotation
		pod.Annotations[PodRankIndexKey] = mockPodRankIndex
		pod.Labels = map[string]string{PodLabelKey: mockJobLabelKey}
		conveyHandlePodAddUpdateEvent(b, podInfo, pod)
		convey.Convey("update successfully finished", func() {
			b.jobReplicasTotal = 2
			b.cachedPodNum = 0
			err := b.handlePodAddUpdateEvent(podInfo, pod)
			convey.So(ModelFramework, convey.ShouldEqual, mockJobLabelKey)
			convey.So(err, convey.ShouldBeNil)
			convey.So(b.cachedPodNum, convey.ShouldEqual, 1)
		})
	})
}

// TestValidateRank test validateRank
func TestValidateRank(t *testing.T) {
	convey.Convey("test validateRank err case", t, func() {
		err := validateRank(errRankNum)
		convey.So(err, convey.ShouldBeError)
	})

	convey.Convey("test validateRank no err", t, func() {
		err := validateRank(noErrRankNum)
		convey.So(err, convey.ShouldBeNil)
	})
}

// TestDeletePodUIDFromList test deletePodUIDFromList
func TestDeletePodUIDFromList(t *testing.T) {
	convey.Convey("test deletePodUIDFromList", t, func() {
		b := mockWorkerInfo()
		podInfo := mockPodIdentifier()
		convey.Convey("delete pod uid", func() {
			b.podSchedulerCache = []string{mockPodUID1, mockPodUID2}
			b.deletePodUIDFromList(podInfo)
			convey.So(len(b.podSchedulerCache), convey.ShouldEqual, 1)
		})
	})
}

// TestHandlePodDelEvent test handlePodDelEvent
func TestHandlePodDelEvent(t *testing.T) {
	convey.Convey("test handlePodDelEvent", t, func() {
		b := mockWorkerInfo()
		podInfo := mockPodIdentifier()
		convey.Convey("should return nil", func() {
			b.podSchedulerCache = []string{mockPodUID1, mockPodUID2}
			mockHandlePodWithoutChip := gomonkey.ApplyMethod(reflect.TypeOf(new(WorkerInfo)),
				"UpdateConfigMap", func(_ *WorkerInfo, _ *podIdentifier, _ string) error {
					return nil
				})
			defer mockHandlePodWithoutChip.Reset()
			b.cachedPodNum = 2
			err := b.handlePodDelEvent(podInfo)
			convey.So(err, convey.ShouldBeNil)
			convey.So(b.cachedPodNum, convey.ShouldEqual, 1)
		})
		convey.Convey("check configMap update failed, error should not be nil", func() {
			mockHandlePodWithoutChip := gomonkey.ApplyMethod(reflect.TypeOf(new(WorkerInfo)),
				"UpdateConfigMap", func(_ *WorkerInfo, _ *podIdentifier, _ string) error {
					return fmt.Errorf("failed to update configMap")
				})
			defer mockHandlePodWithoutChip.Reset()
			err := b.handlePodDelEvent(podInfo)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

// TestModifyStat test modifyStat
func TestModifyStat(t *testing.T) {
	b := mockWorkerInfo()
	convey.Convey("test modifyStat minus", t, func() {
		b.modifyStat(mockNegativeModifyStat)
		convey.So(b.cachedPodNum, convey.ShouldEqual, 0)
	})
	convey.Convey("test modifyStat plus", t, func() {
		b.modifyStat(mockPositiveModifyStat)
		convey.So(b.cachedPodNum, convey.ShouldEqual, 1)
	})
}

// TestCloseStat test CloseStat
func TestCloseStat(t *testing.T) {
	b := mockWorkerInfo()
	convey.Convey("test CloseStat", t, func() {
		b.statSwitch = make(chan struct{})
		b.CloseStat()
		convey.So(b.statStopped, convey.ShouldBeTrue)
	})
}

// TestHandleJobStatus test HandleJobStatus
func TestHandleJobStatus(t *testing.T) {
	b := mockWorkerInfo()
	convey.Convey("test HandleJobStatus", t, func() {
		b.succeedPodNum = 0
		b.cachedPodNum = 1
		convey.Convey("job succeed case", func() {
			jobStatus := b.HandleJobStatus(PhaseJobSucceed)
			convey.So(jobStatus, convey.ShouldEqual, StatusJobSucceed)
		})
		convey.Convey("job still running case", func() {
			b.cachedPodNum = 2
			jobStatus := b.HandleJobStatus(PhaseJobSucceed)
			convey.So(jobStatus, convey.ShouldEqual, PhaseJobRunning)
		})
		convey.Convey("job failed case", func() {
			jobStatus := b.HandleJobStatus(PhaseJobRunning)
			convey.So(jobStatus, convey.ShouldEqual, StatusJobFail)
		})
	})
}

// TestIsReferenceJobSameWithWorker test isReferenceJobSameWithWorker
func TestIsReferenceJobSameWithWorker(t *testing.T) {
	convey.Convey("test isReferenceJobSameWithWorker", t, func() {
		pod := mockPod()
		convey.Convey("job is the same with worker", func() {
			isReferenced := isReferenceJobSameWithWorker(pod, mockJobName, mockPodUID1)
			convey.So(isReferenced, convey.ShouldBeTrue)
		})
		convey.Convey("job is not the same with worker", func() {
			isReferenced := isReferenceJobSameWithWorker(pod, mockJobName, mockPodUID2)
			convey.So(isReferenced, convey.ShouldBeFalse)
		})
	})
}

// TestIsPodAnnotationsReady test isPodAnnotationsReady
func TestIsPodAnnotationsReady(t *testing.T) {
	convey.Convey("test isPodAnnotationsReady", t, func() {
		pod := mockPod()
		convey.Convey("pod annotation is not ready", func() {
			ifPodReady := isPodAnnotationsReady(pod, "")
			convey.So(ifPodReady, convey.ShouldBeFalse)
		})
		convey.Convey("pod annotation is ready", func() {
			pod.Annotations[PodDeviceKey] = ""
			ifPodReady := isPodAnnotationsReady(pod, "")
			convey.So(ifPodReady, convey.ShouldBeTrue)
		})
	})
}

func mockWorkerInfo() *WorkerInfo {
	return &WorkerInfo{
		clientSet:         nil,
		CmMutex:           sync.Mutex{},
		statMu:            sync.Mutex{},
		dryRun:            false,
		statSwitch:        nil,
		podIndexer:        nil,
		CMName:            "",
		CMData:            mockRankTableInit(),
		statStopped:       false,
		rankIndex:         0,
		cachedPodNum:      0,
		jobReplicasTotal:  0,
		podSchedulerCache: make([]string, 0),
	}
}

func mockJobInfo() Info {
	return Info{
		Namespace:         mockNamespace,
		Name:              mockJobName,
		Key:               "",
		Version:           0,
		Uid:               mockJobUID,
		CreationTimestamp: metav1.Time{},
	}
}

func mockWorker() *Worker {
	return &Worker{
		WorkerInfo: *mockWorkerInfo(),
		Info:       mockJobInfo(),
	}
}

func mockPod() *v1.Pod {
	return &v1.Pod{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:              "",
			Namespace:         "",
			UID:               "",
			ResourceVersion:   "",
			OwnerReferences:   mockOwnerReference(),
			CreationTimestamp: metav1.Time{},
			DeletionTimestamp: nil,
			Labels:            make(map[string]string),
			Annotations:       make(map[string]string),
		},
		Spec: v1.PodSpec{
			Containers: make([]v1.Container, 0),
		},
		Status: v1.PodStatus{
			Phase:     "",
			HostIP:    "",
			PodIP:     "",
			PodIPs:    nil,
			StartTime: nil,
		},
	}
}

func mockOwnerReference() []metav1.OwnerReference {
	return []metav1.OwnerReference{
		{
			APIVersion: "",
			Kind:       "",
			Name:       mockJobName,
			UID:        mockPodUID1,
		},
	}
}

func mockPodIdentifier() *podIdentifier {
	return &podIdentifier{
		namespace: mockNamespace,
		name:      mockPodName1,
		jobName:   mockJobName,
		eventType: EventAdd,
		UID:       mockPodUID1,
	}
}

func mockFakePodByAnnotation(anno map[string]string) *v1.Pod {
	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: anno,
		},
	}
}

// TestWorkerInfoSetSharedTorIp test case for WorkerInfoSetSharedTorIp
func TestWorkerInfoSetSharedTorIp(t *testing.T) {
	convey.Convey("Test setSharedTorIp", t, func() {
		b := &WorkerInfo{}

		// Case 1: Pod not have Annotations
		pod1 := &v1.Pod{}
		b.setSharedTorIp(pod1)
		convey.So(b.SharedTorIp, convey.ShouldHaveLength, 0)

		// Case 2: Pod has Annotations, not have  torTag or torTag not equal sharedTor
		pod2 := mockFakePodByAnnotation(map[string]string{"some.other.tag": "value"})
		b.setSharedTorIp(pod2)
		convey.So(b.SharedTorIp, convey.ShouldHaveLength, 0)

		// Case 3: Pod has Annotations，torTag equal sharedTor, not has torIpTag
		pod3 := mockFakePodByAnnotation(map[string]string{torTag: sharedTor})
		b.setSharedTorIp(pod3)
		convey.So(b.SharedTorIp, convey.ShouldHaveLength, 0)

		// Case 4: Pod has Annotations, torTag equal sharedTor, has torIpTag
		pod4 := mockFakePodByAnnotation(map[string]string{torTag: sharedTor, torIpTag: "192.168.1.1"})
		b.setSharedTorIp(pod4)
		convey.So(b.SharedTorIp, convey.ShouldContain, "192.168.1.1")
		convey.So(b.SharedTorIp, convey.ShouldHaveLength, 1)
	})
}

// TestInitMasterAddrByJobType test initMasterAddrByJobType
func TestInitMasterAddrByJobType(t *testing.T) {
	convey.Convey("testInitMasterAddrByJobType", t, func() {
		tmpRT := &RankTable{ServerList: []*ServerHccl{{ServerID: "testId"}}}
		vcjob := &WorkerInfo{JobType: vcJobKind, CMData: tmpRT}
		masterIp := vcjob.initMasterAddrByJobType(mockPodIdentifier())
		convey.So(masterIp, convey.ShouldEqual, "testId")
		acjob := &WorkerInfo{JobType: "AscendJob", CMData: tmpRT}
		patch := gomonkey.ApplyFunc(util.GetServiceIpWithRetry,
			func(k kubernetes.Interface, nameSpace, name string) string {
				return "testIp"
			})
		defer patch.Reset()
		convey.So(acjob.initMasterAddrByJobType(mockPodIdentifier()), convey.ShouldEqual, "testIp")
	})

}

func conveyUpdateConfigMap(workerInfo *WorkerInfo, cm *v1.ConfigMap, podInfo *podIdentifier,
	jobStatus string, assert convey.Assertion) {
	createConfigMaps(workerInfo.clientSet, cm)
	err := workerInfo.UpdateConfigMap(podInfo, jobStatus)
	convey.So(err, assert)
}

func TestUpdateConfigMap(t *testing.T) {
	convey.Convey("Test UpdateConfigMap", t, func() {
		workerInfo, podInfo, cm := mockWorkerInfo(), mockPodIdentifier(), mockConfigMap()
		workerInfo.CMName = cmName
		workerInfo.clientSet = fake.NewSimpleClientset()
		convey.Convey("get configmap failed", func() {
			err := workerInfo.UpdateConfigMap(podInfo, StatusJobRunning)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("old cm ranktable not exists", func() {
			conveyUpdateConfigMap(workerInfo, cm, podInfo, StatusJobRunning, convey.ShouldNotBeNil)
		})
		cm.Data[ConfigmapKey] = ""
		convey.Convey("JobStatus is failed", func() {
			cm.Data[JobStatus] = StatusJobFail
			conveyUpdateConfigMap(workerInfo, cm, podInfo, StatusJobFail, convey.ShouldBeNil)
		})
		convey.Convey("label910 is not exist", func() {
			conveyUpdateConfigMap(workerInfo, cm, podInfo, StatusJobRunning, convey.ShouldNotBeNil)
		})
		(*cm).Labels = map[string]string{Key910: Val910}
		convey.Convey("CMData marshal failed", func() {
			mockMarshal := gomonkey.ApplyFunc(json.Marshal,
				func(_ any) ([]byte, error) {
					return nil, fmt.Errorf("failed to marshal CMData")
				})
			defer mockMarshal.Reset()
			conveyUpdateConfigMap(workerInfo, cm, podInfo, StatusJobRunning, convey.ShouldNotBeNil)

		})
		convey.Convey("update job hccl.json failed", func() {
			mockUpdateJobHccLJson := gomonkey.ApplyPrivateMethod(reflect.TypeOf(workerInfo),
				"updateJobHccLJson", func(_ v1.ConfigMap) error {
					return fmt.Errorf("failed to update ConfigMap for Job")
				})
			defer mockUpdateJobHccLJson.Reset()
			workerInfo.CMData = mockRankTableInit()
			conveyUpdateConfigMap(workerInfo, cm, podInfo, StatusJobRunning, convey.ShouldNotBeNil)
		})
		convey.Convey("JobStatus is not complete", func() {
			mockUpdateJobHccLJson := gomonkey.ApplyPrivateMethod(reflect.TypeOf(workerInfo),
				"updateJobHccLJson", func(b *WorkerInfo, cm v1.ConfigMap) error {
					return nil
				})
			defer mockUpdateJobHccLJson.Reset()
			cm.Data[JobStatus] = StatusJobPending
			conveyUpdateConfigMap(workerInfo, cm, podInfo, StatusJobSucceed, convey.ShouldBeNil)
		})
	})
}

func TestUpdateCMWhenJobEnd(t *testing.T) {
	convey.Convey("Test UpdateCMWhenJobEnd", t, func() {
		workerInfo := mockWorkerInfo()
		workerInfo.CMName = cmName
		workerInfo.clientSet = fake.NewSimpleClientset()
		podKeyInfo := mockPodIdentifier()
		cm := &v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{Name: cmName, Namespace: mockNamespace},
			Data:       map[string]string{},
		}
		pod := mockPod()
		convey.Convey("get configmap failed, error should not be nil", func() {
			err := workerInfo.UpdateCMWhenJobEnd(podKeyInfo)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("when job status is failed, error should be nil", func() {
			cm.Data = map[string]string{JobStatus: StatusJobFail}
			createConfigMaps(workerInfo.clientSet, cm)
			err := workerInfo.UpdateCMWhenJobEnd(podKeyInfo)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("get pod failed, error should not be nil", func() {
			conveyUpdateCMWhenJobEnd(workerInfo, cm, pod, podKeyInfo, convey.ShouldNotBeNil)
		})
		pod.Name = mockPodName1
		pod.Namespace = mockNamespace
		convey.Convey("pod status is Running or Pending, error should be nil", func() {
			pod.Status.Phase = PhaseJobRunning
			conveyUpdateCMWhenJobEnd(workerInfo, cm, pod, podKeyInfo, convey.ShouldBeNil)
		})

		convey.Convey("current job status is Running, error should be nil", func() {
			mockHandleJobStatus := gomonkey.ApplyMethod(reflect.TypeOf(workerInfo),
				"HandleJobStatus", func(_ *WorkerInfo, _ string) string {
					return PhaseJobRunning
				})
			defer mockHandleJobStatus.Reset()
			conveyUpdateCMWhenJobEnd(workerInfo, cm, pod, podKeyInfo, convey.ShouldBeNil)
		})
		convey.Convey("curJobStatus case StatusJobFail", func() {
			conveyUpdateCMWhenJobEndWithStatus(workerInfo, cm, pod, podKeyInfo, StatusJobFail)
		})
		convey.Convey("curJobStatus case default", func() {
			conveyUpdateCMWhenJobEndWithStatus(workerInfo, cm, pod, podKeyInfo, StatusJobSucceed)
		})
	})
}

func conveyUpdateCMWhenJobEndWithStatus(workerInfo *WorkerInfo, cm *v1.ConfigMap,
	pod *v1.Pod, podKeyInfo *podIdentifier, jobStatus string) {
	patch := gomonkey.ApplyMethod(reflect.TypeOf(new(WorkerInfo)), "HandleJobStatus",
		func(_ *WorkerInfo, _ string) string {
			return jobStatus
		})
	defer patch.Reset()
	conveyUpdateCMWhenJobEnd(workerInfo, cm, pod, podKeyInfo, convey.ShouldNotBeNil)
	cm, err := workerInfo.clientSet.CoreV1().ConfigMaps(podKeyInfo.namespace).Get(context.TODO(),
		workerInfo.CMName, metav1.GetOptions{})
	convey.So(cm.Data[JobStatus], convey.ShouldEqual, jobStatus)
	convey.So(err, convey.ShouldBeNil)
}

func conveyUpdateCMWhenJobEnd(workerInfo *WorkerInfo, cm *v1.ConfigMap,
	pod *v1.Pod, podKeyInfo *podIdentifier, assert convey.Assertion) {
	createConfigMaps(workerInfo.clientSet, cm)
	createPod(workerInfo.clientSet, pod)
	err := workerInfo.UpdateCMWhenJobEnd(podKeyInfo)
	convey.So(err, assert)
}

func createConfigMaps(clientSet kubernetes.Interface, cm *v1.ConfigMap) {
	_, err := clientSet.CoreV1().ConfigMaps(mockNamespace).Create(context.TODO(), cm, metav1.CreateOptions{})
	convey.So(err, convey.ShouldBeNil)
}

func createPod(clientSet kubernetes.Interface, pod *v1.Pod) {
	_, err := clientSet.CoreV1().Pods(mockNamespace).Create(context.TODO(), pod, metav1.CreateOptions{})
	convey.So(err, convey.ShouldBeNil)
}

func TestUpdateWithFinish(t *testing.T) {
	convey.Convey("test UpdateWithFinish", t, func() {
		workerInfo := mockWorkerInfo()
		podInfo := mockPodIdentifier()
		patch := gomonkey.ApplyPrivateMethod(reflect.TypeOf(workerInfo), "constructionFinished",
			func(_ *WorkerInfo) bool { return true }).
			ApplyPrivateMethod(reflect.TypeOf(workerInfo), "endConstruction",
				func(_ *WorkerInfo, _ *podIdentifier) error { return fmt.Errorf("update configmap failed") })
		defer patch.Reset()
		err := workerInfo.updateWithFinish(podInfo)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func TestContainerUsedChip(t *testing.T) {
	convey.Convey("test containerUsedChip", t, func() {
		pod := mockPod()
		pod.Spec.Containers = make([]v1.Container, 1)
		patch := gomonkey.ApplyFunc(GetNPUNum, func(c v1.Container) int32 {
			return 1
		})
		defer patch.Reset()
		res := containerUsedChip(pod)
		convey.So(res, convey.ShouldBeTrue)
	})
}

func TestUpdateJobHccLJson(t *testing.T) {
	convey.Convey("test updateJobHccLJson", t, func() {
		cm := mockConfigMap()
		workerInfo := mockWorkerInfo()
		mockGetHccLJsonSlice := gomonkey.ApplyMethod(reflect.TypeOf(new(RankTable)), "GetHccLJsonSlice",
			func(_ *RankTable) []string {
				return make([]string, mockHccLJsonSliceLen)
			})
		defer mockGetHccLJsonSlice.Reset()
		convey.Convey("update job HccLJson failed", func() {
			patch := gomonkey.ApplyFunc(util.CreateOrUpdateCm,
				func(_ kubernetes.Interface, _ *v1.ConfigMap) error {
					return fmt.Errorf("unable to create ConfigMap failed")
				})
			defer patch.Reset()
			err := workerInfo.updateJobHccLJson(*cm)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("update job HccLJson succeed", func() {
			patch := gomonkey.ApplyFunc(util.CreateOrUpdateCm,
				func(_ kubernetes.Interface, _ *v1.ConfigMap) error {
					return nil
				})
			defer patch.Reset()
			err := workerInfo.updateJobHccLJson(*cm)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestGetNPUNum(t *testing.T) {
	convey.Convey("test GetNPUNum", t, func() {
		container := v1.Container{}
		container.Resources.Limits = map[v1.ResourceName]resource.Quantity{
			A910ResourceName: {}}
		result := GetNPUNum(container)
		convey.So(result, convey.ShouldEqual, 0)
		container.Resources.Limits = map[v1.ResourceName]resource.Quantity{
			"test-resource-name": {}}
		result = GetNPUNum(container)
		convey.So(result, convey.ShouldNotEqual, InvalidNPUNum)
	})
}
