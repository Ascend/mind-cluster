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

// Package utils for common utils
package utils

import (
	"errors"
	"os"
	"testing"
)

func TestNormalizeVarRunPath(t *testing.T) {
	for _, tc := range []struct{ name, path, target, want string }{
		{"absolute alias", "/var/run/crio/crio.sock", "/run", "/run/crio/crio.sock"},
		{"relative alias", "/var/run/crio/crio.sock", "../run", "/run/crio/crio.sock"},
		{"other socket", "/var/run/example.sock", "/run", "/run/example.sock"},
		{"directory", "/var/run", "/run", "/run"},
		{"real directory", "/var/run/crio/crio.sock", "", "/var/run/crio/crio.sock"},
		{"untrusted alias", "/var/run/crio/crio.sock", "/tmp/run", "/var/run/crio/crio.sock"},
		{"prefix boundary", "/var/runtime/example.sock", "/run", "/var/runtime/example.sock"},
		{"already canonical", "/run/example.sock", "/run", "/run/example.sock"},
		{"relative path", "var/run/example.sock", "/run", "var/run/example.sock"},
		{"empty path", "", "/run", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := normalizeVarRunPath(tc.path, func(path string) (string, error) {
				if path != "/var/run" {
					t.Fatalf("unexpected link: %s", path)
				}
				if tc.target == "" {
					return "", errors.New("readlink failed")
				}
				return tc.target, nil
			})
			if got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestNormalizeVarRunPathHost(t *testing.T) {
	const path = "/var/run/example.sock"
	want := path
	target, err := os.Readlink("/var/run")
	if err == nil && (target == "/run" || target == "../run") {
		want = "/run/example.sock"
	}
	if got := NormalizeVarRunPath(path); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
