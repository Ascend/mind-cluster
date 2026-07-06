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

const (
	fakePath      = "test-path"
	emptyPath     = ""
	testNamespace = "default"
	testJobName   = "test-job"
	fakeSpecName  = "fake-spec"
	otherVolume   = "fake-volume"
)

var errFake = errors.New("fake error")

func TestHasRankTableVolume(t *testing.T) {
	convey.Convey("TestReadRankTableDir", t, func() {
		job := &mindxdlv1.AscendJob{}
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
		request := map[v1.ResourceName]resource.Quantity{"huawei.com/Ascend910": resource.
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
		convey.Convey("04-pod with npu request and with PodNPUDeviceAnno should return true", func() {
			pod.Annotations = map[string]string{api.PodNPUDeviceAnno: "fake-device"}
			res := PodHasAllocated(pod)
			convey.So(res, convey.ShouldEqual, true)
		})
	})
}

func newRankTableJob(volumeName string) *mindxdlv1.AscendJob {
	job := &mindxdlv1.AscendJob{
		ObjectMeta: metav1.ObjectMeta{Name: testJobName, Namespace: testNamespace},
	}
	spec := &commonv1.ReplicaSpec{}
	source := v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: fakePath}}
	spec.Template.Spec.Volumes = []v1.Volume{newVolume(volumeName, source)}
	job.Spec.ReplicaSpecs = map[commonv1.ReplicaType]*commonv1.ReplicaSpec{fakeSpecName: spec}
	return job
}

func noPatch() *gomonkey.Patches {
	return gomonkey.NewPatches()
}

func patchValidChecker() *gomonkey.Patches {
	return gomonkey.ApplyFuncReturn(utils.PathStringChecker, fakePath, nil)
}

func patchCheckerErr() *gomonkey.Patches {
	return gomonkey.ApplyFuncReturn(utils.PathStringChecker, emptyPath, errFake)
}

func patchSoftlinkErr() *gomonkey.Patches {
	return patchValidChecker().ApplyFuncReturn(utils.IsExist, true).
		ApplyFuncReturn(utils.IsSoftlink, false, errFake)
}

func patchIsSoftlink() *gomonkey.Patches {
	return patchValidChecker().ApplyFuncReturn(utils.IsExist, true).
		ApplyFuncReturn(utils.IsSoftlink, true, nil)
}

func patchNotSoftlink() *gomonkey.Patches {
	return patchValidChecker().ApplyFuncReturn(utils.IsExist, true).
		ApplyFuncReturn(utils.IsSoftlink, false, nil)
}

func patchMkdirErr() *gomonkey.Patches {
	return patchValidChecker().ApplyFuncReturn(utils.IsExist, false).
		ApplyFuncReturn(os.MkdirAll, errFake)
}

func patchMkdirOK() *gomonkey.Patches {
	return patchValidChecker().ApplyFuncReturn(utils.IsExist, false).
		ApplyFuncReturn(os.MkdirAll, nil)
}

type genRankTableDirCase struct {
	name    string
	job     *mindxdlv1.AscendJob
	prepare func() *gomonkey.Patches
	want    string
}

func genRankTableDirCases() []genRankTableDirCase {
	return []genRankTableDirCase{
		{name: "should return empty path when job is nil",
			job: nil, prepare: noPatch, want: emptyPath},
		{name: "should return empty path when job has no ranktable volume",
			job: newRankTableJob(otherVolume), prepare: noPatch, want: emptyPath},
		{name: "should return empty path when path checker returns error",
			job: newRankTableJob(rankTableName), prepare: patchCheckerErr, want: emptyPath},
		{name: "should return empty path when softlink check returns error",
			job: newRankTableJob(rankTableName), prepare: patchSoftlinkErr, want: emptyPath},
		{name: "should return empty path when ranktable dir is softlink",
			job: newRankTableJob(rankTableName), prepare: patchIsSoftlink, want: emptyPath},
		{name: "should return checked path when dir exists and is not softlink",
			job: newRankTableJob(rankTableName), prepare: patchNotSoftlink, want: fakePath},
		{name: "should return empty path when make dir fails",
			job: newRankTableJob(rankTableName), prepare: patchMkdirErr, want: emptyPath},
		{name: "should return checked path when make dir succeeds",
			job: newRankTableJob(rankTableName), prepare: patchMkdirOK, want: fakePath},
	}
}

func TestGenRankTableDir(t *testing.T) {
	for _, tc := range genRankTableDirCases() {
		convey.Convey(tc.name, t, func() {
			patches := tc.prepare()
			defer patches.Reset()
			got := GenRankTableDir(tc.job)
			convey.So(got, convey.ShouldEqual, tc.want)
		})
	}
}
