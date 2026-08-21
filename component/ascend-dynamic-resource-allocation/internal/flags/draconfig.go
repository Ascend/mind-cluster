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

import "ascend-common/common-utils/healthz"

// DRAConfig aggregates all DRA plugin configuration sections.
type DRAConfig struct {
	HwLogConfig      *HWLoggingConfig
	DraOption        *DRAOption
	KubeClientConfig *KubeClientConfig
	DraHealthzConfig *healthz.Config
}

// NewDraConfig creates a DRAConfig with default section instances.
func NewDraConfig() *DRAConfig {
	return &DRAConfig{
		HwLogConfig:      &HWLoggingConfig{},
		DraOption:        &DRAOption{},
		KubeClientConfig: &KubeClientConfig{},
		DraHealthzConfig: healthz.RegisterFlags(),
	}
}

// RegisterFlags registers all DRA config flags.
func (d *DRAConfig) RegisterFlags() {
	d.HwLogConfig.RegisterFlags()
	d.DraOption.RegisterFlags()
	d.KubeClientConfig.RegisterFlags()
}
