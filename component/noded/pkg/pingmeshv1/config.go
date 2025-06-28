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
Package pingmeshv1 is using for checking hccs network
*/
package pingmeshv1

import (
	"ascend-common/common-utils/hwlog"
	"nodeD/pkg/kubeclient"
)

const (
	// DefaultResultMaxAge is the default result max age
	DefaultResultMaxAge = hwlog.DefaultMinSaveAge
	// MinResultMaxAge is the minimum result max age
	MinResultMaxAge = hwlog.DefaultMinSaveAge
	// MaxResultMaxAge is the maximum result max age
	MaxResultMaxAge = hwlog.DefaultMaxSaveAge
	superPodCMKey   = "superPodDevice"
	globalConfigKey = "global"
)

// Config is the configuration for pingmeshv1
type Config struct {
	ResultMaxAge int `json:"result_max_age"`
	KubeClient   *kubeclient.ClientK8s
}
