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

import "testing"

func newWhitelist(defaults, custom []string) *WhitelistManager {
	w := &WhitelistManager{
		defaultWhitelist: buildSet(defaults),
		customWhitelist:  buildSet(custom),
	}
	w.recalcTotal()
	return w
}

func TestBuildSet(t *testing.T) {
	s := buildSet([]string{"a", " b", "", "c", "a"})
	if len(s) != 3 {
		t.Fatalf("expected 3, got %d", len(s))
	}
	if _, ok := s["b"]; !ok {
		t.Error("trimmed 'b' not found")
	}
}

func TestMatchSet_ExactAndPrefix(t *testing.T) {
	s := buildSet([]string{"roce_err_ctr_foo", "roce_dp_ctr_*"})

	tests := []struct {
		name   string
		expect bool
	}{
		{"roce_err_ctr_foo", true},
		{"roce_dp_ctr_bar", true},
		{"roce_rx_ctr_baz", false},
		{"other", false},
	}
	for _, tt := range tests {
		if got := matchSet(tt.name, s); got != tt.expect {
			t.Errorf("matchSet(%q) = %v, want %v", tt.name, got, tt.expect)
		}
	}
}

func TestIsAllowed_DefaultAndCustom(t *testing.T) {
	w := newWhitelist([]string{"roce_err_ctr_*"}, nil)
	if !w.IsAllowed("roce_err_ctr_foo") {
		t.Error("default match failed")
	}
	if w.IsAllowed("roce_rx_foo") {
		t.Error("should not match default")
	}

	w2 := newWhitelist([]string{"roce_err_ctr_*"}, []string{"roce_dp_ctr_*"})
	if w2.IsAllowed("roce_err_ctr_foo") {
		t.Error("custom takes precedence; default should be ignored")
	}
	if !w2.IsAllowed("roce_dp_ctr_bar") {
		t.Error("custom match failed")
	}
}

func TestFilter(t *testing.T) {
	w := newWhitelist([]string{"roce_err_ctr_*"}, nil)
	in := map[string]float64{
		"roce_err_ctr_foo": 1,
		"roce_rx_bar":      2,
	}
	out := w.Filter(in)
	if len(out) != 1 {
		t.Fatalf("expected 1, got %d", len(out))
	}
	if _, ok := out["roce_err_ctr_foo"]; !ok {
		t.Error("filtered metric missing")
	}
}

func TestGetCount(t *testing.T) {
	w := newWhitelist([]string{"a", "b"}, []string{"c"})
	if w.GetCount() != 1 {
		t.Errorf("custom takes precedence; count should be 1, got %d", w.GetCount())
	}
	w2 := newWhitelist([]string{"a", "b"}, nil)
	if w2.GetCount() != 2 {
		t.Errorf("default count should be 2, got %d", w2.GetCount())
	}
}

func TestLoadDefaultAndLoadCustom(t *testing.T) {
	w := newWhitelist(nil, nil)
	w.LoadDefault([]string{"x_*"})
	if !w.IsAllowed("x_y") {
		t.Error("LoadDefault failed")
	}
	w.LoadCustom([]string{"z_*"})
	if !w.IsAllowed("z_y") {
		t.Error("LoadCustom failed")
	}
	if w.IsAllowed("x_y") {
		t.Error("custom should override default")
	}
}

func TestReload(t *testing.T) {
	w := newWhitelist(nil, nil)
	if err := w.Reload([]string{"r_*"}); err != nil {
		t.Fatal(err)
	}
	if !w.IsAllowed("r_val") {
		t.Error("Reload failed")
	}
}
