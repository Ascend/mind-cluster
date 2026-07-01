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

// Package hccn this for npu hccn info
package hccn

import (
	"reflect"
	"strings"
	"testing"

	"ascend-common/devmanager/common"
)

// hccnToolSeparator is the table separator line of `hccn_tool -g -dev_info` output.
const hccnToolSeparator = "+--------+--------+--------+----------+-------------+------------+"

// hccnToolHeader is the table header line of `hccn_tool -g -dev_info` output.
const hccnToolHeader = "| UdieID | PortID | Speed  | PortType | Link Status | Media Type |"

// hccnToolTable builds a `hccn_tool -g -dev_info` output snippet from the given
// data rows. It always emits: separator, header, separator, <data rows>,
// separator — which mirrors the real command output and lets each test case
// only describe its own data rows with proper code indentation.
func hccnToolTable(dataRows ...string) string {
	lines := []string{
		hccnToolSeparator,
		hccnToolHeader,
		hccnToolSeparator,
	}
	lines = append(lines, dataRows...)
	lines = append(lines, hccnToolSeparator)
	return strings.Join(lines, "\n")
}

// sampleDevInfoOutput mimics the real output of `hccn_tool -g -dev_info -i 0`,
// including the header, separator lines and mixed port types / link statuses.
var sampleDevInfoOutput = hccnToolTable(
	"| 0      | 4      | 200    | ETH      | DOWN        | Electrical |",
	"| 0      | 5      | 200    | ETH      | UP          | Electrical |",
	"| 1      | 8      | 200    | UB       | DOWN        | Optical    |",
	"| 1      | 9      | 200    | UB       | UP          | Optical    |",
)

// TestGetAllUBPortsFromHccnLines verifies that GetAllUBports can correctly parse
// the output of `hccn_tool -g -dev_info -i 0`.
func TestGetAllUBPortsFromHccnLines(t *testing.T) {
	tests := []struct {
		name    string
		output  string
		want    []common.UBPort
		wantErr bool
	}{
		{
			name:   "01-parse mixed ETH and UB ports with up/down status",
			output: sampleDevInfoOutput,
			want: []common.UBPort{
				{UDieId: 0, PortID: 4, PortType: BondingPortName, LinkStatus: LinkDown},
				{UDieId: 0, PortID: 5, PortType: BondingPortName, LinkStatus: LinkUp},
				{UDieId: 1, PortID: 8, PortType: UBPortName, LinkStatus: LinkDown},
				{UDieId: 1, PortID: 9, PortType: UBPortName, LinkStatus: LinkUp},
			},
			wantErr: false,
		},
		{
			name:    "02-only header and separators, no data rows",
			output:  hccnToolTable(),
			want:    []common.UBPort{},
			wantErr: false,
		},
		{
			name: "03-data row with too few columns should return error",
			output: hccnToolTable(
				"| 0      | 4      | 200    | ETH      | DOWN        |"),
			want:    nil,
			wantErr: true,
		},
		{
			name: "04-data row with non-integer UdieID should return error",
			output: hccnToolTable(
				"| ab     | 4      | 200    | ETH      | DOWN        | Electrical |"),
			want:    nil,
			wantErr: true,
		},
		{
			name:    "05-empty output should return empty ports without error",
			output:  "",
			want:    []common.UBPort{},
			wantErr: false,
		},
		{
			name: "06-stop parsing at the third separator and ignore trailing data",
			output: hccnToolTable(
				"| 0      | 4      | 200    | ETH      | DOWN        | Electrical |",
				"| 1      | 8      | 200    | UB       | DOWN        | Optical    |",
			),
			want: []common.UBPort{
				{UDieId: 0, PortID: 4, PortType: BondingPortName, LinkStatus: LinkDown},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lines := strings.Split(tt.output, "\n")
			got, err := getAllUBPortsFromHccnLines(lines)
			if (err != nil) != tt.wantErr {
				t.Errorf("getAllUBPortsFromHccnLines() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getAllUBPortsFromHccnLines() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestParseDeviceTable tests the parseDeviceTable function.
func TestParseDeviceTable(t *testing.T) {
	tests := []struct {
		name    string
		lines   []string
		want    map[int][]int
		wantErr bool
	}{
		{
			name:  "parse mixed ETH and UB ports",
			lines: strings.Split(sampleDevInfoOutput, "\n"),
			want:  map[int][]int{0: {4, 5}, 1: {8, 9}},
		},
		{
			name:    "stop at third separator",
			lines:   strings.Split(hccnToolTable("| 0 | 4 | 200 | ETH | DOWN | Electrical |"), "\n"),
			want:    map[int][]int{0: {4}},
			wantErr: false,
		},
		{
			name:    "empty lines",
			lines:   []string{},
			wantErr: true,
		},
		{
			name:    "only header no data",
			lines:   strings.Split(hccnToolTable(), "\n"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseDeviceTable(tt.lines)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseDeviceTable() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseDeviceTable() = %v, want %v", got, tt.want)
			}
		})
	}
}
