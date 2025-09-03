#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# Copyright 2025. Huawei Technologies Co.,Ltd. All rights reserved.
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
# ==============================================================================

import torch

from megatron.core import mpu
from mindio_ttp.adaptor.tft_replica_optimizer import TTPReplicaOptimizer
from mindio_ttp.controller_ttp import ttp_logger
from mindio_ttp.adaptor.utils import FileUtils

from . import common


class TTPElasticTrainingReplicaOptimizer(TTPReplicaOptimizer):
    def save_parameter_state(self, filename: str):
        cur_rank = torch.distributed.get_rank()
        state_dict = self.save_parameter_state_impl()
        check_ret, err_msg, filename = FileUtils.regular_file_path(filename, '/', False)
        if not check_ret:
            ttp_logger.LOGGER.error(f"rank {cur_rank}: save parameter filename is not valid.")
            raise Exception(f"rank {cur_rank}: save parameter filename is not valid, error: {err_msg}")
        if self.error_dump:
            save_rank = self.save_args['rank']
            if cur_rank == save_rank:
                torch.save(state_dict, filename)
                ttp_logger.LOGGER.info(f"error dump rank {cur_rank} save parameters successfully")
        elif common.zit_scale_in_running_state():
            scale_in_dp_group = mpu.get_data_parallel_group()
            if torch.distributed.get_rank(scale_in_dp_group) == 0:
                torch.save(state_dict, filename)
                ttp_logger.LOGGER.info(f"rank {cur_rank} save parameters successfully in scale-in training mode")
        else:
            if torch.distributed.get_rank(self.ori_dp_group) == 0:
                torch.save(state_dict, filename)
                ttp_logger.LOGGER.info(f"normal rank {cur_rank} save parameters successfully")

