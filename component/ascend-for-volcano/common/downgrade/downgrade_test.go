/*
Copyright(C)2026. Huawei Technologies Co.,Ltd. All rights reserved.

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
Package downgrade is using for the scheduler constraint downgrade cross-session
state.
*/
package downgrade

import (
	"sync"
	"testing"

	"volcano.sh/volcano/pkg/scheduler/api"
)

// the fixture configs are the opaque snapshots the store keeps, legacyConfig is
// the scalar the previous version of the feature wrote to the pod markers.
const (
	configSmall  = `{"sp-block":1}`
	configLarge  = `{"sp-block":2}`
	legacyConfig = `"16"`
)

func TestRecordDowngrade(t *testing.T) {
	ResetDowngradeCacheForTest()
	jobID := api.JobID("config-job")
	if _, ok := DowngradeConfig(jobID); ok {
		t.Errorf("downgrade config should not exist before record")
	}
	RecordDowngrade(jobID, configSmall, 160)
	if config, ok := DowngradeConfig(jobID); !ok || config != configSmall {
		t.Errorf("downgrade config = %q, %v, want %q, true", config, ok, configSmall)
	}
	if effectTime, ok := DowngradeEffectTime(jobID); !ok || effectTime != 160 {
		t.Errorf("downgrade effect time = %d, %v, want 160, true", effectTime, ok)
	}
	// the record follows the settled snapshot of the policy
	RecordDowngrade(jobID, configLarge, 320)
	if config, ok := DowngradeConfig(jobID); !ok || config != configLarge {
		t.Errorf("downgrade config = %q, %v, want %q, true", config, ok, configLarge)
	}
	if effectTime, ok := DowngradeEffectTime(jobID); !ok || effectTime != 320 {
		t.Errorf("downgrade effect time = %d, %v, want 320, true", effectTime, ok)
	}
	RecordDowngrade(jobID, "", 480)
	if config, ok := DowngradeConfig(jobID); !ok || config != configLarge {
		t.Errorf("an empty config should not overwrite the record, got %q, %v", config, ok)
	}
	if effectTime, ok := DowngradeEffectTime(jobID); !ok || effectTime != 320 {
		t.Errorf("an empty config should not overwrite the effect time, got %d, %v", effectTime, ok)
	}
	// the record keeps the config as the policy serialized it: the shape check
	// lives on the seed path only, the store takes the handler bytes as handed
	RecordDowngrade(jobID, legacyConfig, 480)
	if config, ok := DowngradeConfig(jobID); !ok || config != legacyConfig {
		t.Errorf("the record should keep the config as handed over, got %q, %v", config, ok)
	}
}

func TestDowngradeState(t *testing.T) {
	ResetDowngradeCacheForTest()
	jobID := api.JobID("state-job")
	if _, _, ok := DowngradeState(jobID); ok {
		t.Errorf("an empty store should hold no state")
	}
	if _, _, ok := DowngradeState(""); ok {
		t.Errorf("an empty job should hold no state")
	}
	RecordDowngrade(jobID, configSmall, 160)
	config, effectTime, ok := DowngradeState(jobID)
	if !ok || config != configSmall || effectTime != 160 {
		t.Errorf("downgrade state = %q, %d, %v, want %q, 160, true", config, effectTime, ok, configSmall)
	}
	// both halves follow the settled record together
	RecordDowngrade(jobID, configLarge, 320)
	config, effectTime, ok = DowngradeState(jobID)
	if !ok || config != configLarge || effectTime != 320 {
		t.Errorf("downgrade state = %q, %d, %v, want %q, 320, true", config, effectTime, ok, configLarge)
	}
	// the narrow accessors read the same record
	if got, ok := DowngradeConfig(jobID); !ok || got != config {
		t.Errorf("downgrade config = %q, %v, want %q, true", got, ok, config)
	}
	if got, ok := DowngradeEffectTime(jobID); !ok || got != effectTime {
		t.Errorf("downgrade effect time = %d, %v, want %d, true", got, ok, effectTime)
	}
	ResetConstraint(jobID)
	if _, _, ok := DowngradeState(jobID); ok {
		t.Errorf("a reset job should hold no state")
	}
}

