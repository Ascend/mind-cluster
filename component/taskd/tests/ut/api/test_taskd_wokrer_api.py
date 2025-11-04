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
from unittest.mock import patch

from taskd.api.taskd_worker_api import init_taskd_worker, destroy_taskd_worker


class WorkerTestCase(unittest.TestCase):
    def test_init_taskd_worker_success(self, mock_worker):
        rank_id = 'not_an_int'
        upper_limit = 5000
        result = init_taskd_worker(rank_id, upper_limit)
        self.assertFalse(result)
        rank_id = 1
        upper_limit = 'not_an_int'
        result = init_taskd_worker(rank_id, upper_limit)
        self.assertFalse(result)
        rank_id = -1
        upper_limit = 500
        result = init_taskd_worker(rank_id, upper_limit)
        self.assertFalse(result)
        rank_id = 1
        upper_limit = 400
        result = init_taskd_worker(rank_id, upper_limit)
        self.assertFalse(result)

    def test_destroy_worker_networker(self, mock_worker):
        result = destroy_taskd_worker()
        self.assertFalse(result)

if __name__ == '__main__':
    unittest.main()
