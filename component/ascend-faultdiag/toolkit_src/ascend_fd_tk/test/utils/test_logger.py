#!/usr/bin/env python3
# Copyright 2026 Huawei Technologies Co., Ltd
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
import tempfile
import unittest

from ascend_fd_tk.utils.logger import SimpleLogger


class TestSimpleLoggerRotation(unittest.TestCase):
    def test_rotation_keeps_backup_count(self):
        with tempfile.TemporaryDirectory() as log_dir:
            simple_logger = SimpleLogger(
                logger_name="st_rotation_logger",
                log_dir=log_dir,
                log_filename="app.log",
                max_size=200,
                backup_count=2,
            )
            logger = simple_logger.get_logger()
            logger.propagate = False
            try:
                for index in range(100):
                    logger.info("log line %s %s", index, "x" * 50)
                for handler in logger.handlers:
                    handler.flush()
                files = os.listdir(log_dir)
                self.assertIn("app.log", files)
                archives = [name for name in files if name.startswith("app.log.")]
                self.assertGreaterEqual(len(archives), 1)
                self.assertLessEqual(len(archives), 2)
            finally:
                for handler in logger.handlers:
                    handler.close()
                logger.handlers = []

    def test_rotation_handler_config(self):
        with tempfile.TemporaryDirectory() as log_dir:
            simple_logger = SimpleLogger(
                logger_name="st_rotation_config_logger",
                log_dir=log_dir,
                log_filename="app.log",
                max_size=1024,
                backup_count=3,
            )
            logger = simple_logger.get_logger()
            try:
                handler = next(handler for handler in logger.handlers if hasattr(handler, "backupCount"))
                self.assertEqual(handler.maxBytes, 1024)
                self.assertEqual(handler.backupCount, 3)
                self.assertTrue(os.path.exists(os.path.join(log_dir, "app.log")))
            finally:
                for handler in logger.handlers:
                    handler.close()
                logger.handlers = []


if __name__ == "__main__":
    unittest.main()
