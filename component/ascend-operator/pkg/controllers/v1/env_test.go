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

	"github.com/agiledragon/gomonkey/v2"
	commonv1 "github.com/kubeflow/common/pkg/apis/common/v1"
	"github.com/smartystreets/goconvey/convey"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"

	"ascend-common/api"
	"ascend-common/api/ascend-operator/apis/batch/v1"
	mindxdlv1 "ascend-operator/pkg/api/v1"
)

const (
	fakeHostNetwork     = "false"
	fakeTaskID          = "123456"
	fakeJobIdLabelValue = "jobIdLabelValue"
	fakeAppLabelValue   = "appLabelValue"
	indexTwo            = 2
	indexSix            = 6

	ascend910            = "huawei.com/Ascend910"
	ascend910vir2c       = "huawei.com/Ascend910-2c"
	chipsPerNode         = "16"
	ascend910DownwardAPI = "metadata.annotations['huawei.com/Ascend910']"
	testEnvKey1          = "TEST_ENV_KEY_1"
	testEnvKey2          = "TEST_ENV_KEY_2"
	testEnvValue1        = "test_value_1"
	testEnvValue2        = "test_value_2"
	testEnvValue3        = "test_value_3"
	commaSeparator       = ","
	msRecoverPrefix      = `'{`
	msRecoverSuffix      = `}'`
	strategy1            = "strategy1"
	strategy2            = "strategy2"
	strategy3            = "strategy3"
	containerIndexZero   = 0
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
		{Name: mindxServerDomainEnv, Value: mindxDefaultServerDomain},
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
			Name: api.DefaultContainerName,
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
		{Name: mindxServerDomainEnv, Value: mindxDefaultServerDomain},
		{Name: hostNetwork, Value: fakeHostNetwork},
		{Name: hcclSuperPodLogicId, Value: "0"},
		{Name: replicaType, Value: string(v1.ReplicaTypeWorker)}}
}

