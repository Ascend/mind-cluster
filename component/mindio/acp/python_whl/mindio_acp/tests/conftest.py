#!/usr/bin/env python
# coding=utf-8
# Copyright (c) Huawei Technologies Co., Ltd. 2024. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
import os
import sys
import atexit
import torch
import pytest
from unittest.mock import MagicMock

sys.modules['mindspeed'] = MagicMock()
sys.modules['mindspeed.patch_utils'] = MagicMock()
sys.modules['mindspeed.patch_utils.Patch'] = MagicMock()

sys.modules['megatron'] = MagicMock()
sys.modules['megatron.core'] = MagicMock()
sys.modules['megatron.core.optimizer'] = MagicMock()
sys.modules['megatron.core.parallel_state'] = MagicMock()
sys.modules['megatron.core.rerun_state_machine'] = MagicMock()
sys.modules['megatron.core.num_microbatches_calculator'] = MagicMock()
sys.modules['megatron.training'] = MagicMock()
sys.modules['megatron.training.utils'] = MagicMock()
sys.modules['megatron.training.checkpointing'] = MagicMock()
sys.modules['megatron.fp16_deprecated'] = MagicMock()
sys.modules['megatron.fp16_deprecated.loss_scaler'] = MagicMock()
sys.modules['mindio_acp._c2python_api'] = MagicMock()

# disable auto patch megatron
os.environ.setdefault("MINDIO_AUTO_PATCH_MEGATRON", "false")


@pytest.fixture(scope='module', autouse=True)
def mock_external_modules(package_mocker):
    torch.npu = package_mocker.MagicMock()
    sys.modules['torch_npu'] = package_mocker.MagicMock()

    from mindio_acp.acc_io.mindio_help import torch_uninitialize_helper
    atexit.unregister(torch_uninitialize_helper)
