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
import unittest
from unittest.mock import patch, MagicMock
import os
import json
import tempfile
from unittest import mock
from taskd.python.adaptor.pytorch.group_info import get_save_path, get_group_info, dump_group_info
from taskd.python.constants.constants import JOB_ID_KEY, GROUP_BASE_DIR_ENV, DEFAULT_GROUP_DIR, GROUP_INFO_NAME


class TestGroupInfoFunctions(unittest.TestCase):

    @patch('os.getenv')
    @patch('os.path.exists')
    @patch('os.makedirs')
    def test_get_save_path(self, mock_makedirs, mock_exists, mock_getenv):
        mock_getenv.side_effect = [None, None]
        result = get_save_path(1)
        self.assertEqual(result, "")

        mock_getenv.side_effect = ['job_id', 'base_dir']
        mock_exists.return_value = True
        result = get_save_path(1)
        self.assertEqual(result, os.path.join('base_dir', 'job_id', '1'))

        mock_getenv.side_effect = ['job_id', 'base_dir']
        mock_exists.return_value = False
        mock_makedirs.side_effect = OSError('Test error')
        result = get_save_path(1)
        self.assertEqual(result, "")


class TestGroupInfo(unittest.TestCase):
    def setUp(self):
        self.original_env = os.environ.copy()
        os.environ[JOB_ID_KEY] = "test_job_id"
        os.environ[GROUP_BASE_DIR_ENV] = tempfile.mkdtemp()
        self.test_rank = 0

    def tearDown(self):
        os.environ.clear()
        os.environ.update(self.original_env)
        if os.path.exists(os.environ.get(GROUP_BASE_DIR_ENV, '')):
            import shutil
            shutil.rmtree(os.environ[GROUP_BASE_DIR_ENV], ignore_errors=True)

    @mock.patch('torch.distributed.is_available')
    def test_get_group_info_distributed_not_available(self, mock_is_available):
        mock_is_available.return_value = False
        
        result = get_group_info(self.test_rank)
        self.assertEqual(result, {})

    @mock.patch('torch.distributed.is_available')
    @mock.patch('torch.distributed.is_initialized')
    def test_get_group_info_distributed_not_initialized(self, mock_is_initialized, mock_is_available):
        mock_is_available.return_value = True
        mock_is_initialized.return_value = False
        result = get_group_info(self.test_rank)
        self.assertEqual(result, {})

    @mock.patch('torch.distributed.is_available')
    @mock.patch('torch.distributed.is_initialized')
    @mock.patch('torch.distributed.get_group_rank')
    @mock.patch('torch.distributed.get_process_group_ranks')
    @mock.patch('torch.distributed.distributed_c10d._get_default_group')
    @mock.patch('torch.distributed.distributed_c10d._world')
    @mock.patch('torch.device')
    def test_get_group_info_success(self, mock_device, mock_distributed_world, mock_get_default_group,
                                    mock_get_process_group_ranks, mock_get_group_rank, mock_is_initialized,
                                    mock_is_available):
        mock_is_available.return_value = True
        mock_is_initialized.return_value = True

        mock_group = mock.MagicMock()
        mock_group_config = ["hccl"]
        mock_distributed_world.pg_map = {mock_group: mock_group_config}

        mock_device.return_value = None
        mock_backend = mock.MagicMock()
        mock_hccl_group = mock.MagicMock()
        mock_hccl_group.get_hccl_comm_name.return_value = "comm_name"
        mock_hccl_group.options.hccl_config = {"group_name": "test_group"}
        mock_backend.return_value = mock_hccl_group
        mock_group._get_backend = mock_backend
        
        mock_get_group_rank.return_value = 0
        mock_get_process_group_ranks.return_value = [0, 1, 2]
        
        mock_default_group = mock.MagicMock()
        mock_default_backend = mock.MagicMock()
        mock_default_hccl_group = mock.MagicMock()
        mock_default_hccl_group.get_hccl_comm_name.return_value = "default_comm_name"
        mock_default_backend.return_value = mock_default_hccl_group
        mock_default_group._get_backend = mock_default_backend
        mock_get_default_group.return_value = mock_default_group
        
        import torch
        torch.distributed.distributed_c10d._world = mock_distributed_world
        
        result = get_group_info(self.test_rank)
        self.assertIsInstance(result, dict)
        self.assertIn("comm_name", result)
        self.assertIn("default_comm_name", result)

    @mock.patch('torch.distributed.is_available')
    def test_get_group_info_exception(self, mock_is_available):
        mock_is_available.side_effect = Exception("test exception")
        
        result = get_group_info(self.test_rank)
        self.assertEqual(result, {})

    def test_get_save_path_invalid_job_id(self):
        os.environ[JOB_ID_KEY] = ""
        result = get_save_path(self.test_rank)
        self.assertEqual(result, "")

    def test_get_save_path_default_dir(self):
        os.environ[GROUP_BASE_DIR_ENV] = "/non/existent/path"
        result = get_save_path(self.test_rank)
        self.assertTrue(result.startswith(DEFAULT_GROUP_DIR))

    @mock.patch('torch.distributed.get_rank')
    @mock.patch('taskd.python.adaptor.pytorch.group_info.get_group_info')
    @mock.patch('taskd.python.adaptor.pytorch.group_info.get_save_path')
    def test_dump_group_info_success(self, mock_get_save_path, mock_get_group_info, mock_get_rank):
        mock_get_rank.return_value = self.test_rank
        mock_get_group_info.return_value = {"comm_name": {"group_name": "test_group"}}
        mock_get_save_path.return_value = tempfile.mkdtemp()
        
        dump_group_info()
        
        full_path = os.path.join(mock_get_save_path.return_value, GROUP_INFO_NAME)
        self.assertTrue(os.path.exists(full_path))
        
        with open(full_path, "r", encoding="utf-8") as f:
            content = json.load(f)
            self.assertEqual(content, {"comm_name": {"group_name": "test_group"}})

    @mock.patch('torch.distributed.get_rank')
    @mock.patch('taskd.python.adaptor.pytorch.group_info.run_log.error')
    def test_dump_group_info_exception(self, mock_error, mock_get_rank):
        mock_get_rank.side_effect = Exception("test exception")
        
        dump_group_info()

        mock_error.assert_any_call('save group info failed: test exception')


if __name__ == '__main__':
    unittest.main()
