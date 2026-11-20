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
import logging
import re
from datetime import datetime

from itertools import chain
from typing import Union

from ascend_fd.model.context import KGParseCtx
from ascend_fd.pkg.parse.parser_saver import LogInfoSaver
from ascend_fd.utils.constant.dev_log_const import (
    PRECHECK_FDM_AICPU_HBM_ERROR,
    PRECHECK_FDM_DMC_MAIN_ERROR,
    PRECHECK_FDM_DMC_MIRROR_ERROR,
    PRECHECK_FDM_MATA_DDRC_MULTI_BIT_ECC,
    PRECHECK_FDM_MATA_UNCORRECTED_ERROR,
    PRECHECK_SEL_AIC_MULTIBIT_ECC,
    PRECHECK_SEL_LOG_LOST_NPU,
    PRECHECK_SEL_MULTIBIT_ECC,
)
from ascend_fd.utils.regular_table import KG_MAX_TIME
from ascend_fd.utils.tool import MultiProcessJob, check_and_format_time_str
from ascend_fd.pkg.parse.knowledge_graph.parser.file_parser import FileParser, EventStorage

kg_logger = logging.getLogger("KNOWLEDGE_GRAPH")


class BMCParser(FileParser):
    TARGET_FILE_PATTERNS = "bmc_log_path"
    SOURCE_FILE = "BMCLog"
    # 日志文件类型按后缀分发：fdm_log 走 FDM 组解析，sel.txt 走 SEL 告警分发
    FDM_LOG_FILE = "fdm_log"
    SEL_TXT_FILE = "sel.txt"
    # ---------- SEL ----------
    SEL_CONNECTION_LOST_KEYWORD = "NPU connection has been lost"
    SEL_TIME_PATTERN = re.compile(r"\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}")
    SEL_NPU_PATTERN = re.compile(r"NPU(\d+)")
    # 多bit ECC 家族：0x56000003 按 Event Code 列匹配，AIC 按 description 关键字 0x80c98008
    CODE_56000003 = "0x56000003"  # 告警编码：NPU 健康状态紧急错误
    MULTI_BIT_ECC_CODE = "0x80e01801"
    SEL_AIC_MULTIBIT_ECC_KEYWORD = "0x80c98008"
    # ---------- FDM ----------
    FDM_GROUP_START_PATTERN = re.compile(r"Error Log Number\]")
    FDM_TIME_PATTERN = re.compile(r"Time:\s*(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2})")
    FDM_SOCKET_PATTERN = re.compile(r"Socket:\s*(NPU\d+)")
    # HAM 行两种格式兼容：`HAM: 2` / `HAM: NpuBoard2`，均提取编号
    FDM_HAM_PATTERN = re.compile(r"HAM:\s*(?:NpuBoard)?(\d+)")
    FDM_MODULE_PATTERN = re.compile(r"Module:\s*(\S+)")
    # Error Type 三种不可纠正错误：UER/UEO/UEU，冒号后及括号前的空格数不确定，用正则兼容
    FDM_MATA_ERROR_TYPE_PATTERN = re.compile(r"Error Type:Uncorrected Error.*\((UER|UEO|UEU)\)", re.IGNORECASE)
    FDM_MATA_MODULE_PREFIX = "MATA"
    DMC_RDATA_MAIN_ERROR = "dmc_rdata_main_error"
    DMC_RDATA_MIRROR_ERROR = "dmc_rdata_mirror_error"
    # MATA 模块 DDRC 多比特：IERRCODE 描述含 DDRC（如 "Response Error From DDRC"）
    FDM_IERR_DDRC_PATTERN = re.compile(r"IERRCODE:.* DDRC")
    # CPU_CORE CE：三个关键句子（括号为字面括号，.* 为通配数字部分）
    FDM_CPU_CORE_MODULE_PREFIX = "CPU_CORE"
    FDM_CE_ERROR_TYPE_KEYWORD = r"Error Type:At least one error was corrected\.\(CE\)"
    FDM_SERR_ASSC_MEM_ERROR_KEYWORD = r"SERRCODE:.*\(Data error from associative memory\)"
    FDM_IERR_UNDEFINED_KEYWORD = r"IERRCODE:.*\(Undefined\)"
    FDM_FAULT_KEYWORDS = {
        DMC_RDATA_MAIN_ERROR: PRECHECK_FDM_DMC_MAIN_ERROR,
        DMC_RDATA_MIRROR_ERROR: PRECHECK_FDM_DMC_MIRROR_ERROR,
    }

    def __init__(self, params: dict):
        super().__init__(params)
        self._fdm_group_lines = []

    def parse(self, parse_ctx: KGParseCtx, task_id: str):
        """
        Parse bmc log file
        :param parse_ctx: knowledge graph parser context
        :param task_id: unique task id
        :return: parse descriptor result
        """
        file_list = self.find_log(parse_ctx.parse_file_path)
        if not file_list:
            return [], {}
        kg_logger.info("%s files parse job started.", self.SOURCE_FILE)
        self.start_time = self.params.get("start_time")
        self.end_time = self.params.get("end_time")
        multiprocess_job = MultiProcessJob("KNOWLEDGE_GRAPH", pool_size=len(file_list), task_id=task_id)
        if self.is_sdk_input:
            results = dict()
            for idx, file_source in enumerate(file_list):
                results.update(
                    {f"{self.SOURCE_FILE}_ID-{idx}_{self._get_filename(file_source)}": self._parse_file(file_source)}
                )
        else:
            for idx, file_source in enumerate(file_list):
                multiprocess_job.add_security_job(
                    f"{self.SOURCE_FILE}_ID-{idx}_{self._get_filename(file_source)}", self._parse_file, file_source
                )
            results, _ = multiprocess_job.join_and_get_results()
        kg_logger.info("%s files parse job is complete.", self.SOURCE_FILE)
        return list(chain(*results.values())), {}

    def _parse_file(self, file_source: Union[str, LogInfoSaver]):
        """
        Parse single bmc log line by line
        :param file_source: log file path
        :return: a list of event dict
        """
        event_storage = EventStorage()
        # 存储掉卡（connection lost）各 NPU 的状态表；多bit ECC / AIC 无状态判定，不入容器
        sel_events = {self.SEL_CONNECTION_LOST_KEYWORD: {}}
        for log_line in self._yield_log(file_source):
            file_path = self._get_file_path(file_source)
            if file_path.endswith(self.SEL_TXT_FILE) or file_path.endswith(self.FDM_LOG_FILE):
                self._match_precheck_line(log_line, file_source, event_storage, sel_events)
                continue
            if self.start_time or self.end_time:
                occur_time = self._filter_log_time(log_line)
            else:
                occur_time = ""
            if self.start_time and occur_time < self.start_time:
                continue
            if self.end_time and occur_time > self.end_time:
                continue
            # 其余文件逐行走通用解析
            event_dict = self.parse_single_line(log_line)
            if not event_dict:
                continue
            occur_time = occur_time or KG_MAX_TIME
            if not event_dict.get("precheck_event"):
                # temporarily assign a max time to the event as the log time format is unknown
                self.supplement_common_info(event_dict, file_source, occur_time)
            event_storage.record_event(event_dict)
        # 文件尾部 flush 最后一组 fdm 数据
        if file_path.endswith(self.FDM_LOG_FILE):
            for final_event in self._build_fdm_event(file_source):
                event_storage.record_event(final_event)
        return event_storage.generate_event_list()

    def _filter_log_time(self, log_line) -> str:
        """
        Filter log time
        :param log_line: log line
        :return: filtered log time
        """
        return ""

    def _match_precheck_line(
        self, log_line: str, file_source: str, event_storage: EventStorage, sel_events: dict = None
    ) -> dict:
        """
        Match precheck event in line
        :param log_line: log line
        :return: precheck event dict or None
        """
        pass

    def _build_fdm_event(self, file_source: Union[str, LogInfoSaver]) -> list:
        """
        Build the precheck events of the completed fdm data group. A group is
        valid only when it has occur time, socket NPU and module. Every
        matched fault keyword of the group produces its own precheck event
        (e.g. a MATA group with both a mirror IERRCODE and an uncorrected
        Error Type produces two events)
        :param file_source: fdm log file
        :return: list of precheck event dicts (possibly empty)
        """
        if not self._fdm_group_lines:
            return []
        joined = "\n".join(self._fdm_group_lines)
        time_match = self.FDM_TIME_PATTERN.search(joined)
        npu_match = self.FDM_SOCKET_PATTERN.search(joined)
        module_match = self.FDM_MODULE_PATTERN.search(joined)
        ham_match = self.FDM_HAM_PATTERN.search(joined)
        if not time_match or not npu_match or not module_match or not ham_match:
            return []
        # 设备号换算：HAM 编号 × Socket 编号 - 1（如 HAM: NpuBoard2 + Socket: NPU2 → device 3）
        npu = str(int(ham_match.group(1)) * int(npu_match.group(1)[3:]) - 1)
        module, occur_time = module_match.group(1), time_match.group(1)
        group_info = (npu, module, occur_time)
        events = []
        if module.startswith(self.FDM_MATA_MODULE_PREFIX):
            events.extend(self._build_mata_events(group_info, file_source))
        if module.startswith(self.FDM_CPU_CORE_MODULE_PREFIX):
            events.extend(self._build_cpu_core_events(group_info, file_source))
        # dmc 关键字（main/mirror）：各命中一个就产出一个事件
        events.extend(self._build_dmc_events(group_info, file_source))
        return events

    def _build_mata_events(self, group_info: tuple, file_source: Union[str, LogInfoSaver]) -> list:
        """
        Build MATA module events: IERRCODE containing DDRC -> DDRC multi-bit
        event; Error Type UER/UEO/UEU -> uncorrected error event. Both can be
        matched in one group and each produces its own event
        :param group_info: (npu, module, occur_time) parsed from the fdm data group
        :param file_source: fdm log file
        :return: list of MATA precheck event dicts (possibly empty)
        """
        events = []
        ierr_ddrc_line = self._find_ierr_ddrc_line()
        if ierr_ddrc_line:
            events.append(
                self._build_fdm_event_dict(
                    PRECHECK_FDM_MATA_DDRC_MULTI_BIT_ECC,
                    "MATA_DDRC",
                    ierr_ddrc_line,
                    group_info,
                    file_source,
                )
            )
        error_line, error_type = self._find_mata_error_type_line()
        if error_type is not None:
            events.append(
                self._build_fdm_event_dict(
                    PRECHECK_FDM_MATA_UNCORRECTED_ERROR,
                    error_type,
                    error_line,
                    group_info,
                    file_source,
                )
            )
        return events

    def _build_cpu_core_events(self, group_info: tuple, file_source: Union[str, LogInfoSaver]) -> list:
        """
        Build CPU_CORE module event: CE + SERRCODE(associative memory) +
        IERRCODE(Undefined) -> AICPU reads error data from HBM
        :param group_info: (npu, module, occur_time) parsed from the fdm data group
        :param file_source: fdm log file
        :return: list of CPU_CORE precheck event dicts (possibly empty)
        """
        cpu_core_fault_line = self._match_cpu_core_ce_error()
        if not cpu_core_fault_line:
            return []
        return [
            self._build_fdm_event_dict(
                PRECHECK_FDM_AICPU_HBM_ERROR,
                "CPU_CORE_CE",
                cpu_core_fault_line,
                group_info,
                file_source,
            )
        ]

    def _build_dmc_events(self, group_info: tuple, file_source: Union[str, LogInfoSaver]) -> list:
        """
        Build dmc keyword events: each matched main/mirror keyword produces
        its own precheck event
        :param group_info: (npu, module, occur_time) parsed from the fdm data group
        :param file_source: fdm log file
        :return: list of dmc precheck event dicts (possibly empty)
        """
        return [
            self._build_fdm_event_dict(
                self.FDM_FAULT_KEYWORDS[fault_keyword],
                fault_keyword,
                fault_line,
                group_info,
                file_source,
            )
            for fault_line, fault_keyword in self._find_fault_keyword_lines()
        ]

    def _find_mata_error_type_line(self) -> (str, str):
        """
        Find the line whose Error Type is one of the uncorrected errors
        UER / UEO / UEU, the value is used as the fault keyword of the event
        :return: (error line, error type), or (None, None) if no such line
        """
        for line in self._fdm_group_lines:
            error_type_match = self.FDM_MATA_ERROR_TYPE_PATTERN.search(line)
            if error_type_match:
                return line, error_type_match.group(1)
        return None, None

    def _find_fault_keyword_lines(self) -> list:
        """
        Find all dmc fault keywords matched by the group, each keyword maps
        to its own precheck code in FDM_FAULT_KEYWORDS and produces one event
        :return: list of (fault line, fault keyword), empty if none matched
        """
        matched = {}
        for line in self._fdm_group_lines:
            for keyword in self.FDM_FAULT_KEYWORDS:
                if keyword in line and keyword not in matched:
                    matched[keyword] = line
        return [(line, keyword) for keyword, line in matched.items()]

    def _find_ierr_ddrc_line(self) -> str:
        """
        Match the MATA DDRC multi-bit line, the IERRCODE description contains
        "DDRC" (case-insensitive)
        :return: the matched IERRCODE line or None
        """
        for line in self._fdm_group_lines:
            if self.FDM_IERR_DDRC_PATTERN.search(line):
                return line
        return None

    def _match_cpu_core_ce_error(self) -> str:
        """
        Match the CPU_CORE CE pattern by the three keyword sentences:
        1) Error Type:At least one error was corrected.(CE)
        2) SERRCODE:.*(Data error from associative memory)
        3) IERRCODE:.*(Undefined)
        :return: the three matched lines joined by newline, or None
        """
        error_type_line = serr_line = ierr_line = None
        for line in self._fdm_group_lines:
            if error_type_line is None and re.search(self.FDM_CE_ERROR_TYPE_KEYWORD, line):
                error_type_line = line
            elif serr_line is None and re.search(self.FDM_SERR_ASSC_MEM_ERROR_KEYWORD, line):
                serr_line = line
            elif ierr_line is None and re.search(self.FDM_IERR_UNDEFINED_KEYWORD, line):
                ierr_line = line
        if error_type_line and serr_line and ierr_line:
            return "\n".join([error_type_line, serr_line, ierr_line])
        return None

    def _build_fdm_event_dict(
        self,
        precheck_code: str,
        keyword: str,
        fault_line: str,
        group_info: tuple,
        file_source: Union[str, LogInfoSaver],
    ) -> dict:
        """
        Build a single precheck event dict of a matched fdm fault keyword
        :param precheck_code: precheck event code of the fault keyword
        :param keyword: matched fault keyword
        :param fault_line: raw fdm log line containing the keyword
        :param group_info: (npu, module, occur_time) parsed from the fdm data group
        :param file_source: fdm log file
        :return: precheck event dict
        """
        npu, module, occur_time = group_info
        return {
            "event_code": precheck_code,
            "fault_keyword": keyword,
            "key_info": fault_line,
            "source_file": self._get_source_file(file_source),
            "source_device": npu,
            "module": module,
            "occur_time": occur_time,
            "is_custom_event": False,
            "type": self.SOURCE_FILE,
            "precheck_event": True,  # 已自带完整字段，跳过通用 supplement
        }


