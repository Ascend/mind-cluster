/*
Copyright(C) 2023. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*
Package controllers is using for reconcile AscendJob.
*/

package v1

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	commonv1 "github.com/kubeflow/common/pkg/apis/common/v1"
	"github.com/smartystreets/goconvey/convey"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	mindxdlv1 "ascend-operator/pkg/api/v1"
	_ "ascend-operator/pkg/testtool"
)

func TestGetContainerExitCode(t *testing.T) {
	expectCode := 0xbeef
	convey.Convey("getContainerExitCode", t, func() {
		pod := &corev1.Pod{
			Status: corev1.PodStatus{
				ContainerStatuses: make([]corev1.ContainerStatus, 1),
			},
		}
		convey.Convey("pod has no default container, should return 0xbeef", func() {
			code := getContainerExitCode(pod)
			convey.ShouldEqual(code, expectCode)
		})
		convey.Convey("pod's default container state is not terminate, should return 0xbeef", func() {
			pod.Status.ContainerStatuses[0] = corev1.ContainerStatus{
				Name: mindxdlv1.DefaultContainerName,
			}
			code := getContainerExitCode(pod)
			convey.ShouldEqual(code, expectCode)
		})
		convey.Convey("pod's default container state is terminate, should return exit code", func() {
			pod.Status.ContainerStatuses[0] = corev1.ContainerStatus{
				Name: mindxdlv1.DefaultContainerName,
				State: corev1.ContainerState{
					Terminated: &corev1.ContainerStateTerminated{
						ExitCode: 1,
					},
				},
			}
			code := getContainerExitCode(pod)
			convey.ShouldEqual(code, 1)
		})
	})
}

func TestInitializeReplicaStatuses(t *testing.T) {
	convey.Convey("initializeReplicaStatuses", t, func() {
		jobStatus := &commonv1.JobStatus{}
		rtype := mindxdlv1.ReplicaTypeWorker
		convey.Convey("01-jobStatus replica status  is nil, should be init", func() {
			initializeReplicaStatuses(jobStatus, rtype)
			convey.ShouldEqual(jobStatus, &commonv1.JobStatus{
				ReplicaStatuses: map[commonv1.ReplicaType]*commonv1.ReplicaStatus{rtype: {}},
			})
		})
	})
}

func TestUpdateJobReplicaStatuses(t *testing.T) {
	convey.Convey("updateJobReplicaStatuses", t, func() {
		rtype := mindxdlv1.ReplicaTypeWorker
		jobStatus := &commonv1.JobStatus{
			ReplicaStatuses: map[commonv1.ReplicaType]*commonv1.ReplicaStatus{rtype: {}},
		}
		pod := &corev1.Pod{}
		convey.Convey("01-pod status is Running, jobStatus active equal 1", func() {
			pod.Status.Phase = corev1.PodRunning
			updateJobReplicaStatuses(jobStatus, rtype, pod)
			convey.ShouldEqual(jobStatus.ReplicaStatuses[rtype].Active, 1)
		})
		convey.Convey("02-pod status is Succeeded, jobStatus succeed equal 1", func() {
			pod.Status.Phase = corev1.PodSucceeded
			updateJobReplicaStatuses(jobStatus, rtype, pod)
			convey.ShouldEqual(jobStatus.ReplicaStatuses[rtype].Succeeded, 1)
		})
		convey.Convey("03-pod status is Failed, jobStatus failed equal 1", func() {
			pod.Status.Phase = corev1.PodFailed
			updateJobReplicaStatuses(jobStatus, rtype, pod)
			convey.ShouldEqual(jobStatus.ReplicaStatuses[rtype].Failed, 1)
		})
	})
}

func TestContainsChiefOrMasterSpec(t *testing.T) {
	convey.Convey("ContainsChiefOrMasterSpec", t, func() {
		convey.Convey("01-nil replicas should return false", func() {
			res := ContainsChiefOrMasterSpec(nil)
			convey.ShouldEqual(res, false)
		})
		convey.Convey("02-replicas with Master should return true", func() {
			replicas := map[commonv1.ReplicaType]*commonv1.ReplicaSpec{
				mindxdlv1.PytorchReplicaTypeMaster: nil,
			}
			res := ContainsChiefOrMasterSpec(replicas)
			convey.ShouldEqual(res, true)
		})
		convey.Convey("03-replicas with Chief should return true", func() {
			replicas := map[commonv1.ReplicaType]*commonv1.ReplicaSpec{
				mindxdlv1.TensorflowReplicaTypeChief: nil,
			}
			res := ContainsChiefOrMasterSpec(replicas)
			convey.ShouldEqual(res, true)
		})
	})
}

