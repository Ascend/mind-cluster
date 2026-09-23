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
# pylint: disable=duplicate-code
import bisect
from datetime import datetime, timedelta
import logging
import os
from pathlib import Path
import re
from typing import Optional, Union

from ascend_fd.model.context import KGParseCtx
from ascend_fd.pkg.parse.knowledge_graph.parser.file_parser import (
    EventStorage,
    FileParser,
)
from ascend_fd.pkg.parse.parser_saver import LogInfoSaver
from ascend_fd.utils.constant.dev_log_const import HIST_DEVICE_OS_PATH_ARRAY
from ascend_fd.utils.constant.str_const import UNKNOWN_DEVICE_ID
from ascend_fd.utils.constant.ub_const import (
    ALL_UBCTL_KEYS,
    PRECHECK_KERNEL_AE_TYPE23,
    PRECHECK_UBCTL_DATA,
    SIDE_AFTER,
    SIDE_BEFORE,
    UBCTL_DUMP_MODE,
    UDIE_MAX_PORT_NUM,
)
from ascend_fd.utils.fault_code import FIBER_OR_COPPER_LINK_FAULT
from ascend_fd.utils.regular_table import (
    NPU_DEVICE_SOURCE,
    NPU_HISTORY_SOURCE,
    NPU_HOST_SOURCE,
    NPU_OS_SOURCE,
    NPU_UBCTL_SOURCE,
)
from ascend_fd.utils.tool import (
    MultiProcessJob,
    check_and_format_time_str,
    safe_read_open,
)

kg_logger = logging.getLogger("KNOWLEDGE_GRAPH")
LOCAL_FAULT_FLAG = "1"
PCS_NORMAL_TH = "0"
NEGATIVE_ONE = "-1"
TS_PATTERN = re.compile(r'^(\d{14})-\d{9}$')
# 路径段级匹配 device-x / dev-os-x：段边界以 / 或 \ 或字符串首（尾）约束，单捕获组直接取设备号。
# 对完整文件路径匹配即可，不依赖固定目录层级（history.log / kernel.log / os_info.txt / hbm.txt / ubctl_log.txt 等任意深度）
DEVICE_ID_PATTERN = re.compile(r"(?:(?<=[\\/])|^)(?:device|dev-os)-(\d{1,3})(?=$|[\\/])")
TIME_FORMAT = "%Y-%m-%d %H:%M:%S"