func fakeExpectEnvsForSetCommonEnv03() []corev1.EnvVar {
	return []corev1.EnvVar{
		fakeRefEnv(api.AscendVisibleDevicesEnv, ascend910DownwardAPI),
		{Name: taskIDEnvKey, Value: fakeTaskID},
		{Name: mindxServerIPEnv, Value: ""},
		{Name: mindxServerDomainEnv, Value: mindxDefaultServerDomain},
		{Name: hostNetwork, Value: fakeHostNetwork},
		{Name: hcclSuperPodLogicId, Value: "0"},
		{Name: replicaType, Value: string(v1.ReplicaTypeWorker)},
	}
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
			Name: api.DefaultContainerName,
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
		convey.Convey("04-pod request soft share device resource, "+
			"no need set ascendVisibleDevicesEnv", func() {
			expectEnvs := fakeExpectEnvsForSetCommonEnv02()
			ei.job.SetLabels(map[string]string{
				api.SchedulerSoftShareDevAicoreQuotaKey: "1",
				api.SchedulerSoftShareDevHbmQuotaKey:    "1",
				api.SchedulerSoftShareDevPolicyKey:      "1"})
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
			Name: api.DefaultContainerName,
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
		convey.Convey("03-rType is worker and job is soft share device, pod info ctReq equal 1", func() {
			ei.isSoftShareDevJob = true
			expectEnvs[1] = corev1.EnvVar{Name: api.MsLocalWorkerEnv, Value: "1"}
			expectEnvs[indexTwo] = corev1.EnvVar{Name: api.MsWorkerNumEnv, Value: strconv.Itoa(ei.npuReplicas)}
			rc.setMindSporeEnv(ei, podTemp)
			convey.So(podTemp.Spec.Containers[0].Env, convey.ShouldResemble, expectEnvs)
		})
		convey.Convey("04-rType is Scheduler, scheduler host equal ei.ip", func() {
			ei.rtype = mindxdlv1.MindSporeReplicaTypeScheduler
			expectEnvs[0] = fakeRefEnv(msSchedHost, statusPodIPDownwardAPI)
			expectEnvs[indexSix] = corev1.EnvVar{Name: msRole, Value: msRoleMap[ei.rtype]}
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
			Name: api.DefaultContainerName,
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
		convey.Convey("03-rType is worker and job is soft share device,  pod info ctReq equal 1", func() {
			ei.isSoftShareDevJob = true
			expectEnvs[0] = corev1.EnvVar{Name: api.PtLocalWorldSizeEnv, Value: "1"}
			expectEnvs[1] = corev1.EnvVar{Name: api.PtWorldSizeEnv, Value: strconv.Itoa(ei.npuReplicas)}
			expectEnvs[indexTwo] = corev1.EnvVar{Name: api.PtLocalRankEnv, Value: localRankStr(1)}
			rc.setPytorchEnv(ei, podTemp)
			convey.So(podTemp.Spec.Containers[0].Env, convey.ShouldResemble, expectEnvs)
		})
		convey.Convey("04-rType is master, scheduler host equal ei.ip", func() {
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
			Name: api.DefaultContainerName,
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
		convey.Convey("03-rType is worker and job is soft share device,  pod info ctReq equal 1", func() {
			ei.isSoftShareDevJob = true
			expectEnvs[1] = corev1.EnvVar{Name: api.TfLocalWorkerEnv, Value: "1"}
			expectEnvs[indexTwo] = corev1.EnvVar{Name: api.TfWorkerSizeEnv, Value: strconv.Itoa(ei.npuReplicas)}
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

func TestAddProcessRecoverEnv(t *testing.T) {
	convey.Convey("when job has empty recover strategy annotation", t, func() {
		pi := &podInfo{job: &mindxdlv1.AscendJob{ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{api.RecoverStrategyKey: ""}}}}
		pod := &corev1.PodTemplateSpec{Spec: corev1.PodSpec{Containers: []corev1.Container{
			{Name: api.DefaultContainerName, Env: []corev1.EnvVar{}}}}}
		addProcessRecoverEnv(pi, pod, 0, api.MindSporeFramework)
		convey.So(len(pod.Spec.Containers[0].Env), convey.ShouldEqual, 0)
	})

	convey.Convey("when job has recover strategy annotation with single strategy - MindSpore", t, func() {
		pi := &podInfo{job: &mindxdlv1.AscendJob{ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{api.RecoverStrategyKey: api.RecoverStrategy}}}}
		pod := &corev1.PodTemplateSpec{Spec: corev1.PodSpec{Containers: []corev1.Container{
			{Name: api.DefaultContainerName, Env: []corev1.EnvVar{}}}}}
		addProcessRecoverEnv(pi, pod, 0, api.MindSporeFramework)
		expectedEnv := map[string]string{
			api.ProcessRecoverEnv: api.EnableFunc, api.ElasticRecoverEnv: api.EnableFlag, api.EnableMS: api.EnableFlag,
			api.MsRecoverEnv: `'{` + api.MsArfStrategy + `}'`, api.MsCloseWatchDogKey: api.MsCloseWatchDogValue,
		}
		checkEnvVars(t, pod.Spec.Containers[0].Env, expectedEnv)
	})

	convey.Convey("when job has recover strategy annotation with single strategy - PyTorch", t, func() {
		pi := &podInfo{job: &mindxdlv1.AscendJob{ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{api.RecoverStrategyKey: api.RecoverStrategy}}}}
		pod := &corev1.PodTemplateSpec{Spec: corev1.PodSpec{Containers: []corev1.Container{
			{Name: api.DefaultContainerName, Env: []corev1.EnvVar{}}}}}
		addProcessRecoverEnv(pi, pod, 0, api.PytorchFramework)
		expectedEnv := map[string]string{
			api.ProcessRecoverEnv: api.EnableFunc, api.ElasticRecoverEnv: api.EnableFlag,
			api.HighAvailableEnv: api.RecoverStrategy, api.PtCloseWatchDogKey: api.PtCloseWatchDogValue}
		checkEnvVars(t, pod.Spec.Containers[0].Env, expectedEnv)
	})
	convey.Convey("when job has recover strategy annotation with multiple strategies - MindSpore", t, func() {
		pi := &podInfo{job: &mindxdlv1.AscendJob{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{
			api.RecoverStrategyKey: api.RecoverStrategy + "," + api.RetryStrategy}}}}
		pod := &corev1.PodTemplateSpec{Spec: corev1.PodSpec{Containers: []corev1.Container{
			{Name: api.DefaultContainerName, Env: []corev1.EnvVar{}}}}}
		addProcessRecoverEnv(pi, pod, 0, api.MindSporeFramework)
		expectedEnv := map[string]string{
			api.ProcessRecoverEnv: api.EnableFunc, api.ElasticRecoverEnv: api.EnableFlag,
			api.MsRecoverEnv: `'{` + api.MsArfStrategy + "," + api.MsUceStrategy + "," + api.MsHcceStrategy + `}'`,
			api.EnableMS:     api.EnableFlag, api.MsCloseWatchDogKey: api.MsCloseWatchDogValue}
		checkEnvVars(t, pod.Spec.Containers[0].Env, expectedEnv)
	})
}

func TestAddProcessRecoverEnv2(t *testing.T) {
	convey.Convey("when job has exit strategy annotation with single strategy - PyTorch", t, func() {
		pi := &podInfo{job: &mindxdlv1.AscendJob{ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{api.RecoverStrategyKey: api.ExitStrategy}}}}
		pod := &corev1.PodTemplateSpec{Spec: corev1.PodSpec{Containers: []corev1.Container{
			{Name: api.DefaultContainerName, Env: []corev1.EnvVar{}}}}}
		addProcessRecoverEnv(pi, pod, 0, api.PytorchFramework)
		expectedEnv := map[string]string{
			api.ProcessRecoverEnv: api.EnableFunc, api.ElasticRecoverEnv: api.EnableFlag,
			api.HighAvailableEnv: api.RecoverStrategy, api.PtCloseWatchDogKey: api.PtCloseWatchDogValue}
		checkEnvVars(t, pod.Spec.Containers[0].Env, expectedEnv)
	})
}

func checkEnvVars(t *testing.T, envVars []corev1.EnvVar, expectedEnv map[string]string) {
	actualEnv := make(map[string]string)
	for _, envVar := range envVars {
		actualEnv[envVar.Name] = envVar.Value
	}
	if len(actualEnv) != len(expectedEnv) {
		t.Errorf("Expected env vars: %v, but got: %v", expectedEnv, actualEnv)
	}
}

func TestAddElasticTrainingEnv(t *testing.T) {
	type args struct {
		env       map[string]string
		trainEnv  sets.String
		framework string
	}
	tests := []struct {
		name string
		args args
		want sets.String
	}{
		{
			name: "pytorch add success",
			args: args{
				env:       make(map[string]string),
				trainEnv:  make(sets.String),
				framework: api.PytorchFramework,
			},
			want: sets.String{api.ElasticTraining: sets.Empty{}},
		},
		{
			name: "no pytorch add failed",
			args: args{
				env:       make(map[string]string),
				trainEnv:  make(sets.String),
				framework: api.MindSporeFramework,
			},
			want: sets.String{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addElasticTrainingEnv(tt.args.env, tt.args.trainEnv, tt.args.framework)
			if tt.args.trainEnv.Len() != tt.want.Len() {
				t.Errorf("get %v, want %v", tt.args.trainEnv.List(), api.ElasticTraining)
			}
		})
	}
}

// TestAddMSPodScheduleEnv test addMSPodScheduleEnv
func TestAddMSPodScheduleEnv(t *testing.T) {
	convey.Convey("addMSPodScheduleEnv", t, func() {
		pi := newCommonPodInfo()
		podTemp := &corev1.PodTemplateSpec{Spec: corev1.PodSpec{
			Containers: make([]corev1.Container, 1),
		}}
		convey.Convey("01-pod schedule strategy is enabled, env should be added", func() {
			pi.job.SetLabels(map[string]string{api.PodScheduleLabel: api.EnableFunc})
			podTemp.Spec.Containers[0] = corev1.Container{
				Name: api.DefaultContainerName,
				Env:  make([]corev1.EnvVar, 0),
			}
			addMSPodScheduleEnv(pi, podTemp, 0)
			convey.So(len(podTemp.Spec.Containers[0].Env), convey.ShouldEqual, 1)
			convey.So(podTemp.Spec.Containers[0].Env[0].Name, convey.ShouldEqual, api.MsRecoverEnv)
			convey.So(podTemp.Spec.Containers[0].Env[0].Value, convey.ShouldEqual, `'{`+api.MsRscStrategy+`}'`)
		})
		convey.Convey("02-pod schedule strategy is disabled, env should not be added", func() {
			pi.job.SetLabels(map[string]string{api.PodScheduleLabel: "off"})
			podTemp.Spec.Containers[0] = corev1.Container{
				Name: api.DefaultContainerName,
				Env:  make([]corev1.EnvVar, 0),
			}
			addMSPodScheduleEnv(pi, podTemp, 0)
			convey.So(len(podTemp.Spec.Containers[0].Env), convey.ShouldEqual, 0)
		})
	})
}

func TestAddSubHealthyEnv(t *testing.T) {
	pi := &podInfo{job: &mindxdlv1.AscendJob{ObjectMeta: metav1.ObjectMeta{UID: "test-uid"}}}
	containerIndex := 0
	const num3 = 3
	const num4 = 4
	cases := buildTestAddSubhealthyEnvCases(num3, num4)
	for _, tc := range cases {
		convey.Convey("When strategy is "+tc.name, t, func() {
			podTemplate := &corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: api.DefaultContainerName, Env: []corev1.EnvVar{}}}},
			}
			patch := gomonkey.ApplyFunc(getSubHealthyStrategy, func(job *mindxdlv1.AscendJob) string {
				return tc.strategy
			})
			defer patch.Reset()

			addSubHealthyEnv(pi, podTemplate, containerIndex, tc.framework)
			convey.So(len(podTemplate.Spec.Containers[containerIndex].Env), convey.ShouldEqual, tc.expectedEnvLen)
			envMap := make(map[string]string)
			for _, env := range podTemplate.Spec.Containers[containerIndex].Env {
				envMap[env.Name] = env.Value
			}
			for k, v := range tc.expectedEnvs {
				convey.So(envMap[k], convey.ShouldEqual, v)
			}
		})
	}
}

