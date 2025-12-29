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
from ascend_fd.pkg.diag.fault_entity import ALL_PROCESS_PREEMPT_FAULT_ENTITY, PART_PROCESS_PREEMPT_FAULT_ENTITY, \
    SINGLE_PROCESS_PREEMPT_FAULT_ENTITY, NODE_DIAG_NORMAL_ENTITY
from ascend_fd.pkg.diag.node_anomaly.resource_preemption.utils.abnormal_detect_tool import (ALL_PROCESS_FAULT,
                                                                                            SINGE_PROCESS_FAULT,
                                                                                            PART_PROCESS_FAULT,
                                                                                            RANDOM_PROCESS_FAULT)
from ascend_fd.utils.i18n import get_label_for_language


class ResultSmoother:
    """
    This class is used to smooth the detection results.
    Remove some anomalies that affect the detection results (noise result).

    key: structure map; value: (noise_length, noise_pos).
    T: contain this check flag; F: not contain this check flag; N: ignore this pos, both contain or not-contain pass.

    In a flag structure sequence, if the inclusion relationship of the check flag conforms to the above sequence,
    the flag inclusion relationship of the entire result sequence is modified to the inclusion relationship
    corresponding to the first value of the sequence (T or F).
    If "TFTT", the check flags are added where they are not included, and vice versa.

    Example:
    origin flag sequence: [20, 0, 20, 20]; flag structure: "TFTT"; check flag: 20;
    modified sequence: [20, 20, 20, 20]

    origin flag sequence: [0, 20, 10, 10, 10]; flag structure: "FTFFF"; check flag: 20;
    modified sequence: [0, 0, 10, 10, 10]
    """
    RANDOM_PROBABILITY_LIMIT = 0.35
    CPU_ABNORMAL = {ALL_PROCESS_FAULT, SINGE_PROCESS_FAULT, PART_PROCESS_FAULT, RANDOM_PROCESS_FAULT}
    FLAG_STRUCTURE_MAP = {
        "TFTT": (1, 1),
        "FTFFF": (1, 1),
        "FFTFF": (1, 2),
        "TTFTT": (1, 2),
        "TFFTT": (2, 1),
        "TTTNNNTT": (3, 3),
        "FFFNNNNFFF": (4, 3)
    }

    @staticmethod
    def check_flag_structure(flag_sequence, flag_structure, check_flag):
        """
        Detect whether the flag sequence conforms to the flag structure.
        :param flag_sequence: the flag sequence
        :param flag_structure: the flag structure
        :param check_flag: the check flag
        :return: flag  true after check
        """
        check_res = ["T" if check_flag == single_flag else "F" for single_flag in flag_sequence]
        compare_flags = [flag in [check_flag, "N"] for check_flag, flag in zip(check_res, flag_structure)]
        return all(compare_flags)

    @staticmethod
    def modify_flag_sequence(fault_flag_list, pos, noise_len, check_flag, tag):
        """
        Modify flag sequence base on tag("T/F").
        :param fault_flag_list: the fault flag list
        :param pos: the tag pos
        :param noise_len: the noise len
        :param check_flag: the check flag
        :param tag: the flag structure tag
        :return: the fault flag list
        """
        if tag == "T":
            remove_value = 0
            add_value = check_flag
        else:
            remove_value = check_flag
            add_value = 0

        for i in range(pos, min(len(fault_flag_list), pos + noise_len)):
            fault_flag_list[i] = add_value if fault_flag_list[i] == remove_value else fault_flag_list[i]
        return fault_flag_list

    @staticmethod
    def modify_pid_sequence(fault_pid_list, pos, noise_len, begin_pos):
        """
        Modify pid flag sequence base on tag "T".
        :param fault_pid_list: the fault pid list
        :param pos: the tag pos
        :param noise_len: the noise len
        :param begin_pos: the beginning pos
        :return: the fault pid list
        """
        begin_pos = begin_pos if begin_pos > 0 else 0
        add_value = fault_pid_list[begin_pos]

        for i in range(pos, min(len(fault_pid_list), pos + noise_len)):
            fault_pid_list[i] = add_value
        return fault_pid_list

    def cpu_smooth(self, cpu_df):
        """
        Change process random preemption flag and smooth abnormal cpu result.
        1. Filter and change process random preemption flag.
        2. Smooth abnormal result based on cpu flag structure and pid flag structure.
        3. Remove 0 from the smooth result and only the error type remains.
        :param cpu_df: the cpu input data
        :return: the cpu data after smooth
        """
        fault_flag_list = list(cpu_df["fault_cpu_flag"])
        fault_pid_list = list(cpu_df["fault_pid"])
        fault_flag_list = self._change_process_random_preemption_flag(fault_flag_list)
        fault_flag_list, fault_pid_list = self._smooth_abnormal_result(fault_flag_list, fault_pid_list)

        cpu_df["fault_cpu_flag"] = fault_flag_list
        cpu_df["fault_pid"] = fault_pid_list
        cpu_df.reset_index(drop=True, inplace=True)
        return cpu_df

    def _change_process_random_preemption_flag(self, fault_flag_list):
        """
        If CPU resource preemption occurs continuously and the random preemption result exceeds 35%,
        all process detection result during this period of time is set to random preemption.
        :param fault_flag_list: the fault flag list
        :return: the fault flag list after change random preemption flag
        """
        i = 0
        while i < len(fault_flag_list):
            if fault_flag_list[i] not in self.CPU_ABNORMAL:
                # if not contain cpu abnormal, skip
                i += 1
                continue

            j = i + 1
            cpu_random_count = 0
            while j < len(fault_flag_list) and fault_flag_list[i] in self.CPU_ABNORMAL:
                # random preemption occur times.
                if fault_flag_list[j] == RANDOM_PROCESS_FAULT:
                    cpu_random_count += 1
                j += 1

            abnormal_cpu_len = j - i
            if abnormal_cpu_len * self.RANDOM_PROBABILITY_LIMIT <= cpu_random_count:
                # remove all other cpu preemption and set as random preemption
                while i < j:
                    fault_flag_list[i] = RANDOM_PROCESS_FAULT
                    i += 1
            i = j
        return fault_flag_list

    def _smooth_abnormal_result(self, fault_flag_list, fault_pid_list=None):
        """
        Check each flag structure and modify the fault flag list.
        :param fault_flag_list: the fault flag list
        :param fault_pid_list: the pid list
        :return: the data after smooth
        """
        # get fault flags after remove 0 flag and start smooth
        for check_flag in sorted(list(set(fault_flag_list) - {0})):
            len_flag = 0
            while len_flag < len(fault_flag_list):
                fault_flag_list, fault_pid_list, jump_len = self._smooth_flag_structure_result(
                    len_flag, check_flag, fault_flag_list, fault_pid_list)
                len_flag += jump_len
        return fault_flag_list, fault_pid_list

    def _smooth_flag_structure_result(self, len_flag, check_flag, fault_flag_list, fault_pid_list=None):
        """
        Smooth data based on flag structure.
        :param len_flag: the fault list len flag
        :param check_flag: the check flag
        :param fault_flag_list: the fault flag list
        :param fault_pid_list: the fault pid list
        :return: the fault flag list and jump len after smooth flag
        """
        jump_len = 1
        for key, value in self.FLAG_STRUCTURE_MAP.items():
            flag_structure = key
            noise_len, noise_pos = value
            begin_pos = len_flag - noise_pos
            end_pos = begin_pos + len(flag_structure)
            flag_sequence = []

            for j in range(begin_pos, end_pos):
                if j < 0 or j >= len(fault_flag_list):
                    flag_sequence.append(0)
                    continue
                flag_sequence.append(fault_flag_list[j])

            # if the flag sequences are inconsistent, modify the flag according to the sequence.
            if self.check_flag_structure(flag_sequence, flag_structure, check_flag):
                self.modify_flag_sequence(fault_flag_list, len_flag, noise_len, check_flag, flag_structure[0])
                if fault_pid_list and flag_structure[0] == "T":
                    self.modify_pid_sequence(fault_pid_list, len_flag, noise_len, begin_pos)
                jump_len = len(flag_structure) - noise_pos
                break
        return fault_flag_list, fault_pid_list, jump_len


