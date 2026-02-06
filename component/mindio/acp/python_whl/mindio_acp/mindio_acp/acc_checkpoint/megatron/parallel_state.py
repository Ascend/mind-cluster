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
from collections import namedtuple
from enum import Enum
from typing import List
from mindio_acp.common import mindio_logger

import torch
from megatron.training import global_vars
from torch.distributed import ProcessGroup
from mindio_acp.acc_checkpoint.utils.utils import SingletonMeta

logging = mindio_logger.LOGGER


class CKPTStage(Enum):
    PreLoad = 1
    LoadDPEP = 2
    LoadDPCP = 3


InitParallelResult = namedtuple("InitParallelResult",
                                "selected_model_rank, selected_optim_rank, process_group, group_ranks")


class InitParallelPolicy(metaclass=SingletonMeta):
    """ separate data parallel group and return source rank, group and global ranks list of load broadcast group
    :param replica_count: the replica count in the dp group
    :return: load_broadcast_src_rank -> the source rank load and broadcast model state dict
            load_broadcast_group -> the group for source rank to load and broadcast model state dict
            load_broadcast_global_ranks -> the ranks list of load_broadcast_group
    """
    _replica_count: int
    _ckpt_stage: CKPTStage
    _selected_model_rank: int
    _selected_optim_rank: int
    _process_group: ProcessGroup
    _group_ranks: List[int]

    def __init__(self, replica_count: int, dp_group, ckpt_stage: CKPTStage):
        self._dp_modulo_exp_ranks = None
        self._replica_count = replica_count
        self._ckpt_stage = ckpt_stage

        result = self.init_parallel_policy(dp_group)
        self._selected_model_rank = result.selected_model_rank
        self._selected_optim_rank = result.selected_optim_rank
        self._process_group = result.process_group
        self._group_ranks = result.group_ranks

    @property
    def selected_model_rank(self) -> int:
        return self._selected_model_rank

    @property
    def selected_optim_rank(self) -> int:
        return self._selected_optim_rank

    @property
    def process_group(self) -> ProcessGroup:
        """
        Returns:
           processGroup for broadcast or scatter.

        """
        return self._process_group

    @property
    def group_ranks(self) -> List[int]:
        """
        Returns: tuple.
            The first item is ProcessGroup for broadcast or scatter.
            The second item is list of global ranks for ProcessGroup.
        """
        return self._group_ranks

    @property
    def dp_modulo_exp(self):
        if self._dp_modulo_exp_ranks is None:
            args = global_vars.get_args()
            world_size = args.world_size

            tp_size = args.tensor_model_parallel_size
            pp_size = args.pipeline_model_parallel_size
            cp_size = args.context_parallel_size
            ep_size = args.expert_model_parallel_size
            dp_size: int = world_size // (tp_size * pp_size * cp_size)

            tp_and_dp_size: int = tp_size * dp_size # 32768
            num_tp_and_dp_groups: int = world_size // tp_and_dp_size # 4
            tp_and_ep_size: int = tp_size * ep_size # 512

            for i in range(num_tp_and_dp_groups):
                start_rank = i * tp_and_dp_size
                end_rank = (i + 1) * tp_and_dp_size
                for j in range(tp_and_ep_size):
                    ranks = range(start_rank + j, end_rank, tp_and_ep_size)
                    if args.rank in ranks:
                        self._dp_modulo_exp_ranks = list(ranks)
        return self._dp_modulo_exp_ranks

    def reset_optim_rank(self):
        self._selected_optim_rank = None

    def init_parallel_policy(self, dp_group):
        """ partition data parallel group and return source rank to save model and optimizer state dict,
            group and global ranks list of partition dp group
        :return: selected_rank_model -> the source rank to save/load model state dict
                selected_rank_optimizer -> the source rank to save/load optimizer state dict
                group -> the ProcessGroup of partition dp group
                global_ranks -> the ranks list of partition dp group
        """
        args = global_vars.get_args()
        world_size = args.world_size
        nproc_per_node = int(os.getenv("LOCAL_WORLD_SIZE", 0))  # the num of NPU each node
        if nproc_per_node <= 0:
            raise ValueError("The env `LOCAL_WORLD_SIZE` value should > 0")
        hosts_num = world_size // nproc_per_node

        dp_global_ranks = self.dp_modulo_exp
        if dp_group is not None:
            dp_global_ranks = torch.distributed.get_process_group_ranks(dp_group)
        if hosts_num < self._replica_count:
            logging.warning(f"hosts_num({hosts_num}) is less than replica_count({self._replica_count}), "
                            f"will set replica_count to 1.")
            self._replica_count = 1

        if self._replica_count <= 1 or (len(dp_global_ranks) // self._replica_count) <= 1:
            return InitParallelResult(selected_model_rank=dp_global_ranks[0], selected_optim_rank=dp_global_ranks[0],
                                      process_group=dp_group, group_ranks=dp_global_ranks)

        dp_size = len(dp_global_ranks)
        # variable for partition dp group
        dp_threshold = dp_size // self._replica_count
        extra_for_last_group = dp_size % self._replica_count
        # need to call new_group for every rank. see pytorch issue: 108732
        for k in range(0, self._replica_count):
            start_rank = k * dp_threshold
            end_rank = (k + 1) * dp_threshold + ((k + 1) // self._replica_count) * extra_for_last_group
            ranks_partition = dp_global_ranks[start_rank:end_rank]
            if args.rank in ranks_partition:
                group = None
                if self._ckpt_stage != CKPTStage.PreLoad:
                    group = torch.distributed.new_group(ranks_partition, use_local_synchronization=True)
                return InitParallelResult(selected_model_rank=ranks_partition[1],
                                          selected_optim_rank=ranks_partition[0],
                                          process_group=group, group_ranks=list(ranks_partition))
        return None
