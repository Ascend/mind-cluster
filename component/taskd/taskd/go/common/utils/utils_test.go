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

// Package utils is to provide go runtime utils
package utils

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"

	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
	"clusterd/pkg/interface/grpc/recover"
	"taskd/common/constant"
)

const (
	tryTime      = 10
	anyPath      = "/a/b/c"
	anyCorrectIp = "127.0.0.1"
	anfWrongIp   = "abc.aaa.333.1111.44"
)

func init() {
	hwlog.InitRunLogger(&hwlog.LogConfig{OnlyToStdout: true}, nil)
}

func TestGetProfilingSwitchValidJson(t *testing.T) {
	t.Run("valid json content", func(t *testing.T) {
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		// Mock read file content
		patches.ApplyFunc(utils.ReadLimitBytes, func(path string) ([]byte, error) {
			return json.Marshal(constant.ProfilingSwitch{
				CommunicationOperator: "ON",
				Step:                  "OFF",
				SaveCheckpoint:        "ON",
				FP:                    "OFF",
				DataLoader:            "ON",
			})
		})
		result, _ := GetProfilingSwitch(anyPath)
		expected := constant.ProfilingSwitch{
			CommunicationOperator: "ON",
			Step:                  "OFF",
			SaveCheckpoint:        "ON",
			FP:                    "OFF",
			DataLoader:            "ON",
		}
		if result != expected {
			t.Errorf("expect %+v，actual %+v", expected, result)
		}
	})
}

func TestGetProfilingSwitchErr(t *testing.T) {
	var lastErr error
	cnt := 0
	for i := 0; i < tryTime; i++ {
		_, err := GetProfilingSwitch(anyPath)
		if !errors.Is(err, lastErr) {
			lastErr = err
			cnt++
		}
	}
	convey.ShouldBeTrue(cnt == 1)
}

func TestObjAndStringConvert(t *testing.T) {
	msg := constant.ProfilingSwitch{
		CommunicationOperator: constant.SwitchON,
		Step:                  constant.SwitchON,
		SaveCheckpoint:        constant.SwitchON,
		FP:                    constant.SwitchON,
		DataLoader:            constant.SwitchON,
	}
	json := ObjToString(msg)
	_, err := StringToObj[constant.ProfilingSwitch](json)
	convey.ShouldBeNil(err)
}

func TestPfSwitchToPfDomainSwitch(t *testing.T) {
	domainSwitch := PfSwitchToPfDomainSwitch(constant.ProfilingSwitch{
		CommunicationOperator: constant.SwitchON,
		Step:                  constant.SwitchON,
		SaveCheckpoint:        constant.SwitchOFF,
		FP:                    constant.SwitchOFF,
		DataLoader:            constant.SwitchOFF,
	})
	convey.ShouldBeTrue(domainSwitch.DefaultDomainAble)
	convey.ShouldBeTrue(domainSwitch.CommDomainAble)
}

func TestProfilingResultToBizCode(t *testing.T) {
	code := ProfilingResultToBizCode(constant.ProfilingResult{
		DefaultDomain: constant.NewProfilingExecRes(constant.On),
		CommDomain:    constant.NewProfilingExecRes(constant.On),
	})
	convey.ShouldEqual(code,
		constant.ProfilingAllCloseCode+constant.ProfilingDefaultOpenInc+constant.ProfilingCommOpenInc)

	code = ProfilingResultToBizCode(constant.ProfilingResult{
		DefaultDomain: constant.NewProfilingExecRes(constant.Exp),
		CommDomain:    constant.NewProfilingExecRes(constant.Exp),
	})
	convey.ShouldEqual(code,
		constant.ProfilingAllCloseCode+constant.ProfilingDefaultExpInc+constant.ProfilingCommExpInc)

	code = ProfilingResultToBizCode(constant.ProfilingResult{
		DefaultDomain: constant.NewProfilingExecRes(constant.Off),
		CommDomain:    constant.NewProfilingExecRes(constant.Off),
	})
	convey.ShouldEqual(code, constant.ProfilingAllCloseCode)
}

func TestBizCodeToProfilingCmd(t *testing.T) {
	cmd, err := BizCodeToProfilingCmd(constant.ProfilingAllOnCmdCode)
	convey.ShouldBeNil(err)
	convey.ShouldBeTrue(cmd.CommDomainAble)
	convey.ShouldBeTrue(cmd.DefaultDomainAble)
	_, err = BizCodeToProfilingCmd(0)
	convey.ShouldNotBeNil(err)
}

