#!/usr/bin/env python
# coding=utf-8
# Copyright (c) 2024, NVIDIA CORPORATION. All rights reserved.
# Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License
# You may obtain a copy of the License at
#    http://www.apache.org/licenses/LICENSE-2.0
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied
# See the License for the specific language governing permissions and
# limitations under the License
#
# Modification description: Patch the Megatron framework save and load functions using
# MindIO's asynchronous save and load methods for acceleration.

import os
import random
import sys
from enum import Enum, auto
from functools import wraps

import numpy as np
import torch
from megatron.core import mpu, tensor_parallel
from megatron.core.num_microbatches_calculator import update_num_microbatches
from megatron.core.rerun_state_machine import get_rerun_state_machine
from megatron.training import global_vars
from megatron.training.checkpointing import check_checkpoint_args
from megatron.training.checkpointing import checkpoint_exists
from megatron.training.checkpointing import find_checkpoint_rank_0
from megatron.training.checkpointing import fix_query_key_value_ordering
from megatron.training.checkpointing import get_checkpoint_name
from megatron.training.checkpointing import get_checkpoint_tracker_filename
from megatron.training.checkpointing import get_checkpoint_version
from megatron.training.checkpointing import get_distributed_optimizer_checkpoint_name
from megatron.training.checkpointing import set_checkpoint_version
from megatron.core.parallel_state import get_data_modulo_expert_parallel_group
from megatron.training.checkpointing import _get_checkpoint_format
from megatron.training.checkpointing import _to_dtensor
from megatron.training.utils import unwrap_model

from mindio_acp.common import mindio_logger
from mindio_acp.acc_checkpoint.megatron.preload_checkpoint_patch import get_replica_count
from mindio_acp.acc_checkpoint.megatron.parallel_state import InitParallelPolicy, CKPTStage
from mindio_acp.acc_checkpoint.megatron.load_optimizer_patch import load_parameter_state

logging = mindio_logger.LOGGER


def import_torch_mindio():
    global print_rank_0, CheckpointHelper
    from mindio_acp.acc_checkpoint.utils.utils import print_rank_0
    from mindio_acp.acc_checkpoint.framework_acp import CheckpointHelper


class CheckpointType(Enum):
    LEGACY = auto()
    LOCAL = auto()
    GLOBAL = auto()
    TORCH_DCP = auto()


def _load_base_checkpoint(
        load_dir,
        args,
        rank0=False,
        sharded_state_dict=None,
        checkpointing_context=None,
):
    """ Load the base state_dict from the given directory

    If rank0 is true, just loads rank 0 checkpoint, ignoring arguments.
    """
    iteration, release = -1, False
    tracker_filename = 'because load directory is not defined'
    if load_dir is not None:
        tracker_filename = get_checkpoint_tracker_filename(load_dir)
        if os.path.isfile(tracker_filename):
            iteration, release = read_metadata(tracker_filename)

    # Allow user to specify the loaded iteration.
    if getattr(args, "ckpt_step", None):
        iteration = args.ckpt_step

    # Otherwise we are dealing with global checkpoints
    # If no tracker file, return nothing
    if iteration == -1:
        if not rank0:
            print_rank_0('WARNING: could not find the metadata file {}'.format(tracker_filename))
            print_rank_0('    will not load any checkpoints and will start from random')
        # Conditionally exit if checkpoint not found.
        if args.exit_on_missing_checkpoint:
            print_rank_0(">> '--exit-on-missing-checkpoint' set ... exiting. <<")
            if torch.distributed.is_initialized():
                torch.distributed.barrier()
            raise ValueError

        return None, "", False, None

    # Determine the type of the checkpoint on disk.
    checkpoint_name = get_checkpoint_name(load_dir, iteration, release, return_base_dir=True)
    ckpt_format = _get_checkpoint_format(checkpoint_name)

    if not rank0:
        dist_infix = "distributed " if ckpt_format == "torch_dist" else ""
        if release:
            print_rank_0(f' loading release {dist_infix}checkpoint from {load_dir}')
        else:
            print_rank_0(
                f' loading {dist_infix}checkpoint from {load_dir} at iteration {iteration}'
            )

    ckpt_type = CheckpointType.LEGACY
    # Handle global legacy checkpoint
    if rank0:
        checkpoint_name = find_checkpoint_rank_0(load_dir, iteration, release)
    else:
        checkpoint_name = get_checkpoint_name(load_dir, iteration, release, return_base_dir=False)

    # adaptor for CheckpointHelper
    dp_ep_group = get_data_modulo_expert_parallel_group()
    dp_global_ranks = torch.distributed.get_process_group_ranks(dp_ep_group)
    if len(dp_global_ranks) > 1:
        replica_count = get_replica_count()
        policy = InitParallelPolicy(replica_count, dp_ep_group, CKPTStage.LoadDPEP)
        load_rank = policy.selected_model_rank
        process_group = policy.process_group
    else:
        load_rank = args.rank
        process_group = None
    try:
        state_dict = CheckpointHelper(args.rank).load_model_checkpoint(checkpoint_name, load_rank, process_group)
    except ModuleNotFoundError:
        from megatron.legacy.fp16_deprecated import loss_scaler

        # For backward compatibility.
        if not rank0:
            print_rank_0(' > deserializing using the old code structure ...')
        sys.modules['fp16.loss_scaler'] = sys.modules['megatron.legacy.fp16_deprecated.loss_scaler']
        sys.modules['megatron.fp16.loss_scaler'] = sys.modules[
            'megatron.legacy.fp16_deprecated.loss_scaler'
        ]
        sys.modules['megatron.model'] = sys.modules['megatron.legacy.model']
        state_dict = torch.load(checkpoint_name, map_location='cpu', weights_only=False)
        sys.modules.pop('fp16.loss_scaler', None)
        sys.modules.pop('megatron.fp16.loss_scaler', None)
        sys.modules.pop('megatron.model', None)
    except Exception as e:
        print_rank_0('could not load the checkpoint')
        print_rank_0(e)
        raise e

    return state_dict, checkpoint_name, release, ckpt_type


