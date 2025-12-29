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

from ascend_fd.pkg.diag.knowledge_graph.kg_engine.graph.graph_builder import GraphBuilder
from ascend_fd.utils.load_kg_config import Schema
from ascend_fd.pkg.diag.knowledge_graph.kg_engine.model.package_data import PackageData
from ascend_fd.pkg.diag.knowledge_graph.kg_engine.model.response import Response
from ascend_fd.utils.tool import safe_write_open, get_version, get_build_time, merge_occurrence, \
    safe_generate_or_merge_json_file
from ascend_fd.configuration.config import DEFAULT_USER_CONF
from ascend_fd.utils.i18n import LANG

kg_logger = logging.getLogger("KNOWLEDGE_GRAPH")


class DataDescriptor:
    """
    Data Descriptor
    """
    VERSION_INFO = "VERSION_INFO"

    def __init__(self):
        self.data = dict()
        self.version_info = dict()
        self.devices = {"Unknown"}
        self.files_parse_info = None

    @staticmethod
    def write_to_json_file(file_path: str, save_data: dict):
        """
        Writes data of the dictionary type to a JSON file
        :param file_path: file save path
        :param save_data: data in dictionary format
        """
        with safe_write_open(file_path, mode="w+", encoding="utf-8") as f_dump:
            f_dump.write(json.dumps(save_data, sort_keys=False, separators=(',', ':'), ensure_ascii=False))
            f_dump.write('\r\n')

    @staticmethod
    def judge_same_event(event: dict, next_event: dict):
        """
        Judge whether the two events are the same,
        :param event: event one
        :param next_event: event two
        :return: same or Not
        """
        for item in event:
            if item not in ["event_code", "source_device"]:
                continue
            if event[item] != next_event[item]:
                return False
        return True

    @staticmethod
    def filter_entity_attributes(attribute: dict):
        if LANG == "en":
            attribute.pop("cause_zh", None)
            attribute.pop("description_zh", None)
            attribute.pop("suggestion_zh", None)
        else:
            attribute.pop("cause_en", None)
            attribute.pop("description_en", None)
            attribute.pop("suggestion_en", None)

    def update_events(self, events_list: list):
        """
        Update events list
        :param events_list: the events_list record
        """
        for event in events_list:
            same_event_list = self.data.setdefault(event.get("event_code", ""), [])
            is_contain = False
            for store_event in same_event_list:
                if not self.judge_same_event(event, store_event):
                    continue
                # occur_time更小时更新，若相等按source_file更新。目的：
                # 按一个标准来保存除["key_info", "occur_time", "source_file"]属性外，其他属性相同的event，而不是不确定的保存。
                if event.get("occur_time", "") < store_event.get("occur_time", ""):
                    merge_occurrence(event, store_event)
                    store_event.update(event)
                if (event.get("occur_time", "") == store_event.get("occur_time", "") and
                        event.get("source_file", "") < store_event.get("source_file", "")):
                    merge_occurrence(event, store_event)
                    store_event.update(event)
                is_contain = True
                break
            if not is_contain:
                same_event_list.append(event)

    def deal_event_data(self):
        """
        该函数包含如下几个功能：
        1、获取所有故障事件的卡信息
        2、设置event_id，诊断时会使用
        3、昇腾组件版本信息调整格式，在清洗文件中落盘，单机诊断也涉及
        """
        count = 1
        for key_name, entities in self.data.items():
            if key_name == self.VERSION_INFO:
                self.version_info = {key: value for key, value in entities[0].items() if key != "event_code"}
                continue
            for event in entities:
                self.devices.add(event.get("source_device", "Unknown"))
                event["event_id"] = f"key{count}"
                count += 1

    def single_worker_fault_analysis(self, file_path: str):
        """
        Analyze the fault of a single work after parse
        :param file_path: output json file path
        """
        final_data = self.get_single_worker_fault_analysis()
        self.write_to_json_file(file_path, final_data)

    def get_single_worker_fault_analysis(self, parse_conf: dict = None):
        """
        Get the fault of a single work after parse
        """
        final_data = {"parse_version": get_version()}
        final_data.update(self.version_info)
        response = {}
        for source_device in self.devices:
            response.update(self.single_device_fault_analysis(source_device, parse_conf))
        final_data.update({"response": response})
        return final_data

    def kg_engine_analyze(self, source_device: str, parse_conf: dict = None):
        """
        Inference engine main function
        :param source_device: source device name
        :param parse_conf: input custom parse config
        :return: inference result
        """
        resp = Response()
        package_data = PackageData([source_device])
        package_data.load_events(self.data)
        if not package_data.event_map:
            return resp
        schema = Schema([DEFAULT_USER_CONF]) if parse_conf is None else Schema(sdk_config_repo=parse_conf)
        graph = GraphBuilder(schema, package_data).build_graph()
        return resp.get_information(graph)

    def single_device_fault_analysis(self, source_device: str, parse_conf: dict = None):
        """
        Analyze the fault of each device in the current worker
        :param source_device: source device name
        :param parse_conf: input custom entity
        """
        resp = Response()
        try:
            resp = self.kg_engine_analyze(source_device, parse_conf)
        except Exception as error:
            resp.error = error
            resp.analyze_success = False
        if not resp.root_causes:
            return {}
        device_causes = {}
        for code, event in resp.root_causes.items():
            self.filter_entity_attributes(event.entities_attribute)
            device_causes.update({
                code: {
                    "code": event.code,
                    "entities_attribute": event.entities_attribute,
                    "events_attribute": event.events_attribute,
                    "chains": event.chains
                }
            })
        response = {
            source_device: {"analyze_success": resp.analyze_success,
                            "error": str(resp.error),
                            "root_causes": device_causes}
        }
        return response

    def dump_to_json_file(self, file_path: str):
        """
        Dump descriptor event data to json file
        :param file_path: output json file path
        """
        final_data = {"parse_version": get_version(), "parse_build_time": get_build_time()}
        for key_name, entities in self.data.items():
            if key_name == self.VERSION_INFO and self.version_info:
                final_data.update(self.version_info)
                continue
            final_data.update({key_name: entities})
        self.write_to_json_file(file_path, final_data)

    def export_server_info_file(self, file_path: str):
        """
        Export server info to json file
        :param file_path: output json file path
        """
        if not self.files_parse_info:
            return
        safe_generate_or_merge_json_file(file_path, self.files_parse_info.trans_parse_info())
