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
import abc
import argparse
import datetime
import getpass
import io
import json
import logging
import logging.handlers
import multiprocessing
import os
import re
import shutil
import signal
import string
import subprocess
import sys
import uuid
from dataclasses import dataclass
from pathlib import Path
from typing import List, Dict, Type, Optional, Any, Union, Callable, BinaryIO, Tuple

from ascend_fd.configuration.config import DEFAULT_USER_CONF, DEFAULT_HOME_PATH, ENV_VAR_KEY, RUN_LOG_FORMAT, \
    OP_LOG_FILE, HOME_PATH
from ascend_fd.utils import regular_table
from ascend_fd.utils.constant.str_const import COMPLEMENT, ERROR_CODE
from ascend_fd.utils.fault_code import CANN_ERRCODE_CUSTOM, PYTORCH_ERRCODE_COMMON, MINDIE_ERRCODE_COMMON
from ascend_fd.utils.status import FileOpenError, InnerError, SuccessRet, FileTooLarge, FileNotExistError, \
    InfoIncorrectError, InfoNotFoundError, ParamError, PathError

VERSION_FILE_READ_LIMIT = 100
MAX_SIZE = 512 * 1024 * 1024
FILES_SIZE_THRESHOLD = 5 * 1024 * 1024 * 1024
PLOG_SIZE_THRESHOLD = 1 * 1024 * 1024 * 1024
FILE_MAX_NUM = 1000000
PLOG_MAX_NUM = 10000
MB_SHIFT = 20
GB_SHIFT = 30
MAX_PATH_LEN = 4096
MAX_CODE_LEN = 50
MAX_PARAM_LEN = 200
LINE_NUM_OF_MULTILINE_MATCH = 10
SHOW_LINES_NUM = 10
PATH_WHITE_LIST_LIN = string.digits + string.ascii_letters + '~-+_./ '
TIME_WHITE_LIST = string.digits + "-:.T+ "
CODE_WHITE_LIST = string.digits + string.ascii_letters + "-_"
ITEM_CHOICES = ["attribute", "rule", "regex"]
CONF_PATH = os.path.join(os.path.dirname(os.path.dirname(os.path.realpath(__file__))), "configuration")
INVALID_HOME_PATH_LIST = ["/tmp"]
LOG_DIR_PATH = os.path.join(HOME_PATH, "RUN_LOG")
DEFAULT_BLACKLIST_CONF = os.path.join(HOME_PATH, 'custom-blacklist.json')
echo = logging.getLogger("ECHO")
fd_logger = logging.getLogger("FAULT_DIAG")
LOG_MAX_BACKUP_COUNT = 10
LOG_MAX_SIZE = 10 * 1024 * 1024
LOG_MODULE = ["FAULT_DIAG", "ROOT_CLUSTER", "KNOWLEDGE_GRAPH", "NODE_ANOMALY", "NET_CONGESTION", "KG_ENGINE"]
DETAIL_FORMAT = '%(levelname)-7s[%(name)s][%(asctime)s][Process %(processName)s:%(process)d]' \
                '[%(filename)s:%(lineno)d] %(message)s'
OPERATION_FORMAT = '[%(asctime)s][%(levelname)s][%(user_name)s@%(host_name)s@%(process)s] %(message)s'
RUN_LOG_DIR_MAX_SIZE = 100 * 1024 * 1024
RUN_LOG_DIR_MAX_NUM = 100
RUN_LOG_DIR_MAX_SAVE_SIZE = 80 * 1024 * 1024
RUN_LOG_DIR_CLEAN_NUM = 20
MAX_PROCESS_DEEP = 5
WORKER_MAX_NUM = 1000000
DIR_MAX_DEPTH = 20
DIAG_CHOICES = ["host", "super_pod"]

SERIAL_NUMBER = "Serial Number"
BOARD_SERIAL_NUMBER = "Board Serial Number"
LCNE_LEVEL = "level"
LCNE_SWITCH_ID = "switchId"
LCNE_CONFIG_NAME = "configName"

SHOW_IP_MAX = 10
SHOE_INFER_GROUP_MAX = 5
DOUBLE_SEP = "============================"


class MultiProcessJob:
    """
    This class is used to execute multiprocess tasks and receive results.
    """
    TIME_OUT = 600
    MAX_SIZE = 20

    def __init__(self, module_name, pool_size, task_id, daemon=True, failed_raise=True):
        """
        Initiate a process task for the module.
        If the parameter is set to true, the subprocess task function has a return result
        :param module_name: module name
        :param pool_size: the process job pool size
        :param task_id: the task unique id
        :param daemon: daemon flag of the process. Bool
        :param failed_raise: indicates whether to raise when all job failed
        """
        self.task_id = task_id
        self.parent_process_id = os.getpid()
        self.job_dict = {}
        self.completed_job_set = set()
        self.module_name = module_name
        self.pool_size = pool_size if pool_size < self.MAX_SIZE else self.MAX_SIZE
        self.pool_queue = multiprocessing.Manager().Queue(self.pool_size)
        self.daemon = daemon
        self.origin_results = []
        self.failed_raise = failed_raise
        self.logger = logging.getLogger(self.module_name)

        for _ in range(self.pool_size):  # use something to fill queue
            self.pool_queue.put(False)

    def __getstate__(self):
        """
        Not pickle the job_dict, which contain other process handle
        """
        state = self.__dict__.copy()
        state['job_dict'] = {}
        return state

    def add_security_job(self, job_name, function, *func_args):
        """
        Add subprocess job and start job
        :param job_name: job name
        :param function: job function
        :param func_args: job args (except result_dict)
        """
        self.safe_get_one_completed_job()
        job = multiprocessing.Process(target=self._security_mask, name=job_name, daemon=self.daemon,
                                      args=(job_name, function, *func_args))
        self.job_dict.update({job_name: job})
        job.start()
        self.logger.info("Start the subjob %s. The run logs are recorded in the %s file.",
                         job_name, RUN_LOG_FORMAT.format(job.pid))

    def join_and_get_results(self):
        """
        Wait for the task to complete and verify the result
        :return: results, failed job list and os error exception
        """
        while len(self.job_dict) > len(self.completed_job_set):  # remaining unfinished tasks
            self.safe_get_one_completed_job()
        return self._check_result()

    def safe_get_one_completed_job(self):
        try:
            job_result = self.pool_queue.get(timeout=self.TIME_OUT)  # cannot get result in time will raise error
        except Exception as e:
            timeout_jobs = sorted(list(set(self.job_dict.keys()) - self.completed_job_set))
            clean_child_process(os.getpid())
            self.logger.error("The execution time of subjobs %s exceeds %s s.", timeout_jobs, self.TIME_OUT)
            raise InnerError(f"The execution time of subjobs {timeout_jobs} exceeds {self.TIME_OUT} s.") from e
        if not job_result:
            return
        self.origin_results.append(job_result)
        _, _, job_name = job_result
        self.completed_job_set.add(job_name)
        job = self.job_dict.get(job_name)
        job.join()
        job.terminate()

    def _check_result(self):
        """
        Check the subtask results and return format results and fail details info.
        The result format is {job_name: job_func_return}, the fail info format is {job_nam: failed error}
        :return: results, failed job details
        """
        failed_details = dict()
        results = dict()
        if not self.job_dict:
            # no multiprocessing tasks, returning directly
            return results, failed_details
        for error_flag, job_result, job_name in self.origin_results:
            if isinstance(error_flag, (Exception, KeyboardInterrupt)):
                failed_details.update({job_name: str(error_flag)})
                self.logger.error("Failed to execute %s job. The error is: [%s].", job_name, error_flag)
                continue
            results.update({job_name: job_result})
            self.logger.info("Job %s is successfully executed.", job_name)
        if len(failed_details) == len(self.job_dict):
            self.logger.error("All subjobs execute failed.")
            if self.failed_raise:
                first_key, first_value = next(iter(failed_details.items()))
                raise InnerError(
                    f"All subjobs execute failed. The first subjob: {first_key}, error reason is: {first_value}")
        elif failed_details:  # record the failed subjobs.
            self.logger.error("Some subjobs %s execute failed.", list(failed_details.keys()))
        return results, failed_details

    def _security_mask(self, job_name, function, *func_args):
        """
        Security mask for subprocess function
        :param job_name: job name
        :param function: function
        :param func_args: function args
        """
        init_echo_log()
        init_run_log(self.task_id)
        logger = logging.getLogger(self.module_name)
        logger.info("The parent process ID is %s. The subprocess job name is %s.", self.parent_process_id, job_name)
        status, result = None, None
        try:
            result = function(*func_args)
        except KeyboardInterrupt as error:
            logger.error("The %s job is executed failed. The reason is KeyboardInterrupt.", job_name)
            status = error
        except Exception as error:
            logger.error("The %s job is executed failed. The reason is %s", job_name, error)
            status = error
        else:
            logger.info("The %s job is executed successfully.", job_name)
            status = SuccessRet()
        try:
            self.pool_queue.put((status, result, job_name))
        except (OSError, EOFError) as error:
            logger.error("The %s job put final result failed. The reason is %s", job_name, error)


