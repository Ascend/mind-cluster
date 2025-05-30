// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package common is grpc common types and functions
package common

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"volcano.sh/apis/pkg/apis/scheduling/v1beta1"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
	"clusterd/pkg/domain/pod"
	"clusterd/pkg/interface/grpc/recover"
	"clusterd/pkg/interface/kube"
)

var deviceNumPerNode = 8
var randomLen = 16
var uuidLen16 = 49
var uuidLen32 = 81
var errorRank = "errorRank"

func init() {
	hwLogConfig := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	hwlog.InitRunLogger(&hwLogConfig, context.Background())
}

func fakeJobInfo() JobBaseInfo {
	return JobBaseInfo{
		JobId:         "testJobId",
		JobName:       "testJobName",
		PgName:        "testPgName",
		Namespace:     "default",
		RecoverConfig: RecoverConfig{},
	}
}

func fakePG() *v1beta1.PodGroup {
	pg := &v1beta1.PodGroup{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{},
		Spec:       v1beta1.PodGroupSpec{},
		Status:     v1beta1.PodGroupStatus{},
	}
	pg.ObjectMeta.Labels = make(map[string]string)
	pg.ObjectMeta.Annotations = make(map[string]string)
	return pg
}

func fakePodMap() map[string]v1.Pod {
	podMap := make(map[string]v1.Pod)
	podMap["0"] = v1.Pod{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			UID:  "rank0PodUid",
			Name: "fakePodName0",
		},
		Spec:   v1.PodSpec{},
		Status: v1.PodStatus{},
	}
	podMap["1"] = v1.Pod{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			UID:  "rank1PodUid",
			Name: "fakePodName1",
		},
		Spec:   v1.PodSpec{},
		Status: v1.PodStatus{},
	}
	podMap["2"] = v1.Pod{}
	return podMap
}

func fakeAllFaultRescheduledPodMap() map[string]string {
	return map[string]string{
		"1": "newRank1PodUid",
	}
}

func fakeFaultPodRescheduleNotComplete() map[string]string {
	return map[string]string{
		"1": "rank1PodUid",
	}
}

func TestChangeProcessRecoverEnableMode(t *testing.T) {
	patch := gomonkey.ApplyFunc(kube.RetryPatchPodGroupLabel, func(pgName, nameSpace string,
		retryTimes int, labels map[string]string) (*v1beta1.PodGroup, error) {
		if labels == nil {
			return nil, errors.New("empty label")
		}
		pg := &v1beta1.PodGroup{}
		pg.Labels = labels
		return pg, nil
	})
	defer patch.Reset()
	convey.Convey("Test ChangeProcessRecoverEnableMode", t, func() {
		pg, err := ChangeProcessRecoverEnableMode(fakeJobInfo(), constant.ProcessRecoverEnable)
		convey.So(err, convey.ShouldBeNil)
		convey.So(pg, convey.ShouldNotBeNil)
		convey.So(pg.Labels, convey.ShouldNotBeNil)
		convey.So(pg.Labels[constant.ProcessRecoverEnableLabel], convey.ShouldEqual, constant.ProcessRecoverEnable)
	})
}

