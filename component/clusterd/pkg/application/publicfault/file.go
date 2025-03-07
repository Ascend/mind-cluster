// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package publicfault funcs about file for public fault
package publicfault

import (
	"context"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"

	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/publicfault"
)

// WatchPubFaultCustomFile watch file /user1/mindx-dl/clusterd/publicCustomization.json
func WatchPubFaultCustomFile(ctx context.Context) {
	eventCh, errCh, err := utils.GetFileWatcherChan(filepath.Dir(constant.PubFaultCustomizationPath))
	if err != nil {
		hwlog.RunLog.Errorf("get file watcher chan failed, error: %v", err)
		return
	}

	for {
		select {
		case _, ok := <-ctx.Done():
			if !ok {
				hwlog.RunLog.Info("stop signal channel closed")
			}
			hwlog.RunLog.Infof("listen file <%s> stop", constant.PubFaultCustomizationName)
			return
		case event, ok := <-eventCh:
			if !ok {
				hwlog.RunLog.Error("event channel is closed")
				return
			}
			if event.Name != constant.PubFaultCustomizationPath {
				continue
			}
			hwlog.RunLog.Infof("watch file %s event: %v", constant.PubFaultCustomizationName, event)
			if event.Op&fsnotify.Create == fsnotify.Create || event.Op&fsnotify.Write == fsnotify.Write ||
				event.Op&fsnotify.Chmod == fsnotify.Chmod {
				hwlog.RunLog.Infof("start trying load public fault config from file %s",
					constant.PubFaultCustomizationName)
				tryLoadPubFaultCfgFromFile()
			}
			if event.Op&fsnotify.Remove == fsnotify.Remove || event.Op&fsnotify.Rename == fsnotify.Rename {
				hwlog.RunLog.Infof("start trying load public fault config from file %s",
					constant.PubFaultCodeFileName)
				publicfault.LoadPubFaultCfgFromFile(constant.PubFaultCodeFilePath)
			}
			UpdateLimiter()
		case watchErr, ok := <-errCh:
			if !ok {
				hwlog.RunLog.Error("error channel is closed")
				return
			}
			hwlog.RunLog.Errorf("watch file %s failed, error: %v", constant.PubFaultCustomizationName, watchErr)
		}
	}
}

// if load new file failed, use original configuration
func tryLoadPubFaultCfgFromFile() {
	const retryTime = 3
	var loadSuc bool
	for i := 0; i < retryTime; i++ {
		if err := publicfault.LoadPubFaultCfgFromFile(constant.PubFaultCustomizationPath); err == nil {
			loadSuc = true
			break
		}
		time.Sleep(1 * time.Second)
	}
	if !loadSuc {
		hwlog.RunLog.Warnf("load fault config from <%s> failed, begin load from <%s>",
			constant.PubFaultCustomizationName, constant.PubFaultCodeFileName)
		if err := publicfault.LoadPubFaultCfgFromFile(constant.PubFaultCodeFilePath); err != nil {
			hwlog.RunLog.Errorf("load from <%s> failed, error: %v", constant.PubFaultCodeFileName, err)
			return
		}
	}
}