class BaseMatcher:
    """
    Pattern Error Base Matcher
    """

    def __init__(self, max_len=None):
        self.max_len = max_len
        self.key_info = ""

    @abc.abstractmethod
    def get_attr_info(self, attr_result: dict, line: str, event_code: str) -> dict:
        """
        Get attr_result
        :param attr_result: attr_regex match result dict
        :param line: the single line
        :param event_code: the event code
        :return: result dict
        """
        pass

    def _forward_in(self, patterns: List[str], text: str) -> (bool, int):
        """
        Match patten keywords from front to back in order
        :param patterns: the keywords list
        :param text: the original string
        :return: indicates whether the matching is successful, unmatched_keyword_id
        """
        idx = 0
        for keyword_id, keyword in enumerate(patterns):
            idx = text.find(keyword, idx)
            if idx == -1:  # not find
                return False, keyword_id
            idx = idx + len(keyword)
        self._set_in_key_info(text, idx)
        return True, -1

    def _set_in_key_info(self, text, idx):
        """
        Set key info
        :param text: the original string
        :param idx: index
        """
        index_sfx = text[idx:].find("\n")
        self.key_info = text[:idx + index_sfx] if index_sfx != -1 else text


class PatternMatcher(BaseMatcher):
    """
    Pattern Error Matcher
    """

    @staticmethod
    def match_attr(attr_regex: str, line: str) -> dict:
        """
        Match the line by the attr_regex
        :param attr_regex: attr_regex rule
        :param line: the single line
        :return: match result dict or {}
        """
        if not attr_regex or not line:
            return {}
        attr_re = re.search(attr_regex, line)
        return attr_re.groupdict() if attr_re else {}

    @staticmethod
    def _match_regex(pattern: str, text: str) -> bool:
        """
        Match "regex" pattern
        :param pattern: regex pattern string
        :param text: the original string
        :return: indicates whether the matching is successful
        """
        if not pattern:
            return False
        return re.search(pattern, text) is not None

    def compare(self, conf: dict, text: str) -> bool:
        """
        Match error patten conf, contain 'in, regex'
        :param conf: error pattern dict
        :param text: the original string
        :return: indicates whether the matching is successful
        """
        if self.max_len and len(text) > self.max_len:
            raise InnerError("The compare original text is too long, exceeds %s" % self.max_len)
        conf = conf or {}  # default conf dict
        return self._match_in(conf.get("in", []), text) or self._match_regex(conf.get("regex", ""), text)

    def get_attr_info(self, attr_result: dict, line: str, event_code: str) -> dict:
        """
        Get attr_result
        :param attr_result: attr_regex match result dict
        :param line: the single line
        :param event_code: the event code
        :return: result dict
        """
        result = {}
        if event_code and event_code.startswith(MINDIE_ERRCODE_COMMON):
            error_code = attr_result.get(ERROR_CODE, "")
            result["event_code"] = event_code + "_" + error_code
            if error_code:
                result[COMPLEMENT] = [error_code]
            return result
        return result

    def _match_in(self, patterns: list, text: str) -> bool:
        """
        Match "in" pattern. Multiple groups of keywords are supported. If one of them is met, the matching is
        successful
        :param patterns: in pattern keywords
        :param text: the original string
        :return: indicates whether the matching is successful
        """
        if not patterns:
            return False
        if isinstance(patterns[0], str):
            patterns_list = [patterns]
        else:
            patterns_list = patterns
        for single_pattern in patterns_list:
            if not self._match_in_for_single_patterns(single_pattern, text):
                continue
            return True
        return False

    def _match_in_for_single_patterns(self, patterns: List[str], text: str) -> bool:
        is_match, _ = self._forward_in(patterns, text)
        return is_match


