// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package statistics test for statistic funcs about fault
package statistics

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	ascendv1 "ascend-common/api/ascend-operator/apis/batch/v1"
)

const (
	name      = "test-name"
	namespace = "test-namespace"
	uid       = "test-uid"
)

func TestSaveJob(t *testing.T) {
	convey.Convey("test SaveJob", t, func() {
		convey.Convey("when job info is nil, cache should remain unchanged", func() {
			oldLen := len(jobManager.jobMap)
			SaveJob(nil)
			newLen := len(jobManager.jobMap)
			convey.So(oldLen, convey.ShouldEqual, newLen)
		})
		convey.Convey("when pod cache less than maxPodNum, cache should save success", func() {
			SaveJob(getDemoJob(uid, name, namespace))
			defer DeleteJob(getDemoJob(uid, name, namespace))
			convey.So(len(jobManager.jobMap), convey.ShouldEqual, 1)
		})
	})
}

func TestDeleteJob(t *testing.T) {
	convey.Convey("test DeleteJob", t, func() {
		convey.Convey("when job info is nil, cache should remain unchanged", func() {
			oldLen := len(jobManager.jobMap)
			DeleteJob(nil)
			newLen := len(jobManager.jobMap)
			convey.So(oldLen, convey.ShouldEqual, newLen)
		})
		convey.Convey("when pod cache less than maxPodNum, cache should delete success", func() {
			SaveJob(getDemoJob(uid, name, namespace))
			DeleteJob(getDemoJob(uid, name, namespace))
			convey.So(len(jobManager.jobMap), convey.ShouldEqual, 0)
		})
	})
}

func TestGetJob(t *testing.T) {
	convey.Convey("test get Pod", t, func() {
		convey.Convey("when pod cache less than maxPodNum, cache should get success", func() {
			SaveJob(getDemoJob(uid, name, namespace))
			jobInfo := GetJob(uid)
			defer DeleteJob(getDemoJob(uid, name, namespace))
			convey.So(jobInfo.GetName(), convey.ShouldEqual, name)
		})
	})
}

func getDemoJob(uid, name, namespace string) metav1.Object {
	jobInfo := &ascendv1.AscendJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			UID:       types.UID(uid),
		},
	}
	return jobInfo
}
