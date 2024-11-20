/* Copyright(C) 2023. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package pkg for noded ut
package pkg

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"huawei.com/npu-exporter/v5/common-utils/hwlog"
)

const (
	illegalLength            = kubeEnvMaxLength + 1
	ubuntuHostName          = "ubuntu-05"
	illegalHeartbeatInterval = MaxHeartbeatInterval + 1
)

// ErrEmpty empty error
var ErrEmpty = errors.New("")
var log = &hwlog.LogConfig{LogFileName: "", OnlyToStdout: true}

func init() {
	if err := hwlog.InitRunLogger(log, context.Background()); err != nil {
		fmt.Printf("hwlog init failed, error is %v", err)
		return
	}
}

// TestValidHeartbeatInterval test function ValidHeartbeatInterval
func TestValidHeartbeatInterval(t *testing.T) {
	ast := assert.New(t)
	testCase := []struct {
		caseName string
		interval int
		want     error
	}{
		{"Heartbeat interval is legal", DefaultHeartbeatInterval, nil},
		{"Heartbeat interval is illegal", illegalHeartbeatInterval, ErrEmpty},
		{"Heartbeat interval is zero", 0, ErrEmpty},
	}
	for _, tCase := range testCase {
		err := ValidHeartbeatInterval(tCase.interval)
		if err != nil {
			ast.Equal(reflect.TypeOf(tCase.want).Kind(), reflect.TypeOf(err).Kind())
			continue
		}
		ast.Nil(err)
	}
}
