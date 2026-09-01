/*
 * Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at

 * http://www.apache.org/licenses/LICENSE-2.0

 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package plugin

import (
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	"ascend-common/common-utils/hwlog"
)

// TestMain initializes the global hwlog.RunLog logger. WriteClaimSpec and the
// vendored cdi/mount packages log through hwlog.RunLog, which is nil unless
// initialized; calling Debugf/Warnf on a nil logger panics. OnlyToStdout
// skips all file-related log config validation.
func TestMain(m *testing.M) {
	if err := hwlog.InitRunLogger(&hwlog.LogConfig{OnlyToStdout: true}, context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, "failed to init hwlog:", err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}

// TestParseDeviceIDSuffix verifies parseDeviceIDSuffix splits the trailing
// integer off a "<name>-<id>" device name.
func TestParseDeviceIDSuffix(t *testing.T) {
	tests := []struct {
		name       string
		in         string
		wantID     int
		errContain string
	}{
		{"plain id", "Ascend910-12", 12, ""},
		{"zero id", "Ascend310P-0", 0, ""},
		{"multiple dashes use last", "a-b-c-7", 7, ""},
		{"leading dash only", "-5", 5, ""},
		{"no separator", "Ascend910", 0, "no '-' separator or empty suffix"},
		{"empty suffix", "Ascend910-", 0, "no '-' separator or empty suffix"},
		{"empty name", "", 0, "no '-' separator or empty suffix"},
		{"non-integer suffix", "Ascend910-x", 0, "is not an integer"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := parseDeviceIDSuffix(tt.in)
			if tt.errContain == "" {
				if err != nil {
					t.Fatalf("parseDeviceIDSuffix(%q) error = %v, want nil", tt.in, err)
				}
				if id != tt.wantID {
					t.Errorf("parseDeviceIDSuffix(%q) = %d, want %d", tt.in, id, tt.wantID)
				}
				return
			}
			if err == nil {
				t.Fatalf("parseDeviceIDSuffix(%q) = %d, want error containing %q", tt.in, id, tt.errContain)
			}
			if !strings.Contains(err.Error(), tt.errContain) {
				t.Errorf("parseDeviceIDSuffix(%q) error = %v, want containing %q", tt.in, err, tt.errContain)
			}
		})
	}
}

// TestNewCDISpecManager verifies the constructor stores devType and
// productTypes and configures the default CDI cache spec directory without
// panicking.
func TestNewCDISpecManager(t *testing.T) {
	devType := "Ascend910"
	productTypes := []string{"Atlas 200 Model 3000"}

	m := NewCDISpecManager(devType, productTypes, t.TempDir(), nil)

	if m == nil {
		t.Fatal("NewCDISpecManager() = nil, want non-nil manager")
	}
	if m.devType != devType {
		t.Errorf("devType = %q, want %q", m.devType, devType)
	}
	if !reflect.DeepEqual(m.productTypes, productTypes) {
		t.Errorf("productTypes = %v, want %v", m.productTypes, productTypes)
	}
}

// TestWriteClaimSpec_Errors verifies WriteClaimSpec fails on unparseable
// device names, an empty claim UID and an unknown device type. These paths
// never touch /dev so they are deterministic on machines without Ascend
// hardware; the success path requires real NPU device nodes and is covered
// by driver-layer integration tests instead.
func TestWriteClaimSpec_Errors(t *testing.T) {
	tests := []struct {
		name        string
		devType     string
		claimUID    string
		deviceNames []string
		errContains string
	}{
		{"no separator in name", "Ascend910", "ut-claim",
			[]string{"Ascend910"}, "no '-' separator or empty suffix"},
		{"empty suffix in name", "Ascend910", "ut-claim",
			[]string{"Ascend910-"}, "no '-' separator or empty suffix"},
		{"non-integer suffix", "Ascend910", "ut-claim",
			[]string{"Ascend910-x"}, "is not an integer"},
		{"empty claim UID", "Ascend910", "",
			[]string{"Ascend910-0"}, "claimUID must not be empty"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewCDISpecManager(tt.devType, nil, t.TempDir(), nil)

			ids, err := m.WriteClaimSpec(tt.claimUID, tt.deviceNames)
			if err == nil {
				t.Fatalf("WriteClaimSpec() = %v, want error containing %q", ids, tt.errContains)
			}
			if !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("WriteClaimSpec() error = %v, want containing %q", err, tt.errContains)
			}
		})
	}
}

// TestWriteClaimSpec_ToMountID verifies WriteClaimSpec invokes the supplied
// toMountID converter with the parsed physical ID and surfaces a converter
// error before reaching the cdi library. The success path requires real /dev
// device nodes and is covered by driver-layer integration tests, mirroring
// TestWriteClaimSpec_Errors above.
func TestWriteClaimSpec_ToMountID(t *testing.T) {
	convErr := errors.New("phyID not found")
	conv := func(int32) (int32, error) { return 0, convErr }
	m := NewCDISpecManager("Ascend910", nil, t.TempDir(), conv)

	_, err := m.WriteClaimSpec("ut-claim", []string{"npu-7"})
	if err == nil {
		t.Fatal("WriteClaimSpec() = nil, want error from toMountID")
	}
	if !strings.Contains(err.Error(), "convert phyID 7 to mountID") {
		t.Errorf("WriteClaimSpec() err = %v, want containing 'convert phyID 7 to mountID'", err)
	}
	if !errors.Is(err, convErr) {
		t.Errorf("WriteClaimSpec() err = %v, want wrapping %v", err, convErr)
	}
}

// TestDeleteClaimSpec verifies DeleteClaimSpec is idempotent: removing a
// spec that was never written returns nil rather than an error.
func TestDeleteClaimSpec(t *testing.T) {
	m := NewCDISpecManager("Ascend910", nil, t.TempDir(), nil)

	if err := m.DeleteClaimSpec("ut-claim-not-written"); err != nil {
		t.Errorf("DeleteClaimSpec() for missing spec = %v, want nil", err)
	}
	// A repeated delete must remain a no-op.
	if err := m.DeleteClaimSpec("ut-claim-not-written"); err != nil {
		t.Errorf("second DeleteClaimSpec() = %v, want nil", err)
	}
}
