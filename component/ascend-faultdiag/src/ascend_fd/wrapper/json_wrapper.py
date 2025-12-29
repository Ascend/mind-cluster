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
import json
import logging

from ascend_fd.pkg.diag.message import NoteMsg
from ascend_fd.pkg.diag.root_cluster.fault_description import FaultDescription
from ascend_fd.utils.i18n import get_label_for_language
from ascend_fd.utils.tool import get_version, get_build_time

logger = logging.getLogger("FAULT_DIAG")


class JsonWrapper:
    """
    This class is used to format the output diagnostic results and save them locally
    """
    ERR_RESULT = {"analyze_success": False}
    LIST_ATTR_KEYS = {"error_case", "fixed_case"}

    def __init__(self, result, failed_details, performance_flag, task_id, single_diag_flag):
        """
        Json result Wrapper
        :param result: diag result
        :param performance_flag: whether the training task contains performance detection
        :param task_id: the job task id
        """
        self.result = result
        self.failed_details = failed_details
        self.performance_flag = performance_flag
        self.single_diag_flag = single_diag_flag
        self.json = {"Version": get_version(), "Build_Time": get_build_time(), "Task_id": task_id}

    @staticmethod
    def parse_note_msgs(note_msgs):
        """
        Convert note msgs into dict
        :param note_msgs: the note_msgs class or list
        :return: note msg dict, contain zh and en str.
        """
        if isinstance(note_msgs, NoteMsg):
            note_msgs = [note_msgs]
        note = ""
        if len(note_msgs) == 1:
            note_msg = note_msgs[0]
            if isinstance(note_msg, dict):
                note_msg = NoteMsg.from_dict(note_msg)
            return {"note": note_msg.note}
        for ind, msg in enumerate(note_msgs):
            if not isinstance(msg, NoteMsg):
                msg = NoteMsg.from_dict(msg)
            note += f"{ind + 1}. {msg.note}\n"
        return {"note": note.rstrip()}

    def format_json(self):
        """
        Format rc and kg result to json
        """
        
        if not self.single_diag_flag:
            rc_result = self.result.get("Rc", self._format_err_result("ROOT_CLUSTER"))
            self.json.update({"Root_Cluster": self.format_rc_result(rc_result)})
        kg_result = self.result.get("Kg", self._format_err_result("KNOWLEDGE_GRAPH"))
        self.json.update({"Knowledge_Graph": self._format_result(kg_result)})
        if not self.performance_flag:
            return
        node_result = self.result.get("Node", self._format_err_result("NODE_ANOMALY"))
        self.json.update({"Node_Anomaly": self._format_result(node_result)})
        net_result = self.result.get("Net", self._format_err_result("NET_CONGESTION"))
        self.json.update({"Net_Congestion": self._format_result(net_result)})

    def export_rc_sdk_results(self):
        rc_result = self.result.get("Rc", self._format_err_result("ROOT_CLUSTER"))
        if not rc_result:
            return self._format_err_result("ROOT_CLUSTER")
        return self.format_rc_result(rc_result, is_sdk_output=True)

    def export_kg_sdk_results(self):
        kg_result = self.result.get("Kg", self._format_err_result("KNOWLEDGE_GRAPH"))
        if not kg_result:
            return self._format_err_result("KNOWLEDGE_GRAPH")
        return self._format_result(self.result)

    def get_format_json(self):
        """
        Get the format json result
        :return: json result
        """
        return json.dumps(self.json, ensure_ascii=False, indent=4)

    def format_rc_result(self, result: dict, is_sdk_output: bool = False):
        """
        Format rc result all content to dict
        :param result: rc result
        :param is_sdk_output: if is sdk output requirements, skip first/last error device
        :return: rc all info dict
        """
        rc_result = dict()
        sdk_field_whitelist = ["analyze_success", "fault_description", "root_cause_device", "device_link",
                               "remote_link", "first_error_device", "last_error_device"]
        for key, value in result.items():
            if is_sdk_output and key not in sdk_field_whitelist:
                continue
            if key == "fault_description":
                fault_description = FaultDescription.from_dict(value)
                rc_result.update({key: {"code": fault_description.code, "string": fault_description.string}})
                continue
            if key == "root_cause_device":
                rc_result.update({key: [str(rank) for rank in value]})
                continue
            if key == "note_msgs":
                rc_result.update(self.parse_note_msgs(value))
                continue
            if key == "fault_filter_time":
                continue
            if key in ["first_error_device", "last_error_device"]:
                rc_result.update({key: f"{value}: {value.err_time}" if value else value})
                continue
            rc_result.update({key: value})
        return rc_result

    def _format_result(self, result: dict) -> dict:
        """
        Format kg, node or net result all content to dict
        :param result: origin result
        :return: new result dict
        """
        new_result = dict()
        for key, value in result.items():
            if key == "fault":
                new_result.update({"fault": self._format_fault(value)})
                continue
            if key == "note_msgs":
                new_result.update(self.parse_note_msgs(value))
                continue
            new_result.update({key: value})
        return new_result

    def _format_fault(self, fault_list: list) -> list:
        """
        Format fault attribute and get kg err flag
        :param fault_list: fault list
        :return: new fault list
        """
        new_fault_list = []
        for fault in fault_list:
            new_fault_list.append(self._format_list_attr(fault))
        return new_fault_list

    def _format_list_attr(self, fault_attrs: dict) -> dict:
        """
        Format the attributes in LIST_ATTR_KEYS in 'fault' field to string format
        :param fault_attrs: the fault attributes of single code
        :return: the fault attributes after format
        """
        for key in self.LIST_ATTR_KEYS:
            attr_value = fault_attrs.get(key)
            if attr_value and isinstance(attr_value, list):
                fault_attrs[key] = "\n".join(attr_value)
        return fault_attrs

    def _format_err_result(self, indicator_name):
        lb = get_label_for_language()
        err_msg = self.failed_details.get(indicator_name, f"{lb.please_check_the_log}")
        return {**self.ERR_RESULT, "err_msgs": str(err_msg)}
