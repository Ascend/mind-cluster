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

import torch
import pytest
from collections import OrderedDict

from mindio_acp.acc_checkpoint.core.model_params_serialization import model_params_save, model_params_load


@pytest.mark.skip(reason="failed without npu")
def test_model_params_save_and_load():
    state_dict = {
        "args": {
            "DDP_impl": "local",
        },
        "model": {
            "language_model": {
                "embedding": {
                    "word_embeddings": OrderedDict(
                        {
                            "weight": torch.rand(1, 2),
                            "position_embeddings": torch.rand(2, 3),
                        }
                    )
                }
            }
        }
    }
    data_buff, tensors = model_params_save(state_dict)

    call_count = 0

    def receiver_tensor(tensor):
        nonlocal call_count
        result = tensors[str(call_count)]
        call_count += 1
        return result

    state_dict2 = model_params_load(data_buff, receiver_tensor)
    assert state_dict2 == state_dict
