/*
Copyright(C) 2023. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*
Package controllers is using for reconcile Job.
*/

package v1

import (
	"errors"
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	mindxdlv1 "ascend-operator/pkg/api/v1"
	_ "ascend-operator/pkg/testtool"
)

func TestIsUnconditionalRetryJob(t *testing.T) {
	convey.Convey("isUnconditionalRetryJob", t, func() {
		rc := &ASJobReconciler{}
		job := &mindxdlv1.Job{}
		convey.Convey("01-gang scheduling config is false, should return false", func() {
			res := rc.isUnconditionalRetryJob(job)
			convey.So(res, convey.ShouldEqual, false)
			convey.ShouldEqual(res, false)
		})
		convey.Convey("gang scheduling config is true", func() {
			rc.Config.EnableGangScheduling = true
			convey.Convey("02-job label is nil, should return false", func() {
				res := rc.isUnconditionalRetryJob(job)
				convey.So(res, convey.ShouldEqual, false)
			})
			job.Labels = make(map[string]string)
			convey.Convey("03-job label contains unconditionalRetryLabel Key, but value is invalid, "+
				"should return false", func() {
				job.Labels[unconditionalRetryLabelKey] = "xxx"
				res := rc.isUnconditionalRetryJob(job)
				convey.So(res, convey.ShouldEqual, false)
			})
			convey.Convey("04-job label contains unconditionalRetryLabel Key, but value is not num, "+
				"should return false", func() {
				job.Labels[unconditionalRetryLabelKey] = "xxx"
				res := rc.isUnconditionalRetryJob(job)
				convey.So(res, convey.ShouldEqual, false)
			})
			convey.Convey("05-job label contains unconditionalRetryLabel Key, but value is 0, "+
				"should return false", func() {
				job.Labels[unconditionalRetryLabelKey] = "0"
				res := rc.isUnconditionalRetryJob(job)
				convey.So(res, convey.ShouldEqual, false)
			})
			convey.Convey("06-job label contains unconditionalRetryLabel Key, and value is 1, "+
				"should return true", func() {
				job.Labels[unconditionalRetryLabelKey] = "1"
				res := rc.isUnconditionalRetryJob(job)
				convey.So(res, convey.ShouldEqual, true)
			})
		})
	})
}

const invalidRetryTimes = -1

func TestGetJobRemainRetryTimesAboutFaultConfigMap(t *testing.T) {
	convey.Convey("getJobRemainRetryTimes", t, func() {
		rc := &ASJobReconciler{}
		job := &mindxdlv1.Job{}
		convey.Convey("01-get fault configmap failed, should return err", func() {
			patch := gomonkey.ApplyPrivateMethod(new(ASJobReconciler), "getConfigmapFromApiserver",
				func(_ *ASJobReconciler, _ string, _ string) (*corev1.ConfigMap, error) {
					return nil, errors.New("not found")
				})
			defer patch.Reset()
			res, err := rc.getJobRemainRetryTimes(job)
			convey.So(res, convey.ShouldEqual, invalidRetryTimes)
			convey.So(err, convey.ShouldResemble, errors.New("not found"))

		})
		convey.Convey("02-no cache of remain-retry-times in fault configmap, should return err", func() {
			cm := &corev1.ConfigMap{
				Data: nil,
			}
			patch := gomonkey.ApplyPrivateMethod(new(ASJobReconciler), "getConfigmapFromApiserver",
				func(_ *ASJobReconciler, _ string, _ string) (*corev1.ConfigMap, error) {
					return cm, nil
				})
			defer patch.Reset()
			res, err := rc.getJobRemainRetryTimes(job)
			convey.So(res, convey.ShouldEqual, invalidRetryTimes)
			convey.So(err, convey.ShouldResemble, fmt.Errorf("volcaco reschedule confimap has no remain-retry-times key"))
		})
	})
}

func TestGetJobRemainRetryTimesAboutCache(t *testing.T) {
	convey.Convey("getJobRemainRetryTimes", t, func() {
		rc := &ASJobReconciler{}
		job := &mindxdlv1.Job{}
		convey.Convey("01-data in cache is invalid, should return err", func() {
			cm := &corev1.ConfigMap{
				Data: map[string]string{cmJobRemainRetryTimes: "3"},
			}
			patch := gomonkey.ApplyPrivateMethod(new(ASJobReconciler), "getConfigmapFromApiserver",
				func(_ *ASJobReconciler, _ string, _ string) (*corev1.ConfigMap, error) {
					return cm, nil
				})
			defer patch.Reset()
			res, err := rc.getJobRemainRetryTimes(job)
			convey.So(res, convey.ShouldEqual, invalidRetryTimes)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("02-data in cache has no current job info, should return err", func() {
			cm := &corev1.ConfigMap{
				Data: map[string]string{cmJobRemainRetryTimes: ""},
			}
			fakeRemainTimes := &RemainRetryTimes{
				UUID:  "111",
				Times: 3,
			}
			rTimes := map[types.UID]*RemainRetryTimes{"111": fakeRemainTimes}
			job.Namespace = "fake-namespace"
			job.Name = "fake-name"
			job.UID = "fake-uid"
			patch1 := gomonkey.ApplyPrivateMethod(new(ASJobReconciler), "getConfigmapFromApiserver",
				func(_ *ASJobReconciler, _ string, _ string) (*corev1.ConfigMap, error) {
					return cm, nil
				})
			defer patch1.Reset()
			patch2 := gomonkey.ApplyFunc(unmarshalRemainRetryTimes, func(_ string) (map[types.UID]*RemainRetryTimes,
				error) {
				return rTimes, nil
			})
			defer patch2.Reset()
			res, err := rc.getJobRemainRetryTimes(job)
			convey.So(res, convey.ShouldEqual, invalidRetryTimes)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}
