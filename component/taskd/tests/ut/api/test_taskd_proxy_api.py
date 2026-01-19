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
from unittest.mock import MagicMock, patch
import os
from taskd.python.cython_api import cython_api
from taskd.python.framework.common.type import DEFAULT_SERVERRANK
from taskd.taskd.api import taskd_proxy_api


class TestTaskdProxyAPI(unittest.TestCase):
    def setUp(self):
        # Backup original cython_api.lib
        self.original_lib = cython_api.lib

    def tearDown(self):
        # Restore original cython_api.lib
        cython_api.lib = self.original_lib

    @patch('os.getenv')
    def test_init_taskd_proxy_success(self, mock_getenv):
        # Mock environment variables
        def mock_getenv_side_effect(key, default=None):
            if key in ["RANK", "MS_NODE_RANK"]:
                return DEFAULT_SERVERRANK
            else:
                return None
        mock_getenv.side_effect = mock_getenv_side_effect

        # Mock cython_api.lib
        cython_api.lib = MagicMock()
        cython_api.lib.InitTaskdProxy.return_value = 0

        config = {}
        result = taskd_proxy_api.init_taskd_proxy(config)
        self.assertTrue(result)

    @patch('os.getenv')
    def test_init_taskd_proxy_lib_not_loaded(self, mock_getenv):
        # Mock environment variables
        def mock_getenv_side_effect(key, default=None):
            if key in ["RANK", "MS_NODE_RANK"]:
                return DEFAULT_SERVERRANK
            else:
                return None
        mock_getenv.side_effect = mock_getenv_side_effect

        cython_api.lib = None
        result = taskd_proxy_api.init_taskd_proxy({})
        self.assertFalse(result)

    @patch('os.getenv')
    def test_init_taskd_proxy_init_failed(self, mock_getenv):
        # Mock environment variables
        def mock_getenv_side_effect(key, default=None):
            if key in ["RANK", "MS_NODE_RANK"]:
                return DEFAULT_SERVERRANK
            else:
                return None
        mock_getenv.side_effect = mock_getenv_side_effect

        # Mock cython_api.lib
        cython_api.lib = MagicMock()
        cython_api.lib.InitTaskdProxy.return_value = 1

        config = {}
        result = taskd_proxy_api.init_taskd_proxy(config)
        self.assertFalse(result)

    @patch('os.getenv')
    def test_init_taskd_proxy_exception(self, mock_getenv):
        # Mock environment variables
        def mock_getenv_side_effect(key, default=None):
            if key in ["RANK", "MS_NODE_RANK"]:
                return DEFAULT_SERVERRANK
            else:
                return None
        mock_getenv.side_effect = mock_getenv_side_effect

        # Mock cython_api.lib
        cython_api.lib = MagicMock()
        cython_api.lib.InitTaskdProxy.side_effect = Exception("Mock exception")

        config = {}
        result = taskd_proxy_api.init_taskd_proxy(config)
        self.assertFalse(result)

    def test_destroy_taskd_proxy_success(self):
        # Mock cython_api.lib
        cython_api.lib = MagicMock()
        cython_api.lib.DestroyTaskdProxy = MagicMock()

        result = taskd_proxy_api.destroy_taskd_proxy()
        self.assertTrue(result)

    def test_destroy_taskd_proxy_lib_not_loaded(self):
        cython_api.lib = None
        result = taskd_proxy_api.destroy_taskd_proxy()
        self.assertFalse(result)

    def test_destroy_taskd_proxy_exception(self):
        # Mock cython_api.lib
        cython_api.lib = MagicMock()
        cython_api.lib.DestroyTaskdProxy.side_effect = Exception("Mock exception")

        result = taskd_proxy_api.destroy_taskd_proxy()
        self.assertFalse(result)


if __name__ == '__main__':
    unittest.main()
    