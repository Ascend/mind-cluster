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

// Package snapshot for the run checkpoint binary test
package snapshot

import (
	"errors"
	"os"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"
)

// TestIsProcessAlive tests whether a process is alive
func TestIsProcessAlive(t *testing.T) {
	// Test whether the current process is alive
	currentPID := os.Getpid()
	if !isProcessAlive(currentPID) {
		t.Errorf("Current process should be alive")
	}
	// Test a non-existent process
	nonExistentPID := 999999
	if isProcessAlive(nonExistentPID) {
		t.Errorf("Non-existent process should not be alive")
	}
}

// TestRunCmd tests the RunCmd function
func TestRunCmd(t *testing.T) {
	// Test normal command execution
	cmd := "echo"
	args := []string{"hello"}
	env := os.Environ()
	exitCode, err := RunCmd(cmd, args, env, -1)
	if err != nil {
		t.Errorf("RunCmd failed: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
	// Test command execution with timeout
	// Use sleep command to simulate long-running operation
	sleepCmd := "sleep"
	sleepArgs := []string{"2"}
	timeout := 1 // 1 second timeout
	exitCode, err = RunCmd(sleepCmd, sleepArgs, env, timeout)
	if err == nil {
		t.Errorf("Expected timeout error, got nil")
	}
	if !errors.Is(err, ErrCommandTimeout) && !errors.Is(err, ErrCommandKillFailed) {
		t.Errorf("Expected timeout error, got %v", err)
	}
	// Test asynchronous execution (timeout=0)
	exitCode, err = RunCmd(cmd, args, env, 0)
	if err != nil {
		t.Errorf("RunCmd with timeout=0 failed: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 for async run, got %d", exitCode)
	}
}

// TestTerminate tests the terminate function
func TestTerminate(t *testing.T) {
	// Test case where terminate function succeeds
	patches := gomonkey.ApplyFunc(isProcessAlive, func(pid int) bool {
		return false
	})
	defer patches.Reset()
	err := terminate(999999999)
	assert.NoError(t, err)
	// Test case where terminate function fails
	patches = gomonkey.ApplyFunc(isProcessAlive, func(pid int) bool {
		return true
	})
	defer patches.Reset()
	err = terminate(999999999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "but process is still running")
}
