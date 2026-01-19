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
    def set_dump_args(self, rank, step, ranks_list):
        """
        In the scale-in running state, if a fault occurs again, it is no longer in the
        scale-in running state, and the flag 'SCALE_IN_RUNNING_STATE' should be updated to False.
        """
        super().set_dump_args(rank, step, ranks_list)
        ttp_logger.LOGGER.info(f"rank={rank}, step={step}, ranks_list={ranks_list},"
                               f" update scale in running state to False")
        common.update_scale_in_flag(False)

    def save_parameter_state(self, filename: str):
        if not self.error_dump and common.zit_scale_in_running_state():
            self.save_parameter_state_scale_in_running(filename)
        else:
            super().save_parameter_state(filename)

    def save_parameter_state_scale_in_running(self, filename: str):
        cur_rank = torch.distributed.get_rank()
        state_dict = self.save_parameter_state_impl()
        check_ret, err_msg, filename = FileUtils.regular_file_path(filename, '/', False)
        if not check_ret:
            ttp_logger.LOGGER.error(f"rank {cur_rank}: save parameter filename is not valid.")
            raise Exception(f"rank {cur_rank}: save parameter filename is not valid, error: {err_msg}")
        scale_in_dp_group = mpu.get_data_parallel_group()
        if torch.distributed.get_rank(scale_in_dp_group) == 0:
            torch.save(state_dict, filename)
            ttp_logger.LOGGER.info(f"rank {cur_rank} save parameters successfully in scale-in training mode")
