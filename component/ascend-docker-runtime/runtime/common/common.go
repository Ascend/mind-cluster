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

// Package common, function and type of grus
package common

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/opencontainers/runtime-spec/specs-go"

	"ascend-common/common-utils/hwlog"
	"ascend-docker-runtime/mindxcheckutils"
)

// Args for calling docker runtime
type Args struct {
	Bundle      string // Bundle path for the container
	Cmd         string // Command to execute
	ContainerID string // Container identifier
	CkptPath    string // Checkpoint path for restore operations
	Root        string // Root directory for the container
}

// ExecRunc exec runc with original args
var ExecRunc = func() error {
	tempRuncPath, err := exec.LookPath(DockerRuncName)
	if err != nil {
		tempRuncPath, err = exec.LookPath(RuncName)
		if err != nil {
			return fmt.Errorf("failed to find the path of runc: %v", err)
		}
	}
	runcPath, err := filepath.EvalSymlinks(tempRuncPath)
	if err != nil {
		return fmt.Errorf("failed to find realpath of runc %v", err)
	}
	if _, err := mindxcheckutils.RealFileChecker(runcPath, true, false, mindxcheckutils.DefaultSize); err != nil {
		return err
	}

	if err := mindxcheckutils.ChangeRuntimeLogMode("runtime-run-"); err != nil {
		return err
	}
	if err = syscall.Exec(runcPath, append([]string{runcPath}, os.Args[1:]...), os.Environ()); err != nil {
		return fmt.Errorf("failed to exec runc: %v", err)
	}

	return nil
}

// WriteSpecFile handles file write-back
func WriteSpecFile(path string, spec *specs.Spec) error {
	jsonFile, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		return fmt.Errorf("cannot reopen spec file %s: %v", path, err)
	}
	defer jsonFile.Close()
	if err = mindxcheckutils.CheckFileInfo(jsonFile, mindxcheckutils.DefaultSize); err != nil {
		hwlog.RunLog.Error(err)
		return err
	}
	jsonOutput, err := json.Marshal(spec)
	if err != nil {
		return fmt.Errorf("failed to marshal OCI spec file: %v", err)
	}

	if err = jsonFile.Truncate(0); err != nil {
		return fmt.Errorf("failed to truncate: %v", err)
	}
	if _, err = jsonFile.WriteAt(jsonOutput, 0); err != nil {
		return fmt.Errorf("failed to write OCI spec file: %v", err)
	}
	return nil
}
