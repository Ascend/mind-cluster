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

// Package main
package process

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"ascend-docker-runtime/mindxcheckutils"
)

const (
	oldString = `{
        "runtimes":     {
                "ascend":       {
                        "path": "/test/runtime",
                        "runtimeArgs":  []
                }
        },
        "default-runtime":      "ascend"
	}`
	defaultRuntime    = `"default-runtime"`
	oldJson           = "old.json"
	createOldFail     = "create old failed %s"
	updateFail        = "update failed %s"
	updateFailAndData = "update failed %s, %v"
)

func jSONBytesEqual(a, b []byte) (bool, error) {
	var contentA, contentB interface{}
	if err := json.Unmarshal(a, &contentA); err != nil {
		return false, err
	}
	if err := json.Unmarshal(b, &contentB); err != nil {
		return false, err
	}
	return reflect.DeepEqual(contentB, contentA), nil
}

func TestCreateJsonStringWholeNew(t *testing.T) {
	data, err := createJsonString("/notExistedFile", "/test/runtime", "add")
	if err != nil {
		t.Fatalf("create string failed %s", err)
	}

	if eq, err := jSONBytesEqual([]byte(oldString), data); err != nil || !eq {
		t.Fatalf("empty create equal failed %s, %v", err, string(data))
	}
}

func TestCreateJsonStringUpdate(t *testing.T) {
	const perm = 0600
	if fid, err := os.OpenFile(oldJson, os.O_CREATE|os.O_RDWR|os.O_TRUNC, perm); err == nil {
		_, err = fid.Write([]byte(oldString))
		closeErr := fid.Close()
		if err != nil || closeErr != nil {
			t.Fatalf(createOldFail, err)
		}
	}
	data, err := createJsonString(oldJson, "/test/runtime1", "add")
	if err != nil {
		t.Fatalf(updateFail, err)
	}
	expectString := `{
        "runtimes":     {
                "ascend":       {
                        "path": "/test/runtime1",
                        "runtimeArgs":  []
                }
        },
        ` + defaultRuntime + `:      "ascend"
}`
	if eq, err := jSONBytesEqual([]byte(expectString), data); err != nil || !eq {
		t.Fatalf(updateFailAndData, err, string(data))
	}
}

func TestCreateJsonStringUpdateWithOtherParam(t *testing.T) {
	const perm = 0600
	oldStringWithParam := `{
        "runtimes":     {
                "ascend":       {
                        "path": "/test/runtime",
                        "runtimeArgs":  [1,2,3]
                },
				"runc2":       {
                        "path": "/test/runtime2",
                        "runtimeArgs":  [1,2,3]
                }
        },
        ` + defaultRuntime + `:      "runc"
}`
	if fid, err := os.OpenFile(oldJson, os.O_CREATE|os.O_RDWR|os.O_TRUNC, perm); err == nil {
		_, err = fid.Write([]byte(oldStringWithParam))
		closeErr := fid.Close()
		if err != nil || closeErr != nil {
			t.Fatalf(createOldFail, err)
		}
	}
	data, err := createJsonString(oldJson, "/test/runtime1", "add")
	if err != nil {
		t.Fatalf(updateFail, err)
	}
	expectString := `{
        "runtimes":     {
                "ascend":       {
                        "path": "/test/runtime1",
                        "runtimeArgs":  [1,2,3]
                },
				"runc2":       {
                        "path": "/test/runtime2",
                        "runtimeArgs":  [1,2,3]
                }
        },
        ` + defaultRuntime + `:      "ascend"
}`
	if eq, err := jSONBytesEqual([]byte(expectString), data); err != nil || !eq {
		t.Fatalf(updateFailAndData, err, string(data))
	}
}

func TestCreateJsonStrinRm(t *testing.T) {
	const perm = 0600
	if fid, err := os.OpenFile(oldJson, os.O_CREATE|os.O_RDWR|os.O_TRUNC, perm); err == nil {
		_, err = fid.Write([]byte(oldString))
		closeErr := fid.Close()
		if err != nil || closeErr != nil {
			t.Fatalf(createOldFail, err)
		}
	}
	data, err := createJsonString(oldJson, "", "rm")
	if err != nil {
		t.Fatalf(updateFail, err)
	}
	expectString := `{
        "runtimes":     {}
	}`
	if eq, err := jSONBytesEqual([]byte(expectString), data); err != nil || !eq {
		t.Fatalf(updateFailAndData, err, string(data))
	}
}

