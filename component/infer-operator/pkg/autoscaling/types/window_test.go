/*
Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.

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

package types

import (
	"testing"
	"time"
)

func TestTimeWindowBucketsAndExpiry(t *testing.T) {
	now := time.Unix(100, 0)
	window := NewTimeWindow(10*time.Second, time.Second)
	window.Record(now, 2)
	window.Record(now.Add(time.Second), 4)
	window.Record(now.Add(1500*time.Millisecond), 8)
	if got := window.Avg(); got != 5 || window.Size() != 2 {
		t.Fatalf("average=%v size=%d", got, window.Size())
	}
	window.Record(now.Add(12*time.Second), 6)
	if window.Size() != 1 || window.Avg() != 6 {
		t.Fatalf("expired window was not pruned")
	}
}
