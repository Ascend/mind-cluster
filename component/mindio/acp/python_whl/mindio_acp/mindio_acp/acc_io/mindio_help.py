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
import atexit
import errno
from typing import Union, Dict, Callable, Any, Optional

from mindio_acp.c2python_api import callback_receiver
from mindio_acp.c2python_api import init_callback_handler
from mindio_acp.acc_io.read_handler import create_read_chain
from mindio_acp.acc_io.serialization import SerializationMixin, PrepareWriteMixin
from mindio_acp.acc_io.write_handler import create_write_chain
from mindio_acp.acc_io.write_handler import import_mindio_sdk_api
from mindio_acp.common import mindio_logger
from mindio_acp.common.check_process import check_process
from mindio_acp.common.utils import get_relative_path
from mindio_acp.launch_server_conf.default_memfs_conf import default_server_info
from mindio_acp.launch_server_conf.launch_server_param import ockiod_path, server_worker_dir

from mindio_acp.acc_io.multi_write_handler import create_multi_write_chain

logging = mindio_logger.LOGGER

OPEN_WAY_MEMFS = 'memfs'
OPEN_WAY_FOPEN = 'fopen'


class _MindioCallbackReciver(callback_receiver):
    def __init__(self):
        super().__init__()

global callback_handler
callback_handler = _MindioCallbackReciver()
init_callback_handler(callback_handler)


class _TorchMultiSaveHelp(SerializationMixin, PrepareWriteMixin):
    def __init__(self):
        super().__init__()

    def __call__(self, ckpt_obj: Union[Dict, bytes], path: list, open_way) -> int:
        import_mindio_sdk_api()
        if torch_initialize_helper(None) != 0:
            logging.warning(f"[mindio_acp] default initialize failed.")
        # 1. serialization ckpt_obj
        data_buff, tensors_dict, record_buff = self.marshal_checkpoint(ckpt_obj)

        # 2. construct write_content
        write_content = self.get_write_content(data_buff, tensors_dict, record_buff)

        # 3. write write_list
        writer = create_multi_write_chain(ckpt_obj, path, open_way)
        marker = writer.handle(write_content)
        return marker


class _TorchSaveHelp(SerializationMixin, PrepareWriteMixin):
    def __init__(self):
        super().__init__()

    def __call__(self, ckpt_obj: Union[Dict, bytes], path: str, open_way) -> int:
        import_mindio_sdk_api()
        if torch_initialize_helper(None) != 0:
            logging.warning(f"[mindio_acp] default initialize failed.")
        # 1. serialization ckpt_obj
        data_buff, tensors_dict, record_buff = self.marshal_checkpoint(ckpt_obj)

        # 2. construct write_content
        write_content = self.get_write_content(data_buff, tensors_dict, record_buff)

        # 3. write write_list
        writer = create_write_chain(ckpt_obj, path, open_way)
        marker = writer.handle(write_content)
        return marker


class _TorchLoadHelp(SerializationMixin):
    def __init__(self):
        super().__init__()

    def __call__(self, path, open_way, map_location) -> object:
        import_mindio_sdk_api()
        try:
            import torch_npu
            torch_npu.utils.serialization._update_cpu_remap_info(map_location)
        except Exception as e:
            logging.debug('Unable to import torch_npu. Exception is: %s', str(e))
            pass
        if torch_initialize_helper(None) != 0:
            logging.warning(f"[mindio_acp] default initialize failed.")
        reader = create_read_chain(path, open_way, map_location)
        return reader.handle()


torch_save_helper = _TorchSaveHelp()
torch_load_helper = _TorchLoadHelp()
multi_save_helper = _TorchMultiSaveHelp()


def torch_initialize_helper(server_info) -> int:
    import_mindio_sdk_api()
    from mindio_acp.acc_io.write_handler import initialize

    if server_info is not None:
        for key, value in server_info.items():
            if key not in default_server_info:
                logging.error('[mindio_acp] initialize invalid key: {}'.format(key))
                return -1
            default_server_info[key] = value
    default_server_info['server.ockiod.path'] = ockiod_path
    default_server_info['server.worker.path'] = server_worker_dir
    try:
        ret = initialize(default_server_info)
    except Exception as e:
        logging.error('[mindio_acp] initialize task failed finally. Exception is: {}.'.format(e))
        return -1
    if ret != 0:
        logging.info('LocalCache client initialize failed {}.'.format(ret))

    return ret


