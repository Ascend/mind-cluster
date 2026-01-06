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
import logging
import json
import re
from itertools import chain

from ascend_fd.model.parse_info import PlogPidParseInfo, SuperPodInfo, LCNInfo
from ascend_fd.utils.status import FileNotExistError
from ascend_fd.utils.tool import safe_write_open, safe_read_open, check_file_num_and_size, \
    safe_generate_or_merge_json_file, PLOG_MAX_NUM, PLOG_SIZE_THRESHOLD, SERIAL_NUMBER, \
    BOARD_SERIAL_NUMBER, LCNE_LEVEL, LCNE_SWITCH_ID, LCNE_CONFIG_NAME
from ascend_fd.utils.regular_table import DEVICE_IP_FILE, TLS_SWITCH, HCCL_IP_INFO, HCCL_IPADDR, \
    MAX_TIME, MIN_TIME, HOST_SN, BMC_COMPLETE_MACHINE_SN, BMC_BOARD_SN, LCNE_BOARD_SN, SERVER_INFO_FILE
from ascend_fd.configuration.config import RC_PARSER_DUMP_NAME
from ascend_fd.pkg.parse.root_cluster.parser import PidFileParser

rc_logger = logging.getLogger("ROOT_CLUSTER")
TLS_SWITCH_PATTERN = re.compile(TLS_SWITCH)
HOST_SN_PATTERN = re.compile(HOST_SN)
BMC_BOARD_SN_PATTERN = re.compile(BMC_BOARD_SN)
BMC_COMPLETE_MACHINE_SN_PATTERN = re.compile(BMC_COMPLETE_MACHINE_SN)
LCNE_BOARD_SN_PATTERN = re.compile(LCNE_BOARD_SN)

SERVER_INDEX = "ServerIndex"
SUPER_POD_ID = "SuperPodId"
SUPER_POD_SIZE = "SuperPodSize"
SUPER_POD_BOARD_PRODUCT_NAME = "IT22SMMB"
A3_SUPER_POD = "Atlas 900 A3 SuperPoD Compute Node"
MAX_UINT32 = 2 ** 32 - 1


def parse_lcne_info(cfg):
    """
    Parse LCNE info
    :param cfg: parse config
    """
    board_sn_info = parse_lcne_sn_info(cfg)
    if board_sn_info:
        lcne_sn_dict = {BOARD_SERIAL_NUMBER: board_sn_info}
        safe_generate_or_merge_json_file(os.path.join(cfg.output_path, SERVER_INFO_FILE), lcne_sn_dict)
    lcne_switch_info = parse_lcne_switch_info(cfg)
    if lcne_switch_info:
        super_info_dict = {LCNE_LEVEL: lcne_switch_info.level, LCNE_SWITCH_ID: lcne_switch_info.switch_id,
                           LCNE_CONFIG_NAME: lcne_switch_info.config_name}
        safe_generate_or_merge_json_file(os.path.join(cfg.output_path, SERVER_INFO_FILE), super_info_dict)


def parse_bmc_info(cfg):
    """
    Parse bmc info
    :param cfg: parse config
    """
    complete_machine_sn_info, board_sn_info = parse_bmc_sn_info(cfg)
    bmc_sn_dict = {}
    if complete_machine_sn_info:
        bmc_sn_dict.update({SERIAL_NUMBER: complete_machine_sn_info})
    if board_sn_info:
        bmc_sn_dict.update({BOARD_SERIAL_NUMBER: board_sn_info})
    if bmc_sn_dict:
        safe_generate_or_merge_json_file(os.path.join(cfg.output_path, SERVER_INFO_FILE), bmc_sn_dict)
    super_pod_info = parse_bmc_super_pod_info(cfg)
    if super_pod_info:
        super_info_dict = {SERVER_INDEX: super_pod_info.server_index, SUPER_POD_ID: super_pod_info.super_pod_id,
                           SUPER_POD_SIZE: super_pod_info.super_pod_size}
        safe_generate_or_merge_json_file(os.path.join(cfg.output_path, SERVER_INFO_FILE), super_info_dict)


