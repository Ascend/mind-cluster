#!/usr/bin/env python3
# coding: utf-8
# Copyright 2026 Huawei Technologies Co., Ltd
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
    base_dir = "/workspace/mind-cluster/tests/st/testcases/"

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


if __name__ == '__main__':
    unittest.main()
