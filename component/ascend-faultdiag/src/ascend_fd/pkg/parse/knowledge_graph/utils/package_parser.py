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
import multiprocessing

from ascend_fd.model.context import KGParseCtx
from ascend_fd.model.parse_info import FilesParseInfo
from ascend_fd.pkg.parse.knowledge_graph.parser.custom_log_parser import CustomLogParser
from ascend_fd.pkg.parse.knowledge_graph.parser.mindio_parser import MindIOLogParser
from ascend_fd.utils.status import InnerError
from ascend_fd.configuration.config import DEFAULT_USER_CONF, KNOWLEDGE_GRAPH_CONF
from ascend_fd.utils.load_kg_config import ParseRegexMap
from ascend_fd.pkg.parse.knowledge_graph.utils.data_descriptor import DataDescriptor
from ascend_fd.pkg.parse.knowledge_graph.parser.cann_log_parser import CANNPlogParser, CANNDeviceLogParser
from ascend_fd.pkg.parse.knowledge_graph.parser.npu_info_parser import NpuInfoParser
from ascend_fd.pkg.parse.knowledge_graph.parser.file_parser import FileParser
from ascend_fd.pkg.parse.knowledge_graph.parser.host_os_parser import HostMsgParser, HostDMesgParser, \
    HostSysMonParser, HostVmCoreParser
from ascend_fd.pkg.parse.knowledge_graph.parser.train_log_parser import TrainLogParser
from ascend_fd.pkg.parse.knowledge_graph.parser.device_plugin_parser import DevicePluginParser
from ascend_fd.pkg.parse.knowledge_graph.parser.noded_log_parser import NodeDLogParser
from ascend_fd.pkg.parse.knowledge_graph.parser.npu_device_parse import NpuHistoryLogParser, NpuOsLogParser, \
    NpuDeviceLogParser
from ascend_fd.pkg.parse.knowledge_graph.parser.volcano_parser import VolcanoSchedulerParser, VolcanoControllerParser
from ascend_fd.pkg.parse.knowledge_graph.parser.common_dl_parser import DockerRuntimeParser, NpuExporterParser
from ascend_fd.pkg.parse.knowledge_graph.parser.amct_log_parser import AMCTLogParser
from ascend_fd.pkg.parse.knowledge_graph.parser.mindie_parser import MindieParser
from ascend_fd.pkg.parse.knowledge_graph.parser.lcne_parser import LCNEParser
from ascend_fd.pkg.parse.knowledge_graph.parser.bmc_parser import BMCParser, BMCLogDumpParser, BMCDeviceDumpParser, \
    BMCAppDumpParser
from ascend_fd.pkg.parse.knowledge_graph.parser.bus_parser import BusParser

kg_logger = logging.getLogger("KNOWLEDGE_GRAPH")
echo = logging.getLogger("ECHO")


def _collect_parser(parser_name: str, parser_collect_result: dict):
    return parser_name, parser_collect_result


