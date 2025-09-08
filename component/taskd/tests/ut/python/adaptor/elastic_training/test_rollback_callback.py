#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# Copyright (c) 2025, Huawei Technologies Co., Ltd. All rights reserved.
import unittest
from unittest import mock
from unittest.mock import patch

from megatron.training import args_test
from taskd.python.adaptor.elastic_training import common
from taskd.python.adaptor.elastic_training import rollback_callback


class TestRollbackCallback(unittest.TestCase):
    @patch('taskd.python.adaptor.elastic_training.rollback_callback.torch.distributed.get_rank')
    @patch('taskd.python.adaptor.elastic_training.rollback_callback.torch')
    def test_rollback_callback(self, mock_torch, mock_get_rank):
        mock_npu = mock.MagicMock()
        mock_npu.set_device = mock.MagicMock()
        mock_torch.npu = mock_npu
        train_args = [[[args_test()], args_test()]]
        params = "{\"scale-out-strategy\": \"DP\"}"
        common.ORIGIN_DP_SIZE = 4
        common.ORIGIN_NUM_MICRO_BATCHES = 1
        rollback_callback.rollback_callback(1, train_args, params)
        mock_get_rank.assert_called()

    def test_build_data_iterator(self):
        train_data_iterator, valid_data_iterator, test_data_iterator = rollback_callback.build_data_iterator([1])
        self.assertEqual(train_data_iterator, [0])
        self.assertEqual(valid_data_iterator, [1])
        self.assertEqual(test_data_iterator, [2])

        args_test.virtual_pipeline_model_parallel_size = None
        train_data_iterator, valid_data_iterator, test_data_iterator = rollback_callback.build_data_iterator([1])
        self.assertEqual(train_data_iterator, 0)
        self.assertEqual(valid_data_iterator, 1)
        self.assertEqual(test_data_iterator, 2)
        args_test.virtual_pipeline_model_parallel_size = 1


if __name__ == '__main__':
    unittest.main()