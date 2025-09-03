#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# Copyright (c) 2025, Huawei Technologies Co., Ltd. All rights reserved.
import unittest

from taskd.python.adaptor.elastic_training import constant


class TestConstant(unittest.TestCase):

    def test_send_or_recieve_rank_repair_args_class(self):
        """Test the send_or_recieve_rank_repair_args class structure"""
        self.assertTrue(hasattr(constant, 'send_or_recieve_rank_repair_args'))

        args = constant.send_or_recieve_rank_repair_args()
        args.src_ranks = None
        args.dest_ranks = None
        args.optim_idxs = None
        args.step = None
        args.rank_list = None
        
        # 测试实例属性是否存在
        self.assertTrue(hasattr(args, 'src_ranks'))
        self.assertTrue(hasattr(args, 'dest_ranks'))
        self.assertTrue(hasattr(args, 'optim_idxs'))
        self.assertTrue(hasattr(args, 'step'))
        self.assertTrue(hasattr(args, 'rank_list'))
        
    def test_send_or_recieve_rank_repair_args_initialization(self):
        """Test initializing send_or_recieve_rank_repair_args with values"""
        args = constant.send_or_recieve_rank_repair_args()
        args.src_ranks = [0, 1]
        args.dest_ranks = [2, 3]
        args.optim_idxs = [0, 1]
        args.step = 100
        args.rank_list = [0, 1, 2, 3]

        self.assertEqual(args.src_ranks, [0, 1])
        self.assertEqual(args.dest_ranks, [2, 3])
        self.assertEqual(args.optim_idxs, [0, 1])
        self.assertEqual(args.step, 100)
        self.assertEqual(args.rank_list, [0, 1, 2, 3])
        
    def test_send_or_recieve_rank_repair_args_empty_initialization(self):
        """Test initializing send_or_recieve_rank_repair_args without values"""
        args = constant.send_or_recieve_rank_repair_args()

        self.assertIsNone(args.src_ranks)
        self.assertIsNone(args.dest_ranks)
        self.assertIsNone(args.optim_idxs)
        self.assertIsNone(args.step)
        self.assertIsNone(args.rank_list)
        

if __name__ == '__main__':
    unittest.main()