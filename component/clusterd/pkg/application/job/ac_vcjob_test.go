// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package job a series of job function
package job

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/npu-exporter/v6/common-utils/hwlog"
	apiCoreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"volcano.sh/apis/pkg/apis/scheduling/v1beta1"
	"volcano.sh/apis/pkg/client/clientset/versioned"
)

const (
	initializing    = "initializing"
	mockJobUID      = "hj235"
	mockJobName     = "test"
	mockJobLabelKey = "volcano.sh/job-name"

	mockNamespace = "vcjob"

	cmName = "job-summary-test"
	cmUID  = "vcjob"

	mockTimeStr   = "1782233"
	logLineLength = 256
)

func init() {
	config := hwlog.LogConfig{
		OnlyToStdout:  true,
		MaxLineLength: logLineLength,
	}
	err := hwlog.InitRunLogger(&config, context.TODO())
	if err != nil {
		fmt.Println(err)
	}
}

// TestAddEvent test AddEvent
func TestAddEvent(t *testing.T) {
	convey.Convey("test AddEvent", t, func() {
		agent := mockAgentEmpty()
		job := mockJobEmpty()
		job.Namespace = mockNamespace
		job.Name = mockJobName
		job.Uid = mockJobUID
		convey.Convey("worker already exist, error should be nil", func() {
			agent.BsWorker[job.Uid] = NewJobWorker(agent, job.Info, nil, 1)
			err := job.AddEvent(agent)
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("check cm created failed, error should not be nil", func() {
			mockInitCM := gomonkey.ApplyFunc(initCM, func(kubeClientSet kubernetes.Interface, job *jobModel) {
				return
			})
			mockCheckCMCreation := gomonkey.ApplyFunc(checkCMCreation, func(namespace, name string,
				kubeClientSet kubernetes.Interface, config *Config) (*apiCoreV1.ConfigMap, error) {
				return nil, fmt.Errorf("failed to get configmap")
			})
			defer mockInitCM.Reset()
			defer mockCheckCMCreation.Reset()
			err := job.AddEvent(agent)
			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("job key does not exist, error should not be nil", func() {
			mockInitCM := gomonkey.ApplyFunc(initCM, func(kubeClientSet kubernetes.Interface, job *jobModel) {
				return
			})
			mockCheckCMCreation := gomonkey.ApplyFunc(checkCMCreation, func(namespace, name string,
				kubeClientSet kubernetes.Interface, config *Config) (*apiCoreV1.ConfigMap, error) {
				return mockConfigMap(), nil
			})
			defer mockInitCM.Reset()
			defer mockCheckCMCreation.Reset()
			err := job.AddEvent(agent)
			convey.So(err, convey.ShouldNotBeNil)
		})

	})
}

// TestEventUpdate test EventUpdate
func TestEventUpdate(t *testing.T) {
	convey.Convey("test EventUpdate", t, func() {
		agent := mockAgentEmpty()
		job := mockJobEmpty()
		job.Namespace = mockNamespace
		job.Name = mockJobName
		job.Uid = mockJobUID
		convey.Convey("job key does not exist, error should not be nil", func() {
			mockAddEvent := gomonkey.ApplyMethod(reflect.TypeOf(new(jobModel)), "AddEvent",
				func(_ *jobModel, _ *Agent) error {
					return fmt.Errorf("add event error")
				})
			defer mockAddEvent.Reset()
			err := job.EventUpdate(agent)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("job key already exist, error should be nil", func() {
			agent.BsWorker[job.Uid] = NewJobWorker(agent, job.Info, nil, 1)
			err := job.EventUpdate(agent)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

// TestDeleteWorker test DeleteWorker
func TestDeleteWorker(t *testing.T) {
	convey.Convey("test DeleteWorker", t, func() {
		agent := mockAgentEmpty()
		job := mockJobEmpty()
		job.Namespace = mockNamespace
		job.Name = mockJobName
		job.Uid = mockJobUID
		identifier := job.Uid
		agent.BsWorker[identifier] = NewJobWorker(agent, job.Info, nil, 1)
		mockUpdateCMOnDelete := gomonkey.ApplyPrivateMethod(reflect.TypeOf(job), "updateCMOnDeleteEvent",
			func(_ *jobModel, kubeClientSet kubernetes.Interface) error {
				return nil
			})
		defer mockUpdateCMOnDelete.Reset()

		convey.Convey("delete not exist worker, current worker should exist", func() {
			job.DeleteWorker(mockNamespace, mockJobName, mockJobUID+mockJobUID, agent)
			_, exist := agent.BsWorker[identifier]
			convey.So(exist, convey.ShouldBeTrue)
		})

		convey.Convey("agent config displayStatistic is true, current worker should not exist", func() {
			agent.Config.DisplayStatistic = true
			job.DeleteWorker(mockNamespace, mockJobName, mockJobUID, agent)
			_, exist := agent.BsWorker[identifier]
			convey.So(exist, convey.ShouldBeFalse)
		})

		convey.Convey("delete exist worker, current worker should not exist", func() {
			job.DeleteWorker(mockNamespace, mockJobName, mockJobUID, agent)
			_, exist := agent.BsWorker[identifier]
			convey.So(exist, convey.ShouldBeFalse)
		})
		convey.Convey("delete exist worker, update configmap failed", func() {
			patch := gomonkey.ApplyPrivateMethod(reflect.TypeOf(job), "updateCMOnDeleteEvent",
				func(_ *jobModel, _ kubernetes.Interface) error {
					return fmt.Errorf("update configmap failed")
				})
			defer patch.Reset()
			job.DeleteWorker(mockNamespace, mockJobName, mockJobUID, agent)
			_, exist := agent.BsWorker[identifier]
			convey.So(exist, convey.ShouldBeFalse)
		})
	})
}

// TestRanktableFactory test RanktableFactory
func TestRanktableFactory(t *testing.T) {
	convey.Convey("test RankTableFactory", t, func() {

		convey.Convey("err ==nil& when RankTableStatus is ok", func() {
			model := &jobModel{replicas: 1}
			rt, re, err := ranktableFactory(model, RankTableStatus{Status: initializing})
			convey.So(err, convey.ShouldEqual, nil)
			convey.So(rt.GetStatus(), convey.ShouldEqual, initializing)
			convey.So(re, convey.ShouldEqual, 1)
			rv := reflect.ValueOf(rt).Elem()
			convey.So(rv.FieldByName("ServerCount").String(), convey.ShouldEqual, "0")
		})
	})
}

// TestGetPGJobInfo test getPGJobInfo
func TestGetPGJobInfo(t *testing.T) {
	convey.Convey("test getPGJobInfo", t, func() {
		obj := &v1beta1.PodGroup{TypeMeta: v1.TypeMeta{}, ObjectMeta: v1.ObjectMeta{Name: mockJobName,
			Namespace: mockNamespace, Labels: map[string]string{mockJobLabelKey: mockJobName},
			GenerateName: "", SelfLink: "", UID: types.UID(mockJobUID), ResourceVersion: "",
			Generation: 0, CreationTimestamp: v1.Now(), DeletionTimestamp: nil,
			DeletionGracePeriodSeconds: nil, Annotations: nil,
			OwnerReferences: []v1.OwnerReference{{Name: mockJobName, UID: types.UID(mockJobUID)}},
			Finalizers:      nil, ManagedFields: nil},
			Spec: v1beta1.PodGroupSpec{}, Status: v1beta1.PodGroupStatus{}}
		metaData, _ := meta.Accessor(obj)
		jobName, uid := getPGJobInfo(metaData)
		convey.So(jobName, convey.ShouldEqual, mockJobName)
		convey.So(uid, convey.ShouldEqual, mockJobUID)
	})
}

// TestGetUnixTime2String test getUnixTime2String
func TestGetUnixTime2String(t *testing.T) {
	convey.Convey("test getUnixTime2String", t, func() {
		timeStr := getUnixTime2String()
		convey.So(timeStr, convey.ShouldHaveSameTypeAs, mockTimeStr)
	})
}

func mockAgentEmpty() *Agent {
	return &Agent{
		Config:        &Config{},
		BsWorker:      map[string]PodWorker{},
		podsInformer:  nil,
		podsIndexer:   nil,
		KubeClientSet: &kubernetes.Clientset{},
		RwMutex:       sync.RWMutex{},
		vcClient:      &versioned.Clientset{},
	}
}

func mockJobEmpty() *jobModel {
	return &jobModel{
		key: "",
		Info: Info{
			Namespace:         "",
			Name:              "",
			Key:               "",
			Version:           0,
			Uid:               "",
			CreationTimestamp: v1.Time{},
		},
		replicas: 0,
		devices:  nil,
	}
}

func mockConfigMap() *apiCoreV1.ConfigMap {
	return &apiCoreV1.ConfigMap{
		TypeMeta: v1.TypeMeta{
			Kind:       "",
			APIVersion: "",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:                       cmName,
			GenerateName:               "",
			Namespace:                  mockNamespace,
			SelfLink:                   "",
			UID:                        cmUID,
			ResourceVersion:            "",
			Generation:                 0,
			CreationTimestamp:          v1.Time{},
			DeletionTimestamp:          nil,
			DeletionGracePeriodSeconds: nil,
			Labels:                     nil,
			Annotations:                nil,
			OwnerReferences:            nil,
			Finalizers:                 nil,
			ManagedFields:              nil,
		},
		Immutable:  nil,
		Data:       map[string]string{},
		BinaryData: nil,
	}
}

func TestUpdateCMOnDeleteEvent(t *testing.T) {
	convey.Convey("test updateCMOnDeleteEvent", t, func() {
		job := mockJobEmpty()
		kubeClientSet := fake.NewSimpleClientset()
		err := job.updateCMOnDeleteEvent(kubeClientSet)
		convey.So(err, convey.ShouldNotBeNil)
		cm := &apiCoreV1.ConfigMap{
			ObjectMeta: v1.ObjectMeta{Name: cmName, Namespace: mockNamespace},
			Data:       map[string]string{},
		}
		job.Info = mockJobInfo()
		_, err = kubeClientSet.CoreV1().ConfigMaps(job.Namespace).Create(context.TODO(), cm, v1.CreateOptions{})
		convey.So(err, convey.ShouldBeNil)
		err = job.updateCMOnDeleteEvent(kubeClientSet)
		convey.So(err, convey.ShouldNotBeNil)
	})
}