func buildTestAddSubhealthyEnvCases(num3 int, num4 int) []struct {
	name           string
	strategy       string
	framework      string
	expectedEnvLen int
	expectedEnvs   map[string]string
} {
	cases := []struct {
		name           string
		strategy       string
		framework      string
		expectedEnvLen int
		expectedEnvs   map[string]string
	}{
		{name: "No strategy",
			strategy:       "",
			framework:      api.PytorchFramework,
			expectedEnvLen: 0,
			expectedEnvs:   map[string]string{}},
		{name: "SubHealthyHotSwitch strategy with PytorchFramework",
			strategy:       api.SubHealthyHotSwitch,
			framework:      api.PytorchFramework,
			expectedEnvLen: num3,
			expectedEnvs: map[string]string{
				api.ProcessRecoverEnv: api.EnableFunc,
				api.ElasticRecoverEnv: api.EnableFlag,
				api.HighAvailableEnv:  api.RecoverStrategy,
			}},
		{name: "SubHealthyHotSwitch strategy with MindSporeFramework",
			strategy:       api.SubHealthyHotSwitch,
			framework:      api.MindSporeFramework,
			expectedEnvLen: num4,
			expectedEnvs: map[string]string{
				api.ProcessRecoverEnv: api.EnableFunc,
				api.ElasticRecoverEnv: api.EnableFlag,
				api.MsRecoverEnv:      `'{` + api.MsArfStrategy + `}'`,
				api.EnableMS:          api.EnableFlag,
			}},
		{name: "SubHealthyHotSwitch strategy with unsupported framework",
			strategy:       api.SubHealthyHotSwitch,
			framework:      "UnsupportedFramework",
			expectedEnvLen: 0,
			expectedEnvs:   map[string]string{}},
	}
	return cases
}

