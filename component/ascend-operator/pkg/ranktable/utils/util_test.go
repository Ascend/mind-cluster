/*
Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*
Package utils is using for generating ranktable.
*/

package utils

import (
	"errors"
	"os"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	commonv1 "github.com/kubeflow/common/pkg/apis/common/v1"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"ascend-common/api"
	"ascend-common/common-utils/utils"
	mindxdlv1 "ascend-operator/pkg/api/v1"
	_ "ascend-operator/pkg/testtool"
)

const fakePath = "test-path"

func TestHasRankTableVolume(t *testing.T) {
	convey.Convey("TestReadRankTableDir", t, func() {
		job := &mindxdlv1.Job{}
		spec := &commonv1.ReplicaSpec{}
		spec.Template.Spec.Volumes = make([]v1.Volume, 1)
		job.Spec.ReplicaSpecs = map[commonv1.ReplicaType]*commonv1.ReplicaSpec{"fake-spec": spec}
		convey.Convey("01-job without volume named ranktable should return empty string", func() {
			volume := newVolume("fake-volume", v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: fakePath}})
			spec.Template.Spec.Volumes[0] = volume
			res := hasRankTableVolume(job)
			convey.So(res, convey.ShouldBeFalse)
		})
		convey.Convey("02-job with volume named ranktable should return true", func() {
			volume := newVolume(rankTableName, v1.VolumeSource{EmptyDir: &v1.EmptyDirVolumeSource{}})
			spec.Template.Spec.Volumes[0] = volume
			res := hasRankTableVolume(job)
			convey.So(res, convey.ShouldBeTrue)
		})
	})
}

func newVolume(name string, src v1.VolumeSource) v1.Volume {
	return v1.Volume{
		Name:         name,
		VolumeSource: src,
	}
}

func TestPodHasAllocated(t *testing.T) {
	convey.Convey("TestPodHasAllocated", t, func() {
		pod := &v1.Pod{}
		convey.Convey("01-pod which has be delete should return false", func() {
			pod.DeletionTimestamp = &metav1.Time{}
			res := PodHasAllocated(pod)
			convey.So(res, convey.ShouldEqual, false)
		})
		pod.DeletionTimestamp = nil
		container := v1.Container{}
		convey.Convey("02-pod without request  should return true", func() {
			request := v1.ResourceList{}
			container.Resources.Requests = request
			pod.Spec.Containers = []v1.Container{container}
			res := PodHasAllocated(pod)
			convey.So(res, convey.ShouldEqual, true)
		})
		request := map[v1.ResourceName]resource.Quantity{api.HuaweiAscend910: resource.
			MustParse("8")}
		container.Resources.Requests = request
		pod.Spec.Containers = []v1.Container{container}
		convey.Convey("02-pod with npu request and without  PodDeviceKey should return false", func() {
			res := PodHasAllocated(pod)
			convey.So(res, convey.ShouldEqual, false)
		})
		convey.Convey("03-pod with npu request and with  PodDeviceKey should return true", func() {
			pod.Annotations = map[string]string{api.Pod910DeviceAnno: "fake-device"}
			res := PodHasAllocated(pod)
			convey.So(res, convey.ShouldEqual, true)
		})
	})
}

func TestGenRankTableDir(t *testing.T) {
	convey.Convey("TestGenRankTableDir", t, func() {
		job := &mindxdlv1.Job{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-job",
				Namespace: "default",
			},
		}
		spec := &commonv1.ReplicaSpec{}
		volume := newVolume(rankTableName, v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: fakePath}})
		spec.Template.Spec.Volumes = make([]v1.Volume, 1)
		job.Spec.ReplicaSpecs = map[commonv1.ReplicaType]*commonv1.ReplicaSpec{"fake-spec": spec}
		convey.Convey("01-job without valid ranktable volume should return empty path", func() {
			path := GenRankTableDir(job)
			convey.So(path, convey.ShouldEqual, "")
		})
		spec.Template.Spec.Volumes[0] = volume
		convey.Convey("02-invalid ranktable path should return empty path", func() {
			patch := gomonkey.ApplyFunc(utils.PathStringChecker, func(string2 string) (string, error) {
				return "", errors.New("fake error")
			})
			defer patch.Reset()
			path := GenRankTableDir(job)
			convey.So(path, convey.ShouldEqual, "")
		})
		patch := gomonkey.ApplyFunc(utils.PathStringChecker, func(string2 string) (string, error) {
			return fakePath, nil
		})
		defer patch.Reset()
		convey.Convey("03-make dir failed should return empty path", func() {
			patch1 := gomonkey.ApplyFunc(os.MkdirAll, func(string, os.FileMode) error {
				return errors.New("make dir error")
			})
			defer patch1.Reset()
			path := GenRankTableDir(job)
			convey.So(path, convey.ShouldEqual, "")
		})
		convey.Convey("04-make dir success should return path", func() {
			patch1 := gomonkey.ApplyFunc(os.MkdirAll, func(string, os.FileMode) error {
				return nil
			})
			defer patch1.Reset()
			path := GenRankTableDir(job)
			convey.So(path, convey.ShouldEqual, fakePath)
		})
	})
}

func TestGenRankTableDir01(t *testing.T) {
	convey.Convey("TestGenRankTableDir", t, func() {
		job := &mindxdlv1.Job{}
		spec := &commonv1.ReplicaSpec{}
		volume := newVolume(rankTableName, v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: fakePath}})
		spec.Template.Spec.Volumes = []v1.Volume{volume}
		job.Spec.ReplicaSpecs = map[commonv1.ReplicaType]*commonv1.ReplicaSpec{"fake-spec": spec}
		patch := gomonkey.ApplyFunc(utils.PathStringChecker, func(string2 string) (string, error) {
			return fakePath, nil
		})
		defer patch.Reset()
		err := os.MkdirAll(fakePath, defaultDirPerm)
		convey.So(err, convey.ShouldBeNil)
		defer os.RemoveAll(fakePath)
		convey.Convey("05-check soft link path failed should return empty path", func() {
			patch1 := gomonkey.ApplyFunc(utils.IsSoftlink, func(string) (bool, error) {
				return true, errors.New("fake error")
			})
			defer patch1.Reset()
			path := GenRankTableDir(job)
			convey.So(path, convey.ShouldEqual, "")
		})
		convey.Convey("06-soft link path should return empty path", func() {
			patch1 := gomonkey.ApplyFunc(utils.IsSoftlink, func(string) (bool, error) {
				return true, nil
			})
			defer patch1.Reset()
			path := GenRankTableDir(job)
			convey.So(path, convey.ShouldEqual, "")
		})
		convey.Convey("07-not soft link path should return empty path", func() {
			patch1 := gomonkey.ApplyFunc(utils.IsSoftlink, func(string) (bool, error) {
				return false, nil
			})
			defer patch1.Reset()
			path := GenRankTableDir(job)
			convey.So(path, convey.ShouldEqual, fakePath)
		})
	})
}
