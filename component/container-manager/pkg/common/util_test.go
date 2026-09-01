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

// Package common test for utils.go
package common

import (
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"syscall"
	"testing"
	"time"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
)

func TestMain(m *testing.M) {
	if err := setup(); err != nil {
		return
	}
	code := m.Run()
	fmt.Printf("exit_code = %v\n", code)
}

func setup() error {
	if err := initLog(); err != nil {
		return err
	}
	return nil
}

func initLog() error {
	logConfig := &hwlog.LogConfig{
		OnlyToStdout: true,
	}
	if err := hwlog.InitRunLogger(logConfig, context.Background()); err != nil {
		fmt.Printf("init hwlog failed, %v\n", err)
		return errors.New("init hwlog failed")
	}
	return nil
}

func TestNewSignWatcher(t *testing.T) {
	tests := []struct {
		name     string
		osSigns  []os.Signal
		wantType chan os.Signal
	}{
		{
			name:     "no signals provided",
			osSigns:  []os.Signal{},
			wantType: make(chan os.Signal, 1),
		},
		{
			name:     "single signal",
			osSigns:  []os.Signal{syscall.SIGINT},
			wantType: make(chan os.Signal, 1),
		},
		{
			name:     "multiple signals",
			osSigns:  []os.Signal{syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP},
			wantType: make(chan os.Signal, 1),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewSignWatcher(tt.osSigns...)
			if got == nil {
				t.Error("NewSignWatcher() returned nil channel")
			}
			if cap(got) != cap(tt.wantType) {
				t.Errorf("NewSignWatcher() channel capacity = %d, want %d", cap(got), cap(tt.wantType))
			}
		})
	}
}

func TestNewSignWatcher2(t *testing.T) {
	t.Run("test concurrent safety", func(t *testing.T) {
		const (
			goroutines     = 10
			timoutDuration = 100 * time.Millisecond
		)
		done := make(chan bool, goroutines)
		for i := 0; i < goroutines; i++ {
			go func(id int) {
				defer func() {
					done <- true
				}()
				signChan := NewSignWatcher(syscall.SIGINT, syscall.SIGTERM)
				if signChan == nil {
					t.Errorf("Goroutine %d: NewSignWatcher() returned nil", id)
				}

				select {
				case signChan <- syscall.SIGINT: // send success
				case <-time.After(timoutDuration):
					t.Errorf("Goroutine %d: Channel is blocked", id)
				}
			}(i)
		}
		for i := 0; i < goroutines; i++ {
			<-done
		}
	})

	t.Run("test multiple signal types", func(t *testing.T) {
		const killSignalDuration = 50 * time.Millisecond
		signals := []os.Signal{
			syscall.SIGINT,
			syscall.SIGTERM,
			syscall.SIGHUP,
		}
		signChan := NewSignWatcher(signals...)
		defer close(signChan)
		testSignal := syscall.SIGINT
		go func() {
			time.Sleep(killSignalDuration)
			signChan <- testSignal
		}()
		select {
		case received := <-signChan:
			if received != testSignal {
				t.Errorf("Expected %v, got %v", testSignal, received)
			}
		case <-time.After(time.Second):
			t.Error("Timeout waiting for test signal")
		}
	})
}

// TestDeepCopy_BasicTypes tests deep copy of basic types
func TestDeepCopy_BasicTypes(t *testing.T) {
	tests := []struct {
		name     string
		src      interface{}
		expected interface{}
	}{
		{
			name:     "int type",
			src:      42,
			expected: 42,
		},
		{
			name:     "string type",
			src:      "hello world",
			expected: "hello world",
		},
		{
			name:     "bool type",
			src:      true,
			expected: true,
		},
		{
			name:     "float64 type",
			src:      3.14159,
			expected: 3.14159,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dst := reflect.New(reflect.TypeOf(tt.src)).Interface()
			if err := DeepCopy(dst, tt.src); err != nil {
				t.Errorf("DeepCopy() error = %v", err)
				return
			}
			actual := reflect.ValueOf(dst).Elem().Interface()
			if !reflect.DeepEqual(actual, tt.expected) {
				t.Errorf("DeepCopy() = %v, want %v", actual, tt.expected)
			}
		})
	}
}

