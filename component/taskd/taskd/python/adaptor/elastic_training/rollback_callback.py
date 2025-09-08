#!/usr/bin/env python
# -*- coding: utf-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2024. All rights reserved.
import torch
from megatron.core import mpu
from megatron.training import get_args
from megatron.core import num_microbatches_calculator
from megatron.core.num_microbatches_calculator import get_num_microbatches
from megatron.training.training import build_train_valid_test_data_iterators, training_log
from megatron.training.utils import calc_params_l2_norm
import megatron.training.global_vars
from mindio_ttp.adaptor.tft_optimizer_data_repair import (LogArgs, get_build_data_args, unset_memory_ckpt, set_load_ckpt,
                                        average_losses_across_microbatches, get_load_ckpt)
from mindio_ttp.adaptor.tft_replica_group import destroy_repair_group
from mindio_ttp.adaptor.utils import (TRAIN_PARAM, MODEL_INDEX, OPTIM_INDEX, SCHEDULER_INDEX)
from mindio_ttp.framework_ttp.ttp_decorator import get_device
from mindio_ttp.controller_ttp import ttp_logger

from . import common


def rollback_callback(step: int, train_args, params: str):
    ttp_logger.LOGGER.info(f"rollback strategy params: {params}, step: {step}")
    common.check_scale_out_params(params)
    args = get_args()
    torch.npu.set_device(get_device())
    # update num_microbatches
    if common.ORIGIN_DP_SIZE is not None and common.ORIGIN_NUM_MICRO_BATCHES is not None:
        args = get_args()
        if args.rampup_batch_size is not None and len(args.rampup_batch_size) == 3:
            new_micro_bsz_times_dp_size = args.micro_batch_size * common.ORIGIN_DP_SIZE
            num_microbatches_calculator._GLOBAL_NUM_MICROBATCHES_CALCULATOR.data_parallel_size = common.ORIGIN_DP_SIZE
            num_microbatches_calculator._GLOBAL_NUM_MICROBATCHES_CALCULATOR.micro_batch_times_data_parallel_size = (
                new_micro_bsz_times_dp_size)
        num_microbatches_calculator._GLOBAL_NUM_MICROBATCHES_CALCULATOR.num_micro_batches = (
            common.ORIGIN_NUM_MICRO_BATCHES)
        ttp_logger.LOGGER.info(f"new num_micro_batches: {get_num_microbatches()}")
    load_ckpt = get_load_ckpt()
    if load_ckpt:
        step = args.iteration
        if args.train_samples is None:
            train_args[TRAIN_PARAM][SCHEDULER_INDEX].num_steps = step * args.global_batch_size
        set_load_ckpt(False)
    #update learning rate
    if args.train_samples is None:
        args.consumed_train_samples = step * args.global_batch_size
    if train_args[TRAIN_PARAM][SCHEDULER_INDEX].num_steps != args.consumed_train_samples:
        train_args[TRAIN_PARAM][SCHEDULER_INDEX].step(args.global_batch_size)
    feature_rollback()
    gather_model_params_from_optimizer(train_args[TRAIN_PARAM][OPTIM_INDEX], step)
    common.build_dataset(train_args)
    unset_memory_ckpt()
    destroy_repair_group()
    training_log_repair(step, train_args)
    rebuild_global_vars(step, args)
    rank = torch.distributed.get_rank()
    ttp_logger.LOGGER.info(f"[rollback] rank {rank} rollback success")


def feature_rollback():
    args = get_args()
    # fix megatron global buffer unsafe datas
    if hasattr(mpu, 'destroy_global_memory_buffer') and hasattr(mpu, '_set_global_memory_buffer'):
        mpu.destroy_global_memory_buffer()
        mpu._set_global_memory_buffer()
    if hasattr(args, "swap_attention") and args.swap_attention:
        # reinit swap prefetch
        from mindspeed.core.memory.adaptive_recomputing.prefetch import SwapPrefetch
        SwapPrefetch.swap_prefetch.prefetch_data_ptr_list = []
        SwapPrefetch.swap_prefetch.prefetch_list = []
        SwapPrefetch.swap_prefetch.slice_tensor_storage_ptr_list = []
        SwapPrefetch.swap_prefetch.swap_tensors = []
        SwapPrefetch.swap_prefetch.data_ptr = {}
        SwapPrefetch.swap_prefetch.cur_micro_num = 0
        SwapPrefetch.swap_prefetch.remove_num = 0
        SwapPrefetch.swap_prefetch.forward_flag = False
        SwapPrefetch.swap_prefetch.slice_tensor_storage_ptr = {}
    if hasattr(args, "num_experts") and args.num_experts:
        mpu._MOE_AUX_LOSSES_LOGGING_TRACKER = {}
    if hasattr(args, "moe_permutation_async_comm") and args.moe_permutation_async_comm:
        from mindspeed.core.transformer.moe import moe_utils
        moe_utils.AG_SHARED_EXPERTS_INPUTS = []


