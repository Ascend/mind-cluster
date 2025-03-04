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

// Package main implements online fault diagnosis.
package main

import (
	"fmt"

	apiv1 "ascend-faultdiag-online/pkg/api/v1"
)

func main() {
	ctx, err := apiv1.CreateFdCtx("D:\\MyRepo\\mind-cluster\\component\\ascend-faultdiag-online\\test.yaml")
	if err != nil {
		fmt.Println("%v", err)
		return
	}
	apiv1.StartService(ctx)
	resp, err := apiv1.Request(ctx, "metric/add", "{"+
		""+
		"}")
	if err != nil {
		fmt.Println("%v", err)
		return
	}
	fmt.Println(resp)
}