class ResultProbComputer:
    """
    This class is used to compute the probability of detection results.
    Probability flag structure:
    Select four neighbor flags on the left and right respectively,
    the correlation decreases from the middle to the two sides,
    the probability of neighbor flags impact is 0.9, self-probability impact is 0.1.
    """
    NEIGHBOUR_LEN = 4
    PROBABILITY = [0.1 * 0.45, 0.2 * 0.45, 0.3 * 0.45, 0.4 * 0.45, 0.1, 0.4 * 0.45, 0.3 * 0.45, 0.2 * 0.45, 0.1 * 0.45]

    def compute(self, res_df, flag):
        """
        Calculate the probability value for each detection result, and record the value in probability column.
        :param res_df: the input data
        :param flag: the fault flag type
        :return: the result after adding the probability column
        """
        fault_flag_list = list(res_df[f"fault_{flag}_flag"])
        all_probability_list = list()

        for index, fault_flag in enumerate(fault_flag_list):
            probability_dict = dict()
            probability_dict[fault_flag] = self._compute_single_flag_probability(fault_flag_list, index, fault_flag)
            all_probability_list.append(probability_dict)
        res_df[f"{flag}_probability"] = all_probability_list
        return res_df

    def _compute_single_flag_probability(self, fault_flag_list, index, fault_flag):
        """
        Calculate the probability value of each fault_flag based on the neighbour flag.
        :param fault_flag_list: the fault flag list
        :param index: the fault flag index
        :param fault_flag: the fault flag
        :return: the probability result of this fault flag
        """
        all_index = [i for i in range(index - self.NEIGHBOUR_LEN, index + self.NEIGHBOUR_LEN + 1)]
        all_weight = []

        for i in all_index:
            if i < 0 or i >= len(fault_flag_list):
                all_weight.append(0)
                continue
            if fault_flag == fault_flag_list[i]:
                all_weight.append(1)
                continue
            all_weight.append(0)
        return sum([b * p for b, p in zip(all_weight, self.PROBABILITY)])