func TestCheckProcessRecoverOpen(t *testing.T) {
	convey.Convey("Test CheckProcessRecoverOpen", t, func() {
		info := fakeJobInfo()
		convey.Convey("case pg get error", func() {
			patch := gomonkey.ApplyFunc(kube.GetPodGroup,
				func(pgName, nameSpace string) (*v1beta1.PodGroup, error) {
					return nil, errors.New("fake get pg error")
				})
			defer patch.Reset()
			isOpen := CheckProcessRecoverOpen(info.PgName, info.Namespace)
			convey.So(isOpen, convey.ShouldBeFalse)
		})
		convey.Convey("case not have process-recover-enable key", func() {
			patch := gomonkey.ApplyFunc(kube.GetPodGroup,
				func(pgName, nameSpace string) (*v1beta1.PodGroup, error) {
					return fakePG(), nil
				})
			defer patch.Reset()
			isOpen := CheckProcessRecoverOpen(info.PgName, info.Namespace)
			convey.So(isOpen, convey.ShouldBeFalse)
		})
		convey.Convey("case process-recover-enable key not open", func() {
			patch := gomonkey.ApplyFunc(kube.GetPodGroup,
				func(pgName, nameSpace string) (*v1beta1.PodGroup, error) {
					pg := fakePG()
					pg.Labels[constant.ProcessRecoverEnableLabel] = constant.ProcessRecoverPause
					return pg, nil
				})
			defer patch.Reset()
			isOpen := CheckProcessRecoverOpen(info.PgName, info.Namespace)
			convey.So(isOpen, convey.ShouldBeFalse)
		})
		convey.Convey("case process-recover-enable key open", func() {
			patch := gomonkey.ApplyFunc(kube.GetPodGroup,
				func(pgName, nameSpace string) (*v1beta1.PodGroup, error) {
					pg := fakePG()
					pg.Labels[constant.ProcessRecoverEnableLabel] = constant.ProcessRecoverEnable
					return pg, nil
				})
			defer patch.Reset()
			isOpen := CheckProcessRecoverOpen(info.PgName, info.Namespace)
			convey.So(isOpen, convey.ShouldBeTrue)
		})
	})
}

func TestFaultPodAllRescheduled(t *testing.T) {
	convey.Convey("Test FaultPodAllRescheduled", t, func() {
		info := fakeJobInfo()
		podMap := fakePodMap()
		convey.Convey("case pod name is empty", func() {
			patch := gomonkey.ApplyFunc(pod.GetPodByRankIndex,
				func(jobId, rankIndex string) v1.Pod {
					emptyPod := v1.Pod{}
					emptyPod.Name = ""
					return emptyPod
				})
			defer patch.Reset()
			completed := FaultPodAllRescheduled(info.JobId, fakeAllFaultRescheduledPodMap())
			convey.So(completed, convey.ShouldBeFalse)
		})
		patch := gomonkey.ApplyFunc(pod.GetPodByRankIndex,
			func(jobId, rankIndex string) v1.Pod {
				if rankPod, ok := podMap[rankIndex]; ok {
					return rankPod
				}
				emptyPod := v1.Pod{}
				emptyPod.Name = ""
				return emptyPod
			})
		defer patch.Reset()
		convey.Convey("case all fault pod rescheduled", func() {
			completed := FaultPodAllRescheduled(info.JobId, fakeAllFaultRescheduledPodMap())
			convey.So(completed, convey.ShouldBeTrue)
		})
		convey.Convey("case fault pod not rescheduled completely", func() {
			completed := FaultPodAllRescheduled(info.JobId, fakeFaultPodRescheduleNotComplete())
			convey.So(completed, convey.ShouldBeFalse)
		})
	})
}

func TestFaults2Ranks(t *testing.T) {
	convey.Convey("Test Faults2Ranks", t, func() {
		faults := []*pb.FaultRank{
			&pb.FaultRank{RankId: "0", FaultType: "0"},
			&pb.FaultRank{RankId: "1", FaultType: "0"},
		}
		ranks := Faults2Ranks(faults)
		convey.So(len(ranks), convey.ShouldEqual, len(faults))
		ranks = Faults2Ranks(nil)
		convey.So(len(ranks), convey.ShouldEqual, 0)
	})
}

func TestFaults2String(t *testing.T) {
	convey.Convey("Test Faults2String", t, func() {
		faults := []*pb.FaultRank{
			&pb.FaultRank{RankId: "0", FaultType: "0"},
			&pb.FaultRank{RankId: "1", FaultType: "0"},
		}
		str := Faults2String(faults)
		convey.So(str, convey.ShouldEqual, "0:0,1:0")
		str = Faults2String(nil)
		convey.So(str, convey.ShouldEqual, "")
	})
}