func TestExtractMsRecoverContent(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "should return empty string when input is empty",
			input:    "",
			expected: "",
		},
		{
			name:     "should extract content when input has correct format",
			input:    `'{strategy1,strategy2}'`,
			expected: "strategy1,strategy2",
		},
		{
			name:     "should extract content when input has spaces",
			input:    `'{ strategy1 , strategy2 }'`,
			expected: "strategy1 , strategy2",
		},
		{
			name:     "should return trimmed content when input has prefix and suffix",
			input:    `'{content}'`,
			expected: "content",
		},
		{
			name:     "should remove suffix when input has no prefix",
			input:    "content}'",
			expected: "content",
		},
		{
			name:     "should remove prefix when input has no suffix",
			input:    `'{content`,
			expected: "content",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := extractMsRecoverContent(tc.input)
			convey.Convey(tc.name, t, func() {
				convey.So(result, convey.ShouldEqual, tc.expected)
			})
		})
	}
}

func TestMergeEnvValue(t *testing.T) {
	testCases := buildTestMergeEnvValueCases()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := mergeEnvValue(tc.envKey, tc.oldValue, tc.newValue)
			convey.Convey(tc.name, t, func() {
				convey.So(result, convey.ShouldEqual, tc.expected)
			})
		})
	}
}

