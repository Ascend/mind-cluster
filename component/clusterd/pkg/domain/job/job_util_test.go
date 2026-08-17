// Copyright (c) Huawei Technologies Co., Ltd. 2024-2026. All rights reserved.

//go:build !race

// Package job a series of job test function
package job

import (
	"encoding/json"
	"strconv"
	"strings"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"volcano.sh/apis/pkg/apis/scheduling/v1beta1"

	"ascend-common/api"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/custom"
	"clusterd/pkg/domain/pod"
)

const (
	vcJobKey      = "job"
	nodeName1     = "node1"
	nodeName2     = "node2"
	podName1      = "pod1"
	podName2      = "pod2"
	podName3      = "pod3"
	podNameSpace1 = "default"
	podUid1       = "123"
	podUid2       = "456"
	podUid3       = "789"
	deviceName1   = "device1"

	pgMinMember2    = 2
	pgMinMember2Str = "2"
	masterIp        = "127.0.0.1"
	mindIeJobId     = "mindie-ms"
	a5DeviceAnno    = `{"pod_name":"p0","devices":[{"device_id":"0","device_ip":"192.168.0.1","levelList":[{"level":0,"info":{"UB":{"net_layer":0,"net_instance_id":"L0"}}},{"level":1,"info":{"UB":{"net_layer":1,"net_instance_id":"L1"}}},{"level":3,"info":{"ROCE":{"net_layer":3,"net_instance_id":"L3"}}}]}]}`
)

func TestPreDeleteCmAndCache(t *testing.T) {
	convey.Convey("test PreDeleteCmAndCache", t, func() {
		jobInfo := getDemoJob(jobName1, jobNameSpace, jobUid1)
		convey.Convey("test job cache is nil", func() {
			PreDeleteCmAndCache(jobUid1)
			jobInfo1, _ := GetJobCache(jobUid1)
			convey.So(jobInfo1.Status, convey.ShouldEqual, "")
		})
		convey.Convey("when job cache is not nil, preDeleteCM success. job status should be delete", func() {
			mockGetHcclSlice := gomonkey.ApplyFunc(getHcclSlice,
				func(table constant.RankTable) []string {
					return []string{"123"}
				})
			defer mockGetHcclSlice.Reset()
			mockPreDeleteCM := gomonkey.ApplyFunc(preDeleteCM,
				func(jobInfo constant.JobInfo, hccls []string) bool {
					return true
				})
			defer mockPreDeleteCM.Reset()
			SaveJobCache(jobUid1, jobInfo)
			defer DeleteJobCache(jobUid1)
			PreDeleteCmAndCache(jobUid1)
			jobInfo1, _ := GetJobCache(jobUid1)
			convey.So(jobInfo1.Status, convey.ShouldEqual, StatusJobFail)
		})
		convey.Convey("when job cache is not nil, job status is completed, preDeleteCM failed. "+
			"job status should be completed", func() {
			mockGetHcclSlice := gomonkey.ApplyFunc(getHcclSlice,
				func(table constant.RankTable) []string {
					return []string{"123"}
				})
			defer mockGetHcclSlice.Reset()
			mockPreDeleteCM := gomonkey.ApplyFunc(preDeleteCM,
				func(jobInfo constant.JobInfo, hccls []string) bool {
					return false
				})
			defer mockPreDeleteCM.Reset()
			jobInfo.Status = StatusJobCompleted
			SaveJobCache(jobUid1, jobInfo)
			defer DeleteJobCache(jobUid1)
			PreDeleteCmAndCache(jobUid1)
			jobInfo1, _ := GetJobCache(jobUid1)
			convey.So(jobInfo1.Status, convey.ShouldEqual, StatusJobCompleted)
		})
	})
}