func TestGetFaultRankIdsInSameNode(t *testing.T) {
	convey.Convey("Test GetFaultRankIdsInSameNode", t, func() {
		faultRanks := []string{"0", "1"}
		res := GetFaultRankIdsInSameNode(faultRanks, deviceNumPerNode)
		convey.So(len(res), convey.ShouldEqual, deviceNumPerNode)
		res = GetFaultRankIdsInSameNode(nil, deviceNumPerNode)
		convey.So(len(res), convey.ShouldEqual, 0)
	})
}

func TestGetPodMap(t *testing.T) {
	convey.Convey("Test GetPodMap", t, func() {
		info := fakeJobInfo()
		podMap := fakePodMap()
		patch := gomonkey.ApplyFunc(pod.GetPodByRankIndex,
			func(jobId, rankIndex string) v1.Pod {
				if rankPod, ok := podMap[rankIndex]; ok {
					return rankPod
				}
				emptyPod := v1.Pod{}
				emptyPod.Name = ""
				return emptyPod
			})
		defer patch.Reset()
		patch1 := gomonkey.ApplyFunc(pod.GetPodDeviceNumByJobId,
			func(jobId string) int {
				return deviceNumPerNode
			})
		defer patch1.Reset()
		mp, err := GetPodMap(info.JobId, []string{"8", "9", "16"})
		convey.So(err, convey.ShouldEqual, nil)
		convey.So(len(mp), convey.ShouldEqual, 1)
		convey.So(mp["1"], convey.ShouldEqual, "rank1PodUid")
		convey.Convey("case device num is zero", caseDeviceNumZeroForGetPodMap)
		convey.Convey("case rank str illegal", caseIllegalRankForGetPodMap)
	})
}

func caseDeviceNumZeroForGetPodMap() {
	info := fakeJobInfo()
	patch := gomonkey.ApplyFunc(pod.GetPodDeviceNumByJobId,
		func(jobId string) int {
			return 0
		})
	defer patch.Reset()
	_, err := GetPodMap(info.JobId, []string{"8"})
	convey.ShouldNotBeNil(err)
}

func caseIllegalRankForGetPodMap() {
	info := fakeJobInfo()
	patch := gomonkey.ApplyFunc(pod.GetPodDeviceNumByJobId,
		func(jobId string) int {
			return deviceNumPerNode
		})
	defer patch.Reset()
	_, err := GetPodMap(info.JobId, []string{errorRank})
	convey.ShouldNotBeNil(err)
}

func TestGetRecoverBaseInfo(t *testing.T) {
	convey.Convey("Test GetRecoverBaseInfo", t, func() {
		info := fakeJobInfo()
		convey.Convey("case get pod group error", func() {
			patch := gomonkey.ApplyFunc(kube.RetryGetPodGroup,
				func(name, namespace string, retryTimes int) (*v1beta1.PodGroup, error) {
					return nil, errors.New("fake get pod group error")
				})
			defer patch.Reset()
			config, code, err := GetRecoverBaseInfo(info.PgName, info.Namespace)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(code, convey.ShouldNotEqual, OK)
			convey.So(config.ProcessRecoverEnable, convey.ShouldBeFalse)
			convey.So(config.PlatFormMode, convey.ShouldBeFalse)
		})
		convey.Convey("case get pod group success, and process-recover-enable off", func() {
			patch := gomonkey.ApplyFunc(kube.RetryGetPodGroup,
				func(name, namespace string, retryTimes int) (*v1beta1.PodGroup, error) {
					pg := fakePG()
					pg.Labels[constant.ProcessRecoverEnableLabel] = constant.ProcessRecoverInit
					return pg, nil
				})
			defer patch.Reset()
			config, code, err := GetRecoverBaseInfo(info.PgName, info.Namespace)
			convey.So(err, convey.ShouldBeNil)
			convey.So(code, convey.ShouldEqual, OK)
			convey.So(config.ProcessRecoverEnable, convey.ShouldBeFalse)
			convey.So(config.GraceExit, convey.ShouldBeFalse)
			convey.So(config.PlatFormMode, convey.ShouldBeFalse)
		})
		convey.Convey("case get pod group success, and process-recover-enable on", func() {
			patch := gomonkey.ApplyFunc(kube.RetryGetPodGroup,
				func(name, namespace string, retryTimes int) (*v1beta1.PodGroup, error) {
					pg := fakePG()
					pg.Labels[constant.ProcessRecoverEnableLabel] = constant.ProcessRecoverEnable
					pg.Labels[constant.SubHealthyStrategy] = constant.SubHealthyGraceExit
					return pg, nil
				})
			defer patch.Reset()
			config, code, err := GetRecoverBaseInfo(info.PgName, info.Namespace)
			convey.So(err, convey.ShouldBeNil)
			convey.So(code, convey.ShouldEqual, OK)
			convey.So(config.ProcessRecoverEnable, convey.ShouldBeTrue)
			convey.So(config.GraceExit, convey.ShouldBeTrue)
			convey.So(config.PlatFormMode, convey.ShouldBeFalse)
		})
		addTestCaseForLabelNotExist(info.PgName, info.Namespace)
	})
}

