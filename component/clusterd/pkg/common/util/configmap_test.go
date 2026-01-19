// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package util a series of util function
package util

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func TestIsNSAndNameMatched(t *testing.T) {
	convey.Convey("Test IsNSAndNameMatched", t, func() {
		convey.Convey("case type is configmap", func() {
			cm := initFakeCMByDataMap(map[string]string{"key": "value"})
			result := IsNSAndNameMatched(cm, cm.Namespace, cm.Name)
			convey.So(result, convey.ShouldBeTrue)
		})
		convey.Convey("case type is not configmap", func() {
			strTypeVal := ""
			result := IsNSAndNameMatched(strTypeVal, "", "")
			convey.So(result, convey.ShouldBeFalse)
		})
	})
}
