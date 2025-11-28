#!/usr/bin/env python
# coding=utf-8
# Copyright (c) Facebook, Inc. and its affiliates.
# All rights reserved.
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
#
# Modification description: Implement data serialization and deserialization using pickle.
import io
import logging
import pickle
import struct
from collections import OrderedDict
from pickle import Pickler, Unpickler
from typing import Union, Tuple, Dict, Optional, cast

import torch
from torch.types import Storage
from torch.storage import _get_dtype_from_pickle_storage_type

if torch.__version__ > torch.torch_version.TorchVersion('1.12'):
    from torch.storage import TypedStorage as TypedStorage
    from torch.storage import UntypedStorage as UntypedStorage
else:
    from torch.storage import _TypedStorage as TypedStorage
    from torch import _UntypedStorage as UntypedStorage

_package_registry = []
TORCH_NPU_EXIST = -1
IMPORTED_TORCH_NPU = None
MINDIO_FLAG = b'mindio\x00'


def _prepare_write(key, storage):
    name = f'data/{key}'
    if storage.device.type != 'cpu':
        storage = _copy_storage_to_cpu_non_blocking(storage)
    # Now that it is on the CPU we can directly copy it into mindio
    num_bytes = storage.nbytes()
    return name, num_bytes, storage


def _copy_storage_to_cpu_non_blocking(storage):
    global IMPORTED_TORCH_NPU
    if not IMPORTED_TORCH_NPU:
        import torch_npu
        IMPORTED_TORCH_NPU = torch_npu

    fake_tensor = IMPORTED_TORCH_NPU._C._tensor_construct_from_storage(storage)
    if torch.__version__ > torch.torch_version.TorchVersion('1.12'):
        return fake_tensor.to('cpu', non_blocking=True).untyped_storage()
    else:
        return fake_tensor.to('cpu', non_blocking=True).storage()


def _rollback_default_to_cpu(backup_to_cpu):
    if backup_to_cpu is not None:
        torch.Tensor.cpu = backup_to_cpu


def _replace_default_to_cpu():
    if _check_torch_npu_exist():
        backup_to_cpu = torch.Tensor.cpu
        torch.Tensor.cpu = _non_blocking_to_cpu
        return backup_to_cpu
    return None


def _non_blocking_to_cpu(self):
    return self.to('cpu', non_blocking=True)


def _check_torch_npu_exist():
    global TORCH_NPU_EXIST
    if TORCH_NPU_EXIST < 0:
        import importlib.util
        if importlib.util.find_spec('torch.npu') is None:
            TORCH_NPU_EXIST = 0
        else:
            TORCH_NPU_EXIST = 1

    return TORCH_NPU_EXIST > 0


def _wait_async_device_tasks():
    if _check_torch_npu_exist():
        try:
            torch.npu.synchronize()
        except RuntimeError as e:
            if 'FORCE STOP' in str(e):
                logging.warning('[torch_mindio] async thread synchronize with ttp err force stop conflict.'
                                'Exception is: {}' .format(e))
            else:
                raise


def register_deserializer(priority, tagger, deserializer):
    queue_elem = (priority, tagger, deserializer)
    _package_registry.append(queue_elem)
    _package_registry.sort()


def get_torch_storage(typed_storage: TypedStorage):
    if torch.__version__ > torch.torch_version.TorchVersion('1.12'):
        return typed_storage.untyped()
    else:
        return typed_storage._storage


def _cpu_tag(obj):
    if torch.__version__ > torch.torch_version.TorchVersion('1.12'):
        if obj.device.type == 'cpu':
            return 'cpu'
    else:
        if type(obj).__module__ == 'torch':
            return 'cpu'
    return None


def _npu_tag(obj):
    if torch.__version__ > torch.torch_version.TorchVersion('1.12'):
        if obj.device.type == 'npu':
            return 'npu'
    else:
        if type(obj).__module__ == 'torch.npu':
            return 'npu:' + str(obj.get_device())
    return None


def _cuda_tag(obj):
    if torch.__version__ > torch.torch_version.TorchVersion('1.12'):
        if obj.device.type == 'cuda':
            return 'cuda'
    else:
        if type(obj).__module__ == 'torch.cuda':
            return 'cuda:' + str(obj.get_device())
    return None


def _cpu_relocation(storage_obj, location):
    if location == 'cpu':
        return storage_obj
    return None


def validate_cuda_device(location):
    device = torch.cuda._utils._get_device_index(location, True)

    if not torch.cuda.is_available():
        raise RuntimeError('The object was serialized with CUDA but torch.cuda.is_available() is False. '
                           'For CPU-only machines, please use torch.load with map_location=torch.device(\'cpu\'))')
    device_count = torch.cuda.device_count()
    if device >= device_count:
        raise RuntimeError(f'Requested CUDA device {device} is unavailable. System has only {device_count} device(s). '
                           f'Please use torch.load with map_location to map your storages to an existing device.')
    return device