// TestDeepCopy_Struct tests func DeepCopy of struct types
func TestDeepCopy_Struct(t *testing.T) {
	type Address struct {
		Street string
		City   string
	}
	type Person struct {
		Name    string
		Address Address
		Hobbies []string
	}
	src := Person{
		Name: "Alice",
		Address: Address{
			Street: "123 Main St",
			City:   "New York",
		},
		Hobbies: []string{"reading", "swimming"},
	}
	dst := Person{}
	if err := DeepCopy(&dst, src); err != nil {
		t.Errorf("DeepCopy() error = %v", err)
		return
	}
	if !reflect.DeepEqual(dst, src) {
		t.Errorf("DeepCopy() = %v, want %v", dst, src)
	}
	// Verification is a deep copy (modifying the original object should not affect the copied object)
	src.Name = "Bob"
	src.Hobbies[0] = "cooking"
	if dst.Name == src.Name {
		t.Error("DeepCopy() did not create a deep copy - Name field was affected")
	}
	if dst.Hobbies[0] == src.Hobbies[0] {
		t.Error("DeepCopy() did not create a deep copy - Hobbies slice was affected")
	}

	// test empty struct
	src = Person{}
	dst = Person{}
	if err := DeepCopy(&dst, src); err != nil {
		t.Errorf("DeepCopy() error = %v", err)
		return
	}
	if !reflect.DeepEqual(dst, src) {
		t.Errorf("DeepCopy() = %v, want %v", dst, src)
	}
}

// TestDeepCopy_NilSource tests behavior when source is nil
func TestDeepCopy_NilSource(t *testing.T) {
	var dst interface{}
	var src interface{} = nil
	if err := DeepCopy(&dst, src); err != nil {
		t.Errorf("DeepCopy() with nil source should not return error, got = %v", err)
	}
}

// TestDeepCopy_TypeMismatch tests behavior when destination type doesn't match source type
func TestDeepCopy_TypeMismatch(t *testing.T) {
	// src type: string, dst type: int
	src := "this is a string"
	var dst int
	if err := DeepCopy(&dst, src); err == nil {
		t.Error("DeepCopy() should return error for type mismatch")
	}
}

// TestDeepCopy_PointerToPointer tests deep copy of pointer to pointer
func TestDeepCopy_PointerToPointer(t *testing.T) {
	value := 100
	src := &value
	var dst *int
	if err := DeepCopy(&dst, src); err != nil {
		t.Errorf("DeepCopy() error = %v", err)
		return
	}
	if *dst != *src {
		t.Errorf("DeepCopy() = %v, want %v", *dst, *src)
	}
	// Verification refers to pointing to different memory addresses
	if dst == src {
		t.Error("DeepCopy() did not create new pointer")
	}
}

// TestDeepCopy_SliceAndMap tests deep copy of slices and maps
func TestDeepCopy_SliceAndMap(t *testing.T) {
	src := struct {
		Slice []int
		Map   map[string]int
	}{
		Slice: []int{1, 2, 3, 4, 5},
		Map:   map[string]int{"one": 1, "two": 2, "three": 3},
	}
	var dst struct {
		Slice []int
		Map   map[string]int
	}
	if err := DeepCopy(&dst, src); err != nil {
		t.Errorf("DeepCopy() error = %v", err)
		return
	}

	// Verify if the content is copied correctly
	if !reflect.DeepEqual(dst, src) {
		t.Errorf("DeepCopy() = %v, want %v", dst, src)
	}
	// Verification is a deep copy
	src.Slice[0] = 999
	src.Map["one"] = 999
	if dst.Slice[0] == 999 || dst.Map["one"] == 999 {
		t.Error("DeepCopy() did not create a deep copy - Slice or Map was affected")
	}
}

