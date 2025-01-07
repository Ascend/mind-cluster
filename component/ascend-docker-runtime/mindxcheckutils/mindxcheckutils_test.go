/* Copyright(C) 2021. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package mindxcheckutils
package mindxcheckutils

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/common-utils/hwlog"
)

const (
	fileMode0600 os.FileMode = 0600
)

var testError = errors.New("test")

func init() {
	ctx, _ := context.WithCancel(context.Background())
	logConfig := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	if err := hwlog.InitRunLogger(&logConfig, ctx); err != nil {
		fmt.Printf("hwlog init failed, error is %v", err)
	}
}

func TestNormalFileCheckRegularFile(t *testing.T) {
	tmpDir, filePath, err := createTestFile(t, "test_file.txt")
	defer removeTmpDir(t, tmpDir)
	err = os.Symlink(filePath, tmpDir+"/syslink")
	if err != nil {
		t.Fatalf("create symlink failed %q: %s", filePath, err)
	}

	if _, _, err = normalFileCheck(tmpDir, true, false); err != nil {
		t.Fatalf("check allow dir failed %q: %s", tmpDir+"/__test__", err)
	}

	if _, _, err = normalFileCheck(tmpDir, false, false); !strings.Contains(err.Error(), "not regular file") {
		t.Fatalf("check not allow dir failed %q: %s", tmpDir+"/__test__", err)
	}

	if _, _, err = normalFileCheck("/dev/zero", true, false); !strings.Contains(err.Error(), "not regular file/dir") {
		t.Fatalf("check /dev/zero failed %q: %s", tmpDir+"/__test__", err)
	}

	if _, _, err = normalFileCheck(tmpDir+"/syslink", false, false); !strings.Contains(err.Error(), "symlinks") {
		t.Fatalf("check symlinks failed %q: %s", tmpDir+"/syslink", err)
	}

	if _, _, err = normalFileCheck(filePath, false, false); err != nil {
		t.Fatalf("check failed %q: %s", filePath, err)
	}

	if _, _, err = normalFileCheck(tmpDir+"/notexisted", false, false); !strings.Contains(err.Error(), "not existed") {
		t.Fatalf("check symlinks failed %q: %s", tmpDir+"/syslink", err)
	}
}

func TestFileCheckRegularFile(t *testing.T) {
	tmpDir, filePath, err := createTestFile(t, "test_file.txt")
	defer removeTmpDir(t, tmpDir)
	err = os.Symlink(filePath, tmpDir+"/syslink")
	if err != nil {
		t.Fatalf("create symlink failed %q: %s", filePath, err)
	}

	if _, err = FileChecker(tmpDir, true, false, false, 0); err != nil {
		t.Fatalf("check allow dir failed %q: %s", tmpDir+"/__test__", err)
	}

	if _, err = FileChecker(tmpDir, false, false, false, 0); err != nil &&
		!strings.Contains(err.Error(), "not regular file") {
		t.Fatalf("check not allow dir failed %q: %s", tmpDir+"/__test__", err)
	}

	if _, err = FileChecker("/dev/zero", true, false, false, 0); err != nil &&
		!strings.Contains(err.Error(), "not regular file/dir") {
		t.Fatalf("check /dev/zero failed %q: %s", tmpDir+"/__test__", err)
	}
}

func TestGetLogPrefix(t *testing.T) {
	logPrefix = ""
	prefix, err := GetLogPrefix()
	if err != nil {
		t.Fatalf("get log prefix failed %v %v", prefix, err)
	}
	if logPrefix == "" || prefix != logPrefix {
		t.Fatalf("get log prefix failed 2 %v %v", prefix, prefix)
	}
}

// TestRealFileChecker test the function RealFileChecker
func TestRealFileChecker(t *testing.T) {
	tmpDir, filePath, err := createTestFile(t, "test_file.txt")
	if err != nil {
		t.Fatalf("create file failed %q: %s", filePath, err)
	}
	defer removeTmpDir(t, tmpDir)
	const permission os.FileMode = 0700
	err = os.WriteFile(filePath, []byte("hello\n"), permission)
	if err != nil {
		t.Fatalf("create file failed %q: %s", filePath, err)
	}
	if _, err = RealFileChecker(filePath, false, true, 0); err == nil {
		t.Fatalf("size check wrong 0 %q: %s", filePath, err)
	}
	if _, err = RealFileChecker(filePath, false, true, 1); err != nil {
		t.Fatalf("size check wrong 1 %q: %s", filePath, err)
	}
}

// TestRealFileChecker1 test the function RealFileChecker patch1
func TestRealFileChecker1(t *testing.T) {
	convey.Convey("test RealFileChecker patch1", t, func() {
		convey.Convey("01-string check fail, should return error", func() {
			patches := gomonkey.ApplyFuncReturn(StringChecker, false)
			defer patches.Reset()
			ret, err := RealFileChecker("", false, false, 0)
			convey.So(ret, convey.ShouldEqual, notValidPath)
			convey.So(err, convey.ShouldBeError)
		})
		convey.Convey("02-file check fail, should return error", func() {
			patches := gomonkey.ApplyFuncReturn(StringChecker, true).
				ApplyFuncReturn(FileChecker, false, testError)
			defer patches.Reset()
			ret, err := RealFileChecker("", false, false, 0)
			convey.So(ret, convey.ShouldEqual, notValidPath)
			convey.So(err, convey.ShouldBeError)
		})
		convey.Convey("03-get file absolute path fail, should return error", func() {
			patches := gomonkey.ApplyFuncReturn(StringChecker, true).
				ApplyFuncReturn(FileChecker, true, nil).
				ApplyFuncReturn(filepath.Abs, "", testError)
			defer patches.Reset()
			ret, err := RealFileChecker("", false, false, 0)
			convey.So(ret, convey.ShouldEqual, notValidPath)
			convey.So(err, convey.ShouldBeError)
		})
		convey.Convey("04-fail to get real path, should return error", func() {
			patches := gomonkey.ApplyFuncReturn(StringChecker, true).
				ApplyFuncReturn(FileChecker, true, nil).
				ApplyFuncReturn(filepath.Abs, "", nil).
				ApplyFuncReturn(filepath.EvalSymlinks, "", testError)
			defer patches.Reset()
			ret, err := RealFileChecker("", false, false, 0)
			convey.So(ret, convey.ShouldEqual, notValidPath)
			convey.So(err, convey.ShouldBeError)
		})
	})
}

// MockFileInfo mock os.FileInfo
type MockFileInfo struct {
	os.FileInfo
}

// IsDir mock the method IsDir, return false
func (MockFileInfo) IsDir() bool {
	return false
}

// TestRealFileChecker2 test the function RealFileChecker patch2
func TestRealFileChecker2(t *testing.T) {
	convey.Convey("test RealFileChecker patch2", t, func() {
		convey.Convey("05-fail to stat, should return error", func() {
			patches := gomonkey.ApplyFuncReturn(StringChecker, true).
				ApplyFuncReturn(FileChecker, true, nil).
				ApplyFuncReturn(filepath.Abs, "", nil).
				ApplyFuncReturn(filepath.EvalSymlinks, "", nil).
				ApplyFuncReturn(os.Stat, MockFileInfo{}, testError)
			defer patches.Reset()
			ret, err := RealFileChecker("", false, false, 0)
			convey.So(ret, convey.ShouldEqual, notValidPath)
			convey.So(err, convey.ShouldBeError)
		})
	})
}

// TestRealDirChecker test the function RealDirChecker
func TestRealDirChecker(t *testing.T) {
	tmpDir, filePath, err := createTestFile(t, "test_file.txt")
	if err != nil {
		t.Fatalf("create file failed %q: %s", filePath, err)
	}
	defer removeTmpDir(t, tmpDir)
	if _, err = RealDirChecker(filePath, false, true); err == nil {
		t.Fatalf("should be dir 0 %q: %s", filePath, err)
	}
	if _, err = RealDirChecker(tmpDir, false, true); err != nil {
		t.Fatalf("should be dir 1 %q: %s", filePath, err)
	}
}

// TestRealDirChecker1 test the function RealDirChecker patch1
func TestRealDirChecker1(t *testing.T) {
	convey.Convey("test RealDirChecker patch1", t, func() {
		convey.Convey("01-string check fail, should return error", func() {
			patches := gomonkey.ApplyFuncReturn(StringChecker, false)
			defer patches.Reset()
			ret, err := RealDirChecker("", false, false)
			convey.So(ret, convey.ShouldEqual, notValidPath)
			convey.So(err, convey.ShouldBeError)
		})
		convey.Convey("02-file check fail, should return error", func() {
			patches := gomonkey.ApplyFuncReturn(StringChecker, true).
				ApplyFuncReturn(FileChecker, false, testError)
			defer patches.Reset()
			ret, err := RealDirChecker("", false, false)
			convey.So(ret, convey.ShouldEqual, notValidPath)
			convey.So(err, convey.ShouldBeError)
		})
		convey.Convey("03-get file abs path fail, should return error", func() {
			patches := gomonkey.ApplyFuncReturn(StringChecker, true).
				ApplyFuncReturn(FileChecker, false, nil).
				ApplyFuncReturn(filepath.Abs, "", testError)
			defer patches.Reset()
			ret, err := RealDirChecker("", false, false)
			convey.So(ret, convey.ShouldEqual, notValidPath)
			convey.So(err, convey.ShouldBeError)
		})
		convey.Convey("04-get real file path fail, should return error", func() {
			patches := gomonkey.ApplyFuncReturn(StringChecker, true).
				ApplyFuncReturn(FileChecker, false, nil).
				ApplyFuncReturn(filepath.Abs, "", nil).
				ApplyFuncReturn(filepath.EvalSymlinks, "", testError)
			defer patches.Reset()
			ret, err := RealDirChecker("", false, false)
			convey.So(ret, convey.ShouldEqual, notValidPath)
			convey.So(err, convey.ShouldBeError)
		})
	})
}

// TestRealDirChecker2 test the function RealDirChecker patch2
func TestRealDirChecker2(t *testing.T) {
	convey.Convey("test RealDirChecker patch2", t, func() {
		convey.Convey("05-stat error, should return error", func() {
			patches := gomonkey.ApplyFuncReturn(StringChecker, true).
				ApplyFuncReturn(FileChecker, false, nil).
				ApplyFuncReturn(filepath.Abs, "", nil).
				ApplyFuncReturn(filepath.EvalSymlinks, "", nil).
				ApplyFuncReturn(os.Stat, MockFileInfo{}, testError)
			defer patches.Reset()
			ret, err := RealDirChecker("", false, false)
			convey.So(ret, convey.ShouldEqual, notValidPath)
			convey.So(err, convey.ShouldBeError)
		})
		convey.Convey("06-not dir, should return error", func() {
			patches := gomonkey.ApplyFuncReturn(StringChecker, true).
				ApplyFuncReturn(FileChecker, false, nil).
				ApplyFuncReturn(filepath.Abs, "", nil).
				ApplyFuncReturn(filepath.EvalSymlinks, "", nil).
				ApplyFuncReturn(os.Stat, MockFileInfo{}, nil)
			defer patches.Reset()
			ret, err := RealDirChecker("", false, false)
			convey.So(ret, convey.ShouldEqual, notValidPath)
			convey.So(err, convey.ShouldBeError)
		})
	})
}

func TestStringChecker(t *testing.T) {
	if ok := StringChecker("0123456789abcABC", 0, DefaultStringSize, ""); !ok {
		t.Fatalf("failed on regular letters")
	}
	const testSize = 3
	if ok := StringChecker("123", 0, testSize, ""); ok {
		t.Fatalf("failed on max length")
	}
	if ok := StringChecker("1234", 0, testSize, ""); ok {
		t.Fatalf("failed on max length")
	}
	if ok := StringChecker("12", 0, testSize, ""); !ok {
		t.Fatalf("failed on max length")
	}
	if ok := StringChecker("", 0, testSize, ""); ok {
		t.Fatalf("failed on min length")
	}
	if ok := StringChecker("123", testSize, DefaultStringSize, ""); ok {
		t.Fatalf("failed on min length")
	}
	if ok := StringChecker("123%", 0, DefaultStringSize, ""); ok {
		t.Fatalf("failed on strange words")
	}
	if ok := StringChecker("123.-/~", 0, DefaultStringSize, DefaultWhiteList); !ok {
		t.Fatalf("failed on strange words")
	}
}

func createTestFile(t *testing.T, fileName string) (string, string, error) {
	tmpDir := os.TempDir()
	const permission os.FileMode = 0700
	if os.MkdirAll(tmpDir+"/__test__", permission) != nil {
		t.Fatalf("MkdirAll failed %q", tmpDir+"/__test__")
	}
	f, err := os.Create(tmpDir + "/__test__" + fileName)
	if err != nil {
		t.Fatalf("create file failed %q: %s", tmpDir+"/__test__", err)
	}
	defer f.Close()
	err = f.Chmod(fileMode0600)
	if err != nil {
		t.Logf("chmod file error: %v", err)
	}
	return tmpDir + "/__test__", tmpDir + "/__test__" + fileName, err
}

func removeTmpDir(t *testing.T, tmpDir string) {
	if os.RemoveAll(tmpDir) != nil {
		t.Logf("removeall %v", tmpDir)
	}
}

// TestChangeRuntimeLogMode tests the function ChangeRuntimeLogMode
func TestChangeRuntimeLogMode(t *testing.T) {
	tests := []struct {
		name    string
		runLog  string
		wantErr bool
	}{
		{
			name:   "case 1",
			runLog: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ChangeRuntimeLogMode(tt.runLog); (err != nil) != tt.wantErr {
				t.Errorf("ChangeRuntimeLogMode() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
