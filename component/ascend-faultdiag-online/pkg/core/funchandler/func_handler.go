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

/*
Package funchandler load and use the func provide by algo_src
*/
package funchandler

import (
	"fmt"

	"ascend-faultdiag-online/pkg/core/model"
)

// ExecuteFunc
type ExecuteFunc func(model.Input) (int, error)

// Handler is the struct to manage the function
type Handler struct {
	// FuncType type of the func
	FuncType string
	// FuncVersion version of the func
	FuncVersion string
	// ExecuteFunc the execute function
	ExecuteFunc ExecuteFunc
}

// GenerateExecuteFunc returns the execute function
func GenerateExecuteFunc(f func(model.Input) int, app string) ExecuteFunc {
	return func(input model.Input) (int, error) {
		ret := f(input)
		if ret != 0 {
			return -1, fmt.Errorf("call [%s] func [Execute] failed, return code: [%d]", app, ret)
		}
		return 0, nil
	}
}
