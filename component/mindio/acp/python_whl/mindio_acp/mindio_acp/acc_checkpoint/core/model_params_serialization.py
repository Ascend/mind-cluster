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

import io
import pickle
from collections import OrderedDict
from typing import cast, Dict

import torch
from torch.serialization import location_tag
from torch.types import Storage

DEFAULT_PROTOCOL = 2


def model_params_save(state_dict, pickle_protocol=DEFAULT_PROTOCOL):
    """
    Use the picker to serialize data, extract the tensor data, and return the serialized data and tensor data.
    """
    tensors_dict = OrderedDict()
    id_map: Dict[int, str] = {}

    tensor_dtypes: Dict[int, torch.dtype] = {}

    def persistent_id(obj):
        if isinstance(obj, torch.Tensor):
            tensor = obj
            tensor_dtype = tensor.dtype
            tensor_shape = tensor.shape
            location = location_tag(tensor)
            storage = cast(Storage, tensor.to('npu'))

            # If storage is allocated, ensure that any other saved storages
            # pointing to the same data all have the same dtype. If storage is
            # not allocated, don't perform this check
            if storage.data_ptr() != 0:
                if storage.data_ptr() in tensor_dtypes:
                    if tensor_dtype != tensor_dtypes[storage.data_ptr()]:
                        raise RuntimeError(
                            'Cannot save multiple tensors or storages that '
                            'view the same data as different types')
                else:
                    tensor_dtypes[storage.data_ptr()] = tensor_dtype

            tensor_key = id_map.setdefault(storage._cdata, len(id_map))
            tensors_dict[tensor_key] = storage

            return ('tensor',
                    tensor_dtype,
                    tensor_key,
                    tensor_shape,
                    location)

        return None

    # Write the pickle data for `obj`
    data_buf = io.BytesIO()
    pickler = pickle.Pickler(data_buf, protocol=pickle_protocol)
    pickler.persistent_id = persistent_id
    pickler.dump(state_dict)
    return data_buf, tensors_dict


def model_params_load(buffer: io.BytesIO, receiver_tensor, **pickle_load_args):
    if 'encoding' not in pickle_load_args.keys():
        pickle_load_args['encoding'] = 'utf-8'

    loaded_tensors = {}

    def load_tensor(dtype, key, shape, location):

        tensor = torch.empty(shape, dtype=dtype, device='npu')
        tensor = receiver_tensor(tensor)
        tensor = tensor.to(location)
        loaded_tensors[key] = tensor

    def persistent_load(saved_id):
        if not isinstance(saved_id, tuple):
            raise TypeError(f"Expected 'saved_id' to be a tuple, but got {type(saved_id)}.")
        typename = saved_id[0]
        data = saved_id[1:]

        if typename != 'tensor':
            raise ValueError(f"Unknown typename for persistent_load, expected 'tensor' but got '{typename}'")
        tensor_dtype, key, shape, location = data
        if key not in loaded_tensors:
            load_tensor(tensor_dtype, key, shape, location)

        return loaded_tensors[key]

    # Load the data (which may in turn use `persistent_load` to load tensors)
    data_file = io.BytesIO(buffer.getvalue())

    unpickler = pickle.Unpickler(data_file, **pickle_load_args)
    unpickler.persistent_load = persistent_load
    state_dict = unpickler.load()

    torch._utils._validate_loaded_sparse_tensors()

    return state_dict
