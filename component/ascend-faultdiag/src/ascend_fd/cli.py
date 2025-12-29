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
import warnings
import sys
import logging
import argparse

from ascend_fd.configuration.config import NAME, COMPONENT
from ascend_fd.controller import router
from ascend_fd.utils.tool import (get_version, file_check, dir_check, file_or_dir_check, str_param_len_check,
                                  code_check, write_path_check, ITEM_CHOICES, init_echo_log, init_home_path,
                                  init_operation_log, init_run_log, clean_run_log_path, clean_child_process,
                                  DIAG_CHOICES, generate_task_id)

op_logger = logging.getLogger("FD_OP")
echo = logging.getLogger("ECHO")


def command_line():
    """
    The command line interface. Commands contain:
    1. version
    2. parse
      -i, --input_path, the input path of origin data file.
      -o, --output_path, the output path of parsed data file.
      -p, --performance, indicates whether the performance evaluation module is enabled.
      --bmc_log, the input path of bmc log file
      --lcne_log, the input path of lcne log file
      --host_log, the input path of host log file
      --device_log, the input path of device log file
      --train_log [ ...], the input paths or names of train log file
      --process_log, the input path of process log file
      --env_check, the input path of environment check file
      --dl_log, the input path of mindx-dl log file
      --mindie_log, the input path of mindie log file
      --amct_log, the input path of amct log file
      --custom_log, the input path of custom log file
      --bus_log, the input path of bus log file
    3. diag
      -i, --input_path, the input path of parsed data file.
      -o, --output_path, the output path of diag result file.
      -p, --performance, indicates whether the performance evaluation module is enabled.
      -s {host,super_pod}, --scene {host,super_pod} indicates a certain diagnosis scene.
    4. entity
      -u, --update, update by the updated data file.
      -d, --delete, the fault entity code need to be deleted.
      -s, --show, the fault entity code need to be shown.
      -c, --check, the path of the user-defined fault entity file which need to check.
      --item, the show entity item, support 'attribute', 'rule' and 'regex'. Only used with '-s --show'.
      -f, --force, indicates whether to skip the deletion confirmation step. Only used with '-d --delete'.
    5. blacklist
      -a, --add, add blacklist keywords, if there are multiple keywords, use blank space to separate.
      -d, --delete, delete blacklist by id,if there are multiple items to be deleted, use blank space to separate.
      -s, --show, show blacklist.
      -f, --file, input the custom file to overwrite the default json.
      --force, indicates whether to skip the deletion confirmation step. Only used with '-d --delete'.
    6. config
      -u, --update, update the path of the configuration file.
      -c, --check, verify the validity of user-defined configuration file
      -s, --show, show the user-defined configuration content
    7. single-diag
      -i, --input_path, the input path of origin data file.
      -o, --output_path, the output path of diag result file.
      --host_log, the input path of host log file
      --device_log, the input path of device log file
      --train_log [ ...], the input paths or names of train log file
      --process_log, the input path of process log file
      --env_check, the input path of environment check file
      --dl_log, the input path of mindx-dl log file
      --mindie_log, the input path of mindie log file
      --amct_log, the input path of amct log file
      --bus_log, the input path of bus log file
    """
    args = argparse.ArgumentParser(add_help=True, description=f"{NAME.capitalize()} Fault Diag")
    sub_arg = args.add_subparsers(dest="cmd", required=True)

    sub_arg.add_parser("version", description=f"show {COMPONENT} version", help=f"show {COMPONENT} version")

    parse_cmd = sub_arg.add_parser("parse", help="parse origin log files")
    add_parse_arguments(parse_cmd)

    diag_cmd = sub_arg.add_parser("diag", help="diag parsed log files")
    add_diag_arguments(diag_cmd)

    parse_black_cmd = sub_arg.add_parser("blacklist", help="filter invalid CANN logs by blacklist for parsing")
    parse_black_cmd.add_argument("--force", action="store_true",
                                 help="Force option after operation delete or file")
    parse_black_cmd_group = parse_black_cmd.add_mutually_exclusive_group(required=True)
    add_blacklist_arguments(parse_black_cmd_group)

    add_custom_config(sub_arg)
    add_entity_cmd(sub_arg)
    add_single_diag_cmd(sub_arg)
    return args.parse_args()