func TestGetDevStatus(t *testing.T) {
	tests := []struct {
		name     string
		faults   []*DevFaultInfo
		expected string
	}{
		{
			name: "01 contain NeedPauseCtrFaultLevels",
			faults: []*DevFaultInfo{
				{FaultLevel: NormalNPU},
				{FaultLevel: FreeRestartNPU},
			},
			expected: StatusNeedPause,
		},
		{
			name: "02 not contain NeedPauseCtrFaultLevels",
			faults: []*DevFaultInfo{
				{FaultLevel: NormalNPU},
				{FaultLevel: SeparateNPU},
			},
			expected: StatusIgnorePause,
		},
		{
			name:     "03 nil faults",
			faults:   []*DevFaultInfo{},
			expected: StatusIgnorePause,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetDevStatus(tt.faults)
			if result != tt.expected {
				t.Errorf("GetDevStatus() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetDevNumPerRing(t *testing.T) {
	for _, tt := range buildGetDevNumPerRingTestCase() {
		t.Run(tt.name, func(t *testing.T) {
			result := GetDevNumPerRing(tt.devType, tt.deviceNum, tt.boardId)
			if result != tt.expected {
				t.Errorf("GetDevNumPerRing() = %v, want %v", result, tt.expected)
			}
		})
	}
}

type getDevNumPerRingTC struct {
	name      string
	devType   string
	devUsage  string
	deviceNum int
	boardId   uint32
	expected  int
}

func buildGetDevNumPerRingTestCase() []getDevNumPerRingTC {
	return []getDevNumPerRingTC{
		{
			name:     "01 A1 device",
			devType:  api.Ascend910A,
			expected: Ascend910RingsNum,
		},
		{
			name:     "02 910B, no hccs",
			devType:  api.Ascend910B,
			boardId:  A800IA2NoneHccsBoardId,
			expected: NoRingNum,
		},
		{
			name:     "03 910B, has hccs",
			devType:  api.Ascend910B,
			boardId:  EmptyBoardId,
			expected: Ascend910BRingsNumTrain,
		},
		{
			name:     "04 910B+train",
			devType:  api.Ascend910B,
			expected: Ascend910BRingsNumTrain,
		},
		{
			name:      "05 A3 device, return device num",
			devType:   api.Ascend910A3,
			deviceNum: 16,
			expected:  Ascend910A3RingsNum,
		},
		{
			name:     "06 invalid devType",
			devType:  "unknown devType",
			expected: NoRingNum,
		},
		{
			name:      "07 910B+train+A200TA2",
			devType:   api.Ascend910B,
			deviceNum: 16,
			expected:  A200TA2RingsNum,
		},
		{
			name:     "08 A5 device, return device num",
			devType:  api.Ascend910A5,
			expected: Ascend910A5RingsNum,
		},
	}
}

func TestParseAddrs(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []string
		wantErr bool
	}{
		{
			name:  " string",
			input: "",
			want:  []string{},
		},
		{
			name:  "whitespace only",
			input: "   ",
			want:  []string{},
		},
		{
			name:  "multiple valid addrs",
			input: "1.2.3.4:8080,5.6.7.8:9090",
			want:  []string{"1.2.3.4:8080", "5.6.7.8:9090"},
		},
		{
			name:  "min port boundary",
			input: "1.2.3.4:1024",
			want:  []string{"1.2.3.4:1024"},
		},
		{
			name:    "invalid ip",
			input:   "abc:8080",
			wantErr: true,
		},
		{
			name:    "invalid port",
			input:   "1.2.3.4:abc",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseAddrs(tt.input)
			if tt.wantErr && err == nil {
				t.Errorf("ParseAddrs(%q) expected error, got none", tt.input)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("ParseAddrs(%q) = %v, want %v", tt.input, got, tt.want)
				}
			}
		})
	}
}

func TestGetNodeId(t *testing.T) {
	hostName, err := os.Hostname()
	if err != nil {
		t.Fatalf("os.Hostname() failed: %v", err)
	}
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "node id with whitespace",
			input: "  node-1  ",
			want:  "node-1",
		},
		{
			name:  "whitespace only returns hostname",
			input: "   ",
			want:  strings.TrimSpace(hostName),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetNodeId(tt.input)
			if err != nil {
				t.Errorf("GetNodeId(%q) unexpected error = %v", tt.input, err)
				return
			}
			if got != tt.want {
				t.Errorf("GetNodeId(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestGetListenAddr(t *testing.T) {
	tests := []struct {
		name string
		ip   string
		port string
		want string
	}{
		{
			name: "empty ip returns empty",
			ip:   "",
			port: "8080",
			want: "",
		},
		{
			name: "empty port uses default port",
			ip:   "1.2.3.4",
			port: "",
			want: "1.2.3.4:8890",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetListenAddr(tt.ip, tt.port); got != tt.want {
				t.Errorf("GetListenAddr(%q, %q) = %q, want %q", tt.ip, tt.port, got, tt.want)
			}
		})
	}
}
