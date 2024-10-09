/*
Copyright(C) 2023. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*
Package controllers is using for reconcile AscendJob.
*/

package v1

import (
	"strconv"
	"testing"

	commonv1 "github.com/kubeflow/common/pkg/apis/common/v1"
	"github.com/smartystreets/goconvey/convey"
	corev1 "k8s.io/api/core/v1"

	mindxdlv1 "ascend-operator/pkg/api/v1"
	_ "ascend-operator/pkg/testtool"
)

func TestSetMindSporeEnv(t *testing.T) {
	convey.Convey("setMindSporeEnv", t, func() {
		ei := newCommonPodInfo()
		podTemp := &corev1.PodTemplateSpec{Spec: corev1.PodSpec{
			Containers: make([]corev1.Container, 1),
		}}
		rc := &ASJobReconciler{}
		msRoleMap := map[commonv1.ReplicaType]string{
			mindxdlv1.MindSporeReplicaTypeScheduler: msSchedulerRole,
			mindxdlv1.ReplicaTypeWorker:             msWorkerRole,
		}
		expectEnvs := []corev1.EnvVar{
			{Name: msSchedHost, Value: ei.ip},
			{Name: msNodeRank, Value: strconv.Itoa(ei.rank)},
			{Name: msSchedPort, Value: ei.port},
			{Name: msServerNum, Value: "0"},
			{Name: msLocalWorker, Value: strconv.Itoa(ei.ctReq)},
			{Name: msWorkerNum, Value: strconv.Itoa(ei.ctReq * ei.npuReplicas)},
			{Name: msRole, Value: msRoleMap[ei.rtype]}}
		convey.Convey("01-pod has no default container, will do nothing", func() {
			rc.setMindSporeEnv(ei, podTemp)
			convey.ShouldBeNil(podTemp.Spec.Containers[0].Env)
		})
		podTemp.Spec.Containers[0] = corev1.Container{
			Name: mindxdlv1.DefaultContainerName,
		}
		convey.Convey("01-rType is worker, scheduler host equal ei.ip", func() {
			rc.setMindSporeEnv(ei, podTemp)
			convey.ShouldEqual(podTemp.Spec.Containers[0].Env, expectEnvs)
		})
		convey.Convey("01-rType is Scheduler, scheduler host equal ei.ip", func() {
			ei.rtype = mindxdlv1.MindSporeReplicaTypeScheduler
			expectEnvs[0] = corev1.EnvVar{
				Name: msSchedHost,
				ValueFrom: &corev1.EnvVarSource{
					FieldRef: &corev1.ObjectFieldSelector{
						FieldPath: statusPodIPDownwardAPI,
					},
				},
			}
			expectEnvs[6] = corev1.EnvVar{
				Name:  msRole,
				Value: msRoleMap[ei.rtype],
			}
			rc.setMindSporeEnv(ei, podTemp)
			convey.ShouldEqual(podTemp.Spec.Containers[0].Env, expectEnvs)
		})
	})
}

func TestSetPytorchEnv(t *testing.T) {
	convey.Convey("SetPytorchEnv", t, func() {
		ei := newCommonPodInfo()
		podTemp := &corev1.PodTemplateSpec{Spec: corev1.PodSpec{
			Containers: make([]corev1.Container, 1),
		}}
		rc := &ASJobReconciler{}
		expectEnvs := []corev1.EnvVar{
			{
				Name:  ptMasterAddr,
				Value: ei.ip,
			},
			{
				Name:  ptMasterPort,
				Value: ei.port,
			},
			{
				Name:  ptLocalRank,
				Value: localRankStr(ei.ctReq),
			},
			{
				Name:  ptRank,
				Value: strconv.Itoa(ei.rank),
			},
			{
				Name:  ptLocalWorldSize,
				Value: strconv.Itoa(ei.ctReq),
			},
			{
				Name:  ptWorldSize,
				Value: strconv.Itoa(ei.ctReq * ei.npuReplicas),
			},
		}
		convey.Convey("01-pod has no default container, will do nothing", func() {
			rc.setMindSporeEnv(ei, podTemp)
			convey.ShouldBeNil(podTemp.Spec.Containers[0].Env)
		})
		podTemp.Spec.Containers[0] = corev1.Container{
			Name: mindxdlv1.DefaultContainerName,
		}
		convey.Convey("02-rType is worker, scheduler host equal ei.ip", func() {
			rc.setMindSporeEnv(ei, podTemp)
			convey.ShouldEqual(podTemp.Spec.Containers[0].Env, expectEnvs)
		})
		convey.Convey("03-rType is master, scheduler host equal ei.ip", func() {
			ei.rtype = mindxdlv1.PytorchReplicaTypeMaster
			rc.setMindSporeEnv(ei, podTemp)
			convey.ShouldEqual(podTemp.Spec.Containers[0].Env, expectEnvs)
		})
	})
}

func TestSetTensorflowEnv(t *testing.T) {
	convey.Convey("setTensorflowEnv", t, func() {
		ei := newCommonPodInfo()
		podTemp := &corev1.PodTemplateSpec{Spec: corev1.PodSpec{
			Containers: make([]corev1.Container, 1),
		}}
		rc := &ASJobReconciler{}
		expectEnvs := []corev1.EnvVar{{Name: tfChiefIP, Value: ei.ip},
			{Name: tfChiefDevice, Value: "0"}, {Name: tfChiefPort, Value: ei.port},
			{Name: tfLocalWorker, Value: strconv.Itoa(ei.ctReq)},
			{Name: tfWorkerSize, Value: strconv.Itoa(ei.ctReq * ei.npuReplicas)},
			{Name: tfWorkerIP, ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{
				FieldPath: statusPodIPDownwardAPI}}}}
		convey.Convey("01-pod has no default container, will do nothing", func() {
			rc.setMindSporeEnv(ei, podTemp)
			convey.ShouldBeNil(podTemp.Spec.Containers[0].Env)
		})
		podTemp.Spec.Containers[0] = corev1.Container{
			Name: mindxdlv1.DefaultContainerName,
		}
		convey.Convey("02-rType is worker, scheduler host equal ei.ip", func() {
			rc.setMindSporeEnv(ei, podTemp)
			convey.ShouldEqual(podTemp.Spec.Containers[0].Env, expectEnvs)
		})
		convey.Convey("03-rType is master, scheduler host equal ei.ip", func() {
			ei.rtype = mindxdlv1.TensorflowReplicaTypeChief
			expectEnvs[0] = corev1.EnvVar{
				Name: tfChiefIP,
				ValueFrom: &corev1.EnvVarSource{
					FieldRef: &corev1.ObjectFieldSelector{
						FieldPath: statusPodIPDownwardAPI,
					},
				},
			}
			rc.setMindSporeEnv(ei, podTemp)
			convey.ShouldEqual(podTemp.Spec.Containers[0].Env, expectEnvs)
		})
	})
}