func TestMarkEffective(t *testing.T) {
	ResetDowngradeCacheForTest()
	jobID := api.JobID("effective-job")
	if HasState(jobID) || IsEffective(jobID) {
		t.Errorf("an empty store should hold no state and no effective job")
	}

	RecordDowngrade(jobID, configSmall, 160)
	if !HasState(jobID) || IsEffective(jobID) {
		t.Errorf("a job which is still deepening holds a state but no effective config")
	}
	MarkEffective(jobID)
	if !HasState(jobID) || !IsEffective(jobID) {
		t.Errorf("an effective job holds a state and the effective mark")
	}
	if config, ok := DowngradeConfig(jobID); !ok || config != configSmall {
		t.Errorf("mark effective should keep the config, got %q, %v", config, ok)
	}
	if effectTime, ok := DowngradeEffectTime(jobID); !ok || effectTime != 160 {
		t.Errorf("mark effective should keep the effect time, got %d, %v", effectTime, ok)
	}

	// marking twice is harmless
	MarkEffective(jobID)
	if !IsEffective(jobID) {
		t.Errorf("mark effective should stay idempotent")
	}

	// a job without a downgrade config is not affected
	plainJob := api.JobID("plain-job")
	MarkEffective(plainJob)
	if IsEffective(plainJob) {
		t.Errorf("a job without a config can not be marked effective")
	}
}

func TestResetConstraint(t *testing.T) {
	ResetDowngradeCacheForTest()
	jobID := api.JobID("reset-job")
	RecordDowngrade(jobID, configSmall, 160)
	MarkEffective(jobID)

	ResetConstraint(jobID)
	if _, ok := DowngradeConfig(jobID); ok {
		t.Errorf("reset constraint should drop the downgrade config")
	}
	if IsEffective(jobID) {
		t.Errorf("reset constraint should drop the effective mark")
	}
	if HasState(jobID) {
		t.Errorf("reset constraint should drop the whole record")
	}
}

func TestSeedDowngrade(t *testing.T) {
	ResetDowngradeCacheForTest()
	jobID := api.JobID("seeded-job")
	SeedDowngrade(jobID, configSmall, 160)
	if config, ok := DowngradeConfig(jobID); !ok || config != configSmall {
		t.Errorf("seeded downgrade config = %q, %v, want %q, true", config, ok, configSmall)
	}
	if effectTime, ok := DowngradeEffectTime(jobID); !ok || effectTime != 160 {
		t.Errorf("seeded downgrade effect time = %d, %v, want 160, true", effectTime, ok)
	}
	if !IsEffective(jobID) {
		t.Errorf("a seeded config comes from scheduled pods, it is effective by construction")
	}

	RecordDowngrade(jobID, configLarge, 320)
	if config, _ := DowngradeConfig(jobID); config != configLarge {
		t.Errorf("record downgrade should overwrite the seeded config, got %q", config)
	}
	SeedDowngrade(jobID, configSmall, 640)
	if config, _ := DowngradeConfig(jobID); config != configLarge {
		t.Errorf("seed should not overwrite the existing state, got %q", config)
	}

	invalidJob := api.JobID("invalid-seed-job")
	SeedDowngrade(invalidJob, configSmall, 0)
	if _, ok := DowngradeConfig(invalidJob); ok {
		t.Errorf("a seed without an effect time should be ignored")
	}
}

type seedValidationTest struct {
	name   string
	config string
	want   bool
}

func buildSeedValidationTestCases() []seedValidationTest {
	return []seedValidationTest{
		{name: "01 a JSON object is accepted", config: `{"sp-block":2}`, want: true},
		{name: "02 an empty JSON object is accepted", config: `{}`, want: true},
		{name: "03 the scalar value of the previous version is refused", config: legacyConfig},
		{name: "04 a bare number is refused", config: "16"},
		{name: "05 a JSON null is refused", config: "null"},
		{name: "06 a JSON array is refused", config: "[16]"},
		{name: "07 a truncated JSON is refused", config: `{"sp-block":`},
		{name: "08 an empty value is refused", config: ""},
	}
}

// TestSeedDowngradeValidatesTheMarker covers the store trust boundary: a config
// which is not a JSON object is refused and the downgrade state rebuilds from zero.
func TestSeedDowngradeValidatesTheMarker(t *testing.T) {
	for _, tt := range buildSeedValidationTestCases() {
		t.Run(tt.name, func(t *testing.T) {
			ResetDowngradeCacheForTest()
			jobID := api.JobID("seed-validation-job")

			SeedDowngrade(jobID, tt.config, 160)

			config, ok := DowngradeConfig(jobID)
			if ok != tt.want {
				t.Errorf("seeded config = %q, %v, want seeded %v", config, ok, tt.want)
			}
		})
	}
}