def torch_register_checker_helper(callback: Callable, check_dict: Dict, user_context: Any,
                                  timeout_sec: int) -> Optional[int]:
    import_mindio_sdk_api()
    from mindio_acp.acc_io.write_handler import register_checker
    if not isinstance(callback, Callable):
        logging.error(f"The parameter must be callable")
        return None
    if not isinstance(check_dict, Dict):
        logging.error(f"Argument must be a dictionary")
        return None
    if not check_dict:
        logging.error(f"input directories is empty.")
        return None
    callback_handler.check_dir_callback = callback
    try:
        ret = register_checker(check_dict, user_context, timeout_sec)
    except Exception as e:
        logging.error('[mindio_acp] register checker task failed finally. Exception is: {}.'.format(e))
        return None
    if ret != 0:
        return None
    return 1


def torch_multi_save_helper(obj: Union[Dict, bytes], path_list: list) -> Optional[int]:
    import_mindio_sdk_api()
    if torch_initialize_helper(None) != 0:
        logging.warning(f"[mindio_acp] default initialize failed.")
    if not isinstance(path_list, list):
        return None
    if len(path_list) == 0:
        logging.warning('[mindio_acp] multi_save: path is none.')
        return 0
    for path in path_list:
        if len(path) > 1024:
            logging.error('the file path length cannot exceed 1024 characters.')
            return None

    if not check_process():
        logging.warning(f'[mindio_acp] ockiod does not exist, using fopen way instead')
        for path in path_list:
            torch_save_helper(obj, os.path.realpath(path), OPEN_WAY_FOPEN)
        return 0
    if len(path_list) == 1:
        return torch_save_helper(obj, os.path.realpath(path_list[0]), OPEN_WAY_MEMFS)
    else:
        return multi_save_helper(obj, path_list, OPEN_WAY_MEMFS)


def torch_preload_helper(*path: str) -> int:
    import_mindio_sdk_api()
    from mindio_acp.acc_io.write_handler import preload
    if torch_initialize_helper(None) != 0:
        return 1
    if not path:
        logging.warning('[mindio_acp] preload: path is none.')
        return 1
    try:
        path_list = [os.path.realpath(single_path) for single_path in path]
        format_path = [get_relative_path(single_path) for single_path in path_list]
    except TypeError as e:
        logging.error('[mindio_acp] preload: path type error. Exception is: {}'.format(e))
        return 1
    if any([(not isinstance(single_path, str) or len(single_path) > 1024) for single_path in path_list]):
        logging.warning('[mindio_acp] preload: path %s invalid.', get_relative_path(path))
        return 1
    try:
        if not check_process():
            logging.error(f'[mindio_acp] preload failed: ockiod does not exist.')
            return 1
        if preload(path_list) != 0:
            logging.warning(f"[mindio_acp] preload failed.")
            return 1
    except Exception as e:
        logging.error('[mindio_acp] preload task failed finally. Exception is: {}.'.format(e))
        return 1

    logging.info('[mindio_acp] preload %s success.', format_path)
    return 0


def torch_wait_flush_helper():
    import_mindio_sdk_api()
    from mindio_acp.acc_io.write_handler import check_background_task

    if not check_process(False):
        return 0

    try:
        ret = check_background_task()
    except Exception as e:
        logging.error('[mindio_acp] check_background_task failed. Exception is: %s', str(e))
        return 1
    if ret != 0:
        logging.error("[mindio_acp] check_background_task failed.")
        return 1
    return 0


def torch_uninitialize_helper():
    import_mindio_sdk_api()
    from mindio_acp.acc_io.write_handler import un_initialize

    try:
        un_initialize()
    except Exception as e:
        logging.error('[mindio_acp] uninitialize failed. Exception is: %s', str(e))
        return 1

    return 0


atexit.register(torch_uninitialize_helper)
atexit.register(torch_wait_flush_helper)
