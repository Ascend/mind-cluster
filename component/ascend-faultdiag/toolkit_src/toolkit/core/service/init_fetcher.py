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

import asyncio
import os
from pathlib import Path
from typing import Dict, Type, List, Tuple

from toolkit.core.collect.collect_config import SwiCliOutputDataType
from toolkit.core.collect.fetcher.dump_log_fetcher.base import DumpLogDirParser
from toolkit.core.collect.fetcher.dump_log_fetcher.bmc.bmc_dump_log_fetcher import BmcDumpLogFetcher
from toolkit.core.collect.fetcher.dump_log_fetcher.bmc.bmc_dump_log_parser import BmcDumpLogParser
from toolkit.core.collect.fetcher.dump_log_fetcher.cli_output_parsed_data import CliOutputParsedData
from toolkit.core.collect.fetcher.dump_log_fetcher.host.host_dump_log_fetcher import HostDumpLogFetcher
from toolkit.core.collect.fetcher.dump_log_fetcher.host.host_log_parser_builder import HostLogParserBuilder
from toolkit.core.collect.fetcher.dump_log_fetcher.switch.diag_info_output.collect_diag_info_log_parser import \
    CollectDiagInfoLogParser
from toolkit.core.collect.fetcher.dump_log_fetcher.switch.swi_cli_output_parser import SwiCliOutputParser
from toolkit.core.collect.fetcher.dump_log_fetcher.switch.swi_cli_output_fetcher import SwiCliOutputFetcher
from toolkit.core.collect.fetcher.dump_log_fetcher.switch.switch_log_path_finder import SwitchLogPathFinder
from toolkit.core.collect.fetcher.ssh_fetcher.bmc_ssh_fetcher import BmcSshFetcher
from toolkit.core.collect.fetcher.ssh_fetcher.host_ssh_fetcher import HostSshFetcher
from toolkit.core.collect.fetcher.ssh_fetcher.switch_ssh_fetcher import SwiSshFetcher
from toolkit.core.common import constants
from toolkit.core.common.diag_enum import CollectType
from toolkit.core.common.path import CommonPath
from toolkit.core.config.conn_config import Conn
from toolkit.core.context.diag_ctx import DiagCtx
from toolkit.core.service.base import DiagService
from toolkit.utils import file_tool
from toolkit.utils.compress_tool import CompressTool
from toolkit.utils.executors import AsyncSSHExecutor
from toolkit.utils.file_tool import convert_log_path
from toolkit.utils.logger import DIAG_LOGGER
from toolkit.utils.ping_tool import PingTool