// TestCreateJsonString1 tests the function createJsonString patch1
func TestCreateJsonString1(t *testing.T) {
	convey.Convey("test createJsonString patch1", t, func() {
		convey.Convey("01-modifyDaemon error, should return error", func() {
			patches := gomonkey.ApplyFuncReturn(os.Stat, nil, nil).
				ApplyFuncReturn(modifyDaemon, nil, testError)
			defer patches.Reset()
			data, err := createJsonString(oldJson, "", "rm")
			convey.So(data, convey.ShouldBeNil)
			convey.So(err, convey.ShouldBeError)
		})
	})
	convey.Convey("test createJsonString patch1", t, func() {
		convey.Convey("02-MarshalIndent error, should return error", func() {
			patches := gomonkey.ApplyFuncReturn(os.Stat, nil, nil).
				ApplyFuncReturn(modifyDaemon, nil, nil).
				ApplyFuncReturn(json.MarshalIndent, []byte{}, testError)
			defer patches.Reset()
			data, err := createJsonString(oldJson, "", "rm")
			convey.So(data, convey.ShouldBeNil)
			convey.So(err, convey.ShouldBeError)
		})
	})
}

// TestCreateJsonString2 tests the function createJsonString patch2
func TestCreateJsonString2(t *testing.T) {
	convey.Convey("test createJsonString patch2", t, func() {
		convey.Convey("03-modifyDaemon error, should return error", func() {
			patches := gomonkey.ApplyFuncReturn(os.Stat, FileInfoMockV2{}, os.ErrNotExist)
			defer patches.Reset()
			reserveDefaultRuntime = true
			data, err := createJsonString(oldJson, "", "rm")
			convey.So(string(data), convey.ShouldEqual, fmt.Sprintf(noDefaultTemplate, ""))
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("04-stat error, should return error", func() {
			patches := gomonkey.ApplyFuncReturn(os.Stat, FileInfoMockV2{}, testError)
			defer patches.Reset()
			reserveDefaultRuntime = true
			data, err := createJsonString(oldJson, "", "rm")
			convey.So(data, convey.ShouldBeNil)
			convey.So(err, convey.ShouldBeError)
		})
	})
}

// TestWriteJson tests the function writeJson
func TestWriteJson(t *testing.T) {
	convey.Convey("test writeJson", t, func() {
		convey.Convey("01-write file error, should return error", func() {
			patches := gomonkey.ApplyFuncReturn(os.Stat, FileInfoMockV2{}, os.ErrNotExist).
				ApplyFuncReturn(os.OpenFile, &os.File{}, nil).
				ApplyMethodReturn(&os.File{}, "Write", 0, testError).
				ApplyMethodReturn(&os.File{}, "Close", nil)
			defer patches.Reset()
			err := writeJson("", []byte{})
			convey.So(err, convey.ShouldBeError)
		})
		convey.Convey("02-write file success, close fail, should return error", func() {
			patches := gomonkey.ApplyFuncReturn(os.Stat, FileInfoMockV2{}, os.ErrNotExist).
				ApplyFuncReturn(os.OpenFile, &os.File{}, nil).
				ApplyMethodReturn(&os.File{}, "Write", 0, nil).
				ApplyMethodReturn(&os.File{}, "Close", testError)
			defer patches.Reset()
			err := writeJson("", []byte{})
			convey.So(err, convey.ShouldBeError)
		})
		convey.Convey("03-write file success, close success, should return nil", func() {
			patches := gomonkey.ApplyFuncReturn(os.Stat, FileInfoMockV2{}, os.ErrNotExist).
				ApplyFuncReturn(os.OpenFile, &os.File{}, nil).
				ApplyMethodReturn(&os.File{}, "Write", 0, nil).
				ApplyMethodReturn(&os.File{}, "Close", nil)
			defer patches.Reset()
			err := writeJson("", []byte{})
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

type testProcessArg struct {
	Name       string
	Command    []string
	WantErr    bool
	WantResult string
}

// TestDockerProcess tests the function DockerProcess
func TestDockerProcess(t *testing.T) {
	tests := getTestDockerProcessCases()
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			if tt.Name == "success case 4" {
				patch := gomonkey.ApplyFunc(os.Stat, func(name string) (os.FileInfo, error) {
					return &FileInfoMock{}, nil
				})
				defer patch.Reset()
				patchRealDirCheck := gomonkey.ApplyFunc(mindxcheckutils.RealDirChecker, func(path string,
					checkParent, allowLink bool) (string, error) {
					return "", nil
				})
				defer patchRealDirCheck.Reset()
				patchRealFileCheck := gomonkey.ApplyFunc(mindxcheckutils.RealFileChecker, func(path string,
					checkParent, allowLink bool, size int) (string, error) {
					return "", nil
				})
				defer patchRealFileCheck.Reset()
				patchSize := gomonkey.ApplyMethod(reflect.TypeOf(&FileInfoMock{}), "Size", func(f *FileInfoMock) int64 {
					return 1
				})
				defer patchSize.Reset()
				patchClose := gomonkey.ApplyMethod(reflect.TypeOf(&os.File{}), "Close", func(_ *os.File) error {
					return nil
				})
				defer patchClose.Reset()
				patchReadAll := gomonkey.ApplyFunc(ioutil.ReadAll, func(r io.Reader) ([]byte, error) {
					testMap := map[string]interface{}{}
					jsonBytes, err := json.Marshal(testMap)
					if err != nil {
						fmt.Println("Error marshaling map:", err)
						return nil, nil
					}
					return jsonBytes, nil
				})
				defer patchReadAll.Reset()
			}
			got, got1 := DockerProcess(tt.Command)
			if (got1 == nil) == tt.WantErr {
				t.Errorf("DockerProcess() got = %v, want %v", got, tt.WantErr)
			}
			if got != tt.WantResult {
				t.Errorf("DockerProcess() got1 = %v, want %v", got1, tt.WantResult)
			}
		})
	}
}

// TestDockerProcess1 tests the function DockerProcess patch1
func TestDockerProcess1(t *testing.T) {
	emptyStr := ""
	destFileTest := "aaa.txt.pid"
	cmds := []string{"add", oldJson, destFileTest, emptyStr, emptyStr, emptyStr, emptyStr}
	convey.Convey("test DockerProcess patch1", t, func() {
		convey.Convey("01-stat error, should return error", func() {
			patches := gomonkey.ApplyFuncReturn(os.Stat, nil, testError).
				ApplyFuncReturn(mindxcheckutils.RealDirChecker, "", testError)
			defer patches.Reset()
			_, err := DockerProcess(cmds)
			convey.So(err, convey.ShouldBeError)
		})
		convey.Convey("02-stat ok, file check fail, should return error", func() {
			patches := gomonkey.ApplyFuncReturn(os.Stat, nil, nil).
				ApplyFuncReturn(mindxcheckutils.RealFileChecker, "", testError)
			defer patches.Reset()
			_, err := DockerProcess(cmds)
			convey.So(err, convey.ShouldBeError)
		})
		convey.Convey("03-file check pass, dir check fail, should return error", func() {
			patches := gomonkey.ApplyFuncReturn(os.Stat, nil, nil).
				ApplyFuncReturn(mindxcheckutils.RealFileChecker, "", nil).
				ApplyFuncReturn(mindxcheckutils.RealDirChecker, "", testError)
			defer patches.Reset()
			_, err := DockerProcess(cmds)
			convey.So(err, convey.ShouldBeError)
		})
	})
}

// TestDockerProcess2 tests the function DockerProcess patch2
func TestDockerProcess2(t *testing.T) {
	emptyStr := ""
	destFileTest := "aaa.txt.pid"
	cmds := []string{"add", oldJson, destFileTest, emptyStr, emptyStr, emptyStr, emptyStr}
	convey.Convey("test DockerProcess patch1", t, func() {
		convey.Convey("04-createJsonString fail, should return error", func() {
			patches := gomonkey.ApplyFuncReturn(os.Stat, FileInfoMockV2{}, nil).
				ApplyFuncReturn(mindxcheckutils.RealFileChecker, "", nil).
				ApplyFuncReturn(createJsonString, []byte{}, testError)
			defer patches.Reset()
			_, err := DockerProcess(cmds)
			convey.So(err, convey.ShouldBeError)
		})
	})
}

func getTestDockerProcessCases() []testProcessArg {
	emptyStr := ""
	addBehavior := "install"
	rmBehavior := "uninstall"
	destFileTest := "aaa.txt.pid"
	return []testProcessArg{
		{
			Name:    "error param case 1",
			Command: []string{"ins"},
			WantErr: true,
		},
		{
			Name:    "error param case 2",
			Command: []string{"add"},
			WantErr: true,
		},
		{
			Name:       "file not exist case 3",
			Command:    []string{"rm", oldJson, emptyStr, emptyStr, emptyStr, emptyStr, emptyStr, emptyStr},
			WantErr:    true,
			WantResult: rmBehavior,
		},
		{
			Name:       "success case 4",
			Command:    []string{"add", oldJson, destFileTest, emptyStr, emptyStr, emptyStr, emptyStr, emptyStr, emptyStr},
			WantErr:    true,
			WantResult: addBehavior,
		},
		{
			Name:       "error param case 5",
			Command:    []string{},
			WantErr:    true,
			WantResult: emptyStr,
		},
	}
}

// FileInfoMock is used to test
type FileInfoMock struct {
	os.FileInfo
}

// Size for FileInfoMock is used to test
func (f FileInfoMock) Size() int64 {
	return maxFileSize + 1
}

// FileInfoMockV2 is used to test
type FileInfoMockV2 struct {
	os.FileInfo
}

// Size for FileInfoMockV2 is used to test
func (f FileInfoMockV2) Size() int64 {
	return maxFileSize - 1
}

// Mode for FileInfoMockV2 is used to test
func (f FileInfoMockV2) Mode() os.FileMode {
	return os.ModePerm
}

// TestLoadOriginJson tests the function loadOriginJson
func TestLoadOriginJson(t *testing.T) {
	convey.Convey("test loadOriginJson", t, func() {
		convey.Convey("01-stat error, should return error", func() {
			patches := gomonkey.ApplyFuncReturn(os.Stat, nil, testError)
			defer patches.Reset()
			_, err := loadOriginJson("")
			convey.So(err, convey.ShouldBeError)
		})
		convey.Convey("02-stat error, should return error", func() {
			patches := gomonkey.ApplyFuncReturn(os.Stat, FileInfoMock{}, nil)
			defer patches.Reset()
			_, err := loadOriginJson("")
			convey.So(err, convey.ShouldBeError)
		})
		convey.Convey("03-open error, should return error", func() {
			patches := gomonkey.ApplyFuncReturn(os.Stat, FileInfoMockV2{}, nil).
				ApplyFuncReturn(os.Open, nil, testError)
			defer patches.Reset()
			_, err := loadOriginJson("")
			convey.So(err, convey.ShouldBeError)
		})
		convey.Convey("04-ReadAll error, should return error", func() {
			patches := gomonkey.ApplyFuncReturn(os.Stat, FileInfoMockV2{}, nil).
				ApplyFuncReturn(os.Open, &os.File{}, nil).
				ApplyFuncReturn(ioutil.ReadAll, nil, testError)
			defer patches.Reset()
			_, err := loadOriginJson("")
			convey.So(err, convey.ShouldBeError)
		})
	})
}

// TestLoadOriginJson1 tests the function loadOriginJson patch1
func TestLoadOriginJson1(t *testing.T) {
	convey.Convey("test loadOriginJson patch1", t, func() {
		convey.Convey("05-Close error, should return error", func() {
			patches := gomonkey.ApplyFuncReturn(os.Stat, FileInfoMockV2{}, nil).
				ApplyFuncReturn(os.Open, &os.File{}, nil).
				ApplyFuncReturn(ioutil.ReadAll, nil, nil).
				ApplyMethodReturn(&os.File{}, "Close", testError)
			defer patches.Reset()
			_, err := loadOriginJson("")
			convey.So(err, convey.ShouldBeError)
		})
		convey.Convey("06-Unmarshal error, should return error", func() {
			patches := gomonkey.ApplyFuncReturn(os.Stat, FileInfoMockV2{}, nil).
				ApplyFuncReturn(os.Open, &os.File{}, nil).
				ApplyFuncReturn(ioutil.ReadAll, nil, nil).
				ApplyMethodReturn(&os.File{}, "Close", nil).
				ApplyFuncReturn(json.Unmarshal, testError)
			defer patches.Reset()
			_, err := loadOriginJson("")
			convey.So(err, convey.ShouldBeError)
		})
	})
}
