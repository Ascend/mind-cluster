/* Copyright(C) 2024. Huawei Technologies Co.,Ltd. All rights reserved.
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

// package parser provides functions to parse device information

package parser

import (
	"context"
	"testing"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
)

const (
	deviceSliceLen2 = 2
	deviceSliceLen3 = 3
	deviceSliceLen4 = 4
	deviceSliceLen7 = 7
)

func init() {
	hwLogConfig := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	hwlog.InitRunLogger(&hwLogConfig, context.Background())
}

func TestParseAscendDeviceInfo_CommaStyle(t *testing.T) {
	env := api.AscendDeviceInfo + "=0,1,2"
	devices := ParseAscendDeviceInfo(env, "test-container")
	if len(devices) != deviceSliceLen3 {
		t.Errorf("expected 3 devices, got %d", len(devices))
	}
	if devices[0] != 0 || devices[1] != 1 || devices[deviceSliceLen2] != deviceSliceLen2 {
		t.Error("incorrect device IDs")
	}
}

func TestParseAscendDeviceInfo_MinusStyle(t *testing.T) {
	env := api.AscendDeviceInfo + "=0-3"
	devices := ParseAscendDeviceInfo(env, "test-container")
	if len(devices) != deviceSliceLen4 {
		t.Errorf("expected 4 devices, got %d", len(devices))
	}
}

func TestParseAscendDeviceInfo_AscendStyle(t *testing.T) {
	env := api.AscendDeviceInfo + "=Ascend910-0,Ascend910-1"
	devices := ParseAscendDeviceInfo(env, "test-container")
	if len(devices) != deviceSliceLen2 {
		t.Errorf("expected 2 devices, got %d", len(devices))
	}
	if devices[0] != 0 || devices[1] != 1 {
		t.Error("incorrect device IDs")
	}
}

func TestParseAscendDeviceInfo_CommaMinusStyle(t *testing.T) {
	env := api.AscendDeviceInfo + "=0-2,4,5-7"
	devices := ParseAscendDeviceInfo(env, "test-container")
	if len(devices) != deviceSliceLen7 {
		t.Errorf("expected 7 devices, got %d", len(devices))
	}
}

func TestParseAscendDeviceInfo_InvalidFormat(t *testing.T) {
	env := "InvalidFormat=0,1"
	devices := ParseAscendDeviceInfo(env, "test-container")
	if devices != nil {
		t.Error("expected nil for invalid format")
	}
}

func TestParseAscendDeviceInfo_TooLong(t *testing.T) {
	const longValueLen = 2000
	longValue := ""
	for i := 0; i < longValueLen; i++ {
		longValue += "0,"
	}
	env := api.AscendDeviceInfo + "=" + longValue
	devices := ParseAscendDeviceInfo(env, "test-container")
	if devices != nil {
		t.Error("expected nil for too long value")
	}
}

func TestParseAscendDeviceInfo_SingleDevice(t *testing.T) {
	env := api.AscendDeviceInfo + "=0"
	devices := ParseAscendDeviceInfo(env, "test-container")
	if len(devices) != 1 {
		t.Errorf("expected 1 device, got %d", len(devices))
	}
	if devices[0] != 0 {
		t.Error("incorrect device ID")
	}
}

func TestParseAscendDeviceInfo_EmptyValue(t *testing.T) {
	env := api.AscendDeviceInfo + "="
	devices := ParseAscendDeviceInfo(env, "test-container")
	if len(devices) != 0 {
		t.Errorf("expected 0 devices, got %d", len(devices))
	}
}

func TestParseAscendDeviceInfo_InvalidRange(t *testing.T) {
	env := api.AscendDeviceInfo + "=5-3"
	devices := ParseAscendDeviceInfo(env, "test-container")
	if len(devices) != 0 {
		t.Errorf("expected 0 devices for invalid range, got %d", len(devices))
	}
}

func TestParseAscendDeviceInfo_InvalidDeviceID(t *testing.T) {
	env := api.AscendDeviceInfo + "=invalid"
	devices := ParseAscendDeviceInfo(env, "test-container")
	if len(devices) != 0 {
		t.Errorf("expected 0 devices for invalid ID, got %d", len(devices))
	}
}

func TestParseAscendDeviceInfo_InvalidAscendFormat(t *testing.T) {
	env := api.AscendDeviceInfo + "=Ascend910"
	devices := ParseAscendDeviceInfo(env, "test-container")
	if len(devices) != 0 {
		t.Errorf("expected 0 devices for invalid Ascend format, got %d", len(devices))
	}
}

func TestParseDeviceIDs_CommaStyle(t *testing.T) {
	devices := parseDeviceIDs("0,1,2", "test-container")
	if len(devices) != deviceSliceLen3 {
		t.Errorf("expected 3 devices, got %d", len(devices))
	}
}

func TestParseDeviceIDs_MinusStyle(t *testing.T) {
	devices := parseDeviceIDs("0-3", "test-container")
	if len(devices) != deviceSliceLen4 {
		t.Errorf("expected 4 devices, got %d", len(devices))
	}
}

func TestParseDeviceIDs_AscendStyle(t *testing.T) {
	devices := parseDeviceIDs("Ascend910-0,Ascend910-1", "test-container")
	if len(devices) != deviceSliceLen2 {
		t.Errorf("expected 2 devices, got %d", len(devices))
	}
}

func TestParseDeviceIDs_CommaMinusStyle(t *testing.T) {
	devices := parseDeviceIDs("0-2,4,5-7", "test-container")
	if len(devices) != deviceSliceLen7 {
		t.Errorf("expected 7 devices, got %d", len(devices))
	}
}

func TestParseCommaStyle_Simple(t *testing.T) {
	devices := parseCommaStyle("0,1,2", "test-container")
	if len(devices) != deviceSliceLen3 {
		t.Errorf("expected 3 devices, got %d", len(devices))
	}
}

func TestParseCommaStyle_WithSpaces(t *testing.T) {
	devices := parseCommaStyle("0, 1, 2", "test-container")
	if len(devices) != deviceSliceLen3 {
		t.Errorf("expected 3 devices, got %d", len(devices))
	}
}

func TestParseCommaStyle_Single(t *testing.T) {
	devices := parseCommaStyle("0", "test-container")
	if len(devices) != 1 {
		t.Errorf("expected 1 device, got %d", len(devices))
	}
}

func TestParseCommaStyle_Empty(t *testing.T) {
	devices := parseCommaStyle("", "test-container")
	if len(devices) != 0 {
		t.Errorf("expected 0 devices, got %d", len(devices))
	}
}

func TestParseMinusStyle_Simple(t *testing.T) {
	devices := parseMinusStyle("0-3", "test-container")
	if len(devices) != deviceSliceLen4 {
		t.Errorf("expected 4 devices, got %d", len(devices))
	}
}

func TestParseMinusStyle_Single(t *testing.T) {
	devices := parseMinusStyle("0-0", "test-container")
	if len(devices) != 1 {
		t.Errorf("expected 1 device, got %d", len(devices))
	}
}

func TestParseMinusStyle_InvalidRange(t *testing.T) {
	devices := parseMinusStyle("5-3", "test-container")
	if len(devices) != 0 {
		t.Errorf("expected 0 devices, got %d", len(devices))
	}
}

func TestParseMinusStyle_InvalidFormat(t *testing.T) {
	devices := parseMinusStyle("invalid", "test-container")
	if len(devices) != 0 {
		t.Errorf("expected 0 devices, got %d", len(devices))
	}
}

func TestParseCommaMinusStyle_Simple(t *testing.T) {
	devices := parseCommaMinusStyle("0-2,4,5-7", "test-container")
	if len(devices) != deviceSliceLen7 {
		t.Errorf("expected 7 devices, got %d", len(devices))
	}
}

func TestParseCommaMinusStyle_SingleRange(t *testing.T) {
	devices := parseCommaMinusStyle("0-3", "test-container")
	if len(devices) != deviceSliceLen4 {
		t.Errorf("expected 4 devices, got %d", len(devices))
	}
}

func TestParseCommaMinusStyle_SingleValue(t *testing.T) {
	devices := parseCommaMinusStyle("0", "test-container")
	if len(devices) != 1 {
		t.Errorf("expected 1 device, got %d", len(devices))
	}
}

func TestParseAscendStyle_Simple(t *testing.T) {
	devices := parseAscendStyle("Ascend910-0,Ascend910-1", "test-container")
	if len(devices) != deviceSliceLen2 {
		t.Errorf("expected 2 devices, got %d", len(devices))
	}
}

func TestParseAscendStyle_Single(t *testing.T) {
	devices := parseAscendStyle("Ascend910-0", "test-container")
	if len(devices) != 1 {
		t.Errorf("expected 1 device, got %d", len(devices))
	}
}

func TestParseAscendStyle_InvalidFormat(t *testing.T) {
	devices := parseAscendStyle("Ascend910", "test-container")
	if len(devices) != 0 {
		t.Errorf("expected 0 devices, got %d", len(devices))
	}
}

func TestParseAscendStyle_InvalidDeviceID(t *testing.T) {
	devices := parseAscendStyle("Ascend910-invalid", "test-container")
	if len(devices) != 0 {
		t.Errorf("expected 0 devices, got %d", len(devices))
	}
}