func TestGetContainerResourceReq(t *testing.T) {
	convey.Convey("getContainerResourceReq", t, func() {
		convey.Convey("01-container with no resources should return 0", func() {
			res := getContainerResourceReq(corev1.Container{})
			convey.ShouldEqual(res, 0)
		})
		convey.Convey("02-container with npu resources should return npu num", func() {
			ct := corev1.Container{
				Resources: corev1.ResourceRequirements{
					Limits: map[corev1.ResourceName]resource.Quantity{"huawei.com/Ascend910": resource.
						MustParse("8")},
				},
			}
			expectRequestRes := 8
			res := getContainerResourceReq(ct)
			convey.ShouldEqual(res, expectRequestRes)
		})
	})
}

func TestGetNpuWorkerSpec(t *testing.T) {
	convey.Convey("getNpuWorkerSpec", t, func() {
		job := &mindxdlv1.AscendJob{}
		expectSpec := &commonv1.ReplicaSpec{}
		convey.Convey("01-job with nil replicas should return nil", func() {
			spec := getNpuWorkerSpec(job)
			convey.ShouldBeNil(spec)
		})
		convey.Convey("02-job with Worker replica should return corresponding spec", func() {
			job.Spec.ReplicaSpecs = map[commonv1.ReplicaType]*commonv1.ReplicaSpec{
				mindxdlv1.ReplicaTypeWorker: expectSpec,
			}
			spec := getNpuWorkerSpec(job)
			convey.ShouldEqual(spec, expectSpec)
		})
		convey.Convey("03-job with Master replica should return corresponding spec", func() {
			job.Spec.ReplicaSpecs = map[commonv1.ReplicaType]*commonv1.ReplicaSpec{
				mindxdlv1.PytorchReplicaTypeMaster: expectSpec,
			}
			spec := getNpuWorkerSpec(job)
			convey.ShouldEqual(spec, expectSpec)
		})
	})
}

func TestGetNpuReqPerPod(t *testing.T) {
	convey.Convey("getNpuReqPerPod", t, func() {
		job := &mindxdlv1.AscendJob{}
		convey.Convey("01-job with no npu worker should return 0", func() {
			patch := gomonkey.ApplyFunc(getNpuWorkerSpec, func(_ *mindxdlv1.AscendJob) *commonv1.ReplicaSpec {
				return nil
			})
			defer patch.Reset()
			res := getNpuReqPerPod(job)
			convey.ShouldEqual(res, 0)
		})
		convey.Convey("02-job with npu worker should return corresponding npu num", func() {
			workerSpec := &commonv1.ReplicaSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{Containers: []corev1.Container{
						{Name: mindxdlv1.DefaultContainerName},
					}},
				},
			}
			patch1 := gomonkey.ApplyFunc(getNpuWorkerSpec, func(_ *mindxdlv1.AscendJob) *commonv1.ReplicaSpec {
				return workerSpec
			})
			defer patch1.Reset()

			patch2 := gomonkey.ApplyFunc(getContainerResourceReq, func(_ corev1.Container) int { return 1 })
			defer patch2.Reset()
			res := getNpuReqPerPod(job)
			convey.ShouldEqual(res, 1)
		})
	})
}

func TestLocalRankStr(t *testing.T) {
	rankRequest := 0
	convey.Convey("localRankStr", t, func() {
		convey.Convey("01-when input is 0, should return empty string", func() {
			res := localRankStr(rankRequest)
			convey.ShouldEqual(res, "")
		})
		convey.Convey("02-when input is 4, should return string", func() {
			rankRequest = 4
			res := localRankStr(rankRequest)
			convey.ShouldEqual(res, "0,1,2,3")
		})
	})
}

func TestGetTotalNpuReplicas(t *testing.T) {
	convey.Convey("getTotalNpuReplicas", t, func() {
		job := &mindxdlv1.AscendJob{
			Spec: mindxdlv1.AscendJobSpec{
				ReplicaSpecs: map[commonv1.ReplicaType]*commonv1.ReplicaSpec{},
			},
		}
		replicas := int32(1)
		spec := &commonv1.ReplicaSpec{
			Replicas: &replicas,
		}
		convey.Convey("01-job with no replicas should return 0", func() {
			res := getTotalNpuReplicas(&mindxdlv1.AscendJob{})
			convey.ShouldEqual(res, 0)
		})
		convey.Convey("02-job with only scheduler should return 0", func() {
			job.Spec.ReplicaSpecs[mindxdlv1.MindSporeReplicaTypeScheduler] = spec
			res := getTotalNpuReplicas(job)
			convey.ShouldEqual(res, 0)
		})
		convey.Convey("03-job with worker should return workerSpec replicas", func() {
			job.Spec.ReplicaSpecs[mindxdlv1.ReplicaTypeWorker] = spec
			res := getTotalNpuReplicas(job)
			convey.ShouldEqual(res, 1)
		})
	})
}

