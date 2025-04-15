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
	"k8s.io/apimachinery/pkg/api/resource"

	"ascend-common/api"
	mindxdlv1 "ascend-operator/pkg/api/v1"
	_ "ascend-operator/pkg/testtool"
)

const (
	fakeHostNetwork     = "false"
	fakeTaskID          = "123456"
	fakeJobIdLabelValue = "jobIdLabelValue"
	fakeAppLabelValue   = "appLabelValue"
	msRoleIndex         = 6

	ascend910            = "huawei.com/Ascend910"
	ascend910vir2c       = "huawei.com/Ascend910-2c"
	chipsPerNode         = "16"
	ascend910DownwardAPI = "metadata.annotations['huawei.com/Ascend910']"
)

// TestIsVirtualResourceReq test isVirtualResourceReq
func TestIsVirtualResourceReq(t *testing.T) {
	convey.Convey("test isVirtualResourceReq", t, func() {
		rc := &ASJobReconciler{}
		convey.Convey("01-pod requests is nil, will return false", func() {
			res := rc.isVirtualResourceReq(nil)
			convey.So(res, convey.ShouldBeFalse)
		})
		convey.Convey("02-pod requests virtual resource, will return true", func() {
			fakeRequests := &corev1.ResourceList{ascend910vir2c: resource.Quantity{}}
			res := rc.isVirtualResourceReq(fakeRequests)
			convey.So(res, convey.ShouldBeTrue)
		})
		convey.Convey("03-pod requests normal resource, will return false", func() {
			fakeRequests := &corev1.ResourceList{ascend910: resource.Quantity{}}
			res := rc.isVirtualResourceReq(fakeRequests)
			convey.So(res, convey.ShouldBeFalse)
		})
	})
}

func fakeExpectEnvsForSetInferEnv01() []corev1.EnvVar {
	return []corev1.EnvVar{
		{Name: taskIDEnvKey, Value: fakeJobIdLabelValue},
		{Name: appTypeEnvKey, Value: fakeAppLabelValue},
		{Name: mindxServerIPEnv, Value: ""},
	}
}

// TestSetInferEnv test setInferEnv
func TestSetInferEnv(t *testing.T) {
	convey.Convey("setInferEnv", t, func() {
		ei := newCommonPodInfo()
		podTemp := &corev1.PodTemplateSpec{Spec: corev1.PodSpec{
			Containers: make([]corev1.Container, 1),
		}}
		rc := &ASJobReconciler{}
		convey.Convey("01-pod has no default container, will do nothing", func() {
			rc.setInferEnv(ei, podTemp)
			convey.ShouldBeNil(podTemp.Spec.Containers[0].Env)
		})
		podTemp.Spec.Containers[0] = corev1.Container{
			Name: mindxdlv1.DefaultContainerName,
		}
		convey.Convey("02-rType is worker, scheduler host equal ei.ip", func() {
			ei.job.SetLabels(map[string]string{
				mindxdlv1.JodIdLabelKey: fakeJobIdLabelValue,
				mindxdlv1.AppLabelKey:   fakeAppLabelValue,
			})
			rc.setInferEnv(ei, podTemp)
			convey.So(podTemp.Spec.Containers[0].Env, convey.ShouldResemble, fakeExpectEnvsForSetInferEnv01())
		})
	})
}

func fakeExpectEnvsForSetCommonEnv02() []corev1.EnvVar {
	return []corev1.EnvVar{
		{Name: taskIDEnvKey, Value: fakeTaskID},
		{Name: mindxServerIPEnv, Value: ""},
		{Name: hostNetwork, Value: fakeHostNetwork},
		{Name: hcclSuperPodLogicId, Value: "0"}}
}

func fakeExpectEnvsForSetCommonEnv03() []corev1.EnvVar {
	return []corev1.EnvVar{
		fakeRefEnv(ascendVisibleDevicesEnv, ascend910DownwardAPI),
		{Name: taskIDEnvKey, Value: fakeTaskID},
		{Name: mindxServerIPEnv, Value: ""},
		{Name: hostNetwork, Value: fakeHostNetwork},
		{Name: hcclSuperPodLogicId, Value: "0"}}
}

