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
	"syscall"
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

func TestCheckPath01(t *testing.T) {
	convey.Convey("test CheckPath", t, func() {
		convey.Convey("01-string check fail, should return error", func() {
			patches := gomonkey.ApplyFuncReturn(StringChecker, false)
			defer patches.Reset()
			err := CheckPath("", false)
			convey.So(err, convey.ShouldBeError)
			convey.So(err.Error(), convey.ShouldEqual, "invalid path")
		})

		convey.Convey("02-path contains .., should return error", func() {
			patches := gomonkey.ApplyFuncReturn(StringChecker, true)
			defer patches.Reset()
			err := CheckPath("/some/../path", false)
			convey.So(err, convey.ShouldBeError)
			convey.So(err.Error(), convey.ShouldContainSubstring, "err path")
		})

		convey.Convey("03-get abs path error, should return error", func() {
			patches := gomonkey.ApplyFuncReturn(StringChecker, true).
				ApplyFuncReturn(filepath.Abs, "", testError)
			defer patches.Reset()
			err := CheckPath("/valid/path", false)
			convey.So(err, convey.ShouldBeError)
			convey.So(err.Error(), convey.ShouldContainSubstring, "get abs path failed")
		})

		convey.Convey("04-path too long, should return error", func() {
			patches := gomonkey.ApplyFuncReturn(StringChecker, true).
				ApplyFuncReturn(filepath.Abs, "/valid/path", nil).
				ApplyFuncReturn(filepath.Base, strings.Repeat("a", DefaultStringSize+1))
			defer patches.Reset()
			err := CheckPath("/valid/path", false)
			convey.So(err, convey.ShouldBeError)
			convey.So(err.Error(), convey.ShouldContainSubstring, "path too long")
		})

		convey.Convey("05-eval symlinks error, should return error", func() {
			patches := gomonkey.ApplyFuncReturn(StringChecker, true).
				ApplyFuncReturn(filepath.Abs, "/valid/path", nil).
				ApplyFuncReturn(filepath.Base, "path").
				ApplyFuncReturn(filepath.EvalSymlinks, "", testError)
			defer patches.Reset()
			err := CheckPath("/valid/path", false)
			convey.So(err, convey.ShouldBeError)
			convey.So(err.Error(), convey.ShouldContainSubstring, "symlinks or not existed")
		})
	})
}

