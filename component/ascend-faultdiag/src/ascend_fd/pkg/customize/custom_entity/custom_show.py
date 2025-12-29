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
import json
import logging

from ascend_fd.utils.tool import ITEM_CHOICES, safe_read_json
from ascend_fd.configuration.config import DEFAULT_USER_CONF

logger = logging.getLogger("FAULT_DIAG")
echo = logging.getLogger("ECHO")


def show_entity(code_list, show_items):
    """
    Displays the specified fault code and its attributes
    :param code_list: List[str], the code list need to show
    :param show_items: List[str], the code's entity item need to show
    """
    user_conf = safe_read_json(DEFAULT_USER_CONF) if os.path.exists(DEFAULT_USER_CONF) else {}
    user_entities = user_conf.setdefault("knowledge-repository", {})

    all_print_json = {}
    failed_json = {}
    code_list = code_list or user_entities.keys()
    for code in code_list:
        if code in all_print_json or code in failed_json:
            continue
        entity = user_entities.get(code, {})
        if not entity:
            logger.error("Can not find fault code %s in user-defined fault entity set "
                         "when the show operation is performed.", code)
            failed_json.update({code: "Not exist in user-defined fault entity set."})
            continue
        print_entity = {}
        for item in ITEM_CHOICES:
            if show_items and item not in show_items:
                continue
            if item == "regex":
                print_entity.update({"source_file": entity.get("source_file")})
            print_entity.update({item: entity.get(item)})
        all_print_json.update({code: print_entity})
    all_print_json.update(failed_json)
    echo.info(json.dumps(all_print_json, ensure_ascii=False, indent=4))
    logger.info("The entity show command is executed successfully.")
