// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package util a series of util function
package util

import (
	"context"
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

func initFakeCMByDataMap(m map[string]string) *v1.ConfigMap {
	return &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-cm",
			Namespace: "test-ns",
		},
		Data: m,
	}
}

// TestIsConfigMapChanged test case for IsConfigMapChanged
func TestIsConfigMapChanged(t *testing.T) {
	convey.Convey("Test IsConfigMapChanged", t, func() {
		// add fake kubernetes client
		fakeClient := fake.NewSimpleClientset()

		//  GetConfigMapWithRetry return same ConfigMap
		cm := initFakeCMByDataMap(map[string]string{"key": "value"})

		patche1 := gomonkey.ApplyFunc(GetConfigMapWithRetry, func(k8s kubernetes.Interface, nameSpace,
			cmName string) (*v1.ConfigMap, error) {
			return cm, nil
		})
		defer patche1.Reset()

		convey.So(IsConfigMapChanged(fakeClient, cm, "test-cm", "test-ns"), convey.ShouldBeFalse)

		//  GetConfigMapWithRetry return different ConfigMap
		diffCm := initFakeCMByDataMap(map[string]string{"key": "diff-value"})

		patche2 := gomonkey.ApplyFunc(GetConfigMapWithRetry, func(k8s kubernetes.Interface, nameSpace,
			cmName string) (*v1.ConfigMap, error) {
			return diffCm, nil
		})
		defer patche2.Reset()
		convey.So(IsConfigMapChanged(fakeClient, cm, "test-cm", "test-ns"), convey.ShouldBeTrue)

		// GetConfigMapWithRetry return err
		patche3 := gomonkey.ApplyFunc(GetConfigMapWithRetry, func(k8s kubernetes.Interface, nameSpace,
			cmName string) (*v1.ConfigMap, error) {
			return nil, fmt.Errorf("")
		})
		defer patche3.Reset()
		convey.So(IsConfigMapChanged(fakeClient, cm, "test-cm", "test-ns"), convey.ShouldBeTrue)
	})
}

// TestGetConfigMapWithRetry test case GetConfigMapWithRetry
func TestGetConfigMapWithRetry(t *testing.T) {
	convey.Convey("TestGetServiceIpWithRetry", t, func() {
		fakeClient := fake.NewSimpleClientset()
		cm, err := GetConfigMapWithRetry(fakeClient, fakeNs, fakeName)
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(cm, convey.ShouldBeNil)
	})
}

func TestGetAndUpdateCmByTotalNum(t *testing.T) {
	convey.Convey("Test GetAndUpdateCmByTotalNum", t, func() {
		convey.Convey("total is empty string", func() {
			err := GetAndUpdateCmByTotalNum("", "", "", map[string]string{}, nil)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("totalNum is less than or equal to 1", func() {
			clientSet := fake.NewSimpleClientset()
			cm := initFakeCMByDataMap(map[string]string{"key": "value"})
			_, err := clientSet.CoreV1().ConfigMaps(cm.Namespace).Create(context.TODO(), cm, metav1.CreateOptions{})
			convey.So(err, convey.ShouldBeNil)
			err = GetAndUpdateCmByTotalNum("1", cm.Name, cm.Namespace, map[string]string{"key": "value"}, clientSet)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("get configMap failed", func() {
			clientSet := fake.NewSimpleClientset()
			cm := initFakeCMByDataMap(map[string]string{"key": "value"})
			_, err := clientSet.CoreV1().ConfigMaps(cm.Namespace).Create(context.TODO(), cm, metav1.CreateOptions{})
			convey.So(err, convey.ShouldBeNil)
			err = GetAndUpdateCmByTotalNum("2", cm.Name, cm.Namespace, map[string]string{"key": "value"}, clientSet)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

func TestCreateOrUpdateConfigMap(t *testing.T) {
	convey.Convey("Test CreateOrUpdateConfigMap", t, func() {
		convey.Convey("create new configMap", func() {
			clientSet := fake.NewSimpleClientset()
			cm := initFakeCMByDataMap(map[string]string{"key": "value"})
			err := CreateOrUpdateConfigMap(clientSet, cm, cm.Name, cm.Namespace)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("configMap not changed", func() {
			clientSet := fake.NewSimpleClientset()
			cm := initFakeCMByDataMap(map[string]string{"key": "value"})
			_, err := clientSet.CoreV1().ConfigMaps(cm.ObjectMeta.Namespace).Create(context.TODO(), cm, metav1.CreateOptions{})
			convey.So(err, convey.ShouldBeNil)
			err = CreateOrUpdateConfigMap(clientSet, cm, cm.Name, cm.Namespace)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestCreateOrUpdateCm(t *testing.T) {
	convey.Convey("Test CreateOrUpdateCm", t, func() {
		convey.Convey("get configMap failed", func() {
			clientSet := fake.NewSimpleClientset()
			cm := initFakeCMByDataMap(map[string]string{"key": "value"})
			mockIsNotFound := gomonkey.ApplyFunc(errors.IsNotFound, func(err error) bool {
				return false
			})
			defer mockIsNotFound.Reset()
			err := CreateOrUpdateCm(clientSet, cm)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("create new configMap", func() {
			clientSet := fake.NewSimpleClientset()
			cm := initFakeCMByDataMap(map[string]string{"key": "value"})
			err := CreateOrUpdateCm(clientSet, cm)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestIsNSAndNameMatched(t *testing.T) {
	convey.Convey("Test IsNSAndNameMatched", t, func() {
		cm := initFakeCMByDataMap(map[string]string{"key": "value"})
		result := IsNSAndNameMatched(cm, cm.Namespace, cm.Name)
		convey.So(result, convey.ShouldBeTrue)
	})
}
