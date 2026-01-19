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

// Package customname a series of device public name function
package customname

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
)

const nameConfigFilePath = "/usr/local/deviceNameCustomization.json"

var deviceTypePattern = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9./]{8,30}[a-zA-Z0-9]$`)
var deviceNamePattern = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9-]{1,15}$`)
var podConfigurationPattern = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9./-]{8,61}[a-zA-Z0-9]$`)

var devNameMap = map[string]DevName{}

// DevName dev public struct
type DevName struct {
	// ResourceType the original name of the device
	ResourceType string
	// DevicePublicType device
	DevicePublicType        string
	OldDevicePublicType     string
	DevicePublicNamePre     string
	OldDevicePublicNamePre  string
	PodConfigurationName    string
	OldPodConfigurationName string
}

// InitPublicNameConfig init public name config
func InitPublicNameConfig() {
	devNames, err := loadFaultCodeFromFile()
	if err != nil {
		hwlog.RunLog.Infof("do not use the custom device name, because %v", err)
		return
	}
	if !checkName(devNames) {
		hwlog.RunLog.Warn("the custom name configuration is invalid")
		return
	}
	for _, devName := range devNames {
		devName = setDefaultName(devName)
		devNameMap[devName.ResourceType] = devName
	}
	hwlog.RunLog.Infof("the custom name takes effect: %v", devNameMap)
}

func setDefaultName(devName DevName) DevName {
	switch devName.ResourceType {
	case api.Ascend910:
		devName.OldDevicePublicType = api.HuaweiAscend910
		devName.OldDevicePublicNamePre = api.Ascend910MinuxPrefix
		devName.OldPodConfigurationName = api.Pod910DeviceAnno
	case api.Ascend310:
		devName.OldDevicePublicType = api.HuaweiAscend310
		devName.OldDevicePublicNamePre = api.Ascend310MinuxPrefix
	case api.Ascend310P:
		devName.OldDevicePublicType = api.HuaweiAscend310P
		devName.OldDevicePublicNamePre = api.Ascend310PMinuxPrefix
	default:
		hwlog.RunLog.Errorf("devicePublicType(%s) is invalid ", devName.DevicePublicType)
	}
	return devName
}

// ReplaceDevicePublicName replace device name with public name
func ReplaceDevicePublicName(resourceType string, oldName string) string {
	devName := devNameMap[resourceType]
	if len(devName.DevicePublicNamePre) == 0 {
		return oldName
	}
	return strings.ReplaceAll(oldName, devName.OldDevicePublicNamePre, devName.DevicePublicNamePre)
}

// ReplaceDeviceInnerName replace device name with inner name
func ReplaceDeviceInnerName(resourceType string, oldNames []string) []string {
	devName := devNameMap[resourceType]
	if len(devName.DevicePublicNamePre) == 0 {
		return oldNames
	}
	newNames := make([]string, 0, len(oldNames))
	for _, oldName := range oldNames {
		newName := strings.ReplaceAll(oldName, devName.DevicePublicNamePre, devName.OldDevicePublicNamePre)
		newNames = append(newNames, newName)
	}
	return newNames
}

// ReplaceDevicePublicType replace device type with public type
func ReplaceDevicePublicType(resourceType string, oldName string) string {
	devName := devNameMap[resourceType]
	if len(devName.DevicePublicType) == 0 {
		return oldName
	}
	return strings.ReplaceAll(oldName, devName.OldDevicePublicType, devName.DevicePublicType)
}

// ReplacePodAnnotation replace pod annotation name and value with public name
func ReplacePodAnnotation(resourceType string, annotation map[string]string) map[string]string {
	devName := devNameMap[resourceType]
	if len(devName.DevicePublicNamePre) == 0 {
		return annotation
	}
	newAnnotation := make(map[string]string, len(annotation))
	for key, value := range annotation {
		if key == devName.OldDevicePublicType {
			key = devName.DevicePublicType
		}
		if key == devName.OldPodConfigurationName {
			key = devName.PodConfigurationName
		}
		strings.ReplaceAll(value, devName.OldDevicePublicNamePre, devName.DevicePublicNamePre)
		newAnnotation[key] = value
	}
	return newAnnotation
}

func ReplaceDeviceInfoPublicName(resourceType string, deviceList map[string]string,
	deviceName string) (map[string]string, string) {
	devName := devNameMap[resourceType]
	if len(devName.DevicePublicNamePre) == 0 {
		return deviceList, deviceName
	}
	newDeviceList := make(map[string]string, len(deviceList))
	for key, value := range deviceList {
		newKey := strings.ReplaceAll(key, devName.OldDevicePublicType, devName.DevicePublicType)
		newValue := strings.ReplaceAll(value, devName.OldDevicePublicNamePre, devName.DevicePublicNamePre)
		newDeviceList[newKey] = newValue
	}
	newDeviceName := strings.ReplaceAll(deviceName, devName.OldDevicePublicNamePre, devName.DevicePublicNamePre)
	return newDeviceList, newDeviceName
}

func checkName(devNames []DevName) bool {
	for _, devName := range devNames {
		if len(devName.DevicePublicNamePre) == 0 || len(devName.ResourceType) == 0 ||
			len(devName.DevicePublicType) == 0 {
			hwlog.RunLog.Error("all name should not be null")
			return false
		}
		if devName.ResourceType != api.Ascend310 && devName.ResourceType != api.Ascend310P &&
			devName.ResourceType != api.Ascend910 {
			hwlog.RunLog.Errorf("resourceType only support %s, %s, %s",
				api.Ascend310, api.Ascend310P, api.Ascend910)
			return false
		}
		if !deviceTypePattern.MatchString(devName.DevicePublicType) {
			hwlog.RunLog.Errorf("devicePublicType(%s) is invalid", devName.DevicePublicType)
			return false
		}
		if !deviceNamePattern.MatchString(devName.DevicePublicNamePre) {
			hwlog.RunLog.Errorf("devicePublicNamePre(%s) is invalid", devName.DevicePublicNamePre)
			return false
		}
		if len(devName.PodConfigurationName) != 0 &&
			!podConfigurationPattern.MatchString(devName.PodConfigurationName) {
			hwlog.RunLog.Errorf("podConfigurationName(%s) is invalid", devName.PodConfigurationName)
			return false
		}
	}
	return true
}

func loadFaultCodeFromFile() ([]DevName, error) {
	faultCodeBytes, err := utils.LoadFile(nameConfigFilePath)
	if err != nil {
		return nil, fmt.Errorf("load name config json failed: %v", err)
	}
	var devNames []DevName
	if err := json.Unmarshal(faultCodeBytes, &devNames); err != nil {
		return nil, fmt.Errorf("unmarshal fault code byte failed: %v", err)
	}
	return devNames, nil
}