class BaseNpuLogParser(FileParser):
    @staticmethod
    def _get_occur_time(line, dir_name) -> str:
        pass

    @staticmethod
    def _filter_log_list(log_list, start_time, end_time, time_format):
        """
        Filter log list by time, only the files in the training period are retained.
        :param log_list: log list
        :param start_time: start time
        :param end_time: end time
        :param time_format: time format
        :return: new log list
        """
        if start_time >= end_time:
            return log_list
        start_file = time_format.format(start_time.replace("-", "").replace(":", "").replace(" ", ""))
        end_file = time_format.format(end_time.replace("-", "").replace(":", "").replace(" ", ""))
        # use bisect to get the start time file and end time file index, obtaining logs in the Training Time Range
        start_idx = max(bisect.bisect(log_list, start_file) - 1, 0)
        end_idx = bisect.bisect(log_list, end_file)
        return log_list[start_idx:end_idx]

    def parse(self, parse_ctx: KGParseCtx, task_id):
        pass

    def process_parse_file_list(self, file_source_list: list, task_id: str):
        """
        Parse multiple files, multiple processes them in cmd input scene
        :param file_source_list: the log file list
        :param task_id: the task unique id
        :return:
        """
        events_list = []
        kg_logger.info("%s files parse job started.", self.SOURCE_FILE)
        if self.is_sdk_input:
            results = dict()
            for idx, file_source in enumerate(file_source_list):
                results.update(
                    {
                        f"{self.SOURCE_FILE}_ID-{idx}_{self._get_filename(file_source)}": self._parse_single_file(
                            file_source
                        )
                    }
                )
        else:
            if not file_source_list:
                kg_logger.info("No %s files matched, skip the %s log parse job.", self.SOURCE_FILE, self.SOURCE_FILE)
                return [], {}
            multiprocess_job = MultiProcessJob("KNOWLEDGE_GRAPH", pool_size=len(file_source_list), task_id=task_id)
            for idx, file_source in enumerate(file_source_list):
                multiprocess_job.add_security_job(
                    f"{self.SOURCE_FILE}_ID-{idx}_{self._get_filename(file_source)}",
                    self._parse_single_file,
                    file_source,
                )
            results, _ = multiprocess_job.join_and_get_results()
        for event_list in results.values():
            events_list.extend(event_list)
        kg_logger.info("%s files parse job is complete.", self.SOURCE_FILE)
        return events_list, {}

    def _parse_single_file(self, file_source: Union[str, LogInfoSaver]):
        """
        Parse single slog file
        :param file_source: slog file path
        :return: parse result
        """
        if not self.is_sdk_input and not os.path.isfile(file_source):
            return []
        device_id = self._determine_device_id(file_source)
        dir_name = os.path.dirname(file_source.path) if self.is_sdk_input else os.path.dirname(file_source)
        event_storage = EventStorage()
        for line in self._yield_log(file_source):
            occur_time = self._get_occur_time(line, dir_name)
            if not occur_time or occur_time < self.resuming_training_time:
                continue  # when cannot get time or this line is before resuming training, ignore this line
            # not in the Training Time Range, ignore it
            if self.start_time and occur_time < self.start_time:
                continue
            if self.end_time and occur_time > self.end_time:
                continue
            event_dict = self.parse_single_line(line)
            if not event_dict:
                # kg-config.json 未命中时，尝试匹配 PRECHECK 行
                event_dict = self._match_precheck_line(line)
            if not event_dict:
                continue
            # if rf_lf does not exist or when rf_lf is 1, pcs_err_cnt is equal to 0, the fault condition is not met
            rf_lf = event_dict.get("rf_lf", "")
            pcs_err_cnt = event_dict.get("pcs_err_cnt", PCS_NORMAL_TH)
            is_custom_fault = event_dict.get("event_code", "") == FIBER_OR_COPPER_LINK_FAULT
            local_fault_not_met = (rf_lf == LOCAL_FAULT_FLAG) and (pcs_err_cnt == PCS_NORMAL_TH)
            if is_custom_fault and (not rf_lf or local_fault_not_met):
                continue
            event_dict.update({"source_device": device_id})
            self.supplement_common_info(event_dict, file_source, occur_time)
            event_storage.record_event(event_dict)
        return event_storage.generate_event_list()

    def _match_precheck_line(self, log_line: str) -> dict:
        """
        Hook for subclasses: match a PRECHECK line not covered by kg-config.json.
        Override to generate a PRECHECK event, return {} by default.
        :param log_line: single log line
        :return: PRECHECK event dict or {}
        """
        return {}

    def _determine_device_id(self, file_source):
        dir_name = os.path.dirname(file_source.path) if self.is_sdk_input else os.path.dirname(file_source)
        # 对完整路径做段级匹配，取最后一个（离文件最近的）device-x / dev-os-x 段，
        # 兼容 hisi/slog/ubctl 的不同目录层级，不再依赖固定层级向上取 device 目录
        device_match = None
        for match in DEVICE_ID_PATTERN.finditer(dir_name):
            device_match = match
        if device_match:
            device_id = device_match.group(1)
        else:
            device_id = getattr(file_source, "device_id_str", "Unknown")
            if device_id == "Unknown":
                kg_logger.warning("The %s may not be a regular file, please check.", self._get_filename(file_source))
        return device_id