func addTestCaseForLabelNotExist(name, namespace string) {
	convey.Convey("case pod group don't have key ProcessRecoverEnableLabel", func() {
		patch := gomonkey.ApplyFunc(kube.RetryGetPodGroup,
			func(name, namespace string, retryTimes int) (*v1beta1.PodGroup, error) {
				pg := fakePG()
				pg.Annotations[constant.RecoverStrategies] = constant.ProcessRetryStrategyName
				return pg, nil
			})
		defer patch.Reset()
		config, _, _ := GetRecoverBaseInfo(name, namespace)
		convey.So(config.ProcessRecoverEnable, convey.ShouldBeFalse)
	})
}

func TestIsUceFault(t *testing.T) {
	convey.Convey("Test IsUceFault", t, func() {
		convey.Convey("case uce fault", func() {
			faults := []*pb.FaultRank{
				&pb.FaultRank{RankId: "0", FaultType: "0"},
				&pb.FaultRank{RankId: "1", FaultType: "0"},
			}
			flag := IsUceFault(faults)
			convey.So(flag, convey.ShouldBeTrue)
		})
		convey.Convey("case normal fault", func() {
			faults := []*pb.FaultRank{
				&pb.FaultRank{RankId: "0", FaultType: "0"},
				&pb.FaultRank{RankId: "1", FaultType: "1"},
			}
			flag := IsUceFault(faults)
			convey.So(flag, convey.ShouldBeFalse)
		})
	})
}

func TestLabelFaultPod(t *testing.T) {
	convey.Convey("Test LabelFaultPod", t, func() {
		info := fakeJobInfo()
		convey.Convey("case label success", func() {
			patch := gomonkey.ApplyFunc(pod.GetPodDeviceNumByJobId,
				func(jobId string) int {
					return deviceNumPerNode
				})
			defer patch.Reset()
			patch1 := gomonkey.ApplyFuncReturn(labelPodFault, map[string]string{"1": "rank1PodUid"}, nil)
			defer patch1.Reset()
			mp, err := LabelFaultPod(info.JobId, []string{"8"}, nil, "")
			convey.So(err, convey.ShouldBeNil)
			convey.So(len(mp), convey.ShouldEqual, 1)
		})
		convey.Convey("case label fail", func() {
			patch := gomonkey.ApplyFunc(pod.GetPodDeviceNumByJobId,
				func(jobId string) int {
					return deviceNumPerNode
				})
			defer patch.Reset()
			patch1 := gomonkey.ApplyFuncReturn(labelPodFault,
				map[string]string{"1": "rank1PodUid"}, errors.New("fake error"))
			defer patch1.Reset()
			mp, err := LabelFaultPod(info.JobId, []string{"8"}, nil, "")
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(len(mp), convey.ShouldEqual, 1)
		})
		convey.Convey("case device num per node is zero", caseDeviceNumPerNodeZero)
		convey.Convey("case rank string illegal", caseRankStrIllegal)
	})
}

