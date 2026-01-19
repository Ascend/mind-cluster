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

import json

from megatron.training import get_args
from megatron.core import mpu
from megatron.training.training import build_train_valid_test_data_iterators
from mindio_ttp.adaptor import tft_optimizer_data_repair
from mindio_ttp.adaptor import utils
from mindio_ttp.controller_ttp import ttp_logger

# ORIGIN_DP_SIZE origin dp size
ORIGIN_DP_SIZE = None
# ORIGIN_NUM_MICRO_BATCHES origin num_micro_vatches
ORIGIN_NUM_MICRO_BATCHES = None
# SCALE_IN_WORLD_GROUP scale-in training world group
SCALE_IN_WORLD_GROUP = None
# SCALE_IN_DP_CP_REPLICA_GROUP scale-in training dp_cp_replica group
SCALE_IN_DP_CP_REPLICA_GROUP = None
# SCALE_IN_DP_CP_REPLICA_GROUP_GLOO scale-in training dp_cp_replica gloo group
SCALE_IN_DP_CP_REPLICA_GROUP_GLOO = None
# FAULT_RANK_IN_DP_CP_REPLICA_GROUP whether fault rank in this dp_cp_replica group
FAULT_RANK_IN_DP_CP_REPLICA_GROUP = False
# IS_FAULT_REPLICA_RANK whether replica rank of fault rank
IS_FAULT_REPLICA_RANK = False
# FAULT_REPLICA_RANK the replica rank of fault rank
FAULT_REPLICA_RANK = None
# SCALE_IN_RUNNING_STATE whether in scale-in training
SCALE_IN_RUNNING_STATE = False
# HAS_DATA whether global_batch_size % dp is 0
HAS_DATA = None


def update_scale_in_flag(new_state):
    global SCALE_IN_RUNNING_STATE
    SCALE_IN_RUNNING_STATE = new_state


def zit_get_has_data_index():
    return HAS_DATA


def zit_get_scale_in_world_group():
    return SCALE_IN_WORLD_GROUP


def zit_is_fault_replica_rank():
    return IS_FAULT_REPLICA_RANK


def zit_fault_rank_in_dp_cp_replica_group():
    return FAULT_RANK_IN_DP_CP_REPLICA_GROUP


def zit_scale_in_running_state():
    return SCALE_IN_RUNNING_STATE


def zit_get_scale_in_dp_cp_replica_group():
    return SCALE_IN_DP_CP_REPLICA_GROUP


def zit_get_scale_in_dp_cp_replica_group_gloo():
    return SCALE_IN_DP_CP_REPLICA_GROUP_GLOO


def zit_get_fault_replica_rank():
    return FAULT_REPLICA_RANK


def build_dataset(args):
    train_data_iterator, valid_data_iterator, test_data_iterator \
        = build_data_iterator(args[utils.TRAIN_PARAM][utils.MODEL_INDEX])
    args[utils.TRAIN_PARAM][utils.TRAIN_DATA_INDEX] = train_data_iterator
    args[utils.TRAIN_PARAM][utils.VALID_DATA_INDEX] = valid_data_iterator
    args[utils.TEST_DATA_ITER][0] = test_data_iterator


def build_data_iterator(model):
    args = get_args()
    _, _, tft_optimizer_data_repair.train_valid_test_datasets_provider_ = tft_optimizer_data_repair.get_build_data_args()
    if args.virtual_pipeline_model_parallel_size is not None:
        train_data_iterator, valid_data_iterator, test_data_iterator = [], [], []
        for i in range(len(model)):
            mpu.set_virtual_pipeline_model_parallel_rank(i)
            iterators = build_train_valid_test_data_iterators(
                tft_optimizer_data_repair.train_valid_test_datasets_provider_)
            train_data_iterator.append(iterators[0])
            valid_data_iterator.append(iterators[1])
            test_data_iterator.append(iterators[2])
    else:
        train_data_iterator, valid_data_iterator, test_data_iterator \
            = build_train_valid_test_data_iterators(tft_optimizer_data_repair.train_valid_test_datasets_provider_)
    return train_data_iterator, valid_data_iterator, test_data_iterator


def check_scale_out_params(params: str):
    parsed_dict = json.loads(params)
    ttp_logger.LOGGER.info(f"params: {parsed_dict}")
    strategy = parsed_dict.get("scale-out-strategy", "")
    if strategy != "DP":
        ttp_logger.LOGGER.error(f"Only support DP strategy, but got {strategy}")
        raise Exception(f"Only support DP strategy, but got {strategy}")


def check_scale_in_params(params: str):
    parsed_dict = json.loads(params)
    ttp_logger.LOGGER.info(f"params: {parsed_dict}")
    strategy = parsed_dict.get("scale-in-strategy", "")
    if strategy != "DP":
        ttp_logger.LOGGER.error(f"Only support DP strategy, but got {strategy}")
        raise Exception(f"Only support DP strategy, but got {strategy}")