class NpuHistoryLogParser(BaseNpuLogParser):
    TARGET_FILE_PATTERNS = "hisi_logs_path"
    SOURCE_FILE = NPU_HISTORY_SOURCE
    DIR_ERR_NUM = 100

    # 日志格式为UDMA xxxxxxxx ae xxxxx, event type is 2, sub type is 3
    KERNEL_AE_KEYWORDS = ("UDMA", "ae", "event type is 2", "sub type is 3")
    # ---------- kernel.log 时间精确化：device_os.log 时间参照 ----------
    # device_os.log 相对时间戳子目录路径：<子目录>/log/slog/debug/device_os.log
    _DEVICE_OS_LOG_REL_PATH = os.path.join(*HIST_DEVICE_OS_PATH_ARRAY)
    # 参照行必须以 [ERROR] KERNEL 开头，且只取第一条
    _DEVICE_OS_REFERENCE_HEAD = "[ERROR] KERNEL"
    # 参照行示例：[ERROR] KERNEL(3869,slogd):2026-09-20-12:45:34.644.410 [slogd_kernel_log.c:254][90032.054298] ...
    # 提取 "KERNEL(pid,组):" 之后的实际时间（到空格为止），即参照实际时间 W0
    _DEVICE_OS_REAL_TIME_PATTERN = re.compile(r"KERNEL\([^)]*\):\s*([^ ]+)")
    # 提取方括号中的开机时长（形如 [90032.054298]，单位秒），即参照开机时长 T0
    _DEVICE_OS_UPTIME_PATTERN = re.compile(r"\[(\d+\.\d+)\]")
    # kernel.log 行首开机时长（形如 [95186.485753]）
    _HISI_UPTIME_PATTERN = re.compile(r"^\d+(?:\.\d+)?$")

    def __init__(self, params: dict):
        super().__init__(params)
        # 非 SDK 离线日志场景：{时间戳子目录: device_os.log 时间参照}，parse() 时构建
        self.time_reference_map = {}

    @staticmethod
    def _get_occur_time(line, dir_name):
        """
        Get the time info.
        history.log e.g: [yyyy-mm-dd-hh:mm:ss.******] *****
        :param line: log line
        :return: time info
        """
        time_str = line[line.find("[") + 1 : line.find("]")]
        occur_time = check_and_format_time_str(time_str.strip())
        if not occur_time:
            time_str = NpuHistoryLogParser._extract_timestamp_from_dir(dir_name)
            occur_time = check_and_format_time_str(time_str.strip())
        return occur_time

    def _yield_log(self, file_source: Union[str, LogInfoSaver]):  # pylint: disable=arguments-differ
        """
        非 SDK（离线日志目录）场景解析 kernel.log 时：行首为开机时长且该子目录有时间参照，
        先把行首的开机时长换算为实际时间再产出，使基类解析流程的 _get_occur_time 与
        任务 [start, end] 窗口过滤基于精确时间生效；其余场景（SDK 输入、history.log 等）原样产出。
        父类 FileParser._yield_log 为静态方法，此处需访问实例状态（is_sdk_input/time_reference_map），
        因此覆盖为实例方法（self 在绑定调用中由实例提供，外部调用方式不变）。
        """
        if self.is_sdk_input or self._get_filename(file_source) != "kernel.log":
            yield from super()._yield_log(file_source)
            return
        reference = self._get_kernel_file_reference(file_source)
        for line in super()._yield_log(file_source):
            yield self._convert_line_uptime(line, reference)

    def _convert_line_uptime(self, line: str, reference: Optional[tuple]) -> str:
        """
        仅对命中 PRECHECK 关键字（KERNEL_AE_KEYWORDS）且行首为开机时长的 kernel.log 行，
        把开机时长替换为换算后的实际时间；其余行原样返回，避免逐行转换。
        :param line: log line
        :param reference: 时间参照 (W0 实际时间, T0 开机时长) 或 None
        """
        if not reference or not self._is_ae_keyword_line(line):
            return line
        time_str = line[line.find("[") + 1 : line.find("]")].strip()
        if not self._HISI_UPTIME_PATTERN.match(time_str):
            return line
        real_time = self._convert_uptime_to_real_time(reference, float(time_str))
        return "[{}]{}".format(real_time, line[line.find("]") + 1 :])

    @staticmethod
    def _is_ae_keyword_line(line: str) -> bool:
        """
        判断行是否命中 PRECHECK 关键字（与 _match_precheck_line 的判定一致）。
        :param line: log line
        """
        return all(keyword in line for keyword in NpuHistoryLogParser.KERNEL_AE_KEYWORDS)

    def _get_kernel_file_reference(self, file_source) -> Optional[tuple]:
        """
        取 kernel.log 文件对应时间戳子目录的时间参照（(W0 实际时间, T0 开机时长)），没有则返回 None。
        :param file_source: kernel.log 文件路径
        """
        return getattr(self, "time_reference_map", {}).get(os.path.dirname(os.path.dirname(file_source)))

    @staticmethod
    def _parse_time_reference(device_os_path: str) -> Optional[tuple]:
        """
        解析 device_os.log 第一条 [ERROR] KERNEL 行，得到时间参照：同一时刻的实际时间 W0 与开机时长 T0。
        :param device_os_path: device_os.log 绝对路径
        :return: (W0 实际时间 datetime, T0 开机时长)；文件缺失、无 [ERROR] KERNEL 行或解析失败返回 None
        """
        if not device_os_path or not os.path.isfile(device_os_path):
            return None
        try:
            with safe_read_open(device_os_path, "r", encoding="UTF-8") as file_stream:
                for raw_line in file_stream:
                    line = raw_line.strip()
                    if not line.startswith(NpuHistoryLogParser._DEVICE_OS_REFERENCE_HEAD):
                        continue
                    # 第一条 [ERROR] KERNEL 行解析 W0/T0，任一失败即视为参照不可用
                    real_time_match = NpuHistoryLogParser._DEVICE_OS_REAL_TIME_PATTERN.search(line)
                    uptime_match = NpuHistoryLogParser._DEVICE_OS_UPTIME_PATTERN.search(line)
                    if not real_time_match or not uptime_match:
                        return None
                    real_time = check_and_format_time_str(real_time_match.group(1))
                    if not real_time:
                        return None
                    real_dt = datetime.strptime(real_time, "%Y-%m-%d %H:%M:%S.%f")
                    return real_dt, float(uptime_match.group(1))
        except (OSError, ValueError):
            return None
        return None

    @staticmethod
    def _convert_uptime_to_real_time(reference: tuple, uptime: float) -> str:
        """
        把 kernel.log 行首的开机时长换算为实际时间：实际时间 = W0 + (本行开机时长 - T0)。
        :param reference: _parse_time_reference 返回的时间参照 (W0 实际时间, T0 开机时长)
        :param uptime: kernel.log 行首开机时长（秒）
        :return: "%Y-%m-%d %H:%M:%S.%f" 格式的实际时间字符串
        """
        real_dt = reference[0] + timedelta(seconds=uptime - reference[1])
        return real_dt.strftime("%Y-%m-%d %H:%M:%S.%f")

    def _prepare_time_references(self, log_list: list) -> dict:
        """
        构建 {时间戳子目录: 时间参照} 映射，供 kernel.log 开机时长换算实际时间使用。
        :param log_list: hisi_logs_path 全量文件列表（含 kernel.log / history.log / os_info.txt 等）
        :return: {子目录绝对路径: (W0 实际时间, T0 开机时长) 或 None}
        """
        references = {}
        for file_source in log_list:
            log_path = file_source if isinstance(file_source, str) else getattr(file_source, "path", "")
            if not log_path or os.path.basename(log_path) != "kernel.log":
                continue
            # <...>/device-X/<时间戳子目录>/log/kernel.log -> 时间戳子目录
            sub_dir = os.path.dirname(os.path.dirname(log_path))
            references.setdefault(
                sub_dir, self._parse_time_reference(os.path.join(sub_dir, self._DEVICE_OS_LOG_REL_PATH))
            )
        return references

    @staticmethod
    def _extract_timestamp_from_dir(path: str) -> Optional[str]:
        """
        Extract the timestamp from log dir path.
        path e.g: /hisi_logs/device-0/yyyymmddhhmmss-xxxxxxxxx/log/kernel.log
        :param path: dir path of log
        :return: timestamp string
        """
        p = Path(path)
        parts = p.parts
        try:
            idx = parts.index("hisi_logs")
        except ValueError:
            return ""
        for segment in parts[idx + 1 :]:
            match = TS_PATTERN.match(segment)
            if match:
                return match.group(1)
        return ""

    def parse(self, parse_ctx: KGParseCtx, task_id):
        """
        Parse hisi logs file, contain history.log
        :param parse_ctx: knowledge graph parser context
        :param task_id: the task unique id
        :return: hisi logs parse result
        """
        self.start_time = self.params.get("start_time")
        self.end_time = self.params.get("end_time")
        self.resuming_training_time = parse_ctx.resuming_training_time
        self.is_sdk_input = parse_ctx.is_sdk_input
        # 非 SDK（离线日志目录）场景：预构建 device_os.log 时间参照，供 kernel.log 开机时长换算实际时间；
        # SDK 场景无本地目录，沿用原逻辑
        parse_file_path = self.find_log(parse_ctx.parse_file_path)
        if not self.is_sdk_input:
            self.time_reference_map = self._prepare_time_references(parse_file_path)
        return self.process_parse_file_list(parse_file_path, task_id)

    def _match_precheck_line(self, log_line: str) -> dict:
        """
        Match the UDMA AE line in hisi kernel.log and generate a PRECHECK event.
        Line sample: "... UDMA ... ae ... event type is 2 ... sub type is 3 ..."
        :param log_line: single log line
        :return: PRECHECK event dict or {}
        """
        if not all(keyword in log_line for keyword in self.KERNEL_AE_KEYWORDS):
            return {}
        return {
            "event_code": PRECHECK_KERNEL_AE_TYPE23,
            "key_info": log_line,
            "attribute": {},
            "is_custom_event": False,
        }