type TestEnvValueCase struct {
	name           string
	envKey         string
	oldValue       string
	newValue       string
	expected       string
	initialEnvs    []corev1.EnvVar
	envValue       string
	expectedEnvs   []corev1.EnvVar
	expectedLength int
}

func buildTestMergeEnvValueCases() []TestEnvValueCase {
	highAvailMerged := strategy1 + commaSeparator + strategy2 + commaSeparator + strategy3
	msRecoverMerged := msRecoverPrefix + strategy1 + commaSeparator + strategy2 + commaSeparator +
		strategy3 + msRecoverSuffix
	return []TestEnvValueCase{
		{name: "should merge and deduplicate when envKey is HighAvailableEnv",
			envKey:   api.HighAvailableEnv,
			oldValue: strategy1 + commaSeparator + strategy2,
			newValue: strategy2 + commaSeparator + strategy3,
			expected: highAvailMerged},
		{name: "should return new value when oldValue is empty for HighAvailableEnv",
			envKey:   api.HighAvailableEnv,
			oldValue: "",
			newValue: strategy1 + commaSeparator + strategy2,
			expected: strategy1 + commaSeparator + strategy2},
		{name: "should return old value when newValue is empty for HighAvailableEnv",
			envKey:   api.HighAvailableEnv,
			oldValue: strategy1 + commaSeparator + strategy2,
			newValue: "",
			expected: strategy1 + commaSeparator + strategy2},
		{name: "should merge and deduplicate when envKey is MsRecoverEnv",
			envKey:   api.MsRecoverEnv,
			oldValue: msRecoverPrefix + strategy1 + commaSeparator + strategy2 + msRecoverSuffix,
			newValue: msRecoverPrefix + strategy2 + commaSeparator + strategy3 + msRecoverSuffix,
			expected: msRecoverMerged},
		{name: "should return new value when oldValue is empty for MsRecoverEnv",
			envKey:   api.MsRecoverEnv,
			oldValue: "",
			newValue: msRecoverPrefix + strategy1 + msRecoverSuffix,
			expected: msRecoverPrefix + strategy1 + msRecoverSuffix},
		{name: "should return new value when envKey is not special",
			envKey:   testEnvKey1,
			oldValue: testEnvValue1,
			newValue: testEnvValue2,
			expected: testEnvValue2},
		{name: "should handle spaces in HighAvailableEnv values",
			envKey:   api.HighAvailableEnv,
			oldValue: " " + strategy1 + " , " + strategy2 + " ",
			newValue: strategy2 + " , " + strategy3,
			expected: strategy1 + commaSeparator + strategy2 + commaSeparator + strategy3},
	}
}

