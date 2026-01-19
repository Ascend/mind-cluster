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
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"ascend-common/common-utils/hwlog"
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
	var cg = map[string]any{}
	assert.False(t, checkConfigExist(cg, enum.Stop))
	assert.False(t, checkConfigExist(cg, "STOP"))
	cg["sharedFilePath"] = "."
	assert.False(t, checkConfigExist(cg, enum.Stop))
	cg["localFilePath"] = "."
	assert.False(t, checkConfigExist(cg, enum.Start))
}

func TestCheckInvalidInput(t *testing.T) {
	// no command
	var input = &model.Input{}
	assert.False(t, checkInvalidInput(input))
	// invalid command
	input.Command = "invalid"
	assert.False(t, checkInvalidInput(input))
	// no target
	input.Target = "invalid"
	assert.False(t, checkInvalidInput(input))
	// command equals registerCallBack and& no func in input
	input.Target = enum.Cluster
	input.Command = enum.Register
	assert.False(t, checkInvalidInput(input))
	// command equals start and no model in input
	input.Command = enum.Start
	input.Model = nil
	assert.False(t, checkInvalidInput(input))
	// command equals start and model in input but wrong type
	input.Model = "model"
	assert.False(t, checkInvalidInput(input))
	// command equals start and& model in input and correct type
	input.Model = map[string]any{"1": 1}
	assert.False(t, checkInvalidInput(input))
	// invalid command
	input.Model = "shishis"
	assert.False(t, checkInvalidInput(input))
	// command equals stop
	input.Model = "stop"
	assert.False(t, checkInvalidInput(input))
}
