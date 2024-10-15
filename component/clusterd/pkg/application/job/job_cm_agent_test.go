// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package job a series of job function
package job

import (
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"volcano.sh/apis/pkg/client/clientset/versioned"
)

const (
	vcJob = "volcano.sh/job-name"
	acJob = "job-name"
)

// TestGetWorkName test GetWorkName
func TestGetWorkName(t *testing.T) {
	convey.Convey("agent GetWorkName", t, func() {
		labels := make(map[string]string, 1)

		convey.Convey("return volcano-job when label contains volcano.sh/job-nam", func() {
			labels[vcJob] = vcJob
			labels[acJob] = acJob
			work := getWorkName(labels)
			convey.So(work, convey.ShouldEqual, vcJob)
		})
		convey.Convey("return ascend-job when label contains job-name", func() {
			labels[acJob] = acJob
			work := getWorkName(labels)
			convey.So(work, convey.ShouldEqual, acJob)
		})
		convey.Convey("return empty string when label neither contains volcano.sh/job-nam nor job-name ", func() {
			work := getWorkName(labels)
			convey.So(work, convey.ShouldEqual, "")
		})
	})
}

// TestGetPodInfo test getPodInfo
func TestGetPodInfo(t *testing.T) {
	convey.Convey("test getPodInfo", t, func() {
		convey.Convey("obj is not a metaData", func() {
			obj := struct{}{}
			info, _ := getPodInfo(obj, "test")
			convey.So(info, convey.ShouldBeNil)
		})
		convey.Convey("obj is a metaData", func() {
			obj := &v1.Pod{TypeMeta: metav1.TypeMeta{}, ObjectMeta: metav1.ObjectMeta{
				Name:      mockPodName1,
				Namespace: mockNamespace,
				Labels:    map[string]string{vcJob: mockJobName},
				UID:       types.UID(mockPodUID1),
			},
				Spec: v1.PodSpec{}, Status: v1.PodStatus{}}
			expect := &podIdentifier{
				namespace: mockNamespace,
				name:      mockPodName1,
				jobName:   mockJobName,
				eventType: EventAdd,
				UID:       mockPodUID1,
			}
			info, _ := getPodInfo(obj, EventAdd)
			convey.So(info, convey.ShouldResemble, expect)
		})
	})
}

func TestNewAgent(t *testing.T) {
	convey.Convey("test NewAgent", t, func() {
		convey.Convey("new agent failed", func() {
			patch := gomonkey.ApplyFunc(labels.NewRequirement,
				func(_ string, _ selection.Operator, _ []string, opts ...field.PathOption) (*labels.Requirement, error) {
					return nil, fmt.Errorf("new agent failed")
				})
			defer patch.Reset()
			_, err := NewAgent(nil, nil, nil)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("new agent success", func() {
			kubeClientSet, err := kubernetes.NewForConfig(&rest.Config{})
			convey.So(err, convey.ShouldBeNil)
			config := NewConfig()
			vcClient := &versioned.Clientset{}
			_, err = NewAgent(kubeClientSet, config, vcClient)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}
