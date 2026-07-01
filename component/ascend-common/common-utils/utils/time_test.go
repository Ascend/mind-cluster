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

// Package utils this file for time utils
package utils

import (
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
)

const (
	fiveSeconds  = 5 * time.Second
	longDuration = 20 * 365 * 24 * time.Hour
)

func TestNowMono(t *testing.T) {
	tests := []struct {
		name    string
		elapsed time.Duration
		want    int64
	}{
		{
			name:    "should return zero when no time has elapsed",
			elapsed: 0,
			want:    0,
		},
		{
			name:    "should return elapsed nanos when seconds have passed",
			elapsed: fiveSeconds,
			want:    int64(fiveSeconds),
		},
		{
			name:    "should return large nanos when long duration has passed",
			elapsed: longDuration,
			want:    int64(longDuration),
		},
	}
	for _, tt := range tests {
		convey.Convey(tt.name, t, func() {
			patches := gomonkey.ApplyFuncReturn(time.Since, tt.elapsed)
			defer patches.Reset()
			convey.So(NowMono(), convey.ShouldEqual, tt.want)
		})
	}
}
