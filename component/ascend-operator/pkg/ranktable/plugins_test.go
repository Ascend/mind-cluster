/*
Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package ranktable

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"volcano.sh/apis/pkg/apis/batch/v1alpha1"

	"ascend-common/api"
	mindxdlv1 "ascend-operator/pkg/api/v1"
)

func TestShouldGenerateRankTableOfDeploymentPlugin(t *testing.T) {
	convey.Convey("TestDeploymentPlugin_ShouldGenerateRankTable", t, func() {
		plugin := &deploymentPlugin{}

		convey.Convey("01-with AtlasTaskLabel should return true", func() {
			deploy := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "test-deploy",
					Labels: map[string]string{api.AtlasTaskLabel: ""},
				},
			}
			convey.So(plugin.ShouldGenerateRankTable(deploy), convey.ShouldBeTrue)
		})

		convey.Convey("02-with ranktable volume mount should return true", func() {
			deploy := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{Name: "test-deploy"},
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Volumes: []corev1.Volume{
								{Name: "ranktable"},
							},
						},
					},
				},
			}
			convey.So(plugin.ShouldGenerateRankTable(deploy), convey.ShouldBeTrue)
		})

		convey.Convey("03-without label and volume mount should return false", func() {
			deploy := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{Name: "test-deploy"},
			}
			convey.So(plugin.ShouldGenerateRankTable(deploy), convey.ShouldBeFalse)
		})

		convey.Convey("04-non-deployment object should return false", func() {
			sts := &appsv1.StatefulSet{}
			convey.So(plugin.ShouldGenerateRankTable(sts), convey.ShouldBeFalse)
		})
	})
}

func TestShouldGenerateRankTableOfStatefulSetPlugin(t *testing.T) {
	convey.Convey("TestStatefulSetPlugin_ShouldGenerateRankTable", t, func() {
		plugin := &statefulSetPlugin{}

		convey.Convey("01-with AtlasTaskLabel should return true", func() {
			sts := &appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "test-sts",
					Labels: map[string]string{api.AtlasTaskLabel: ""},
				},
			}
			convey.So(plugin.ShouldGenerateRankTable(sts), convey.ShouldBeTrue)
		})

		convey.Convey("02-with ranktable volume mount should return true", func() {
			sts := &appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{Name: "test-sts"},
				Spec: appsv1.StatefulSetSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Volumes: []corev1.Volume{
								{Name: "ranktable"},
							},
						},
					},
				},
			}
			convey.So(plugin.ShouldGenerateRankTable(sts), convey.ShouldBeTrue)
		})

		convey.Convey("03-without label and volume mount should return false", func() {
			sts := &appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{Name: "test-sts"},
			}
			convey.So(plugin.ShouldGenerateRankTable(sts), convey.ShouldBeFalse)
		})
	})
}

func TestShouldGenerateRankTableVcJobPlugin(t *testing.T) {
	convey.Convey("TestVcJobPlugin_ShouldGenerateRankTable", t, func() {
		plugin := &vcJobPlugin{}

		convey.Convey("01-with AtlasTaskLabel should return true", func() {
			vcjob := &v1alpha1.Job{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "test-vcjob",
					Labels: map[string]string{api.AtlasTaskLabel: ""},
				},
			}
			convey.So(plugin.ShouldGenerateRankTable(vcjob), convey.ShouldBeTrue)
		})

		convey.Convey("02-with ranktable volume mount in task should return true", func() {
			vcjob := &v1alpha1.Job{
				ObjectMeta: metav1.ObjectMeta{Name: "test-vcjob"},
				Spec: v1alpha1.JobSpec{
					Tasks: []v1alpha1.TaskSpec{
						{
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Volumes: []corev1.Volume{
										{Name: "ranktable"},
									},
								},
							},
						},
					},
				},
			}
			convey.So(plugin.ShouldGenerateRankTable(vcjob), convey.ShouldBeTrue)
		})

		convey.Convey("03-without label and volume mount should return false", func() {
			vcjob := &v1alpha1.Job{
				ObjectMeta: metav1.ObjectMeta{Name: "test-vcjob"},
			}
			convey.So(plugin.ShouldGenerateRankTable(vcjob), convey.ShouldBeFalse)
		})
	})
}

func TestFindPluginForObject(t *testing.T) {
	convey.Convey("TestFindPluginForObject", t, func() {
		convey.Convey("01-find deployment plugin", func() {
			deploy := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{Name: "test-deploy"},
			}
			plugin := FindPluginForObject(deploy)
			convey.So(plugin, convey.ShouldNotBeNil)
			convey.So(plugin.Name(), convey.ShouldEqual, mindxdlv1.DeploymentPlugin)
		})

		convey.Convey("02-find statefulset plugin", func() {
			sts := &appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{Name: "test-sts"},
			}
			plugin := FindPluginForObject(sts)
			convey.So(plugin, convey.ShouldNotBeNil)
			convey.So(plugin.Name(), convey.ShouldEqual, mindxdlv1.StatefulsetPlugin)
		})

		convey.Convey("03-find vcjob plugin", func() {
			vcjob := &v1alpha1.Job{
				ObjectMeta: metav1.ObjectMeta{Name: "test-vcjob"},
			}
			plugin := FindPluginForObject(vcjob)
			convey.So(plugin, convey.ShouldNotBeNil)
			convey.So(plugin.Name(), convey.ShouldEqual, mindxdlv1.VcJobPlugin)
		})

		convey.Convey("04-unsupported object should return nil", func() {
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: "test-pod"},
			}
			plugin := FindPluginForObject(pod)
			convey.So(plugin, convey.ShouldBeNil)
		})
	})
}
