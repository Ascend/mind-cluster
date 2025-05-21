/* Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package kube a series of slow node test function
package slownode

import (
	"context"
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/common-utils/hwlog"
	"ascend-faultdiag-online/pkg/model/feature/slownode"
)

var (
	zero          = 0
	one           = 1
	two           = 2
	three         = 3
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

func TestStopInformer(t *testing.T) {
	convey.Convey("TestStopInformer", t, func() {
		convey.So(StopInformer, convey.ShouldNotPanic)
	})
}

func TestCleanFuncs(t *testing.T) {
	convey.Convey("TestCleanFuncs", t, func() {
		CleanFunc()
		convey.So(len(jobFuncList), convey.ShouldEqual, zero)
	})
}

func TestAddCmSLFeatFunc(t *testing.T) {
	convey.Convey("TestAddCmSLFeatFunc", t, func() {
		convey.Convey("add one SLFeat func", func() {
			AddCMHandler(&jobFuncList, func(info *slownode.SlowNodeJob, info2 *slownode.SlowNodeJob, s string) {})
			convey.So(len(jobFuncList), convey.ShouldEqual, one)
		})
		convey.Convey("add two SLFeat func", func() {
			AddCMHandler(&jobFuncList, func(info *slownode.SlowNodeJob, info2 *slownode.SlowNodeJob, s string) {})
			convey.So(len(jobFuncList), convey.ShouldEqual, two)
		})
		convey.Convey("add two different business func", func() {
			AddCMHandler(&jobFuncList, func(info *slownode.SlowNodeJob, info2 *slownode.SlowNodeJob, s string) {})
			convey.So(len(jobFuncList), convey.ShouldEqual, three)
		})
	})
}

func TestCheckConfigMapIsSlowNodeFeatConf(t *testing.T) {
	convey.Convey("test filterSlowNodeJob", t, func() {
		var obj interface{}
		mockMatchedTrue := gomonkey.ApplyFunc(IsNameMatched, func(any, string) bool {
			return true
		})
		defer mockMatchedTrue.Reset()
		slowNodeCheck := filterSlowNodeFeature(obj)
		convey.So(slowNodeCheck, convey.ShouldBeTrue)
	})
}
