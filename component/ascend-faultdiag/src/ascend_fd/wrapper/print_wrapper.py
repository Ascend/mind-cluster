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
import re
import logging
from typing import List

from ascend_fd.configuration.config import COMPONENT
from ascend_fd.pkg.diag.root_cluster.fault_description import FaultDescription
from ascend_fd.utils.fault_code import NORMAL_CODE_LIST, NODE_AND_NETWORK_CODE_LIST
from ascend_fd.pkg.diag.knowledge_graph.kg_diag_job import MAX_WORKER_CHAIN_NUM
from ascend_fd.pkg.diag.message import (NoteMsg, MAX_RANK_NOTE_MSG, MAX_DEVICE_NOTE_MSG, MAX_WORKER_CHAINS_NOTE_MSG,
                                        REMOTE_LINKS_NOTE, REMOTE_LINKS_MAX_NOTE)
from ascend_fd.utils.regular_table import PLOG_PARSED, VERSION_INFO_LABEL_LIST, SHOW_LABEL_LIST
from ascend_fd.utils.tool import get_version
from ascend_fd.wrapper.pretty_table import PrettyTable, Style
from ascend_fd.utils.i18n import get_label_for_language, LANG

logger = logging.getLogger("FAULT_DIAG")
lb = get_label_for_language()
REQUIRE_LEN = 80
DEFAULT_LEN = 200
BASE_INFO_LEN = 32 if LANG == "zh" else 64  # width of the first two columns
MAX_WORKER_PRINT = 16
MAX_CQE_LINK_PRINT = 8
HALF_MAX_REMOTE_LINK_PRINT = 8
FAULT_CODE = "code"
FAULT_EVENT = "fault"
SPLIT_SEP = " -> "
MAX_FAULT_EVENT_INDEX = 2
ROOT_CAUSE_EVENT_INDEX = 0


