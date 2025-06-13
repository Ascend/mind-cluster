/* Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package pingmesh for
package pingmesh

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	"ascend-common/api/slownet"
	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
	"clusterd/pkg/common/constant"
)

const (
	suppressedPeriod = 0
	networkType      = 1
	pingType         = 0
	pingTimes        = 5
	pingInterval     = 1
	period           = 15

	startIndex        = 1
	confFileRetryTime = 3
)

// NewCathelperConf new CathelperConf info
func NewCathelperConf() constant.CathelperConf {
	return constant.CathelperConf{
		SuppressedPeriod: suppressedPeriod,
		NetworkType:      networkType,
		PingType:         pingType,
		PingTimes:        pingTimes,
		PingInterval:     pingInterval,
		Period:           period,
	}
}

func saveConfigToFile(superpodID string, conf *constant.CathelperConf) error {
	if conf == nil {
		hwlog.RunLog.Errorf("config is nil")
		return fmt.Errorf("config is nil")
	}
	rasConfig := NewCathelperConf()
	path, err := slownet.GetConfigPathForDetect(superpodID)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get config path, err: %v", err)
		return err
	}
	err = writeConfigToFile(&rasConfig, path)
	if err != nil {
		hwlog.RunLog.Errorf("failed to save config to file, err: %v", err)
		return err
	}
	return nil
}

func writeConfigToFile(rasConfig *constant.CathelperConf, filePath string) error {
	if _, errInfo := utils.CheckPath(filePath); errInfo != nil {
		hwlog.RunLog.Errorf("file path is invalid, err: %v", errInfo)
		return errInfo
	}

	var file *os.File = nil
	defer func(file *os.File) {
		if file == nil {
			return
		}
		errClose := file.Close()
		if errClose != nil {
			hwlog.RunLog.Errorf("close file failed, err: %v", errClose)
		}
	}(file)
	for i := startIndex; i <= confFileRetryTime; i++ {
		var err error
		file, err = os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, defaultPerm)
		if err == nil {
			break
		}
		hwlog.RunLog.Errorf("failed to create or overwrite file: %v and retry times : %d", err, i)
		if i == confFileRetryTime {
			return err
		}
		time.Sleep(time.Second)
	}
	if err := writeToFile(rasConfig, file); err != nil {
		return err
	}
	hwlog.RunLog.Infof("write cathelper.conf file to path<%s> success", filePath)
	return nil
}

func writeToFile(rasConfig *constant.CathelperConf, file *os.File) error {
	if file == nil || rasConfig == nil {
		return fmt.Errorf("the file pointer or config paramters are nil")
	}

	val := reflect.ValueOf(rasConfig).Elem()
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		jsonTag := fieldType.Tag.Get("json")
		fieldName := strings.Split(jsonTag, ",")[0]

		line := fmt.Sprintf("%s=%v\n", fieldName, field.Interface())
		if _, err := file.WriteString(line); err != nil {
			hwlog.RunLog.Errorf("failed to write to file: %v", err)
			return err
		}
	}
	return nil
}