func TestDeleteCmAndCache(t *testing.T) {
	convey.Convey("test DeleteCmAndCache", t, func() {
		jobInfo := getDemoJob(jobName1, jobNameSpace, jobUid1)
		jobInfo.IsPreDelete = true
		convey.Convey("when job cache is nil", func() {
			DeleteCmAndCache(jobUid1)
			jobInfo1, _ := GetJobCache(jobUid1)
			convey.So(jobInfo1.Status, convey.ShouldEqual, "")
		})
		convey.Convey("when job cache is not nil, deleteCm failed. job should be exists", func() {
			mockPreDeleteCM := gomonkey.ApplyFunc(deleteCm,
				func(jobInfo constant.JobInfo) bool {
					return false
				})
			defer mockPreDeleteCM.Reset()
			SaveJobCache(jobUid1, jobInfo)
			defer DeleteJobCache(jobUid1)
			DeleteCmAndCache(jobUid1)
			_, ok := GetJobCache(jobUid1)
			convey.So(ok, convey.ShouldEqual, true)
		})
		convey.Convey("when job cache is not nil, deleteCm success. job should not be exists", func() {
			mockPreDeleteCM := gomonkey.ApplyFunc(deleteCm,
				func(jobInfo constant.JobInfo) bool {
					return true
				})
			defer mockPreDeleteCM.Reset()
			SaveJobCache(jobUid1, jobInfo)
			defer DeleteJobCache(jobUid1)
			DeleteCmAndCache(jobUid1)
			_, ok := GetJobCache(jobUid1)
			convey.So(ok, convey.ShouldEqual, false)
		})
		convey.Convey("when job cache is not nil, deleteCm failed but have job with same name. "+
			"job should not be exists", func() {
			mockPreDeleteCM := gomonkey.ApplyFunc(deleteCm,
				func(jobInfo constant.JobInfo) bool {
					return false
				})
			defer mockPreDeleteCM.Reset()
			SaveJobCache(jobUid1, jobInfo)
			defer DeleteJobCache(jobUid1)
			jobInfo2 := getDemoJob(jobName1, jobNameSpace, jobUid2)
			SaveJobCache(jobUid2, jobInfo2)
			defer DeleteJobCache(jobUid2)
			DeleteCmAndCache(jobUid1)
			_, ok := GetJobCache(jobUid1)
			convey.So(ok, convey.ShouldEqual, false)
		})
	})
}

func TestInitCmAndCache(t *testing.T) {
	convey.Convey("test InitCmAndCache", t, func() {
		newPGInfo := getDemoPodGroup(jobName1, jobNameSpace, jobUid1)
		convey.Convey("when pg name is nil, job cache should be nil", func() {
			newPGInfo.Name = ""
			InitCmAndCache(*newPGInfo, nil)
			jobInfoMap := GetAllJobCache()
			convey.So(len(jobInfoMap), convey.ShouldEqual, 0)
		})
		convey.Convey("when pg name is not nil, initCm success. job cache should not be nil", func() {
			mockInitCM := gomonkey.ApplyFunc(initCM,
				func(jobInfo constant.JobInfo) bool {
					return true
				})
			defer mockInitCM.Reset()
			InitCmAndCache(*newPGInfo, nil)
			defer DeleteJobCache(jobUid1)
			jobInfoMap := GetAllJobCache()
			convey.So(len(jobInfoMap), convey.ShouldEqual, 1)
		})
		convey.Convey("when pg name is not nil, initCm failed. job cache should be nil", func() {
			mockInitCM := gomonkey.ApplyFunc(initCM,
				func(jobInfo constant.JobInfo) bool {
					return false
				})
			defer mockInitCM.Reset()
			InitCmAndCache(*newPGInfo, nil)
			defer DeleteJobCache(jobUid1)
			jobInfoMap := GetAllJobCache()
			convey.So(len(jobInfoMap), convey.ShouldEqual, 0)
		})
	})
}

func TestGetJobBasicInfoByPodGroup(t *testing.T) {
	convey.Convey("test getJobBasicInfoByPodGroup success", t, func() {
		newPGInfo := getDemoPodGroup(jobName1, jobNameSpace, jobUid1)
		jobInfo := getJobBasicInfoByPG(*newPGInfo, nil)
		convey.So(jobInfo.Name, convey.ShouldEqual, jobName1)
	})
}

