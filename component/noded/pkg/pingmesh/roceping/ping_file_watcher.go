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
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
)

const (
	checkFileExt = ".check"
	defaultPerm  = 0400
)

// FileMetaInfo for file meta data info
type FileMetaInfo struct {
	Name       string
	Size       int64
	ModifyTime int64
	DataHash   string
}

// FileEvent for file content change event
type FileEvent struct {
	FileMetaInfo
}

// FileWatcherLoop for watching file info by loop
type FileWatcherLoop struct {
	ctx           context.Context
	watchedFile   string
	checkFilePath string
	savePath      string
	intervalSec   int
	eventChan     chan FileEvent
	curFileHash   string
	fileHashLock  *sync.RWMutex
}

// NewFileWatcherLoop for create instance of FileWatcherLoop
func NewFileWatcherLoop(ctx context.Context, interval int, savePath string) *FileWatcherLoop {
	return &FileWatcherLoop{
		ctx:          ctx,
		intervalSec:  interval,
		savePath:     savePath,
		eventChan:    make(chan FileEvent),
		curFileHash:  "",
		fileHashLock: &sync.RWMutex{},
	}
}

// AddListenPath for adding listen file path
func (w *FileWatcherLoop) AddListenPath(filePath string) error {
	if w == nil {
		return errors.New("file watcher is empty")
	}
	if len(filePath) == 0 {
		return errors.New("file path is empty")
	}
	w.watchedFile = filePath
	fileName := filepath.Base(filePath)
	if !utils.IsLexist(w.savePath) {
		return fmt.Errorf("%s file dir is not exist", w.savePath)
	}
	if _, err := utils.CheckPath(w.savePath); err != nil {
		return fmt.Errorf("save path is invalid: %v", err)
	}
	w.checkFilePath = filepath.Join(w.savePath, fmt.Sprintf("%s%s", fileName, checkFileExt))
	return nil
}

// ListenEvents for listening file change events
func (w *FileWatcherLoop) ListenEvents() {
	hwlog.RunLog.Infof("start goroutine for listening filepath(%s) events:", w.watchedFile)
	ticker := time.NewTicker(time.Second * time.Duration(w.intervalSec))
	defer ticker.Stop()

	for {
		select {
		case <-w.ctx.Done():
			hwlog.RunLog.Info("FileWatcher received ctx.Done event, will stop listen events")
			return
		case _, ok := <-ticker.C:
			if !ok {
				hwlog.RunLog.Info("period ticker channel is closed, will return")
				return
			}
			hwlog.RunLog.Infof("period ticker for %s begins to check", filepath.Base(w.watchedFile))
			event, err := w.checkWatchedFileChanged()
			if err != nil {
				hwlog.RunLog.Errorf("check file info failed, err : %v", err)
				continue
			}
			if err = w.sendFileChangeEvent(event); err != nil {
				hwlog.RunLog.Errorf("send file change event failed, err : %v", err)
			}
		}
	}
}

// GetEventChan for getting file events channel
func (w *FileWatcherLoop) GetEventChan() chan FileEvent {
	return w.eventChan
}

// GetCurFileHash for getting current file content hash value
func (w *FileWatcherLoop) GetCurFileHash() string {
	w.fileHashLock.RLock()
	val := w.curFileHash
	w.fileHashLock.RUnlock()
	return val
}

func (w *FileWatcherLoop) updateCurFileHash(fileHash string) {
	w.fileHashLock.Lock()
	w.curFileHash = fileHash
	w.fileHashLock.Unlock()
}

// UpdateCheckFile for updating check file when first read success
func (w *FileWatcherLoop) UpdateCheckFile() error {
	if !utils.IsLexist(w.watchedFile) {
		return nil
	}
	info, err := w.getWatchedFileInfo()
	if err != nil {
		hwlog.RunLog.Errorf("get file %s info failed, err: %v", w.watchedFile, err)
		return err
	}
	if err = w.saveMetaInfoToCheckFile(info); err != nil {
		return err
	}
	w.updateCurFileHash(info.DataHash)
	hwlog.RunLog.Infof("update meta info to check file %s success", w.checkFilePath)
	return nil
}

func (w *FileWatcherLoop) checkWatchedFileChanged() (*FileEvent, error) {
	if !utils.IsLexist(w.watchedFile) {
		return nil, nil
	}
	info, err := w.getWatchedFileInfo()
	if err != nil {
		return nil, err
	}
	hwlog.RunLog.Infof("%s current meta info: %v", w.watchedFile, *info)
	if !utils.IsLexist(w.checkFilePath) {
		hwlog.RunLog.Infof("will send file change event, because of the check file %s not exist", w.checkFilePath)
		fileEvent := &FileEvent{
			FileMetaInfo: *info,
		}
		return fileEvent, nil
	}

	oldInfo, err := w.readMetaInfoFromCheckFile()
	if err != nil {
		hwlog.RunLog.Errorf("read old info from check file %s failed, err: %v", w.checkFilePath, err)
		return nil, err
	}

	hwlog.RunLog.Infof("read meta info from file %s success", w.checkFilePath)
	if oldInfo.DataHash == info.DataHash {
		hwlog.RunLog.Info("file content is not changed")
		return nil, nil
	}

	hwlog.RunLog.Infof("file content changed, oldInfo: %v, newInfo: %v", *oldInfo, *info)
	fileEvent := &FileEvent{
		FileMetaInfo: *info,
	}
	return fileEvent, nil
}

