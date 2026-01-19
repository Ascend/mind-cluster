#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# Copyright (c) 2025, Huawei Technologies Co., Ltd. All rights reserved.
import unittest
from unittest.mock import patch, MagicMock, call
import sys
import json

from taskd.python.adaptor.elastic_training import common


class TestCommon(unittest.TestCase):

    def setUp(self):
        # Reset global variables before each test
        common.ORIGIN_DP_SIZE = None
        common.ORIGIN_NUM_MICRO_BATCHES = None
        common.SCALE_IN_WORLD_GROUP = None
        common.SCALE_IN_DP_CP_REPLICA_GROUP = None
        common.SCALE_IN_DP_CP_REPLICA_GROUP_GLOO = None
        common.FAULT_RANK_IN_DP_CP_REPLICA_GROUP = False
        common.IS_FAULT_REPLICA_RANK = False
        common.FAULT_REPLICA_RANK = None
        common.SCALE_IN_RUNNING_STATE = False
        common.HAS_DATA = None

    def test_update_scale_in_flag(self):
        """Test updating scale in flag"""
        common.update_scale_in_flag(True)
        self.assertTrue(common.SCALE_IN_RUNNING_STATE)

        common.update_scale_in_flag(False)
        self.assertFalse(common.SCALE_IN_RUNNING_STATE)

    def test_zit_get_has_data_index(self):
        """Test getting has data index"""
        common.HAS_DATA = 5
        self.assertEqual(common.zit_get_has_data_index(), 5)

        common.HAS_DATA = None
        self.assertIsNone(common.zit_get_has_data_index())

    def test_zit_get_scale_in_world_group(self):
        """Test getting scale in world group"""
        common.SCALE_IN_WORLD_GROUP = "test_group"
        self.assertEqual(common.zit_get_scale_in_world_group(), "test_group")

        common.SCALE_IN_WORLD_GROUP = None
        self.assertIsNone(common.zit_get_scale_in_world_group())

    def test_zit_is_fault_replica_rank(self):
        """Test checking if is fault replica rank"""
        common.IS_FAULT_REPLICA_RANK = True
        self.assertTrue(common.zit_is_fault_replica_rank())

        common.IS_FAULT_REPLICA_RANK = False
        self.assertFalse(common.zit_is_fault_replica_rank())

    def test_zit_fault_rank_in_dp_cp_replica_group(self):
        """Test checking if fault rank in dp cp replica group"""
        common.FAULT_RANK_IN_DP_CP_REPLICA_GROUP = True
        self.assertTrue(common.zit_fault_rank_in_dp_cp_replica_group())

        common.FAULT_RANK_IN_DP_CP_REPLICA_GROUP = False
        self.assertFalse(common.zit_fault_rank_in_dp_cp_replica_group())

    def test_zit_scale_in_running_state(self):
        """Test checking scale in running state"""
        common.SCALE_IN_RUNNING_STATE = True
        self.assertTrue(common.zit_scale_in_running_state())

        common.SCALE_IN_RUNNING_STATE = False
        self.assertFalse(common.zit_scale_in_running_state())

    def test_zit_get_scale_in_dp_cp_replica_group(self):
        """Test getting scale in dp cp replica group"""
        common.SCALE_IN_DP_CP_REPLICA_GROUP = "test_dp_cp_group"
        self.assertEqual(common.zit_get_scale_in_dp_cp_replica_group(), "test_dp_cp_group")

        common.SCALE_IN_DP_CP_REPLICA_GROUP = None
        self.assertIsNone(common.zit_get_scale_in_dp_cp_replica_group())

    def test_zit_get_scale_in_dp_cp_replica_group_gloo(self):
        """Test getting scale in dp cp replica group gloo"""
        common.SCALE_IN_DP_CP_REPLICA_GROUP_GLOO = "test_gloo_group"
        self.assertEqual(common.zit_get_scale_in_dp_cp_replica_group_gloo(), "test_gloo_group")

        common.SCALE_IN_DP_CP_REPLICA_GROUP_GLOO = None
        self.assertIsNone(common.zit_get_scale_in_dp_cp_replica_group_gloo())

    def test_zit_get_fault_replica_rank(self):
        """Test getting fault replica rank"""
        common.FAULT_REPLICA_RANK = 3
        self.assertEqual(common.zit_get_fault_replica_rank(), 3)

        common.FAULT_REPLICA_RANK = None
        self.assertIsNone(common.zit_get_fault_replica_rank())

    @patch('taskd.python.adaptor.elastic_training.common.ttp_logger')
    def test_check_scale_out_params_valid(self, mock_logger):
        """Test checking valid scale out params"""
        params = '{"scale-out-strategy": "DP"}'

        # Should not raise exception
        try:
            common.check_scale_out_params(params)
        except Exception as e:
            self.fail(f"check_scale_out_params raised exception: {e}")

        mock_logger.LOGGER.info.assert_called_once()

    @patch('taskd.python.adaptor.elastic_training.common.ttp_logger')
    def test_check_scale_out_params_invalid(self, mock_logger):
        """Test checking invalid scale out params"""
        params = '{"scale-out-strategy": "INVALID"}'

        with self.assertRaises(Exception) as context:
            common.check_scale_out_params(params)

        self.assertIn("Only support DP strategy", str(context.exception))
        mock_logger.LOGGER.info.assert_called_once()

    @patch('taskd.python.adaptor.elastic_training.common.ttp_logger')
    def test_check_scale_out_params_empty(self, mock_logger):
        """Test checking empty scale out params"""
        params = '{}'

        with self.assertRaises(Exception) as context:
            common.check_scale_out_params(params)

        self.assertIn("Only support DP strategy", str(context.exception))
        mock_logger.LOGGER.info.assert_called_once()

    @patch('taskd.python.adaptor.elastic_training.common.ttp_logger')
    def test_check_scale_out_params_invalid_json(self, mock_logger):
        """Test checking invalid JSON scale out params"""
        params = 'invalid json'

        with self.assertRaises(json.JSONDecodeError):
            common.check_scale_out_params(params)

    @patch('taskd.python.adaptor.elastic_training.common.build_data_iterator')
    @patch('taskd.python.adaptor.elastic_training.common.utils')
    def test_build_dataset(self, mock_utils, mock_build_data_iterator):
        """Test building dataset"""
        mock_train = MagicMock()
        mock_valid = MagicMock()
        mock_test = MagicMock()
        mock_build_data_iterator.return_value = (mock_train, mock_valid, mock_test)
        train_param = 'train_param'
        model_index = 'model_index'
        train_data_index = 'train_data_index'
        valid_data_index = 'valid_data_index'
        test_data_iter = 'test_data_iter'
        mock_utils.TRAIN_PARAM = train_param
        mock_utils.MODEL_INDEX = model_index
        mock_utils.TRAIN_DATA_INDEX = train_data_index
        mock_utils.VALID_DATA_INDEX = valid_data_index
        mock_utils.TEST_DATA_ITER = test_data_iter

        args = {
            train_param: {
                model_index: MagicMock()
            },
            test_data_iter: [None]
        }

        common.build_dataset(args)

        self.assertEqual(args[train_param][train_data_index], mock_train)
        self.assertEqual(args[train_param][valid_data_index], mock_valid)
        self.assertEqual(args[test_data_iter][0], mock_test)
        mock_build_data_iterator.assert_called_once_with(args[train_param][model_index])

    @patch('taskd.python.adaptor.elastic_training.common.get_args')
    @patch('taskd.python.adaptor.elastic_training.common.tft_optimizer_data_repair.get_build_data_args')
    @patch('taskd.python.adaptor.elastic_training.common.mpu.set_virtual_pipeline_model_parallel_rank')
    @patch('taskd.python.adaptor.elastic_training.common.build_train_valid_test_data_iterators')
    def test_build_data_iterator_with_virtual_pipeline(self, mock_build_iterators, mock_set_rank,
                                                       mock_get_build_args, mock_get_args):
        """Test building data iterator with virtual pipeline"""
        mock_args = MagicMock()
        mock_args.virtual_pipeline_model_parallel_size = 2
        mock_get_args.return_value = mock_args

        mock_get_build_args.return_value = (None, None, "datasets_provider")

        mock_iterators = [MagicMock(), MagicMock(), MagicMock()]
        mock_build_iterators.return_value = mock_iterators

        model = [MagicMock(), MagicMock()]  # Two models for virtual pipeline

        train, valid, test = common.build_data_iterator(model)

        self.assertEqual(len(train), 2)
        self.assertEqual(len(valid), 2)
        self.assertEqual(len(test), 2)
        mock_set_rank.assert_has_calls([call(0), call(1)])
        mock_build_iterators.assert_has_calls([call("datasets_provider"), call("datasets_provider")])

    @patch('taskd.python.adaptor.elastic_training.common.get_args')
    @patch('taskd.python.adaptor.elastic_training.common.tft_optimizer_data_repair.get_build_data_args')
    @patch('taskd.python.adaptor.elastic_training.common.build_train_valid_test_data_iterators')
    def test_build_data_iterator_without_virtual_pipeline(self, mock_build_iterators,
                                                          mock_get_build_args, mock_get_args):
        """Test building data iterator without virtual pipeline"""
        mock_args = MagicMock()
        mock_args.virtual_pipeline_model_parallel_size = None
        mock_get_args.return_value = mock_args

        mock_get_build_args.return_value = (None, None, "datasets_provider")
        mock_build_iterators.return_value = ("train", "valid", "test")

        model = MagicMock()

        train, valid, test = common.build_data_iterator(model)

        self.assertEqual(train, "train")
        self.assertEqual(valid, "valid")
        self.assertEqual(test, "test")
        mock_build_iterators.assert_called_once_with("datasets_provider")

    @patch('taskd.python.adaptor.elastic_training.common.get_args')
    @patch('taskd.python.adaptor.elastic_training.common.tft_optimizer_data_repair.get_build_data_args')
    @patch('taskd.python.adaptor.elastic_training.common.build_train_valid_test_data_iterators')
    def test_build_data_iterator_empty_model(self, mock_build_iterators,
                                             mock_get_build_args, mock_get_args):
        """Test building data iterator with empty model"""
        mock_args = MagicMock()
        mock_args.virtual_pipeline_model_parallel_size = 2
        mock_get_args.return_value = mock_args

        mock_get_build_args.return_value = (None, None, "datasets_provider")
        mock_build_iterators.return_value = (MagicMock(), MagicMock(), MagicMock())

        model = []  # Empty model list

        train, valid, test = common.build_data_iterator(model)

        self.assertEqual(train, [])
        self.assertEqual(valid, [])
        self.assertEqual(test, [])


if __name__ == '__main__':
    unittest.main()