class PackageParser(object):
    DEFAULT_PRIMARY_PARSERS = [CANNPlogParser, BMCParser, LCNEParser, BMCLogDumpParser,
                               BMCDeviceDumpParser, BMCAppDumpParser, NpuInfoParser]
    MAX_POOL_SIZE = 20

    def __init__(self, kg_parse_ctx: KGParseCtx, primary_parsers: list = None, other_parsers: list = None,
                 parse_conf: dict = None):
        """
        Log parser base class
        :param kg_parse_ctx: kg parse ctx
        """
        super(PackageParser, self).__init__()
        self.kg_parse_ctx = kg_parse_ctx

        self.primary_parsers = list()
        self.other_parsers = list()
        self.desc = DataDescriptor()
        self.params = {
            "default_conf": ParseRegexMap([KNOWLEDGE_GRAPH_CONF]).get_parse_regex(),
            "user_conf": ParseRegexMap([DEFAULT_USER_CONF]).get_parse_regex()
            if parse_conf is None else ParseRegexMap([DEFAULT_USER_CONF], sdk_config_repo=parse_conf).get_parse_regex()
        }
        self.add_primary_parsers(primary_parsers)
        self.add_other_parsers(other_parsers)

    @classmethod
    def init_sdk_package_parser(cls, parse_ctx: KGParseCtx, parser_classes_list: set, parse_conf: dict):
        """
        sdk pacakge parser initialization
        """
        primary_parsers = []
        other_parsers = []
        primary_parser_names = [parser.__name__ for parser in cls.DEFAULT_PRIMARY_PARSERS]
        for parser in parser_classes_list:
            if parser.__name__ in primary_parser_names:
                primary_parsers.append(parser)
            else:
                other_parsers.append(parser)
        return cls(parse_ctx, primary_parsers, other_parsers, parse_conf)

    def add_parser(self, target_parser_list: list, parser_cls_list: list):
        """
        Add parser class object
        :param target_parser_list: a parser class list to store primary or other parser
        :param parser_cls_list: a parser class list to be added
        """
        for parser_cls in parser_cls_list:
            if not issubclass(parser_cls, FileParser):
                kg_logger.error(f"The {parser_cls} must be sub-class of FileParser.")
                raise InnerError(f"The {parser_cls} must be sub-class of FileParser.")
            target_parser_list.append(parser_cls(self.params))

    def add_primary_parsers(self, input_parser_list: list = None):
        """
        Add parsers to the primary parsers list
        """
        parser_cls_list = input_parser_list if input_parser_list is not None \
            else self.DEFAULT_PRIMARY_PARSERS
        self.add_parser(self.primary_parsers, parser_cls_list)

    def add_other_parsers(self, input_parser_list: list = None):
        """
        Add parsers to the other parsers list
        """
        parser_cls_list = input_parser_list if input_parser_list is not None \
            else [CANNDeviceLogParser, HostMsgParser, HostDMesgParser, HostSysMonParser, HostVmCoreParser,
                  TrainLogParser, NpuHistoryLogParser, NpuOsLogParser, NpuDeviceLogParser, NodeDLogParser,
                  DevicePluginParser, VolcanoSchedulerParser, VolcanoControllerParser, DockerRuntimeParser,
                  NpuExporterParser, AMCTLogParser, MindieParser, CustomLogParser, BusParser, MindIOLogParser]
        self.add_parser(self.other_parsers, parser_cls_list)

    def parse(self, task_id):
        """
        Parse log file with two-phase execution:
        Phase 1: Serial execution of primary_parsers (they set params for other parsers)
        Phase 2: Parallel collect() for other_parsers, then serial filter_events()
        :param task_id: the task unique id
        """
        err_dict = dict()
        issue_dict = dict()
        primary_err, primary_issue = self._parse_primary_group(task_id)
        err_dict.update(primary_err)
        issue_dict.update(primary_issue)
        other_err, other_issue = self._parse_other_group_parallel(task_id)
        err_dict.update(other_err)
        issue_dict.update(other_issue)
        if err_dict:
            kg_logger.warning("These %s parsers parse log failed.", list(err_dict.keys()))
            for parser_name, err_msg in err_dict.items():
                echo.warning("The parser %s failed. The error is: [%s].", parser_name, err_msg)
        for parser_name, err_msg in issue_dict.items():
            echo.warning("The parser %s partially failed. The error is: [%s].", parser_name, ", ".join(err_msg))
        self.desc.deal_event_data()

    def _parse_primary_group(self, task_id: str):
        """
        Parse primary parsers serially (they may set params for other parsers)
        """
        execution_err = dict()
        execution_issue = dict()
        for parser in self.primary_parsers:
            try:
                result, parser_err_dict = parser.parse(self.kg_parse_ctx, task_id)
            except Exception as error:
                kg_logger.warning("The %s parser parse log failed. The reason is: %s",
                                  parser.__class__.__name__, error)
                execution_err.update({parser.__class__.__name__: str(error)})
            else:
                if isinstance(result, FilesParseInfo):
                    self.desc.update_events(result.event_list)
                    self.desc.files_parse_info = result
                else:
                    self.desc.update_events(result)
                execution_issue.update(parser_err_dict)
        return execution_err, execution_issue

    def _parse_other_group_parallel(self, task_id: str):
        """
        Parse other parsers in parallel using collect() then filter_events()
        """
        execution_err = dict()
        execution_issue = dict()
        if not self.other_parsers:
            return execution_err, execution_issue
        collect_results = self._parallel_collect(task_id)
        for parser in self.other_parsers:
            parser_name = parser.__class__.__name__
            collect_result = collect_results.get(parser_name, {})
            if isinstance(collect_result, Exception):
                execution_err.update({parser_name: str(collect_result)})
                continue
            events_list = collect_result.get("events_list", [])
            err_dict = collect_result.get("err_dict", {})
            collect_data = collect_result.get("collect_result", {})
            if isinstance(events_list, FilesParseInfo):
                events_list = events_list.event_list
            try:
                filtered_events = parser.filter_events(events_list, collect_data)
            except Exception as error:
                kg_logger.warning("The %s parser filter events failed. The reason is: %s",
                                  parser_name, error)
                execution_err.update({parser_name: str(error)})
                continue
            if isinstance(events_list, FilesParseInfo):
                self.desc.update_events(filtered_events.event_list)
                self.desc.files_parse_info = events_list
            else:
                self.desc.update_events(filtered_events)
            execution_issue.update(err_dict)
        return execution_err, execution_issue

    def _parallel_collect(self, task_id: str):
        """
        Execute collect() for all other parsers in parallel
        """
        pool_size = min(len(self.other_parsers), self.MAX_POOL_SIZE)
        if self.kg_parse_ctx.is_sdk_input or pool_size <= 1:
            return self._serial_collect()
        manager = multiprocessing.Manager()
        result_dict = manager.dict()
        processes = []
        for parser in self.other_parsers:
            parser_name = parser.__class__.__name__
            p = multiprocessing.Process(
                target=self._collect_worker,
                args=(parser, parser_name, self.kg_parse_ctx, task_id, result_dict)
            )
            p.start()
            processes.append((parser_name, p))
        for parser_name, p in processes:
            p.join()
            if p.exitcode != 0 and parser_name not in result_dict:
                result_dict[parser_name] = Exception(f"Process exited with code {p.exitcode}")
        return dict(result_dict)

    def _serial_collect(self):
        """
        Execute collect() serially (for SDK input or single parser)
        """
        results = {}
        for parser in self.other_parsers:
            parser_name = parser.__class__.__name__
            try:
                events_list, collect_result, err_dict = parser.collect(self.kg_parse_ctx, "")
                results[parser_name] = {
                    "events_list": events_list,
                    "collect_result": collect_result,
                    "err_dict": err_dict
                }
            except Exception as e:
                results[parser_name] = e
        return results

    @staticmethod
    def _collect_worker(parser, parser_name: str, kg_parse_ctx, task_id: str, result_dict: dict):
        """
        Worker function for parallel collect execution
        """
        try:
            events_list, collect_result, err_dict = parser.collect(kg_parse_ctx, task_id)
            result_dict[parser_name] = {
                "events_list": events_list,
                "collect_result": collect_result,
                "err_dict": err_dict
            }
        except Exception as e:
            result_dict[parser_name] = e