def acp_load_checkpoint(ddp_model, optimizer, opt_param_scheduler, load_arg='load', strict=True,
                        checkpointing_context=None, skip_load_to_model_and_opt=False):
    args = global_vars.get_args()
    import_torch_mindio()
    load_dir = getattr(args, load_arg)
    model = unwrap_model(ddp_model)
    torch_dist_str = "torch_dist"
    model_str = "model"

    ckpt_format = args.ckpt_format
    not_support_format = args.auto_detect_ckpt_format or ckpt_format == torch_dist_str
    should_skip_loading = skip_load_to_model_and_opt or optimizer.is_stub_optimizer
    unsupported_configuration = not_support_format or should_skip_loading
    if unsupported_configuration:
        raise NotImplementedError("Unsupported Configuration")

    load_kwargs = {}
    state_dict, checkpoint_name, release, ckpt_type = _load_base_checkpoint(
        load_dir, args, rank0=False, checkpointing_context=checkpointing_context,
        **load_kwargs
    )

    # Checkpoint not loaded.
    if state_dict is None:
        # Iteration and num_floating_point_operations_so_far default to 0.
        return 0, 0

    # Set checkpoint version.
    set_checkpoint_version(state_dict.get('checkpoint_version', 0))

    if release or args.finetune:
        iteration = 0
    else:
        iteration = state_dict.get('iteration', state_dict.get('total_iters'))
        if iteration is None:
            raise KeyError(f'Unable to load iteration from checkpoint {checkpoint_name}, exiting')
    num_floating_point_operations_so_far = state_dict.get('num_floating_point_operations_so_far', 0)

    have_valid_args = 'args' in state_dict and not args.finetune
    if have_valid_args:
        checkpoint_args = state_dict.get('args')
        check_checkpoint_args(checkpoint_args)
        args.consumed_train_samples = getattr(checkpoint_args,
                                              'consumed_train_samples', 0)
        args.skipped_train_samples = getattr(checkpoint_args,
                                             'skipped_train_samples', 0)
        update_num_microbatches(consumed_samples=args.consumed_train_samples, verbose=True)
        args.consumed_valid_samples = getattr(checkpoint_args,
                                              'consumed_valid_samples', 0)
    else:
        print_rank_0('could not find arguments in the checkpoint ...')

    # Model.
    strict = False if args.retro_add_retriever else strict
    if len(ddp_model) == 1:
        ddp_model[0].load_state_dict(state_dict.get(model_str), strict=strict)
    else:
        for i, model in enumerate(ddp_model):
            mpu.set_virtual_pipeline_model_parallel_rank(i)
            model.load_state_dict(state_dict.get('model%d' % i), strict=strict)

    # Fix up query/key/value matrix ordering if needed.
    ckpt_version = get_checkpoint_version()
    print_rank_0(f' checkpoint version : {ckpt_version}')
    fix_query_key_value_ordering(model, ckpt_version)

    # load Optimizer.
    is_optim_load_disabled = args.no_load_optim or release or args.finetune
    if not is_optim_load_disabled:
        # Load state dict.
        optimizer.load_state_dict(state_dict.get('optimizer'))

        # Load distributed optimizer's custom parameter state.
        # For distributed checkpoint it's already loaded in load_state_dict above
        if args.use_distributed_optimizer:
            # NOTE: this is a manual read of the tracker file.
            # This code should not be reached when reading from a non_persistent checkpoint
            tracker_filename = get_checkpoint_tracker_filename(load_dir)
            iteration, release = read_metadata(tracker_filename)
            model_checkpoint_name = \
                get_checkpoint_name(load_dir, iteration, release)
            optim_checkpoint_name = \
                get_distributed_optimizer_checkpoint_name(
                    model_checkpoint_name)
            load_parameter_state(optimizer, optim_checkpoint_name)

        # Load scheduler.
        if opt_param_scheduler is not None:
            if 'lr_scheduler' in state_dict:  # backward compatbility
                opt_param_scheduler.load_state_dict(state_dict.get('lr_scheduler'))
            else:
                opt_param_scheduler.load_state_dict(state_dict.get('opt_param_scheduler'))
    else:
        should_reload_params = (args.fp16 or args.bf16) and optimizer is not None
        if should_reload_params:
            optimizer.reload_model_params()

    # rerun state
    if 'rerun_state_machine' in state_dict:
        get_rerun_state_machine().load_state_dict(state_dict.get('rerun_state_machine'))

    rng_state_str = "rng_state"
    rng_tracker_states_str = "rng_tracker_states"
    # rng states.
    if not release and not args.finetune and not args.no_load_rng:
        random_key = 'random_rng_state'
        numpy_key = 'np_rng_state'
        torch_key = 'torch_rng_state'
        cuda_key = 'cuda_rng_state'
        if rng_state_str in state_dict:
            # access rng_state for data parallel rank
            rng_state_list = state_dict.get(rng_state_str)
            rng_state = rng_state_list[0]
            random.setstate(rng_state.get(random_key))
            np.random.set_state(rng_state.get(numpy_key))
            torch.set_rng_state(rng_state.get(torch_key))
            torch.cuda.set_rng_state(rng_state.get(cuda_key))
            if not rng_state.get(rng_tracker_states_str):
                raise KeyError
            tensor_parallel.get_cuda_rng_tracker().set_states(
                rng_state.get(rng_tracker_states_str))
        else:  # backward compatability
            random.setstate(state_dict.get(random_key))
            np.random.set_state(state_dict.get(numpy_key))
            torch.set_rng_state(state_dict.get(torch_key))
            torch.cuda.set_rng_state(state_dict.get(cuda_key))
            if not state_dict.get(rng_tracker_states_str):
                raise KeyError
            tensor_parallel.get_cuda_rng_tracker().set_states(
                state_dict.get(rng_tracker_states_str))

    torch.distributed.barrier()

    print_rank_0(f'  successfully loaded checkpoint from {load_dir} '
                 f'[ t {mpu.get_tensor_model_parallel_rank() + 1}/{mpu.get_tensor_model_parallel_world_size()}, '
                 f'p {mpu.get_pipeline_model_parallel_rank() + 1}/{mpu.get_pipeline_model_parallel_world_size()} ] '
                 f'at iteration {iteration}')

    torch.cuda.empty_cache()

    return iteration, num_floating_point_operations_so_far


