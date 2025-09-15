#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# Copyright (c) 2025, Huawei Technologies Co., Ltd. All rights reserved.
import unittest
from unittest.mock import patch

from taskd.python.adaptor.elastic_training import common
from taskd.python.adaptor.elastic_training.optimizer import TTPElasticTrainingReplicaOptimizer


class TestTTPElasticTrainingReplicaOptimizer(unittest.TestCase):
    def test_save_parameter_state(self):
        file_name = "file_name"
        optimizer = TTPElasticTrainingReplicaOptimizer(True)
        optimizer.save_parameter_state(file_name)
        self.assertEqual(file_name, optimizer.filename)

    @patch('mindio_ttp.adaptor.utils.FileUtils.regular_file_path')
    @patch('torch.distributed.get_rank')
    @patch('torch.save')
    def test_save_parameter_state_scale_in_running(self, mock_save, mock_get_rank, mock_regular_file_path):
        mock_regular_file_path.return_value = False, "", 'regular_file_path'
        mock_get_rank.return_value = 0
        file_name = "file_name"
        common.SCALE_IN_RUNNING_STATE = True
        try:
            TTPElasticTrainingReplicaOptimizer(False).save_parameter_state(file_name)
        except Exception as e:
            self.assertIn("rank 0: save parameter filename is not valid, error:", str(e))
        mock_regular_file_path.return_value = True, "", 'regular_file_path'
        TTPElasticTrainingReplicaOptimizer(False).save_parameter_state(file_name)
        assert mock_save.called
        TTPElasticTrainingReplicaOptimizer(False).save_parameter_state(file_name)
        self.assertEqual(5, mock_get_rank.call_count)
        assert mock_save.called


    def test_set_dump_args(self):
        common.update_scale_in_flag(True)
        TTPElasticTrainingReplicaOptimizer(True).set_dump_args(0, 1, [0, 4])
        self.assertEqual(False, common.zit_scale_in_running_state())


if __name__ == '__main__':
    unittest.main()