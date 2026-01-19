/*
Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.

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

/*
Package scaling is using for scale AscendJobs.
*/

package scaling

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes/fake"
	"volcano.sh/apis/pkg/apis/scheduling/v1beta1"
	volschedulinglisters "volcano.sh/apis/pkg/client/listers/scheduling/v1beta1"

	apiv1 "ascend-operator/pkg/api/v1"
	_ "ascend-operator/pkg/testtool"
)

type fakePgLister struct {
	*fakeNamespacedPgLister
}

type fakeNamespacedPgLister struct {
	podGroups []*v1beta1.PodGroup
}

func (f *fakeNamespacedPgLister) List(selector labels.Selector) ([]*v1beta1.PodGroup, error) {
	return f.podGroups, nil
}

func (f *fakeNamespacedPgLister) Get(name string) (*v1beta1.PodGroup, error) {
	return nil, nil
}

func (f *fakePgLister) PodGroups(namespace string) volschedulinglisters.PodGroupNamespaceLister {
	return f.fakeNamespacedPgLister
}

func TestValidJob(t *testing.T) {
	convey.Convey("test scaling.Controller.ValidJob", t, func() {
		client := fake.NewSimpleClientset()
		pgLister := &fakePgLister{}
		sc := New(client, pgLister)
		job := &apiv1.Job{}
		convey.Convey("01-job without label should return nil", func() {
			err := sc.ValidJob(job)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("02-job with label of scaling-rule but without group-name return false", func() {
			job.Labels = map[string]string{
				scalingRuleKey: "test",
			}
			err := sc.ValidJob(job)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("03-job with label of scaling-rule and group-name but without jobID return false", func() {
			job.Labels = map[string]string{
				scalingRuleKey: "test",
				groupNameKey:   "group0",
			}
			err := sc.ValidJob(job)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("04-job with label of scaling-rule and group-name but without jobID return false", func() {
			job.Labels = map[string]string{
				scalingRuleKey:  "test",
				groupNameKey:    "group0",
				jobGroupNameKey: "test-job",
			}
			err := sc.ValidJob(job)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestCanCreatePod02(t *testing.T) {
	convey.Convey("test scaling.Controller.CanCreatePod 02", t, func() {
		sc := New(fake.NewSimpleClientset(), &fakePgLister{})
		job := &apiv1.Job{}
		job.Labels = map[string]string{
			scalingRuleKey:  "test",
			groupNameKey:    "group0",
			jobGroupNameKey: "test-job",
		}
		convey.Convey("05-get scaling-rule failed return false", func() {
			res := sc.CanCreatePod(job)
			convey.So(res, convey.ShouldEqual, false)
		})
		patch1 := gomonkey.ApplyPrivateMethod(sc, "getScalingRule", func(_ *Controller,
			_, _ string) ([]map[string]*groupInfo, error) {
			return []map[string]*groupInfo{
				{"group0": {2, 4}, "group1": {1, 8}},
				{"group0": {1, 4}, "group1": {1, 8}},
			}, nil
		})
		defer patch1.Reset()
		convey.Convey("06-get current groups state failed should return false", func() {
			patch2 := gomonkey.ApplyPrivateMethod(sc, "getRuleRefPodGroups", func(_ *Controller,
				_, _, _ string) (map[string]int, error) {
				return nil, errors.New("get current groups state failed")
			})
			defer patch2.Reset()
			res := sc.CanCreatePod(job)
			convey.So(res, convey.ShouldEqual, false)
		})
		convey.Convey("07-scaling-rule not covered current state should return false", func() {
			patch2 := gomonkey.ApplyPrivateMethod(sc, "getRuleRefPodGroups", func(_ *Controller,
				_, _, _ string) (map[string]int, error) {
				return map[string]int{"group0": 2, "group": 1}, nil
			})
			defer patch2.Reset()
			res := sc.CanCreatePod(job)
			convey.So(res, convey.ShouldEqual, false)
		})
		convey.Convey("08-scaling-rule covered current state should return false", func() {
			patch2 := gomonkey.ApplyPrivateMethod(sc, "getRuleRefPodGroups", func(_ *Controller,
				_, _, _ string) (map[string]int, error) {
				return map[string]int{}, nil
			})
			defer patch2.Reset()
			res := sc.CanCreatePod(job)
			convey.So(res, convey.ShouldEqual, true)
		})
	})
}

func TestGetRuleRefPodGroups(t *testing.T) {
	convey.Convey("test scaling.Controller.getRuleRefPodGroups", t, func() {
		lister := &fakePgLister{
			fakeNamespacedPgLister: &fakeNamespacedPgLister{
				podGroups: make([]*v1beta1.PodGroup, 1),
			}}
		sc := New(fake.NewSimpleClientset(), lister)
		namespace, rule, jobID := "default", "scaling-rule", "test-job"
		convey.Convey("01-construct label-selector failed should return error", func() {
			patch := gomonkey.ApplyFunc(metav1.LabelSelectorAsSelector, func(_ *metav1.LabelSelector) (labels.Selector, error) {
				return nil, errors.New("construct label-selector failed")
			})
			defer patch.Reset()
			_, err := sc.getRuleRefPodGroups(namespace, rule, jobID)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("02-list pod-group failed should return error", func() {
			patch1 := gomonkey.ApplyMethod(new(fakeNamespacedPgLister), "List", func(_ *fakeNamespacedPgLister,
				selector labels.Selector) (ret []*v1beta1.PodGroup, err error) {
				return nil, errors.New("list pod-group failed")
			})
			defer patch1.Reset()
			_, err := sc.getRuleRefPodGroups(namespace, rule, jobID)
			convey.So(err, convey.ShouldNotBeNil)
		})
		pg := &v1beta1.PodGroup{}
		convey.Convey("03-pod-group without label of group-name should return error", func() {
			lister.podGroups[0] = pg
			_, err := sc.getRuleRefPodGroups(namespace, rule, jobID)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("04-pod-group without label of group-name should return error", func() {
			pg.Labels = map[string]string{groupNameKey: "group0"}
			pg.Status.Phase = v1beta1.PodGroupRunning
			lister.podGroups[0] = pg
			groups, err := sc.getRuleRefPodGroups(namespace, rule, jobID)
			convey.So(err, convey.ShouldBeNil)
			convey.So(groups, convey.ShouldResemble, map[string]int{"group0": 1})
		})
	})
}

func TestGetScalingRule(t *testing.T) {
	convey.Convey("test scaling.Controller.getScalingRule", t, func() {
		convey.Convey("01-get scaling-rule cm failed should return error", func() {
			client := fake.NewSimpleClientset()
			sc := New(client, &fakePgLister{})
			namespace, name := "default", "scaling-rule"
			convey.Convey("01-get scaling-rule configmap failed should return error", func() {
				_, err := sc.getScalingRule(namespace, name)
				convey.So(err, convey.ShouldNotBeNil)
			})
			cm := &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: namespace,
					Name:      name,
				},
				Data: map[string]string{},
			}
			convey.Convey("02-scaling-rule cm without rule.json should return error", func() {
				_, err := client.CoreV1().ConfigMaps(namespace).Create(context.TODO(), cm, metav1.CreateOptions{})
				convey.So(err, convey.ShouldBeNil)
				_, err = sc.getScalingRule(namespace, name)
				convey.So(err, convey.ShouldNotBeNil)
			})
			convey.Convey("03-scaling-rule cm without invalid rule.json should return error", func() {
				cm.Data[configmapRuleKey] = fakeRuleString()
				_, err := client.CoreV1().ConfigMaps(namespace).Create(context.TODO(), cm, metav1.CreateOptions{})
				convey.So(err, convey.ShouldBeNil)
				rule, err := sc.getScalingRule(namespace, name)
				convey.So(err, convey.ShouldBeNil)
				convey.So(rule, convey.ShouldResemble, []map[string]*groupInfo{
					{
						"group0": {groupNum: 2, serverNumPerGroup: 8},
						"group1": {groupNum: 1, serverNumPerGroup: 8},
					},
					{
						"group0": {groupNum: 1, serverNumPerGroup: 8},
						"group1": {groupNum: 1, serverNumPerGroup: 8},
					},
				})
			})
		})
	})
}

func fakeRuleString() string {
	rule := &Rule{
		Version: "1.0",
		ElasticScalingList: []Item{
			{GroupList: []Group{
				{
					GroupName:         "group0",
					GroupNum:          "2",
					ServerNumPerGroup: "8",
				},
				{
					GroupName:         "group1",
					GroupNum:          "1",
					ServerNumPerGroup: "8",
				},
			}},
			{GroupList: []Group{
				{
					GroupName:         "group0",
					GroupNum:          "1",
					ServerNumPerGroup: "8",
				},
				{
					GroupName:         "group1",
					GroupNum:          "1",
					ServerNumPerGroup: "8",
				},
			}},
		},
	}
	bt, err := json.Marshal(rule)
	if err != nil {
		return ""
	}
	return string(bt)
}
