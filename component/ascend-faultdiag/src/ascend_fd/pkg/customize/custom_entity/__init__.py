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
from ascend_fd.utils.status import ParamError
from ascend_fd.pkg.customize.custom_entity.custom_delete import delete_entity
from ascend_fd.pkg.customize.custom_entity.custom_update import update_entity
from ascend_fd.pkg.customize.custom_entity.custom_show import show_entity
from ascend_fd.pkg.customize.custom_entity.custom_check import check_entity


def start_entity_job(args):
    """
    Start the entity job, contain update, delete, show and check cmd
    :param args: the command-line arguments
    """
    if args.item and not isinstance(args.show, list):
        raise ParamError("The '--item' parameter can only be used together with '-s --show'.")
    if args.force and not args.delete:
        raise ParamError("The '-f --force' parameter can only be used together with '-d --delete'.")
    if isinstance(args.show, list):
        show_entity(args.show, args.item)
        return
    if args.update:
        update_entity(data_path=args.update)
        return
    if args.check:
        check_entity(args.check)
        return
    delete_entity(args.delete, args.force)
