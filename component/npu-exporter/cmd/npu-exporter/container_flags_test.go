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

package main

import (
	"context"
	"testing"

	"ascend-common/common-utils/hwlog"
	"huawei.com/npu-exporter/v6/collector/container"
)

func TestReadCntMonitoringFlags(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	if err := hwlog.InitRunLogger(&hwlog.LogConfig{OnlyToStdout: true}, ctx); err != nil {
		t.Fatal(err)
	}
	oldMode, oldCRI, oldOCI := containerMode, endpoint, containerd
	t.Cleanup(func() { containerMode, endpoint, containerd = oldMode, oldCRI, oldOCI })
	tests := []struct {
		mode, cri, oci       string
		kind                 int
		wantCRI, wantOCI     string
		backupCRI, backupOCI bool
	}{
		{"crio", "", "", container.EndpointTypeCrio, "unix:///var/run/crio/crio.sock", "", false, false},
		{"crio", "unix:///custom/crio.sock", "", container.EndpointTypeCrio, "unix:///custom/crio.sock", "", false, false},
		{"docker", "", "", container.EndpointTypeDockerd, container.DefaultDockerShim, container.DefaultDockerAddr, true, true},
		{"containerd", "", "", container.EndpointTypeContainerd, container.DefaultContainerdAddr, container.DefaultContainerdAddr, true, true},
		{"isula", "", "", container.EndpointTypeIsula, container.DefaultIsuladAddr, container.DefaultIsuladAddr, true, true},
		{"containerd", "unix:///custom/cri.sock", "unix:///custom/oci.sock", container.EndpointTypeContainerd, "unix:///custom/cri.sock", "unix:///custom/oci.sock", false, false},
		{"invalid", "", "", container.EndpointTypeDockerd, container.DefaultDockerShim, container.DefaultDockerAddr, true, true},
	}
	for _, tc := range tests {
		t.Run(tc.mode+tc.cri, func(t *testing.T) {
			containerMode, endpoint, containerd = tc.mode, tc.cri, tc.oci
			got := readCntMonitoringFlags()
			want := container.CntNpuMonitorOpts{EndpointType: tc.kind, CriEndpoint: tc.wantCRI, OciEndpoint: tc.wantOCI,
				UseCriBackup: tc.backupCRI, UseOciBackup: tc.backupOCI}
			if got != want {
				t.Fatalf("got %+v; want %+v", got, want)
			}
		})
	}
}

func TestCheckCntMonitoringFlags(t *testing.T) {
	oldMode, oldCRI, oldOCI := containerMode, endpoint, containerd
	t.Cleanup(func() { containerMode, endpoint, containerd = oldMode, oldCRI, oldOCI })
	const wantError = "-containerd is not supported when -containerMode=crio; use -endpoint instead"
	for _, tc := range []struct {
		name, mode, cri, oci string
		wantErr              bool
	}{
		{"crio default", "crio", "", "", false},
		{"crio explicit endpoint", "crio", "unix:///custom/crio.sock", "", false},
		{"crio incompatible containerd", "crio", "", "unix:///containerd.sock", true},
		{"crio both endpoints", "crio", "unix:///custom/crio.sock", "unix:///containerd.sock", true},
		{"crio invalid containerd address", "crio", "", "invalid", true},
		{"docker override", "docker", "", "unix:///containerd.sock", false},
		{"containerd override", "containerd", "", "unix:///containerd.sock", false},
		{"isula override", "isula", "", "unix:///isulad.sock", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			containerMode, endpoint, containerd = tc.mode, tc.cri, tc.oci
			err := checkCntMonitoringFlags()
			if (err != nil) != tc.wantErr {
				t.Fatalf("error=%v, wantErr=%v", err, tc.wantErr)
			}
			if err != nil && err.Error() != wantError {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
