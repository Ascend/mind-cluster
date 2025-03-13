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

// Package profiling contains functions that support dynamically collecting profiling data
package profiling

import (
	"encoding/json"

	"ascend-common/common-utils/hwlog"
)

// MsptiActivityKind The kinds of activity records.
// Each kind is associated with a activity record structure that holds the information associated with the kind.
type MsptiActivityKind int32

// MsptiActivitySourceKind The source kinds of mark data.
// Each mark activity record kind represents information about host or device
type MsptiActivitySourceKind int32

// MsptiActivityFlag Flags associated with activity records.
// Activity record flags. Flags can be combined by bitwise OR to associated multiple flags with an activity record.
// specific to a certain activity kind, as noted below.
type MsptiActivityFlag int32

type MsptiObjectID struct {
	// Pt A process object requires that we identify the process ID.
	// A thread object requires that we identify both the process and thread ID.
	Pt struct {
		ProcessId uint32
		ThreadId  uint32
	}
	// Ds A device object requires that we identify the device ID.
	// A context object requires that we identify both the device and
	// A stream object requires that we identify device and stream ID.
	Ds struct {
		DeviceId uint32
		StreamId uint32
	}
}

type MsptiActivity interface {
	Marshal() []byte
}

type MsptiActivityMark struct {
	// Kind The activity record kind, must be msptiActivityKindMarker.
	Kind MsptiActivityKind
	// Flag The flags associated with the marker MsptiActivityFlag
	Flag MsptiActivityFlag
	// SourceKind The source kinds of mark data.
	SourceKind MsptiActivitySourceKind
	// Timestamp The timestamp for the marker,
	// A value of 0 indicates that timestamp information could not be collected for the marker.
	Timestamp uint64
	// Id The marker ID
	Id uint64
	// MsptiObjectId The identifier for the activity object associated with this marker.
	// 'objectKind' indicates which ID is valid for this record.
	MsptiObjectId MsptiObjectID
	// Name The marker name for an instantaneous or start marker. This will be NULL for an end marker.
	Name *string
	// Domain The name of the domain to which this marker belongs to.This will be NULL for default domain.
	Domain *string
}

func (mark MsptiActivityMark) Marshal() []byte {
	name, domain := "NULL", "NULL"
	if mark.Name != nil {
		name = *mark.Name
	}
	if mark.Domain != nil {
		domain = *mark.Domain
	}
	mark.Name = &name
	mark.Domain = &domain
	bytes, err := json.Marshal(mark)
	if err != nil {
		hwlog.RunLog.Errorf("failed to marshal MsptiActivityMark, err:%s", err.Error())
		return []byte{}
	}
	return bytes
}

// MsptiActivityApi obtain api kind activity from mspti
type MsptiActivityApi struct {
	// Kind The activity record kind, must be msptiActivityKindApi.
	Kind MsptiActivityKind
	// Start The start timestamp for the api, in ns.
	Start uint64
	// End The end timestamp for the api, in ns.
	End uint64
	// Pt A thread object requires that we identify both the process and thread ID.
	Pt struct {
		ProcessId uint32
		ThreadId  uint32
	}
	// CorrelationId The correlation ID of the kernel.
	CorrelationId uint64
	// Name The api name.
	Name *string
}

func (api MsptiActivityApi) Marshal() []byte {
	name := "NULL"
	if api.Name != nil {
		name = *api.Name
	}
	api.Name = &name
	bytes, err := json.Marshal(api)
	if err != nil {
		hwlog.RunLog.Errorf("failed to marshal MsptiActivityApi, err:%s", err.Error())
		return []byte{}
	}
	return bytes
}

type MsptiActivityKernel struct {
	// Kind The activity record kind, must be msptiActivityKindKernel.
	Kind MsptiActivityKind
	// Start The start timestamp for the api, in ns.
	Start uint64
	// End The end timestamp for the api, in ns.
	End uint64
	// Ds A stream object requires that we identify device and stream ID.
	Ds struct {
		DeviceId uint32
		StreamId uint32
	}
	// CorrelationId The correlation ID of the kernel.
	CorrelationId uint64
	// Type The kernel type.
	Type *string
	// Name The kernel name.
	Name *string
}

func (kernel MsptiActivityKernel) Marshal() []byte {
	name, Type := "NULL", "NULL"
	if kernel.Name != nil {
		name = *kernel.Name
	}
	if kernel.Type != nil {
		Type = *kernel.Type
	}
	kernel.Name = &name
	kernel.Type = &Type
	bytes, err := json.Marshal(kernel)
	if err != nil {
		hwlog.RunLog.Errorf("failed to marshal MsptiActivityKernel, err:%s", err.Error())
		return []byte{}
	}
	return bytes
}
