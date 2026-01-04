/*
 * Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * 		http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"fmt"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/google/uuid"
	resourceapi "k8s.io/api/resource/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"

	"ascend/dra-example-driver/pkg/consts"
)

func enumerateAllPossibleDevices(numNPUs int) (AllocatableDevices, error) {
	seed := os.Getenv("NODE_NAME")
	uuids := generateUUIDs(seed, numNPUs)

	alldevices := make(AllocatableDevices)
	for i, uuid := range uuids {
		device := resourceapi.Device{
			Name: fmt.Sprintf("npu-%d", i),
			Attributes: map[resourceapi.QualifiedName]resourceapi.DeviceAttribute{
				"index": {
					IntValue: ptr.To(int64(i)),
				},
				"uuid": {
					StringValue: ptr.To(uuid),
				},
				"model": {
					StringValue: ptr.To("LATEST-NPU-MODEL"),
				},
				"driverVersion": {
					VersionValue: ptr.To("0.1.0"),
				},
			},
			Capacity: map[resourceapi.QualifiedName]resourceapi.DeviceCapacity{
				"memory": {
					Value: resource.MustParse("80Gi"),
				},
			},
		}
		alldevices[device.Name] = device
	}
	klog.Infoln("cmd/dra-npu-kubeletplugin/discovery.go::enumerateAllPossibleDevices [alldevices]:")
	spew.Dump(alldevices)
	return alldevices, nil
}