func TestLabelPodFault(t *testing.T) {
	convey.Convey("Test labelPodFault", t, func() {
		info := fakeJobInfo()
		podMap := fakePodMap()
		patch := gomonkey.ApplyFunc(pod.GetPodByRankIndex,
			func(jobId, rankIndex string) v1.Pod {
				if rankPod, ok := podMap[rankIndex]; ok {
					return rankPod
				}
				emptyPod := v1.Pod{}
				emptyPod.Name = ""
				return emptyPod
			})
		defer patch.Reset()
		convey.Convey("case patch pod success", func() {
			patch1 := gomonkey.ApplyFunc(kube.PatchPodLabel,
				func(podName string, podNamespace string, labels map[string]string) (*v1.Pod, error) {
					return nil, nil
				})
			defer patch1.Reset()
			_, err := labelPodFault(info.JobId, []string{"1", "2"}, map[string]string{"1": "rank1PodUid"}, "")
			convey.So(err, convey.ShouldEqual, nil)
		})
		convey.Convey("case patch pod fail", func() {
			patch1 := gomonkey.ApplyFunc(kube.PatchPodLabel,
				func(podName string, podNamespace string, labels map[string]string) (*v1.Pod, error) {
					return nil, errors.New("fake patch error")
				})
			defer patch1.Reset()
			_, err := labelPodFault(info.JobId, []string{"1", "2"}, map[string]string{"1": "rank1PodUid"}, "")
			convey.So(err, convey.ShouldEqual, nil)
		})
		convey.Convey("case labeled map is nil", func() {
			patch1 := gomonkey.ApplyFuncReturn(kube.PatchPodLabel, nil, nil)
			defer patch1.Reset()
			_, err := labelPodFault(info.JobId, []string{"1", "2"}, nil, "")
			convey.So(err, convey.ShouldEqual, nil)
		})
	})
}

func caseDeviceNumPerNodeZero() {
	info := fakeJobInfo()
	patch := gomonkey.ApplyFunc(pod.GetPodDeviceNumByJobId,
		func(jobId string) int {
			return 0
		})
	defer patch.Reset()
	_, err := LabelFaultPod(info.JobId, []string{"8"}, nil, "")
	convey.ShouldNotBeNil(err)
}

func caseRankStrIllegal() {
	info := fakeJobInfo()
	patch := gomonkey.ApplyFunc(pod.GetPodDeviceNumByJobId,
		func(jobId string) int {
			return deviceNumPerNode
		})
	defer patch.Reset()
	_, err := LabelFaultPod(info.JobId, []string{errorRank}, nil, "")
	convey.ShouldNotBeNil(err)
}

func TestNewEventId(t *testing.T) {
	convey.Convey("Test NewEventId", t, func() {
		uuid := NewEventId(randomLen)
		convey.So(len(uuid), convey.ShouldEqual, uuidLen16)
		uuid = NewEventId(randomLen + randomLen + randomLen)
		convey.So(len(uuid), convey.ShouldEqual, uuidLen32)
	})
}

func TestRemoveSliceDuplicateFaults(t *testing.T) {
	convey.Convey("Test RemoveSliceDuplicateFaults", t, func() {
		convey.Convey("case have same rank", func() {
			faults := []*pb.FaultRank{
				&pb.FaultRank{RankId: "1", FaultType: "1"},
				&pb.FaultRank{RankId: "1", FaultType: "1"},
			}
			newFaults := RemoveSliceDuplicateFaults(faults)
			convey.So(len(newFaults), convey.ShouldEqual, 1)
		})
		convey.Convey("case not have same fault", func() {
			faults := []*pb.FaultRank{
				&pb.FaultRank{RankId: "0", FaultType: "1"},
				&pb.FaultRank{RankId: "1", FaultType: "0"},
				&pb.FaultRank{RankId: "1", FaultType: "1"},
			}
			newFaults := RemoveSliceDuplicateFaults(faults)
			convey.So(len(newFaults), convey.ShouldEqual, len(faults)-1)
		})
	})
}