func TestUpdateCmAndCache(t *testing.T) {
	pgDemo := getDemoPodGroup(jobName1, jobNameSpace, jobUid1)
	mockInitRankTableByPod := gomonkey.ApplyFunc(pod.ConstructRankTableByPod, getDemoRankTable)
	defer mockInitRankTableByPod.Reset()
	convey.Convey("test UpdateCmAndCache Job status", t, func() {
		convey.Convey("when jobInfo is nil, updateCM success. job status should be running", func() {
			mockUpdateCM := gomonkey.ApplyFunc(updateCM,
				func(jobInfo constant.JobInfo, index int, hccl string) bool {
					return true
				})
			defer mockUpdateCM.Reset()
			UpdateCmAndCache(StatusJobRunning, "", *pgDemo, map[string]v1.Pod{})
			defer DeleteJobCache(jobUid1)
			jobInfo, ok := GetJobCache(jobUid1)
			convey.So(ok, convey.ShouldEqual, true)
			convey.So(jobInfo.Status, convey.ShouldEqual, StatusJobRunning)
		})
		convey.Convey("when jobInfo is not nil, updateCM success. job status should be running", func() {
			mockUpdateCM := gomonkey.ApplyFunc(updateCM,
				func(jobInfo constant.JobInfo, index int, hccl string) bool {
					return true
				})
			defer mockUpdateCM.Reset()
			UpdateCmAndCache(StatusJobRunning, jobUid1, *pgDemo, map[string]v1.Pod{})
			defer DeleteJobCache(jobUid1)
			jobInfo, ok := GetJobCache(jobUid1)
			convey.So(ok, convey.ShouldEqual, true)
			convey.So(jobInfo.Status, convey.ShouldEqual, StatusJobRunning)
		})
		convey.Convey("when jobInfo is not nil, updateCM failed. job should be nil", func() {
			mockUpdateCM := gomonkey.ApplyFunc(updateCM,
				func(jobInfo constant.JobInfo, index int, hccl string) bool {
					return false
				})
			defer mockUpdateCM.Reset()
			UpdateCmAndCache(StatusJobRunning, jobUid1, *pgDemo, map[string]v1.Pod{})
			defer DeleteJobCache(jobUid1)
			_, ok := GetJobCache(jobUid1)
			convey.So(ok, convey.ShouldEqual, false)
		})
	})
}

func TestInitJobShareTorInfo(t *testing.T) {
	podDemo1 := getDemoPod(podName1, podNameSpace1, podUid1)
	podDemo2 := getDemoPod(podName2, podNameSpace1, podUid2)
	podDemo3 := getDemoPod(podName3, podNameSpace1, podUid3)
	podMapDemo := map[string]v1.Pod{
		podUid1: *podDemo1,
		podUid2: *podDemo2,
		podUid3: *podDemo3,
	}
	convey.Convey("test initJobShareTorInfo", t, func() {
		convey.Convey("when job framework is not pytorch. job masterAddr be nil", func() {
			jobDemo := getDemoJob(jobName1, jobNameSpace, jobUid1)
			initJobShareTorInfo(&jobDemo, podMapDemo)
			convey.So(jobDemo.MasterAddr, convey.ShouldEqual, "")
		})
		convey.Convey("when job framework is pytorch. job masterAddr is masterIp", func() {
			jobDemo := getDemoJob(jobName1, jobNameSpace, jobUid1)
			jobDemo.Framework = ptFramework
			initJobShareTorInfo(&jobDemo, podMapDemo)
			convey.So(jobDemo.MasterAddr, convey.ShouldEqual, masterIp)
		})
		convey.Convey("when job framework is pytorch, jobType is vcjob. job masterAddr is masterIp", func() {
			jobDemo := getDemoJob(jobName1, jobNameSpace, jobUid1)
			jobDemo.JobType = vcJobKind
			jobDemo.Framework = ptFramework
			serverHccl := constant.ServerHccl{
				HostIp: masterIp,
			}
			jobDemo.JobRankTable.ServerList = append(jobDemo.JobRankTable.ServerList, serverHccl)
			initJobShareTorInfo(&jobDemo, podMapDemo)
			convey.So(jobDemo.MasterAddr, convey.ShouldEqual, masterIp)
		})
	})
}

func getDemoPod(name, nameSpace, podUid string) *v1.Pod {
	p := &v1.Pod{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{},
		Spec: v1.PodSpec{
			Containers: []v1.Container{{
				Env: []v1.EnvVar{{
					Name:      masterAddr,
					Value:     masterIp,
					ValueFrom: nil,
				}},
			}},
		},
		Status: v1.PodStatus{},
	}
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
	return p
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
	podGroupInfo.Spec.MinMember = pgMinMember2
	podGroupInfo.Spec.MinResources = &v1.ResourceList{"huawei/Ascend910": resource.Quantity{}}
	return podGroupInfo
}

func getDemoRankTable(_ map[string]v1.Pod, _ int) (constant.RankTable, int) {
	rankTable := constant.RankTable{
		Status:      StatusJobCompleted,
		ServerCount: pgMinMember2Str,
		Total:       1,
		ServerList:  []constant.ServerHccl{},
	}
	return rankTable, pgMinMember2
}