class InitFetcher(DiagService):
    _FIND_DEPTH = 3

    def __init__(self, diag_ctx: DiagCtx, collect_type: CollectType = CollectType.ALL):
        super().__init__(diag_ctx)
        self.collect_type = collect_type
        self._add_fetcher_task_map = {
            CollectType.SSH: self.add_ssh_fetchers,
            CollectType.LOCAL: self.add_local_file_fetchers,
        }
        self._add_fetcher_tasks = []
        if self.collect_type == CollectType.ALL:
            self._add_fetcher_tasks.extend(list(self._add_fetcher_task_map.values()))
        else:
            self._add_fetcher_tasks.append(self._add_fetcher_task_map[self.collect_type])

    @staticmethod
    async def _check_and_add_executor(conn, fetchers_map, fetcher_type):
        executor = AsyncSSHExecutor(conn.host, conn.port, conn.username, conn.password, conn.private_key)
        executor.ensure_shell_session()
        if executor.shell_channel:
            fetchers_map[executor.host] = fetcher_type(executor)

    async def run(self):
        tasks = [fetcher() for fetcher in self._add_fetcher_tasks]
        await asyncio.gather(*tasks)

    async def add_ssh_fetchers(self):
        if not os.path.exists(CommonPath.ENCRYPTED_CONN_CONFIG_PATH):
            DIAG_LOGGER.warning("加密配置文件不存在，从默认路径获取配置信息")
            self.diag_ctx.encrypt_conn_config()
        res = self.diag_ctx.load_conn_config()
        if res:
            DIAG_LOGGER.error(res)
            return
        if not self.diag_ctx.conn_config:
            DIAG_LOGGER.warning("未获取到有用的配置信息")
            return
        futures = [
            *self._ping_ssh_conn(self.diag_ctx.conn_config.switch_conn, self.diag_ctx.switch_fetchers, SwiSshFetcher),
            *self._ping_ssh_conn(self.diag_ctx.conn_config.host_conn, self.diag_ctx.host_fetchers, HostSshFetcher),
            *self._ping_ssh_conn(self.diag_ctx.conn_config.bmc_conn, self.diag_ctx.bmcs_fetchers, BmcSshFetcher),
        ]
        async_tasks = []
        for conn, future, fetchers_map, fetcher_type in futures:
            success, info = await future
            if not success:
                DIAG_LOGGER.warning(f"ping {conn.host} failed: {info}")
                continue
            async_tasks.append(self._check_and_add_executor(conn, fetchers_map, fetcher_type))
        await asyncio.gather(*async_tasks)

    async def add_local_file_fetchers(self):
        await asyncio.gather(
            # 用户输入的目录或者默认项目目录下的Host日志
            self._uncompress_and_add_host_file_fetchers(self.diag_ctx.dump_log_dir_config.host_dump_log_dir),
            # 执行目录下的Host日志
            self._uncompress_and_add_host_file_fetchers(CommonPath.CUR_PATH_HOST_DUMP_LOG_DIR),
            # 用户输入的目录或者默认项目目录下的BMC日志
            self._uncompress_and_add_bmc_file_fetchers(self.diag_ctx.dump_log_dir_config.bmc_dump_log_dir),
            # 通过工具自动收集的BMC日志
            self._uncompress_and_add_bmc_file_fetchers(CommonPath.TOOL_HOME_BMC_DUMP_CACHE_DIR),
            # 执行目录下的BMC日志
            self._uncompress_and_add_bmc_file_fetchers(CommonPath.CUR_PATH_BMC_DUMP_LOG_DIR),
            # 用户输入的目录或者默认项目目录下的交换机日志
            self._add_swi_file_fetchers(self.diag_ctx.dump_log_dir_config.switch_dump_log_dir),
            # 执行目录下的Switch日志
            self._add_swi_file_fetchers(CommonPath.CUR_PATH_SWITCH_DUMP_LOG_DIR),
        )

    def _ping_ssh_conn(self, conn_list: List[Conn], fetchers_map: Dict, fetcher_type: Type):
        futures = []
        for conn in conn_list:
            future = self.diag_ctx.submit_multi_process_task(PingTool.ping, conn.host)
            futures.append((conn, future, fetchers_map, fetcher_type))
        return futures

    async def _uncompress_and_add_host_file_fetchers(self, host_dump_log_dir):
        if not os.path.exists(host_dump_log_dir):
            return
        await self._uncompress_log_pkgs(host_dump_log_dir)
        await self._add_host_fetchers(host_dump_log_dir)

    async def _uncompress_and_add_bmc_file_fetchers(self, bmc_dump_log_dir):
        if not os.path.exists(bmc_dump_log_dir):
            return
        await self._uncompress_log_pkgs(bmc_dump_log_dir)
        await self._add_bmc_fetchers(bmc_dump_log_dir)

    async def _uncompress_log_pkgs(self, root_dir):
        root_path = convert_log_path(root_dir)
        if not root_path:
            return
        futures = []
        # 解压不超过3层的压缩包
        file_paths = file_tool.find_all_sub_paths(str(root_path), "*.tar.gz", self._FIND_DEPTH)
        for file_path in file_paths:
            # file_path: 压缩包的完整 Path 对象（如 /data/subdir/file.tar.gz）
            file_path = Path(file_path)
            targz_abspath = str(file_path)  # 压缩包绝对路径
            targz_dir = str(file_path.parent)  # 压缩包所在目录
            future = self.diag_ctx.submit_multi_process_task(CompressTool.extract_tar_gz, targz_abspath, targz_dir)
            futures.append([targz_abspath, future])
        for targz_abspath, future in futures:
            try:
                await future
            except Exception as e:
                DIAG_LOGGER.error(f"文件{targz_abspath}解压任务执行失败: {e}")

    async def _add_host_fetchers(self, host_dump_log_dir: str):
        parsers = HostLogParserBuilder.build(host_dump_log_dir)
        await self._add_file_fetchers_by_parser(self.diag_ctx.host_fetchers, parsers, HostDumpLogFetcher)

    async def _add_bmc_fetchers(self, bmc_dump_log_dir: str):
        log_collect_dirs = file_tool.find_all_sub_paths(bmc_dump_log_dir,
                                                        constants.TOOL_BMC_LOG_COLLECT_DIR_NAME, self._FIND_DEPTH)
        parsers = [BmcDumpLogParser(bmc_dump_log_dir, log_collect_dir) for log_collect_dir in log_collect_dirs]
        return await self._add_file_fetchers_by_parser(self.diag_ctx.bmcs_fetchers, parsers, BmcDumpLogFetcher)

    async def _add_file_fetchers_by_parser(self, fetchers_map: Dict, parsers: List[DumpLogDirParser],
                                           fetcher_type: Type):
        futures = []
        for parser in parsers:
            futures.append([parser.parse_dir, self.diag_ctx.submit_multi_process_task(parser.parse)])
        for log_collect_dir, future in futures:
            try:
                data_dict = await future
                if not data_dict:
                    continue
                fetcher = fetcher_type(log_collect_dir, CliOutputParsedData(data_dict))
                fetcher_id = await fetcher.fetch_id()
                fetchers_map[fetcher_id] = fetcher
            except Exception as e:
                DIAG_LOGGER.error(f"parse dir {log_collect_dir} failed: {e}")

    async def _add_swi_file_fetchers(self, switch_dump_log_dir: str):
        root_dir = Path(switch_dump_log_dir)
        if not root_dir.exists() or not root_dir.is_dir():
            return
        await self._unzip_diag_info_zip(switch_dump_log_dir)
        cli_output_futures, diag_info_futures = await self._parse_cli_and_log(switch_dump_log_dir)
        await self._add_swi_file_fetcher_by_future(cli_output_futures, diag_info_futures)

    async def _unzip_diag_info_zip(self, switch_dump_log_dir: str):
        # 找到zip包并解压
        diag_info_log_paths = file_tool.find_all_sub_paths(switch_dump_log_dir, "*.zip", self._FIND_DEPTH)
        unzip_futures = []
        for diag_info_log_path in diag_info_log_paths:
            file_path = Path(diag_info_log_path)
            zip_abspath = str(file_path)  # 压缩包绝对路径
            zip_dir = str(file_path.parent)  # 压缩包所在目录
            future = self.diag_ctx.submit_multi_process_task(CompressTool.extract_zip_recursive,
                                                             zip_abspath, zip_dir, 4)
            unzip_futures.append(future)
        for future in unzip_futures:
            try:
                await future
            except Exception as e:
                DIAG_LOGGER.error(f"unzip zip failed: {e}")

    async def _parse_cli_and_log(self, switch_dump_log_dir: str) -> Tuple[List[asyncio.Future], List[asyncio.Future]]:
        diag_info_dirs, cli_output_txt_paths = SwitchLogPathFinder.find(switch_dump_log_dir)
        # 解析回显
        cli_output_futures = []
        for file_path in cli_output_txt_paths:
            future = self.diag_ctx.submit_multi_process_task(SwiCliOutputParser.parse, file_path)
            cli_output_futures.append(future)
        # 解析日志
        diag_info_futures = []
        for diag_info_dir in diag_info_dirs:
            future = self.diag_ctx.submit_multi_process_task(CollectDiagInfoLogParser.parse, diag_info_dir)
            diag_info_futures.append(future)
        return cli_output_futures, diag_info_futures

    async def _add_swi_file_fetcher_by_future(self, cli_output_futures: List[asyncio.Future],
                                              diag_info_futures: List[asyncio.Future]):
        # 获取回显fetcher
        local_fetchers: Dict[str, SwiCliOutputFetcher] = {}
        for future in cli_output_futures:
            data_dict = await future
            if not data_dict:
                continue
            fetcher = SwiCliOutputFetcher(CliOutputParsedData(data_dict))
            local_fetchers[await fetcher.get_switch_name()] = fetcher
            # 添加日志到fetcher
        for diag_info_future in diag_info_futures:
            diag_info_parse_result = await diag_info_future
            swi_name = diag_info_parse_result.swi_name
            find_results = diag_info_parse_result.find_log_results
            port_down_status = diag_info_parse_result.port_down_status
            # 日志有配套的回显
            if swi_name in local_fetchers:
                fetcher = local_fetchers[swi_name]
            # 没有配套回显
            elif swi_name:
                fetcher = SwiCliOutputFetcher(CliOutputParsedData())
                local_fetchers[swi_name] = fetcher
                fetcher.parsed_data.add_data([SwiCliOutputDataType.SWI_NAME.name], swi_name)
            else:
                continue
            fetcher.parsed_data.add_data([SwiCliOutputDataType.DIAG_INFO_LOG.name], find_results)
            fetcher.parsed_data.add_data([SwiCliOutputDataType.PORT_DOWN_STATUS.name], port_down_status)
        self.diag_ctx.switch_fetchers.update({await fetcher.fetch_id(): fetcher for fetcher in local_fetchers.values()})
