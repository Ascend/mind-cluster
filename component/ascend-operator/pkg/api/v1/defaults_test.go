/*
Copyright(C) 2023. Huawei Technologies Co.,Ltd. All rights reserved.
*/

package v1

import (
	"testing"

	commonv1 "github.com/kubeflow/common/pkg/apis/common/v1"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"ascend-common/api"
)

func TestInt32(t *testing.T) {
	convey.Convey("TestInt32", t, func() {
		convey.So(*Int32(0), convey.ShouldEqual, 0)
	})
}

func TestSetDefaultPort(t *testing.T) {
	convey.Convey("TestSetDefaultPort", t, func() {
		spec := &v1.PodSpec{}
		spec.Containers = []v1.Container{{
			Name:  api.DefaultContainerName,
			Ports: []v1.ContainerPort{},
		}}
		setDefaultPort(spec)
		convey.So(spec.Containers[0].Ports[0].ContainerPort, convey.ShouldEqual, DefaultPort)
		convey.So(spec.Containers[0].Ports[0].Name, convey.ShouldEqual, api.DefaultPortName)
	})
}

func TestSetDefaultReplicas(t *testing.T) {
	convey.Convey("TestSetDefaultReplicas", t, func() {
		spec := &commonv1.ReplicaSpec{}
		setDefaultReplicas(spec)
		convey.So(spec.Replicas, convey.ShouldResemble, Int32(1))
		convey.So(spec.RestartPolicy, convey.ShouldEqual, DefaultRestartPolicy)
	})
}

func TestSetTypeNamesToCamelCase(t *testing.T) {
	convey.Convey("TestSetTypeNamesToCamelCase", t, func() {
		job := &AscendJob{}
		job.Spec.ReplicaSpecs = map[commonv1.ReplicaType]*commonv1.ReplicaSpec{
			"scheduler": {},
			"chief":     {},
			"master":    {},
			"worker":    {},
		}
		setTypeNamesToCamelCase(job)
		_, ok := job.Spec.ReplicaSpecs[MindSporeReplicaTypeScheduler]
		convey.So(ok, convey.ShouldEqual, true)
		_, ok = job.Spec.ReplicaSpecs[ReplicaTypeWorker]
		convey.So(ok, convey.ShouldEqual, true)
		_, ok = job.Spec.ReplicaSpecs[PytorchReplicaTypeMaster]
		convey.So(ok, convey.ShouldEqual, true)
		_, ok = job.Spec.ReplicaSpecs[TensorflowReplicaTypeChief]
		convey.So(ok, convey.ShouldEqual, true)
	})
}

func TestSetDefaultsAscendJob(t *testing.T) {
	convey.Convey("TestSetDefaultsAscendJob", t, func() {
		job := &AscendJob{
			Spec: AscendJobSpec{
				ReplicaSpecs: map[commonv1.ReplicaType]*commonv1.ReplicaSpec{
					ReplicaTypeWorker: {},
				},
			},
		}
		template := v1.PodTemplateSpec{
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					{
						Ports: []v1.ContainerPort{
							{
								Name: api.DefaultPortName,
							},
						},
					},
				},
			},
		}
		job.Spec.ReplicaSpecs[ReplicaTypeWorker] = &commonv1.ReplicaSpec{Template: template}
		SetDefaultsAscendJob(job)
		convey.So(*job.Spec.RunPolicy.CleanPodPolicy, convey.ShouldEqual, commonv1.CleanPodPolicyNone)
		convey.So(*job.Spec.SuccessPolicy, convey.ShouldEqual, SuccessPolicyDefault)
	})
}

func TestDeepCopyJob(t *testing.T) {
	convey.Convey("TestDeepCopyJob", t, func() {
		successpolicy := SuccessPolicyDefault
		jobList := &AscendJobList{
			Items: []AscendJob{
				{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec: AscendJobSpec{
						SuccessPolicy: &successpolicy,
						ReplicaSpecs: map[commonv1.ReplicaType]*commonv1.ReplicaSpec{
							"Worker": {},
						},
					},
					Status: commonv1.JobStatus{},
				},
			},
		}
		newJobList := jobList.DeepCopy()
		convey.So(jobList, convey.ShouldResemble, newJobList)
	})
}

func TestGetJobFramework(t *testing.T) {
	convey.Convey("TestGetJobFramework", t, func() {
		convey.Convey("01-nil job should return error", func() {
			_, err := GetJobFramework(nil)
			convey.So(err, convey.ShouldNotBeNil)
		})
		job := &AscendJob{}
		convey.Convey("02-nil labels should return error", func() {
			_, err := GetJobFramework(job)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("03-invalid labels should return error", func() {
			job.Labels = map[string]string{FrameworkKey: "invalid"}
			_, err := GetJobFramework(job)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("04-valid labels should return framework name", func() {
			job.Labels = map[string]string{FrameworkKey: MindSporeFrameworkName}
			framework, err := GetJobFramework(job)
			convey.So(err, convey.ShouldBeNil)
			convey.So(framework, convey.ShouldEqual, MindSporeFrameworkName)
		})
	})
}

func TestAddDefaultingFuncs(t *testing.T) {
	convey.Convey("TestAddDefaultingFuncs", t, func() {
		err := addDefaultingFuncs(runtime.NewScheme())
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestSetObjectDefaultsAscendJobList(t *testing.T) {
	convey.Convey("TestSetObjectDefaultsAscendJobList", t, func() {
		jobList := &AscendJobList{
			Items: []AscendJob{{}},
		}
		SetObjectDefaultsAscendJobList(jobList)
		convey.So(*jobList.Items[0].Spec.RunPolicy.CleanPodPolicy, convey.ShouldResemble, commonv1.CleanPodPolicyNone)
	})
}