// TestSetCommonEnv test setCommonEnv
func TestSetCommonEnv(t *testing.T) {
	convey.Convey("test setCommonEnv", t, func() {
		ei := newCommonPodInfo()
		podTemp := &corev1.PodTemplateSpec{Spec: corev1.PodSpec{
			Containers: make([]corev1.Container, 1),
		}}
		rc := &ASJobReconciler{}
		convey.Convey("01-pod has no default container, will do nothing", func() {
			rc.setCommonEnv(ei, podTemp)
			convey.So(podTemp.Spec.Containers[0].Env, convey.ShouldBeNil)
		})
		podTemp.Spec.Containers[0] = corev1.Container{
			Name: mindxdlv1.DefaultContainerName,
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					ascend910vir2c: resource.MustParse(chipsPerNode),
				},
			},
		}
		convey.Convey("02-pod request virtual resource, "+
			"no need set ascendVisibleDevicesEnv", func() {
			expectEnvs := fakeExpectEnvsForSetCommonEnv02()
			rc.setCommonEnv(ei, podTemp)
			convey.So(podTemp.Spec.Containers[0].Env, convey.ShouldResemble, expectEnvs)
		})
		convey.Convey("03-pod request normal resource, "+
			"set ascendVisibleDevicesEnv", func() {
			podTemp.Spec.Containers[0].Resources.Requests = corev1.ResourceList{
				ascend910: resource.MustParse(chipsPerNode),
			}
			expectEnvs := fakeExpectEnvsForSetCommonEnv03()
			rc.setCommonEnv(ei, podTemp)
			convey.So(podTemp.Spec.Containers[0].Env, convey.ShouldResemble, expectEnvs)
		})
	})
}

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
			{Name: api.MsLocalWorkerEnv, Value: strconv.Itoa(ei.ctReq)},
			{Name: api.MsWorkerNumEnv, Value: strconv.Itoa(ei.ctReq * ei.npuReplicas)},
			{Name: msNodeRank, Value: strconv.Itoa(ei.rank)},
			{Name: msSchedPort, Value: ei.port},
			{Name: msServerNum, Value: "0"},
			{Name: msRole, Value: msRoleMap[ei.rtype]},
			{Name: npuPod, Value: "false"}}
		convey.Convey("01-pod has no default container, will do nothing", func() {
			rc.setMindSporeEnv(ei, podTemp)
			convey.So(podTemp.Spec.Containers[0].Env, convey.ShouldBeNil)
		})
		podTemp.Spec.Containers[0] = corev1.Container{
			Name: mindxdlv1.DefaultContainerName,
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					ascend910: resource.MustParse(chipsPerNode),
				},
			},
		}
		convey.Convey("02-rType is worker, scheduler host equal ei.ip", func() {
			rc.setMindSporeEnv(ei, podTemp)
			convey.So(podTemp.Spec.Containers[0].Env, convey.ShouldResemble, expectEnvs)
		})
		convey.Convey("03-rType is Scheduler, scheduler host equal ei.ip", func() {
			ei.rtype = mindxdlv1.MindSporeReplicaTypeScheduler
			expectEnvs[0] = fakeRefEnv(msSchedHost, statusPodIPDownwardAPI)
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
			{Name: api.PtLocalWorldSizeEnv, Value: strconv.Itoa(ei.ctReq)},
			{Name: api.PtWorldSizeEnv, Value: strconv.Itoa(ei.ctReq * ei.npuReplicas)},
			{Name: api.PtLocalRankEnv, Value: localRankStr(ei.ctReq)},
			{Name: ptMasterAddr, Value: ei.ip},
			{Name: ptMasterPort, Value: ei.port},
			{Name: ptRank, Value: strconv.Itoa(ei.rank)}}
		convey.Convey("01-pod has no default container, will do nothing", func() {
			rc.setPytorchEnv(ei, podTemp)
			convey.So(podTemp.Spec.Containers[0].Env, convey.ShouldBeNil)
			convey.ShouldBeNil(podTemp.Spec.Containers[0].Env)
		})
		podTemp.Spec.Containers[0] = corev1.Container{
			Name: mindxdlv1.DefaultContainerName,
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					ascend910: resource.MustParse(chipsPerNode),
				},
			},
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
			{Name: api.TfLocalWorkerEnv, Value: strconv.Itoa(ei.ctReq)},
			{Name: api.TfWorkerSizeEnv, Value: strconv.Itoa(ei.ctReq * ei.npuReplicas)},
			{Name: tfChiefPort, Value: ei.port},
			{Name: tfRank, Value: "1"},
			{Name: tfChiefDevice, Value: "0"},
			{Name: tfWorkerIP, ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{
				FieldPath: statusPodIPDownwardAPI}}}}
		convey.Convey("01-pod has no default container, will do nothing", func() {
			rc.setTensorflowEnv(ei, podTemp)
			convey.So(podTemp.Spec.Containers[0].Env, convey.ShouldBeNil)
		})
		podTemp.Spec.Containers[0] = corev1.Container{
			Name: mindxdlv1.DefaultContainerName,
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					ascend910: resource.MustParse(chipsPerNode),
				},
			},
		}
		convey.Convey("02-rType is worker, scheduler host equal ei.ip", func() {
			rc.setTensorflowEnv(ei, podTemp)
			convey.So(podTemp.Spec.Containers[0].Env, convey.ShouldResemble, expectEnvs)
		})
		convey.Convey("03-rType is master, scheduler host equal ei.ip", func() {
			ei.rtype = mindxdlv1.TensorflowReplicaTypeChief
			expectEnvs[0] = fakeRefEnv(tfChiefIP, statusPodIPDownwardAPI)
			rc.setTensorflowEnv(ei, podTemp)
			convey.So(podTemp.Spec.Containers[0].Env, convey.ShouldResemble, expectEnvs)
		})
	})
}

func fakeRefEnv(name string, downwardAPI string) corev1.EnvVar {
	return corev1.EnvVar{
		Name: name,
		ValueFrom: &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{
				FieldPath: downwardAPI,
			},
		},
	}
}
