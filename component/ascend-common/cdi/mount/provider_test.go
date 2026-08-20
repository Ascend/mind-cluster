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

// Package mount — provider_test.go
//
// Tests for Provider interface and Build.
package mount

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	cdispec "tags.cncf.io/container-device-interface/specs-go"
)

func TestBuild_ProviderError(t *testing.T) {
	_, err := Build(&errorProvider{err: errors.New("provider failure")}, "")
	if err == nil {
		t.Fatal("expected error from provider.GetMounts")
	}
}

type errorProvider struct{ err error }

func (p *errorProvider) GetMounts() ([]*cdispec.Mount, error) { return nil, p.err }

func TestBuild_HostRootPrefixesTopology(t *testing.T) {
	hostRoot := t.TempDir()
	topoFile := filepath.Join(hostRoot, "etc", "hccl_rootinfo.json")
	if err := os.MkdirAll(filepath.Dir(topoFile), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(topoFile, []byte(`{"version":"1.0"}`), 0644); err != nil {
		t.Fatal(err)
	}

	orig := TopologyItems
	defer func() { TopologyItems = orig }()
	TopologyItems = []TopologyItem{{HostPath: "/etc/hccl_rootinfo.json", Options: []string{"rbind", "rprivate", "ro"}}}

	// Empty provider (no .list mounts); only the topology item should be emitted.
	provider := &FileProvider{Dir: t.TempDir()}
	mounts, err := Build(provider, hostRoot)
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	if len(mounts) != 1 {
		t.Fatalf("expected 1 topology mount, got %d", len(mounts))
	}
	// HostPath must carry the original host path, without the hostRoot prefix.
	if mounts[0].HostPath != "/etc/hccl_rootinfo.json" {
		t.Errorf("HostPath = %q, want /etc/hccl_rootinfo.json", mounts[0].HostPath)
	}
}