def start_rc_parse_job(cfg):
    """
    Start rc parse job
    :param cfg: parse config
    """
    if cfg.bmc_log:
        parse_bmc_info(cfg)
    if cfg.lcne_log:
        parse_lcne_info(cfg)
    plog_files = cfg.log_saver.get_plog_dict() if cfg.log_saver is not None else dict()
    check_file_num_and_size(list(chain(*plog_files.values())), rc_logger, file_num=PLOG_MAX_NUM,
                            file_size=PLOG_SIZE_THRESHOLD)
    if not plog_files:
        if cfg.bmc_log or cfg.lcne_log:
            return
        rc_logger.error("No plog file that meets the path specifications is found.")
        raise FileNotExistError("No plog file that meets the path specifications is found.")
    device_ip_map = parse_device_ip_map(cfg)
    tls_status_map = parse_tls_status(cfg)
    pid_parse_result = dict()
    device_pid_map = dict()  # use device id and pid to filter the latest task
    resuming_training_dict = cfg.log_saver.resuming_training_record
    n_second_recovery_record = cfg.log_saver.n_seconds_recovery_record
    device_log_dict = cfg.log_saver.get_device_log_dict()

    for pid, file_list in plog_files.items():
        pid_file_parser = PidFileParser(pid, device_ip_map, resuming_training_dict.get(pid, MIN_TIME),
                                        n_second_recovery_record.get(pid, MIN_TIME))
        for file_path in file_list:
            pid_file_parser.parse_log(file_path)
        this_pid_result = pid_file_parser.get_result()
        if not this_pid_result:
            continue
        pid_file_parser.save_pid_log(cfg.output_path)  # 落盘原始日志，后续删除
        if check_device_id_repeat(pid, this_pid_result, device_pid_map, pid_parse_result):
            continue
        tls_status = tls_status_map.get(pid)
        if tls_status:
            this_pid_result.tls_status = tls_status
        this_pid_result.aicpu_notify_wait_remote = filter_device_error(device_log_dict.get(pid))
        pid_parse_result.update({pid: this_pid_result.to_dict()})

    with safe_write_open(os.path.join(cfg.output_path, RC_PARSER_DUMP_NAME), mode="w+", encoding="utf-8") \
            as file_stream:
        file_stream.write(json.dumps(pid_parse_result, ensure_ascii=False))
        file_stream.write('\r\n')
    # save device_ip_info.json for old version
    device_info_dict = {"device_ip": device_ip_map, "device_tls": tls_status_map}
    with safe_write_open(os.path.join(cfg.output_path, DEVICE_IP_FILE), mode="w+", encoding="utf-8") as file_stream:
        file_stream.write(json.dumps(device_info_dict, ensure_ascii=False))
        file_stream.write('\r\n')
    if parse_host_sn_info(cfg):
        host_sn_dict = {SERIAL_NUMBER: parse_host_sn_info(cfg)}
        safe_generate_or_merge_json_file(os.path.join(cfg.output_path, SERVER_INFO_FILE), host_sn_dict)
    rc_logger.info("The plog parsing result is saved in dir %s.", os.path.basename(cfg.output_path))


def check_device_id_repeat(pid: str, result: PlogPidParseInfo, device_pid_map: dict, pid_parse_result: dict):
    """
    Check whether duplicate device IDs exist and use the latest training data
    """
    base_info = result.base
    device_id = base_info.phy_device_id or base_info.logic_device_id
    if not device_id:
        return False
    if device_id not in device_pid_map:
        device_pid_map.update({device_id: pid})
        return False
    store_pid = device_pid_map.get(device_id)
    store_pid_result = pid_parse_result.get(store_pid)
    if not store_pid_result:
        device_pid_map.update({device_id: pid})
        return False
    store_start_time = store_pid_result.get("start_train_time", MAX_TIME)
    new_start_time = result.start_train_time
    if store_start_time <= new_start_time:
        pid_parse_result.pop(store_pid)
        device_pid_map.update({device_id: pid})
        return False
    return True


def get_tls_switch_info(device_log_file):
    """
    Get TLS SWITCH info from device log file
    """
    with safe_read_open(device_log_file, "r", encoding="UTF-8") as file_stream:
        for line in file_stream:
            tls_switch_match = TLS_SWITCH_PATTERN.search(line.strip())
            if tls_switch_match:
                return tls_switch_match.group(1)
    return ""