class NpuSlogParser(BaseNpuLogParser):
    @staticmethod
    def _get_occur_time(line, dir_name):
        """
        Get the time info.
        e.g:
        [****] ****:yyyy-mm-dd-hh:mm:ss.***.*** *******
        [****] ****: yyyy-mm-dd-hh:mm:ss.***.*** ******* (There's an extra space after the colon.)
        :param line: log line
        :return: time info
        """
        time_str = line[line.find(":") + 1 :].strip()
        time_str = time_str.split(" ")[0]
        return check_and_format_time_str(time_str.strip())

    def parse(self, parse_ctx: KGParseCtx, task_id):
        """
        Parse slog file
        :param parse_ctx: knowledge graph parser context
        :param task_id: the task unique id
        :return: slog parse result
        """
        self.start_time = self.params.get("start_time")
        self.end_time = self.params.get("end_time")
        self.resuming_training_time = parse_ctx.resuming_training_time
        self.is_sdk_input = parse_ctx.is_sdk_input
        file_source_list = []
        slog_dict = self.find_log(parse_ctx.parse_file_path)
        if not self.is_sdk_input:
            slog_dict = self._filter_slog_by_time(slog_dict)
        for log_dir, log_list in slog_dict.items():
            for file_source in log_list:
                if isinstance(file_source, str):
                    file_source_list.append(os.path.join(log_dir, file_source))
                elif isinstance(file_source, LogInfoSaver):
                    file_source_list.append(file_source)
        return self.process_parse_file_list(file_source_list, task_id)

    def _filter_slog_by_time(self, slog_dict: dict):
        """
        Filter the slog dict by time
        Two log format:
        1) slog/dev-os-*/(debug/run)/device-os/device-os_***.log
        2) slog/dev-os-*/device-*/device-*_***.log
        :param slog_dict: slog dict
        :return: new slog dict
        """
        new_slog_dict = dict()
        if not self.start_time or not self.end_time:  # not get the train time interval
            for log_dir, log_list in slog_dict.items():
                new_slog_dict.update({log_dir: log_list[-2:]})  # use latest 2 files
            return new_slog_dict
        for log_dir, log_list in slog_dict.items():
            # file name timestamp use 17 numbers, start time and end time have 20 numbers, so need to cut
            log_format = '{}{}'.format(os.path.basename(log_dir), "_{:17}.log")
            new_log_list = self._filter_log_list(log_list, self.start_time, self.end_time, log_format)
            new_slog_dict.update({log_dir: new_log_list[-2:]})  # use latest 2 files
        return new_slog_dict


