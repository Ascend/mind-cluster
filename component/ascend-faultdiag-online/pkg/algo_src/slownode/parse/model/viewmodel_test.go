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

// Package model is a DT collection for functions in model
package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNameMapping(t *testing.T) {
	var nameView = NameView{Name: "test"}
	assert.ElementsMatch(t, []any{&nameView.Name}, NameMapping(&nameView))
}

func TestHdDurMapping(t *testing.T) {
	hd := HostDeviceDuration{
		HostDuration:   123,
		DeviceDuration: 321,
	}
	assert.ElementsMatch(t, []any{&hd.HostDuration, &hd.DeviceDuration}, HdDurMapping(&hd))
}

func TestDurationMapping(t *testing.T) {
	duration := Duration{Dur: 123}
	assert.ElementsMatch(t, []any{&duration.Dur}, DurationMapping(&duration))
}

func TestStepStartEndNsMapping(t *testing.T) {
	step := StepStartEndNs{Id: 1, StartNs: 123, EndNs: 1234}
	assert.ElementsMatch(t, []any{&step.Id, &step.StartNs, &step.EndNs}, StepStartEndNsMapping(&step))
}

func TestStartEndNsMapping(t *testing.T) {
	startEndNs := StartEndNs{StartNs: 123, EndNs: 1234}
	assert.ElementsMatch(t, []any{&startEndNs.StartNs, &startEndNs.EndNs}, StartEndNsMapping(&startEndNs))
}

func TestValueViewMapping(t *testing.T) {
	valueView := ValueView{Value: "test"}
	assert.ElementsMatch(t, []any{&valueView.Value}, ValueViewMapping(&valueView))
}

func TestIdViewMapping(t *testing.T) {
	idView := IdView{Id: 1}
	assert.ElementsMatch(t, []any{&idView.Id}, IdViewMapping(&idView))
}

func TestStringIdsMapping(t *testing.T) {
	stringIds := StringIdsView{Id: 1, Value: "name"}
	assert.ElementsMatch(t, []any{&stringIds.Id, &stringIds.Value}, StringIdsMapping(&stringIds))
}
