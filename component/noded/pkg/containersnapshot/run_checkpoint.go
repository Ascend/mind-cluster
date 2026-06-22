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

// Package snapshot for the run checkpoint binary
package snapshot

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"syscall"
	"time"

	"ascend-common/common-utils/hwlog"
)

const (
	maxAttempts  = 5
	waitInterval = 1 * time.Second
	killWaitTime = 100 * time.Millisecond
)

var (
	// ErrCommandTimeout is returned when a command times out
	ErrCommandTimeout = errors.New("command timeout")
	// ErrCommandKillFailed is command timeout but failed to kill process group
	ErrCommandKillFailed = errors.New("command timeout and failed to kill process group")
)

func isProcessAlive(pid int) bool {
	return syscall.Kill(pid, 0) == nil
}

func terminate(pid int) error {
	// Wait for up to 5 seconds
	for i := 0; i < maxAttempts; i++ {
		if err := syscall.Kill(-pid, syscall.SIGTERM); err != nil {
			hwlog.RunLog.Errorf("send SIGTERM to pgid -%d failed: %v", pid, err)
		}
		time.Sleep(waitInterval)
		if !isProcessAlive(pid) {
			return nil // process gone
		}
	}

	// Force kill
	if err := syscall.Kill(-pid, syscall.SIGKILL); err != nil && err != syscall.ESRCH {
		hwlog.RunLog.Errorf("send SIGKILL to pgid -%d failed: %v", pid, err)
		return err
	}
	time.Sleep(killWaitTime)
	if !isProcessAlive(pid) {
		return nil // process gone
	}

	return fmt.Errorf("SIGKILL sent to process group %d, but process is still running", -pid)
}

// RunCmd handled timeout's three situations
func RunCmd(command string, args, env []string, timeout int) (int, error) {
	if timeout == 0 {
		go RunCmd(command, args, env, -1)
		return 0, nil
	}

	var cancel context.CancelFunc
	ctx := context.Background()

	if timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
		defer cancel()
	}

	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Env = env
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf
	if err := cmd.Start(); err != nil {
		return -1, fmt.Errorf("failed to start command: %v", err)
	}

	var waitErr error
	waitChan := make(chan struct{})

	go func() {
		defer close(waitChan)
		waitErr = cmd.Wait()
		if waitErr != nil {
			hwlog.RunLog.Errorf("wait command err: %v, command err: %v", waitErr, stderrBuf.String())
		}
	}()

	var err error
	select {
	case <-ctx.Done():
		if terminateErr := terminate(cmd.Process.Pid); terminateErr != nil {
			err = fmt.Errorf("%v: %v", ErrCommandKillFailed, terminateErr)
		} else {
			err = ErrCommandTimeout
		}
	case <-waitChan:
		err = waitErr
	}

	return cmd.ProcessState.ExitCode(), err
}