func TestGetRestartCondition(t *testing.T) {
	convey.Convey("getRestartCondition", t, func() {
		conditions := make([]commonv1.JobCondition, 0, 2)
		convey.Convey("01-nil conditions will return nil", func() {
			res := getRestartCondition(nil)
			convey.ShouldBeNil(res)
		})
		convey.Convey("02-conditions without restart condition will return nil", func() {
			conditions = append(conditions, commonv1.JobCondition{
				Type: commonv1.JobRunning,
			})
			res := getRestartCondition(conditions)
			convey.ShouldBeNil(res)
		})
		convey.Convey("03-conditions with restart condition will return right result", func() {
			expectCondition := commonv1.JobCondition{Type: commonv1.JobRestarting, Reason: "fake reason",
				Message: "fake message"}
			conditions = append(conditions, expectCondition)
			res := getRestartCondition(conditions)
			convey.ShouldEqual(res, &commonv1.JobCondition{
				Reason:  "fake reason",
				Message: "fake message",
			})
		})
	})
}

func mockRplsWithNPU() map[commonv1.ReplicaType]*commonv1.ReplicaSpec {
	replicas := int32(1)
	quantityMap := map[corev1.ResourceName]resource.Quantity{"huawei.com/Ascend910": resource.MustParse("8")}
	return map[commonv1.ReplicaType]*commonv1.ReplicaSpec{
		mindxdlv1.MindSporeReplicaTypeScheduler: {
			Replicas: &replicas,
			Template: corev1.PodTemplateSpec{Spec: corev1.PodSpec{Containers: []corev1.Container{{
				Name: mindxdlv1.DefaultContainerName,
				Resources: corev1.ResourceRequirements{
					Limits:   quantityMap,
					Requests: quantityMap,
				},
			}}}},
		},
		mindxdlv1.ReplicaTypeWorker: {
			Replicas: &replicas,
			Template: corev1.PodTemplateSpec{Spec: corev1.PodSpec{Containers: []corev1.Container{{
				Name: mindxdlv1.DefaultContainerName,
				Resources: corev1.ResourceRequirements{
					Limits:   quantityMap,
					Requests: quantityMap,
				},
			}}}},
		}}
}

func TestCheckNonWorkerRplMountChips(t *testing.T) {
	convey.Convey("checkNonWorkerRplMountChips", t, func() {
		convey.Convey("01-conditions with non-worker replicaSpec not mount npu condition will return false", func() {
			ji := &jobInfo{rpls: map[commonv1.ReplicaType]*commonv1.ReplicaSpec{
				mindxdlv1.MindSporeFrameworkName: {},
				mindxdlv1.ReplicaTypeWorker:      {},
			}}
			res := checkNonWorkerRplMountChips(ji)
			convey.So(res, convey.ShouldBeFalse)
		})
		convey.Convey("02-conditions with non-worker replicaSpec mount npu condition will return true", func() {
			ji := &jobInfo{rpls: mockRplsWithNPU()}
			res := checkNonWorkerRplMountChips(ji)
			convey.So(res, convey.ShouldBeTrue)
		})
	})
}

func TestGetNonWorkerPodMountChipStatus(t *testing.T) {
	convey.Convey("getNonWorkerPodMountChipStatus", t, func() {
		job := newCommonAscendJob()
		convey.Convey("01-annotation not found target key will return false", func() {
			res := getNonWorkerPodMountChipStatus(job)
			convey.So(res, convey.ShouldBeFalse)
		})
		convey.Convey("02-annotation found target key will return real value", func() {
			job.SetAnnotations(map[string]string{nonWorkerPodMountChipStatus: "true"})
			res := getNonWorkerPodMountChipStatus(job)
			convey.So(res, convey.ShouldBeTrue)
		})
	})
}

func TestCheckNpuPod(t *testing.T) {
	convey.Convey("checkNpuPod", t, func() {
		pi := newCommonPodInfo()
		convey.Convey("01-pod with no npu will return false", func() {
			res := checkNpuPod(pi)
			convey.So(res, convey.ShouldBeFalse)
		})
		convey.Convey("02-pod with npu should return true", func() {
			pi.job.Spec.ReplicaSpecs = mockRplsWithNPU()
			res := checkNpuPod(pi)
			convey.So(res, convey.ShouldBeTrue)
		})
	})
}
