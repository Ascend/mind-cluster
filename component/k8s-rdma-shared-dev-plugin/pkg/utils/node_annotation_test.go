/*
   Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.
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

package utils

import (
	"context"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"ascend-common/common-utils/hwlog"
)

func init() {
	_ = hwlog.InitRunLogger(&hwlog.LogConfig{OnlyToStdout: true}, context.Background())
}

func TestUpdateNodeAnnotationSuccessAndMissingNode(t *testing.T) {
	const nodeName = "test-node"
	const key = "huawei.com/dpu.resource.name"
	const value = "huawei.com/ub_rdma_1,huawei.com/ub_rdma_2"

	convey.Convey("When node exists, annotation should be patched successfully", t, func() {
		node := &corev1.Node{
			ObjectMeta: metav1.ObjectMeta{Name: nodeName, Annotations: map[string]string{}},
		}
		clientset := fake.NewSimpleClientset(node)

		err := UpdateNodeAnnotation(clientset, nodeName, key, value)
		convey.So(err, convey.ShouldBeNil)

		updated, getErr := clientset.CoreV1().Nodes().Get(context.TODO(), nodeName, metav1.GetOptions{})
		convey.So(getErr, convey.ShouldBeNil)
		convey.So(updated.Annotations[key], convey.ShouldEqual, value)
	})

	convey.Convey("When node is missing, should return error after retries", t, func() {
		patches := gomonkey.ApplyFunc(time.Sleep, func(_ time.Duration) {})
		defer patches.Reset()

		clientset := fake.NewSimpleClientset()
		err := UpdateNodeAnnotation(clientset, "missing-node", key, value)
		convey.So(err, convey.ShouldNotBeNil)
	})
}
