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

package flags

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/dynamic-resource-allocation/kubeletplugin"

	"ascend-common/api"
	"ascend-dynamic-resource-allocation/pkg/consts"
)

// DRAOption is the option struct for DRA plugin.
type DRAOption struct {
	// NodeName follows the NODE_NAME environment variable, set by the
	// deployment YAML via spec.nodeName fieldRef.
	NodeName                      string
	CdiRoot                       string
	KubeletRegistrarDirectoryPath string
	KubeletPluginsDirectoryPath   string
	DeviceResetTimeout            int
}

// RegisterFlags registers DRA options flags using standard library flag package
func (d *DRAOption) RegisterFlags() {
	d.NodeName = os.Getenv(consts.NodeNameEnv)

	flag.StringVar(&d.CdiRoot,
		"cdi-root",
		consts.DefaultCDIRoot,
		"Absolute path to the directory where CDI files will be generated.")

	flag.StringVar(&d.KubeletRegistrarDirectoryPath,
		"kubelet-registrar-directory-path",
		kubeletplugin.KubeletRegistryDir,
		"Absolute path to the directory where kubelet stores plugin registrations.")

	flag.StringVar(&d.KubeletPluginsDirectoryPath,
		"kubelet-plugins-directory-path",
		kubeletplugin.KubeletPluginsDir,
		"Absolute path to the directory where kubelet stores plugin data.")

	flag.IntVar(&d.DeviceResetTimeout, api.DeviceResetTimeout, api.DefaultDeviceResetTimeout,
		"when device-plugin starts, if the number of chips is insufficient, the maximum duration to wait for "+
			"the driver to report all chips, unit second, range [10, 600]")
}

// Validate ensures required directories exist (creating them if necessary).
func (d *DRAOption) Validate() error {
	if err := ensureDir(d.CdiRoot); err != nil {
		return fmt.Errorf("cdi-root path validate failed: %w", err)
	}

	if err := ensureDir(d.KubeletRegistrarDirectoryPath); err != nil {
		return fmt.Errorf("kubelet-registrar-directory-path validate failed: %w", err)
	}

	if err := ensureDir(d.KubeletPluginsDirectoryPath); err != nil {
		return fmt.Errorf("kubelet-plugins-directory-path validate failed: %w", err)
	}

	return nil
}

func ensureDir(path string) error {
	info, err := os.Stat(path)
	switch {
	case err != nil && os.IsNotExist(err):
		if err := os.MkdirAll(path, consts.DefaultDirMode); err != nil {
			return err
		}
	case err != nil:
		return err
	case !info.IsDir():
		return fmt.Errorf("path '%s' exists but is not a directory", path)
	}
	return nil
}

// DriverPluginPath returns the per-driver plugin data directory under the
// kubelet plugins directory.
func (d *DRAOption) DriverPluginPath() string {
	return filepath.Join(d.KubeletPluginsDirectoryPath, consts.DriverName)
}