func TestGetJobServerInfoMap(t *testing.T) {
	convey.Convey("test GetJobServerInfoMap", t, func() {
		convey.Convey("when job cache is nil, jobServerInfoMap should be nil", func() {
			jobServerInfoMap := GetJobServerInfoMap()
			convey.So(len(jobServerInfoMap.InfoMap), convey.ShouldEqual, 0)
		})
		convey.Convey("when job cache length is 1, jobServerInfoMap length is 1", func() {
			jobDemo := getDemoJob(jobName1, jobNameSpace, jobUid1)
			serverHccl := constant.ServerHccl{
				ServerID: masterIp,
			}
			deviceList := constant.Device{
				DeviceID: "",
			}
			serverHccl.DeviceList = append(serverHccl.DeviceList, deviceList)
			jobDemo.PreServerList = append(jobDemo.PreServerList, serverHccl)
			SaveJobCache(jobUid1, jobDemo)
			defer DeleteJobCache(jobUid1)
			jobServerInfoMap := GetJobServerInfoMap()
			convey.So(len(jobServerInfoMap.InfoMap), convey.ShouldEqual, 1)
		})
	})
}

func TestGetJobIsRunning(t *testing.T) {
	convey.Convey("test GetJobIsRunning", t, func() {
		jobDemo := getDemoJob(jobName1, jobNameSpace, jobUid1)
		convey.Convey("when job cache is running, return true", func() {
			jobDemo.Status = StatusJobRunning
			SaveJobCache(jobUid1, jobDemo)
			defer DeleteJobCache(jobUid1)
			convey.So(GetJobIsRunning(jobUid1), convey.ShouldBeTrue)
		})
		convey.Convey("when job cache is not running, return false", func() {
			jobDemo.Status = StatusJobPending
			SaveJobCache(jobUid1, jobDemo)
			defer DeleteJobCache(jobUid1)
			convey.So(GetJobIsRunning(jobUid1), convey.ShouldBeFalse)
		})
	})
}

func TestGetJobIsExists(t *testing.T) {
	convey.Convey("test GetJobIsExists", t, func() {
		convey.Convey("when job cache is not exists, return false", func() {
			convey.So(GetJobIsExists(jobUid1), convey.ShouldBeFalse)
		})
		convey.Convey("when job cache is exists, return true", func() {
			jobDemo := getDemoJob(jobName1, jobNameSpace, jobUid1)
			SaveJobCache(jobUid1, jobDemo)
			defer DeleteJobCache(jobUid1)
			convey.So(GetJobIsExists(jobUid1), convey.ShouldBeTrue)
		})
	})
}

func TestFlushLastUpdateTime(t *testing.T) {
	convey.Convey("test FlushLastUpdateTime", t, func() {
		convey.Convey("when job cache is not exists, flush should be failed", func() {
			FlushLastUpdateTime(jobUid1)
			_, ok := GetJobCache(jobUid1)
			convey.So(ok, convey.ShouldBeFalse)
		})
		convey.Convey("when job cache is exists, return true", func() {
			jobDemo := getDemoJob(jobName1, jobNameSpace, jobUid1)
			SaveJobCache(jobUid1, jobDemo)
			defer DeleteJobCache(jobUid1)
			FlushLastUpdateTime(jobUid1)
			jonInfo, ok := GetJobCache(jobUid1)
			convey.So(ok, convey.ShouldBeTrue)
			convey.So(jonInfo.LastUpdatedCmTime, convey.ShouldNotBeZeroValue)
		})
	})
}

func TestIsInferenceJob(t *testing.T) {
	convey.Convey("test IsMindIeServerPod", t, func() {
		podInfo := getDemoPod(jobName1, jobNameSpace, podUid1)
		convey.Convey("when job is not mindie server job, return false", func() {
			convey.So(IsMindIeServerPod(*podInfo), convey.ShouldBeFalse)
		})
		convey.Convey("when job is mindie server job, return true", func() {
			podInfo.Labels = map[string]string{}
			podInfo.Labels[constant.MindIeJobIdLabelKey] = mindIeJobId
			podInfo.Labels[constant.MindIeAppTypeLabelKey] = constant.ServerAppType
			convey.So(IsMindIeServerPod(*podInfo), convey.ShouldBeTrue)
		})
	})
}

