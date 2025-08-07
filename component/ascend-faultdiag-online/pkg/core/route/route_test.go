/*
Copyright(C)2025. Huawei Technologies Co.,Ltd. All rights reserved.

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

// Package route provides test cases for the router
package route

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"ascend-faultdiag-online/pkg/core/api"
)

func TestNewRouter(t *testing.T) {
	assert.NotNil(t, NewRouter())
}

func TestHandleApi(t *testing.T) {

	router := NewRouter()
	router.RootApi = &api.Api{}

	// non-exist api
	apiFunc, err := router.HandleApi("v1/invalid")
	assert.Error(t, err)
	assert.Nil(t, apiFunc)
}