func TestGetClusterdAddr(t *testing.T) {
	patches := gomonkey.NewPatches()
	defer patches.Reset()
	patches.ApplyFunc(os.Getenv, func(string) string {
		return anyCorrectIp
	})
	addr, err := GetClusterdAddr()
	convey.ShouldBeNil(err)
	convey.ShouldEqual(addr, anyCorrectIp+constant.ClusterdPort)
	patches.ApplyFunc(os.Getenv, func(string) string {
		return anfWrongIp
	})
	addr, err = GetClusterdAddr()
	convey.ShouldBeNil(err)
}

func TestGetFaultRanksMapByList(t *testing.T) {
	type args struct {
		faultRanks []*pb.FaultRank
	}
	tests := []struct {
		name string
		args args
		want map[int]int
	}{
		{
			name: "get fault ranks map by list",
			args: args{
				faultRanks: []*pb.FaultRank{
					{RankId: "a"},
					{RankId: "1", FaultType: "a"},
					{RankId: "2", FaultType: "2"},
				},
			},
			want: map[int]int{2: 2},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetFaultRanksMapByList(tt.args.faultRanks); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetFaultRanksMapByList() = %v, want %v", got, tt.want)
			}
		})
	}
}

func buildLogNormalTestCases() []struct {
	name           string
	logFileName    string
	envValue       string
	setupMocks     func() *gomonkey.Patches
	expectError    bool
	expectedErrMsg string
} {
	return []struct {
		name           string
		logFileName    string
		envValue       string
		setupMocks     func() *gomonkey.Patches
		expectError    bool
		expectedErrMsg string
	}{
		{
			name:        "normal case, log env is nil, use default path",
			logFileName: testAbsoluteLogPath,
			envValue:    "",
			setupMocks: func() *gomonkey.Patches {
				patches := gomonkey.ApplyFuncReturn(filepath.Abs, "/default/path/test.log", nil)
				return patches
			},
			expectError: false,
		},
		{
			name:        "normal case, log env is not nil, use define path",
			logFileName: testAbsoluteLogPath,
			envValue:    "/custom/log/path",
			setupMocks: func() *gomonkey.Patches {
				patches := gomonkey.ApplyFuncReturn(filepath.Abs, "/custom/log/path/test.log", nil)
				return patches
			},
			expectError: false,
		},
	}
}

func buildLogAbnormalTestCases() []struct {
	name           string
	logFileName    string
	envValue       string
	setupMocks     func() *gomonkey.Patches
	expectError    bool
	expectedErrMsg string
} {
	return []struct {
		name           string
		logFileName    string
		envValue       string
		setupMocks     func() *gomonkey.Patches
		expectError    bool
		expectedErrMsg string
	}{
		{
			name:        "abnormal case, get absolute path failed",
			logFileName: testRelativeLogPath,
			envValue:    "",
			setupMocks: func() *gomonkey.Patches {
				patches := gomonkey.ApplyFuncReturn(filepath.Abs, "", mockErr)
				return patches
			},
			expectError:    true,
			expectedErrMsg: fmt.Sprintf("get abs log file path error: %s", mockErr.Error()),
		},
		{
			name:        "abnormal case, init runlogger failed",
			logFileName: testAbsoluteLogPath,
			envValue:    "",
			setupMocks: func() *gomonkey.Patches {
				patches := gomonkey.ApplyFuncReturn(filepath.Abs, "/default/path/test.log", nil).
					ApplyFuncReturn(hwlog.InitRunLogger, mockErr)
				return patches
			},
			expectError:    true,
			expectedErrMsg: mockErr.Error(),
		},
	}
}

func TestInitHwLog(t *testing.T) {
	for _, tt := range append(buildLogNormalTestCases(), buildLogAbnormalTestCases()...) {
		t.Run(tt.name, func(t *testing.T) {
			if err := dealEnv(t, tt.envValue); err != nil {
				return
			}

			var patches *gomonkey.Patches
			if tt.setupMocks != nil {
				patches = tt.setupMocks()
			}
			if patches != nil {
				defer patches.Reset()
			}

			err := InitHwLog(tt.logFileName, context.Background())
			if tt.expectError {
				assert.Error(t, err)
				if tt.expectedErrMsg != "" {
					assert.Contains(t, err.Error(), tt.expectedErrMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