def add_custom_config(sub_arg):
    config_cmd = sub_arg.add_parser("config", help="custom configuration parsing files")
    config_cmd_group = config_cmd.add_mutually_exclusive_group(required=True)
    config_cmd_group.add_argument("-u", "--update", type=file_check, metavar='path',
                                  help="update the path of the configuration file (in JSON format), "
                                       "use this file to update the user-defined configuration information")
    config_cmd_group.add_argument("-c", "--check", action="store_true",
                                  help="verify the validity of user-defined configuration file")
    config_cmd_group.add_argument("-s", "--show", action="store_true",
                                  help="show the user-defined configuration content")


def add_blacklist_arguments(parse_black_cmd_group):
    parse_black_cmd_group.add_argument("-a", "--add", type=str_param_len_check, nargs='+', metavar='',
                                       help="add blacklist keywords, "
                                            "need to set keywords, "
                                            "if there are multiple keywords, use blank space to separate, "
                                            "e.g. --add 'ERROR' 'No Such File or directory'")
    parse_black_cmd_group.add_argument("-d", "--delete", type=int, nargs='+', metavar='',
                                       help="delete blacklist by id,"
                                            "if there are multiple items to be deleted, use blank space to separate")
    parse_black_cmd_group.add_argument("-s", "--show", action='store_true',
                                       help="show blacklist")
    parse_black_cmd_group.add_argument("-f", "--file", type=file_check,
                                       help="input the custom file to overwrite the default json")


def add_diag_arguments(diag_cmd):
    diag_cmd.add_argument("-i", "--input_path", type=dir_check, required=True, metavar='',
                          help="the input path of parsed data file")
    diag_cmd.add_argument("-p", "--performance", action="store_true",
                          help="Indicates whether the performance evaluation module is enabled. If this parameter is "
                               "specified, diag jobs related to node anomaly and network congestion are executed.")
    diag_cmd.add_argument("-o", "--output_path", type=write_path_check, required=True, metavar='',
                          help="the output path of diag result file")
    diag_cmd.add_argument("-s", "--scene", choices=DIAG_CHOICES, default="host",
                          help="diag scene: %(choices)s")


def add_parse_arguments(parse_cmd):
    parse_cmd.add_argument("-i", "--input_path", type=dir_check, metavar='',
                           help="the input path of origin data file")
    parse_cmd.add_argument("--bmc_log", type=dir_check, metavar='',
                           help="the input path of bmc log file")
    parse_cmd.add_argument("--lcne_log", type=dir_check, metavar='',
                           help="the input path of lcne log file")
    parse_cmd.add_argument("--host_log", type=dir_check, metavar='',
                           help="the input path of host log file")
    parse_cmd.add_argument("--device_log", type=dir_check, metavar='',
                           help="the input path of device log file")
    parse_cmd.add_argument("--train_log", type=file_or_dir_check, nargs='+', metavar='',
                           help="the input paths or names of train log file")
    parse_cmd.add_argument("--process_log", type=dir_check, metavar='',
                           help="the input path of process log file")
    parse_cmd.add_argument("--env_check", type=dir_check, metavar='',
                           help="the input path of environment check file")
    parse_cmd.add_argument("--dl_log", type=dir_check, metavar='',
                           help="the input path of mindx-dl log file")
    parse_cmd.add_argument("--mindie_log", type=dir_check, metavar='',
                           help="the input path of mindie log file")
    parse_cmd.add_argument("--amct_log", type=dir_check, metavar='',
                           help="the input path of amct log file")
    parse_cmd.add_argument("--bus_log", type=dir_check, metavar='',
                           help="the input path of bus log file")
    parse_cmd.add_argument("--custom_log", type=dir_check, metavar='',
                           help="the input path of custom log file")
    parse_cmd.add_argument("-p", "--performance", action="store_true",
                           help="Indicates whether the performance evaluation module is enabled. If this parameter is "
                                "specified, parse jobs related to node anomaly and network congestion are executed.")
    parse_cmd.add_argument("-o", "--output_path", type=write_path_check, required=True, metavar='',
                           help="the output path of parsed data file")


