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

// Package utils is a DT collection for func in get_detection_groups
package applicablescenejudgment

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"

	"ascend-common/common-utils/hwlog"
	"ascend-faultdiag-online/pkg/algo_src/slownode/config"
)

func init() {
	config := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	err := hwlog.InitRunLogger(&config, nil)
	if err != nil {
		fmt.Println(err)
	}
}

func updateTempFile(t *testing.T, tmpFile *os.File, s string) {
	// 清空 tmpFile 的内容
	err := tmpFile.Truncate(0)
	if err != nil {
		t.Fatal(err)
	}

	// 重新设置文件的读写位置到文件开头
	_, err = tmpFile.Seek(0, 0)
	if err != nil {
		t.Fatal(err)
	}

	_, err = tmpFile.WriteString(s)
	if err != nil {
		t.Fatal(err)
	}
}

// checkEPContent 函数的测试
func TestCheckEPContent(t *testing.T) {
	// 创建临时文件用于测试
	tmpFile, err := os.CreateTemp("", "test.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	// 测试用例1: 文件读取错误
	updateTempFile(t, tmpFile, `{"EP": [[1, 2]], "TP": [[1, 2]]}`)
	_, err = checkEPContent("/path/to/nonexistent/file.json")
	assert.Error(t, err)

	// 测试用例2: JSON文件解析错误
	updateTempFile(t, tmpFile, `{"EP: [[1, 2]], "TP": [[1, 2]`)
	_, err = checkEPContent(tmpFile.Name())
	assert.Error(t, err)

	// 测试用例3: EP字段不存在
	updateTempFile(t, tmpFile, `{"TP": [[1, 2]]}`)
	_, err = checkEPContent(tmpFile.Name())
	assert.NoError(t, err)

	// 测试用例4: EP通信域只有一张卡
	updateTempFile(t, tmpFile, `{"EP": [[1]], "TP": [[1, 2]]}`)
	_, err = checkEPContent(tmpFile.Name())
	assert.NoError(t, err)

	// 测试用例5: EP和TP不完全相同
	updateTempFile(t, tmpFile, `{"EP": [[1, 2]], "TP": [[1, 3]]}`)
	_, err = checkEPContent(tmpFile.Name())
	assert.NoError(t, err)

	// 测试用例6: EP和TP完全相同
	updateTempFile(t, tmpFile, `{"EP": [[1, 2]], "TP": [[1, 2]]}`)
	_, err = checkEPContent(tmpFile.Name())
	assert.NoError(t, err)
}

// checkCPContent 函数的测试
func TestCheckCPContent(t *testing.T) {
	// 测试用例1: 文件不存在
	_, err := checkCPContent("nonexistent.json")
	assert.Error(t, err)

	// 测试用例2: 文件存在但内容为空
	tmpFile, err := os.CreateTemp("", "test.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	updateTempFile(t, tmpFile, `{}`)
	exists, err := checkCPContent(tmpFile.Name())
	assert.NoError(t, err)
	assert.False(t, exists)

	// 测试用例3: 文件存在且CP字段为空
	updateTempFile(t, tmpFile, `{"TP": [[1, 2]]}`)
	exists, err = checkCPContent(tmpFile.Name())
	assert.NoError(t, err)
	assert.False(t, exists)

	// 测试用例4: 文件存在且CP通信域只有一张卡
	updateTempFile(t, tmpFile, `{"CP": [[1]]}`)
	exists, err = checkCPContent(tmpFile.Name())
	assert.NoError(t, err)
	assert.False(t, exists)

	// 测试用例5: 文件存在且CP通信域人有多张卡
	updateTempFile(t, tmpFile, `{"CP": [[1, 2]]}`)
	exists, err = checkCPContent(tmpFile.Name())
	assert.NoError(t, err)
	assert.True(t, exists)
}

func TestCheckApplicableScene(t *testing.T) {
	convey.Convey("Test CheckApplicableScene", t, func() {
		var sndConfig *config.DetectionConfig
		var taskName = "taskName"
		convey.Convey("sndConfig is nil", func() {
			res := CheckApplicableScene(sndConfig, taskName)
			convey.So(res, convey.ShouldBeFalse)
		})

		sndConfig = &config.DetectionConfig{}
		convey.Convey("checkEPContent failed", func() {
			patch := gomonkey.ApplyFuncReturn(checkEPContent, false, errors.New("mock checkEPContent failed"))
			defer patch.Reset()
			res := CheckApplicableScene(sndConfig, taskName)
			convey.So(res, convey.ShouldBeFalse)
		})

		convey.Convey("checkEPContent get true", func() {
			patch := gomonkey.ApplyFuncReturn(checkEPContent, true, nil)
			defer patch.Reset()
			res := CheckApplicableScene(sndConfig, taskName)
			convey.So(res, convey.ShouldBeFalse)
		})

		patches := gomonkey.ApplyFuncReturn(checkEPContent, false, nil)
		defer patches.Reset()
		convey.Convey("checkCPContent failed", func() {
			patch := gomonkey.ApplyFuncReturn(checkCPContent, false, errors.New("mock checkCPContent failed"))
			defer patch.Reset()
			res := CheckApplicableScene(sndConfig, taskName)
			convey.So(res, convey.ShouldBeFalse)
		})

		convey.Convey("checkCPContent get true", func() {
			patch := gomonkey.ApplyFuncReturn(checkCPContent, true, nil)
			defer patch.Reset()
			res := CheckApplicableScene(sndConfig, taskName)
			convey.So(res, convey.ShouldBeFalse)
		})

		patches.ApplyFuncReturn(checkCPContent, false, nil)

		convey.Convey("normal case", func() {
			res := CheckApplicableScene(sndConfig, taskName)
			convey.So(res, convey.ShouldBeTrue)
		})
	})
}
