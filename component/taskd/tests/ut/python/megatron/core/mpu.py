#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# Copyright (c) 2025, Huawei Technologies Co., Ltd. All rights reserved.
_EMBEDDING_GROUP = None
_DATA_PARALLEL_GROUP = None
_DATA_PARALLEL_GROUP_GLOO = None
_DATA_PARALLEL_GLOBAL_RANKS = [0, 4, 8, 12]
_DATA_PARALLEL_GROUP_WITH_CP = None
_DATA_PARALLEL_GROUP_WITH_CP_GLOO = None
_DATA_PARALLEL_GLOBAL_RANKS_WITH_CP = [0, 4, 8, 12]
_CONTEXT_PARALLEL_GROUP = None
_CONTEXT_PARALLEL_GROUP_GLOO = None
_MODEL_PARALLEL_GROUP = None
_MODEL_PARALLEL_GROUP_GLOO = None
_TENSOR_MODEL_PARALLEL_GROUP = None
_TENSOR_MODEL_PARALLEL_GROUP_GLOO = None
_PIPELINE_MODEL_PARALLEL_GROUP = None
_PIPELINE_MODEL_PARALLEL_GROUP_GLOO = None
_POSITION_EMBEDDING_GROUP = None
_POSITION_EMBEDDING_GROUP_GLOO = None
_DATA_MODULO_EXPERT_PARALLEL_GROUP = None
_DATA_MODULO_EXPERT_PARALLEL_GROUP_GLOO = None


def set_virtual_pipeline_model_parallel_rank(rank):
    pass


def get_data_parallel_group(with_context_parallel=False):
    return None


def get_pipeline_model_parallel_world_size():
    return 1


def get_tensor_model_parallel_world_size():
    return 4


def get_context_parallel_world_size():
    return 1


