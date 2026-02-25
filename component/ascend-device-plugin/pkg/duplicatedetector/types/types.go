/* Copyright(C) 2024. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package types defines types used in the duplicate detector
package types

import (
	"time"
)

// ContainerNPUInfo stores NPU device information for a container
type ContainerNPUInfo struct {
	ID        string // Container ID
	Name      string // Container name
	Namespace string // Container namespace
	PodName   string // Pod name (if available)
	PodNS     string // Pod namespace (if available)
	Devices   []int  // NPU device IDs mounted to this container
}

// DuplicateMountInfo represents a duplicate mount scenario
type DuplicateMountInfo struct {
	DeviceID   int
	Containers []*ContainerNPUInfo
}

// DetectorConfig contains configuration for the duplicate detector
type DetectorConfig struct {
	// CriEndpoint is the containerd or docker socket endpoint (e.g., unix:///run/containerd/containerd.sock)
	CriEndpoint string

	// RuntimeType is the runtime type used by the containers (e.g., docker, containerd)
	RuntimeType string
}

// ContainerEventType represents the type of container event
type ContainerEventType string

const (
	// ContainerEventCreate represents container creation event
	ContainerEventCreate ContainerEventType = "create"
	// ContainerEventDestroy represents container destroy event
	ContainerEventDestroy ContainerEventType = "destroy"
)

// ContainerEvent represents a container lifecycle event
type ContainerEvent struct {
	Type        ContainerEventType
	ContainerID string
	Namespace   string
	Timestamp   time.Time
}

// EventHandler is a callback function for container events
type EventHandler func(event ContainerEvent)