def get_host_sn_info(dmidecode_log):
    """
    Get host sn info from dmidecode log file
    """
    with safe_read_open(dmidecode_log, "r", encoding="UTF-8") as file_stream:
        lines = file_stream.readlines()
        selected_lines = []
        for i, line in enumerate(lines):
            if "System Information" in line:
                # Capture the current line and the next 5 lines
                end_index = i + 6
                block = lines[i:end_index]
                selected_lines.extend(block)
        for line in selected_lines:
            sn_match = HOST_SN_PATTERN.search(line.strip())
            if sn_match:
                return sn_match.group(1)
    return ""


def get_lcne_sn_info(lcne_log):
    """
    Get LCNE serial number info from lcne log file
    """
    with safe_read_open(lcne_log, "r", encoding="UTF-8") as file_stream:
        lines = file_stream.readlines()
        for line in lines:
            board_sn_match = LCNE_BOARD_SN_PATTERN.search(line.strip())
            if board_sn_match:
                return board_sn_match.group(1)
    return ""


def get_bmc_sn_info(bmc_log):
    """
    Get bmc sn info from bmc log file
    """
    complete_machine_sn_info = ""
    board_sn_info = ""

    with safe_read_open(bmc_log, "r", encoding="UTF-8") as file_stream:
        lines = file_stream.readlines()
        selected_lines = []
        for i, line in enumerate(lines):
            if "Atlas 900 A3 SuperPoD Compute Node" in line or SUPER_POD_BOARD_PRODUCT_NAME in line:
                # Capture the current line and the next 5 lines
                end_index = i + 6
                block = lines[i:end_index]
                selected_lines.extend(block)
        for line in selected_lines:
            complete_machine_sn_match = BMC_COMPLETE_MACHINE_SN_PATTERN.search(line.strip())
            if complete_machine_sn_match:
                complete_machine_sn_info = complete_machine_sn_match.group(1)
            board_sn_match = BMC_BOARD_SN_PATTERN.search(line.strip())
            if board_sn_match:
                board_sn_info = board_sn_match.group(1)
    return complete_machine_sn_info, board_sn_info


def get_bmc_super_pod_info(bmc_log):
    """
    Get BMC super pod info from bmc log file
    """
    super_pod_info_dict = {}
    current_key = None
    targets = {SERVER_INDEX, SUPER_POD_ID, SUPER_POD_SIZE}

    with safe_read_open(bmc_log, "r", encoding="UTF-8") as file_stream:
        lines = file_stream.readlines()
        for line in lines:
            line = line.strip()
            if line.startswith('.'):
                key = line[1:]  # Remove the leading "."
                current_key = key if key in targets else None
            elif current_key and line.startswith('value:'):
                # Extract value and convert it to integer
                value = line.split(':', 1)[1].strip()
                super_pod_info_dict[current_key] = int(value) if value.isdigit() else 0

    return SuperPodInfo(super_pod_info_dict.get(SERVER_INDEX, 0), super_pod_info_dict.get(SUPER_POD_ID, 0),
                        super_pod_info_dict.get(SUPER_POD_SIZE, 0))


def get_lcne_switch_info(lcne_log):
    """
    Get LCNE switch info from lcne log file
    """
    result = {}
    current_block = {}

    # get lcne level and switchId. eg: set switch-position level 1 switchId 22
    switch_pattern = re.compile(r'^set switch-position level (\d{1,2}) switchId (\d{1,3})$')
    # get lcne config. eg: superpod configuration typical atlas900t_384_topo1
    superpod_pattern = re.compile(r'^superpod configuration typical (\S{1,25})$')

    with safe_read_open(lcne_log, "r", encoding="UTF-8") as file_stream:
        lines = file_stream.readlines()
        for line in lines:
            if not line or line == "--":
                continue
            switch_match = switch_pattern.search(line.strip())
            superpod_match = superpod_pattern.search(line.strip())
            if switch_match:
                current_block.update({
                    LCNE_LEVEL: int(switch_match.group(1)),
                    LCNE_SWITCH_ID: int(switch_match.group(2))
                })
            elif superpod_match:
                current_block[LCNE_CONFIG_NAME] = superpod_match.group(1)
                result = current_block.copy()
                current_block = {}  # reset current block

    return LCNInfo(result.get(LCNE_LEVEL, 0), result.get(LCNE_SWITCH_ID, 0),
                   result.get(LCNE_CONFIG_NAME, ""))


