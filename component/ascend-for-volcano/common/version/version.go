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

// Package version for version info
package version

import (
	"encoding/json"
	"runtime"
)

var (
	// Version semantic version number
	Version string
	// GitCommit git commit hash
	GitCommit string
	// GitBranch git branch name
	GitBranch string
	// BuildOS build operating system
	BuildOS string
	// BuildArch build architecture
	BuildArch string
	// GoVersion Go compiler version
	GoVersion string
)

// Get returns the complete version information injected at compile time.
// If some fields are not injected (empty) at compile time, fill with runtime defaults.
func Get() Info {
	info := Info{
		Version:   Version,
		GitCommit: GitCommit,
		GitBranch: GitBranch,
		BuildOS:   BuildOS,
		BuildArch: BuildArch,
		GoVersion: GoVersion,
	}

	if info.BuildOS == "" {
		info.BuildOS = runtime.GOOS
	}
	if info.BuildArch == "" {
		info.BuildArch = runtime.GOARCH
	}
	if info.GoVersion == "" {
		info.GoVersion = runtime.Version()
	}

	return info
}

// Info single component version information (without component name - identified by annotation key / configmap key)
type Info struct {
	Version   string `json:"version"`   // semantic version number
	GitCommit string `json:"gitCommit"` // git commit hash
	GitBranch string `json:"gitBranch"` // git branch name
	BuildOS   string `json:"buildOS"`   // build operating system
	BuildArch string `json:"buildArch"` // build architecture
	GoVersion string `json:"goVersion"` // Go version
}

// ToJSON serializes version information to JSON string
func ToJSON(i any) string {
	b, _ := json.Marshal(i)
	return string(b)
}

// VersionSummary daemonSet component version info summary
type VersionSummary struct {
	Type         string         `json:"type"`
	Versions     map[string]int `json:"versions"`
	TotalNodes   int            `json:"totalNodes"`
	QueryCommand string         `json:"queryCommand"`
}
