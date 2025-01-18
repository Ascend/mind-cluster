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

const (
	fakeHostNetwork = "false"
	fakeTaskID      = "123456"
	msRoleIndex     = 8
)

// TestSetMindSporeEnv test setMindSporeEnv
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
			{Name: msLocalWorker, Value: strconv.Itoa(ei.ctReq)},
			{Name: msWorkerNum, Value: strconv.Itoa(ei.ctReq * ei.npuReplicas)},
			{Name: taskIDEnvKey, Value: fakeTaskID},
			{Name: mindxServerIPEnv, Value: ""},
			{Name: msNodeRank, Value: strconv.Itoa(ei.rank)},
			{Name: msSchedPort, Value: ei.port},
			{Name: msServerNum, Value: "0"},
			{Name: msRole, Value: msRoleMap[ei.rtype]},
			{Name: hostNetwork, Value: fakeHostNetwork},
			{Name: npuPod, Value: "false"},
		}
		convey.Convey("01-pod has no default container, will do nothing", func() {
			rc.setMindSporeEnv(ei, podTemp)
			convey.So(podTemp.Spec.Containers[0].Env, convey.ShouldBeNil)
		})
		podTemp.Spec.Containers[0] = corev1.Container{
			Name: mindxdlv1.DefaultContainerName,
		}
		convey.Convey("01-rType is worker, scheduler host equal ei.ip", func() {
			rc.setMindSporeEnv(ei, podTemp)
			convey.So(podTemp.Spec.Containers[0].Env, convey.ShouldResemble, expectEnvs)
		})
		convey.Convey("01-rType is Scheduler, scheduler host equal ei.ip", func() {
			ei.rtype = mindxdlv1.MindSporeReplicaTypeScheduler
			expectEnvs[0] = fakeRefEnv(msSchedHost)
			expectEnvs[msRoleIndex] = corev1.EnvVar{
				Name:  msRole,
				Value: msRoleMap[ei.rtype],
			}
			rc.setMindSporeEnv(ei, podTemp)
			convey.So(podTemp.Spec.Containers[0].Env, convey.ShouldResemble, expectEnvs)
		})
	})
}

// TestSetPytorchEnv test setPytorchEnv
func TestSetPytorchEnv(t *testing.T) {
	convey.Convey("SetPytorchEnv", t, func() {
		ei := newCommonPodInfo()
		podTemp := &corev1.PodTemplateSpec{Spec: corev1.PodSpec{
			Containers: make([]corev1.Container, 1),
		}}
		rc := &ASJobReconciler{}
		expectEnvs := []corev1.EnvVar{
			{Name: ptLocalWorldSize, Value: strconv.Itoa(ei.ctReq)},
			{Name: ptWorldSize, Value: strconv.Itoa(ei.ctReq * ei.npuReplicas)},
			{Name: ptLocalRank, Value: localRankStr(ei.ctReq)},
			{Name: ptMasterAddr, Value: ei.ip},
			{Name: ptMasterPort, Value: ei.port},
			{Name: ptRank, Value: strconv.Itoa(ei.rank)},
			{Name: taskIDEnvKey, Value: fakeTaskID},
			{Name: mindxServerIPEnv, Value: ""},
			{Name: hostNetwork, Value: fakeHostNetwork},
		}
		convey.Convey("01-pod has no default container, will do nothing", func() {
			rc.setPytorchEnv(ei, podTemp)
			convey.So(podTemp.Spec.Containers[0].Env, convey.ShouldBeNil)
			convey.ShouldBeNil(podTemp.Spec.Containers[0].Env)
		})
		podTemp.Spec.Containers[0] = corev1.Container{
			Name: mindxdlv1.DefaultContainerName,
		}
		convey.Convey("02-rType is worker, scheduler host equal ei.ip", func() {
			rc.setPytorchEnv(ei, podTemp)
			convey.So(podTemp.Spec.Containers[0].Env, convey.ShouldResemble, expectEnvs)
		})
		convey.Convey("03-rType is master, scheduler host equal ei.ip", func() {
			ei.rtype = mindxdlv1.PytorchReplicaTypeMaster
			rc.setPytorchEnv(ei, podTemp)
			convey.So(podTemp.Spec.Containers[0].Env, convey.ShouldResemble, expectEnvs)
		})
	})
}

// TestSetMindSporeEnv test setMindSporeEnv
func TestSetTensorflowEnv(t *testing.T) {
	convey.Convey("setTensorflowEnv", t, func() {
		ei := newCommonPodInfo()
		podTemp := &corev1.PodTemplateSpec{Spec: corev1.PodSpec{
			Containers: make([]corev1.Container, 1),
		}}
		rc := &ASJobReconciler{}
		expectEnvs := []corev1.EnvVar{
			{Name: tfChiefIP, Value: ei.ip},
			{Name: tfLocalWorker, Value: strconv.Itoa(ei.ctReq)},
			{Name: tfWorkerSize, Value: strconv.Itoa(ei.ctReq * ei.npuReplicas)},
			{Name: tfChiefPort, Value: ei.port},
			{Name: tfRank, Value: "1"},
			{Name: taskIDEnvKey, Value: fakeTaskID},
			{Name: mindxServerIPEnv, Value: ""},
			{Name: tfChiefDevice, Value: "0"},
			{Name: hostNetwork, Value: fakeHostNetwork},
			{Name: tfWorkerIP, ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{
				FieldPath: statusPodIPDownwardAPI}}}}
		convey.Convey("01-pod has no default container, will do nothing", func() {
			rc.setTensorflowEnv(ei, podTemp)
			convey.So(podTemp.Spec.Containers[0].Env, convey.ShouldBeNil)
		})
		podTemp.Spec.Containers[0] = corev1.Container{
			Name: mindxdlv1.DefaultContainerName,
		}
		convey.Convey("02-rType is worker, scheduler host equal ei.ip", func() {
			rc.setTensorflowEnv(ei, podTemp)
			convey.So(podTemp.Spec.Containers[0].Env, convey.ShouldResemble, expectEnvs)
		})
		convey.Convey("03-rType is master, scheduler host equal ei.ip", func() {
			ei.rtype = mindxdlv1.TensorflowReplicaTypeChief
			expectEnvs[0] = fakeRefEnv(tfChiefIP)
			rc.setTensorflowEnv(ei, podTemp)
			convey.So(podTemp.Spec.Containers[0].Env, convey.ShouldResemble, expectEnvs)
		})
	})
}

func fakeRefEnv(name string) corev1.EnvVar {
	return corev1.EnvVar{
		Name: name,
		ValueFrom: &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{
				FieldPath: statusPodIPDownwardAPI,
			},
		},
	}
}