func (w *FileWatcherLoop) getWatchedFileInfo() (*FileMetaInfo, error) {
	if len(w.watchedFile) == 0 {
		return nil, errors.New("watched file path is empty")
	}

	if !utils.IsLexist(w.watchedFile) {
		return nil, fmt.Errorf("the watched file not exist: %s", w.watchedFile)
	}

	if _, err := utils.CheckPath(w.watchedFile); err != nil {
		return nil, err
	}

	fileInfo, err := os.Stat(w.watchedFile)
	if err != nil {
		return nil, err
	}

	size := fileInfo.Size()
	if size == 0 || size > maxFileSize {
		return nil, fmt.Errorf("file size is invalid: %d, which should in range [1, %d] bytes", size, maxFileSize)
	}
	dataHash, err := GetFileDataHash(w.watchedFile, maxFileSize)
	if err != nil {
		return nil, err
	}
	info := &FileMetaInfo{
		Name:       filepath.Base(w.watchedFile),
		Size:       size,
		ModifyTime: fileInfo.ModTime().UnixMilli(),
		DataHash:   dataHash,
	}
	return info, nil
}

func (w *FileWatcherLoop) sendFileChangeEvent(event *FileEvent) error {
	if event == nil {
		return nil
	}

	timer := time.NewTimer(time.Duration(w.intervalSec) * time.Second)
	defer timer.Stop()

	select {
	case <-w.ctx.Done():
		hwlog.RunLog.Infof("received signal from ctx done when sending file event, will return")
		return nil
	case <-timer.C:
		hwlog.RunLog.Error("send file change event timed out")
		return errors.New("send file change event timed out")
	case w.eventChan <- *event:
		hwlog.RunLog.Infof("send file change event to channel success")
		return nil
	}
}

func (w *FileWatcherLoop) readMetaInfoFromCheckFile() (*FileMetaInfo, error) {
	if len(w.checkFilePath) == 0 {
		return nil, errors.New("the check file path is empty")
	}

	if !utils.IsLexist(w.checkFilePath) {
		return nil, errors.New("the check file path is not exist")
	}

	if _, err := utils.CheckPath(w.checkFilePath); err != nil {
		return nil, err
	}

	dataBytes, err := utils.ReadLimitBytes(w.checkFilePath, maxFileSize)
	if err != nil {
		return nil, err
	}

	var info FileMetaInfo
	err = json.Unmarshal(dataBytes, &info)
	if err != nil {
		return nil, err
	}
	return &info, nil
}

func (w *FileWatcherLoop) saveMetaInfoToCheckFile(info *FileMetaInfo) error {
	if info == nil {
		return errors.New("input is empty")
	}
	if len(w.checkFilePath) == 0 {
		return errors.New("the check file path is empty")
	}

	if _, err := utils.CheckPath(w.checkFilePath); err != nil {
		return fmt.Errorf("the check file path is invalid, err: %s", err)
	}

	dataBytes, err := json.Marshal(*info)
	if err != nil {
		return err
	}

	openFlag := os.O_RDWR | os.O_CREATE | os.O_TRUNC
	f, err := os.OpenFile(w.checkFilePath, openFlag, defaultPerm)
	if err != nil {
		return err
	}
	defer func() {
		errClose := f.Close()
		if errClose != nil {
			hwlog.RunLog.Errorf("close file %s failed, err: %v", w.checkFilePath, errClose)
			return
		}
	}()
	err = f.Chmod(defaultPerm)
	if err != nil {
		return fmt.Errorf("chmod file %s failed, err: %v", w.checkFilePath, err)
	}

	_, err = f.Write(dataBytes)
	return err
}

// GetFileDataHash for calculating hash value of the file content data
func GetFileDataHash(filePath string, limitLength int) (string, error) {
	if len(filePath) == 0 {
		return "", errors.New("file path is empty")
	}
	if !utils.IsLexist(filePath) {
		return "", errors.New("file path is not exist")
	}
	if _, err := utils.CheckPath(filePath); err != nil {
		return "", err
	}
	data, err := utils.ReadLimitBytes(filePath, limitLength)
	if err != nil {
		return "", err
	}
	h := sha256.New()
	if _, err = h.Write(data); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
