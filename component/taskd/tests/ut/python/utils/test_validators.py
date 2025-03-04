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
import tempfile
import unittest

from taskd.python.utils.validator import *

BIN_PATH = "/usr/bin"
DIRECTORY_BLACKLIST_PATH = "/abc/d/e"


class TestValidators(unittest.TestCase):
    """
    Test validator functions
    """

    def setUp(self):
        super().setUp()

    def tearDown(self):
        super().tearDown()

    def test_validator_should_return_default_if_invalid(self):
        validation = Validator("aa")
        validation.register_checker(lambda x: len(x) < 5, "length of string should be less than 5")
        self.assertTrue(validation.is_valid())
        validation = Validator("123456")
        validation.register_checker(lambda x: len(x) < 5, "length of string should be less than 5")
        self.assertFalse(validation.is_valid())
        self.assertEqual(validation.get_value("DEFAULT"), "DEFAULT")

    def test_string_validator_max_len_parameter(self):
        self.assertFalse(StringValidator("aa.1245", max_len=3).check_string_length().check().is_valid())
        self.assertTrue(StringValidator("aa.1245", max_len=30).check_string_length().check().is_valid())
        # default infinity
        self.assertTrue(StringValidator("aa.1234564646546").check_string_length().check().is_valid())

    def test_string_validator_min_len_parameter(self):
        self.assertFalse(StringValidator("aa", min_len=3).check_string_length().check().is_valid())
        self.assertTrue(StringValidator("aaa", min_len=3).check_string_length().check().is_valid())
        # default infinity
        self.assertTrue(StringValidator("a").check_string_length().check().is_valid())

    def test_string_validator_can_be_transformed2int(self):
        self.assertFalse(StringValidator("a").can_be_transformed2int().check().is_valid())
        self.assertFalse(StringValidator("9" * 20).can_be_transformed2int().check().is_valid())
        self.assertFalse(StringValidator("1,2").can_be_transformed2int().check().is_valid())
        self.assertTrue(StringValidator("12").can_be_transformed2int().check().is_valid())
        self.assertFalse(StringValidator("12").can_be_transformed2int(min_value=100, max_value=200).check().is_valid())

    def test_string_validator_contain_sensitive_words(self):
        self.assertFalse(StringValidator("passwordme").check_not_contain_black_element("pass")
                         .check_string_length().check().is_valid())

    def test_map_validator_should_contain_inclusive_keys(self):
        map_validator = MapValidator({"a": True, "b": {"c": "1234"}}, inclusive_keys=["a", "b"])
        self.assertTrue(map_validator.is_valid())

    def test_directory_black_list(self):
        self.assertFalse(DirectoryValidator(DIRECTORY_BLACKLIST_PATH).with_blacklist(
            lst=[DIRECTORY_BLACKLIST_PATH]).check().is_valid())
        self.assertTrue(DirectoryValidator(DIRECTORY_BLACKLIST_PATH).with_blacklist(
            lst=[""]).check().is_valid())
        self.assertTrue(DirectoryValidator(DIRECTORY_BLACKLIST_PATH).with_blacklist(["/abc/d/"]).check().is_valid())
        self.assertTrue(DirectoryValidator(DIRECTORY_BLACKLIST_PATH).with_blacklist(["/abc/d/"], exact_compare=True)
                        .check().is_valid())
        # if not exact compare, the /abc/d/e is chirldren path of /abc/d/, so it is invalid
        self.assertFalse(DirectoryValidator(DIRECTORY_BLACKLIST_PATH)
                         .with_blacklist(["/abc/d/"], exact_compare=False).check().is_valid())
        self.assertTrue(DirectoryValidator("/usr/bin/bash").with_blacklist().check().is_valid())
        self.assertFalse(DirectoryValidator("/usr/bin/bash").with_blacklist(exact_compare=False).check().is_valid())

    def test_remove_prefix(self):
        self.assertEqual(DirectoryValidator.remove_prefix(BIN_PATH, None)[1], BIN_PATH)
        self.assertEqual(DirectoryValidator.remove_prefix(BIN_PATH, "")[1], BIN_PATH)
        self.assertIsNone(DirectoryValidator.remove_prefix(None, "abc")[1])
        self.assertEqual(DirectoryValidator.remove_prefix("/usr/bin/python", BIN_PATH)[1], "/python")

    def test_directory_white_list(self):
        self.assertTrue(DirectoryValidator.check_is_children_path("/abc/d", DIRECTORY_BLACKLIST_PATH))
        self.assertTrue(DirectoryValidator.check_is_children_path("/abc/d", "/abc/d"))
        self.assertFalse(DirectoryValidator.check_is_children_path("/abc/d", "/abc/de"))
        self.assertTrue(DirectoryValidator.check_is_children_path("/usr/bin", "/usr/bin/bash"))

    def test_directory_soft_link(self):
        tmp = tempfile.NamedTemporaryFile(delete=True)
        temp_dir = tempfile.mkdtemp()
        path = os.path.join(temp_dir, "link.ink")
        # make a soft link
        os.symlink(tmp.name, path)

        try:
            # do stuff with temp
            tmp.write(b"stuff")
            self.assertFalse(DirectoryValidator(path).check_not_soft_link().check().is_valid())
        finally:
            tmp.close()
            os.remove(path)
            os.removedirs(temp_dir)

    def test_directory_check(self):
        self.assertFalse(DirectoryValidator("a/b/.././c/a.txt").check_is_not_none().check_dir_name().check().is_valid())
        self.assertFalse(DirectoryValidator("").check_is_not_none().check_dir_name().check().is_valid())
        self.assertFalse(DirectoryValidator(None).check_is_not_none().check_dir_name().check().is_valid())
        self.assertTrue(DirectoryValidator("a/bc/d").check_is_not_none().check_dir_name().check().is_valid())
        self.assertTrue(DirectoryValidator("/user/restore/fault/config", max_len=255).
                        check_is_not_none().check_dir_name().
                        path_should_exist(is_file=True, msg="can not find the fault ranks config file")
                        .should_not_contains_sensitive_words().with_blacklist().check())
        self.assertTrue(DirectoryValidator(os.path.dirname(__file__), max_len=255)
                        .check_is_not_none().check_dir_name().check_dir_file_number()
                        .path_should_exist(is_file=False, msg="can not find the fault ranks config file")
                        .should_not_contains_sensitive_words().with_blacklist().check())
        self.assertFalse(DirectoryValidator(os.path.dirname(__file__)).path_should_not_exist().check().is_valid())

    def test_rank_size_check(self):
        self.assertFalse(RankSizeValidator(4096).check_rank_size_valid().check().is_valid())
        self.assertFalse(RankSizeValidator(0).check_device_num_valid().check().is_valid())
        self.assertTrue(RankSizeValidator(1).check_rank_size_valid().check().is_valid())

    def test_file_check(self):
        file_path = os.path.join(os.path.dirname(__file__), "test_data", "test.txt")
        print("file path is", file_path)
        self.assertFalse(FileValidator(file_path).check_file_size().check().is_valid())
        self.assertTrue(FileValidator(file_path).check_not_soft_link().check().is_valid())
        os.chown(file_path, os.getuid(), os.getgid())
        self.assertTrue(FileValidator(file_path).check_user_group().check().is_valid())

    def test_int_check(self):
        self.assertTrue(IntValidator(1, min_value=0, max_value=12).check_value().check().is_valid())

    def test_class_check(self):
        self.assertTrue(ClassValidator(2, int).check_isinstance().is_valid())


if __name__ == "__main__":
    unittest.main()
