/*
Copyright(C)2025. Huawei Technologies Co.,Ltd. All rights reserved.

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

// Package roceping for ping by icmp in RoCE mesh net between super pods in A5
package roceping

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/common-utils/utils"
)

func TestNewFileWatcherLoop(t *testing.T) {
	convey.Convey("test NewFileWatcherLoop func", t, func() {
		convey.Convey("01-should not be nil when input ok", func() {
			fw := NewFileWatcherLoop(context.Background(), 1, "/")
			convey.So(fw, convey.ShouldNotBeNil)
		})
	})
}

func TestAddListenPath(t *testing.T) {
	convey.Convey("test FileWatcherLoop method AddListenPath", t, func() {
		convey.Convey("01-should return err when file path is empty", func() {
			fw := NewFileWatcherLoop(context.Background(), 1, "/")
			err := fw.AddListenPath("")
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("02-should return err when save path is not exist", func() {
			fw := NewFileWatcherLoop(context.Background(), 1, "")
			err := fw.AddListenPath("/xxx")
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("03-should return err when save path is invalid", func() {
			fw := NewFileWatcherLoop(context.Background(), 1, "/")
			patch := gomonkey.ApplyFuncReturn(utils.IsLexist, true)
			defer patch.Reset()
			patchCheck := gomonkey.ApplyFuncReturn(utils.CheckPath, "", errors.New("path invalid"))
			defer patchCheck.Reset()
			err := fw.AddListenPath("/xxx")
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("04-should return nil when all is valid", func() {
			fw := NewFileWatcherLoop(context.Background(), 1, "/")
			patch := gomonkey.ApplyFuncReturn(utils.IsLexist, true)
			defer patch.Reset()
			patchCheck := gomonkey.ApplyFuncReturn(utils.CheckPath, "/", nil)
			defer patchCheck.Reset()
			err := fw.AddListenPath("/xxx")
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestListenEvents(t *testing.T) {
	convey.Convey("test FileWatcherLoop method ListenEvents", t, func() {
		convey.Convey("01-should stop when received ctx done", func() {
			ctx, cancel := context.WithCancel(context.Background())
			fw := NewFileWatcherLoop(ctx, 1, "/")
			go fw.ListenEvents()
			const waitingPeriod = 2
			time.Sleep(time.Duration(waitingPeriod) * time.Second)
			ret := runtime.NumGoroutine()
			cancel()
			time.Sleep(1 * time.Second)
			ret1 := runtime.NumGoroutine()
			convey.So(ret, convey.ShouldEqual, ret1+1)
		})
		convey.Convey("02-should stop when send file changed event", func() {
			ctx, cancel := context.WithCancel(context.Background())
			fw := NewFileWatcherLoop(ctx, 1, "/")
			patch := gomonkey.ApplyPrivateMethod(fw, "checkWatchedFileChanged",
				func(w *FileWatcherLoop) (*FileEvent, error) {
					return &FileEvent{}, nil
				},
			)
			defer patch.Reset()
			go fw.ListenEvents()
			const waitingPeriod = 2
			time.Sleep(time.Duration(waitingPeriod) * time.Second)
			ret := runtime.NumGoroutine()
			cancel()
			time.Sleep(1 * time.Second)
			ret1 := runtime.NumGoroutine()
			convey.So(ret, convey.ShouldEqual, ret1+1)
		})
	})
}

func TestGetEventChan(t *testing.T) {
	convey.Convey("test FileWatcherLoop method GetEventChan", t, func() {
		fw := NewFileWatcherLoop(context.Background(), 1, "")
		convey.So(fw.GetEventChan(), convey.ShouldEqual, fw.eventChan)
	})
}

func TestCheckWatchedFileChanged(t *testing.T) {
	convey.Convey("test FileWatcherLoop method checkWatchedFileChanged", t, func() {
		testCheckWatchedFileChangedCase1()
		testCheckWatchedFileChangedCase2()
		testCheckWatchedFileChangedCase3()
	})
}

func testCheckWatchedFileChangedCase1() {
	fw := NewFileWatcherLoop(context.Background(), 1, "")
	convey.Convey("01-should return nil when watched file is not exist", func() {
		fw.watchedFile = "/xxx/not/exist"
		event, err := fw.checkWatchedFileChanged()
		convey.So(err, convey.ShouldBeNil)
		convey.So(event, convey.ShouldBeNil)
	})

	convey.Convey("02-should return err when get watched file info failed", func() {
		fw.watchedFile = "/xxx/not/exist"
		patch := gomonkey.ApplyFuncReturn(utils.IsLexist, true)
		defer patch.Reset()
		patchGet := gomonkey.ApplyPrivateMethod(fw, "getWatchedFileInfo",
			func() (*FileMetaInfo, error) {
				return &FileMetaInfo{}, errors.New("get watched file info failed")
			},
		)
		defer patchGet.Reset()
		event, err := fw.checkWatchedFileChanged()
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(event, convey.ShouldBeNil)
	})

	convey.Convey("03-should return event when check file not exist", func() {
		fw.watchedFile = "/xxx/not/exist"
		exitFuncCalledCnt := 0
		patch := gomonkey.ApplyFunc(utils.IsLexist, func(filePath string) bool {
			defer func() { exitFuncCalledCnt++ }()
			if exitFuncCalledCnt == 0 {
				return true
			}
			return false
		})
		defer patch.Reset()
		patchGet := gomonkey.ApplyPrivateMethod(fw, "getWatchedFileInfo",
			func() (*FileMetaInfo, error) {
				return &FileMetaInfo{}, nil
			},
		)
		defer patchGet.Reset()
		event, err := fw.checkWatchedFileChanged()
		convey.So(err, convey.ShouldBeNil)
		convey.So(event, convey.ShouldNotBeNil)
	})
}

func testCheckWatchedFileChangedCase2() {
	fw := NewFileWatcherLoop(context.Background(), 1, "")
	convey.Convey("04-should return event when read check file failed", func() {
		fw.watchedFile = "/xxx/not/exist"
		patch := gomonkey.ApplyFuncReturn(utils.IsLexist, true)
		defer patch.Reset()
		patchGet := gomonkey.ApplyPrivateMethod(fw, "getWatchedFileInfo",
			func() (*FileMetaInfo, error) {
				return &FileMetaInfo{}, nil
			},
		)
		defer patchGet.Reset()
		patchRead := gomonkey.ApplyPrivateMethod(fw, "readMetaInfoFromCheckFile",
			func() (*FileMetaInfo, error) {
				return &FileMetaInfo{}, errors.New("invalid file")
			},
		)
		defer patchRead.Reset()
		event, err := fw.checkWatchedFileChanged()
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(event, convey.ShouldBeNil)
	})
	convey.Convey("05-should return event when data hash is equal", func() {
		fw.watchedFile = "/xxx/not/exist"
		patch := gomonkey.ApplyFuncReturn(utils.IsLexist, true)
		defer patch.Reset()
		patchGet := gomonkey.ApplyPrivateMethod(fw, "getWatchedFileInfo",
			func() (*FileMetaInfo, error) {
				return &FileMetaInfo{}, nil
			},
		)
		defer patchGet.Reset()
		patchRead := gomonkey.ApplyPrivateMethod(fw, "readMetaInfoFromCheckFile",
			func() (*FileMetaInfo, error) {
				return &FileMetaInfo{}, nil
			},
		)
		defer patchRead.Reset()
		event, err := fw.checkWatchedFileChanged()
		convey.So(err, convey.ShouldBeNil)
		convey.So(event, convey.ShouldBeNil)
	})
}

func testCheckWatchedFileChangedCase3() {
	fw := NewFileWatcherLoop(context.Background(), 1, "")
	convey.Convey("06-should return event when data hash is not equal", func() {
		fw.watchedFile = "/xxx/not/exist"
		patch := gomonkey.ApplyFuncReturn(utils.IsLexist, true)
		defer patch.Reset()
		patchGet := gomonkey.ApplyPrivateMethod(fw, "getWatchedFileInfo",
			func() (*FileMetaInfo, error) {
				return &FileMetaInfo{DataHash: "aaa"}, nil
			},
		)
		defer patchGet.Reset()
		patchRead := gomonkey.ApplyPrivateMethod(fw, "readMetaInfoFromCheckFile",
			func() (*FileMetaInfo, error) {
				return &FileMetaInfo{DataHash: "bbb"}, nil
			},
		)
		defer patchRead.Reset()
		event, err := fw.checkWatchedFileChanged()
		convey.So(err, convey.ShouldBeNil)
		convey.So(event, convey.ShouldNotBeNil)
	})
}

func TestSendFileChangeEvent(t *testing.T) {
	convey.Convey("test FileWatcherLoop method sendFileChangeEvent", t, func() {
		convey.Convey("01-should return nil when event is nil", func() {
			fw := NewFileWatcherLoop(context.Background(), 1, "/")
			err := fw.sendFileChangeEvent(nil)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("02-should stop when send file changed event", func() {
			ctx, cancel := context.WithCancel(context.Background())
			fw := NewFileWatcherLoop(ctx, 1, "/")
			go func() {
				_ = fw.sendFileChangeEvent(&FileEvent{})
			}()
			ret := runtime.NumGoroutine()
			cancel()
			time.Sleep(1 * time.Second)
			ret1 := runtime.NumGoroutine()
			convey.So(ret, convey.ShouldEqual, ret1+1)
		})
	})
}

func TestReadMetaInfoFromCheckFile(t *testing.T) {
	convey.Convey("test FileWatcherLoop method readMetaInfoFromCheckFile", t, func() {
		fwl := NewFileWatcherLoop(context.Background(), 1, "/")
		convey.Convey("01-should return err when check file path is empty", func() {
			fwl.checkFilePath = ""
			_, err := fwl.readMetaInfoFromCheckFile()
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("02-should return err when check file path is not exist", func() {
			fwl.checkFilePath = "/xxx/not/exist/file"
			_, err := fwl.readMetaInfoFromCheckFile()
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("03-should return err when check file path is invalid", func() {
			fwl.checkFilePath = "/xxx/not/exist/file"
			patch := gomonkey.ApplyFuncReturn(utils.IsLexist, true)
			defer patch.Reset()
			patchCheck := gomonkey.ApplyFuncReturn(utils.CheckPath, "/", errors.New("invalid path"))
			defer patchCheck.Reset()
			_, err := fwl.readMetaInfoFromCheckFile()
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("04-should return err when check file path is invalid", func() {
			fwl.checkFilePath = "/xxx/not/exist/file"
			patch := gomonkey.ApplyFuncReturn(utils.IsLexist, true)
			defer patch.Reset()
			patchCheck := gomonkey.ApplyFuncReturn(utils.CheckPath, fwl.checkFilePath, nil)
			defer patchCheck.Reset()
			patchRead := gomonkey.ApplyFuncReturn(utils.ReadLimitBytes, nil, errors.New("invalid size"))
			defer patchRead.Reset()
			_, err := fwl.readMetaInfoFromCheckFile()
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("05-should return err when check file path is invalid", func() {
			fwl.checkFilePath = "/xxx/not/exist/file"
			patch := gomonkey.ApplyFuncReturn(utils.IsLexist, true)
			defer patch.Reset()
			patchCheck := gomonkey.ApplyFuncReturn(utils.CheckPath, fwl.checkFilePath, nil)
			defer patchCheck.Reset()
			dataBytes, err := makeFileMetaInfoData()
			convey.So(err, convey.ShouldBeNil)
			patchRead := gomonkey.ApplyFuncReturn(utils.ReadLimitBytes, dataBytes, nil)
			defer patchRead.Reset()
			info, err := fwl.readMetaInfoFromCheckFile()
			convey.So(err, convey.ShouldBeNil)
			convey.So(info, convey.ShouldNotBeNil)
		})
	})
}

func makeFileMetaInfoData() ([]byte, error) {
	info := FileMetaInfo{
		Name:       "a",
		Size:       1,
		ModifyTime: time.Now().UnixMilli(),
		DataHash:   "b",
	}
	return json.Marshal(info)
}

func TestGetWatchedFileInfo(t *testing.T) {
	convey.Convey("test FileWatcherLoop method getWatchedFileInfo", t, func() {
		fwl := NewFileWatcherLoop(context.Background(), 1, "/")
		convey.Convey("01-should return err when watched file path is empty", func() {
			fwl.watchedFile = ""
			_, err := fwl.getWatchedFileInfo()
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("02-should return err when check file path is not exist", func() {
			fwl.watchedFile = "/xxx/not/exist/file"
			_, err := fwl.getWatchedFileInfo()
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("03-should return err when check file path is invalid", func() {
			fwl.watchedFile = "/xxx/not/exist/file"
			patch := gomonkey.ApplyFuncReturn(utils.IsLexist, true)
			defer patch.Reset()
			patchCheck := gomonkey.ApplyFuncReturn(utils.CheckPath, "/", errors.New("invalid path"))
			defer patchCheck.Reset()
			_, err := fwl.getWatchedFileInfo()
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("04-should return err when stat file failed", func() {
			fwl.watchedFile = "/xxx/not/exist/file"
			patch := gomonkey.ApplyFuncReturn(utils.IsLexist, true)
			defer patch.Reset()
			patchCheck := gomonkey.ApplyFuncReturn(utils.CheckPath, fwl.checkFilePath, nil)
			defer patchCheck.Reset()
			_, err := fwl.getWatchedFileInfo()
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("05-should return err when file size invalid", func() {
			fwl.watchedFile = "/xxx/not/exist/file"
			patch := gomonkey.ApplyFuncReturn(utils.IsLexist, true)
			defer patch.Reset()
			patchCheck := gomonkey.ApplyFuncReturn(utils.CheckPath, fwl.checkFilePath, nil)
			defer patchCheck.Reset()
			patchStat := gomonkey.ApplyFunc(os.Stat, func(fileName string) (os.FileInfo, error) {
				return mockOsFileInfoErr{}, nil
			})
			defer patchStat.Reset()
			_, err := fwl.getWatchedFileInfo()
			convey.So(err, convey.ShouldNotBeNil)
		})
		testGetWatchedFileInfoWhenGetHash()
	})
}

func testGetWatchedFileInfoWhenGetHash() {
	fwl := NewFileWatcherLoop(context.Background(), 1, "/")
	convey.Convey("01-should return err when get data hash failed", func() {
		fwl.watchedFile = "/xxx/not/exist/file"
		patch := gomonkey.ApplyFuncReturn(utils.IsLexist, true)
		defer patch.Reset()
		patchCheck := gomonkey.ApplyFuncReturn(utils.CheckPath, fwl.checkFilePath, nil)
		defer patchCheck.Reset()
		patchStat := gomonkey.ApplyFunc(os.Stat, func(fileName string) (os.FileInfo, error) {
			return mockOsFileInfo{}, nil
		})
		defer patchStat.Reset()
		patchHash := gomonkey.ApplyFuncReturn(GetFileDataHash, "", errors.New("invalid file"))
		defer patchHash.Reset()
		_, err := fwl.getWatchedFileInfo()
		convey.So(err, convey.ShouldNotBeNil)
	})
	convey.Convey("02-should return nil when get file hash success", func() {
		fwl.watchedFile = "/xxx/not/exist/file"
		patch := gomonkey.ApplyFuncReturn(utils.IsLexist, true)
		defer patch.Reset()
		patchCheck := gomonkey.ApplyFuncReturn(utils.CheckPath, fwl.checkFilePath, nil)
		defer patchCheck.Reset()
		patchStat := gomonkey.ApplyFunc(os.Stat, func(fileName string) (os.FileInfo, error) {
			return mockOsFileInfo{}, nil
		})
		defer patchStat.Reset()
		patchHash := gomonkey.ApplyFuncReturn(GetFileDataHash, "aaa", nil)
		defer patchHash.Reset()
		info, err := fwl.getWatchedFileInfo()
		convey.So(err, convey.ShouldBeNil)
		convey.So(info, convey.ShouldNotBeNil)
	})
}

type mockOsFileInfo struct {
}

func (f mockOsFileInfo) Name() string {
	return "mockOsFileInfo"
}
func (f mockOsFileInfo) Size() int64 {
	return 1
}

func (f mockOsFileInfo) Mode() os.FileMode {
	return os.ModePerm
}

func (f mockOsFileInfo) ModTime() time.Time {
	return time.Now()
}

func (f mockOsFileInfo) IsDir() bool {
	return false
}

func (f mockOsFileInfo) Sys() any {
	return nil
}

type mockOsFileInfoErr struct {
}

func (f mockOsFileInfoErr) Name() string {
	return "mockOsFileInfoErr"
}
func (f mockOsFileInfoErr) Size() int64 {
	return 1 + maxFileSize
}

func (f mockOsFileInfoErr) Mode() os.FileMode {
	return os.ModePerm
}

func (f mockOsFileInfoErr) ModTime() time.Time {
	return time.Now()
}

func (f mockOsFileInfoErr) IsDir() bool {
	return false
}

func (f mockOsFileInfoErr) Sys() any {
	return nil
}

func TestSaveMetaInfoToCheckFile(t *testing.T) {
	convey.Convey("test FileWatcherLoop method saveMetaInfoToCheckFile", t, func() {
		fwl := NewFileWatcherLoop(context.Background(), 1, "/")
		convey.Convey("01-should return err when info is empty", func() {
			err := fwl.saveMetaInfoToCheckFile(nil)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("02-should return err when check file path is empty", func() {
			err := fwl.saveMetaInfoToCheckFile(&FileMetaInfo{})
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("03-should return err when check file path is not exist", func() {
			fwl.checkFilePath = "/xxx/not/exist/file.check"
			patchCheck := gomonkey.ApplyFuncReturn(utils.CheckPath, fwl.checkFilePath, errors.New("invalid file"))
			defer patchCheck.Reset()
			err := fwl.saveMetaInfoToCheckFile(&FileMetaInfo{})
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("04-should return err when json Marshal filed", func() {
			fwl.checkFilePath = "/xxx/not/exist/file"
			patch := gomonkey.ApplyFuncReturn(utils.IsLexist, true)
			defer patch.Reset()
			patchCheck := gomonkey.ApplyFuncReturn(utils.CheckPath, fwl.checkFilePath, nil)
			defer patchCheck.Reset()
			patchMarshal := gomonkey.ApplyFuncReturn(json.Marshal, nil, errors.New("json marshal failed"))
			defer patchMarshal.Reset()
			err := fwl.saveMetaInfoToCheckFile(&FileMetaInfo{})
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("05-should return err when open file failed", func() {
			fwl.checkFilePath = "/xxx/not/exist/file"
			patch := gomonkey.ApplyFuncReturn(utils.IsLexist, true)
			defer patch.Reset()
			patchCheck := gomonkey.ApplyFuncReturn(utils.CheckPath, fwl.checkFilePath, nil)
			defer patchCheck.Reset()
			patchOpen := gomonkey.ApplyFuncReturn(os.OpenFile, nil, errors.New("open file failed"))
			defer patchOpen.Reset()
			err := fwl.saveMetaInfoToCheckFile(&FileMetaInfo{})
			convey.So(err, convey.ShouldNotBeNil)
		})
		testSaveMetaInfoToCheckFileWhenFileOp(fwl)
	})
}

func testSaveMetaInfoToCheckFileWhenFileOp(fwl *FileWatcherLoop) {
	convey.Convey("06-should return err when chmod file failed", func() {
		fwl.checkFilePath = "/xxx/not/exist/file"
		patch := gomonkey.ApplyFuncReturn(utils.IsLexist, true)
		defer patch.Reset()
		patchCheck := gomonkey.ApplyFuncReturn(utils.CheckPath, fwl.checkFilePath, nil)
		defer patchCheck.Reset()
		patchOpen := gomonkey.ApplyFuncReturn(os.OpenFile, &os.File{}, nil)
		defer patchOpen.Reset()
		patchChmod := gomonkey.ApplyMethodReturn(&os.File{}, "Chmod", errors.New("chmod failed"))
		defer patchChmod.Reset()
		err := fwl.saveMetaInfoToCheckFile(&FileMetaInfo{})
		convey.So(err, convey.ShouldNotBeNil)
	})
	convey.Convey("07-should return err when write file failed", func() {
		fwl.checkFilePath = "/xxx/not/exist/file"
		patch := gomonkey.ApplyFuncReturn(utils.IsLexist, true)
		defer patch.Reset()
		patchCheck := gomonkey.ApplyFuncReturn(utils.CheckPath, fwl.checkFilePath, nil)
		defer patchCheck.Reset()
		patchOpen := gomonkey.ApplyFuncReturn(os.OpenFile, &os.File{}, nil)
		defer patchOpen.Reset()
		patchChmod := gomonkey.ApplyMethodReturn(&os.File{}, "Chmod", nil)
		defer patchChmod.Reset()
		patchWrite := gomonkey.ApplyMethodReturn(&os.File{}, "Write", 0, errors.New("write failed"))
		defer patchWrite.Reset()
		err := fwl.saveMetaInfoToCheckFile(&FileMetaInfo{})
		convey.So(err, convey.ShouldNotBeNil)
	})
	convey.Convey("07-should return nil when write file success", func() {
		fwl.checkFilePath = "/xxx/not/exist/file"
		patch := gomonkey.ApplyFuncReturn(utils.IsLexist, true)
		defer patch.Reset()
		patchCheck := gomonkey.ApplyFuncReturn(utils.CheckPath, fwl.checkFilePath, nil)
		defer patchCheck.Reset()
		patchOpen := gomonkey.ApplyFuncReturn(os.OpenFile, &os.File{}, nil)
		defer patchOpen.Reset()
		patchChmod := gomonkey.ApplyMethodReturn(&os.File{}, "Chmod", nil)
		defer patchChmod.Reset()
		patchWrite := gomonkey.ApplyMethodReturn(&os.File{}, "Write", 1, nil)
		defer patchWrite.Reset()
		err := fwl.saveMetaInfoToCheckFile(&FileMetaInfo{})
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestGetFileDataHash(t *testing.T) {
	convey.Convey("test GetFileDataHash func", t, func() {
		convey.Convey("01-should return err when file path empty", func() {
			_, err := GetFileDataHash("", maxFileSize)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("02-should return err when file path not exist", func() {
			patch := gomonkey.ApplyFuncReturn(utils.IsLexist, false)
			defer patch.Reset()
			_, err := GetFileDataHash("/path/to/file", maxFileSize)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("03-should return err when file path invalid", func() {
			patch := gomonkey.ApplyFuncReturn(utils.IsLexist, true)
			defer patch.Reset()
			patchCheck := gomonkey.ApplyFuncReturn(utils.CheckPath, "", errors.New("path invalid"))
			defer patchCheck.Reset()
			_, err := GetFileDataHash("/path/to/file", maxFileSize)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("04-should return err when file path read failed", func() {
			targetFilePath := "/path/to/file"
			patch := gomonkey.ApplyFuncReturn(utils.IsLexist, true)
			defer patch.Reset()
			patchCheck := gomonkey.ApplyFuncReturn(utils.CheckPath, targetFilePath, nil)
			defer patchCheck.Reset()
			patchRead := gomonkey.ApplyFuncReturn(utils.ReadLimitBytes, nil, errors.New("invalid size"))
			defer patchRead.Reset()
			_, err := GetFileDataHash(targetFilePath, maxFileSize)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("05-should return nil when file path read success", func() {
			targetFilePath := "/path/to/file"
			patch := gomonkey.ApplyFuncReturn(utils.IsLexist, true)
			defer patch.Reset()
			patchCheck := gomonkey.ApplyFuncReturn(utils.CheckPath, targetFilePath, nil)
			defer patchCheck.Reset()
			patchRead := gomonkey.ApplyFuncReturn(utils.ReadLimitBytes, []byte{1, 0, 1}, nil)
			defer patchRead.Reset()
			_, err := GetFileDataHash(targetFilePath, maxFileSize)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestUpdateCheckFile(t *testing.T) {
	w := &FileWatcherLoop{watchedFile: "1", fileHashLock: &sync.RWMutex{}}
	convey.Convey("test UpdateCheckFile case 1", t, func() {
		mock1 := gomonkey.ApplyFunc(utils.IsLexist, func(_ string) bool {
			return false
		})
		defer mock1.Reset()
		convey.So(w.UpdateCheckFile(), convey.ShouldBeNil)
	})
	mock1 := gomonkey.ApplyFunc(utils.IsLexist, func(_ string) bool { return true })
	defer mock1.Reset()
	convey.Convey("test UpdateCheckFile case 2", t, func() {
		mock2 := gomonkey.ApplyFunc((*FileWatcherLoop).getWatchedFileInfo,
			func(_ *FileWatcherLoop) (*FileMetaInfo, error) { return nil, errors.New("fake error 1") })
		defer mock2.Reset()
		convey.So(w.UpdateCheckFile().Error(), convey.ShouldEqual, "fake error 1")
	})
	info := FileMetaInfo{DataHash: "123"}
	mock2 := gomonkey.ApplyFunc((*FileWatcherLoop).getWatchedFileInfo,
		func(_ *FileWatcherLoop) (*FileMetaInfo, error) { return &info, nil })
	defer mock2.Reset()
	convey.Convey("test UpdateCheckFile case 3", t, func() {
		mock3 := gomonkey.ApplyFunc((*FileWatcherLoop).saveMetaInfoToCheckFile,
			func(_ *FileWatcherLoop, _ *FileMetaInfo) error { return errors.New("fake error 2") })
		defer mock3.Reset()
		convey.So(w.UpdateCheckFile().Error(), convey.ShouldEqual, "fake error 2")
	})
	mock3 := gomonkey.ApplyFunc((*FileWatcherLoop).saveMetaInfoToCheckFile,
		func(_ *FileWatcherLoop, _ *FileMetaInfo) error { return nil })
	defer mock3.Reset()
	convey.Convey("test UpdateCheckFile case 4", t, func() {
		convey.So(w.UpdateCheckFile(), convey.ShouldBeNil)
	})
}

func TestGetCurFileHash(t *testing.T) {
	convey.Convey("test GetCurFileHash func", t, func() {
		hash := "1"
		w := &FileWatcherLoop{
			curFileHash:  hash,
			fileHashLock: &sync.RWMutex{},
		}
		convey.So(w.GetCurFileHash(), convey.ShouldEqual, hash)
	})
}
