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
import collections
import io
import json
import os
import unittest
from contextlib import redirect_stdout
from unittest.mock import patch, Mock

from mindcluster_tools.eid_generator import EIDGenerator
from mindcluster_tools.tools_parser import main
from mindcluster_tools.utils import parse_eid, product_type_enum
from mindcluster_tools.topo import Topo


class TestRootInfo(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.eid_generator = EIDGenerator()
        # Framework for rootinfo.json, used for validation
        cls.frame = {
            "version": str,
            "rank_list": (
                list,
                {
                    "device_id": int,
                    "local_id": int,
                    "level_list": (
                        list,
                        {
                            "net_layer": int,
                            "net_instance_id": str,
                            "net_type": str,
                            "net_attr": str,
                            "rank_addr_list": (
                                list,
                                {
                                    "addr_type": str,
                                    "addr": str,
                                    "ports": (list, str),
                                    "plane_id": str,
                                },
                            ),
                        },
                    ),
                },
            ),
        }

        # Simulate PEER2NET type topology data items
        cls.mock_data = {
            "edge_list": [
                {
                    "net_layer": 0,
                    "link_type": "PEER2NET",
                    "local_a": 0,
                    "local_a_ports": ["0/1", "0/2", "1/1", "1/2"],
                    "protocols": ["UB_CTP", "UB_MEM"],
                },
                {
                    "net_layer": 0,
                    "link_type": "PEER2NET",
                    "local_a": 1,
                    "local_a_ports": ["0/1", "0/2", "1/1", "1/2"],
                    "protocols": ["UB_CTP", "UB_MEM"],
                },
            ]
        }
        cls.tmp_topo_path = "tmp.json"
        with open(cls.tmp_topo_path, "w") as f:
            json.dump(cls.mock_data, f, indent=2)

    @classmethod
    def tearDownClass(cls):
        if os.path.exists(cls.tmp_topo_path):
            os.remove(cls.tmp_topo_path)

    def test_parse_eid(self):
        """Test parsing EID"""
        sample1 = ["000000000000002000100000df00027f"]
        sample2 = [
            "000000000000002000100000df000282",
            "000000000000002000100000df000283",
        ]
        res1 = [
            {
                "port_id": 127,
                "board_id": 2,
                "chassis_id": 0,
                "fe_id": 1,
                "super_pod_id": 0,
            }
        ]
        res2 = [
            {
                "port_id": 130,
                "board_id": 2,
                "chassis_id": 0,
                "fe_id": 1,
                "super_pod_id": 0,
            },
            {
                "port_id": 131,
                "board_id": 2,
                "chassis_id": 0,
                "fe_id": 1,
                "super_pod_id": 0,
            },
        ]
        self.assertEqual(parse_eid.main(sample1), res1)
        self.assertEqual(parse_eid.main(sample2), res2)

    def test_topo(self):
        """Test topology file query"""
        topo = Topo(self.__class__.tmp_topo_path)
        self.assertEqual(topo.get_ports_by_level_and_die(0, 0, 0), [1, 2])
        self.assertEqual(topo.get_ports_by_level_and_die(1, 0, 1), [1, 2])

    def test_eid_generator(self):
        """Test EID generator"""
        pairs = [
            (
                (2, 0, 7, 1, 1, parse_eid.EID_TYPE_PHY),
                "000000000000002000100000dfdf102c",
            ),
            (
                (5, 1, 5, 0, 1, parse_eid.EID_TYPE_PHY),
                "000000000000000000100000dfdf1069",
            ),
            (
                (31, 1, 4, 1, 1, parse_eid.EID_TYPE_PHY),
                "000000000000002000100000dfdf138c",
            ),
            (
                (7, 0, 2, 0, 1, parse_eid.EID_TYPE_PHY),
                "000000000000000000100000dfdf1081",
            ),
            (
                (3, 0, 1, 0, 1, parse_eid.EID_TYPE_LOGIC),
                "000000000000000000100000dfdf10c1",
            ),
            (
                (5, 1, 2, 1, 1, parse_eid.EID_TYPE_LOGIC),
                "000000000000002000100000dfdf10cc",
            ),
            (
                (18, 0, 1, 1, 1, parse_eid.EID_TYPE_LOGIC),
                "000000000000002000100000dfdf12bd",
            ),
            (
                (7, 1, 2, 0, 1, parse_eid.EID_TYPE_LOGIC),
                "000000000000000000100000dfdf10d4",
            ),
        ]
        for params, target in pairs:
            eid = self.__class__.eid_generator.query(*params)
            self.assertEqual(eid, target)

    def assertIsRootinfoDict(self, rootinfo):
        """Test if rootinfo.json has correct format"""

        def validate_dict_frame(cur_frame, cur_data, work_deque):
            """Validate if cur_data satisfies dictionary type format"""
            #  compare key
            if set(cur_frame.keys()) != set(cur_data.keys()):
                self.fail(
                    self._formatMessage(
                        None,
                        f"the key in frame {set(cur_frame.keys())} does not match {set(cur_data.keys())}",
                    )
                )
            #  compare type of value
            for k, v in cur_frame.items():
                # Composite type queue comparison
                if isinstance(v, tuple):
                    if isinstance(cur_data[k], v[0]):
                        work_deque.append((v[1], cur_data[k]))
                    else:
                        self.fail(
                            self._formatMessage(
                                None,
                                f"the attributr [{k}] does not match type [{v[0]}]",
                            )
                        )
                else:
                    if not isinstance(cur_data[k], v):
                        self.fail(
                            self._formatMessage(
                                None, f"the [{k}] does not match type [{v}]"
                            )
                        )

        def validate_single_type(cur_frame, cur_data):
            """Validate if cur_data satisfies single type format"""
            if not isinstance(cur_data, cur_frame):
                self.fail(
                    self._formatMessage(
                        None,
                        f"the value [{cur_data}] does not match type [{cur_frame}]",
                    )
                )

        work_deque = collections.deque()
        work_deque.append((self.__class__.frame, rootinfo))
        while work_deque:
            cur_frame, cur_datas = work_deque.popleft()
            if not isinstance(cur_datas, list):
                cur_datas = [cur_datas]
            if isinstance(cur_frame, dict):
                for cur_data in cur_datas:
                    validate_dict_frame(cur_frame, cur_data, work_deque)
            else:
                for cur_data in cur_datas:
                    validate_single_type(cur_frame, cur_data)

    @patch("builtins.exit", side_effect=SystemExit)
    @patch("builtins.print")
    def test_version_print(self, mock_print, mock_exit):
        """Test version printing"""
        with self.assertRaises(SystemExit):
            main(["-v"])
        self.assertEqual(mock_print.call_count, 2)
        self.assertEqual(mock_exit.call_count, 1)

    @patch("mindcluster_tools.rootinfo.TopoSingleFactory.get_topo")
    @patch("mindcluster_tools.rootinfo.DCMIQuerier")
    def test_rootinfo(self, mock_dcmi_querier, mock_get_topo):
        """Test integrated rootinfo generation, using EIDGenerator to replace DCMI queries, using fixed [1,2] ports instead of querying topology files"""
        mock_dcmi_querier_instance = Mock()
        mock_dcmi_querier_instance.query.side_effect = self.eid_generator.query
        mock_dcmi_querier.return_value = mock_dcmi_querier_instance
        mock_topo_instance = Mock()
        mock_topo_instance.get_ports_by_level_and_die.return_value = [1, 2]
        mock_get_topo.return_value = mock_topo_instance
        print("\n--------------level 1---------------")
        output_buffer = io.StringIO()
        with redirect_stdout(output_buffer):
            main(
                [
                    "rootinfo",
                    "-t",
                    "superpod_2d.json",
                    "-l",
                    "1",
                    "--super_pod_id",
                    "0",
                    "--chassis_id",
                    "0",
                    "-r",
                    "64",
                ]
            )
            ret = output_buffer.getvalue()
            self.assertIsRootinfoDict(json.loads(ret))
        print("\n--------------level 2---------------")
        output_buffer = io.StringIO()
        with redirect_stdout(output_buffer):
            main(
                [
                    "rootinfo",
                    "-t",
                    "superpod_2d.json",
                    "-l",
                    "2",
                    "--super_pod_id",
                    "0",
                    "--chassis_id",
                    "0",
                    "-r",
                    "64",
                ]
            )
            ret = output_buffer.getvalue()
            self.assertIsRootinfoDict(json.loads(ret))

    @patch("mindcluster_tools.rootinfo.TopoSingleFactory.get_topo")
    @patch("mindcluster_tools.rootinfo.dcmi.get_local_id")
    @patch("mindcluster_tools.rootinfo.dcmi.get_device_board_info")
    @patch("mindcluster_tools.rootinfo.dcmi.get_mainboard_id")
    def test_rootinfo_with_dcmi_when_1d_and_2d(
        self,
        mock_get_mainboard_id,
        mock_get_device_board_info,
        mock_get_local_id,
        mock_get_topo,
    ):
        """Test pod 1D/2D building rootinfo through DCMI queries"""
        mock_topo_instance = Mock()
        mock_topo_instance.get_ports_by_level_and_die = Mock(return_value=[1, 2])
        mock_get_topo.return_value = mock_topo_instance
        card_num = 64
        mock_get_local_id.side_effect = [i for i in range(card_num)]
        mock_get_mainboard_id.return_value = 104
        mock_board_info = Mock()
        mock_board_info.board_id = 40
        mock_get_device_board_info.return_value = mock_board_info
        os.environ["MOCK_SPOD_ID"] = str(0)
        os.environ["MOCK_SPOD_SIZE"] = str(128)
        os.environ["MOCK_CHASSIS_ID"] = str(10)
        print("\n--------------2d---------------")
        os.environ["PRODUCT_TYPE"] = str(product_type_enum.ProductType.POD_2D.value)
        output_buffer = io.StringIO()
        with redirect_stdout(output_buffer):
            main(["rootinfo"])
            ret = output_buffer.getvalue()
            self.assertIsRootinfoDict(json.loads(ret))
        print("\n--------------1d---------------")
        os.environ["PRODUCT_TYPE"] = str(product_type_enum.ProductType.POD_1D.value)
        output_buffer = io.StringIO()
        with redirect_stdout(output_buffer):
            main(["rootinfo"])
            ret = output_buffer.getvalue()
            self.assertIsRootinfoDict(json.loads(ret))
        self.assertEqual(mock_get_mainboard_id.call_count, 2)
        self.assertEqual(mock_get_device_board_info.call_count, 2)

    @patch("mindcluster_tools.rootinfo.TopoSingleFactory.get_topo")
    @patch("mindcluster_tools.rootinfo.dcmi.get_local_id")
    @patch("mindcluster_tools.rootinfo.dcmi.get_device_board_info")
    @patch("mindcluster_tools.rootinfo.dcmi.get_mainboard_id")
    def test_rootinfo_with_dcmi_when_server_8p_and16p(
        self,
        mock_get_mainboard_id,
        mock_get_device_board_info,
        mock_get_local_id,
        mock_get_topo,
    ):
        """Test server 8p/16p building rootinfo through DCMI queries"""
        mock_topo_instance = Mock()
        mock_topo_instance.get_ports_by_level_and_die = Mock(return_value=[1, 2])
        mock_get_topo.return_value = mock_topo_instance
        card_num = 64
        mock_get_local_id.side_effect = [i for i in range(card_num)]
        mock_get_mainboard_id.return_value = 104
        mock_board_info = Mock()
        mock_board_info.board_id = 40
        mock_get_device_board_info.return_value = mock_board_info
        os.environ["MOCK_SPOD_ID"] = str(0)
        os.environ["MOCK_SPOD_SIZE"] = str(128)
        os.environ["MOCK_CHASSIS_ID"] = str(10)
        print("\n--------------server 8p---------------")
        os.environ["PRODUCT_TYPE"] = str(product_type_enum.ProductType.SERVER_8P.value)
        output_buffer = io.StringIO()
        with redirect_stdout(output_buffer):
            main(["rootinfo"])
            ret = output_buffer.getvalue()
            self.assertIsRootinfoDict(json.loads(ret))
        print("\n--------------server 16p---------------")
        os.environ["PRODUCT_TYPE"] = str(product_type_enum.ProductType.SERVER_16P.value)
        output_buffer = io.StringIO()
        with redirect_stdout(output_buffer):
            main(["rootinfo"])
            ret = output_buffer.getvalue()
            self.assertEqual(mock_get_mainboard_id.call_count, 2)
        self.assertEqual(mock_get_device_board_info.call_count, 2)
        self.assertIsRootinfoDict(json.loads(ret))

    @patch("mindcluster_tools.rootinfo.TopoSingleFactory.get_topo")
    @patch("mindcluster_tools.rootinfo.dcmi.get_device_board_info")
    @patch("mindcluster_tools.rootinfo.dcmi.get_mainboard_id")
    def test_rootinfo_with_dcmi_when_standard_1p_and_4p(
        self, mock_get_mainboard_id, mock_get_device_board_info, mock_get_topo
    ):
        """Test standard card 1p/4p building rootinfo through DCMI queries"""
        mock_topo_instance = Mock()
        mock_topo_instance.get_ports_by_level_and_die = Mock(return_value=[1, 2])
        mock_get_topo.return_value = mock_topo_instance
        mock_board_info = Mock()
        mock_board_info.board_id = 27
        mock_get_device_board_info.return_value = mock_board_info
        os.environ["MOCK_SPOD_ID"] = str(-1)
        os.environ["MOCK_SPOD_SIZE"] = str(-1)
        os.environ["MOCK_CHASSIS_ID"] = str(-1)
        print("\n--------------standard 1p---------------")
        os.environ["PRODUCT_TYPE"] = str(
            product_type_enum.ProductType.STANDARD_1P.value
        )
        mock_get_mainboard_id.return_value = 104
        output_buffer = io.StringIO()
        with redirect_stdout(output_buffer):
            main(["rootinfo"])
            ret = output_buffer.getvalue()
            self.assertIsRootinfoDict(json.loads(ret))
        print("\n--------------standard 4p---------------")
        os.environ["PRODUCT_TYPE"] = str(
            product_type_enum.ProductType.STANDARD_4P.value
        )
        mock_get_mainboard_id.return_value = 108
        output_buffer = io.StringIO()
        with redirect_stdout(output_buffer):
            main(["rootinfo"])
            ret = output_buffer.getvalue()
            self.assertEqual(mock_get_mainboard_id.call_count, 2)
        self.assertEqual(mock_get_device_board_info.call_count, 2)
        self.assertIsRootinfoDict(json.loads(ret))