func TestGetMindIeServerJobDeviceInfoMap(t *testing.T) {
	convey.Convey("test GetCustomFilterFaultJobAndUsedDeviceInfoMap", t, func() {
		convey.Convey("if there is an mindie server job on the current node, return mindie server jobId", func() {
			demoPod := getDemoPod(jobName1, jobNameSpace, podUid1)
			demoPod.Spec = v1.PodSpec{NodeName: nodeName1}
			patch := gomonkey.ApplyFuncReturn(GetAllJobCache, map[string]constant.JobInfo{
				jobUid1: getDemoJob(jobName1, jobNameSpace, jobUid1),
			}).ApplyFuncReturn(pod.GetPodByJobId, map[string]v1.Pod{
				podUid1: *demoPod,
			}).ApplyFuncReturn(custom.JudgeFilterFaultAnnosByJobKey, true)
			defer patch.Reset()
			jobInfoMap, deviceInfoMap := GetCustomFilterFaultJobAndUsedDeviceInfoMap()
			convey.So(jobInfoMap, convey.ShouldNotBeEmpty)
			convey.So(deviceInfoMap, convey.ShouldNotBeEmpty)
		})
		convey.Convey("if there is no mindie server job on the current node, return mindie server jobId", func() {
			jobInfoMap, deviceInfoMap := GetCustomFilterFaultJobAndUsedDeviceInfoMap()
			convey.So(jobInfoMap, convey.ShouldBeEmpty)
			convey.So(deviceInfoMap, convey.ShouldBeEmpty)
		})
	})
}

func TestDeepCopyJobInfo(t *testing.T) {
	convey.Convey("test DeepCopyJobInfo", t, func() {
		jobDemo := getDemoJob(jobName1, jobNameSpace, jobUid1)
		jobDemo.NodeNames = map[string]string{podName1: nodeName1}
		serverHccl := constant.ServerHccl{
			ServerID: masterIp,
		}
		deviceList := constant.Device{
			DeviceID: deviceName1,
		}
		serverHccl.DeviceList = append(serverHccl.DeviceList, deviceList)
		jobDemo.PreServerList = append(jobDemo.PreServerList, serverHccl)
		jobDemo.JobRankTable = constant.RankTable{
			Status:      StatusJobCompleted,
			ServerCount: pgMinMember2Str,
			Total:       1,
			ServerList:  jobDemo.PreServerList,
		}

		copyJobInfo := DeepCopyJobInfo(&jobDemo)
		convey.So(*copyJobInfo, convey.ShouldResemble, jobDemo)

		jobDemo.Name = jobName2
		jobDemo.Key = jobUid2
		jobDemo.NodeNames = map[string]string{podName2: nodeName2}
		jobDemo.PreServerList = []constant.ServerHccl{}
		jobDemo.JobRankTable.ServerList = []constant.ServerHccl{}
		convey.So(jobDemo.Key, convey.ShouldNotEqual, copyJobInfo.Key)
		convey.So(jobDemo.NodeNames, convey.ShouldNotResemble, copyJobInfo.NodeNames)
		convey.So(jobDemo.PreServerList, convey.ShouldNotResemble, copyJobInfo.PreServerList)
		convey.So(jobDemo.JobRankTable.ServerList, convey.ShouldNotResemble, copyJobInfo.JobRankTable.ServerList)
	})
}

func TestGetSidFromLabels(t *testing.T) {
	convey.Convey("test getSidFromLabels", t, func() {
		convey.Convey("nil labels", func() {
			convey.So(getSidFromLabels(nil), convey.ShouldEqual, "")
		})

		convey.Convey("only CustomJobKeyLabel", func() {
			convey.So(getSidFromLabels(map[string]string{constant.CustomJobKeyLabel: "custom-key"}), convey.ShouldEqual, "")
		})

		convey.Convey("CustomJobKeyLabel with value", func() {
			labels := map[string]string{constant.CustomJobKeyLabel: "custom-key", "custom-key": "sid-value"}
			convey.So(getSidFromLabels(labels), convey.ShouldEqual, "sid-value")
		})

		convey.Convey("only CustomJobIdLabel", func() {
			convey.So(getSidFromLabels(map[string]string{constant.CustomJobIdLabel: "job-id-value"}), convey.ShouldEqual, "job-id-value")
		})

		convey.Convey("CustomJobIdLabel with spaces", func() {
			convey.So(getSidFromLabels(map[string]string{constant.CustomJobIdLabel: "  job-id-value  "}), convey.ShouldEqual, "job-id-value")
		})

		convey.Convey("CustomJobKeyLabel with spaces in value", func() {
			labels := map[string]string{constant.CustomJobKeyLabel: "custom-key", "custom-key": "  sid-value  "}
			convey.So(getSidFromLabels(labels), convey.ShouldEqual, "sid-value")
		})

		convey.Convey("CustomJobIdLabel only spaces", func() {
			convey.So(getSidFromLabels(map[string]string{constant.CustomJobIdLabel: "   "}), convey.ShouldEqual, "")
		})

		convey.Convey("both labels: CustomJobKeyLabel precedence", func() {
			labels := map[string]string{constant.CustomJobKeyLabel: "custom-key", "custom-key": "sid-value", constant.CustomJobIdLabel: "job-id-value"}
			convey.So(getSidFromLabels(labels), convey.ShouldEqual, "sid-value")
		})

		convey.Convey("no relevant labels", func() {
			convey.So(getSidFromLabels(map[string]string{"other-label": "other-value"}), convey.ShouldEqual, "")
		})
	})
}