class PatternSingleOrMultiLineMatcher(PatternMatcher):
    """
    Pattern Error Matcher for single-line or multiline
    """
    COMPLEMENT = "complement"
    ERROR_CODE = "error_code"

    def __init__(self, max_len=None, file_stream: io.StringIO = None, log_lines: list = None):
        super().__init__(max_len)
        self.file_stream = file_stream
        self.log_lines = log_lines
        self.cur_list_idx = 0 if log_lines is not None else None
        self.buffer = ""
        self.key_info = ""

    def update_stream(self, file_stream: io.StringIO = None):
        self.file_stream = file_stream
        self.buffer = ""
        self.key_info = ""

    def get_attr_info(self, attr_result: dict, line: str, event_code: str) -> dict:
        """
        Gets complement, key_info, and event_code of CANN custom event
        :param attr_result: attr_regex match result dict
        :param line: the single line
        :param event_code: the event code
        :return: result dict
        """
        result = dict()
        if event_code and event_code.startswith(PYTORCH_ERRCODE_COMMON):
            result["event_code"] = event_code + "_" + attr_result.get(self.ERROR_CODE, "")
            if attr_result.get(self.ERROR_CODE) and attr_result.get("module"):
                result[self.COMPLEMENT] = [attr_result.get("module", ""), attr_result.get(self.ERROR_CODE, "")]
            return result
        else:
            complement_value = attr_result.get(self.COMPLEMENT, "")
            if not complement_value or len(complement_value) < 2:
                return result
            # obtain component_name based on the second digit of complement_value
            component_name = regular_table.ERROR_CODE_COMPONENT.get(complement_value[1], "Unknown")
            result[self.COMPLEMENT] = [component_name, complement_value]
            result["event_code"] = CANN_ERRCODE_CUSTOM + "_" + complement_value

        keyinfo = self.read_multi_line(SHOW_LINES_NUM, line)
        if keyinfo:
            result["key_info"] = keyinfo
        return result

    def read_multi_line(self, line_num: int, text: str) -> str:
        """
        Read lines forward from the current position
        :param line_num: lines num for reading backwards
        :param text: the single-line original string
        :return: the multi-line original string
        """
        if self.log_lines is not None:
            return self._read_multi_line_in_list()
        if not self.file_stream:
            return ""
        buffer_list = [text]
        cur_tell = self.file_stream.tell()
        for _ in range(line_num):
            buffer_list.append(self.file_stream.readline())
        self.file_stream.seek(cur_tell)
        return ''.join(buffer_list)

    def update_line_index(self, line_idx: int):
        self.cur_list_idx = line_idx

    def _match_multi_line(self, patterns, text, line_num=LINE_NUM_OF_MULTILINE_MATCH) -> bool:
        """
        Match patten keywords from front to back in order for multi line
        :param patterns: the keywords list
        :param text: the single-line original string
        :param line_num: lines num for reading backwards
        :return: indicates whether the matching is successful
        """
        if self.log_lines is None:
            if not self.buffer:
                self.buffer = self.read_multi_line(line_num, text)
        else:
            self.buffer = self._read_multi_line_in_list()
        is_match, _ = self._forward_in(patterns, self.buffer)
        return is_match

    def _read_multi_line_in_list(self):
        """
        Read series of lines through the log line list
        :return: a combined list of multi lines
        """
        return "\n".join(self.log_lines[self.cur_list_idx:self.cur_list_idx + LINE_NUM_OF_MULTILINE_MATCH])

    def _match_in_for_single_patterns(self, patterns: List[str], text: str) -> bool:
        """
        Match patten keywords from front to back in order for single-line or multiline
        :param patterns: the keywords list
        :param text: the single-line original string
        :return: indicates whether the matching is successful
        """
        is_match, unmatched_keyword_id = self._forward_in(patterns, text)
        if is_match:
            return True
        if unmatched_keyword_id == 0:  # the first keyword was not matched, exiting directly
            return False
        return self._match_multi_line(patterns, text)


class MyRotatingFileHandler(logging.handlers.RotatingFileHandler):
    """
    This class is used to control logs permissions.
    """

    def doRollover(self):
        logging.handlers.RotatingFileHandler.doRollover(self)
        # set the new log file to 440
        safe_chmod(self.baseFilename, mode=0o640)
        # set the done log file to 440
        new_log_file = self.baseFilename + ".1"
        if os.path.exists(new_log_file):
            safe_chmod(new_log_file, mode=0o440)

    def handleError(self, record):
        _, err_v, _ = sys.exc_info()
        # err_v is the Error value
        echo.error("Log file record failed. Then reason is: %s", err_v)


class TimeBoundExtractor:
    """
    通用文件时间边界提取工具（支持大文件分块处理，内存高效）
    """

    def __init__(
            self, line_time_parser: Callable[[str], Optional[datetime.datetime]],
            line_validator: Callable[[str], bool] = lambda line: True, chunk_size: int = 4096
    ):
        # 行时间戳解析函数
        self.line_time_parser = line_time_parser
        # 行有效性验证函数
        self.line_validator = line_validator
        self.chunk_size = chunk_size

    @staticmethod
    def _process_reverse_chunk(chunk: str, line_buffer: List[str]) -> List[str]:
        """
        处理逆向块，维护行缓冲
        """
        # 分割块内容为行列表
        lines = chunk.split('\n')

        # 将缓冲区的行拼接到当前块的首行（正确顺序合并跨块行）
        if line_buffer and lines:
            lines[0] = line_buffer.pop() + lines[0]

        # 保存当前块的末行到缓冲（为前一个块的拼接做准备)
        if lines and lines[-1]:
            line_buffer.append(lines[-1])

        # 返回有效行（排除末行，因为已保存到缓冲）
        return [ln.strip() for ln in lines[:-1] if ln.strip()]

    @staticmethod
    def _process_forward_chunk(chunk: bytes, buffer: str) -> Tuple[str, List[str]]:
        """
        处理正向块，维护行缓冲
        """
        buffer += chunk.decode('utf-8', errors='replace')
        lines = buffer.splitlines()
        if not lines:
            return '', []
        if buffer[-1] not in ('\n', '\r'):
            return lines[-1], lines[:-1]
        return '', lines

    def get_time_bound(self, file_path: str, mode: str = "latest") -> Optional[datetime.datetime]:
        """
        获取文件的边界时间戳（最早或最晚）
        :param file_path: 文件路径
        :param mode: 'earliest'（正向搜索）或 'latest'（逆向搜索）
        :return: datetime.datetime 或 None（无法解析时）
        """
        if mode not in ("earliest", "latest"):
            raise ValueError("mode must be 'earliest' or 'latest'")

        with safe_read_open(file_path, 'rb') as f:
            return self._find_epoch_time(f, reverse=(mode == "latest"))

    def _find_epoch_time(self, file_obj, reverse: bool) -> Optional[datetime.datetime]:
        """
        根据搜索方向选择处理逻辑
        """
        return (
            self._reverse_search(file_obj)
            if reverse
            else self._forward_search(file_obj)
        )

    def _reverse_search(self, file_obj: BinaryIO) -> Optional[datetime.datetime]:
        """
        从文件末尾逆向搜索时间戳
        """
        file_obj.seek(0, os.SEEK_END)
        # 剩余需要读取的字节数
        remaining = file_obj.tell()
        # 处理跨块的行片段
        line_buffer: List[str] = []

        while remaining > 0:
            # 处理跨块的行片段
            chunk, remaining = self._read_reverse_chunk(file_obj, remaining)
            # 处理原始数据块，返回有效行列表（已处理截断行）
            lines = self._process_reverse_chunk(chunk, line_buffer)
            # 扫描有效行寻找时间戳
            dt = self._scan_valid_line(lines, reverse=True)
            if dt is not None:
                return dt

        # 扫描有效行寻找时间戳
        return self._check_remaining_buffer(line_buffer)

    def _forward_search(self, file_obj: BinaryIO) -> Optional[datetime.datetime]:
        """
        从文件开头正向搜索时间戳
        """
        buffer = ''
        while True:
            chunk = file_obj.read(self.chunk_size)
            # 处理缓冲区和数据块，返回新缓冲区和完整行列表
            buffer, lines = self._process_forward_chunk(chunk, buffer)
            dt = self._scan_valid_line(lines)
            if dt is not None:
                return dt
            if not chunk and not buffer:
                break
        return None

    def _read_reverse_chunk(self, file_obj: BinaryIO, remaining: int) -> Tuple[str, int]:
        """
        逆向读取数据块
        """
        # 计算本次读取大小和偏移量（逆向：从后往前读）
        read_size = min(self.chunk_size, remaining)
        offset = max(remaining - read_size, 0)
        file_obj.seek(offset)
        # 更新剩余需要处理的字节数
        remaining -= read_size
        # 读取并解码数据块
        chunk = file_obj.read(read_size).decode('utf-8', errors='replace')
        return chunk, remaining

    def _scan_valid_line(self, lines: List[str], reverse: bool = False) -> Optional[datetime.datetime]:
        """
        扫描行列表查找有效时间戳
        """
        iterator = reversed(lines) if reverse else iter(lines)
        for line in iterator:
            if not self.line_validator(line):
                continue
            dt = self.line_time_parser(line)
            if dt is not None:
                return dt
        return None

    def _check_remaining_buffer(self, line_buffer: List[str]) -> Optional[datetime.datetime]:
        """
        检查逆向搜索后的残留行缓冲
        """
        if not line_buffer:
            return None
        return self._scan_valid_line(
            [line.strip() for line in line_buffer],
            reverse=True
        )


