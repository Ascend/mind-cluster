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

package configmanager

import (
	"testing"
	"time"
)

// TestCollectorIntervalSetGet covers the guard branches of SetCollectorInterval
// and the miss/fallback/sentinel paths of GetCollectorInterval, using isolated
// test keys that do not interfere with the real cache keys.
func TestCollectorIntervalSetGet(t *testing.T) {
	const keyA, keyB = "utIntervalKeyA", "utIntervalKeyB"
	defer func() {
		collectorIntervalMap.Delete(keyA)
		collectorIntervalMap.Delete(keyB)
	}()

	// empty key: Set is a no-op and Get returns the fallback
	SetCollectorInterval("", time.Second)
	if got := GetCollectorInterval("", 5*time.Second); got != 5*time.Second {
		t.Errorf("GetCollectorInterval(empty key) = %v, want fallback 5s", got)
	}

	// non-positive interval (except the collect-once sentinel) is not stored
	SetCollectorInterval(keyA, 0)
	SetCollectorInterval(keyA, -5*time.Second)
	if got := GetCollectorInterval(keyA, 5*time.Second); got != 5*time.Second {
		t.Errorf("non-positive interval stored, got %v, want fallback 5s", got)
	}

	// positive interval round-trip
	SetCollectorInterval(keyA, 7*time.Second)
	if got := GetCollectorInterval(keyA, 5*time.Second); got != 7*time.Second {
		t.Errorf("round-trip = %v, want 7s", got)
	}

	// CollectOnceInterval sentinel is stored and returned as-is
	SetCollectorInterval(keyA, CollectOnceInterval)
	if got := GetCollectorInterval(keyA, 5*time.Second); got != CollectOnceInterval {
		t.Errorf("collect-once = %v, want %v", got, CollectOnceInterval)
	}

	// missing key returns the fallback
	if got := GetCollectorInterval("utIntervalMissing", 9*time.Second); got != 9*time.Second {
		t.Errorf("missing key = %v, want fallback 9s", got)
	}

	// stored non-duration value falls back
	collectorIntervalMap.Store(keyB, "not-a-duration")
	if got := GetCollectorInterval(keyB, 9*time.Second); got != 9*time.Second {
		t.Errorf("non-duration value = %v, want fallback 9s", got)
	}

	// disabled (<= 0) stored value falls back
	collectorIntervalMap.Store(keyB, DisabledInterval)
	if got := GetCollectorInterval(keyB, 9*time.Second); got != 9*time.Second {
		t.Errorf("disabled value = %v, want fallback 9s", got)
	}
}

// TestIntervalConfigApply covers IntervalConfig.Apply and GetDpuListRefreshInterval.
func TestIntervalConfigApply(t *testing.T) {
	defer func() {
		SetCollectorInterval(CacheKeyHinicadm5, time.Duration(DefaultHinicadm5IntervalSeconds)*time.Second)
		SetCollectorInterval(CacheKeySysfs, time.Duration(DefaultSysfsIntervalSeconds)*time.Second)
		SetCollectorInterval(CacheKeyDpuListRefresh, time.Duration(DefaultDpuListRefreshIntervalSeconds)*time.Second)
	}()

	(&IntervalConfig{
		Hinicadm5CollectorInterval: 25,
		SysfsCollectorInterval:     26,
		DpuListRefreshInterval:     27,
	}).Apply()

	if got := GetCollectorInterval(CacheKeyHinicadm5, 0); got != 25*time.Second {
		t.Errorf("hinicadm5 interval = %v, want 25s", got)
	}
	if got := GetCollectorInterval(CacheKeySysfs, 0); got != 26*time.Second {
		t.Errorf("sysfs interval = %v, want 26s", got)
	}
	if got := GetDpuListRefreshInterval(0); got != 27*time.Second {
		t.Errorf("dpu list refresh interval = %v, want 27s", got)
	}

	// unconfigured key returns the fallback
	collectorIntervalMap.Delete(CacheKeyDpuListRefresh)
	if got := GetDpuListRefreshInterval(11 * time.Second); got != 11*time.Second {
		t.Errorf("dpu list refresh fallback = %v, want 11s", got)
	}
}

func TestValidateInterval(t *testing.T) {
	tests := []struct {
		name    string
		seconds int
		want    time.Duration
		wantErr bool
	}{
		{"collect-once", -1, CollectOnceInterval, false},
		{"min valid", 20, 20 * time.Second, false},
		{"max valid", 3600, 3600 * time.Second, false},
		{"zero", 0, 0, true},
		{"below min", 19, 0, true},
		{"above max", 3601, 0, true},
		{"negative two", -2, 0, true},
		{"far above", 86400, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateInterval(tt.seconds)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateInterval(%d) error = %v, wantErr %v", tt.seconds, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("ValidateInterval(%d) = %v, want %v", tt.seconds, got, tt.want)
			}
		})
	}
}

func TestApplyInvalidValueUsesDefault(t *testing.T) {
	defer func() {
		SetCollectorInterval(CacheKeyHinicadm5, time.Duration(DefaultHinicadm5IntervalSeconds)*time.Second)
	}()

	(&IntervalConfig{
		Hinicadm5CollectorInterval: 10,
	}).Apply()

	if got := GetCollectorInterval(CacheKeyHinicadm5, 0); got != time.Duration(DefaultHinicadm5IntervalSeconds)*time.Second {
		t.Errorf("invalid interval should use hinicadm5 default %ds, got %v",
			DefaultHinicadm5IntervalSeconds, got)
	}
}
