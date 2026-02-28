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

import json
import os.path
from typing import List, Dict

from ascend_fd_tk.core.common.path import CommonPath
from ascend_fd_tk.core.log_parser import parse_script
from ascend_fd_tk.core.log_parser.base import LogParsePattern, FindResult, LogParser
from ascend_fd_tk.utils.executors import AsyncExecutor, CmdTask


class RemoteLogPyScriptParser(LogParser):
    """
    使用python脚本远程清洗
    """

    def __init__(self, executor: AsyncExecutor):
        self.executor = executor
        self.local_parse_result_path = os.path.join(CommonPath.PARSE_CACHE_DIR,
                                                    f"{self.executor.host}_parse_result.json")

    async def find(self, parse_dir="", log_pattern_map: Dict[str, LogParsePattern] = None) -> List[FindResult]:
        if os.path.exists(self.local_parse_result_path):
            os.remove(self.local_parse_result_path)
        configs = [log_pattern.keyword_config.to_dict() for log_pattern in log_pattern_map.values()]
        os.makedirs(CommonPath.PARSE_CACHE_DIR, exist_ok=True)
        with open(CommonPath.PARSE_CONFIG_FILE, 'w', encoding='utf-8') as f:
            f.write(json.dumps(configs, ensure_ascii=False))
        await self.executor.run_cmd(CmdTask("mkdir -p ~/.ascend-fd-tk"))
        await self.executor.upload_file(CommonPath.PARSE_CONFIG_FILE, "~/.ascend-fd-tk/parse_config.json")
        await self.executor.upload_file(parse_script.SCRIPT_PATH, "~/.ascend-fd-tk/parse_script.py")
        await self.executor.run_cmd(CmdTask(f"python3 ~/.ascend-fd-tk/parse_script.py {parse_dir}"))
        await self.executor.download_file("~/.ascend-fd-tk/parse_result.json", self.local_parse_result_path)
        if not os.path.exists(self.local_parse_result_path):
            return []
        with open(self.local_parse_result_path, 'r', encoding='utf-8') as f:
            parse_results = json.loads(f.read()) or []
        find_results = [FindResult.from_dict(parse_result) for parse_result in parse_results]
        return self.fill_search_info(find_results, log_pattern_map)