class PrintWrapper:
    """
    This class is used to format the output diagnostic results and print them in the terminal in tabular form
    """
    SEP = "--------"
    ERR_RESULT = {"analyze_success": False}
    ATTRIBUTE_MAP = {
        "fault_source": lb.fault_source, f"cause_{LANG}": lb.cause, f"description_{LANG}": lb.fault_description,
        "error_case": lb.error_case, "fixed_case": lb.fixed_case, f"suggestion_{LANG}": lb.suggestion
    }
    SIMPLE_ATTRIBUTE_MAP = {
        "fault_source": lb.fault_source, f"cause_{LANG}": lb.cause, f"description_{LANG}": lb.fault_description
    }
    LIST_ATTR_KEYS = {lb.suggestion, lb.error_case, lb.fixed_case}

    def __init__(self, result, failed_details, performance_flag, single_diag_flag):
        """
        Table result Wrapper
        :param result: diag result
        :param performance_flag: whether the training task contains performance detection
        """
        self.version_info = get_version()
        self.result = result
        self.failed_details = failed_details
        self.performance_flag = performance_flag
        self.single_diag_flag = single_diag_flag
        self.table = PrettyTable()

        try:
            terminal_width = os.get_terminal_size().columns
        except OSError:
            # OS has no way of getting rows and columns of a Simulated Terminal
            terminal_width = DEFAULT_LEN

        self.max_len = terminal_width - BASE_INFO_LEN if terminal_width > REQUIRE_LEN else REQUIRE_LEN - BASE_INFO_LEN
        self.table.title = f"{COMPONENT.capitalize()} Fault-Diag Report"
        self.table.field_names = [lb.version_info, lb.label_type, lb.version]
        style = Style(
            align={"l": {lb.version}},
            max_width={lb.version: self.max_len}
        )
        self.table.style = style
        self.add_result_rows()

    @staticmethod
    def _parse_note_msgs(note_msgs):
        """
        Convert note msgs into Chinese str
        :param note_msgs: the note_msgs class or list
        :return: Chinese note msg str.
        """
        if isinstance(note_msgs, NoteMsg):
            note_msgs = [note_msgs]
        note = ""
        if len(note_msgs) == 1:
            note_msg = note_msgs[0]
            if isinstance(note_msg, dict):
                note_msg = NoteMsg.from_dict(note_msg)
            return note_msg.note
        for ind, msg in enumerate(note_msgs):
            if not isinstance(msg, NoteMsg):
                msg = NoteMsg.from_dict(msg)
            note += f"{ind + 1}. {msg.note}\n"
        return note.rstrip()

    @staticmethod
    def _parse_remote_link(remote_links: str):
        if not remote_links:
            return remote_links, []

        notes = [REMOTE_LINKS_NOTE]
        devices = remote_links.split(SPLIT_SEP)
        if len(devices) <= HALF_MAX_REMOTE_LINK_PRINT * 2:
            return remote_links, notes

        notes.append(REMOTE_LINKS_MAX_NOTE)
        left_remote_links = SPLIT_SEP.join(devices[:HALF_MAX_REMOTE_LINK_PRINT])
        right_remote_links = SPLIT_SEP.join(devices[-HALF_MAX_REMOTE_LINK_PRINT:])
        return ' -> ... -> '.join([left_remote_links, right_remote_links]), notes

    @staticmethod
    def _long_str_format(long_str: str) -> str:
        """
        Format the long str
        :param long_str: the origin long str
        :return: new str
        """
        new_str = long_str.expandtabs(tabsize=4)
        return repr(new_str).strip("'\"")

    def add_result_rows(self):
        """
        Add rc and kg result rows
        """
        description_str = lb.description
        indicator_name_str = lb.label_type
        kg_result = self.result.get("Kg", self.ERR_RESULT)
        self._add_version_info(kg_result)
        if not self.single_diag_flag:
            rc_result = self.result.get("Rc", self.ERR_RESULT)
            self.table.add_row(self._format_rows(lb.root_cluster_analysis, indicator_name_str, description_str),
                               divider=True)
            self.add_rc_rows(rc_result)
        self.table.add_row(self._format_rows(lb.knowledge_graph_analysis, indicator_name_str, description_str),
                           divider=True)
        self._add_result_rows(kg_result, "KNOWLEDGE_GRAPH")
        if not self.performance_flag:
            return
        node_result = self.result.get("Node", self.ERR_RESULT)
        self.table.add_row(self._format_rows(lb.node_anomaly_analysis, indicator_name_str, description_str),
                           divider=True)
        self._add_result_rows(node_result, "NODE_ANOMALY")
        net_result = self.result.get("Net", self.ERR_RESULT)
        self.table.add_row(self._format_rows(lb.net_congestion_analysis, indicator_name_str, description_str),
                           divider=True)
        self._add_result_rows(net_result, "NET_CONGESTION")

    def get_format_table(self):
        """
        Get tabular information str
        :return: result table str
        """
        return self.table.get_string()

    def add_rc_rows(self, result):
        """
        Add rc result rows. Contain root ranks, root workers, error content and note msg
        :param result: rc result dict
        """
        rc_rows = []
        if not result.get("analyze_success", False):
            rc_rows.append(
                self._format_rows("", lb.analysis_failed,
                                  self.failed_details.get("ROOT_CLUSTER", lb.please_check_the_log)))
            self._add_paragraph(rc_rows)
            return
        root_cause_device = result.get("root_cause_device")
        note_msgs = result.get("note_msgs", [])
        if len(root_cause_device) > MAX_WORKER_PRINT:
            root_cause_device = root_cause_device[:MAX_WORKER_PRINT]
            root_cause_device.append("...")  # 超过MAX_WORKER_PRINT个时加省略号
            note_msgs.append(MAX_RANK_NOTE_MSG)
        remote_link_str, remote_note_list = self._parse_remote_link(result.get("remote_link"))
        if remote_note_list:
            note_msgs.extend(remote_note_list)
        if note_msgs:
            msg_rows = [self._format_rows("", lb.note, self._parse_note_msgs(note_msgs))]
            self._add_paragraph(msg_rows)
        rc_rows.append(self._format_rows("", lb.root_cause_device, root_cause_device))
        fault_description = result.get("fault_description")
        if fault_description:
            rc_rows.append(self._format_rows("", lb.case_description,
                                             FaultDescription.from_dict(fault_description).string))
        device_link = result.get("device_link")
        if device_link:
            device_link_string = '\n'.join(device_link[:MAX_CQE_LINK_PRINT])
            device_link_string += '\n...' if len(device_link) > MAX_CQE_LINK_PRINT else ''  # 超过MAX_CQE_LINK_PRINT个时加省略号
            rc_rows.append(self._format_rows("", lb.root_cause_device_chain, device_link_string))
        if remote_link_str:
            rc_rows.append(self._format_rows("", lb.remote_link_chain, remote_link_str))
        self._add_first_and_last_error_device(result, rc_rows)
        device_info = result.get("show_device_info")
        if not device_info or device_info.get("device_type") != "first_error_device":
            self._add_paragraph(rc_rows)
            return
        self._add_plog_specifications(device_info, rc_rows)

    def _add_plog_specifications(self, device_info, rc_rows):
        device = device_info.get("device")
        error_log = device_info.get("error_log")
        if device and error_log:
            rc_rows.append(self._format_rows(
                "",
                lb.plog_log,
                lb.plog_of_the_first_error_device_shown_below.format(device=device) + error_log.strip()))
        plog_file_path = device_info.get("plog_file_path")
        plog_re = re.match(PLOG_PARSED, os.path.basename(plog_file_path))
        if plog_re:
            rc_rows.append(self._format_rows(
                "",
                lb.log_specification,
                lb.refer_plog_of_corresponding_pid.format(device=device.split(' ')[0], plog_info=plog_re[1])))
        self._add_paragraph(rc_rows)

    def _add_version_info(self, kg_result):
        """
        Add a paragraph of version info to the table
        :param kg_result: the parse result of kg
        """
        result_rows = [self._format_rows("", "Fault-Diag", self.version_info)]
        if kg_result.get("analyze_success", False):
            version_dict = kg_result.get("version_info", {})
            for label, version_label in zip(SHOW_LABEL_LIST, VERSION_INFO_LABEL_LIST):
                cur_version = version_dict.get(version_label, "")
                if cur_version:
                    result_rows.append(self._format_rows("", label, cur_version))
        self._add_paragraph(result_rows)

    def _add_first_and_last_error_device(self, result, rc_rows: list):
        """
        Add first and last error device info
        :param result: rc result dict
        """
        first_error_rank = result.get("first_error_device")
        if first_error_rank:
            rc_rows.append(self._format_rows(
                "",
                lb.first_error_device,
                f"{first_error_rank}: {first_error_rank.err_time}"))
        last_error_rank = result.get("last_error_device")
        if last_error_rank:
            rc_rows.append(self._format_rows(
                "",
                lb.last_error_device,
                f"{last_error_rank}: {last_error_rank.err_time}")
            )

    def _add_result_rows(self, result: dict, indicator_name: str):
        """
        Add result rows. Contain kg, node and net
        :param result: result dict
        """
        if not result.get("analyze_success", False):
            result_rows = [
                self._format_rows("", lb.analysis_failed,
                                  self.failed_details.get(indicator_name, lb.please_check_the_log))]
            self._add_paragraph(result_rows)
            return
        add_root_device_note = any(
            len(fault.get("fault_source", "")) > MAX_WORKER_PRINT for fault in result.get(FAULT_EVENT, []) if
            (fault.get(FAULT_CODE) and fault.get(FAULT_CODE) != "NORMAL_OR_UNSUPPORTED"))  # 诊断正常时不打印该提示
        add_chains_note = False
        fault_list = result.get(FAULT_EVENT, [])
        for fault in fault_list:
            if not fault.get(FAULT_CODE):
                continue
            for chain_dict in fault.get("fault_chains", []):
                worker_list = chain_dict.get("worker", [])
                if len(worker_list) > MAX_WORKER_PRINT:
                    add_chains_note = True
                    break
        note_msgs = result.get("note_msgs", [])
        if add_root_device_note:
            note_msgs.append(MAX_DEVICE_NOTE_MSG)
        if add_chains_note:
            note_msgs.append(MAX_WORKER_CHAINS_NOTE_MSG)
        if note_msgs:
            result_rows = [self._format_rows("", lb.note, self._parse_note_msgs(note_msgs))]
            self._add_paragraph(result_rows)
        # format fault attribute from fault list.
        for index, fault in enumerate(fault_list):
            if not fault.get(FAULT_CODE):
                continue
            result_rows = self._format_fault_attr(fault, index)
            self._add_paragraph(result_rows)

    def _add_fault_details(self, fault_details: List):
        """
        Add fault details rows.
        :param fault_details: fault_details list
        :return:
        """
        node_rows = []
        sep_flag = True if len(fault_details) > 1 else False
        if sep_flag:
            node_rows.append(self._format_rows("", self.SEP, ""))
        for detail in fault_details:
            node_rows.extend(self._parse_fault_detail(detail))
            if sep_flag:
                node_rows.append(self._format_rows("", self.SEP, ""))
        return node_rows

    def _add_paragraph(self, rows):
        """
        Add a dividing line after the last line of each paragraph of information
        :param rows: the paragraph rows
        """
        if not rows:
            return
        pre_rows = rows[:-1]
        if pre_rows:
            self.table.add_rows(pre_rows)
        self.table.add_row(rows[-1], divider=True)

    def _format_fault_attr(self, fault: dict, index: int) -> list:
        """
        Format fault attribute from fault field.
        :param fault: fault code
        :param index: index
        :return: error content rows
        """
        fault_code = fault.get("code")
        # 当单机诊断时，不显示"疑似根因故障"提示
        rows = [self._format_rows(lb.suspected_root_cause_fault
                                  if (not self.single_diag_flag and index == ROOT_CAUSE_EVENT_INDEX)
                                  else "", lb.status_code, fault_code)]
        if fault_code in NORMAL_CODE_LIST:
            rows.append(self._format_rows("", lb.result_description, fault.get(f"description_{LANG}")))
            return rows
        # format fault categories to 184 faults by kg.
        if fault_code not in NODE_AND_NETWORK_CODE_LIST:
            rows.append(self._format_fault_class(fault))
        # format fault details.
        if fault.get('fault_details'):
            rows.extend(self._add_fault_details(fault.get('fault_details')))
        # format fault suggestion, description, etc.
        # 当故障事件大于3时，简略展示故障事件分析描述信息；当单机诊断时，不简略展示
        format_flag = not self.single_diag_flag and index > MAX_FAULT_EVENT_INDEX
        for key, value in (self.SIMPLE_ATTRIBUTE_MAP.items() if format_flag else self.ATTRIBUTE_MAP.items()):
            fault_value = fault.get(key)
            if not fault_value:
                continue
            if key == "fault_source" and len(fault_value) > MAX_WORKER_PRINT:
                fault_value = fault_value[:MAX_WORKER_PRINT]
                fault_value.append("...")  # 超过MAX_WORKER_PRINT个时加省略号
            this_row = self._format_rows("", value, fault_value)
            rows.append(this_row)
        # 当故障事件大于3时，省略下面故障事件分析描述信息；当单机诊断时，不省略
        if format_flag:
            return rows
        # format key log. get key log rows from the fault event key_info field
        event_attr = fault.get('event_attr')
        fault_source = fault.get('fault_source')
        if event_attr and fault_source and isinstance(fault_source, list):
            work_source = event_attr.get(fault_source[0])
            if work_source and isinstance(work_source, list) and isinstance(work_source[0], dict):
                key_info = work_source[0].get("key_info", "").strip()
                rows.append(self._format_rows("", lb.key_info, key_info))

        # format fault chains.
        fault_chains = fault.get("fault_chains")
        if fault_chains:
            rows.extend(self._format_fault_chains(fault_chains))
        return rows

    def _parse_fault_detail(self, fault_detail):
        """
        Convert fault detail into Chinese info
        :param fault_detail: fault detail dict
        :return: fault detail rows
        """
        rows = []
        key_value = {
            "worker": lb.faulty_workers, "device_list": lb.fault_source, "device": lb.faulty_device,
            "periods": lb.fault_occurrence_period, "process_id": lb.fault_process,
            "fault_period_probability": lb.fault_probability
        }
        for key, value in key_value.items():
            if not fault_detail.get(key, None):
                continue
            if key == "worker":
                worker_name = fault_detail.get(key)
                rows.append(self._format_rows("", value, worker_name))
                continue
            if key == "device":
                worker_name, npu_id = fault_detail.get(key)
                device_name = f"{worker_name} device-{npu_id}"
                rows.append(self._format_rows("", value, device_name))
                continue
            if key == "periods":
                fault_periods = [str(period) for period in fault_detail.get(key)]
                # each row print two periods
                if len(fault_periods) > 4:  # if periods list length exceeds 4, only show 3 and save all in json reports
                    fault_periods = fault_periods[:3] + [f"......          {lb.complete_fault_details_refer_to_json}"]
                row_nums = (len(fault_periods) + 1) // 2
                new_periods = [" ".join(fault_periods[i * 2: i * 2 + 2]) for i in range(row_nums)]
                rows.append(self._format_rows("", value, "\n".join(new_periods)))
                continue
            if key == "fault_period_probability":
                periods = "\n".join(map(str, fault_detail.get(key)))
                rows.append(self._format_rows("", value, periods))
                continue
            rows.append(self._format_rows("", value, fault_detail.get(key)))
        return rows

    def _format_rows(self, title_name, indicator_name, description):
        """
        Format long description str and return new row contain [title_name, indicator_name, description]
        :param title_name: title name
        :param indicator_name: indicator Name
        :param description: indicator description
        :return: row list [title_name, indicator_name, description]
        """
        if indicator_name not in self.LIST_ATTR_KEYS or not isinstance(description, list):
            description = str(description).split("\n")
        description = "\n".join(map(self._long_str_format, description))
        return [title_name, indicator_name, description]

    def _format_fault_class(self, fault_class: dict):
        """
        Format fault class info
        :param fault_class: fault class dict
        :return: row contains fault class info
        """
        unknown_str = "UNKNOWN"
        class_val = fault_class.get("class", unknown_str)
        component_val = fault_class.get("component", unknown_str)
        module_val = fault_class.get("module", unknown_str)
        format_info = f'{lb.label_type}:{class_val} {lb.component}:{component_val} {lb.module}:{module_val}'
        row = self._format_rows("", lb.fault_category, format_info)
        return row

    def _format_fault_chains(self, fault_chains: list) -> list:
        """
        Format fault chains from fault_chains field
        :param fault_chains: fault chains
        :return: row contains fault chains info
        """
        # if the max number of workers that own the faulty chain is more than 3, only the first three are displayed
        if len(fault_chains) > MAX_WORKER_CHAIN_NUM:
            fault_chains = fault_chains[:MAX_WORKER_CHAIN_NUM]

        result_rows = []
        diff_chain_num = len(fault_chains)
        for index, chains_dict in enumerate(fault_chains):
            worker_name = chains_dict.get("worker", [])
            if len(worker_name) > MAX_WORKER_PRINT:
                worker_name = worker_name[:MAX_WORKER_PRINT]
                worker_name.append("...")  # 超过MAX_WORKER_PRINT个时加省略号
            chains = chains_dict.get("chains", "")
            if index == 0:
                result_rows.append(self._format_rows("", lb.fault_propagation_chain, worker_name))
            else:
                result_rows.append(self._format_rows("", "", worker_name))
            result_rows.append(self._format_rows("", "", chains))

            # if multiple worker have different fault chains, use "" to separate them.
            if index < diff_chain_num - 1:
                result_rows.append(self._format_rows("", "", ""))
        return result_rows
