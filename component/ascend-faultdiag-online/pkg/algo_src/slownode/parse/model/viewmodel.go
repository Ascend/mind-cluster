/* Copyright 2025 Huawei Technologies Co., Ltd.

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

/*
Package model.
*/
package model

import "ascend-common/common-utils/hwlog"

// NameView is a view for filtering operator names
type NameView struct {
	// Name contains name information
	Name string
}

// HostDeviceDuration is a view for filtering operator duration, distinguishing between host and device sides
type HostDeviceDuration struct {
	// HostDuration is the duration of the operator on the host side
	HostDuration int64
	// DeviceDuration is the duration of the operator on the device side
	DeviceDuration int64
}

// Duration is a view for duration data
type Duration struct {
	// Dur represents duration data
	Dur int64
}

// StepStartEndNs is a view for step time (start and end)
type StepStartEndNs struct {
	// Id is the unique identifier
	Id int64
	// StartNs is the start timestamp
	StartNs int64
	// EndNs is the end timestamp
	EndNs int64
}

// StartEndNs is a view for time (start and end)
type StartEndNs struct {
	// StartNs is the start timestamp
	StartNs int64
	// EndNs is the end timestamp
	EndNs int64
}

// ValueView is a view for data values
type ValueView struct {
	// Value represents the data value
	Value string
}

// IdView is a view for identifiers
type IdView struct {
	// Id is the unique identifier
	Id int64
}

// StringIdsView is a view containing both identifier and value
type StringIdsView struct {
	// Id is the unique identifier
	Id int64
	// Value represents the data value
	Value string
}

// NameMapping 名称映射
func NameMapping(nameView *NameView) []any {
	if nameView == nil {
		hwlog.RunLog.Error("[SLOWNODE PARSE]invalid nil nameView")
		return nil
	}
	return []any{&nameView.Name}
}

// HdDurMapping HostDeviceDuration映射到指针
func HdDurMapping(hdDur *HostDeviceDuration) []any {
	if hdDur == nil {
		hwlog.RunLog.Error("[SLOWNODE PARSE]invalid nil hdDur")
		return nil
	}
	return []any{&hdDur.HostDuration, &hdDur.DeviceDuration}
}

// DurationMapping Duration映射到指针
func DurationMapping(duration *Duration) []any {
	if duration == nil {
		hwlog.RunLog.Error("[SLOWNODE PARSE]invalid nil duration")
		return nil
	}
	return []any{&duration.Dur}
}

// StepStartEndNsMapping StartEndNs映射到指针
func StepStartEndNsMapping(stepStartEndNs *StepStartEndNs) []any {
	if stepStartEndNs == nil {
		hwlog.RunLog.Error("[SLOWNODE PARSE]invalid nil stepStartEndNs")
		return nil
	}
	return []any{&stepStartEndNs.Id, &stepStartEndNs.StartNs, &stepStartEndNs.EndNs}
}

// StartEndNsMapping StartEndNs映射到指针
func StartEndNsMapping(startEndNs *StartEndNs) []any {
	if startEndNs == nil {
		hwlog.RunLog.Error("[SLOWNODE PARSE]invalid nil startEndNs")
		return nil
	}
	return []any{&startEndNs.StartNs, &startEndNs.EndNs}
}

// ValueViewMapping ValueView映射到指针
func ValueViewMapping(valueView *ValueView) []any {
	if valueView == nil {
		hwlog.RunLog.Error("[SLOWNODE PARSE]invalid nil valueView")
		return nil
	}
	return []any{&valueView.Value}
}

// IdViewMapping ValueView映射到指针
func IdViewMapping(idView *IdView) []any {
	if idView == nil {
		hwlog.RunLog.Error("[SLOWNODE PARSE]invalid nil idView")
		return nil
	}
	return []any{&idView.Id}
}

// IdMapping ValueView映射到指针
func IdMapping(id int64) []any {
	return []any{&id}
}

// StringIdsMapping StringIds映射到指针
func StringIdsMapping(stringIds *StringIdsView) []any {
	if stringIds == nil {
		hwlog.RunLog.Error("[SLOWNODE PARSE]invalid nil stringIds")
		return nil
	}
	return []any{&stringIds.Id, &stringIds.Value}
}