def build_data_iterator(model):
    args = get_args()
    _, _, train_valid_test_datasets_provider_ = get_build_data_args()
    if args.virtual_pipeline_model_parallel_size is not None:
        train_data_iterator = []
        valid_data_iterator = []
        test_data_iterator = []
        for i in range(len(model)):
            mpu.set_virtual_pipeline_model_parallel_rank(i)
            iterators = build_train_valid_test_data_iterators(
            train_valid_test_datasets_provider_)
            train_data_iterator.append(iterators[0])
            valid_data_iterator.append(iterators[1])
            test_data_iterator.append(iterators[2])
    else:
        train_data_iterator, valid_data_iterator, test_data_iterator \
        = build_train_valid_test_data_iterators(
        train_valid_test_datasets_provider_)
    return train_data_iterator, valid_data_iterator, test_data_iterator


def rebuild_global_vars(step, args):
    args.iteration = step
    from megatron.training.global_vars import _set_timers
    megatron.training.global_vars._GLOBAL_TIMERS = None
    _set_timers(args)
    from megatron.core.rerun_state_machine import destroy_rerun_state_machine
    destroy_rerun_state_machine()


def training_log_repair(iteration: int, train_args: list):
    # Average losses across micro_batches
    if LogArgs.losses_reduced_ and len(LogArgs.losses_reduced_) > 1:
        LogArgs.losses_reduced_ = average_losses_across_microbatches(LogArgs.losses_reduced_)
    args = get_args()
    losses_reduced = LogArgs.losses_reduced_
    if iteration == args.iteration or losses_reduced is None:
        ttp_logger.LOGGER.info(f"rank:{args.rank} Skip the train log repair. repair_step:{iteration} "
                               f"args.iteration:{args.iteration}.")
        return
    # get necessary parameters
    loss_scale = train_args[TRAIN_PARAM][OPTIM_INDEX].get_loss_scale().item()
    params_norm = None
    if args.log_params_norm:
        params_norm = calc_params_l2_norm(train_args[TRAIN_PARAM][MODEL_INDEX])
    learning_rate = None
    decoupled_learning_rate = None
    for param_group in train_args[TRAIN_PARAM][OPTIM_INDEX].param_groups:
        if param_group['is_decoupled_lr']:
            decoupled_learning_rate = param_group['lr']
        else:
            learning_rate = param_group['lr']
    report_memory_flag = False
    skipped_iter = 0
    total_loss_dict = {}
    # get loss from losses_reduced.
    loss_dict = {}
    if LogArgs.losses_reduced_:
        if len(LogArgs.losses_reduced_) == 1:
            loss_dict = LogArgs.losses_reduced_[0]
        else:
            ttp_logger.LOGGER.warning(f"lm loss might be not correct, please check the usage of tft_set_losses_reduced."
                                      f"loss_dict:{LogArgs.losses_reduced_}")
    # do repair log
    ttp_logger.LOGGER.info(f"rank:{args.rank} repair training log at iteration: {iteration}")
    training_log(loss_dict, total_loss_dict, learning_rate, decoupled_learning_rate, iteration,
    loss_scale, report_memory_flag, skipped_iter, LogArgs.grad_norm_, params_norm, LogArgs.num_zeros_in_grad_)
    return


def gather_model_params_from_optimizer(optimizer, step):
    args = get_args()

    if hasattr(optimizer, 'set_update_successful'):
        optimizer.set_update_successful(True)
    if getattr(args, "reuse_fp32_param", False):
        optimizer.fp32_tensor_to_fp16_tensor()
    else:
        optimizer._copy_main_params_to_model_params()

    optimizer.sync_gather_all_model_params(force_sync=True)
    ttp_logger.LOGGER.info(f'rank:{args.rank} successfully gather and rollback at iteration {step}')