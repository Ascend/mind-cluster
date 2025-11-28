#!/usr/bin/env python
# coding=utf-8
# Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.
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
import io
import os
from typing import Callable

import torch
import torch.distributed

import mindio_acp
from mindio_acp.common import mindio_logger
from mindio_acp.acc_checkpoint.core.model_params_serialization import model_params_save, model_params_load
from mindio_acp.acc_checkpoint.utils.utils import time_used, retry
from mindio_acp.common.utils import get_relative_path

logging = mindio_logger.LOGGER


class CheckpointRapidLoaderMixin(object):

    def __init__(self, rank):
        self.__rank = rank

    @staticmethod
    def __broadcast_checkpoint_from_src_rank(checkpoint_name: str, broadcast: Callable[[torch.Tensor], None]):
        model_state_dict = mindio_acp.load(checkpoint_name, map_location="cpu")
        model_buffer, tensors_dict = model_params_save(model_state_dict)
        model_bytes_tensor = torch.ByteTensor(torch.ByteStorage.from_buffer(model_buffer.getvalue())).to("npu")
        model_bytes_size_tensor = torch.tensor([model_bytes_tensor.numel()], dtype=torch.long, device="npu")
        # broadcast the size of model states tensor
        if broadcast(model_bytes_size_tensor) is None:
            logging.debug('broadcast the size of model states tensor error.')
        # broadcast the serialized model state tensor data
        if broadcast(model_bytes_tensor) is None:
            logging.debug('broadcast the serialized model state tensor data error.')

        # broadcast the model params tensors
        for _, tensor in tensors_dict.items():
            if broadcast(tensor) is None:
                logging.debug('broadcast the model params tensors error.')
        return model_state_dict

    @staticmethod
    def __receive_checkpoint_at_dst_rank(broadcast: Callable[[torch.Tensor], None]):
        model_bytes_size_tensor = torch.empty(1, dtype=torch.long, device='npu')
        # receive the size of model states tensor
        if broadcast(model_bytes_size_tensor) is None:
            logging.debug('receive the size of model states tensor error.')
        model_bytes_tensor = torch.empty(model_bytes_size_tensor.data, dtype=torch.uint8, device='npu')
        # receives the serialized model state tensor data
        if broadcast(model_bytes_tensor) is None:
            logging.debug('receives the serialized model state tensor data error.')

        # load model params
        model_buffer = model_bytes_tensor.cpu().numpy().tobytes()
        buffer = io.BytesIO(model_buffer)
        model_state_dict = model_params_load(buffer, broadcast)
        return model_state_dict

    @staticmethod
    @retry(wait_min=100, wait_max=800)
    def __get_iteration_from_tracker_file(load_path: str):
        """Get iteration code from tracker file."""
        tracker_filename = os.path.join(load_path, 'latest_checkpointed_iteration.txt')
        # If no tracker file, print log and raise exception
        if not os.path.isfile(tracker_filename):
            raise ValueError(f'could not find the metadata file {get_relative_path(tracker_filename)}')

        # Otherwise, read the tracker file and get the iteration
        with open(tracker_filename, 'r') as f:
            iteration = int(f.read().strip())
        return iteration

    @staticmethod
    def __preload_checkpoint(load_path: str, iteration: int):
        """Get checkpoint file name and preload it into cache.
        """
        # generate checkpoint file name.
        file_paths = []
        directory = os.path.join(load_path, 'iter_{:07d}'.format(iteration))
        for root, _, files in os.walk(directory):
            for file in files:
                file_path = os.path.join(root, file)
                file_paths.append(file_path)

        # Preload the checkpoint.
        ret = mindio_acp.preload(*file_paths)
        if ret != 0:
            logging.warning('failed preload checkpoint from %s at iteration %d.',
                            get_relative_path(load_path), iteration)

    @time_used
    def rapid_load_model_checkpoint(self, checkpoint_name: str, load_rank: int, process_group):
        """ Load the model parameters from the given checkpoint_name
        """
        if not process_group:
            model_state_dict = mindio_acp.load(checkpoint_name, map_location="cpu")
            return model_state_dict

        def broadcast(tensor):
            try:
                torch.distributed.broadcast(
                    tensor,
                    src=load_rank,
                    group=process_group)
                return tensor
            except Exception as e:
                logging.error('Broadcast failed. Exception is: %s', str(e))
                return None

        if self.__rank == load_rank:
            model_state_dict = self.__broadcast_checkpoint_from_src_rank(checkpoint_name, broadcast)
        else:
            model_state_dict = self.__receive_checkpoint_at_dst_rank(broadcast)

        torch.npu.synchronize()
        return model_state_dict

    @time_used
    def async_preload_checkpoint(self, load_path: str):
        # Read the tracker file and get the iteration.
        iteration = self.__get_iteration_from_tracker_file(load_path)

        if iteration is None:
            logging.warning('rank=%d, failed to get iteration from tracker file', self.__rank)
            return

        self.__preload_checkpoint(load_path, iteration)
