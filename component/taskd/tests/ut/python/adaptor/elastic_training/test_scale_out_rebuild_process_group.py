#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# Copyright (c) 2025, Huawei Technologies Co., Ltd. All rights reserved.
import os
import unittest
from unittest import mock
from unittest.mock import patch

from megatron.core import mpu
from megatron.training import args_test, get_timers, get_args

from taskd.python.adaptor.elastic_training import scale_out_rebuild_process_group_callback, common


class TestScaleOutRebuildProcessCallback(unittest.TestCase):
    @patch('torch.distributed.destroy_process_group')
    @patch('torch.distributed.init_process_group')
    @patch('taskd.python.adaptor.elastic_training.scale_out_rebuild_process_group_callback.destroy_all_process_group')
    @patch('taskd.python.adaptor.elastic_training.scale_out_rebuild_process_group_callback.get_timers')
    @patch('megatron.core.rerun_state_machine.destroy_rerun_state_machine')
    @patch('torch.distributed.get_world_size')
    @patch('torch.distributed.get_process_group_ranks')
    @patch('torch.distributed.new_group')
    @patch('torch.distributed.get_rank')
    def test_scale_out_rebuild_process_group_callback(self, mock_get_rank, mock_new_group, mock_group_ranks,
                               mock_world_size, mock_destroy_rerun_state_machine, mock_get_timers,
                                                      mock_destroy_all_process_group, mock_init_group,
                                                      mock_destroy_process_group):
        mock_get_timers.return_value = get_timers
        mock_get_rank.return_value = 0
        mock_world_size.return_value = 16
        train_args = [[[args_test()], args_test()]]
        params = "{\"scale-out-strategy\": \"DP\"}"
        scale_out_rebuild_process_group_callback.scale_out_rebuild_process_group_callback(
                [8], train_args, params)
        self.assertEqual('0', os.environ['TORCH_DIST_INIT_BARRIER'])

        common.SCALE_IN_WORLD_GROUP = None
        scale_out_rebuild_process_group_callback.scale_out_rebuild_process_group_callback(
            [8], train_args, params)
        self.assertEqual('0', os.environ['TORCH_DIST_INIT_BARRIER'])
        self.assertIsNotNone(mpu._EMBEDDING_GLOBAL_RANKS)

    @patch('torch.distributed')
    def test_init_context_parallel_group(self, mock_distributed):
        mock_process_group_nccl = mock.MagicMock()
        mock_process_group_nccl.Options = get_args
        mock_distributed.ProcessGroupNCCL = mock_process_group_nccl
        mock_distributed.get_world_size.return_value = 16
        mock_distributed.get_rank.return_value = 0
        nccl_common_cfg = {"cp": {}}
        scale_out_rebuild_process_group_callback.init_context_parallel_group(
                args_test(), 30, nccl_common_cfg)
        self.assertEqual(mpu._CONTEXT_PARALLEL_GLOBAL_RANKS, range(0, 4, 4))



if __name__ == '__main__':
    unittest.main()