func TestRetryWriteResetCM(t *testing.T) {
	convey.Convey("Test RetryWriteResetCM", t, func() {
		info := fakeJobInfo()
		convey.Convey("case write success", func() {
			patch1 := gomonkey.ApplyFuncReturn(WriteResetInfoToCM, &v1.ConfigMap{}, nil)
			defer patch1.Reset()
			_, err := RetryWriteResetCM(info.JobName, info.Namespace, []string{"8"}, constant.ClearOperation)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("case write fail", func() {
			patch1 := gomonkey.ApplyFuncReturn(WriteResetInfoToCM, &v1.ConfigMap{}, errors.New("fake error"))
			defer patch1.Reset()
			_, err := RetryWriteResetCM(info.JobName, info.Namespace, []string{"8"}, constant.ClearOperation)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

type successSender struct {
}

func (s *successSender) Send(signal *pb.ProcessManageSignal) error {
	return nil
}

type failSender struct {
}

func (s *failSender) Send(signal *pb.ProcessManageSignal) error {
	return errors.New("fake error")
}

func TestSendRetry(t *testing.T) {
	convey.Convey("Test SendRetry", t, func() {
		convey.Convey("case send success", func() {
			err := SendRetry(&successSender{}, nil, constant.RetryTime)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("case send fail", func() {
			err := SendRetry(&failSender{}, nil, constant.RetryTime)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

func TestStrategySupported(t *testing.T) {
	convey.Convey("Test StrategySupported", t, func() {
		supportRetry := StrategySupported(constant.ProcessRetryStrategyName)
		supportRecover := StrategySupported(constant.ProcessRecoverStrategyName)
		supportDump := StrategySupported(constant.ProcessDumpStrategyName)
		supportExit := StrategySupported(constant.ProcessExitStrategyName)
		supportOthers := StrategySupported("")
		convey.So(supportRetry, convey.ShouldBeTrue)
		convey.So(supportRecover, convey.ShouldBeTrue)
		convey.So(supportDump, convey.ShouldBeTrue)
		convey.So(supportExit, convey.ShouldBeTrue)
		convey.So(supportOthers, convey.ShouldBeFalse)
	})
}

func TestString2Faults(t *testing.T) {
	convey.Convey("Test String2Faults", t, func() {
		convey.Convey("case un format str", func() {
			faults := String2Faults("")
			convey.So(len(faults), convey.ShouldEqual, 0)
			faults = String2Faults(",,,")
			convey.So(len(faults), convey.ShouldEqual, 0)
			faults = String2Faults(" ,1:1,, ")
			convey.So(len(faults), convey.ShouldEqual, 1)
			faults = String2Faults(" ,1:1,2:1:1, ")
			convey.So(len(faults), convey.ShouldEqual, 1)
		})
		convey.Convey("case format str", func() {
			faults := []*pb.FaultRank{
				&pb.FaultRank{RankId: "0", FaultType: "1"},
				&pb.FaultRank{RankId: "1", FaultType: "1"},
				&pb.FaultRank{RankId: "2", FaultType: "1"},
			}
			convertFaults := String2Faults("0:1,1:1,2:1")
			convey.So(len(faults), convey.ShouldEqual, len(convertFaults))
			convey.So(faults, convey.ShouldResemble, convertFaults)
		})
	})
}

func fakeResetInfo() TaskResetInfo {
	return TaskResetInfo{
		RankList: []*TaskDevInfo{
			&TaskDevInfo{
				RankId:       0,
				DevFaultInfo: DevFaultInfo{},
			},
			&TaskDevInfo{
				RankId:       1,
				DevFaultInfo: DevFaultInfo{},
			},
		},
		UpdateTime:    0,
		RetryTime:     0,
		FaultFlushing: false,
		GracefulExit:  0,
	}
}

func fakeRecoverResetInfo() TaskResetInfo {
	return TaskResetInfo{
		RankList: []*TaskDevInfo{
			&TaskDevInfo{
				RankId: 0,
				DevFaultInfo: DevFaultInfo{
					Policy: constant.HotResetPolicy,
				},
			},
		},
		UpdateTime:    0,
		RetryTime:     0,
		FaultFlushing: false,
		GracefulExit:  0,
	}
}

func fakeResetBody(withRecover bool) string {
	var info TaskResetInfo
	if withRecover {
		info = fakeRecoverResetInfo()
	} else {
		info = fakeResetInfo()
	}
	bs, err := json.Marshal(info)
	if err != nil {
		return ""
	}
	return string(bs)
}

func caseHaveRestKey() {
	info := fakeJobInfo()
	patch0 := gomonkey.ApplyFuncReturn(kube.GetConfigMap, &v1.ConfigMap{
		Data: map[string]string{
			constant.ResetInfoCMDataKey: fakeResetBody(false),
		},
	}, nil)
	defer patch0.Reset()
	patch1 := gomonkey.ApplyFunc(kube.UpdateConfigMap, func(newCm *v1.ConfigMap) (*v1.ConfigMap, error) {
		return newCm, nil
	})
	defer patch1.Reset()
	cm, err := WriteResetInfoToCM(info.JobName, info.Namespace, []string{"8"}, constant.ClearOperation)
	convey.So(err, convey.ShouldBeNil)
	var newReset TaskResetInfo
	err = json.Unmarshal([]byte(cm.Data[constant.ResetInfoCMDataKey]), &newReset)
	convey.So(err, convey.ShouldBeNil)
	convey.So(len(newReset.RankList), convey.ShouldEqual, 0)
	cm, err = WriteResetInfoToCM(info.JobName, info.Namespace, []string{"8"}, constant.RestartAllProcessOperation)
	convey.So(err, convey.ShouldBeNil)
	err = json.Unmarshal([]byte(cm.Data[constant.ResetInfoCMDataKey]), &newReset)
	convey.So(err, convey.ShouldBeNil)
	convey.So(newReset.RetryTime, convey.ShouldEqual, 1)
	cm, err = WriteResetInfoToCM(info.JobName, info.Namespace, []string{"8"}, constant.NotifyFaultListOperation)
	convey.So(err, convey.ShouldBeNil)
	err = json.Unmarshal([]byte(cm.Data[constant.ResetInfoCMDataKey]), &newReset)
	convey.So(err, convey.ShouldBeNil)
	convey.So(len(newReset.RankList), convey.ShouldEqual, 1)
	cm, err = WriteResetInfoToCM(info.JobName, info.Namespace, []string{"8"}, constant.NotifyFaultFlushingOperation)
	convey.So(err, convey.ShouldBeNil)
	err = json.Unmarshal([]byte(cm.Data[constant.ResetInfoCMDataKey]), &newReset)
	convey.So(err, convey.ShouldBeNil)
	convey.So(newReset.FaultFlushing, convey.ShouldBeTrue)
	patch2 := gomonkey.ApplyFuncReturn(json.Unmarshal, errors.New("fake unmarshal error"))
	defer patch2.Reset()
	cm, err = WriteResetInfoToCM(info.JobName, info.Namespace, []string{"8"}, constant.NotifyFaultFlushingOperation)
	convey.ShouldNotBeNil(err)
}

func TestWriteResetInfoToCM(t *testing.T) {
	convey.Convey("Test WriteResetInfoToCM", t, func() {
		info := fakeJobInfo()
		convey.Convey("case get config map error", func() {
			patch := gomonkey.ApplyFuncReturn(kube.GetConfigMap, &v1.ConfigMap{}, errors.New("fake error"))
			defer patch.Reset()
			_, err := WriteResetInfoToCM(info.JobName, info.Namespace, []string{"8"}, constant.ClearOperation)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("case not have reset key", func() {
			patch0 := gomonkey.ApplyFuncReturn(kube.GetConfigMap, &v1.ConfigMap{
				Data: make(map[string]string),
			}, nil)
			defer patch0.Reset()
			_, err := WriteResetInfoToCM(info.JobName, info.Namespace, []string{"8"}, constant.ClearOperation)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("case have reset key", caseHaveRestKey)
		convey.Convey("case setNewTaskInfo return error", setNewTaskInfoError)
		convey.Convey("case json marshal error", jsonMarshalError)
	})
}

func setNewTaskInfoError() {
	info := fakeJobInfo()
	convey.Convey("case policy error", func() {
		patch0 := gomonkey.ApplyFuncReturn(kube.GetConfigMap, &v1.ConfigMap{
			Data: map[string]string{
				constant.ResetInfoCMDataKey: fakeResetBody(true),
			},
		}, nil)
		defer patch0.Reset()
		patch1 := gomonkey.ApplyFunc(kube.UpdateConfigMap, func(newCm *v1.ConfigMap) (*v1.ConfigMap, error) {
			return newCm, nil
		})
		defer patch1.Reset()
		_, err := WriteResetInfoToCM(info.JobName, info.Namespace, []string{""}, constant.ClearOperation)
		convey.ShouldNotBeNil(err)
	})
	convey.Convey("case rank error", func() {
		patch0 := gomonkey.ApplyFuncReturn(kube.GetConfigMap, &v1.ConfigMap{
			Data: map[string]string{
				constant.ResetInfoCMDataKey: fakeResetBody(false),
			},
		}, nil)
		defer patch0.Reset()
		patch1 := gomonkey.ApplyFunc(kube.UpdateConfigMap, func(newCm *v1.ConfigMap) (*v1.ConfigMap, error) {
			return newCm, nil
		})
		defer patch1.Reset()
		_, err := WriteResetInfoToCM(info.JobName, info.Namespace, []string{errorRank}, constant.NotifyFaultListOperation)
		convey.ShouldNotBeNil(err)
	})
}

func jsonMarshalError() {
	info := fakeJobInfo()
	patch0 := gomonkey.ApplyFuncReturn(kube.GetConfigMap, &v1.ConfigMap{
		Data: map[string]string{
			constant.ResetInfoCMDataKey: fakeResetBody(false),
		},
	}, nil).ApplyFunc(kube.UpdateConfigMap, func(newCm *v1.ConfigMap) (*v1.ConfigMap, error) {
		return newCm, nil
	}).ApplyFuncReturn(util.MakeDataHash, "").
		ApplyFuncReturn(json.Marshal, nil, errors.New("fake marshal error"))
	defer patch0.Reset()
	_, err := WriteResetInfoToCM(info.JobName, info.Namespace, []string{""}, constant.ClearOperation)
	convey.ShouldNotBeNil(err)
}

func TestSortRecoverStrategies(t *testing.T) {
	convey.Convey("Test SortRecoverStrategies", t, func() {
		convey.Convey("01-reverse strategy slice", func() {
			strategies := []string{constant.ProcessExitStrategyName, constant.ProcessDumpStrategyName,
				constant.ProcessRecoverStrategyName, constant.ProcessRetryStrategyName}
			SortRecoverStrategies(strategies)
			targets := []string{constant.ProcessRetryStrategyName, constant.ProcessRecoverStrategyName,
				constant.ProcessDumpStrategyName, constant.ProcessExitStrategyName}
			convey.So(strategies, convey.ShouldResemble, targets)
		})
		convey.Convey("02-sort slice when exist unknown strategy", func() {
			unknownStrategy := "unknownStrategy"
			strategies := []string{constant.ProcessExitStrategyName, constant.ProcessDumpStrategyName, unknownStrategy}
			SortRecoverStrategies(strategies)
			targets := []string{constant.ProcessDumpStrategyName, constant.ProcessExitStrategyName, unknownStrategy}
			convey.So(strategies, convey.ShouldResemble, targets)
		})
	})
}