func TestAddEnvValueWithDedup(t *testing.T) {
	testCases := buildTestAddEnvValueWithDedupCases()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pod := createPodTemplateWithEnvs(tc.initialEnvs)
			addEnvValueWithDedup(pod, tc.envKey, tc.envValue, containerIndexZero)
			convey.Convey(tc.name, t, func() {
				verifyEnvVars(t, pod, tc.expectedEnvs, tc.expectedLength)
			})
		})
	}
}

func buildTestAddEnvValueWithDedupCases() []TestEnvValueCase {
	highAvailMergedValue := strategy1 + commaSeparator + strategy2 + commaSeparator + strategy3
	msRecoverMergedValue := msRecoverPrefix + strategy1 + commaSeparator + strategy2 + msRecoverSuffix
	return []TestEnvValueCase{
		{name: "should append new env when env does not exist",
			initialEnvs:    []corev1.EnvVar{},
			envKey:         testEnvKey1,
			envValue:       testEnvValue1,
			expectedEnvs:   []corev1.EnvVar{{Name: testEnvKey1, Value: testEnvValue1}},
			expectedLength: 1},
		{name: "should overwrite existing env when envKey is not special",
			initialEnvs:    []corev1.EnvVar{{Name: testEnvKey1, Value: testEnvValue1}},
			envKey:         testEnvKey1,
			envValue:       testEnvValue2,
			expectedEnvs:   []corev1.EnvVar{{Name: testEnvKey1, Value: testEnvValue2}},
			expectedLength: 1},
		{name: "should merge values when envKey is HighAvailableEnv and env exists",
			initialEnvs:    []corev1.EnvVar{{Name: api.HighAvailableEnv, Value: strategy1 + commaSeparator + strategy2}},
			envKey:         api.HighAvailableEnv,
			envValue:       strategy2 + commaSeparator + strategy3,
			expectedEnvs:   []corev1.EnvVar{{Name: api.HighAvailableEnv, Value: highAvailMergedValue}},
			expectedLength: 1},
		{name: "should merge values when envKey is MsRecoverEnv and env exists",
			initialEnvs: []corev1.EnvVar{
				{Name: api.MsRecoverEnv, Value: msRecoverPrefix + strategy1 + msRecoverSuffix},
			},
			envKey:         api.MsRecoverEnv,
			envValue:       msRecoverPrefix + strategy2 + msRecoverSuffix,
			expectedEnvs:   []corev1.EnvVar{{Name: api.MsRecoverEnv, Value: msRecoverMergedValue}},
			expectedLength: 1},
		{name: "should append new env when envKey exists but different key",
			initialEnvs: []corev1.EnvVar{{Name: testEnvKey1, Value: testEnvValue1}},
			envKey:      testEnvKey2,
			envValue:    testEnvValue2,
			expectedEnvs: []corev1.EnvVar{
				{Name: testEnvKey1, Value: testEnvValue1},
				{Name: testEnvKey2, Value: testEnvValue2},
			},
			expectedLength: 2},
		{name: "should handle multiple envs and update correct one",
			initialEnvs: []corev1.EnvVar{
				{Name: testEnvKey1, Value: testEnvValue1},
				{Name: testEnvKey2, Value: testEnvValue2}},
			envKey:   testEnvKey1,
			envValue: testEnvValue3,
			expectedEnvs: []corev1.EnvVar{
				{Name: testEnvKey1, Value: testEnvValue3},
				{Name: testEnvKey2, Value: testEnvValue2}},
			expectedLength: 2},
	}
}

func createPodTemplateWithEnvs(initialEnvs []corev1.EnvVar) *corev1.PodTemplateSpec {
	return &corev1.PodTemplateSpec{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Env: initialEnvs},
			},
		},
	}
}

func verifyEnvVars(t *testing.T, pod *corev1.PodTemplateSpec, expectedEnvs []corev1.EnvVar, expectedLength int) {
	actualEnvs := pod.Spec.Containers[containerIndexZero].Env
	convey.So(len(actualEnvs), convey.ShouldEqual, expectedLength)
	for i, expectedEnv := range expectedEnvs {
		convey.So(actualEnvs[i].Name, convey.ShouldEqual, expectedEnv.Name)
		convey.So(actualEnvs[i].Value, convey.ShouldEqual, expectedEnv.Value)
	}
}
