/* Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.
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

package logger

import (
	"log"
	"testing"
)

// TestInitLogger covers that the logger can be initialized successfully.
func TestInitLogger(t *testing.T) {
	if err := InitLogger(); err != nil {
		t.Fatalf("InitLogger() error = %v", err)
	}
}

// TestLogFunctions covers that all log-level functions do not panic.
func TestLogFunctions(t *testing.T) {
	// Ensure logger is initialized
	_ = InitLogger()

	Debug("debug message")
	Info("info message")
	Warn("warn message")
	Error("error message")

	Debugf("debug %s", "fmt")
	Infof("info %s", "fmt")
	Warnf("warn %s", "fmt")
	Errorf("error %s", "fmt")
}

// TestSetStdLog covers that SetStdLog does not panic.
func TestSetStdLog(t *testing.T) {
	_ = InitLogger()
	SetStdLog()
	// Verify the standard log output was redirected by writing a line
	log.Println("test std log redirect")
}

// TestHwLogConfig_Defaults covers that the default config values are set.
func TestHwLogConfig_Defaults(t *testing.T) {
	if HwLogConfig.LogFileName != DefaultLogFile {
		t.Errorf("LogFileName = %s, want %s", HwLogConfig.LogFileName, DefaultLogFile)
	}
	if HwLogConfig.MaxLineLength != maxLogLineLength {
		t.Errorf("MaxLineLength = %d, want %d", HwLogConfig.MaxLineLength, maxLogLineLength)
	}
	if DefaultLogFile != "/var/log/mindx-dl/dpu-exporter/dpu-exporter.log" {
		t.Errorf("DefaultLogFile = %s, want /var/log/mindx-dl/dpu-exporter/dpu-exporter.log", DefaultLogFile)
	}
}
