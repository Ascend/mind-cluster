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


class MindclusterTest0001(unittest.TestCase):
    """
        Template of the presmoke test
        To complete it, you need to:
        1. set the variable of where you set the job yaml
        2. prepare the environment in setup func
        3. run your job with test cases
        4. clean the environment with tear down func
        please to take care not making effect to others' testcase
    """

    @classmethod
    def setUpClass(cls) -> None:
        pass

    @classmethod
    def tearDownClass(cls):
        pass

    def setUp(self, methodName='mindcluster_ascend800ta2_schedule_0001'):
        pass

    def test_001(self):
        pass

    def tearDown(self) -> None:
        pass