func TestCheckPath02(t *testing.T) {
	convey.Convey("test CheckPath", t, func() {
		convey.Convey("06-symlink not allowed but path is symlink, should return error", func() {
			patches := gomonkey.ApplyFuncReturn(StringChecker, true).
				ApplyFuncReturn(filepath.Abs, "/valid/path", nil).
				ApplyFuncReturn(filepath.Base, "path").
				ApplyFuncReturn(filepath.EvalSymlinks, "/different/path", nil)
			defer patches.Reset()
			err := CheckPath("/valid/path", false)
			convey.So(err, convey.ShouldBeError)
			convey.So(err.Error(), convey.ShouldContainSubstring, "symlinks or not existed")
		})

		convey.Convey("07-symlink not allowed and path is not symlink, should return nil", func() {
			patches := gomonkey.ApplyFuncReturn(StringChecker, true).
				ApplyFuncReturn(filepath.Abs, "/valid/path", nil).
				ApplyFuncReturn(filepath.Base, "path").
				ApplyFuncReturn(filepath.EvalSymlinks, "/valid/path", nil)
			defer patches.Reset()
			err := CheckPath("/valid/path", false)
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("08-symlink allowed and path is symlink, should return nil", func() {
			patches := gomonkey.ApplyFuncReturn(StringChecker, true).
				ApplyFuncReturn(filepath.Abs, "/symlink/path", nil).
				ApplyFuncReturn(filepath.Base, "path").
				ApplyFuncReturn(filepath.EvalSymlinks, "/real/path", nil)
			defer patches.Reset()
			err := CheckPath("/symlink/path", true)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestCheckFileInfo01(t *testing.T) {
	convey.Convey("test CheckFileInfo", t, func() {
		tmpDir, filePath, err := createTestFile(t, "test_file.txt")
		if err != nil {
			t.Fatalf("create file failed %q: %s", filePath, err)
		}
		defer removeTmpDir(t, tmpDir)
		file, err := os.Open(filePath)
		convey.So(err, convey.ShouldBeNil)

		dir, err := os.Open(tmpDir)
		convey.So(err, convey.ShouldBeNil)

		defer file.Close()
		const validSize, invalidSize, tooBigSize = 10, -1, 1000000
		convey.Convey("01-get file stat error, should return error", func() {
			patch := gomonkey.ApplyMethodReturn(new(os.File), "Stat", nil, errors.New("stat error"))
			defer patch.Reset()
			err := CheckFileInfo(file, validSize)
			convey.So(err, convey.ShouldBeError)
			convey.So(err.Error(), convey.ShouldEqual, "invalid file")
		})

		convey.Convey("02-not regular file, should return error", func() {
			err := CheckFileInfo(dir, validSize)
			convey.So(err, convey.ShouldBeError)
			convey.So(err.Error(), convey.ShouldEqual, "invalid regular file")
		})

		convey.Convey("03-size too large, should return error", func() {
			err := CheckFileInfo(file, invalidSize)
			convey.So(err, convey.ShouldBeError)
			convey.So(err.Error(), convey.ShouldEqual, "invalid size")
		})

		convey.Convey("05-size too large, should return error", func() {
			err := CheckFileInfo(file, tooBigSize)
			convey.So(err, convey.ShouldBeError)
			convey.So(err.Error(), convey.ShouldEqual, "invalid size")
		})
	})
}

// TestCheckFileInfo 测试 CheckFileInfo 函数的各种情况
func TestCheckFileInfo02(t *testing.T) {
	convey.Convey("test CheckFileInfo", t, func() {
		tmpDir, filePath, err := createTestFile(t, "test_file.txt")
		if err != nil {
			t.Fatalf("create file failed %q: %s", filePath, err)
		}
		defer removeTmpDir(t, tmpDir)
		file, err := os.Open(filePath)
		convey.So(err, convey.ShouldBeNil)
		defer func() {
			err := file.Close()
			convey.So(err, convey.ShouldBeNil)
		}()
		stat, err := file.Stat()
		convey.So(err, convey.ShouldBeNil)
		convey.Convey("06-permission length not right, should return error", func() {
			patches := gomonkey.ApplyMethodReturn(stat.Mode().Perm(), "String", "wrong_length_perm")
			defer patches.Reset()
			err := CheckFileInfo(file, 1)
			convey.So(err, convey.ShouldBeError)
			convey.So(err.Error(), convey.ShouldContainSubstring, "permission not right")
		})
		convey.Convey("07-write permission not right for group, should return error", func() {
			patches := gomonkey.ApplyMethodReturn(stat.Mode().Perm(), "String", "-rw-rw-r--")
			defer patches.Reset()
			err := CheckFileInfo(file, 1)
			convey.So(err, convey.ShouldBeError)
			convey.So(err.Error(), convey.ShouldContainSubstring, "write permission not right")
		})
		convey.Convey("08-write permission not right for other, should return error", func() {
			patches := gomonkey.ApplyMethodReturn(stat.Mode().Perm(), "String", "-rw-r--rw-")
			defer patches.Reset()
			err := CheckFileInfo(file, 1)
			convey.So(err, convey.ShouldBeError)
			convey.So(err.Error(), convey.ShouldContainSubstring, "write permission not right")
		})
		convey.Convey("09-can not get stat, should return error", func() {
			patches := gomonkey.ApplyMethodReturn(stat, "Sys", "not_stat_t")
			defer patches.Reset()
			err := CheckFileInfo(file, 1)
			convey.So(err, convey.ShouldBeError)
			convey.So(err.Error(), convey.ShouldContainSubstring, "can not get stat")
		})
	})
}

func TestCheckFileInfo03(t *testing.T) {
	convey.Convey("test CheckFileInfo", t, func() {
		tmpDir, filePath, err := createTestFile(t, "test_file.txt")
		if err != nil {
			t.Fatalf("create file failed %q: %s", filePath, err)
		}
		defer removeTmpDir(t, tmpDir)
		file, err := os.Open(filePath)
		convey.So(err, convey.ShouldBeNil)

		defer func() {
			err := file.Close()
			convey.So(err, convey.ShouldBeNil)
		}()

		const expectedUid, actualUid = 9999, 1000
		stat, err := file.Stat()
		convey.So(err, convey.ShouldBeNil)
		convey.Convey("10-owner not right, should return error", func() {
			mockStat := &syscall.Stat_t{Uid: expectedUid}
			patches := gomonkey.ApplyMethodReturn(stat, "Sys", mockStat).
				ApplyFuncReturn(os.Getuid, actualUid)
			defer patches.Reset()
			err := CheckFileInfo(file, 1)
			convey.So(err, convey.ShouldBeError)
			convey.So(err.Error(), convey.ShouldContainSubstring, "owner not right")
		})
		convey.Convey("11-setuid not allowed, should return error", func() {
			patches := gomonkey.ApplyMethodReturn(stat, "Mode", os.ModeSetuid)
			defer patches.Reset()
			err := CheckFileInfo(file, 1)
			convey.So(err, convey.ShouldBeError)
			convey.So(err.Error(), convey.ShouldContainSubstring, "setuid not allowed")
		})
		convey.Convey("12-setgid not allowed, should return error", func() {
			patches := gomonkey.ApplyMethodReturn(stat, "Mode", os.ModeSetgid)
			defer patches.Reset()
			err := CheckFileInfo(file, 1)
			convey.So(err, convey.ShouldBeError)
			convey.So(err.Error(), convey.ShouldContainSubstring, "setgid not allowed")
		})
		convey.Convey("13-valid file, should return nil", func() {
			mockStat := &syscall.Stat_t{Uid: 0}
			patches := gomonkey.ApplyMethodReturn(stat, "Sys", mockStat).
				ApplyMethodReturn(stat.Mode().Perm(), "String", "-rw-r--r--").
				ApplyMethodReturn(stat, "Mode", os.FileMode(0))
			defer patches.Reset()
			err := CheckFileInfo(file, 1)
			convey.So(err, convey.ShouldBeNil)
		})
	})
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
	convey.Convey("01-logPrefix not empty, should return nil", t, func() {
		logPrefix = "test"
		defer func() {
			logPrefix = ""
		}()
		_, err := GetLogPrefix()
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("02-uid less than 0, should return error", t, func() {
		patch := gomonkey.ApplyFuncReturn(os.Geteuid, -1)
		defer patch.Reset()
		_, err := GetLogPrefix()
		convey.So(err, convey.ShouldBeError)
	})
	convey.Convey("03-EvalSymlinks error, should return error", t, func() {
		patch := gomonkey.ApplyFuncReturn(os.Geteuid, 1).
			ApplyFuncReturn(filepath.EvalSymlinks, "", testError)
		defer patch.Reset()
		ret, _ := GetLogPrefix()
		convey.So(strings.Contains(ret, "unknown"), convey.ShouldBeTrue)
	})
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

// Mode mock the method Mode
func (MockFileInfo) Mode() os.FileMode {
	return os.ModePerm
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
	convey.Convey("Walk error, should return nil", t, func() {
		patch := gomonkey.ApplyFuncReturn(filepath.Walk, testError)
		defer patch.Reset()
		convey.So(ChangeRuntimeLogMode(""), convey.ShouldBeNil)
	})
}

// TestFileChecker tests the function FileChecker
func TestFileChecker(t *testing.T) {
	convey.Convey("test FileChecker", t, func() {
		const maxDepth, groupWriteIndex, otherWriteIndex, permLength int = 99, 5, 8, 10
		convey.Convey("01-deep over maxDepth, should return error", func() {
			_, err := FileChecker("", false, false, false, maxDepth+1)
			convey.So(err, convey.ShouldBeError)
		})
		convey.Convey("02-contains .., should return error", func() {
			_, err := FileChecker("..", false, false, false, 1)
			convey.So(err, convey.ShouldBeError)
		})
		convey.Convey("03-get abs path error, should return error", func() {
			patch := gomonkey.ApplyFuncReturn(filepath.Abs, "", testError)
			defer patch.Reset()
			_, err := FileChecker("", false, false, false, 1)
			convey.So(err, convey.ShouldBeError)
		})
		patches := gomonkey.ApplyFuncReturn(filepath.Abs, "", nil)
		defer patches.Reset()
		convey.Convey("04-over DefaultStringSize, should return error", func() {
			const strLen = 257
			patch := gomonkey.ApplyFuncReturn(filepath.Base, strings.Repeat("1", strLen))
			defer patch.Reset()
			_, err := FileChecker("", false, false, false, 1)
			convey.So(err, convey.ShouldBeError)
		})
		patches.ApplyFuncReturn(filepath.Base, "0").
			ApplyFuncReturn(normalFileCheck, MockFileInfo{}, false, nil)
		convey.Convey("05-len(perm) not equal permLength should return error", func() {
			var a os.FileMode
			patch := gomonkey.ApplyMethodReturn(a, "String", "")
			defer patch.Reset()
			_, err := FileChecker("", false, false, false, 1)
			convey.So(err, convey.ShouldBeError)
		})
		convey.Convey("06-write permission error, should return error", func() {
			_, err := FileChecker("", false, false, false, 1)
			convey.So(err, convey.ShouldBeError)
		})
	})
}