def load_checkpoint_wrapper(fn):
    @wraps(fn)
    def wrapper(ddp_model, optimizer, opt_param_scheduler, load_arg='load', strict=True,
                checkpointing_context=None, skip_load_to_model_and_opt=False):
        """Load a model checkpoint and return the iteration.
        strict (bool): whether to strictly enforce that the keys in
            :attr:`state_dict` of the checkpoint match the names of
            parameters and buffers in model.
        skip_load_to_model_and_opt (bool): whether to call `load_state_dict`
            for :attr:`model` and :attr:`optimizer`. In case of running FSDP2 with mcore distributed
            checkpointing, the tensors are already loaded in-place by `_load_base_checkpoint`.
        """
        args = global_vars.get_args()

        if args.use_dist_ckpt or args.async_save:
            return fn(ddp_model=ddp_model,
                      optimizer=optimizer,
                      opt_param_scheduler=opt_param_scheduler,
                      load_arg=load_arg,
                      strict=strict,
                      checkpointing_context=checkpointing_context,
                      skip_load_to_model_and_opt=skip_load_to_model_and_opt)

        return acp_load_checkpoint(ddp_model=ddp_model,
                                   optimizer=optimizer,
                                   opt_param_scheduler=opt_param_scheduler,
                                   load_arg=load_arg,
                                   strict=strict,
                                   checkpointing_context=checkpointing_context,
                                   skip_load_to_model_and_opt=skip_load_to_model_and_opt)

    return wrapper


def read_metadata(tracker_file):
    last_iteration = 0
    release = False
    with open(tracker_file, 'r') as f:
        metastring = f.read().strip()
        try:
            last_iteration = int(metastring)
        except ValueError as e:
            release = metastring == 'release'
            if not release:
                print_rank_0(f'ERROR: Metadata file {tracker_file} is invalid . Exiting')
                raise e

    return last_iteration, release
