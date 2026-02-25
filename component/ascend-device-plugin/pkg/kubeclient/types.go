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

package kubeclient

// Event is the event that will be sent to the watcher
type Event struct {
	// ResourceType is the type of the resource that is being watched
	Resource ResourceType
	// Key is namespace/name
	Key string
	// Type is the type of the event
	Type EventType
}

// EventType string type of event
type EventType string

const (
	// EventTypeAdd is used when a new resource is created
	EventTypeAdd EventType = "add"
	// EventTypeUpdate is used when an existing resource is modified
	EventTypeUpdate EventType = "update"
	// EventTypeDelete is used when an existing resource is deleted
	EventTypeDelete EventType = "delete"
)

// ResourceType string type of resource
type ResourceType string

const (
	// PodResource is used when the event is related to a pod
	PodResource ResourceType = "pod"
	// CMResource is used when the event is related to a configmap
	CMResource ResourceType = "configmap"
)

const (
	DockerRuntime     = "docker"
	ContainerdRuntime = "containerd"
)
