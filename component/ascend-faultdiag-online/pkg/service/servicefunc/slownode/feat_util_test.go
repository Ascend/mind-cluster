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

// Package feature a series of feature test function
package slownode

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	core "k8s.io/api/core/v1"

	"ascend-common/common-utils/hwlog"
	"ascend-faultdiag-online/pkg/model/feature/slownode"
	sm "ascend-faultdiag-online/pkg/module/slownode"
)

var (
	testCmName = "ras-feature-slownode"
)

func init() {
	hwLogConfig := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	hwlog.InitRunLogger(&hwLogConfig, context.Background())
}

func marshalData(data any) []byte {
	dataBuffer, err := json.Marshal(data)
	if err != nil {
		hwlog.RunLog.Errorf("[FD-OL SLOWNODE]json marshal data failed: %v", err)
		return nil
	}
	return dataBuffer
}

func TestParseSlowNodeFeatConfCM(t *testing.T) {
	convey.Convey("TestParseSlowNodeFeatConfCM", t, func() {
		convey.Convey("obj is nil", func() {
			err := ParseCMResult(nil, sm.SlowNodeFatureCMKey, nil)
			convey.So(err.Error(), convey.ShouldEqual, "[FD-OL SLOWNODE]source %!s(<nil>) is not a feature configmap")
		})
		convey.Convey("obj without FeatConf key", func() {
			cm := &core.ConfigMap{}
			cm.Name = testCmName
			err := ParseCMResult(cm, sm.SlowNodeFatureCMKey, nil)
			convey.So(err.Error(), convey.ShouldEndWith, sm.SlowNodeFatureCMKey)
		})
		convey.Convey("obj is valid", func() {
			cm := &core.ConfigMap{}
			cm.Name = testCmName
			job := slownode.SlowNodeJob{}
			job.SlowNode = 1
			cm.Data = map[string]string{}
			cm.Data[sm.SlowNodeFatureCMKey] = string(marshalData(job))
			err := ParseCMResult(cm, sm.SlowNodeFatureCMKey, &job)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}
