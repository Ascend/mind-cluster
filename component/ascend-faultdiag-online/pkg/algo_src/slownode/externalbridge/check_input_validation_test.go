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

// Package externalbridge is a DT collection for func in check_input_validation
package externalbridge

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"

	"ascend-common/common-utils/hwlog"
	"ascend-faultdiag-online/pkg/algo_src/slownode/config"
	"ascend-faultdiag-online/pkg/core/model"
	"ascend-faultdiag-online/pkg/core/model/enum"
)

const logLineLength = 256

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

func TestCheckConfigValid(t *testing.T) {
	convey.Convey("test checkConfigValid", t, func() {
		var conf = config.AlgoInputConfig{}
		// no detectionLevel
		convey.So(checkConfigValid(conf), convey.ShouldBeFalse)
		// wrong detectionLevel
		conf.DetectionLevel = "wrongdata"
		convey.So(checkConfigValid(conf), convey.ShouldBeFalse)
		conf.DetectionLevel = "cluster"
		// wrong FilePath
		convey.So(checkConfigValid(conf), convey.ShouldBeFalse)
		conf.FilePath = "path"
		// wrong jobName
		convey.So(checkConfigValid(conf), convey.ShouldBeFalse)
		conf.JobName = "jobname"
		// wrong Nsigma
		conf.Nsigma = -1
		convey.So(checkConfigValid(conf), convey.ShouldBeFalse)
		conf.Nsigma = 1
		// wrong NormalNumber
		convey.So(checkConfigValid(conf), convey.ShouldBeFalse)
		conf.NormalNumber = 2
		// wrong NconsecAnomaliesSignifySlow
		convey.So(checkConfigValid(conf), convey.ShouldBeFalse)
		conf.NconsecAnomaliesSignifySlow = 2
		// wrong NsecondsOneDetection
		conf.NsecondsOneDetection = -1
		convey.So(checkConfigValid(conf), convey.ShouldBeFalse)
		conf.NsecondsOneDetection = 1
		// wrong DegradationPercentage
		convey.So(checkConfigValid(conf), convey.ShouldBeFalse)
		conf.DegradationPercentage = 1
		// wrong ClusterMeanDistance
		convey.So(checkConfigValid(conf), convey.ShouldBeFalse)
		conf.ClusterMeanDistance = 2
		// wrong CardsOneNode
		convey.So(checkConfigValid(conf), convey.ShouldBeFalse)
		conf.CardsOneNode = 1
		// normal
		convey.So(checkConfigValid(conf), convey.ShouldBeTrue)
	})
}

func TestCheckConfigDigit(t *testing.T) {
	var cg = map[string]any{}
	assert.False(t, checkConfigDigit(cg))
	cg["normalNumber"] = 1
	assert.False(t, checkConfigDigit(cg))
	cg["nSigma"] = 1
	assert.False(t, checkConfigDigit(cg))
	cg["cardOneNode"] = 1
	assert.False(t, checkConfigDigit(cg))
	cg["nSecondsDoOneDetection"] = 1
	assert.False(t, checkConfigDigit(cg))
	cg["nConsecAnomaliesSignifySlow"] = 1
	assert.False(t, checkConfigDigit(cg))
	cg["degradationPercentage"] = 1
	assert.False(t, checkConfigDigit(cg))
	cg["clusterMeanDistance"] = 1
	assert.True(t, checkConfigDigit(cg))
}

func TestCheckConfigExist(t *testing.T) {
	convey.Convey("test checkConfigExist", t, func() {
		var conf any
		var cmdStr enum.Command
		convey.Convey("conf is nil", func() {
			convey.So(checkConfigExist(conf, cmdStr), convey.ShouldBeFalse)
		})
		conf = map[string]string{}
		convey.Convey("json unmarshal failed", func() {
			patch := gomonkey.ApplyFuncReturn(json.Unmarshal, errors.New("mock json unmarshal failed"))
			defer patch.Reset()
			convey.So(checkConfigExist(conf, cmdStr), convey.ShouldBeFalse)
		})
		convey.Convey("lack of jobId", func() {
			convey.So(checkConfigExist(conf, cmdStr), convey.ShouldBeFalse)
		})
		conf = map[string]string{"jobId": "jobId"}
		convey.So(checkConfigExist(conf, enum.Stop), convey.ShouldBeTrue)
		convey.Convey("lack of filePath", func() {
			convey.So(checkConfigExist(conf, cmdStr), convey.ShouldBeFalse)
		})
		conf = map[string]string{"jobId": "jobId", "filePath": "path"}
		patch := gomonkey.ApplyFuncReturn(checkConfigDigit, true)
		convey.So(checkConfigExist(conf, enum.Start), convey.ShouldBeTrue)
		patch.Reset()
	})
}

func TestCheckInvalidInput(t *testing.T) {
	convey.Convey("test checkInvalidInput", t, func() {
		convey.Convey("target is not node or cluster", func() {
			convey.So(checkInvalidInput(&model.Input{}), convey.ShouldBeFalse)
		})
		convey.Convey("eventType is not slownode or dataparse", func() {
			convey.So(checkInvalidInput(&model.Input{Target: enum.Cluster}), convey.ShouldBeFalse)
		})
		input := &model.Input{Target: enum.Cluster, Command: enum.Register, EventType: enum.DataParse}
		convey.Convey("command is register", func() {
			// func is nil
			convey.So(checkInvalidInput(input), convey.ShouldBeFalse)
			// func is not nil
			input.Func = func(s string) {}
			convey.So(checkInvalidInput(input), convey.ShouldBeTrue)
		})
		convey.Convey("command is not start or not reload", func() {
			input.Command = enum.Stop
			convey.So(checkInvalidInput(input), convey.ShouldBeTrue)
		})
		convey.Convey("eventType is slownode or dataparse", func() {
			patch := gomonkey.ApplyFuncReturn(checkConfigExist, false)
			defer patch.Reset()
			input.Command = enum.Start
			input.EventType = enum.SlowNodeAlgo
			convey.So(checkInvalidInput(input), convey.ShouldBeFalse)
			input.EventType = enum.DataParse
			convey.So(checkInvalidInput(input), convey.ShouldBeTrue)
		})
	})
}

func TestTransformJsonToStruct(t *testing.T) {
	convey.Convey("test transformJsonToStruct", t, func() {
		var input = &model.Input{}
		var cg = &config.DataParseModel{}
		convey.Convey("json marshal failed", func() {
			patch := gomonkey.ApplyFuncReturn(json.Marshal, nil, errors.New("mock json marshal failed"))
			err := transformJsonToStruct(input, cg)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldEqual, "mock json marshal failed")
			patch.Reset()
		})
		patches := gomonkey.ApplyFuncReturn(json.Marshal, nil, nil)
		patches.ApplyFuncReturn(json.Unmarshal, nil)
		defer patches.Reset()
		convey.Convey("normal", func() {
			err := transformJsonToStruct(input, cg)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}
