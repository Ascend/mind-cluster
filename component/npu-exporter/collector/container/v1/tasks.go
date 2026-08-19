/* Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package v1 implement the containerd client
package v1

import (
	"context"
	"fmt"
	"math"

	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = fmt.Errorf
var _ = math.Inf
var _ = proto.Marshal

// ListPidsRequest is the request message for the containerd Tasks.ListPids RPC.
type ListPidsRequest struct {
	// ContainerId is the ID of the container whose running process PIDs are requested.
	ContainerId          string   `protobuf:"bytes,1,opt,name=container_id,json=containerId,proto3" json:"container_id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

// Reset resets the object
func (m *ListPidsRequest) Reset() { *m = ListPidsRequest{} }

// String
func (m *ListPidsRequest) String() string { return proto.CompactTextString(m) }

// ProtoMessage
func (*ListPidsRequest) ProtoMessage() {}

// GetContainerId returns the container ID.
func (m *ListPidsRequest) GetContainerId() string {
	if m != nil {
		return m.ContainerId
	}
	return ""
}

// ProcessInfo describes a process running in a container task.
type ProcessInfo struct {
	// Pid is the process ID on the host.
	Pid                  uint32   `protobuf:"varint,1,opt,name=pid,proto3" json:"pid,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

// Reset resets the object
func (m *ProcessInfo) Reset() { *m = ProcessInfo{} }

// String
func (m *ProcessInfo) String() string { return proto.CompactTextString(m) }

// ProtoMessage
func (*ProcessInfo) ProtoMessage() {}

// GetPid returns the process ID.
func (m *ProcessInfo) GetPid() uint32 {
	if m != nil {
		return m.Pid
	}
	return 0
}

// ListPidsResponse is the response message for the containerd Tasks.ListPids RPC.
type ListPidsResponse struct {
	// Processes are the processes running in the requested container.
	Processes            []*ProcessInfo `protobuf:"bytes,1,rep,name=processes,proto3" json:"processes,omitempty"`
	XXX_NoUnkeyedLiteral struct{}       `json:"-"`
	XXX_unrecognized     []byte         `json:"-"`
	XXX_sizecache        int32          `json:"-"`
}

// Reset resets the object
func (m *ListPidsResponse) Reset() { *m = ListPidsResponse{} }

// String
func (m *ListPidsResponse) String() string { return proto.CompactTextString(m) }

// ProtoMessage
func (*ListPidsResponse) ProtoMessage() {}

// GetProcesses returns the processes running in the container.
func (m *ListPidsResponse) GetProcesses() []*ProcessInfo {
	if m != nil {
		return m.Processes
	}
	return nil
}

func init() {
	proto.RegisterType((*ListPidsRequest)(nil), "containerd.services.tasks.v1.ListPidsRequest")
	proto.RegisterType((*ListPidsResponse)(nil), "containerd.services.tasks.v1.ListPidsResponse")
	proto.RegisterType((*ProcessInfo)(nil), "containerd.services.tasks.v1.ProcessInfo")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConnInterface

// TasksClient is the client API for the containerd Tasks service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type TasksClient interface {
	// ListPids lists the process PIDs running in the given container.
	ListPids(ctx context.Context, in *ListPidsRequest, opts ...grpc.CallOption) (*ListPidsResponse, error)
}

type tasksClient struct {
	cc grpc.ClientConnInterface
}

// NewTasksClient creates a Tasks client on the given connection.
func NewTasksClient(cc grpc.ClientConnInterface) TasksClient {
	return &tasksClient{cc}
}

func (c *tasksClient) ListPids(ctx context.Context, in *ListPidsRequest, opts ...grpc.CallOption) (*ListPidsResponse, error) {
	out := new(ListPidsResponse)
	err := c.cc.Invoke(ctx, "/containerd.services.tasks.v1.Tasks/ListPids", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}
