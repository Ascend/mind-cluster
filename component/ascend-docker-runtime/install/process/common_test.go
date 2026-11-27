/* Copyright(C) 2025-2026. Huawei Technologies Co.,Ltd. All rights reserved.
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

package process

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/pelletier/go-toml"
	"github.com/smartystreets/goconvey/convey"

	"ascend-docker-runtime/mindxcheckutils"
)

// TestCheckParamAndGetBehavior test the function checkParamAndGetBehavior
func TestCheckParamAndGetBehavior(t *testing.T) {
	type args struct {
		action  string
		command []string
	}
	addCmds := []string{"0", "0", "0", "0", "0", "0"}
	rmCmds := []string{"0", "0", "0", "0", "0"}
	var tests = []struct {
		name  string
		args  args
		want  bool
		want1 string
	}{
		{
			name:  "01-add command,should return behavior install",
			args:  args{action: addCommand, command: addCmds},
			want:  true,
			want1: "install",
		}, {
			name:  "01-add command,should return behavior uninstall",
			args:  args{action: rmCommand, command: rmCmds},
			want:  true,
			want1: "uninstall",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := checkParamAndGetBehavior(tt.args.action, tt.args.command)
			if got != tt.want {
				t.Errorf("checkParamAndGetBehavior() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("checkParamAndGetBehavior() got = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestCheckParamLength(t *testing.T) {
	emptyStr := ""
	command := []string{"rm", oldJson, emptyStr, emptyStr, emptyStr}
	if got := CheckParamLength(command); !got {
		t.Errorf("CheckParamLength() got = %v, want %v", got, true)
	}
	command = []string{"rm", oldJson}
	if got := CheckParamLength(command); got {
		t.Errorf("CheckParamLength() got = %v, want %v", got, false)
	}
}

func TestParseAndValidateCommand(t *testing.T) {
	failure := errors.New("unsafe path")
	cases := []struct {
		name     string
		command  []string
		missing  bool
		failAt   int
		behavior string
		calls    []string
	}{
		{"empty", nil, false, 0, "", nil},
		{"unknown action", []string{"bad", "src", "dest", "runtime", "0", "0"}, false, 0, "", nil},
		{"short add", []string{"add"}, false, 0, "", nil},
		{"wrong add length", []string{"add", "src", "dest", "0", "0"}, false, 0, "", nil},
		{"wrong remove length", []string{"rm", "src", "dest", "runtime", "0", "0"}, false, 0, "", nil},
		{"add", []string{"add", "src", "dest", "runtime", "0", "0"}, false, 0, "install", []string{"file:src", "dir:.", "file:runtime"}},
		{"remove", []string{"rm", "src", "dest", "0", "0"}, false, 0, "uninstall", []string{"file:src", "dir:."}},
		{"missing source", []string{"add", "src", "dest", "runtime", "0", "0"}, true, 0, "install", []string{"dir:.", "dir:.", "file:runtime"}},
		{"unsafe source", []string{"add", "src", "dest", "runtime", "0", "0"}, false, 1, "install", []string{"file:src"}},
		{"unsafe source parent", []string{"add", "src", "dest", "runtime", "0", "0"}, true, 1, "install", []string{"dir:."}},
		{"unsafe destination parent", []string{"add", "src", "dest", "runtime", "0", "0"}, false, 2, "install", []string{"file:src", "dir:."}},
		{"unsafe runtime", []string{"add", "src", "dest", "runtime", "0", "0"}, false, 3, "install", []string{"file:src", "dir:.", "file:runtime"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var calls []string
			patches := gomonkey.NewPatches()
			defer patches.Reset()
			patches.ApplyFunc(os.Stat, func(string) (os.FileInfo, error) {
				if tc.missing {
					return nil, os.ErrNotExist
				}
				return nil, nil
			})
			check := func(kind, path string, parent, link bool) (string, error) {
				if !parent || link {
					t.Fatal("security flags changed")
				}
				calls = append(calls, kind+":"+path)
				if len(calls) == tc.failAt {
					return "", failure
				}
				return path, nil
			}
			patches.ApplyFunc(mindxcheckutils.RealDirChecker, func(path string, parent, link bool) (string, error) {
				return check("dir", path, parent, link)
			})
			patches.ApplyFunc(mindxcheckutils.RealFileChecker, func(path string, parent, link bool, size int) (string, error) {
				if size != mindxcheckutils.DefaultSize {
					t.Fatal("size limit changed")
				}
				return check("file", path, parent, link)
			})
			arg, behavior, err := parseAndValidateCommand(tc.command)
			wantError := tc.calls == nil || tc.failAt != 0
			if (err != nil) != wantError || behavior != tc.behavior || !reflect.DeepEqual(calls, tc.calls) {
				t.Fatalf("arg=%v behavior=%q err=%v calls=%v", arg, behavior, err, calls)
			}
			if tc.failAt != 0 && !errors.Is(err, failure) {
				t.Fatal("lost validation error")
			}
			if !wantError {
				want := &commandArgs{action: tc.command[0], srcFilePath: "src", destFilePath: "dest"}
				if want.action == addCommand {
					want.runtimeFilePath = "runtime"
				}
				if !reflect.DeepEqual(arg, want) {
					t.Fatalf("args=%v want=%v", arg, want)
				}
			}
		})
	}
}

func TestDeleteRuntimeAndEmptyParents(t *testing.T) {
	cases := []struct {
		name, input, expected string
	}{
		{"empty ancestors", "[root.runtimes.ascend]\npath='runtime'", ""},
		{"sibling runtime", "[root.runtimes.ascend]\npath='runtime'\n[root.runtimes.runc]\npath='runc'", "[root.runtimes.runc]\npath='runc'"},
		{"parent setting", "[root]\ndefault='runc'\n[root.runtimes.ascend]\npath='runtime'", "[root]\ndefault='runc'"},
		{"missing runtime", "[root.runtimes.runc]\npath='runc'", "[root.runtimes.runc]\npath='runc'"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tree, err := toml.Load(tc.input)
			if err != nil {
				t.Fatal(err)
			}
			if err := deleteRuntimeAndEmptyParents(tree, []string{"root", "runtimes", "ascend"}); err != nil {
				t.Fatal(err)
			}
			want, err := toml.Load(tc.expected)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(tree.ToMap(), want.ToMap()) {
				t.Fatalf("tree=%v want=%v", tree.ToMap(), want.ToMap())
			}
		})
	}
	if err := deleteRuntimeAndEmptyParents(nil, nil); err != nil {
		t.Fatal(err)
	}
}

func TestDeleteRuntimeErrors(t *testing.T) {
	for _, message := range []string{tomlPathNotFoundMessage, tomlNoSuchKeyMessage, "unexpected failure"} {
		t.Run(message, func(t *testing.T) {
			tree, err := toml.Load("")
			if err != nil {
				t.Fatal(err)
			}
			failure := errors.New(message)
			patches := gomonkey.ApplyMethodReturn(tree, "DeletePath", failure)
			defer patches.Reset()
			err = deleteRuntimeAndEmptyParents(tree, []string{"runtime"})
			if (err != nil) != (message == "unexpected failure") {
				t.Fatalf("unexpected deletion result: %v", err)
			}
		})
	}
}

func TestDeleteEmptyParentFailure(t *testing.T) {
	tree, err := toml.Load("[root.ascend]\npath='runtime'")
	if err != nil {
		t.Fatal(err)
	}
	failure := errors.New("parent deletion failed")
	originalDelete := (*toml.Tree).DeletePath
	patches := gomonkey.NewPatches()
	defer patches.Reset()
	patches.ApplyMethod(tree, "DeletePath", func(current *toml.Tree, path []string) error {
		if len(path) == 1 {
			return failure
		}
		var result error
		patches.Origin(func() { result = originalDelete(current, path) })
		return result
	})
	if err := deleteRuntimeAndEmptyParents(tree, []string{"root", "ascend"}); !errors.Is(err, failure) {
		t.Fatalf("lost parent deletion failure: %v", err)
	}
}

// secureConfigTestDir checks the full parent chain without changing shared directory permissions.
func secureConfigTestDir(t *testing.T) string {
	t.Helper()
	candidates := []string{os.TempDir()}
	if cwd, err := os.Getwd(); err == nil {
		candidates = append(candidates, cwd)
	}
	if home, err := os.UserHomeDir(); err == nil {
		candidates = append(candidates, home)
	}
	candidates = append(candidates, "/")
	for _, parent := range candidates {
		if _, err := mindxcheckutils.FileChecker(parent, true, true, false, 0); err != nil {
			t.Logf("unsafe test parent %q: %v", parent, err)
			continue
		}
		dir, err := os.MkdirTemp(parent, "ascend-config-test-")
		if err != nil {
			t.Logf("cannot create test directory in %q: %v", parent, err)
			continue
		}
		t.Cleanup(func() {
			if err := os.RemoveAll(dir); err != nil {
				t.Errorf("remove test directory %q: %v", dir, err)
			}
		})
		if _, err := mindxcheckutils.FileChecker(dir, true, true, false, 0); err != nil {
			t.Fatalf("created test directory is unsafe: %v", err)
		}
		return dir
	}
	t.Fatal("no safe writable test parent; set TMPDIR to a directory with a safe parent chain")
	return ""
}

func assertUnsafeMissingSourceParent(t *testing.T, entry func([]string) (string, error)) {
	t.Helper()
	patches := gomonkey.NewPatches()
	defer patches.Reset()
	patches.ApplyFunc(os.Stat, func(string) (os.FileInfo, error) { return nil, os.ErrNotExist })
	failure := errors.New("unsafe source parent")
	patches.ApplyFunc(mindxcheckutils.RealDirChecker, func(path string, parent, link bool) (string, error) {
		if path != filepath.Dir("src") || !parent || link {
			t.Fatal("incorrect parent validation")
		}
		return "", failure
	})
	behavior, err := entry([]string{"add", "src", "dest", "runtime", "0", "0"})
	if behavior != "install" || !errors.Is(err, failure) {
		t.Fatalf("behavior=%q err=%v", behavior, err)
	}
}

func assertUnsafeMissingSourceParentReal(t *testing.T, engine string, entry func([]string) (string, error)) {
	t.Helper()
	dir := secureConfigTestDir(t)
	unsafe := filepath.Join(dir, "unsafe")
	if err := os.Mkdir(unsafe, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(unsafe, 0777); err != nil {
		t.Fatal(err)
	}
	dest := filepath.Join(dir, "destination.conf")
	behavior, err := entry([]string{"add", filepath.Join(unsafe, "missing.conf"), dest, filepath.Join(dir, "unused-runtime"), "no", engine})
	if err == nil || behavior != "install" {
		t.Fatalf("unsafe source accepted: %q %v", behavior, err)
	}
	if !strings.Contains(err.Error(), "write permission not right "+unsafe+" ") {
		t.Fatalf("expected unsafe source parent rejection, got: %v", err)
	}
	if _, err := os.Stat(dest); !os.IsNotExist(err) {
		t.Fatalf("destination was touched: %v", err)
	}
}

// exerciseConfigRoundTrip uses real validation, parsing and persistence; each engine supplies its assertions.
func exerciseConfigRoundTrip(t *testing.T, engine, source string, entry func([]string) (string, error),
	assertRuntime func(*testing.T, string, string, bool)) {
	t.Helper()
	dir := secureConfigTestDir(t)
	src, dest, removed, bin := filepath.Join(dir, "source.conf"), filepath.Join(dir, "added.conf"), filepath.Join(dir, "removed.conf"), filepath.Join(dir, "runtime")
	for path, content := range map[string]string{src: source, bin: "test runtime"} {
		if err := os.WriteFile(path, []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
	}
	behavior, err := entry([]string{"add", src, dest, bin, "no", engine})
	if err != nil || behavior != "install" {
		t.Fatalf("add: behavior=%q error=%v", behavior, err)
	}
	assertRuntime(t, dest, bin, true)
	behavior, err = entry([]string{"rm", dest, removed, "no", engine})
	if err != nil || behavior != "uninstall" {
		t.Fatalf("remove: behavior=%q error=%v", behavior, err)
	}
	assertRuntime(t, removed, bin, false)
}

// TestWriteTomlConfigToFile tests the shared writer
func TestWriteTomlConfigToFile(t *testing.T) {
	convey.Convey("Test writeTomlConfigToFile", t, func() {
		convey.Convey("tree is nil, should return error", func() {
			err := writeTomlConfigToFile(nil, "test.toml")
			convey.So(err, convey.ShouldBeError)
		})
		convey.Convey("write a real tree, content persisted", func() {
			tree, _ := toml.Load(`runtime_type = "oci"`)
			tmpFile := filepath.Join(os.TempDir(), "ascend-crio-ut.conf")
			defer os.Remove(tmpFile)
			convey.So(writeTomlConfigToFile(tree, tmpFile), convey.ShouldBeNil)
			data, rErr := os.ReadFile(tmpFile)
			convey.So(rErr, convey.ShouldBeNil)
			convey.So(string(data), convey.ShouldContainSubstring, "oci")
		})
	})
}

// TestGetSubtreeByPath tests the function getSubtreeByPath
func TestGetSubtreeByPath(t *testing.T) {
	convey.Convey("Test getSubtreeByPath", t, func() {
		tree, _ := toml.Load(`
[plugins.cri.containerd.runtimes.runc]
runtime_type = "io.containerd.runc.v2"
`)
		convey.Convey("path exists", func() {
			keys := []string{pluginsKey, version1RuntimePluginName, containerdKey, runtimesKey, "runc"}
			subtree := getSubtreeByPath(keys, tree)
			convey.So(subtree, convey.ShouldNotBeNil)
		})
		convey.Convey("path not exists", func() {
			keys := []string{pluginsKey, version1RuntimePluginName, containerdKey, runtimesKey, "invalid"}
			subtree := getSubtreeByPath(keys, tree)
			convey.So(subtree, convey.ShouldBeNil)
		})
	})
}

// TestCopy tests the function copy
func TestCopy(t *testing.T) {
	convey.Convey("Test copy", t, func() {
		convey.Convey("tree is nil", func() {
			convey.So(copy(nil), convey.ShouldBeNil)
		})
		convey.Convey("tree is not nil", func() {
			tree, _ := toml.Load("key = 'value'")
			copyTree := copy(tree)
			convey.So(copyTree, convey.ShouldNotBeNil)
			convey.So(copyTree.Get("key"), convey.ShouldEqual, "value")
		})
	})
}