def validate_npu_device(location):
    device = torch.npu.utils._get_device_index(location, True)
    if not torch.npu.is_available():
        raise RuntimeError('Attempting to deserialize object on a NPU '
                           'device but torch.npu.is_available() is False. '
                           'If you are running on a CPU-only machine, '
                           'please use torch.load with map_location=torch.device(\'cpu\') '
                           'to map your storages to the CPU.')
    device_count = torch.npu.device_count()
    if device >= device_count:
        raise RuntimeError('Attempting to deserialize object on NPU device '
                           f'{device} but torch.cuda.device_count() is {device_count}. Please use '
                           'torch.load with map_location to map your storages '
                           'to an existing device.')
    return device


def _npu_relocation(obj, location):
    if location.startswith('npu'):
        device = validate_npu_device(location)
        if getattr(obj, "_torch_load_uninitialized", False):
            with torch.npu.device(device):
                return torch.UntypedStorage(obj.nbytes(), device=torch.device(location))
        else:
            return obj.npu(device)
    return None


def _cuda_relocation(obj, location):
    if location.startswith('cuda'):
        device = validate_cuda_device(location)
        if getattr(obj, "_torch_load_uninitialized", False):
            with torch.cuda.device(device):
                return torch.UntypedStorage(obj.nbytes(), device=torch.device(location))
        else:
            return obj.cuda(device)
    return None


register_deserializer(10, _cpu_tag, _cpu_relocation)
register_deserializer(15, _npu_tag, _npu_relocation)
register_deserializer(20, _cuda_tag, _cuda_relocation)


def location_tag(storage):
    for _, tagger, _ in _package_registry:
        location = tagger(storage)
        if location:
            return location
    raise RuntimeError("don't know how to determine data location of " + torch.typename(storage))


def _maybe_decode_ascii(bytes_str: Union[bytes, str]) -> str:
    if isinstance(bytes_str, bytes):
        return bytes_str.decode('ascii')
    return bytes_str


def default_restore_function(storage, location):
    for _, _, fn in _package_registry:
        result = fn(storage, location)
        if result is not None:
            return result
    raise RuntimeError("don't know how to restore data location of " + torch.typename(storage) + " (tagged with "
                       + location + ")")


def normalize_storage_type(storage_type):
    return getattr(torch, storage_type.__name__)


def _get_restore_func(map_location):
    if map_location is None:
        restore_loc = default_restore_function
    elif isinstance(map_location, dict):
        def restore_loc(storage, location):
            location = map_location.get(location, location)
            return default_restore_function(storage, location)
    elif isinstance(map_location, (str, bytes)):
        def restore_loc(storage, location):
            return default_restore_function(storage, map_location)
    elif isinstance(map_location, torch.device):
        def restore_loc(storage, location):
            return default_restore_function(storage, str(map_location))
    else:
        def restore_loc(storage, location):
            result = map_location(storage, location)
            if result is None:
                result = default_restore_function(storage, location)
            return result
    return restore_loc


