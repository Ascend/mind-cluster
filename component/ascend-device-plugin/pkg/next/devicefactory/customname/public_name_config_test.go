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

// Package customname a series of device public name test function
package customname

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
)

func init() {
	hwLogConfig := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	err := hwlog.InitRunLogger(&hwLogConfig, context.Background())
	if err != nil {
		fmt.Printf("log init failed, error is %v\n", err)
	}
}

// TestReplaceDevicePublicName tests ReplaceDevicePublicName
func TestReplaceDevicePublicName(t *testing.T) {
	resourceType := api.Ascend910
	oldName := "ascend910-1"

	devNameMap[resourceType] = DevName{
		ResourceType:           resourceType,
		DevicePublicNamePre:    "custom910-",
		OldDevicePublicNamePre: api.Ascend910MinuxPrefix,
	}

	newName := ReplaceDevicePublicName(resourceType, oldName)
	assert.Equal(t, oldName, newName, "when OldDevicePublicNamePre in devNameMap does not match oldName, "+
		"original name should be returned")
	oldName = api.Ascend910MinuxPrefix + "1"
	newName = ReplaceDevicePublicName(resourceType, oldName)
	assert.Equal(t, "custom910-1", newName,
		"when OldDevicePublicNamePre in devNameMap matches oldName, replaced name should be returned")
	delete(devNameMap, resourceType)
	newName = ReplaceDevicePublicName(resourceType, oldName)
	assert.Equal(t, oldName, newName,
		"when corresponding resourceType does not exist in devNameMap, original name should be returned")
}

// TestReplaceDeviceInnerName tests ReplaceDeviceInnerName
func TestReplaceDeviceInnerName(t *testing.T) {
	resourceType := api.Ascend310
	oldNames := []string{"Ascend310-1", "Ascend310-2"}

	devNameMap[resourceType] = DevName{
		ResourceType:           resourceType,
		DevicePublicNamePre:    "custom310-",
		OldDevicePublicNamePre: api.Ascend310MinuxPrefix,
	}

	newNames := ReplaceDeviceInnerName(resourceType, oldNames)
	assert.Equal(t, oldNames, newNames, "when OldDevicePublicNamePre in devNameMap does not match "+
		"oldNames, original name list should be returned")

	oldNames = []string{"custom310-1", "custom310-2"}
	newNames = ReplaceDeviceInnerName(resourceType, oldNames)
	assert.Equal(t, []string{"Ascend310-1", "Ascend310-2"}, newNames,
		"when OldDevicePublicNamePre in devNameMap matches oldNames, "+
			"replaced name list should be returned")
}

// TestReplaceDevicePublicType tests ReplaceDevicePublicType
func TestReplaceDevicePublicType(t *testing.T) {
	resourceType := api.Ascend310P
	oldName := "huawei.com/Ascend310P"

	devNameMap[resourceType] = DevName{
		ResourceType:        resourceType,
		DevicePublicType:    "CustomAscend310P",
		OldDevicePublicType: api.HuaweiAscend310P,
	}

	newName := ReplaceDevicePublicType(resourceType, oldName)
	assert.Equal(t, "CustomAscend310P", newName,
		"when OldDevicePublicType in devNameMap matches oldName, replaced type should be returned")

	delete(devNameMap, resourceType)
	newName = ReplaceDevicePublicType(resourceType, oldName)
	assert.Equal(t, oldName, newName,
		"when corresponding resourceType does not exist in devNameMap, original type should be returned")
}

// TestReplacePodAnnotation tests ReplacePodAnnotation
func TestReplacePodAnnotation(t *testing.T) {
	resourceType := api.Ascend910
	annotation := map[string]string{
		api.HuaweiAscend910: "value1",
		"other-key":         "value3",
	}

	devNameMap[resourceType] = DevName{
		ResourceType:           resourceType,
		DevicePublicType:       "CustomAscend910",
		OldDevicePublicType:    api.HuaweiAscend910,
		DevicePublicNamePre:    "custom910-",
		OldDevicePublicNamePre: api.Ascend910MinuxPrefix,
	}

	newAnnotation := ReplacePodAnnotation(resourceType, annotation)

	assert.Contains(t, newAnnotation, "CustomAscend910", "device type key should be replaced")
	assert.Contains(t, newAnnotation, "other-key", "other keys should not be modified")
}

