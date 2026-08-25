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

// Package mount — profile_test.go
//
// Tests for MountProfile, DefaultMountProfile, WriteMountProfile, and
// readProfileEntries.
package mount

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// mustWriteFile writes content to path, failing the test on error.
func mustWriteFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestDefaultMountProfile(t *testing.T) {
	cfg := DefaultMountProfile()
	if len(cfg[defaultGenerationKey]) != 1 {
		t.Fatalf("expected 1 default group, got %d", len(cfg[defaultGenerationKey]))
	}
	if len(cfg[ascend950Gen]) != 2 {
		t.Fatalf("expected 2 Ascend950 groups (ub + base/topo), got %d", len(cfg[ascend950Gen]))
	}
	for _, p := range []string{"/usr/lib64/libummu*", "/usr/lib64/liburma*", "/usr/lib64/libnl*"} {
		found := false
		for _, e := range cfg[ascend950Gen] {
			for _, ep := range e.Paths {
				if ep != p {
					continue
				}
				found = true
				if e.Type != ubType {
					t.Errorf("ub entry %q: Type = %q, want %q", p, e.Type, ubType)
				}
			}
		}
		if !found {
			t.Errorf("ub entry %q missing from Ascend950 mounts", p)
		}
	}
	for _, p := range []string{hcclRootInfoPath, topoDirPath} {
		found := false
		for _, e := range cfg[ascend950Gen] {
			for _, ep := range e.Paths {
				if ep != p {
					continue
				}
				found = true
				if e.Type == ubType {
					t.Errorf("topology entry %q: Type = %q, want empty", p, e.Type)
				}
			}
		}
		if !found {
			t.Errorf("topology entry %q missing from Ascend950 mounts", p)
		}
	}
}

func TestWriteMountProfile_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	cfg := MountProfile{
		defaultGenerationKey: []MountEntry{{Paths: []string{"/c1"}}, {Paths: []string{"/c2"}}},
		ascend950Gen:         []MountEntry{{Paths: []string{"/g1"}}, {Paths: []string{"/g2"}, Type: ubType}},
	}
	if err := WriteMountProfile(dir, cfg); err != nil {
		t.Fatalf("WriteMountProfile returned error: %v", err)
	}

	path := filepath.Join(dir, "mounts.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("mounts.json not readable: %v", err)
	}
	var got MountProfile
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("mounts.json is not valid JSON: %v", err)
	}
	if len(got) != len(cfg) {
		t.Errorf("map length = %d, want %d", len(got), len(cfg))
	}
	for key, wantEntries := range cfg {
		if len(got[key]) != len(wantEntries) {
			t.Errorf("key %q entry count = %d, want %d", key, len(got[key]), len(wantEntries))
		}
	}
	// The temp file must not survive the atomic rename.
	if _, err := os.Stat(path + ".tmp"); !os.IsNotExist(err) {
		t.Errorf("mounts.json.tmp should not exist after WriteMountProfile")
	}
}

func TestWriteMountProfile_CreatesDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "mounts")
	if err := WriteMountProfile(dir, DefaultMountProfile()); err != nil {
		t.Fatalf("WriteMountProfile returned error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "mounts.json")); err != nil {
		t.Fatalf("mounts.json missing after MkdirAll: %v", err)
	}
}

func TestReadProfileEntries_MissingFile(t *testing.T) {
	entries, ubIncluded, err := readProfileEntries(t.TempDir(), "Ascend910A5", false)
	if err != nil {
		t.Fatalf("missing mounts.json must not be an error, got: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(entries))
	}
	if ubIncluded {
		t.Error("ubIncluded = true, want false")
	}
}

// TestReadProfileEntries_GenerationLookup covers the generation-key lookup and
// UB filter: the DevType is mapped via generationKey (Ascend910A5→Ascend950)
// and unknown DevTypes fall back to the "default" entry.
func TestReadProfileEntries_GenerationLookup(t *testing.T) {
	dir := t.TempDir()
	cfg := MountProfile{
		defaultGenerationKey: []MountEntry{{Paths: []string{"/c1"}}},
		ascend950Gen:         []MountEntry{{Paths: []string{"/g1"}}, {Paths: []string{"/g2"}, Type: ubType}},
	}
	if err := WriteMountProfile(dir, cfg); err != nil {
		t.Fatal(err)
	}

	// Ascend910A5 → Ascend950 generation: full list (g1+g2), ubIncluded=true.
	entries, ubIncluded, err := readProfileEntries(dir, "Ascend910A5", false)
	if err != nil {
		t.Fatalf("readProfileEntries returned error: %v", err)
	}
	if !ubIncluded {
		t.Error("ubIncluded = false, want true")
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries (g1,g2), got %d: %v", len(entries), entries)
	}
	if entries[0].Paths[0] != "/g1" || entries[1].Paths[0] != "/g2" {
		t.Errorf("unexpected entries: %v", entries)
	}

	// disableUBMounts=true → the UB-marked entry is dropped.
	entries, ubIncluded, err = readProfileEntries(dir, "Ascend910A5", true)
	if err != nil {
		t.Fatalf("readProfileEntries returned error: %v", err)
	}
	if ubIncluded {
		t.Error("ubIncluded = true, want false when UB mounts disabled")
	}
	if len(entries) != 1 || entries[0].Paths[0] != "/g1" {
		t.Fatalf("expected 1 entry (g1), got %d: %v", len(entries), entries)
	}

	// Unknown devType → fallback to the "default" entry.
	entries, ubIncluded, err = readProfileEntries(dir, "Ascend310P", false)
	if err != nil {
		t.Fatalf("readProfileEntries returned error: %v", err)
	}
	if ubIncluded {
		t.Error("ubIncluded = true, want false for fallback entry")
	}
	if len(entries) != 1 || entries[0].Paths[0] != "/c1" {
		t.Fatalf("expected 1 fallback entry (/c1), got %d: %v", len(entries), entries)
	}
}

// TestGenerationKey covers the DevType → generation-key mapping: the Ascend
// 950 generation reports DevType "Ascend910A5" but is keyed "Ascend950";
// every other DevType (including empty) maps to itself.
func TestGenerationKey(t *testing.T) {
	tests := []struct{ input, want string }{
		{"Ascend910A5", ascend950Gen},
		{"Ascend910", "Ascend910"},
		{"", ""},
	}
	for _, tt := range tests {
		if got := generationKey(tt.input); got != tt.want {
			t.Errorf("generationKey(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestReadProfileEntries_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	mustWriteFile(t, filepath.Join(dir, "mounts.json"), "{not json")
	if _, _, err := readProfileEntries(dir, "Ascend910A5", false); err == nil {
		t.Fatal("expected error for invalid mounts.json")
	}
}
