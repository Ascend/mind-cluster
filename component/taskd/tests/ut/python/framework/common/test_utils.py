#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# Copyright 2026. Huawei Technologies Co.,Ltd. All rights reserved.
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
import os
from taskd.python.framework.common.utils import get_report_fault_timeout
from taskd.python.toolkit.constants import constants

class TestGetReportTimeout(unittest.TestCase):
    def setUp(self):
        self.original_env = os.environ.copy()

    def tearDown(self):
        os.environ.clear()
        os.environ.update(self.original_env)

    def test_get_report_timeout_disabled(self):
        os.environ[constants.REPORT_FAULT_TIMEOUT_ENV] = '-1'
        result = get_report_fault_timeout()
        self.assertEqual(result, constants.REPORT_FAULT_TIMEOUT_DISABLED)

    def test_get_report_timeout_valid_value(self):
        os.environ[constants.REPORT_FAULT_TIMEOUT_ENV] = '450'
        result = get_report_fault_timeout()
        self.assertEqual(result, 450)

    def test_get_report_timeout_invalid_value(self):
        os.environ[constants.REPORT_FAULT_TIMEOUT_ENV] = 'invalid'
        result = get_report_fault_timeout()
        self.assertEqual(result, constants.REPORT_FAULT_TIMEOUT_DISABLED)

    def test_get_report_timeout_above_max(self):
        os.environ[constants.REPORT_FAULT_TIMEOUT_ENV] = '700'
        result = get_report_fault_timeout()
        self.assertEqual(result, constants.REPORT_FAULT_TIMEOUT_DISABLED)