func TestGetSidForJobInfo(t *testing.T) {
	convey.Convey("test getSidForJobInfo", t, func() {
		pgInfo := *getDemoPodGroup(jobName1, jobNameSpace, jobUid1)
		pod1 := *getDemoPod(podName1, jobNameSpace, podUid1)
		pod2 := *getDemoPod(podName2, jobNameSpace, podUid2)
		podsInJob := map[string]v1.Pod{podUid1: pod1, podUid2: pod2}

		convey.Convey("podgroup has sid", func() {
			pgInfo.Labels = map[string]string{constant.CustomJobIdLabel: "pg-sid"}
			convey.So(getSidForJobInfo(pgInfo, podsInJob), convey.ShouldEqual, "pg-sid")
		})

		convey.Convey("podgroup no sid, first pod has sid", func() {
			pgInfo.Labels = nil
			pod1.Labels = map[string]string{constant.CustomJobIdLabel: "pod1-sid"}
			pod2.Labels = nil
			podsInJob[podUid1] = pod1
			podsInJob[podUid2] = pod2
			convey.So(getSidForJobInfo(pgInfo, podsInJob), convey.ShouldEqual, "pod1-sid")
		})

		convey.Convey("podgroup no sid, second pod has sid", func() {
			pgInfo.Labels = nil
			pod1.Labels = nil
			pod2.Labels = map[string]string{constant.CustomJobIdLabel: "pod2-sid"}
			podsInJob[podUid1] = pod1
			podsInJob[podUid2] = pod2
			convey.So(getSidForJobInfo(pgInfo, podsInJob), convey.ShouldEqual, "pod2-sid")
		})

		convey.Convey("neither has sid", func() {
			pgInfo.Labels = nil
			pod1.Labels = nil
			pod2.Labels = nil
			podsInJob[podUid1] = pod1
			podsInJob[podUid2] = pod2
			expectedJobKey := string(pgInfo.GetOwnerReferences()[0].UID)
			convey.So(getSidForJobInfo(pgInfo, podsInJob), convey.ShouldEqual, expectedJobKey)
		})
	})
}

func TestGetHcclSlice(t *testing.T) {
	convey.Convey("test getHcclSlice", t, func() {
		convey.Convey("non A5 job: rankList is nil, hccl json should not contain v2.0 fields", func() {
			table := constant.RankTable{
				ServerCount: "3",
				Total:       1,
				ServerList: []constant.ServerHccl{
					{ServerID: "192.168.0.1", HostIp: "192.168.0.1", ServerName: nodeName1,
						DeviceList: []constant.Device{{DeviceID: "0", DeviceIP: "192.168.0.1", RankID: "0"}}},
					{ServerID: "192.168.0.2", HostIp: "192.168.0.2", ServerName: nodeName2,
						DeviceList: []constant.Device{{DeviceID: "0", DeviceIP: "192.168.0.2", RankID: "1"}}},
					{ServerID: "192.168.0.3", HostIp: "192.168.0.3", ServerName: podName3,
						DeviceList: []constant.Device{{DeviceID: "0", DeviceIP: "192.168.0.3", RankID: "2"}}},
				},
			}
			hccls := getHcclSlice(table)
			convey.So(len(hccls), convey.ShouldEqual, 1)
			rankTable := constant.RankTable{}
			convey.So(json.Unmarshal([]byte(hccls[0]), &rankTable), convey.ShouldBeNil)
			convey.So(rankTable.RankList, convey.ShouldBeNil)
			convey.So(rankTable.RankCount, convey.ShouldEqual, 0)
			convey.So(rankTable.Version, convey.ShouldEqual, "")
			convey.So(len(rankTable.ServerList), convey.ShouldEqual, 3)
		})
		convey.Convey("A5 single slice job: hccl json should serialize the whole table including rankList", func() {
			table := buildServerRankTable(1)
			table.ServerList[0].ServerName = nodeName1
			table.ServerList[0].DeviceList = table.ServerList[0].DeviceList[:1]
			table.RankList = table.RankList[:1]
			table.RankCount = 1
			hccls := getHcclSlice(table)
			convey.So(len(hccls), convey.ShouldEqual, 1)
			convey.So(strings.Contains(hccls[0], `"rank_list"`), convey.ShouldBeTrue)
			convey.So(strings.Contains(hccls[0], `"version":"2.0"`), convey.ShouldBeTrue)
			rankTable := constant.RankTable{}
			convey.So(json.Unmarshal([]byte(hccls[0]), &rankTable), convey.ShouldBeNil)
			convey.So(len(rankTable.RankList), convey.ShouldEqual, 1)
			convey.So(rankTable.RankList[0].RankID, convey.ShouldEqual, 0)
			convey.So(rankTable.Version, convey.ShouldEqual, "2.0")
			convey.So(rankTable.RankCount, convey.ShouldEqual, 1)
		})
	})
}

