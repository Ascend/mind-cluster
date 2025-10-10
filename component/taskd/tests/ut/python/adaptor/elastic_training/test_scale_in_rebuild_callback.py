#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# Copyright (c) 2025, Huawei Technologies Co., Ltd. All rights reserved.
import unittest
from unittest import mock
from unittest.mock import patch

import torch
from megatron.training import args_test
from mindio_ttp.adaptor import tft_replica_group

from taskd.python.adaptor.elastic_training import common
from taskd.python.adaptor.elastic_training import scale_in_rebuild_callback


class TestScaleInRebuildCallback(unittest.TestCase):
    @patch('megatron.core.rerun_state_machine.destroy_rerun_state_machine')
    @patch('torch.distributed.get_world_size')
    @patch('torch.distributed.new_group')
    @patch('torch.distributed.get_rank')
    def test_scale_in_rebuild_callback(self, mock_get_rank, mock_new_group,
                               mock_world_size, mock_destroy_rerun_state_machine):
        mock_get_rank.return_value = 0
        mock_world_size.return_value = 16
        origin_get_process_group_ranks = torch.distributed.get_process_group_ranks
        torch.distributed.get_process_group_ranks = get_process_group_ranks
        train_args = [[[args_test()], args_test()]]
        params = "{\"scale-in-strategy\": \"DP\"}"
        new_dp_ranks = [0, 4]
        new_world_ranks = [0, 1, 2, 3, 4, 5, 6, 7]
        args_test.expert_model_parallel_size = 2
        try:
            scale_in_rebuild_callback.scale_in_rebuild_callback(new_dp_ranks, new_world_ranks, train_args, params)
        except Exception as e:
            self.assertIn('not support ep or cp bigger than 1', str(e))

        args_test.expert_model_parallel_size = 1
        scale_in_rebuild_callback.scale_in_rebuild_callback(new_dp_ranks, new_world_ranks, train_args, params)
        mock_destroy_rerun_state_machine.assert_called()
        torch.distributed.get_process_group_ranks = origin_get_process_group_ranks

    @patch('torch.distributed.new_group')
    @patch('torch.distributed.get_rank')
    def test_create_scale_in_replica_group(self, mock_get_rank, mock_new_group):
        mock_new_group.return_value = 1
        mock_get_rank.return_value = 0
        common.IS_FAULT_REPLICA_RANK = False
        scale_in_rebuild_callback.create_scale_in_replica_group(True, [0, 4], [8, 12])
        self.assertEqual(2, mock_new_group.call_count)
        self.assertEqual(1, tft_replica_group.DP_CP_REPLICA_GROUP)
        self.assertEqual(1, tft_replica_group.DP_CP_REPLICA_GROUP_GLOO)

        mock_new_group.return_value = 2
        mock_get_rank.return_value = 8
        scale_in_rebuild_callback.create_scale_in_replica_group(False, [0, 4], [8, 12])
        self.assertEqual(4, mock_new_group.call_count)
        self.assertEqual(2, tft_replica_group.DP_CP_REPLICA_GROUP)
        self.assertEqual(2, tft_replica_group.DP_CP_REPLICA_GROUP_GLOO)

    @patch('taskd.python.adaptor.elastic_training.scale_in_rebuild_callback.build_new_dp_cp_group')
    def test_get_fault_msgs(self, mock_build_new_dp_cp_group):
        old_dp = [0, 4, 8, 12]
        new_dp = [0, 4]
        dp_cp_replica_ranks = [0, 4]
        fault_idxs, fault_local_idxs, fault_first_group = scale_in_rebuild_callback.get_fault_msgs(0, old_dp,
                                                                                                   old_dp, new_dp,
                                                                                                 dp_cp_replica_ranks)
        self.assertEqual([2, 3], fault_idxs)
        self.assertEqual([0, 1], fault_local_idxs)
        self.assertFalse(fault_first_group)
        self.assertFalse(common.FAULT_RANK_IN_DP_CP_REPLICA_GROUP)

        old_dp = [0, 4, 8, 12]
        new_dp = [4, 8, 12]
        dp_cp_replica_ranks = [8, 12]
        fault_idxs, fault_local_idxs, fault_first_group = scale_in_rebuild_callback.get_fault_msgs(8, old_dp,
                                                                                                   old_dp, new_dp,
                                                                                                   dp_cp_replica_ranks)
        self.assertEqual([0], fault_idxs)
        self.assertEqual([0], fault_local_idxs)
        self.assertTrue(fault_first_group)
        self.assertFalse(common.FAULT_RANK_IN_DP_CP_REPLICA_GROUP)

        old_dp = [0, 4, 8, 12]
        new_dp = [0, 4, 8]
        dp_cp_replica_ranks = [8, 12]
        fault_idxs, fault_local_idxs, fault_first_group = scale_in_rebuild_callback.get_fault_msgs(8, old_dp, old_dp,
                                                                                                   new_dp,
                                                                                                   dp_cp_replica_ranks)
        self.assertEqual([3], fault_idxs)
        self.assertEqual([1], fault_local_idxs)
        self.assertFalse(fault_first_group)
        self.assertTrue(common.FAULT_RANK_IN_DP_CP_REPLICA_GROUP)

    def test_get_ranks_after_change_left(self):
        dp_cp_replica_ranks_length = 2
        fault_idxs = [1, 2]
        old_dp_ranks = [0, 8, 16, 24]
        changed_old_dp_ranks = scale_in_rebuild_callback.get_ranks_after_change_left(dp_cp_replica_ranks_length,
                                                                                     fault_idxs,
                                                                                     old_dp_ranks,
                                                                                     24)
        self.assertEqual(changed_old_dp_ranks, [0, 24, 16, 24])
        self.assertTrue(common.IS_FAULT_REPLICA_RANK)

    @patch('torch.distributed.new_group')
    @patch('torch.distributed.get_rank')
    def test_create_new_replica_group_for_changed_left(self, mock_get_rank, mock_new_group):
        mock_new_group.return_value = 1
        mock_get_rank.return_value = 24
        common.IS_FAULT_REPLICA_RANK = False
        scale_in_rebuild_callback.create_new_replica_group_for_changed_left([0, 24])
        self.assertEqual(2, mock_new_group.call_count)
        self.assertEqual(1, tft_replica_group.DP_CP_REPLICA_GROUP)
        self.assertEqual(1, tft_replica_group.DP_CP_REPLICA_GROUP_GLOO)
        self.assertFalse(common.FAULT_RANK_IN_DP_CP_REPLICA_GROUP)

    @patch('torch.distributed.get_rank')
    def test_build_scale_in_dp_cp_replica_group_not_build(self, mock_get_rank):
        mock_get_rank.return_value = 24
        common.IS_FAULT_REPLICA_RANK = True
        scale_in_rebuild_callback.build_scale_in_dp_cp_replica_group([0, 1], False, False,
                                                                     [0, 1, 2, 3])
        self.assertFalse(common.IS_FAULT_REPLICA_RANK)

    @patch('taskd.python.adaptor.elastic_training.scale_in_rebuild_callback.torch')
    def test_init_pipeline_parallel_group(self, mock_torch):
        mock_torch.distributed.reinit_process_group = mock_reinit_process_group
        scale_in_rebuild_callback.init_pipeline_parallel_group(0)
        global called
        self.assertEqual(0, called)
        from megatron.core import mpu
        mpu._PIPELINE_GLOBAL_RANKS = [0]
        mpu._PIPELINE_MODEL_PARALLEL_GROUP = 1
        mpu._EMBEDDING_GLOBAL_RANKS = [0]
        mpu._EMBEDDING_GROUP = 1
        mpu._POSITION_EMBEDDING_GLOBAL_RANKS = [0]
        mpu._POSITION_EMBEDDING_GROUP = 1
        scale_in_rebuild_callback.init_pipeline_parallel_group(0)
        self.assertEqual(3, called)
        called = 0
        mpu._PIPELINE_GLOBAL_RANKS = []
        mpu._PIPELINE_MODEL_PARALLEL_GROUP = None
        mpu._EMBEDDING_GLOBAL_RANKS = []
        mpu._EMBEDDING_GROUP = None
        mpu._POSITION_EMBEDDING_GLOBAL_RANKS = []
        mpu._POSITION_EMBEDDING_GROUP = None

    @patch('taskd.python.adaptor.elastic_training.scale_in_rebuild_callback.torch')
    def test_init_context_parallel_group(self, mock_torch):
        mock_torch.distributed.reinit_process_group = mock_reinit_process_group
        global called
        from megatron.core import mpu
        mpu._CONTEXT_PARALLEL_GROUP = 1
        scale_in_rebuild_callback.init_context_parallel_group()
        self.assertEqual(1, called)
        called = 0
        mpu._CONTEXT_PARALLEL_GROUP = None

    @patch('taskd.python.adaptor.elastic_training.scale_in_rebuild_callback.torch')
    def test_init_model_parallel_group(self, mock_torch):
        mock_torch.distributed.reinit_process_group = mock_reinit_process_group
        global called
        from megatron.core import mpu
        mpu._MODEL_PARALLEL_GROUP = 1
        scale_in_rebuild_callback.init_model_parallel_group()
        self.assertEqual(1, called)
        called = 0
        mpu._MODEL_PARALLEL_GROUP = None

    @patch('taskd.python.adaptor.elastic_training.scale_in_rebuild_callback.torch')
    def test_init_tensor_parallel_group(self, mock_torch):
        mock_torch.distributed.reinit_process_group = mock_reinit_process_group
        global called
        from megatron.core import mpu
        mpu._TENSOR_MODEL_PARALLEL_GROUP = 1
        scale_in_rebuild_callback.init_tensor_parallel_group()
        self.assertEqual(1, called)
        called = 0
        mpu._TENSOR_MODEL_PARALLEL_GROUP = None


if __name__ == '__main__':
    unittest.main()


def get_process_group_ranks(group):
    if group == 'dp_cp_replica_group':
        return [0, 4]
    return [0, 4, 8, 12]

called = 0
def mock_reinit_process_group(group, rebuild_link):
    global called
    called = called + 1