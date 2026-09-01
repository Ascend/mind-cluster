/*
 * Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at

 * http://www.apache.org/licenses/LICENSE-2.0

 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package flags

import (
	"flag"
	"testing"
)

// withFreshFlagSet replaces the global flag.CommandLine with an empty FlagSet
// for the duration of the test. All flag registrations in the flags package
// target flag.CommandLine, so this prevents "flag redefined" panics when
// several tests register the same flags within one test binary.
func withFreshFlagSet(t *testing.T) {
	t.Helper()
	orig := flag.CommandLine
	flag.CommandLine = flag.NewFlagSet(t.Name(), flag.ContinueOnError)
	t.Cleanup(func() {
		flag.CommandLine = orig
	})
}

// flagCheck is a single named equality assertion on a config field.
type flagCheck struct {
	name string
	got  interface{}
	want interface{}
}

// runFlagChecks asserts every check and reports mismatches with check names.
func runFlagChecks(t *testing.T, checks []flagCheck) {
	t.Helper()
	for _, c := range checks {
		if c.got != c.want {
			t.Errorf("%s = %v, want %v", c.name, c.got, c.want)
		}
	}
}
