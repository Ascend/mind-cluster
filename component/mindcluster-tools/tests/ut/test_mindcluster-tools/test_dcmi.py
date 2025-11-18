#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# Copyright 2025. Huawei Technologies Co.,Ltd. All rights reserved.
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
import os
import unittest
from mindcluster_tools.dcmi import dcmi
from mindcluster_tools.utils.product_type_enum import ProductType

class TestDCMI(unittest.TestCase):

    def test_init(self):
        ret = dcmi.dcmi_init()
        self.assertEqual(ret, 0)

    def test_get_all_device_count(self):
        cnt = dcmi.get_all_device_count()
        self.assertEqual(cnt, 8)

    def test_should_raise_error_when_param_invalid(self):
        with self.assertRaises(dcmi.DcmiReturnValueError):
            os.environ["PRODUCT_TYPE"] = str(ProductType.POD_1D.value)
            dcmi.get_eid_list_by_urma_dev_index(-1, 0, 0)
        with self.assertRaises(dcmi.DcmiReturnValueError):
            os.environ["PRODUCT_TYPE"] = str(ProductType.POD_2D.value)
            dcmi.get_eid_list_by_urma_dev_index(0, 1, 0)
        with self.assertRaises(dcmi.DcmiReturnValueError):
            dcmi.get_urma_device_cnt(-1, 0)
        with self.assertRaises(dcmi.DcmiReturnValueError):
            dcmi.get_device_id_in_card(-1)

    def test_get_1d_eid_list(self):
        os.environ["PRODUCT_TYPE"] = str(ProductType.POD_1D.value)
        eid_list = dcmi.get_eid_list_by_urma_dev_index(0, 0, 0)
        self.assertEqual(len(eid_list), 11)

    def test_get_2d_eid_list(self):
        os.environ["PRODUCT_TYPE"] = str(ProductType.POD_2D.value)
        eid_list = dcmi.get_eid_list_by_urma_dev_index(0, 0, 0)
        for eid in eid_list:
            print(f"eid=[{eid}]")
        self.assertEqual(len(eid_list), 20)

    def test_get_urma_device_cnt(self):
        cnt = dcmi.get_urma_device_cnt(0, 0)
        self.assertEqual(cnt, 4)

    def test_get_device_id_max(self):
        device_id_max, mcu_id, cpu_id = dcmi.get_device_id_in_card(0)
        self.assertEqual(device_id_max, 1)
        self.assertEqual(mcu_id, 0)
        self.assertEqual(cpu_id, 0)

    def test_dcmi_get_card_list(self):
        expected_card_num = 8
        expected_card_list = [i for i in range(expected_card_num)]
        card_num, card_list = dcmi.dcmi_get_card_list()
        self.assertEqual(card_num, 8)
        self.assertEqual(card_list, expected_card_list)

    def test_get_super_pod_info(self):
        expected_spod_id = 8
        expected_spod_size = 512
        expected_chassis_id = 10
        os.environ["PRODUCT_TYPE"] = str(ProductType.POD_2D.value)
        os.environ["MOCK_SPOD_ID"] = str(expected_spod_id)
        os.environ["MOCK_SPOD_SIZE"] = str(expected_spod_size)
        os.environ["MOCK_CHASSIS_ID"] = str(expected_chassis_id)
        spod_info = dcmi.get_super_pod_info()
        super_pod_type = int.from_bytes(spod_info.super_pod_type, byteorder='big')
        super_pod_id = spod_info.super_pod_id
        super_pod_size = spod_info.super_pod_size
        chassis_id = spod_info.chassis_id
        self.assertEqual(super_pod_type, int(ProductType.POD_2D.value))
        self.assertEqual(super_pod_id, expected_spod_id)
        self.assertEqual(super_pod_size, expected_spod_size)
        self.assertEqual(chassis_id, expected_chassis_id)
