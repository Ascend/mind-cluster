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

// Package profiling contains functions that support dynamically collecting profiling data
package profiling

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"
)

func TestMsptiActivityApiMarshal(t *testing.T) {
	t.Run("name is not null", func(t *testing.T) {
		name := "testName"
		msp := MsptiActivityApi{Name: &name}
		res := msp.Marshal()
		msp2 := MsptiActivityApi{}
		err := json.Unmarshal(res, &msp2)
		assert.NoError(t, err)
		assert.Equal(t, name, *msp2.Name)
	})

	t.Run("marshal err", func(t *testing.T) {
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		patches.ApplyFunc(json.Marshal, func(v any) ([]byte, error) {
			return nil, errors.New("marshal error")
		})
		name := "testName"
		msp := MsptiActivityApi{Name: &name}
		res := msp.Marshal()
		assert.Equal(t, 0, len(res))
	})
}

func TestMsptiActivityKernelMarshal(t *testing.T) {
	t.Run("name and type is not null", func(t *testing.T) {
		name := "testName"
		typeName := "testType"
		msp := MsptiActivityKernel{Name: &name, Type: &typeName}
		res := msp.Marshal()
		msp2 := MsptiActivityKernel{}
		err := json.Unmarshal(res, &msp2)
		assert.NoError(t, err)
		assert.Equal(t, name, *msp2.Name)
		assert.Equal(t, typeName, *msp2.Type)
	})

	t.Run("marshal err", func(t *testing.T) {
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		patches.ApplyFunc(json.Marshal, func(v any) ([]byte, error) {
			return nil, errors.New("marshal error")
		})
		name := "testName"
		msp := MsptiActivityKernel{Name: &name}
		res := msp.Marshal()
		assert.Equal(t, 0, len(res))
	})
}

func TestMsptiActivityMarkMarshal(t *testing.T) {
	t.Run("name and domain is not null", func(t *testing.T) {
		name := "testName"
		domainName := "testDomain"
		msp := MsptiActivityMark{Name: &name, Domain: &domainName}
		res := msp.Marshal()
		msp2 := MsptiActivityMark{}
		err := json.Unmarshal(res, &msp2)
		assert.NoError(t, err)
		assert.Equal(t, name, *msp2.Name)
		assert.Equal(t, domainName, *msp2.Domain)
	})

	t.Run("marshal err", func(t *testing.T) {
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		patches.ApplyFunc(json.Marshal, func(v any) ([]byte, error) {
			return nil, errors.New("marshal error")
		})
		name := "testName"
		msp := MsptiActivityMark{Name: &name}
		res := msp.Marshal()
		assert.Equal(t, 0, len(res))
	})
}
