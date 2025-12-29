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
import warnings

from ascend_fd.utils.tool import MultiProcessJob
from ascend_fd.model.cfg import DiagCFG
from ascend_fd.utils.fault_code import NODE_DIAGNOSIS_NORMAL
from ascend_fd.pkg.diag.message import SOME_SUBTASKS_FAILED
from ascend_fd.pkg.diag.node_anomaly.npu_anomaly import npu_anomaly_job
from ascend_fd.pkg.diag.node_anomaly.resource_preemption import resource_preemption_job

warnings.filterwarnings("ignore")
node_logger = logging.getLogger("NODE_ANOMALY")


def start_node_diag_job(cfg: DiagCFG):
    """
    Start node diag job
    :param cfg: diag config
    :return: node diag result dict
    result format:
    {"analyze_success": True or False,
     "fault": [
        {"code": "", "cause_zh": "", "description_zh": "", "suggestion_zh": "",
         "fault_details": [
             {"fault_period_probability": [[],[]], "process_id": [], "worker": ""}  # cpu resource preemption
             or {"device": ("worker_name", "npu_id")}  # npu overload frequency]
        },
        {...}
     ],
     "note_msgs": "NoteMsg"}
    """
    result = {'analyze_success': True}
    worker_causes = dict()
    multiprocess_job = MultiProcessJob("NODE_ANOMALY", pool_size=20, task_id=cfg.task_id)
    worker_dict = cfg.parsed_saver.get_all_worker_dir_path()
    for worker_name, worker_dir in worker_dict.items():
        multiprocess_job.add_security_job(f"NPU_ANOMALY_{worker_name}", npu_anomaly_job,
                                          worker_dir, worker_name)
        multiprocess_job.add_security_job(f"RESUORCE_PREEMPTION_{worker_name}", resource_preemption_job,
                                          worker_dir, worker_name)
    results, failed_details = multiprocess_job.join_and_get_results()
    for _, result_dict in results.items():
        for fault_entity, fault_detail_list in result_dict.items():
            # combine fault detail list by fault code.
            worker_cause = worker_causes.setdefault(fault_entity.code, fault_entity.attribute)
            worker_cause.setdefault("fault_details", []).extend(fault_detail_list)
    # if worker causes contains other fault, remove NODE_DIAGNOSIS_NORMAL code
    if NODE_DIAGNOSIS_NORMAL in worker_causes and set(worker_causes.keys()) - {NODE_DIAGNOSIS_NORMAL}:
        worker_causes.pop(NODE_DIAGNOSIS_NORMAL)
    result["fault"] = list(worker_causes.values())
    if failed_details:  # some subtasks (not all) failed.
        result["note_msgs"] = SOME_SUBTASKS_FAILED
        result["failed_jobs"] = failed_details
    return result
