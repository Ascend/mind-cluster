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

// Package mount — provider_test.go
//
// Tests for Provider interface and Build.
package mount

import (
	"errors"
	"testing"

	cdispec "tags.cncf.io/container-device-interface/specs-go"
)

func TestBuild_ProviderError(t *testing.T) {
	_, err := Build(&errorProvider{err: errors.New("provider failure")})
	if err == nil {
		t.Fatal("expected error from provider.GetMounts")
	}
}

type errorProvider struct{ err error }

func (p *errorProvider) GetMounts() ([]*cdispec.Mount, error) { return nil, p.err }