class NpuOsLogParser(NpuSlogParser):
    TARGET_FILE_PATTERNS = "slog_path"
    SOURCE_FILE = NPU_OS_SOURCE


class NpuDeviceLogParser(NpuSlogParser):
    TARGET_FILE_PATTERNS = "slog_path"
    SOURCE_FILE = NPU_DEVICE_SOURCE


class NpuHostLogParser(NpuSlogParser):
    TARGET_FILE_PATTERNS = "slog_host_path"
    SOURCE_FILE = NPU_HOST_SOURCE

    @staticmethod
    def _get_occur_time(line, dir_name):
        """
        Get the time info form line.
        line e.g:
        [2025-04-16-03:08:13.818489] [ascend] [drv_pcie] xxx
        :param line: log line
        :return: time info
        """
        time_str = line.split(']')[0].lstrip('[')
        return check_and_format_time_str(time_str.strip())

    def parse(self, parse_ctx: KGParseCtx, task_id):
        """
        Parse host logs file, contain host_kernel.log
        :param parse_ctx: knowledge graph parser context
        :param task_id: the task unique id
        :return: host logs parse result
        """
        self.start_time = self.params.get("start_time")
        self.end_time = self.params.get("end_time")
        self.resuming_training_time = parse_ctx.resuming_training_time
        self.is_sdk_input = parse_ctx.is_sdk_input
        return self.process_parse_file_list(self.find_log(parse_ctx.parse_file_path), task_id)


