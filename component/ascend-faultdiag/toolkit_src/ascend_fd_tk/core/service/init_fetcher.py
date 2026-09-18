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

from ascend_fd_tk.core.collect.collect_config import SwiCliOutputDataType
from ascend_fd_tk.core.collect.fetcher.dump_log_fetcher.base import DumpLogDirParser
from ascend_fd_tk.core.collect.fetcher.dump_log_fetcher.bmc.bmc_dump_log_fetcher import BmcDumpLogFetcher
from ascend_fd_tk.core.collect.fetcher.dump_log_fetcher.bmc.bmc_dump_log_parser import BmcDumpLogParser
from ascend_fd_tk.core.collect.fetcher.dump_log_fetcher.cli_output_parsed_data import CliOutputParsedData
from ascend_fd_tk.core.collect.fetcher.dump_log_fetcher.host.host_dump_log_fetcher import HostDumpLogFetcher
from ascend_fd_tk.core.collect.fetcher.dump_log_fetcher.host.host_log_parser_builder import HostLogParserBuilder
from ascend_fd_tk.core.collect.fetcher.dump_log_fetcher.switch.diag_info_output.collect_diag_info_log_parser import (
    CollectDiagInfoLogParser,
)
from ascend_fd_tk.core.collect.fetcher.dump_log_fetcher.switch.swi_cli_output_parser import SwiCliOutputParser
from ascend_fd_tk.core.collect.fetcher.dump_log_fetcher.switch.swi_cli_output_fetcher import SwiCliOutputFetcher
from ascend_fd_tk.core.collect.fetcher.dump_log_fetcher.switch.switch_log_path_finder import SwitchLogPathFinder
from ascend_fd_tk.core.collect.fetcher.ssh_fetcher.bmc_ssh_fetcher import BmcSshFetcher
from ascend_fd_tk.core.collect.fetcher.ssh_fetcher.generation_probe.factory import create_generation_probe
from ascend_fd_tk.core.collect.fetcher.ssh_fetcher.host_ssh_fetcher import HostSshFetcher
from ascend_fd_tk.core.collect.fetcher.ssh_fetcher.host_ssh_fetcher_factory import create_host_ssh_fetcher
from ascend_fd_tk.core.collect.fetcher.ssh_fetcher.podmanager_ssh_fetcher import PoDManagerSshFetcher
from ascend_fd_tk.core.collect.fetcher.ssh_fetcher.switch_ssh_fetcher import SwiSshFetcher
from ascend_fd_tk.core.common import constants
from ascend_fd_tk.core.common.diag_enum import CollectType
from ascend_fd_tk.core.common.path import CommonPath
from ascend_fd_tk.core.config.conn_config import Conn
from ascend_fd_tk.core.context.diag_ctx import DiagCtx
from ascend_fd_tk.core.model.conn_key import ConnKey
from ascend_fd_tk.core.service.base import DiagService
from ascend_fd_tk.utils import file_tool
from ascend_fd_tk.utils.compress_tool import CompressTool
from ascend_fd_tk.utils.executors import AsyncSSHExecutor
from ascend_fd_tk.utils.file_tool import convert_log_path
from ascend_fd_tk.utils.logger import DIAG_LOGGER
from ascend_fd_tk.utils.ping_tool import PingTool


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
    async def _add_pod_manager_executors(conn, fetchers_map):
        """PoDManager 按分片建连接：槽位均分到有限个连接，连接内串行切换槽位执行。

        建连失败的连接不丢弃其负责的槽位：按成功连接数重新轮询分片，
        保证每个槽位都被剩余连接接管，避免漏采。
        """

        async def _create(conn_idx):
            executor = AsyncSSHExecutor(conn.host, conn.port, conn.username, conn.password, conn.private_key)
            await asyncio.get_running_loop().run_in_executor(None, executor.ensure_shell_session)
            return conn_idx, executor

        created = await asyncio.gather(*(_create(i) for i in range(constants.POD_MANAGER_CONN_NUM)))
        ok_executors = []
        for conn_idx, executor in created:
            if not executor.shell_channel:
                DIAG_LOGGER.warning(
                    "PoDManager %s 第 %s 个连接建立 SSH 连接失败，其槽位由其余连接接管", conn.host, conn_idx
                )
                continue
            ok_executors.append((conn_idx, executor))
        if not ok_executors:
            DIAG_LOGGER.warning("PoDManager %s 所有连接建立失败，跳过该设备采集", conn.host)
            return
        # 重新轮询分片：槽位均分到成功连接，连接内串行切换
        ok_conn_num = len(ok_executors)
        npu_slot_ids = constants.POD_MANAGER_NPU_SLOT_IDS
        sfu_slot_ids = constants.POD_MANAGER_SFU_SLOT_IDS
        for offset, (conn_idx, executor) in enumerate(ok_executors):
            fetcher_slots = sfu_slot_ids[offset::ok_conn_num]
            # A5 NPU 槽位同样按轮询分片到各连接，用于跳转 NPU 槽位采集 switch QoS credit
            fetcher_npu_slots = npu_slot_ids[offset::ok_conn_num]
            fetcher = PoDManagerSshFetcher(executor, fetcher_slots, fetcher_npu_slots)
            fetchers_map[ConnKey(executor.host, conn_idx)] = fetcher
        DIAG_LOGGER.info(
            "PoDManager %s 初始化成功：共创建 %s 个连接，槽位分片：%s",
            conn.host,
            ok_conn_num,
            {conn_idx: sfu_slot_ids[offset::ok_conn_num] for offset, (conn_idx, _) in enumerate(ok_executors)},
        )

    async def _check_and_add_executor(self, conn, fetchers_map, fetcher_type):
        # PoDManager 一个 IP 按内置 slot 表建立多个 SSH 连接
        if fetcher_type is PoDManagerSshFetcher:
            await self._add_pod_manager_executors(conn, fetchers_map)
            return
        executor = AsyncSSHExecutor(conn.host, conn.port, conn.username, conn.password, conn.private_key)
        executor.ensure_shell_session()
        if not executor.shell_channel:
            return
        # 需要代际探测的设备（host）先探测再按代际构造；switch/bmc 直接构造
        probe = create_generation_probe(fetcher_type)
        if probe is not None:
            generation, extra = await probe.probe(executor)
            if fetcher_type is HostSshFetcher:
                fetchers_map[executor.host] = create_host_ssh_fetcher(executor, generation, npu_mapping_cache=extra)
        else:
            fetchers_map[executor.host] = fetcher_type(executor)

    async def run(self):
        tasks = [fetcher() for fetcher in self._add_fetcher_tasks]
        await asyncio.gather(*tasks)

    async def add_ssh_fetchers(self):
        if not os.path.exists(CommonPath.ENCRYPTED_CONN_CONFIG_PATH):
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
            *self._ping_ssh_conn(
                self.diag_ctx.conn_config.pod_manager_conn, self.diag_ctx.pod_manager_fetchers, PoDManagerSshFetcher
            ),
        ]
        async_tasks = []
        for conn, future, fetchers_map, fetcher_type in futures:
            success, info = await future
            if not success:
                DIAG_LOGGER.warning("ping %s failed: %s", conn.host, info)
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
        host_dump_log_dir = convert_log_path(host_dump_log_dir)
        if not host_dump_log_dir:
            return
        await self._uncompress_log_pkgs(host_dump_log_dir)
        await self._add_host_fetchers(host_dump_log_dir)

    async def _uncompress_and_add_bmc_file_fetchers(self, bmc_dump_log_dir):
        bmc_dump_log_dir = convert_log_path(bmc_dump_log_dir)
        if not bmc_dump_log_dir:
            return
        await self._uncompress_log_pkgs(bmc_dump_log_dir)
        await self._add_bmc_fetchers(bmc_dump_log_dir)

    async def _uncompress_log_pkgs(self, root_dir):
        if not root_dir:
            return
        futures = []
        # 解压不超过3层的压缩包
        file_paths = file_tool.find_all_sub_paths(root_dir, "*.tar.gz", self._FIND_DEPTH)
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
                DIAG_LOGGER.error("文件%s解压任务执行失败: %s", targz_abspath, str(e))

    async def _add_host_fetchers(self, host_dump_log_dir: str):
        parsers = HostLogParserBuilder.build(host_dump_log_dir)
        await self._add_file_fetchers_by_parser(self.diag_ctx.host_fetchers, parsers, HostDumpLogFetcher)

    async def _add_bmc_fetchers(self, bmc_dump_log_dir: str):
        log_collect_dirs = file_tool.find_all_sub_paths(
            bmc_dump_log_dir, constants.TOOL_BMC_LOG_COLLECT_DIR_NAME, self._FIND_DEPTH
        )
        parsers = [
            BmcDumpLogParser(bmc_dump_log_dir, log_collect_dir)
            for log_collect_dir in log_collect_dirs
            if os.path.isdir(log_collect_dir)  # 过滤掉名字为dump_info的文件
        ]
        return await self._add_file_fetchers_by_parser(self.diag_ctx.bmcs_fetchers, parsers, BmcDumpLogFetcher)

    async def _add_file_fetchers_by_parser(
        self, fetchers_map: Dict, parsers: List[DumpLogDirParser], fetcher_type: Type
    ):
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
                DIAG_LOGGER.error("parse dir %s failed: %s", log_collect_dir, str(e))

    async def _add_swi_file_fetchers(self, switch_dump_log_dir: str):
        switch_dump_log_dir = convert_log_path(switch_dump_log_dir)
        if not switch_dump_log_dir:
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
            future = self.diag_ctx.submit_multi_process_task(
                CompressTool.extract_zip_recursive, zip_abspath, zip_dir, 4
            )
            unzip_futures.append(future)
        for future in unzip_futures:
            try:
                await future
            except Exception as e:
                DIAG_LOGGER.error("unzip zip failed: %s", str(e))

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

    async def _add_swi_file_fetcher_by_future(
        self, cli_output_futures: List[asyncio.Future], diag_info_futures: List[asyncio.Future]
    ):
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