// TestReplaceDeviceInfoPublicName tests ReplaceDeviceInfoPublicName
func TestReplaceDeviceInfoPublicName(t *testing.T) {
	resourceType := api.Ascend310
	deviceList := map[string]string{
		api.HuaweiAscend310: "value1",
		"other-key":         "value2",
	}
	deviceName := api.Ascend310MinuxPrefix + "1"

	// Set up mock devNameMap
	devNameMap[resourceType] = DevName{
		ResourceType:           resourceType,
		DevicePublicType:       "CustomAscend310",
		OldDevicePublicType:    api.HuaweiAscend310,
		DevicePublicNamePre:    "custom310-",
		OldDevicePublicNamePre: api.Ascend310MinuxPrefix,
	}

	newDeviceList, newDeviceName := ReplaceDeviceInfoPublicName(resourceType, deviceList, deviceName)

	assert.Contains(t, newDeviceList, "CustomAscend310", "device type key should be replaced")
	assert.Equal(t, "custom310-1", newDeviceName, "device name prefix should be replaced")
}

// TestCheckName tests CheckName
func TestCheckName(t *testing.T) {
	validDevNames := []DevName{
		{
			ResourceType:        api.Ascend910,
			DevicePublicType:    "CustomAscend910",
			DevicePublicNamePre: "custom910",
		},
	}
	assert.True(t, checkName(validDevNames), "valid names should pass the check")
	invalidDevNames1 := []DevName{
		{
			ResourceType:        "InvalidType",
			DevicePublicType:    "CustomAscend910",
			DevicePublicNamePre: "custom910-",
		},
	}
	assert.False(t, checkName(invalidDevNames1), "invalid resource types should fail the check")

	invalidDevNames2 := []DevName{
		{
			ResourceType:        api.Ascend910,
			DevicePublicType:    "inv@lid",
			DevicePublicNamePre: "custom910-",
		},
	}
	assert.False(t, checkName(invalidDevNames2), "invalid device public type format should fail the check")

	invalidDevNames3 := []DevName{
		{
			ResourceType:        api.Ascend910,
			DevicePublicType:    "CustomAscend910-",
			DevicePublicNamePre: "inv@lid",
		},
	}
	assert.False(t, checkName(invalidDevNames3), "invalid device public name prefix format should fail the check")

	invalidDevNames4 := []DevName{
		{
			ResourceType:        api.Ascend910,
			DevicePublicType:    "",
			DevicePublicNamePre: "",
		},
	}
	assert.False(t, checkName(invalidDevNames4), "all name should not be null")
}

// TestSetDefaultName tests SetDefaultName
func TestSetDefaultName(t *testing.T) {
	devName910 := DevName{
		ResourceType: api.Ascend910,
	}
	devName910 = setDefaultName(devName910)
	assert.Equal(t, api.HuaweiAscend910, devName910.OldDevicePublicType)
	assert.Equal(t, api.Ascend910MinuxPrefix, devName910.OldDevicePublicNamePre)

	devName310 := DevName{
		ResourceType: api.Ascend310,
	}
	devName310 = setDefaultName(devName310)
	assert.Equal(t, api.HuaweiAscend310, devName310.OldDevicePublicType)
	assert.Equal(t, api.Ascend310MinuxPrefix, devName310.OldDevicePublicNamePre)

	devName310P := DevName{
		ResourceType: api.Ascend310P,
	}
	devName310P = setDefaultName(devName310P)
	assert.Equal(t, api.HuaweiAscend310P, devName310P.OldDevicePublicType)
	assert.Equal(t, api.Ascend310PMinuxPrefix, devName310P.OldDevicePublicNamePre)
}