class UbctlLogParser(BaseNpuLogParser):
    TARGET_FILE_PATTERNS = "ubctl_log_path"
    SOURCE_FILE = NPU_UBCTL_SOURCE

    # 目录名正则：ub_info/dev-os-x/ubctl/2026_08_23_13_22_44/ubctl_log.txt
    UBCTL_TIMESTAMP_DIR_PATTERN = re.compile(r"^(\d{4})_(\d{2})_(\d{2})_(\d{2})_(\d{2})_(\d{2})$")

    # 数据块划分行正则：日志打印整行带双引号，参数顺序固定 -c -> -d -> -m -> -p（引号可选，兼容测试直接写字面）
    #   "Command: ubctl -c 0 -d 1 -m dump -p 0"
    #   -d 后 udie(0/1)；-m 后采集模式；-p 后 port(0~8)
    UBCTL_COMMAND_PATTERN = re.compile(
        r'^"?Command:\s*ubctl\s+-c\s+\S+\s+-d\s+(?P<udie>[01])\s+-m\s+(?P<mode>\S+)\s+-p\s+(?P<port>\d+)"?$'
    )

    # 辅助解析记录丢包数据过程处理的类：每个 udie 的每个 port 都有一份全部指标
    class _CalcPkg:
        def __init__(self):
            # udie0/udie1 -> port(0~8) -> {指标: 原始16进制值字符串}，全量采集不过滤，重复则覆盖
            # port 是否有数据取决于该 udie 实际使用的端口（部分 udie 并非 0~8 全用）
            self.udie0 = {port: dict() for port in range(UDIE_MAX_PORT_NUM)}
            self.udie1 = {port: dict() for port in range(UDIE_MAX_PORT_NUM)}
            # 当前 Command 数据块归属的 udie/port（-d/-p 精确定位），未进入合法数据块时为 None
            self.current_udie = None
            self.current_port = None
            # 出现 -m 非 dump（新类型采集数据）后置位，整个文件不再继续读
            self.parse_finished = False

    def __init__(self, params: dict):
        super().__init__(params)
        self.calc_pkg = None

    def parse(self, parse_ctx: KGParseCtx, task_id):
        #  1. 对齐 hisi/slog：标准三行上下文初始化（必须）
        self.start_time = self.params.get("start_time")
        self.end_time = self.params.get("end_time")
        self.resuming_training_time = parse_ctx.resuming_training_time or ""
        self.is_sdk_input = parse_ctx.is_sdk_input

        #  2. 取 all ubctl_log.txt 文件路径（DevLogSaver 已按 before/after 槽组装好路径；
        #     before/after 都空时走 legacy 单字段兜底）
        before_roots = list(parse_ctx.parse_file_path.ubctl_before_files or [])
        after_roots = list(parse_ctx.parse_file_path.ubctl_after_files or [])
        if not before_roots and not after_roots:
            # legacy 兜底：走 TARGET_FILE_PATTERNS=ubctl_log_path 单字段（filter_log收的 ubctl_log.txt）
            legacy_files = self.find_log(parse_ctx.parse_file_path) or []
            after_roots = legacy_files
        if not before_roots and not after_roots:
            kg_logger.info("No %s files matched, skip.", self.SOURCE_FILE)
            return [], {}

        kg_logger.info(
            "%s parse start (files: before=%s, after=%s, start=%s, end=%s).",
            self.SOURCE_FILE,
            len(before_roots),
            len(after_roots),
            self.start_time,
            self.end_time,
        )

        #  3. 目录级时间戳过滤：选 before=任务前最后一个；after=任务后第一个
        target_before = self._pick_ubctl_file(before_roots, side=SIDE_BEFORE)  # {dev_id: ubctl_log_path}
        target_after = self._pick_ubctl_file(after_roots, side=SIDE_AFTER)
        kg_logger.info(
            "%s directory pick done. before devs=%s, after devs=%s.",
            self.SOURCE_FILE,
            list(target_before.keys()),
            list(target_after.keys()),
        )

        collect_before_result = self._parse_ubctl_content(target_before, side=SIDE_BEFORE, task_id=task_id)
        collect_after_result = self._parse_ubctl_content(target_after, side=SIDE_AFTER, task_id=task_id)

        events = self._assemble_precheck_events(parse_ctx, collect_before_result, collect_after_result)

        kg_logger.info("%s parse complete, user emit %s PRECHECK events.", self.SOURCE_FILE, len(events))
        return events, {}

    def _pick_ubctl_file(self, ubctl_in_list, side: str):
        """
        入参 ubctl_in_list：list[ubctl_log.txt 绝对路径]
          路径组装已在 DevLogSaver 完成（对齐 hisi_logs_list/slog_dict 的"路径初始化在 saver"）：
            ub_info/<dev-os-X>/ubctl/<时间戳目录>/ubctl_log.txt
          这里只做内存筛选，不再扫目录：
            dev 分组（从路径提取 dev_id）→ resuming/start/end 时间过滤 → 每 dev_id 选 1 份
        返回：{dev_id(str): (ubctl_log.txt 绝对路径(str), occur_time(str))}
          occur_time = 上级时间戳目录名（采集时刻），eg "2026_08_23_13_20_00" → "2026-08-23 13:20:00"
        """
        if not ubctl_in_list:
            return {}
        # List[(timestamp_dt: datetime, timestamp_str, ubctl_log_txt_path, dev_id)]
        timestamp_pairs = []
        for f in ubctl_in_list:
            f_path = self._filename_safe(f)
            timestamp_dir_name = os.path.basename(os.path.dirname(f_path))
            timestamp_dt = self._parse_timestamp_dir_to_datetime(timestamp_dir_name)
            if not timestamp_dt:
                continue
            dev_id = self._determine_device_id(f) or UNKNOWN_DEVICE_ID
            timestamp_pairs.append((timestamp_dt, timestamp_dir_name, f_path, dev_id))
        timestamp_pairs.sort(key=lambda x: x[0])  # 时间升序
        if not timestamp_pairs:
            return {}
        return self._select_by_train_time(timestamp_pairs, side)

    def _select_by_train_time(self, timestamp_pairs, side: str):
        """
        参考 hisi/slog：按 start_time/end_time 过滤采集点（文件路径按时间戳目录排序），
          before=任务前最后一个（<=start_time 的最大值）
          after =任务后第一个（>=end_time 的最小值）
        返回值：{dev_id(str): (ubctl_log.txt 绝对路径(str), occur_time(str))} → 每个 dev_id 选 1 个目标文件
          occur_time 取上级时间戳目录名（采集时刻），格式 "YYYY-MM-DD HH:MM:SS"
        """
        start_dt = self._parse_iso_or_raise(self.start_time)
        end_dt = self._parse_iso_or_raise(self.end_time)
        resuming_dt = self._parse_iso_or_raise(self.resuming_training_time)
        result_per_dev: dict = {}

        # 训练重启时间过滤（断点分界）：before 数据在断点之前（<=resuming），after 数据在断点之后（>=resuming）
        if resuming_dt:
            if side == SIDE_BEFORE:
                timestamp_pairs = [p for p in timestamp_pairs if p[0] <= resuming_dt]
            else:
                timestamp_pairs = [p for p in timestamp_pairs if p[0] >= resuming_dt]

        def _assign_grouped_by_dev(candidate_list, side: str):
            # 同 dev_id 多文件命中时：before取时间最大（最近），after取时间最小（最早）
            grouped: dict = {}
            for t_dt, _, path, dev_id in candidate_list:
                grouped.setdefault(dev_id, []).append((t_dt, path))
            for dev_id, items in grouped.items():
                items.sort(key=lambda x: x[0])
                if side == SIDE_BEFORE:
                    # before场景按照升序排序，每个 dev_id 取最后一个（最近）数据
                    t_dt, path = items[-1]
                else:
                    # after场景按照升序排序，每个 dev_id 取第一个（最早）数据
                    t_dt, path = items[0]
                occur_time = t_dt.strftime("%Y-%m-%d %H:%M:%S")
                result_per_dev[dev_id] = (path, occur_time)

        # ---- 场景1：有 start_time + end_time（标准场景） ----
        if start_dt and end_dt:
            if side == SIDE_BEFORE:
                target_list = [p for p in timestamp_pairs if p[0] <= start_dt]
            else:
                target_list = [p for p in timestamp_pairs if p[0] >= end_dt]
            _assign_grouped_by_dev(target_list, side)
            return result_per_dev
        # ---- 场景2：拿不到 start/end，参考 slog 兜底取最新两个目录 ----
        # 先按 dev_id 排序，再按时间升序
        # 目的：每个 dev_id 最新时间在最后，最早时间在最前
        timestamp_pairs.sort(key=lambda x: (x[3], x[0]))
        _assign_grouped_by_dev(timestamp_pairs, side)
        return result_per_dev

    @staticmethod
    def _parse_iso_or_raise(time_str):
        """字符串转 datetime（兼容项目里 yyyy-mm-dd hh:mm:ss 格式）"""
        if not time_str:
            return None
        try:
            return datetime.strptime(time_str.strip()[:19], TIME_FORMAT)
        except ValueError:
            return None

    @staticmethod
    def _filename_safe(f):
        if isinstance(f, LogInfoSaver):
            return f.path or f.filename or ""
        return str(f)

    @staticmethod
    def _parse_timestamp_dir_to_datetime(timestamp_dir_name: str):
        """
        时间戳目录名 -> datetime（统一解析入口）。
        目录名格式：2026_08_23_13_20_00（匹配 UBCTL_TIMESTAMP_DIR_PATTERN）
        :return: datetime；目录名不匹配或非法则返回 None
        """

        m = UbctlLogParser.UBCTL_TIMESTAMP_DIR_PATTERN.match(timestamp_dir_name)
        if not m:
            return None
        try:
            return datetime(
                int(m.group(1)),
                int(m.group(2)),
                int(m.group(3)),
                int(m.group(4)),
                int(m.group(5)),
                int(m.group(6)),
            )
        except ValueError:
            return None

    def _parse_ubctl_content(self, target_files: dict, side: str, task_id: str):
        """
        解析 ubctl_log.txt 内容，按规则解析（rc_full等自定义指标）
        输入：
          target_files: dict[str, tuple]  {dev_id: (ubctl_log.txt 绝对路径, occur_time)}  目录过滤选中的文件（可能空）
          occur_time 来自 ubctl_log.txt 上级目录名时间戳（采集时刻），不是文件内首行
          side: str  before/after
        """

        # 组装每个文件的独立解析任务（多进程按文件粒度并行，sdk 场景串行）
        file_infos = []
        for dev_id, (ubctl_log_txt_path, occur_time) in target_files.items():
            file_infos.append((dev_id, ubctl_log_txt_path, occur_time, side))

        result = {}
        if self.is_sdk_input or len(file_infos) <= 1:
            for file_info in file_infos:
                dev_id, dev_result = self._parse_ubctl_single_file(file_info)
                result[dev_id] = dev_result
            return result

        multiprocess_job = MultiProcessJob("KNOWLEDGE_GRAPH", pool_size=len(file_infos), task_id=task_id)
        for idx, file_info in enumerate(file_infos):
            multiprocess_job.add_security_job(
                f"{self.SOURCE_FILE}_ID-{idx}_{self._filename_safe(file_info[1])}",
                self._parse_ubctl_single_file,
                file_info,
            )
        raw_results, _ = multiprocess_job.join_and_get_results()
        for dev_id, dev_result in raw_results.values():
            result[dev_id] = dev_result
        return result

    def _parse_ubctl_single_file(self, file_info):
        """
        单个 ubctl_log.txt 的解析任务（多进程 worker / sdk 串行共用）。
        :param file_info: (dev_id, ubctl_log.txt 绝对路径, occur_time, side)
        :return: (dev_id, {"udie0":..., "udie1":..., "source_path":..., "occur_time":...})
        """
        dev_id, ubctl_log_txt_path, occur_time, _side = file_info
        # 每个文件一个独立的采集状态机
        self.calc_pkg = UbctlLogParser._CalcPkg()
        for line in self._yield_log(ubctl_log_txt_path):
            self.feed_line(line)
            if self.calc_pkg.parse_finished:
                # -m 非 dump：采集的是新类型数据，不再继续往下读
                break
        return dev_id, {
            "udie0": self.calc_pkg.udie0,
            "udie1": self.calc_pkg.udie1,
            "source_path": ubctl_log_txt_path,
            "occur_time": occur_time,
        }

    # ====================== 以下函数为解析ubctl_log.txt中指标数据的统计辅助函数 ======================
    def feed_line(self, line: str):
        """
        上层逐行调用，每次传入一行日志；日志按 "Command: ubctl ..." 行划分数据块：
          同一块内（下一个 Command 行之前）的指标行归属该 Command 的 (udie, port)。
        -m 非 dump 表示采集新类型数据：置 parse_finished，上层停止读整个文件。
        """
        line = line.strip()
        if not line:
            return

        command_match = self.UBCTL_COMMAND_PATTERN.match(line)
        if command_match:
            if command_match.group("mode") != UBCTL_DUMP_MODE:
                # 本次读到 -m 非 dump：dump 块已读完，采集的是新类型数据，不再继续往下读
                self.calc_pkg.parse_finished = True
                return
            # 落盘数据块：以 Command 行的 -d/-p 精确定位 udie/port；port 越界则整块忽略
            udie = int(command_match.group("udie"))
            port = int(command_match.group("port"))
            if 0 <= port < UDIE_MAX_PORT_NUM:
                self.calc_pkg.current_udie = udie
                self.calc_pkg.current_port = port
            else:
                self.calc_pkg.current_udie = None
                self.calc_pkg.current_port = None
            return

        if self.calc_pkg.parse_finished or self.calc_pkg.current_port is None:
            return
        udie_data = self.calc_pkg.udie0 if self.calc_pkg.current_udie == 0 else self.calc_pkg.udie1
        port = self.calc_pkg.current_port
        # 通用提取：任一指标行，全量采集原始16进制值，同 udie 同 port 同指标覆盖赋值
        for key in ALL_UBCTL_KEYS:
            if line.startswith(f"{key}:"):
                _, v_str = line.split(":", maxsplit=1)
                udie_data[port][key] = v_str.strip()
                break

    def _assemble_precheck_events(self, parse_ctx: KGParseCtx, collect_before_result: dict, collect_after_result: dict):
        events_collector = []

        # 按 dev_id 合并 before/after 的 udie0/udie1 原始数据树，emit 一个 PRECHECK_ 事件
        all_dev_ids = set(collect_before_result) | set(collect_after_result)
        for dev_id in all_dev_ids:
            before = collect_before_result.get(dev_id, {})
            after = collect_after_result.get(dev_id, {})

            # 树的形状：{"udie0": {port: {指标: 原始hex}}, "udie1": {...}}
            before_tree = {"udie0": before.get("udie0", {}), "udie1": before.get("udie1", {})}
            after_tree = {"udie0": after.get("udie0", {}), "udie1": after.get("udie1", {})}

            # 都取 after的，默认都会在after里面
            occur_time = after.get("occur_time") or ""
            source_file = "\n".join([before.get("source_path") or "", after.get("source_path") or ""]).strip()
            event_dict = {
                "event_code": PRECHECK_UBCTL_DATA,  # 后续 diag 阶段按此码取 attribute 判断
                "key_info": f"ubctl before/after raw data for dev {dev_id}",
                "source_file": source_file,
                "source_device": dev_id,
                "occur_time": occur_time,
                "is_custom_event": False,
                "type": self.SOURCE_FILE,
                "occurrence": [],
                "attribute": {
                    "before": before_tree,
                    "after": after_tree,
                    "source_lines": {"before": [], "after": []},
                },
            }
            events_collector.append(event_dict)

        return events_collector