def read_version_info():
    """
    read version info and build time
    """
    src_path = Path(__file__).absolute().parent.parent
    version_file = src_path.joinpath("Version.info")
    if os.path.islink(version_file):
        raise FileOpenError(f"{os.path.basename(version_file)} should not be a symbolic link file.")
    with open(os.path.realpath(version_file)) as file_stream:
        version_info = file_stream.read(VERSION_FILE_READ_LIMIT).splitlines()
    return version_info


def get_version():
    """
    Get the version info
    """
    version_info = read_version_info()  # e.g: ["6.0.0", "2024-10-21"]
    if not version_info:
        raise InfoNotFoundError("Failed to obtain the version info.")
    return version_info[0]


def str_param_len_check(input_str: str):
    if len(input_str) > MAX_PARAM_LEN:
        raise argparse.ArgumentTypeError('the input string length cannot over %d', MAX_PARAM_LEN)
    return input_str


def get_build_time():
    """
    Get the build time
    """
    version_info = read_version_info()  # e.g: ["6.0.0", "2024-10-21"]
    if len(version_info) < 2:
        raise InfoNotFoundError("Failed to obtain the build time.")
    return version_info[1]


def chinese_check(char: str) -> bool:
    """
    Check whether the value is a Chinese character.
    [\u4e00, \u9fa5] is Chinese characters
    [\u3000, \u303F] or [\uFF00, \uFFEF] is Chinese punctuation
    :param char: character
    :return: Chinese or not
    """
    if len(char) != 1:  # 长度不为1时返回false
        return False
    return '\u4e00' <= char <= '\u9fa5' or '\u3000' <= char <= '\u303F' or '\uFF00' <= char <= '\uFFEF'


def white_check(check_str: str, white_list: str, allow_zh: bool = False):
    """
    White check for string
    :param check_str: the origin string
    :param white_list: white list str
    :param allow_zh: whether Chinese characters are supported
    :return: Check whether the check is passed
    """
    for char in check_str:
        if char not in white_list and not (allow_zh and chinese_check(char)):
            return False
    return True


def code_check(code: str):
    """
    Check the cmd input code
    :param code: fault code string
    :return: code string
    """
    if len(code) < 1 or len(code) > MAX_CODE_LEN:
        raise argparse.ArgumentTypeError(
            f"The fault code is invalid.\n"
            f"The fault code length exceeds the maximum code length({MAX_CODE_LEN}).")
    if not white_check(code, CODE_WHITE_LIST):
        raise argparse.ArgumentTypeError(
            "The fault code is invalid.\n"
            "The fault code can contain only digits, uppercase and lowercase letters, "
            "and following characters: ['-', '_']")
    return code


def path_check(path: str):
    """
    Check the cmd input_path
    :param path: the data path
    :return: data real path after check
    """
    if path is None:
        return path
    if not white_check(path, PATH_WHITE_LIST_LIN):
        raise argparse.ArgumentTypeError(
            "The path is invalid.\n"
            "The path can contain only digits, uppercase and lowercase letters, "
            "and following characters: ['~', '+', '-', '_', '.', ' ']")
    if len(path) < 1 or len(path) > MAX_PATH_LEN:
        raise argparse.ArgumentTypeError(
            f"The path is invalid.\n"
            f"The path length exceeds the maximum path length({MAX_PATH_LEN}).")
    try:
        real_path = check_symlink(path)
    except (PathError, FileNotExistError) as err:
        raise argparse.ArgumentTypeError(f"Failed to check path: {err}")
    return real_path


def check_symlink(path):
    """
    Check symlink
    :return: real path
    """
    if not path:
        raise PathError()
    # 解析"~"
    expanded_path = os.path.expanduser(path)
    abs_path = os.path.abspath(expanded_path)
    if not os.path.exists(abs_path):
        raise FileNotExistError()
    real_path = os.path.realpath(abs_path)
    if abs_path != real_path:
        raise PathError("The path should not be a symbolic link file.")
    return real_path


def file_check(path: str):
    """
    Check the cmd input_path is a file or not
    :param path: the data path
    :return: data real path after check
    """
    path = path_check(path)
    if not os.path.isfile(path):
        raise argparse.ArgumentTypeError("The path should be a file.")
    return path


def dir_check(path: str):
    """
    Check the cmd input_path is a directory or not
    :param path: the data path
    :return: data real path after check
    """
    path = path_check(path)
    if not os.path.isdir(path):
        raise argparse.ArgumentTypeError("The path should be a directory.")
    return path


def file_or_dir_check(path: str):
    """
    Check the input argument whether a legal directory or filename
    :param path: a path
    :return: str of real path
    """
    path = path_check(path)
    if not os.path.isdir(path) and not os.path.isfile(path):
        raise argparse.ArgumentTypeError("The path should be either a file or a directory")
    return path


def write_path_check(path: str):
    """
    Check the cmd output_path, add check user owner
    :param path: the data path
    :return: data path after check owner
    """
    real_path = dir_check(path)
    if not check_owner(real_path):
        raise argparse.ArgumentTypeError("The path is not owned by current user or root.")
    return real_path


def check_owner(path):
    """
    Check the path owner
    :param path: the input path
    :return: whether the path owner is current user or root
    """
    path_stat = os.stat(path)
    path_owner, path_gid = path_stat.st_uid, path_stat.st_gid
    user_check = path_owner == os.getuid() and path_owner == os.geteuid()
    return path_owner == 0 or path_gid in os.getgroups() or user_check


def safe_read_line(file_path: str):
    """
    safe read file line
    :param file_path: the log file
    :return: log line
    """
    with safe_read_open(file_path, "r", encoding="UTF-8") as file_stream:
        while True:
            line = file_stream.readline()
            if not line:
                break
            yield line.strip()