// buildServerRankTable builds n servers (each 2 devices) plus an aligned rankList 0..2n-1.
func buildServerRankTable(n int) constant.RankTable {
	serverList := make([]constant.ServerHccl, 0, n)
	for s := 0; s < n; s++ {
		serverList = append(serverList, constant.ServerHccl{
			ServerID: "192.168.0." + strconv.Itoa(s+1),
			HostIp:   "192.168.0." + strconv.Itoa(s+1),
			DeviceList: []constant.Device{
				{DeviceID: "0", RankID: strconv.Itoa(s * 2)},
				{DeviceID: "1", RankID: strconv.Itoa(s*2 + 1)},
			},
		})
	}
	rankList := make([]constant.Rank, 0, n*2)
	for r := 0; r < n*2; r++ {
		rankList = append(rankList, constant.Rank{RankID: r, LocalID: r, DeviceID: r})
	}
	return constant.RankTable{
		Version:     "2.0",
		RankCount:   n * 2,
		ServerCount: strconv.Itoa(n),
		ServerList:  serverList,
		RankList:    rankList,
	}
}

// sliceMaxSize returns the marshaled byte size of the first serverCount servers (+ their ranks).
func sliceMaxSize(table constant.RankTable, serverCount int) int {
	part := table
	part.ServerList = table.ServerList[:serverCount]
	part.RankList = table.RankList[:serverCount*2]
	b, _ := json.Marshal(part)
	return len(b)
}

func TestGetHcclSliceMultiSliceAlignment(t *testing.T) {
	convey.Convey("test getHcclSliceBySize rankList alignment across multiple slices", t, func() {
		table := buildServerRankTable(5)
		// maxSize = marshaled size of a 2-server sub-table; byte threshold splits 5 servers into slices [2,2,1].
		maxSize := sliceMaxSize(table, 2)
		hccls := getHcclSliceBySize(table, maxSize)
		convey.So(len(hccls), convey.ShouldEqual, 3)

		expectedRanks := [][]int{{0, 1, 2, 3}, {4, 5, 6, 7}, {8, 9}}
		expectedServers := []int{2, 2, 1}
		var allRankIds []int
		for i, hccl := range hccls {
			part := constant.RankTable{}
			convey.So(json.Unmarshal([]byte(hccl), &part), convey.ShouldBeNil)
			convey.So(len(part.ServerList), convey.ShouldEqual, expectedServers[i])
			convey.So(len(part.RankList), convey.ShouldEqual, len(expectedRanks[i]))
			for j, rank := range part.RankList {
				convey.So(rank.RankID, convey.ShouldEqual, expectedRanks[i][j])
				allRankIds = append(allRankIds, rank.RankID)
			}
		}
		convey.So(allRankIds, convey.ShouldResemble, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9})
	})
}

