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

// Package policy is used for processing superpod information
package policy

import (
	"os"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/common-utils/hwlog"
	"ascend-faultdiag-online/pkg/algo_src/netfault/controllerflags"
)

func TestLoopWaitFile(t *testing.T) {
	convey.Convey("test func loopWaitFile", t, func() {
		convey.Convey("return false when not exist", func() {
			mockCheck := gomonkey.ApplyFuncReturn(CheckCurSuperPodConfigSwitch, true)
			defer mockCheck.Reset()
			mockStat := gomonkey.ApplyFuncReturn(os.Stat, nil, os.ErrNotExist)
			defer mockStat.Reset()
			mockSleep := gomonkey.ApplyFunc(time.Sleep, func(_ time.Duration) {})
			defer mockSleep.Reset()
			ret := loopWaitFile("filePath", "DirPath")
			convey.So(ret, convey.ShouldBeFalse)
		})
		convey.Convey("return false when controllered exitd", func() {
			mockStat := gomonkey.ApplyFuncReturn(os.Stat, nil, nil)
			defer mockStat.Reset()
			mockCheck := gomonkey.ApplyFuncReturn(CheckCurSuperPodConfigSwitch, true)
			defer mockCheck.Reset()
			controllerflags.IsControllerExited.SetState(true)
			ret := loopWaitFile("filePath", "DirPath")
			convey.So(ret, convey.ShouldBeFalse)
		})
	})
}

func TestIsPureLetter(t *testing.T) {
	convey.Convey("TestIsPureLetter", t, func() {
		// 测试纯字母字符串
		convey.Convey("when_str_is_pure_letter", func() {
			convey.So(isPureLetter("HelloWorld"), convey.ShouldBeTrue)
			convey.So(isPureLetter("abc"), convey.ShouldBeTrue)
			convey.So(isPureLetter("ABC"), convey.ShouldBeTrue)
		})

		// 测试包含数字的字符串
		convey.Convey("when_str_contains_digit", func() {
			convey.So(isPureLetter("Hello1"), convey.ShouldBeFalse)
			convey.So(isPureLetter("a1b2c3"), convey.ShouldBeFalse)
		})

		// 测试包含特殊字符的字符串
		convey.Convey("when_str_contains_special_char", func() {
			convey.So(isPureLetter("Hello@World"), convey.ShouldBeFalse)
			convey.So(isPureLetter("abc!"), convey.ShouldBeFalse)
		})
	})
}

func TestIsPureNumber(t *testing.T) {
	convey.Convey("TestIsPureNumber", t, func() {
		// 测试纯数字字符串
		convey.Convey("when_str_is_pure_number", func() {
			convey.So(isPureNumber("12345"), convey.ShouldEqual, true)
			convey.So(isPureNumber("0"), convey.ShouldEqual, true)
		})

		// 测试包含字母的字符串
		convey.Convey("when_str_contains_letter", func() {
			convey.So(isPureNumber("123abc"), convey.ShouldBeFalse)
			convey.So(isPureNumber("abc123"), convey.ShouldBeFalse)
		})

		// 测试包含特殊字符的字符串
		convey.Convey("when_str_contains_special_char", func() {
			convey.So(isPureNumber("123@45"), convey.ShouldBeFalse)
			convey.So(isPureNumber("123!45"), convey.ShouldBeFalse)
		})
	})
}

func TestReadConfigFromFile(t *testing.T) {
	convey.Convey("TestReadConfigFromFile", t, func() {
		fileContent := []byte(`
supperssedPeriod=0
networkType=1
pingType=0
pingTimes=5
pingInterval=1
period=10
netFault=on
`)
		targetKeys := []string{"networkType", "pingType", "pingTimes", "pingInterval", "suppressedPeriod", "period"}
		result := ReadConfigFromFile(fileContent, targetKeys)

		convey.So(result, convey.ShouldNotBeEmpty)
	})
}

func TestCheckCurSuperPodConfigSwitch(t *testing.T) {
	convey.Convey("test CheckCurSuperPodConfigSwitch", t, func() {
		res := CheckCurSuperPodConfigSwitch(".")
		convey.So(res, convey.ShouldBeFalse)
		err := createTmpConfigFile()
		convey.So(err, convey.ShouldBeNil)
		defer removeTmpConfigFile()
		res = CheckCurSuperPodConfigSwitch(".")
		convey.So(res, convey.ShouldBeTrue)
	})
}

func createTmpConfigFile() error {
	configPath := "./cathelper.conf"
	fileContent := `
supperssedPeriod=0
networkType=1
pingType=0
pingTimes=5
pingInterval=1
period=10
netFault=on
`
	var fileMode0644 os.FileMode = 0644
	file, err := os.OpenFile(configPath, os.O_CREATE|os.O_RDWR, fileMode0644)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(fileContent)
	return err
}

func removeTmpConfigFile() {
	configPath := "./cathelper.conf"
	err := os.Remove(configPath)
	if err != nil {
		hwlog.RunLog.Errorf("remove temp config file %s failed: %v", configPath, err)
	}
}
