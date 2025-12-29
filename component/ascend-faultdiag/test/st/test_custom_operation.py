#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# Copyright 2025 Huawei Technologies Co., Ltd
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
# ==============================================================================
import os
import sys
import logging
import unittest
from unittest.mock import patch
from io import StringIO
from collections import namedtuple

from ascend_fd.pkg.customize.custom_entity import update_entity, show_entity, delete_entity, check_entity
from ascend_fd.pkg.parse.blacklist import blacklist_op, start_blacklist_job
from ascend_fd.utils.status import ParamError
from ascend_fd.utils.tool import safe_read_json as safe_read_json_from_tool
from ascend_fd.utils.tool import DEFAULT_BLACKLIST_CONF
from ascend_fd.configuration.config import DEFAULT_USER_CONF, HOME_PATH

TEST_DIR = os.path.dirname(os.path.dirname(os.path.realpath(__file__)))
BASE_DATA_PATH = os.path.join(TEST_DIR, "custom_operation", "data.json")


class TestUpdateEntity(unittest.TestCase):
    """
    Need deal update -> show -> delete
    """

    @patch('sys.stdout', new_callable=StringIO)
    def setUp(self, mock_string_io) -> None:
        os.makedirs(HOME_PATH, 0o700, exist_ok=True)
        echo_handler = logging.StreamHandler(sys.stdout)
        echo_handler.setFormatter(logging.Formatter('%(message)s'))
        echo_logger = logging.getLogger("ECHO")
        echo_logger.addHandler(echo_handler)
        echo_logger.setLevel(logging.INFO)
        self.mock_string_io = mock_string_io
        self.blacklist_manager = blacklist_op.BlackListManager()

    def test_a_update(self):
        update_entity(BASE_DATA_PATH)
        user_conf = safe_read_json_from_tool(DEFAULT_USER_CONF) if os.path.exists(DEFAULT_USER_CONF) else {}
        self.assertIn("Test_Code_1", user_conf.get("knowledge-repository", {}))
        self.assertIn("Test_Code_2", user_conf.get("knowledge-repository", {}))
        self.assertNotIn("Error_Code_3", user_conf.get("knowledge-repository", {}))
        self.assertIn("Fault codes ['Error_Code_3'] fail to verify parameters", self.mock_string_io.getvalue())

    def test_b_show(self):
        show_entity(["Test_Code_2", "Error_Code_3"], [])
        self.assertIn("Test_Code_2", self.mock_string_io.getvalue())
        self.assertIn("test_description_zh2", self.mock_string_io.getvalue())
        self.assertIn("Not exist in user-defined fault entity set.", self.mock_string_io.getvalue())

    def test_c_delete(self):
        delete_entity(["Test_Code_1", "Error_Code_3"], True)
        user_conf = safe_read_json_from_tool(DEFAULT_USER_CONF) if os.path.exists(DEFAULT_USER_CONF) else {}
        self.assertNotIn("Test_Code_1", user_conf.get("knowledge-repository", {}))
        self.assertIn("Test_Code_2", user_conf.get("knowledge-repository", {}))
        self.assertIn("Fault codes ['Error_Code_3'] does not exist in user-defined fault entity set",
                      self.mock_string_io.getvalue())

    def test_blacklist_add(self):
        new_blacklist = ["ERROR", "FAULT"]
        self.blacklist_manager.add_blacklist(new_blacklist)
        user_blacklist = safe_read_json_from_tool(DEFAULT_BLACKLIST_CONF) if os.path.exists(
            DEFAULT_BLACKLIST_CONF) else {}
        blacklist_key = "blacklist"
        self.assertIn("ERROR", user_blacklist.get(blacklist_key, [])[0])
        self.assertIn("FAULT", user_blacklist.get(blacklist_key, [])[0])
        self.assertNotIn("Error_Code_3", user_blacklist.get(blacklist_key, [])[0])

    def test_blacklist_show(self):
        self.blacklist_manager.show_blacklist()
        self.assertIn("BLACKLIST", self.mock_string_io.getvalue())

    def test_blacklist_file(self):
        self.blacklist_manager.switch_custom_file(DEFAULT_BLACKLIST_CONF)

    def test_blacklist_delete(self):
        delete_ids = [0]
        self.blacklist_manager.delete_blacklist_keywords_by_key(delete_ids)
        user_blacklist = safe_read_json_from_tool(DEFAULT_BLACKLIST_CONF) if os.path.exists(
            DEFAULT_BLACKLIST_CONF) else {}
        self.assertNotIn("ERROR", user_blacklist.get("blacklist", {}))
        self.assertNotIn("FAULT", user_blacklist.get("blacklist", {}))

    def test_blacklist_ops1(self):
        Args = namedtuple('args', ['force', 'add', 'delete', 'show', 'file'])
        args = Args(force=True, add=[], delete=None, show=None, file=None)
        try:
            start_blacklist_job(args)
        except ParamError:
            pass

    def test_blacklist_ops2(self):
        Args = namedtuple('args', ['force', 'add', 'delete', 'show', 'file'])
        args = Args(force=False, add=["ERROR", "TEST_ERROR"], delete=None, show=None, file=None)
        start_blacklist_job(args)
        self.assertIn("add blacklist success", self.mock_string_io.getvalue())

    def test_blacklist_ops3(self):
        Args = namedtuple('args', ['force', 'add', 'delete', 'show', 'file'])
        args = Args(force=False, add=None, delete=None, show=True, file=None)
        start_blacklist_job(args)

    def test_blacklist_ops4(self):
        Args = namedtuple('args', ['force', 'add', 'delete', 'show', 'file'])
        args = Args(force=True, add=None, delete=None, show=None, file=DEFAULT_BLACKLIST_CONF)
        start_blacklist_job(args)

    def test_blacklist_ops5(self):
        Args = namedtuple('args', ['force', 'add', 'delete', 'show', 'file'])
        args = Args(force=True, add=None, delete=[0], show=None, file=None)
        start_blacklist_job(args)

    def test_func_check_entity_pass(self):
        self.assertTrue(
            check_entity(os.path.join(TEST_DIR, "custom_operation", "test_valid_custom_ascend_kg_config.json")))
