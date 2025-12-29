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
from dataclasses import dataclass, fields

from ascend_fd.model.node_info import FaultFilterTime
from ascend_fd.utils import regular_table
from ascend_fd.pkg.parse.parser_saver import ProcessLogSaver, EnvInfoSaver, TrainLogSaver, HostLogSaver, BMCLogSaver, \
    LCNELogSaver, DevLogSaver, DlLogSaver, AMCTLogSaver, MindieLogSaver, ParsedDataSaver, CustomLogSaver


@dataclass
class DiagCFG:
    """
    Diag Config
    """
    task_id: str
    input_path: str
    output_path: str
    parsed_saver: ParsedDataSaver
    root_worker_devices = {}
    fault_filter_time = FaultFilterTime(regular_table.MIN_TIME, regular_table.MAX_TIME)


@dataclass
class ParseCFG:
    """
    Parse Config
    """
    task_id: str
    input_path: str
    output_path: str
    lcne_log: str
    bmc_log: str
    # the input allows absense of savers
    log_saver: ProcessLogSaver = None
    env_info_saver: EnvInfoSaver = None
    train_log_saver: TrainLogSaver = None
    host_log_saver: HostLogSaver = None
    dev_log_saver: DevLogSaver = None
    dl_log_saver: DlLogSaver = None
    amct_log_saver: AMCTLogSaver = None
    mindie_log_saver: MindieLogSaver = None
    bmc_log_saver: BMCLogSaver = None
    lcne_log_saver: LCNELogSaver = None
    custom_log_saver: CustomLogSaver = None

    @property
    def is_sdk_input(self):
        return not self.input_path and not self.output_path

    @classmethod
    def config_saver(cls, kwargs, saver_list):
        for saver in saver_list:
            for f in fields(cls):
                if f.name in kwargs:
                    continue
                if f.type == type(saver):
                    kwargs[f.name] = saver
                    break

    @classmethod
    def sdk_config(cls, task_id: str, saver_list: list) -> 'ParseCFG':
        """
        Automatically allocating saver to config
        """
        kwargs = {
            "task_id": task_id,
            "input_path": "",
            "output_path": "",
            "lcne_log": "",
            "bmc_log": ""
        }
        cls.config_saver(kwargs, saver_list)
        return cls(**kwargs)

    @classmethod
    def cmd_config(cls, args, saver_list: list) -> 'ParseCFG':
        kwargs = {
            "task_id": args.task_id,
            "input_path": args.input_path,
            "output_path": args.output_path,
            "lcne_log": getattr(args, "lcne_log", ""),
            "bmc_log": getattr(args, 'bmc_log', ""),
        }
        cls.config_saver(kwargs, saver_list)
        return cls(**kwargs)
