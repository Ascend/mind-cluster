#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# Copyright 2025. Huawei Technologies Co.,Ltd. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
# ==============================================================================

import argparse
import sys
from datetime import datetime, timezone

from mindcluster_tools import __version__
from mindcluster_tools.error.error import (
    ParamError,
    DcmiReturnValueError,
    TopoMissMatchError,
    GetIpError,
)
from mindcluster_tools.rootinfo import construct_rootinfo


def show_version():
    print("mindcluster-tools version {}".format(__version__))
    print(
        "Copyright {}. Huawei Technologies Co.,Ltd. All rights reserved.".format(
            datetime.now(tz=timezone.utc).year
        )
    )


def main(args_list=None):
    parser = argparse.ArgumentParser(
        description="mindcluster-tools cli", prog="mindcluster-tools"
    )
    parser.add_argument(
        "-v", "--version", action="store_true", help="version of mindcluster-tools"
    )
    subparsers = parser.add_subparsers(
        dest="command", help="you can use subcommands as follows:"
    )
    parser_rootinfo = subparsers.add_parser("rootinfo", help="construct rootinfo.json")
    parser_rootinfo.add_argument(
        "-t", "--topo_path", type=str, required=False, help="the path of topo file"
    )
    parser_rootinfo.add_argument(
        "-r", "--rank_count", type=int, required=False, help="rank_count"
    )
    parser_rootinfo.add_argument(
        "-l", "--level_count", type=int, required=False, help="the level of network"
    )
    parser_rootinfo.add_argument(
        "--super_pod_id", type=int, required=False, help="super pod id"
    )
    parser_rootinfo.add_argument(
        "--chassis_id", type=int, required=False, help="chassis id"
    )
    args = parser.parse_args() if args_list is None else parser.parse_args(args_list)

    A5_DIE_COUNT = 2
    A5_DIE_PORT_COUNT = 9

    args.die_count = A5_DIE_COUNT
    args.die_port_count = A5_DIE_PORT_COUNT

    if args.version:
        show_version()
        exit(0)
    if not args.command:
        parser.print_help()
        sys.exit(0)
    params = dict(vars(args))
    params.pop("command")
    params.pop("version")
    try:
        ret = construct_rootinfo(params)
        print(ret)
    except ParamError as e:
        print(f"{e}, please input following parameters:")
        parser.print_help()
    except DcmiReturnValueError as e:
        print(e)
    except TopoMissMatchError as e:
        print(e)
    except OSError as e:
        print(f"{e}, please check your NPU driver")
    except GetIpError as e:
        print(e)
    except Exception:
        print("Unknown error occurred")