def safe_read_open(file, *args, **kwargs):
    """
    Safe open file.
    Function will check if the file is a soft link or the file size is too large
    :param file: file path
    :param args: func parameters
    :param kwargs: func parameters
    :return: file_stream
    """
    if os.path.islink(file):
        raise FileOpenError(f"The {os.path.basename(file)} should not be a symbolic link file.")
    file_real_path = os.path.realpath(file)
    file_stream = open(file_real_path, *args, **kwargs)
    file_info = os.stat(file_stream.fileno())
    if file_info.st_size > MAX_SIZE:
        file_stream.close()
        raise FileOpenError(f"The size of {os.path.basename(file)} should be less than {MAX_SIZE >> MB_SHIFT} MB.")
    return file_stream


def safe_read_open_with_size(file: str, size: int = 100 * 1024, *args, **kwargs):
    """
    Safe open file and read specify size.
    If file size <= specify size, read all; if file size > specify size, read the latest specify size data
    :param file: file path
    :param size: read size (Byte)
    :param args: func parameters
    :param kwargs: func parameters
    :return: file_stream
    """
    if os.path.islink(file):
        raise FileOpenError(f"The {os.path.basename(file)} should not be a symbolic link file.")
    file_real_path = os.path.realpath(file)
    kwargs.update({"errors": "ignore"})
    file_stream = open(file_real_path, *args, **kwargs)
    file_info = os.stat(file_stream.fileno())
    if file_info.st_size > size:
        file_stream.seek(file_info.st_size - size + 1)
    return file_stream


def safe_generate_or_merge_json_file(file_path, new_data):
    existing_data = {}
    if os.path.exists(file_path):
        try:
            existing_data = safe_read_json(file_path)
        except FileOpenError:
            pass

    # Merge data (new data will overwrite the same keys in old data)
    merged_data = {**existing_data, **new_data}
    with safe_write_open(file_path, mode="w+", encoding="utf-8") as file_stream:
        file_stream.write(json.dumps(merged_data, ensure_ascii=False))
        file_stream.write('\r\n')


def safe_write_open(file, open_flags=os.O_WRONLY | os.O_CREAT | os.O_TRUNC, open_mode=0o640, *args, **kwargs):
    """
    Safe open file for writing
    :param file: file path
    :param open_flags: the 'flags' parameter for 'os.open' func, default os.O_WRONLY | os.O_CREAT | os.O_TRUNC
    :param open_mode: the 'mode' parameter for 'os.open' func, default 0o640
    :param args: func parameters
    :param kwargs: func parameters
    :return: file_stream
    """
    if os.path.islink(file):
        raise FileOpenError(f"The {os.path.basename(file)} should not be a symbolic link file.")
    file_real_path = os.path.realpath(file)
    if not os.path.exists(file_real_path):
        return os.fdopen(os.open(file_real_path, open_flags, open_mode), *args, **kwargs)
    # check file owner
    if not check_owner(file_real_path):
        raise FileOpenError(f"{os.path.basename(file)} is not owned by current user or root.")
    return os.fdopen(os.open(file_real_path, open_flags, open_mode), *args, **kwargs)


def safe_read_json(file: str, *args, **kwargs) -> dict:
    """
    Safe read json file.
    Will use safe open func to open json file
    :param file: file path
    :param args: func parameters
    :param kwargs: func parameters
    :return: dict
    """
    with safe_read_open(file, "r", encoding='utf-8') as file_stream:
        try:
            data = json.load(file_stream, *args, **kwargs)
        except Exception as err:
            raise FileOpenError(f"Open {os.path.basename(file)} json failed: {err}") from err
    return data


def load_json_data(pkg_data_path):
    """
    Load JSON file
    :param pkg_data_path: JSON file path
    :return: data in JSON format
    """
    if not os.path.exists(pkg_data_path):
        raise FileNotExistError("The JSON file path %s does not exist." % pkg_data_path)
    events_json = safe_read_json(pkg_data_path)
    if not events_json or not isinstance(events_json, dict):
        raise InfoIncorrectError("Failed to load data in JSON format.")
    return events_json


def safe_chmod(file: str, mode):
    """
    Safe chmod file
    Will use safe chmod func to chmod file
    :param file: file path
    :param mode: file mode
    """
    with safe_read_open(file) as file_stream:
        os.fchmod(file_stream.fileno(), mode)


def get_user_info():
    """
    Get the user info for operate log
    :return: user host, user name
    """
    user_name = getpass.getuser()
    user_host = get_user_host()
    try:
        current_terminal = os.ttyname(sys.stdout.fileno())
    except OSError:
        current_terminal = "UnKnown"
    user_id = os.getuid()
    return f'{user_host}@{current_terminal}', f'{user_name}({user_id})'


def get_user_host():
    """
    Get the user host info
    :return: user host
    """
    user_host = 'localhost'
    who_cmd = subprocess.run(['/usr/bin/who', '-m'], shell=False, stdout=subprocess.PIPE)
    if who_cmd.returncode != 0:
        return user_host
    who = who_cmd.stdout.decode()
    if not who:
        return user_host
    current_logged = who.split()
    host = current_logged[-1].strip("()")
    if host not in (':0.0', ':0'):
        user_host = host
    return user_host


def safe_list_dir(file_dir: str, max_loop_size: int = WORKER_MAX_NUM):
    """
    :param file_dir: dir file path
    :param max_loop_size: max loop number, default 1000000
    :return: files path
    """
    dir_list = os.listdir(file_dir)
    if len(dir_list) > max_loop_size:
        raise FileTooLarge("The number of files in directory %s exceeds %s." % (file_dir, max_loop_size))
    return dir_list


def safe_walk(file_dir: str, dir_depth: int = DIR_MAX_DEPTH, max_loop_size: int = WORKER_MAX_NUM):
    """
    Traverse files by specifying the directory depth
    :param file_dir: dir file path
    :param dir_depth: dir depth, default 20
    :param max_loop_size: max loop number, default 1000000
    :return: file path
    """
    loop_num = 0
    input_path_depth = file_dir.rstrip().count(os.sep)
    for root, dirs, files in os.walk(file_dir, onerror=raise_error):
        if loop_num > max_loop_size:
            raise FileTooLarge("The number of files in directory %s exceeds %s." % (file_dir, max_loop_size))
        root_depth = root.rstrip().count(os.sep)
        if root_depth - input_path_depth > dir_depth:
            dirs[:] = []  # 清空 dirs 列表，不再继续向里查找
            fd_logger.warning("The search depth of the current dir has reached %s. "
                              "The part that is too deep will be ignored.", dir_depth)
            continue
        yield root, dirs, files
        loop_num += 1


def raise_error(error):
    """
    Raise the captured exception, os.Walk callback function
    :param error: exceptions caught
    """
    raise error


def get_parse_json(kg_parser_file: str) -> dict:
    """
    Get parse json file
    :param kg_parser_file: json file in the parsed results
    :return: parse json
    """
    if not os.path.exists(kg_parser_file):
        raise FileNotExistError(f'The {kg_parser_file} dir is not exist.')
    kg_json = safe_read_json(kg_parser_file)
    return kg_json if kg_json else {}