class FaultTimePeriodFormatter:
    """
    This class is used to get fault detail info in each fault time period.
    """

    @staticmethod
    def _find_consecutive_fault_pos(fault_flag_list):
        """
        Get the consecutive fault period [start, end).
        RANDOM_PROCESS_FAULT can be regarded as any error
        :param fault_flag_list: the fault flog list
        :return: the consecutive fault period list
        """
        start_pos, period_flag = None, None
        pos_list = []
        fault_flag_list.append(0)  # add an end flag.
        for index, flag in enumerate(fault_flag_list):
            if flag == 0:
                if start_pos:
                    pos_list.append([start_pos, index])
                    start_pos, period_flag = None, None
                continue
            if not start_pos:
                start_pos = index
            # record first valid fault flag.
            if flag != RANDOM_PROCESS_FAULT and not period_flag:
                period_flag = flag
            if flag not in [period_flag, RANDOM_PROCESS_FAULT]:
                pos_list.append([start_pos, index])
                start_pos = index
                period_flag = flag
        return pos_list

    @staticmethod
    def _get_error_info(period_df, flag):
        """
        Get error type、 error probability and error pid in this time period.
        :param period_df: data in this time period
        :param flag: the fault flag type
        :return: fault type、 fault probability and fault pid in this time period
        """
        # fault_flag_list up to two values: fault flag and random flag.
        fault_flag_list = list(period_df[f"fault_{flag}_flag"])
        fault_flag = list(set(fault_flag_list) - {RANDOM_PROCESS_FAULT})
        if not fault_flag:
            return "", "", ""

        # compute probability in this period.
        fault_prob = 0.
        for prob_dict in list(period_df[f"{flag}_probability"]):
            if fault_flag and fault_flag[0] in prob_dict:
                fault_prob += prob_dict.get(fault_flag[0], 0.) / len(period_df)

        fault_pid = period_df["fault_pid"].iloc[0]
        return fault_flag[0], "%.3f" % fault_prob, sorted(fault_pid)

    def format(self, res_df, flag):
        """
        Extract specific fault information from detect results.
        :param flag: the fault flag type
        :param res_df: the input data
        :return: error info dict for continuous faults
        """
        error_info_dict = {}
        pos_list = self._find_consecutive_fault_pos(list(res_df[f"fault_{flag}_flag"]))
        for pos in pos_list:
            period_df = res_df.copy()
            period_df = period_df[pos[0]: pos[1]]
            if period_df.empty:
                continue
            # get every fault period and probability and format error info.
            begin_time = period_df["time"].iloc[0]
            end_time = period_df["time"].iloc[-1]
            fault_flag, fault_prob, fault_pid = self._get_error_info(period_df, flag)
            if not fault_flag:
                continue
            type_dict = error_info_dict.setdefault(fault_flag, dict())
            lb = get_label_for_language()
            type_dict.setdefault(str(fault_pid), list()).append([(begin_time, end_time),
                                                                 f"{lb.fault_probability} : {fault_prob}"])
        return error_info_dict


def wrap_error_result(errors_dict, worker_name):
    """
    Convert the detection results to the final output format.
    :param errors_dict: the error info dict
    :param worker_name: the worker name
    :return: fault_code and fault_detail info
    """
    fault_map = {
        ALL_PROCESS_FAULT: ALL_PROCESS_PREEMPT_FAULT_ENTITY,
        SINGE_PROCESS_FAULT: SINGLE_PROCESS_PREEMPT_FAULT_ENTITY,
        PART_PROCESS_FAULT: PART_PROCESS_PREEMPT_FAULT_ENTITY
    }

    causes = dict()
    if not errors_dict:
        return {NODE_DIAG_NORMAL_ENTITY: []}

    for fault_flag, fault_details in errors_dict.items():
        fault_entity = fault_map.get(fault_flag)
        if not fault_entity:
            continue
        for fault_pid, fault_period_prob in fault_details.items():
            fault_detail = {
                "fault_period_probability": fault_period_prob,
                "process_id": fault_pid,
                "worker": worker_name
            }
            causes.setdefault(fault_entity, []).append(fault_detail)
    return causes
