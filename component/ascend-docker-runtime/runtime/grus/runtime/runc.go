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

// Package runtime, runc implement of runtime
package runtime

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/containerd/go-runc"

	"ascend-common/common-utils/hwlog"
	"ascend-docker-runtime/runtime/common"
)

const (
	// RuncName runc runtime name
	RuncName = "runc"
)

// RuncRuntime is the runc implementation of RuntimeAPI
type RuncRuntime struct {
	client *runc.Runc
}

// NewRuntimeRunc creates a new runc runtime instance
func NewRuntimeRunc() RuntimeAPI {
	return &RuncRuntime{}
}

// Init initializes the runc client with the specified root path
func (r *RuncRuntime) Init(rootPath string) {
	r.client = &runc.Runc{Root: rootPath}
}

// Pause pauses the container with the specified ID
func (r *RuncRuntime) Pause(id string) error {
	if r.client == nil {
		return fmt.Errorf("runc client not init")
	}
	return r.client.Pause(context.Background(), id)
}

// Resume resumes the paused container with the specified ID
func (r *RuncRuntime) Resume(id string) error {
	if r.client == nil {
		return fmt.Errorf("runc client not init")
	}
	return r.client.Resume(context.Background(), id)
}

// State returns the state information of the container with the specified ID
func (r *RuncRuntime) State(id string) (*StateInfo, error) {
	if r.client == nil {
		return nil, fmt.Errorf("runc client not init")
	}
	con, err := r.client.State(context.Background(), id)
	if err != nil {
		return nil, err
	}
	return &StateInfo{ID: con.ID, Pid: con.Pid, Status: con.Status, Bundle: con.Bundle, Rootfs: con.Rootfs}, nil
}

// Checkpoint creates a checkpoint of the container at the specified path
func (r *RuncRuntime) Checkpoint(ckptPath, id string) error {
	if r.client == nil {
		return fmt.Errorf("runc client not init")
	}

	imagePath := filepath.Join(ckptPath, CRIU_IMG_DIR)
	workPath := filepath.Join(imagePath, "work")
	var actions []runc.CheckpointAction
	opts := &runc.CheckpointOpts{
		ImagePath:                imagePath,
		WorkDir:                  workPath,
		AllowOpenTCP:             false,
		AllowExternalUnixSockets: true,
		AllowTerminal:            false,
		FileLocks:                true,
	}
	actions = append(actions, runc.LeaveRunning)

	if err := os.MkdirAll(imagePath, common.COMMON_DIR_MODE); err != nil {
		hwlog.RunLog.Errorf("Create image path: %s, err: %v", imagePath, err)
		return err
	}
	if err := os.MkdirAll(workPath, common.COMMON_DIR_MODE); err != nil {
		hwlog.RunLog.Errorf("Create work path: %s, err: %v", workPath, err)
		return err
	}

	conIDFile := filepath.Join(ckptPath, CONTAINER_ID_FILE)
	if err := os.WriteFile(conIDFile, []byte(id), common.COMMON_FILE_MODE); err != nil {
		hwlog.RunLog.Errorf("Create container id file: %s, err: %v", id, err)
		return err
	}

	return r.client.Checkpoint(context.Background(), id, opts, actions...)
}

// Restore restores the container from a checkpoint at the specified path
func (r *RuncRuntime) Restore(ckptPath, id, ns string, externalEnvs []string) (err error) {
	imagePath := filepath.Join(ckptPath, CRIU_IMG_DIR)
	args := os.Args[1 : len(os.Args)-1]
	for i, targ := range args {
		if targ == "create" {
			args[i] = "restore"
		}
	}

	/*
		dir layout:
		    /var/log/ascend-docker-runtime/restore/
		        - ${ns}_${container1}/work
		        - ${ns}_${container2}/work
	*/
	workPath := fmt.Sprintf("%s/%s_%s/work", common.RESTORE_ROOT_DIR, ns, id)
	if err = os.MkdirAll(workPath, common.COMMON_DIR_MODE); err != nil {
		hwlog.RunLog.Errorf("create work dir failed: %v", err)
		return
	}

	args = append(args,
		"--detach",
		"--ext-unix-sk",
		"--file-locks",
		"--image-path", imagePath,
		"--work-path", workPath,
		id,
	)

	hwlog.RunLog.Infof("calling runc restore args: %v", args)
	cmd := exec.Command(RuncName, args...)
	cmd.Env = append(os.Environ(), externalEnvs...)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err = cmd.Run(); err != nil {
		hwlog.RunLog.Errorf("calling runc restore failed: %v", err)
		return
	}
	hwlog.RunLog.Infof("calling runc restore %s success", id)
	return nil
}
