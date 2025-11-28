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

from functools import wraps

from .patch_utils import PatchManager

try:
    import mindio_ttp
    has_mindio_ttp = True
except Exception:
    has_mindio_ttp = False


def exec_adaptation():
    patch_initialize()
    patch_training()
    patch_core_optimizer()
    patch_miscellaneous()
    PatchManager.apply_patches()


def patch_initialize():
    PatchManager.register_patch('megatron.training.initialize.parse_args', parse_args_decorator)


def patch_training():
    from .save_checkpoint_patch import save_checkpoint_wrapper
    from .load_checkpoint_patch import load_checkpoint_wrapper
    from .preload_checkpoint_patch import initialize_model_parallel_wrapper
    PatchManager.register_patch('megatron.training.checkpointing.save_checkpoint', save_checkpoint_wrapper)
    PatchManager.register_patch('megatron.training.checkpointing.load_checkpoint', load_checkpoint_wrapper)
    PatchManager.register_patch('megatron.core.parallel_state.initialize_model_parallel',
                                initialize_model_parallel_wrapper)


def patch_miscellaneous():
    PatchManager.register_patch('megatron.training.arguments.parse_args', parse_args_decorator)


def patch_core_optimizer():
    from .save_checkpoint_patch import distrib_optimizer_step_wrapper
    PatchManager.register_patch('megatron.core.optimizer.optimizer.ChainedOptimizer.step',
                                distrib_optimizer_step_wrapper)
    PatchManager.register_patch('megatron.core.optimizer.optimizer.MixedPrecisionOptimizer.step',
                                distrib_optimizer_step_wrapper)
    PatchManager.register_patch('megatron.core.optimizer.optimizer.FP32Optimizer.step',
                                distrib_optimizer_step_wrapper)
    if has_mindio_ttp:
        PatchManager.register_patch('mindio_ttp.adaptor.TTPReplicaOptimizer.step',
                                    distrib_optimizer_step_wrapper)
        PatchManager.register_patch('mindio_ttp.adaptor.TTPFP16ReplicaOptimizer.step',
                                    distrib_optimizer_step_wrapper)
        PatchManager.register_patch('mindio_ttp.adaptor.TTPFP32ReplicaOptimizer.step',
                                    distrib_optimizer_step_wrapper)
        PatchManager.register_patch('mindio_ttp.adaptor.TTPReplicaChainedOptimizer.step',
                                    distrib_optimizer_step_wrapper)


def extra_args_provider_decorator(extra_args_provider):
    @wraps(extra_args_provider)
    def wrapper(parser):
        if extra_args_provider is not None:
            parser = extra_args_provider(parser)
        parser = process_args(parser)
        return parser

    return wrapper


def parse_args_decorator(parse_args):
    @wraps(parse_args)
    def wrapper(extra_args_provider=None, ignore_unknown_args=False):
        decorated_provider = extra_args_provider_decorator(extra_args_provider)
        return parse_args(decorated_provider, ignore_unknown_args)

    return wrapper


def process_args(parser):
    parser.conflict_handler = 'resolve'
    parser = _add_high_availability_args(parser)
    return parser


def _add_high_availability_args(parser):
    group = parser.add_argument_group(title='high_availability')

    group.add_argument('--replica-count',
                       type=int,
                       default=1,
                       help='high availability for replica count')
    return parser