func TestSweep(t *testing.T) {
	ResetDowngradeCacheForTest()

	leftJob := api.JobID("left-job")
	keptJob := api.JobID("kept-job")
	RecordDowngrade(leftJob, configLarge, 100)
	RecordDowngrade(keptJob, configSmall, 200)
	Sweep(func(jobID api.JobID) bool { return jobID == keptJob })

	if _, ok := DowngradeConfig(leftJob); ok {
		t.Errorf("sweep should remove the downgrade config of the dead job")
	}
	if _, ok := DowngradeEffectTime(leftJob); ok {
		t.Errorf("sweep should remove the downgrade effect time of the dead job")
	}
	if config, ok := DowngradeConfig(keptJob); !ok || config != configSmall {
		t.Errorf("sweep should keep the downgrade config of the alive job, got %q, %v", config, ok)
	}
	if effectTime, ok := DowngradeEffectTime(keptJob); !ok || effectTime != 200 {
		t.Errorf("sweep should keep the effect time of the alive job, got %d, %v", effectTime, ok)
	}
}

// TestStampSession verifies the snapshot carries the all-waiting fact of the job and
// is rebuilt from scratch by every stamp, the level of the wait being a pure function
// of it (no mark of an already taken step is stored).
func TestStampSession(t *testing.T) {
	ResetDowngradeCacheForTest()
	waitingJob := api.JobID("stamp-waiting-job")
	partialJob := api.JobID("stamp-partial-job")

	StampSession(map[api.JobID]bool{waitingJob: true, partialJob: false})

	if !SessionAllWaiting(waitingJob) {
		t.Errorf("the all-waiting fact should come from the stamp input")
	}
	if SessionAllWaiting(partialJob) {
		t.Errorf("the all-waiting fact should come from the stamp input")
	}
	if SessionAllWaiting(api.JobID("unknown-job")) {
		t.Errorf("an unknown job is not all waiting")
	}

	// the next session rebuilds the snapshot from scratch
	StampSession(map[api.JobID]bool{waitingJob: false})
	if SessionAllWaiting(waitingJob) {
		t.Errorf("the new session should read the new all-waiting fact")
	}
	if SessionAllWaiting(partialJob) {
		t.Errorf("the previous session snapshot should be dropped")
	}
}

// TestEmptyJobSafe verifies an empty job id is a no-op on every entry of the
// package api.
func TestEmptyJobSafe(t *testing.T) {
	ResetDowngradeCacheForTest()
	RecordDowngrade("", configSmall, 160)
	SeedDowngrade("", configSmall, 160)
	MarkEffective("")
	ResetConstraint("")
	StampSession(map[api.JobID]bool{"": true})
	Sweep(func(api.JobID) bool { return true })
	if _, ok := DowngradeConfig(""); ok {
		t.Errorf("an empty job should hold no downgrade config")
	}
	if _, ok := DowngradeEffectTime(""); ok {
		t.Errorf("an empty job should hold no downgrade effect time")
	}
	if SessionAllWaiting("") {
		t.Errorf("an empty job should not be all waiting")
	}
	if HasState("") || IsEffective("") {
		t.Errorf("an empty job should hold no state and no effective config")
	}
}

// TestConcurrentAccess verifies the store survives concurrent writers: the
// lock serializes the read-modify-write of the downgrade state.
func TestConcurrentAccess(t *testing.T) {
	ResetDowngradeCacheForTest()
	jobID := api.JobID("concurrent-job")
	const workerNum = 16
	var waitGroup sync.WaitGroup
	for i := 0; i < workerNum; i++ {
		waitGroup.Add(2)
		go func() {
			defer waitGroup.Done()
			RecordDowngrade(jobID, configSmall, 100)
		}()
		go func() {
			defer waitGroup.Done()
			HasState(jobID)
		}()
	}
	waitGroup.Wait()
	config, ok := DowngradeConfig(jobID)
	if !ok || config != configSmall {
		t.Errorf("concurrent records should settle one config, got %q, %v", config, ok)
	}
	if effectTime, ok := DowngradeEffectTime(jobID); !ok || effectTime != 100 {
		t.Errorf("concurrent records should settle one effect time, got %d, %v", effectTime, ok)
	}
}
