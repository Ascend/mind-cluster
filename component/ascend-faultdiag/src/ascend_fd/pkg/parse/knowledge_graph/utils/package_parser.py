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

from ascend_fd.model.context import KGParseCtx
from ascend_fd.model.parse_info import FilesParseInfo
from ascend_fd.pkg.parse.knowledge_graph.parser.custom_log_parser import CustomLogParser
from ascend_fd.pkg.parse.knowledge_graph.parser.mindio_parser import MindIOLogParser
from ascend_fd.utils.status import InnerError
from ascend_fd.configuration.config import DEFAULT_USER_CONF
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


class PackageParser(object):
    DEFAULT_PRIMARY_PARSERS = [CANNPlogParser, BMCParser, LCNEParser, BMCLogDumpParser,
                               BMCDeviceDumpParser, BMCAppDumpParser, NpuInfoParser]

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
        self.params = {"regex_conf": ParseRegexMap([DEFAULT_USER_CONF]).get_parse_regex()} if parse_conf is None \
            else {"regex_conf": ParseRegexMap(sdk_config_repo=parse_conf).get_parse_regex()}
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
        Parse log file
        :param task_id: the task unique id
        """
        err_dict = dict()
        issue_dict = dict()
        for parser_group in [self.primary_parsers, self.other_parsers]:
            execution_err, execution_issue = self.parse_group(parser_group, task_id)
            err_dict.update(execution_err)
            issue_dict.update(execution_issue)
        if err_dict:
            kg_logger.warning("These %s parsers parse log failed.", list(err_dict.keys()))
            for parser_name, err_msg in err_dict.items():
                echo.warning("The parser %s failed. The error is: [%s].", parser_name, err_msg)
        for parser_name, err_msg in issue_dict.items():
            echo.warning("The parser %s partially failed. The error is: [%s].", parser_name, ", ".join(err_msg))
        self.desc.deal_event_data()

    def parse_group(self, parser_group: list, task_id: str):
        execution_err = dict()
        execution_issue = dict()
        for parser in parser_group:
            try:
                result, parser_err_dict = parser.parse(self.kg_parse_ctx, task_id)
            except Exception as error:
                kg_logger.warning("The %s parser parse log failed. The reason is: %s",
                                  parser.__class__.__name__, error)
                execution_err.update({parser.__class__.__name__: error})
            else:
                # 此处先特殊判定，MindieParser.parse返回files_parse_info，其他parser.parse返回event_list
                if isinstance(result, FilesParseInfo):
                    self.desc.update_events(result.event_list)
                    self.desc.files_parse_info = result
                else:
                    self.desc.update_events(result)
                execution_issue.update(parser_err_dict)
        return execution_err, execution_issue
