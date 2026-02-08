#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# Copyright 2026 Huawei Technologies Co., Ltd
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
# ==============================================================================

from diag_tool.core.inspection.config.base import OpticalThreshold, InspectionConfig, BerThreshold

MayiConfig = InspectionConfig(
    hccs_swi_optical_threshold=OpticalThreshold(
        media_snr_error_threshold=20,
        txrx_power_threshold=-3.2,
        txrx_power_diff_threshold=3
    ),
    roce_swi_optical_threshold=OpticalThreshold(
        media_snr_error_threshold=19,
        txrx_power_threshold=-3.2,
        txrx_power_diff_threshold=3
    ),
    host_roce_optical_threshold=OpticalThreshold(
        media_snr_error_threshold=18,
        txrx_power_threshold=-4.2,
        txrx_power_diff_threshold=7
    ),
    ber_threshold=BerThreshold(
        operations_phase_err=2e-4,
        operations_phase_warn=5e-5,
        delivery_phase_err=5e-6
    )
)