def add_entity_cmd(sub_arg):
    entity_cmd = sub_arg.add_parser("entity", help="perform operations on the user-defined faulty entity.")
    op_group = entity_cmd.add_mutually_exclusive_group(required=True)
    op_group.add_argument("-u", "--update", type=file_check, metavar='path',
                          help="the path of updated data file (json format). "
                               "Use file to update the user-defined fault entity")
    op_group.add_argument("-d", "--delete", type=code_check, nargs="+", metavar='code',
                          help="the fault entity code need to be deleted. "
                               "One or multiple codes can be transferred. Use spaces to separate multiple codes.")
    op_group.add_argument("-s", "--show", type=code_check, nargs="*", metavar='code',
                          help="the fault entity code need to be shown. "
                               "One or multiple codes can be transferred. Use spaces to separate multiple codes.")
    op_group.add_argument("-c", "--check", type=file_check, metavar='path',
                          help="the path of the user-defined fault entity file which need to check.")
    entity_cmd.add_argument("--item", choices=ITEM_CHOICES, nargs="+", metavar='item',
                            help="the entity item need to be shown, support 'attribute', 'rule' and 'regex'. "
                                 "Use spaces to separate multiple items. Only used with '-s --show'. "
                                 "Default 'attr rule regex', show all items.")
    entity_cmd.add_argument("-f", "--force", action="store_true",
                            help="indicates whether to skip the deletion confirmation step and forcibly execute "
                                 "delete command. Only used with '-d --delete'. "
                                 "Default false, re-confirmation is required.")


def add_single_diag_cmd(sub_arg):
    single_diag_cmd = sub_arg.add_parser("single-diag", help="single parse and diag log files")
    single_diag_cmd.add_argument("-i", "--input_path", type=dir_check, metavar='',
                                 help="the input path of origin data file")
    single_diag_cmd.add_argument("--host_log", type=dir_check, metavar='',
                                 help="the input path of host log file")
    single_diag_cmd.add_argument("--device_log", type=dir_check, metavar='',
                                 help="the input path of device log file")
    single_diag_cmd.add_argument("--train_log", type=file_or_dir_check, nargs='+', metavar='',
                                 help="the input paths or names of train log file")
    single_diag_cmd.add_argument("--process_log", type=dir_check, metavar='',
                                 help="the input path of process log file")
    single_diag_cmd.add_argument("--env_check", type=dir_check, metavar='',
                                 help="the input path of environment check file")
    single_diag_cmd.add_argument("--dl_log", type=dir_check, metavar='',
                                 help="the input path of mindx-dl log file")
    single_diag_cmd.add_argument("--mindie_log", type=dir_check, metavar='',
                                 help="the input path of mindie log file")
    single_diag_cmd.add_argument("--amct_log", type=dir_check, metavar='',
                                 help="the input path of amct log file")
    single_diag_cmd.add_argument("--bus_log", type=dir_check, metavar='',
                                 help="the input path of bus log file")
    single_diag_cmd.add_argument("-o", "--output_path", type=write_path_check, required=True, metavar='',
                                 help="the output path of diag result file")


def show_version():
    echo.info("%s v%s", COMPONENT, get_version())


def run(cmd):
    """
    The Component entry
    """
    op_logger.info("Start to execute cmd [%s %s].", COMPONENT, cmd)
    args = command_line()
    if args.cmd == "version":
        show_version()
    else:
        args.task_id = generate_task_id()  # set the task_id to the args instance
        init_run_log(args.task_id)
        router(args)


def init() -> bool:
    """
    Init the home path and all logger. The home path used to save custom conf file and log file
    :return: init success flag.
    """
    init_echo_log()  # the first thing is setting echo log, then can use "ECHO" logger to print anything
    home_path = init_home_path()
    if not home_path:
        return False
    try:
        clean_run_log_path()
    except (FileNotFoundError, PermissionError) as err:
        echo.warning("Clean the run log failed, the reason is: %s", err)
    init_operation_log()
    return True


def main():
    warnings.filterwarnings("ignore")
    if not init():
        return
    cmd = sys.argv[1] if len(sys.argv) > 1 else ""
    err_msg = ""
    try:
        run(cmd)
    except Exception as err:
        err_msg = f"Exception: {err}"
    except KeyboardInterrupt:
        err_msg = "KeyboardInterrupt"
    except SystemExit as err:
        # 0 means -h exit, other means err
        if err.code != 0:
            err_msg = f"exit code: {err.code}"
    finally:
        clean_child_process(os.getpid())
        if err_msg:
            echo.error("Execute cmd [%s %s] failed. The error is: [%s].", COMPONENT, cmd, err_msg)
            op_logger.error("Execute cmd [%s %s] failed. The error is: [%s].", COMPONENT, cmd, err_msg)
        else:
            op_logger.info("Execute cmd [%s %s] successfully.", COMPONENT, cmd)
