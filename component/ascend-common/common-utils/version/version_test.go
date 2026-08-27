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

// Package version for version test
package version

import (
	"errors"
	"math/rand"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
)

func TestGet(t *testing.T) {
	convey.Convey("Test Get function", t, func() {
		convey.Convey("some fields empty, use runtime defaults", func() {
			// Set test values with some empty
			Version = "v2.0.0"
			GitCommit = "xyz789"
			GitBranch = "develop"
			BuildOS = ""
			BuildArch = ""
			GoVersion = ""

			info := Get()
			convey.So(info.Version, convey.ShouldEqual, "v2.0.0")
			convey.So(info.GitCommit, convey.ShouldEqual, "xyz789")
			convey.So(info.GitBranch, convey.ShouldEqual, "develop")
			convey.So(info.BuildOS, convey.ShouldNotBeEmpty)
			convey.So(info.BuildArch, convey.ShouldNotBeEmpty)
			convey.So(info.GoVersion, convey.ShouldNotBeEmpty)
		})
	})
}

func TestToJSON(t *testing.T) {
	convey.Convey("Test ToJSON function", t, func() {
		convey.Convey("serialize Info struct to JSON", func() {
			testInfo := Info{
				Version:   "v1.0.0",
				GitCommit: "abc123",
				GitBranch: "main",
				BuildOS:   "linux",
				BuildArch: "amd64",
				GoVersion: "go1.21",
			}

			jsonStr := ToJSON(testInfo)
			convey.So(jsonStr, convey.ShouldNotBeEmpty)
			convey.So(jsonStr, convey.ShouldContainSubstring, `"version":"v1.0.0"`)
			convey.So(jsonStr, convey.ShouldContainSubstring, `"gitCommit":"abc123"`)
			convey.So(jsonStr, convey.ShouldContainSubstring, `"gitBranch":"main"`)
			convey.So(jsonStr, convey.ShouldContainSubstring, `"buildOS":"linux"`)
			convey.So(jsonStr, convey.ShouldContainSubstring, `"buildArch":"amd64"`)
			convey.So(jsonStr, convey.ShouldContainSubstring, `"goVersion":"go1.21"`)
		})
	})
}

type mockAnnotationAdder struct {
	keys   []string
	values []string
	err    error
	calls  int
	failN  int // fail the first failN calls
}

func (m *mockAnnotationAdder) AddAnnotation(key, value string) error {
	m.calls++
	m.keys = append(m.keys, key)
	m.values = append(m.values, value)
	if m.failN > 0 {
		m.failN--
		return m.err
	}
	return nil
}

func TestReportVersionToNodeAnnotation(t *testing.T) {
	testInfo := Info{
		Version:   "v1.0.0",
		GitCommit: "abc123",
		GitBranch: "main",
		BuildOS:   "linux",
		BuildArch: "amd64",
		GoVersion: "go1.21",
	}
	testComponent := "test-component"
	expectedKey := versionAnnotationKeyPrefix + testComponent + ".version"
	expectedValue := ToJSON(testInfo)

	convey.Convey("Test ReportVersionToNodeAnnotation", t, func() {
		convey.Convey("client is nil, return error immediately", func() {
			err := ReportVersionToNodeAnnotation(nil, testInfo, testComponent)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldContainSubstring, "k8s client is nil")
		})

		convey.Convey("AddAnnotation succeeds on first attempt", func() {
			mc := &mockAnnotationAdder{}
			patches := gomonkey.NewPatches()
			defer patches.Reset()
			// Stub time.Sleep (jitter + backoff) to avoid real sleeping
			patches.ApplyFunc(time.Sleep, func(d time.Duration) {})
			// Stub rand.Intn to return 0 to minimize jitter
			patches.ApplyFunc(rand.Intn, func(n int) int { return 0 })

			err := ReportVersionToNodeAnnotation(mc, testInfo, testComponent)

			convey.So(err, convey.ShouldBeNil)
			convey.So(mc.calls, convey.ShouldEqual, 1)
			convey.So(mc.keys[0], convey.ShouldEqual, expectedKey)
			convey.So(mc.values[0], convey.ShouldEqual, expectedValue)
		})

		convey.Convey("AddAnnotation fails twice then succeeds on third attempt", func() {
			mc := &mockAnnotationAdder{failN: 2, err: errors.New("annotation failed")}
			patches := gomonkey.NewPatches()
			defer patches.Reset()
			patches.ApplyFunc(time.Sleep, func(d time.Duration) {})
			patches.ApplyFunc(rand.Intn, func(n int) int { return 0 })

			err := ReportVersionToNodeAnnotation(mc, testInfo, testComponent)

			convey.So(err, convey.ShouldBeNil)
			convey.So(mc.calls, convey.ShouldEqual, 3)
			convey.So(mc.keys[2], convey.ShouldEqual, expectedKey)
			convey.So(mc.values[2], convey.ShouldEqual, expectedValue)
		})

		convey.Convey("AddAnnotation always fails, return wrapped error after max attempts", func() {
			mc := &mockAnnotationAdder{failN: 100, err: errors.New("annotation always fails")}
			patches := gomonkey.NewPatches()
			defer patches.Reset()
			patches.ApplyFunc(time.Sleep, func(d time.Duration) {})
			patches.ApplyFunc(rand.Intn, func(n int) int { return 0 })

			err := ReportVersionToNodeAnnotation(mc, testInfo, testComponent)

			convey.So(err, convey.ShouldNotBeNil)
			convey.So(mc.calls, convey.ShouldEqual, VersionReportMaxAttempts)
			convey.So(err.Error(), convey.ShouldContainSubstring, testComponent)
			convey.So(err.Error(), convey.ShouldContainSubstring, "annotation always fails")
		})

		convey.Convey("jitter sleep is invoked before first attempt", func() {
			mc := &mockAnnotationAdder{}
			patches := gomonkey.NewPatches()
			defer patches.Reset()
			sleepCount := 0
			patches.ApplyFunc(time.Sleep, func(d time.Duration) { sleepCount++ })
			patches.ApplyFunc(rand.Intn, func(n int) int { return 100 })

			err := ReportVersionToNodeAnnotation(mc, testInfo, testComponent)

			convey.So(err, convey.ShouldBeNil)
			// jitter sleep + 0 backoff sleeps (since first attempt succeeds)
			convey.So(sleepCount, convey.ShouldBeGreaterThanOrEqualTo, 1)
			convey.So(mc.calls, convey.ShouldEqual, 1)
		})
	})
}