def get_component_version(kg_parsed_results):
    """
    Get version of Driver, Firm, NNAE and CANN
    :param kg_parsed_results: the parsed results
    :return: dict of component version
    """
    version_dict = dict()
    for label in regular_table.VERSION_INFO_LABEL_LIST:
        version_info = kg_parsed_results.get(label, "")
        if version_info:
            version_dict.update({label: version_info})
    return version_dict


def log_filter_character_defence(log_record: logging.LogRecord):
    """
    Log filter, to filter escape character, for example: '\n', '\r', '\b', '\f'
    :param log_record: logging LogRecord
    """
    log_record.msg = repr(str(log_record.msg)).strip("'\"")
    log_record.args = tuple([repr(str(arg)).strip("'\"") for arg in log_record.args])


def log_filter_for_run(log_record: logging.LogRecord) -> bool:
    """
    Log filter for format
    :param log_record: logging LogRecord
    :return: bool
    """
    if log_record.levelname == "WARNING":
        log_record.levelname = "WARN"
    log_record.levelname = f"[{log_record.levelname}]"
    log_filter_character_defence(log_record)
    return True


def log_filter_for_operation(log_record: logging.LogRecord) -> bool:
    """
    Log filter for OP_LOG
    :param log_record: logging LogRecord
    :return: bool
    """
    log_record.host_name, log_record.user_name = get_user_info()
    log_filter_character_defence(log_record)
    return True


def log_filter_for_echo(log_record: logging.LogRecord) -> bool:
    """
    Log filter for echo
    :param log_record: logging LogRecord
    :return: bool
    """
    level_map = {
        logging.WARNING: "Warn: ",
        logging.ERROR: "Error: "
    }
    log_record.msg = f"{level_map.get(log_record.levelno, '')}{log_record.msg}"
    return True


def init_run_log(task_id):
    """
    Init the all logger, Contains the QueueListener and all QueueHandler.
    Run log uses the queueListener to implement multiprocess logging
    :param task_id: this task id, use the time stamp
    """
    task_log_path = os.path.join(LOG_DIR_PATH, task_id)
    os.makedirs(task_log_path, 0o700, exist_ok=True)
    log_file = RUN_LOG_FORMAT.format(os.getpid())
    file_handler = MyRotatingFileHandler(os.path.join(task_log_path, log_file),
                                         maxBytes=LOG_MAX_SIZE, backupCount=LOG_MAX_BACKUP_COUNT)
    safe_chmod(os.path.join(task_log_path, log_file), mode=0o640)
    file_handler.setFormatter(logging.Formatter(DETAIL_FORMAT))
    for name in LOG_MODULE:
        logger = logging.getLogger(name)
        clean_logger(logger)
        logger.addHandler(file_handler)
        logger.addFilter(log_filter_for_run)
        logger.setLevel(logging.INFO)


def init_operation_log():
    """
    Init the operation log.
    """
    log_file = os.path.join(HOME_PATH, OP_LOG_FILE)
    file_handler = MyRotatingFileHandler(log_file, maxBytes=LOG_MAX_SIZE, backupCount=LOG_MAX_BACKUP_COUNT)
    safe_chmod(log_file, mode=0o640)  # the file is generated after the handler is initialized.
    file_handler.setFormatter(logging.Formatter(OPERATION_FORMAT))
    op_log = logging.getLogger("FD_OP")
    op_log.addHandler(file_handler)
    op_log.addFilter(log_filter_for_operation)
    op_log.setLevel(logging.INFO)


def init_echo_log(log_lever: int = logging.INFO):
    """
    Init the echo log. Use to print version and err info on terminal screen
    """
    echo_handler = logging.StreamHandler(sys.stdout)
    echo_handler.setFormatter(logging.Formatter('%(message)s'))
    echo_logger = logging.getLogger("ECHO")
    clean_logger(echo_logger)
    echo_logger.addHandler(echo_handler)
    echo_logger.addFilter(log_filter_for_echo)
    echo_logger.setLevel(log_lever)


def init_home_path() -> bool:
    """
    Init the home path, which can be set through env variable.
    Then create the blacklist and custom-entity configuration files in home path.
    Run logs and operation logs will also store in this folder.
    :return: init home path success flag
    """
    # check the HOME_PATH
    if HOME_PATH in INVALID_HOME_PATH_LIST:
        echo.error("Environment variable %s is set to %s. Please set it to another valid path", ENV_VAR_KEY, HOME_PATH)
        return False
    if HOME_PATH == DEFAULT_HOME_PATH:
        _create_default_conf_file()
        return True
    return _init_home_path_by_env()


def _create_default_conf_file():
    """
    Creat the blacklist conf file and user-defined entity conf file
    """
    os.makedirs(HOME_PATH, 0o700, exist_ok=True)
    os.makedirs(LOG_DIR_PATH, 0o700, exist_ok=True)
    for conf_file in (DEFAULT_USER_CONF, DEFAULT_BLACKLIST_CONF):
        if not os.path.exists(conf_file):  # write an empty dict to json
            with safe_write_open(conf_file, mode="w+") as file_stream:
                file_stream.write("{}")


def _init_home_path_by_env() -> bool:
    """
    Check the home path which set by env variable
    :return: check result
    """
    if not os.path.exists(HOME_PATH):
        echo.error("The %s path must exist.", ENV_VAR_KEY)
        return False
    if not os.path.isdir(HOME_PATH):
        echo.error("The %s path already exists but is not a folder.", ENV_VAR_KEY)
        return False
    if not check_owner(HOME_PATH):
        echo.error("The owner of %s path does not belong to the root user or the user who executes the program.",
                   ENV_VAR_KEY)
        return False
    try:
        _create_default_conf_file()
    except OSError as error:
        echo.error("The %s path init failed because %s", ENV_VAR_KEY, error)
        return False
    return True


def clean_run_log_path():
    """
    Clean the run log dir file if the folder's num exceeds the maximum or dir size exceeds the maximum
    when one task start.
    If the RUN_LOG task num > RUN_LOG_DIR_MAX_NUM, will clean RUN_LOG_DIR_CLEAN_NUM task log.
    If the RUN_LOG path > RUN_LOG_DIR_MAX_SIZE, will clean RUN_LOG_DIR_CLEAN_SIZE in RUN LOG dir.
    """
    # num clean
    folder_list = safe_list_dir(LOG_DIR_PATH)
    folder_list = sorted(folder_list)
    if len(folder_list) >= RUN_LOG_DIR_MAX_NUM:
        for clean_file in folder_list[:RUN_LOG_DIR_CLEAN_NUM]:
            file_path = os.path.join(LOG_DIR_PATH, clean_file)
            if os.path.exists(file_path):
                shutil.rmtree(file_path)

    # size clean
    folder_list = safe_list_dir(LOG_DIR_PATH)
    folder_list = sorted(folder_list)
    run_log_size = get_file_or_folder_size(LOG_DIR_PATH)
    if run_log_size < RUN_LOG_DIR_MAX_SIZE:
        return
    clean_path_set = set()
    clean_size = 0
    for folder_name in folder_list:
        folder_path = os.path.join(LOG_DIR_PATH, folder_name)
        clean_size += get_file_or_folder_size(folder_path)
        clean_path_set.add(folder_path)
        if run_log_size - clean_size <= RUN_LOG_DIR_MAX_SAVE_SIZE:
            break
    for clean_path in clean_path_set:
        if os.path.exists(clean_path):
            shutil.rmtree(clean_path)


