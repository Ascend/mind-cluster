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

// Package hwlog test file
package hwlog

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/common-utils/utils"
)

func TestNewLogger(t *testing.T) {
	convey.Convey("test api", t, func() {
		convey.Convey("test setLogger func", func() {
			lgConfig := &LogConfig{
				OnlyToStdout: true,
			}
			lg := new(logger)
			err := lg.setLogger(lgConfig)
			convey.So(err, convey.ShouldBeNil)
			// test for log file
			mockPathCheck := gomonkey.ApplyFunc(utils.CheckPath, func(_ string) (string, error) {
				return "", nil
			})
			mockMkdir := gomonkey.ApplyFunc(os.Chmod, func(_ string, _ fs.FileMode) error {
				return nil
			})
			defer mockPathCheck.Reset()
			defer mockMkdir.Reset()
			lgConfig = &LogConfig{
				LogFileName: path.Join(filepath.Dir(os.Args[0]), "t.log"),
				OnlyToFile:  true,
				MaxBackups:  DefaultMaxBackups,
				MaxAge:      DefaultMinSaveAge,
				CacheSize:   DefaultCacheSize,
				ExpiredTime: DefaultExpiredTime,
			}
			err = lg.setLogger(lgConfig)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestLoggerPrint(t *testing.T) {
	convey.Convey("test api", t, func() {
		convey.Convey("test logger print func", func() {
			lgConfig := &LogConfig{
				OnlyToStdout: true,
				LogLevel:     -1,
			}
			lg := new(logger)
			err := lg.setLogger(lgConfig)
			convey.So(err, convey.ShouldBeNil)
			lg.Debug("test debug")
			lg.Debugf("test debugf")
			lg.Info("test info")
			lg.Infof("test infof")
			lg.Warn("test warn")
			lg.Warnf("test warnf")
			lg.Error("test error")
			lg.Errorf("test errorf")
			lg.Critical("test critical")
			lg.Criticalf("test criticalf")
			lg.setLoggerLevel(maxLogLevel + 1)
			lg.Debug("test debug")
			lg.Debugf("test debugf")
			lg.Info("test info")
			lg.Infof("test infof")
			lg.Warn("test warn")
			lg.Warnf("test warnf")
			lg.Error("test error")
			lg.Errorf("test errorf")
			lg.Critical("test critical")
			lg.Criticalf("test criticalf")
		})
	})
}
func TestLoggerPrintWithLimit(t *testing.T) {
	convey.Convey("test api", t, func() {
		convey.Convey("test logger print func with limit", func() {
			lgConfig := &LogConfig{
				OnlyToStdout: true,
				LogLevel:     -1,
			}
			lg := new(logger)
			err := lg.setLogger(lgConfig)
			convey.So(err, convey.ShouldBeNil)
			domain := "hccs"
			logicId := 1

			errFormat := "collect failed ,err:%v"
			collectErr := fmt.Errorf("detail errs,logicId(%d)", logicId)
			lg.ErrorfWithLimit(domain, logicId, errFormat, collectErr)
			lg.ErrorfWithLimit(domain, logicId, errFormat, collectErr)
			lg.ErrorfWithLimit(domain, logicId, errFormat, collectErr)
			lg.ErrorfWithLimit(domain, logicId, errFormat, collectErr)
			ResetErrCnt(domain, logicId)
			lg.ErrorfWithLimit(domain, logicId, errFormat, collectErr)
			lg.ErrorfWithLimit(domain, logicId, errFormat, collectErr)
		})
	})
}

func TestWarnfWithLimit(t *testing.T) {
	convey.Convey("test api", t, func() {
		convey.Convey("test warn logger print func with limit", func() {
			lgConfig := &LogConfig{
				OnlyToStdout: true,
				LogLevel:     -1,
			}
			lg := new(logger)
			err := lg.setLogger(lgConfig)
			convey.So(err, convey.ShouldBeNil)
			domain := "hccs"
			logicId := 1

			errFormat := "collect failed ,err:%v"
			collectErr := fmt.Errorf("detail errs,logicId(%d)", logicId)
			lg.WarnfWithLimit(domain, logicId, errFormat, collectErr)
			lg.WarnfWithLimit(domain, logicId, errFormat, collectErr)
			lg.WarnfWithLimit(domain, logicId, errFormat, collectErr)
			lg.WarnfWithLimit(domain, logicId, errFormat, collectErr)
			ResetErrCnt(domain, logicId)
			lg.WarnfWithLimit(domain, logicId, errFormat, collectErr)
			lg.WarnfWithLimit(domain, logicId, errFormat, collectErr)
		})
	})
}

func TestValidate(t *testing.T) {
	convey.Convey("test api", t, func() {
		convey.Convey("test validate", func() {
			lg := new(logger)
			res := lg.validate()
			convey.So(res, convey.ShouldBeFalse)
			lgConfig := &LogConfig{
				OnlyToStdout: true,
			}
			err := lg.setLogger(lgConfig)
			convey.So(err, convey.ShouldBeNil)
			res = lg.validate()
			convey.So(res, convey.ShouldBeTrue)
		})
	})
}

// createTestLogFile creates a temporary log file and returns its path
func createTestLogFile(t *testing.T) string {
	// Create temporary log directory
	testLogsDir := "./logs"
	if _, err := os.Stat(testLogsDir); os.IsNotExist(err) {
		var fileMode os.FileMode = 0755
		if createFileErr := os.MkdirAll(testLogsDir, fileMode); createFileErr != nil {
			t.Fatalf("Failed to create log directory: %v", createFileErr)
		}
	}

	// Generate unique log file name
	logFileName := path.Join(testLogsDir, "test_log_"+time.Now().Format("20060102_150405")+".log")

	// Clean up temporary file after test
	defer func() {
		if err := os.Remove(logFileName); err != nil {
			t.Logf("Failed to clean up test log file: %v", err)
		}
	}()

	return logFileName
}

// setupLogger initializes a logger with the specified configuration
func setupLogger(t *testing.T, logFileName string) *logger {
	config := &LogConfig{
		LogFileName:   logFileName,
		LogLevel:      0, // INFO level
		MaxLineLength: 2048,
		FileMaxSize:   20,
		MaxBackups:    30,
		MaxAge:        7,
		OnlyToStdout:  false,
		OnlyToFile:    false,
	}

	hwlog := new(logger)
	if err := hwlog.setLogger(config); err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	return hwlog
}

// validateLogFormat checks if log lines match the expected format with microsecond precision and optional DST flag
func validateLogFormat(t *testing.T, logContent string, logLevel int) {
	logLines := strings.Split(logContent, "\n")
	timeFormatRegex := regexp.MustCompile(`^\[\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\.\d{6}(DST)?\]`)

	for _, line := range logLines {
		if line == "" {
			continue
		}

		// Validate time format with brackets
		if !timeFormatRegex.MatchString(line) {
			t.Errorf("Log format validation failed, line '%s' does not match expected time format with brackets", line)
		}

		// Validate log level filtering
		if strings.Contains(line, "[DEBUG]") && logLevel > 0 {
			t.Errorf("Log level filtering failed, DEBUG logs should not appear at INFO level: %s", line)
		}
	}
}

// validateLogContent checks if log content contains all expected log messages
func validateLogContent(t *testing.T, logContent string) {
	expectedMessages := []string{
		"This is an info log",
		"This is a warning log",
		"This is an error log",
		"This is a critical log",
	}

	for _, expected := range expectedMessages {
		if !strings.Contains(logContent, expected) {
			t.Errorf("Log content validation failed, expected message not found: %s", expected)
		}
	}
}

// TestLogFormatAndContentValidation tests the log format and content validation
func TestLogFormatAndContentValidation(t *testing.T) {
	// Create test log file
	logFileName := createTestLogFile(t)

	// Set up logger
	hwlog := setupLogger(t, logFileName)

	// Write test logs
	hwlog.Debug("This is a debug log")
	hwlog.Info("This is an info log")
	hwlog.Warn("This is a warning log")
	hwlog.Error("This is an error log")
	hwlog.Critical("This is a critical log")

	// Wait for logs to be written
	time.Sleep(1 * time.Second)

	// Read log content
	logContent, err := os.ReadFile(logFileName)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	// Validate log format and content
	validateLogFormat(t, string(logContent), 0) // 0 = INFO level
	validateLogContent(t, string(logContent))
}
