// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package kube a series of kube function
package kube

import (
	"errors"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"volcano.sh/apis/pkg/apis/scheduling/v1beta1"
	"volcano.sh/apis/pkg/client/clientset/versioned"
)

func TestNewVCClientK8s(t *testing.T) {
	convey.Convey("Test newVCClientK8s", t, func() {
		convey.Convey("Should return error when BuildConfigFromFlags fails", func() {
			patches := gomonkey.ApplyFunc(clientcmd.BuildConfigFromFlags, func(_, _ string) (*rest.Config, error) {
				return nil, errors.New("config error")
			})
			defer patches.Reset()

			client, err := newVCClientK8s()
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(client, convey.ShouldBeNil)
		})

		convey.Convey("Should return error when NewForConfig fails", func() {
			patches := gomonkey.ApplyFunc(versioned.NewForConfig, func(_ *rest.Config) (*versioned.Clientset, error) {
				return nil, errors.New("client error")
			})
			defer patches.Reset()

			client, err := newVCClientK8s()
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(client, convey.ShouldBeNil)
		})
	})
}

func mockBuildConfigFromFlags(masterUrl, kubeconfigPath string) (*rest.Config, error) {
	return &rest.Config{
		Host: "http://example.com",
	}, nil
}

func mockNewForConfig(config *rest.Config) (*versioned.Clientset, error) {
	return &versioned.Clientset{}, nil
}

func mockGetPodGroup(name, namespace string) (*v1beta1.PodGroup, error) {
	return &v1beta1.PodGroup{}, nil
}

func mockUpdatePodGroup(pg *v1beta1.PodGroup) (*v1beta1.PodGroup, error) {
	return &v1beta1.PodGroup{}, nil
}

func TestInitClientVolcano(t *testing.T) {
	convey.Convey("Test InitClientVolcano", t, func() {
		patcher1 := gomonkey.ApplyFunc(clientcmd.BuildConfigFromFlags, mockBuildConfigFromFlags)
		defer patcher1.Reset()

		patcher2 := gomonkey.ApplyFunc(versioned.NewForConfig, mockNewForConfig)
		defer patcher2.Reset()

		client, err := InitClientVolcano()
		convey.So(err, convey.ShouldBeNil)
		convey.So(client, convey.ShouldNotBeNil)
	})
}

func TestGetClientVolcano(t *testing.T) {
	convey.Convey("Test GetClientVolcano", t, func() {
		patcher1 := gomonkey.ApplyFunc(clientcmd.BuildConfigFromFlags, mockBuildConfigFromFlags)
		defer patcher1.Reset()

		patcher2 := gomonkey.ApplyFunc(versioned.NewForConfig, mockNewForConfig)
		defer patcher2.Reset()

		_, err := InitClientVolcano()
		convey.So(err, convey.ShouldBeNil)

		client := GetClientVolcano()
		convey.So(client, convey.ShouldNotBeNil)
	})
}

func TestRetryGetPodGroup(t *testing.T) {
	convey.Convey("Test RetryGetPodGroup", t, func() {
		name := "testName"
		namespace := "testNamespace"
		retryTimes := 3

		patcher := gomonkey.ApplyFunc(GetPodGroup, mockGetPodGroup)
		defer patcher.Reset()

		pg, err := RetryGetPodGroup(name, namespace, retryTimes)
		convey.So(err, convey.ShouldBeNil)
		convey.So(pg, convey.ShouldNotBeNil)
	})
}

func TestRetryPatchPodGroupAnnotations(t *testing.T) {
	convey.Convey("Test RetryPatchPodGroupAnnotations", t, func() {
		pgName := "testPgName"
		pgNamespace := "testPgNamespace"
		retryTimes := 1
		annotations := map[string]string{"key": "value"}

		patcher := gomonkey.ApplyFunc(patchPodGroupAnnotation,
			func(_, _ string, _ map[string]string) (*v1beta1.PodGroup, error) {
				return &v1beta1.PodGroup{}, nil
			})
		defer patcher.Reset()

		pg, err := RetryPatchPodGroupAnnotations(pgName, pgNamespace, retryTimes, annotations)
		convey.So(err, convey.ShouldBeNil)
		convey.So(pg, convey.ShouldNotBeNil)
	})
}

func TestRetryPatchPodGroupLabel(t *testing.T) {
	convey.Convey("Test RetryPatchPodGroupLabel", t, func() {
		pgName := "testPgName"
		nameSpace := "testNameSpace"
		retryTimes := 3
		labels := map[string]string{"key": "value"}

		patcher := gomonkey.ApplyFunc(patchPodGroupLabel,
			func(_, _ string, _ map[string]string) (*v1beta1.PodGroup, error) {
				return &v1beta1.PodGroup{}, nil
			})
		defer patcher.Reset()

		pg, err := RetryPatchPodGroupLabel(pgName, nameSpace, retryTimes, labels)
		convey.So(err, convey.ShouldBeNil)
		convey.So(pg, convey.ShouldNotBeNil)
	})
}