func TestGetHcclSliceBySize(t *testing.T) {
	convey.Convey("test getHcclSliceBySize", t, func() {
		convey.Convey("byte-threshold slicing: each slice fits maxSize and keeps full-table semantics", func() {
			table := buildServerRankTable(6)
			// +64 slack for rank-id digit growth in later slices (e.g. rank 10/11), still far below a 3-server part.
			maxSize := sliceMaxSize(table, 2) + 64
			hccls := getHcclSliceBySize(table, maxSize)
			convey.So(len(hccls), convey.ShouldEqual, 3)
			for _, hccl := range hccls {
				part := constant.RankTable{}
				convey.So(json.Unmarshal([]byte(hccl), &part), convey.ShouldBeNil)
				rb, err := json.Marshal(part)
				convey.So(err, convey.ShouldBeNil)
				convey.So(len(rb), convey.ShouldBeLessThanOrEqualTo, maxSize)
				convey.So(part.Total, convey.ShouldEqual, len(hccls))
				convey.So(part.ServerCount, convey.ShouldEqual, "6")
			}
		})
		convey.Convey("empty ServerList returns nil", func() {
			convey.So(getHcclSliceBySize(constant.RankTable{}, 1024), convey.ShouldBeNil)
		})
		convey.Convey("single oversized server is emitted as one slice with its ranks", func() {
			table := buildServerRankTable(1)
			table.ServerList[0].ContainerIds = map[string]string{"c0": strings.Repeat("x", 4*1024)}
			maxSize := 1024 // smaller than the single server's marshaled size
			hccls := getHcclSliceBySize(table, maxSize)
			convey.So(len(hccls), convey.ShouldEqual, 1)
			part := constant.RankTable{}
			convey.So(json.Unmarshal([]byte(hccls[0]), &part), convey.ShouldBeNil)
			convey.So(len(part.ServerList), convey.ShouldEqual, 1)
			convey.So(len(part.RankList), convey.ShouldEqual, 2)
			convey.So(part.Total, convey.ShouldEqual, 1)
		})
		convey.Convey("non-A5 regression: nil RankList slices keep v1.0 semantics", func() {
			serverList := make([]constant.ServerHccl, 0, 4)
			for s := 0; s < 4; s++ {
				serverList = append(serverList, constant.ServerHccl{
					ServerID:   "192.168.0." + strconv.Itoa(s+1),
					HostIp:     "192.168.0." + strconv.Itoa(s+1),
					ServerName: nodeName1,
					DeviceList: []constant.Device{{DeviceID: "0", DeviceIP: "192.168.0." + strconv.Itoa(s+1), RankID: strconv.Itoa(s)}},
				})
			}
			table := constant.RankTable{ServerCount: "4", ServerList: serverList}
			part := table
			part.ServerList = table.ServerList[0:1]
			b, _ := json.Marshal(part)
			maxSize := len(b)
			hccls := getHcclSliceBySize(table, maxSize)
			convey.So(len(hccls), convey.ShouldEqual, 4)
			for _, hccl := range hccls {
				convey.So(strings.Contains(hccl, `"rank_list"`), convey.ShouldBeFalse)
				part := constant.RankTable{}
				convey.So(json.Unmarshal([]byte(hccl), &part), convey.ShouldBeNil)
				convey.So(part.RankList, convey.ShouldBeNil)
				convey.So(part.RankCount, convey.ShouldEqual, 0)
				convey.So(part.Version, convey.ShouldEqual, "")
				convey.So(len(part.ServerList), convey.ShouldEqual, 1)
			}
		})
	})
}

func TestUpdateCmAndCacheA5(t *testing.T) {
	convey.Convey("test UpdateCmAndCache for A5 job", t, func() {
		convey.Convey("when job resourceType is npu, rankList should be generated with version 2.0", func() {
			podDemo := getDemoPod(podName1, podNameSpace1, podUid1)
			podDemo.Spec.NodeName = nodeName1
			podDemo.Annotations = map[string]string{
				api.PodRankIndexAnno: "0",
				api.PodNPUDeviceAnno: a5DeviceAnno,
			}
			podsInJob := map[string]v1.Pod{podUid1: *podDemo}
			pgDemo := getDemoPodGroup(jobName1, jobNameSpace, jobUid1)
			jobInfo := getDemoJob(jobName1, jobNameSpace, jobUid1)
			jobInfo.ResourceType = api.NPULowerCase
			jobInfo.Replicas = 1
			jobInfo.TotalCmNum = 1
			SaveJobCache(jobUid1, jobInfo)
			defer DeleteJobCache(jobUid1)
			mockUpdateCM := gomonkey.ApplyFunc(updateCM,
				func(jobInfo constant.JobInfo, index int, hccl string) bool {
					return true
				})
			defer mockUpdateCM.Reset()
			UpdateCmAndCache(StatusJobRunning, jobUid1, *pgDemo, podsInJob)
			jobInfo, ok := GetJobCache(jobUid1)
			convey.So(ok, convey.ShouldBeTrue)
			convey.So(jobInfo.JobRankTable.Version, convey.ShouldEqual, "2.0")
			convey.So(jobInfo.JobRankTable.RankCount, convey.ShouldEqual, 1)
			convey.So(len(jobInfo.JobRankTable.RankList), convey.ShouldEqual, 1)
			convey.So(jobInfo.JobRankTable.RankList[0].RankID, convey.ShouldEqual, 0)
			convey.So(len(jobInfo.JobRankTable.ServerList), convey.ShouldEqual, 1)
		})
	})
}