def parse_npu_info_file(npu_info_file):
    """
    Parse npu_info_before(after).txt file
    """
    device_to_rank = dict()
    with safe_read_open(npu_info_file, "r", encoding="UTF-8") as file_stream:
        content = file_stream.read()
        event_message_list = content.split("\n\n")
        for event_message in event_message_list:
            event_message = event_message.strip()
            if "ipaddr" not in event_message:
                continue
            """
            examples of some contents in the npu_info_before(after).txt file 
                hccn_tool -i 0 -ip -g
                ipaddr:x.x.x.x
                netmask:x.x.x.x

                hccn_tool -i 1 -ip -g
                ipaddr:x.x.x.x
                netmask:x.x.x.x
            """
            hccl_info_re = re.search(HCCL_IP_INFO, event_message)
            ipaddr_re = re.search(HCCL_IPADDR, event_message)
            if not hccl_info_re or not ipaddr_re:
                continue
            device_id = hccl_info_re[1]
            device_ip = ipaddr_re[1]
            device_to_rank[device_id] = device_ip

    return device_to_rank


def parse_device_ip_map(cfg):
    """
    Parse the device ip map from npu_info_before/after.txt
    """
    device_ip_map = dict()
    if cfg.env_info_saver is None:
        return device_ip_map
    npu_info_list = cfg.env_info_saver.get_npu_info_list()
    if npu_info_list:
        for npu_info in npu_info_list:
            if not os.path.exists(npu_info):
                continue
            device_ip_map.update(parse_npu_info_file(npu_info))
    return device_ip_map


def parse_tls_status(cfg):
    """
    Parse the device tls status from CANN_Device log
    """
    device_tls_status = dict()
    for pid, file_path_list in cfg.log_saver.get_device_log_dict().items():
        for file in file_path_list:
            tls_status = get_tls_switch_info(file)
            if tls_status:
                device_tls_status.update({pid: tls_status})
                break
    return device_tls_status


def filter_device_error(device_log_files):
    """
    Parse the device tls status from CANN_Device log
    """
    if not device_log_files:
        return ""
    for file in device_log_files:
        remote_rank_id = ""
        with safe_read_open(file, "r", encoding="UTF-8") as file_stream:
            for line in file_stream:
                if not remote_rank_id:
                    remote_rank_id = find_remote_rank_id(line)
                    continue
                if remote_rank_id:
                    return ("{}:{}".format(find_group(line), remote_rank_id)) if find_group(line) else ""
    return ""


def find_remote_rank_id(log_line):
    # base information is streamId:89, sqid:89, head:1691, tail:1695, type:NOTIFY WAIT, localRank:1, remoteRank:0
    if "remoteRank:" not in log_line or "NOTIFY WAIT" not in log_line:
        return ""
    items = log_line.split(",")
    for item in items:
        if "remoteRank:" in item:
            rank_id = item.split(":")[-1].strip()
            if rank_id != str(MAX_UINT32):
                return rank_id
    return ""


def find_group(log_line):
    # opData information is tag:worldCommSendRecv_1_0_group_name, group:group_name_2881, opIndex:13004809
    if "group:" not in log_line:
        return ""
    items = log_line.split(",")
    for item in items:
        if "group:" in item:
            return item.split(":")[-1].strip()
    return ""


def parse_host_sn_info(cfg):
    if cfg.host_log_saver is None:
        return ""
    for file in cfg.host_log_saver.get_dmidecode_log():
        complete_machine_sn_info = get_host_sn_info(file)
        if complete_machine_sn_info:
            return complete_machine_sn_info
    return ""


def parse_bmc_sn_info(cfg):
    for file in cfg.bmc_log_saver.get_fruinfo_log():
        return get_bmc_sn_info(file)
    return "", ""


def parse_bmc_super_pod_info(cfg):
    for file in cfg.bmc_log_saver.get_mdb_info_log():
        return get_bmc_super_pod_info(file)
    return None


def parse_lcne_sn_info(cfg):
    for file in cfg.lcne_log_saver.get_devm_bddvadp_log():
        return get_lcne_sn_info(file)
    return ""


def parse_lcne_switch_info(cfg):
    for file in cfg.lcne_log_saver.get_diag_display_info_log():
        return get_lcne_switch_info(file)
    return None
