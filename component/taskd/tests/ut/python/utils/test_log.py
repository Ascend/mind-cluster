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
import datetime
import os
import tempfile
import time
import unittest
import re
from unittest.mock import patch

from taskd.python.constants.constants import LOG_BACKUP_FORMAT, LOG_BACKUP_PATTERN
from taskd.python.utils.log.logger import CustomRotatingHandler


class TestLogger(unittest.TestCase):
    def test_logger_init(self):
        from taskd.python.utils.log import run_log
        assert run_log is not None


class TestCustomRotationHandler(unittest.TestCase):
    def setUp(self):
        # create a temporary dir
        self.temp_dir = tempfile.TemporaryDirectory()
        self.test_dir = self.temp_dir.name
        self.base_filename = os.path.join(self.test_dir, "customlog.log")

    def tearDown(self):
        # clear the temporary dir
        self.temp_dir.cleanup()

    def create_handler(self, max_bytes=0, backup_count=0, delay=False):
        """
        create CustomRotationHandler instance
        """

        handler = CustomRotatingHandler(
            filename=self.base_filename,
            maxBytes=max_bytes,
            backupCount=backup_count,
            delay=delay,
        )
        return handler

    def test_rotation_filename_format(self):
        """
        test backup file name format
        """
        fixed_time = datetime.datetime(2023, 10, 5, 12, 34, 56)
        with patch("datetime.datetime") as mock_datetime:
            mock_datetime.now.return_value = fixed_time
            handler = self.create_handler()
            rotated_name = handler.rotation_filename("unused")
            base = os.path.splitext(self.base_filename)[0]
            expected_name = f"{base}-2023-10-05T12-34-56.000.log"
            self.assertEqual(rotated_name, expected_name)

    def test_do_rollover_creates_backup(self):
        """
        test backup file creation
        """
        handler = self.create_handler(backup_count=2)
        with open(self.base_filename, "w") as f:
            f.write("test log")
        handler.doRollover()
        backups = [f for f in os.listdir(self.test_dir) if f != "customlog.log"]
        self.assertEqual(len(backups), 1)
        self.assertRegex(backups[0], re.compile(LOG_BACKUP_PATTERN))

    def test_backup_count_enforcement(self):
        """
        test backup count enforcement
        """
        handler = self.create_handler(backup_count=2)
        old_backups = [
            "customlog-2023-10-05T12-00-00.000.log",
            "customlog-2023-10-05T12-10-00.000.log",
            "customlog-2023-10-05T12-20-00.000.log"
        ]
        for fname in old_backups:
            with open(os.path.join(self.test_dir, fname), "w") as f:
                f.write("old log")
        with open(self.base_filename, "w") as f:
            f.write("new log")
        handler.doRollover()
        remaining = sorted(f for f in os.listdir(self.test_dir) if f != "customlog.log")
        self.assertEqual(len(remaining), 2)
        self.assertIn("customlog-2023-10-05T12-20-00.000.log", remaining)

    def test_no_backup_deletion_when_count_zero(self):
        """
        test backup deletion when backup count is zero
        """
        handler = self.create_handler(backup_count=0)
        with open(self.base_filename, "w") as f:
            f.write("test log")
        handler.doRollover()
        # sleep to ensure that file creation timestamps are not duplicated
        time.sleep(0.1)
        with open(self.base_filename, "w") as f:
            f.write("test2 log")
        handler.doRollover()
        # sleep to ensure that file creation timestamps are not duplicated
        time.sleep(0.1)
        with open(self.base_filename, "w") as f:
            f.write("test3 log")
        handler.doRollover()
        backups = [f for f in os.listdir(self.test_dir) if f != "customlog.log"]
        self.assertEqual(len(backups), 3)

    def test_multiple_rollovers(self):
        """
        test after multiple rollovers, backup files are deleted in chronological order
        """
        handler = self.create_handler(backup_count=2)
        old_backups = [
            "customlog-2023-10-05T12-00-00.000.log",
            "customlog-2023-10-05T12-10-00.000.log",
            "customlog-2023-10-05T12-20-00.000.log"
        ]
        for fname in old_backups:
            with open(os.path.join(self.test_dir, fname), "w") as f:
                f.write("old log")
        for _ in range(4):
            with open(self.base_filename, "w") as f:
                f.write("new log")
            handler.doRollover()
            time.sleep(0.1)
        backups = [f for f in os.listdir(self.test_dir) if f != "customlog.log"]
        self.assertEqual(len(backups), 2)
        self.assertNotIn("customlog-2023-10-05T12-00-00.000.log", backups)
        self.assertNotIn("customlog-2023-10-05T12-10-00.000.log", backups)
        self.assertNotIn("customlog-2023-10-05T12-20-00.000.log", backups)

    def test_ignores_invalid_filenames(self):
        """
        test ignoring invalid filenames
        """
        handler = self.create_handler(backup_count=1)
        valid_file = "customlog-2023-10-05T12-00-00.000.log"
        invalid_files = [
            "customlog-invalid.log",
            "otherfile.log",
            "customlog-2023-10-05T12-00-00.000.txt"
        ]
        for fname in [valid_file] + invalid_files:
            with open(os.path.join(self.test_dir, fname), "w") as f:
                f.write("new log")
        with open(self.base_filename, "w") as f:
            f.write("new log")
        handler.doRollover()
        remaining = os.listdir(self.test_dir)
        for invalid in invalid_files:
            self.assertIn(invalid, remaining)
        backups = [f for f in os.listdir(self.test_dir) if re.match(rf"customlog-{LOG_BACKUP_PATTERN}.log", f)]
        self.assertEqual(len(backups), 1)


if __name__ == "__main__":
    unittest.main()
