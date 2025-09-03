#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# Copyright (c) 2025, Huawei Technologies Co., Ltd. All rights reserved.
import unittest
from unittest.mock import patch

from taskd.python.adaptor.elastic_training import repair_callback


class TestRepairCallback(unittest.TestCase):
    @patch('torch.distributed.get_rank')
    def test_repair_callback_invalid_step(self, mock_get_rank):
        mock_get_rank.return_value = 0
        """Test repair_callback with invalid step"""
        step = 0
        need_rebuild = False
        error_ranks = []
        repair_info = {'repair_type': 'send'}
        train_args = {'train_param': {}}
        params = "{\"scale-out-strategy\": \"DP\"}"
        try:
            repair_callback.repair_callback(step, need_rebuild, error_ranks, repair_info, train_args, params)
        except Exception as e:
            self.assertIn("module 'torch' has no attribute 'npu'", str(e))


class TestSendRankRepair(unittest.TestCase):
    @patch('mindio_ttp.adaptor.tft_optimizer_data_repair.convert_log_tensors_to_args')
    @patch('torch.distributed.send')
    @patch('torch.tensor')
    @patch('torch.ByteTensor')
    @patch('torch.save')
    @patch('taskd.python.adaptor.elastic_training.repair_callback.save_memory_ckpt')
    def test_repair_callback(self, mock_save_ckpt, mock_save, mock_byte_tensor,
                                       mock_tensor, mock_send, mock_convert_log):
        src_ranks = [0]
        dest_ranks = [1]
        rank_list = [0, 1]
        optim_idxs = [0]
        from taskd.python.adaptor.elastic_training.constant import send_or_recieve_rank_repair_args
        args = send_or_recieve_rank_repair_args(src_ranks=src_ranks, dest_ranks=dest_ranks, rank_list=rank_list,
                                                optim_idxs=optim_idxs)
        # test src is not equal rank
        try:
            repair_callback.send_rank_repair(args, None, 1)
        except Exception as e:
            self.assertIn("src rank is not equal to current rank", str(e))
        # test src is equal dest
        dest_ranks = [0]
        args = send_or_recieve_rank_repair_args(src_ranks=src_ranks, dest_ranks=dest_ranks, rank_list=rank_list,
                                                optim_idxs=optim_idxs)
        try:
            repair_callback.send_rank_repair(args, None, 0)
        except Exception as e:
            self.assertIn("src rank is equal to dest rank", str(e))
        # test normal
        mock_save_ckpt.return_value = {}
        dest_ranks = [1]
        args = send_or_recieve_rank_repair_args(src_ranks=src_ranks, dest_ranks=dest_ranks, rank_list=rank_list,
                                                optim_idxs=optim_idxs)
        from megatron.training.utils import test_optimizer
        train_args = [test_optimizer()]
        repair_callback.send_rank_repair(args, train_args, 0)
        assert mock_convert_log.called


class TestRecvRankRepair(unittest.TestCase):
    @patch('mindio_ttp.adaptor.tft_optimizer_data_repair.convert_log_tensors_to_args')
    @patch('torch.distributed.recv')
    @patch('torch.tensor')
    @patch('torch.ByteTensor')
    @patch('torch.load')
    @patch('mindio_ttp.adaptor.tft_optimizer_data_repair.convert_log_tensors_to_args')
    @patch('torch.empty')
    def test_recv_rank_repair(self, mock_empty, mock_save_ckpt, mock_load, mock_byte_tensor,
                                       mock_tensor, mock_recv, mock_convert_log):
        from megatron.training.utils import test_optimizer
        mock_tensor.return_value = test_optimizer()
        mock_empty.return_value = test_optimizer()
        src_ranks = [0]
        dest_ranks = [1]
        rank_list = [0, 1]
        optim_idxs = [0]
        from taskd.python.adaptor.elastic_training.constant import send_or_recieve_rank_repair_args
        args = send_or_recieve_rank_repair_args(src_ranks=src_ranks, dest_ranks=dest_ranks, rank_list=rank_list,
                                                optim_idxs=optim_idxs)
        train_args = [test_optimizer()]
        # test src != rank
        try:
            repair_callback.recv_rank_repair(args, False, train_args, 0)
        except Exception as e:
            self.assertIn("dest rank is not equal current rank", str(e))
        # test src is equal dest
        dest_ranks = [0]
        args = send_or_recieve_rank_repair_args(src_ranks=src_ranks, dest_ranks=dest_ranks, rank_list=rank_list,
                                                optim_idxs=optim_idxs)
        try:
            repair_callback.recv_rank_repair(args, False, train_args, 0)
        except Exception as e:
            self.assertIn("src rank is equal to dest rank", str(e))
        # test load recv to npu
        mock_save_ckpt.return_value = {}
        dest_ranks = [1]
        args = send_or_recieve_rank_repair_args(src_ranks=src_ranks, dest_ranks=dest_ranks, rank_list=rank_list,
                                                optim_idxs=optim_idxs)
        try:
            repair_callback.recv_rank_repair(args, False, train_args, 1)
        except Exception as e:
            self.assertIn("module 'torch' has no attribute 'npu'", str(e))


class TestBuildModelAndOptimizer(unittest.TestCase):
    def test_build_model_and_optimizer(self):
        result = repair_callback.build_model_and_optimizer(None, 1, True)
        self.assertIsNotNone(result)


class TestLoadMemoryCkpt(unittest.TestCase):
    @patch('torch.distributed.get_rank')
    def test_load_memory_ckpt(self, mock_get_rank):
        from megatron.training import get_args
        args_str = 'args'
        state_dict = {
            'iteration': 0,
            args_str: get_args(),
            'num_floating_point_operations_so_far': False,
            'optimizer': None,
            'opt_param_scheduler': None,
        }
        result = repair_callback.load_memory_ckpt(None, get_args(), get_args(),
                                                  state_dict, 0)
        self.assertEqual(result, 0)

    def test_save_memory_ckpt(self):
        from megatron.training import get_args
        args_str = 'args'
        state_dict = {
            'iteration': 0,
            args_str: get_args(),
            'num_floating_point_operations_so_far': False,
            'optimizer': None,
            'opt_param_scheduler': None,
        }
        result = repair_callback.save_memory_ckpt(get_args(), get_args(), 1,
                                                  0, 0)
        self.assertIsNotNone(result)


if __name__ == '__main__':
    unittest.main()