/*
Copyright(C)2025. Huawei Technologies Co.,Ltd. All rights reserved.

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

// package slownode is a DT collection for func in slownode
package slownode

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestKeyGenerator(t *testing.T) {
	convey.Convey("Test KeyGenerator", t, func() {
		convey.Convey("test NodeAlgoResult key generator", func() {
			result := NodeAlgoResult{
				Namespace: "test-namespace",
			}
			result.JobName = "test-job"
			key := result.KeyGenerator()
			expectedKey := "test-namespace/test-job"
			convey.So(key, convey.ShouldEqual, expectedKey)
		})
		convey.Convey("test NodeDataProfilingResultt key generator", func() {
			result := NodeDataProfilingResult{
				Namespace: "test-namespace",
			}
			result.JobName = "test-job"
			key := result.KeyGenerator()
			expectedKey := "test-namespace/test-job"
			convey.So(key, convey.ShouldEqual, expectedKey)
		})
		convey.Convey("test Job key generator", func() {
			result := Job{
				Namespace: "test-namespace",
			}
			result.JobName = "test-job"
			key := result.KeyGenerator()
			expectedKey := "test-namespace/test-job"
			convey.So(key, convey.ShouldEqual, expectedKey)
		})
	})
}