class SerializationMixin:

    @staticmethod
    def save_bytes(data: bytes):
        record_map = {}
        logging.info('save directly bytes data size = %d', len(data))
        record_map['mindio_save_data_type'] = 'directly_bytes'
        record_map['bytes_length'] = len(data)
        record_map_str = pickle.dumps(record_map)
        return data, None, record_map_str

    @staticmethod
    def unmarshal_checkpoint(handler, record_map, data_pkl_bytes, map_location=None, **pickle_load_args) -> Union[
        dict, bytes]:
        restore_loc = _get_restore_func(map_location)

        loaded_storages = {}
        loaded_storages_multi_ptr = []
        loaded_storages_reference = []

        def load_tensor_prepare(dtype, nbytes, key, location):
            name = f'data/{key}'

            storage = cast(Storage, UntypedStorage(nbytes))
            start, size = record_map[name]
            loaded_storages_multi_ptr.append([storage.data_ptr(), start, size])
            # Increase the reference count
            loaded_storages_reference.append([storage])

            loaded_storages[key] = TypedStorage(
                wrap_storage=restore_loc(storage, location),
                dtype=dtype)

        def multi_load_tensors():
            ret = handler.file.multi_read(loaded_storages_multi_ptr)

        def persistent_load(saved_id):
            if not isinstance(saved_id, tuple):
                raise RuntimeError("saved_id is not tuple.")
            typename = _maybe_decode_ascii(saved_id[0])
            data = saved_id[1:]

            if not typename == 'storage':
                raise RuntimeError(f"Unknown typename for persistent_load, expected 'storage' but got '{typename}'")

            storage_type, key, location, numel = data
            dtype = storage_type.dtype

            if key not in loaded_storages:
                nbytes = numel * torch._utils._element_size(dtype)
                load_tensor_prepare(dtype, nbytes, key, _maybe_decode_ascii(location))

            return loaded_storages[key]

        load_module_mapping: Dict[str, str] = {
            'torch.tensor': 'torch._tensor'
        }

        class UnpicklerWrapper(Unpickler):  # type: ignore[name-defined]
            def find_class(self, module_name, name):
                if isinstance(name, str) and 'Storage' in name:
                    try:
                        return StorageType(name)
                    except KeyError:
                        pass
                module_name = load_module_mapping.get(module_name, module_name)
                return super().find_class(module_name, name)

        data_file = io.BytesIO(data_pkl_bytes)

        unpickler = UnpicklerWrapper(data_file, **pickle_load_args)
        unpickler.persistent_load = persistent_load

        result = unpickler.load()

        if len(loaded_storages_multi_ptr) > 0:
            multi_load_tensors()

        torch._utils._validate_loaded_sparse_tensors()

        return result

    def marshal_checkpoint(self, ckpt_obj: Union[dict, bytes]) -> Tuple[bytes, Optional[OrderedDict], Dict]:
        """
        Using the pickle to serialize checkpoint object
        Args:
            ckpt_obj (dict/bytes): checkpoint object to be serialized

        Returns: the serialized data, tensors dict and the record_map
        """
        record_map = {}
        if isinstance(ckpt_obj, bytes):
            return self.save_bytes(ckpt_obj)

        serialized_storages = OrderedDict()
        id_map: Dict[int, str] = {}
        storage_dtypes: Dict[int, torch.dtype] = {}
        start = 0

        def persistent_id(obj):
            if isinstance(obj, TypedStorage) or torch.is_storage(obj):
                if isinstance(obj, TypedStorage):
                    storage = get_torch_storage(obj)
                    stg_dtype = obj.dtype
                    stg_type_str = obj.pickle_storage_type()
                    stg_type = getattr(torch, stg_type_str)
                    stg_numel = obj.size()
                else:
                    storage = obj
                    stg_dtype = storage.dtype
                    stg_type = normalize_storage_type(type(obj))
                    stg_numel = storage.nbytes()

                storage = cast(Storage, storage)
                data_ptr = storage.data_ptr()
                if data_ptr != 0:
                    if data_ptr in storage_dtypes and stg_dtype != storage_dtypes[data_ptr]:
                        raise RuntimeError(
                            'Cannot save multiple tensors or storages that '
                            'view the same data as different types')
                    elif data_ptr not in storage_dtypes:
                        storage_dtypes[data_ptr] = stg_dtype

                location = location_tag(storage)
                storage_key = id_map.setdefault(storage._cdata, str(len(id_map)))
                serialized_storages[storage_key] = storage

                return ('storage', stg_type, storage_key, location, stg_numel)

            return None

        data_buf = io.BytesIO()
        pickler = Pickler(data_buf, protocol=2)
        pickler.persistent_id = persistent_id
        backup_to_cpu_method = _replace_default_to_cpu()
        pickler.dump(ckpt_obj)
        _rollback_default_to_cpu(backup_to_cpu_method)
        data_value = data_buf.getvalue()

        record_map['data.pkl'] = (start, len(data_value))
        start += len(data_value)
        for key in serialized_storages:
            name, num_bytes, serialized_storages[key] = _prepare_write(key, serialized_storages[key])
            record_map[name] = (start, num_bytes)
            start += num_bytes
        _wait_async_device_tasks()
        record_buff = pickle.dumps(record_map)
        return data_value, serialized_storages, record_buff


class PrepareWriteMixin:
    @staticmethod
    def get_write_content(data_buff: bytes, tensors_dict: Optional[OrderedDict], record_buff: bytes):
        """
        get write content for data_buff tensors_dict and record_buff
        And add MindioFileTail to the tail of marshalled bytes.
        Args:
            data_buff:
            tensors_dict:
            record_buff:

        Returns:
        """
        if not tensors_dict:
            # prepare MindioFileTail.
            start = len(data_buff)
            start_bytes = struct.pack('Q', start)
            record_map_length = len(record_buff)
            record_map_length_bytes = struct.pack('Q', record_map_length)

            # add MindioFileTail to the tail of marshalled record_map bytes.
            write_content = b''.join([data_buff, record_buff, start_bytes, record_map_length_bytes, MINDIO_FLAG])
            return write_content
        else:
            write_content_list = list()
            start = 0
            write_content_list.append((data_buff, len(data_buff)))
            start += len(data_buff)
            for key in tensors_dict:
                num_bytes = tensors_dict[key].nbytes()
                write_content_list.append((tensors_dict[key].data_ptr(), num_bytes))
                start += num_bytes

            # prepare MindioFileTail.
            start_bytes = struct.pack('Q', start)
            record_map_length = len(record_buff)
            record_map_length_bytes = struct.pack('Q', record_map_length)
            # add MindioFileTail to the tail of marshalled record_map bytes.
            record_buff_with_tail = b''.join([record_buff, start_bytes, record_map_length_bytes, MINDIO_FLAG])
            write_content_list.append((record_buff_with_tail, len(record_buff_with_tail)))
            return write_content_list


class StorageType():
    def __init__(self, name):
        self.dtype = _get_dtype_from_pickle_storage_type(name)

    def __str__(self):
        return f'StorageType(dtype={self.dtype})'