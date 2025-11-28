#!/usr/bin/env python
# coding=utf-8
# Copyright (c) Huawei Technologies Co., Ltd. 2023. All rights reserved.
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
import stat
import inspect
from typing import Callable, List, Union

import mindio_acp
from mindio_acp.acc_checkpoint.core.checkpoint_saver import CheckpointSaverMixin
from mindio_acp.acc_checkpoint.core.checkpoint_async_saver import CheckpointAsyncSaverMixin
from mindio_acp.acc_checkpoint.core.checkpoint_rapid_loader import CheckpointRapidLoaderMixin
from mindio_acp.acc_checkpoint.utils.utils import time_used, SingletonMeta
from mindio_acp.common.utils import get_relative_path
from mindio_acp.common import mindio_logger

logging = mindio_logger.LOGGER

MAX_FILE_PATH_LENGTH = 1024


class CheckpointHelper(CheckpointAsyncSaverMixin, CheckpointRapidLoaderMixin, CheckpointSaverMixin,
                       metaclass=SingletonMeta):
    """This class is used to help save checkpoint files.

    Example::
        >>> # (1) Preparing to Save Model Parameters:
        >>> ...
        >>> # Arguments, iteration, and model.
        >>> state_dict = dict()
        >>> state_dict['args'] = args
        >>> state_dict['checkpoint_version'] = 3.0
        >>> state_dict['iteration'] = iteration
        >>> if len(model) == 1:
        >>>     state_dict['model'] = model[0].state_dict_for_save_checkpoint()
        >>> else:
        >>>     for i in range(len(model)):
        >>>         mpu.set_virtual_pipeline_model_parallel_rank(i)
        >>>         state_dict['model%d' % i] = model[i].state_dict_for_save_checkpoint()
        >>> ...
        >>> # (2) save model parameters:
        >>> CheckpointHelper(rank).save_model_checkpoint(
        >>>     checkpoint_name=checkpoint_name,
        >>>     model=state_dict,
        >>> )
        >>>
        >>> # (3) save optimizer parameters:
        >>> CheckpointHelper(rank).save_optimizer_checkpoint(
        >>>     checkpoint_name=checkpoint_name,
        >>>     get_parameter_state_func=get_parameter_state_func,
        >>> )
        >>>
        >>> # (4) submit the task for updating the tracker file.
        >>> CheckpointHelper(rank).async_write_tracker_file(
        >>>     iteration=iteration,
        >>>     iteration_dir=checkpoint_path,
        >>>     total_file_count=file_count
        >>>     tracker_filename=filename,
        >>> )
        >>>
        >>> # (5) Wait for the parameters to be copied before updating the parameters in the next iteration.
        >>> # modify the method `def step(self) in class DistributedOptimizer`
        >>> class DistributedOptimizer(MixedPrecisionOptimizer):
        >>>     ...
        >>>     @torch.no_grad()
        >>>     def step(self):
        >>>         # Step the optimizer.
        >>>         from mindio_acp import CheckpointHelper
        >>>         CheckpointHelper(rank).wait_d2h_finished()
        >>>
        >>>         self.update_successful, grad_norm, num_zeros_in_grad = super().step()
        >>>         ...
        >>>
    """

    def __init__(self, rank):
        self.rank = rank
        CheckpointAsyncSaverMixin.__init__(self, rank)
        CheckpointRapidLoaderMixin.__init__(self, rank)
        CheckpointSaverMixin.__init__(self)

    @time_used
    def async_write_tracker_file(self, iteration: int, iteration_dir: str, total_file_count: int, tracker_filename: str,
                                 callback=None):
        """Submit the asynchronous task, that is updating the tracker file after finishing saving all checkpoints in
        this iteration.

        Arguments:
            iteration(int): iteration number
            iteration_dir(str): full directory path for this iteration
            total_file_count(int): total count of checkpoint files saved for this iteration
            tracker_filename(int): full path for tracker file
            callback: Callback function that will be invoked when the asynchronous task finished. The prototype is:
                      callback(iteration:int, result: int), result = 0 means success, otherwise means failed.
        """

        def _save_post_process(result, step):
            logging.info('tracker task for iteration=%d finished with result=%d', step, result)
            if result != 0:
                logging.error('watching checkpoint failed with result=%d for iteration %d', result, step)
                if callback is not None:
                    callback(step, result)
                return

            flags = os.O_WRONLY | os.O_CREAT
            mode = stat.S_IWUSR | stat.S_IRUSR
            real_path = os.path.realpath(tracker_filename)
            with os.fdopen(os.open(real_path, flags, mode), 'w') as f:
                f.write(str(step))

            if callback is not None:
                callback(step, 0)

        if len(iteration_dir) > 1024 or len(tracker_filename) > 1024:
            raise ValueError('the path length cannot exceed 1024 characters.')

        if not os.path.isdir(iteration_dir):
            raise ValueError('the iteration_dir is not a directory.')

        real_iteration_dir = os.path.realpath(iteration_dir)
        logging.info('Rank: %d register tracker task for iteration=%d, path=%s, file_count=%d',
                     self.rank, iteration, get_relative_path(iteration_dir), total_file_count)
        mindio_acp.register_checker(_save_post_process, {real_iteration_dir: total_file_count}, iteration, 300)

    @time_used
    def async_save_checkpoint(self, thread_func: Callable, kwargs):
        """ Use CheckpointHelper instance asynchronous thread run thread_func.

        :param thread_func: asynchronous function object
        :param kwargs: asynchronous function parameter
        :return:
        """
        self._async_save_checkpoint(thread_func, kwargs)

    def async_checkpoint_count(self):
        """ Decrement reference count. This function should be called when D2H(data HBM->HOST) complete.

        :return:
        """
        self._async_checkpoint_count()

    def get_checkpoint_copy_stream(self):
        """ Create torch_npu.npu.Stream to do D2H.

        :return:
        """
        self._get_checkpoint_copy_stream()

    @time_used
    def save_model_checkpoint(self, checkpoint_name: Union[str, List[str]], model: dict):
        r"""
        Save a model checkpoint file.

        A simple checkpoint, where the entire contents of the checkpoint specified by :attr:`model`.

        Arguments:
            checkpoint_name(str|list): Full path of the checkpoint file or files
            model(dict): Model parameters checkpoint to be saved
        """
        if isinstance(checkpoint_name, str):
            checkpoint_names = [checkpoint_name]
        else:
            checkpoint_names = list(checkpoint_name)
        for name in checkpoint_names:
            if len(name) > MAX_FILE_PATH_LENGTH:
                raise ValueError(f"The file path length cannot exceed {MAX_FILE_PATH_LENGTH} characters")

        self.async_save_model_checkpoint(checkpoint_name, model)

    @time_used
    def save_optimizer_checkpoint(self, checkpoint_name: Union[str, List[str]], get_parameter_state_func: Callable):
        r"""
        Save an optimizer checkpoint file.

        A complex checkpoint, We need to perform gather in the specified group to obtain the entire contents of
        the checkpoint.

        Arguments:
            checkpoint_name(str|list): Full path of the checkpoint file or files
            get_parameter_state_func(Callback): Within the optimizer group, only the ranks that need to be saved
                will return states. Other ranks should return None.
        """
        if isinstance(checkpoint_name, str):
            checkpoint_names = [checkpoint_name]
        else:
            checkpoint_names = list(checkpoint_name)
        for name in checkpoint_names:
            if len(name) > MAX_FILE_PATH_LENGTH:
                raise ValueError(f"The file path length cannot exceed {MAX_FILE_PATH_LENGTH} characters")

        sig = inspect.signature(get_parameter_state_func)
        if len(sig.parameters) > 1:
            raise ValueError("The number of get_parameter_state_func parameters should be 0 or 1")

        has_sig_para = True if len(sig.parameters) == 1 else False
        if has_sig_para and "notify_callback" not in sig.parameters:
            raise ValueError("The name of get_parameter_state_func parameters should be 'notify_callback'")

        self.async_save_optimizer_checkpoint(checkpoint_name, get_parameter_state_func, has_sig_para)

    @time_used
    def wait_d2h_finished(self):
        r"""
        Wait for asynchronous replication to complete. This function should be called before updating parameters in
        the next iteration.
        """
        self.wait_d2h_checkpoint_finished()

    @time_used
    def async_preload(self, load_path: str):
        """ Preload checkpoint files from load_dir.

        Arguments:
            load_path(str): Checkpoint load path. It should include `latest_checkpointed_iteration.txt` file.

        step1: (rank0) Obtain the iteration to be restored by reading the tracker file.
        step2: Exchange iteration information through the store.
        step3: Calculate the checkpoint file to be loaded and
               invoke mindio_acp.preload() to load checkpoint files.
        """
        if len(load_path) > MAX_FILE_PATH_LENGTH:
            raise ValueError(f"The load path length cannot exceed {MAX_FILE_PATH_LENGTH} characters")

        self.async_preload_checkpoint(load_path)

    @time_used
    def load_model_checkpoint(self, checkpoint_name: str, load_rank: int, process_group=None):
        """Load from a checkpoint file.

        Loading model params checkpoint. Only the first three parameters are required(Don't specify optimizer).

        Arguments:
            checkpoint_name(str): Full path of the checkpoint file
            load_rank(int): Specified rank to load checkpoint.
            process_group: Set process_group if you need broadcast the checkpoint_name via process_group.
                Otherwise, the checkpoint will be directly loaded using the load function.
        """
        if len(checkpoint_name) > 1024:
            raise ValueError("The load path length cannot exceed 1024 characters")
        if load_rank < 0:
            raise ValueError("The load_rank should >= 0")

        return self.rapid_load_model_checkpoint(checkpoint_name, load_rank, process_group)
