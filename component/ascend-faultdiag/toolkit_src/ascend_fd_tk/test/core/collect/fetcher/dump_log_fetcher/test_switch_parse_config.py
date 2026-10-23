#!/usr/bin/env python3
# -*- coding: utf-8 -*-
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

"""交换机 CLI 回显解析配置单元测试：alarm 分段判断与配置项。"""

import unittest

from ascend_fd_tk.core.collect.collect_config import SwiCliOutputDataType
from ascend_fd_tk.core.collect.fetcher.dump_log_fetcher.switch.parse_config import (
    PARSE_CONFIGS,
    _display_alarm_judge_func,
    _display_alarm_active_judge_func,
    _display_alarm_history_judge_func,
)


class TestDisplayAlarmJudge(unittest.TestCase):
    def test_judge_history(self):
        self.assertTrue(_display_alarm_judge_func("0> 2026-08-29 10:00:00 h detail\n"))
        self.assertTrue(_display_alarm_history_judge_func("0> 2026-08-29 10:00:00 h detail\n"))
        self.assertFalse(_display_alarm_active_judge_func("0> 2026-08-29 10:00:00 h detail\n"))

    def test_judge_active(self):
        self.assertFalse(_display_alarm_judge_func("no separator"))
        self.assertTrue(_display_alarm_active_judge_func("0> 2026-08-29 10:00:00\n"))
        self.assertFalse(_display_alarm_history_judge_func("0> 2026-08-29 10:00:00\n"))


class TestSwiCliOutputParseConfigs(unittest.TestCase):
    def test_parse_configs_cover_key_types(self):
        data_types = {config.data_type for config in PARSE_CONFIGS}
        self.assertIn(SwiCliOutputDataType.CLOCK, data_types)
        self.assertIn(SwiCliOutputDataType.ALARM_ACTIVE, data_types)
        self.assertIn(SwiCliOutputDataType.ALARM_HISTORY, data_types)
        self.assertIn(SwiCliOutputDataType.OPTICAL_MODULE, data_types)
        self.assertTrue(all(config.primary_keys for config in PARSE_CONFIGS))


if __name__ == "__main__":
    unittest.main()
