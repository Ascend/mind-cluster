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
from ascend_fd.pkg import start_rc_parse_job, start_kg_parse_job, start_kg_diag_job
from ascend_fd.utils.status import InnerError


def kg_diag_job(cfg):
    """
    Used to perform Kg diag job
    :param cfg: config dict
    :return: kg diag result dict, {"Kg": result dict}
    """
    kg_result = start_kg_diag_job(cfg)
    return {"Kg": kg_result}


def node_diag_job(cfg):
    """
    Used to perform Node diag job
    :param cfg: config dict
    :return: node diag result dict, {"Node": result dict}
    """
    try:
        from ascend_fd.pkg.diag.node_anomaly import start_node_diag_job
    except ImportError as e:
        raise InnerError(f"Import Node Anomaly diag module failed: [{e}]") from e
    result = start_node_diag_job(cfg)
    return {"Node": result}


def net_diag_job(cfg):
    """
    Used to perform Net diag job
    :param cfg: config dict
    :return: net diag result dict, {"Net": result dict}
    """
    try:
        from ascend_fd.pkg.diag.network_congestion import start_net_diag_job
    except ImportError as e:
        raise InnerError(f"Import Network Congestion diag module failed: [{e}]") from e
    result = start_net_diag_job(cfg)
    return {"Net": result}


def generate_parse_job(flag):
    """
    Generate the parse job dict
    :param flag: the performance flag
    :return: parse job dict
    """
    parse_jobs = {
        "ROOT_CLUSTER": start_rc_parse_job,
        "KNOWLEDGE_GRAPH": start_kg_parse_job
    }
    if flag:
        try:
            from ascend_fd.pkg.parse.node_anomaly import start_node_parse_job
        except ImportError as e:
            raise InnerError(f"Import Node Anomaly parse module failed: [{e}]") from e
        try:
            from ascend_fd.pkg.parse.network_congestion import start_net_parse_job
        except ImportError as e:
            raise InnerError(f"Import Network Congestion parse module failed: [{e}]") from e
        parse_jobs.update({"NODE_ANOMALY": start_node_parse_job, "NET_CONGESTION": start_net_parse_job})
    return parse_jobs


def generate_diag_job(flag):
    """
    Generate the diag job dict
    :param flag: the performance flag
    :return: diag job dict
    """
    diag_jobs = {
        "KNOWLEDGE_GRAPH": kg_diag_job
    }
    if flag:
        diag_jobs.update({"NODE_ANOMALY": node_diag_job, "NET_CONGESTION": net_diag_job})
    return diag_jobs