def get_file_or_folder_size(specified_path):
    """
    Get the file or folder size
    :param specified_path: the file or folder path
    :return: folder path size
    """
    if os.path.isfile(specified_path):
        return os.stat(specified_path).st_size
    total_size = 0
    if not os.path.isdir(specified_path):
        return total_size
    for dir_path, _, file_names in safe_walk(specified_path):
        for file_name in file_names:
            file_path = os.path.join(dir_path, file_name)
            total_size += os.stat(file_path).st_size
    return total_size


def clean_logger(logger: logging.Logger):
    """
    Clean the logger configurations when child process inherits logger from a parent process
    :param logger: the logging.Logger of child process
    """
    for used_handler in logger.handlers[:]:
        logger.removeHandler(used_handler)
    for used_filter in logger.filters[:]:
        logger.removeFilter(used_filter)


def clean_child_process(parent_pid):
    """
    Clean all child process when main process exit
    :param parent_pid: the main process pid
    """
    for pid in get_all_descendant_pids(parent_pid, recursion_count=0):
        try:
            os.kill(pid, signal.SIGTERM)
        except (ValueError, OSError):
            continue


def get_children_pids(pid):
    """
    Get the children pid list of a parent process
    :param pid: the parent pid
    :return: children pid list
    """
    with safe_read_open(f'/proc/{pid}/task/{pid}/children', "r") as f:
        children = [int(pid) for pid in f.read().strip().split() if pid.isdigit()]
    return children


def get_all_descendant_pids(pid, recursion_count):
    """
    Get all the children, grandchildren and deeper of a parent process
    :param pid: the parent pid
    :param recursion_count: the number of recursion layers
    :return: children pid list
    """
    descendants = []
    if recursion_count > MAX_PROCESS_DEEP:
        return descendants
    try:
        children = get_children_pids(pid)
    except FileNotFoundError:
        return descendants
    for child_pid in children:
        descendants.extend(get_all_descendant_pids(child_pid, recursion_count + 1))
    descendants.extend(children)
    return descendants


def check_and_format_time_str(time_str, timezone_trans_flag=False):
    """
    Check whether the character string is a time character string (containing only digits,'-',':','.','T','+','').
    If the verification is successful, the value is formatted into "YYYY-MM-DD hh-mm-ss.******"
    :param time_str: the origin time str
    :param timezone_trans_flag: the timezone trans flag
    :return: A unified time str if valid
    """
    for char in time_str:
        if char not in TIME_WHITE_LIST:
            return ""

    # replace "YYYY-MM-DD-hh:mm:ss.***.***" to "YYYY-MM-DD-hh:mm:ss.******"
    modified_time_str = re.sub(r'\.(\d{1,5})\.(\d{1,5})', r'.\1\2', time_str)

    formats = [
        "%Y-%m-%d-%H:%M:%S.%f",  # 格式1：带连字符分隔
        "%Y-%m-%dT%H:%M:%S",  # 格式2：ISO格式不带时区
        "%Y-%m-%dT%H:%M:%S%z",  # 格式3：ISO格式带时区
        "%Y-%m-%d %H:%M:%S.%f%z",  # 格式4：空格分隔带毫秒和时区
        "%Y-%m-%d %H:%M:%S.%f",  # 格式5：空格字符分割
        "%Y%m%d%H%M%S"  # 格式6：紧凑型日期时间
    ]
    dt = None
    for fmt in formats:
        try:
            dt = datetime.datetime.strptime(modified_time_str, fmt)
            break
        except ValueError:
            continue
    if not dt:
        return ""
    # Convert to UTC time and format when contains utc and timezone trans flag is true
    if dt.tzinfo is not None and timezone_trans_flag:
        dt = dt.astimezone(datetime.timezone.utc)
    else:
        # Assume no time zone and treat it as UTC
        dt = dt.replace(tzinfo=datetime.timezone.utc)

    return dt.strftime("%Y-%m-%d %H:%M:%S.%f")


def check_file_num_and_size(files, logger, file_num=FILE_MAX_NUM, file_size=FILES_SIZE_THRESHOLD):
    """
    Check the number and size of files
    :param files: all files list
    :param logger: logger
    :param file_num: files num, default 1000000
    :param file_size: files size, default 5G
    """
    if len(files) > file_num:
        logger.error("The number of files is too large to be parsed.")
        raise FileTooLarge("The number of files is too large to be parsed.")
    if all_file_size_sum(files) > file_size:
        echo.warning(f"All files size exceeds %s GB, it may take a long time to parse these files.",
                     file_size >> GB_SHIFT)
        logger.warning("All files size exceeds %s GB, it may take a long time to parse these files.",
                       file_size >> GB_SHIFT)


def check_scikit_learn_version(threshold: str = "1.3.0"):
    """
    Check scikit-learn version >= threshold
    """
    import sklearn
    sklearn_version = sklearn.__version__
    if sklearn_version.split(".") < threshold.split("."):
        raise ImportError(
            f"The scikit-learn version ({sklearn_version}) is too low. "
            "Please upgrade to version 1.3.0 or higher."
        )


def all_file_size_sum(files):
    """
    All files size sum
    :param files: all files list
    :return: all files size sum
    """
    sum_size = 0
    if not files:
        return 0
    for file in files:
        if os.path.isfile(file):
            stats = os.stat(file)
            sum_size += stats.st_size
    return sum_size


def merge_occurrence(dst_dict: dict, src_dict: dict):
    """
    Merge occurrence between two events, always keep the quantity lower than the limit
    :param dst_dict: the dict that will be expanded
    :param src_dict: the dict that will be extracted and extended to res_dict
    """
    occurrence = "occurrence"
    if occurrence not in dst_dict or occurrence not in src_dict:
        return
    record_limit = 50
    remaining_available_record = record_limit - len(dst_dict[occurrence])
    if remaining_available_record <= 0:
        return
    dst_dict[occurrence].extend(src_dict[occurrence][:remaining_available_record])


def get_log_module_and_time(log_line: str):
    """
    Get the log time and module of plog
    """
    # The log example:
    # "[INFO/ERROR/***] XXXX(**,**):20yy-mm-dd-xx:xx:xx.xxx.xxx ********************"
    log_module = log_line.split()[1].split("(")[0].strip()
    times = log_line.split()[1].split(")")[1].strip(":")
    log_time = times[:-4] + times[-3:]  # "20yy-mm-dd-xx:xx:xx.xxx.xxx" -> "20yy-mm-dd-xx:xx:xx.xxxxxx"
    return log_module, log_time


