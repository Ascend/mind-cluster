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
	"strings"
	"sync"
)

// WhitelistManager manages metric whitelists for DPU collectors.
// It allows filtering the 300+ hinicadm5 metrics down to only the desired ones.
// The whitelist can be hot-reloaded at runtime.
type WhitelistManager struct {
	mu               sync.RWMutex
	defaultWhitelist map[string]struct{}
	customWhitelist  map[string]struct{}
	totalCount       int
}

var (
	whitelistInstance *WhitelistManager
	whitelistOnce     sync.Once
)

// DefaultWhitelistPatterns is the built-in default whitelist for DPU metrics.
// These patterns use prefix matching (trailing * means "match any suffix").
// The parsed metric names are lowercased, e.g. "roce_cmdq_ctr_ext_cmd".
var DefaultWhitelistPatterns = []string{
	"roce_rx_*",
	"roce_tx_*",
	"roce_cmdq_*",
	"roce_cq_*",
	"roce_qp_*",
	"roce_mr_*",
	"roce_srq_*",
	"roce_ext_*",
}

// GetWhitelist returns the singleton WhitelistManager instance with default whitelist loaded
func GetWhitelist() *WhitelistManager {
	whitelistOnce.Do(func() {
		whitelistInstance = &WhitelistManager{
			defaultWhitelist: buildSet(DefaultWhitelistPatterns),
			customWhitelist:  make(map[string]struct{}),
		}
		whitelistInstance.recalcTotal()
	})
	return whitelistInstance
}

// LoadDefault loads the built-in default whitelist patterns
func (w *WhitelistManager) LoadDefault(patterns []string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.defaultWhitelist = buildSet(patterns)
	w.recalcTotal()
}

// LoadCustom loads user-provided whitelist patterns from config file
func (w *WhitelistManager) LoadCustom(patterns []string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.customWhitelist = buildSet(patterns)
	w.recalcTotal()
}

// IsAllowed checks if a metric name matches any pattern in the active whitelist.
// If customWhitelist is non-empty, it takes precedence; otherwise defaultWhitelist is used.
// Supports exact match and prefix match (pattern ending with *).
func (w *WhitelistManager) IsAllowed(metricName string) bool {
	w.mu.RLock()
	defer w.mu.RUnlock()

	if len(w.customWhitelist) > 0 {
		return matchSet(metricName, w.customWhitelist)
	}
	return matchSet(metricName, w.defaultWhitelist)
}

// Filter removes metrics not in the active whitelist from the given map.
// If customWhitelist is non-empty, it takes precedence; otherwise defaultWhitelist is used.
func (w *WhitelistManager) Filter(metrics map[string]float64) map[string]float64 {
	w.mu.RLock()
	defer w.mu.RUnlock()

	active := w.defaultWhitelist
	if len(w.customWhitelist) > 0 {
		active = w.customWhitelist
	}

	filtered := make(map[string]float64, len(metrics))
	for name, val := range metrics {
		if matchSet(name, active) {
			filtered[name] = val
		}
	}
	return filtered
}

// GetCount returns the number of patterns in the combined whitelist
func (w *WhitelistManager) GetCount() int {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.totalCount
}

// Reload reloads the custom whitelist from the given patterns (convenience for hot-reload)
func (w *WhitelistManager) Reload(patterns []string) error {
	w.LoadCustom(patterns)
	return nil
}

func (w *WhitelistManager) recalcTotal() {
	if len(w.customWhitelist) > 0 {
		w.totalCount = len(w.customWhitelist)
	} else {
		w.totalCount = len(w.defaultWhitelist)
	}
}

func buildSet(patterns []string) map[string]struct{} {
	s := make(map[string]struct{}, len(patterns))
	for _, p := range patterns {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		s[p] = struct{}{}
	}
	return s
}

func matchSet(metricName string, s map[string]struct{}) bool {
	if _, ok := s[metricName]; ok {
		return true
	}
	for pattern := range s {
		if strings.HasSuffix(pattern, "*") {
			prefix := strings.TrimSuffix(pattern, "*")
			if strings.HasPrefix(metricName, prefix) {
				return true
			}
		}
	}
	return false
}
