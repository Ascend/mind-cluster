// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package job a series of job function
package job

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"volcano.sh/apis/pkg/apis/scheduling"
	"volcano.sh/apis/pkg/client/clientset/versioned"

	"clusterd/pkg/common/constant"
)

// TestSyncJob test SyncJob
func TestSyncJob(t *testing.T) {
	convey.Convey("test SyncJob", t, func() {
		convey.Convey("error should not be nil", func() {
			pg := mockPodGroup()
			agent := mockAgentEmpty()
			err := SyncJob(pg.ObjectMeta, constant.AddOperator, nil, agent)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

// TestHandlePGAddOrUpdateEvent test HandlePGAddOrUpdateEvent
func TestHandlePGAddOrUpdateEvent(t *testing.T) {
	convey.Convey("test HandlePGAddOrUpdateEvent", t, func() {
		agent := mockAgentEmpty()
		job := mockJobEmpty()
		convey.Convey("add event, error should be nil", func() {
			mockAddEvent := gomonkey.ApplyMethod(reflect.TypeOf(new(jobModel)), "AddEvent",
				func(_ *jobModel, _ *Agent) error {
					return nil
				})
			defer mockAddEvent.Reset()
			err := HandlePGAddOrUpdateEvent(EventAdd, agent, job)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("update event, error should be nil", func() {
			mockUpdateEvent := gomonkey.ApplyMethod(reflect.TypeOf(new(jobModel)), "EventUpdate",
				func(_ *jobModel, _ *Agent) error {
					return nil
				})
			defer mockUpdateEvent.Reset()
			err := HandlePGAddOrUpdateEvent(EventUpdate, agent, job)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("other event, error should not be nil", func() {
			err := HandlePGAddOrUpdateEvent(EventDelete, agent, job)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

// TestNewVCClientK8s test NewVCClientK8s
func TestNewVCClientK8s(t *testing.T) {
	convey.Convey("test NewVCClientK8s", t, func() {
		convey.Convey("build vcClient config failed", func() {
			_, err := NewVCClientK8s()
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("get vcClient failed", func() {
			patch := gomonkey.ApplyFunc(clientcmd.BuildConfigFromFlags, func(_, _ string) (*rest.Config, error) {
				return nil, nil
			}).ApplyFunc(versioned.NewForConfig, func(c *rest.Config) (*versioned.Clientset, error) {
				return nil, fmt.Errorf("get vcClient failed")
			})
			defer patch.Reset()
			_, err := NewVCClientK8s()
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("new VCClientK8s success", func() {
			patch := gomonkey.ApplyFunc(clientcmd.BuildConfigFromFlags, func(_, _ string) (*rest.Config, error) {
				return nil, nil
			}).ApplyFunc(versioned.NewForConfig, func(c *rest.Config) (*versioned.Clientset, error) {
				return nil, nil
			})
			defer patch.Reset()
			_, err := NewVCClientK8s()
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

// TestNewConfig test NewConfig
func TestNewConfig(t *testing.T) {
	convey.Convey("create new config", t, func() {
		configs := NewConfig()
		convey.So(configs.DryRun, convey.ShouldEqual, dryRun)
		convey.So(configs.DisplayStatistic, convey.ShouldEqual, displayStatistic)
		convey.So(configs.CmCheckInterval, convey.ShouldEqual, cmCheckInterval)
		convey.So(configs.CmCheckTimeout, convey.ShouldEqual, cmCheckTimeout)
	})
}

func mockPodGroup() *scheduling.PodGroup {
	return &scheduling.PodGroup{
		TypeMeta: v1.TypeMeta{
			APIVersion: "scheduling.volcano.sh/v1beta1",
			Kind:       "PodGroup",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      "podgroup-7a09885b-b753-4924-9fba-77c0836bac20",
			Namespace: mockNamespace,
			Labels:    make(map[string]string),
			OwnerReferences: []v1.OwnerReference{
				{
					APIVersion: "v1",
					Kind:       "Pod",
					Name:       "pod1",
					UID:        "7a09885b-b753-4924-9fba-77c0836bac20",
				},
			},
		},
		Spec: scheduling.PodGroupSpec{
			MinMember: 1,
		},
		Status: scheduling.PodGroupStatus{
			Phase: scheduling.PodGroupRunning,
		},
	}
}

func TestDeleteJobSummaryCM(t *testing.T) {
	convey.Convey("Test deleteJobSummaryCM", t, func() {
		mockClient := fake.NewSimpleClientset()
		err := deleteJobSummaryCM(mockClient, false, nil)
		convey.So(err, convey.ShouldBeNil)
	})
}