def validate_type(data, param_type: type, filed_name: str):
    """
    Validate data type, raise error if invalid
    :param data: data to be validated
    :param param_type: type to be verified
    :param filed_name: original filed name
    """
    if not isinstance(data, param_type):
        raise ParamError(
            "Invalid parameter type for '{}', it should be '{}'.".format(filed_name, param_type.__name__))


def validate_list_length(param_list: list, limit: int, label: str):
    """
    Validate the list length
    :param param_list: the list that is going to be validated
    :param limit: the limit for the list
    :param label: the label to fill in the error echo
    :return:
    """
    input_list_size = len(param_list)
    if input_list_size > limit:
        raise ParamError("The size of {} exceeds the limit, which is {}, "
                         "whereas the current size is {}.".format(label, limit, input_list_size))


def generate_task_id():
    """
    Get the task id for this job
    :return: unique task id
    """
    china_tz = datetime.timezone(datetime.timedelta(hours=8))
    timestamp = datetime.datetime.now(tz=china_tz).strftime("%Y%m%d%H%M%S%f")
    unique_id = uuid.uuid4()
    return f"{timestamp}_{unique_id}"


@dataclass
class Field:
    type: Type
    mandatory: bool = True
    allow_empty: bool = False
    default: Optional[Any] = None
    sub_schema: Optional[Union[Dict[str, Any], Type]] = None
    choices: Optional[Union[List, range]] = None
    size_limit: Optional[int] = None
    sub_element_type: Optional[Type] = None
    custom_validator: Optional[Callable[[Any, str], Any]] = None
    statistic_callback: Optional[Callable[[Any], None]] = None


class SchemaValidator:
    def __init__(self, schema):
        self.schema = schema
        self.cur_key_path = ""

    @staticmethod
    def _validate_type(value, cur_key_path, value_type: type):
        if not isinstance(value, value_type):
            raise ParamError("Type mismatch for '{}', expected {}, got {} instead"
                             .format(cur_key_path, value_type.__name__, type(value).__name__))

    def validate(self, data: dict, schema: dict = None, root: str = ""):
        """
        A series of validation process for various parameters
        :param data: data source
        :param schema: a schema of how the data is structured
        :param root: root path
        """
        schema = schema or self.schema
        for key, rule in schema.items():
            self.cur_key_path = f"{root}.{key}" if root else f"{key}"
            if key not in data:
                self._validate_existence(rule.mandatory)
                continue
            value = data[key]
            value_type = rule.type
            self._validate_type(value, self.cur_key_path, value_type)
            if value_type in (str, list, tuple, dict, set) and len(value) == 0:
                self._validate_empty(data, key, rule.allow_empty, rule.default)
                continue
            self._validate_choices(value, value_type, rule.choices)
            if value_type in (list, str, dict):
                self._validate_length(value, rule.size_limit)
            if rule.custom_validator is not None:
                new_value = rule.custom_validator(value, self.cur_key_path)
                data[key] = new_value
            if rule.sub_schema:
                self._validate_sub_schema(value, rule.sub_schema, value_type)
            elif rule.sub_element_type:
                self._validate_sub_element_type(value, rule.sub_element_type)
            if rule.statistic_callback is not None:
                rule.statistic_callback(value)

    def _validate_existence(self, is_mandatory: bool):
        """
        Validate whether the absence is legal
        """
        if is_mandatory:
            raise ParamError("Missing required field: '{}'".format(self.cur_key_path))

    def _validate_empty(self, data: dict, key: str, allow_empty: bool, default):
        """
        Validate whether the empty state is legal and fill in a default value if it is set
        """
        if not allow_empty:
            raise ParamError("Empty value not allowed for {}".format(self.cur_key_path))
        if default is not None:
            data[key] = default

    def _validate_choices(self, value, value_type: type, choices: Union[list, range]):
        """
        Validate whether the value is in range or pre-defined choices
        """
        if isinstance(choices, range) and value_type in (int, float) and value not in choices:
            raise ParamError("Validation error at '{}': Value '{}' not in the allowed range: [{}-{}]"
                             .format(self.cur_key_path, value, choices.start, choices.stop - 1))

        if choices is not None and value not in choices:
            raise ParamError("Validation error at '{}': Value '{}' not in the allowed choices: {}"
                             .format(self.cur_key_path, value, choices))

    def _validate_length(self, value: Union[list, dict, str], size_limit: int):
        """
        Validate whether the length exceeds the limit
        """
        if size_limit is not None and len(value) > size_limit:
            raise ParamError("Value size exceeds the limit: {}, expected no greater than {}, got {}"
                             .format(self.cur_key_path, size_limit, len(value)))

    def _validate_sub_schema(self, value, sub_schema, value_type: type):
        """
        Recursively validating sub schema
        """
        if sub_schema and value_type is dict:
            original_cur_key_path = self.cur_key_path
            self.validate(value, sub_schema, self.cur_key_path)
            self.cur_key_path = original_cur_key_path
        elif sub_schema and value_type is list:
            for idx, item in enumerate(value):
                original_cur_key_path = self.cur_key_path
                self.cur_key_path = f"{self.cur_key_path}[{idx}]"
                if not isinstance(item, dict):
                    raise ParamError("Validation error at '{}': Type mismatch, expected {}, got {} instead"
                                     .format(self.cur_key_path, dict.__name__, type(item).__name__))
                self.validate(item, sub_schema, self.cur_key_path)
                self.cur_key_path = original_cur_key_path

    def _validate_sub_element_type(self, container, sub_type: type):
        """
        Validate the type of sub elements for the container.
        """
        if isinstance(container, list):
            for idx, element in enumerate(container):
                self._validate_type(element, f"{self.cur_key_path}[{idx}]", sub_type)

        if isinstance(container, dict):
            for key, value in container.items():
                self._validate_type(value, f"{self.cur_key_path}.{key}", sub_type)


def sort_results_by_id(multiprocess_results):
    """
    Sort results by its idx 1 element, usually used an int as the criteria
    """
    results_and_idx = list(multiprocess_results.values())
    length_requirement = 2
    if not results_and_idx or len(next(iter(results_and_idx))) < length_requirement:
        results = [val for val, _ in results_and_idx]
        return results
    results_and_idx.sort(key=lambda result: result[1])
    results = [val for val, _ in results_and_idx]
    return results


def convert_sets_to_lists(obj):
    """
    Convert all sets in a nested data structure to lists.
    """
    if isinstance(obj, set):
        return list(obj)
    elif isinstance(obj, dict):
        return {key: convert_sets_to_lists(value) for key, value in obj.items()}
    elif isinstance(obj, list):
        return [convert_sets_to_lists(element) for element in obj]
    else:
        return obj


def collect_parse_results(path, key_words: str):
    path_list = [os.path.join(path, target) for target in safe_list_dir(path) if key_words in target]
    return path_list


def init_sdk_task():
    try:
        clean_run_log_path()
    except (FileNotFoundError, PermissionError):
        # This error is not reported as it will not affect the procedure
        # Whereas it has a high probability of occurrence in a multiprocessing scenario
        pass
    task_id = generate_task_id()
    init_echo_log(log_lever=logging.CRITICAL)
    init_run_log(task_id)
    return task_id