class BMCAppDumpParser(BMCParser):
    TARGET_FILE_PATTERNS = "bmc_app_dump_log_path"
    SOURCE_FILE = "BMCAppDumpLog"
    TIME_FORMAT = "%Y-%m-%d %H:%M:%S"

    def _match_precheck_line(
        self, log_line: str, file_source: str, event_storage: EventStorage, sel_events: dict = None
    ) -> dict:
        """
        Match precheck event in line
        :param log_line: log line
        :return: precheck event dict or None
        """
        file_path = self._get_file_path(file_source)
        if file_path.endswith(self.FDM_LOG_FILE):
            # fdm 一组数据（Error Log Number] 开头的一组）可能同时命中多个故障
            # 关键字，每类命中各自产出一个 precheck 事件，逐个记录后直接跳过通用解析
            fdm_events = self.parse_fdm_log_line(log_line, file_source)
            for event_dict in fdm_events:
                event_storage.record_event(event_dict)
            return

        # sel.txt 仅走 sel 专用分发：掉卡 / 多bit ECC / AIC，一行可能同时命中多个
        # 告警，每类命中各自产出一个 precheck 事件；未命中的行不回落通用解析
        if file_path.endswith(self.SEL_TXT_FILE):
            sel_event_list = self.parse_sel_txt_line(log_line, file_source, sel_events)
            for event_dict in sel_event_list:
                event_storage.record_event(event_dict)
            return

    def parse_fdm_log_line(self, log_line: str, file_source: Union[str, LogInfoSaver]) -> list:
        """
        Parse fdm log line by line, each data group starts with
        "[Error Log Number]". When a new group header is met, the previous
        group is finalized and its precheck events are returned
        :param log_line: single fdm log line
        :param file_source: fdm log file
        :return: parsed, event_list, parsed is always True for fdm log,
                 event_list is the completed previous group's precheck events
        """
        fdm_events = []
        # 非一组数据开头，就结算判定这组数据
        if self.FDM_GROUP_START_PATTERN.search(log_line):
            fdm_events = self._build_fdm_event(file_source)
            self._fdm_group_lines = [log_line]
        elif log_line:
            self._fdm_group_lines.append(log_line)
        return fdm_events

    def parse_sel_txt_line(self, log_line: str, file_source: Union[str, LogInfoSaver], sel_events: dict) -> list:
        """
        Dispatch a single sel.txt line to the matched alarm parsers. Each
        matched alarm type produces its own precheck event, a line can match
        several types at once:
        - "NPU connection has been lost" → 掉卡（按 NPU 逆序状态机）
        - Event Code 0x56000003 且 description 含 0x80e01801 → 健康紧急多bit ECC
        - description 含 0x80c98008 → AIC 多bit ECC（不看 Event Code 列）
        :param log_line: single sel.txt line
        :param file_source: sel.txt log file
        :param sel_events: npu state table keyed by event keyword（仅掉卡使用）
        :return: 命中的 precheck 事件列表（未命中或行无效时可能为空）；
                 调用方对 sel.txt 一律不回落通用解析
        """
        fields = [field.strip() for field in log_line.split("|")]
        event_list = []
        # 6 列 BMC sel 格式的行才做专用告警匹配；列数不足视为 AppDump 简化
        # 行（"序号|时间|事件描述"），跳过专用匹配直接走下方通用解析
        if len(fields) >= 6:
            event_code = fields[3]
            event_description = fields[5]
            # 掉卡事件（按 description 关键字命中）
            if self.SEL_CONNECTION_LOST_KEYWORD in event_description:
                lost_event = self.parse_npu_connection_lost_line(log_line, fields, file_source, sel_events)
                if lost_event:
                    event_list.append(lost_event)
            # 健康紧急多bit ECC（按 Event Code 列 + description 错误码命中）
            if self.CODE_56000003 == event_code and self.MULTI_BIT_ECC_CODE in event_description.lower():
                ecc_event = self.parse_multibit_ecc_line(log_line, fields, file_source, PRECHECK_SEL_MULTIBIT_ECC)
                if ecc_event:
                    event_list.append(ecc_event)
            # AIC 多bit ECC（仅按 description 关键字命中）
            if self.SEL_AIC_MULTIBIT_ECC_KEYWORD in event_description.lower():
                aic_event = self.parse_multibit_ecc_line(log_line, fields, file_source, PRECHECK_SEL_AIC_MULTIBIT_ECC)
                if aic_event:
                    event_list.append(aic_event)

        # 未命中以上类型（含列数不足的 AppDump 简化行）走通用解析
        if not event_list:
            if self.start_time or self.end_time:
                occur_time = self._filter_log_time(log_line)
            else:
                occur_time = ""
            if self.start_time and occur_time < self.start_time:
                return []
            if self.end_time and occur_time > self.end_time:
                return []
            event_dict = self.parse_single_line(log_line)
            if not event_dict.get("precheck_event"):
                occur_time = occur_time or KG_MAX_TIME
                self.supplement_common_info(event_dict, file_source, occur_time)
                event_list.append(event_dict)

        return event_list

    def parse_npu_connection_lost_line(
        self, log_line: str, fields: list, file_source: Union[str, LogInfoSaver], sel_events: dict
    ) -> dict:
        """
        Parse "NPU connection has been lost" line
        :param log_line: single sel.txt line
        :param fields: sel.txt line split by "|" and stripped
        :param file_source: sel.txt log file
        :param sel_events: dict keyed by event keyword, npu state of each event type
        :return: event dict or None
        """
        generation_time, status, description = fields[1], fields[4], fields[5]
        time_match = self.SEL_TIME_PATTERN.search(generation_time)
        npu_match = self.SEL_NPU_PATTERN.search(description)
        if not time_match or not npu_match:
            return None
        npu = str(int(npu_match.group(1)) - 1)  # sel.txt 中 NPU 编号 1-8，设备号 0-7 需减一
        npu_state = sel_events[self.SEL_CONNECTION_LOST_KEYWORD]
        if not self._should_keep_sel_state(npu_state, npu, status):
            return None
        return self._build_sel_event(PRECHECK_SEL_LOG_LOST_NPU, log_line, npu, time_match.group(0), file_source)

    def parse_multibit_ecc_line(
        self,
        log_line: str,
        fields: list,
        file_source: Union[str, LogInfoSaver],
        precheck_code: str,
    ) -> dict:
        """
        Parse an Asserted multi-bit ECC sel.txt line, e.g. NPU health critical
        (Event Code 0x56000003, Error Code 0x80e01801) or AIC (description
        contains 0x80c98008). Every Asserted occurrence is a problem and is
        recorded as a precheck event, other statuses are ignored. The event
        code / description keyword is validated by the dispatcher
        :param log_line: single sel.txt line
        :param fields: sel.txt line split by "|" and stripped
        :param file_source: sel.txt log file
        :param precheck_code: precheck event code of the matched alarm type
        :return: precheck event dict or None
        """
        generation_time, status, description = fields[1], fields[4], fields[5]
        if status != "Asserted":
            return None
        time_match = self.SEL_TIME_PATTERN.search(generation_time)
        npu_match = self.SEL_NPU_PATTERN.search(description)
        if not time_match or not npu_match:
            return None
        npu = str(int(npu_match.group(1)) - 1)  # sel.txt 中 NPU 编号 1-8，设备号 0-7 需减一
        return self._build_sel_event(precheck_code, log_line, npu, time_match.group(0), file_source)

    def _should_keep_sel_state(self, npu_state: dict, npu: str, status: str) -> bool:
        """
        State machine of the reverse ordered sel.txt: the newest line decides the
        fate, and the first Deasserted (recovery boundary) stops the recording
        :param npu_state: npu state table of one sel event type,
                         npu -> True(all recording stopped)/False(still in fault)
        :param npu: npu name
        :param status: Asserted / Deasserted
        :return: True if the line should be recorded into occurrence
        """
        recovered = True
        in_fault = False
        is_asserted = status == "Asserted"
        # 首次遇到该 npu：最新行若是 Deasserted 说明它已恢复，无需记录；是 Asserted 则视为仍在故障中
        npu_state.setdefault(npu, in_fault if is_asserted else recovered)
        if npu_state[npu] == recovered:
            return False  # 已遇到恢复边界，丢弃之后更旧的行
        if not is_asserted:
            npu_state[npu] = recovered  # 故障期间遇到 Deasserted：这是恢复边界，停止后续记录
            return False
        return True  # 故障期间的 Asserted：记录该行

    def _build_sel_event(
        self, precheck_code: str, log_line: str, npu: str, occur_time: str, file_source: Union[str, LogInfoSaver]
    ) -> dict:
        """
        Build the precheck event of a sel.txt keyword alarm
        :param precheck_code: fixed precheck event code of the alarm type
        :param log_line: raw sel.txt line
        :param npu: npu name
        :param occur_time: event occur time
        :param file_source: sel.txt log file
        :return: precheck event dict
        """
        return {
            "event_code": precheck_code,  # 固定码，设备维度由 source_device 区分
            "key_info": log_line,
            "source_file": self._get_source_file(file_source),
            "source_device": npu,
            "occur_time": occur_time,
            "is_custom_event": False,
            "type": self.SOURCE_FILE,
            "precheck_event": True,  # 已自带完整字段，跳过通用 supplement
        }

    def _filter_log_time(self, log_line) -> str:
        """
        Filter log time
        :param log_line: log line
        :return: filtered log time
        """
        log_parts = [p.strip() for p in log_line.split('|')]
        if len(log_parts) <= 1:
            return ""
        try:
            time_obj = datetime.strptime(log_parts[1], self.TIME_FORMAT)
            return str(time_obj) + ".000000"
        except ValueError:
            return ""


class BMCDeviceDumpParser(BMCParser):
    TARGET_FILE_PATTERNS = "bmc_device_dump_log_path"
    SOURCE_FILE = "BMCDeviceDumpLog"


class BMCLogDumpParser(BMCParser):
    TARGET_FILE_PATTERNS = "bmc_log_dump_log_path"
    SOURCE_FILE = "BMCLogDumpLog"
    TIME_REGEX = re.compile(r"^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{6}\+\d{2}:\d{2}")
    TIME_FORMAT = "%Y-%m-%dT%H:%M:%S.%f%z"

    def _filter_log_time(self, log_line) -> str:
        """
        Filter log time
        :param log_line: log line
        :return: filtered log time
        """
        find_time_ret = self.TIME_REGEX.findall(log_line)
        if not find_time_ret:
            return ""
        try:
            time_obj = datetime.strptime(find_time_ret[0], self.TIME_FORMAT)
            return check_and_format_time_str(str(time_obj))
        except ValueError:
            return ""
