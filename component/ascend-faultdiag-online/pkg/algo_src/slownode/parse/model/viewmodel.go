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

/*
Package model.
*/
package model

// NameView 筛选算子的名称
type NameView struct {
	Name string
}

// HostDeviceDuration 筛选算子的耗时视图
type HostDeviceDuration struct {
	HostDuration   int64
	DeviceDuration int64
}

// Duration 耗时视图
type Duration struct {
	Dur int64
}

// StepStartEndNs step时间开始结束视图
type StepStartEndNs struct {
	Id      int64
	StartNs int64
	EndNs   int64
}

// StartEndNs 时间开始结束视图
type StartEndNs struct {
	StartNs int64
	EndNs   int64
}

// ValueView 值视图
type ValueView struct {
	Value string
}

// IdView 值视图
type IdView struct {
	Id int64
}

// StringIdsView Id和Value视图
type StringIdsView struct {
	Id    int64
	Value string
}

// NameMapping 名称映射
func NameMapping(nameView *NameView) []any {
	return []any{&nameView.Name}
}

// HdDurMapping HostDeviceDuration映射到指针
func HdDurMapping(hdDur *HostDeviceDuration) []any {
	return []any{&hdDur.HostDuration, &hdDur.DeviceDuration}
}

// DurationMapping Duration映射到指针
func DurationMapping(duration *Duration) []any {
	return []any{&duration.Dur}
}

// StepStartEndNsMapping StartEndNs映射到指针
func StepStartEndNsMapping(stepStartEndNs *StepStartEndNs) []any {
	return []any{&stepStartEndNs.Id, &stepStartEndNs.StartNs, &stepStartEndNs.EndNs}
}

// StartEndNsMapping StartEndNs映射到指针
func StartEndNsMapping(startEndNs *StartEndNs) []any {
	return []any{&startEndNs.StartNs, &startEndNs.EndNs}
}

// ValueViewMapping ValueView映射到指针
func ValueViewMapping(valueView *ValueView) []any {
	return []any{&valueView.Value}
}

// IdViewMapping ValueView映射到指针
func IdViewMapping(idView *IdView) []any {
	return []any{&idView.Id}
}

// IdMapping ValueView映射到指针
func IdMapping(id int64) []any {
	return []any{&id}
}

// StringIdsMapping StringIds映射到指针
func StringIdsMapping(stringIds *StringIdsView) []any {
	return []any{&stringIds.Id, &stringIds.Value}
}
