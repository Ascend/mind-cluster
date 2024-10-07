/*
Copyright(C) 2023. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*
Package controllers is using for reconcile AscendJob.
*/

package v1

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/api/core/v1"

	_ "ascend-operator/pkg/testtool"
)

func TestGetVcRescheduleCM(t *testing.T) {
	convey.Convey("getVcRescheduleCM", t, func() {
		rc := &ASJobReconciler{}
		convey.Convey("01-get configmap with nil patch should return right result", func() {
			cm := &v1.ConfigMap{}
			patch := gomonkey.ApplyPrivateMethod(new(ASJobReconciler), "getConfigmap",
				func(_ *ASJobReconciler, _ string, _ string) (*v1.ConfigMap, error) {
					return cm, nil
				})
			defer patch.Reset()
			cm, err := rc.getVcRescheduleCM()
			convey.ShouldBeNil(err)
			convey.ShouldNotBeNil(cm)
		})
	})
}
